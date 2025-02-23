package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/rand"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// executeCommand 创建一个新的 SSH 会话并执行命令
func executeCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("无法创建 SSH 会话: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}
	return string(output), nil
}

func (a *App) Uninstall(sshhost string, sshpassword string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         30 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("无法连接远程服务器: %v", err)
	}
	defer client.Close()

	// 用于收集每个步骤的结果
	results := []string{}

	// 1. 停止 5189 端口的程序
	stopCommand := "sudo kill $(sudo lsof -t -i:5189) || true"
	output, err := executeCommand(client, stopCommand)
	if err != nil {
		results = append(results, fmt.Sprintf("停止 5189 端口程序失败: %v, 输出: %s", err, output))
	} else {
		results = append(results, fmt.Sprintf("停止 5189 端口程序成功: %s", output))
	}

	// 2. 删除文件夹
	deleteFolderCommand := "sudo rm -rf /home/dist || true"
	output, err = executeCommand(client, deleteFolderCommand)
	if err != nil {
		results = append(results, fmt.Sprintf("删除文件夹失败: %v, 输出: %s", err, output))
	} else {
		results = append(results, fmt.Sprintf("删除文件夹成功: %s", output))
	}

	// 3. 删除文件 /home/agent
	deleteFileCommand := "sudo rm -f /home/agent || true"
	output, err = executeCommand(client, deleteFileCommand)
	if err != nil {
		results = append(results, fmt.Sprintf("删除文件 /home/agent 失败: %v, 输出: %s", err, output))
	} else {
		results = append(results, fmt.Sprintf("删除文件 /home/agent 成功: %s", output))
	}

	// 4. 删除文件 /home/agent.tar.gz
	deleteFileCommand2 := "sudo rm -f /home/agent.tar.gz || true"
	output, err = executeCommand(client, deleteFileCommand2)
	if err != nil {
		results = append(results, fmt.Sprintf("删除文件 /home/agent.tar.gz 失败: %v, 输出: %s", err, output))
	} else {
		results = append(results, fmt.Sprintf("删除文件 /home/agent.tar.gz 成功: %s", output))
	}
	fmt.Print(results)
	// 返回所有步骤的结果
	return "success"
}
func (a *App) Install(sshhost string, sshpassword string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         30 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("Failed to dial: %s %s %s", err, sshpassword, sshhost)
	}
	defer client.Close()

	// 创建一个新的会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("Failed to create session: %s", err)
	}
	defer session.Close()

	// 使用io.Pipe代替标准输入
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Sprintf("Failed to create stdin pipe: %s", err)
	}
	defer stdinPipe.Close()

	// 请求一个伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度 = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // 输出速度 = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Sprintf("Request for pseudo terminal failed: %s", err)
	}

	// 使用当前时间的纳秒级时间戳作为种子
	rand.Seed(time.Now().UnixNano())

	// 生成一个 0 到 99 之间的随机整数
	randomNumber := rand.String(5)

	// 启动一个命令并保持会话
	cmd := "cd /home && wget -O agent.tar.gz https://down.tapcode.work/agent.tar.gz?v=" + randomNumber + " && tar -xzvf agent.tar.gz"
	fmt.Print(cmd)
	if err := session.Start(cmd); err != nil {
		return fmt.Sprintf("Failed to start command: %s", err)
	}

	if err := session.Wait(); err != nil {
		return fmt.Sprintf("Session failed: %s", err)
	}
	return "success"
}

func (a *App) Setprivatekey(sshhost string, sshpassword string, siyao string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root", // 替换为你的 SSH 用户名
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
	}

	// 连接远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("SSH 连接失败: %v", err)
	}
	defer client.Close()

	// 备份原始配置文件
	remoteConfigPath := "/home/bot/config.yaml"
	backupConfigPath := "/home/bot/config.yaml.example"
	backupCmd := fmt.Sprintf("cp %s %s", remoteConfigPath, backupConfigPath)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建 SSH 会话失败: %v", err)
	}
	_, err = session.Output(backupCmd)
	session.Close()
	if err != nil {
		return fmt.Sprintf("备份配置文件失败: %v", err)
	}

	// 读取远程配置文件
	catCmd := fmt.Sprintf("cat %s", remoteConfigPath)
	session, err = client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建 SSH 会话失败: %v", err)
	}
	configData, err := session.Output(catCmd)
	session.Close()
	if err != nil {
		return fmt.Sprintf("读取远程配置文件失败: %v", err)
	}

	// 解析 YAML 到 map
	var config2 map[string]interface{}
	err = yaml.Unmarshal(configData, &config2)
	if err != nil {
		return fmt.Sprintf("解析 YAML 失败: %v", err)
	}

	// 修改 private_key
	if _, ok := config2["private_key"]; ok {
		config2["private_key"] = siyao
	} else {
		return "配置文件中未找到 private_key 字段"
	}

	// 确保 not_support_tokens 字段的值不为 null，并设置为空列表
	if notSupportTokens, ok := config2["not_support_tokens"]; ok {
		if notSupportTokens == nil {
			config2["not_support_tokens"] = []interface{}{}
		}
	} else {
		config2["not_support_tokens"] = []interface{}{}
	}

	// 将修改后的内容写回 YAML
	newData, err := yaml.Marshal(&config2)
	if err != nil {
		return fmt.Sprintf("生成 YAML 失败: %v", err)
	}

	// 将修改后的配置文件上传到远程服务器
	tmpFile := filepath.Join(os.TempDir(), "config_new.yaml")

	// 确保目录存在
	dir := filepath.Dir(tmpFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("创建目录失败: %v", err)
	}

	// 检查文件是否存在并删除
	if _, err := os.Stat(tmpFile); err == nil {
		if err := os.Remove(tmpFile); err != nil {
			return fmt.Sprintf("删除已存在文件失败: %v", err)
		}
	}

	// 使用 os.WriteFile 写入文件
	if err := os.WriteFile(tmpFile, newData, 0644); err != nil {
		return fmt.Sprintf("创建临时文件失败: %v", err)
	}

	// 确保文件在使用后被删除
	defer os.Remove(tmpFile)

	// 上传文件到远程服务器
	remoteTmpPath := "/home/bot/config_new.yaml"
	err = scpUpload(client, tmpFile, remoteTmpPath)
	if err != nil {
		return fmt.Sprintf("上传文件失败: %v", err)
	}

	// 替换远程配置文件
	mvCmd := fmt.Sprintf("mv %s %s", remoteTmpPath, remoteConfigPath)
	session, err = client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建 SSH 会话失败: %v", err)
	}
	_, err = session.Output(mvCmd)
	session.Close()
	if err != nil {
		return fmt.Sprintf("替换配置文件失败: %v", err)
	}

	// 检查 /home/bot/run.sh 是否存在
	checkRunShCmd := "test -f /home/bot/run.sh && echo exists || echo not_exists"
	session, err = client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建 SSH 会话失败: %v", err)
	}
	output, err := session.Output(checkRunShCmd)
	session.Close()
	if err != nil {
		return fmt.Sprintf("检查 run.sh 文件失败: %v", err)
	}

	if strings.TrimSpace(string(output)) != "exists" {
		return "run.sh 文件不存在"
	}

	// 启动 run.sh，不等待其完成
	runCmd := "cd /home/bot && chmod +x run.sh && nohup ./run.sh > /dev/null 2>&1 &"
	session, err = client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建 SSH 会话失败: %v", err)
	}
	defer session.Close()

	// 使用 Start 而不是 Output，以避免阻塞
	if err := session.Start(runCmd); err != nil {
		return fmt.Sprintf("启动 run.sh 失败: %v", err)
	}

	// 等待 /home/bot/PRIVATEKEY 文件出现
	for {
		checkCmd := "test -f /home/bot/PRIVATEKEY && echo exists || echo not_exists"
		session, err = client.NewSession()
		if err != nil {
			return fmt.Sprintf("创建 SSH 会话失败: %v", err)
		}
		output, err := session.Output(checkCmd)
		session.Close()
		if err != nil {
			return fmt.Sprintf("检查 PRIVATEKEY 文件失败: %v", err)
		}

		if strings.TrimSpace(string(output)) == "exists" {
			// 恢复备份文件到原始配置文件
			restoreCmd := fmt.Sprintf("cp %s %s", backupConfigPath, remoteConfigPath)
			session, err = client.NewSession()
			if err != nil {
				return fmt.Sprintf("创建 SSH 会话失败: %v", err)
			}
			_, err = session.Output(restoreCmd)
			session.Close()
			if err != nil {
				return fmt.Sprintf("恢复备份文件失败: %v", err)
			}

			// 删除备份文件
			rmBackupCmd := fmt.Sprintf("rm -f %s", backupConfigPath)
			session, err = client.NewSession()
			if err != nil {
				return fmt.Sprintf("创建 SSH 会话失败: %v", err)
			}
			_, err = session.Output(rmBackupCmd)
			session.Close()
			if err != nil {
				return fmt.Sprintf("删除备份文件失败: %v", err)
			}

			break
		}
		time.Sleep(1 * time.Second) // 每隔 1 秒检查一次
	}
	return "success"
}

// scpUpload 通过 SCP 上传文件到远程服务器
func scpUpload(client *ssh.Client, localPath string, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// 读取本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// 通过 SCP 上传文件
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C0644", fileInfo.Size(), filepath.Base(remotePath))
		io.Copy(w, file)
		fmt.Fprint(w, "\x00")
	}()

	return session.Run(fmt.Sprintf("scp -t %s", remotePath))
}

func (a *App) AddServerIP(sshhost string, sshpassword string, newip string, iface string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         30 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("Failed to dial: %s %s %s", err, sshpassword, sshhost)
	}
	defer client.Close()

	// 创建一个新的会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("Failed to create session: %s", err)
	}
	defer session.Close()

	// 使用io.Pipe代替标准输入
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Sprintf("Failed to create stdin pipe: %s", err)
	}
	defer stdinPipe.Close()

	// 请求一个伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度 = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // 输出速度 = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Sprintf("Request for pseudo terminal failed: %s", err)
	}

	// 启动一个命令并保持会话
	cmd := "sudo ip addr add " + newip + "/24 dev " + iface
	fmt.Print(cmd)
	if err := session.Start(cmd); err != nil {
		return fmt.Sprintf("Failed to start command: %s", err)
	}

	if err := session.Wait(); err != nil {
		return fmt.Sprintf("Session failed: %s", err)
	}
	return "success"
}
func sshSession(sshhost string, sshpassword string) (*ssh.Session, error) {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         10 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// 创建一个会话
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session, nil
}
func (a *App) CheckPort(sshhost string, sshpassword string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         10 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("Failed to dial: %s %s %s", err, sshpassword, sshhost)
	}
	defer client.Close()

	// 创建一个新的会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("Failed to create session: %s", err)
	}
	defer session.Close()

	pid, err := getProcessID(client, "5189")
	if err != nil {
		return fmt.Sprintf("Failed to find process: %s", err)
	}
	if pid == "" {
		return "No process found using port 5189."
	}
	return "success"
	// // 执行命令检查端口是否被占用
	// command := "netstat -tuln | grep :5189"
	// output, err := session.Output(command)
	// if err != nil {
	// 	fmt.Printf("Session ended successfully. %s", err)
	// 	return "false"
	// }
	// fmt.Println(string(output))
	// if len(output) > 0 {
	// 	return "success" // 端口被占用
	// } else {
	// 	return "false" // 端口未被占用
	// }
}

func (a *App) CloseServer(sshhost string, sshpassword string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         10 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("Failed to connect to SSH server: %s", err)
	}
	defer client.Close()

	// 查找占用5189端口的进程ID
	pid, err := getProcessID(client, "5189")
	if err != nil {
		return fmt.Sprintf("Failed to find process: %s", err)
	}
	if pid == "" {
		return "No process found using port 5189."
	}

	// 杀死进程
	err = killProcess(client, pid)
	if err != nil {
		return fmt.Sprintf("Failed to kill process: %s", err)
	}

	return "success"
}

// 获取占用指定端口的进程ID
func getProcessID(client *ssh.Client, port string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %s", err)
	}
	defer session.Close()

	command := fmt.Sprintf("lsof -i :%s | awk 'NR==2 {print $2}'", port)
	output, err := session.Output(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %s", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// 杀死指定进程
func killProcess(client *ssh.Client, pid string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err)
	}
	defer session.Close()

	killCommand := fmt.Sprintf("kill -9 %s", pid)
	_, err = session.Output(killCommand)
	if err != nil {
		return fmt.Errorf("failed to kill process: %s", err)
	}

	return nil
}

func (a *App) RunServer(sshhost string, sshpassword string, token string) string {
	fmt.Printf("RunServer %s %s", sshhost, sshpassword)
	// SSH 配置
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         10 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":22", config)
	if err != nil {
		return fmt.Sprintf("Failed to dial: %s %s %s", err, sshpassword, sshhost)
	}
	defer client.Close()

	// 创建一个新的会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("Failed to create session: %s", err)
	}
	defer session.Close()

	// 创建缓冲区来收集标准输出和标准错误输出
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	session.Stdin = os.Stdin

	// 请求一个伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度 = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // 输出速度 = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Sprintf("Request for pseudo terminal failed: %s", err)
	}

	// 执行后续命令（设置环境变量、启动 agent）
	startCmd := fmt.Sprintf(
		"export API_TOKEN=%s && export SSH_PWD=%s && cd /home && chmod +x agent && ./agent",
		token, sshpassword,
	)
	fmt.Printf("Executing start command: %s\n", startCmd)

	// 创建一个新的会话来执行启动命令
	startSession, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("Failed to create start session: %s", err)
	}
	defer startSession.Close()

	// 执行启动命令
	var startStdout, startStderr bytes.Buffer
	startSession.Stdout = &startStdout
	startSession.Stderr = &startStderr

	if err := startSession.Start(startCmd); err != nil {
		return fmt.Sprintf("Failed to start command: %s", err)
	}

	// 使用 WaitGroup 同步 Goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	// 使用 Goroutine 保持会话
	go func() {
		defer wg.Done()
		if err := startSession.Wait(); err != nil { // 等待会话结束
			fmt.Printf("Session ended with error: %s\n", err)
		} else {
			fmt.Println("Session ended successfully.")
		}
	}()

	// 等待 Goroutine 结束
	wg.Wait()

	// 返回启动成功的状态和日志
	fmt.Println("Server started successfully. Session is kept alive.")
	logs := fmt.Sprintf(
		"Start command stdout: %s\nStart command stderr: %s",
		startStdout.String(), startStderr.String(),
	)
	return logs
}

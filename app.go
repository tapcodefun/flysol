package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"runtime"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/rand"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
	// "os/exec"
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

func (a *App) Uninstall(sshhost string, sshpassword string, sshuser string, sshport string, private_key string) string {
	// 连接到远程服务器
	client, err := sshClient(sshhost, private_key, sshpassword, sshuser, sshport)
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
func (a *App) Install(sshhost string, privateKey string, sshpassword string, sshuser string, sshport string, directory string) string {
	client, err := sshClient(sshhost, privateKey, sshpassword, sshuser, sshport)
	if err != nil {
		return fmt.Sprintf("Failed Client: %s", err)
	}
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
	var cmd string
	if privateKey == "" {
		cmd = "cd /" + directory + " && wget -O agent.tar.gz https://down.tapcode.work/agent.tar.gz?v=" + randomNumber + " && tar -xzvf agent.tar.gz"
	} else {
		cmd = fmt.Sprintf("echo '%s' > /%s/privateKey && chmod 600 /home/privateKey && cd /home && wget -O agent.tar.gz https://down.tapcode.work/agent.tar.gz?v=%s && tar -xzvf agent.tar.gz", privateKey, directory, randomNumber)
	}
	fmt.Print(cmd)
	if err := session.Start(cmd); err != nil {
		return fmt.Sprintf("Failed to start command: %s", err)
	}

	if err := session.Wait(); err != nil {
		return fmt.Sprintf("Session failed: %s", err)
	}
	return "success"
}

func (a *App) Setprivatekey(sshhost string, sshpassword string, siyao string, sshuser string, sshport string, private_key string) string {
	// 连接远程服务器
	client, err := sshClient(sshhost, private_key, sshpassword, sshuser, sshport)
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

	return session.Run(fmt.Sprintf("scp -t -f %s", remotePath))
}
func (a *App) AddServerIP(sshhost string, sshpassword string, newip string, iface string, sshuser string, sshport string, private_key string) string {
	// 连接到远程服务器
	client, err := sshClient(sshhost, private_key, sshpassword, sshuser, sshport)
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
func sshClient(sshhost string, privateKey string, sshpassword string, sshuser string, sshport string) (*ssh.Client, error) {
	var config *ssh.ClientConfig
	var signer ssh.Signer
	var err error
	if privateKey != "" {
		// 解析带密码的私钥
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(sshpassword))
		if err != nil {
			return nil, err
		}
		// SSH 配置
		config = &ssh.ClientConfig{
			User: sshuser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
			Timeout:         30 * time.Second,            // 连接超时时间
		}
	} else {
		// SSH 配置
		config = &ssh.ClientConfig{
			User: sshuser,
			Auth: []ssh.AuthMethod{
				ssh.Password(sshpassword),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
			Timeout:         30 * time.Second,            // 连接超时时间
		}
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":"+sshport, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
func (a *App) CheckPort(sshhost string, sshpassword string, sshuser string, sshport string) string {
	// SSH 配置
	config := &ssh.ClientConfig{
		User: sshuser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshpassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查
		Timeout:         10 * time.Second,            // 连接超时时间
	}

	// 连接到远程服务器
	client, err := ssh.Dial("tcp", sshhost+":"+sshport, config)
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

type Host struct {
	Token    string  `json:"token"`
	Cpu      float64 `json:"cpu"`
	Pid      int     `json:"pid"`
	Version  string  `json:"version"`
	MasterIp string  `json:"masterIp"`
}

func (a *App) Fetchost(sshhost string) string {
	// 初始化 Host 结构体
	var host Host

	// 从 /metrics 接口提取 Token 和 Cpu
	metricsURL := "http://" + sshhost + "/metrics"
	response, err := http.Get(metricsURL)
	if err != nil {
		fmt.Printf("请求 /metrics 接口失败: %v\n", err)
		return toJSON(host) // 返回部分数据
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("读取 /metrics 响应失败: %v\n", err)
		return toJSON(host) // 返回部分数据
	}

	// 使用 map[string]interface{} 来解析 JSON 数据
	var metricsData map[string]interface{}
	if err := json.Unmarshal(body, &metricsData); err != nil {
		fmt.Printf("解析 /metrics JSON 失败: %v\n", err)
	} else {
		// 提取所需的字段
		if token, ok := metricsData["token"].(string); ok {
			host.Token = token
		}
		if cpu, ok := metricsData["cpu"].(float64); ok {
			host.Cpu = cpu
		}
		if version, ok := metricsData["version"].(string); ok {
			host.Version = version
		}
	}

	// 从 /pid 接口提取 Pid
	pidURL := "http://" + sshhost + "/pid"
	response2, err := http.Get(pidURL)
	if err != nil {
		fmt.Printf("请求 /pid 接口失败: %v\n", err)
		return toJSON(host) // 返回部分数据
	}
	defer response2.Body.Close()

	body2, err := io.ReadAll(response2.Body)
	if err != nil {
		fmt.Printf("读取 /pid 响应失败: %v\n", err)
		return toJSON(host) // 返回部分数据
	}

	var pidData struct {
		Pid int `json:"pid"`
	}
	if err := json.Unmarshal(body2, &pidData); err != nil {
		fmt.Printf("解析 /pid JSON 失败: %v\n", err)
	} else {
		host.Pid = pidData.Pid
	}

	return toJSON(host)
}

// toJSON 将 Host 结构体序列化为 JSON 字符串
func toJSON(host Host) string {
	jsonData, err := json.Marshal(host)
	if err != nil {
		fmt.Printf("序列化 JSON 失败: %v\n", err)
		return "{}" // 返回空 JSON 对象
	}
	return string(jsonData)
}
func (a *App) CloseServer(sshhost string, sshpassword string, sshuser string, sshport string, private_key string) string {
	client, err := sshClient(sshhost, private_key, sshpassword, sshuser, sshport)
	if err != nil {
		return fmt.Sprintf("Failed Client: %s", err)
	}

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
func (a *App) RunServer(sshhost string, token string, sshpassword string, sshuser string, sshport string, private_key string) string {
	fmt.Printf("RunServer %s %s", sshhost, sshpassword)
	client, err := sshClient(sshhost, private_key, sshpassword, sshuser, sshport)
	if err != nil {
		return fmt.Sprintf("Failed to dial: %s %s", err, sshpassword)
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

	sshType := "password"
	if private_key != "" {
		sshType = "private_key"
	}
	// 执行后续命令（设置环境变量、启动 agent）
	startCmd := fmt.Sprintf(
		"export API_TOKEN=%s && export SSH_PWD=%s && export SSH_USER=%s && export SSH_PORT=%s && export SSH_TYPE=%s && cd /home && chmod +x agent && ./agent",
		token, sshpassword, sshuser, sshport, sshType,
	)
	// 启动命令示例
	// export API_TOKEN=token && export SSH_PWD=SiDZtDe?nk && export SSH_USER=root && export SSH_PORT=22 && export SSH_TYPE=password && cd /home && chmod +x agent && ./agent
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

// 创建并返回一个SSH客户端
func (a *App) CreateSSHClient(host string, user string, password string, port string) (*ssh.Client, error) {
	// 检查是否是私钥认证
	var authMethod ssh.AuthMethod
	if strings.HasPrefix(password, "-----BEGIN") {
		// 尝试解析私钥
		signer, err := ssh.ParsePrivateKey([]byte(password))
		if err != nil {
			// 如果解析失败，尝试使用空密码解析带密码的私钥
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(password), []byte(""))
			if err != nil {
				// 如果仍然失败，则返回错误
				return nil, fmt.Errorf("无法解析私钥: %v", err)
			}
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		// 使用密码认证
		authMethod = ssh.Password(password)
	}

	// 创建SSH配置
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 5,
	}

	// 建立SSH连接
	return ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
}

// TransferFile 接收前端传来的文件内容，通过SSH连接上传到指定主机路径
// func (a *App) TransferFile(sshhost string, sshuser string, sshpassword string, sshport string, fileContent string, remotePath string) string {
// 	// 创建SSH连接
// 	client, err := sshClient(sshhost, sshpassword, sshuser, sshport)
// 	if err != nil {
// 		return fmt.Sprintf("SSH连接失败: %v", err)
// 	}
// 	defer client.Close()

// 	// 创建临时文件存储前端传来的文件内容
// 	tmpFile := filepath.Join(os.TempDir(), "transfer_tmp_"+rand.String(8))

// 	// 确保目录存在
// 	dir := filepath.Dir(tmpFile)
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return fmt.Sprintf("创建临时目录失败: %v", err)
// 	}

// 	// 写入文件内容到临时文件
// 	if err := os.WriteFile(tmpFile, []byte(fileContent), 0644); err != nil {
// 		return fmt.Sprintf("创建临时文件失败: %v", err)
// 	}

// 	// 确保临时文件在使用后被删除
// 	defer os.Remove(tmpFile)

// 	// 上传文件到远程服务器
// 	err = scpUpload(client, tmpFile, remotePath)
// 	if err != nil {
// 		return fmt.Sprintf("文件上传失败: %v", err)
// 	}

// 	return "success"
// }

// uploadPrivatekey 从远程主机获取私钥文件内容
// UploadFileToRemoteHost 将文件上传到远程主机
// scpUpload 通过 SCP 上传文件到远程服务器
func scpupload(client *ssh.Client, localPath string, remotePath string) error {
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
	return session.Run(fmt.Sprintf("scp -t %s", filepath.ToSlash(remotePath)))
}

func (a *App) UploadFileToRemoteHost(host string, user string, password string, port string, remoteDir string, fileName string, fileContent string, isUpzip string) string {
	// 连接到远程服务器
	client, err := sshClient(host, "", password, user, port)
	if err != nil {
		return fmt.Sprintf("SSH连接失败: %v", err)
	}
	defer client.Close()

	// 创建临时文件
	tmpFile := filepath.Join(os.TempDir(), fileName)

	// 解码Base64内容
	decoded, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		return fmt.Sprintf("文件解码失败: %v", err)
	}

	// 写入临时文件
	err = os.WriteFile(tmpFile, decoded, 0644)
	if err != nil {
		return fmt.Sprintf("临时文件创建失败: %v", err)
	}
	defer os.Remove(tmpFile) // 确保临时文件被删除

	// 确保远程目录存在
	remoteDirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建SSH会话失败: %v", err)
	}
	_, err = session.CombinedOutput(remoteDirCmd)
	session.Close()
	if err != nil {
		return fmt.Sprintf("创建远程目录失败: %v", err)
	}

	// 构建远程文件路径并转换为Linux风格的路径
	remotePath := filepath.ToSlash(filepath.Join(remoteDir, fileName))
	fmt.Printf("文件路径：%v\n", remotePath)

	// 上传文件
	err = scpupload(client, tmpFile, remotePath)
	if err != nil {
		return fmt.Sprintf("文件上传失败: %v", err)
	}

	// // 如果是压缩文件，执行解压命令
	// fileExt := strings.ToLower(filepath.Ext(fileName))
	// validCompressExts := []string{".tar.gz", ".tgz", ".gz", ".zip", ".tar"}
	// isValidCompress := false
	// for _, ext := range validCompressExts {
	//     if fileExt == ext || (ext == ".tar.gz" && strings.HasSuffix(strings.ToLower(fileName), ext)) {
	//         isValidCompress = true
	//         break
	//     }
	// }
	// if isUpzip == "true" && isValidCompress {
	//     // 创建新的SSH会话用于执行解压命令
	//     session, err := client.NewSession()
	//     if err != nil {
	//         return fmt.Sprintf("创建SSH会话失败: %v", err)
	//     }
	//     defer session.Close()

	//     // 构建解压命令
	//     extension := strings.ToLower(filepath.Ext(fileName))

	//     // 根据文件扩展名选择解压命令
	//     var unpackCmd string
	//     switch extension {
	//     case ".zip":
	//         unpackCmd = fmt.Sprintf("cd %s && unzip -o %s", remoteDir, fileName)
	//     case ".tar":
	//         unpackCmd = fmt.Sprintf("cd %s && tar -xf %s", remoteDir, fileName)
	//     case ".gz", ".tgz":
	//         unpackCmd = fmt.Sprintf("cd %s && tar -xzf %s", remoteDir, fileName)
	//     default:
	//         return fmt.Sprintf("不支持的文件格式: %s", extension)
	//     }

	//     _, err = session.CombinedOutput(unpackCmd)
	//     if err != nil {
	//         return fmt.Sprintf("文件解压失败: %v", err)
	//     }

	//     // 创建新会话列出目录内容
	//     lsSession, err := client.NewSession()
	//     if err != nil {
	//         return fmt.Sprintf("创建列目录会话失败: %v", err)
	//     }
	//     defer lsSession.Close()

	//     // 执行ls命令显示目录内容
	//     lsCmd := fmt.Sprintf("cd %s && ls -la", remoteDir)
	//     lsOutput, err := lsSession.CombinedOutput(lsCmd)
	//     if err != nil {
	//         return fmt.Sprintf("列出目录内容失败: %v", err)
	//     }
	//     fmt.Printf("解压后的目录内容:\n%s\n", string(lsOutput))
	// }

	return "success"
}

// UploadFolderToRemoteHost 上传文件夹到远程主机，保持目录结构
func (a *App) UploadFolderToRemoteHost(host string, user string, password string, port string, remoteDir string, folderPath string, folderContent map[string]string) string {
	// 连接到远程服务器
	client, err := sshClient(host, "", password, user, port)
	if err != nil {
		return fmt.Sprintf("SSH连接失败: %v", err)
	}
	defer client.Close()

	// 创建临时目录用于存放解码后的文件
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("upload_%d", time.Now().UnixNano()))
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return fmt.Sprintf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 确保临时目录被删除

	// 确保远程根目录存在
	remoteDirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Sprintf("创建SSH会话失败: %v", err)
	}
	_, err = session.CombinedOutput(remoteDirCmd)
	session.Close()
	if err != nil {
		return fmt.Sprintf("创建远程根目录失败: %v", err)
	}

	// 遍历文件内容映射
	for relativePath, fileContent := range folderContent {
		// 构建本地临时文件路径
		localFilePath := filepath.Join(tempDir, relativePath)
		localDir := filepath.Dir(localFilePath)

		// 确保本地临时目录存在
		err = os.MkdirAll(localDir, 0755)
		if err != nil {
			return fmt.Sprintf("创建本地临时子目录失败: %v", err)
		}

		// 解码Base64内容
		decoded, err := base64.StdEncoding.DecodeString(fileContent)
		if err != nil {
			return fmt.Sprintf("文件解码失败 %s: %v", relativePath, err)
		}

		// 写入临时文件
		err = os.WriteFile(localFilePath, decoded, 0644)
		if err != nil {
			return fmt.Sprintf("临时文件创建失败 %s: %v", relativePath, err)
		}

		// 构建远程目录路径
		remoteFileDir := filepath.ToSlash(filepath.Join(remoteDir, filepath.Dir(relativePath)))

		// 确保远程子目录存在
		if filepath.Dir(relativePath) != "." {
			remoteDirCmd := fmt.Sprintf("mkdir -p %s", remoteFileDir)
			session, err := client.NewSession()
			if err != nil {
				return fmt.Sprintf("创建SSH会话失败: %v", err)
			}
			_, err = session.CombinedOutput(remoteDirCmd)
			session.Close()
			if err != nil {
				return fmt.Sprintf("创建远程子目录失败 %s: %v", remoteFileDir, err)
			}
		}

		// 构建远程文件路径并转换为Linux风格的路径
		remotePath := filepath.ToSlash(filepath.Join(remoteDir, relativePath))
		fmt.Printf("上传文件：%s -> %s\n", relativePath, remotePath)

		// 上传文件
		err = scpupload(client, localFilePath, remotePath)
		if err != nil {
			return fmt.Sprintf("文件上传失败 %s: %v", relativePath, err)
		}
	}

	return "success"
}

func (a *App) UploadPrivatekey(category string, host string, user string, password string, port string) string {
	// 获取用户主目录
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("获取用户主目录失败: %v", err)
	}

	// 判断操作系统
	var privatekeyDir string
	if runtime.GOOS == "windows" {
		// 如果是 Windows 系统，将文件夹创建在 C 盘
		privatekeyDir = filepath.Join("C:\\", "privatekey")
	} else {
		// 其他操作系统将文件夹创建在桌面
		privatekeyDir = filepath.Join(userHomeDir, "Desktop", "privatekey")
	}

	// 创建privatekey主目录
	if err := os.MkdirAll(privatekeyDir, 0700); err != nil {
		return fmt.Sprintf("创建私钥目录失败: %v", err)
	}

	// 创建以分类名命名的子文件夹
	hostDir := filepath.Join(privatekeyDir, category)
	if err := os.MkdirAll(hostDir, 0700); err != nil {
		return fmt.Sprintf("创建IP子目录失败: %v", err)
	}

	// 创建SSH连接
	client, err := a.CreateSSHClient(host, user, password, port)
	if err != nil {
		return fmt.Sprintf("SSH连接失败: %v", err)
	}
	defer client.Close()

	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Sprintf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	// 打开远程文件
	remoteFile, err := sftpClient.Open("/home/bot/PRIVATE_KEY")
	if err != nil {
		return fmt.Sprintf("打开远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	// 生成唯一的文件名
	localFileName := fmt.Sprintf("%v", host)
	localFilePath := filepath.Join(hostDir, localFileName)

	// 创建本地文件
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Sprintf("创建本地文件失败: %v", err)
	}
	defer localFile.Close()

	// 复制文件内容
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Sprintf("下载文件失败: %v", err)
	}

	return fmt.Sprintf("私钥文件已成功保存到: %s", localFilePath)
}

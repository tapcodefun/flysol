package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
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

	// 启动一个命令并保持会话
	cmd := "cd /home && wget https://down.tapcode.work/agent.tar.gz && tar -xzvf agent.tar.gz"
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
func (a *App) RunServer(sshhost string, sshpassword string) string {
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

	// 设置标准输入、输出和错误输出
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
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

	// 启动一个命令并保持会话
	cmd := "export API_TOKEN=nextcoderdev && export SSH_PWD=" + sshpassword + " && cd /home && chmod +x agent && ./agent"
	fmt.Print(cmd)
	if err := session.Start(cmd); err != nil {
		return fmt.Sprintf("Failed to start command: %s", err)
	}

	// 使用 WaitGroup 同步 Goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	// 使用 Goroutine 保持会话
	go func() {
		defer wg.Done()
		if err := session.Wait(); err != nil { // 等待会话结束
			fmt.Printf("Session ended with error: %s\n", err)
		} else {
			fmt.Println("Session ended successfully.")
		}
	}()

	// 返回启动成功的状态
	fmt.Println("Server started successfully. Session is kept alive.")
	wg.Wait() // 等待 Goroutine 结束
	return "success"
}

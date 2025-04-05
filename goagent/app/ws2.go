package app

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func Runcmd(c *gin.Context) {
	cmd := "wget https://sourceforge.net/projects/rust-mev-bot/files/rust-mev-bot-1.0.5.zip"
	sshClient, err := connectToRemoteServer()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "连接失败"})
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "session失败"})
		return
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		log.Printf("命令执行失败: %s, stderr: %s", err, stderr.String())
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "执行失败"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": false, "message": stdout.String()})
	}
}

// 获取所有 screen 会话的 ID
func getScreenSessionIDs() ([]string, error) {
	// 执行 screen -ls 命令
	cmd := exec.Command("screen", "-ls")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析输出，提取会话 ID
	var sessionIDs []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 检查行是否包含会话信息
		if strings.Contains(line, "\t") && strings.Contains(line, ".") {
			// 提取会话 ID（例如：12345.pts-0.hostname）
			sessionID := strings.Fields(line)[0]
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

func Screen(c *gin.Context) {
	status, err := checkCmd("screen")
	if status == "未安装" {
		c.JSON(http.StatusOK, gin.H{
			"error": "unstall",
		})
		return
	}
	sessionIDs, err := getScreenSessionIDs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to check screen sessions: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"screens": sessionIDs,
		"error":   "",
	})
}

func SaveConfig(c *gin.Context) {
	// 定义一个结构体来绑定 JSON 数据
	var jsonData struct {
		File    string `json:"file"`
		Content string `json:"content"`
	}

	// 绑定 JSON 数据到结构体
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 读取 SFTP 文件内容
	content, err := saveSFTPFile("/home/bot/config.yaml", jsonData.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "message": content})
}

func GetConfig(c *gin.Context) {
	// 读取 SFTP 文件内容
	content, err := readSFTPFile("/home/bot/config.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "message": content})
}

func saveSFTPFile(remoteFilePath string, yaml string) (string, error) {
	sshClient, err := connectToRemoteServer()
	if err != nil {
		return "", fmt.Errorf("fSSH连接失败: %v", err)
	}
	defer sshClient.Close()

	// 创建 SFTP 客户端
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		log.Fatalf("Failed to create SFTP client: %v", err)
	}
	defer client.Close()

	file, err := client.Create(remoteFilePath)
	if err != nil {
		log.Fatalf("Failed to create remote file: %v", err)
	}
	defer file.Close()

	// 要写入的内容
	content := []byte(yaml)

	// 将内容写入文件
	_, err = file.Write(content)
	if err != nil {
		log.Fatalf("Failed to write to remote file: %v", err)
	}

	return "", nil
}

func readSFTPFile(remoteFilePath string) (string, error) {
	sshClient, err := connectToRemoteServer()
	if err != nil {
		return "", fmt.Errorf("fSSH连接失败: %v", err)
	}
	defer sshClient.Close()

	// 创建 SFTP 客户端
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return "", fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer client.Close()

	// 打开远程文件
	file, err := client.Open(remoteFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open remote file: %v", err)
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}

func executeRemoteCommand2(command string, sshClient *ssh.Client, conn *websocket.Conn) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		return fmt.Errorf("创建终端失败: %v", err)
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	if err := session.Start(command); err != nil {
		return fmt.Errorf("启动命令失败: %v", err)
	}

	go io.Copy(&wsWriter{conn}, stdout)
	go io.Copy(&wsWriter{conn}, stderr)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket读取错误:", err)
				return
			}
			msg := strings.TrimSpace(string(message))
			switch msg {
			case "Ctrl+C":
				stdin.Write([]byte("\x03"))
			case "Ctrl+A+D":
				stdin.Write([]byte("\x01d"))
			default:
				stdin.Write(append(message, '\n'))
			}
		}
	}()
	return session.Wait()
}

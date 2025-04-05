package app

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var mu sync.Mutex

func Which(c *gin.Context) {
	cmd := c.Query("cmd")
	message, err := checkCmd(cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "未安装"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": false, "message": message})
	}
}

func WebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级错误:", err)
		return
	}
	defer conn.Close()

	sshClient, err := connectToRemoteServer()
	if err != nil {
		log.Printf("SSH连接失败: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("SSH连接失败"))
		return
	}
	defer sshClient.Close()

	msgChan := make(chan []byte)
	defer close(msgChan)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket读取错误:", err)
				return
			}
			msgChan <- message
		}
	}()

	for {
		select {
		case message := <-msgChan:
			mu.Lock()
			msg := strings.TrimSpace(string(message))
			log.Printf("收到命令: %s", msg)

			if handlePreChecks(msg, conn) {
				mu.Unlock()
				continue
			}

			if err := executeRemoteCommand(msg, sshClient, conn, msgChan); err != nil {
				log.Printf("命令执行错误: %v", err)
				conn.WriteMessage(websocket.TextMessage, []byte("错误: "+err.Error()))
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte("执行完成!"))
			}
			mu.Unlock()
		}
	}
}

func WebSocket2(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级错误:", err)
		return
	}
	defer conn.Close()

	sshClient, err := connectToRemoteServer()
	if err != nil {
		log.Printf("SSH连接失败: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("SSH连接失败"))
		return
	}
	defer sshClient.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket读取错误2:", err)
			return
		}
		session, err := sshClient.NewSession()
		if err != nil {
			log.Printf("创建会话失败: %v", err)
			return
		}
		defer session.Close()

		if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
			log.Printf("创建终端失败: %v", err)
			return
		}

		stdin, _ := session.StdinPipe()
		stdout, _ := session.StdoutPipe()
		stderr, _ := session.StderrPipe()

		if err := session.Start(string(message)); err != nil {
			log.Printf("启动命令失败: %v", err)
			return
		}

		go io.Copy(&wsWriter{conn}, stdout)
		go io.Copy(&wsWriter{conn}, stderr)
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("WebSocket22读取错误:", err)
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
		session.Wait()
		return
	}
}

func executeRemoteCommand(command string, sshClient *ssh.Client, conn *websocket.Conn, msgChan <-chan []byte) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		return fmt.Errorf("创建终端失败: %v", err)
	}

	//stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	if err := session.Start(command); err != nil {
		return fmt.Errorf("启动命令失败: %v", err)
	}

	go io.Copy(&wsWriter{conn}, stdout)
	go io.Copy(&wsWriter{conn}, stderr)

	return session.Wait()
}

func handlePreChecks(msg string, conn *websocket.Conn) bool {
	if strings.Contains(msg, "wget") && strings.Contains(msg, "rust-mev-bot") {
		lastSlash := strings.LastIndex(msg, "/")
		target := strings.TrimSpace(msg[lastSlash+1:])
		if exists, _ := checkFileExists("/home/" + target); exists {
			conn.WriteMessage(websocket.TextMessage, []byte("文件已存在"))
			return true
		}
	}

	if strings.Contains(msg, "unzip") && strings.Contains(msg, "rust-mev-bot") {
		if exists, _ := checkFileExists("/home/bot"); exists {
			conn.WriteMessage(websocket.TextMessage, []byte("已解压"))
			return true
		}
	}

	if strings.Contains(msg, "config.yaml.example") {
		if exists, _ := checkFileExists("/home/bot/config.yaml"); exists {
			conn.WriteMessage(websocket.TextMessage, []byte("配置文件已存在"))
			return true
		}
	}

	if msg == "apt install unzip" {
		if _, err := checkCmd("unzip"); err == nil {
			conn.WriteMessage(websocket.TextMessage, []byte("已安装"))
			return true
		}
	}

	if msg == "apt install screen" {
		if _, err := checkCmd("screen"); err == nil {
			conn.WriteMessage(websocket.TextMessage, []byte("已安装"))
			return true
		}
	}
	return false
}

func checkCmd(cmd string) (string, error) {
	sshClient, err := connectToRemoteServer()
	if err != nil {
		return "连接错误", err
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return "会话错误", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	if err := session.Run("which " + cmd); err != nil {
		return "未安装", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func checkFileExists(filePath string) (bool, error) {
	conn, err := connectToRemoteServer()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return false, err
	}
	defer client.Close()

	_, err = client.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func connectToRemoteServer() (*ssh.Client, error) {
	port := os.Getenv("SSH_PORT")
	if port == "" {
		port = "22"
	}

	user := os.Getenv("SSH_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("SSH_PWD")
	if password == "" {
		return nil, fmt.Errorf("SSH_PWD environment variable is not set")
	}

	client, err := ssh.Dial("tcp", "127.0.0.1:"+port, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return client, nil
}

type wsWriter struct{ conn *websocket.Conn }

func (w *wsWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

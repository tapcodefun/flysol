package app

import (
	gonet "net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type iptype struct {
	Iface string `json:"iface"`
	Ip    string `json:"ip"`
}

func getLocalIPs() []iptype {
	var ips []iptype

	// 获取所有网络接口
	interfaces, err := gonet.Interfaces()
	if err != nil {
		return ips
	}

	// 遍历每个网络接口
	for _, iface := range interfaces {
		// 获取该接口的IP地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		// 打印每个IP地址
		for _, addr := range addrs {
			if ipnet, ok := addr.(*gonet.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, iptype{
						Iface: iface.Name,
						Ip:    ipnet.IP.String(),
					})
				}
			}
		}
	}
	return ips
}
func PidHandler(c *gin.Context) {
	connections, err := net.Connections("tcp")
	if err != nil {
		data := map[string]interface{}{
			"pid": -1,
		}
		c.JSON(http.StatusOK, data)
		return
	}

	for _, conn := range connections {
		if conn.Status == "LISTEN" && int(conn.Laddr.Port) == 5189 {
			data := map[string]interface{}{
				"pid": int(conn.Pid),
			}
			c.JSON(http.StatusOK, data)
			return
		}
	}
	data := map[string]interface{}{
		"pid": 0,
	}
	c.JSON(http.StatusOK, data)
}
func MetricsHandler(c *gin.Context) {
	// 收集数据
	cpuPercent, _ := cpu.Percent(0, false)
	memInfo, _ := mem.VirtualMemory()
	ips := getLocalIPs()
	data := map[string]interface{}{
		"version": "1.0.5",
		"token":   os.Getenv("API_TOKEN"),
		"cpu":     cpuPercent[0],
		"memory":  memInfo,
		"ips":     ips,
	}
	c.JSON(http.StatusOK, data)
}

type ProcessConnect struct {
	Type   string   `json:"type"`
	Status string   `json:"status"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
	PID    int32    `json:"PID"`
	Name   string   `json:"name"`
}

var NetTypes = [...]string{"tcp", "udp"}

func ProgressHandler(c *gin.Context) {
	var err error
	var (
		result []ProcessConnect
		proc   *process.Process
	)
	for _, netType := range NetTypes {
		connections, _ := net.Connections(netType)
		if err == nil {
			for _, conn := range connections {
				proc, err = process.NewProcess(conn.Pid)
				if err == nil {
					name, _ := proc.Name()
					result = append(result, ProcessConnect{
						Type:   netType,
						Status: conn.Status,
						Laddr:  conn.Laddr,
						Raddr:  conn.Raddr,
						PID:    conn.Pid,
						Name:   name,
					})
				}
			}
		}
	}
	c.JSON(http.StatusOK, result)
}

package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nextcoder/hostagent/app"
)

func main() {
	r := gin.Default()
	// 配置 CORS 中间件
	r.Use(cors.Default())

	// 设置静态文件服务，将 dist 目录作为静态文件根目录
	r.Static("/assets", "./dist/assets")

	// 加载 index.html
	r.GET("/", func(c *gin.Context) {
		code := c.Query("code")
		expectedToken := os.Getenv("API_TOKEN")
		if expectedToken == code {
			c.File("./dist/index.html")
		} else {
			c.String(200, "API_TOKEN不正确")
		}
	})

	// 注册路由
	r.GET("/metrics", cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}), app.MetricsHandler)
	r.GET("/pid", cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}), app.PidHandler)
	r.GET("/progress", app.ProgressHandler)
	r.GET("/runcmd", app.Runcmd)
	r.GET("/which", app.Which)
	r.GET("/screen", app.Screen)
	r.GET("/ws", app.WebSocket)
	r.GET("/wsrun", app.WebSocket2)
	r.GET("/config", app.GetConfig)
	r.POST("/config", app.SaveConfig)

	r.Run(":5189")
}

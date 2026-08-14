package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	//:=写法简短形式省去了类型，表示让go编译器自动判断类型
	//以下写法等于var router = gin.Default()
	//创建默认的Gin引擎
	router := gin.Default()

	//定义路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	//启动u服务器,默认端口8080
	router.Run() // listens on 0.0.0.0:8080 by default

}

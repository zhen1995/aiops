package main

import (
	"path/filepath"
	"runtime"

	"aiops/configs"
	"aiops/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// getConfigPath 获取配置文件路径（自动定位到项目根目录）
func getConfigPath() string {
	// 获取当前文件(main.go)的绝对路径
	_, filename, _, _ := runtime.Caller(0)
	// main.go 位于 backend/cmd/center/main.go，向上回退两级到 backend 目录
	dir := filepath.Dir(filename)
	dir = filepath.Join(dir, "..", "..")
	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic("获取项目根目录失败: " + err.Error())
	}
	return filepath.Join(absDir, "configs", "config.yaml")
}

func main() {
	// 加载配置文件
	cfg, err := configs.LoadConfig(getConfigPath())
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败")
	}

	//:=写法简短形式省去了类型，表示让go编译器自动判断类型
	//以下写法等于var router = gin.Default()
	//创建默认的Gin引擎
	router := gin.Default()

	//定义路由
	router.GET("/login", controllers.LoginHandler(db))

	//启动服务器,默认端口8080
	router.Run(cfg.Server.Port)
}

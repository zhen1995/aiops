package main

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"runtime"

	"aiops/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// getConfigPath 获取配置文件路径（自动定位到项目根目录）
func getConfigPath() string {
	// 获取当前文件(main.go)的绝对路径
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "config.yaml")
}

func main() {
	// 加载配置文件
	cfg, err := LoadConfig(getConfigPath())
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
	router.GET("/login", func(c *gin.Context) {
		username := c.Query("username")
		password := c.Query("password")
		//gorm 使用泛型API查询
		ctx := context.Background()
		var result string
		sysUser, err := gorm.G[model.SysUser](db).Where("username = ?", username).First(ctx)
		if err != nil {
			//如果记录不存在
			if errors.Is(err, gorm.ErrRecordNotFound) {
				result = "不存在此用户"
				c.JSON(http.StatusOK, gin.H{
					"message": result,
				})
				return
			}
			println("查询报错了")
		}

		if password == sysUser.Password {
			result = "登录成功"
		} else {
			result = "登录失败,用户名或者密码错误"
		}
		c.JSON(http.StatusOK, gin.H{
			"message": result,
		})
	})
	//启动u服务器,默认端口8080
	router.Run(cfg.Server.Port)

}

package controllers

import (
	"context"
	"errors"
	"net/http"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginHandler 处理 /login 登录请求
func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Query("username")
		password := c.Query("password")
		//gorm 使用泛型API查询
		ctx := context.Background()
		var result string
		sysUser, err := gorm.G[models.SysUser](db).Where("username = ?", username).First(ctx)
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
	}
}

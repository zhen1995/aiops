package controllers

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// tokenStore 内存 Token 存储（token -> 用户信息）
type tokenStore struct {
	sync.Map
}

// TokenInfo Token 对应的用户信息
type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// GlobalTokenStore 全局 Token 存储
var GlobalTokenStore = &tokenStore{}

const tokenExpireDuration = 24 * time.Hour

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthController 认证控制器
type AuthController struct {
	DB *gorm.DB
}

// NewAuthController 创建认证控制器
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "用户名和密码不能为空",
		})
		return
	}

	var user models.SysUser
	if err := c.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"message": "用户名或密码错误",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "服务器内部错误",
		})
		return
	}

	// 验证密码
	if user.Password != req.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 生成 Token
	token := uuid.New().String()
	info := TokenInfo{
		UserID:    user.ID,
		Username:  user.Username,
		Name:      user.Name,
		CreatedAt: time.Now(),
	}
	GlobalTokenStore.Store(token, info)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"name":     user.Name,
			},
		},
		"message": "登录成功",
	})
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	// 支持 "Bearer <token>" 格式
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if token != "" {
		GlobalTokenStore.Delete(token)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "登出成功",
	})
}

// GetUserInfo 获取当前用户信息（根据 Token）
func (c *AuthController) GetUserInfo(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	value, ok := GlobalTokenStore.Load(token)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"message": "登录已过期，请重新登录",
		})
		return
	}

	info := value.(TokenInfo)
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"id":       info.UserID,
			"username": info.Username,
			"name":     info.Name,
		},
	})
}
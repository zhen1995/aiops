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

// getUserPermissions 获取指定用户的所有权限名称列表
func (c *AuthController) getUserPermissions(userID string) []string {
	permissions := make([]string, 0)

	// 查询用户-角色关联
	var userRoles []models.SysUserRoleRelation
	if err := c.DB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return permissions
	}

	// 收集角色 ID
	roleIDs := make([]string, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	if len(roleIDs) == 0 {
		return permissions
	}

	// 查询角色-权限关联，通过 sys_role_auth_relation 的 auth_id 关联 sys_auth 获取权限名称
	var roleAuths []models.SysRoleAuthRelation
	if err := c.DB.Where("role_id IN ?", roleIDs).Find(&roleAuths).Error; err != nil {
		return permissions
	}

	// 如果 role_auth_relation 中已冗余存储了 auth_name，直接使用
	// 否则通过 auth_id 查询 sys_auth 表获取 name
	if len(roleAuths) > 0 && roleAuths[0].AuthName != "" {
		seen := make(map[string]bool)
		for _, ra := range roleAuths {
			if ra.AuthName != "" && !seen[ra.AuthName] {
				seen[ra.AuthName] = true
				permissions = append(permissions, ra.AuthName)
			}
		}
	} else {
		// 通过 auth_id 查询 sys_auth 表
		authIDs := make([]string, 0, len(roleAuths))
		for _, ra := range roleAuths {
			authIDs = append(authIDs, ra.AuthID)
		}
		var auths []models.SysAuth
		if err := c.DB.Where("id IN ?", authIDs).Find(&auths).Error; err != nil {
			return permissions
		}
		seen := make(map[string]bool)
		for _, a := range auths {
			if a.Name != "" && !seen[a.Name] {
				seen[a.Name] = true
				permissions = append(permissions, a.Name)
			}
		}
	}

	return permissions
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名和密码不能为空",
		})
		return
	}

	var user models.SysUser
	if err := c.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户名或密码错误",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "服务器内部错误",
		})
		return
	}

	// 验证密码
	if user.Password != req.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
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

	// 获取用户权限
	permissions := c.getUserPermissions(user.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":          user.ID,
				"username":    user.Username,
				"name":        user.Name,
				"permissions": permissions,
			},
		},
		"message": "登录成功",
	})
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if token != "" {
		GlobalTokenStore.Delete(token)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
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
			"code":    401,
			"message": "登录已过期，请重新登录",
		})
		return
	}

	info := value.(TokenInfo)

	// 获取用户权限
	permissions := c.getUserPermissions(info.UserID)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"id":          info.UserID,
			"username":    info.Username,
			"name":        info.Name,
			"permissions": permissions,
		},
	})
}

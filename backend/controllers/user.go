package controllers

import (
	"errors"
	"net/http"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController 用户管理控制器
type UserController struct {
	DB *gorm.DB
}

// NewUserController 创建控制器
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// UserListRequest 用户列表查询参数
type UserListRequest struct {
	Keyword string `form:"keyword" binding:"omitempty"`
}

// List 获取用户列表
func (c *UserController) List(ctx *gin.Context) {
	var req UserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
		})
		return
	}

	var users []models.SysUser
	query := c.DB.Model(&models.SysUser{})

	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR name LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": users,
	})
}

// Get 根据 ID 获取用户
func (c *UserController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.SysUser
	if err := c.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "用户不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}

// CreateRequest 创建用户请求
type CreateRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// Create 创建用户
func (c *UserController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existing models.SysUser
	if err := c.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "用户名已存在",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	user := models.SysUser{
		Username: req.Username,
		Password: req.Password,
		Name:     req.Name,
	}

	if err := c.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "创建失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}

// UpdateRequest 更新用户请求
type UpdateRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

// Update 更新用户
func (c *UserController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var user models.SysUser
	if err := c.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "用户不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 如果修改了用户名，检查是否重复
	if req.Username != "" && req.Username != user.Username {
		var existing models.SysUser
		if err := c.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"message": "用户名已存在",
			})
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"message": "查询失败",
			})
			return
		}
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	now := time.Now()
	user.UpdateAt = &now

	if err := c.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&user, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}

// Delete 删除用户（软删除）
func (c *UserController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var user models.SysUser
	if err := c.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "用户不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	// 不允许删除自己
	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	if value, ok := GlobalTokenStore.Load(token); ok {
		if info, ok := value.(TokenInfo); ok && info.UserID == id {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"message": "不能删除自己",
			})
			return
		}
	}

	if err := c.DB.Delete(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "删除失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// ResetPassword 重置用户密码
func (c *UserController) ResetPassword(ctx *gin.Context) {
	id := ctx.Param("id")

	var user models.SysUser
	if err := c.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "用户不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
		})
		return
	}

	now := time.Now()
	user.Password = req.Password
	user.UpdateAt = &now

	if err := c.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "重置密码失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "密码重置成功",
	})
}

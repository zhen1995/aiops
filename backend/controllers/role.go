package controllers

import (
	"errors"
	"net/http"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoleController 角色管理控制器
type RoleController struct {
	DB *gorm.DB
}

// NewRoleController 创建控制器
func NewRoleController(db *gorm.DB) *RoleController {
	return &RoleController{DB: db}
}

// RoleListItem 角色列表项（含成员数和权限名）
type RoleListItem struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	CreatedAt   *time.Time `json:"created_at"`
	Updated     *time.Time `json:"updated"`
	UserCount   int64      `json:"userCount"`
	Permissions []string   `json:"permissions"`
}

// List 获取角色列表
func (c *RoleController) List(ctx *gin.Context) {
	var roles []models.SysRole
	if err := c.DB.Order("created_at DESC").Find(&roles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询角色列表失败",
		})
		return
	}

	result := make([]RoleListItem, 0, len(roles))
	for _, role := range roles {
		item := RoleListItem{
			ID:        role.ID,
			Name:      role.Name,
			CreatedAt: role.CreatedAt,
			Updated:   role.Updated,
		}

		// 统计成员数
		var count int64
		c.DB.Model(&models.SysUserRoleRelation{}).
			Where("role_id = ?", role.ID).
			Count(&count)
		item.UserCount = count

		// 获取权限列表
		var relations []models.SysRoleAuthRelation
		c.DB.Where("role_id = ?", role.ID).Find(&relations)
		for _, r := range relations {
			if r.AuthName != "" {
				item.Permissions = append(item.Permissions, r.AuthName)
			}
		}

		result = append(result, item)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// Get 根据 ID 获取角色详情
func (c *RoleController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var role models.SysRole
	if err := c.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 获取已分配的权限 ID 列表
	var relations []models.SysRoleAuthRelation
	c.DB.Where("role_id = ?", id).Find(&relations)
	authIDs := make([]string, 0, len(relations))
	for _, r := range relations {
		authIDs = append(authIDs, r.AuthID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"id":        role.ID,
			"name":      role.Name,
			"createdAt": role.CreatedAt,
			"updated":   role.Updated,
			"authIds":   authIDs,
		},
	})
}

// CreateRequest 创建/更新角色请求
type CreateRoleRequest struct {
	Name string `json:"name" binding:"required"`
}

// Create 创建角色
func (c *RoleController) Create(ctx *gin.Context) {
	var req CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	role := models.SysRole{Name: req.Name}
	if err := c.DB.Create(&role).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": role})
}

// Update 更新角色
func (c *RoleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var role models.SysRole
	if err := c.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	var req CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	role.Name = req.Name
	now := time.Now()
	role.Updated = &now

	if err := c.DB.Save(&role).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": role})
}

// Delete 删除角色（软删除）
func (c *RoleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	var role models.SysRole
	if err := c.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 检查是否有用户关联此角色
	var userCount int64
	c.DB.Model(&models.SysUserRoleRelation{}).Where("role_id = ?", id).Count(&userCount)
	if userCount > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该角色下还有用户，无法删除",
		})
		return
	}

	if err := c.DB.Delete(&role).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// SetAuthsRequest 设置角色权限请求
type SetAuthsRequest struct {
	AuthIDs []string `json:"authIds"`
}

// SetAuths 设置角色的权限关联
func (c *RoleController) SetAuths(ctx *gin.Context) {
	id := ctx.Param("id")
	var role models.SysRole
	if err := c.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	var req SetAuthsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 使用事务
	err := c.DB.Transaction(func(tx *gorm.DB) error {
		// 删除旧的权限关联
		if err := tx.Where("role_id = ?", id).Delete(&models.SysRoleAuthRelation{}).Error; err != nil {
			return err
		}

		// 添加新的权限关联
		for _, authID := range req.AuthIDs {
			// 获取权限名称
			var auth models.SysAuth
			if err := tx.First(&auth, "id = ?", authID).Error; err != nil {
				continue // 跳过不存在的权限
			}

			rel := models.SysRoleAuthRelation{
				RoleID:   id,
				AuthID:   authID,
				RoleName: role.Name,
				AuthName: auth.Name,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "设置权限失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "权限设置成功"})
}

// ListAuths 获取所有权限列表（供分配使用）
func (c *RoleController) ListAuths(ctx *gin.Context) {
	var auths []models.SysAuth
	if err := c.DB.Where("type = ?", 1).Order("created_at ASC").Find(&auths).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询权限列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": auths})
}

// ListRolesSimple 获取所有角色简要列表（供用户分配使用）
func (c *RoleController) ListRolesSimple(ctx *gin.Context) {
	var roles []models.SysRole
	if err := c.DB.Order("created_at ASC").Find(&roles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询角色列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": roles})
}

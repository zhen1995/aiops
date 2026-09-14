package controllers

import (
	"errors"
	"net/http"
	"strings"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BusinessGroupController 业务分组控制器：对告警规则进行分组管理
type BusinessGroupController struct {
	DB *gorm.DB
}

// NewBusinessGroupController 创建控制器
func NewBusinessGroupController(db *gorm.DB) *BusinessGroupController {
	return &BusinessGroupController{DB: db}
}

// businessGroupItem 列表项：分组 + 关联规则数
type businessGroupItem struct {
	models.BusinessGroup
	RuleCount int64 `json:"rule_count"`
}

// List 获取业务分组列表（含每个分组下的告警规则数量）
func (c *BusinessGroupController) List(ctx *gin.Context) {
	var groups []models.BusinessGroup
	if err := c.DB.Order("created_at asc").Find(&groups).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询业务分组失败", "error": err.Error()})
		return
	}

	items := make([]businessGroupItem, 0, len(groups))
	for _, g := range groups {
		var count int64
		if err := c.DB.Model(&models.AlertRule{}).Where("group_id = ?", g.ID).Count(&count).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "统计分组规则数失败", "error": err.Error()})
			return
		}
		items = append(items, businessGroupItem{BusinessGroup: g, RuleCount: count})
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// validateGroupName 校验分组名称：非空且在未删除记录中唯一（软删除记录不占用名称）
func (c *BusinessGroupController) validateGroupName(name string, excludeID string) error {
	if name == "" {
		return errors.New("分组名称不能为空")
	}
	var count int64
	q := c.DB.Model(&models.BusinessGroup{}).Where("name = ?", name)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("分组名称已存在")
	}
	return nil
}

// Create 创建业务分组
func (c *BusinessGroupController) Create(ctx *gin.Context) {
	var group models.BusinessGroup
	if err := ctx.ShouldBindJSON(&group); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	group.Name = strings.TrimSpace(group.Name)
	if err := c.validateGroupName(group.Name, ""); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := c.DB.Create(&group).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建业务分组失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": group})
}

// Update 重命名业务分组
func (c *BusinessGroupController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var group models.BusinessGroup
	if err := c.DB.First(&group, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "业务分组不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询业务分组失败", "error": err.Error()})
		return
	}

	var updateData struct {
		Name string `json:"name"`
	}
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	name := strings.TrimSpace(updateData.Name)
	if err := c.validateGroupName(name, id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := c.DB.Model(&group).Update("name", name).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新业务分组失败", "error": err.Error()})
		return
	}
	group.Name = name
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": group})
}

// Delete 删除业务分组（软删除）：先将该分组下告警规则的 group_id 置空，避免悬空引用
func (c *BusinessGroupController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var group models.BusinessGroup
	if err := c.DB.First(&group, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "业务分组不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询业务分组失败", "error": err.Error()})
		return
	}

	if err := c.DB.Model(&models.AlertRule{}).Where("group_id = ?", id).Update("group_id", "").Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解除分组与规则的关联失败", "error": err.Error()})
		return
	}

	if err := c.DB.Delete(&group).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除业务分组失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

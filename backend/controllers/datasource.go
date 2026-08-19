package controllers

import (
	"errors"
	"net/http"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DatasourceController 数据源控制器
type DatasourceController struct {
	DB *gorm.DB
}

// NewDatasourceController 创建控制器
func NewDatasourceController(db *gorm.DB) *DatasourceController {
	return &DatasourceController{DB: db}
}

// List 获取所有数据源列表
func (c *DatasourceController) List(ctx *gin.Context) {
	var list []models.Datasource
	if err := c.DB.Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": list,
	})
}

// Get 根据 ID 获取单个数据源
func (c *DatasourceController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var ds models.Datasource
	if err := c.DB.First(&ds, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "数据源不存在",
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
		"data": ds,
	})
}

// Create 创建数据源
func (c *DatasourceController) Create(ctx *gin.Context) {
	var ds models.Datasource
	if err := ctx.ShouldBindJSON(&ds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
			"error": err.Error(),
		})
		return
	}

	if err := c.DB.Create(&ds).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "创建失败",
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": ds,
	})
}

// Update 更新数据源
func (c *DatasourceController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var ds models.Datasource
	if err := c.DB.First(&ds, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "数据源不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	var updateData models.Datasource
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
			"error": err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"name":        updateData.Name,
		"type":        updateData.Type,
		"url":         updateData.URL,
		"timeout":     updateData.Timeout,
		"username":    updateData.Username,
		"password":    updateData.Password,
		"is_skip_ssl": updateData.IsSkipSSL,
		"remark":      updateData.Remark,
		"is_enabled":  updateData.IsEnabled,
	}

	if err := c.DB.Model(&ds).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "更新失败",
			"error": err.Error(),
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&ds, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": ds,
	})
}

// Delete 删除数据源（软删除）
func (c *DatasourceController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var ds models.Datasource
	if err := c.DB.First(&ds, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "数据源不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	if err := c.DB.Delete(&ds).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "删除失败",
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "删除成功",
	})
}

// ToggleEnabled 切换启用状态
func (c *DatasourceController) ToggleEnabled(ctx *gin.Context) {
	id := ctx.Param("id")

	var ds models.Datasource
	if err := c.DB.First(&ds, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "数据源不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	newStatus := 0
	if ds.IsEnabled == 0 {
		newStatus = 1
	}

	if err := c.DB.Model(&ds).Update("is_enabled", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "操作失败",
			"error": err.Error(),
		})
		return
	}

	if err := c.DB.First(&ds, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": ds,
	})
}

// TypeList 获取数据源类型列表
func (c *DatasourceController) TypeList(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": models.DatasourceTypeOptions,
	})
}
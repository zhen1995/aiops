package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aiops/internal/notify"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotifyMediaController 通知媒介控制器
type NotifyMediaController struct {
	DB *gorm.DB
}

// NewNotifyMediaController 创建控制器
func NewNotifyMediaController(db *gorm.DB) *NotifyMediaController {
	return &NotifyMediaController{DB: db}
}

// validateConfig 校验媒介类型与配置 JSON 是否合法
func validateMediaConfig(mediaType, rawConfig string) (*notify.MediaConfig, error) {
	if mediaType != models.NotifyMediaTypeWebhook && mediaType != models.NotifyMediaTypeDingtalk {
		return nil, errors.New("媒介类型必须是 webhook 或 dingtalk")
	}
	cfg, err := notify.ParseConfig(rawConfig)
	if err != nil {
		return nil, err
	}
	switch mediaType {
	case models.NotifyMediaTypeWebhook:
		if cfg.URL == "" {
			return nil, errors.New("webhook 回调地址不能为空")
		}
	case models.NotifyMediaTypeDingtalk:
		if cfg.Webhook == "" {
			return nil, errors.New("钉钉机器人 Webhook 地址不能为空")
		}
	}
	return cfg, nil
}

// List 获取所有通知媒介列表
func (c *NotifyMediaController) List(ctx *gin.Context) {
	var list []models.NotifyMedia
	if err := c.DB.Order("created_at desc").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": list,
	})
}

// Get 根据 ID 获取单个通知媒介
func (c *NotifyMediaController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var media models.NotifyMedia
	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "通知媒介不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": media,
	})
}

// Create 创建通知媒介
func (c *NotifyMediaController) Create(ctx *gin.Context) {
	var media models.NotifyMedia
	if err := ctx.ShouldBindJSON(&media); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if media.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "媒介名称不能为空",
		})
		return
	}
	if _, err := validateMediaConfig(media.Type, media.Config); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	if err := c.DB.Create(&media).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": media,
	})
}

// Update 更新通知媒介
func (c *NotifyMediaController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var media models.NotifyMedia
	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "通知媒介不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	var updateData models.NotifyMedia
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if updateData.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "媒介名称不能为空",
		})
		return
	}
	// 编辑时类型不可变更，沿用原有类型
	if _, err := validateMediaConfig(media.Type, updateData.Config); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"name":       updateData.Name,
		"type":       media.Type,
		"config":     updateData.Config,
		"remark":     updateData.Remark,
		"is_enabled": updateData.IsEnabled,
	}

	if err := c.DB.Model(&media).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": media,
	})
}

// Delete 删除通知媒介（软删除）
func (c *NotifyMediaController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var media models.NotifyMedia
	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "通知媒介不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	if err := c.DB.Delete(&media).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
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

// ToggleEnabled 切换启用状态
func (c *NotifyMediaController) ToggleEnabled(ctx *gin.Context) {
	id := ctx.Param("id")

	var media models.NotifyMedia
	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "通知媒介不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	newStatus := 0
	if media.IsEnabled == 0 {
		newStatus = 1
	}

	if err := c.DB.Model(&media).Update("is_enabled", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "操作失败",
			"error":   err.Error(),
		})
		return
	}

	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": media,
	})
}

// TypeList 获取通知媒介类型列表
func (c *NotifyMediaController) TypeList(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": models.NotifyMediaTypeOptions,
	})
}

// Test 发送测试消息，支持通过请求体自定义测试内容
func (c *NotifyMediaController) Test(ctx *gin.Context) {
	id := ctx.Param("id")
	var media models.NotifyMedia
	if err := c.DB.First(&media, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知媒介不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	if _, err := validateMediaConfig(media.Type, media.Config); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "message": "测试失败: " + err.Error()})
		return
	}

	// 请求体可选，content 非空时使用用户自定义内容，否则使用默认文案
	var body struct {
		Content string `json:"content"`
	}
	_ = ctx.ShouldBindJSON(&body)

	content := strings.TrimSpace(body.Content)
	if content == "" {
		content = fmt.Sprintf("【AIOPS】通知媒介「测试」消息，时间：%s", time.Now().Format("2006-01-02 15:04:05"))
	}

	if err := notify.Send(media.Type, media.Config, content); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "message": "测试失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "测试消息发送成功"})
}

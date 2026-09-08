package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"aiops/internal/notify"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotifyTemplateController 消息模板控制器
type NotifyTemplateController struct {
	DB     *gorm.DB
	Domain string // 站点地址（如 http://localhost:5173），用于模板里的 {{$.domain}} 变量
}

func NewNotifyTemplateController(db *gorm.DB, domain string) *NotifyTemplateController {
	return &NotifyTemplateController{DB: db, Domain: domain}
}

func (c *NotifyTemplateController) List(ctx *gin.Context) {
	var list []models.NotifyTemplate
	if err := c.DB.Order("created_at desc").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (c *NotifyTemplateController) Create(ctx *gin.Context) {
	var t models.NotifyTemplate
	if err := ctx.ShouldBindJSON(&t); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if t.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "模板名称不能为空"})
		return
	}
	if t.Content == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "模板内容不能为空"})
		return
	}
	if err := notify.ValidateTemplate(t.Content); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := c.DB.Create(&t).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": t})
}

func (c *NotifyTemplateController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var t models.NotifyTemplate
	if err := c.DB.First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "模板不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	var body struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		Type          string `json:"type"`
		MediaType     string `json:"media_type"`
		Content       string `json:"content"`
		VariablesHint string `json:"variables_hint"`
		IsEnabled     int    `json:"is_enabled"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if body.Content != "" {
		if err := notify.ValidateTemplate(body.Content); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
	}
	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Description != "" {
		updates["description"] = body.Description
	}
	if body.Type != "" {
		updates["type"] = body.Type
	}
	if body.MediaType != "" {
		updates["media_type"] = body.MediaType
	}
	if body.Content != "" {
		updates["content"] = body.Content
	}
	updates["variables_hint"] = body.VariablesHint
	updates["is_enabled"] = body.IsEnabled
	if err := c.DB.Model(&t).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.DB.First(&t, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": t})
}

func (c *NotifyTemplateController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.DB.Delete(&models.NotifyTemplate{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0})
}

// Validate 校验模板语法
func (c *NotifyTemplateController) Validate(ctx *gin.Context) {
	var body struct {
		Content string `json:"content"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := notify.ValidateTemplate(body.Content); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": gin.H{"ok": false, "message": err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"ok": true}})
}

// Preview 预览模板渲染效果（给前端编辑器的实时预览用）
// 请求体可选 evt 字段（夜莺事件原始对象）和 domain 字段，缺省则用 Mock 数据 / config 默认站点地址
func (c *NotifyTemplateController) Preview(ctx *gin.Context) {
	var body struct {
		Content string                 `json:"content"`
		Event   map[string]interface{} `json:"event,omitempty"`
		Domain  string                 `json:"domain,omitempty"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if body.Content == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "模板内容不能为空"})
		return
	}

	var evt *notify.AlertEvent
	if len(body.Event) > 0 {
		evt = notify.ParseEventFromRaw(body.Event)
	} else {
		evt = sampleAlertEvent()
	}

	domain := body.Domain
	if domain == "" {
		domain = c.Domain
	}

	result := notify.RenderTemplateJSON(body.Content, evt, domain)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// SampleEvent 返回一个示例告警事件（前端预览模板用）
func (c *NotifyTemplateController) SampleEvent(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": sampleAlertEvent()})
}

// ---------- 默认模板种子 ----------

// EnsureDefaultTemplate 如果不存在则创建默认模板
func EnsureDefaultTemplate(db *gorm.DB) error {
	var count int64
	db.Model(&models.NotifyTemplate{}).Count(&count)
	if count > 0 {
		return nil
	}
	tpl := models.NotifyTemplate{
		Name:        "默认钉钉告警模板",
		Description: "钉钉机器人告警通知模板（markdown 格式），包含触发/恢复场景",
		Type:        "all",
		MediaType:   "dingtalk",
		Content:     notify.DefaultTemplateFiring,
		IsEnabled:   1,
	}
	if err := db.Create(&tpl).Error; err != nil {
		return fmt.Errorf("创建默认模板失败: %w", err)
	}
	return nil
}

// sampleAlertEvent 构造一个示例 AlertEvent（给预览和 sample_event 接口用）
func sampleAlertEvent() *notify.AlertEvent {
	tags := map[string]string{
		"__name__":  "mysql_slave_status_slave_sql_running",
		"alertName": "connection",
		"instance":  "10.2.211.21:9104",
		"ip":        "10.2.211.21",
		"service":   "mysql",
		"system":    "data_center",
		"type":      "mysql",
	}
	now := time.Now()
	return &notify.AlertEvent{
		Id:             "102738452",
		RuleID:         76,
		RuleName:       "MySQL数据库-复制SQL线程状态停滞",
		RuleNote:       "监控 MySQL 从库复制状态，SQL 线程停滞超过 1 分钟触发告警",
		Cluster:        "VictoriaMetrics(聚合多个Prometheus的数据)",
		BusiGroupID:    22,
		BusiGroupName:  "mysql数据库",
		Severity:       2,
		SeverityLabel:  "P2-警告",
		TriggerValue:   "0.00",
		TriggerTime:    now.Add(-3 * time.Minute),
		FirstTrigger:   now.Add(-3 * time.Minute),
		LastTrigger:    now.Add(-1 * time.Second),
		LastEvalTime:   now.Add(-1 * time.Second),
		DurationSec:    180,
		IsRecovered:    false,
		Location:       "10.2.211.21:9104",
		Cate:           "prometheus",
		TargetIdent:    "10.2.211.21:9104",
		Datasource:     "VictoriaMetrics",
		Tags:           tags,
		TagsMap:        tags,
		TagsJSON:       `{"__name__":"mysql_slave_status_slave_sql_running","alertName":"connection","instance":"10.2.211.21:9104","ip":"10.2.211.21","service":"mysql","type":"mysql"}`,
		AnnotationsJSON: map[string]string{
			"dashboard": "https://grafana.example.com/d/mysql-replica",
			"runbook":   "https://wiki.example.com/runbooks/mysql-replica-lag",
		},
	}
}

// suppress unused
var _ = json.Marshal

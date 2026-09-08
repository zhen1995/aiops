package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotifyTemplate 通知消息模板（Go template 语法）
type NotifyTemplate struct {
	ID          string         `gorm:"primaryKey;size:64" json:"id"`
	Name        string         `gorm:"size:64;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	// Type 模板适用场景：firing=告警触发 / recovered=告警恢复 / all=通用
	Type        string         `gorm:"size:16;not null;default:all" json:"type"`
	// MediaType 媒介类型，与 notify_media.type 对应：dingtalk / webhook / email / wecom
	MediaType   string         `gorm:"size:32;not null;default:dingtalk" json:"media_type"`
	// Content Go template 内容（text/template）
	Content     string         `gorm:"type:text;not null" json:"content"`
	// VariablesHint 变量提示 JSON，描述模板中可引用的变量（前端展示用）
	VariablesHint string       `gorm:"type:text" json:"variables_hint"`
	IsEnabled   int            `gorm:"default:1" json:"is_enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`
}

func (NotifyTemplate) TableName() string { return "notify_template" }

func (t *NotifyTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.IsEnabled == 0 {
		t.IsEnabled = 1
	}
	if t.Type == "" {
		t.Type = "all"
	}
	return nil
}

// NotifyRule 通知规则 —— 单媒介 × 单模板 关联 + 触发条件
type NotifyRule struct {
	ID         string         `gorm:"primaryKey;size:64" json:"id"`
	Name       string         `gorm:"size:64;not null" json:"name"`
	Remark     string         `gorm:"size:255" json:"remark"`

	// MediaID 关联的通知媒介（单个外键）
	MediaID    string         `gorm:"size:64" json:"media_id"`
	// TemplateID 关联的消息模板（单个外键）
	TemplateID string         `gorm:"size:64" json:"template_id"`

	// TriggerTypes 触发类型：firing=告警触发 / recovered=告警恢复 / all=通用
	TriggerTypes string      `gorm:"size:16;default:all" json:"trigger_types"`

	// SeverityFilter 级别过滤（JSON 数组字符串）：[1,2,3]，空表示不过滤
	SeverityFilter string    `gorm:"size:128" json:"severity_filter"`

	IsEnabled   int          `gorm:"default:1" json:"is_enabled"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`
}

func (NotifyRule) TableName() string { return "notify_rule" }

func (r *NotifyRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.IsEnabled == 0 {
		r.IsEnabled = 1
	}
	if r.TriggerTypes == "" {
		r.TriggerTypes = "all"
	}
	return nil
}

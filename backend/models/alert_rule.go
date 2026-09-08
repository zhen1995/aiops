package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertRule 告警规则（系统自建，不再依赖 Nightingale）
type AlertRule struct {
	ID           string         `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Name         string         `gorm:"size:128;not null;comment:规则名称" json:"name"`
	PromQL       string         `gorm:"type:text;not null;comment:PromQL 表达式" json:"prom_ql"`
	EvalInterval int            `gorm:"comment:执行频率(秒)" json:"eval_interval"`
	Duration     int            `gorm:"comment:持续时间(秒)" json:"duration"`
	Severity     int            `gorm:"comment:告警级别 1-P1紧急 2-P2警告 3-P3提醒" json:"severity"`
	// NotifyRuleID 关联的通知规则（可选）——触发告警/恢复时自动发送通知
	NotifyRuleID string         `gorm:"size:64;comment:通知规则ID" json:"notify_rule_id"`
	// RepeatIntervalMinutes 重复通知间隔（分钟），0 表示不重复提醒，仅首次触发时发一次
	RepeatIntervalMinutes int   `gorm:"default:0;comment:重复通知间隔(分钟)" json:"repeat_interval_minutes"`
	// MaxSendCount 最大发送次数，0 表示不限制
	MaxSendCount          int   `gorm:"default:0;comment:最大发送次数 0=不限制" json:"max_send_count"`
	IsEnabled    int            `gorm:"comment:启用状态 0-停用 1-启用" json:"is_enabled"`
	CreatedAt    time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (AlertRule) TableName() string {
	return "alert_rules"
}

// BeforeCreate 自动生成 UUID 并设置默认值
func (r *AlertRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.IsEnabled == 0 {
		r.IsEnabled = 1
	}
	return nil
}

// AlertRuleSeverityOptions 告警级别选项
var AlertRuleSeverityOptions = []struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}{
	{Value: 1, Label: "P1-紧急"},
	{Value: 2, Label: "P2-警告"},
	{Value: 3, Label: "P3-提醒"},
}

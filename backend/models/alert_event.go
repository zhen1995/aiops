package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 告警事件类型与状态常量
const (
	AlertEventTypeAlert    = "alert"    // 告警事件
	AlertEventTypeRecovery = "recovery" // 告警恢复事件

	AlertEventStatusFiring   = "firing"   // 告警中（未恢复）
	AlertEventStatusResolved = "resolved" // 已恢复/已解决
)

// AlertEvent 告警事件（由系统按告警规则自动探测生成，不再依赖 Nightingale）
type AlertEvent struct {
	ID           string     `gorm:"primaryKey;size:64;comment:id" json:"id"`
	RuleID       string     `gorm:"size:64;index;comment:关联告警规则ID" json:"rule_id"`
	RuleName     string     `gorm:"size:128;comment:规则名称（冗余，便于展示）" json:"rule_name"`
	Severity     int        `gorm:"comment:告警级别 1-P1紧急 2-P2警告 3-P3提醒" json:"severity"`
	Type         string     `gorm:"size:16;index;comment:事件类型 alert-告警事件 recovery-告警恢复事件" json:"type"`
	Status       string     `gorm:"size:16;index;comment:状态 firing-告警中 resolved-已恢复" json:"status"`
	TargetIdent  string     `gorm:"size:256;comment:告警对象（如 instance）" json:"target_ident"`
	Tags         string     `gorm:"type:text;comment:标签序列化 k=v,..." json:"tags"`
	TriggerValue string     `gorm:"size:64;comment:触发值" json:"trigger_value"`
	TriggerTime  time.Time  `gorm:"index;comment:触发时间" json:"trigger_time"`
	RecoveredAt  *time.Time `gorm:"comment:恢复时间" json:"recovered_at"`
	NotifyCount  int        `gorm:"default:0;comment:已发送通知次数" json:"notify_count"`
	LastNotifiedAt *time.Time `gorm:"comment:上次发送通知时间" json:"last_notified_at"`
	CreatedAt    time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

func (AlertEvent) TableName() string {
	return "alert_events"
}

// BeforeCreate 自动生成 UUID
func (e *AlertEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

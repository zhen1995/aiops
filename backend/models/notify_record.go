package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 通知记录状态常量
const (
	NotifyRecordStatusSuccess      = "success"      // 已发送
	NotifyRecordStatusFailed       = "failed"       // 发送失败
	NotifyRecordStatusIntercepted  = "intercepted"  // 已拦截（被降噪策略拦截）
	NotifyRecordStatusSkipped      = "skipped"      // 已跳过（通知链路中不满足发送条件）
)

// NotifyRecord 统一通知记录：每一次告警通知的发送轨迹（成功/失败/被拦截/被跳过）
type NotifyRecord struct {
	ID             string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	EventID        string    `gorm:"size:64;index;comment:关联告警事件ID" json:"event_id"`
	RuleID         string    `gorm:"size:64;comment:关联告警规则ID" json:"rule_id"`
	RuleName       string    `gorm:"size:128;comment:规则名称" json:"rule_name"`
	Severity       int       `gorm:"comment:告警级别 1-P1紧急 2-P2警告 3-P3提醒" json:"severity"`
	EventType      string    `gorm:"size:16;comment:事件类型 alert/recovery" json:"event_type"`
	TargetIdent    string    `gorm:"size:256;comment:告警对象" json:"target_ident"`
	Tags           string    `gorm:"type:text;comment:标签序列化 k=v,..." json:"tags"`
	TriggerValue   string    `gorm:"size:64;comment:触发值" json:"trigger_value"`
	TriggerTime    time.Time `gorm:"comment:触发时间" json:"trigger_time"`
	Status         string    `gorm:"size:16;index;comment:状态 success/failed/intercepted/skipped" json:"status"`
	Strategy       string    `gorm:"size:64;comment:拦截策略（intercepted 时）window_aggregation/topology_suppression" json:"strategy"`
	Reason         string    `gorm:"size:512;comment:原因说明" json:"reason"`
	NotifyRuleName string    `gorm:"size:128;comment:通知规则名称" json:"notify_rule_name"`
	MediaName      string    `gorm:"size:128;comment:通知媒介名称" json:"media_name"`
	MediaType      string    `gorm:"size:32;comment:通知媒介类型" json:"media_type"`
	TemplateName   string    `gorm:"size:128;comment:消息模板名称" json:"template_name"`
	Content        string    `gorm:"type:text;comment:渲染后的消息正文" json:"content"`
	CreatedAt      time.Time `gorm:"index;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (NotifyRecord) TableName() string {
	return "notify_records"
}

// BeforeCreate 自动生成 UUID
func (r *NotifyRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

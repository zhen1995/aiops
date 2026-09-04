package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 根因分析状态常量
const (
	RootCauseStatusRunning   = "running"   // 分析中
	RootCauseStatusCompleted = "completed" // 分析完成
	RootCauseStatusFailed    = "failed"    // 分析失败
)

// RootCauseAnalysis 告警事件根因分析任务与结果
type RootCauseAnalysis struct {
	ID           string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	AlertEventID string    `gorm:"size:64;index;comment:关联告警事件ID" json:"alert_event_id"`
	RuleName     string    `gorm:"size:128;comment:规则名称（冗余，便于展示）" json:"rule_name"`
	TargetIdent  string    `gorm:"size:256;comment:告警对象（如 instance）" json:"target_ident"`
	Tags         string    `gorm:"type:text;comment:标签序列化 k=v,..." json:"tags"`
	Severity     int       `gorm:"comment:告警级别 1-P1紧急 2-P2警告 3-P3提醒" json:"severity"`
	TriggerTime  time.Time `gorm:"comment:告警触发时间" json:"trigger_time"`
	Status       string    `gorm:"size:16;index;comment:状态 running/completed/failed" json:"status"`
	Error        string    `gorm:"type:text;comment:失败原因" json:"error"`
	Result       string    `gorm:"type:text;comment:结构化分析结果 JSON" json:"result"`
	Content      string    `gorm:"type:text;comment:完整分析报告 Markdown" json:"content"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (RootCauseAnalysis) TableName() string {
	return "root_cause_analyses"
}

// BeforeCreate 自动生成 UUID
func (e *RootCauseAnalysis) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

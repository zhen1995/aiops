package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 降噪策略常量
const (
	DenoiseStrategyWindowAggregation   = "window_aggregation"   // 时间窗口聚合
	DenoiseStrategyTopologySuppression = "topology_suppression" // 拓扑抑制
)

// DenoisePolicy 告警降噪策略配置（每种策略一行，strategy 唯一）
type DenoisePolicy struct {
	ID          string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Strategy    string    `gorm:"size:64;uniqueIndex;comment:策略类型 window_aggregation/topology_suppression" json:"strategy"`
	Name        string    `gorm:"size:128;comment:策略名称" json:"name"`
	Description string    `gorm:"size:256;comment:策略描述" json:"description"`
	Enabled     bool      `gorm:"comment:是否启用" json:"enabled"`
	Config      string    `gorm:"type:text;comment:策略配置 JSON 文本（窗口策略存 window_minutes）" json:"config"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (DenoisePolicy) TableName() string {
	return "denoise_policies"
}

// BeforeCreate 自动生成 UUID
func (p *DenoisePolicy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// DenoiseRecord 降噪拦截记录（被降噪策略拦截、未生成告警事件的原始告警）
type DenoiseRecord struct {
	ID          string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Strategy    string    `gorm:"size:64;index;comment:拦截策略 window_aggregation/topology_suppression" json:"strategy"`
	RuleID      string    `gorm:"size:64;comment:关联告警规则ID" json:"rule_id"`
	RuleName    string    `gorm:"size:128;comment:规则名称" json:"rule_name"`
	Severity    int       `gorm:"comment:告警级别 1-P1紧急 2-P2警告 3-P3提醒" json:"severity"`
	TargetIdent string    `gorm:"size:256;comment:告警对象" json:"target_ident"`
	Tags        string    `gorm:"type:text;comment:标签序列化 k=v,..." json:"tags"`
	CreatedAt   time.Time `gorm:"index;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (DenoiseRecord) TableName() string {
	return "denoise_records"
}

// BeforeCreate 自动生成 UUID
func (r *DenoiseRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

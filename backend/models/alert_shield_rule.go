package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertShieldRule 告警屏蔽规则：屏蔽期间内，事件标签命中规则的事件不产生告警事件与通知
type AlertShieldRule struct {
	ID        string         `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Name      string         `gorm:"size:128;not null;comment:规则名称" json:"name"`
	Tags      string         `gorm:"type:text;comment:事件标签 JSON 数组 [\"k=v\",...]，全部命中即屏蔽" json:"tags"`
	StartAt   time.Time      `gorm:"comment:屏蔽开始时间" json:"start_at"`
	EndAt     time.Time      `gorm:"comment:屏蔽结束时间" json:"end_at"`
	IsEnabled int            `gorm:"default:1;comment:是否启用 1/0" json:"is_enabled"`
	Remark    string         `gorm:"size:500;comment:屏蔽原因" json:"remark"`
	CreatedBy string         `gorm:"size:64;comment:创建人" json:"created_by"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (AlertShieldRule) TableName() string {
	return "alert_shield_rule"
}

// BeforeCreate 自动生成 UUID
func (r *AlertShieldRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

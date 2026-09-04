package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 通知媒介类型
const (
	NotifyMediaTypeWebhook  = "webhook"
	NotifyMediaTypeDingtalk = "dingtalk"
)

// NotifyMedia 通知媒介
type NotifyMedia struct {
	ID        string         `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Name      string         `gorm:"size:60;not null;comment:媒介名称" json:"name"`
	Type      string         `gorm:"size:20;not null;comment:类型 webhook/dingtalk" json:"type"`
	Config    string         `gorm:"type:text;comment:媒介配置JSON" json:"config"`
	Remark    string         `gorm:"size:200;comment:备注" json:"remark"`
	IsEnabled int            `gorm:"comment:状态 0-停用 1-启用" json:"is_enabled"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (NotifyMedia) TableName() string {
	return "notify_media"
}

// BeforeCreate 自动生成 UUID 并设置默认值
func (m *NotifyMedia) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.IsEnabled == 0 {
		m.IsEnabled = 1
	}
	return nil
}

// NotifyMediaTypeOptions 通知媒介类型选项
var NotifyMediaTypeOptions = []string{
	NotifyMediaTypeWebhook,
	NotifyMediaTypeDingtalk,
}

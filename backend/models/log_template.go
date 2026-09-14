package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LogTemplate 日志模板表：Drain 挖掘出的常量模板 + 变量参数模式
type LogTemplate struct {
	ID           string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	ServiceID    string    `gorm:"size:64;index;comment:关联服务ID" json:"service_id"`
	Template     string    `gorm:"type:text;comment:Drain模板 如 Connection to database <*> failed after <*>ms" json:"template"`
	Level        string    `gorm:"size:16;comment:级别 error/warn/info" json:"level"`
	Count        int64     `gorm:"comment:累计条数" json:"count"`
	SampleRaw    string    `gorm:"type:text;comment:最近一条原始日志" json:"sample_raw"`
	SampleParams string    `gorm:"type:text;comment:最近样本JSON数组 [{raw,params}] 最多保留3条" json:"sample_params"`
	FirstSeenAt  time.Time `gorm:"comment:首次出现" json:"first_seen_at"`
	LastSeenAt   time.Time `gorm:"comment:最近出现" json:"last_seen_at"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (LogTemplate) TableName() string {
	return "log_templates"
}

// BeforeCreate 自动生成 UUID
func (t *LogTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertEngineConfig struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name      string         `gorm:"size:30;not null;comment:名称" json:"name"`
	BaseURL   string         `gorm:"size:100;not null;comment:夜莺地址" json:"base_url"`
	Token     string         `gorm:"size:100;not null;comment:用户Token" json:"token"`
	Gids      string         `gorm:"size:100;comment:业务组ID,逗号分隔,空则使用Token可访问的全部组" json:"gids"`
	IsEnabled int            `gorm:"default:1;comment:0-禁用 1-启用" json:"is_enabled"`
	IsDefault int            `gorm:"default:0;comment:0-否 1-是" json:"is_default"`
	Remark    string         `gorm:"size:100;comment:备注" json:"remark"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (AlertEngineConfig) TableName() string {
	return "alert_engine_config"
}

func (c *AlertEngineConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.IsEnabled == 0 {
		c.IsEnabled = 1
	}
	return nil
}

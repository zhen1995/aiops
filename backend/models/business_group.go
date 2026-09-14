package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BusinessGroup 业务分组：用于对告警规则进行分组管理
type BusinessGroup struct {
	ID        string         `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Name      string         `gorm:"size:128;not null;uniqueIndex;comment:分组名称" json:"name"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (BusinessGroup) TableName() string {
	return "business_groups"
}

// BeforeCreate 自动生成 UUID
func (g *BusinessGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}

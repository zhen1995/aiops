package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SysRole 角色表
type SysRole struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name      string         `gorm:"size:30;comment:名称" json:"name"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	Updated   *time.Time     `gorm:"column:updated;comment:更新时间" json:"updated"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

// BeforeCreate 创建前自动生成 ID 并设置时间戳
func (r *SysRole) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	now := time.Now()
	r.CreatedAt = &now
	r.Updated = &now
	return nil
}

// BeforeUpdate 更新前自动刷新 Updated
func (r *SysRole) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	r.Updated = &now
	return nil
}

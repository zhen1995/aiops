package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SysUser 用户表
type SysUser struct {
	ID        string         `gorm:"primaryKey;size:32;comment:id" json:"id"`
	Username  string         `gorm:"size:30;not null;comment:用户名" json:"username"`
	Password  string         `gorm:"size:30;not null;comment:密码" json:"password"`
	Name      string         `gorm:"size:10;not null;comment:姓名" json:"name"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
	UpdateAt  *time.Time     `gorm:"column:update_at;comment:更新时间" json:"update_at"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

// BeforeCreate 创建前自动生成 ID 并设置时间戳
func (u *SysUser) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	now := time.Now()
	u.CreatedAt = &now
	u.UpdateAt = &now
	return nil
}

// BeforeUpdate 更新前自动刷新 UpdateAt
func (u *SysUser) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	u.UpdateAt = &now
	return nil
}

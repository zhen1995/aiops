package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SysUserRoleRelation 用户-角色关联表
type SysUserRoleRelation struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	UserID    string         `gorm:"size:40;not null;comment:用户id" json:"user_id"`
	RoleID    string         `gorm:"size:40;not null;comment:角色id" json:"role_id"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (SysUserRoleRelation) TableName() string {
	return "sys_user_role_relation"
}

// BeforeCreate 创建前自动生成 ID 并设置时间戳
func (ur *SysUserRoleRelation) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == "" {
		ur.ID = uuid.New().String()
	}
	now := time.Now()
	ur.CreatedAt = &now
	ur.UpdatedAt = &now
	return nil
}

// BeforeUpdate 更新前自动刷新 UpdatedAt
func (ur *SysUserRoleRelation) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	ur.UpdatedAt = &now
	return nil
}

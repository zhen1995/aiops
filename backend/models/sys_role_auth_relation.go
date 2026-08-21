package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SysRoleAuthRelation 角色-权限关联表
type SysRoleAuthRelation struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	RoleID    string         `gorm:"column:role_id;size:40;not null;comment:角色id" json:"roleId"`
	AuthID    string         `gorm:"column:auth_id;size:40;not null;comment:权限id" json:"auth_id"`
	RoleName  string         `gorm:"column:role_name;size:40;comment:角色名称" json:"role_name"`
	AuthName  string         `gorm:"column:auth_name;size:40;comment:权限名称" json:"auth_name"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;comment:更新时间" json:"updatedAt"`
	Updated   *time.Time     `gorm:"column:updated;comment:更新时间" json:"updated"`
	RoleIdAlt string         `gorm:"column:roleId;size:40;not null;comment:角色id(冗余列)" json:"-"`
}

func (SysRoleAuthRelation) TableName() string {
	return "sys_role_auth_relation"
}

// BeforeCreate 创建前自动生成 ID 并设置时间戳
func (r *SysRoleAuthRelation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	// 同时写入 role_id 和 roleId 两个列
	r.RoleIdAlt = r.RoleID
	now := time.Now()
	r.CreatedAt = &now
	r.UpdatedAt = &now
	r.Updated = &now
	return nil
}

// BeforeUpdate 更新前自动刷新时间戳
func (r *SysRoleAuthRelation) BeforeUpdate(tx *gorm.DB) error {
	// 保持 roleId 列与 role_id 列同步
	r.RoleIdAlt = r.RoleID
	now := time.Now()
	r.UpdatedAt = &now
	r.Updated = &now
	return nil
}

package models

import "gorm.io/gorm"

type SysUser struct {
	gorm.Model
	Username string
	Password string
	Name     string
}

func (SysUser) TableName() string {
	return "sys_user"
}

package models

import "gorm.io/gorm"

// SeedSysAuth 确保所有必需的权限项存在
func SeedSysAuth(db *gorm.DB) error {
	names := []string{
		"告警引擎", "告警规则", "告警事件",
		"巡检任务", "巡检报告",
		"通知媒介", "消息模板", "通知规则",
		"夜莺引擎配置", "夜莺告警规则", "夜莺告警事件",
		"服务注册",
	}
	for _, name := range names {
		var count int64
		if err := db.Model(&SysAuth{}).Where("name = ?", name).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			auth := SysAuth{Name: name, Type: 1}
			if err := db.Create(&auth).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// SeedAdminRoleAuth 确保 admin/管理员 角色拥有所有权限（幂等，可重复运行）
// 给 admin 角色分配所有 sys_auth 记录；如果 admin 角色不存在则不做任何事
func SeedAdminRoleAuth(db *gorm.DB) error {
	// 找出 admin 角色（可能叫 "管理员" 或 "admin"）
	var adminRole SysRole
	if err := db.Where("name = ? OR name = ?", "管理员", "admin").First(&adminRole).Error; err != nil {
		// 没找到 admin 角色就跳过，不报错
		return nil
	}

	// 找出所有权限
	var allAuths []SysAuth
	if err := db.Find(&allAuths).Error; err != nil {
		return err
	}

	for _, auth := range allAuths {
		// 检查关系是否已存在
		var count int64
		db.Model(&SysRoleAuthRelation{}).
			Where("role_id = ? AND auth_id = ?", adminRole.ID, auth.ID).
			Count(&count)
		if count > 0 {
			continue
		}
		// 不存在则插入
		rel := SysRoleAuthRelation{
			RoleID:   adminRole.ID,
			AuthID:   auth.ID,
			RoleName: adminRole.Name,
			AuthName: auth.Name,
		}
		if err := db.Create(&rel).Error; err != nil {
			return err
		}
	}
	return nil
}

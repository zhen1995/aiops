package models

import "gorm.io/gorm"

func SeedSysAuth(db *gorm.DB) error {
	names := []string{"告警引擎", "告警规则"}
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

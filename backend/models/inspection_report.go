package models

import (
	"time"

	"gorm.io/gorm"
)

// InspectionReport 巡检报告表
type InspectionReport struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	TaskID    uint   `gorm:"index" json:"task_id"`            // 关联任务 ID
	TaskName  string `gorm:"size:128" json:"task_name"`       // 冗余任务名，方便列表展示
	Title     string `gorm:"size:256;not null" json:"title"`  // 报告标题
	Content   string `gorm:"type:longtext;not null" json:"content"` // Markdown 报告正文
	Summary   string `gorm:"size:512" json:"summary"`          // 简短摘要（列表展示用）
	Score     int    `gorm:"default:0" json:"score"`          // 综合评分 0-100
	Status    string `gorm:"size:32;default:'completed'" json:"status"` // generating / completed / failed
	Error     string `gorm:"size:512" json:"error"`           // 失败原因
	CreatedAt time.Time `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InspectionReport) TableName() string {
	return "inspection_reports"
}

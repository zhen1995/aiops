package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KBDocument 知识库文档表（向量存 Qdrant，本表只管元数据与状态）
type KBDocument struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name      string         `gorm:"size:200;not null;comment:文件名" json:"name"`
	Type      string         `gorm:"size:10;not null;comment:扩展名" json:"type"`
	Size      int64          `gorm:"comment:文件大小(字节)" json:"size"`
	Status    string         `gorm:"size:10;not null;default:pending;comment:状态 pending/indexing/indexed/failed" json:"status"`
	ErrorMsg  string         `gorm:"size:500;comment:失败原因" json:"error_msg"`
	Uploader  string         `gorm:"size:30;comment:上传人" json:"uploader"`
	FilePath  string         `gorm:"size:300;comment:文件存储路径" json:"file_path"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (KBDocument) TableName() string {
	return "kb_document"
}

// BeforeCreate 创建前自动生成 ID 并设置时间戳
func (d *KBDocument) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	now := time.Now()
	d.CreatedAt = &now
	d.UpdatedAt = &now
	return nil
}

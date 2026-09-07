package knowledge

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiops/internal/chat"
	"aiops/models"

	"gorm.io/gorm"
)

// Service 知识库编排：文件落盘、异步索引、删除清理、检索
type Service struct {
	DB        *gorm.DB
	Client    *Client
	UploadDir string
}

func NewService(db *gorm.DB, client *Client, uploadDir string) *Service {
	return &Service{DB: db, Client: client, UploadDir: uploadDir}
}

// SaveUpload 保存上传文件并落库（pending），随后调用方启动 IndexDocument
func (s *Service) SaveUpload(ctx context.Context, fileName string, size int64, data io.Reader, uploader string) (*models.KBDocument, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	if ext == "" {
		return nil, errors.New("无法识别的文件类型")
	}
	doc := &models.KBDocument{Name: fileName, Type: ext, Size: size, Status: "pending", Uploader: uploader}
	if err := s.DB.Create(doc).Error; err != nil {
		return nil, err
	}
	doc.FilePath = filepath.Join(s.UploadDir, doc.ID+"."+ext)
	if err := s.saveFile(doc.FilePath, data); err != nil {
		s.DB.Delete(doc)
		return nil, err
	}
	if err := s.DB.Model(doc).Update("file_path", doc.FilePath).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

// IndexDocument 异步执行：解析→向量化→写 chunk→更新状态
func (s *Service) IndexDocument(docID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		var doc models.KBDocument
		if err := s.DB.First(&doc, "id = ?", docID).Error; err != nil {
			return
		}
		s.DB.Model(&doc).Updates(map[string]interface{}{"status": "indexing", "error_msg": ""})

		fail := func(msg string) {
			s.DB.Model(&doc).Updates(map[string]interface{}{"status": "failed", "error_msg": msg})
		}

		f, err := os.Open(doc.FilePath)
		if err != nil {
			fail("读取文件失败: " + err.Error())
			return
		}
		defer f.Close()

		emb, err := chat.DefaultConfig(s.DB, models.LLMModelTypeEmbedding)
		if err != nil {
			fail("加载向量化模型配置失败: " + err.Error())
			return
		}

		if _, err := s.Client.Index(ctx, doc.ID, doc.Name, emb, doc.Name, f); err != nil {
			fail(err.Error())
			return
		}

		// 记录分块元数据（内容在 Qdrant，这里只清理旧记录，避免双写大文本）
		s.DB.Where("document_id = ?", doc.ID).Delete(&models.KBChunk{})
		s.DB.Model(&doc).Updates(map[string]interface{}{"status": "indexed"})
	}()
}

// Reindex 删除旧向量后重新索引
func (s *Service) Reindex(ctx context.Context, docID string) error {
	if err := s.Client.DeleteDocument(ctx, docID); err != nil {
		return err
	}
	s.IndexDocument(docID)
	return nil
}

// Delete 删除文档：向量、分块记录、文件、元数据
func (s *Service) Delete(ctx context.Context, docID string) error {
	var doc models.KBDocument
	if err := s.DB.First(&doc, "id = ?", docID).Error; err != nil {
		return err
	}
	if err := s.Client.DeleteDocument(ctx, docID); err != nil {
		return err
	}
	if err := s.DB.Where("document_id = ?", docID).Delete(&models.KBChunk{}).Error; err != nil {
		return err
	}
	if doc.FilePath != "" {
		_ = os.Remove(doc.FilePath)
	}
	return s.DB.Delete(&doc).Error
}

// Retrieve 语义检索，直接透传 Python 服务结果
func (s *Service) Retrieve(ctx context.Context, query string, topK int) ([]ChunkHit, error) {
	emb, err := chat.DefaultConfig(s.DB, models.LLMModelTypeEmbedding)
	if err != nil {
		return nil, err
	}
	hits, err := s.Client.Retrieve(ctx, query, topK, emb)
	if err != nil {
		return nil, err
	}
	// 空命中时保证返回非 nil 切片，避免 JSON 序列化为 null
	if hits == nil {
		hits = make([]ChunkHit, 0)
	}
	return hits, nil
}

func (s *Service) saveFile(path string, data io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, data)
	return err
}

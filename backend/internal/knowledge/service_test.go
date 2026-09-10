package knowledge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiops/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 用内存 SQLite 代替 MySQL（结构一致，仅用于单测）
// 每个测试使用独立的内存库名，避免 cache=shared 串数据
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.KBDocument{}, &models.KBChunk{}, &models.LLMConfig{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSaveUpload(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	svc := NewService(db, NewClient("http://unused"), dir)

	doc, err := svc.SaveUpload(context.Background(), "手册.md", 12, strings.NewReader("# 内容内容内容"), "root")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Status != "pending" || doc.Type != "md" {
		t.Errorf("doc = %+v", doc)
	}
	if _, err := os.Stat(filepath.Join(dir, doc.ID+".md")); err != nil {
		t.Errorf("文件未落盘: %v", err)
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()

	var deletedVec bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletedVec = true
		}
		w.Write([]byte(`{"document_id":"x","chunks":1}`))
	}))
	defer srv.Close()

	svc := NewService(db, NewClient(srv.URL), dir)
	doc, _ := svc.SaveUpload(context.Background(), "a.md", 2, strings.NewReader("hi"), "root")
	svc.IndexDocument(doc.ID) // 假服务返回 chunks，走完流程

	if err := svc.Delete(context.Background(), doc.ID); err != nil {
		t.Fatal(err)
	}
	if !deletedVec {
		t.Error("未调用向量删除")
	}
	var count int64
	db.Model(&models.KBDocument{}).Count(&count)
	if count != 0 {
		t.Errorf("文档未删除, count=%d", count)
	}
	if _, err := os.Stat(filepath.Join(dir, doc.ID+".md")); !os.IsNotExist(err) {
		t.Error("文件未删除")
	}
}

// seedEmbeddingConfig 写入一条默认启用的向量化模型配置，让 IndexDocument 能走到 Client.Index
func seedEmbeddingConfig(t *testing.T, db *gorm.DB) {
	t.Helper()
	cfg := models.LLMConfig{
		Name:      "测试向量化模型",
		ModelType: models.LLMModelTypeEmbedding,
		Model:     "test-embedding",
		BaseURL:   "http://unused",
		APIKey:    "test-key",
		IsDefault: 1,
		IsEnabled: 1,
	}
	if err := db.Create(&cfg).Error; err != nil {
		t.Fatal(err)
	}
}

// TestDeleteDuringIndexingRollsBackVectors 复现原线上故障：
// 索引进行中删除文档，索引完成后不得残留向量（goroutine 需回滚）
func TestDeleteDuringIndexingRollsBackVectors(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	seedEmbeddingConfig(t, db)

	indexStarted := make(chan struct{})
	var deleteCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost: // /index：模拟向量化耗时
			close(indexStarted)
			time.Sleep(300 * time.Millisecond)
			w.Write([]byte(`{"document_id":"x","chunks":3}`))
		case http.MethodDelete:
			deleteCalls++
			w.Write([]byte(`{"deleted":true}`))
		}
	}))
	defer srv.Close()

	svc := NewService(db, NewClient(srv.URL), dir)
	doc, _ := svc.SaveUpload(context.Background(), "a.md", 2, strings.NewReader("hi"), "root")
	svc.IndexDocument(doc.ID)

	// 等索引请求在途后立即删除文档
	<-indexStarted
	time.Sleep(50 * time.Millisecond)
	if err := svc.Delete(context.Background(), doc.ID); err != nil {
		t.Fatal(err)
	}

	// 等索引协程收尾（含回滚）
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if deleteCalls >= 2 { // 1 次来自用户删除，1 次来自索引回滚
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if deleteCalls < 2 {
		t.Errorf("索引完成后未回滚向量，deleteCalls=%d", deleteCalls)
	}
}

// TestDeleteReturnsErrorWhenFileLocked 文件删除失败时必须报错且保留文档记录
func TestDeleteReturnsErrorWhenFileLocked(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	svc := NewService(db, NewClient(srv.URL), dir)
	doc, _ := svc.SaveUpload(context.Background(), "a.md", 2, strings.NewReader("hi"), "root")

	// 用同名非空目录顶替文件，强制 os.Remove 失败（跨平台，不依赖 Windows 锁语义）
	filePath := filepath.Join(dir, doc.ID+".md")
	if err := os.Remove(filePath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filePath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filePath, "inner.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := svc.Delete(context.Background(), doc.ID)
	if err == nil {
		t.Fatal("文件删除失败时应返回错误")
	}
	var count int64
	db.Model(&models.KBDocument{}).Count(&count)
	if count != 1 {
		t.Errorf("文件删除失败时文档记录应保留, count=%d", count)
	}
	if _, statErr := os.Stat(filePath); statErr != nil {
		t.Errorf("文件不应被删除: %v", statErr)
	}
}

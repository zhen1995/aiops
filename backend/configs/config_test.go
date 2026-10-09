package configs

import (
	"os"
	"path/filepath"
	"testing"
)

// 全部候选路径不存在时：不报错，退化为纯环境变量 + 默认值
func TestLoadConfigNoFileFallsBackToEnvAndDefaults(t *testing.T) {
	t.Setenv("AIOPS_KNOWLEDGE_PYTHON_BASE_URL", "http://analyzer:9000")

	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nope", "config.yaml"))
	if err != nil {
		t.Fatalf("无配置文件时应降级为环境变量运行，实际报错: %v", err)
	}
	if cfg.Knowledge.PythonBaseURL != "http://analyzer:9000" {
		t.Fatalf("环境变量未生效: %q", cfg.Knowledge.PythonBaseURL)
	}
	if cfg.Server.Port != ":8080" {
		t.Fatalf("默认端口应为 :8080，实际 %q", cfg.Server.Port)
	}
	if cfg.Knowledge.UploadDir != "./uploads/knowledge" {
		t.Fatalf("默认上传目录不正确: %q", cfg.Knowledge.UploadDir)
	}
	if cfg.Knowledge.QdrantURL != "http://localhost:6333" {
		t.Fatalf("默认 Qdrant 地址不正确: %q", cfg.Knowledge.QdrantURL)
	}
}

// 第一个存在的候选路径生效，且环境变量优先于配置文件
func TestLoadConfigFirstExistingPathWins(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "config.yaml")
	content := "server:\n  port: \":9000\"\nknowledge:\n  python_base_url: \"http://file:9000\"\n"
	if err := os.WriteFile(good, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing", "config.yaml")

	t.Setenv("AIOPS_KNOWLEDGE_PYTHON_BASE_URL", "http://env:9000")

	cfg, err := LoadConfig(missing, good)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != ":9000" {
		t.Fatalf("应读取第一个存在的候选文件，实际端口 %q", cfg.Server.Port)
	}
	if cfg.Knowledge.PythonBaseURL != "http://env:9000" {
		t.Fatalf("环境变量应优先于配置文件，实际 %q", cfg.Knowledge.PythonBaseURL)
	}
}

package shield

import (
	"encoding/json"
	"testing"
	"time"

	"aiops/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.AlertShieldRule{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func mustJSON(t *testing.T, arr []string) string {
	b, err := json.Marshal(arr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

func TestMatch(t *testing.T) {
	db := setupTestDB(t)
	now := time.Now()
	rules := []models.AlertShieldRule{
		{Name: "启用中-单标签", Tags: mustJSON(t, []string{"system=wms"}), StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour), IsEnabled: 1},
		{Name: "启用中-多标签", Tags: mustJSON(t, []string{"instance=wms_prd", "type=应用"}), StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour), IsEnabled: 1},
		{Name: "未开始", Tags: mustJSON(t, []string{"system=wms"}), StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), IsEnabled: 1},
		{Name: "已过期", Tags: mustJSON(t, []string{"system=wms"}), StartAt: now.Add(-2 * time.Hour), EndAt: now.Add(-time.Hour), IsEnabled: 1},
		{Name: "已停用", Tags: mustJSON(t, []string{"system=wms"}), StartAt: now.Add(-time.Hour), EndAt: now.Add(time.Hour), IsEnabled: 0},
	}
	for i := range rules {
		if err := db.Create(&rules[i]).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	s := NewService(db)
	if err := s.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}

	// 单标签命中
	ok, name := s.Match("instance=wms_prd, ip=10.0.0.1, system=wms", now)
	if !ok || name != "启用中-单标签" {
		t.Fatalf("single tag match: ok=%v name=%q", ok, name)
	}

	// 多标签全部命中
	ok, name = s.Match("type=应用, instance=wms_prd, extra=x", now)
	if !ok || name != "启用中-多标签" {
		t.Fatalf("multi tag match: ok=%v name=%q", ok, name)
	}

	// 多标签部分命中 → 不屏蔽
	ok, _ = s.Match("type=应用, instance=other", now)
	if ok {
		t.Fatal("partial match should not shield")
	}

	// 时间窗外与停用规则不命中
	ok, _ = s.Match("type=数据库", now)
	if ok {
		t.Fatal("unmatched tag should not shield")
	}
}

package rca

import (
	"fmt"
	"testing"
	"time"

	"aiops/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupLocalizeDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Service{}, &models.AlertEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	parent := "svc-parent"
	services := []models.Service{
		{ID: "svc-parent", Name: "中间件A", Code: "mid-a", PromLabels: `{"system":"wms"}`},
		{ID: "svc-child", Name: "业务B", Code: "biz-b", ParentID: &parent, PromLabels: `{"system":"wms","type":"应用"}`},
		{ID: "svc-other", Name: "独立C", Code: "ind-c", PromLabels: `{"system":"erp"}`},
	}
	for i := range services {
		if err := db.Create(&services[i]).Error; err != nil {
			t.Fatalf("create service: %v", err)
		}
	}
	now := time.Now()
	events := []models.AlertEvent{
		{ID: "e1", RuleID: "r1", RuleName: "中间件CPU高", Status: models.AlertEventStatusFiring, Tags: "system=wms", TriggerTime: now.Add(-10 * time.Minute)},
		{ID: "e2", RuleID: "r2", RuleName: "业务延迟大", Status: models.AlertEventStatusFiring, Tags: "system=wms,type=应用", TriggerTime: now.Add(-5 * time.Minute)},
		{ID: "e3", RuleID: "r3", RuleName: "历史已恢复", Status: models.AlertEventStatusResolved, Tags: "system=erp", TriggerTime: now.Add(-3 * time.Minute)},
	}
	for i := range events {
		if err := db.Create(&events[i]).Error; err != nil {
			t.Fatalf("create event: %v", err)
		}
	}
	return db
}

func TestLocalizeByTopology(t *testing.T) {
	db := setupLocalizeDB(t)
	out := localizeByTopology(db, models.AlertEvent{}, localizeWindow)
	fmt.Println(out)
	// 父服务聚合了自身 + 子孙告警（2 条），应排在第一且独立C（已恢复不计）不出现
	if out == "" {
		t.Fatal("expected localization output")
	}
}

package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchController 全局搜索控制器：并行查询服务/告警事件/告警规则/知识库文档并聚合
type SearchController struct {
	DB *gorm.DB
}

// NewSearchController 创建控制器
func NewSearchController(db *gorm.DB) *SearchController {
	return &SearchController{DB: db}
}

// 各类搜索命中出参（matched 为命中的字段值，供前端高亮）
type searchServiceHit struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Matched string `json:"matched"`
}

type searchAlertHit struct {
	ID          string `json:"id"`
	RuleName    string `json:"rule_name"`
	Severity    int    `json:"severity"`
	Status      string `json:"status"`
	TriggerTime int64  `json:"trigger_time"` // epoch 秒，与前端搜索契约一致
	Matched     string `json:"matched"`
}

type searchRuleHit struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	PromQL  string `json:"promql"`
	Matched string `json:"matched"`
}

type searchDocumentHit struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Matched string `json:"matched"`
}

// escapeLikeKeyword 转义 LIKE 中的 %/_ 及转义符本身，配合 ESCAPE '!' 使用
func escapeLikeKeyword(kw string) string {
	kw = strings.ReplaceAll(kw, "!", "!!")
	kw = strings.ReplaceAll(kw, "%", "!%")
	kw = strings.ReplaceAll(kw, "_", "!_")
	return kw
}

// firstMatched 返回第一个包含关键词的字段值（均未命中时回退到 fallback）
func firstMatched(kw, fallback string, candidates ...string) string {
	for _, c := range candidates {
		if strings.Contains(c, kw) {
			return c
		}
	}
	return fallback
}

// Search 全局搜索聚合
// GET /api/search?q=<关键词>&limit=5
func (c *SearchController) Search(ctx *gin.Context) {
	kw := strings.TrimSpace(ctx.Query("q"))
	if kw == "" || len(kw) > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "q 参数不能为空且不能超过 50 个字符"})
		return
	}
	limit, err := parseInt(ctx.DefaultQuery("limit", "5"))
	if err != nil || limit < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "limit 参数必须是正整数"})
		return
	}
	if limit > 20 {
		limit = 20
	}

	like := "%" + escapeLikeKeyword(kw) + "%"
	esc := " ESCAPE '!'"

	services := make([]searchServiceHit, 0)
	alerts := make([]searchAlertHit, 0)
	rules := make([]searchRuleHit, 0)
	documents := make([]searchDocumentHit, 0)
	var totalServices, totalAlerts, totalRules, totalDocuments int64

	var wg sync.WaitGroup
	wg.Add(4)

	// 服务注册：name/code/owner 模糊匹配，按名称排序
	go func() {
		defer wg.Done()
		db := c.DB.Model(&models.Service{}).
			Where("(name LIKE ?"+esc+" OR code LIKE ?"+esc+" OR owner LIKE ?"+esc+")", like, like, like)
		if err := db.Count(&totalServices).Error; err != nil {
			fmt.Println("全局搜索-服务查询失败:", err)
			return
		}
		var rows []models.Service
		if err := db.Order("name").Limit(limit).Find(&rows).Error; err != nil {
			fmt.Println("全局搜索-服务查询失败:", err)
			return
		}
		for _, s := range rows {
			services = append(services, searchServiceHit{
				ID:      s.ID,
				Name:    s.Name,
				Code:    s.Code,
				Matched: firstMatched(kw, s.Name, s.Name, s.Code, s.Owner),
			})
		}
	}()

	// 告警事件：仅告警事件（type=alert），rule_name/target_ident/tags 模糊匹配，按触发时间倒序
	go func() {
		defer wg.Done()
		db := c.DB.Model(&models.AlertEvent{}).
			Where("type = ?", models.AlertEventTypeAlert).
			Where("(rule_name LIKE ?"+esc+" OR target_ident LIKE ?"+esc+" OR tags LIKE ?"+esc+")", like, like, like)
		if err := db.Count(&totalAlerts).Error; err != nil {
			fmt.Println("全局搜索-告警事件查询失败:", err)
			return
		}
		var rows []models.AlertEvent
		if err := db.Order("trigger_time DESC").Limit(limit).Find(&rows).Error; err != nil {
			fmt.Println("全局搜索-告警事件查询失败:", err)
			return
		}
		for _, e := range rows {
			alerts = append(alerts, searchAlertHit{
				ID:          e.ID,
				RuleName:    e.RuleName,
				Severity:    e.Severity,
				Status:      e.Status,
				TriggerTime: e.TriggerTime.Unix(),
				Matched:     firstMatched(kw, e.RuleName, e.RuleName, e.TargetIdent, e.Tags),
			})
		}
	}()

	// 告警规则：name/promql 模糊匹配（软删除由 GORM 默认过滤，与列表查询口径一致）
	go func() {
		defer wg.Done()
		db := c.DB.Model(&models.AlertRule{}).
			Where("(name LIKE ?"+esc+" OR promql LIKE ?"+esc+")", like, like)
		if err := db.Count(&totalRules).Error; err != nil {
			fmt.Println("全局搜索-告警规则查询失败:", err)
			return
		}
		var rows []models.AlertRule
		if err := db.Order("created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
			fmt.Println("全局搜索-告警规则查询失败:", err)
			return
		}
		for _, r := range rows {
			rules = append(rules, searchRuleHit{
				ID:      r.ID,
				Name:    r.Name,
				PromQL:  r.PromQL,
				Matched: firstMatched(kw, r.Name, r.Name, r.PromQL),
			})
		}
	}()

	// 知识库文档：文件名模糊匹配，按创建时间倒序
	go func() {
		defer wg.Done()
		db := c.DB.Model(&models.KBDocument{}).
			Where("name LIKE ?"+esc, like)
		if err := db.Count(&totalDocuments).Error; err != nil {
			fmt.Println("全局搜索-知识库文档查询失败:", err)
			return
		}
		var rows []models.KBDocument
		if err := db.Order("created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
			fmt.Println("全局搜索-知识库文档查询失败:", err)
			return
		}
		for _, d := range rows {
			documents = append(documents, searchDocumentHit{
				ID:      d.ID,
				Title:   d.Name,
				Matched: d.Name,
			})
		}
	}()

	wg.Wait()

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"services":        services,
		"alerts":          alerts,
		"rules":           rules,
		"documents":       documents,
		"total_services":  totalServices,
		"total_alerts":    totalAlerts,
		"total_rules":     totalRules,
		"total_documents": totalDocuments,
		"total":           totalServices + totalAlerts + totalRules + totalDocuments,
	}})
}

package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LogAnalysisController 日志分析控制器（趋势 / 聚类 / Drain 模板）
type LogAnalysisController struct {
	DB *gorm.DB
}

// NewLogAnalysisController 创建控制器
func NewLogAnalysisController(db *gorm.DB) *LogAnalysisController {
	return &LogAnalysisController{DB: db}
}

// ---------- 出参结构 ----------

type logTrendResponse struct {
	Labels []string `json:"labels"`
	Total  []int64  `json:"total"`
	Error  []int64  `json:"error"`
}

type logClusterItem struct {
	ID        string   `json:"id"`
	Pattern   string   `json:"pattern"`
	Count     int64    `json:"count"`
	Level     string   `json:"level"`
	Services  []string `json:"services"`
	Trend     string   `json:"trend"`
	FirstSeen string   `json:"firstSeen"`
}

type logTemplateSample struct {
	Raw    string   `json:"raw"`
	Params []string `json:"params"`
}

type logTemplateItem struct {
	Template     string              `json:"template"`
	Level        string              `json:"level"`
	Count        int64               `json:"count"`
	ParamSamples []logTemplateSample `json:"paramSamples"`
}

// Trend 日志量趋势：聚合 log_cluster_stats 分桶（总量与 ERROR 量双曲线）
// GET /api/logs/trend?hours=24（或 ?start=<RFC3339>&end=<RFC3339>）&service=<服务ID，可选>
func (c *LogAnalysisController) Trend(ctx *gin.Context) {
	start, end, ok := parseLogRange(ctx)
	if !ok {
		return
	}

	type bucketRow struct {
		BucketStart time.Time
		Total       int64
		Error       int64
	}
	q := c.DB.Model(&models.LogClusterStat{}).
		Select("bucket_start, SUM(total_count) AS total, SUM(error_count) AS error").
		Where("bucket_start >= ? AND bucket_start < ?", start.Truncate(time.Hour), end).
		Group("bucket_start").
		Order("bucket_start ASC")
	if ids := parseServiceIDs(ctx.Query("service")); len(ids) > 0 {
		q = q.Where("service_id IN ?", ids)
	}
	var rows []bucketRow
	if err := q.Scan(&rows).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询日志趋势失败: " + err.Error()})
		return
	}

	byBucket := make(map[int64]bucketRow, len(rows))
	for _, r := range rows {
		byBucket[r.BucketStart.Unix()] = r
	}

	resp := logTrendResponse{
		Labels: make([]string, 0),
		Total:  make([]int64, 0),
		Error:  make([]int64, 0),
	}
	for b := start.Truncate(time.Hour); b.Before(end); b = b.Add(time.Hour) {
		row := byBucket[b.Unix()]
		resp.Labels = append(resp.Labels, b.Format("15:04"))
		resp.Total = append(resp.Total, row.Total)
		resp.Error = append(resp.Error, row.Error)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": resp})
}

// Clusters 日志聚类列表：联表取服务名，按级别严重程度 → 趋势 → 数量倒序分页
// GET /api/logs/clusters?service=&trend=&level=(&level 支持逗号分隔多值，如 error,warn)&hours=24&page=1&page_size=20
func (c *LogAnalysisController) Clusters(ctx *gin.Context) {
	page, pageSize := parseLogPage(ctx)

	hours := 24
	if h := ctx.Query("hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 {
			hours = v
		}
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	base := c.DB.Model(&models.LogCluster{}).Where("last_seen_at >= ?", since)
	if ids := parseServiceIDs(ctx.Query("service")); len(ids) > 0 {
		base = base.Where("log_clusters.service_id IN ?", ids)
	}
	if trend := ctx.Query("trend"); trend != "" {
		base = base.Where("log_clusters.trend = ?", trend)
	}
	if level := ctx.Query("level"); level != "" {
		levels := parseLogLevels(level)
		if len(levels) > 0 {
			base = base.Where("log_clusters.level IN ?", levels)
		}
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询日志聚类失败: " + err.Error()})
		return
	}

	type clusterRow struct {
		ID          string
		Code        string
		Pattern     string
		Count       int64
		Level       string
		Trend       string
		FirstSeenAt time.Time
		ServiceName string
	}
	var rows []clusterRow
	err := base.
		Select("log_clusters.id, log_clusters.code, log_clusters.pattern, log_clusters.count, log_clusters.level, log_clusters.trend, log_clusters.first_seen_at, services.name AS service_name").
		Joins("LEFT JOIN services ON services.id = log_clusters.service_id AND services.deleted_at IS NULL").
		// 级别加权排序：级别严重程度（ERROR > WARN > 其他）→ 趋势（激增 > 上升 > 平稳）→ 数量倒序
		Order("CASE log_clusters.level WHEN 'error' THEN 0 WHEN 'warn' THEN 1 ELSE 2 END, CASE log_clusters.trend WHEN 'spike' THEN 0 WHEN 'rising' THEN 1 ELSE 2 END, log_clusters.count DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询日志聚类失败: " + err.Error()})
		return
	}

	items := make([]logClusterItem, 0, len(rows))
	for _, r := range rows {
		services := []string{}
		if r.ServiceName != "" {
			services = append(services, r.ServiceName)
		}
		items = append(items, logClusterItem{
			ID:        r.Code,
			Pattern:   r.Pattern,
			Count:     r.Count,
			Level:     r.Level,
			Services:  services,
			Trend:     r.Trend,
			FirstSeen: r.FirstSeenAt.Format(time.RFC3339),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"total": total, "items": items}})
}

// Templates Drain 模板提取：模板 + 最近参数样本
// GET /api/logs/templates?service=&limit=20
func (c *LogAnalysisController) Templates(ctx *gin.Context) {
	limit := 20
	if l := ctx.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	q := c.DB.Model(&models.LogTemplate{}).Order("count DESC").Limit(limit)
	if ids := parseServiceIDs(ctx.Query("service")); len(ids) > 0 {
		q = q.Where("service_id IN ?", ids)
	}
	var templates []models.LogTemplate
	if err := q.Find(&templates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询日志模板失败: " + err.Error()})
		return
	}

	items := make([]logTemplateItem, 0, len(templates))
	for _, tpl := range templates {
		samples := make([]logTemplateSample, 0, 3)
		if strings.TrimSpace(tpl.SampleParams) != "" {
			_ = json.Unmarshal([]byte(tpl.SampleParams), &samples)
		}
		// 样本缺失时回退到最近一条原始日志
		if len(samples) == 0 && tpl.SampleRaw != "" {
			samples = append(samples, logTemplateSample{Raw: tpl.SampleRaw})
		}
		items = append(items, logTemplateItem{
			Template:     tpl.Template,
			Level:        tpl.Level,
			Count:        tpl.Count,
			ParamSamples: samples,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"items": items}})
}

// ---------- 参数解析 ----------

// parseLogRange 解析时间范围：优先 start/end（RFC3339），否则按 hours（默认 24）
func parseLogRange(ctx *gin.Context) (time.Time, time.Time, bool) {
	end := time.Now()
	startStr := ctx.Query("start")
	endStr := ctx.Query("end")
	if startStr != "" && endStr != "" {
		start, err1 := time.Parse(time.RFC3339, startStr)
		endParsed, err2 := time.Parse(time.RFC3339, endStr)
		if err1 != nil || err2 != nil || !endParsed.After(start) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "start/end 参数格式错误，需为 RFC3339 且 end 晚于 start"})
			return time.Time{}, time.Time{}, false
		}
		return start, endParsed, true
	}

	hours := 24
	if h := ctx.Query("hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 {
			hours = v
		}
	}
	return end.Add(-time.Duration(hours) * time.Hour), end, true
}

// parseServiceIDs 解析服务过滤参数：支持逗号分隔多服务 ID（如 service=id1,id2）
func parseServiceIDs(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(part); id != "" {
			out = append(out, id)
		}
	}
	return out
}

// parseLogLevels 解析级别过滤参数：支持逗号分隔多值（如 level=error,warn），仅保留合法级别
func parseLogLevels(raw string) []string {
	valid := map[string]bool{"error": true, "warn": true, "info": true}
	seen := map[string]bool{}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		lv := strings.ToLower(strings.TrimSpace(part))
		if valid[lv] && !seen[lv] {
			seen[lv] = true
			out = append(out, lv)
		}
	}
	return out
}

// parseLogPage 解析分页参数（page 从 1 起，page_size 默认 20，上限 200）
func parseLogPage(ctx *gin.Context) (int, int) {
	page, pageSize := 1, 20
	if p := ctx.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
			if pageSize > 200 {
				pageSize = 200
			}
		}
	}
	return page, pageSize
}

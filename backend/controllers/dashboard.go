package controllers

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardController 总览大盘控制器
type DashboardController struct {
	DB *gorm.DB
}

// NewDashboardController 创建控制器
func NewDashboardController(db *gorm.DB) *DashboardController {
	return &DashboardController{DB: db}
}

// ---------- 出参结构 ----------

type dashboardKPI struct {
	ActiveAlerts         int64   `json:"active_alerts"`
	ActiveAlertsDeltaPct float64 `json:"active_alerts_delta_pct"`
	AnomalyToday         int64   `json:"anomaly_today"`
	AnomalyDelta         int64   `json:"anomaly_delta"`
	CompressionRate      float64 `json:"compression_rate"`
	CompressionDeltaPct  float64 `json:"compression_delta_pct"`
	MttrMinutes          float64 `json:"mttr_minutes"`
	MttrDeltaPct         float64 `json:"mttr_delta_pct"`
}

type dashboardTrend struct {
	Buckets         []string `json:"buckets"`
	Raw             []int64  `json:"raw"`
	Denoised        []int64  `json:"denoised"`
	CompressionRate float64  `json:"compression_rate"`
}

type severityItem struct {
	Level string `json:"level"`
	Count int64  `json:"count"`
}

type serviceHealthItem struct {
	Service         string `json:"service"`
	Score           int    `json:"score"`
	AbnormalMetrics int64  `json:"abnormal_metrics"`
	Trend           string `json:"trend"`
}

type latestAlertItem struct {
	ID       string `json:"id"`
	Severity int    `json:"severity"`
	Content  string `json:"content"`
	Service  string `json:"service"`
	Time     string `json:"time"`
}

// Overview 总览大盘聚合数据
// GET /api/dashboard/overview?hours=24 或 ?start=<RFC3339>&end=<RFC3339>
func (c *DashboardController) Overview(ctx *gin.Context) {
	start, end, ok := c.parseRange(ctx)
	if !ok {
		return
	}

	kpi := c.buildKPI(start, end)
	trend := c.buildTrend(start, end)
	severity := c.buildSeverityDist(start, end)
	health := c.buildServiceHealth(start, end)
	latest := c.buildLatestAlerts(start, end)

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"kpi":            kpi,
		"trend":          trend,
		"severity_dist":  severity,
		"service_health": health,
		"latest_alerts":  latest,
	}})
}

// parseRange 解析时间范围：优先 start/end（RFC3339），否则按 hours（默认 24）
func (c *DashboardController) parseRange(ctx *gin.Context) (time.Time, time.Time, bool) {
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

// ---------- KPI ----------

func (c *DashboardController) buildKPI(start, end time.Time) dashboardKPI {
	var kpi dashboardKPI

	// 当前未恢复告警数（不限时间范围）
	c.DB.Model(&models.AlertEvent{}).
		Where("status = ? AND type = ?", models.AlertEventStatusFiring, models.AlertEventTypeAlert).
		Count(&kpi.ActiveAlerts)

	// 今日 0 点（本地时区）
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := dayStart.Add(-24 * time.Hour)
	yesterdayEnd := now.Add(-24 * time.Hour)

	todayAlerts := c.countAlertEvents(dayStart, now)
	yesterdayAlerts := c.countAlertEvents(yesterdayStart, yesterdayEnd)
	kpi.AnomalyToday = todayAlerts
	kpi.AnomalyDelta = todayAlerts - yesterdayAlerts
	// 活动告警「较昨日变化」：对比 24 小时前的未恢复告警快照
	kpi.ActiveAlertsDeltaPct = round1(pctChange(float64(kpi.ActiveAlerts), float64(c.countActiveAt(yesterdayEnd))))

	// 降噪压缩率：今日 records / (alert 事件 + records)
	todayRecords := c.countDenoiseRecords(dayStart, now)
	yesterdayRecords := c.countDenoiseRecords(yesterdayStart, yesterdayEnd)
	todayRate := compressionRate(todayAlerts, todayRecords)
	yesterdayRate := compressionRate(yesterdayAlerts, yesterdayRecords)
	kpi.CompressionRate = round1(todayRate)
	kpi.CompressionDeltaPct = round1(todayRate - yesterdayRate)

	// MTTR：范围内已恢复事件的平均恢复耗时（分钟）
	kpi.MttrMinutes = round1(c.avgMttr(start, end))
	span := end.Sub(start)
	kpi.MttrDeltaPct = round1(pctChange(c.avgMttr(start, end), c.avgMttr(start.Add(-span), start)))

	return kpi
}

// countAlertEvents 统计范围内 type=alert 事件数
func (c *DashboardController) countAlertEvents(start, end time.Time) int64 {
	var n int64
	c.DB.Model(&models.AlertEvent{}).
		Where("type = ? AND trigger_time >= ? AND trigger_time < ?", models.AlertEventTypeAlert, start, end).
		Count(&n)
	return n
}

// countActiveAt 统计指定时间点处于未恢复状态的告警事件数（触发于该时点之前且当时尚未恢复）
func (c *DashboardController) countActiveAt(t time.Time) int64 {
	var n int64
	c.DB.Model(&models.AlertEvent{}).
		Where("type = ? AND trigger_time < ? AND (recovered_at IS NULL OR recovered_at > ?)",
			models.AlertEventTypeAlert, t, t).
		Count(&n)
	return n
}

// countDenoiseRecords 统计范围内降噪拦截记录数
func (c *DashboardController) countDenoiseRecords(start, end time.Time) int64 {
	var n int64
	c.DB.Model(&models.DenoiseRecord{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&n)
	return n
}

// avgMttr 范围内 recovered_at 非空事件的平均恢复耗时（分钟），无数据返回 0
func (c *DashboardController) avgMttr(start, end time.Time) float64 {
	var avgSeconds sql.NullFloat64
	err := c.DB.Model(&models.AlertEvent{}).
		Select("AVG(TIMESTAMPDIFF(SECOND, trigger_time, recovered_at))").
		Where("recovered_at IS NOT NULL AND trigger_time >= ? AND trigger_time < ?", start, end).
		Scan(&avgSeconds).Error
	if err != nil || !avgSeconds.Valid {
		return 0
	}
	return avgSeconds.Float64 / 60
}

// ---------- 趋势 ----------

func (c *DashboardController) buildTrend(start, end time.Time) dashboardTrend {
	span := end.Sub(start)
	// >24h 按天桶，否则按小时桶
	var width time.Duration
	dayBucket := span > 24*time.Hour
	if dayBucket {
		days := int64(math.Ceil(span.Hours() / 24))
		if days < 1 {
			days = 1
		}
		width = time.Duration(days) * 24 * time.Hour
	} else {
		width = time.Hour
	}

	bucketCount := int(math.Ceil(span.Seconds() / width.Seconds()))
	if bucketCount < 1 {
		bucketCount = 1
	}

	trend := dashboardTrend{
		Buckets:  make([]string, 0, bucketCount),
		Raw:      make([]int64, 0, bucketCount),
		Denoised: make([]int64, 0, bucketCount),
	}

	var totalEvents, totalRecords int64
	for i := 0; i < bucketCount; i++ {
		bStart := start.Add(time.Duration(i) * width)
		bEnd := bStart.Add(width)
		if bEnd.After(end) {
			bEnd = end
		}

		events := c.countAlertEvents(bStart, bEnd)
		records := c.countDenoiseRecords(bStart, bEnd)
		totalEvents += events
		totalRecords += records

		if dayBucket {
			trend.Buckets = append(trend.Buckets, bStart.Format("01-02"))
		} else {
			trend.Buckets = append(trend.Buckets, bStart.Format("15:04"))
		}
		trend.Raw = append(trend.Raw, events+records)
		trend.Denoised = append(trend.Denoised, events)
	}

	trend.CompressionRate = round1(compressionRate(totalEvents, totalRecords))
	return trend
}

// ---------- 级别分布 ----------

func (c *DashboardController) buildSeverityDist(start, end time.Time) []severityItem {
	type row struct {
		Severity int
		Count    int64
	}
	var rows []row
	c.DB.Model(&models.AlertEvent{}).
		Select("severity, COUNT(*) AS count").
		Where("type = ? AND trigger_time >= ? AND trigger_time < ?", models.AlertEventTypeAlert, start, end).
		Group("severity").
		Scan(&rows)

	result := make([]severityItem, 0, len(rows))
	for _, r := range rows {
		if r.Severity < 1 || r.Severity > 3 || r.Count == 0 {
			continue
		}
		result = append(result, severityItem{
			Level: "P" + strconv.Itoa(r.Severity),
			Count: r.Count,
		})
	}
	return result
}

// ---------- 服务健康 ----------

func (c *DashboardController) buildServiceHealth(start, end time.Time) []serviceHealthItem {
	var services []models.Service
	if err := c.DB.Find(&services).Error; err != nil || len(services) == 0 {
		return []serviceHealthItem{}
	}

	span := end.Sub(start)
	prevStart := start.Add(-span)

	items := make([]serviceHealthItem, 0, len(services))
	for _, svc := range services {
		abnormal := c.countServiceFiring(svc)
		score := 100 - 20*int(abnormal)
		if svc.Verified != 1 {
			score -= 10
		}
		if score < 0 {
			score = 0
		}

		cur := c.countServiceAlerts(svc, start, end)
		prev := c.countServiceAlerts(svc, prevStart, start)
		trendLabel := "平稳"
		if cur > prev {
			trendLabel = "恶化"
		} else if cur < prev {
			trendLabel = "好转"
		}

		items = append(items, serviceHealthItem{
			Service:         svc.Name,
			Score:           score,
			AbnormalMetrics: abnormal,
			Trend:           trendLabel,
		})
	}

	// 分数差的排前面，最多取 8 个（服务数量少，直接插入排序）
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j-1].Score > items[j].Score; j-- {
			items[j-1], items[j] = items[j], items[j-1]
		}
	}
	if len(items) > 8 {
		items = items[:8]
	}
	return items
}

// serviceTagCond 事件 tags 含服务 code 或 name 的条件（参数化 LIKE）
func serviceTagCond(svc models.Service) (string, []interface{}) {
	conds := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if svc.Code != "" {
		conds = append(conds, "tags LIKE ?")
		args = append(args, "%"+svc.Code+"%")
	}
	if svc.Name != "" {
		conds = append(conds, "tags LIKE ?")
		args = append(args, "%"+svc.Name+"%")
	}
	if len(conds) == 0 {
		return "1 = 0", nil
	}
	return "(" + strings.Join(conds, " OR ") + ")", args
}

// countServiceFiring 该服务当前 firing 事件数
func (c *DashboardController) countServiceFiring(svc models.Service) int64 {
	cond, args := serviceTagCond(svc)
	var n int64
	q := c.DB.Model(&models.AlertEvent{}).
		Where("status = ? AND type = ?", models.AlertEventStatusFiring, models.AlertEventTypeAlert).
		Where(cond, args...)
	q.Count(&n)
	return n
}

// countServiceAlerts 该服务范围内 alert 事件数
func (c *DashboardController) countServiceAlerts(svc models.Service, start, end time.Time) int64 {
	cond, args := serviceTagCond(svc)
	var n int64
	q := c.DB.Model(&models.AlertEvent{}).
		Where("type = ? AND trigger_time >= ? AND trigger_time < ?", models.AlertEventTypeAlert, start, end).
		Where(cond, args...)
	q.Count(&n)
	return n
}

// ---------- 最新告警 ----------

func (c *DashboardController) buildLatestAlerts(start, end time.Time) []latestAlertItem {
	var events []models.AlertEvent
	c.DB.Where("type = ? AND severity <= 2 AND trigger_time >= ? AND trigger_time < ?",
		models.AlertEventTypeAlert, start, end).
		Order("trigger_time DESC").
		Limit(10).
		Find(&events)

	result := make([]latestAlertItem, 0, len(events))
	for _, e := range events {
		content := e.RuleName
		if content == "" {
			content = e.TargetIdent
		}
		result = append(result, latestAlertItem{
			ID:       e.ID,
			Severity: e.Severity,
			Content:  content,
			Service:  parseTagValue(e.Tags, "service"),
			Time:     e.TriggerTime.Format("1-2 15:04"),
		})
	}
	return result
}

// ---------- 工具函数 ----------

// pctChange 百分比变化：(cur - prev) / prev × 100，prev 为 0 时返回 0
func pctChange(cur, prev float64) float64 {
	if prev == 0 {
		return 0
	}
	return (cur - prev) * 100 / prev
}

// compressionRate 压缩率：records / (events + records) × 100，分母 0 返回 0
func compressionRate(events, records int64) float64 {
	total := events + records
	if total == 0 {
		return 0
	}
	return float64(records) * 100 / float64(total)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// parseTagValue 从 "k=v,k2=v2" 文本中解析指定 key 的值，无则返回 "-"
func parseTagValue(tags, key string) string {
	for _, pair := range strings.Split(tags, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == key {
			if v := strings.TrimSpace(kv[1]); v != "" {
				return v
			}
		}
	}
	return "-"
}

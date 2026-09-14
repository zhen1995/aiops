package loganalysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"aiops/internal/datasource"
	"aiops/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultInterval = 5 * time.Minute    // 默认分析间隔
	pythonBatchSize = 500                // 单次调用 Python 的日志批大小
	maxSamples      = 3                  // 模板保留的最近样本条数
	statRetention   = 7 * 24 * time.Hour // 分桶统计保留时长

	spikeMultiple  = 3.0 // spike：当前 >= 上一窗口 × 3
	risingMultiple = 1.5 // rising：当前 >= 上一窗口 × 1.5
	spikeMinCount  = 50  // spike 最小阈值
)

// errPythonUnavailable Python 解析服务整体不可用的哨兵错误（本轮其余服务跳过）
var errPythonUnavailable = errors.New("python 日志解析服务不可用")

// logSample 模板最近样本：原始日志 + 参数列表
type logSample struct {
	Raw    string   `json:"raw"`
	Params []string `json:"params"`
}

// Analyzer 日志周期分析器：按服务增量拉取 ES 日志，经 Python Drain 解析后落库
type Analyzer struct {
	db       *gorm.DB
	client   *Client
	interval time.Duration

	mu      sync.Mutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
}

// NewAnalyzer 创建分析器；intervalStr 如 "5m"，解析失败或非法时兜底默认 5 分钟
func NewAnalyzer(db *gorm.DB, pythonBaseURL, intervalStr string) *Analyzer {
	interval := defaultInterval
	if d, err := time.ParseDuration(strings.TrimSpace(intervalStr)); err == nil && d > 0 {
		interval = d
	}
	return &Analyzer{
		db:       db,
		client:   NewClient(pythonBaseURL),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动分析器：立即执行一轮，此后按间隔周期执行
func (a *Analyzer) Start() {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.mu.Unlock()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		log.Printf("[日志分析] 分析器已启动，间隔 %s", a.interval)
		a.runRound()
		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()
		for {
			select {
			case <-a.stopCh:
				return
			case <-ticker.C:
				a.runRound()
			}
		}
	}()
}

// Stop 优雅停止分析器
func (a *Analyzer) Stop() {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return
	}
	close(a.stopCh)
	a.running = false
	a.mu.Unlock()
	a.wg.Wait()
	log.Printf("[日志分析] 分析器已停止")
}

// runRound 执行一轮分析：逐个服务串行处理，单服务失败仅记日志不影响其他服务
func (a *Analyzer) runRound() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var services []models.Service
	if err := a.db.WithContext(ctx).
		Where("status = 1 AND es_index_patterns IS NOT NULL AND es_index_patterns != ''").
		Find(&services).Error; err != nil {
		log.Printf("[日志分析] 查询服务列表失败: %v", err)
		return
	}
	if len(services) == 0 {
		return
	}

	pythonDown := false
	for i := range services {
		if pythonDown {
			log.Printf("[日志分析] Python 解析服务不可用，跳过服务 %s", services[i].Name)
			continue
		}
		if err := a.analyzeService(ctx, &services[i]); err != nil {
			if errors.Is(err, errPythonUnavailable) {
				pythonDown = true
				log.Printf("[日志分析] %v", err)
			} else {
				log.Printf("[日志分析] 服务 %s 分析失败: %v", services[i].Name, err)
			}
		}
	}

	a.cleanupStats(ctx)
}

// logEntry 单条拉取到的日志：原文 + 时间 + 级别提示
type logEntry struct {
	raw       string
	ts        time.Time
	levelHint string
}

// analyzeService 分析单个服务的一轮增量日志
func (a *Analyzer) analyzeService(ctx context.Context, svc *models.Service) error {
	start, end, err := a.cursorRange(svc.ID)
	if err != nil {
		return fmt.Errorf("读取分析游标失败: %w", err)
	}
	patterns := parseIndexPatterns(svc.ESIndexPatterns)
	if len(patterns) == 0 {
		return nil
	}

	// 加载服务关联的 ES 数据源（只读）
	var ds models.Datasource
	if err := a.db.WithContext(ctx).First(&ds, "id = ?", svc.ESDatasourceID).Error; err != nil {
		return fmt.Errorf("ES 数据源不可用: %w", err)
	}
	if ds.IsEnabled != 1 {
		return nil
	}
	esClient := datasource.NewElasticsearchClient(&ds)

	// 按索引模式增量拉取（累计不超过单服务本轮上限）
	entries := make([]logEntry, 0, datasource.MaxLogAnalysisHits)
	for _, pattern := range patterns {
		if len(entries) >= datasource.MaxLogAnalysisHits {
			break
		}
		res, err := esClient.SearchLogs(ctx, pattern,
			start.Format(time.RFC3339), end.Format(time.RFC3339),
			datasource.MaxLogAnalysisHits-len(entries))
		if err != nil {
			log.Printf("[日志分析] 服务 %s 索引 %s 拉取失败: %v", svc.Name, pattern, err)
			continue
		}
		for _, h := range res.Hits {
			raw := extractMessage(h.Source)
			if raw == "" {
				continue
			}
			entries = append(entries, logEntry{
				raw:       raw,
				ts:        extractTimestamp(h.Source, time.Now()),
				levelHint: extractLevel(h.Source),
			})
		}
	}

	// 无日志也推进游标，避免重复扫描空窗
	if len(entries) == 0 {
		a.saveCursor(svc.ID, end)
		return nil
	}

	// 批量调用 Python Drain 解析（Python 失败不推进游标，下一轮重试）
	logs := make([]string, len(entries))
	for i, e := range entries {
		logs[i] = e.raw
	}
	results, err := a.parseBatches(ctx, svc.ID, logs)
	if err != nil {
		return fmt.Errorf("%w: 服务 %s 调用解析失败: %v", errPythonUnavailable, svc.Name, err)
	}

	if err := a.mergeResults(ctx, svc, entries, results); err != nil {
		return fmt.Errorf("结果落库失败: %w", err)
	}
	a.saveCursor(svc.ID, end)
	log.Printf("[日志分析] 服务 %s 本轮分析 %d 条日志", svc.Name, len(results))
	return nil
}

// parseBatches 分批调用 Python 解析并合并结果
func (a *Analyzer) parseBatches(ctx context.Context, serviceID string, logs []string) ([]ParseResult, error) {
	all := make([]ParseResult, 0, len(logs))
	for from := 0; from < len(logs); from += pythonBatchSize {
		to := from + pythonBatchSize
		if to > len(logs) {
			to = len(logs)
		}
		batch, err := a.client.Parse(ctx, serviceID, logs[from:to])
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
	}
	return all, nil
}

// ---------- 结果合并落库 ----------

// templateAgg 单模板本轮的归并结果
type templateAgg struct {
	template string
	level    string
	count    int64
	samples  []logSample // 最新在前，最多 maxSamples 条
	firstAt  time.Time
	lastAt   time.Time
}

// bucketAgg 单分桶本轮的计数
type bucketAgg struct {
	total int64
	err   int64
}

// clusterAgg 单聚类本轮的归并结果
type clusterAgg struct {
	code     string
	template string
	level    string
	count    int64
	firstAt  time.Time
	lastAt   time.Time
	buckets  map[int64]*bucketAgg // bucketStart(unix) → 计数
}

// mergeResults 按 模板/聚类/分桶 归并统计并落库，随后刷新趋势打标
func (a *Analyzer) mergeResults(ctx context.Context, svc *models.Service, entries []logEntry, results []ParseResult) error {
	tplMap := make(map[string]*templateAgg)
	clMap := make(map[string]*clusterAgg)

	n := len(entries)
	if len(results) < n {
		n = len(results)
	}
	for i := 0; i < n; i++ {
		r := results[i]
		e := entries[i]
		level := normalizeLevel(r.Level, e.levelHint)

		tpl := tplMap[r.Template]
		if tpl == nil {
			tpl = &templateAgg{template: r.Template, firstAt: e.ts, lastAt: e.ts}
			tplMap[r.Template] = tpl
		}
		tpl.count++
		if level != "" {
			tpl.level = level
		}
		tpl.samples = pushSample(tpl.samples, logSample{Raw: e.raw, Params: r.Params})
		if e.ts.Before(tpl.firstAt) {
			tpl.firstAt = e.ts
		}
		if e.ts.After(tpl.lastAt) {
			tpl.lastAt = e.ts
		}

		code := fmt.Sprintf("C-%03d", r.ClusterID)
		cl := clMap[code]
		if cl == nil {
			cl = &clusterAgg{
				code:     code,
				template: r.Template,
				firstAt:  e.ts,
				lastAt:   e.ts,
				buckets:  make(map[int64]*bucketAgg),
			}
			clMap[code] = cl
		}
		cl.count++
		if level != "" {
			cl.level = level
		}
		if e.ts.Before(cl.firstAt) {
			cl.firstAt = e.ts
		}
		if e.ts.After(cl.lastAt) {
			cl.lastAt = e.ts
		}
		bucketStart := e.ts.Truncate(time.Hour).Unix()
		b := cl.buckets[bucketStart]
		if b == nil {
			b = &bucketAgg{}
			cl.buckets[bucketStart] = b
		}
		b.total++
		if level == "error" {
			b.err++
		}
	}

	// 库内已有首现时间可能更早，落库时取更小值
	tplIDs := make(map[string]string, len(tplMap))
	clIDs := make(map[string]string, len(clMap))
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 模板：按 service_id + template 幂等更新
		for _, agg := range tplMap {
			var tpl models.LogTemplate
			err := tx.Where("service_id = ? AND template = ?", svc.ID, agg.template).First(&tpl).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tpl = models.LogTemplate{
					ServiceID:    svc.ID,
					Template:     agg.template,
					Level:        orDefault(agg.level, "info"),
					Count:        agg.count,
					SampleRaw:    agg.samples[0].Raw,
					SampleParams: marshalSamples(agg.samples),
					FirstSeenAt:  agg.firstAt,
					LastSeenAt:   agg.lastAt,
				}
				if err := tx.Create(&tpl).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				samples := mergeSamples(parseSamples(tpl.SampleParams), agg.samples)
				updates := map[string]interface{}{
					"count":         tpl.Count + agg.count,
					"level":         orDefault(agg.level, tpl.Level),
					"sample_raw":    orDefault(agg.samples[0].Raw, tpl.SampleRaw),
					"sample_params": marshalSamples(samples),
					"first_seen_at": minTime(tpl.FirstSeenAt, agg.firstAt),
					"last_seen_at":  maxTime(tpl.LastSeenAt, agg.lastAt),
				}
				if err := tx.Model(&tpl).Updates(updates).Error; err != nil {
					return err
				}
			}
			tplIDs[agg.template] = tpl.ID
		}

		// 2) 聚类：按 service_id + code 幂等更新
		for _, agg := range clMap {
			var cl models.LogCluster
			err := tx.Where("service_id = ? AND code = ?", svc.ID, agg.code).First(&cl).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cl = models.LogCluster{
					ServiceID:   svc.ID,
					TemplateID:  tplIDs[agg.template],
					Code:        agg.code,
					Pattern:     agg.template,
					Level:       orDefault(agg.level, "info"),
					Count:       agg.count,
					Trend:       models.LogClusterTrendFlat,
					FirstSeenAt: agg.firstAt,
					LastSeenAt:  agg.lastAt,
				}
				if err := tx.Create(&cl).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				updates := map[string]interface{}{
					"template_id":   tplIDs[agg.template],
					"pattern":       agg.template,
					"level":         orDefault(agg.level, cl.Level),
					"count":         cl.Count + agg.count,
					"first_seen_at": minTime(cl.FirstSeenAt, agg.firstAt),
					"last_seen_at":  maxTime(cl.LastSeenAt, agg.lastAt),
				}
				if err := tx.Model(&cl).Updates(updates).Error; err != nil {
					return err
				}
			}
			clIDs[agg.code] = cl.ID
		}

		// 3) 分桶统计：按 cluster_id + bucket_start 累加
		for code, agg := range clMap {
			for unix, b := range agg.buckets {
				stat := models.LogClusterStat{
					ClusterID:   clIDs[code],
					ServiceID:   svc.ID,
					BucketStart: time.Unix(unix, 0),
					TotalCount:  b.total,
					ErrorCount:  b.err,
				}
				err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "cluster_id"}, {Name: "bucket_start"}},
					DoUpdates: clause.Assignments(map[string]interface{}{
						"total_count": gorm.Expr("total_count + VALUES(total_count)"),
						"error_count": gorm.Expr("error_count + VALUES(error_count)"),
					}),
				}).Create(&stat).Error
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 4) 趋势打标：本日已聚合分桶 vs 上一等长窗口
	for code, clID := range clIDs {
		trend := a.computeTrend(clID)
		if err := a.db.WithContext(ctx).Model(&models.LogCluster{}).
			Where("id = ?", clID).Update("trend", trend).Error; err != nil {
			log.Printf("[日志分析] 更新聚类 %s 趋势失败: %v", code, err)
		}
	}
	return nil
}

// computeTrend 比较聚类本日（dayStart→now）与上一等长窗口的分桶总量，输出 spike/rising/flat
func (a *Analyzer) computeTrend(clusterID string) string {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	prevStart := dayStart.Add(-now.Sub(dayStart))

	var cur, prev int64
	a.db.Model(&models.LogClusterStat{}).
		Where("cluster_id = ? AND bucket_start >= ?", clusterID, dayStart).
		Select("COALESCE(SUM(total_count), 0)").
		Scan(&cur)
	a.db.Model(&models.LogClusterStat{}).
		Where("cluster_id = ? AND bucket_start >= ? AND bucket_start < ?", clusterID, prevStart, dayStart).
		Select("COALESCE(SUM(total_count), 0)").
		Scan(&prev)

	if cur >= spikeMinCount && float64(cur) >= spikeMultiple*float64(prev) {
		return models.LogClusterTrendSpike
	}
	if prev > 0 && float64(cur) >= risingMultiple*float64(prev) {
		return models.LogClusterTrendRising
	}
	return models.LogClusterTrendFlat
}

// ---------- 游标与清理 ----------

// cursorRange 读取服务增量游标：无记录或异常时取本轮 end 前一个间隔
func (a *Analyzer) cursorRange(serviceID string) (time.Time, time.Time, error) {
	end := time.Now()
	var cursor models.LogAnalysisCursor
	err := a.db.Where("service_id = ?", serviceID).First(&cursor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return end.Add(-a.interval), end, nil
	}
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start := cursor.LastAnalyzedAt
	if start.IsZero() || !start.Before(end) {
		start = end.Add(-a.interval)
	}
	return start, end, nil
}

// saveCursor 推进游标 last_analyzed_at = 本轮 end
func (a *Analyzer) saveCursor(serviceID string, t time.Time) {
	cursor := models.LogAnalysisCursor{ServiceID: serviceID, LastAnalyzedAt: t}
	err := a.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "service_id"}},
		UpdateAll: true,
	}).Create(&cursor).Error
	if err != nil {
		log.Printf("[日志分析] 推进服务 %s 游标失败: %v", serviceID, err)
	}
}

// cleanupStats 清理 7 天前的分桶统计
func (a *Analyzer) cleanupStats(ctx context.Context) {
	deadline := time.Now().Add(-statRetention)
	if err := a.db.WithContext(ctx).
		Where("bucket_start < ?", deadline).
		Delete(&models.LogClusterStat{}).Error; err != nil {
		log.Printf("[日志分析] 清理过期分桶失败: %v", err)
	}
}

// ---------- 工具函数 ----------

// parseIndexPatterns 解析服务的 ES 索引模式（JSON 数组字符串，兼容单个模式）
func parseIndexPatterns(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var arr []string
	if json.Unmarshal([]byte(raw), &arr) == nil {
		out := make([]string, 0, len(arr))
		for _, p := range arr {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return []string{raw}
}

// extractMessage 取日志文本：优先 message/msg，其次第一个字符串字段，最后整条 _source
func extractMessage(src map[string]interface{}) string {
	for _, k := range []string{"message", "msg"} {
		if v, ok := src[k].(string); ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	for k, v := range src {
		if k == "@timestamp" || k == "timestamp" {
			continue
		}
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	if len(src) == 0 {
		return ""
	}
	b, _ := json.Marshal(src)
	return string(b)
}

// extractTimestamp 从 _source 提取日志时间，失败时回退 fallback
func extractTimestamp(src map[string]interface{}, fallback time.Time) time.Time {
	for _, k := range []string{"@timestamp", "timestamp"} {
		if v, ok := src[k].(string); ok && v != "" {
			if t, err := parseLogTime(v); err == nil {
				return t
			}
		}
	}
	return fallback
}

// parseLogTime 解析常见日志时间格式
func parseLogTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
}

// extractLevel 从 _source 提取级别提示（loglevel/level）
func extractLevel(src map[string]interface{}) string {
	for _, k := range []string{"level", "loglevel"} {
		if v, ok := src[k].(string); ok && v != "" {
			return strings.ToLower(v)
		}
	}
	return ""
}

// normalizeLevel 归一化级别：Python 结果优先，其次 ES 字段提示，兜底 info
func normalizeLevel(primary, hint string) string {
	if l := strings.ToLower(strings.TrimSpace(primary)); l != "" {
		return l
	}
	switch strings.ToLower(strings.TrimSpace(hint)) {
	case "error", "err", "fatal", "critical":
		return "error"
	case "warn", "warning":
		return "warn"
	case "info", "debug", "trace":
		return "info"
	}
	return "info"
}

// pushSample 前插样本（按 raw 去重），最多保留 maxSamples 条
func pushSample(samples []logSample, s logSample) []logSample {
	out := make([]logSample, 0, maxSamples)
	out = append(out, s)
	for _, old := range samples {
		if old.Raw == s.Raw {
			continue
		}
		if len(out) >= maxSamples {
			break
		}
		out = append(out, old)
	}
	return out
}

// mergeSamples 合并库内已有样本（旧）与本轮新样本（新在前），去重后最多保留 maxSamples 条
func mergeSamples(existing, incoming []logSample) []logSample {
	seen := make(map[string]bool)
	out := make([]logSample, 0, maxSamples)
	for _, s := range incoming {
		if seen[s.Raw] {
			continue
		}
		seen[s.Raw] = true
		out = append(out, s)
	}
	for _, s := range existing {
		if len(out) >= maxSamples {
			break
		}
		if seen[s.Raw] {
			continue
		}
		seen[s.Raw] = true
		out = append(out, s)
	}
	return out
}

func parseSamples(raw string) []logSample {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var samples []logSample
	if json.Unmarshal([]byte(raw), &samples) != nil {
		return nil
	}
	return samples
}

func marshalSamples(samples []logSample) string {
	if len(samples) == 0 {
		return "[]"
	}
	b, err := json.Marshal(samples)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func minTime(a, b time.Time) time.Time {
	if a.IsZero() || b.Before(a) {
		return b
	}
	return a
}

func maxTime(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

package datasource

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"aiops/models"
)

// PrometheusClient Prometheus 只读查询客户端
type PrometheusClient struct {
	ds models.Datasource
	hc *http.Client
}

// NewPrometheusClient 创建 Prometheus 客户端
func NewPrometheusClient(ds *models.Datasource) *PrometheusClient {
	timeout := ds.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout * 1000 // ms
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if ds.IsSkipSSL == 1 {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &PrometheusClient{
		ds: *ds,
		hc: &http.Client{
			Timeout:   time.Duration(timeout) * time.Millisecond,
			Transport: transport,
		},
	}
}

// QueryRange 执行 PromQL range 查询，并做聚合/截断
func (c *PrometheusClient) QueryRange(ctx context.Context, query, start, end, step string, topn, limit int) (*PrometheusQueryResult, error) {
	if query == "" {
		return nil, fmt.Errorf("promql 不能为空")
	}

	now := time.Now()
	startT, err := ParseTime(start, now)
	if err != nil {
		return nil, fmt.Errorf("解析 start 失败: %w", err)
	}
	endT, err := ParseTime(end, now)
	if err != nil {
		return nil, fmt.Errorf("解析 end 失败: %w", err)
	}
	if endT.Before(startT) {
		endT = now
	}

	if step == "" {
		step = autoStep(startT, endT)
	}

	base, err := url.Parse(c.ds.URL)
	if err != nil {
		return nil, fmt.Errorf("解析数据源 URL 失败: %w", err)
	}
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), "api/v1/query_range")

	q := base.Query()
	q.Set("query", query)
	q.Set("start", strconv.FormatInt(startT.Unix(), 10))
	q.Set("end", strconv.FormatInt(endT.Unix(), 10))
	q.Set("step", step)
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.ds.Username != "" {
		req.SetBasicAuth(c.ds.Username, c.ds.Password)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Prometheus 失败: %w", err)
	}
	defer resp.Body.Close()

	var payload promResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析 Prometheus 响应失败: %w", err)
	}
	if payload.Status != "success" {
		return nil, fmt.Errorf("Prometheus 查询失败: %s", payload.Error)
	}

	result := &PrometheusQueryResult{
		Query:      query,
		DataSource: c.ds.Name,
		Series:     make([]*PrometheusSeries, 0, len(payload.Data.Result)),
	}

	for _, r := range payload.Data.Result {
		s := &PrometheusSeries{
			Labels:  r.Metric,
			Values:  make([][2]interface{}, 0, len(r.Values)),
			Summary: map[string]float64{},
		}
		s.Name = seriesName(r.Metric)

		var minV, maxV, sumV float64
		var count int
		var latest float64
		for _, pair := range r.Values {
			if len(pair) < 2 {
				continue
			}
			ts, ok1 := parseTimestamp(pair[0])
			v, ok2 := parseFloat(pair[1])
			if !ok1 || !ok2 {
				continue
			}
			s.Values = append(s.Values, [2]interface{}{ts, v})
			latest = v
			if count == 0 || v < minV {
				minV = v
			}
			if count == 0 || v > maxV {
				maxV = v
			}
			sumV += v
			count++
		}
		if count == 0 {
			continue
		}
		s.Summary["min"] = minV
		s.Summary["max"] = maxV
		s.Summary["avg"] = sumV / float64(count)
		s.Summary["latest"] = latest
		s.Values = downsample(s.Values, MaxPointsPerSeries)
		result.Series = append(result.Series, s)
	}

	result.Series = topNSeries(result.Series, topn, MaxSeries)
	return result, nil
}

// 内部响应结构
type promResponse struct {
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
	Data      struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string  `json:"metric"`
			Values [][2]interface{}   `json:"values,omitempty"`
			Value  [2]interface{}     `json:"value,omitempty"`
		} `json:"result"`
	} `json:"data"`
}

func seriesName(metric map[string]string) string {
	if name, ok := metric["__name__"]; ok {
		return name
	}
	if len(metric) == 0 {
		return "unknown"
	}
	parts := make([]string, 0, len(metric))
	for k, v := range metric {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func parseTimestamp(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case string:
		if i, err := strconv.ParseInt(t, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

func parseFloat(v interface{}) (float64, bool) {
	switch f := v.(type) {
	case float64:
		return f, true
	case string:
		if fl, err := strconv.ParseFloat(f, 64); err == nil {
			return fl, true
		}
	}
	return 0, false
}

func autoStep(start, end time.Time) string {
	d := end.Sub(start)
	switch {
	case d <= time.Hour:
		return "15s"
	case d <= 6*time.Hour:
		return "1m"
	case d <= 24*time.Hour:
		return "5m"
	case d <= 7*24*time.Hour:
		return "30m"
	default:
		return "1h"
	}
}

func downsample(values [][2]interface{}, max int) [][2]interface{} {
	if len(values) <= max {
		return values
	}
	step := float64(len(values)) / float64(max)
	out := make([][2]interface{}, 0, max)
	for i := 0; i < max; i++ {
		idx := int(float64(i) * step)
		if idx >= len(values) {
			idx = len(values) - 1
		}
		out = append(out, values[idx])
	}
	return out
}

func topNSeries(series []*PrometheusSeries, topn, max int) []*PrometheusSeries {
	limit := max
	if topn > 0 && topn < limit {
		limit = topn
	}
	if len(series) <= limit {
		return series
	}
	sort.SliceStable(series, func(i, j int) bool {
		return series[i].Summary["latest"] > series[j].Summary["latest"]
	})
	return series[:limit]
}

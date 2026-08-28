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

// PyroscopeClient Pyroscope 只读查询客户端
type PyroscopeClient struct {
	ds models.Datasource
	hc *http.Client
}

// NewPyroscopeClient 创建 Pyroscope 客户端
func NewPyroscopeClient(ds *models.Datasource) *PyroscopeClient {
	timeout := ds.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout * 1000
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if ds.IsSkipSSL == 1 {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &PyroscopeClient{
		ds: *ds,
		hc: &http.Client{
			Timeout:   time.Duration(timeout) * time.Millisecond,
			Transport: transport,
		},
	}
}

// Render 调用 Pyroscope /pyroscope/render 接口，返回聚合后的 Top 函数与时间线
func (c *PyroscopeClient) Render(ctx context.Context, query, start, end string, maxNodes int) (*PyroscopeQueryResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query 不能为空")
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

	if maxNodes <= 0 {
		maxNodes = 1024
	}

	base, err := url.Parse(c.ds.URL)
	if err != nil {
		return nil, fmt.Errorf("解析数据源 URL 失败: %w", err)
	}
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), "pyroscope/render")

	q := base.Query()
	q.Set("query", query)
	q.Set("from", strconv.FormatInt(startT.Unix(), 10))
	q.Set("until", strconv.FormatInt(endT.Unix(), 10))
	q.Set("format", "json")
	q.Set("maxNodes", strconv.Itoa(maxNodes))
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
		return nil, fmt.Errorf("请求 Pyroscope 失败: %w", err)
	}
	defer resp.Body.Close()

	var payload pyroResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析 Pyroscope 响应失败: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("Pyroscope 查询失败: %v", payload.Error)
	}

	result := &PyroscopeQueryResult{
		DataSource:  c.ds.Name,
		Query:       query,
		ProfileType: payload.Metadata.Name,
		Units:       string(payload.Metadata.Units),
		SampleRate:  payload.Metadata.SampleRate,
		NumTicks:    payload.Flamebearer.NumTicks,
		TopFunctions: aggregateFlamebearer(payload.Flamebearer.Names, payload.Flamebearer.Levels, payload.Flamebearer.NumTicks, MaxSeries),
	}

	if payload.Timeline != nil {
		result.Timeline = &PyroscopeTimeline{
			StartTime:     payload.Timeline.StartTime,
			DurationDelta: payload.Timeline.DurationDelta,
			Samples:       payload.Timeline.Samples,
		}
	}
	return result, nil
}

// 内部响应结构
type pyroResponse struct {
	Error    map[string]interface{} `json:"error,omitempty"`
	Flamebearer struct {
		Names    []string `json:"names"`
		Levels   [][]int  `json:"levels"`
		NumTicks int      `json:"numTicks"`
		MaxSelf  int      `json:"maxSelf"`
	} `json:"flamebearer"`
	Metadata struct {
		Format     string `json:"format"`
		SpyName    string `json:"spyName"`
		SampleRate uint32 `json:"sampleRate"`
		Units      string `json:"units"`
		Name       string `json:"name"`
	} `json:"metadata"`
	Timeline *struct {
		StartTime     int64    `json:"startTime"`
		Samples       []uint64 `json:"samples"`
		DurationDelta int64    `json:"durationDelta"`
	} `json:"timeline,omitempty"`
}

func aggregateFlamebearer(names []string, levels [][]int, numTicks, topN int) []*PyroscopeFunction {
	selfMap := make(map[string]int)
	totalMap := make(map[string]int)

	for _, level := range levels {
		for j := 0; j+3 < len(level); j += 4 {
			// single 格式：[offset, total, self, nameIndex]
			total := level[j+1]
			self := level[j+2]
			nameIdx := level[j+3]
			if nameIdx < 0 || nameIdx >= len(names) {
				continue
			}
			name := names[nameIdx]
			selfMap[name] += self
			totalMap[name] += total
		}
	}

	funcs := make([]*PyroscopeFunction, 0, len(selfMap))
	for name, self := range selfMap {
		percent := 0.0
		if numTicks > 0 {
			percent = float64(self) / float64(numTicks) * 100
		}
		funcs = append(funcs, &PyroscopeFunction{
			Name:        name,
			Self:        self,
			Total:       totalMap[name],
			SelfPercent: round2(percent),
		})
	}

	sort.SliceStable(funcs, func(i, j int) bool {
		return funcs[i].Self > funcs[j].Self
	})

	if len(funcs) > topN {
		funcs = funcs[:topN]
	}
	return funcs
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

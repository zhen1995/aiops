package datasource

// 数据源类型常量，与 models.Datasource.Type 保持一致
const (
	TypePrometheus    = "Prometheus"
	TypeElasticsearch = "ElasticSearch"
	TypePyroscope     = "Pyroscope"
)

// 通用限制常量
const (
	DefaultTimeout     = 15 // 秒
	MaxSeries          = 10
	MaxPointsPerSeries = 100
	MaxESHits          = 20
	MaxResultBytes     = 64 * 1024
)

// PrometheusSeries 单个时序序列的聚合结果
type PrometheusSeries struct {
	Name    string            `json:"name"`
	Labels  map[string]string `json:"labels,omitempty"`
	Summary map[string]float64 `json:"summary"`
	Values  [][2]interface{}  `json:"values,omitempty"`
}

// InstantSample 即时查询返回的单个样本
type InstantSample struct {
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
}

// InstantQueryResult PromQL 即时查询结果（告警规则评估用）
type InstantQueryResult struct {
	DataSource string          `json:"data_source"`
	Hit        bool            `json:"hit"` // 查询结果是否非空
	Samples    []InstantSample `json:"samples"`
}

// PrometheusQueryResult Prometheus 查询返回给 LLM 的结构
type PrometheusQueryResult struct {
	Query      string              `json:"query"`
	DataSource string              `json:"data_source"`
	Series     []*PrometheusSeries `json:"series"`
}

// ESLogEntry 单条日志/文档的精简结果
type ESLogEntry struct {
	ID      string                 `json:"id"`
	Source  map[string]interface{} `json:"source"`
}

// ESQueryResult Elasticsearch 查询返回给 LLM 的结构
type ESQueryResult struct {
	DataSource string       `json:"data_source"`
	Index      string       `json:"index"`
	Total      int          `json:"total"`
	Hits       []*ESLogEntry `json:"hits"`
}

// PyroscopeFunction 火焰图中某个函数的聚合结果
type PyroscopeFunction struct {
	Name       string  `json:"name"`
	Self       int     `json:"self"`
	Total      int     `json:"total"`
	SelfPercent float64 `json:"self_percent"`
}

// PyroscopeQueryResult Pyroscope 查询返回给 LLM 的结构
type PyroscopeQueryResult struct {
	DataSource string               `json:"data_source"`
	Query      string               `json:"query"`
	ProfileType string              `json:"profile_type,omitempty"`
	Units      string               `json:"units,omitempty"`
	SampleRate uint32               `json:"sample_rate,omitempty"`
	NumTicks   int                  `json:"num_ticks"`
	TopFunctions []*PyroscopeFunction `json:"top_functions"`
	Timeline   *PyroscopeTimeline   `json:"timeline,omitempty"`
}

// PyroscopeTimeline 性能剖析时间线
type PyroscopeTimeline struct {
	StartTime     int64    `json:"start_time"`
	DurationDelta int64    `json:"duration_delta"`
	Samples       []uint64 `json:"samples"`
}

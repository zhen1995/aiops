package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"aiops/internal/anomaly"
	"aiops/internal/datasource"
	"aiops/internal/knowledge"
	"aiops/internal/wiki"
	"aiops/models"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// Registry 维护一组可调用工具，内部依赖 datasource.Manager 获取配置
type Registry struct {
	db            *gorm.DB
	manager       *datasource.Manager
	kb            *knowledge.Service
	wiki          *wiki.Service
	pythonBaseURL string // Python 算法服务地址（异常检测技能）
}

// NewRegistry 创建工具注册表
func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{db: db, manager: datasource.NewManager(db)}
}

// SetKnowledge 注入知识库编排服务（须在首次构建 Agent 前调用）
func (r *Registry) SetKnowledge(svc *knowledge.Service) {
	r.kb = svc
}

// SetWiki 注入知识库 Wiki 服务（须在首次构建 Agent 前调用）
func (r *Registry) SetWiki(svc *wiki.Service) {
	r.wiki = svc
}

// SetPythonBaseURL 注入 Python 算法服务地址（异常检测技能依赖；须在首次构建 Agent 前调用）
func (r *Registry) SetPythonBaseURL(baseURL string) {
	r.pythonBaseURL = baseURL
}

// Tools 返回所有工具实例及其 ToolInfo，用于绑定到 ChatModel
func (r *Registry) Tools(ctx context.Context) ([]tool.InvokableTool, []*schema.ToolInfo, error) {
	tm, infos, err := r.ToolMap(ctx)
	if err != nil {
		return nil, nil, err
	}
	tools := make([]tool.InvokableTool, 0, len(tm))
	for _, t := range tm {
		tools = append(tools, t)
	}
	return tools, infos, nil
}

// ToolMap 返回工具名到工具实例的映射，以及所有 ToolInfo，便于按名调用
func (r *Registry) ToolMap(ctx context.Context) (map[string]tool.InvokableTool, []*schema.ToolInfo, error) {
	builders := []func() (tool.InvokableTool, error){
		r.listDataSourcesTool,
		r.listServicesTool,
		r.getServiceInfoTool,
		r.queryPrometheusTool,
		r.queryElasticsearchTool,
		r.queryPyroscopeTool,
		r.searchKnowledgeBaseTool,
		r.searchWikiTool,
		r.detectAnomaliesTool,
	}

	tm := make(map[string]tool.InvokableTool, len(builders))
	infos := make([]*schema.ToolInfo, 0, len(builders))
	for _, b := range builders {
		t, err := b()
		if err != nil {
			return nil, nil, err
		}
		info, err := t.Info(ctx)
		if err != nil {
			return nil, nil, err
		}
		tm[info.Name] = t
		infos = append(infos, info)
	}
	return tm, infos, nil
}

// MustInfos 是 Tools 的便捷版本，失败时 panic，适合在 init 阶段使用
func (r *Registry) MustInfos(ctx context.Context) []*schema.ToolInfo {
	_, infos, err := r.Tools(ctx)
	if err != nil {
		panic(fmt.Sprintf("初始化工具失败: %v", err))
	}
	return infos
}

// Manager 暴露底层数据源管理器，供 Agent 构建 system prompt 使用
func (r *Registry) Manager() *datasource.Manager {
	return r.manager
}

// ---------- list_data_sources ----------

type ListDataSourcesInput struct{}

type DataSourceInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ListDataSourcesOutput struct {
	Sources []DataSourceInfo `json:"sources"`
}

func (r *Registry) listDataSourcesTool() (tool.InvokableTool, error) {
	return utils.InferTool[ListDataSourcesInput, ListDataSourcesOutput](
		"list_data_sources",
		"列出当前系统中所有已启用、可用于查询的数据源名称和类型。当用户的问题涉及具体数据源但没说名字时，优先调用此工具。",
		func(ctx context.Context, in ListDataSourcesInput) (ListDataSourcesOutput, error) {
			list, err := r.manager.LoadEnabled(ctx)
			if err != nil {
				return ListDataSourcesOutput{}, err
			}
			out := ListDataSourcesOutput{Sources: make([]DataSourceInfo, 0, len(list))}
			for _, ds := range list {
				out.Sources = append(out.Sources, DataSourceInfo{Name: ds.Name, Type: ds.Type})
			}
			return out, nil
		},
	)
}

// ---------- list_services ----------

type ListServicesInput struct{}

// ServiceInfo 单个启用服务的数据源定位信息
type ServiceInfo struct {
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	ESDatasource    string            `json:"es_datasource"`
	ESIndexPatterns []string          `json:"es_index_patterns"`
	PromDatasource  string            `json:"prom_datasource"`
	PromLabels      map[string]string `json:"prom_labels"`
}

type ListServicesOutput struct {
	Services []ServiceInfo `json:"services"`
}

func (r *Registry) listServicesTool() (tool.InvokableTool, error) {
	return utils.InferTool[ListServicesInput, ListServicesOutput](
		"list_services",
		"列出所有已启用的服务注册信息（code、名称、ES 数据源名称与索引模式、Prometheus 数据源名称与标签选择器）。查询指定服务的日志或指标前，先用本工具获取该服务对应的 ES 数据源名称、索引模式和 Prometheus 标签选择器，避免猜错索引或数据源名。",
		func(ctx context.Context, in ListServicesInput) (ListServicesOutput, error) {
			var list []models.Service
			if err := r.db.WithContext(ctx).
				Where("status = ?", 1).
				Order("name ASC").
				Find(&list).Error; err != nil {
				return ListServicesOutput{}, err
			}

			// 收集引用的数据源 ID，批量查出名称（供 query_elasticsearch 等工具的 data_source_name 直接使用）
			dsIDs := map[string]bool{}
			for _, svc := range list {
				if svc.ESDatasourceID != "" {
					dsIDs[svc.ESDatasourceID] = true
				}
				if svc.PromDatasourceID != "" {
					dsIDs[svc.PromDatasourceID] = true
				}
			}
			dsNames := map[string]string{}
			if len(dsIDs) > 0 {
				ids := make([]string, 0, len(dsIDs))
				for id := range dsIDs {
					ids = append(ids, id)
				}
				var dsList []models.Datasource
				if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&dsList).Error; err != nil {
					return ListServicesOutput{}, err
				}
				for _, ds := range dsList {
					dsNames[ds.ID] = ds.Name
				}
			}

			out := ListServicesOutput{Services: make([]ServiceInfo, 0, len(list))}
			for _, svc := range list {
				info := ServiceInfo{
					Code:            svc.Code,
					Name:            svc.Name,
					ESDatasource:    dsNames[svc.ESDatasourceID],
					PromDatasource:  dsNames[svc.PromDatasourceID],
					ESIndexPatterns: []string{},
					PromLabels:      map[string]string{},
				}
				if svc.ESIndexPatterns != "" {
					var patterns []string
					if err := json.Unmarshal([]byte(svc.ESIndexPatterns), &patterns); err == nil {
						info.ESIndexPatterns = patterns
					}
				}
				if svc.PromLabels != "" {
					var labels map[string]string
					if err := json.Unmarshal([]byte(svc.PromLabels), &labels); err == nil {
						info.PromLabels = labels
					}
				}
				out.Services = append(out.Services, info)
			}
			return out, nil
		},
	)
}

// ---------- get_service_info ----------

type GetServiceInfoInput struct {
	Keyword string `json:"keyword" jsonschema:"required,description=服务编码或名称，支持模糊匹配，例如 order-service 或 订单"`
}

// ServiceDetail 服务注册详情（含数据源名称与元信息）
type ServiceDetail struct {
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	ESDatasource    string            `json:"es_datasource"`
	ESIndexPatterns []string          `json:"es_index_patterns"`
	PromDatasource  string            `json:"prom_datasource"`
	PromLabels      map[string]string `json:"prom_labels"`
	PyroscopeApp    string            `json:"pyroscope_app"`
	Owner           string            `json:"owner"`
	Description     string            `json:"description"`
	Status          int               `json:"status"`
	Verified        int               `json:"verified"`
	VerifyMessage   string            `json:"verify_message"`
}

type GetServiceInfoOutput struct {
	Services []ServiceDetail `json:"services"`
}

func (r *Registry) getServiceInfoTool() (tool.InvokableTool, error) {
	return utils.InferTool[GetServiceInfoInput, GetServiceInfoOutput](
		"get_service_info",
		"按服务编码或名称查询服务注册的完整详情（ES 索引模式、Prometheus 标签选择器、Pyroscope 应用名、负责人、探测状态等），支持模糊匹配。已知服务名称、需要它的数据定位信息或负责人等元信息时调用本工具。",
		func(ctx context.Context, in GetServiceInfoInput) (GetServiceInfoOutput, error) {
			kw := strings.TrimSpace(in.Keyword)
			var list []models.Service
			// 先精确匹配编码，未命中再按编码/名称模糊匹配
			if err := r.db.WithContext(ctx).
				Where("code = ?", kw).
				Limit(20).Find(&list).Error; err != nil {
				return GetServiceInfoOutput{}, err
			}
			if len(list) == 0 {
				if err := r.db.WithContext(ctx).
					Where("code LIKE ? OR name LIKE ?", "%"+kw+"%", "%"+kw+"%").
					Order("name ASC").
					Limit(20).Find(&list).Error; err != nil {
					return GetServiceInfoOutput{}, err
				}
			}

			// 收集引用的数据源 ID，批量查出名称
			dsIDs := map[string]bool{}
			for _, svc := range list {
				if svc.ESDatasourceID != "" {
					dsIDs[svc.ESDatasourceID] = true
				}
				if svc.PromDatasourceID != "" {
					dsIDs[svc.PromDatasourceID] = true
				}
			}
			dsNames := map[string]string{}
			if len(dsIDs) > 0 {
				ids := make([]string, 0, len(dsIDs))
				for id := range dsIDs {
					ids = append(ids, id)
				}
				var dsList []models.Datasource
				if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&dsList).Error; err != nil {
					return GetServiceInfoOutput{}, err
				}
				for _, ds := range dsList {
					dsNames[ds.ID] = ds.Name
				}
			}

			out := GetServiceInfoOutput{Services: make([]ServiceDetail, 0, len(list))}
			for _, svc := range list {
				detail := ServiceDetail{
					Code:            svc.Code,
					Name:            svc.Name,
					ESDatasource:    dsNames[svc.ESDatasourceID],
					PromDatasource:  dsNames[svc.PromDatasourceID],
					PyroscopeApp:    svc.PyroscopeApp,
					Owner:           svc.Owner,
					Description:     svc.Description,
					Status:          svc.Status,
					Verified:        svc.Verified,
					VerifyMessage:   svc.VerifyMessage,
					ESIndexPatterns: []string{},
					PromLabels:      map[string]string{},
				}
				if svc.ESIndexPatterns != "" {
					var patterns []string
					if err := json.Unmarshal([]byte(svc.ESIndexPatterns), &patterns); err == nil {
						detail.ESIndexPatterns = patterns
					}
				}
				if svc.PromLabels != "" {
					var labels map[string]string
					if err := json.Unmarshal([]byte(svc.PromLabels), &labels); err == nil {
						detail.PromLabels = labels
					}
				}
				out.Services = append(out.Services, detail)
			}
			return out, nil
		},
	)
}

// ---------- query_prometheus ----------

type QueryPrometheusInput struct {
	DataSourceName string `json:"data_source_name" jsonschema:"required,description=要查询的 Prometheus 数据源名称，例如 指标平台"`
	Query          string `json:"query" jsonschema:"required,description=PromQL 查询表达式，例如 topk(5, rate(container_cpu_usage_seconds_total[5m]))"`
	Start          string `json:"start,omitempty" jsonschema:"description=开始时间，支持 1h、30m、ISO 等，默认 1 小时前"`
	End            string `json:"end,omitempty" jsonschema:"description=结束时间，支持 now、ISO 等，默认现在"`
	Step           string `json:"step,omitempty" jsonschema:"description=采样步长，如 15s、1m，默认自动计算"`
	TopN           int    `json:"topn,omitempty" jsonschema:"description=返回的序列数量上限，默认 10"`
}

func (r *Registry) queryPrometheusTool() (tool.InvokableTool, error) {
	return utils.InferTool[QueryPrometheusInput, datasource.PrometheusQueryResult](
		"query_prometheus",
		"在指定的 Prometheus 数据源上执行 PromQL 查询，返回聚合后的时序数据（含序列名称、统计摘要、采样点）。",
		func(ctx context.Context, in QueryPrometheusInput) (datasource.PrometheusQueryResult, error) {
			ds, err := r.manager.FindEnabled(ctx, in.DataSourceName)
			if err != nil {
				return datasource.PrometheusQueryResult{}, err
			}
			if ds.Type != datasource.TypePrometheus {
				return datasource.PrometheusQueryResult{}, fmt.Errorf("数据源 %s 不是 Prometheus 类型", in.DataSourceName)
			}
			if in.Start == "" {
				in.Start = "1h"
			}
			if in.End == "" {
				in.End = "now"
			}
			client := datasource.NewPrometheusClient(ds)
			res, err := client.QueryRange(ctx, in.Query, in.Start, in.End, in.Step, in.TopN, 0)
			if err != nil {
				return datasource.PrometheusQueryResult{}, err
			}
			return *res, nil
		},
	)
}

// ---------- query_elasticsearch ----------

type QueryElasticsearchInput struct {
	DataSourceName string `json:"data_source_name" jsonschema:"required,description=要查询的 Elasticsearch 数据源名称，例如 日志平台"`
	IndexPattern   string `json:"index_pattern" jsonschema:"required,description=索引模式，例如 nginx-*、logstash-*"`
	QueryString    string `json:"query_string,omitempty" jsonschema:"description=Lucene 查询字符串，例如 level:ERROR AND service:order"`
	Start          string `json:"start,omitempty" jsonschema:"description=开始时间，支持 1h、30m、ISO 等，默认 1 小时前"`
	End            string `json:"end,omitempty" jsonschema:"description=结束时间，支持 now、ISO 等，默认现在"`
	Limit          int    `json:"limit,omitempty" jsonschema:"description=返回日志条数上限，默认 20"`
}

func (r *Registry) queryElasticsearchTool() (tool.InvokableTool, error) {
	return utils.InferTool[QueryElasticsearchInput, datasource.ESQueryResult](
		"query_elasticsearch",
		"在指定的 Elasticsearch 数据源上检索日志/文档，返回按时间倒排、字段精简后的命中列表。",
		func(ctx context.Context, in QueryElasticsearchInput) (datasource.ESQueryResult, error) {
			ds, err := r.manager.FindEnabled(ctx, in.DataSourceName)
			if err != nil {
				return datasource.ESQueryResult{}, err
			}
			if ds.Type != datasource.TypeElasticsearch {
				return datasource.ESQueryResult{}, fmt.Errorf("数据源 %s 不是 ElasticSearch 类型", in.DataSourceName)
			}
			if in.Start == "" {
				in.Start = "1h"
			}
			if in.End == "" {
				in.End = "now"
			}
			client := datasource.NewElasticsearchClient(ds)
			res, err := client.Search(ctx, in.IndexPattern, in.QueryString, in.Start, in.End, in.Limit)
			if err != nil {
				return datasource.ESQueryResult{}, err
			}
			return *res, nil
		},
	)
}

// ---------- query_pyroscope ----------

type QueryPyroscopeInput struct {
	DataSourceName string `json:"data_source_name" jsonschema:"required,description=要查询的 Pyroscope 数据源名称"`
	Query          string `json:"query" jsonschema:"required,description=Pyroscope 查询表达式，例如 process_cpu:cpu:nanoseconds:cpu:nanoseconds{service_name=\"my-app\"}"`
	Start          string `json:"start,omitempty" jsonschema:"description=开始时间，支持 1h、30m、ISO 等，默认 1 小时前"`
	End            string `json:"end,omitempty" jsonschema:"description=结束时间，支持 now、ISO 等，默认现在"`
	MaxNodes       int    `json:"max_nodes,omitempty" jsonschema:"description=火焰图最大节点数，默认 1024"`
}

func (r *Registry) queryPyroscopeTool() (tool.InvokableTool, error) {
	return utils.InferTool[QueryPyroscopeInput, datasource.PyroscopeQueryResult](
		"query_pyroscope",
		"在指定的 Pyroscope 数据源上查询性能剖析火焰图，返回按 Self 采样数排序的 Top 函数及时间线摘要。",
		func(ctx context.Context, in QueryPyroscopeInput) (datasource.PyroscopeQueryResult, error) {
			ds, err := r.manager.FindEnabled(ctx, in.DataSourceName)
			if err != nil {
				return datasource.PyroscopeQueryResult{}, err
			}
			if ds.Type != datasource.TypePyroscope {
				return datasource.PyroscopeQueryResult{}, fmt.Errorf("数据源 %s 不是 Pyroscope 类型", in.DataSourceName)
			}
			if in.Start == "" {
				in.Start = "1h"
			}
			if in.End == "" {
				in.End = "now"
			}
			client := datasource.NewPyroscopeClient(ds)
			res, err := client.Render(ctx, in.Query, in.Start, in.End, in.MaxNodes)
			if err != nil {
				return datasource.PyroscopeQueryResult{}, err
			}
			return *res, nil
		},
	)
}

// ---------- search_knowledge_base ----------

// SearchKnowledgeBaseInput search_knowledge_base 工具入参
type SearchKnowledgeBaseInput struct {
	Query string `json:"query" jsonschema:"required,description=检索问题或关键词，例如：MySQL 连接超时如何处理"`
	TopK  int    `json:"top_k" jsonschema:"description=返回的片段数量，默认 5，最大 10"`
}

// SearchKnowledgeBaseOutput 单个命中片段
type SearchKnowledgeBaseOutput struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

func (r *Registry) searchKnowledgeBaseTool() (tool.InvokableTool, error) {
	return utils.InferTool[SearchKnowledgeBaseInput, []*SearchKnowledgeBaseOutput](
		"search_knowledge_base",
		"在运维知识库中做语义检索，返回与问题最相关的文档片段（含来源文档与相似度）。当用户询问运维经验、故障处理手册、排查步骤、平台使用说明等知识性问题时调用。",
		func(ctx context.Context, in SearchKnowledgeBaseInput) ([]*SearchKnowledgeBaseOutput, error) {
			if r.kb == nil {
				return nil, fmt.Errorf("知识库服务未初始化")
			}
			topK := in.TopK
			if topK <= 0 || topK > 10 {
				topK = 5
			}
			hits, err := r.kb.Retrieve(ctx, in.Query, topK)
			if err != nil {
				return nil, err
			}
			out := make([]*SearchKnowledgeBaseOutput, 0, len(hits))
			for _, h := range hits {
				out = append(out, &SearchKnowledgeBaseOutput{
					ChunkID: h.ChunkID, DocumentID: h.DocumentID, Title: h.Title,
					Content: h.Content, Score: h.Score,
				})
			}
			return out, nil
		},
	)
}

// ---------- search_wiki ----------

// SearchWikiInput search_wiki 工具入参
type SearchWikiInput struct {
	Query string `json:"query" jsonschema:"required,description=检索关键词，例如：MySQL 主从延迟处理"`
	Limit int    `json:"limit" jsonschema:"description=返回词条数量，默认 5，最大 10"`
}

// SearchWikiOutput 单个命中词条
type SearchWikiOutput struct {
	EntryID      string `json:"entry_id"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Content      string `json:"content"`
	CategoryName string `json:"category_name"`
	SourceName   string `json:"source_name"`
}

func (r *Registry) searchWikiTool() (tool.InvokableTool, error) {
	return utils.InferTool[SearchWikiInput, []*SearchWikiOutput](
		"search_wiki",
		"检索知识库 Wiki 中的结构化词条（已发布的权威知识页，含概念说明、操作步骤、故障处理等）。当用户询问运维知识、操作方法、故障处理流程、平台使用说明等结构性问题时优先调用本工具，其次才是 search_knowledge_base 的文档片段检索。",
		func(ctx context.Context, in SearchWikiInput) ([]*SearchWikiOutput, error) {
			if r.wiki == nil {
				return nil, fmt.Errorf("Wiki 服务未初始化")
			}
			limit := in.Limit
			if limit <= 0 || limit > 10 {
				limit = 5
			}
			hits, err := r.wiki.SearchEntries(in.Query, limit, false)
			if err != nil {
				return nil, err
			}
			out := make([]*SearchWikiOutput, 0, len(hits))
			for _, h := range hits {
				out = append(out, &SearchWikiOutput{
					EntryID: h.ID, Title: h.Title, Summary: h.Summary,
					Content: h.Content, CategoryName: h.CategoryName, SourceName: h.SourceName,
				})
			}
			return out, nil
		},
	)
}

// ---------- detect_anomalies ----------

// DetectAnomaliesInput detect_anomalies 工具入参
type DetectAnomaliesInput struct {
	DataSourceName string `json:"data_source_name" jsonschema:"required,description=要查询的 Prometheus 数据源名称"`
	Query          string `json:"query" jsonschema:"required,description=PromQL 查询表达式，例如 sum by (instance) (rate(http_requests_total[5m]))"`
	Start          string `json:"start,omitempty" jsonschema:"description=开始时间，支持 6h、1d、ISO 等，默认 6 小时前"`
	End            string `json:"end,omitempty" jsonschema:"description=结束时间，支持 now、ISO 等，默认现在"`
	CompareOffset  string `json:"compare_offset,omitempty" jsonschema:"description=历史同时段对比偏移，如 7d、24h、168h；设置后自动多拉一个偏移前的同 query 序列做周期对比，输出 periodic_deviation 信号"`
}

// DetectAnomaliesOutput 异常检测信号摘要（不含原始点）
type DetectAnomaliesOutput struct {
	DataSource string           `json:"data_source"`
	Query      string           `json:"query"`
	Summary    string           `json:"summary"`
	Signals    []anomaly.Signal `json:"signals"`
}

// detectAnomaliesTool 服务端闭环：拉取原始时序（程序侧）→ Python 确定性检测 → 只把异常信号摘要返回模型
func (r *Registry) detectAnomaliesTool() (tool.InvokableTool, error) {
	return utils.InferTool[DetectAnomaliesInput, *DetectAnomaliesOutput](
		"detect_anomalies",
		"对 Prometheus 指标做确定性异常检测（基线漂移、3-sigma/IQR 离群、突变点），返回异常信号摘要而非原始数据。适合长时间窗、多序列的排障分析；简单看值请用 query_prometheus。",
		func(ctx context.Context, in DetectAnomaliesInput) (*DetectAnomaliesOutput, error) {
			ds, err := r.manager.FindEnabled(ctx, in.DataSourceName)
			if err != nil {
				return nil, err
			}
			if ds.Type != datasource.TypePrometheus {
				return nil, fmt.Errorf("数据源 %s 不是 Prometheus 类型", in.DataSourceName)
			}
			if in.Start == "" {
				in.Start = "6h"
			}
			if in.End == "" {
				in.End = "now"
			}
			// 原始时序仅在程序侧流转，不进模型上下文
			res, err := datasource.NewPrometheusClient(ds).QueryRange(ctx, in.Query, in.Start, in.End, "", 0, 0)
			if err != nil {
				return nil, err
			}

			// 历史同时段对比：设置 compare_offset 时，服务端自动多拉一个偏移前的同 query 序列
			var history map[string][][2]interface{}
			if offset := strings.TrimSpace(in.CompareOffset); offset != "" {
				d, err := parseOffset(offset)
				if err != nil {
					return nil, err
				}
				histRes, err := queryShifted(ctx, ds, in.Query, in.Start, in.End, d)
				if err != nil {
					return nil, fmt.Errorf("历史同时段查询失败: %w", err)
				}
				history = make(map[string][][2]interface{}, len(histRes.Series))
				for _, s := range histRes.Series {
					history[seriesKey(s.Labels)] = s.Values
				}
			}

			series := make([]anomaly.SeriesInput, 0, len(res.Series))
			for _, s := range res.Series {
				item := anomaly.SeriesInput{Metric: s.Name, Labels: s.Labels, Points: s.Values}
				if history != nil {
					item.HistoryPoints = history[seriesKey(s.Labels)]
				}
				series = append(series, anomaly.SeriesInput(item))
			}
			detectResp, err := anomaly.NewClient(r.pythonBaseURL).Detect(ctx, &anomaly.DetectRequest{Series: series})
			if err != nil {
				return nil, err
			}
			return &DetectAnomaliesOutput{
				DataSource: in.DataSourceName,
				Query:      in.Query,
				Summary:    detectResp.Summary,
				Signals:    detectResp.Signals,
			}, nil
		},
	)
}

// parseOffset 解析偏移量，支持 30m / 12h / 7d
func parseOffset(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("偏移量不能为空")
	}
	num, unit := s[:len(s)-1], s[len(s)-1]
	var mul time.Duration
	switch unit {
	case 'm':
		mul = time.Minute
	case 'h':
		mul = time.Hour
	case 'd':
		mul = 24 * time.Hour
	default:
		return 0, fmt.Errorf("非法偏移量 %q（支持 30m/12h/7d）", s)
	}
	v, err := strconv.Atoi(num)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("非法偏移量 %q（支持 30m/12h/7d）", s)
	}
	return time.Duration(v) * mul, nil
}

// seriesKey 序列标签的稳定键（用于当前序列与历史序列按标签对齐）
func seriesKey(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, "%s=%s", k, labels[k])
	}
	b.WriteString("}")
	return b.String()
}

// queryShifted 查询 [start-offset, end-offset] 的同 query 序列（历史同时段）
func queryShifted(ctx context.Context, ds *models.Datasource, query, start, end string, offset time.Duration) (*datasource.PrometheusQueryResult, error) {
	now := time.Now()
	startT, err := datasource.ParseTime(start, now)
	if err != nil {
		return nil, err
	}
	endT, err := datasource.ParseTime(end, now)
	if err != nil {
		return nil, err
	}
	return datasource.NewPrometheusClient(ds).QueryRange(ctx, query,
		startT.Add(-offset).Format(time.RFC3339), endT.Add(-offset).Format(time.RFC3339), "", 0, 0)
}

// MarshalResult 将工具结果序列化为 JSON 字符串，并做大小兜底
func MarshalResult(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if len(b) > datasource.MaxResultBytes {
		return "", fmt.Errorf("工具结果超过 %d 字节，请缩小查询范围", datasource.MaxResultBytes)
	}
	return string(b), nil
}

package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"aiops/internal/datasource"
	"aiops/internal/knowledge"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// Registry 维护一组可调用工具，内部依赖 datasource.Manager 获取配置
type Registry struct {
	db      *gorm.DB
	manager *datasource.Manager
	kb      *knowledge.Service
}

// NewRegistry 创建工具注册表
func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{db: db, manager: datasource.NewManager(db)}
}

// SetKnowledge 注入知识库编排服务（须在首次构建 Agent 前调用）
func (r *Registry) SetKnowledge(svc *knowledge.Service) {
	r.kb = svc
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
		r.queryPrometheusTool,
		r.queryElasticsearchTool,
		r.queryPyroscopeTool,
		r.searchKnowledgeBaseTool,
	}
	builders = append(builders, r.n9eTools()...)

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

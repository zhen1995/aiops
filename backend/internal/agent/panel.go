// 多智能体专家会诊编排：Planner 拆题 → 专科专家并发分析 → 主持人汇总。
// 专家执行复用本包的单循环 Agent；每个专家独立 ChatModel（BindTools 有状态，不可并发共享）。
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// maxSubTasks 单次会诊最多拆出的子任务数（控制 LLM 调用成本）
const maxSubTasks = 4

// Specialist 专科专家定义
type Specialist struct {
	Key          string   // 标识：metrics/logs/profiling/service/knowledge
	Title        string   // 展示名：指标专家/日志专家/...
	Instructions string   // 角色 system prompt
	ToolFilter   []string // 该专家可见的工具（空=全部）
}

// serviceTools 服务注册类工具的公共集合
var serviceTools = []string{"list_services", "get_service_info", "list_data_sources"}

// Specialists 返回内置的专科专家名册
func Specialists() []Specialist {
	return []Specialist{
		{
			Key:   "metrics",
			Title: "指标专家",
			Instructions: "你是运维指标分析专家，擅长 Prometheus 监控指标解读、时序趋势分析、容量与性能评估。" +
				"只基于 query_prometheus 等工具返回的真实数据下结论，数据不足时明确说明，禁止编造指标值。",
			ToolFilter: append([]string{"query_prometheus"}, serviceTools...),
		},
		{
			Key:   "logs",
			Title: "日志专家",
			Instructions: "你是日志分析专家，擅长 Elasticsearch 日志检索、错误模式聚类、调用链线索提取。" +
				"只基于 query_elasticsearch 等工具返回的真实日志下结论，数据不足时明确说明，禁止编造日志内容。" +
				"查询日志前先用 list_services 获取目标服务的 ES 数据源名称与索引模式，不要猜数据源名或索引名；" +
				"用 query_elasticsearch 组合查询条件一次取足数据（limit 10~20），最多查询 2 次后必须给出结论。",
			ToolFilter: append([]string{"query_elasticsearch"}, serviceTools...),
		},
		{
			Key:   "profiling",
			Title: "剖析专家",
			Instructions: "你是性能剖析专家，擅长 Pyroscope 火焰图解读、CPU/内存热点定位、性能瓶颈分析。" +
				"只基于 query_pyroscope 等工具返回的真实剖析数据下结论，数据不足时明确说明，禁止编造。",
			ToolFilter: append([]string{"query_pyroscope"}, serviceTools...),
		},
		{
			Key:   "service",
			Title: "服务注册专家",
			Instructions: "你是服务注册与元信息专家，熟悉系统内服务的索引模式、Prometheus 标签选择器、负责人、连通性状态等注册信息。" +
				"只基于服务注册工具返回的真实信息回答。",
			ToolFilter: serviceTools,
		},
		{
			Key:   "knowledge",
			Title: "知识库专家",
			Instructions: "你是运维知识库专家，擅长通过语义检索定位故障处理手册、排查步骤与平台使用说明。" +
				"回答需引用 search_knowledge_base 返回的文档片段内容，并注明来源文档标题。",
			ToolFilter: []string{"search_knowledge_base"},
		},
	}
}

// SubTask  Planner 拆分出的子任务：指定由哪个专家回答什么问题
type SubTask struct {
	Specialist string `json:"specialist"`
	Question   string `json:"question"`
}

// PanelCallbacks 会诊过程回调（供 SSE 控制器转发给前端）
type PanelCallbacks struct {
	OnPanelStart   func(subtasks []SubTask)                // 会诊开始（专家名单确定）
	OnSpecialist   func(key, title, status, conclusion string) // status: running/completed/failed
	OnToolCall     func(expert, name, args string)         // 专家的工具调用
	OnToolResult   func(expert, name, result string)       // 专家的工具结果
	OnChunk        func(chunk string)                      // 主持人的最终答案片段
}

// ChatModelFactory 创建一个新的 ChatModel（每次调用必须返回独立实例）
type ChatModelFactory func(ctx context.Context) (model.ChatModel, error)

// Panel 专家会诊编排器
type Panel struct {
	registry *Registry
}

// NewPanel 创建会诊编排器
func NewPanel(registry *Registry) *Panel {
	return &Panel{registry: registry}
}

// specialistByKey 按 key 找专家定义
func specialistByKey(key string) (Specialist, bool) {
	for _, s := range Specialists() {
		if s.Key == key {
			return s, true
		}
	}
	return Specialist{}, false
}

// SpecialistInfo 按 key 导出专家定义（供控制器组装 SSE 事件）
func SpecialistInfo(key string) (Specialist, bool) {
	return specialistByKey(key)
}

// Plan 让 Planner 判断问题是否需要专家会诊并拆分为子任务。
// forcePanel 为 true 表示用户明确要求多 Agent 会诊：只要问题可拆分就必须拆（拆不了才回退单 Agent）。
// 返回 nil 表示不需要会诊（调用方应回退到单 Agent 路径）；返回非空表示需要会诊。
// 规划失败（LLM 调用失败/JSON 解析失败/非法子任务）一律视为不需要会诊，保证可用性优先。
func Plan(ctx context.Context, factory ChatModelFactory, question string, forcePanel bool) []SubTask {
	cm, err := factory(ctx)
	if err != nil {
		return nil
	}

	var b strings.Builder
	b.WriteString("你是运维智能会诊的调度员。请判断用户问题是否需要多个专科专家会诊（涉及两类及以上数据源/领域的交叉分析），并拆分为子任务。\n")
	b.WriteString("可用专家：")
	for _, s := range Specialists() {
		fmt.Fprintf(&b, "%s(%s) ", s.Key, s.Title)
	}
	b.WriteString("\n输出严格 JSON（不要输出其他任何内容）：{\"need_panel\":true,\"subtasks\":[{\"specialist\":\"metrics\",\"question\":\"子问题\"}]}\n")
	if forcePanel {
		fmt.Fprintf(&b, "规则：用户已明确要求使用多专家会诊——只要问题涉及运维数据、可多角度分析或可跨领域拆分，need_panel 必须为 true 并拆分为 2~%d 个子任务；只有无法拆分的单一事实问答才允许 need_panel=false（此时 subtasks 为空数组）。specialist 只能取上述 key。\n", maxSubTasks)
	} else {
		fmt.Fprintf(&b, "规则：need_panel 为 false 时 subtasks 为空数组；子任务最多 %d 个；specialist 只能取上述 key；简单问答、单一领域问题、纯知识性问题一律 need_panel=false。\n", maxSubTasks)
	}

	messages := []*schema.Message{
		schema.SystemMessage(b.String()),
		schema.UserMessage(question),
	}
	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		fmt.Printf("[专家会诊] Planner 调用失败，回退单 Agent: %v\n", err)
		return nil
	}

	subtasks := parsePlan(resp.Content)
	if len(subtasks) == 0 {
		return nil
	}
	return subtasks
}

// parsePlan 从 Planner 输出中解析子任务列表（容忍模型输出多余文本）
func parsePlan(content string) []SubTask {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return nil
	}
	var plan struct {
		NeedPanel bool      `json:"need_panel"`
		SubTasks  []SubTask `json:"subtasks"`
	}
	if err := json.Unmarshal([]byte(content[start:end+1]), &plan); err != nil || !plan.NeedPanel {
		return nil
	}
	valid := make([]SubTask, 0, len(plan.SubTasks))
	seen := map[string]bool{}
	for _, st := range plan.SubTasks {
		if _, ok := specialistByKey(st.Specialist); !ok {
			continue
		}
		st.Question = strings.TrimSpace(st.Question)
		if st.Question == "" || seen[st.Specialist] {
			continue
		}
		seen[st.Specialist] = true
		valid = append(valid, st)
		if len(valid) >= maxSubTasks {
			break
		}
	}
	return valid
}

// Run 执行会诊：并发执行各专家分析，最后由主持人汇总输出最终答案。
// 单个专家失败不中断整体（结论记为失败原因）；全部失败才返回 error。
func (p *Panel) Run(ctx context.Context, factory ChatModelFactory, question string, subtasks []SubTask, cb *PanelCallbacks) (*schema.Message, error) {
	if cb != nil && cb.OnPanelStart != nil {
		cb.OnPanelStart(subtasks)
	}

	// 并发执行各专家
	conclusions := make([]string, len(subtasks))
	var wg sync.WaitGroup
	for i, st := range subtasks {
		wg.Add(1)
		go func(idx int, st SubTask) {
			defer wg.Done()
			conclusions[idx] = p.runSpecialist(ctx, factory, st, cb)
		}(i, st)
	}
	wg.Wait()

	// 统计失败数，全部失败则整体失败
	failed := 0
	for _, c := range conclusions {
		if strings.HasPrefix(c, "【分析失败】") {
			failed++
		}
	}
	if failed == len(subtasks) {
		return nil, fmt.Errorf("所有专家分析均失败")
	}

	// 主持人汇总
	return p.runChair(ctx, factory, question, subtasks, conclusions, cb)
}

// runSpecialist 执行单个专家分析，返回结论文本（失败时返回"【分析失败】原因"）
func (p *Panel) runSpecialist(ctx context.Context, factory ChatModelFactory, st SubTask, cb *PanelCallbacks) string {
	sp, ok := specialistByKey(st.Specialist)
	if !ok {
		return "【分析失败】未知专家: " + st.Specialist
	}

	if cb != nil && cb.OnSpecialist != nil {
		cb.OnSpecialist(sp.Key, sp.Title, "running", "")
	}

	cm, err := factory(ctx)
	if err != nil {
		if cb != nil && cb.OnSpecialist != nil {
			cb.OnSpecialist(sp.Key, sp.Title, "failed", "")
		}
		return "【分析失败】初始化大模型失败: " + err.Error()
	}

	ag := New(cm, p.registry, Options{
		Instructions:      sp.Instructions,
		EnforceDataSource: true,
		ToolFilter:        sp.ToolFilter,
		MaxIterations:     8,
	})

	var innerErr error
	var callbacks *Callbacks
	if cb != nil {
		callbacks = &Callbacks{
			OnStatus: func(status string) {},
			OnToolCall: func(name, args string) {
				if cb.OnToolCall != nil {
					cb.OnToolCall(sp.Key, name, args)
				}
			},
			OnToolResult: func(name, result string) {
				if cb.OnToolResult != nil {
					cb.OnToolResult(sp.Key, name, result)
				}
			},
		}
	}

	messages := []*schema.Message{
		schema.SystemMessage(sp.Instructions),
		schema.UserMessage(st.Question + "\n\n请给出分析结论（200 字以内），只基于工具返回的真实数据。" +
			"工具使用要高效：先用 list_services / get_service_info 定位数据源与索引，再用领域查询工具一次取足数据（返回条数 limit 10~20），最多 2 次数据查询后必须给出结论，不要反复试探。"),
	}
	resp, innerErr := ag.Run(ctx, messages, callbacks)
	if innerErr != nil {
		if cb != nil && cb.OnSpecialist != nil {
			cb.OnSpecialist(sp.Key, sp.Title, "failed", "")
		}
		return "【分析失败】" + innerErr.Error()
	}

	conclusion := strings.TrimSpace(resp.Content)
	if conclusion == "" {
		conclusion = "（该专家未给出结论）"
	}
	if cb != nil && cb.OnSpecialist != nil {
		cb.OnSpecialist(sp.Key, sp.Title, "completed", conclusion)
	}
	return conclusion
}

// runChair 主持人：交叉验证各专家结论，输出最终答案（流式）
func (p *Panel) runChair(ctx context.Context, factory ChatModelFactory, question string, subtasks []SubTask, conclusions []string, cb *PanelCallbacks) (*schema.Message, error) {
	cm, err := factory(ctx)
	if err != nil {
		return nil, fmt.Errorf("初始化主持人模型失败: %w", err)
	}

	var b strings.Builder
	b.WriteString("你是运维专家会诊的主持人。下面是用户原始问题与各位专家的独立分析结论。\n")
	b.WriteString("你的职责：1) 交叉验证各结论，发现矛盾时指出并以证据更充分者为准；2) 综合形成统一、可执行的最终回答；3) 关键判断标注来源专家。\n")
	b.WriteString("如果某位专家分析失败，向用户简要说明该维度未能覆盖，不要编造该维度内容。\n\n")
	b.WriteString("【用户问题】\n" + question + "\n\n【各专家结论】\n")
	for i, st := range subtasks {
		sp, _ := specialistByKey(st.Specialist)
		fmt.Fprintf(&b, "--- %s ---\n子问题：%s\n结论：%s\n\n", sp.Title, st.Question, conclusions[i])
	}

	messages := []*schema.Message{
		schema.SystemMessage(b.String()),
		schema.UserMessage("请输出给用户的最终中文回答（Markdown 格式）。"),
	}
	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("主持人汇总失败: %w", err)
	}
	if cb != nil && cb.OnChunk != nil && resp.Content != "" {
		cb.OnChunk(resp.Content)
	}
	return resp, nil
}

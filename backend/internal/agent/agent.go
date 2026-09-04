// Package agent 提供公共的 AI Agent 执行框架：数据源工具注册表 +
// 单循环 Function Calling 执行逻辑，供对话、巡检报告等所有 Agent 复用。
package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// defaultMaxIterations Agent 默认最大迭代次数，防止死循环
const defaultMaxIterations = 5

// Callbacks 供上层（如 SSE 控制器、巡检执行器）观察 Agent 执行过程
type Callbacks struct {
	OnStatus     func(status string)       // 中间状态，例如"正在查询指标平台..."
	OnToolCall   func(name, args string)   // 模型决定调用某个工具
	OnToolResult func(name, result string) // 工具执行完毕
	OnChunk      func(chunk string)        // 最终答案的文本片段
}

// Options Agent 配置
type Options struct {
	// Instructions 该 Agent 的角色/任务描述，会作为 system prompt 的开头
	Instructions string
	// EnforceDataSource 是否启用防幻觉强校验：
	// 用户问题命中数据源关键词时，必须调用 query_* 工具，否则拒绝回答。
	// 对话等开放式问答场景建议开启；巡检等任务式场景由模型自行决策，一般不开启。
	EnforceDataSource bool
	// MaxIterations 最大迭代次数，默认 5
	MaxIterations int
}

// Agent 封装单循环 Function Calling 的执行逻辑
type Agent struct {
	cm       model.ChatModel
	registry *Registry
	opts     Options
}

// New 创建 Agent
func New(cm model.ChatModel, registry *Registry, opts Options) *Agent {
	if opts.MaxIterations <= 0 {
		opts.MaxIterations = defaultMaxIterations
	}
	return &Agent{cm: cm, registry: registry, opts: opts}
}

// Run 执行 Agent 循环，返回最终答案消息（含 ResponseMeta）
func (a *Agent) Run(ctx context.Context, messages []*schema.Message, cb *Callbacks) (*schema.Message, error) {
	toolMap, infos, err := a.registry.ToolMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("加载工具失败: %w", err)
	}

	if len(infos) > 0 {
		if err := a.cm.BindTools(infos); err != nil {
			return nil, fmt.Errorf("绑定工具失败: %w", err)
		}
	}

	// 用动态 system prompt 替换第一条系统消息
	prompt, err := a.buildSystemPrompt(ctx, infos)
	if err != nil {
		return nil, err
	}
	messages = replaceSystemPrompt(messages, prompt)

	// 判断当前问题是否明显需要查询数据源（仅对话场景开启强校验）
	requireData := a.opts.EnforceDataSource && requiresDataSource(lastUserContent(messages))
	var invokedQueryTools []string
	var reminderSent bool

	for i := 0; i < a.opts.MaxIterations; i++ {
		resp, err := a.cm.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("大模型调用失败: %w", err)
		}

		if len(resp.ToolCalls) == 0 {
			// 模型未调用工具直接给出答案
			if requireData && len(invokedQueryTools) == 0 {
				if !reminderSent && i < a.opts.MaxIterations-1 {
					// 先追加一条强提醒，让模型重新决策
					messages = append(messages, schema.SystemMessage(
						"注意：该问题涉及具体运维数据，你必须先调用 query_prometheus / query_elasticsearch / query_pyroscope 获取真实数据，禁止凭先验知识或编造数据回答。如果工具调用失败，请向用户说明无法获取数据。",
					))
					reminderSent = true
					continue
				}
				// 提醒后仍不查询，返回兜底拒绝消息，避免幻觉
				return refusalMessage("该问题需要查询运维数据源才能准确回答，但我未能获取到真实数据。请检查数据源配置，或更具体地描述你想查询的指标、日志、节点、服务或火焰图。"), nil
			}

			// 最终答案
			if cb != nil && cb.OnChunk != nil && resp.Content != "" {
				cb.OnChunk(resp.Content)
			}
			return resp, nil
		}

		// 记录本次 tool_calls 中是否有真正的查询工具
		for _, tc := range resp.ToolCalls {
			if strings.HasPrefix(tc.Function.Name, "query_") {
				invokedQueryTools = append(invokedQueryTools, tc.Function.Name)
			}
		}

		// 记录模型发出的 tool_calls
		messages = append(messages, schema.AssistantMessage(resp.Content, resp.ToolCalls))

		if cb != nil && cb.OnStatus != nil {
			cb.OnStatus(fmt.Sprintf("正在调用 %d 个数据查询工具...", len(resp.ToolCalls)))
		}

		// 并发执行工具
		results := a.executeToolCalls(ctx, resp.ToolCalls, toolMap, cb)

		// 将工具结果回填到上下文
		for idx, tc := range resp.ToolCalls {
			content := results[idx]
			if content == "" {
				content = "{}"
			}
			messages = append(messages, schema.ToolMessage(
				content,
				tc.ID,
				schema.WithToolName(tc.Function.Name),
			))
		}
	}

	return nil, fmt.Errorf("Agent 超过最大迭代次数 %d", a.opts.MaxIterations)
}

func (a *Agent) executeToolCalls(ctx context.Context, calls []schema.ToolCall, toolMap map[string]tool.InvokableTool, cb *Callbacks) []string {
	type result struct {
		idx     int
		content string
	}

	results := make([]string, len(calls))
	if len(calls) == 0 {
		return results
	}

	var wg sync.WaitGroup
	resCh := make(chan result, len(calls))

	for idx, tc := range calls {
		wg.Add(1)
		go func(idx int, tc schema.ToolCall) {
			defer wg.Done()

			if cb != nil && cb.OnToolCall != nil {
				cb.OnToolCall(tc.Function.Name, tc.Function.Arguments)
			}

			t, ok := toolMap[tc.Function.Name]
			var content string
			var err error
			if !ok {
				err = fmt.Errorf("未知工具: %s", tc.Function.Name)
			} else {
				content, err = t.InvokableRun(ctx, tc.Function.Arguments)
			}
			if err != nil {
				content = fmt.Sprintf(`{"error": %q}`, err.Error())
			}

			if cb != nil && cb.OnToolResult != nil {
				cb.OnToolResult(tc.Function.Name, content)
			}
			resCh <- result{idx: idx, content: content}
		}(idx, tc)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	for r := range resCh {
		results[r.idx] = r.content
	}
	return results
}

func (a *Agent) buildSystemPrompt(ctx context.Context, infos []*schema.ToolInfo) (string, error) {
	var b strings.Builder
	b.WriteString(a.opts.Instructions)
	b.WriteString("\n\n你可以使用以下工具查询已配置的数据源（仅只读）：\n")
	for _, info := range infos {
		b.WriteString(fmt.Sprintf("- %s: %s\n", info.Name, info.Desc))
	}
	b.WriteString("\n强制规则：\n")
	b.WriteString("1. 如果用户问题涉及具体指标、日志、性能剖析、节点、容器、服务、应用、TopN 排行等运维数据，你必须先调用对应的数据源工具获取真实数据，禁止凭先验知识或编造数据回答。\n")
	b.WriteString("2. 如果你不确定该用哪个数据源，先调用 list_data_sources 获取可用数据源。\n")
	b.WriteString("3. 只有工具返回的数据才能作为回答依据；如果工具调用失败或返回空，向用户说明无法获取数据，而不是编造。\n")
	b.WriteString("4. 时间范围未指定时，默认使用最近 1 小时。\n")
	b.WriteString("5. 禁止执行修改、删除、写入类操作。\n")
	b.WriteString("\n示例：\n")
	b.WriteString("- 用户问“最近1小时 CPU Top5” → 必须调用 query_prometheus\n")
	b.WriteString("- 用户问“my-service 的火焰图热点” → 必须调用 query_pyroscope\n")
	b.WriteString("- 用户问“nginx 最近 ERROR 日志” → 必须调用 query_elasticsearch\n")
	b.WriteString("- 用户问“什么是 SRE” → 可以不调用工具\n")

	list, err := a.registry.Manager().LoadEnabled(ctx)
	if err != nil {
		return "", fmt.Errorf("加载数据源失败: %w", err)
	}
	if len(list) > 0 {
		b.WriteString("\n当前已启用的数据源：\n")
		for _, ds := range list {
			b.WriteString(fmt.Sprintf("- %s (%s)\n", ds.Name, ds.Type))
		}
	}
	return b.String(), nil
}

func replaceSystemPrompt(messages []*schema.Message, prompt string) []*schema.Message {
	if len(messages) > 0 && messages[0].Role == schema.System {
		messages[0].Content = prompt
		return messages
	}
	return append([]*schema.Message{schema.SystemMessage(prompt)}, messages...)
}

func lastUserContent(messages []*schema.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.User {
			return messages[i].Content
		}
	}
	return ""
}

func refusalMessage(content string) *schema.Message {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
	}
}

// Agent 循环行为测试：最大迭代部分结果、工具结果收口、重复调用去重、正常收敛。
package agent

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"unicode/utf8"

	"aiops/models"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// fakeChatModel 按预置序列逐次返回响应；序列用完后停在最后一条。
type fakeChatModel struct {
	responses []*schema.Message
	calls     int
}

func (f *fakeChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if len(f.responses) == 0 {
		return nil, fmt.Errorf("fakeChatModel: 无预置响应")
	}
	idx := f.calls
	if idx >= len(f.responses) {
		idx = len(f.responses) - 1
	}
	f.calls++
	return f.responses[idx], nil
}

func (f *fakeChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("fakeChatModel: Stream 未实现")
}

func (f *fakeChatModel) BindTools(tools []*schema.ToolInfo) error { return nil }

func testRegistry(t *testing.T) *Registry {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 sqlite 失败: %v", err)
	}
	if err := db.AutoMigrate(&models.Datasource{}, &models.Service{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return NewRegistry(db)
}

func toolCallMsg(name, args, thought string) *schema.Message {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: thought,
		ToolCalls: []schema.ToolCall{{
			ID:       "call_" + name,
			Function: schema.FunctionCall{Name: name, Arguments: args},
		}},
	}
}

// 达到最大迭代次数时返回部分结果（含最后一轮中间分析），而不是 error。
func TestRun_MaxIterationsReturnsPartialResult(t *testing.T) {
	reg := testRegistry(t)
	cm := &fakeChatModel{responses: []*schema.Message{
		toolCallMsg("list_services", "{}", "第一步：先列出所有服务。"),
		toolCallMsg("list_services", "{}", "第二步：再确认服务详情。"),
		toolCallMsg("list_services", "{}", "第三步：继续分析。"),
	}}
	ag := New(cm, reg, Options{Instructions: "测试", MaxIterations: 3})

	var streamed string
	resp, err := ag.Run(context.Background(), []*schema.Message{schema.UserMessage("分析系统状态")}, &Callbacks{
		OnChunk: func(chunk string) { streamed += chunk },
	})
	if err != nil {
		t.Fatalf("期望部分结果而非错误，实际: %v", err)
	}
	if !strings.Contains(resp.Content, "工具调用上限") {
		t.Errorf("部分结果应说明达到上限，实际: %q", resp.Content)
	}
	if !strings.Contains(resp.Content, "第三步：继续分析。") {
		t.Errorf("部分结果应包含最后一轮中间分析，实际: %q", resp.Content)
	}
	if !strings.Contains(streamed, resp.Content) {
		t.Errorf("部分结果应经 OnChunk 流出")
	}
}

// 正常收敛路径：模型直接给出最终答案时原样返回。
func TestRun_ConvergesWithoutToolCalls(t *testing.T) {
	reg := testRegistry(t)
	cm := &fakeChatModel{responses: []*schema.Message{
		{Role: schema.Assistant, Content: "最终答案"},
	}}
	ag := New(cm, reg, Options{Instructions: "测试"})

	resp, err := ag.Run(context.Background(), []*schema.Message{schema.UserMessage("什么是 SRE")}, nil)
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	if resp.Content != "最终答案" {
		t.Errorf("期望原样返回最终答案，实际: %q", resp.Content)
	}
}

// 超大工具结果按上限收口：不超上限、UTF-8 完整、含截断提示。
func TestCapToolResult(t *testing.T) {
	short := strings.Repeat("短", 10)
	if got := capToolResult(short); got != short {
		t.Errorf("短结果不应被截断")
	}

	long := strings.Repeat("这是一段比较长的中文工具结果内容。", 2000) // ~64KB+
	got := capToolResult(long)
	if len(got) > liveObservationCapBytes {
		t.Errorf("截断结果 %d 字节超过上限 %d", len(got), liveObservationCapBytes)
	}
	if !utf8.ValidString(got) {
		t.Errorf("截断结果包含非法 UTF-8（中文被切碎）")
	}
	if !strings.Contains(got, "结果过长已截断") {
		t.Errorf("截断结果应含提示头")
	}
	if !strings.Contains(got, "中间省略") {
		t.Errorf("截断结果应含中缝标记")
	}
	// 首尾保留：开头与结尾的原文片段都应可见
	if !strings.HasPrefix(got, "(结果过长已截断") || !strings.Contains(got, "中文工具结果内容。") {
		t.Errorf("截断结果应保留首尾两段")
	}
}

// 同 Run 内同名同参的重复工具调用只真正执行一次（executeToolCalls 级验证）。
func TestExecuteToolCalls_Dedup(t *testing.T) {
	var invoked int32
	fake, err := utils.InferTool[struct {
		Q string `json:"q"`
	}, struct {
		R string `json:"r"`
	}](
		"fake_query", "测试工具",
		func(ctx context.Context, in struct {
			Q string `json:"q"`
		}) (struct {
			R string `json:"r"`
		}, error) {
			atomic.AddInt32(&invoked, 1)
			return struct {
				R string `json:"r"`
			}{R: "ok"}, nil
		})
	if err != nil {
		t.Fatalf("构造工具失败: %v", err)
	}

	a := &Agent{}
	toolMap := map[string]tool.InvokableTool{"fake_query": fake}
	dedup := map[string]string{}
	calls := []schema.ToolCall{
		{ID: "c1", Function: schema.FunctionCall{Name: "fake_query", Arguments: `{"q":"1"}`}},
		{ID: "c2", Function: schema.FunctionCall{Name: "fake_query", Arguments: `{"q":"1"}`}},
		{ID: "c3", Function: schema.FunctionCall{Name: "fake_query", Arguments: `{"q":"2"}`}},
	}

	// 同一次 executeToolCalls 内的同名同参并发去重 + 跨迭代去重
	first := a.executeToolCalls(context.Background(), calls, toolMap, nil, dedup)
	second := a.executeToolCalls(context.Background(), calls[:1], toolMap, nil, dedup)

	if atomic.LoadInt32(&invoked) != 2 {
		t.Errorf("同名同参调用应只执行一次（q=1 与 q=2 各一次），实际执行 %d 次", invoked)
	}
	for i, r := range first {
		if r == "" {
			t.Errorf("第 %d 个调用应有结果", i)
		}
	}
	if !strings.Contains(first[1], "未重复执行") {
		t.Errorf("同轮去重结果应附注说明，实际: %q", first[1])
	}
	if !strings.Contains(second[0], "未重复执行") {
		t.Errorf("跨轮去重结果应附注说明，实际: %q", second[0])
	}
}

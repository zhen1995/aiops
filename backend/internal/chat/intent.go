package chat

import (
	"strings"
	"unicode"
)

// dataSourceKeywords 用于判断用户问题是否明显需要查询数据源。
// 只保留较强的信号词，避免把纯知识问答误判为必须查询。
var dataSourceKeywords = []string{
	// 指标 / 监控
	"cpu", "内存", "memory", "qps", "tps", "rps", "延迟", "latency",
	"请求量", "吞吐量", "错误率", "error rate", "失败率", "超时",
	"响应时间", "流量", "并发", "负载", "饱和度", "使用率", "utilization",
	"指标", "metric", "prometheus", "promql",
	// 日志
	"日志", "log", "elasticsearch", "es", "报错", "异常", "exception",
	"stacktrace", "trace",
	// 性能剖析
	"pyroscope", "火焰图", "profile", "profiling", "热点", "函数耗时", "cpu profile",
	"alloc", "inuse", "heap", "goroutine", "mutex", "block",
	// 资源 / 实体
	"节点", "node", "容器", "container", "pod", "服务", "service",
	"应用", "app", "实例", "主机", "host", "namespace", "命名空间",
	"deployment", "集群", "cluster",
	// 聚合 / 排序 / 时间范围
	"top", "排行", "排名", "最高", "最低", "平均", "最大", "最小", "总和",
	"最近", "过去", "小时", "分钟", "天", "周内",
}

// requiresDataSource 判断用户问题是否需要查询真实数据源。
// 返回 true 时，Agent 会强制要求模型调用 query_* 工具，否则拒绝回答以避免编造数据。
func requiresDataSource(text string) bool {
	if text == "" {
		return false
	}
	text = strings.ToLower(simpleNormalize(text))
	for _, kw := range dataSourceKeywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// simpleNormalize 简单归一化：去掉常见标点，便于关键词匹配
func simpleNormalize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			b.WriteRune(' ')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

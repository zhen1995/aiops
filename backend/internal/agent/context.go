// 工具结果收口：单条工具结果回填进 messages 前按上限截断，防止超大查询结果
// 在剩余迭代中反复携带、撑爆上下文 token。纯确定性策略，零 LLM 成本。
// 思路借鉴夜莺 aiagent/context_manager.go 的 capObservation/truncateObservation：
// 保留首尾而非只留头——头部承载关键数据，尾部承载收尾/错误提示；
// 提示头写明原始大小，模型据此收窄条件重查而不是接受残缺输出。
package agent

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// liveObservationCapBytes 单条工具结果进 messages 的字节上限。
const liveObservationCapBytes = 32 * 1024

// capToolResult 把超长工具结果截成「提示头 + 正文头 + 中缝标记 + 正文尾」，
// 返回值总字节数不超过 liveObservationCapBytes。SSE 回调推送的是截断前原文，
// 这里只收口喂给模型的上下文。
func capToolResult(content string) string {
	if len(content) <= liveObservationCapBytes {
		return content
	}

	header := fmt.Sprintf("(结果过长已截断：原始 %d 字节 / %d 行，下面只保留首尾两段；需要完整内容请收窄查询条件后重新调用工具)\n",
		len(content), strings.Count(content, "\n")+1)
	const markerTpl = "\n...(中间省略 %d 字节)...\n"
	body := liveObservationCapBytes - len(header) - len(fmt.Sprintf(markerTpl, len(content)))
	if body <= 0 {
		return runePrefix(content, liveObservationCapBytes)
	}

	head := body / 2
	prefix := runePrefix(content, head)
	suffix := runeSuffix(content, body-head)
	return header + prefix + fmt.Sprintf(markerTpl, len(content)-len(prefix)-len(suffix)) + suffix
}

// runePrefix 返回 s 中不超过 n 字节、且不切碎 UTF-8 的最长前缀。
func runePrefix(s string, n int) string {
	if n >= len(s) {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// runeSuffix 返回 s 中不超过 n 字节、且不切碎 UTF-8 的最长后缀。
func runeSuffix(s string, n int) string {
	if n >= len(s) {
		return s
	}
	i := len(s) - n
	for i < len(s) && !utf8.RuneStart(s[i]) {
		i++
	}
	return s[i:]
}

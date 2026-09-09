package datasource

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// promLabelKeyRegexp 合法的 Prom 标签名
var promLabelKeyRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// BuildPromSelector 将标签选择器构造成 PromQL 匹配器形式 {k1="v1",k2="v2"}；
// 非法标签名跳过，value 用 strconv.Quote 转义；空 map 返回 {}
func BuildPromSelector(labels map[string]string) string {
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		if !promLabelKeyRegexp.MatchString(k) {
			continue
		}
		parts = append(parts, k+"="+strconv.Quote(v))
	}
	sort.Strings(parts)
	return "{" + strings.Join(parts, ",") + "}"
}

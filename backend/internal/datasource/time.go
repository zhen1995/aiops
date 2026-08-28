package datasource

import (
	"fmt"
	"strings"
	"time"
)

// ParseTime 解析相对时间或绝对时间。
//
// 支持格式：
//   - "now"：参考时间
//   - "1h" / "30m" / "1d"：参考时间往前推对应时长
//   - "-1h" / "+30m"：显式相对偏移
//   - ISO / RFC3339 时间字符串
func ParseTime(input string, ref time.Time) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" || strings.EqualFold(input, "now") {
		return ref, nil
	}

	// 显式偏移
	if strings.HasPrefix(input, "+") || strings.HasPrefix(input, "-") {
		d, err := time.ParseDuration(input)
		if err == nil {
			return ref.Add(d), nil
		}
	}

	// 裸时长，默认往前推
	if d, err := time.ParseDuration(input); err == nil {
		return ref.Add(-d), nil
	}

	// 尝试常见绝对时间格式
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, input); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析时间: %s", input)
}

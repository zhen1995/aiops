package agent

import (
	"testing"
	"time"
)

func TestParseOffset(t *testing.T) {
	cases := map[string]time.Duration{
		"30m": 30 * time.Minute,
		"12h": 12 * time.Hour,
		"7d":  7 * 24 * time.Hour,
	}
	for in, want := range cases {
		got, err := parseOffset(in)
		if err != nil || got != want {
			t.Fatalf("parseOffset(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "abc", "7x", "-1h", "0d"} {
		if _, err := parseOffset(bad); err == nil {
			t.Fatalf("parseOffset(%q) should fail", bad)
		}
	}
}

func TestSeriesKey(t *testing.T) {
	a := seriesKey(map[string]string{"b": "2", "a": "1"})
	b := seriesKey(map[string]string{"a": "1", "b": "2"})
	if a != b || a == "" {
		t.Fatalf("seriesKey unstable: %q vs %q", a, b)
	}
}

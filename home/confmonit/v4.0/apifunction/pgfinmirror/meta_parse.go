package pgfinmirror

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func stringFromAny(v any) string {
	switch n := v.(type) {
	case string:
		return strings.TrimSpace(n)
	case json.Number:
		return n.String()
	case float64:
		if n == float64(int64(n)) {
			return fmt.Sprintf("%d", int64(n))
		}
		return fmt.Sprintf("%v", n)
	case int:
		return fmt.Sprintf("%d", n)
	case int64:
		return fmt.Sprintf("%d", n)
	default:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func floatFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func timeFromAny(v any) *time.Time {
	switch n := v.(type) {
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return nil
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05+0000", "2006-01-02 15:04:05-0700"} {
			if t, err := time.Parse(layout, s); err == nil {
				return &t
			}
		}
	case float64:
		return msToTime(int64(n))
	case int:
		return msToTime(int64(n))
	case int64:
		return msToTime(n)
	case json.Number:
		i, _ := n.Int64()
		return msToTime(i)
	}
	return nil
}

func msToTime(ms int64) *time.Time {
	if ms <= 0 {
		return nil
	}
	t := time.UnixMilli(ms).UTC()
	return &t
}

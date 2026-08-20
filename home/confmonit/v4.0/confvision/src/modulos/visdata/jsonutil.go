package visdata

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"
)

func nullStr(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func nullBool(v sql.NullBool) any {
	if v.Valid {
		return v.Bool
	}
	return nil
}

func nullInt(v sql.NullInt64) any {
	if v.Valid {
		return int(v.Int64)
	}
	return nil
}

func nullFloat(v sql.NullFloat64) any {
	if v.Valid {
		return v.Float64
	}
	return nil
}

func nullTime(v sql.NullTime) any {
	if v.Valid {
		return v.Time.UTC().Format(time.RFC3339)
	}
	return nil
}

func normalizeValue(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(t)
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	case bool, float64, int64, string:
		return t
	default:
		return t
	}
}

func writeJSON(status int) ([]byte, int, error) {
	return nil, status, nil
}

func marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func parseJSON(body []byte, dst any) error {
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, dst)
}

func strVal(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	if s != "" {
		return s
	}
	return trimAny(v)
}

func boolVal(m map[string]any, key string) *bool {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case bool:
		return &t
	case string:
		b := t == "true" || t == "1"
		return &b
	default:
		return nil
	}
}

func intVal(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint64:
		return int(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	default:
		return 0
	}
}

func floatVal(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	default:
		return 0
	}
}

func trimAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		b, _ := json.Marshal(t)
		if len(b) >= 2 && b[0] == '"' {
			var s string
			_ = json.Unmarshal(b, &s)
			return s
		}
		return string(b)
	}
}

func intSliceFromAny(v any) []int {
	switch t := v.(type) {
	case []any:
		out := make([]int, 0, len(t))
		for _, el := range t {
			switch n := el.(type) {
			case float64:
				out = append(out, int(n))
			case int:
				out = append(out, n)
			case json.Number:
				i, _ := n.Int64()
				out = append(out, int(i))
			}
		}
		return out
	case []int:
		return t
	default:
		return nil
	}
}

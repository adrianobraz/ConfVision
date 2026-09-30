package pgatendimento

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

func decodeGruposJSON(raw sql.NullString) ([]string, error) {
	out := []string{}
	if !raw.Valid {
		return out, nil
	}
	s := strings.TrimSpace(raw.String)
	if s == "" || s == "[]" || s == "null" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	if out == nil {
		return []string{}, nil
	}
	return out, nil
}

func strMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		b, _ := json.Marshal(t)
		s := strings.TrimSpace(string(b))
		if len(s) >= 2 && s[0] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
}

func intMap(m map[string]any, key string) int {
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
	case int64:
		return int(t)
	default:
		return 0
	}
}

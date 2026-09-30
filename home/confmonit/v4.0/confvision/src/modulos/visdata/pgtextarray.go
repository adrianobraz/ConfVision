package visdata

import (
	"database/sql"
	"encoding/json"
	"strings"
)

const sqlGruposFromJSON = `COALESCE(ARRAY(SELECT json_array_elements_text($1::json)), ARRAY[]::text[])`

func encodeGruposJSON(grupos []string) (string, error) {
	if grupos == nil {
		grupos = []string{}
	}
	b, err := json.Marshal(grupos)
	if err != nil {
		return "[]", err
	}
	return string(b), nil
}

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

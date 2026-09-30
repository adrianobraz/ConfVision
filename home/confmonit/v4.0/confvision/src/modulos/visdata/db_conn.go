package visdata

import (
	"database/sql/driver"
	"errors"
	"strings"
)

func isBadConn(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, driver.ErrBadConn) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "eof")
}

func resetDB() {
	dbMu.Lock()
	defer dbMu.Unlock()
	if dbInst != nil {
		_ = dbInst.Close()
		dbInst = nil
	}
	dbErr = nil
}

func humanizeDBErr(err error) error {
	if err == nil {
		return nil
	}
	if isBadConn(err) {
		return errors.New("Postgres indisponivel — verifique POSTGRES_URL no servidor e reinicie o ConfVision")
	}
	return err
}

func intFromAny(v any) int {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return intVal(map[string]any{"v": v}, "v")
	}
}

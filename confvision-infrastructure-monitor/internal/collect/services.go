package collect

import (
	"strings"
)

// PapelDoContainer classifica pelo nome real. Nome desconhecido não vira erro.
func PapelDoContainer(nome string) string {
	n := strings.ToLower(strings.TrimPrefix(nome, "/"))
	switch {
	case strings.Contains(n, "worker"):
		return "worker"
	case strings.Contains(n, "rtsp-relay") || strings.Contains(n, "rtsp_relay") || strings.Contains(n, "relay"):
		return "rtsp-relay"
	case strings.Contains(n, "mediamtx") || n == "mtx":
		return "mediamtx"
	case strings.Contains(n, "motion"):
		return "motion"
	case strings.Contains(n, "timelapse"):
		return "timelapse"
	case strings.Contains(n, "dvr"):
		return "dvr"
	case strings.Contains(n, "sensor"):
		return "sensor"
	case strings.Contains(n, "sync-agent") || strings.Contains(n, "sync_agent") || strings.Contains(n, "syncagent"):
		return "sync-agent"
	case strings.Contains(n, "redis"):
		return "redis"
	default:
		return ""
	}
}

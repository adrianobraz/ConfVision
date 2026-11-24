package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Listen       string
	BehindProxy  bool
	HostRoot     string
	ServerID     string
	ServerNome   string
	Secret       string
	Interval     time.Duration
	TLSCert      string
	TLSKey       string
	MetricsLog   string
	MediaMTXURL  string
	MediaMTXUser string
	MediaMTXPass string
	RedisURL     string
	DockerSocket string
	CPUQuota     *float64
	RAMQuotaMB   *float64
	AlertCPUSus  float64
	AlertCPUAl   float64
	AlertRAMSus  float64
	AlertRAMAl   float64
	AlertDiskSus float64
	AlertDiskAl  float64
	AlertYoloP95 float64
	RedisKeys    []string
}

func Load() (Config, error) {
	cfg := Config{
		Listen:       env("MONITOR_LISTEN", "127.0.0.1:8090"),
		BehindProxy:  envBool("MONITOR_BEHIND_PROXY", false),
		HostRoot:     env("MONITOR_HOST_ROOT", "/"),
		ServerID:     env("MONITOR_SERVER_ID", "vps-local"),
		ServerNome:   env("MONITOR_SERVER_NOME", "VPS Hostinger"),
		Secret:       strings.TrimSpace(os.Getenv("MONITOR_API_SECRET")),
		Interval:     time.Duration(envInt("MONITOR_INTERVAL_SEC", 10)) * time.Second,
		TLSCert:      strings.TrimSpace(os.Getenv("MONITOR_TLS_CERT")),
		TLSKey:       strings.TrimSpace(os.Getenv("MONITOR_TLS_KEY")),
		MetricsLog:   strings.TrimSpace(os.Getenv("METRICS_LOG_PATH")),
		MediaMTXURL:  strings.TrimRight(strings.TrimSpace(os.Getenv("MEDIAMTX_API_BASE")), "/"),
		MediaMTXUser: os.Getenv("MEDIAMTX_API_USER"),
		MediaMTXPass: os.Getenv("MEDIAMTX_API_PASS"),
		RedisURL:     strings.TrimSpace(os.Getenv("REDIS_URL")),
		DockerSocket: env("DOCKER_SOCKET", "/var/run/docker.sock"),
		AlertCPUSus:  envFloat("ALERT_CPU_SUSPEITA", 70),
		AlertCPUAl:   envFloat("ALERT_CPU_ALERTA", 85),
		AlertRAMSus:  envFloat("ALERT_RAM_SUSPEITA", 75),
		AlertRAMAl:   envFloat("ALERT_RAM_ALERTA", 90),
		AlertDiskSus: envFloat("ALERT_DISCO_SUSPEITA", 80),
		AlertDiskAl:  envFloat("ALERT_DISCO_ALERTA", 90),
		AlertYoloP95: envFloat("ALERT_YOLO_P95_MS", 400),
		RedisKeys: []string{
			env("REDIS_EVENT_QUEUE_KEY", "confvision:eventos"),
			env("REDIS_YOLO_QUEUE_KEY", "confvision:yolo:queue"),
		},
	}
	if cfg.Interval < time.Second {
		cfg.Interval = 10 * time.Second
	}
	if v, ok := envOptionalFloat("MONITOR_CPU_QUOTA"); ok {
		cfg.CPUQuota = &v
	}
	if v, ok := envOptionalFloat("MONITOR_RAM_QUOTA_MB"); ok {
		cfg.RAMQuotaMB = &v
	}
	if err := cfg.ValidateListen(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// ValidateListen recusa bind fora do loopback sem certificado.
// A API entre servidores precisa de HTTPS. Este processo não abre firewall.
func (c Config) ValidateListen() error {
	host, _, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return fmt.Errorf("MONITOR_LISTEN invalido")
	}
	if host == "" {
		host = "0.0.0.0"
	}
	if isLoopback(host) || c.BehindProxy {
		return nil
	}
	if c.TLSCert == "" || c.TLSKey == "" {
		return fmt.Errorf("bind fora do loopback exige MONITOR_TLS_CERT e MONITOR_TLS_KEY")
	}
	return nil
}

func (c Config) TLSEnabled() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "sim", "yes":
		return true
	case "0", "false", "nao", "não", "no":
		return false
	default:
		return fallback
	}
}

func env(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envFloat(key string, fallback float64) float64 {
	v, ok := envOptionalFloat(key)
	if !ok {
		return fallback
	}
	return v
}

func envOptionalFloat(key string) (float64, bool) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

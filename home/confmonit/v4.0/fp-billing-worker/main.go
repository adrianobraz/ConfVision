package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIBase        string
	WorkerKey      string
	DiasAntecedencia int
	Timeout        time.Duration
	DryRun         bool
	RunOnce        bool
	ScheduleWeekday time.Weekday // Monday
	ScheduleHour   int
	ScheduleMinute int
	TZ             *time.Location
}

type XanoClient struct {
	cfg    Config
	client *http.Client
}

func main() {
	loadDotEnv()
	cfg := loadConfig()
	log.Printf("fp-billing-worker iniciado | api=%s | dry_run=%v | run_once=%v", cfg.APIBase, cfg.DryRun, cfg.RunOnce)

	worker := &BillingWorker{
		xano: NewXanoClient(cfg),
		cfg:  cfg,
	}

	if cfg.RunOnce {
		if err := worker.RunWeekly(); err != nil {
			log.Fatalf("erro no ciclo: %v", err)
		}
		return
	}

	for {
		wait := durationUntilNext(cfg)
		log.Printf("proxima execucao em %s (segunda %02d:%02d %s)", wait.Round(time.Second), cfg.ScheduleHour, cfg.ScheduleMinute, cfg.TZ.String())
		time.Sleep(wait)
		if err := worker.RunWeekly(); err != nil {
			log.Printf("erro no ciclo: %v", err)
		}
	}
}

func NewXanoClient(cfg Config) *XanoClient {
	return &XanoClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *XanoClient) Post(path string, payload any) ([]byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url := strings.TrimRight(c.cfg.APIBase, "/") + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("xano HTTP %d: %s", res.StatusCode, string(raw))
	}
	return raw, nil
}

func loadConfig() Config {
	tzName := getenv("TZ", "America/Sao_Paulo")
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		log.Printf("TZ invalido %q, usando UTC", tzName)
		loc = time.UTC
	}
	return Config{
		APIBase:          getenv("XANO_API_FINANCEIRO", "https://xpcy-oyme-lno7.b2.xano.io/api:-WvTZ3QM"),
		WorkerKey:        getenv("WORKER_SECRET", ""),
		DiasAntecedencia: getenvInt("DIAS_ANTECEDENCIA_FATURA", 5),
		Timeout:          time.Duration(getenvInt("HTTP_TIMEOUT_SEC", 60)) * time.Second,
		DryRun:           strings.EqualFold(getenv("DRY_RUN", "false"), "true"),
		RunOnce:          hasArg("--run-once"),
		ScheduleWeekday:  time.Monday,
		ScheduleHour:     getenvInt("RUN_HOUR", 3),
		ScheduleMinute:   getenvInt("RUN_MINUTE", 0),
		TZ:               loc,
	}
}

func durationUntilNext(cfg Config) time.Duration {
	now := time.Now().In(cfg.TZ)
	target := time.Date(now.Year(), now.Month(), now.Day(), cfg.ScheduleHour, cfg.ScheduleMinute, 0, 0, cfg.TZ)
	for !target.After(now) || target.Weekday() != cfg.ScheduleWeekday {
		target = target.Add(24 * time.Hour)
	}
	return target.Sub(now)
}

func hasArg(flag string) bool {
	for _, a := range os.Args[1:] {
		if a == flag {
			return true
		}
	}
	return false
}

func loadDotEnv() {
	paths := []string{".env"}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), ".env"))
	}
	for _, p := range paths {
		if applyDotEnvFile(p) {
			return
		}
	}
}

func applyDotEnvFile(filename string) bool {
	b, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	text := strings.ReplaceAll(string(b), "\r\n", "\n")
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return true
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

func runConfVisionGradeLoop(client *http.Client, cfg Config, shutdown <-chan struct{}) {
	ticker := time.NewTicker(cfg.ConfVisionGradeInterval)
	defer ticker.Stop()

	runConfVisionGradeOnce(client, cfg)
	for {
		select {
		case <-shutdown:
			log.Printf("confvision_grade: loop encerrado")
			return
		case <-ticker.C:
			runConfVisionGradeOnce(client, cfg)
		}
	}
}

func runConfVisionGradeOnce(client *http.Client, cfg Config) {
	if !atomic.CompareAndSwapInt32(&runningConfVisionGrade, 0, 1) {
		log.Printf("confvision_grade: execucao anterior em andamento, pulando ciclo")
		return
	}
	defer atomic.StoreInt32(&runningConfVisionGrade, 0)

	start := time.Now()
	attempts := cfg.Retries + 1
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			backoff := backoffForAttempt(attempt)
			log.Printf("confvision_grade retry=%d backoff=%s", attempt-1, backoff)
			time.Sleep(backoff)
		}

		err := postCvgWorkerTick(client, cfg)
		if err == nil {
			log.Printf("confvision_grade sucesso | attempt=%d/%d | duracao=%s", attempt, attempts, time.Since(start))
			return
		}

		lastErr = err
		log.Printf("confvision_grade falha | attempt=%d/%d | erro=%v", attempt, attempts, err)
	}

	log.Printf("confvision_grade erro final | duracao=%s | erro=%v", time.Since(start), lastErr)
}

func postCvgWorkerTick(client *http.Client, cfg Config) error {
	if strings.TrimSpace(cfg.ConfVisionGradeTickURL) == "" {
		return fmt.Errorf("CONFVISION_GRADE_TICK_URL nao configurada")
	}

	loc, err := time.LoadLocation(cfg.ConfVisionGradeLocation)
	if err != nil {
		return fmt.Errorf("timezone invalido %q: %w", cfg.ConfVisionGradeLocation, err)
	}

	now := time.Now().In(loc)
	payload := map[string]any{
		"dia_semana": cvgDiaSemana(now),
		"hora":       now.Format("15:04"),
		"data_ref":   now.Format("2006-01-02"),
	}

	if strings.TrimSpace(cfg.ConfVisionGradeWorkerKey) != "" {
		payload["worker_key"] = cfg.ConfVisionGradeWorkerKey
	}

	body, status, err := doJSONRequestWithHeaders(client, http.MethodPost, cfg.ConfVisionGradeTickURL, payload, map[string]string{
		"X-CVG-Worker-Key": cfg.ConfVisionGradeWorkerKey,
	})
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("status=%d body=%s", status, compactBody(body))
	}

	logCvgTickResult(body)
	return nil
}

func logCvgTickResult(body []byte) {
	var resp struct {
		DataRef   string `json:"data_ref"`
		Hora      string `json:"hora"`
		Total     int    `json:"total"`
		Executados int   `json:"executados"`
		Ignorados int    `json:"ignorados"`
		Erros     int    `json:"erros"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Printf("confvision_grade tick ok | body=%s", compactBody(body))
		return
	}
	if resp.Total == 0 {
		log.Printf("confvision_grade tick | %s %s | nenhum slot pendente", resp.DataRef, resp.Hora)
		return
	}
	log.Printf("confvision_grade tick | %s %s | total=%d executados=%d ignorados=%d erros=%d",
		resp.DataRef, resp.Hora, resp.Total, resp.Executados, resp.Ignorados, resp.Erros)
}

func cvgDiaSemana(t time.Time) int {
	wd := t.Weekday()
	if wd == time.Sunday {
		return 7
	}
	return int(wd)
}

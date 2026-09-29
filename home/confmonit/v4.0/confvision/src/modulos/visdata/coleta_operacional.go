package visdata

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed coleta_operacional_migration.sql
var sqlColetaOperacionalMigration string

var (
	coletaMu        sync.Mutex
	coletaMigration sync.Once
	coletaMigrationErr error
)

func ApplyColetaOperacionalMigration(ctx context.Context) error {
	coletaMigration.Do(func() {
		db, err := DB()
		if err != nil {
			coletaMigrationErr = err
			return
		}
		_, coletaMigrationErr = db.ExecContext(ctx, sqlColetaOperacionalMigration)
	})
	return coletaMigrationErr
}

func coletaInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv("COLETA_RELATORIO_INTERVAL"))
	if raw == "" {
		return 30 * time.Minute
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < time.Minute {
		return 30 * time.Minute
	}
	return d
}

func coletaEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COLETA_RELATORIO_ENABLED")))
	if v == "0" || v == "false" || v == "off" {
		return false
	}
	return true
}

// StartColetaOperacionalBackground aplica migration e roda coleta periódica (additive).
func StartColetaOperacionalBackground() {
	if !coletaEnabled() {
		return
	}
	go func() {
		ctx := context.Background()
		if err := ApplyColetaOperacionalMigration(ctx); err != nil {
			fmt.Printf("[coleta] migration falhou: %v\n", err)
			return
		}
		fmt.Printf("[coleta] migration OK; intervalo=%s\n", coletaInterval())
		time.Sleep(15 * time.Second)
		runColetaOperacional(ctx)
		tick := time.NewTicker(coletaInterval())
		defer tick.Stop()
		for range tick.C {
			runColetaOperacional(ctx)
		}
	}()
}

func runColetaOperacional(ctx context.Context) {
	coletaMu.Lock()
	defer coletaMu.Unlock()
	if err := collectHealthAndMetrics(ctx); err != nil {
		fmt.Printf("[coleta] health/metric: %v\n", err)
	}
	if err := collectStreamRelatorioFromCameras(ctx); err != nil {
		fmt.Printf("[coleta] stream relatorio: %v\n", err)
	}
}

func collectHealthAndMetrics(ctx context.Context) error {
	db, err := DB()
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	_, err = db.ExecContext(ctx, `
INSERT INTO vis_sistema_health (componente, status, mensagem, detalhe_json)
VALUES ('postgres', 'ok', 'ping ok', $1)`, jsonRaw(map[string]any{"coletado": now.Format(time.RFC3339)}))
	if err != nil {
		return err
	}

	rows, err := FetchAllProcessorCapacity(ctx)
	if err == nil {
		for _, row := range rows {
			st := "ok"
			if !row.Reachable {
				st = "down"
			} else if strings.EqualFold(row.CapacityState, "critical") {
				st = "degraded"
			}
			detail, _ := json.Marshal(row)
			_, _ = db.ExecContext(ctx, `
INSERT INTO vis_sistema_health (componente, base_url, status, mensagem, detalhe_json)
VALUES ($1, $2, $3, $4, $5)`,
				"rust_processor", row.BaseURL, st, row.Error+row.LoadAdvisory, detail)
		}
		metricBlob, _ := json.Marshal(map[string]any{
			"processors": rows,
			"coletado_em": now.Format(time.RFC3339),
		})
		_, _ = db.ExecContext(ctx, `
INSERT INTO vis_sistema_metric (escopo, chave, metricas_json)
VALUES ('global', 'rust_processors', $1)`, metricBlob)
	}

	var total, comFalha, pausadas, bloqueadas int
	_ = db.QueryRowContext(ctx, `
SELECT
  COUNT(*)::int,
  COUNT(*) FILTER (WHERE stream_falhas_consecutivas > 0)::int,
  COUNT(*) FILTER (WHERE analitico_pausado IS TRUE)::int,
  COUNT(*) FILTER (WHERE bloqueado IS TRUE)::int
FROM vis_camera`).Scan(&total, &comFalha, &pausadas, &bloqueadas)

	metricCam, _ := json.Marshal(map[string]any{
		"cameras_total": total, "stream_falhas": comFalha,
		"analitico_pausadas": pausadas, "bloqueadas": bloqueadas,
	})
	_, err = db.ExecContext(ctx, `
INSERT INTO vis_sistema_metric (escopo, chave, metricas_json)
VALUES ('global', 'vis_camera_resumo', $1)`, metricCam)
	return err
}

func collectStreamRelatorioFromCameras(ctx context.Context) error {
	db, err := DB()
	if err != nil {
		return err
	}
	qrows, err := db.QueryContext(ctx, `
SELECT id, id_franqueado, stream_falhas_consecutivas, stream_motivo_pausa,
       stream_ultimo_erro, stream_erro_classe, bloqueado, analitico_pausado
FROM vis_camera
WHERE stream_falhas_consecutivas > 0
   OR (stream_motivo_pausa IS NOT NULL AND BTRIM(stream_motivo_pausa) <> '')
   OR (stream_ultimo_erro IS NOT NULL AND BTRIM(stream_ultimo_erro) <> '')
   OR bloqueado IS TRUE`)
	if err != nil {
		return err
	}
	defer qrows.Close()

	for qrows.Next() {
		var id int
		var idFra sql.NullString
		var falhas int
		var motivo, ultErr, errClass sql.NullString
		var bloq, pausa sql.NullBool
		if err := qrows.Scan(&id, &idFra, &falhas, &motivo, &ultErr, &errClass, &bloq, &pausa); err != nil {
			return err
		}
		codigo := "stream_cadastro"
		titulo := "Sinal de problema no cadastro/stream"
		dica := "Verifique RTSP, RTMP, analítico pausado ou bloqueio."
		if bloq.Valid && bloq.Bool {
			codigo = "rtmp_bloqueado"
			titulo = "Câmera com RTMP bloqueado"
		} else if motivo.Valid && strings.HasPrefix(motivo.String, "sistema_stream") {
			codigo = "sistema_stream"
			titulo = "Analítico pausado pelo sistema (stream)"
		} else if falhas > 0 {
			codigo = "stream_falhas_consecutivas"
			titulo = fmt.Sprintf("Falhas consecutivas de stream (%d)", falhas)
		}
		detail := map[string]any{
			"stream_falhas_consecutivas": falhas,
			"stream_motivo_pausa":        nullStrVal(motivo),
			"stream_ultimo_erro":         nullStrVal(ultErr),
			"stream_erro_classe":         nullStrVal(errClass),
			"analitico_pausado":          pausa.Valid && pausa.Bool,
			"bloqueado":                  bloq.Valid && bloq.Bool,
		}
		dj, _ := json.Marshal(detail)
		dedupe := fmt.Sprintf("vis_camera:%d:%s", id, codigo)
		var exists int
		_ = db.QueryRowContext(ctx, `
SELECT 1 FROM vis_stream_relatorio
WHERE referencia_dedupe = $1 AND created_at > NOW() - INTERVAL '30 minutes'
LIMIT 1`, dedupe).Scan(&exists)
		if exists == 1 {
			continue
		}
		_, err = db.ExecContext(ctx, `
INSERT INTO vis_stream_relatorio (
  id_franqueado, vis_camera_id, fonte, motivo_codigo, titulo, dica, severidade, detalhe_json, referencia_dedupe
) VALUES ($1,$2,'vis_camera',$3,$4,$5,'warn',$6,$7)`,
			nullStrVal(idFra), id, codigo, titulo, dica, dj, dedupe)
		if err != nil {
			return err
		}
	}
	return qrows.Err()
}

func nullStrVal(s sql.NullString) any {
	if s.Valid && strings.TrimSpace(s.String) != "" {
		return s.String
	}
	return nil
}

func jsonRaw(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func ListStreamRelatorio(ctx context.Context, idFranqueado string, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var rows *sql.Rows
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado != "" {
		rows, err = db.QueryContext(ctx, `
SELECT id, created_at, id_franqueado, vis_camera_id, fonte, motivo_codigo, titulo, dica, severidade, detalhe_json
FROM vis_stream_relatorio
WHERE id_franqueado = $1
ORDER BY created_at DESC
LIMIT $2`, idFranqueado, limit)
	} else {
		rows, err = db.QueryContext(ctx, `
SELECT id, created_at, id_franqueado, vis_camera_id, fonte, motivo_codigo, titulo, dica, severidade, detalhe_json
FROM vis_stream_relatorio
ORDER BY created_at DESC
LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanColetaRows(rows)
}

func ListSistemaHealth(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, coletado_em, componente, base_url, status, http_status, latencia_ms, mensagem, detalhe_json
FROM vis_sistema_health
ORDER BY coletado_em DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanColetaRows(rows)
}

func ListSistemaMetric(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, coletado_em, escopo, id_franqueado, chave, valor_num, valor_text, tags_json, metricas_json
FROM vis_sistema_metric
ORDER BY coletado_em DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanColetaRows(rows)
}

func scanColetaRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		dest := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range dest {
			ptrs[i] = &dest[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := make(map[string]any, len(cols))
		for i, col := range cols {
			m[col] = normalizeValue(dest[i])
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func parseRelatorioLimit(q urlQueryGetter, def int) int {
	if q == nil {
		return def
	}
	n, _ := strconv.Atoi(strings.TrimSpace(q.Get("limit")))
	if n < 1 {
		return def
	}
	return n
}

type urlQueryGetter interface {
	Get(string) string
}

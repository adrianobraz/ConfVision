// Pesquisa operacional Postgres (ConfVision). Uso: go run ./scripts/pg-audit
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	url := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if url == "" {
		url = loadPostgresURLFromEnvFile()
	}
	if url == "" {
		fmt.Fprintln(os.Stderr, "POSTGRES_URL ausente")
		os.Exit(2)
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ping: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK — Postgres conectado")

	runSection(db, "Câmera 15 (piloto)", `
SELECT id, ativo, bloqueado, deteccao_humano, analitico_pausado, worker_id,
       vis_mediamtx_node_id,
       CASE WHEN rtsp_url_sec IS NULL OR rtsp_url_sec = '' THEN 'vazio' ELSE 'ok' END AS rtsp,
       ultimo_ping_em, plano, status
FROM vis_camera WHERE id = 15`)

	runSection(db, "Sync Rust pilot-b (mesmas regras do Go)", `
SELECT id, nome, worker_id, analitico_pausado, vis_mediamtx_node_id
FROM vis_camera c
WHERE c.ativo = TRUE
  AND c.deteccao_humano = TRUE
  AND (c.analitico_pausado IS NOT TRUE)
  AND c.worker_id = 'rust-processor-pilot-b-02'
ORDER BY id`)

	runSection(db, "Contagem analíticas por worker_id (Rust)", `
SELECT worker_id, COUNT(*) AS n
FROM vis_camera
WHERE ativo = TRUE AND deteccao_humano = TRUE AND (analitico_pausado IS NOT TRUE)
  AND worker_id LIKE 'rust-processor%'
GROUP BY worker_id
ORDER BY worker_id`)

	runSection(db, "Últimos eventos câmera 15", `
SELECT id, created_at, tipo_deteccao, status, confianca
FROM vis_evento
WHERE vis_camera_id = 15
ORDER BY id DESC
LIMIT 8`)

	runSection(db, "Analíticas pausadas (excluídas do sync)", `
SELECT COUNT(*) AS pausadas FROM vis_camera
WHERE ativo = TRUE AND deteccao_humano = TRUE AND analitico_pausado IS TRUE`)

	runSection(db, "Pilot-a elegíveis", `
SELECT id, nome, status, LEFT(COALESCE(rtsp_url_sec,''),50) AS rtsp_prefix
FROM vis_camera
WHERE ativo = TRUE AND deteccao_humano = TRUE AND (analitico_pausado IS NOT TRUE)
  AND worker_id = 'rust-processor-pilot-a-01'
ORDER BY id`)

	runSection(db, "Pilot-b — todas elegíveis", `
SELECT id, nome, status, somente_armado, modo_deteccao, id_dispositivo, confianca_min, cooldown_seg
FROM vis_camera
WHERE ativo = TRUE AND deteccao_humano = TRUE AND (analitico_pausado IS NOT TRUE)
  AND worker_id = 'rust-processor-pilot-b-02'
ORDER BY id`)

	runSection(db, "Áreas ativas câmeras 15 e 21", `
SELECT vis_camera_id, COUNT(*) AS areas, SUM(CASE WHEN ativo IS TRUE THEN 1 ELSE 0 END) AS areas_ativas
FROM vis_camera_area
WHERE vis_camera_id IN (15, 21)
GROUP BY vis_camera_id`)

	runSection(db, "Eventos recentes (qualquer câmera Rust)", `
SELECT vis_camera_id, id, created_at, tipo_deteccao, status
FROM vis_evento
WHERE vis_camera_id IN (
  SELECT id FROM vis_camera WHERE worker_id LIKE 'rust-processor%'
)
ORDER BY id DESC
LIMIT 10`)

	runSection(db, "Coleta operacional (health recente)", `
SELECT componente, status, coletado_em
FROM vis_sistema_health
ORDER BY coletado_em DESC
LIMIT 6`)

	runSection(db, "vis_worker ping (Rust)", `
SELECT worker_id, cameras_ativas, ultimo_ping_em, ativo, cpu_percent, mem_percent
FROM vis_worker
WHERE worker_tipo = 'rust_processor'
ORDER BY ultimo_ping_em DESC NULLS LAST
LIMIT 6`)

	runSection(db, "Stream OK piloto-b (ultimo_stream_ok_em)", `
SELECT id, nome, status, ultimo_stream_ok_em, stream_falhas_consecutivas
FROM vis_camera
WHERE id IN (3, 5, 15, 21)
ORDER BY id`)
}

func runSection(db *sql.DB, title, query string) {
	fmt.Printf("\n=== %s ===\n", title)
	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query erro: %v\n", err)
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(cols, "\t"))

	vals := make([]any, len(cols))
	scan := make([]any, len(cols))
	for i := range vals {
		scan[i] = &vals[i]
	}
	n := 0
	for rows.Next() {
		if err := rows.Scan(scan...); err != nil {
			fmt.Fprintf(os.Stderr, "scan: %v\n", err)
			return
		}
		line := make([]string, len(cols))
		for i, v := range vals {
			line[i] = fmt.Sprint(v)
		}
		fmt.Fprintln(w, strings.Join(line, "\t"))
		n++
	}
	w.Flush()
	if n == 0 {
		fmt.Println("(sem linhas)")
	}
}

func loadPostgresURLFromEnvFile() string {
	root, _ := os.Getwd()
	for _, rel := range []string{".env", filepath.Join("..", ".env")} {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "POSTGRES_URL=") {
				return strings.TrimSpace(strings.TrimPrefix(line, "POSTGRES_URL="))
			}
		}
	}
	return ""
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	url := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if url == "" {
		fmt.Println("POSTGRES_URL ausente")
		os.Exit(2)
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== vis_evento ultimas 24h (amostra) ===")
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, vis_camera_id, status, origem, worker_id
FROM vis_evento
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY id DESC
LIMIT 10`)
	if err != nil {
		// fallback sem colunas opcionais
		rows, err = db.QueryContext(ctx, `
SELECT id, created_at, vis_camera_id, status
FROM vis_evento
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY id DESC LIMIT 10`)
	}
	if err != nil {
		fmt.Println("query vis_evento:", err)
		os.Exit(1)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
		var id, cam int
		var st string
		var created time.Time
		// try scan with optional cols
		var origem, worker sql.NullString
		if err := rows.Scan(&id, &created, &cam, &st, &origem, &worker); err != nil {
			_ = rows.Scan(&id, &created, &cam, &st)
			fmt.Printf("  id=%d cam=%d status=%s at=%s\n", id, cam, st, created.Format(time.RFC3339))
		} else {
			fmt.Printf("  id=%d cam=%d status=%s origem=%s worker=%s at=%s\n",
				id, cam, st, nullS(origem), nullS(worker), created.Format(time.RFC3339))
		}
	}
	if n == 0 {
		fmt.Println("  (nenhum)")
	}
}

func nullS(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "-"
}

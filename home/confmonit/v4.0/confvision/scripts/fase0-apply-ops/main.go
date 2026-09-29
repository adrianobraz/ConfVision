// Aplica sql/fase0_pilot_ops.sql no Postgres (POSTGRES_URL).
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	url := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if url == "" {
		fmt.Fprintln(os.Stderr, "POSTGRES_URL ausente")
		os.Exit(2)
	}
	sqlPath := os.Getenv("FASE0_SQL")
	if sqlPath == "" {
		// repo: core4-rust-pilot relativo ou env
		candidates := []string{
			`c:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor\sql\fase0_pilot_ops.sql`,
			filepath.Join("..", "..", "..", "..", "core4-rust-pilot", "confvision-rust-processor", "sql", "fase0_pilot_ops.sql"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				sqlPath = c
				break
			}
		}
	}
	if sqlPath == "" {
		fmt.Fprintln(os.Stderr, "defina FASE0_SQL apontando para fase0_pilot_ops.sql")
		os.Exit(2)
	}
	raw, err := os.ReadFile(sqlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read sql: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping: %v\n", err)
		os.Exit(1)
	}

	script := string(raw)
	start := strings.Index(strings.ToUpper(script), "BEGIN;")
	if start < 0 {
		start = strings.Index(strings.ToUpper(script), "BEGIN")
	}
	end := strings.LastIndex(strings.ToUpper(script), "COMMIT;")
	if start < 0 || end < 0 || end <= start {
		fmt.Fprintln(os.Stderr, "sql sem BEGIN/COMMIT")
		os.Exit(1)
	}
	txBody := script[start : end+len("COMMIT;")]

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "begin: %v\n", err)
		os.Exit(1)
	}
	for _, stmt := range splitSQL(txBody) {
		up := strings.ToUpper(strings.TrimSpace(stmt))
		if up == "BEGIN" || up == "COMMIT" || strings.HasPrefix(up, "--") || up == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			_ = tx.Rollback()
			fmt.Fprintf(os.Stderr, "exec failed: %v\nstmt: %.120s\n", err, stmt)
			os.Exit(1)
		}
		fmt.Println("OK:", strings.Split(strings.TrimSpace(stmt), "\n")[0])
	}
	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "commit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Fase0 SQL aplicado.")

	after := `
SELECT id, nome, worker_id, analitico_pausado
FROM vis_camera WHERE id IN (3,4,5,15) ORDER BY id`
	rows, err := db.QueryContext(ctx, after)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var nome, worker sql.NullString
		var paused sql.NullBool
		_ = rows.Scan(&id, &nome, &worker, &paused)
		fmt.Printf("  cam %d worker=%s pausado=%v %s\n", id, worker.String, paused.Bool, nome.String)
	}
}

func splitSQL(s string) []string {
	var out []string
	var b strings.Builder
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
		if strings.HasSuffix(trim, ";") {
			out = append(out, b.String())
			b.Reset()
		}
	}
	if t := strings.TrimSpace(b.String()); t != "" {
		out = append(out, t)
	}
	return out
}

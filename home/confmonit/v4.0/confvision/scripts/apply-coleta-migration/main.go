// Aplica coleta_operacional_migration.sql (POSTGRES_URL). Idempotente.
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
	sqlPath := os.Getenv("COLETA_SQL")
	if sqlPath == "" {
		sqlPath = filepath.Join("src", "modulos", "visdata", "coleta_operacional_migration.sql")
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping: %v\n", err)
		os.Exit(1)
	}

	for _, stmt := range splitSQL(string(raw)) {
		up := strings.ToUpper(strings.TrimSpace(stmt))
		if up == "" || strings.HasPrefix(up, "--") {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			fmt.Fprintf(os.Stderr, "exec failed: %v\nstmt: %.160s\n", err, stmt)
			os.Exit(1)
		}
		first := strings.Split(strings.TrimSpace(stmt), "\n")[0]
		fmt.Println("OK:", first)
	}
	fmt.Println("Coleta migration aplicada.")

	checks := []string{
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'vis_sistema_health'`,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'vis_sistema_metric'`,
		`SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'vis_camera' AND column_name = 'stream_motivo_pausa'`,
	}
	for _, q := range checks {
		var n int
		if err := db.QueryRowContext(ctx, q).Scan(&n); err != nil {
			fmt.Fprintf(os.Stderr, "verify: %v\n", err)
			os.Exit(1)
		}
		if n < 1 {
			fmt.Fprintf(os.Stderr, "verify falhou: %s\n", q)
			os.Exit(1)
		}
	}
	fmt.Println("Verify: vis_sistema_health, vis_sistema_metric, vis_camera.stream_motivo_pausa OK")
}

func splitSQL(s string) []string {
	var out []string
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
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

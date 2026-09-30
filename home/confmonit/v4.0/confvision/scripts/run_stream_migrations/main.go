// Aplica migrations de stream policy + coleta (vis_stream_relatorio).
// Uso (na raiz confvision): POSTGRES_URL=... go run ./scripts/run_stream_migrations/
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var migrationFiles = []string{
	"20260326_vis_camera_stream_policy.sql",
	"20260326_vis_camera_stream_error_diag.sql",
	"20260928_vis_coleta_operacional.sql",
}

func main() {
	url := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if len(os.Args) > 1 {
		url = strings.TrimSpace(os.Args[1])
	}
	if url == "" {
		fmt.Fprintln(os.Stderr, "usage: POSTGRES_URL=... go run ./scripts/run_stream_migrations/")
		os.Exit(1)
	}

	root, err := confvisionRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	for _, name := range migrationFiles {
		path := filepath.Join(root, "sql", "migrations", name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("=== %s ===\n", name)
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			fmt.Fprintf(os.Stderr, "apply %s: %v\n", name, err)
			os.Exit(1)
		}
	}

	var colCount int
	err = conn.QueryRow(ctx, `
SELECT COUNT(*) FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'vis_camera' AND column_name LIKE 'stream_%'`).Scan(&colCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify columns: %v\n", err)
		os.Exit(1)
	}
	var relTable *string
	err = conn.QueryRow(ctx, `SELECT to_regclass('public.vis_stream_relatorio')::text`).Scan(&relTable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify relatorio: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK: stream columns on vis_camera = %d; vis_stream_relatorio = %v\n", colCount, derefStr(relTable))
}

func confvisionRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "sql", "migrations")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("nao encontrei sql/migrations a partir de %s", wd)
}

func derefStr(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// One-off: go run ./scripts/migrate-coleta (POSTGRES_URL no .env ou ambiente)
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	root, _ := os.Getwd()
	sqlPath := filepath.Join(root, "sql", "migrations", "20260928_vis_coleta_operacional.sql")
	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ler migration: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ping postgres: %v\n", err)
		os.Exit(1)
	}

	if _, err := db.Exec(string(sqlBytes)); err != nil {
		fmt.Fprintf(os.Stderr, "exec migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK — migration 20260928_vis_coleta_operacional aplicada.")
}

func loadPostgresURLFromEnvFile() string {
	for _, name := range []string{".env", filepath.Join("..", ".env")} {
		b, err := os.ReadFile(name)
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

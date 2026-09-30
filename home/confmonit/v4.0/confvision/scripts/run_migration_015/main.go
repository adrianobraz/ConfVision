// One-off: apply 015 dupla_comunicacao migration. Usage:
// go run ./scripts/run_migration_015 "postgres://..."
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	url := os.Getenv("POSTGRES_URL")
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	if url == "" {
		fmt.Fprintln(os.Stderr, "usage: POSTGRES_URL=... go run . OR go run . <postgres-url>")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	sql := `ALTER TABLE vis_integracao_config ADD COLUMN IF NOT EXISTS dupla_comunicacao BOOLEAN NOT NULL DEFAULT FALSE`
	if _, err := conn.Exec(ctx, sql); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	var colExists bool
	err = conn.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_name = 'vis_integracao_config' AND column_name = 'dupla_comunicacao'
)`).Scan(&colExists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(1)
	}

	if colExists {
		fmt.Println("OK: coluna dupla_comunicacao existe em vis_integracao_config")
	} else {
		fmt.Fprintln(os.Stderr, "ERRO: coluna nao encontrada apos migration")
		os.Exit(1)
	}

	var total int
	_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM vis_integracao_config`).Scan(&total)
	fmt.Printf("vis_integracao_config: %d registro(s)\n", total)

	rows, err := conn.Query(ctx, `
SELECT id_franqueado, sistema, ativo, COALESCE(dupla_comunicacao, false)
FROM vis_integracao_config ORDER BY updated_at DESC LIMIT 5`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var idFra, sistema string
			var ativo, dupla bool
			if rows.Scan(&idFra, &sistema, &ativo, &dupla) == nil {
				fmt.Printf("  franqueado=%s sistema=%s ativo=%v dupla=%v\n", idFra, sistema, ativo, dupla)
			}
		}
	}
}

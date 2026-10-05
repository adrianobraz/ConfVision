// One-off: apply 019 integracao inbound migration.
// Usage: go run . "postgres://..."
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

	sql := `
ALTER TABLE vis_integracao_config
    ADD COLUMN IF NOT EXISTS webhook_inbound_ativo BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS codigo_evento_armar TEXT,
    ADD COLUMN IF NOT EXISTS codigo_evento_desarmar TEXT`
	if _, err := conn.Exec(ctx, sql); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	var count int
	err = conn.QueryRow(ctx, `
SELECT COUNT(*) FROM information_schema.columns
WHERE table_name = 'vis_integracao_config'
  AND column_name IN ('webhook_inbound_ativo', 'codigo_evento_armar', 'codigo_evento_desarmar')`).Scan(&count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(1)
	}
	if count < 3 {
		fmt.Fprintf(os.Stderr, "verify: esperado 3 colunas, encontrado %d\n", count)
		os.Exit(1)
	}

	fmt.Println("OK: migration 019_integracao_inbound aplicada")
}

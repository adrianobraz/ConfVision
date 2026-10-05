// One-off: apply 020 grupos visualizacao migration.
// Usage: POSTGRES_URL=... go run . OR go run . <postgres-url>
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	sql := `
CREATE TABLE IF NOT EXISTS vis_grupo_visualizacao (
    id              SERIAL PRIMARY KEY,
    id_franqueado   VARCHAR(64) NOT NULL,
    nome            VARCHAR(255) NOT NULL,
    descricao       TEXT,
    ativo           BOOLEAN NOT NULL DEFAULT TRUE,
    layout_mosaic   JSONB NOT NULL DEFAULT '{"modo":"auto"}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_fra ON vis_grupo_visualizacao (id_franqueado);
CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_ativo ON vis_grupo_visualizacao (id_franqueado, ativo);

CREATE TABLE IF NOT EXISTS vis_grupo_visualizacao_cliente (
    id          SERIAL PRIMARY KEY,
    grupo_id    INT NOT NULL REFERENCES vis_grupo_visualizacao (id) ON DELETE CASCADE,
    id_cliente  VARCHAR(64) NOT NULL,
    UNIQUE (grupo_id, id_cliente)
);

CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cli ON vis_grupo_visualizacao_cliente (id_cliente);
CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cli_grupo ON vis_grupo_visualizacao_cliente (grupo_id);

CREATE TABLE IF NOT EXISTS vis_grupo_visualizacao_camera (
    id              SERIAL PRIMARY KEY,
    grupo_id        INT NOT NULL REFERENCES vis_grupo_visualizacao (id) ON DELETE CASCADE,
    id_cliente      VARCHAR(64) NOT NULL,
    vis_camera_id   INT NOT NULL REFERENCES vis_camera (id) ON DELETE CASCADE,
    ordem           INT NOT NULL DEFAULT 0,
    ativo           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (grupo_id, vis_camera_id)
);

CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cam_grupo ON vis_grupo_visualizacao_camera (grupo_id, ordem);
CREATE INDEX IF NOT EXISTS idx_vis_grupo_vis_cam_cli ON vis_grupo_visualizacao_camera (grupo_id, id_cliente)`

	if _, err := conn.Exec(ctx, sql); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	var n int
	err = conn.QueryRow(ctx, `
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'vis_grupo_visualizacao',
    'vis_grupo_visualizacao_cliente',
    'vis_grupo_visualizacao_camera'
  )`).Scan(&n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(1)
	}
	if n < 3 {
		fmt.Fprintf(os.Stderr, "verify: esperado 3 tabelas, encontrado %d\n", n)
		os.Exit(1)
	}

	fmt.Println("OK: migration 020_grupos_visualizacao aplicada")
}

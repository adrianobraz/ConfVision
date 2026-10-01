// Aplica SQL no Postgres central ConfVision.
// Uso:
//   POSTGRES_URL=postgres://... apply.exe
//   POSTGRES_URL=postgres://... apply.exe ../../003_import_xano.sql
//   POSTGRES_URL=postgres://... apply.exe -check
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

func main() {
	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "Defina POSTGRES_URL no ambiente ou .env")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conexao falhou: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Ping falhou: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "-check-atendimento" {
		var cfg, pol, ia, autofim int
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM ops_cliente_atendimento_config`).Scan(&cfg)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM ops_franqueado_atendimento_politica`).Scan(&pol)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM ops_cliente_atendimento_config WHERE inteligencia_artificial`).Scan(&ia)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM ops_cliente_atendimento_config WHERE finalizacao_automatica`).Scan(&autofim)
		fmt.Printf("ops_cliente_atendimento_config=%d (ia=%d autofim=%d)\n", cfg, ia, autofim)
		fmt.Printf("ops_franqueado_atendimento_politica=%d\n", pol)
		idFra := os.Getenv("ID_FRA")
		if idFra != "" {
			var iaPol, autofimPol bool
			var iaModo, autofimModo string
			err := conn.QueryRow(ctx, `SELECT inteligencia_artificial, ia_modo, finalizacao_automatica, autofim_modo FROM ops_franqueado_atendimento_politica WHERE id_franqueado=$1`, idFra).Scan(&iaPol, &iaModo, &autofimPol, &autofimModo)
			if err == nil {
				fmt.Printf("politica_fra ia=%v modo=%s autofim=%v modo=%s\n", iaPol, iaModo, autofimPol, autofimModo)
			} else {
				fmt.Println("politica_fra: (sem linha, fail-open na #22)")
			}
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "-check" {
		var nodes, cameras, areas, licencas, gradeSlots, armeJanelas, autofimQueue int
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM vis_mediamtx_node`).Scan(&nodes)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM vis_camera`).Scan(&cameras)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM vis_camera_area`).Scan(&areas)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM vis_licenca`).Scan(&licencas)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM vis_cliente_grade_slot`).Scan(&gradeSlots)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM ops_arme_janela`).Scan(&armeJanelas)
		_ = conn.QueryRow(ctx, `SELECT COUNT(*) FROM ops_bot_finalizaeventoauto`).Scan(&autofimQueue)
		fmt.Printf("vis_mediamtx_node: %d\n", nodes)
		fmt.Printf("vis_camera: %d\n", cameras)
		fmt.Printf("vis_camera_area: %d\n", areas)
		fmt.Printf("vis_licenca: %d\n", licencas)
		fmt.Printf("vis_cliente_grade_slot: %d\n", gradeSlots)
		fmt.Printf("ops_arme_janela: %d\n", armeJanelas)
		fmt.Printf("ops_bot_finalizaeventoauto: %d\n", autofimQueue)
		return
	}

	sqlFile := filepath.Join("..", "..", "002_central_schema.sql")
	if len(os.Args) > 1 {
		sqlFile = os.Args[1]
	}
	sqlFile, _ = filepath.Abs(sqlFile)

	body, err := os.ReadFile(sqlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler %s: %v\n", sqlFile, err)
		os.Exit(1)
	}
	sqlText := strings.TrimPrefix(string(body), "\ufeff")

	fmt.Printf("Conectado. Aplicando %s ...\n", sqlFile)
	if _, err := conn.Exec(ctx, sqlText); err != nil {
		fmt.Fprintf(os.Stderr, "Exec falhou: %v\n", err)
		os.Exit(1)
	}

	var tables []string
	rows, err := conn.Query(ctx, `
SELECT tablename FROM pg_tables
WHERE schemaname = 'public' AND tablename LIKE 'vis_%'
ORDER BY tablename`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			if rows.Scan(&name) == nil {
				tables = append(tables, name)
			}
		}
	}

	fmt.Println("SQL aplicado com sucesso.")
	if len(tables) > 0 {
		fmt.Println("Tabelas vis_*:", tables)
	}
}

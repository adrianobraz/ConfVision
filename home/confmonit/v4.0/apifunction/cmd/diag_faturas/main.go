package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	url := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if url == "" {
		url = "postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	ctx := context.Background()

	fmt.Println("=== ops_parceiro_ativacao ===")
	q1(ctx, db, `SELECT id, id_parceiro, fp_fatura_id, status, qtd_clientes, preco_unitario, ciclo_ref
FROM ops_parceiro_ativacao WHERE fp_fatura_id IN (6,8) OR id_franqueado='2026072204185539042285696' ORDER BY id`)

	fmt.Println("=== ops_parceiro_excecao ===")
	q1(ctx, db, `SELECT id, id_parceiro, id_parceiro_anterior, id_vinculo_anterior, fp_fatura_id, status, ciclo_ref
FROM ops_parceiro_excecao WHERE fp_fatura_id IN (6,8) OR id_franqueado='2026072204185539042285696' ORDER BY id`)
}

func q1(ctx context.Context, db *sql.DB, q string) {
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		_ = rows.Scan(ptrs...)
		parts := make([]string, len(cols))
		for i, c := range cols {
			parts[i] = fmt.Sprintf("%s=%v", c, vals[i])
		}
		fmt.Println(strings.Join(parts, " | "))
	}
}

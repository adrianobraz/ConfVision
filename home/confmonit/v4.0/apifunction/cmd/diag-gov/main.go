package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dsn := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if dsn == "" {
		log.Fatal("POSTGRES_URL obrigatorio")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	fmt.Println("=== fp_gov_restricao (ativo=true) ===")
	printRows(db, `
		SELECT entidade_tipo, entidade_id, fatura_id, nivel, vencimento_em, dias_restantes, ativo, updated_at
		FROM fp_gov_restricao
		WHERE ativo = TRUE
		ORDER BY updated_at DESC
	`)

	fmt.Println("\n=== fp_gov_repasse (status=aberta) ===")
	printRows(db, `
		SELECT fatura_id, tipo, id_representante, id_central, status, vencimento_em, valor_total, synced_at
		FROM fp_gov_repasse
		WHERE status = 'aberta'
		ORDER BY vencimento_em ASC NULLS LAST
	`)

	fmt.Println("\n=== fp_gov_repasse (fatura_id=121 ou relacionados) ===")
	printRows(db, `
		SELECT fatura_id, tipo, id_representante, id_central, status, vencimento_em, valor_total
		FROM fp_gov_repasse
		WHERE fatura_id = 121 OR fatura_id IN (
			SELECT fatura_id FROM fp_gov_restricao WHERE ativo = TRUE
		)
		ORDER BY fatura_id
	`)

	fmt.Println("\n=== fp_gov_suspensao_cascata (ativo=true) ===")
	printRows(db, `
		SELECT id, id_franqueado, assinatura_id, id_representante, fatura_repasse_id, motivo_interno, ativo, suspenso_em
		FROM fp_gov_suspensao_cascata
		WHERE ativo = TRUE
		ORDER BY suspenso_em DESC
		LIMIT 20
	`)

	filter := strings.TrimSpace(os.Getenv("ID_REP"))
	if filter != "" {
		fmt.Printf("\n=== filtro id_representante=%s ===\n", filter)
		printRows(db, `
			SELECT entidade_tipo, entidade_id, fatura_id, nivel, dias_restantes, ativo
			FROM fp_gov_restricao
			WHERE ativo = TRUE AND entidade_tipo = 'REP' AND entidade_id = $1
		`, filter)
		printRows(db, `
			SELECT fatura_id, tipo, status, vencimento_em, valor_total, id_central
			FROM fp_gov_repasse
			WHERE id_representante = $1
			ORDER BY fatura_id DESC
		`, filter)
	}
}

func printRows(db *sql.DB, query string, args ...any) {
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("query erro: %v", err)
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		row := map[string]any{}
		for i, c := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[c] = string(v)
			default:
				row[c] = v
			}
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		fmt.Println("(nenhum registro)")
		return
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

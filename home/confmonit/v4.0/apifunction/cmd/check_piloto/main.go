// One-off: consulta e reparo pago_ate cs_parceiro_vinculo no piloto.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	url := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if url == "" {
		url = "postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
	}
	idFra := "2026072204185539042285696"
	if len(os.Args) > 1 {
		idFra = os.Args[1]
	}

	db, err := sql.Open("postgres", url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	ctx := context.Background()

	fmt.Println("=== Antes ===")
	printRows(ctx, db, idFra)

	rows, err := db.QueryContext(ctx, `
SELECT e.id_vinculo_ativo, COALESCE(e.pago_ate::text, '')
FROM ops_parceiro_excecao e
WHERE e.id_franqueado = $1 AND e.status = 'paga'
  AND e.id_vinculo_ativo IS NOT NULL AND TRIM(e.id_vinculo_ativo) <> ''`, idFra)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var vinculo, pagoAte string
		if err := rows.Scan(&vinculo, &pagoAte); err != nil {
			continue
		}
		vinculo = strings.TrimSpace(vinculo)
		if vinculo == "" {
			continue
		}
		ate := pagoAte
		if strings.TrimSpace(ate) == "" {
			ate = time.Now().AddDate(0, 0, 14).Format("2006-01-02")
		}
		res, err := db.ExecContext(ctx, `
UPDATE fp_servico_cobranca
SET pago_ate = $3::date, status_ciclo = 'normal', updated_at = NOW()
WHERE id_franqueado = $1 AND ref_tipo = 'cs_parceiro_vinculo' AND ref_id = $2 AND ativo = TRUE
  AND (pago_ate IS NULL OR pago_ate < $3::date)`, idFra, vinculo, ate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update %s: %v\n", vinculo, err)
			continue
		}
		n, _ := res.RowsAffected()
		fmt.Printf("backfill vinculo=%s pago_ate=%s rows=%d\n", vinculo, ate, n)
	}

	fmt.Println("\n=== Depois ===")
	printRows(ctx, db, idFra)
}

func printRows(ctx context.Context, db *sql.DB, idFra string) {
	rows, err := db.QueryContext(ctx, `
SELECT e.id, e.status, e.pago_ate, e.id_vinculo_ativo,
       s.ref_id, s.pago_ate, s.ativo
FROM ops_parceiro_excecao e
LEFT JOIN fp_servico_cobranca s
  ON s.id_franqueado = e.id_franqueado
 AND s.ref_tipo = 'cs_parceiro_vinculo'
 AND s.ref_id = e.id_vinculo_ativo
WHERE e.id_franqueado = $1
ORDER BY e.id DESC`, idFra)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var st string
		var exPago, vinculo, refID sql.NullString
		var svcPago sql.NullTime
		var ativo sql.NullBool
		_ = rows.Scan(&id, &st, &exPago, &vinculo, &refID, &svcPago, &ativo)
		svc := ""
		if svcPago.Valid {
			svc = svcPago.Time.Format("2006-01-02")
		}
		fmt.Printf("excecao=%d status=%s ex_pago=%s vinculo=%s svc_pago=%s ativo=%v\n",
			id, st, exPago.String, vinculo.String, svc, ativo.Bool)
	}
}

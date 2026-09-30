// Gera SQL INSERT vis_licenca a partir de JSON do Xano (fp_confvision_resumo_franqueado).
// go run . ../../tmp_xano_res.json > ../../sql/sync_franqueado_test.sql
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	path := "tmp_xano_res.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	lics, _ := resp["licencas"].([]any)
	fmt.Println("-- Sync vis_licenca Xano -> Postgres (gerado automaticamente)")
	fmt.Println("BEGIN;")
	for _, el := range lics {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		fmt.Println(upsertSQL(m))
	}
	fmt.Println(`SELECT setval(pg_get_serial_sequence('vis_licenca','id'), COALESCE((SELECT MAX(id) FROM vis_licenca), 1));`)
	fmt.Println("COMMIT;")
}

func upsertSQL(m map[string]any) string {
	id := intAny(m["id"])
	created := tsAny(m["created_at"])
	pago := tsAny(m["pago_em"])
	valido := tsAny(m["valido_ate"])
	idFra := strAny(m["id_franqueado"])
	plano := strAny(m["plano"])
	unidade := strAny(m["unidade"])
	if unidade == "" {
		unidade = "camera"
	}
	valor := floatAny(m["valor"])
	status := strAny(m["status"])
	if status == "" {
		status = "pendente"
	}
	obs := strings.ReplaceAll(strAny(m["observacao"]), "'", "''")
	idFat := strAny(m["id_fatura"])
	idPag := strAny(m["id_pagamento"])

	return fmt.Sprintf(`INSERT INTO vis_licenca (id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate, status, id_fatura, id_pagamento, observacao)
VALUES (%d, %s, '%s', '%s', '%s', %.2f, %s, %s, '%s', %s, %s, '%s')
ON CONFLICT (id) DO UPDATE SET
  id_franqueado=EXCLUDED.id_franqueado, plano=EXCLUDED.plano, unidade=EXCLUDED.unidade,
  valor=EXCLUDED.valor, pago_em=EXCLUDED.pago_em, valido_ate=EXCLUDED.valido_ate,
  status=EXCLUDED.status, id_fatura=EXCLUDED.id_fatura, id_pagamento=EXCLUDED.id_pagamento,
  observacao=EXCLUDED.observacao;`,
		id, sqlTs(created), idFra, plano, unidade, valor,
		sqlTsNullable(pago), sqlTsNullable(valido), status,
		sqlText(idFat), sqlText(idPag), obs)
}

func intAny(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	}
	return 0
}

func floatAny(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func strAny(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func tsAny(v any) time.Time {
	switch t := v.(type) {
	case float64:
		if t > 1e12 {
			return time.UnixMilli(int64(t)).UTC()
		}
		return time.Unix(int64(t), 0).UTC()
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
			if n > 1e12 {
				return time.UnixMilli(n).UTC()
			}
			return time.Unix(n, 0).UTC()
		}
	}
	return time.Time{}
}

func sqlTs(t time.Time) string {
	if t.IsZero() {
		return "NOW()"
	}
	return "'" + t.Format("2006-01-02 15:04:05+00") + "'"
}

func sqlTsNullable(t time.Time) string {
	if t.IsZero() {
		return "NULL"
	}
	return sqlTs(t)
}

func sqlText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "NULL"
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

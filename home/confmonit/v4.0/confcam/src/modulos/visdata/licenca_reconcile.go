package visdata

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"confvision/src/xanopro"
)

type ativarLicencaPagamentoHook func(ctx context.Context, licID, faturaID, pagamentoID int, pagoEm time.Time) (map[string]any, error)

// ativarLicencaPagamentoFn permite mock em testes.
var ativarLicencaPagamentoFn ativarLicencaPagamentoHook = AtivarLicencaPagamento

// ReconciliarLicencasPagasFranqueado ativa vis_licenca pendente ligada a faturas ConfVision
// ja pagas no Xano (leitura fp_fatura_minhas + ativacao local no Postgres).
func ReconciliarLicencasPagasFranqueado(ctx context.Context, idFranqueado string) (int, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return 0, fmt.Errorf("id_franqueado obrigatorio")
	}

	faturas, err := fetchFaturasPagasComItens(idFranqueado)
	if err != nil {
		return 0, err
	}
	return aplicarAtivacaoLicencasDeFaturas(ctx, faturas)
}

// ReconciliarLicencasPagasTodosPendentes percorre franqueados com licenca pendente no Postgres.
func ReconciliarLicencasPagasTodosPendentes(ctx context.Context) (map[string]int, error) {
	ids, err := ListFranqueadosComLicencaStatus(ctx, "pendente")
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(ids))
	for _, idFra := range ids {
		n, err := ReconciliarLicencasPagasFranqueado(ctx, idFra)
		if err != nil {
			out[idFra] = -1
			continue
		}
		out[idFra] = n
	}
	return out, nil
}

// ListFranqueadosComLicencaStatus retorna id_franqueado distintos com licencas no status informado.
func ListFranqueadosComLicencaStatus(ctx context.Context, status string) ([]string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return nil, fmt.Errorf("status obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT DISTINCT id_franqueado FROM vis_licenca
WHERE status = $1 AND TRIM(id_franqueado) <> ''
ORDER BY id_franqueado`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, rows.Err()
}

// DiagnoseReconciliarLicencas resume faturas pagas e itens vis_licenca (sem alterar Postgres).
func DiagnoseReconciliarLicencas(idFranqueado string) (map[string]any, error) {
	faturas, err := fetchFaturasPagasComItens(idFranqueado)
	if err != nil {
		return nil, err
	}
	cvFat := 0
	itensLic := 0
	for _, fat := range faturas {
		tipo := strings.TrimSpace(strValMap(fat, "tipo"))
		if tipo != "" && !strings.Contains(tipo, "confvision") {
			continue
		}
		cvFat++
		for _, item := range mapSliceAny(fat["itens"]) {
			if strings.TrimSpace(strValMap(item, "ref_tipo")) == "vis_licenca" {
				itensLic++
			}
		}
	}
	return map[string]any{
		"faturas_pagas_total":      len(faturas),
		"faturas_confvision":       cvFat,
		"itens_vis_licenca":       itensLic,
		"incluir_itens_necessario": itensLic == 0 && cvFat > 0,
	}, nil
}

func fetchFaturasPagasComItens(idFranqueado string) ([]map[string]any, error) {
	raw, err := xanopro.Post("/fp_fatura_minhas", map[string]any{
		"id_franqueado": idFranqueado,
		"status":        "paga",
		"incluir_itens": true,
	})
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return mapSliceAny(resp["dados"]), nil
}

// ReconciliarLicencasPagasFranqueadoVerbose igual ao reconciliar padrao, retornando erros por licenca.
func ReconciliarLicencasPagasFranqueadoVerbose(ctx context.Context, idFranqueado string) (int, []string, error) {
	faturas, err := fetchFaturasPagasComItens(idFranqueado)
	if err != nil {
		return 0, nil, err
	}
	n, errs := aplicarAtivacaoLicencasDeFaturasVerbose(ctx, faturas)
	return n, errs, nil
}

func aplicarAtivacaoLicencasDeFaturas(ctx context.Context, faturas []map[string]any) (int, error) {
	n, _ := aplicarAtivacaoLicencasDeFaturasVerbose(ctx, faturas)
	return n, nil
}

func aplicarAtivacaoLicencasDeFaturasVerbose(ctx context.Context, faturas []map[string]any) (int, []string) {
	faturasCV := make([]map[string]any, 0, len(faturas))
	for _, fat := range faturas {
		tipo := strings.TrimSpace(strValMap(fat, "tipo"))
		if tipo != "" && !strings.Contains(tipo, "confvision") {
			continue
		}
		faturasCV = append(faturasCV, fat)
	}
	sort.Slice(faturasCV, func(i, j int) bool {
		return intFromAny(faturasCV[i]["id"]) < intFromAny(faturasCV[j]["id"])
	})

	pendentesPorFranq := map[string][]int{}
	ativadas := 0
	var errs []string

	for _, fat := range faturasCV {
		idFra := strings.TrimSpace(strValMap(fat, "id_franqueado"))
		faturaID := intFromAny(fat["id"])
		pagoEm := parseTimeAny(fat["pago_em"])
		if pagoEm.IsZero() {
			pagoEm = time.Now().UTC()
		}

		if idFra != "" && pendentesPorFranq[idFra] == nil {
			ids, err := listPendenteLicencaIDsFn(ctx, idFra)
			if err != nil {
				errs = append(errs, fmt.Sprintf("fatura=%d: listar pendentes: %v", faturaID, err))
				continue
			}
			pendentesPorFranq[idFra] = ids
		}
		fila := pendentesPorFranq[idFra]

		for _, item := range mapSliceAny(fat["itens"]) {
			if strings.TrimSpace(strValMap(item, "ref_tipo")) != "vis_licenca" {
				continue
			}
			licID := intFromAny(item["ref_id"])
			if licID <= 0 && len(fila) > 0 {
				licID = fila[0]
				fila = fila[1:]
			}
			if licID <= 0 {
				errs = append(errs, fmt.Sprintf("fatura=%d: item vis_licenca sem ref_id e sem pendente disponivel", faturaID))
				continue
			}
			if _, err := ativarLicencaPagamentoFn(ctx, licID, faturaID, 0, pagoEm); err != nil {
				errs = append(errs, fmt.Sprintf("licenca=%d fatura=%d: %v", licID, faturaID, err))
				continue
			}
			ativadas++
		}
		pendentesPorFranq[idFra] = fila
	}
	return ativadas, errs
}

var listPendenteLicencaIDsFn = listPendenteLicencaIDs

func listPendenteLicencaIDs(ctx context.Context, idFranqueado string) ([]int, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id FROM vis_licenca
WHERE id_franqueado = $1 AND status = 'pendente'
ORDER BY id ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func strValMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(v)
	}
}

func mapSliceAny(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, el := range t {
			if m, ok := el.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return t
	default:
		return nil
	}
}

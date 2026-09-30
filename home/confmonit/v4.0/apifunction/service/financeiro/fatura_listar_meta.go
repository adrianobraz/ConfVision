package financeiro

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"apifunction/auth"
	"apifunction/config"
	"apifunction/xano"
)

const (
	metaTableFatura     = 121
	metaTableFaturaItem = 122
)

func listarFaturasMeta(ctx context.Context, sess auth.SessaoAdm, f ListarFiltro) ([]Fatura, error) {
	cli := xano.NewMeta(config.XanoMetaBaseURL, config.XanoMetaAccessToken, config.XanoMetaWorkspaceID)
	if cli == nil || !cli.Enabled() {
		return nil, errMetaNaoConfigurado
	}

	idCen := idCentralSessao(sess)
	cenKeys, err := chavesCentral(ctx, idCen)
	if err != nil {
		return nil, err
	}
	reps, err := repsDaCentral(ctx, idCen)
	if err != nil {
		return nil, err
	}
	if sess.UserTipo == "REP" {
		reps = []string{strings.TrimSpace(sess.IDRepresentante)}
	} else if f.IDRepresentante != "" {
		reps = []string{strings.TrimSpace(f.IDRepresentante)}
	}
	fraIDs, err := franqueadosDaCentral(ctx, reps)
	if err != nil {
		return nil, err
	}

	rawFaturas, err := cli.ListTableContent(metaTableFatura)
	if err != nil {
		return nil, err
	}

	repSet := toStringSet(reps)
	fraSet := toStringSet(fraIDs)

	var out []Fatura
	for _, row := range rawFaturas {
		fat := faturaFromMetaRow(row)
		if !faturaNoEscopo(fat, cenKeys, repSet, fraSet) {
			continue
		}
		if st := strings.TrimSpace(f.Status); st != "" && strings.TrimSpace(fat.Status) != st {
			continue
		}
		if tp := strings.TrimSpace(f.Tipo); tp != "" && strings.TrimSpace(fat.Tipo) != tp {
			continue
		}
		if idf := strings.TrimSpace(f.IDFranqueado); idf != "" && strings.TrimSpace(fat.IDFranqueado) != idf {
			continue
		}
		out = append(out, fat)
	}

	sort.Slice(out, func(i, j int) bool {
		pi := statusPriority(out[i].Status)
		pj := statusPriority(out[j].Status)
		if pi != pj {
			return pi < pj
		}
		return out[i].ID > out[j].ID
	})

	if len(out) == 0 {
		return out, nil
	}

	rawItens, err := cli.ListTableContent(metaTableFaturaItem)
	if err != nil {
		return nil, err
	}
	ids := map[int64]struct{}{}
	for _, fat := range out {
		ids[fat.ID] = struct{}{}
	}
	itensMap := map[int64][]FaturaItem{}
	for _, row := range rawItens {
		fid := int64FromAny(row["fp_fatura_id"])
		if _, ok := ids[fid]; !ok {
			continue
		}
		itensMap[fid] = append(itensMap[fid], faturaItemFromMetaRow(row))
	}
	for i := range out {
		out[i].Itens = itensMap[out[i].ID]
	}
	return out, nil
}

var errMetaNaoConfigurado = errorString("configure XANO_META_ACCESS_TOKEN no apifunction")

type errorString string

func (e errorString) Error() string { return string(e) }

func toStringSet(vals []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" {
			out[v] = struct{}{}
		}
	}
	return out
}

func faturaNoEscopo(fat Fatura, cenKeys map[string]struct{}, repSet, fraSet map[string]struct{}) bool {
	idCentral := strings.TrimSpace(fat.IDCentral)
	idRep := strings.TrimSpace(fat.IDRepresentante)
	idFra := strings.TrimSpace(fat.IDFranqueado)

	if idCentral != "" && len(cenKeys) > 0 {
		if _, ok := cenKeys[idCentral]; ok {
			return true
		}
	}
	if idRep != "" {
		if _, ok := repSet[idRep]; ok {
			return true
		}
	}
	if idFra != "" {
		if _, ok := fraSet[idFra]; ok {
			return true
		}
	}
	return false
}

func statusPriority(status string) int {
	switch strings.TrimSpace(status) {
	case "aberta":
		return 0
	case "paga":
		return 1
	case "cancelada":
		return 2
	default:
		return 9
	}
}

func faturaFromMetaRow(row map[string]any) Fatura {
	fat := Fatura{
		ID:                  int64FromAny(row["id"]),
		IDFranqueado:        stringFromAny(row["id_franqueado"]),
		IDRepresentante:     stringFromAny(row["id_representante"]),
		IDCentral:           stringFromAny(row["id_central"]),
		Referencia:          stringFromAny(row["referencia"]),
		Status:              stringFromAny(row["status"]),
		Tipo:                stringFromAny(row["tipo"]),
		ValorTotal:          floatFromAny(row["valor_total"]),
		ValorPisoCentral:    floatFromAny(row["valor_piso_central"]),
		ValorPisoBreakglass: floatFromAny(row["valor_piso_breakglass"]),
		MargemCentral:       floatFromAny(row["margem_central"]),
		MargemRep:           floatFromAny(row["margem_rep"]),
		FaturaOrigemID:      int64FromAny(row["fatura_origem_id"]),
		CicloRef:            stringFromAny(row["ciclo_ref"]),
		Observacao:          stringFromAny(row["observacao"]),
	}
	if t := timeFromAny(row["created_at"]); t != nil {
		fat.CreatedAt = t
	}
	if t := timeFromAny(row["vencimento_em"]); t != nil {
		fat.VencimentoEm = t
	}
	if t := timeFromAny(row["pago_em"]); t != nil {
		fat.PagoEm = t
	}
	return fat
}

func faturaItemFromMetaRow(row map[string]any) FaturaItem {
	return FaturaItem{
		ID:            int64FromAny(row["id"]),
		FPFaturaID:    int64FromAny(row["fp_fatura_id"]),
		Descricao:     stringFromAny(row["descricao"]),
		Quantidade:    intFromAny(row["quantidade"]),
		ValorUnitario: floatFromAny(row["valor_unitario"]),
		ValorTotal:    floatFromAny(row["valor_total"]),
		RefTipo:       stringFromAny(row["ref_tipo"]),
		RefID:         stringFromAny(row["ref_id"]),
		ValorPiso:     floatFromAny(row["valor_piso"]),
		MargemCentral: floatFromAny(row["margem_central"]),
		MargemRep:     floatFromAny(row["margem_rep"]),
	}
}

func stringFromAny(v any) string {
	switch n := v.(type) {
	case string:
		return strings.TrimSpace(n)
	case json.Number:
		return n.String()
	case float64:
		if n == float64(int64(n)) {
			return strings.TrimSpace(formatInt(int64(n)))
		}
		return strings.TrimSpace(formatFloat(n))
	case int:
		return formatInt(int64(n))
	case int64:
		return formatInt(n)
	default:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func int64FromAny(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

func intFromAny(v any) int {
	return int(int64FromAny(v))
}

func floatFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func timeFromAny(v any) *time.Time {
	switch n := v.(type) {
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return nil
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05+0000", "2006-01-02 15:04:05-0700"} {
			if t, err := time.Parse(layout, s); err == nil {
				return &t
			}
		}
	case float64:
		return msToTime(int64(n))
	case int:
		return msToTime(int64(n))
	case int64:
		return msToTime(n)
	case json.Number:
		i, _ := n.Int64()
		return msToTime(i)
	}
	return nil
}

func msToTime(ms int64) *time.Time {
	if ms <= 0 {
		return nil
	}
	t := time.UnixMilli(ms).UTC()
	return &t
}

func formatInt(v int64) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(jsonNumber(v), `"`, ""), " ", ""))
}

func formatFloat(v float64) string {
	b, _ := json.Marshal(v)
	return strings.Trim(string(b), `"`)
}

func jsonNumber(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

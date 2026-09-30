package pgcatalogo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"apifunction/config"
	"apifunction/xano"
)

// SyncXanoResult resume espelhamento Postgres -> Xano (assinaturas / FranqueadoPro).
type SyncXanoResult struct {
	PacotesAtualizados  int      `json:"pacotes_atualizados"`
	ProdutosAtualizados int      `json:"produtos_atualizados"`
	Avisos              []string `json:"avisos,omitempty"`
}

func SyncXanoEnabled() bool {
	cli := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	return cli != nil && cli.Enabled() && strings.TrimSpace(config.WorkerAdminToken) != ""
}

func xanoFinanceiro() *xano.Client {
	return xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
}

func adminPayload(idCentral string, extra map[string]any) map[string]any {
	p := map[string]any{
		"admin_token": config.WorkerAdminToken,
		"id_central":  strings.TrimSpace(idCentral),
	}
	for k, v := range extra {
		p[k] = v
	}
	return p
}

// SyncCentralToXano alinha fp_pacote_cota e fp_produto_catalogo no Xano com o Postgres da Central.
func SyncCentralToXano(ctx context.Context, idCentral string) (SyncXanoResult, error) {
	res := SyncXanoResult{}
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		return res, fmt.Errorf("id_central obrigatorio")
	}
	if !SyncXanoEnabled() {
		return res, fmt.Errorf("sync xano nao configurado (XANO_API_FINANCEIRO + WORKER_ADMIN_TOKEN)")
	}
	cli := xanoFinanceiro()

	pgPacotes, err := ListarPacotes(ctx, idCentral, "", "")
	if err != nil {
		return res, err
	}
	xanoPacotes, err := xanoListPacotes(cli, idCentral)
	if err != nil {
		return res, fmt.Errorf("listar pacotes xano: %w", err)
	}
	byQty := map[int]int{}
	for _, p := range xanoPacotes {
		q := syncIntFromAny(p["quantidade"])
		id := syncIntFromAny(p["id"])
		if q > 0 && id > 0 {
			byQty[q] = id
		}
	}
	pgToXanoPacote := map[int]int{}
	for _, pg := range pgPacotes {
		xid := byQty[pg.Quantidade]
		payload := adminPayload(idCentral, map[string]any{
			"nome":       pg.Nome,
			"quantidade": pg.Quantidade,
			"valor":      pg.Valor,
			"ativo":      normalizeAtivo(pg.Ativo),
			"observacao": pg.Observacao,
		})
		if xid > 0 {
			payload["id"] = xid
		}
		out, err := cli.Post("/fp_pacote_cota_salvar", payload)
		if err != nil {
			return res, fmt.Errorf("pacote %q: %w", pg.Nome, err)
		}
		newID := syncRecordID(out)
		if newID <= 0 {
			res.Avisos = append(res.Avisos, fmt.Sprintf("pacote %s: resposta xano sem id", pg.Nome))
			continue
		}
		pgToXanoPacote[pg.ID] = newID
		byQty[pg.Quantidade] = newID
		res.PacotesAtualizados++
	}

	pgProdutos, _, err := ListarProdutos(ctx, idCentral, "", "")
	if err != nil {
		return res, err
	}
	xanoProdutos, err := xanoListProdutos(cli, idCentral)
	if err != nil {
		return res, fmt.Errorf("listar catalogo xano: %w", err)
	}
	byPlano := map[string]int{}
	for _, p := range xanoProdutos {
		key := syncProdKey(syncStr(p["produto"]), syncStr(p["plano"]))
		id := syncIntFromAny(p["id"])
		if key != "" && id > 0 {
			byPlano[key] = id
		}
	}

	for _, pg := range pgProdutos {
		key := syncProdKey(pg.Produto, pg.Plano)
		payload := adminPayload(idCentral, map[string]any{
			"produto":                pg.Produto,
			"plano":                  pg.Plano,
			"nome_exibicao":          pg.NomeExibicao,
			"valor_mensal":           pg.ValorMensal,
			"periodicidade_padrao":   strings.TrimSpace(pg.PeriodicidadePadrao),
			"retencao_dias":          pg.RetencaoDias,
			"ativo":                  normalizeAtivo(pg.Ativo),
			"observacao":             pg.Observacao,
		})
		if xid := byPlano[key]; xid > 0 {
			payload["id"] = xid
		}
		if len(pg.LimitesJSON) > 0 && string(pg.LimitesJSON) != "null" {
			payload["limites_json"] = jsonRawToAny(pg.LimitesJSON)
		}
		if len(pg.ModulosJSON) > 0 && string(pg.ModulosJSON) != "null" {
			payload["modulos_json"] = jsonRawToAny(pg.ModulosJSON)
		}
		if strings.EqualFold(pg.Produto, "franqueadopro") && pg.FPPacoteCotaID != nil && *pg.FPPacoteCotaID > 0 {
			xPacote, ok := pgToXanoPacote[*pg.FPPacoteCotaID]
			if !ok {
				return res, fmt.Errorf("plano %s: pacote postgres id %d sem id xano", pg.Plano, *pg.FPPacoteCotaID)
			}
			payload["fp_pacote_cota_id"] = xPacote
		}
		out, err := cli.Post("/fp_catalogo_salvar", payload)
		if err != nil {
			return res, fmt.Errorf("catalogo %s/%s: %w", pg.Produto, pg.Plano, err)
		}
		if syncRecordID(out) <= 0 {
			res.Avisos = append(res.Avisos, fmt.Sprintf("produto %s/%s: resposta xano sem id", pg.Produto, pg.Plano))
			continue
		}
		res.ProdutosAtualizados++
	}
	return res, nil
}

func xanoListPacotes(cli *xano.Client, idCentral string) ([]map[string]any, error) {
	out, err := cli.Post("/fp_pacote_cota_listar", adminPayload(idCentral, map[string]any{
		"ativo": "",
	}))
	if err != nil {
		return nil, err
	}
	return syncExtractList(out), nil
}

func xanoListProdutos(cli *xano.Client, idCentral string) ([]map[string]any, error) {
	out, err := cli.Post("/fp_catalogo_listar", adminPayload(idCentral, map[string]any{}))
	if err != nil {
		return nil, err
	}
	return syncExtractList(out), nil
}

func syncExtractList(out map[string]any) []map[string]any {
	raw, _ := out["dados"].([]any)
	var list []map[string]any
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			list = append(list, m)
		}
	}
	return list
}

func syncRecordID(out map[string]any) int {
	if id := syncIntFromAny(out["id"]); id > 0 {
		return id
	}
	for _, key := range []string{"model", "salvo", "dados"} {
		if nested, ok := out[key].(map[string]any); ok {
			if id := syncIntFromAny(nested["id"]); id > 0 {
				return id
			}
		}
	}
	return 0
}

func syncProdKey(produto, plano string) string {
	return strings.ToLower(strings.TrimSpace(produto)) + "|" + strings.ToLower(strings.TrimSpace(plano))
}

func syncStr(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func syncIntFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func jsonRawToAny(raw json.RawMessage) any {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return map[string]any{}
	}
	return v
}

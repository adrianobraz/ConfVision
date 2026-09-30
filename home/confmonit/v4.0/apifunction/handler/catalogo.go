package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"apifunction/auth"
	"apifunction/pgcatalogo"
	"apifunction/service/tarifa"
)

func (h *Handler) FinanceiroCatalogoListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo == "REP" {
		writeErr(w, http.StatusBadRequest, "Representante deve usar fp_preco_rep_listar")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		Produto string `json:"produto"`
		Ativo   string `json:"ativo"`
	}
	_ = decodeJSONPermissive(r, &body)
	lista, modo, err := pgcatalogo.ListarProdutos(r.Context(), idCentral, body.Produto, body.Ativo)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"dados":      lista,
		"total":      len(lista),
		"id_central": idCentral,
		"modo_preco": modo,
		"breakglass": isBreakglassRequest(r, sess),
		"fonte":      "postgres",
	})
}

func (h *Handler) FinanceiroCatalogoSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode alterar o preco piso do catalogo")
		return
	}
	var in pgcatalogo.SalvarProdutoInput
	if err := decodeJSONPermissive(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := pgcatalogo.SalvarProduto(r.Context(), sess, in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	idCentral, _ := sess.RequireIDCentralFinanceiro()
	modo, _ := pgcatalogo.ModoPreco(r.Context(), idCentral)
	resp := map[string]any{
		"ok":                true,
		"id":                  item.ID,
		"produto":             item.Produto,
		"plano":               item.Plano,
		"nome_exibicao":       item.NomeExibicao,
		"valor_mensal":        item.ValorMensal,
		"valor_piso_breakglass": item.ValorPisoBreakglass,
		"valor_piso_minimo":   item.ValorPisoMinimo,
		"modo_preco":          modo,
		"id_central":          idCentral,
		"fonte":               "postgres",
	}
	resp["sync_xano"] = catalogoSyncXanoMeta(r.Context(), idCentral)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) FinanceiroCatalogoSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "perfil nao autorizado para seed")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		Escopo  string `json:"escopo"`
		Produto string `json:"produto"`
	}
	_ = decodeJSONPermissive(r, &body)
	escopo := strings.ToLower(strings.TrimSpace(body.Escopo))
	if escopo == "" && strings.TrimSpace(body.Produto) == "confvision_licenca" {
		escopo = "confvision_licenca"
	}
	var n int
	switch escopo {
	case "confvision_licenca":
		n, err = pgcatalogo.SeedConfvisionLicencas(r.Context(), idCentral)
	default:
		n, err = pgcatalogo.SeedProdutos(r.Context(), idCentral)
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	modo, _ := pgcatalogo.ModoPreco(r.Context(), idCentral)
	if isBreakglassRequest(r, sess) || modo == "piso" {
		_ = pgcatalogo.SyncPisoBreakglass(r.Context(), idCentral)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"inseridos":   n,
		"id_central":  idCentral,
		"modo_preco":  modo,
		"fonte":       "postgres",
		"sync_xano":   catalogoSyncXanoMeta(r.Context(), idCentral),
	})
}

func (h *Handler) FinanceiroPacoteCotaListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		Ativo string `json:"ativo"`
	}
	_ = decodeJSONPermissive(r, &body)
	idRep := ""
	if sess.UserTipo == "REP" {
		idRep = strings.TrimSpace(sess.IDRepresentante)
	}
	lista, err := pgcatalogo.ListarPacotes(r.Context(), idCentral, idRep, body.Ativo)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg, _ := pgcatalogo.ConfigCentralPreco(r.Context(), idCentral)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                          true,
		"dados":                       lista,
		"total":                       len(lista),
		"id_central":                  idCentral,
		"id_representante":            idRep,
		"userTipo":                    sess.UserTipo,
		"modo_preco":                  cfg.ModoPreco,
		"valor_unitario_minimo_cota":  cfg.ValorUnitarioMinimoCota,
		"fonte":                       "postgres",
	})
}

func (h *Handler) FinanceiroPacoteCotaPrecoRepSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "REP" {
		writeErr(w, http.StatusForbidden, "somente Representante define preco de venda de cota")
		return
	}
	var in pgcatalogo.SalvarPrecoPacoteRepInput
	if err := decodeJSONPermissive(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := pgcatalogo.SalvarPrecoPacoteRep(r.Context(), sess, in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) FinanceiroPacoteCotaSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode cadastrar Pacotes de Cotas")
		return
	}
	var in pgcatalogo.SalvarPacoteInput
	if err := decodeJSONPermissive(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := pgcatalogo.SalvarPacote(r.Context(), sess, in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	idCentral, _ := sess.RequireIDCentralFinanceiro()
	resp := map[string]any{}
	if b, err := json.Marshal(item); err == nil {
		_ = json.Unmarshal(b, &resp)
	}
	resp["sync_xano"] = catalogoSyncXanoMeta(r.Context(), idCentral)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) FinanceiroPacoteCotaSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	n, err := pgcatalogo.SeedPacotes(r.Context(), idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"inseridos":  n,
		"total":      n,
		"id_central": idCentral,
		"fonte":      "postgres",
		"sync_xano":  catalogoSyncXanoMeta(r.Context(), idCentral),
	})
}

func (h *Handler) FinanceiroCentralPrecoConfigListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": []any{}, "total": 0, "api_ok": false})
		return
	}
	lista, err := pgcatalogo.ListarPrecoConfig(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"dados":  lista,
		"total":  len(lista),
		"api_ok": true,
		"fonte":  "postgres",
	})
}

func (h *Handler) FinanceiroCentralPrecoConfigSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	if strings.ToUpper(strings.TrimSpace(r.Header.Get("X-Breakglass"))) != "S" {
		writeErr(w, http.StatusForbidden, "somente Break-glass pode alterar modo de preco")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var body struct {
		IDCentral               string   `json:"id_central"`
		ModoPreco               string   `json:"modo_preco"`
		Observacao              string   `json:"observacao"`
		ValorUnitarioMinimoCota *float64 `json:"valor_unitario_minimo_cota"`
	}
	if err := decodeJSONPermissive(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	idCentral := strings.TrimSpace(body.IDCentral)
	if idCentral == "" {
		idCentral, err = sess.RequireIDCentralFinanceiro()
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	modo := strings.ToLower(strings.TrimSpace(body.ModoPreco))
	if modo != "" {
		if err := pgcatalogo.SalvarModoPreco(r.Context(), idCentral, modo, body.Observacao, true); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if modo == "piso" {
			_ = tarifa.SyncPisoBreakglass(idCentral)
		}
	}
	if body.ValorUnitarioMinimoCota != nil {
		if err := pgcatalogo.SalvarUnitarioMinimoCota(r.Context(), idCentral, *body.ValorUnitarioMinimoCota, true); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	cfg, _ := pgcatalogo.ConfigCentralPreco(r.Context(), idCentral)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                         true,
		"id_central":                 idCentral,
		"modo_preco":                 cfg.ModoPreco,
		"valor_unitario_minimo_cota": cfg.ValorUnitarioMinimoCota,
		"fonte":                      "postgres",
	})
}

func isBreakglassRequest(r *http.Request, sess auth.SessaoAdm) bool {
	if strings.ToUpper(strings.TrimSpace(r.Header.Get("X-Breakglass"))) == "S" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(sess.IDUsuario), "BREAKGLASS")
}

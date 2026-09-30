package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"apifunction/auth"
	"apifunction/service/financeiro"
)

func (h *Handler) FinanceiroFaturaListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req struct {
		AdminToken      string `json:"admin_token"`
		Status          string `json:"status"`
		Tipo            string `json:"tipo"`
		IDFranqueado    string `json:"id_franqueado"`
		IDRepresentante string `json:"id_representante"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	adminToken := strings.TrimSpace(req.AdminToken)
	if adminToken == "" {
		adminToken = strings.TrimSpace(r.Header.Get("X-Adm-Token"))
	}

	filtro := financeiro.ListarFiltro{
		Status:          req.Status,
		Tipo:            req.Tipo,
		IDFranqueado:    req.IDFranqueado,
		IDRepresentante: req.IDRepresentante,
		AdminToken:      adminToken,
	}

	lista, fonte, err := financeiro.ListarFaturas(r.Context(), sess, filtro)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if lista == nil {
		lista = []financeiro.Fatura{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"dados":          lista,
		"total":          len(lista),
		"userTipo":       sess.UserTipo,
		"idCentral":      centralSessao(sess),
		"idCentralUUID":  strings.TrimSpace(sess.IDCentralUUID),
		"fonte":          fonte,
		"permite_global": false,
	})
}

func centralSessao(sess auth.SessaoAdm) string {
	if u := strings.TrimSpace(sess.IDCentralUUID); u != "" {
		return u
	}
	return strings.TrimSpace(sess.IDCentralCatalogo)
}

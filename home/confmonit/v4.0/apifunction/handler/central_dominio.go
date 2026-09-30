package handler

import (
	"net/http"
	"strings"

	"apifunction/auth"
	"apifunction/pgcentraldominio"
	"apifunction/service/contrato"
)

func (h *Handler) FinanceiroCentralDominioCarregar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentraldominio.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode gerenciar dominios da marca")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg, err := pgcentraldominio.Carregar(r.Context(), idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": reg})
}

func (h *Handler) FinanceiroCentralDominioSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentraldominio.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode gerenciar dominios da marca")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		Dominio    string `json:"dominio"`
		Subdominio string `json:"subdominio"`
		App        string `json:"app"`
	}
	if err := decodeJSONPermissive(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg, prov, err := pgcentraldominio.Salvar(r.Context(), pgcentraldominio.SalvarInput{
		IDCentral:  idCentral,
		Dominio:    body.Dominio,
		Subdominio: body.Subdominio,
		App:        body.App,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"dados":     reg,
		"provision": prov,
		"ssl_ok":    prov.SSLOK,
		"fqdn":      reg.Apps[strings.ToLower(strings.TrimSpace(body.App))].FQDN,
	})
}

func (h *Handler) FinanceiroCentralDominioRemover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentraldominio.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode gerenciar dominios da marca")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		App string `json:"app"`
	}
	if err := decodeJSONPermissive(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg, err := pgcentraldominio.Remover(r.Context(), idCentral, body.App)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": reg})
}

func (h *Handler) FinanceiroCentralDominioRetentarSSL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentraldominio.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode gerenciar dominios da marca")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		App string `json:"app"`
	}
	if err := decodeJSONPermissive(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg, prov, err := pgcentraldominio.RetentarSSL(r.Context(), idCentral, body.App)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"dados":     reg,
		"provision": prov,
		"ssl_ok":    prov.SSLOK,
	})
}

func (h *Handler) FinanceiroCentralDominioURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentraldominio.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	fqdn := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("fqdn")))
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	idCentral := strings.TrimSpace(r.URL.Query().Get("id_central"))

	if fqdn != "" && fqdn != "localhost" && fqdn != "127.0.0.1" {
		idCen, _, ok, err := pgcentraldominio.ResolverCentralPorFQDN(r.Context(), fqdn)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if ok && idCen != "" {
			idCentral = idCen
		}
	}
	if idCentral == "" && idFra != "" {
		_, idCen, err := contrato.FranqueadoRepresentante(idFra)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		idCentral = idCen
	}
	if idCentral == "" {
		writeErr(w, http.StatusBadRequest, "fqdn, id_franqueado ou id_central obrigatorio")
		return
	}
	urls, err := pgcentraldominio.URLsHub(r.Context(), idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "urls": urls, "id_central": idCentral})
}

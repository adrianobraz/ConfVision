package handler

import (
	"net/http"
	"strings"

	"apifunction/auth"
	"apifunction/pgcentralwhitelabel"
)

func (h *Handler) FinanceiroCentralWhitelabelCarregar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentralwhitelabel.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode gerenciar white label")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg, err := pgcentralwhitelabel.Obter(r.Context(), idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": reg})
}

func (h *Handler) FinanceiroCentralWhitelabelSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentralwhitelabel.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	if sess.UserTipo != "CEN" {
		writeErr(w, http.StatusForbidden, "somente a Central pode gerenciar white label")
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		TemaJSON  map[string]any    `json:"tema_json"`
		LogosJSON map[string]string `json:"logos_json"`
	}
	if err := decodeJSONPermissive(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg := &pgcentralwhitelabel.Registro{
		IDCentral: idCentral,
		TemaJSON:  body.TemaJSON,
		LogosJSON: body.LogosJSON,
	}
	if err := pgcentralwhitelabel.Salvar(r.Context(), reg); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out, _ := pgcentralwhitelabel.Obter(r.Context(), idCentral)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": out})
}

func (h *Handler) MarcaResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcentralwhitelabel.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres nao configurado")
		return
	}
	fqdn := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("fqdn")))
	idCentral := strings.TrimSpace(r.URL.Query().Get("id_central"))
	app := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("app")))

	var res *pgcentralwhitelabel.ResolveMarca
	var ok bool
	var err error

	if fqdn != "" {
		res, ok, err = pgcentralwhitelabel.ResolvePorFQDN(r.Context(), fqdn)
	} else if idCentral != "" && app != "" {
		res, ok, err = pgcentralwhitelabel.ResolvePorCentralApp(r.Context(), idCentral, app)
	} else {
		writeErr(w, http.StatusBadRequest, "informe fqdn ou id_central+app")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok || res == nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": nil, "app": "", "id_central": ""})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"dados":      res.Dados,
		"app":        res.App,
		"id_central": res.IDCentral,
		"fqdn":       res.FQDN,
	})
}

package handler

import (
	"net/http"
	"strings"

	"apifunction/pgcredito"
	"apifunction/pggovernanca"
)

func (h *Handler) GovernancaEstado(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcredito.Configurado() {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "pode_operar": true, "nivel": pggovernanca.NivelOK,
		})
		return
	}
	q := r.URL.Query()
	userTipo := strings.TrimSpace(q.Get("user_tipo"))
	idVinculo := strings.TrimSpace(q.Get("id_vinculo"))
	idCentral := strings.TrimSpace(q.Get("id_central"))
	breakglass := strings.EqualFold(q.Get("breakglass"), "true") || q.Get("breakglass") == "1"

	rt := pggovernanca.NovoRuntime()
	est := rt.EstadoEntidade(userTipo, idVinculo, idCentral, breakglass)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"dados": est,
	})
}

func (h *Handler) GovernancaFranqueado(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	if !pgcredito.Configurado() {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": pggovernanca.FranqueadoEstado{Motivo: "ok"}})
		return
	}
	rt := pggovernanca.NovoRuntime()
	est := rt.FranqueadoEstado(idFra)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": est})
}

func (h *Handler) GovernancaWorkerTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey string `json:"worker_key"`
	}
	_ = decodeJSONPermissive(r, &req)
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	rt := pggovernanca.NovoRuntime()
	res, err := rt.WorkerTick()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

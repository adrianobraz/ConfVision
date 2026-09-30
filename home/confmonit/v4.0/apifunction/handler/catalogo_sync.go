package handler

import (
	"context"
	"net/http"
	"strings"

	"apifunction/pgcatalogo"
)

func (h *Handler) FinanceiroCatalogoSyncXano(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcatalogo.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "postgres catalogo nao configurado")
		return
	}
	var req struct {
		WorkerKey  string `json:"worker_key"`
		IDCentral  string `json:"id_central"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	idCentral := strings.TrimSpace(req.IDCentral)
	if idCentral == "" {
		writeErr(w, http.StatusBadRequest, "id_central obrigatorio")
		return
	}
	res, err := pgcatalogo.SyncCentralToXano(r.Context(), idCentral)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"dados": res,
	})
}

func catalogoSyncXanoMeta(ctx context.Context, idCentral string) map[string]any {
	if !pgcatalogo.SyncXanoEnabled() {
		return map[string]any{"ok": false, "motivo": "desabilitado"}
	}
	res, err := pgcatalogo.SyncCentralToXano(ctx, idCentral)
	if err != nil {
		return map[string]any{"ok": false, "erro": err.Error()}
	}
	return map[string]any{
		"ok":                    true,
		"pacotes_atualizados":   res.PacotesAtualizados,
		"produtos_atualizados":  res.ProdutosAtualizados,
		"avisos":                res.Avisos,
	}
}

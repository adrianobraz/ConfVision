package handler

import (
	"net/http"
	"strings"
	"sync"

	"apifunction/pgreceptordns"
)

var (
	receptorDNSSvc     *pgreceptordns.Service
	receptorDNSSvcOnce sync.Once
)

// receptorDNS inicializa apos config.Carregar() no main (primeira requisicao).
func receptorDNS() *pgreceptordns.Service {
	receptorDNSSvcOnce.Do(func() {
		receptorDNSSvc = pgreceptordns.NovoService()
	})
	return receptorDNSSvc
}

func (h *Handler) ReceptorDNSObter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	reg, err := pgreceptordns.Obter(idFra)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": reg})
}

func (h *Handler) ReceptorDNSDisponibilidade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	var req struct {
		Subdominio string `json:"subdominio"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	disp, err := receptorDNS().VerificarDisponibilidade(req.Subdominio)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": disp})
}

func (h *Handler) ReceptorDNSSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	var req struct {
		IDFranqueado string `json:"id_franqueado"`
		Subdominio   string `json:"subdominio"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	reg, err := receptorDNS().Salvar(req.IDFranqueado, req.Subdominio)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": reg})
}

func (h *Handler) ReceptorDNSRemover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	var req struct {
		IDFranqueado string `json:"id_franqueado"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := receptorDNS().Remover(req.IDFranqueado); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ReceptorDNSEndpoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	idFab := strings.TrimSpace(r.URL.Query().Get("id_fabricante"))
	endpoints, reg, err := receptorDNS().Endpoints(idFra, idFab)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	dnsProprioAtivo := reg != nil && reg.Status == "ativo" && strings.TrimSpace(reg.FQDN) != ""
	usandoPadrao := !dnsProprioAtivo && len(endpoints) > 0
	dados := map[string]any{
		"dns":       reg,
		"endpoints": endpoints,
	}
	if usandoPadrao {
		dados["usando_dns_padrao"] = true
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"dados": dados,
	})
}

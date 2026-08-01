package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"fp-dominio-provision/apache"
	"fp-dominio-provision/config"
)

type payload struct {
	FQDN string `json:"fqdn"`
	App  string `json:"app"`
}

type Handler struct {
	Apache *apache.Manager
}

func Novo() *Handler {
	return &Handler{
		Apache: &apache.Manager{
			SitesDir:     config.ApacheSites,
			ProxyTarget:  config.ProxyTarget,
			CertbotEmail: config.CertbotEmail,
		},
	}
}

func (h *Handler) auth(w http.ResponseWriter, r *http.Request) bool {
	key := r.Header.Get("X-Provisioner-Key")
	if key == "" || key != config.ApiKey {
		http.Error(w, `{"message":"nao autorizado"}`, http.StatusUnauthorized)
		return false
	}
	return true
}

func (h *Handler) readPayload(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"message":"body invalido"}`, http.StatusBadRequest)
		return "", "", false
	}
	var p payload
	if len(body) > 0 {
		_ = json.Unmarshal(body, &p)
	}
	fqdn := strings.ToLower(strings.TrimSpace(p.FQDN))
	if fqdn == "" {
		http.Error(w, `{"message":"fqdn obrigatorio"}`, http.StatusBadRequest)
		return "", "", false
	}
	app := strings.ToLower(strings.TrimSpace(p.App))
	if app == "" {
		app = "franqueadopro"
	}
	return fqdn, app, true
}

func (h *Handler) managerForApp(app string) *apache.Manager {
	target := config.ProxyTarget
	if app == "confvision" && config.ProxyTargetCV != "" {
		target = config.ProxyTargetCV
	}
	return &apache.Manager{
		SitesDir:     config.ApacheSites,
		ProxyTarget:  target,
		CertbotEmail: config.CertbotEmail,
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, res apache.Result) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *Handler) Provisionar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.auth(w, r) {
		return
	}
	fqdn, app, ok := h.readPayload(w, r)
	if !ok {
		return
	}
	h.writeJSON(w, h.managerForApp(app).Provisionar(fqdn))
}

func (h *Handler) Remover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.auth(w, r) {
		return
	}
	fqdn, app, ok := h.readPayload(w, r)
	if !ok {
		return
	}
	h.writeJSON(w, h.managerForApp(app).Remover(fqdn))
}

func (h *Handler) RetentarSSL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.auth(w, r) {
		return
	}
	fqdn, app, ok := h.readPayload(w, r)
	if !ok {
		return
	}
	h.writeJSON(w, h.managerForApp(app).RetentarSSL(fqdn))
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

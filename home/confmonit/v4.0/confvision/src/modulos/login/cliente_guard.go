package login

import (
	"net/http"
	"strings"

	"confvision/src/seguranca"
)

// RestringirCliente bloqueia rotas administrativas para sessão CLI.
func RestringirCliente(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, erro := seguranca.LerCookies(r)
		if erro != nil || !seguranca.EhCliente(cookie) {
			next(w, r)
			return
		}

		if caminhoPermitidoCliente(r.Method, r.URL.Path) {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") || strings.Contains(r.Header.Get("Accept"), "application/json") {
			http.Error(w, `{"status":"acesso restrito ao portal do cliente"}`, http.StatusForbidden)
			return
		}
		http.Redirect(w, r, "/carregar-menu-confvision", http.StatusFound)
	}
}

func caminhoPermitidoCliente(method, path string) bool {
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	switch path {
	case "/carregar-menu-confvision",
		"/eventos",
		"/ao-vivo",
		"/mosaicos",
		"/gravacoes/timeline",
		"/gravacoes/dvr",
		"/relatorio-armado",
		"/logout",
		"/CarregarDados":
		return method == http.MethodGet || path == "/CarregarDados" || path == "/logout"
	}

	if strings.HasPrefix(path, "/ao-vivo/") && method == http.MethodGet {
		return true
	}
	if strings.HasPrefix(path, "/mosaicos/") && method == http.MethodGet {
		return true
	}

	switch {
	case path == "/api/cameras" && method == http.MethodGet:
		return true
	case strings.HasPrefix(path, "/api/cameras/") && method == http.MethodGet && !strings.Contains(path, "/areas") && !strings.Contains(path, "/snapshot") && !strings.Contains(path, "/gravacao") && !strings.Contains(path, "/licenca"):
		return true
	case path == "/api/eventos" && method == http.MethodGet:
		return true
	case strings.HasPrefix(path, "/api/eventos/") && strings.HasSuffix(path, "/clips") && method == http.MethodGet:
		return true
	case path == "/api/gravacao-segmentos" && method == http.MethodGet:
		return true
	case path == "/api/gravacao-segmento/video" && method == http.MethodGet:
		return true
	case path == "/api/grupos-visualizacao" && method == http.MethodGet:
		return true
	case path == "/api/grupos-visualizacao/disponiveis" && method == http.MethodGet:
		return true
	case strings.HasPrefix(path, "/api/grupos-visualizacao/") && method == http.MethodGet:
		return true
	case path == "/api/dispositivos" && method == http.MethodPost:
		return true
	case path == "/api/dispositivos/set-armado" && method == http.MethodPost:
		return true
	case path == "/api/dispositivos/get-armado" && method == http.MethodPost:
		return true
	case path == "/api/dispositivo-dados" && method == http.MethodPost:
		return true
	case path == "/api/clientes" && method == http.MethodPost:
		return true
	case path == "/api/dispositivos/por-franqueado" && method == http.MethodPost:
		// Armado usa esta rota; no proxy forçamos escopo do cliente
		return true
	case path == "/cvWhitelabelCarregar" && method == http.MethodPost:
		// Cliente precisa do logo whitelabel após login
		return true
	case method == http.MethodPost && (path == "/notificacoes/minhas" ||
		path == "/notificacoes/contagem" ||
		path == "/notificacoes/marcar-visto" ||
		path == "/notificacoes/marcar-lido" ||
		path == "/feedback/enviar"):
		return true
	}

	return false
}

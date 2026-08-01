package administrator

import (
	"confvision/src/seguranca"
	"net/http"
	"strings"
)

// RestringirEscopoPapel limita ADM à área administrator/RTMP e impede FRA/CLI de acessá-la.
func RestringirEscopoPapel(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, erro := seguranca.LerCookies(r)
		if erro != nil {
			next(w, r)
			return
		}

		path := strings.TrimSuffix(r.URL.Path, "/")
		if path == "" {
			path = "/"
		}

		if seguranca.EhAdministrator(cookie) {
			if caminhoPermitidoAdministrator(r.Method, path) {
				next(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.Contains(r.Header.Get("Accept"), "application/json") {
				http.Error(w, `{"status":"acesso restrito ao administrator"}`, http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/administrator/rtmp-falhas", http.StatusFound)
			return
		}

		if strings.HasPrefix(path, "/administrator") {
			responderProibido(w, r, "acesso restrito ao administrator")
			return
		}
		if path == "/rtmp-falhas" || strings.HasPrefix(path, "/api/rtmp-") {
			responderProibido(w, r, "relatorio Falhas RTMP restrito ao administrator")
			return
		}

		next(w, r)
	}
}

func caminhoPermitidoAdministrator(method, path string) bool {
	switch path {
	case "/administrator/rtmp-falhas":
		return method == http.MethodGet
	case "/logout":
		return method == http.MethodGet
	case "/api/rtmp-falhas", "/api/rtmp-falhas/resumo", "/api/rtmp-online", "/api/rtmp-bans":
		return method == http.MethodGet
	case "/api/rtmp-bans/unban", "/api/rtmp-bans/ban":
		return method == http.MethodPost
	case "/api/administrator/cameras":
		return method == http.MethodGet
	}
	if strings.HasPrefix(path, "/api/cameras/") && strings.HasSuffix(path, "/bloquear") {
		return method == http.MethodPost
	}
	return false
}

func responderProibido(w http.ResponseWriter, r *http.Request, msg string) {
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.Contains(r.Header.Get("Accept"), "application/json") {
		http.Error(w, `{"status":"`+msg+`"}`, http.StatusForbidden)
		return
	}
	http.Redirect(w, r, "/carregar-menu-confvision", http.StatusFound)
}

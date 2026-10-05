package seguranca

import (
	"net/http"
	"strings"
)

// TokenDaRequisicao obtem o JWT do cookie HttpOnly ou do header Authorization.
func TokenDaRequisicao(r *http.Request) string {
	if cookie, err := LerCookies(r); err == nil {
		if t := strings.TrimSpace(cookie["token"]); t != "" {
			return t
		}
	}

	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}

	return ""
}

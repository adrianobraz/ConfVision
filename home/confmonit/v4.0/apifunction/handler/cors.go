package handler

import (
	"net/http"
	"strings"

	"apifunction/config"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range config.CORSOrigins {
			if o == "*" || strings.EqualFold(o, origin) {
				allowed = true
				if o == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				break
			}
		}
		if allowed && origin != "" && origin != "*" {
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Breakglass, X-Adm-Central, X-Adm-Token, X-Adm-Tipo, X-Adm-Representante")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

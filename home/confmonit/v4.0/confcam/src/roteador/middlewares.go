package roteador

import (
	"confvision/src/seguranca"
	"log"
	"net/http"
	"strings"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s", r.Method, r.RequestURI, r.Host)
		next(w, r)
	}
}

func Autenticar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, erro := seguranca.LerCookies(r); erro != nil {
			if seguranca.TokenDaRequisicao(r) == "" {
				if strings.HasPrefix(r.URL.Path, "/administrator") {
					http.Redirect(w, r, "/administrator", http.StatusFound)
					return
				}
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}
		next(w, r)
	}
}

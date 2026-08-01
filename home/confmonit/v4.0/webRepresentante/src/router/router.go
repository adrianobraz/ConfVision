package router

import (
	"fmt"
	"log"
	"net/http"
	"webRepresentante/src/modulos/financeiro"
	"webRepresentante/src/modulos/financeiroFechamento"
	"webRepresentante/src/modulos/financeiroPacote"
	"webRepresentante/src/modulos/financeiroRelatorio"
	"webRepresentante/src/modulos/franqueado"
	"webRepresentante/src/modulos/franqueadoOs"
	"webRepresentante/src/modulos/home"
	"webRepresentante/src/modulos/login"
	"webRepresentante/src/modulos/relatorio"
	"webRepresentante/src/seguranca"
)

func Carregar() *http.ServeMux {
	// Login
	rotas := login.Rotas

	// Home
	rotas = append(rotas, home.Rotas...)

	// Franqueado
	rotas = append(rotas, franqueado.Rotas...)
	rotas = append(rotas, franqueadoOs.Rotas...)

	// Financeiro
	rotas = append(rotas, financeiro.Rotas...)
	rotas = append(rotas, financeiroFechamento.Rotas...)
	rotas = append(rotas, financeiroPacote.Rotas...)
	rotas = append(rotas, financeiroRelatorio.Rotas...)

	// Relatorio
	rotas = append(rotas, relatorio.Rotas...)

	mux := http.NewServeMux()

	// Rota statica
	mux.Handle(
		"GET /public/",
		http.StripPrefix("/public/", http.FileServer(http.Dir("public"))),
	)

	for _, r := range rotas {

		if r.Seguro {
			mux.HandleFunc(
				fmt.Sprintf("%s %s", r.Metodo, r.Uri),
				Logger(Autenticar(r.Controle)),
			)
		} else {
			mux.HandleFunc(
				fmt.Sprintf("%s %s", r.Metodo, r.Uri),
				Logger(r.Controle),
			)
		}

	}
	return mux
}

/// MIDDLEWARE ///

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s", r.Method, r.RequestURI, r.Host)
		next(w, r)
	}
}

func Autenticar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := seguranca.LerCookies(r); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/login", 302)
			return
		}
		next(w, r)
	}
}

package router

import (
	"fmt"
	"log"
	"net/http"
	"webCliente/src/modulos/atenderProcesso"
	"webCliente/src/modulos/atendimento"
	"webCliente/src/modulos/grade"
	"webCliente/src/modulos/home"
	"webCliente/src/modulos/login"
	"webCliente/src/modulos/meusDados"
	"webCliente/src/modulos/msgAtendente"
	relatorioEvento "webCliente/src/modulos/relatorio/evento"
	relatorioMenu "webCliente/src/modulos/relatorio/menu"
	"webCliente/src/seguranca"
	"webCliente/src/tipos"
)

func Carregar() *http.ServeMux {
	// Login
	rotas := []tipos.Rota{}

	// Login
	rotas = append(rotas, login.Rotas...)

	// Home
	rotas = append(rotas, home.Rotas...)

	rotas = append(rotas, meusDados.Rotas...)

	rotas = append(rotas, grade.Rotas...)

	rotas = append(rotas, msgAtendente.Rotas...)

	rotas = append(rotas, relatorioMenu.Rotas...)

	rotas = append(rotas, relatorioEvento.Rotas...)

	rotas = append(rotas, atendimento.Rotas...)

	rotas = append(rotas, atenderProcesso.Rotas...)

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

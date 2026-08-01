package router

import (
	"fmt"
	"log"
	"net/http"
	"webCentral/src/modulos/administrator"
	"webCentral/src/modulos/central"
	"webCentral/src/modulos/centralCadastro"
	"webCentral/src/modulos/centralUsuario"
	"webCentral/src/modulos/configuracao"
	centralAlarme "webCentral/src/modulos/configuracaoCentralAlarme"
	"webCentral/src/modulos/financeiro"
	"webCentral/src/modulos/financeiroFechamento"
	"webCentral/src/modulos/financeiroLancamento"
	"webCentral/src/modulos/financeiroLiberacao"
	"webCentral/src/modulos/financeiroPacote"
	"webCentral/src/modulos/financeiroRelatorio"
	"webCentral/src/modulos/home"
	"webCentral/src/modulos/login"
	"webCentral/src/modulos/relatorio"
	"webCentral/src/modulos/relatorioAtendimento"
	"webCentral/src/modulos/representante"
	"webCentral/src/modulos/representanteOs"
	"webCentral/src/seguranca"
)

func Carregar() *http.ServeMux {
	// Login
	rotas := login.Rotas

	// Break-glass Administrator
	rotas = append(rotas, administrator.Rotas...)

	// Home
	rotas = append(rotas, home.Rotas...)

	// Central
	rotas = append(rotas, central.Rotas...)
	rotas = append(rotas, centralCadastro.Rotas...)
	rotas = append(rotas, centralUsuario.Rotas...)

	// Franqueado
	rotas = append(rotas, representante.Rotas...)
	rotas = append(rotas, representanteOs.Rotas...)

	// Financeiro
	rotas = append(rotas, financeiro.Rotas...)
	rotas = append(rotas, financeiroPacote.Rotas...)
	rotas = append(rotas, financeiroFechamento.Rotas...)
	rotas = append(rotas, financeiroLancamento.Rotas...)
	rotas = append(rotas, financeiroLiberacao.Rotas...)
	rotas = append(rotas, financeiroRelatorio.Rotas...)

	// Relatorio
	rotas = append(rotas, relatorio.Rotas...)
	rotas = append(rotas, relatorioAtendimento.Rotas...)

	// Configuração
	rotas = append(rotas, configuracao.Rotas...)
	rotas = append(rotas, centralAlarme.Rotas...)

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

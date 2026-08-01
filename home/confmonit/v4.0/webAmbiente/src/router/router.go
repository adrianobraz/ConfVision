package router

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"webAmbiente/src/modulos/ambiente"
	"webAmbiente/src/modulos/centroOperacional"
	"webAmbiente/src/modulos/editor"
	"webAmbiente/src/modulos/home"
	"webAmbiente/src/modulos/login"
	"webAmbiente/src/modulos/mapas"
	"webAmbiente/src/modulos/monitor"
	"webAmbiente/src/modulos/notificacao"
	"webAmbiente/src/modulos/feedback"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
)

func Carregar() *http.ServeMux {
	rotas := []tipos.Rota{}

	rotas = append(rotas, login.Rotas...)
	rotas = append(rotas, home.Rotas...)
	rotas = append(rotas, mapas.Rotas...)
	rotas = append(rotas, monitor.Rotas...)
	rotas = append(rotas, centroOperacional.Rotas...)
	rotas = append(rotas, ambiente.Rotas...)
	rotas = append(rotas, editor.Rotas...)
	rotas = append(rotas, notificacao.Rotas...)
	rotas = append(rotas, feedback.Rotas...)

	mux := http.NewServeMux()

	mux.Handle(
		"GET /public/",
		http.StripPrefix("/public/", http.FileServer(http.Dir("public"))),
	)

	mux.Handle(
		"GET /audio/",
		http.StripPrefix("/audio/", http.FileServer(http.Dir("audio"))),
	)

	for _, r := range rotas {
		handler := r.Controle
		if r.Seguro {
			handler = Autenticar(handler)
		}
		if r.Master {
			handler = ExigirMaster(handler)
		}
		mux.HandleFunc(
			fmt.Sprintf("%s %s", r.Metodo, r.Uri),
			Logger(handler),
		)
	}
	return mux
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s", r.Method, r.RequestURI, r.Host)
		next(w, r)
	}
}

func Autenticar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := seguranca.LerOperador(r); err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !seguranca.EhOperadorTerminal(r) {
			seguranca.Deletar(w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func ExigirMaster(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !seguranca.EhMaster(r) {
			if r.Method == http.MethodGet {
				http.Redirect(w, r, "/home", http.StatusFound)
				return
			}
			resposta.Erro(w, http.StatusForbidden, errors.New("acesso restrito a usuario master"))
			return
		}
		next(w, r)
	}
}

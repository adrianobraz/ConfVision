package roteador

import (
	"log"
	"net/http"
	"terminal/assets/modulos/ateEventosDetalhe"
	"terminal/assets/modulos/ateProDados"
	"terminal/assets/modulos/cliDados"
	"terminal/assets/modulos/ctrClienteBuscar"
	"terminal/assets/modulos/ctrMudarSenha"
	"terminal/assets/modulos/ctrOrdemServico"
	"terminal/assets/modulos/modCliente"
	"terminal/assets/modulos/modDiscador"
	"terminal/assets/modulos/modEvtDesagrupado"
	"terminal/assets/modulos/modGrade"
	"terminal/assets/modulos/modProcedimento"
	"terminal/assets/modulos/modSetores"
	"terminal/assets/modulos/modTecnico"
	"terminal/assets/modulos/modUsuarios"
	"terminal/assets/modulos/modViatura"
	"terminal/assets/modulos/proAtendimento"
	"terminal/assets/modulos/proEventoDetalhe"
	"terminal/assets/modulos/proFranqFiltro"
	"terminal/assets/pagina/home"
	"terminal/assets/pagina/login"
	"terminal/assets/pagina/notFound"
	aux "terminal/src/auxiliar"

	"github.com/gorilla/mux"
)

func ConfigurarRotas() *mux.Router {
	r := mux.NewRouter()

	// Login
	rotas := aux.RotasGravaCusto

	rotas = append(rotas, login.Rotas...)

	// Home
	rotas = append(rotas, home.Rotas...)

	//########## MODULOS ##########\\
	//proAtendimento
	rotas = append(rotas, proAtendimento.Rotas...)

	//proFranqFitro
	rotas = append(rotas, proFranqFiltro.Rotas...)

	//proEventoDetalhe
	rotas = append(rotas, proEventoDetalhe.Rotas...)

	rotas = append(rotas, ctrClienteBuscar.Rotas...)

	rotas = append(rotas, cliDados.Rotas...)

	rotas = append(rotas, modCliente.Rotas...)

	rotas = append(rotas, modViatura.Rotas...)

	rotas = append(rotas, modTecnico.Rotas...)

	rotas = append(rotas, ateProDados.Rotas...)

	rotas = append(rotas, ateEventosDetalhe.Rotas...)

	rotas = append(rotas, modProcedimento.Rotas...)

	rotas = append(rotas, modGrade.Rotas...)

	rotas = append(rotas, modUsuarios.Rotas...)

	rotas = append(rotas, modSetores.Rotas...)

	rotas = append(rotas, modProcedimento.Rotas...)

	rotas = append(rotas, modEvtDesagrupado.Rotas...)

	rotas = append(rotas, ctrMudarSenha.Rotas...)

	rotas = append(rotas, ctrOrdemServico.Rotas...)

	rotas = append(rotas, modDiscador.Rotas...)

	r.NotFoundHandler = http.HandlerFunc(notFound.Pagina)

	for _, rota := range rotas {
		r.HandleFunc(rota.Uri, logger(rota.Funcao)).Methods(rota.Metodo)
	}

	fileServer := http.FileServer(http.Dir("./assets/"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fileServer))

	return r
}

func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s", r.Method, r.RequestURI, r.Host)
		next(w, r)
	}
}

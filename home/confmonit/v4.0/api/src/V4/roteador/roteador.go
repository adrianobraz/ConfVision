package roteadorV4

import (
	admfinanceiroV4 "api/src/V4/modulos/admfinanceiro"
	benuvemV4 "api/src/V4/modulos/benuvem"
	centralV4 "api/src/V4/modulos/central"
	centralModeloV4 "api/src/V4/modulos/centralModelo"
	clienteV4 "api/src/V4/modulos/cliente"
	confserviceparceiroV4 "api/src/V4/modulos/confserviceparceiro"
	contactidV4 "api/src/V4/modulos/contactid"
	custoAtendimentoV4 "api/src/V4/modulos/custoAtendimento"
	dispositivoV4 "api/src/V4/modulos/dispositivo"
	emailEventoV4 "api/src/V4/modulos/emailEvento"
	"api/src/V4/modulos/erroConexao"
	eventoV4 "api/src/V4/modulos/evento"
	fabricanteV4 "api/src/V4/modulos/fabricante"
	faturaV4 "api/src/V4/modulos/fatura"
	franqueadoV4 "api/src/V4/modulos/franqueado"
	gradeV4 "api/src/V4/modulos/grade"
	"api/src/V4/modulos/informativo"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	listaEnvioV4 "api/src/V4/modulos/listaEnvio"
	"api/src/V4/modulos/listaInformativo"
	"api/src/V4/modulos/feedback"
	"api/src/V4/modulos/notificacao"
	pacotev4 "api/src/V4/modulos/pacote"
	procedimentosV4 "api/src/V4/modulos/procedimentos"
	processoV4 "api/src/V4/modulos/processo"
	receptorV4 "api/src/V4/modulos/receptor"
	"api/src/V4/modulos/receptorEvento"
	representanteV4 "api/src/V4/modulos/representante"
	setorV4 "api/src/V4/modulos/setor"
	setupRelatorioV4 "api/src/V4/modulos/setupRelatorio"
	tarifacaoV4 "api/src/V4/modulos/tarifacao"
	tecnicoV4 "api/src/V4/modulos/tecnico"
	terminalV4 "api/src/V4/modulos/terminal"
	ticketV4 "api/src/V4/modulos/ticket"
	usuariosV4 "api/src/V4/modulos/usuarios"
	usuariosAlarmeV4 "api/src/V4/modulos/usuariosAlarme"
	viaturaV4 "api/src/V4/modulos/viatura"
	voipV4 "api/src/V4/modulos/voip"
	"api/src/V4/respApp"
	"api/src/V4/seguranca"
	tiposV4 "api/src/V4/tipos"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Configurar() *mux.Router {
	r := mux.NewRouter()

	var roteamento = []tiposV4.Rota{}
	roteamento = append(roteamento, admfinanceiroV4.Rotas...)
	roteamento = append(roteamento, benuvemV4.Rotas...)
	roteamento = append(roteamento, centralV4.Rotas...)
	roteamento = append(roteamento, centralModeloV4.Rotas...)
	roteamento = append(roteamento, clienteV4.Rotas...)
	roteamento = append(roteamento, confserviceparceiroV4.Rotas...)
	roteamento = append(roteamento, contactidV4.Rota...)
	roteamento = append(roteamento, custoAtendimentoV4.Rotas...)
	roteamento = append(roteamento, dispositivoV4.Rotas...)
	roteamento = append(roteamento, emailEventoV4.Rotas...)
	roteamento = append(roteamento, eventoV4.Rotas...)
	roteamento = append(roteamento, fabricanteV4.Rotas...)
	roteamento = append(roteamento, faturaV4.Rotas...)
	roteamento = append(roteamento, franqueadoV4.Rotas...)
	roteamento = append(roteamento, gradeV4.Rotas...)
	roteamento = append(roteamento, informativo.Rotas...)
	roteamento = append(roteamento, listaBoqueioV4.Rotas...)
	roteamento = append(roteamento, listaEnvioV4.Rotas...)
	roteamento = append(roteamento, listaInformativo.Rotas...)
	roteamento = append(roteamento, notificacao.Rotas...)
	roteamento = append(roteamento, feedback.Rotas...)
	roteamento = append(roteamento, pacotev4.Rotas...)
	roteamento = append(roteamento, procedimentosV4.Rotas...)
	roteamento = append(roteamento, processoV4.Rotas...)
	roteamento = append(roteamento, receptorV4.Rotas...)
	roteamento = append(roteamento, receptorEvento.Rotas...)
	roteamento = append(roteamento, representanteV4.Rotas...)
	roteamento = append(roteamento, setorV4.Rotas...)
	roteamento = append(roteamento, setupRelatorioV4.Rotas...)
	roteamento = append(roteamento, tarifacaoV4.Rotas...)
	roteamento = append(roteamento, tecnicoV4.Rotas...)
	roteamento = append(roteamento, terminalV4.Rotas...)
	roteamento = append(roteamento, ticketV4.Rotas...)
	roteamento = append(roteamento, usuariosV4.Rotas...)
	roteamento = append(roteamento, usuariosAlarmeV4.Rotas...)
	roteamento = append(roteamento, viaturaV4.Rotas...)
	roteamento = append(roteamento, voipV4.Rotas...)
	roteamento = append(roteamento, erroConexao.Rotas...)

	for _, rota := range roteamento {
		if rota.Seguro {
			r.HandleFunc(rota.URI, logger(
				autenticar(rota.Funcao),
			)).Methods(rota.Metodo)
		} else {
			r.HandleFunc(rota.URI, logger(rota.Funcao)).Methods(rota.Metodo)
		}

	}
	return r
}

func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s %s", r.Method, r.RequestURI, r.Host, r.RemoteAddr)
		next(w, r)
	}
}

func autenticar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if erro := seguranca.ValidarToken(r); erro != nil {
			fmt.Println(erro)
			respApp.Erro(w, http.StatusUnauthorized, erro)
			return
		}
		next(w, r)
	}
}

package roteador

import (
	dominioSaas "confvision/recursos/modulos/minha-empresa/dominio-saas"
	gerenciarCliente "confvision/recursos/modulos/minha-empresa/gerenciar-cliente"
	minhasLicencas "confvision/recursos/modulos/minha-empresa/minhas-licencas"
	whitelabel "confvision/recursos/modulos/minha-empresa/whitelabel"
	"confvision/src/modulos/administrator"
	"confvision/src/modulos/confvision"
	gerenciarDispositivo "confvision/src/modulos/configuracao/gerenciar-dispositivo"
	gerenciarSetoresAlarme "confvision/src/modulos/configuracao/gerenciar-setores-alarme"
	"confvision/src/modulos/login"
	"confvision/src/modulos/modAuxiliar"
	"confvision/src/modulos/notificacao"
	"confvision/src/modulos/feedback"
	"confvision/src/modulos/visapi"
	"net/http"

	"github.com/gorilla/mux"
)

func ConfigurarRotas() *mux.Router {
	r := mux.NewRouter()

	rotas := login.Rotas
	rotas = append(rotas, administrator.Rotas...)
	rotas = append(rotas, modAuxiliar.Rotas...)
	rotas = append(rotas, notificacao.Rotas...)
	rotas = append(rotas, feedback.Rotas...)
	rotas = append(rotas, confvision.Rotas...)
	rotas = append(rotas, gerenciarCliente.Rotas...)
	rotas = append(rotas, minhasLicencas.Rotas...)
	rotas = append(rotas, whitelabel.Rotas...)
	rotas = append(rotas, dominioSaas.Rotas...)
	rotas = append(rotas, gerenciarDispositivo.Rotas...)
	rotas = append(rotas, gerenciarSetoresAlarme.Rotas...)

	visapi.RegistrarRotasWorkerAPI(r)

	for _, rota := range rotas {
		if rota.Aberto {
			r.HandleFunc(rota.URI, Logger(rota.Funcao)).Methods(rota.Metodo)
		} else {
			r.HandleFunc(rota.URI, Logger(Autenticar(login.VerificarAcessoConfVision(administrator.RestringirEscopoPapel(login.RestringirCliente(rota.Funcao)))))).Methods(rota.Metodo)
		}
	}

	fileServer := http.FileServer(http.Dir("./recursos/"))
	r.PathPrefix("/recursos/").Handler(http.StripPrefix("/recursos/", fileServer))
	return r
}

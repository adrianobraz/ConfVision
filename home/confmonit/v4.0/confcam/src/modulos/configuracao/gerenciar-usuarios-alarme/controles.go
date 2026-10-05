package gerenciarUsuarioAlarme

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-gerenciar-usuarios-alarme",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarUsuarioAlarme,
		Aberto: false,
	},
	{
		URI:    "/UsuarioAlarmeListrar",
		Metodo: http.MethodPost,
		Funcao: UsuarioAlarmeListar,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeClienteListar",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeClienteListar,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeDispositivoListar",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeDispositivoListar,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeHabilitar",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeHabilitar,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeBuscar",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeBuscar,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeExcluir",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeExcluir,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeAlterar",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeAlterar,
		Aberto: false,
	},
	{
		URI:    "/UsuariosAlarmeInserir",
		Metodo: http.MethodPost,
		Funcao: UsuariosAlarmeInserir,
		Aberto: false,
	},
}

func CarregarGerenciarUsuarioAlarme(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
	} else {
		var d auxiliar.Pagina

		// Carrega title do site
		d.TituloSite = config.TituloSite

		d.NomeTela = "Gerenciar Usuario Alarme"

		d.LinkRetorno = "/carregar-menu-configuracoes"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-usuarios-alarme.html", d)
	}
}

func UsuarioAlarmeListar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuarioAlarme/listaByIdDispositivo", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeClienteListar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeDispositivoListar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

///////

func UsuariosAlarmeHabilitar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuarioAlarme/inverteAtivoByid", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeInserir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuarioAlarme/insere", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(corpo))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeBuscar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuarioAlarme/getDadosById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeAlterar(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuarioAlarme/alteraById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(corpo))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeExcluir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/usuarioAlarme/deletaById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(corpo))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosAlarmeResetarSenha(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	url := fmt.Sprintf("%s/usuarios-alarme-resetar-senha", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost, url, bytes.NewBuffer(corpo),
	)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

func UsuariosValidarEmail(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]

	url := fmt.Sprintf("%s/usuarios-alarme-validar-email/%s", config.ApiUrl, email)

	resp, erro := seguranca.RequisiacaoAutenticada(
		r, http.MethodGet, url, nil,
	)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, corpo)
}

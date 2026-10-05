package gerenciarDispositivo

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
		URI:    "/carregar-gerenciar-dispositivo",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarDispositivo,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivosFabricanteListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivosFabricanteListar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoCarregarCliente",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoCarregarCliente,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoListarPorCliente",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoListarPorCliente,
		Aberto: false,
	},

	//////////////////
	{
		URI:    "/GerenciarDispositivoListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoListar,
		Aberto: false,
	},

	{
		URI:    "/GerenciarDispositivoInserir",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoInserir,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoBuscar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoBuscar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoAlterar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoAlterar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoExcluir",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoExcluir,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoHabilitar",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoHabilitar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoGerarContaAlarme",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoGerarContaAlarme,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoVerificarContaAlarme",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoVerificarContaAlarme,
		Aberto: false,
	},

	{
		URI:    "/GerenciarDispositivosModelosCentraisListarFabricante",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivosModelosCentraisListarFabricante,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoInserirCamera",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoInserirCamera,
		Aberto: false,
	},
	{
		URI:    "/GerenciarDispositivoAlterarCamera",
		Metodo: http.MethodPost,
		Funcao: GerenciarDispositivoAlterarCamera,
		Aberto: false,
	},

	/*
		{
			URI:    "/base-resetar-senha",
			Metodo: http.MethodPost,
			Funcao: BaseResetarSenha,
			Aberto: false,
		},
	*/
}

func CarregarGerenciarDispositivo(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Dispositivo"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-confvision"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-dispositivo.html", d)
	}
}

func GerenciarDispositivoListar(w http.ResponseWriter, r *http.Request) {
	idFranqueado := mux.Vars(r)["idFranqueado"]

	url := fmt.Sprintf("%s/dispositivo-listar/%s", config.ApiUrl, idFranqueado)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodGet, url, nil)
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

func GerenciarDispositivoListarPorCliente(w http.ResponseWriter, r *http.Request) {

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

func GerenciarDispositivoHabilitar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/inverteAtivoById", config.ApiUrl)

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

func GerenciarDispositivoInserir(w http.ResponseWriter, r *http.Request) {

	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	provedorVideo := extrairProvedorVideo(corpo)
	payloadAPI, erro := removerProvedorVideo(corpo)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/insere", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(
		r, http.MethodPost, url, bytes.NewBuffer(payloadAPI),
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
		return
	}

	idDispositivo, erro := extrairIdDispositivoInserido(corpo)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if erro := gravarProvedorVideo(idDispositivo, provedorVideo); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}

func GerenciarDispositivoBuscar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/getDadosById", config.ApiUrl)

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

func GerenciarDispositivoAlterar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/alteraById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func GerenciarDispositivoExcluir(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/deletaById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	auxiliar.RespostaAPP(w, body)
}

func GerenciarDispositivoGerarContaAlarme(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/gerarContaByIdFranqueado", config.ApiUrl)

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

func GerenciarDispositivoVerificarContaAlarme(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/vericaContaByIdFranquado", config.ApiUrl)

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

func GerenciarDispositivoCarregarCliente(w http.ResponseWriter, r *http.Request) {
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

func GerenciarDispositivosFabricanteListar(w http.ResponseWriter, r *http.Request) {

	url := fmt.Sprintf("%s/v4/fabricante/listar", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, nil)
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

func GerenciarDispositivosModelosCentraisListarFabricante(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/centralModelo/listarByIdFabricante", config.ApiUrl)

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

package gerenciarSetoresAlarme

import (
	"webFranqueado/src/auxiliar"
	"webFranqueado/src/config"
	"webFranqueado/src/seguranca"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-gerenciar-setores-alarme",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarSetoresAlarme,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeListarClientes",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeClientesListar,
		Aberto: false,
	},
	{
		URI:    "/SetoresAlarmeDispositivoListarPorCliente",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeDispositivoListarPorCliente,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeListar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeHabilitar",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeHabilitar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeBuscar",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeBuscar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeAlterar",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeAlterar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeExcluir",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeExcluir,
		Aberto: false,
	},
	{
		URI:    "/GerenciarSetoresAlarmeInserir",
		Metodo: http.MethodPost,
		Funcao: GerenciarSetoresAlarmeInserir,
		Aberto: false,
	},
}

func CarregarGerenciarSetoresAlarme(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Setores Alarme"

		d.LinkRetorno = "/carregar-menu-configuracoes"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-setores-alarme.html", d)
	}
}

func GerenciarSetoresAlarmeClientesListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func GerenciarSetoresAlarmeDispositivoListarPorCliente(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requidição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Respode ao requisitante
	auxiliar.RespostaAPP(w, corpo)
}

func GerenciarSetoresAlarmeListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para consulta na API
	url := fmt.Sprintf("%s/v4/setor/listaByIdDispositivo", config.ApiUrl)

	// Executa a requisição a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se algum codigo de erro foi retornado
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera as informações retornada da API
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Retorna a informação ao requerinte
	auxiliar.RespostaAPP(w, corpo)
}

////////////////////////////////////////////////////////

func GerenciarSetoresAlarmeHabilitar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do cormo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para consulta a API
	url := fmt.Sprintf("%s/v4/setor/inverteSetorAtivaById", config.ApiUrl)
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

func GerenciarSetoresAlarmeInserir(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/setor/insere", config.ApiUrl)

	// Efetua a consulta na API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se retornou codigo de erro
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo da da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Retorna o conteudo para o requisitante
	auxiliar.RespostaAPP(w, body)
}

func GerenciarSetoresAlarmeBuscar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/setor/getDadosById", config.ApiUrl)

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

func GerenciarSetoresAlarmeAlterar(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/setor/alteraById", config.ApiUrl)

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

func GerenciarSetoresAlarmeExcluir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/setor/deletaById", config.ApiUrl)

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

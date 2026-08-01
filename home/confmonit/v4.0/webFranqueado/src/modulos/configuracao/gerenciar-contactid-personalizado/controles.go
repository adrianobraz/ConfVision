package gerenciarContactIdPersonalizado

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

var RotasContactIdPersonalizado = []auxiliar.Rota{
	{
		URI:    "/carregar-gerenciar-contactid-personalizado",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarContactIdPersonalizado,
		Aberto: false,
	},
	{
		URI:    "/listarContacid",
		Metodo: http.MethodPost,
		Funcao: listarContacid,
		Aberto: false,
	},
	{
		URI:    "/modalidadeEventoListar",
		Metodo: http.MethodPost,
		Funcao: modalidadeEventoListar,
		Aberto: false,
	},
	{
		URI:    "/carregarGrupo",
		Metodo: http.MethodPost,
		Funcao: carregarGrupo,
		Aberto: false,
	},
	{
		URI:    "/buscar",
		Metodo: http.MethodPost,
		Funcao: buscar,
		Aberto: false,
	},
	{
		URI:    "/alterar",
		Metodo: http.MethodPost,
		Funcao: alterar,
		Aberto: false,
	},
	{
		URI:    "/inserir",
		Metodo: http.MethodPost,
		Funcao: inserir,
		Aberto: false,
	},
	{
		URI:    "/excluir",
		Metodo: http.MethodPost,
		Funcao: excluir,
		Aberto: false,
	},
}

func CarregarGerenciarContactIdPersonalizado(w http.ResponseWriter, r *http.Request) {

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

		d.NomeTela = "Gerenciar Contactid Personalizado"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-configuracoes"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-contactid-personalizado.html", d)
	}
}

func listarContacid(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/contactid/listaByIdVinculo", config.ApiUrl)

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
 
func modalidadeEventoListar(w http.ResponseWriter, r *http.Request) {

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/contactid/listaPadrao", config.ApiUrl)

	// Efetua a consulta na API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, nil)
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
	body, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Retorna o conteudo para o requisitante
	auxiliar.RespostaAPP(w, body)
}

func carregarGrupo(w http.ResponseWriter, r *http.Request) {

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/contactid/listaGrupos", config.ApiUrl)

	// Efetua a consulta na API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, nil)
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
	body, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Retorna o conteudo para o requisitante
	auxiliar.RespostaAPP(w, body)
}

func buscar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para consulta a API
	url := fmt.Sprintf("%s/v4/contactid/getDadosById", config.ApiUrl)

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

func alterar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/contactid/alteraById", config.ApiUrl)

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	auxiliar.RespostaJsonOK(w)
}

func inserir(w http.ResponseWriter, r *http.Request) {

	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/contactid/insere", config.ApiUrl)

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
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}

func excluir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/contactid/deletaById", config.ApiUrl)
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
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}

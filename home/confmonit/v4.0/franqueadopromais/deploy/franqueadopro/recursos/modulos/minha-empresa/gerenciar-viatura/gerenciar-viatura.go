package gerenciarViatura

import (
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"
)

var RotasGerenciarViatura = []auxiliar.Rota{
	{
		URI:    "/CarregarPaginaGerenciarViatura",
		Metodo: http.MethodGet,
		Funcao: CarregarPaginaGerenciarViatura,
		Aberto: false,
	},

	{
		URI:    "/ViaturaInserir",
		Metodo: http.MethodPost,
		Funcao: ViaturaInserir,
		Aberto: false,
	},

	{
		URI:    "/ViaturaBuscar",
		Metodo: http.MethodPost,
		Funcao: ViaturaBuscar,
		Aberto: false,
	},

	{
		URI:    "/ViaturaAlterar",
		Metodo: http.MethodPost,
		Funcao: ViaturaAlterar,
		Aberto: false,
	},

	{
		URI:    "/ViaturaExcluir",
		Metodo: http.MethodPost,
		Funcao: ViaturaExcluir,
		Aberto: false,
	},
	{
		URI:    "/ViaturaCarregarTabela",
		Metodo: http.MethodPost,
		Funcao: ViaturaCarregarTabela,
		Aberto: false,
	},

	{
		URI:    "/ViaturaHabilitar",
		Metodo: http.MethodPost,
		Funcao: ViaturaHabilitar,
		Aberto: false,
	},
}

func CarregarPaginaGerenciarViatura(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Viatura"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-operacional"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-viatura.html", d)
	}
}

func ViaturaInserir(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a url para consulta na API
	url := fmt.Sprintf("%s/v4/viatura/insere", config.ApiUrl)

	// Consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verificar se a API retornou algo
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Envia  o conteudo retornado para o requerinte
	auxiliar.RespostaAPP(w, body)
}

func ViaturaBuscar(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a url para consulta na API
	url := fmt.Sprintf("%s/v4/viatura/getDadosById", config.ApiUrl)

	// Consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verificar se a API retornou algo
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Envia  o conteudo retornado para o requerinte
	auxiliar.RespostaAPP(w, body)
}

func ViaturaAlterar(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a url para consulta na API
	url := fmt.Sprintf("%s/v4/viatura/alteraById", config.ApiUrl)

	// Consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verificar se a API retornou algo
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Envia  o conteudo retornado para o requerinte
	auxiliar.RespostaAPP(w, body)
}

// ViaturaExcluir exclui uma viatura
func ViaturaExcluir(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a url para consulta na API
	url := fmt.Sprintf("%s/v4/viatura/deleteById", config.ApiUrl)

	// Consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verificar se a API retornou algo
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Envia  o conteudo retornado para o requerinte
	auxiliar.RespostaAPP(w, body)
}

// ViaturaCarregarTabela busca uma lista de viatura do franqueado
func ViaturaCarregarTabela(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a url para consulta na API
	url := fmt.Sprintf("%s/v4/viatura/listaByIdFranqueado", config.ApiUrl)

	// Consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verificar se a API retornou algo
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Envia  o conteudo retornado para o requerinte
	auxiliar.RespostaAPP(w, body)
}

// ViaturaHabilitar habilita ou desabilita uma viatura
func ViaturaHabilitar(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a url para consulta na API
	url := fmt.Sprintf("%s/v4/viatura/inverteAtivoById", config.ApiUrl)

	// Consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verificar se a API retornou algo
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo retornado
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Envia  o conteudo retornado para o requerinte
	auxiliar.RespostaAPP(w, body)
}

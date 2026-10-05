package gerenciarProcedimentoAtendimento

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	// Carrega a pagina Gerenciar Procedimento Atendimento
	{
		URI:    "/carregar-gerenciar-procedimento-atendimento",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarProcedimentoAtentendimento,
		Aberto: false,
	},
	// Lista todos os procedimento de atendimento de um determinado franqueado
	{
		URI:    "/GerenciarProcedimentoAtendimentoListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoListar,
		Aberto: false,
	},
	// Lista os clientes pertencentes a um determinado franqueado
	{
		URI:    "/GerenciarProcedimentoAtendimentoClienteListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoClienteListar,
		Aberto: false,
	},
	// Lista todos os grupos de atendimento da central
	{
		URI:    "/GerenciarProcedimentoAtendimentoGruposEventosListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoGruposEventosListar,
		Aberto: false,
	},
	//Habilita e desabilita o procedimento
	{
		URI:    "/GerenciarProcedimentoAtendimentoHabilitar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoHabilitar,
		Aberto: false,
	},
	// Busca os dados do procedimento pelo seu id
	{
		URI:    "/GerenciarProcedimentoAtendimentoBuscar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoBuscar,
		Aberto: false,
	},
	// Deleta o procedimento
	{
		URI:    "/GerenciarProcedimentoAtendimentoDeletar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoDeletar,
		Aberto: false,
	},
	// Altera o procedimento
	{
		URI:    "/GerenciarProcedimentoAtendimentoAlterar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoAlterar,
		Aberto: false,
	},
	//////////////////////////////////////////////////////

	{
		URI:    "/GerenciarProcedimentoAtendimentoInserir",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoInserir,
		Aberto: false,
	},
	{
		URI:    "/GerenciarProcedimentoAtendimentoHabilitar",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoHabilitar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarProcedimentoAtendimentoInserir",
		Metodo: http.MethodPost,
		Funcao: GerenciarProcedimentoAtendimentoInserir,
		Aberto: false,
	},
}

func CarregarGerenciarProcedimentoAtentendimento(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Procedimento de Atendimento"

		d.LogoMarca = "logo2Id6.png"

		d.LinkRetorno = "/carregar-menu-configuracoes"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-procedimento-atendimento.html", d)
	}
}

func GerenciarProcedimentoAtendimentoListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisiçao
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/procedimento/listaByAllIdFranqueado", config.ApiUrl)
	// Efetua o acesso a API
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
	// Recupera o conteudo do corpo da resposta
	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retornao o conteudo
	auxiliar.RespostaAPP(w, corpo)
}

func GerenciarProcedimentoAtendimentoClienteListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisiçao
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)

	// Efetua o acesso a API
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
	// Recupera o conteudo do corpo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retornao o conteudo
	auxiliar.RespostaAPP(w, body)
}

func GerenciarProcedimentoAtendimentoGruposEventosListar(w http.ResponseWriter, r *http.Request) {

	// Cria a URL de consulta a API
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

	// Recupera o conteudo da resposta
	body, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retorna o conteudo
	auxiliar.RespostaAPP(w, body)
}

func GerenciarProcedimentoAtendimentoHabilitar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do cormpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL de consulta a API
	url := fmt.Sprintf("%s/v4/procedimento/inverteAtivoByid", config.ApiUrl)

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

	// Recupera o conteudo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retorna o conteudo
	auxiliar.RespostaAPP(w, body)
}

func GerenciarProcedimentoAtendimentoBuscar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do cormpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL de consulta a API
	url := fmt.Sprintf("%s/v4/procedimento/getDadosById", config.ApiUrl)

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

	// Recupera o conteudo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retorna o conteudo
	auxiliar.RespostaAPP(w, body)
}

/////////

func GerenciarProcedimentoAtendimentoInserir(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do cormpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL de consulta a API
	url := fmt.Sprintf("%s/v4/procedimento/insere", config.ApiUrl)

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

	// Recupera o conteudo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retorna o conteudo
	auxiliar.RespostaAPP(w, body)
}

func GerenciarProcedimentoAtendimentoAlterar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do cormpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL de consulta a API
	url := fmt.Sprintf("%s/v4/procedimento/alteraById", config.ApiUrl)

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

	// Recupera o conteudo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retorna o conteudo
	auxiliar.RespostaAPP(w, body)
}

func GerenciarProcedimentoAtendimentoDeletar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do cormpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL de consulta a API
	url := fmt.Sprintf("%s/v4/procedimento/deletaById", config.ApiUrl)

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

	// Recupera o conteudo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Retorna o conteudo
	auxiliar.RespostaAPP(w, body)
}

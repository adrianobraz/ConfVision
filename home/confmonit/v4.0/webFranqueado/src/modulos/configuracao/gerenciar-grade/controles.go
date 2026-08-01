package gerenciarGrade

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

var RotasGerenciarGrade = []auxiliar.Rota{
	{
		URI:    "/carregar-gerenciar-grade",
		Metodo: http.MethodGet,
		Funcao: CarregarGerenciarGrade,
		Aberto: false,
	},
	{
		URI:    "/GerenciarGradeClienteListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeClienteListar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarGradeListar",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeListar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarGradeHabilitar",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeHabilitar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarGradeBuscar",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeBuscar,
		Aberto: false,
	},
	{
		URI:    "/GerenciarGradeAlterar",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeAlterar,
		Aberto: false,
	},

	{
		URI:    "/GerenciarGradeInserir",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeInserir,
		Aberto: false,
	},
	{
		URI:    "/GerenciarGradeExcluir",
		Metodo: http.MethodPost,
		Funcao: GerenciarGradeExcluir,
		Aberto: false,
	},
	////////////////////////////////////////

	{
		URI:    "/listarDispositivoByIdCliente",
		Metodo: http.MethodPost,
		Funcao: listarDispositivoByIdCliente,
		Aberto: false,
	},
}

func CarregarGerenciarGrade(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Gerenciar Grade de Horario"

		d.LinkRetorno = "/carregar-menu-configuracoes"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "gerenciar-grade.html", d)
	}
}

// OK
func GerenciarGradeClienteListar(w http.ResponseWriter, r *http.Request) {
	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica codigo de erro
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

// OK
func GerenciarGradeListar(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/grade/listarByIdDispositivo", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica codigo de erro
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

///////////////////////]

func GerenciarGradeHabilitar(w http.ResponseWriter, r *http.Request) {
	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/grade/inverteAtivo", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica codigo de erro
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	auxiliar.RespostaJsonOK(w)
}

func GerenciarGradeInserir(w http.ResponseWriter, r *http.Request) {

	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria URL para acesso a API
	url := fmt.Sprintf("%s/v4/grade/insere", config.ApiUrl)

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

func GerenciarGradeBuscar(w http.ResponseWriter, r *http.Request) {
	// Recupera os dados do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/grade/getDadosById", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica codigo de erro
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera o conteudo do corpo da resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Retorna o conteudo pata o APP
	auxiliar.RespostaAPP(w, body)
}

func GerenciarGradeAlterar(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/grade/alteraById", config.ApiUrl)

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

func GerenciarGradeExcluir(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/grade/deletaById", config.ApiUrl)

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

func listarDispositivoByIdCliente(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)

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

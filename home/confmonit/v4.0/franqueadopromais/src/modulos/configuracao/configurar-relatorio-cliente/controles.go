package configurarRelatorioCliente

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"

	"github.com/joho/godotenv"
)

var RotasConfigurarRelatorioCliente = []auxiliar.Rota{
	{
		URI:    "/carregar-configurar-relatorio-cliente",
		Metodo: http.MethodGet,
		Funcao: CarregarConfigurarRelatorioCliente,
		Aberto: false,
	},
	{
		URI:    "/ConfigurarRelatorioClientesListar",
		Metodo: http.MethodPost,
		Funcao: ConfigurarRelatorioClientesListar,
		Aberto: false,
	},
	{
		URI:    "/RelatorioEventoPermisaoBuscar",
		Metodo: http.MethodPost,
		Funcao: RelatorioEventoPermisaoBuscar,
		Aberto: false,
	},
	{
		URI:    "/RelatorioEventoPermisaoAlterar",
		Metodo: http.MethodPost,
		Funcao: RelatorioEventoPermisaoAlterar,
		Aberto: false,
	},
}

func CarregarConfigurarRelatorioCliente(w http.ResponseWriter, r *http.Request) {
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

		d.NomeTela = "Relatório de Eventos por Grupo"

		d.LinkRetorno = "/carregar-menu-eventos"
		// Carrega o HTML
		auxiliar.ExecutarTemplate(w, "configurar-relatorio-cliente.html", d)
	}
}

func ConfigurarRelatorioClientesListar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
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

	// Verifica se foi retornado algum erro
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera a resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Encia a resposta
	auxiliar.RespostaAPP(w, body)
}

func RelatorioEventoPermisaoBuscar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/setupRelatorio/getDadosByIdCliente", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se foi retornado algum erro
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera a resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Encia a resposta
	auxiliar.RespostaAPP(w, body)
}

func RelatorioEventoPermisaoAlterar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo do corpo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria a URL para acesso a API
	url := fmt.Sprintf("%s/v4/setupRelatorio/alteraById", config.ApiUrl)

	// Executa a consulta a API
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se foi retornado algum erro
	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Recupera a resposta
	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)

	}

	// Encia a resposta
	auxiliar.RespostaAPP(w, body)
}

package menuPrincipal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var RotasMenuPrincipal = []auxiliar.Rota{
	{
		URI:    "/carregar-menu-principal",
		Metodo: http.MethodGet,
		Funcao: CarregarMenuPrincipal,
		Aberto: false,
	},
	{
		URI:    "/MenuPrincipalCarregarInformativo",
		Metodo: http.MethodPost,
		Funcao: MenuPrincipalCarregarInformativo,
		Aberto: false,
	},
	{
		URI:    "/MenuPrincipalBuscarListaInformativo",
		Metodo: http.MethodPost,
		Funcao: MenuPrincipalBuscarListaInformativo,
		Aberto: false,
	},
	{
		URI:    "/MenuPrincipalCarregarSemComunicar",
		Metodo: http.MethodPost,
		Funcao: MenuPrincipalCarregarSemComunicar,
		Aberto: false,
	},
	{
		URI:    "/DashboardContarSemComunicacao",
		Metodo: http.MethodPost,
		Funcao: DashboardContarSemComunicacao,
		Aberto: false,
	},
	{
		URI:    "/DashboardListarSemComunicacao",
		Metodo: http.MethodPost,
		Funcao: DashboardListarSemComunicacao,
		Aberto: false,
	},
	{
		URI:    "/DashboardContarClientes",
		Metodo: http.MethodPost,
		Funcao: DashboardContarClientes,
		Aberto: false,
	},
	{
		URI:    "/carregar-dashboard-sem-comunicacao",
		Metodo: http.MethodGet,
		Funcao: CarregarDashboardSemComunicacao,
		Aberto: false,
	},
	{
		URI:    "/DashboardContarEventosPeriodo",
		Metodo: http.MethodPost,
		Funcao: DashboardContarEventosPeriodo,
		Aberto: false,
	},
	{
		URI:    "/DashboardContarEventosGrupo",
		Metodo: http.MethodPost,
		Funcao: DashboardContarEventosGrupo,
		Aberto: false,
	},
	{
		URI:    "/DashboardListarEventos",
		Metodo: http.MethodPost,
		Funcao: DashboardListarEventos,
		Aberto: false,
	},
	{
		URI:    "/carregar-dashboard-eventos",
		Metodo: http.MethodGet,
		Funcao: CarregarDashboardEventos,
		Aberto: false,
	},

	{
		URI:    "/carregar-tickets/{idFranqueado}",
		Metodo: http.MethodGet,
		Funcao: CarregarCarregarTickets,
		Aberto: false,
	},
	{
		URI:    "/excluir-tickets/{idTicket}",
		Metodo: http.MethodGet,
		Funcao: CarregarCarregarTickets,
		Aberto: false,
	},
}

// CarregarPaginaMenuPrincipal carrega a pagina do menu principal
func CarregarMenuPrincipal(w http.ResponseWriter, r *http.Request) {
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

		auxiliar.ExecutarTemplate(w, "menu-principal.html", d)
	}
}

// OK
func MenuPrincipalCarregarInformativo(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Monta a url para requisição na API
	url := fmt.Sprintf(`%s/v4/informativo/getDadosById`, config.ApiUrl)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func MenuPrincipalBuscarListaInformativo(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Monta a url para requisição na API
	url := fmt.Sprintf(`%s/v4/listaInformativo/listaByIdAlvo`, config.ApiUrl)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

// --
func MenuPrincipalCarregarSemComunicar(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Monta a url para requisição na API
	// url := fmt.Sprintf(`%s/DispositivoCarregarSemComunicacao`, config.ApiUrl)
	url := fmt.Sprintf(`%s/v4/dispositivo/listarSemComunicacaoByIdFranqueado`, config.ApiUrl)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)

}

func DashboardContarSemComunicacao(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf(`%s/v4/dispositivo/contarSemComunicacaoFaixas`, config.ApiUrl)

	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func DashboardListarSemComunicacao(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var req struct {
		IdFranqueado string `json:"idFranqueado"`
		Faixa        string `json:"faixa"`
	}
	_ = json.Unmarshal(body, &req)

	url := fmt.Sprintf(`%s/v4/dispositivo/listarSemComunicacaoByFaixa`, config.ApiUrl)
	if req.Faixa == "todos" {
		url = fmt.Sprintf(`%s/v4/dispositivo/listarSemComunicacaoByIdFranqueado`, config.ApiUrl)
	}

	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func DashboardContarClientes(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf(`%s/v4/cliente/listarByIdFranqueado`, config.ApiUrl)

	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func CarregarDashboardSemComunicacao(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite
	d.NomeTela = "Sem Comunicação"
	d.LinkRetorno = "/carregar-menu-principal"
	d.IconePagina = "bi-wifi-off"
	d.LogoMarca = "logo2Id6.png"

	auxiliar.ExecutarTemplate(w, "dashboard-sem-comunicacao.html", d)
}

func DashboardContarEventosPeriodo(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf(`%s/v4/evento/contarByIdFranqueadoPeriodo`, config.ApiUrl)
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func DashboardContarEventosGrupo(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf(`%s/v4/evento/contarByIdFranqueadoGrupo`, config.ApiUrl)
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func DashboardListarEventos(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf(`%s/v4/evento/listarByIdFranqueadoStartEndGrupo`, config.ApiUrl)
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)
}

func CarregarDashboardEventos(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	d.TituloSite = config.TituloSite
	d.NomeTela = "Eventos"
	d.LinkRetorno = "/carregar-menu-principal"
	d.IconePagina = "bi-lightning-fill"
	d.LogoMarca = "logo2Id6.png"

	auxiliar.ExecutarTemplate(w, "dashboard-eventos.html", d)
}

func CarregarCarregarTickets(w http.ResponseWriter, r *http.Request) {
	idFranqueado := mux.Vars(r)["idFranqueado"]

	// Monta a url para requisição na API
	url := fmt.Sprintf(
		`%s/tickets-franqueado-carregar-todos/%s`, config.ApiUrl, idFranqueado,
	)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodGet, url, nil)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	corpo, erro := io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, corpo)

}

package menuPrincipal

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

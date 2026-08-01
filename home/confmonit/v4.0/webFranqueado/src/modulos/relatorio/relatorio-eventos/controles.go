package relatorioEventos

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"webFranqueado/src/auxiliar"
	"webFranqueado/src/config"
	"webFranqueado/src/seguranca"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-relatorio-eventos",
		Metodo: http.MethodGet,
		Funcao: CarregarRelatorioEventos,
		Aberto: false,
	},
	{
		URI:    "/RelatorioEventosCarregarClientes",
		Metodo: http.MethodPost,
		Funcao: RelatorioEventosCarregarClientes,
		Aberto: false,
	},
	{
		URI:    "/relatorioEventosListar",
		Metodo: http.MethodPost,
		Funcao: relatorioEventosListar,
		Aberto: false,
	},
	{
		URI:    "/relatorioEventosCarregarDispositivos",
		Metodo: http.MethodPost,
		Funcao: relatorioEventosCarregarDispositivos,
		Aberto: false,
	},
}

func CarregarRelatorioEventos(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina

	// Carrega title do site
	d.TituloSite = config.TituloSite

	d.NomeTela = "Relatório dos Eventos"

	d.LogoMarca = "logo2Id6.png"

	d.LinkRetorno = "/carregar-menu-relatorio"
	// Carrega o HTML
	auxiliar.ExecutarTemplate(w, "relatorio-eventos.html", d)
}

func RelatorioEventosCarregarClientes(w http.ResponseWriter, r *http.Request) {

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

func relatorioEventosListar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/evento/listarByDispStartEnd", config.ApiUrl)

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

func relatorioEventosCarregarDispositivos(w http.ResponseWriter, r *http.Request) {

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

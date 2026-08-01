package relatorioEvento

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"webCliente/src/config"
	"webCliente/src/resposta"
	"webCliente/src/seguranca"
	"webCliente/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/relatorio/evento/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/relatorio/evento/carregarDispositivo",
		Metodo:   http.MethodPost,
		Controle: carregarDispositivo,
		Seguro:   true,
	},
	{
		Uri:      "/relatorio/evento/filtrar",
		Metodo:   http.MethodPost,
		Controle: filtrar,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/relatorio/evento/evento.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Relatório Evento", config.TituloSite),
		NavbarTitulo: "Relatório Evento",
	}
	if err := page.ExecuteTemplate(w, "evento.html", d); err != nil {
		fmt.Println(err)
	}
}

func carregarDispositivo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.Api),
		bytes.NewBuffer(body),
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		resposta.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.App(w, body)
}

func filtrar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/evento/listarByDispStartEndGrupo", config.Api),
		bytes.NewBuffer(body),
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		resposta.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.App(w, body)
}

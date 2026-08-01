package meusDados

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
		Uri:      "/meusDadosPage",
		Metodo:   http.MethodGet,
		Controle: meusDadosPage,
		Seguro:   true,
	},
	{
		Uri:      "/carregarDados",
		Metodo:   http.MethodPost,
		Controle: carregarDados,
		Seguro:   true,
	},
}

func meusDadosPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/meusDados/meusDados.html",
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
		Titulo:       fmt.Sprintf("%s - Meus Dados", config.TituloSite),
		NavbarTitulo: "Meus Dados",
	}
	if err := page.ExecuteTemplate(w, "meusDados.html", d); err != nil {
		fmt.Println(err)
	}
}

func carregarDados(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/cliente/getDadosById", config.Api),
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

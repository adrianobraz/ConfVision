package atenderProcesso

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
		Uri:      "/atenderProcesso/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/atenderProcesso/getProcessoById",
		Metodo:   http.MethodPost,
		Controle: getProcessoById,
		Seguro:   true,
	},
	{
		Uri:      "/atenderProcesso/listarEventosByProcesso",
		Metodo:   http.MethodPost,
		Controle: listarEventosByProcesso,
		Seguro:   true,
	},
	{
		Uri:      "/atenderProcesso/finalizarProcesso",
		Metodo:   http.MethodPost,
		Controle: finalizarProcesso,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/atenderProcesso/atenderProcesso.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Atender Processo", config.TituloSite),
		NavbarTitulo: "Atender Processo",
	}
	if err := page.ExecuteTemplate(w, "atenderProcesso.html", d); err != nil {
		fmt.Println(err)
	}
}

func getProcessoById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("aqui")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/terminal/getDadosProcessoById", config.Api),
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

func listarEventosByProcesso(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/terminal/listarEventosByProcesso", config.Api),
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

func finalizarProcesso(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/terminal/finalizarProcesso", config.Api),
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

//

package representanteOs

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"webCentral/src/config"
	"webCentral/src/resposta"
	"webCentral/src/seguranca"
	"webCentral/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/representante/os",
		Metodo:   http.MethodGet,
		Controle: representanteOsPage,
		Seguro:   true,
	},
	{
		Uri:      "/representante/os/listar",
		Metodo:   http.MethodPost,
		Controle: listar,
		Seguro:   true,
	},
	{
		Uri:      "/representante/os/atualizar",
		Metodo:   http.MethodPost,
		Controle: atualizar,
		Seguro:   true,
	},
	{
		Uri:      "/representante/os/inserir",
		Metodo:   http.MethodPost,
		Controle: inserir,
		Seguro:   true,
	},
	{
		Uri:      "/representante/os/buscar",
		Metodo:   http.MethodPost,
		Controle: buscar,
		Seguro:   true,
	},
	{
		Uri:      "/representante/os/carregarRepresentante",
		Metodo:   http.MethodPost,
		Controle: carregarRepresentante,
		Seguro:   true,
	},
}

func representanteOsPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/representanteOs/representanteOs.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       "Franqueado Ordem Serviço",
		NavbarLink:   "/franqueado",
		NavbarIcon:   "bi bi-clipboard-check",
		NavbarTitulo: "Ordem Serviço",
	}
	if err := page.ExecuteTemplate(w, "representanteOs.html", d); err != nil {
		fmt.Println(err)
	}
}

func listar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/ticket/listarByIdMaster", config.Api),
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

func atualizar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	fmt.Println("aqui")

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/ticket/alteraById", config.Api),
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

func inserir(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/ticket/insere", config.Api),
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

func buscar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/ticket/getDadosById", config.Api),
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

func carregarRepresentante(w http.ResponseWriter, r *http.Request) {

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/representante/listar", config.Api),
		nil,
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if res.StatusCode >= 400 {
		resposta.TratarStatusCodeDeErro(w, res)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.App(w, body)
}

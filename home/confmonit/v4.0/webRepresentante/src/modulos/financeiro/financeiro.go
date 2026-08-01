package financeiro

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"webRepresentante/src/config"
	"webRepresentante/src/resposta"
	"webRepresentante/src/seguranca"
	"webRepresentante/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/financeiro",
		Metodo:   http.MethodGet,
		Controle: financeiroPage,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/listarFranq",
		Metodo:   http.MethodPost,
		Controle: listarFranq,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/reverterCancelarFranq",
		Metodo:   http.MethodPost,
		Controle: reverterCancelarFranq,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/deletarFranq",
		Metodo:   http.MethodPost,
		Controle: deletarFranq,
		Seguro:   true,
	},
}

func financeiroPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/financeiro/financeiro.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       fmt.Sprintf("%s - Financeiro", config.TituloSite),
		NavbarIcon:   "bi bi-bank",
		NavbarTitulo: "Financeiro",
	}
	if err := page.ExecuteTemplate(w, "financeiro.html", d); err != nil {
		fmt.Println(err)
	}

}

func listarFranq(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf("%s/v4/franqueado/listarByIdRepresentante", config.Api),
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

func reverterCancelarFranq(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/franqueado/reverteCancelaById", config.Api),
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

func deletarFranq(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/franqueado/deletaById", config.Api),
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

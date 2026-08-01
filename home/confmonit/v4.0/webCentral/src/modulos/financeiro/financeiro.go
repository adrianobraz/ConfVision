package financeiro

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
		Uri:      "/financeiro",
		Metodo:   http.MethodGet,
		Controle: financeiroPage,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/listarRep",
		Metodo:   http.MethodPost,
		Controle: listar,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/cancelaRep",
		Metodo:   http.MethodPost,
		Controle: cancelaRep,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/reverteCancelaRep",
		Metodo:   http.MethodPost,
		Controle: reverteCancelaRep,
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
		Titulo:       "Financeiro",
		NavbarLink:   "/home",
		NavbarIcon:   "bi bi-bank",
		NavbarTitulo: "Financeiro",
	}
	if err := page.ExecuteTemplate(w, "financeiro.html", d); err != nil {
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
		fmt.Sprintf("%s/v4/representante/listar", config.Api),
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

func cancelaRep(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/representante/cancelaById", config.Api),
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

func reverteCancelaRep(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/representante/reverteCancelaById", config.Api),
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

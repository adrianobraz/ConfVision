package financeiroLancamento

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
		Uri:      "/financeiro/lancamento",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/lancamento/listar",
		Metodo:   http.MethodPost,
		Controle: listar,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/lancamento/inserir",
		Metodo:   http.MethodPost,
		Controle: inserir,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/lancamento/deletar",
		Metodo:   http.MethodPost,
		Controle: deletar,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/financeiroLancamento/financeiroLancamento.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}
	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}
	d := tipos.Page{
		Titulo:       "Inserir Lançamento",
		NavbarLink:   "/financeiro",
		NavbarIcon:   "bi bi-plus-circle",
		NavbarTitulo: "Inserir Lançamento",
	}
	if err := page.ExecuteTemplate(w, "financeiroLancamento.html", d); err != nil {
		fmt.Println(err)
	}
}

func listar(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/tarifacao/listaLancamentosPendentes", config.Api))
}

func inserir(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/tarifacao/insereLancamento", config.Api))
}

func deletar(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/tarifacao/deleteLancamento", config.Api))
}

func proxyPost(w http.ResponseWriter, r *http.Request, url string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	res, err := seguranca.ReqAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
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

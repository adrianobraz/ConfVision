package financeiroFechamento

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
		Uri:      "/financeiro/fechamento",
		Metodo:   http.MethodGet,
		Controle: fechamentoPage,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/faturaListarByCentral",
		Metodo:   http.MethodPost,
		Controle: faturaListarByCentral,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/faturaGetDadosById",
		Metodo:   http.MethodPost,
		Controle: faturaGetDadosById,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/recebeFaturaById",
		Metodo:   http.MethodPost,
		Controle: recebeFaturaById,
		Seguro:   true,
	},
}

func fechamentoPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/financeiroFechamento/financeiroFechamento.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       "Fechamento",
		NavbarLink:   "/financeiro",
		NavbarIcon:   "bi bi-calendar2-week",
		NavbarTitulo: "Fechamento",
	}
	if err := page.ExecuteTemplate(w, "financeiroFechamento.html", d); err != nil {
		fmt.Println(err)
	}

}

func faturaListarByCentral(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/fatura/listaByIdOrigem", config.Api))
}

func faturaGetDadosById(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/fatura/getDadosById", config.Api))
}

func recebeFaturaById(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/fatura/recebeFaturaById", config.Api))
}

func proxyPost(w http.ResponseWriter, r *http.Request, url string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		url,
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

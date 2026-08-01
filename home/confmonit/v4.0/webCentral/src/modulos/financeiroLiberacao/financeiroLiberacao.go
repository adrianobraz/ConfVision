package financeiroLiberacao

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
		Uri:      "/financeiro/liberacao",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/liberacao/gravar",
		Metodo:   http.MethodPost,
		Controle: gravar,
		Seguro:   true,
	},
	{
		Uri:      "/financeiro/liberacao/listaBloqueados",
		Metodo:   http.MethodPost,
		Controle: listaBloqueados,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/financeiroLiberacao/financeiroLiberacao.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}
	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}
	d := tipos.Page{
		Titulo:       "Liberação Provisória",
		NavbarLink:   "/financeiro",
		NavbarIcon:   "bi bi-unlock",
		NavbarTitulo: "Liberação Provisória",
	}
	if err := page.ExecuteTemplate(w, "financeiroLiberacao.html", d); err != nil {
		fmt.Println(err)
	}
}

func gravar(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/listaBloqueio/liberarProvisorio", config.Api))
}

func listaBloqueados(w http.ResponseWriter, r *http.Request) {
	proxyPost(w, r, fmt.Sprintf("%s/v4/representante/listarToDesativado", config.Api))
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

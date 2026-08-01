package centralCadastro

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
		Uri:      "/cadastro/centrais",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/cadastro/centrais/listar",
		Metodo:   http.MethodPost,
		Controle: listar,
		Seguro:   true,
	},
	{
		Uri:      "/cadastro/centrais/inserir",
		Metodo:   http.MethodPost,
		Controle: inserir,
		Seguro:   true,
	},
	{
		Uri:      "/cadastro/centrais/alterar",
		Metodo:   http.MethodPost,
		Controle: alterar,
		Seguro:   true,
	},
}

func exigeBreakglass(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := seguranca.LerCookies(r)
	if err != nil || cookie["idUsuario"] != "BREAKGLASS" {
		resposta.Erro(w, http.StatusForbidden, fmt.Errorf("acesso restrito ao administrador"))
		return false
	}
	return true
}

func page(w http.ResponseWriter, r *http.Request) {
	cookie, _ := seguranca.LerCookies(r)
	if cookie["idUsuario"] != "BREAKGLASS" {
		http.Redirect(w, r, "/home", 302)
		return
	}

	templ := []string{
		"public/templates/centralCadastro/centralCadastro.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/tabela.html",
	}
	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}
	d := tipos.Page{
		Titulo:       "Cadastro de Centrais",
		NavbarLink:   "/home",
		NavbarIcon:   "bi bi-building",
		NavbarTitulo: "Centrais",
	}
	if err := page.ExecuteTemplate(w, "centralCadastro.html", d); err != nil {
		fmt.Println(err)
	}
}

func listar(w http.ResponseWriter, r *http.Request) {
	if !exigeBreakglass(w, r) {
		return
	}
	proxyPost(w, r, fmt.Sprintf("%s/v4/central/lista", config.Api))
}

func inserir(w http.ResponseWriter, r *http.Request) {
	if !exigeBreakglass(w, r) {
		return
	}
	proxyPost(w, r, fmt.Sprintf("%s/v4/central/insere", config.Api))
}

func alterar(w http.ResponseWriter, r *http.Request) {
	if !exigeBreakglass(w, r) {
		return
	}
	proxyPost(w, r, fmt.Sprintf("%s/v4/central/atualiza", config.Api))
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

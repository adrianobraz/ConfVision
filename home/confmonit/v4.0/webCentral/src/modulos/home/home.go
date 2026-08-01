package home

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
		Uri:      "/home",
		Metodo:   http.MethodGet,
		Controle: homePage,
		Seguro:   true,
	},

	{
		Uri:      "/getRepBloqueado",
		Metodo:   http.MethodPost,
		Controle: getRepBloqueado,
		Seguro:   true,
	},
	{
		Uri:      "/repHabilitar",
		Metodo:   http.MethodPost,
		Controle: repHabilitar,
		Seguro:   true,
	},
	{
		Uri:      "/faturaListarByCentral",
		Metodo:   http.MethodPost,
		Controle: faturaListarByCentral,
		Seguro:   true,
	},
	{
		Uri:      "/usuarioCentralAlterarSenha",
		Metodo:   http.MethodPost,
		Controle: usuarioCentralAlterarSenha,
		Seguro:   true,
	},
}

func homePage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/home/home.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo:       "Home",
		NavbarLink:   "/home",
		NavbarIcon:   "bi bi-house",
		NavbarTitulo: "Home",
	}
	if err := page.ExecuteTemplate(w, "home.html", d); err != nil {
		fmt.Println(err)
	}
}

func getRepBloqueado(w http.ResponseWriter, r *http.Request) {

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/representante/listarToDesativado", config.Api),
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

func repHabilitar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/representante/setAtivoById", config.Api),
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

// faturaListarByCentral lista todas as fatura da central
func faturaListarByCentral(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/faturaListarByCentral", config.Api),
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

func usuarioCentralAlterarSenha(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/usuario/alterarSenhaById", config.Api),
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

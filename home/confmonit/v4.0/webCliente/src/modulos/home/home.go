package home

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
		Uri:      "/home",
		Metodo:   http.MethodGet,
		Controle: homePage,
		Seguro:   true,
	},
	{
		Uri:      "/home/carregarDispositivo",
		Metodo:   http.MethodPost,
		Controle: carregarDispositivo,
		Seguro:   true,
	},
	{
		Uri:      "/alterarSenha",
		Metodo:   http.MethodPost,
		Controle: alterarSenha,
		Seguro:   true,
	},
	{
		Uri:      "/enviarPanico",
		Metodo:   http.MethodPost,
		Controle: enviarPanico,
		Seguro:   true,
	},
	{
		Uri:      "/armarDispositivo",
		Metodo:   http.MethodPost,
		Controle: armarDispositivo,
		Seguro:   true,
	},
	{
		Uri:      "/setDispPadraoBtnPanico",
		Metodo:   http.MethodPost,
		Controle: setDispPadraoBtnPanico,
		Seguro:   true,
	},
}

func homePage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/home/home.html",
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
		Titulo:       fmt.Sprintf("%s - Home", config.TituloSite),
		NavbarTitulo: "Home",
	}
	if err := page.ExecuteTemplate(w, "home.html", d); err != nil {
		fmt.Println(err)
	}
}

func carregarDispositivo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.Api),
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

func alterarSenha(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/cliente/alterarSenhaById", config.Api),
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

func enviarPanico(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := http.Post(config.RecWeb.Url, "application/json", bytes.NewBuffer(body))
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

func armarDispositivo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	res, err := http.Post(config.RecComando.Url, "application/json", bytes.NewBuffer(body))
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

	fmt.Println(string(body))
	resposta.App(w, body)
}

func setDispPadraoBtnPanico(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/cliente/setDispPadraoBtnPanico", config.Api),
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

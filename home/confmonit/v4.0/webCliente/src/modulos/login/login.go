package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"webCliente/src/config"
	"webCliente/src/resposta"
	"webCliente/src/seguranca"
	"webCliente/src/tipos"
)

type login struct {
	Email     string `json:"email"`
	Senha     string `json:"senha"`
	Token     string `json:"token"`
	IdUsuario string `json:"idUsuario"`
}

var Rotas = []tipos.Rota{
	{
		Uri:      "/",
		Metodo:   http.MethodGet,
		Controle: loginPage,
		Seguro:   false,
	},
	{
		Uri:      "/login",
		Metodo:   http.MethodGet,
		Controle: loginPage,
		Seguro:   false,
	},
	{
		Uri:      "/logar",
		Metodo:   http.MethodPost,
		Controle: logar,
		Seguro:   false,
	},
	{
		Uri:      "/logout",
		Metodo:   http.MethodGet,
		Controle: logout,
		Seguro:   false,
	},
	{
		Uri:      "/carregaDadosReceptor",
		Metodo:   http.MethodPost,
		Controle: carregaDadosReceptor,
		Seguro:   false,
	},
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	cookie, _ := seguranca.LerCookies(r)
	if cookie["token"] != "" {

		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	templ := []string{
		"public/templates/login/login.html",
		"public/templates/components/pagina.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo: fmt.Sprintf("%s - Login", config.TituloSite),
	}
	if err := page.ExecuteTemplate(w, "login.html", d); err != nil {
		fmt.Println(err)
	}
}

func logar(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Faz a requisição na API
	resp, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf(`%s/v4/cliente/logar`, config.Api),
		bytes.NewBuffer(body),
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if resp.StatusCode >= 400 {
		resposta.TratarStatusCodeDeErro(w, resp)
		return
	}

	// Pega o json no copro da resposta
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto login
	var respLogin struct {
		Dados struct {
			login
		} `json:"dados"`
	}

	// Popula o objeto login
	if err = json.Unmarshal(body, &respLogin); err != nil {

		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	// Cria um cookie com os dados do login
	if err = seguranca.SalvarCookies(
		w,
		respLogin.Dados.IdUsuario,
		respLogin.Dados.Token,
	); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde ao APP
	resposta.App(w, body)
}

func logout(w http.ResponseWriter, r *http.Request) {
	seguranca.Deletar(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func carregaDadosReceptor(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	res, err := seguranca.ReqAutenticada(
		r, http.MethodPost,
		fmt.Sprintf("%s/v4/receptorEvento/carregarDados", config.Api),
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

	var dados struct {
		Status string `json:"status"`
		Dados  []struct {
			Nome   string `json:"nome"`
			Porta  string `json:"porta"`
			Rota   string `json:"rota"`
			Modulo string `json:"modulo"`
			Senha  string `json:"senha"`
		}
	}

	if err := json.Unmarshal(body, &dados); err != nil {

		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	for _, i := range dados.Dados {
		if i.Modulo == "REC_COMANDO" {
			config.RecComando.Senha = i.Senha
			config.RecComando.Url = fmt.Sprintf("http://185.130.61.3:%s/armar", i.Porta)
		} else if i.Modulo == "REC_WEB" {
			config.RecWeb.Url = fmt.Sprintf("http://185.130.61.3:%s/recebe-evento", i.Porta)
			config.RecWeb.Senha = i.Senha
		}
	}

	resposta.App(w, body)
}

package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"webRepresentante/src/config"
	"webRepresentante/src/resposta"
	"webRepresentante/src/seguranca"
	"webRepresentante/src/tipos"
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
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	cookie, _ := seguranca.LerCookies(r)
	if cookie["token"] != "" {

		http.Redirect(w, r, "/home", 302)
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
		fmt.Sprintf(`%s/v4/representante/logar`, config.Api),
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
	http.Redirect(w, r, "/login", 302)
}

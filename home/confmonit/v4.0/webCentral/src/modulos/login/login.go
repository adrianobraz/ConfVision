package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"webCentral/src/config"
	"webCentral/src/resposta"
	"webCentral/src/seguranca"
	"webCentral/src/tipos"
)

type login struct {
	Email     string `json:"email"`
	Senha     string `json:"senha"`
	Token     string `json:"token"`
	IdUsuario string `json:"idUsuario"`
}

type centralLogin struct {
	Dados struct {
		Token          string `json:"token"`
		IdUsuario      string `json:"idUsuario"`
		IdVinculo      string `json:"idVinculo"`
		Tipo           string `json:"tipo"`
		Nome           string `json:"nome"`
		Nick           string `json:"nick"`
		Email1         string `json:"email1"`
		Email2         string `json:"email2"`
		EnviarEmail    string `json:"enviarEmail"`
		Senha          string `json:"senha"`
		Telefone1      string `json:"telefone1"`
		Telefone2      string `json:"telefone2"`
		UsuarioTeminal string `json:"usuarioTerminal"`
		UsuarioWeb     string `json:"usuarioWeb"`
		Master         string `json:"master"`
		Ativo          string `json:"ativo"`
		DataCadastro   string `json:"dataCriação"`
	} `json:"dados"`
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
		Titulo: "Login Central",
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
		fmt.Sprintf(`%s/v4/central/logar`, config.Api),
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
	var respLogin centralLogin

	// Popula o objeto login
	if err = json.Unmarshal(body, &respLogin); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if respLogin.Dados.Ativo == "S" &&
		respLogin.Dados.Tipo == "CEN" &&
		respLogin.Dados.UsuarioWeb == "S" {
		// Cria um cookie com os dados do login
		if err = seguranca.SalvarCookies(
			w,
			respLogin.Dados.IdUsuario,
			respLogin.Dados.Token,
		); err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
	}

	// Responde ao APP
	resposta.App(w, body)
}

func logout(w http.ResponseWriter, r *http.Request) {
	seguranca.Deletar(w)
	http.Redirect(w, r, "/login", 302)
}

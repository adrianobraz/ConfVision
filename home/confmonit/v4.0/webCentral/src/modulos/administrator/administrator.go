package administrator

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

type adminLoginResp struct {
	Dados struct {
		Token          string `json:"token"`
		IdUsuario      string `json:"idUsuario"`
		IdVinculo      string `json:"idVinculo"`
		Tipo           string `json:"tipo"`
		Nome           string `json:"nome"`
		Nick           string `json:"nick"`
		Email1         string `json:"email1"`
		UsuarioTeminal string `json:"usuarioTerminal"`
		UsuarioWeb     string `json:"usuarioWeb"`
		Master         string `json:"master"`
		Ativo          string `json:"ativo"`
	} `json:"dados"`
}

var Rotas = []tipos.Rota{
	{
		Uri:      "/administrator",
		Metodo:   http.MethodGet,
		Controle: administratorPage,
		Seguro:   false,
	},
	{
		Uri:      "/administrator/logar",
		Metodo:   http.MethodPost,
		Controle: logar,
		Seguro:   false,
	},
}

func administratorPage(w http.ResponseWriter, r *http.Request) {
	cookie, _ := seguranca.LerCookies(r)
	if cookie["token"] != "" {
		http.Redirect(w, r, "/home", 302)
		return
	}

	templ := []string{
		"public/templates/administrator/administrator.html",
		"public/templates/components/pagina.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	var d = tipos.Page{
		Titulo: "Administrator",
	}
	if err := page.ExecuteTemplate(w, "administrator.html", d); err != nil {
		fmt.Println(err)
	}
}

func logar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resp, err := seguranca.ReqAutenticada(
		r,
		http.MethodPost,
		fmt.Sprintf(`%s/v4/central/logarAdministrador`, config.Api),
		bytes.NewBuffer(body),
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if resp.StatusCode >= 400 {
		resposta.TratarStatusCodeDeErro(w, resp)
		return
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var respLogin adminLoginResp
	if err = json.Unmarshal(body, &respLogin); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if respLogin.Dados.Ativo == "S" &&
		respLogin.Dados.Tipo == "CEN" &&
		respLogin.Dados.UsuarioWeb == "S" {
		if err = seguranca.SalvarCookies(
			w,
			respLogin.Dados.IdUsuario,
			respLogin.Dados.Token,
		); err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
	}

	resposta.App(w, body)
}

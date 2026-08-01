package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"webFranqueado/src/auxiliar"
	"webFranqueado/src/config"
	"webFranqueado/src/seguranca"

	"github.com/joho/godotenv"
)

// CarregarPaginaLogin carrega a pagina para efetuar o login
func CarregarLogin(w http.ResponseWriter, r *http.Request) {

	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
	} else {
		cookie, _ := seguranca.LerCookies(r)
		if cookie["token"] != "" {
			http.Redirect(w, r, "/carregar-menu-principal", 302)
			return
		}

		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "login.html", d)
	}
}

// LoginLogar efetua o login no sistema
func LoginLogar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Monta a url para requisição na API
	// url := fmt.Sprintf(`%s/logar`, config.ApiUrl)
	url := fmt.Sprintf(`%s/v4/franqueado/logar`, config.ApiUrl)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	corpo, erro := io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	// Cria um objeto login
	var login struct {
		Dados struct {
			Token     string `json:"token"`
			IdUsuario string `json:"idUsuario"`
		} `json:"dados"`
	}

	// Popula o objeto login
	if erro = json.Unmarshal(corpo, &login); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Cria um cookie com os dados do login
	if erro = seguranca.SalvarCookies(w, login.Dados.IdUsuario, login.Dados.Token); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Responde ao APP
	auxiliar.RespostaAPP(w, corpo)
}

func CarregarDados(w http.ResponseWriter, r *http.Request) {
	json, erro := seguranca.CarregarDados(r)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, json)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	seguranca.Deletar(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func getFranqDadosById(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Monta a url para requisição na API
	url := fmt.Sprintf(`%s/v4/franqueado/getDadosById`, config.ApiUrl)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	corpo, erro := io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Responde ao APP
	auxiliar.RespostaAPP(w, corpo)
}

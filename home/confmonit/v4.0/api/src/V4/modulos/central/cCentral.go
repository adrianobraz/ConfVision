package centralV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func logar(w http.ResponseWriter, r *http.Request) {

	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber e manipular as informações do login
	var obj struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var login centralLogin

	// Efetua o login
	if err := login.logar(obj.Email, obj.Senha); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	login.Senha = ""
	// Envia as credenciais para o app solicitante
	respApp.Dados(w, http.StatusOK, login)
}

func logarAdministrador(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var obj struct {
		Usuario string `json:"usuario"`
		Email   string `json:"email"`
		Senha   string `json:"senha"`
	}
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	usuario := obj.Usuario
	if usuario == "" {
		usuario = obj.Email
	}

	var login centralLogin
	if err := login.logarAdministrador(usuario, obj.Senha); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, login)
}

func lista(w http.ResponseWriter, r *http.Request) {
	var c Central
	var lista []Central
	if err := c.Lista(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(lista) == 0 {
		respApp.Vazio(w)
		return
	}
	respApp.Dados(w, http.StatusOK, lista)
}

func insere(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var c Central
	if err := json.Unmarshal(body, &c); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := c.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, c)
}

func atualiza(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var c Central
	if err := json.Unmarshal(body, &c); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := c.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, c)
}

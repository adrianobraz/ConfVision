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

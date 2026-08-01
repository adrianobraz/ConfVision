package admfinanceiroV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func logar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var obj struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	// Aceita tambem "usuario" (alias de email) no body
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	if v, ok := raw["email"].(string); ok {
		obj.Email = v
	}
	if obj.Email == "" {
		if v, ok := raw["usuario"].(string); ok {
			obj.Email = v
		}
	}
	if v, ok := raw["senha"].(string); ok {
		obj.Senha = v
	}

	var login admLogin
	if err := login.logar(obj.Email, obj.Senha); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, login)
}

func setFlag(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var obj struct {
		IDUsuario     string `json:"idUsuario"`
		AdmFinanceiro string `json:"admFinanceiro"`
	}
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := setAdmFinanceiro(obj.IDUsuario, obj.AdmFinanceiro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, map[string]string{
		"idUsuario":     strings.TrimSpace(obj.IDUsuario),
		"admFinanceiro": strings.ToUpper(strings.TrimSpace(obj.AdmFinanceiro)),
	})
}

package seguranca

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"webCentral/src/config"

	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

func ConfigurarCookies() {
	s = securecookie.New(config.HashKey, config.BlockKey)
}

// SalvarCookies codifica os dados e cria o cookie no navegador
func SalvarCookies(w http.ResponseWriter, idUsuario, token string) error {

	// Recebe os dados para codificar
	dados := map[string]string{
		"idUsuario": idUsuario,
		"token":     token,
	}

	// Codifica os dados do cookie
	encoded, erro := s.Encode("central", dados)
	if erro != nil {
		return erro
	}
	fmt.Println(dados)

	http.SetCookie(w, &http.Cookie{
		Name:     "central",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	})

	return nil
}

// LerCookies recupera o cookie no navegador e decodifica seus valores
func LerCookies(r *http.Request) (map[string]string, error) {
	cookie, erro := r.Cookie("central")
	if erro != nil {
		return nil, erro
	}

	dados := make(map[string]string)

	if erro = s.Decode("central", cookie.Value, &dados); erro != nil {
		return nil, erro
	}
	return dados, nil
}

// CarregarDados retorna um json com os dados do cookie
func CarregarDados(r *http.Request) ([]byte, error) {
	cookie, erro := LerCookies(r)
	if erro != nil {
		return nil, erro
	}

	var dados = struct {
		IdUsuario string `json:"idUsuario"`
		Token     string `json:"token"`
	}{
		IdUsuario: cookie["idUsuario"],
		Token:     cookie["token"],
	}

	retJson, erro := json.Marshal(dados)
	if erro != nil {
		return nil, nil
	}
	return retJson, nil
}

// Deletar destroi o cookie do navegador
func Deletar(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "central",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
}

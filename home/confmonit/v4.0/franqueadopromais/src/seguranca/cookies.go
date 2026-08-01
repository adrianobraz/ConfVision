package seguranca

import (
	"franqueadopro/src/config"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

func ConfigurarCookies() {
	s = securecookie.New(config.HashKey, config.BlockKey)
}

//login.IdUsuario, login.NivelAcesso, login.NomeUsuario,
//login.Email, login.IdVinculo, login.Token,

func SalvarCookies(w http.ResponseWriter, idUsuario, token string) error {
	return SalvarCookiesCompleto(w, idUsuario, token, "", "")
}

func SalvarCookiesCompleto(w http.ResponseWriter, idUsuario, token, master, idFranqueado string) error {
	dados := map[string]string{
		"idUsuario":    idUsuario,
		"token":        token,
		"master":       master,
		"idFranqueado": idFranqueado,
	}

	dadosCodificados, erro := s.Encode("franqueado", dados)
	if erro != nil {
		return erro
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "franqueado",
		Value:    dadosCodificados,
		Path:     "/",
		HttpOnly: true,
	})

	return nil
}

func LerCookies(r *http.Request) (map[string]string, error) {
	cookie, erro := r.Cookie("franqueado")
	if erro != nil {
		return nil, erro
	}

	valores := make(map[string]string)
	if erro = s.Decode(
		"franqueado", cookie.Value, &valores,
	); erro != nil {
		return nil, erro
	}
	return valores, nil
}

func CarregarDados(r *http.Request) ([]byte, error) {
	cookie, erro := LerCookies(r)
	if erro != nil {
		return nil, erro
	}

	var dados = struct {
		IdUsuario string `json:"idUsuario"`
	}{

		IdUsuario: cookie["idUsuario"],
	}

	retJson, erro := json.Marshal(dados)
	if erro != nil {
		return nil, nil
	}

	return retJson, nil
}

func Deletar(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "franqueado",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
}

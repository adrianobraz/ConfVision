package seguranca

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/config"

	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

const cookieOperador = "operador"

func ConfigurarCookies() {
	s = securecookie.New(config.HashKey, config.BlockKey)
}

func SalvarOperador(w http.ResponseWriter, dados map[string]string) error {
	encoded, err := s.Encode(cookieOperador, dados)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieOperador,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func LerOperador(r *http.Request) (map[string]string, error) {
	cookie, err := r.Cookie(cookieOperador)
	if err != nil {
		return nil, err
	}
	dados := make(map[string]string)
	if err = s.Decode(cookieOperador, cookie.Value, &dados); err != nil {
		return nil, err
	}
	return dados, nil
}

func Deletar(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieOperador,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
}

func EhMaster(r *http.Request) bool {
	op, err := LerOperador(r)
	if err != nil {
		return false
	}
	return op["userMaster"] == "S"
}

func EhOperadorTerminal(r *http.Request) bool {
	op, err := LerOperador(r)
	if err != nil {
		return false
	}
	return op["usuarioTerminal"] == "S"
}

func IdFranqueadoOperador(r *http.Request, bodyIdFranqueado string) (string, error) {
	return resolverIdFranqueado(r, bodyIdFranqueado, "")
}

// ResolverIdFranqueado — FRA usa vínculo; CEN/REP exigem body ou derivam do idCliente (MySQL).
func ResolverIdFranqueado(r *http.Request, bodyIdFranqueado, idCliente string) (string, error) {
	return resolverIdFranqueado(r, bodyIdFranqueado, idCliente)
}

func resolverIdFranqueado(r *http.Request, bodyIdFranqueado, idCliente string) (string, error) {
	op, err := LerOperador(r)
	if err != nil {
		return "", err
	}
	bodyIdFranqueado = strings.TrimSpace(bodyIdFranqueado)
	idCliente = strings.TrimSpace(idCliente)

	switch op["userTipo"] {
	case "FRA":
		return strings.TrimSpace(op["idVinculo"]), nil
	case "CEN", "REP":
		vinculo := strings.TrimSpace(op["idVinculo"])
		if bodyIdFranqueado != "" && bodyIdFranqueado != vinculo {
			return bodyIdFranqueado, nil
		}
		if idCliente != "" {
			return auxiliar.FranqueadoDoCliente(idCliente)
		}
		return "", errors.New("selecione um franqueado")
	default:
		if bodyIdFranqueado != "" {
			return bodyIdFranqueado, nil
		}
		return strings.TrimSpace(op["idVinculo"]), nil
	}
}

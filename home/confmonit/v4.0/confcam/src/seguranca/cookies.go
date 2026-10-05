package seguranca

import (
	"confvision/src/conexao"
	"confvision/src/config"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

func ConfigurarCookies() {
	s = securecookie.New(config.HashKey, config.BlockKey)
}

//login.IdUsuario, login.NivelAcesso, login.NomeUsuario,
//login.Email, login.IdVinculo, login.Token,

func SalvarCookies(w http.ResponseWriter, idUsuario, idVinculo, token string) error {
	return SalvarSessao(w, Sessao{
		IdUsuario: idUsuario,
		IdVinculo: idVinculo,
		Token:     token,
		Tipo:      "FRA",
	})
}

// Sessao — sessão ConfVision (FRA = franqueado, CLI = cliente).
type Sessao struct {
	IdUsuario string
	IdVinculo string
	Token     string
	Tipo      string
	IdCliente string
}

func SalvarSessao(w http.ResponseWriter, sess Sessao) error {
	tipo := strings.ToUpper(strings.TrimSpace(sess.Tipo))
	if tipo == "" {
		tipo = "FRA"
	}
	dados := map[string]string{
		"idUsuario": sess.IdUsuario,
		"idVinculo": sess.IdVinculo,
		"token":     sess.Token,
		"tipo":      tipo,
		"idCliente": strings.TrimSpace(sess.IdCliente),
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
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
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

// IdFranqueadoDoCookie — no ConfVision o tenant fica em idVinculo (não idFranqueado).
func IdFranqueadoDoCookie(cookie map[string]string) string {
	if cookie == nil {
		return ""
	}
	if v := strings.TrimSpace(cookie["idVinculo"]); v != "" {
		return v
	}
	return strings.TrimSpace(cookie["idFranqueado"])
}

func TipoSessao(cookie map[string]string) string {
	if cookie == nil {
		return ""
	}
	t := strings.ToUpper(strings.TrimSpace(cookie["tipo"]))
	if t == "" {
		return "FRA"
	}
	return t
}

func EhCliente(cookie map[string]string) bool {
	return TipoSessao(cookie) == "CLI"
}

func EhAdministrator(cookie map[string]string) bool {
	return TipoSessao(cookie) == "ADM"
}

func IdClienteDoCookie(cookie map[string]string) string {
	if cookie == nil {
		return ""
	}
	return strings.TrimSpace(cookie["idCliente"])
}

// TextoDoPayload lê string de map JSON (string, número ou outro).
func TextoDoPayload(payload map[string]any, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case string:
			if s := strings.TrimSpace(v); s != "" {
				return s
			}
		case float64:
			return strings.TrimSpace(fmt.Sprintf("%.0f", v))
		case json.Number:
			if s := strings.TrimSpace(v.String()); s != "" {
				return s
			}
		default:
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func idFranqueadoByUsuario(idUsuario string) string {
	idUsuario = strings.TrimSpace(idUsuario)
	if idUsuario == "" {
		return ""
	}
	db, err := conexao.Conectar()
	if err != nil {
		return ""
	}
	defer db.Close()

	var idVinculo sql.NullString
	if err := db.QueryRow(`
		SELECT ID_Vinculo
		FROM usuarios
		WHERE ID_Usuario = ?
		LIMIT 1
	`, idUsuario).Scan(&idVinculo); err != nil {
		return ""
	}
	return strings.TrimSpace(idVinculo.String)
}

// ResolverIdFranqueado: body (não vazio) > cookie idVinculo > MySQL via idUsuario.
func ResolverIdFranqueado(cookie map[string]string, payload map[string]any) string {
	if id := TextoDoPayload(payload, "id_franqueado", "idFranqueado"); id != "" {
		return id
	}
	if id := IdFranqueadoDoCookie(cookie); id != "" {
		return id
	}
	if cookie != nil {
		return idFranqueadoByUsuario(cookie["idUsuario"])
	}
	return ""
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
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
}

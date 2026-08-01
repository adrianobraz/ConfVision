package login

import (
	"confvision/src/conexao"
	"confvision/src/seguranca"
	"confvision/src/xanopro"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func getUsaConfVision(idFranqueado string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return "", errors.New("id do franqueado invalido")
	}

	db, erro := conexao.Conectar()
	if erro != nil {
		return "", fmt.Errorf("falha ao conectar no banco: %w", erro)
	}
	defer db.Close()

	var usa sql.NullString
	erro = db.QueryRow(`
		SELECT UsaConfVision
		FROM franqueado
		WHERE ID_Franqueado = ?
		LIMIT 1
	`, idFranqueado).Scan(&usa)

	if erro != nil {
		if errors.Is(erro, sql.ErrNoRows) {
			return "", errors.New("franqueado nao encontrado")
		}
		return "", erro
	}

	valor := strings.ToUpper(strings.TrimSpace(usa.String))
	if valor == "" {
		return "N", nil
	}

	return valor, nil
}

// confVisionLiberadoXano consulta a fonte canonica:
// plano ConfVision, bundle FP Pro+ ou licenca de camera paga.
// Retorna (liberado, ok) — ok=false se Xano indisponivel.
func confVisionLiberadoXano(idFranqueado string) (liberado bool, ok bool) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return false, true
	}

	raw, err := xanopro.Post("/fp_confvision_resumo_franqueado", map[string]any{
		"id_franqueado": idFranqueado,
	})
	if err != nil || len(raw) == 0 {
		return false, false
	}

	var resp struct {
		Liberado      bool   `json:"liberado"`
		UsaConfVision string `json:"usa_confvision"`
	}
	if json.Unmarshal(raw, &resp) != nil {
		// resposta pode vir aninhada em alguns proxies
		var m map[string]any
		if json.Unmarshal(raw, &m) != nil {
			return false, false
		}
		if v, exists := m["liberado"]; exists {
			switch t := v.(type) {
			case bool:
				return t, true
			case string:
				return strings.EqualFold(t, "true") || t == "1" || strings.EqualFold(t, "S"), true
			}
		}
		if v, exists := m["usa_confvision"]; exists {
			s := strings.ToUpper(strings.TrimSpace(fmt.Sprint(v)))
			return s == "S", true
		}
		return false, false
	}

	if resp.Liberado {
		return true, true
	}
	usa := strings.ToUpper(strings.TrimSpace(resp.UsaConfVision))
	return usa == "S", true
}

// confVisionAcessoPermitido: Xano primeiro; se indisponivel, flag MySQL legada.
func confVisionAcessoPermitido(idFranqueado string) (bool, error) {
	if liberado, ok := confVisionLiberadoXano(idFranqueado); ok {
		return liberado, nil
	}

	usa, err := getUsaConfVision(idFranqueado)
	if err != nil {
		return false, err
	}
	return usa == "S", nil
}

func getIdFranqueadoByUsuario(idUsuario string) (string, error) {
	idUsuario = strings.TrimSpace(idUsuario)
	if idUsuario == "" {
		return "", errors.New("id do usuario invalido")
	}

	db, erro := conexao.Conectar()
	if erro != nil {
		return "", fmt.Errorf("falha ao conectar no banco: %w", erro)
	}
	defer db.Close()

	var idVinculo sql.NullString
	erro = db.QueryRow(`
		SELECT ID_Vinculo
		FROM usuarios
		WHERE ID_Usuario = ?
		LIMIT 1
	`, idUsuario).Scan(&idVinculo)
	if erro != nil {
		if errors.Is(erro, sql.ErrNoRows) {
			return "", errors.New("usuario nao encontrado")
		}
		return "", erro
	}

	valor := strings.TrimSpace(idVinculo.String)
	if valor == "" {
		return "", errors.New("usuario sem vinculo de franqueado")
	}

	return valor, nil
}

func idFranqueadoDaSessao(cookie map[string]string) string {
	if cookie == nil {
		return ""
	}

	idVinculo := strings.TrimSpace(cookie["idVinculo"])
	if idVinculo != "" {
		return idVinculo
	}

	idUsuario := strings.TrimSpace(cookie["idUsuario"])
	if idUsuario == "" {
		return ""
	}

	if fraId, erro := getIdFranqueadoByUsuario(idUsuario); erro == nil {
		return fraId
	}

	return ""
}

func VerificarAcessoConfVision(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/logout" {
			next(w, r)
			return
		}

		cookie, erro := seguranca.LerCookies(r)
		if erro != nil {
			next(w, r)
			return
		}

		if seguranca.EhAdministrator(cookie) {
			next(w, r)
			return
		}

		idFranqueado := idFranqueadoDaSessao(cookie)
		if idFranqueado == "" {
			next(w, r)
			return
		}

		ok, erro := confVisionAcessoPermitido(idFranqueado)
		if erro != nil || !ok {
			seguranca.Deletar(w)
			if strings.Contains(r.Header.Get("Accept"), "application/json") || strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, `{"status":"modulo ConfVision nao contratado"}`, http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r)
	}
}

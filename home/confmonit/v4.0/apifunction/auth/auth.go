package auth

import (
	"apifunction/config"
	"apifunction/db"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

type UsuarioAdm struct {
	IDUsuario     string
	IDVinculo     string
	Master        string
	AdmFinanceiro string
}

func ExtrairBearer(r *http.Request) (string, error) {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		return "", errors.New("authorization obrigatorio")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("authorization invalido")
	}
	tok := strings.TrimSpace(parts[1])
	if tok == "" {
		return "", errors.New("token vazio")
	}
	return tok, nil
}

func IDUsuarioDoToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de assinatura inesperado")
		}
		return config.APIKey, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("token invalido")
	}
	id := strings.TrimSpace(fmt.Sprintf("%v", claims["id"]))
	if id == "" || id == "<nil>" {
		return "", errors.New("token sem id de usuario")
	}
	return id, nil
}

func RequireAdm(r *http.Request) (UsuarioAdm, error) {
	tok, err := ExtrairBearer(r)
	if err != nil {
		return UsuarioAdm{}, err
	}
	idUsuario, err := IDUsuarioDoToken(tok)
	if err != nil {
		return UsuarioAdm{}, err
	}

	var u UsuarioAdm
	var master, admFin sql.NullString
	err = db.Conn.QueryRow(`
		SELECT
			usuarios.ID_Usuario,
			usuarios.ID_Vinculo,
			COALESCE(usuarios.Master, 'N'),
			COALESCE(usuarios.AdmFinanceiro, 'N')
		FROM usuarios
		LEFT JOIN listaBloqueio AS bl ON usuarios.ID_Usuario = bl.ID_Alvo
		WHERE usuarios.ID_Usuario = ?
		  AND bl.ID_Alvo IS NULL
		LIMIT 1
	`, idUsuario).Scan(&u.IDUsuario, &u.IDVinculo, &master, &admFin)
	if err == sql.ErrNoRows {
		return UsuarioAdm{}, errors.New("usuario nao encontrado ou bloqueado")
	}
	if err != nil {
		return UsuarioAdm{}, err
	}
	u.Master = strings.ToUpper(strings.TrimSpace(master.String))
	u.AdmFinanceiro = strings.ToUpper(strings.TrimSpace(admFin.String))
	if u.Master != "S" {
		return UsuarioAdm{}, errors.New("acesso exige usuario Master")
	}
	if u.AdmFinanceiro != "S" {
		return UsuarioAdm{}, errors.New("acesso exige AdmFinanceiro = S")
	}
	return u, nil
}

package seguranca

import (
	"api/src/V4/config"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var SecretKey = "xqH3jc53EAsYY3kpFb8V0NeiBmzkOHllIBQZ6BJQbKCZTPgPGsCzBcJygTC7giBpB0/XL+0qFmfqN31yraR4CA=="

func CriarToken(id string) (string, error) {
	permicoes := jwt.MapClaims{}
	permicoes["authorized"] = true                           // autorizado
	permicoes["exp"] = time.Now().Add(time.Hour * 12).Unix() // Expira
	//permicoes["id"] = "magno"
	permicoes["id"] = id
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, permicoes)
	return token.SignedString(config.SecretKey) // chave secret
}

func ValidarToken(r *http.Request) error {
	tokenString := extrairToken(r)
	token, erro := jwt.Parse(tokenString, retornarChaveVerificacao)
	if erro != nil {
		return fmt.Errorf("token -> %s", erro.Error())
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	}
	return errors.New("tokem inválido")
}

func extrairToken(r *http.Request) string {
	token := r.Header.Get("Authorization")

	// Verifica o tamanho da palavra se = a 2
	if len(strings.Split(token, " ")) == 2 {
		return strings.Split(token, " ")[1]
	}
	return ""
}

func retornarChaveVerificacao(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

		return nil, fmt.Errorf("método de assinatura inesperado! %v", token.Header["alg"])
	}
	return config.SecretKey, nil
}

/*
func ExtrairIdUsuario(r *http.Request) (string, error) {
	tokenString := extrairToken(r)
	token, erro := jwt.Parse(tokenString, retornarChaveVerificacao)
	if erro != nil {
		return "", erro
	}

	if permicaoes, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idUsuario := fmt.Sprintf("%v", permicaoes["usuarioId"])
		if erro != nil {
			return "", nil
		}
		return idUsuario, nil
	}
	return "", errors.New("token inválido")
}
*/

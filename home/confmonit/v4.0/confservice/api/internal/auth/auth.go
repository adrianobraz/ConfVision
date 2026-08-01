package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"confservice/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	ParceiroID string `json:"parceiroId"`
	Email      string `json:"email"`
	Role       string `json:"role"` // parceiro | admin | franqueado
	jwt.RegisteredClaims
}

func HashSenha(senha string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	return string(b), err
}

func ChecarSenha(hash, senha string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)) == nil
}

func EmitirToken(parceiroID, email, role string, dur time.Duration) (string, error) {
	claims := Claims{
		ParceiroID: parceiroID,
		Email:      email,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(dur)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(config.JWTSecret))
}

func ParseToken(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("token invalido")
	}
	return claims, nil
}

func Bearer(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("sem authorization")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("authorization invalido")
	}
	return strings.TrimSpace(parts[1]), nil
}

func RequireAPIKey(r *http.Request) bool {
	if config.APIKey == "" {
		return false
	}
	return r.Header.Get("X-Api-Key") == config.APIKey
}

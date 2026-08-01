package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	Porta             string
	APIKey            string
	JWTSecret         string
	MySQLDSN          string
	CORSOrigins       []string
	ComissaoPadraoPct float64
	AdminUser         string
	AdminPass         string
)

func Carregar() {
	_ = godotenv.Load()

	Porta = env("PORTA", "2020")
	APIKey = os.Getenv("API_KEY")
	JWTSecret = env("JWT_SECRET", "dev-jwt-secret-change-me")
	MySQLDSN = os.Getenv("MYSQL_DSN")
	if MySQLDSN == "" {
		log.Fatal("MYSQL_DSN obrigatorio no .env")
	}

	origins := env("CORS_ORIGINS", "http://localhost:5173")
	for _, o := range strings.Split(origins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			CORSOrigins = append(CORSOrigins, o)
		}
	}

	pct, err := strconv.ParseFloat(env("COMISSAO_PADRAO_PCT", "20"), 64)
	if err != nil {
		pct = 20
	}
	ComissaoPadraoPct = pct

	AdminUser = env("ADMIN_USER", "admin")
	AdminPass = env("ADMIN_PASS", "admin")

	if APIKey == "" {
		log.Println("aviso: API_KEY vazio — rotas internas /internal/* ficam abertas so se autenticadas por JWT")
	}
}

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

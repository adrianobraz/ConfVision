package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	Api         string
	Porta       string
	HashKey     []byte
	BlockKey    []byte
	SiteHttps   bool
	Certificado string
	Chave       string
	TituloSite  string
)

var RecWeb struct {
	Url   string
	Senha string
}

var RecComando struct {
	Url   string
	Senha string
}

func Configurar() {
	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	Porta = os.Getenv("PORTA")

	Api = os.Getenv("API")

	RecWeb.Url = os.Getenv("REC_WEB_URL")
	RecWeb.Senha = os.Getenv("REC_WEB_SENHA")

	RecComando.Url = os.Getenv("REC_COMANDO_URL")
	RecComando.Senha = os.Getenv("REC_COMANDO_SENHA")

	HashKey = []byte(os.Getenv("HASHKEY"))
	BlockKey = []byte(os.Getenv("BLOCKKEY"))

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}

	TituloSite = "WEB CLIENTE"
}

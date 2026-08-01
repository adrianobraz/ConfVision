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
)

func Configurar() {
	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	Porta = os.Getenv("PORTA")

	Api = os.Getenv("API")

	HashKey = []byte(os.Getenv("HASHKEY"))
	BlockKey = []byte(os.Getenv("BLOCKKEY"))

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}
}

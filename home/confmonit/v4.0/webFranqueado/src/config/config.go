package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	TituloSite  string
	ApiUrl      string
	WebPorta    string
	HashKey     []byte
	BlockKey    []byte
	SiteHttps   bool
	Certificado string
	Chave       string
	UrlReceptor string
	UrlComando  string

	Manual struct {
		GeradorEvento string
	}
)

func ConfigurarApp() {
	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	TituloSite = os.Getenv("TITULO_SITE")
	ApiUrl = os.Getenv("URL_API")
	WebPorta = os.Getenv("PORTA")
	HashKey = []byte(os.Getenv("HASHKEY"))
	BlockKey = []byte(os.Getenv("BLOCKKEY"))

	UrlReceptor = os.Getenv("URL_RECEPTOR")
	UrlComando = os.Getenv("URL_COMANDO")

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}

	Manual.GeradorEvento = os.Getenv("GERADOR_EVENTO")

}

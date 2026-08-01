package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	Porta          string
	ApiKey         string
	ProxyTarget    string
	ProxyTargetCV  string
	ApacheSites    string
	CertbotEmail   string
)

func Carregar() {
	_ = godotenv.Load()

	Porta = os.Getenv("PORTA")
	if Porta == "" {
		Porta = "2015"
	}
	ApiKey = os.Getenv("API_KEY")
	ProxyTarget = os.Getenv("PROXY_TARGET")
	if ProxyTarget == "" {
		ProxyTarget = "http://127.0.0.1:2005/"
	}
	ProxyTargetCV = os.Getenv("PROXY_TARGET_CONFVISION")
	if ProxyTargetCV == "" {
		// ConfVision em producao usa PORTA=8086
		ProxyTargetCV = "http://127.0.0.1:8086/"
	}
	ApacheSites = os.Getenv("APACHE_SITES_DIR")
	if ApacheSites == "" {
		ApacheSites = "/etc/apache2/sites-available"
	}
	CertbotEmail = os.Getenv("CERTBOT_EMAIL")
	if CertbotEmail == "" {
		CertbotEmail = "suporte@kitsite.com.br"
	}

	if ApiKey == "" {
		log.Fatal("API_KEY obrigatorio no .env do fp-dominio-provision")
	}
}

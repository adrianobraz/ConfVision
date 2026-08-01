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
	XanoApi     string
	XanoApiPro  string
	XanoApiCentralDisparos string
	ElevenLabsAPIKey       string

	ProvisionerURL string
	ProvisionerKey string

	// ConfService (parceiros de monitoramento / webhook)
	ConfServiceURL string
	ConfServiceKey string

	// Banco MySQL (mesmo do API V4) — usado no relatorio de ligacoes
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

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
	XanoApi = os.Getenv("XANO_API_FRANQUEADO")
	XanoApiPro = os.Getenv("XANO_API_FRANQUEADO_PRO")
	XanoApiCentralDisparos = os.Getenv("XANO_API_CENTRAL_DISPAROS")
	ElevenLabsAPIKey = os.Getenv("ELEVENLABS_API_KEY")
	ProvisionerURL = os.Getenv("PROVISIONER_URL")
	ProvisionerKey = os.Getenv("PROVISIONER_KEY")
	ConfServiceURL = stringsTrim(os.Getenv("CONFSERVICE_URL"))
	if ConfServiceURL == "" {
		ConfServiceURL = "http://127.0.0.1:2020"
	}
	ConfServiceKey = stringsTrim(os.Getenv("CONFSERVICE_API_KEY"))

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}

	Manual.GeradorEvento = os.Getenv("GERADOR_EVENTO")

	DBHost = stringsTrim(os.Getenv("BD_HOST_MV4"))
	DBPort = stringsTrim(os.Getenv("BD_PORT_MV4"))
	DBUser = stringsTrim(os.Getenv("BD_USER_MV4"))
	DBPass = stringsTrim(os.Getenv("BD_PASS_MV4"))
	DBName = stringsTrim(os.Getenv("BD_BASE_MV4"))
	if DBPort == "" {
		DBPort = "3306"
	}
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[0] == '"' || s[0] == '\'') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == '"' || s[len(s)-1] == '\'') {
		s = s[:len(s)-1]
	}
	return s
}

// DBDSN monta a conexao MySQL do FranqueadoPro (relatorio de ligacoes).
func DBDSN() string {
	if DBHost == "" || DBUser == "" || DBName == "" {
		return ""
	}
	return DBUser + ":" + DBPass + "@tcp(" + DBHost + ":" + DBPort + ")/" + DBName + "?charset=utf8&parseTime=True&loc=Local"
}


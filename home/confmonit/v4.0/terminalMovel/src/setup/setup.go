package setup

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	Servidor           string
	Uri                string
	Porta              string
	StringConexaoBanco string
	ReceptorUrl        string
	ReceptorSenha      string
	CodBenuvemCentral  string
	SiteHttps          bool
	UrlReceptor        string
	Certificado        string
	Chave              string
	OpenAIApiKey       string
	OpenAIModel        string
	XanoBaseUrl        string
	XanoToken          string
	ConfVisionHlsBase  string
)

func ConfigurarApp() {
	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	// Carrega dados do servidor
	Servidor = os.Getenv("SERVIDOR")
	Uri = os.Getenv("URI")
	Porta = os.Getenv("PORTA")

	ReceptorUrl = os.Getenv("RECEPTOR_URL")
	ReceptorSenha = os.Getenv("RECEPTOR_SENHA")
	CodBenuvemCentral = os.Getenv("COD_BENUVEM_CENTRAL")
	// Carrega dados da conexão
	StringConexaoBanco = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("BD_USER"),
		os.Getenv("BD_PASS"),
		os.Getenv("BD_HOST"),
		os.Getenv("BD_PORT"),
		os.Getenv("BD_BASE"),
	)

	UrlReceptor = os.Getenv("URL_RECEPTOR")

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}

	OpenAIApiKey = os.Getenv("OPENAI_API_KEY")
	if m := os.Getenv("OPENAI_MODEL"); m != "" {
		OpenAIModel = m
	} else {
		OpenAIModel = "gpt-5-nano"
	}

	XanoBaseUrl = os.Getenv("XANO_BASE_URL")
	XanoToken = os.Getenv("XANO_TOKEN")
	if XanoToken == "" {
		XanoToken = "1e2d2eef75ccd7cfea2e06d7ce1c63c4"
	}
	ConfVisionHlsBase = os.Getenv("CONFVISION_HLS_BASE")
	if ConfVisionHlsBase == "" {
		ConfVisionHlsBase = "https://vision.confmonit2.com.br"
	}
}

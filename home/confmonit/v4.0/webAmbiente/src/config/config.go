package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	Api                string
	Xano               string
	XanoCenterOperacion string
	UrlComando         string
	SenhaWebComando    string
	XanoConfVision     string
	XanoConfVisionToken string
	ConfVisionHlsBase  string
	Porta              string
	BindHost           string
	CookieSecure       bool
	StringConexaoBanco string
	HashKey            []byte
	BlockKey           []byte
	SiteHttps          bool
	Certificado        string
	Chave              string
	TituloSite         string

	ContaboS3AccessKey string
	ContaboS3SecretKey string
	ContaboS3Endpoint  string
	ContaboS3Bucket    string
	ContaboS3Region    string
	ContaboS3TenantId  string

	ContaboMapaMaxBytes       int
	ContaboMapaMaxWidth       int
	ContaboMapaMaxUploadBytes int
)

func Configurar() {
	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	Porta = os.Getenv("PORTA")
	BindHost = os.Getenv("BIND_HOST")
	CookieSecure = os.Getenv("COOKIE_SECURE") == "SIM"
	Api = os.Getenv("API")
	Xano = os.Getenv("XANO_API")
	XanoCenterOperacion = strings.TrimSpace(os.Getenv("XANO_CENTER_OPERACION_API"))
	if XanoCenterOperacion == "" {
		XanoCenterOperacion = "https://xpcy-oyme-lno7.b2.xano.io/api:CiSZf6eF"
	}
	UrlComando = strings.TrimSpace(os.Getenv("URL_COMANDO"))
	if UrlComando == "" && Api != "" {
		UrlComando = strings.TrimRight(Api, "/") + "/armar"
	}
	SenhaWebComando = os.Getenv("SENHA_WEB_COMANDO")
	XanoConfVision = strings.TrimSpace(os.Getenv("XANO_CONFVISION_API"))
	XanoConfVisionToken = strings.TrimSpace(os.Getenv("XANO_CONFVISION_TOKEN"))
	ConfVisionHlsBase = strings.TrimRight(strings.TrimSpace(os.Getenv("CONFVISION_HLS_BASE")), "/")

	StringConexaoBanco = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("BD_USER"),
		os.Getenv("BD_PASS"),
		os.Getenv("BD_HOST"),
		os.Getenv("BD_PORT"),
		os.Getenv("BD_BASE"),
	)

	HashKey = []byte(os.Getenv("HASHKEY"))
	BlockKey = []byte(os.Getenv("BLOCKKEY"))

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}

	TituloSite = os.Getenv("TITULO_SITE")
	if TituloSite == "" {
		TituloSite = "ConfMonit Ambiente"
	}

	ContaboS3AccessKey = os.Getenv("CONTABO_S3_ACCESS_KEY")
	ContaboS3SecretKey = os.Getenv("CONTABO_S3_SECRET_KEY")
	ContaboS3Endpoint = os.Getenv("CONTABO_S3_ENDPOINT")
	if ContaboS3Endpoint == "" {
		ContaboS3Endpoint = "https://usc1.contabostorage.com"
	}
	ContaboS3Bucket = os.Getenv("CONTABO_S3_BUCKET")
	if ContaboS3Bucket == "" {
		ContaboS3Bucket = "mapa"
	}
	ContaboS3Region = os.Getenv("CONTABO_S3_REGION")
	if ContaboS3Region == "" {
		ContaboS3Region = "us-east-1"
	}
	ContaboS3TenantId = strings.TrimSpace(os.Getenv("CONTABO_S3_TENANT_ID"))

	ContaboMapaMaxBytes = envInt("CONTABO_MAPA_MAX_BYTES", 153600)
	ContaboMapaMaxWidth = envInt("CONTABO_MAPA_MAX_WIDTH", 1920)
	ContaboMapaMaxUploadBytes = envInt("CONTABO_MAPA_MAX_UPLOAD_BYTES", 10<<20)
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

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
	TituloSite            string
	ApiUrl                string
	ApiUrlSetor           string
	XanoBaseUrl             string
	XanoCvgBaseUrl          string
	XanoApiPro              string
	ProvisionerURL          string
	ProvisionerKey          string
	MediamtxHlsBase         string
	MediamtxRtmpPublishBase string
	MediamtxRtspBase        string
	MediamtxHlsPublic       string
	MediamtxRtmpPublic      string
	RtmpWatchURL            string
	RtmpGuardURL            string
	RtmpGuardAdminKey       string
	RtmpPublishSecret       string
	AdministratorUser       string
	AdministratorPass       string
	WebPorta                string
	BindHost                string
	CookieSecure            bool
	HashKey                 []byte
	BlockKey                []byte
	SiteHttps               bool
	Certificado             string
	Chave                   string
	UrlReceptor             string
	UrlComando              string

	ReceptorWebURL          string
	ReceptorWebSenha        string
	TerminalNotifyEnabled   bool
	TerminalNotifyRetries   int
	TerminalContactID       string

	ImagemPublicSecret      string
	ImagemPublicBaseURL     string
	ImagemRateLimitPerMin   int
	ImagemRateLimitPerHour  int
	IntegracaoDispatchEnabled bool

	ContaboS3AccessKey string
	ContaboS3SecretKey string
	ContaboS3Endpoint  string
	ContaboS3Bucket    string
	ContaboS3Region    string
	ContaboS3TenantId  string
	CadSnapshotMaxBytes int
	CadSnapshotMaxWidth int

	Manual struct {
		GeradorEvento string
	}

	ConexaoMySQL string

	ConexaoPostgres    string
	VisPostgresEnabled bool
	VisWorkerAPIKey    string
	VisReceptorBearer  string
)

func ConfigurarApp() {
	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	TituloSite = os.Getenv("TITULO_SITE")
	ApiUrl = os.Getenv("URL_API")
	ApiUrlSetor = os.Getenv("URL_API_SETOR")
	XanoBaseUrl = os.Getenv("XANO_BASE_URL")
	XanoCvgBaseUrl = os.Getenv("XANO_CVG_BASE_URL")
	if XanoCvgBaseUrl == "" {
		XanoCvgBaseUrl = XanoBaseUrl
	}
	XanoApiPro = os.Getenv("XANO_API_FRANQUEADO_PRO")
	ProvisionerURL = strings.TrimSpace(os.Getenv("PROVISIONER_URL"))
	ProvisionerKey = strings.TrimSpace(os.Getenv("PROVISIONER_KEY"))
	MediamtxHlsBase = os.Getenv("MEDIAMTX_HLS_BASE")
	MediamtxRtmpPublishBase = os.Getenv("MEDIAMTX_RTMP_PUBLISH_BASE")
	MediamtxRtspBase = os.Getenv("MEDIAMTX_RTSP_BASE")
	if MediamtxRtspBase == "" {
		MediamtxRtspBase = "rtsp://127.0.0.1:8554"
	}
	MediamtxHlsPublic = os.Getenv("MEDIAMTX_HLS_PUBLIC")
	if MediamtxHlsPublic == "" {
		MediamtxHlsPublic = MediamtxHlsBase
	}
	MediamtxRtmpPublic = os.Getenv("MEDIAMTX_RTMP_PUBLIC")
	if MediamtxRtmpPublic == "" {
		MediamtxRtmpPublic = MediamtxRtmpPublishBase
	}
	RtmpWatchURL = strings.TrimRight(strings.TrimSpace(os.Getenv("RTMP_WATCH_URL")), "/")
	RtmpGuardURL = strings.TrimRight(strings.TrimSpace(os.Getenv("RTMP_GUARD_URL")), "/")
	RtmpGuardAdminKey = strings.TrimSpace(os.Getenv("RTMP_GUARD_ADMIN_KEY"))
	RtmpPublishSecret = strings.TrimSpace(os.Getenv("RTMP_PUBLISH_SECRET"))
	AdministratorUser = strings.TrimSpace(os.Getenv("ADMINISTRATOR_USER"))
	AdministratorPass = strings.TrimSpace(os.Getenv("ADMINISTRATOR_PASS"))
	WebPorta = os.Getenv("PORTA")
	BindHost = os.Getenv("BIND_HOST")
	CookieSecure = os.Getenv("COOKIE_SECURE") == "SIM"
	HashKey = []byte(os.Getenv("HASHKEY"))
	BlockKey = []byte(os.Getenv("BLOCKKEY"))

	UrlReceptor = os.Getenv("URL_RECEPTOR")
	UrlComando = os.Getenv("URL_COMANDO")

	ReceptorWebURL = strings.TrimRight(strings.TrimSpace(os.Getenv("RECEPTOR_WEB_URL")), "/")
	if ReceptorWebURL == "" {
		ReceptorWebURL = strings.TrimRight(strings.TrimSpace(UrlReceptor), "/")
	}
	ReceptorWebSenha = strings.TrimSpace(os.Getenv("RECEPTOR_WEB_SENHA"))
	TerminalNotifyEnabled = envBool("TERMINAL_NOTIFY_ENABLED", true)
	TerminalNotifyRetries = envInt("TERMINAL_NOTIFY_RETRIES", 3)
	TerminalContactID = strings.TrimSpace(os.Getenv("TERMINAL_CONTACT_ID"))
	if TerminalContactID == "" {
		TerminalContactID = "CV01"
	}

	ImagemPublicSecret = strings.TrimSpace(os.Getenv("IMAGEM_PUBLIC_SECRET"))
	ImagemPublicBaseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("IMAGEM_PUBLIC_BASE_URL")), "/")
	ImagemRateLimitPerMin = envInt("IMAGEM_RATE_LIMIT_PER_MIN", 60)
	ImagemRateLimitPerHour = envInt("IMAGEM_RATE_LIMIT_PER_HOUR", 500)
	IntegracaoDispatchEnabled = envBool("INTEGRACAO_DISPATCH_ENABLED", true)

	if os.Getenv("HTTPS") == "SIM" {
		SiteHttps = true
		Certificado = os.Getenv("CERTIFICADO")
		Chave = os.Getenv("CHAVE")
	} else {
		SiteHttps = false
	}

	Manual.GeradorEvento = os.Getenv("GERADOR_EVENTO")

	ContaboS3AccessKey = os.Getenv("CONTABO_S3_ACCESS_KEY")
	ContaboS3SecretKey = os.Getenv("CONTABO_S3_SECRET_KEY")
	ContaboS3Endpoint = os.Getenv("CONTABO_S3_ENDPOINT")
	if ContaboS3Endpoint == "" {
		ContaboS3Endpoint = "https://usc1.contabostorage.com"
	}
	ContaboS3Bucket = os.Getenv("CONTABO_S3_BUCKET")
	if ContaboS3Bucket == "" {
		ContaboS3Bucket = "confvision"
	}
	ContaboS3Region = os.Getenv("CONTABO_S3_REGION")
	if ContaboS3Region == "" {
		ContaboS3Region = "us-east-1"
	}
	ContaboS3TenantId = strings.TrimSpace(os.Getenv("CONTABO_S3_TENANT_ID"))
	CadSnapshotMaxBytes = envInt("CAD_SNAPSHOT_MAX_BYTES", 51200)
	CadSnapshotMaxWidth = envInt("CAD_SNAPSHOT_MAX_WIDTH", 1280)

	ConexaoMySQL = montarConexaoMySQL()
	ConexaoPostgres = montarConexaoPostgres()
	VisPostgresEnabled = ConexaoPostgres != ""
	VisWorkerAPIKey = trimEnv(os.Getenv("VIS_WORKER_API_KEY"))
	VisReceptorBearer = trimEnv(os.Getenv("VIS_RECEPTOR_BEARER"))
	if VisReceptorBearer == "" {
		VisReceptorBearer = "1e2d2eef75ccd7cfea2e06d7ce1c63c4"
	}
}

func trimEnv(v string) string {
	return strings.TrimSpace(strings.Trim(v, "\r"))
}

func montarConexaoPostgres() string {
	if u := trimEnv(os.Getenv("POSTGRES_URL")); u != "" {
		return u
	}

	user := trimEnv(envOu("PG_USER", "PG_USER_CONFMONIT"))
	pass := trimEnv(envOu("PG_PASS", "PG_PASS_CONFMONIT"))
	host := trimEnv(envOu("PG_HOST", "PG_HOST_CONFMONIT"))
	port := trimEnv(envOu("PG_PORT", "PG_PORT_CONFMONIT"))
	base := trimEnv(envOu("PG_BASE", "PG_BASE_CONFMONIT"))
	ssl := trimEnv(envOu("PG_SSLMODE", ""))

	if user == "" || host == "" || base == "" {
		return ""
	}
	if port == "" {
		port = "5432"
	}
	if ssl == "" {
		ssl = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, base, ssl,
	)
}

func montarConexaoMySQL() string {
	user := envOu("BD_USER", "BD_USER_MV4")
	pass := envOu("BD_PASS", "BD_PASS_MV4")
	host := envOu("BD_HOST", "BD_HOST_MV4")
	port := envOu("BD_PORT", "BD_PORT_MV4")
	base := envOu("BD_BASE", "BD_BASE_MV4")

	if user == "" || host == "" || base == "" {
		return ""
	}

	if port == "" {
		port = "3306"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user, pass, host, port, base,
	)
}

func envOu(principal, alternativa string) string {
	if v := strings.TrimSpace(os.Getenv(principal)); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(alternativa))
}

func envInt(name string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envBool(name string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on", "sim", "s":
		return true
	case "0", "false", "no", "off", "nao", "n":
		return false
	default:
		return fallback
	}
}

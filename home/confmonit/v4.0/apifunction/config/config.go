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
	Porta              string
	APIKey             []byte
	MySQLDSN           string
	PostgresDSN        string
	CORSOrigins        []string
	XanoAPIFinanceiro   string
	XanoMetaAccessToken string
	XanoMetaBaseURL     string
	XanoMetaWorkspaceID int
	WorkerSecret       string
	WorkerAdminToken   string
	GovDiasGraca       int
	ConfServiceURL     string
	ConfServiceAPIKey  string

	CloudflareAPIToken string
	CloudflareZoneID   string
	ReceptorDNSZona    string
	ReceptorDNSIP      string
	FaturaListarFonte  string

	ReceptorDiagAgentURL     string
	ReceptorDiagAgentKey     string
	ReceptorDiagLabKey       string
	ApiFunctionInternalURL   string

	ProvisionerURL      string
	ProvisionerKey      string
	DominioMarcaDNSIP   string
)

func Carregar() {
	_ = godotenv.Load()

	Porta = env("APIFUNCTION_PORT", env("PORTA", "20001"))
	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatal("API_KEY obrigatorio no .env (mesma chave JWT da API V4)")
	}
	APIKey = []byte(key)

	user := os.Getenv("BD_USER_MV4")
	pass := os.Getenv("BD_PASS_MV4")
	host := env("BD_HOST_MV4", "127.0.0.1")
	port := env("BD_PORT_MV4", "3306")
	base := env("BD_BASE_MV4", "confmonitV4")
	if user == "" || base == "" {
		log.Fatal("BD_USER_MV4 e BD_BASE_MV4 obrigatorios no .env")
	}
	MySQLDSN = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user, pass, host, port, base,
	)

	for _, o := range strings.Split(env("CORS_ORIGINS", "*"), ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			CORSOrigins = append(CORSOrigins, o)
		}
	}

	XanoAPIFinanceiro = strings.TrimRight(strings.TrimSpace(os.Getenv("XANO_API_FINANCEIRO")), "/")
	XanoMetaAccessToken = strings.TrimSpace(os.Getenv("XANO_META_ACCESS_TOKEN"))
	XanoMetaBaseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("XANO_META_BASE_URL")), "/")
	XanoMetaWorkspaceID = envInt("XANO_META_WORKSPACE_ID", 1)
	if XanoMetaBaseURL == "" && XanoAPIFinanceiro != "" {
		if i := strings.Index(XanoAPIFinanceiro, "/api:"); i >= 0 {
			XanoMetaBaseURL = XanoAPIFinanceiro[:i] + "/api:meta"
		}
	}
	WorkerSecret = strings.TrimSpace(os.Getenv("WORKER_SECRET"))
	WorkerAdminToken = strings.TrimSpace(os.Getenv("WORKER_ADMIN_TOKEN"))
	PostgresDSN = strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	GovDiasGraca = envInt("GOV_DIAS_GRACA", 6)
	ConfServiceURL = strings.TrimRight(strings.TrimSpace(os.Getenv("CONFSERVICE_URL")), "/")
	ConfServiceAPIKey = strings.TrimSpace(os.Getenv("CONFSERVICE_API_KEY"))

	CloudflareAPIToken = strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN"))
	CloudflareZoneID = strings.TrimSpace(os.Getenv("CLOUDFLARE_ZONE_ID"))
	ReceptorDNSZona = env("RECEPTOR_DNS_ZONA", "dnsid.com.br")
	ReceptorDNSIP = env("RECEPTOR_DNS_IP", "185.130.61.3")
	FaturaListarFonte = strings.ToLower(env("FATURA_LISTAR_FONTE", "auto"))

	ReceptorDiagAgentURL = strings.TrimRight(strings.TrimSpace(os.Getenv("RECEPTOR_DIAG_AGENT_URL")), "/")
	ReceptorDiagAgentKey = strings.TrimSpace(os.Getenv("RECEPTOR_DIAG_AGENT_KEY"))
	ReceptorDiagLabKey = strings.TrimSpace(os.Getenv("RECEPTOR_DIAG_LAB_KEY"))
	ApiFunctionInternalURL = strings.TrimRight(strings.TrimSpace(os.Getenv("APIFUNCTION_INTERNAL_URL")), "/")

	ProvisionerURL = strings.TrimRight(strings.TrimSpace(os.Getenv("PROVISIONER_URL")), "/")
	ProvisionerKey = strings.TrimSpace(os.Getenv("PROVISIONER_KEY"))
	DominioMarcaDNSIP = env("DOMINIO_MARCA_DNS_IP", "185.130.61.4")
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func env(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

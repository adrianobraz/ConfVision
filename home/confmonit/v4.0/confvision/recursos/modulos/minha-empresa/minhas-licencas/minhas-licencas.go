package minhasLicencas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/xanopro"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/minhas-licencas",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaResumo",
		Metodo: http.MethodPost,
		Funcao: licencaResumo,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaPlanos",
		Metodo: http.MethodPost,
		Funcao: licencaPlanos,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaComprar",
		Metodo: http.MethodPost,
		Funcao: licencaComprar,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaFaturas",
		Metodo: http.MethodPost,
		Funcao: licencaFaturas,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina
		d.TituloSite = config.TituloSite
		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
		return
	}

	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Minhas licenças"
	d.LinkRetorno = "/carregar-menu-confvision"
	auxiliar.ExecutarTemplate(w, "minhas-licencas.html", d)
}

func licencaResumo(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_confvision_resumo_franqueado")
}

func licencaPlanos(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_confvision_planos_listar_franqueado")
}

func licencaComprar(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_confvision_comprar")
}

func licencaFaturas(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_fatura_minhas")
}

func proxyPro(w http.ResponseWriter, r *http.Request, path string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}

	if _, ok := payload["id_franqueado"]; !ok {
		if id, ok := payload["idFranqueado"].(string); ok && id != "" {
			payload["id_franqueado"] = id
		}
	}

	status, raw, err := xanopro.PostWithStatus(path, payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(raw)
}

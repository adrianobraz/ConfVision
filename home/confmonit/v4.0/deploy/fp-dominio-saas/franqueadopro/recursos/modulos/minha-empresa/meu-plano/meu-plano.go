package meuPlano

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/licenca"
	"franqueadopro/src/seguranca"
	"franqueadopro/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-meu-plano",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/licencaResumo",
		Metodo: http.MethodPost,
		Funcao: licencaResumo,
		Aberto: false,
	},
	{
		URI:    "/licencaMinhasFaturas",
		Metodo: http.MethodPost,
		Funcao: licencaMinhasFaturas,
		Aberto: false,
	},
	{
		URI:    "/licencaContratar",
		Metodo: http.MethodPost,
		Funcao: licencaContratar,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Meu Plano"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-principal"
	auxiliar.PreencherEhMaster(r, &d)
	auxiliar.ExecutarTemplate(w, "meu-plano.html", d)
}

func licencaResumo(w http.ResponseWriter, r *http.Request) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	raw, err := proxyProRaw(w, r, "/fp_assinatura_resumo")
	if err != nil || len(raw) == 0 {
		return
	}
	licenca.SincronizarCache(cookie["idFranqueado"], "franqueadopro", raw)
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaMinhasFaturas(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_fatura_minhas")
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaContratar(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_assinatura_contratar")
	if err != nil || len(raw) == 0 {
		return
	}
	cookie, _ := seguranca.LerCookies(r)
	if cookie != nil {
		licenca.InvalidarCache(cookie["idFranqueado"], "franqueadopro")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func proxyProRaw(w http.ResponseWriter, r *http.Request, path string) ([]byte, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return nil, err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return nil, err
	}

	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["id_franqueado"] = cookie["idFranqueado"]

	raw, err := xanopro.Post(path, payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano pro: %w", err))
		return nil, err
	}
	return raw, nil
}

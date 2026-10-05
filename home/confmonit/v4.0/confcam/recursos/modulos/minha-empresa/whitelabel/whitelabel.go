package whitelabel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"confvision/src/apifunction"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"confvision/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-whitelabel",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/cvWhitelabelCarregar",
		Metodo: http.MethodPost,
		Funcao: whitelabelCarregar,
		Aberto: false,
	},
	{
		URI:    "/cvWhitelabelSalvar",
		Metodo: http.MethodPost,
		Funcao: whitelabelSalvar,
		Aberto: false,
	},
	{
		URI:    "/cvWhitelabelByFqdn",
		Metodo: http.MethodPost,
		Funcao: whitelabelByFqdn,
		Aberto: true,
	},
	{
		URI:    "/centralMarcaPorCentral",
		Metodo: http.MethodPost,
		Funcao: marcaPorCentral,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Minha marca"
	d.LinkRetorno = "/carregar-menu-confvision"
	auxiliar.ExecutarTemplate(w, "whitelabel.html", d)
}

func whitelabelCarregar(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_whitelabel_get")
}

func whitelabelSalvar(w http.ResponseWriter, r *http.Request) {
	_, payload, err := lerPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	// ConfVision nao envia tema_json — API preserva o existente
	raw, err := xanopro.Post("/fp_whitelabel_salvar", payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func whitelabelByFqdn(w http.ResponseWriter, r *http.Request) {
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

	fqdn, _ := payload["fqdn"].(string)
	fqdn = normalizarHost(fqdn)
	if fqdn == "" {
		fqdn = normalizarHost(r.Host)
	}
	payload["fqdn"] = fqdn

	if out, err := apifunction.MarcaResolve(fqdn, "", ""); err == nil {
		if dados, _ := out["dados"].(map[string]any); dados != nil && len(dados) > 0 {
			auxiliar.RespostaJSON(w, http.StatusOK, out)
			return
		}
	}

	raw, err := xanopro.Post("/fp_whitelabel_by_fqdn", payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func marcaPorCentral(w http.ResponseWriter, r *http.Request) {
	if _, err := seguranca.LerCookies(r); err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	idCen, _ := payload["id_central"].(string)
	app, _ := payload["app"].(string)
	if strings.TrimSpace(app) == "" {
		app = "confvision"
	}
	out, err := apifunction.MarcaResolve("", strings.TrimSpace(idCen), app)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	auxiliar.RespostaJSON(w, http.StatusOK, out)
}

func normalizarHost(host string) string {
	h := strings.ToLower(strings.TrimSpace(host))
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	if i := strings.Index(h, "/"); i >= 0 {
		h = h[:i]
	}
	if i := strings.Index(h, ":"); i >= 0 {
		h = h[:i]
	}
	return h
}

func lerPayload(r *http.Request) (map[string]string, map[string]any, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}

	idFra := seguranca.ResolverIdFranqueado(cookie, payload)
	if idFra == "" {
		return cookie, payload, fmt.Errorf("id_franqueado obrigatorio — faca login novamente")
	}
	payload["id_franqueado"] = idFra
	return cookie, payload, nil
}

func proxyPro(w http.ResponseWriter, r *http.Request, path string) {
	_, payload, err := lerPayload(r)
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "id_franqueado") {
			status = http.StatusBadRequest
		}
		auxiliar.RespostaErro(w, status, err)
		return
	}

	raw, err := xanopro.Post(path, payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

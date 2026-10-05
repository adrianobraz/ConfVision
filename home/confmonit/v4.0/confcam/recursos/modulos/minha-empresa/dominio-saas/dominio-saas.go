package dominioSaas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/provisioner"
	"confvision/src/seguranca"
	"confvision/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-dominio",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/cvDominioCarregar",
		Metodo: http.MethodPost,
		Funcao: dominioCarregar,
		Aberto: false,
	},
	{
		URI:    "/cvDominioSalvar",
		Metodo: http.MethodPost,
		Funcao: dominioSalvar,
		Aberto: false,
	},
	{
		URI:    "/cvDominioRemover",
		Metodo: http.MethodPost,
		Funcao: dominioRemover,
		Aberto: false,
	},
	{
		URI:    "/cvDominioRetentarSSL",
		Metodo: http.MethodPost,
		Funcao: dominioRetentarSSL,
		Aberto: false,
	},
}

type salvarResposta struct {
	Dados       json.RawMessage `json:"dados"`
	Provisionar bool            `json:"provisionar"`
	FQDN        string          `json:"fqdn"`
}

type dominioDados struct {
	Subdominio      string `json:"subdominio"`
	SubdominioCV    string `json:"subdominio_cv"`
	Dominio         string `json:"dominio"`
	FQDN            string `json:"fqdn"`
	FQDNCV          string `json:"fqdn_cv"`
	DominioStatus   string `json:"dominio_status"`
	DominioStatusCV string `json:"dominio_status_cv"`
	DominioErro     string `json:"dominio_erro"`
	DominioErroCV   string `json:"dominio_erro_cv"`
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Domínio personalizado"
	d.LinkRetorno = "/carregar-menu-confvision"
	auxiliar.ExecutarTemplate(w, "dominio-saas.html", d)
}

func dominioCarregar(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_whitelabel_get")
}

func dominioSalvar(w http.ResponseWriter, r *http.Request) {
	_, payload, err := lerPayload(r)
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "id_franqueado") {
			status = http.StatusBadRequest
		}
		auxiliar.RespostaErro(w, status, err)
		return
	}
	payload["app"] = "confvision"
	idFra := seguranca.TextoDoPayload(payload, "id_franqueado")

	raw, err := xanopro.Post("/fp_dominio_salvar", payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}

	var resp salvarResposta
	if err := json.Unmarshal(raw, &resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(raw)
		return
	}

	fqdn := resp.FQDN
	if fqdn == "" {
		var d dominioDados
		if len(resp.Dados) > 0 {
			_ = json.Unmarshal(resp.Dados, &d)
		}
		fqdn = montarFQDNCV(d)
	}

	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)

	if resp.Provisionar && fqdn != "" {
		provOut := executarProvisionamento(idFra, fqdn)
		for k, v := range provOut {
			out[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	enc, _ := json.Marshal(out)
	w.Write(enc)
}

func montarFQDNCV(d dominioDados) string {
	if d.FQDNCV != "" {
		return strings.ToLower(strings.TrimSpace(d.FQDNCV))
	}
	dom := strings.ToLower(strings.TrimSpace(d.Dominio))
	if dom == "" {
		return ""
	}
	sub := strings.ToLower(strings.TrimSpace(d.SubdominioCV))
	if sub == "" || sub == "@" {
		return dom
	}
	return sub + "." + dom
}

func dominioRemover(w http.ResponseWriter, r *http.Request) {
	_, payload, err := lerPayload(r)
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "id_franqueado") {
			status = http.StatusBadRequest
		}
		auxiliar.RespostaErro(w, status, err)
		return
	}
	idFra := seguranca.TextoDoPayload(payload, "id_franqueado")

	getRaw, err := xanopro.Post("/fp_whitelabel_get", map[string]any{
		"id_franqueado": idFra,
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}

	var getResp struct {
		Dados dominioDados `json:"dados"`
	}
	_ = json.Unmarshal(getRaw, &getResp)
	fqdn := montarFQDNCV(getResp.Dados)

	if fqdn == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("nenhum dominio ConfVision configurado para remover"))
		return
	}

	if config.ProvisionerURL != "" {
		if _, err := provisioner.Remover(fqdn); err != nil {
			auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro ao desativar dominio no servidor: %w", err))
			return
		}
	}

	raw, err := xanopro.Post("/fp_whitelabel_dominio_remover", map[string]any{
		"id_franqueado": idFra,
		"app":           "confvision",
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("falhou ao limpar banco: %w", err))
		return
	}

	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	out["fqdn_removido"] = fqdn

	w.Header().Set("Content-Type", "application/json")
	enc, _ := json.Marshal(out)
	w.Write(enc)
}

func dominioRetentarSSL(w http.ResponseWriter, r *http.Request) {
	_, payload, err := lerPayload(r)
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "id_franqueado") {
			status = http.StatusBadRequest
		}
		auxiliar.RespostaErro(w, status, err)
		return
	}
	idFra := seguranca.TextoDoPayload(payload, "id_franqueado")

	getRaw, err := xanopro.Post("/fp_whitelabel_get", map[string]any{
		"id_franqueado": idFra,
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano: %w", err))
		return
	}

	var getResp struct {
		Dados dominioDados `json:"dados"`
	}
	_ = json.Unmarshal(getRaw, &getResp)
	fqdn := montarFQDNCV(getResp.Dados)
	if fqdn == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("nenhum dominio ConfVision configurado"))
		return
	}

	prov, provErr := provisioner.RetentarSSL(fqdn)
	status := "erro"
	erroMsg := ""
	if provErr != nil {
		erroMsg = provErr.Error()
	} else if prov.OK {
		status = prov.Status
		if prov.Status == "" {
			if prov.SSLOK {
				status = "ativo"
			} else {
				status = "pendente_dns"
			}
		}
		if !prov.SSLOK && prov.Message != "" {
			erroMsg = prov.Message
		}
	} else {
		erroMsg = prov.Message
	}

	_, _ = xanopro.Post("/fp_whitelabel_dominio_status", map[string]any{
		"id_franqueado":  idFra,
		"dominio_status": status,
		"dominio_erro":   erroMsg,
		"app":            "confvision",
	})

	out, _ := json.Marshal(map[string]any{
		"dominio_status": status,
		"dominio_erro":   erroMsg,
		"ssl_ok":         prov.SSLOK,
		"message":        prov.Message,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func atualizarStatusDominio(idFranqueado, status, erroMsg string) {
	_, _ = xanopro.Post("/fp_whitelabel_dominio_status", map[string]any{
		"id_franqueado":  idFranqueado,
		"dominio_status": status,
		"dominio_erro":   erroMsg,
		"app":            "confvision",
	})
}

func executarProvisionamento(idFranqueado, fqdn string) map[string]any {
	if config.ProvisionerURL == "" {
		atualizarStatusDominio(idFranqueado, "pendente_dns", "PROVISIONER_URL nao configurado — configure o DNS e SSL manualmente")
		return map[string]any{
			"dominio_status": "pendente_dns",
			"dominio_erro":   "PROVISIONER_URL nao configurado",
			"provision_msg":  "Dominio salvo. Configure DNS e SSL no servidor.",
			"ssl_ok":         false,
			"provisionar":    true,
			"fqdn":           fqdn,
		}
	}

	prov, provErr := provisioner.Provisionar(fqdn)
	status := "erro"
	erroMsg := ""
	if provErr != nil {
		erroMsg = provErr.Error()
	} else if prov.OK {
		status = prov.Status
		if status == "" {
			if prov.SSLOK {
				status = "ativo"
			} else {
				status = "pendente_dns"
			}
		}
		if !prov.SSLOK && prov.Message != "" {
			erroMsg = prov.Message
		}
	} else {
		erroMsg = prov.Message
	}
	atualizarStatusDominio(idFranqueado, status, erroMsg)
	return map[string]any{
		"dominio_status": status,
		"dominio_erro":   erroMsg,
		"provision_msg":  prov.Message,
		"ssl_ok":         prov.SSLOK,
		"provisionar":    true,
		"fqdn":           fqdn,
	}
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

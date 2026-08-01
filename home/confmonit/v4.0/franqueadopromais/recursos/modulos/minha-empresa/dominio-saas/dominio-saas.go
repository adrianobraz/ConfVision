package dominioSaas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/provisioner"
	"franqueadopro/src/seguranca"
	"franqueadopro/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-dominio-saas",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/dominioSaasCarregar",
		Metodo: http.MethodPost,
		Funcao: dominioSaasCarregar,
		Aberto: false,
	},
	{
		URI:    "/dominioSaasSalvar",
		Metodo: http.MethodPost,
		Funcao: dominioSaasSalvar,
		Aberto: false,
	},
	{
		URI:    "/dominioSaasRemover",
		Metodo: http.MethodPost,
		Funcao: dominioSaasRemover,
		Aberto: false,
	},
	{
		URI:    "/dominioSaasRetentarSSL",
		Metodo: http.MethodPost,
		Funcao: dominioSaasRetentarSSL,
		Aberto: false,
	},
}

type salvarResposta struct {
	Dados       json.RawMessage `json:"dados"`
	Provisionar bool            `json:"provisionar"`
	FQDN        string          `json:"fqdn"`
}

type dominioDados struct {
	Subdominio    string `json:"subdominio"`
	Dominio       string `json:"dominio"`
	FQDN          string `json:"fqdn"`
	DominioStatus string `json:"dominio_status"`
	DominioErro   string `json:"dominio_erro"`
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Domínio Personalizado"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-gestao"
	auxiliar.PreencherEhMaster(r, &d)
	auxiliar.ExecutarTemplate(w, "dominio-saas.html", d)
}

func dominioSaasCarregar(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_whitelabel_get")
}

func dominioSaasSalvar(w http.ResponseWriter, r *http.Request) {
	cookie, payload, err := lerPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	raw, err := xanopro.Post("/fp_dominio_salvar", payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano pro: %w", err))
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
		fqdn = d.FQDN
	}

	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)

	if resp.Provisionar && fqdn != "" {
		provOut := executarProvisionamento(cookie["idFranqueado"], fqdn)
		for k, v := range provOut {
			out[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	enc, _ := json.Marshal(out)
	w.Write(enc)
}

func montarFQDN(d dominioDados) string {
	if d.FQDN != "" {
		return strings.ToLower(strings.TrimSpace(d.FQDN))
	}
	dom := strings.ToLower(strings.TrimSpace(d.Dominio))
	if dom == "" {
		return ""
	}
	sub := strings.ToLower(strings.TrimSpace(d.Subdominio))
	if sub == "" || sub == "@" {
		return dom
	}
	return sub + "." + dom
}

func dominioSaasRemover(w http.ResponseWriter, r *http.Request) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	getRaw, err := xanopro.Post("/fp_whitelabel_get", map[string]any{
		"id_franqueado": cookie["idFranqueado"],
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano pro: %w", err))
		return
	}

	var getResp struct {
		Dados dominioDados `json:"dados"`
	}
	_ = json.Unmarshal(getRaw, &getResp)
	fqdn := montarFQDN(getResp.Dados)

	if fqdn == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("nenhum dominio configurado para remover"))
		return
	}

	prov, err := provisioner.Remover(fqdn)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro ao desativar dominio no servidor: %w", err))
		return
	}

	raw, err := xanopro.Post("/fp_whitelabel_dominio_remover", map[string]any{
		"id_franqueado": cookie["idFranqueado"],
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("dominio desativado no servidor, mas falhou ao limpar banco: %w", err))
		return
	}

	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	out["provision_msg"] = prov.Message
	out["fqdn_removido"] = fqdn

	w.Header().Set("Content-Type", "application/json")
	enc, _ := json.Marshal(out)
	w.Write(enc)
}

func dominioSaasRetentarSSL(w http.ResponseWriter, r *http.Request) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	getRaw, err := xanopro.Post("/fp_whitelabel_get", map[string]any{
		"id_franqueado": cookie["idFranqueado"],
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano pro: %w", err))
		return
	}

	var getResp struct {
		Dados struct {
			FQDN string `json:"fqdn"`
		} `json:"dados"`
	}
	_ = json.Unmarshal(getRaw, &getResp)
	if getResp.Dados.FQDN == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("nenhum dominio configurado"))
		return
	}

	prov, provErr := provisioner.RetentarSSL(getResp.Dados.FQDN)
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
		"id_franqueado":  cookie["idFranqueado"],
		"dominio_status": status,
		"dominio_erro":   erroMsg,
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
	})
}

func executarProvisionamento(idFranqueado, fqdn string) map[string]any {
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
	payload["id_franqueado"] = cookie["idFranqueado"]
	return cookie, payload, nil
}

func proxyPro(w http.ResponseWriter, r *http.Request, path string) {
	_, payload, err := lerPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	raw, err := xanopro.Post(path, payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano pro: %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

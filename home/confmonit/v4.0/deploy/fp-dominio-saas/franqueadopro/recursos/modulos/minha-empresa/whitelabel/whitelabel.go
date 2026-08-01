package whitelabel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/provisioner"
	"franqueadopro/src/seguranca"
	"franqueadopro/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-whitelabel",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/whitelabelCarregar",
		Metodo: http.MethodPost,
		Funcao: whitelabelCarregar,
		Aberto: false,
	},
	{
		URI:    "/whitelabelSalvar",
		Metodo: http.MethodPost,
		Funcao: whitelabelSalvar,
		Aberto: false,
	},
	{
		URI:    "/whitelabelDominioRemover",
		Metodo: http.MethodPost,
		Funcao: whitelabelDominioRemover,
		Aberto: false,
	},
	{
		URI:    "/whitelabelDominioRetentarSSL",
		Metodo: http.MethodPost,
		Funcao: whitelabelDominioRetentarSSL,
		Aberto: false,
	},
}

type salvarResposta struct {
	Dados       json.RawMessage `json:"dados"`
	Provisionar bool            `json:"provisionar"`
	FQDN        string          `json:"fqdn"`
}

type removerResposta struct {
	Dados        json.RawMessage `json:"dados"`
	FQDNRemovido string          `json:"fqdn_removido"`
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "White Label"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-gestao"
	auxiliar.PreencherEhMaster(r, &d)
	auxiliar.ExecutarTemplate(w, "whitelabel.html", d)
}

func whitelabelCarregar(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_whitelabel_get")
}

func whitelabelSalvar(w http.ResponseWriter, r *http.Request) {
	cookie, payload, err := lerPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	raw, err := xanopro.Post("/fp_whitelabel_salvar", payload)
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

	if resp.Provisionar && resp.FQDN != "" {
		prov, provErr := provisioner.Provisionar(resp.FQDN)
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
			"dados":         resp.Dados,
			"provisionar":   true,
			"fqdn":          resp.FQDN,
			"dominio_status": status,
			"dominio_erro":  erroMsg,
			"provision_msg": prov.Message,
			"ssl_ok":        prov.SSLOK,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func whitelabelDominioRemover(w http.ResponseWriter, r *http.Request) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	raw, err := xanopro.Post("/fp_whitelabel_dominio_remover", map[string]any{
		"id_franqueado": cookie["idFranqueado"],
	})
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("erro xano pro: %w", err))
		return
	}

	var resp removerResposta
	if err := json.Unmarshal(raw, &resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(raw)
		return
	}

	if resp.FQDNRemovido != "" {
		_, _ = provisioner.Remover(resp.FQDNRemovido)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func whitelabelDominioRetentarSSL(w http.ResponseWriter, r *http.Request) {
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

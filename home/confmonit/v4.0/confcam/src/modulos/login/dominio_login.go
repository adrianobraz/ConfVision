package login

import (
	"encoding/json"
	"strings"

	"confvision/src/config"
	"confvision/src/xanopro"
)

const errAcessoNegado = "acesso negado"

func hostLoginPermiteBypass(host string) bool {
	h := normalizarHost(host)
	return h == "" || h == "localhost" || h == "127.0.0.1"
}

// clientePermitidoNoHost: CLI so bloqueado se o franqueado tem fqdn_cv e o Host e diferente.
func clientePermitidoNoHost(hostLogin, idFranqueado string) (bool, error) {
	if hostLoginPermiteBypass(hostLogin) {
		return true, nil
	}
	fqdn, err := fqdnCVFranqueado(idFranqueado)
	if err != nil {
		return true, nil
	}
	if fqdn == "" {
		return true, nil
	}
	return normalizarHost(hostLogin) == normalizarHost(fqdn), nil
}

func fqdnCVFranqueado(idFranqueado string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return "", nil
	}
	if config.XanoApiPro == "" {
		return "", nil
	}

	raw, err := xanopro.Post("/fp_whitelabel_get", map[string]any{
		"id_franqueado": idFranqueado,
	})
	if err != nil {
		return "", err
	}

	var resp struct {
		Dados map[string]any `json:"dados"`
	}
	if json.Unmarshal(raw, &resp) != nil || resp.Dados == nil {
		return "", nil
	}

	fqdn := strings.TrimSpace(anyToTrimString(resp.Dados["fqdn_cv"]))
	if fqdn == "" {
		return "", nil
	}

	status := strings.ToLower(strings.TrimSpace(anyToTrimString(resp.Dados["dominio_status_cv"])))
	if status == "removido" || status == "erro" {
		return "", nil
	}

	return fqdn, nil
}

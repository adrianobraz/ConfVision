package pgcentraldominio

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"apifunction/config"
)

var fqdnRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)
var subRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func MontarFQDN(sub, dom string) string {
	dom = strings.ToLower(strings.TrimSpace(dom))
	if dom == "" {
		return ""
	}
	sub = strings.ToLower(strings.TrimSpace(sub))
	if sub == "" || sub == "@" {
		return dom
	}
	return sub + "." + dom
}

func normalizarDominio(d string) string {
	d = strings.ToLower(strings.TrimSpace(d))
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	if i := strings.Index(d, "/"); i >= 0 {
		d = d[:i]
	}
	return strings.TrimSuffix(d, ".")
}

func DNSIP() string {
	if ip := strings.TrimSpace(config.DominioMarcaDNSIP); ip != "" {
		return ip
	}
	return "185.130.61.4"
}

func Carregar(ctx context.Context, idCentral string) (*Registro, error) {
	reg, err := Obter(ctx, idCentral)
	if err != nil {
		return nil, err
	}
	reg.DNSIP = DNSIP()
	return reg, nil
}

type SalvarInput struct {
	IDCentral  string
	Dominio    string
	Subdominio string
	App        string
}

func Salvar(ctx context.Context, in SalvarInput) (*Registro, ProvResult, error) {
	app := strings.ToLower(strings.TrimSpace(in.App))
	if !AppValido(app) {
		return nil, ProvResult{}, errors.New("app invalido")
	}
	dom := normalizarDominio(in.Dominio)
	if dom == "" {
		return nil, ProvResult{}, errors.New("dominio obrigatorio")
	}
	if !fqdnRe.MatchString(dom) {
		return nil, ProvResult{}, errors.New("dominio invalido")
	}
	sub := strings.ToLower(strings.TrimSpace(in.Subdominio))
	if sub == "" {
		return nil, ProvResult{}, errors.New("subdominio obrigatorio")
	}
	if !subRe.MatchString(sub) {
		return nil, ProvResult{}, errors.New("subdominio invalido")
	}

	reg, err := Obter(ctx, in.IDCentral)
	if err != nil {
		return nil, ProvResult{}, err
	}
	if reg.Dominio != "" && reg.Dominio != dom {
		return nil, ProvResult{}, fmt.Errorf("dominio base ja configurado (%s). Remova os produtos antes de alterar a base", reg.Dominio)
	}
	reg.Dominio = dom

	fqdn := MontarFQDN(sub, dom)
	if !fqdnRe.MatchString(fqdn) {
		return nil, ProvResult{}, errors.New("fqdn invalido")
	}

	// Unicidade entre apps da mesma central
	for otherApp, slot := range reg.Apps {
		if otherApp == app {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(slot.FQDN), fqdn) {
			return nil, ProvResult{}, fmt.Errorf("este fqdn ja esta em uso pelo produto %s nesta central", otherApp)
		}
	}

	ocupado, err := ListarTodosFQDN(ctx, in.IDCentral)
	if err != nil {
		return nil, ProvResult{}, err
	}
	if _, ok := ocupado[fqdn]; ok {
		return nil, ProvResult{}, errors.New("este dominio ja esta em uso por outra central")
	}

	old := reg.Apps[app]
	if old.FQDN != "" && !strings.EqualFold(old.FQDN, fqdn) {
		_, _ = RemoverApp(old.FQDN, app)
	}

	reg.Apps[app] = AppSlot{
		Subdominio: sub,
		FQDN:       fqdn,
		Status:     "provisionando",
		Erro:       "",
	}
	if err := SalvarRow(ctx, reg); err != nil {
		return nil, ProvResult{}, err
	}

	prov, provErr := ProvisionarApp(fqdn, app)
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
	reg.Apps[app] = AppSlot{
		Subdominio: sub,
		FQDN:       fqdn,
		Status:     status,
		Erro:       erroMsg,
	}
	_ = SalvarRow(ctx, reg)
	reg.DNSIP = DNSIP()
	return reg, prov, nil
}

func Remover(ctx context.Context, idCentral, app string) (*Registro, error) {
	app = strings.ToLower(strings.TrimSpace(app))
	if !AppValido(app) {
		return nil, errors.New("app invalido")
	}
	reg, err := Obter(ctx, idCentral)
	if err != nil {
		return nil, err
	}
	slot := reg.Apps[app]
	if slot.FQDN != "" {
		_, _ = RemoverApp(slot.FQDN, app)
	}
	reg.Apps[app] = AppSlot{}
	if err := SalvarRow(ctx, reg); err != nil {
		return nil, err
	}
	reg.DNSIP = DNSIP()
	return reg, nil
}

func RetentarSSL(ctx context.Context, idCentral, app string) (*Registro, ProvResult, error) {
	app = strings.ToLower(strings.TrimSpace(app))
	if !AppValido(app) {
		return nil, ProvResult{}, errors.New("app invalido")
	}
	reg, err := Obter(ctx, idCentral)
	if err != nil {
		return nil, ProvResult{}, err
	}
	slot := reg.Apps[app]
	fqdn := strings.TrimSpace(slot.FQDN)
	if fqdn == "" {
		return nil, ProvResult{}, errors.New("nenhum dominio configurado para este produto")
	}
	prov, provErr := RetentarSSLApp(fqdn, app)
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
	reg.Apps[app] = AppSlot{
		Subdominio: slot.Subdominio,
		FQDN:       fqdn,
		Status:     status,
		Erro:       erroMsg,
	}
	_ = SalvarRow(ctx, reg)
	reg.DNSIP = DNSIP()
	return reg, prov, nil
}

// URLsHub devolve mapa produto_hub -> url https (somente status ativo).
func URLsHub(ctx context.Context, idCentral string) (map[string]string, error) {
	reg, err := Carregar(ctx, idCentral)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for app, slot := range reg.Apps {
		if strings.ToLower(strings.TrimSpace(slot.Status)) != "ativo" {
			continue
		}
		fqdn := strings.TrimSpace(slot.FQDN)
		if fqdn == "" {
			continue
		}
		hub := HubProduto(app)
		if hub == "" {
			continue
		}
		out[hub] = "https://" + fqdn
	}
	return out, nil
}

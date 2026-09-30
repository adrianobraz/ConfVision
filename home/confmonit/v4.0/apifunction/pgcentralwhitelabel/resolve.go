package pgcentralwhitelabel

import (
	"context"
	"strings"

	"apifunction/pgcentraldominio"
)

// ResolveMarca por FQDN (domínio da central) ou id_central + app explícito.
type ResolveMarca struct {
	IDCentral string         `json:"id_central"`
	App       string         `json:"app"`
	FQDN      string         `json:"fqdn,omitempty"`
	Dados     map[string]any `json:"dados"`
}

func ResolvePorFQDN(ctx context.Context, fqdn string) (*ResolveMarca, bool, error) {
	fqdn = strings.ToLower(strings.TrimSpace(fqdn))
	if fqdn == "" {
		return nil, false, nil
	}
	idCen, app, ok, err := pgcentraldominio.ResolverCentralPorFQDN(ctx, fqdn)
	if err != nil || !ok {
		return nil, false, err
	}
	return montarResolve(ctx, idCen, app, fqdn)
}

func ResolvePorCentralApp(ctx context.Context, idCentral, app string) (*ResolveMarca, bool, error) {
	idCentral = strings.TrimSpace(idCentral)
	app = strings.ToLower(strings.TrimSpace(app))
	if idCentral == "" || app == "" {
		return nil, false, nil
	}
	return montarResolve(ctx, idCentral, app, "")
}

func montarResolve(ctx context.Context, idCentral, app, fqdn string) (*ResolveMarca, bool, error) {
	reg, err := Obter(ctx, idCentral)
	if err != nil {
		return nil, false, err
	}
	logo := strings.TrimSpace(reg.LogosJSON[app])
	tema := reg.TemaJSON
	if logo == "" && len(tema) == 0 {
		return nil, false, nil
	}
	dados := map[string]any{
		"id_central": idCentral,
		"app":        app,
	}
	if logo != "" {
		dados["logo_data"] = logo
	}
	if len(tema) > 0 {
		dados["tema_json"] = tema
	}
	return &ResolveMarca{
		IDCentral: idCentral,
		App:       app,
		FQDN:      fqdn,
		Dados:     dados,
	}, true, nil
}

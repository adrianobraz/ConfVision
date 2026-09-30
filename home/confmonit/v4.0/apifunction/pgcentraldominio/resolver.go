package pgcentraldominio

import (
	"context"
	"encoding/json"
	"strings"

	"apifunction/pgcredito"
)

// ResolverCentralPorFQDN encontra central e app pelo FQDN configurado em Domínios da marca.
func ResolverCentralPorFQDN(ctx context.Context, fqdn string) (idCentral, app string, ok bool, err error) {
	fqdn = strings.ToLower(strings.TrimSpace(fqdn))
	if fqdn == "" {
		return "", "", false, nil
	}
	db, err := pgcredito.DB()
	if err != nil {
		return "", "", false, err
	}
	rows, err := db.QueryContext(ctx, `SELECT id_central, apps FROM fp_central_dominio_marca`)
	if err != nil {
		return "", "", false, err
	}
	defer rows.Close()
	for rows.Next() {
		var idCen string
		var raw []byte
		if err := rows.Scan(&idCen, &raw); err != nil {
			return "", "", false, err
		}
		apps := map[string]AppSlot{}
		_ = json.Unmarshal(raw, &apps)
		for appName, slot := range apps {
			if strings.EqualFold(strings.TrimSpace(slot.FQDN), fqdn) {
				return strings.TrimSpace(idCen), strings.ToLower(strings.TrimSpace(appName)), true, nil
			}
		}
	}
	return "", "", false, nil
}

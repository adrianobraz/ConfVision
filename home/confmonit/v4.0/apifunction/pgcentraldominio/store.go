package pgcentraldominio

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"apifunction/pgcredito"
)

type AppSlot struct {
	Subdominio string `json:"subdominio"`
	FQDN       string `json:"fqdn"`
	Status     string `json:"status"`
	Erro       string `json:"erro"`
}

type Registro struct {
	IDCentral string             `json:"id_central"`
	Dominio   string             `json:"dominio"`
	Apps      map[string]AppSlot `json:"apps"`
	DNSIP     string             `json:"dns_ip"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func emptyApps() map[string]AppSlot {
	m := make(map[string]AppSlot, len(AppsSuportados))
	for _, a := range AppsSuportados {
		m[a] = AppSlot{}
	}
	return m
}

func Obter(ctx context.Context, idCentral string) (*Registro, error) {
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		return nil, errors.New("id_central obrigatorio")
	}
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	var dominio string
	var raw []byte
	var updated time.Time
	err = db.QueryRowContext(ctx, `
		SELECT dominio, apps, updated_at
		FROM fp_central_dominio_marca
		WHERE id_central = $1
	`, idCentral).Scan(&dominio, &raw, &updated)
	if err == sql.ErrNoRows {
		return &Registro{
			IDCentral: idCentral,
			Dominio:   "",
			Apps:      emptyApps(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	apps := emptyApps()
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &apps)
	}
	for _, a := range AppsSuportados {
		if _, ok := apps[a]; !ok {
			apps[a] = AppSlot{}
		}
	}
	return &Registro{
		IDCentral: idCentral,
		Dominio:   strings.ToLower(strings.TrimSpace(dominio)),
		Apps:      apps,
		UpdatedAt: updated,
	}, nil
}

func SalvarRow(ctx context.Context, reg *Registro) error {
	if reg == nil || strings.TrimSpace(reg.IDCentral) == "" {
		return errors.New("id_central obrigatorio")
	}
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(reg.Apps)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO fp_central_dominio_marca (id_central, dominio, apps, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (id_central) DO UPDATE SET
			dominio = EXCLUDED.dominio,
			apps = EXCLUDED.apps,
			updated_at = NOW()
	`, reg.IDCentral, strings.ToLower(strings.TrimSpace(reg.Dominio)), raw)
	return err
}

func ListarTodosFQDN(ctx context.Context, excetoCentral string) (map[string]string, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `SELECT id_central, apps FROM fp_central_dominio_marca`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var idCen string
		var raw []byte
		if err := rows.Scan(&idCen, &raw); err != nil {
			return nil, err
		}
		if strings.EqualFold(strings.TrimSpace(idCen), strings.TrimSpace(excetoCentral)) {
			continue
		}
		apps := map[string]AppSlot{}
		_ = json.Unmarshal(raw, &apps)
		for _, slot := range apps {
			fq := strings.ToLower(strings.TrimSpace(slot.FQDN))
			if fq != "" {
				out[fq] = idCen
			}
		}
	}
	return out, nil
}

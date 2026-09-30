package pgcentralwhitelabel

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"apifunction/pgcredito"
)

type Registro struct {
	IDCentral string            `json:"id_central"`
	TemaJSON  map[string]any    `json:"tema_json"`
	LogosJSON map[string]string `json:"logos_json"`
	UpdatedAt time.Time         `json:"updated_at"`
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
	var temaRaw, logosRaw []byte
	var updated time.Time
	err = db.QueryRowContext(ctx, `
		SELECT tema_json, logos_json, updated_at
		FROM fp_central_whitelabel WHERE id_central = $1
	`, idCentral).Scan(&temaRaw, &logosRaw, &updated)
	if err == sql.ErrNoRows {
		return &Registro{
			IDCentral: idCentral,
			TemaJSON:  map[string]any{},
			LogosJSON: map[string]string{},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	reg := &Registro{IDCentral: idCentral, UpdatedAt: updated}
	reg.TemaJSON = map[string]any{}
	reg.LogosJSON = map[string]string{}
	if len(temaRaw) > 0 {
		_ = json.Unmarshal(temaRaw, &reg.TemaJSON)
	}
	if len(logosRaw) > 0 {
		_ = json.Unmarshal(logosRaw, &reg.LogosJSON)
	}
	return reg, nil
}

func Salvar(ctx context.Context, reg *Registro) error {
	if reg == nil || strings.TrimSpace(reg.IDCentral) == "" {
		return errors.New("id_central obrigatorio")
	}
	if reg.TemaJSON == nil {
		reg.TemaJSON = map[string]any{}
	}
	if reg.LogosJSON == nil {
		reg.LogosJSON = map[string]string{}
	}
	temaRaw, err := json.Marshal(reg.TemaJSON)
	if err != nil {
		return err
	}
	logosRaw, err := json.Marshal(reg.LogosJSON)
	if err != nil {
		return err
	}
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO fp_central_whitelabel (id_central, tema_json, logos_json, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (id_central) DO UPDATE SET
			tema_json = EXCLUDED.tema_json,
			logos_json = EXCLUDED.logos_json,
			updated_at = NOW()
	`, reg.IDCentral, temaRaw, logosRaw)
	return err
}

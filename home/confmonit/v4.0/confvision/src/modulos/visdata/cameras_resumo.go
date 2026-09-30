package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type CamerasResumoSQL struct {
	Total          int
	Ativas         int
	Inativas       int
	SemComunicacao int
	Pausadas       int
	Bloqueadas     int
}

type CameraResumoRow struct {
	ID            int
	Ativo         bool
	Bloqueado     bool
	Plano         string
	Pausado       bool
	SomenteArmado bool
	IDDispositivo string
}

func CountCamerasResumoSQL(ctx context.Context, idFranqueado string, staleMin int) (CamerasResumoSQL, error) {
	out := CamerasResumoSQL{}
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return out, fmt.Errorf("id_franqueado obrigatorio")
	}
	if staleMin < 1 {
		staleMin = 15
	}

	db, err := DB()
	if err != nil {
		return out, err
	}

	limite := time.Now().UTC().Add(-time.Duration(staleMin) * time.Minute)

	err = db.QueryRowContext(ctx, `
SELECT
  COUNT(*)::int AS total,
  COUNT(*) FILTER (WHERE c.ativo IS TRUE)::int AS ativas,
  COUNT(*) FILTER (WHERE c.ativo IS NOT TRUE)::int AS inativas,
  COUNT(*) FILTER (
    WHERE c.ativo IS TRUE
      AND (c.bloqueado IS NOT TRUE)
      AND (c.ultimo_ping_em IS NULL OR c.ultimo_ping_em < $2)
  )::int AS sem_comunicacao,
  COUNT(*) FILTER (
    WHERE c.ativo IS TRUE AND c.analitico_pausado IS TRUE
  )::int AS pausadas,
  COUNT(*) FILTER (WHERE c.bloqueado IS TRUE)::int AS bloqueadas
FROM vis_camera c
WHERE c.id_franqueado = $1`, idFranqueado, limite).Scan(
		&out.Total, &out.Ativas, &out.Inativas, &out.SemComunicacao, &out.Pausadas, &out.Bloqueadas,
	)
	return out, err
}

func ListCamerasResumoRows(ctx context.Context, idFranqueado string) ([]CameraResumoRow, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT c.id, c.ativo, c.bloqueado, COALESCE(c.plano, ''), c.analitico_pausado, c.somente_armado, COALESCE(c.id_dispositivo, '')
FROM vis_camera c
WHERE c.id_franqueado = $1`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CameraResumoRow
	for rows.Next() {
		var row CameraResumoRow
		var ativo, bloq, pausado, somente sql.NullBool
		var plano, idDisp sql.NullString
		if err := rows.Scan(&row.ID, &ativo, &bloq, &plano, &pausado, &somente, &idDisp); err != nil {
			return nil, err
		}
		row.Ativo = ativo.Valid && ativo.Bool
		row.Bloqueado = bloq.Valid && bloq.Bool
		row.Pausado = pausado.Valid && pausado.Bool
		row.SomenteArmado = somente.Valid && somente.Bool
		if plano.Valid {
			row.Plano = strings.TrimSpace(plano.String)
		}
		if idDisp.Valid {
			row.IDDispositivo = strings.TrimSpace(idDisp.String)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func GetCameraIDFranqueado(ctx context.Context, cameraID int) (string, error) {
	db, err := DB()
	if err != nil {
		return "", err
	}
	var idFra sql.NullString
	err = db.QueryRowContext(ctx, `SELECT id_franqueado FROM vis_camera WHERE id = $1`, cameraID).Scan(&idFra)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(idFra.String)
	if s == "" {
		return "", fmt.Errorf("camera sem id_franqueado")
	}
	return s, nil
}

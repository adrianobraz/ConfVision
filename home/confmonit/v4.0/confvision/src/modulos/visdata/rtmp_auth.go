package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const rtmpAuthSelectSQL = `
SELECT c.id, c.id_franqueado, c.ativo, c.bloqueado, c.protocolo, c.plano,
       c.vis_licenca_id, l.status, l.valido_ate,
       c.analitico_pausado, c.stream_motivo_pausa
FROM vis_camera c
LEFT JOIN vis_licenca l ON l.id = c.vis_licenca_id
`

type rtmpAuthRow struct {
	id                              int
	idFranqueado, protocolo, plano  sql.NullString
	ativo, bloqueado                sql.NullBool
	licID                           sql.NullInt64
	licStatus                       sql.NullString
	licValido                       sql.NullTime
	analiticoPausado                sql.NullBool
	streamMotivoPausa               sql.NullString
}

func scanRTMPAuth(scanner interface {
	Scan(dest ...any) error
}) (rtmpAuthRow, error) {
	var row rtmpAuthRow
	err := scanner.Scan(
		&row.id, &row.idFranqueado, &row.ativo, &row.bloqueado, &row.protocolo, &row.plano,
		&row.licID, &row.licStatus, &row.licValido,
		&row.analiticoPausado, &row.streamMotivoPausa,
	)
	return row, err
}

func mapRTMPAuth(row rtmpAuthRow) map[string]any {
	out := map[string]any{
		"id":            row.id,
		"id_franqueado": nullStr(row.idFranqueado),
		"ativo":         nullBool(row.ativo),
		"bloqueado":     nullBool(row.bloqueado),
		"protocolo":     nullStr(row.protocolo),
		"plano":         nullStr(row.plano),
	}
	if row.licID.Valid {
		out["vis_licenca_id"] = row.licID.Int64
	} else {
		out["vis_licenca_id"] = nil
	}
	if row.licStatus.Valid {
		out["licenca_status"] = row.licStatus.String
	}
	if row.licValido.Valid {
		out["licenca_valido_ate"] = row.licValido.Time.UTC().Format(time.RFC3339)
	}
	out["analitico_pausado"] = nullBool(row.analiticoPausado)
	if row.streamMotivoPausa.Valid && strings.TrimSpace(row.streamMotivoPausa.String) != "" {
		out["stream_motivo_pausa"] = strings.TrimSpace(row.streamMotivoPausa.String)
	} else {
		out["stream_motivo_pausa"] = nil
	}
	return out
}

func GetCameraRTMPAuth(ctx context.Context, cameraID int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	row, err := scanRTMPAuth(db.QueryRowContext(ctx, rtmpAuthSelectSQL+` WHERE c.id = $1`, cameraID))
	if err != nil {
		return nil, err
	}
	return mapRTMPAuth(row), nil
}

// ListCamerasRTMPAuthSync lista cameras elegiveis para publish RTMP no no (cache Redis).
func ListCamerasRTMPAuthSync(ctx context.Context, nodeID int) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	query := rtmpAuthSelectSQL + `
WHERE (c.bloqueado IS NOT TRUE)
  AND (c.ativo = TRUE OR LOWER(COALESCE(c.plano, '')) = 'online')`
	args := []any{}
	if nodeID > 0 {
		query += ` AND c.vis_mediamtx_node_id = $1`
		args = append(args, nodeID)
	}
	query += ` ORDER BY c.id ASC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		row, err := scanRTMPAuth(rows)
		if err != nil {
			return nil, fmt.Errorf("scan rtmp auth sync: %w", err)
		}
		out = append(out, mapRTMPAuth(row))
	}
	return out, rows.Err()
}

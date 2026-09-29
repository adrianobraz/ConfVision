package visdata

import (
	"context"
	"database/sql"
	"fmt"
)

type MediamtxNode struct {
	ID               int
	Nome             string
	RtmpPublic       sql.NullString
	HlsPublic        sql.NullString
	RtspInternal     sql.NullString
	MaxCameras       int
	Ordem            int
	Status           string
	CamerasAtribuidas int
}

func PickMediamtxNode(ctx context.Context) (*MediamtxNode, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
SELECT id, nome, rtmp_public, hls_public, rtsp_internal,
       COALESCE(max_cameras, 200), COALESCE(ordem, 1), COALESCE(status, 'ativo')
FROM vis_mediamtx_node
WHERE status = 'ativo'
ORDER BY ordem ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *MediamtxNode
	bestCount := 999999

	for rows.Next() {
		var n MediamtxNode
		if err := rows.Scan(&n.ID, &n.Nome, &n.RtmpPublic, &n.HlsPublic, &n.RtspInternal,
			&n.MaxCameras, &n.Ordem, &n.Status); err != nil {
			return nil, err
		}
		var total int
		if err := db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_camera WHERE vis_mediamtx_node_id = $1`, n.ID).Scan(&total); err != nil {
			return nil, err
		}
		n.CamerasAtribuidas = total
		if total < n.MaxCameras && total < bestCount {
			c := n
			best = &c
			bestCount = total
		}
	}
	if best == nil {
		return nil, fmt.Errorf("Todos os nos MediaMTX estao cheios — cadastre um novo servidor ou aumente max_cameras")
	}
	return best, nil
}

func SyncMediamtxNode(ctx context.Context, nodeID int) error {
	if nodeID <= 0 {
		return nil
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `SELECT vis_mediamtx_node_recalc_pontos($1)`, nodeID)
	return err
}

func ListMediamtxNodes(ctx context.Context) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT n.id, n.created_at, n.nome, n.rtmp_public, n.hls_public, n.rtsp_internal,
       COALESCE(n.max_cameras, 200), COALESCE(n.ordem, 1), COALESCE(n.status, 'ativo'), n.observacao,
       (SELECT COUNT(*) FROM vis_camera c WHERE c.vis_mediamtx_node_id = n.id) AS cameras_atribuidas
FROM vis_mediamtx_node n
ORDER BY n.ordem ASC, n.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var id, maxCam, ordem, atrib int
		var created sql.NullTime
		var nome, rtmp, hls, rtsp, status, obs sql.NullString
		if err := rows.Scan(&id, &created, &nome, &rtmp, &hls, &rtsp, &maxCam, &ordem, &status, &obs, &atrib); err != nil {
			return nil, err
		}
		vagas := maxCam - atrib
		out = append(out, map[string]any{
			"id":                 id,
			"created_at":         nullTime(created),
			"nome":               nullStr(nome),
			"rtmp_public":        nullStr(rtmp),
			"hls_public":         nullStr(hls),
			"rtsp_internal":      nullStr(rtsp),
			"max_cameras":        maxCam,
			"ordem":              ordem,
			"status":             nullStr(status),
			"observacao":         nullStr(obs),
			"cameras_atribuidas": atrib,
			"vagas_restantes":    vagas,
		})
	}
	return out, rows.Err()
}

func loadMediamtxNodeByID(ctx context.Context, nodeID int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id, maxCam, ordem, atrib int
	var created sql.NullTime
	var nome, rtmp, hls, rtsp, status, obs sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT n.id, n.created_at, n.nome, n.rtmp_public, n.hls_public, n.rtsp_internal,
       COALESCE(n.max_cameras, 200), COALESCE(n.ordem, 1), COALESCE(n.status, 'ativo'), n.observacao,
       (SELECT COUNT(*) FROM vis_camera c WHERE c.vis_mediamtx_node_id = n.id) AS cameras_atribuidas
FROM vis_mediamtx_node n
WHERE n.id = $1`, nodeID).Scan(&id, &created, &nome, &rtmp, &hls, &rtsp, &maxCam, &ordem, &status, &obs, &atrib)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no MediaMTX nao encontrado")
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                    id,
		"vis_mediamtx_node_id":  id,
		"created_at":            nullTime(created),
		"nome":                  nullStr(nome),
		"rtmp_public":           nullStr(rtmp),
		"hls_public":            nullStr(hls),
		"rtsp_internal":         nullStr(rtsp),
		"max_cameras":           maxCam,
		"ordem":                 ordem,
		"status":                nullStr(status),
		"observacao":            nullStr(obs),
		"cameras_atribuidas":    atrib,
		"vagas_restantes":       maxCam - atrib,
	}, nil
}

// ResolveMediamtxForCamera retorna URLs do no MediaMTX da camera; opcionalmente atribui no se ausente.
func ResolveMediamtxForCamera(ctx context.Context, cameraID int, atribuirSeAusente bool) (map[string]any, error) {
	if cameraID <= 0 {
		return nil, fmt.Errorf("camera_id invalido")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var nodeID sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT vis_mediamtx_node_id FROM vis_camera WHERE id = $1`, cameraID).Scan(&nodeID); err == sql.ErrNoRows {
		return nil, fmt.Errorf("camera nao encontrada")
	} else if err != nil {
		return nil, err
	}

	if !nodeID.Valid || nodeID.Int64 <= 0 {
		if !atribuirSeAusente {
			return nil, fmt.Errorf("camera sem no MediaMTX atribuido")
		}
		node, err := PickMediamtxNode(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := db.ExecContext(ctx, `UPDATE vis_camera SET vis_mediamtx_node_id = $2 WHERE id = $1`, cameraID, node.ID); err != nil {
			return nil, err
		}
		_ = SyncMediamtxNode(ctx, node.ID)
		return loadMediamtxNodeByID(ctx, node.ID)
	}

	out, err := loadMediamtxNodeByID(ctx, int(nodeID.Int64))
	if err != nil {
		return nil, err
	}
	out["vis_camera_id"] = cameraID
	return out, nil
}

package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
SELECT n.id, n.nome, n.rtmp_public, n.hls_public, n.rtsp_internal,
       COALESCE(n.max_cameras, 200), COALESCE(n.ordem, 1), COALESCE(n.status, 'ativo'),
       (SELECT COUNT(*) FROM vis_camera c WHERE c.vis_mediamtx_node_id = n.id) AS cameras_atribuidas
FROM vis_mediamtx_node n
WHERE n.status = 'ativo'
ORDER BY n.ordem ASC, n.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *MediamtxNode
	bestCount := 999999

	for rows.Next() {
		var n MediamtxNode
		var total int
		if err := rows.Scan(&n.ID, &n.Nome, &n.RtmpPublic, &n.HlsPublic, &n.RtspInternal,
			&n.MaxCameras, &n.Ordem, &n.Status, &total); err != nil {
			return nil, err
		}
		n.CamerasAtribuidas = total
		if total < n.MaxCameras && total < bestCount {
			c := n
			best = &c
			bestCount = total
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
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

func getMediamtxNodeByID(ctx context.Context, nodeID int) (*MediamtxNode, int, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	var n MediamtxNode
	var total int
	err = db.QueryRowContext(ctx, `
SELECT n.id, n.nome, n.rtmp_public, n.hls_public, n.rtsp_internal,
       COALESCE(n.max_cameras, 200), COALESCE(n.ordem, 1), COALESCE(n.status, 'ativo'),
       (SELECT COUNT(*) FROM vis_camera c WHERE c.vis_mediamtx_node_id = n.id) AS cameras_atribuidas
FROM vis_mediamtx_node n
WHERE n.id = $1`, nodeID).Scan(
		&n.ID, &n.Nome, &n.RtmpPublic, &n.HlsPublic, &n.RtspInternal,
		&n.MaxCameras, &n.Ordem, &n.Status, &total,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, fmt.Errorf("No MediaMTX da camera nao encontrado")
		}
		return nil, 0, err
	}
	n.CamerasAtribuidas = total
	return &n, total, nil
}

func assignMediamtxNodeToCamera(ctx context.Context, cameraID, nodeID int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET vis_mediamtx_node_id = $2 WHERE id = $1`, cameraID, nodeID)
	return err
}

// ResolveMediamtxForCamera retorna URLs RTMP/HLS/RTSP do no da camera (atribui no se ausente).
func ResolveMediamtxForCamera(ctx context.Context, cameraID int, atribuirSeAusente bool) (map[string]any, error) {
	cam, err := GetCameraByID(ctx, cameraID)
	if err != nil {
		return nil, err
	}

	nodeID := intVal(cam, "vis_mediamtx_node_id")
	if nodeID <= 0 && atribuirSeAusente {
		picked, err := PickMediamtxNode(ctx)
		if err != nil {
			return nil, err
		}
		if err := assignMediamtxNodeToCamera(ctx, cameraID, picked.ID); err != nil {
			return nil, err
		}
		_ = SyncMediamtxNode(ctx, picked.ID)
		nodeID = picked.ID
	}
	if nodeID <= 0 {
		return nil, fmt.Errorf("Camera sem no MediaMTX — nenhum servidor disponivel")
	}

	node, total, err := getMediamtxNodeByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	rtmpPublic := sqlStr(node.RtmpPublic)
	hlsPublic := sqlStr(node.HlsPublic)
	rtspInternal := sqlStr(node.RtspInternal)
	if rtspInternal == "" {
		rtspInternal = "rtsp://127.0.0.1:8554"
	}

	return map[string]any{
		"vis_camera_id":        cameraID,
		"vis_mediamtx_node_id": node.ID,
		"nome":                 node.Nome,
		"rtmp_public":          rtmpPublic,
		"hls_public":           hlsPublic,
		"rtsp_internal":        rtspInternal,
		"max_cameras":          node.MaxCameras,
		"cameras_atribuidas":   total,
		"status":               node.Status,
	}, nil
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

func sqlStr(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return strings.TrimSpace(v.String)
}

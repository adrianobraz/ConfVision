package visdata

import (
	"confvision/src/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RtmpPublicToRtspPublic deriva base RTSP publica a partir de rtmp://host:1935.
func RtmpPublicToRtspPublic(rtmpPublic string) string {
	s := strings.TrimSpace(rtmpPublic)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil || u.Hostname() == "" {
		return ""
	}
	port := u.Port()
	if port == "" || port == "1935" {
		return fmt.Sprintf("rtsp://%s:8554", u.Hostname())
	}
	return fmt.Sprintf("rtsp://%s:%s", u.Hostname(), port)
}

// RtspURLSecForCamera monta rtsp://host:8554/cam/{hash} (Opção A).
func RtspURLSecForCamera(ctx context.Context, cameraID int) (string, error) {
	if cameraID < 1 {
		return "", fmt.Errorf("camera id invalido")
	}
	path := CameraStreamPath(cameraID)
	if path == "" {
		return "", fmt.Errorf("nao foi possivel gerar path cam/hash (RTMP_PUBLISH_SECRET?)")
	}
	mtx, err := ResolveMediamtxForCamera(ctx, cameraID, false)
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(strings.TrimSpace(config.MediamtxRtspBase), "/")
	if pub := strings.TrimSpace(fmt.Sprint(mtx["rtsp_public"])); pub != "" && pub != "<nil>" {
		base = strings.TrimRight(pub, "/")
	} else if rtmp := strings.TrimSpace(fmt.Sprint(mtx["rtmp_public"])); rtmp != "" && rtmp != "<nil>" {
		if derived := RtmpPublicToRtspPublic(rtmp); derived != "" {
			base = strings.TrimRight(derived, "/")
		}
	}
	if base == "" {
		return "", fmt.Errorf("camera %d sem base RTSP publica", cameraID)
	}
	return base + "/" + path, nil
}

// EnsureRtspURLSecForCamera preenche rtsp_url_sec se vazio ou legado live/.
func EnsureRtspURLSecForCamera(ctx context.Context, cameraID int) (updated bool, rtsp string, err error) {
	cam, err := GetCameraByID(ctx, cameraID)
	if err != nil {
		return false, "", err
	}
	cur := strings.TrimSpace(fmt.Sprint(cam["rtsp_url_sec"]))
	if cur != "" && cur != "<nil>" && !strings.Contains(strings.ToLower(cur), "/live/") {
		return false, cur, nil
	}
	rtsp, err = RtspURLSecForCamera(ctx, cameraID)
	if err != nil {
		return false, "", err
	}
	db, err := DB()
	if err != nil {
		return false, "", err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET rtsp_url_sec = $2 WHERE id = $1`, cameraID, rtsp)
	if err != nil {
		return false, "", err
	}
	return true, rtsp, nil
}

// BackfillMissingRtspURLSec preenche rtsp_url_sec em todas as cameras com campo vazio/legado.
func BackfillMissingRtspURLSec(ctx context.Context) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id FROM vis_camera
WHERE rtsp_url_sec IS NULL OR TRIM(rtsp_url_sec) = '' OR rtsp_url_sec ILIKE '%/live/%'
ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	updated := 0
	var fails []map[string]any
	for _, id := range ids {
		ok, _, err := EnsureRtspURLSecForCamera(ctx, id)
		if err != nil {
			fails = append(fails, map[string]any{"id": id, "erro": err.Error()})
			continue
		}
		if ok {
			updated++
		}
	}
	return map[string]any{
		"candidatas": len(ids),
		"atualizadas": updated,
		"falhas":      fails,
	}, nil
}

func markCamerasStreamOnline(ctx context.Context, cameraIDs []int) (int, error) {
	if len(cameraIDs) == 0 {
		return 0, nil
	}
	db, err := DB()
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	tag, err := db.ExecContext(ctx, `
UPDATE vis_camera SET
  status = 'online',
  ultimo_stream_ok_em = $2,
  stream_falhas_consecutivas = 0,
  stream_erro_classe = NULL,
  stream_ultimo_erro = NULL,
  stream_ultimo_erro_em = NULL,
  stream_motivo_pausa = NULL
WHERE id = ANY($1::int[])`, cameraIDs, now)
	if err != nil {
		return 0, err
	}
	n, _ := tag.RowsAffected()
	return int(n), nil
}

func markActiveCamerasStreamOfflineExcept(ctx context.Context, onlineIDs []int) (int, error) {
	db, err := DB()
	if err != nil {
		return 0, err
	}
	var n int64
	if len(onlineIDs) == 0 {
		res, err := db.ExecContext(ctx, `
UPDATE vis_camera SET status = 'offline'
WHERE ativo IS TRUE AND COALESCE(status, '') = 'online'`)
		if err != nil {
			return 0, err
		}
		n, _ = res.RowsAffected()
	} else {
		res, err := db.ExecContext(ctx, `
UPDATE vis_camera SET status = 'offline'
WHERE ativo IS TRUE
  AND COALESCE(status, '') = 'online'
  AND NOT (id = ANY($1::int[]))`, onlineIDs)
		if err != nil {
			return 0, err
		}
		n, _ = res.RowsAffected()
	}
	return int(n), nil
}

type guardOnlineItem struct {
	Path     string `json:"path"`
	CameraID int    `json:"camera_id"`
}

// SyncStreamStatusFromGuard consulta RTMP Guard /online e atualiza status no Postgres.
func SyncStreamStatusFromGuard(ctx context.Context) (map[string]any, error) {
	base := strings.TrimRight(strings.TrimSpace(config.RtmpGuardURL), "/")
	key := strings.TrimSpace(config.RtmpGuardAdminKey)
	if base == "" {
		return nil, fmt.Errorf("RTMP_GUARD_URL nao configurado")
	}
	if key == "" {
		return nil, fmt.Errorf("RTMP_GUARD_ADMIN_KEY nao configurado")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/online", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-RTMP-Guard-Key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("guard /online HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		Dados []guardOnlineItem `json:"dados"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse guard online: %w", err)
	}
	var onlineIDs []int
	for _, it := range parsed.Dados {
		if it.CameraID > 0 {
			onlineIDs = append(onlineIDs, it.CameraID)
		}
	}
	onMarked, err := markCamerasStreamOnline(ctx, onlineIDs)
	if err != nil {
		return nil, err
	}
	offMarked, err := markActiveCamerasStreamOfflineExcept(ctx, onlineIDs)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"online_guard":     len(onlineIDs),
		"camera_ids":       onlineIDs,
		"marcadas_online":  onMarked,
		"marcadas_offline": offMarked,
	}, nil
}

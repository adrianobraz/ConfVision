package confvision

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"confvision/src/config"
	"confvision/src/modulos/visdata"
)

func getCameraOperacional(ctx context.Context, cameraID int) (map[string]any, error) {
	if config.VisPostgresEnabled {
		return visdata.GetCameraByID(ctx, cameraID)
	}
	return fetchCameraMapXano(cameraID)
}

func resolveMediamtxOperacional(ctx context.Context, cameraID int) (map[string]string, error) {
	if config.VisPostgresEnabled {
		out, err := visdata.ResolveMediamtxForCamera(ctx, cameraID, true)
		if err != nil {
			return nil, err
		}
		mtx := map[string]string{
			"rtmp_public":          strings.TrimSpace(fmt.Sprint(out["rtmp_public"])),
			"hls_public":           strings.TrimSpace(fmt.Sprint(out["hls_public"])),
			"rtsp_internal":        strings.TrimSpace(fmt.Sprint(out["rtsp_internal"])),
			"nome":                 strings.TrimSpace(fmt.Sprint(out["nome"])),
			"vis_mediamtx_node_id": strings.TrimSpace(fmt.Sprint(out["vis_mediamtx_node_id"])),
		}
		if mtx["rtmp_public"] == "" || mtx["rtmp_public"] == "<nil>" {
			return nil, fmt.Errorf("mediamtx sem rtmp_public para camera %d", cameraID)
		}
		return mtx, nil
	}
	return fetchMediamtxForCamera(cameraID)
}

func fetchCameraMapXano(cameraID int) (map[string]any, error) {
	if strings.TrimSpace(config.XanoBaseUrl) == "" {
		return nil, fmt.Errorf("postgres indisponivel e XANO_BASE_URL nao configurado")
	}
	id := strconv.Itoa(cameraID)
	u := fmt.Sprintf("%s/vis_camera/%s?vis_camera_id=%s",
		strings.TrimRight(config.XanoBaseUrl, "/"), id, id)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("camera nao encontrada")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xano HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var cam map[string]any
	if err := json.Unmarshal(raw, &cam); err != nil {
		return nil, fmt.Errorf("parse camera: %w", err)
	}
	if d, ok := cam["dados"].(map[string]any); ok {
		return d, nil
	}
	return cam, nil
}

func cameraMapToRecord(cam map[string]any) *visCameraRecord {
	if cam == nil {
		return nil
	}
	rec := &visCameraRecord{
		IdFranqueado: strings.TrimSpace(fmt.Sprint(cam["id_franqueado"])),
		SnapshotURL:  strings.TrimSpace(fmt.Sprint(cam["snapshot_url"])),
		Plano:        strings.TrimSpace(fmt.Sprint(cam["plano"])),
	}
	switch v := cam["id"].(type) {
	case float64:
		rec.ID = int(v)
	case int:
		rec.ID = v
	case json.Number:
		n, _ := v.Int64()
		rec.ID = int(n)
	default:
		fmt.Sscanf(fmt.Sprint(v), "%d", &rec.ID)
	}
	rec.Bloqueado = TruthyCameraBool(cam["bloqueado"])
	rec.Ativo = TruthyCameraBool(cam["ativo"])
	return rec
}

func buscarCameraOperacional(ctx context.Context, cameraID string) (*visCameraRecord, error) {
	id, err := strconv.Atoi(strings.TrimSpace(cameraID))
	if err != nil || id < 1 {
		return nil, fmt.Errorf("id de camera invalido")
	}
	cam, err := getCameraOperacional(ctx, id)
	if err != nil {
		return nil, err
	}
	rec := cameraMapToRecord(cam)
	if rec == nil || rec.ID == 0 {
		return nil, fmt.Errorf("camera nao encontrada")
	}
	return rec, nil
}

func atualizarSnapshotOperacional(ctx context.Context, cameraID int, url *string) error {
	if config.VisPostgresEnabled {
		val := ""
		if url != nil {
			val = *url
		}
		return visdata.UpdateCameraSnapshot(ctx, cameraID, val)
	}
	idStr := strconv.Itoa(cameraID)
	if url == nil {
		return limparCadSnapshotXano(idStr)
	}
	return atualizarCadSnapshotXano(idStr, url)
}

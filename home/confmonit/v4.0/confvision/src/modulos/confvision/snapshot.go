package confvision

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/storage"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type snapshotRequest struct {
	IdFranqueado string `json:"id_franqueado"`
}

type visCameraRecord struct {
	ID           int    `json:"id"`
	IdFranqueado string `json:"id_franqueado"`
	SnapshotURL  string `json:"snapshot_url"`
	Bloqueado    bool   `json:"bloqueado"`
	Ativo        bool   `json:"ativo"`
	Plano        string `json:"plano"`
}

func TirarSnapshotCamera(w http.ResponseWriter, r *http.Request) {
	cameraID := mux.Vars(r)["id"]
	if strings.TrimSpace(cameraID) == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id da camera obrigatorio"))
		return
	}

	var req snapshotRequest
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				auxiliar.RespostaErro(w, http.StatusBadRequest, err)
				return
			}
		}
	}

	cam, err := buscarCameraOperacional(r.Context(), cameraID)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if cam.ID == 0 {
		auxiliar.RespostaErro(w, http.StatusNotFound, fmt.Errorf("camera nao encontrada"))
		return
	}

	idFranqueado := strings.TrimSpace(req.IdFranqueado)
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}
	if strings.TrimSpace(cam.IdFranqueado) != idFranqueado {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("camera nao pertence ao franqueado"))
		return
	}

	if ok, motivo := CameraPodeStream(cam.Bloqueado, cam.Ativo, cam.Plano); !ok {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("%s", cameraPodeStreamMsg(motivo)))
		return
	}

	chave := ChaveRtmp(cam.ID)
	if chave == "" {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("RTMP_PUBLISH_SECRET nao configurado"))
		return
	}
	pathLive := StreamPath(cam.ID)
	rtspURL := fmt.Sprintf("%s/%s", strings.TrimRight(config.MediamtxRtspBase, "/"), pathLive)
	ctx := r.Context()
	log.Printf("confvision snapshot camera=%s rtsp=%s franqueado=%s", cameraID, rtspURL, idFranqueado)

	raw, err := storage.CapturarFrameRTSP(ctx, rtspURL)
	if err != nil {
		log.Printf("confvision snapshot captura falhou camera=%s: %v", cameraID, err)
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	url, err := storage.SubstituirCadSnapshot(ctx, cam.ID, raw)
	if err != nil {
		log.Printf("confvision snapshot upload falhou camera=%s: %v", cameraID, err)
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	if err := atualizarSnapshotOperacional(ctx, cam.ID, &url); err != nil {
		log.Printf("confvision snapshot persistencia falhou camera=%s: %v", cameraID, err)
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	log.Printf("confvision snapshot ok camera=%s url=%s", cameraID, url)

	auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
		"snapshot_url": url,
	})
}

func RemoverSnapshotCamera(w http.ResponseWriter, r *http.Request) {
	cameraID := mux.Vars(r)["id"]
	if strings.TrimSpace(cameraID) == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id da camera obrigatorio"))
		return
	}

	var req snapshotRequest
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				auxiliar.RespostaErro(w, http.StatusBadRequest, err)
				return
			}
		}
	}

	cam, err := buscarCameraOperacional(r.Context(), cameraID)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if cam.ID == 0 {
		auxiliar.RespostaErro(w, http.StatusNotFound, fmt.Errorf("camera nao encontrada"))
		return
	}

	idFranqueado := strings.TrimSpace(req.IdFranqueado)
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}
	if strings.TrimSpace(cam.IdFranqueado) != idFranqueado {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("camera nao pertence ao franqueado"))
		return
	}

	ctx := r.Context()
	if err := storage.RemoverCadSnapshot(ctx, cam.ID); err != nil {
		log.Printf("confvision snapshot remover storage falhou camera=%s: %v", cameraID, err)
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	if err := atualizarSnapshotOperacional(ctx, cam.ID, nil); err != nil {
		log.Printf("confvision snapshot remover persistencia falhou camera=%s: %v", cameraID, err)
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	log.Printf("confvision snapshot removido camera=%s", cameraID)

	auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
		"snapshot_url": nil,
	})
}

func buscarCameraXano(cameraID string) (*visCameraRecord, error) {
	url := fmt.Sprintf("%s/vis_camera/%s?vis_camera_id=%s", config.XanoBaseUrl, cameraID, cameraID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xano GET camera: %s", string(body))
	}

	var cam visCameraRecord
	if err := json.Unmarshal(body, &cam); err != nil {
		var wrapped struct {
			Dados visCameraRecord `json:"dados"`
		}
		if err2 := json.Unmarshal(body, &wrapped); err2 != nil {
			return nil, err
		}
		cam = wrapped.Dados
	}
	return &cam, nil
}

func atualizarCadSnapshotXano(cameraID string, url *string) error {
	path := fmt.Sprintf("%s/vis_camera/snapshot/%s?vis_camera_id=%s",
		config.XanoBaseUrl, cameraID, cameraID)

	cameraIDInt := 0
	fmt.Sscanf(cameraID, "%d", &cameraIDInt)

	var snapshotVal interface{}
	if url != nil {
		snapshotVal = *url
	}

	payload, err := json.Marshal(map[string]interface{}{
		"vis_camera_id": cameraIDInt,
		"snapshot_url":  snapshotVal,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("xano PUT snapshot: %s", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	log.Printf("confvision snapshot xano resposta camera=%s: %s", cameraID, string(body))
	return nil
}

func limparCadSnapshotXano(cameraID string) error {
	return atualizarCadSnapshotXano(cameraID, nil)
}

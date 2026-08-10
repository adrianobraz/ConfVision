package visapi

import (
	"confvision/src/conexao"
	"confvision/src/config"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

var (
	repoOnce sync.Once
	repo     *Repository
	repoErr  error
)

func repoInstance() (*Repository, error) {
	repoOnce.Do(func() {
		if !config.VisPostgresEnabled {
			repoErr = sql.ErrConnDone
			return
		}
		db, err := conexao.ConectarPostgres()
		if err != nil {
			repoErr = err
			return
		}
		repo = NewRepository(db)
	})
	return repo, repoErr
}

func postgresReady() bool {
	if !config.VisPostgresEnabled {
		return false
	}
	_, err := repoInstance()
	return err == nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseNodeID(r *http.Request) int {
	v := strings.TrimSpace(r.URL.Query().Get("vis_mediamtx_node_id"))
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func HandleCameraQueryAtivas(w http.ResponseWriter, r *http.Request) {
	if !postgresReady() {
		http.Error(w, `{"erro":"postgres nao configurado"}`, http.StatusServiceUnavailable)
		return
	}
	repository, err := repoInstance()
	if err != nil {
		http.Error(w, `{"erro":"postgres indisponivel"}`, http.StatusServiceUnavailable)
		return
	}

	workerID := strings.TrimSpace(r.URL.Query().Get("worker_id"))
	nodeID := parseNodeID(r)

	cameras, err := repository.ListCamerasAnaliticas(workerID, nodeID)
	if err != nil {
		log.Printf("[visapi] query_ativas: %v", err)
		http.Error(w, `{"erro":"falha ao listar cameras"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"dados": cameras})
}

func HandleCameraSyncAtivas(w http.ResponseWriter, r *http.Request) {
	if !postgresReady() {
		http.Error(w, `{"erro":"postgres nao configurado"}`, http.StatusServiceUnavailable)
		return
	}
	repository, err := repoInstance()
	if err != nil {
		http.Error(w, `{"erro":"postgres indisponivel"}`, http.StatusServiceUnavailable)
		return
	}

	workerID := strings.TrimSpace(r.URL.Query().Get("worker_id"))
	nodeID := parseNodeID(r)
	sinceVersion := strings.TrimSpace(r.URL.Query().Get("since_version"))
	includeGravacao := strings.EqualFold(r.URL.Query().Get("include_gravacao"), "true")

	cameras, err := repository.ListCamerasAnaliticas(workerID, nodeID)
	if err != nil {
		log.Printf("[visapi] sync_ativas cameras: %v", err)
		http.Error(w, `{"erro":"falha ao listar cameras"}`, http.StatusInternalServerError)
		return
	}

	areas, err := repository.ListAreasAtivas()
	if err != nil {
		log.Printf("[visapi] sync_ativas areas: %v", err)
		http.Error(w, `{"erro":"falha ao listar areas"}`, http.StatusInternalServerError)
		return
	}

	configVersion := buildConfigVersion(cameras, areas)
	if sinceVersion != "" && sinceVersion == configVersion {
		writeJSON(w, http.StatusOK, map[string]any{
			"unchanged":      true,
			"config_version": configVersion,
		})
		return
	}

	payload := map[string]any{
		"unchanged":      false,
		"config_version": configVersion,
		"cameras":        cameras,
		"areas":          areas,
		"gravacao":       []any{},
	}

	if includeGravacao {
		gravacao, err := repository.ListGravacaoCameras(workerID, nodeID)
		if err != nil {
			log.Printf("[visapi] sync_ativas gravacao: %v", err)
			http.Error(w, `{"erro":"falha ao listar gravacao"}`, http.StatusInternalServerError)
			return
		}
		payload["gravacao"] = gravacao
	}

	writeJSON(w, http.StatusOK, payload)
}

func HandleCameraAreaQueryAtivas(w http.ResponseWriter, r *http.Request) {
	if !postgresReady() {
		http.Error(w, `{"erro":"postgres nao configurado"}`, http.StatusServiceUnavailable)
		return
	}
	repository, err := repoInstance()
	if err != nil {
		http.Error(w, `{"erro":"postgres indisponivel"}`, http.StatusServiceUnavailable)
		return
	}

	areas, err := repository.ListAreasAtivas()
	if err != nil {
		log.Printf("[visapi] area_query_ativas: %v", err)
		http.Error(w, `{"erro":"falha ao listar areas"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"dados": areas})
}

func HandleCameraRTMPAuth(w http.ResponseWriter, r *http.Request) {
	if !postgresReady() {
		http.Error(w, `{"erro":"postgres nao configurado"}`, http.StatusServiceUnavailable)
		return
	}
	repository, err := repoInstance()
	if err != nil {
		http.Error(w, `{"erro":"postgres indisponivel"}`, http.StatusServiceUnavailable)
		return
	}

	idStr := mux.Vars(r)["vis_camera_id"]
	cameraID, err := strconv.Atoi(idStr)
	if err != nil || cameraID <= 0 {
		http.Error(w, `{"erro":"camera invalida"}`, http.StatusBadRequest)
		return
	}

	row, err := repository.GetCameraRTMPAuth(cameraID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"erro":"nao encontrado"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("[visapi] rtmp_auth: %v", err)
		http.Error(w, `{"erro":"falha auth"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, row)
}

func HandleWorkerPing(w http.ResponseWriter, r *http.Request) {
	if !postgresReady() {
		http.Error(w, `{"erro":"postgres nao configurado"}`, http.StatusServiceUnavailable)
		return
	}
	repository, err := repoInstance()
	if err != nil {
		http.Error(w, `{"erro":"postgres indisponivel"}`, http.StatusServiceUnavailable)
		return
	}

	var input WorkerPingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"erro":"json invalido"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.WorkerID) == "" {
		http.Error(w, `{"erro":"worker_id obrigatorio"}`, http.StatusBadRequest)
		return
	}

	model, err := repository.UpsertWorkerPing(input)
	if err != nil {
		log.Printf("[visapi] worker_ping: %v", err)
		http.Error(w, `{"erro":"falha ao registrar ping"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, model)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	if !postgresReady() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":  "postgres_off",
			"enabled": config.VisPostgresEnabled,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"enabled": true,
	})
}

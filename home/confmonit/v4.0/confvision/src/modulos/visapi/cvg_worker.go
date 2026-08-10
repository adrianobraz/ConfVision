package visapi

import (
	"confvision/src/modulos/visdata"
	"encoding/json"
	"io"
	"net/http"
)

func handleCvgWorkerTick(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	var payload map[string]any
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}

	workerKey := r.Header.Get("X-CVG-Worker-Key")
	status, raw, err := visdata.CvgWorkerTickPOST(r.Context(), payload, workerKey)
	if err != nil {
		http.Error(w, `{"erro":"`+visdata.ErrorMessage(err)+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

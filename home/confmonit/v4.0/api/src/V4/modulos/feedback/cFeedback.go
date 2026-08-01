package feedback

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func enviar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var f Feedback
	if err := json.Unmarshal(body, &f); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := f.Enviar(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, map[string]string{
		"idFeedback": f.ID_Feedback,
	})
}

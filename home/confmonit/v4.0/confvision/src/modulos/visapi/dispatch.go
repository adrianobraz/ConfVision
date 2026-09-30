package visapi

import (
	"bytes"
	"confvision/src/modulos/visdata"
	"io"
	"net/http"
)

func handleDispatch(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	ctx := visdata.ContextWithHTTPRequest(r.Context(), r)
	status, raw, err := visdata.Dispatch(ctx, r.Method, path, bytes.NewReader(bodyBytes))
	if err != nil {
		if err == visdata.ErrNotHandled {
			http.Error(w, `{"erro":"rota nao implementada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"erro":"`+visdata.ErrorMessage(err)+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

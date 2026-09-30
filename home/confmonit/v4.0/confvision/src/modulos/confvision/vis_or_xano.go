package confvision

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/modulos/visdata"
	"errors"
	"io"
	"net/http"
)

// proxyVisOrXano encaminha para visdata.Dispatch (Postgres) ou fallback Xano.
func proxyVisOrXano(w http.ResponseWriter, r *http.Request, metodo, path string) {
	var body []byte
	if metodo != http.MethodGet && metodo != http.MethodHead {
		corpo, err := io.ReadAll(r.Body)
		if err != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
		body = corpo
	}

	if config.VisPostgresEnabled {
		status, raw, err := visdata.Dispatch(r.Context(), metodo, path, bytes.NewReader(body))
		if err == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			_, _ = w.Write(raw)
			return
		}
		if !errors.Is(err, visdata.ErrNotHandled) {
			auxiliar.RespostaErro(w, http.StatusBadGateway, err)
			return
		}
	}

	if config.XanoBaseUrl == "" {
		http.Error(w, `{"erro":"Postgres indisponivel e XANO_BASE_URL nao configurado"}`, http.StatusBadGateway)
		return
	}

	url := config.XanoBaseUrl + path
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(r.Context(), metodo, url, reader)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(raw)
		return
	}
	auxiliar.RespostaAPP(w, raw)
}

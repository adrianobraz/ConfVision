package confservice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"apifunction/config"
)

// ListEventosFranqueado GET /internal/eventos?idFranqueado=...
func ListEventosFranqueado(idFranqueado, de, ate, status, idCliente string, limite int) (map[string]any, error) {
	base := config.ConfServiceURL
	if base == "" {
		return nil, fmt.Errorf("CONFSERVICE_URL nao configurado")
	}
	q := url.Values{}
	q.Set("idFranqueado", idFranqueado)
	if de != "" {
		q.Set("de", de)
	}
	if ate != "" {
		q.Set("ate", ate)
	}
	if status != "" {
		q.Set("status", status)
	}
	if idCliente != "" {
		q.Set("idCliente", idCliente)
	}
	if limite > 0 {
		q.Set("limite", fmt.Sprintf("%d", limite))
	}
	req, err := http.NewRequest(http.MethodGet, base+"/internal/eventos?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("confservice HTTP %d: %s", res.StatusCode, string(raw))
	}
	var out map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

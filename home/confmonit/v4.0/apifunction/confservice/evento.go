package confservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"apifunction/config"
)

// EnfileirarEvento POST /internal/evento — fila webhook parceiro (Moni).
func EnfileirarEvento(payload map[string]any) (map[string]any, error) {
	base := config.ConfServiceURL
	if base == "" {
		return nil, fmt.Errorf("CONFSERVICE_URL nao configurado")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, base+"/internal/evento", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	client := &http.Client{Timeout: 12 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("confservice HTTP %d: %s", res.StatusCode, string(raw))
	}
	var out map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out, nil
}

package confserviceparceiroV4

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"api/src/V4/config"
)

func csRequest(method, path string, body []byte) ([]byte, int, error) {
	base := strings.TrimRight(config.ConfServiceURL, "/")
	if base == "" {
		return nil, 0, fmt.Errorf("CONFSERVICE_URL nao configurado")
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceKey)
	}
	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return raw, resp.StatusCode, err
}

func csErroMsg(raw []byte) string {
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err == nil {
		if e, ok := parsed["erro"].(string); ok && e != "" {
			return e
		}
	}
	return "confservice erro"
}

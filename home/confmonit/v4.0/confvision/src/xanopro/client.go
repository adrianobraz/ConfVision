package xanopro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"confvision/src/config"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

func Post(path string, payload any) ([]byte, error) {
	status, raw, err := PostWithStatus(path, payload)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("xano pro HTTP %d: %s", status, string(raw))
	}
	return raw, nil
}

func PostWithStatus(path string, payload any) (int, []byte, error) {
	api := config.XanoApiPro
	if api == "" {
		return 0, nil, fmt.Errorf("XANO_API_FRANQUEADO_PRO nao configurado")
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}

	url := api + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}
	return res.StatusCode, raw, nil
}

package xano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"franqueadopro/src/config"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

func Do(method, path string, payload any) ([]byte, error) {
	if config.XanoApi == "" {
		return nil, fmt.Errorf("XANO_API_FRANQUEADO nao configurado")
	}

	var body io.Reader
	if payload != nil && method != http.MethodGet && method != http.MethodDelete {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(b)
	}

	url := config.XanoApi + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("xano HTTP %d: %s", res.StatusCode, string(raw))
	}
	return raw, nil
}

func Get(path string) ([]byte, error) {
	return Do(http.MethodGet, path, nil)
}

func Put(path string, payload any) ([]byte, error) {
	return Do(http.MethodPut, path, payload)
}

func Delete(path string) ([]byte, error) {
	return Do(http.MethodDelete, path, nil)
}

func Post(path string, payload any) ([]byte, error) {
	if config.XanoApi == "" {
		return nil, fmt.Errorf("XANO_API_FRANQUEADO nao configurado")
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := config.XanoApi + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("xano HTTP %d: %s", res.StatusCode, string(raw))
	}
	return raw, nil
}

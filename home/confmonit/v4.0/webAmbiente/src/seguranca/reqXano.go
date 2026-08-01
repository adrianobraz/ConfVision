package seguranca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"webAmbiente/src/config"
)

var xanoHTTP = &http.Client{Timeout: 20 * time.Second}

func ReqXano(metodo, path string, body io.Reader) (*http.Response, error) {
	if config.Xano == "" {
		return nil, fmt.Errorf("XANO_API nao configurado")
	}
	url := config.Xano + path
	req, err := http.NewRequest(metodo, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return xanoHTTP.Do(req)
}

func ReqXanoGET(path string) (*http.Response, error) {
	return ReqXano(http.MethodGet, path, nil)
}

func ReqXanoJSON(path string, payload any) (*http.Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return ReqXano(http.MethodPost, path, bytes.NewBuffer(b))
}

func ReqXanoPUT(path string, payload any) (*http.Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return ReqXano(http.MethodPut, path, bytes.NewBuffer(b))
}

// ReqXanoParaBase POST JSON em uma base Xano alternativa (ex.: centerOperacion).
func ReqXanoParaBase(baseURL, path string, payload any) (*http.Response, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL Xano nao configurada")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return xanoHTTP.Do(req)
}

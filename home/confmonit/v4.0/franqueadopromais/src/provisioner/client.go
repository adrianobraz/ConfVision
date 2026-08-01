package provisioner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"franqueadopro/src/config"
)

type Result struct {
	OK      bool   `json:"ok"`
	SSLOK   bool   `json:"ssl_ok"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Slug    string `json:"slug"`
}

var httpClient = &http.Client{Timeout: 180 * time.Second}

func post(path string, fqdn string) (Result, error) {
	var out Result
	if config.ProvisionerURL == "" {
		return out, fmt.Errorf("PROVISIONER_URL nao configurado")
	}

	body, _ := json.Marshal(map[string]string{"fqdn": fqdn})
	req, err := http.NewRequest(http.MethodPost, config.ProvisionerURL+path, bytes.NewBuffer(body))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.ProvisionerKey != "" {
		req.Header.Set("X-Provisioner-Key", config.ProvisionerKey)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}
	if res.StatusCode >= 400 {
		return out, fmt.Errorf("provisioner HTTP %d: %s", res.StatusCode, string(raw))
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

func Provisionar(fqdn string) (Result, error) {
	return post("/provisionar", fqdn)
}

func Remover(fqdn string) (Result, error) {
	res, err := post("/remover", fqdn)
	if err != nil {
		return res, err
	}
	if !res.OK {
		return res, fmt.Errorf("%s", res.Message)
	}
	return res, nil
}

func RetentarSSL(fqdn string) (Result, error) {
	return post("/retentar-ssl", fqdn)
}

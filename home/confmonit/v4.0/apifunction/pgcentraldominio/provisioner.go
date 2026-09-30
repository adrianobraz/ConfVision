package pgcentraldominio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"apifunction/config"
)

type ProvResult struct {
	OK      bool   `json:"ok"`
	SSLOK   bool   `json:"ssl_ok"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

var provHTTP = &http.Client{Timeout: 180 * time.Second}

func provisionerPost(path, fqdn, app string) (ProvResult, error) {
	var out ProvResult
	url := config.ProvisionerURL + path
	if url == path || config.ProvisionerURL == "" {
		return out, fmt.Errorf("PROVISIONER_URL nao configurado")
	}
	if app == "" {
		app = "franqueadopro"
	}
	body, _ := json.Marshal(map[string]string{"fqdn": fqdn, "app": app})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.ProvisionerKey != "" {
		req.Header.Set("X-Provisioner-Key", config.ProvisionerKey)
	}
	res, err := provHTTP.Do(req)
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
	_ = json.Unmarshal(raw, &out)
	return out, nil
}

func ProvisionarApp(fqdn, app string) (ProvResult, error) {
	return provisionerPost("/provisionar", fqdn, app)
}

func RemoverApp(fqdn, app string) (ProvResult, error) {
	return provisionerPost("/remover", fqdn, app)
}

func RetentarSSLApp(fqdn, app string) (ProvResult, error) {
	return provisionerPost("/retentar-ssl", fqdn, app)
}

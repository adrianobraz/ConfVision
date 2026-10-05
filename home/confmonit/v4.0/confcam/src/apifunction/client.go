package apifunction

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"confvision/src/config"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

func baseURL() string {
	u := strings.TrimRight(strings.TrimSpace(config.ApiFunctionURL), "/")
	if u == "" {
		return "http://127.0.0.1:20001"
	}
	return u
}

func MarcaResolve(fqdn, idCentral, app string) (map[string]any, error) {
	q := url.Values{}
	if strings.TrimSpace(fqdn) != "" {
		q.Set("fqdn", strings.TrimSpace(fqdn))
	}
	if strings.TrimSpace(idCentral) != "" {
		q.Set("id_central", strings.TrimSpace(idCentral))
	}
	if strings.TrimSpace(app) != "" {
		q.Set("app", strings.TrimSpace(app))
	}
	reqURL := baseURL() + "/marca/resolve?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
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
		return nil, fmt.Errorf("apifunction HTTP %d", res.StatusCode)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

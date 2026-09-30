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

// MarcaResolve consulta apifunction (Postgres fp_whitelabel / central) por fqdn ou id_central.
// Retorna o JSON decodificado (espera chaves ok, dados, app, id_central).
func MarcaResolve(fqdn, idCentral, app string) (map[string]any, error) {
	base := strings.TrimSpace(config.ApiFunctionURL)
	if base == "" {
		return nil, fmt.Errorf("APIFUNCTION_URL nao configurado")
	}

	q := url.Values{}
	if s := strings.TrimSpace(fqdn); s != "" {
		q.Set("fqdn", s)
	}
	if s := strings.TrimSpace(idCentral); s != "" {
		q.Set("id_central", s)
	}
	if s := strings.TrimSpace(app); s != "" {
		q.Set("app", s)
	}

	u := strings.TrimRight(base, "/") + "/marca/resolve"
	if enc := q.Encode(); enc != "" {
		u += "?" + enc
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("apifunction HTTP %d: %s", res.StatusCode, string(raw))
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("apifunction resposta invalida: %w", err)
	}
	return out, nil
}

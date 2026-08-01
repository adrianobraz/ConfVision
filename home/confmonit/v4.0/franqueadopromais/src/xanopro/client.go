package xanopro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"franqueadopro/src/config"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

func Post(path string, payload any) ([]byte, error) {
	api := config.XanoApiPro
	if api == "" {
		api = config.XanoApi
	}
	if api == "" {
		return nil, fmt.Errorf("XANO_API_FRANQUEADO_PRO nao configurado")
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := api + path
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
		if msg := extrairMsgXano(raw); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("nao foi possivel concluir a operacao")
	}
	return raw, nil
}

// extrairMsgXano pega só o "message" amigável do JSON de erro do Xano.
func extrairMsgXano(raw []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if m, ok := payload["message"].(string); ok {
		m = strings.TrimSpace(m)
		if m != "" {
			return m
		}
	}
	return ""
}

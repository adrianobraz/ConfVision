package pgatendimento

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var tokenCache struct {
	mu    sync.Mutex
	token string
	expAt time.Time
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func httpJSON(ctx context.Context, method, rawURL string, payload any, headers map[string]string) ([]byte, int, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

func extractDados(body []byte) (map[string]any, error) {
	var wrap map[string]any
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	if d, ok := wrap["dados"].(map[string]any); ok {
		return d, nil
	}
	if r, ok := wrap["result"].(map[string]any); ok {
		if d, ok := r["dados"].(map[string]any); ok {
			return d, nil
		}
	}
	return nil, fmt.Errorf("dados ausente")
}

func extractToken(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		for _, key := range []string{"token", "Token", "access_token", "jwt"} {
			if vv, ok := x[key]; ok {
				if t := extractToken(vv); t != "" {
					return t
				}
			}
		}
	case []any:
		for _, el := range x {
			if t := extractToken(el); t != "" {
				return t
			}
		}
	}
	return ""
}

func webLogarToken(ctx context.Context) (string, error) {
	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()

	if tokenCache.token != "" && time.Now().Before(tokenCache.expAt.Add(-5*time.Minute)) {
		return tokenCache.token, nil
	}

	urlLogar := envOr("OPS_WEBLOGAR_URL", envOr("CVG_WEBLOGAR_URL", "http://185.130.61.4:2010/v4/cliente/webLogar"))
	senha := envOr("OPS_WEBLOGAR_SENHA", envOr("CVG_WEBLOGAR_SENHA", ""))
	if senha == "" {
		return "", fmt.Errorf("OPS_WEBLOGAR_SENHA nao configurada")
	}

	body, status, err := httpJSON(ctx, http.MethodPost, urlLogar, map[string]string{"senha": senha}, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("webLogar status=%d", status)
	}

	var resp struct {
		Status string `json:"status"`
		Dados  any    `json:"dados"`
		Result struct {
			Dados any `json:"dados"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	token := extractToken(resp.Dados)
	if token == "" {
		token = extractToken(resp.Result.Dados)
	}
	if token == "" {
		return "", fmt.Errorf("token nao encontrado na resposta webLogar")
	}

	tokenCache.token = token
	tokenCache.expAt = time.Now().Add(12 * time.Hour)
	return token, nil
}

func getDadosProcessoByID(ctx context.Context, idProcesso string) (map[string]any, error) {
	base := envOr("OPS_CONFMONIT_BASE", "http://185.130.61.4:2010")
	urlProc := envOr("OPS_GET_PROCESSO_URL", base+"/v4/terminal/getDadosProcessoById")

	token, err := webLogarToken(ctx)
	if err != nil {
		return nil, err
	}

	body, status, err := httpJSON(ctx, http.MethodPost, urlProc, map[string]any{
		"idProcesso": idProcesso,
	}, map[string]string{"Authorization": "Bearer " + token})
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("getDadosProcessoById status=%d", status)
	}
	return extractDados(body)
}

func processoEnd(ctx context.Context, idProcesso, motivo, perfil string) (map[string]any, string, error) {
	descricao := fmt.Sprintf("[AUTO] %s [%s]", motivo, perfil)

	procDados, err := getDadosProcessoByID(ctx, idProcesso)
	if err != nil {
		return nil, "", err
	}

	dataAtenFim := ""
	if procDados != nil {
		if v, ok := procDados["dataAtenFim"].(string); ok {
			dataAtenFim = strings.TrimSpace(v)
		}
	}

	if dataAtenFim != "" && dataAtenFim != "01/01/0001 00:00:00" {
		return map[string]any{"dados": []any{}}, "JA_FINALIZADO", nil
	}

	base := envOr("OPS_CONFMONIT_BASE", "http://185.130.61.4:2010")
	urlFinalizar := envOr("OPS_FINALIZAR_PROCESSO_URL", base+"/v4/terminal/finalizarProcesso")

	token, err := webLogarToken(ctx)
	if err != nil {
		return nil, "", err
	}

	body, status, err := httpJSON(ctx, http.MethodPost, urlFinalizar, map[string]any{
		"idProcesso": idProcesso,
		"idCliente":  "0ROBOAUTO",
		"descricao":  descricao,
		"nome":       "ROBO AUTO",
	}, map[string]string{"Authorization": "Bearer " + token})
	if err != nil {
		return nil, "", err
	}
	if status < 200 || status >= 300 {
		return nil, "", fmt.Errorf("finalizarProcesso status=%d", status)
	}

	dados, _ := extractDados(body)
	ret := map[string]any{"dados": dados}
	acao := "FINALIZOU"
	if dados == nil {
		acao = "JA_FINALIZADO"
		ret["dados"] = []any{}
	}
	return ret, acao, nil
}

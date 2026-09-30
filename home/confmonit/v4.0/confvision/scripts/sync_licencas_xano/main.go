// Sync licencas Xano -> Postgres via API publica ConfVision.
// Se sync_lote (404), faz POST /vis_licenca por licenca (sem preservar ID Xano).
//
// go run .
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	idFranqueado = "2025050602383727046281876"
	xanoURL      = "https://xpcy-oyme-lno7.b2.xano.io/api:mNF05uZd/fp_confvision_resumo_franqueado"
	cvBase       = "https://vision.confmonit2.com.br"
	workerKey    = "a7f3c9e2-8b1d-4f6a-9c0e-vis-bridge-2026"
)

func main() {
	client := &http.Client{Timeout: 60 * time.Second}
	key := os.Getenv("VIS_WORKER_API_KEY")
	if key == "" {
		key = workerKey
	}

	raw, err := post(client, xanoURL, map[string]any{"id_franqueado": idFranqueado}, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "xano:", err)
		os.Exit(1)
	}
	var xano map[string]any
	_ = json.Unmarshal(raw, &xano)
	lics, _ := xano["licencas"].([]any)
	fmt.Printf("Xano: %d licencas\n", len(lics))

	body, _ := json.Marshal(map[string]any{"licencas": lics})
	code, resp, err := postRaw(client, cvBase+"/ops/vis_licenca/sync_lote", body, key)
	if err == nil && code < 400 {
		fmt.Printf("sync_lote OK (%d): %s\n", code, resp)
		return
	}
	fmt.Printf("sync_lote indisponivel (%d): %s — fallback POST /vis_licenca\n", code, resp)

	ok, fail := 0, 0
	for _, el := range lics {
		m, _ := el.(map[string]any)
		payload := map[string]any{
			"id_franqueado": idFranqueado,
			"plano":         str(m["plano"]),
			"unidade":       str(m["unidade"]),
			"status":        str(m["status"]),
			"observacao":    str(m["observacao"]),
		}
		if v := tsRFC(m["pago_em"]); v != "" {
			payload["pago_em"] = v
		}
		if v := tsRFC(m["valido_ate"]); v != "" {
			payload["valido_ate"] = v
		}
		b, _ := json.Marshal(payload)
		c, r, e := postRaw(client, cvBase+"/vis_licenca", b, key)
		if e != nil || c >= 400 {
			fail++
			fmt.Printf("  FAIL plano=%s: %v %s\n", payload["plano"], e, r)
			continue
		}
		ok++
	}
	fmt.Printf("Fallback: %d ok, %d fail\n", ok, fail)

	c, r, _ := get(client, cvBase+"/vis_licenca_by_franqueado?id_franqueado="+idFranqueado, key)
	fmt.Printf("Postgres resumo HTTP %d: %s\n", c, trunc(r, 400))
}

func post(client *http.Client, url string, payload any, key string) ([]byte, error) {
	b, _ := json.Marshal(payload)
	c, r, err := postRaw(client, url, b, key)
	if err != nil {
		return nil, err
	}
	if c >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", c, r)
	}
	return []byte(r), nil
}

func postRaw(client *http.Client, url string, body []byte, key string) (int, string, error) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-Vis-Worker-Key", key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

func get(client *http.Client, url, key string) (int, string, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	if key != "" {
		req.Header.Set("X-Vis-Worker-Key", key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func tsRFC(v any) string {
	var ms int64
	switch t := v.(type) {
	case float64:
		ms = int64(t)
	default:
		return ""
	}
	if ms > 1e12 {
		return time.UnixMilli(ms).UTC().Format(time.RFC3339)
	}
	return time.Unix(ms, 0).UTC().Format(time.RFC3339)
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

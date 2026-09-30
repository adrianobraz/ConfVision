package pgreceptordiag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"apifunction/config"
)

func RunAgentChecks(modulo, porta, conta, idFisico string) (*AgentRunResult, error) {
	base := strings.TrimRight(strings.TrimSpace(config.ReceptorDiagAgentURL), "/")
	if base == "" {
		return nil, fmt.Errorf("RECEPTOR_DIAG_AGENT_URL nao configurado")
	}
	q := url.Values{}
	q.Set("modulo", modulo)
	q.Set("porta", porta)
	if conta != "" {
		q.Set("conta", conta)
	}
	if idFisico != "" {
		q.Set("id_fisico", idFisico)
	}
	reqURL := base + "/diag/run?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(config.ReceptorDiagAgentKey)
	if key != "" {
		req.Header.Set("X-Diag-Key", key)
	}
	client := &http.Client{Timeout: 25 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("agente HTTP %d: %s", res.StatusCode, string(raw))
	}
	var wrap struct {
		OK    bool            `json:"ok"`
		Dados AgentRunResult  `json:"dados"`
		Erro  string          `json:"erro"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, err
	}
	if !wrap.OK {
		if wrap.Erro != "" {
			return nil, fmt.Errorf(wrap.Erro)
		}
		return nil, fmt.Errorf("agente retornou erro")
	}
	return &wrap.Dados, nil
}

func TCPPortaAcessivel(host, porta string, timeout time.Duration) (bool, string) {
	host = strings.TrimSpace(host)
	porta = strings.TrimSpace(porta)
	if host == "" || porta == "" {
		return false, ""
	}
	addr := net.JoinHostPort(host, porta)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	cmd := fmt.Sprintf("Test-NetConnection / TCP dial %s", addr)
	if err != nil {
		return false, cmd
	}
	_ = conn.Close()
	return true, cmd
}

func PostPresencaInterno(p ReqPresenca) error {
	base := strings.TrimRight(strings.TrimSpace(config.ApiFunctionInternalURL), "/")
	if base == "" {
		base = "http://127.0.0.1:" + config.Porta
	}
	body, _ := json.Marshal(p)
	req, err := http.NewRequest(http.MethodPost, base+"/receptor/diag/presenca", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	key := strings.TrimSpace(config.WorkerSecret)
	if key != "" {
		req.Header.Set("X-Worker-Key", key)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("presenca HTTP %d: %s", res.StatusCode, string(raw))
	}
	return nil
}

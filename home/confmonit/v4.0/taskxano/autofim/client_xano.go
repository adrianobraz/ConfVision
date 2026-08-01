package autofim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type pendingResponse struct {
	Dados []pendingRow `json:"dados"`
}

type pendingRow struct {
	ID            int    `json:"id"`
	IDProcesso    string `json:"idProcesso"`
	IDDispositivo string `json:"idDispositivo"`
	Status        string `json:"status"`
	Tentativas    int    `json:"tentativas"`
	AlarmEventsID int    `json:"alarm_events_id"`
}

type claimResponse struct {
	Dados struct {
		Claimed bool       `json:"claimed"`
		Row     pendingRow `json:"row"`
	} `json:"dados"`
}

type contextResponse struct {
	Dados struct {
		Evento struct {
			ID            int    `json:"id"`
			IDProcesso    string `json:"idProcesso"`
			IDDispositivo string `json:"idDispositivo"`
		} `json:"evento"`
		EventosProcesso      []contextEvent `json:"eventosProcesso"`
		HistAgg              histAgg        `json:"histAgg"`
		ProcessoJaFinalizado bool           `json:"processoJaFinalizado"`
		DataAtenFim          string         `json:"dataAtenFim"`
	} `json:"dados"`
}

type contextEvent struct {
	CtiGrupo     string `json:"ctiGrupo"`
	CtiDescricao string `json:"ctiDescricao"`
	Particao     string `json:"particao"`
	ZonaUser     string `json:"zonaUser"`
	CreatedAt    any    `json:"created_at"`
}

type histAgg struct {
	QtdCiclos3mDiffProc      int    `json:"qtd_ciclos_3m_diff_proc"`
	QtdCiclos3mDiffProcCamel int    `json:"qtdCiclos3mDiffProc"`
	PerfilLocal              string `json:"perfil_local"`
	PerfilLocalCamel         string `json:"perfilLocal"`
}

type processEndResponse struct {
	Dados struct {
		Acao               string         `json:"acao"`
		RetornoProcessoEnd map[string]any `json:"retornoProcessoEnd"`
	} `json:"dados"`
}

func getPending(client *http.Client, rawURL string, limit int) ([]pendingRow, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	var out pendingResponse
	if err := doJSONRequest(client, http.MethodGet, u.String(), nil, &out); err != nil {
		return nil, err
	}
	return out.Dados, nil
}

func claim(client *http.Client, rawURL string, id int, lockToken string) (claimResponse, error) {
	var out claimResponse
	err := doJSONRequest(client, http.MethodPost, rawURL, map[string]any{
		"id":         id,
		"lock_token": lockToken,
	}, &out)
	return out, err
}

func getContext(client *http.Client, rawURL string, idEvento int) (contextResponse, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return contextResponse{}, err
	}
	q := u.Query()
	q.Set("idEvento", strconv.Itoa(idEvento))
	u.RawQuery = q.Encode()

	var out contextResponse
	err = doJSONRequest(client, http.MethodGet, u.String(), nil, &out)
	return out, err
}

func callProcessEnd(client *http.Client, rawURL, idProcesso, motivo, perfilLocal string) (processEndResponse, error) {
	var out processEndResponse
	err := doJSONRequest(client, http.MethodPost, rawURL, map[string]any{
		"idProcesso":  idProcesso,
		"motivo":      motivo,
		"perfilLocal": perfilLocal,
	}, &out)
	return out, err
}

func sendLog(client *http.Client, rawURL string, payload map[string]any) error {
	var out map[string]any
	return doJSONRequest(client, http.MethodPost, rawURL, payload, &out)
}

func markSuccess(client *http.Client, rawURL string, id int, lockToken, motivo, acao string, payloadResultado map[string]any) error {
	var out map[string]any
	return doJSONRequest(client, http.MethodPost, rawURL, map[string]any{
		"id":                id,
		"lock_token":        lockToken,
		"motivo_final":      motivo,
		"acao_final":        acao,
		"payload_resultado": payloadResultado,
	}, &out)
}

func markFail(client *http.Client, rawURL string, id int, lockToken, erro string, retrySeconds int) error {
	var out map[string]any
	return doJSONRequest(client, http.MethodPost, rawURL, map[string]any{
		"id":                  id,
		"lock_token":          lockToken,
		"erro":                erro,
		"retry_delay_seconds": retrySeconds,
	}, &out)
}

func doJSONRequest(client *http.Client, method, rawURL string, payload any, out any) error {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d body=%s", resp.StatusCode, compactBody(respBody))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("json inválido: %w", err)
	}
	return nil
}

func compactBody(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 220 {
		return s[:220] + "..."
	}
	return s
}

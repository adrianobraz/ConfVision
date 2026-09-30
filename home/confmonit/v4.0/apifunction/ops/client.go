package ops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"apifunction/pgatendimento"
)

func useLocal() bool {
	return pgatendimento.Configurado()
}

func baseURL() string {
	u := strings.TrimSpace(os.Getenv("CONFVISION_OPS_URL"))
	if u == "" {
		u = "http://127.0.0.1:8086"
	}
	return strings.TrimRight(u, "/")
}

func post(path string, payload map[string]any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL()+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return fmt.Errorf("ops HTTP %d: %s", res.StatusCode, string(raw))
	}
	return nil
}

func put(path string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, baseURL()+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return fmt.Errorf("ops HTTP %d: %s", res.StatusCode, string(raw))
	}
	return nil
}

func Creditar(idFranqueado, servico, tipo string, valor float64, idFatura int64, obs, criadoPor string) error {
	if useLocal() {
		var idFat *int64
		if idFatura > 0 {
			idFat = &idFatura
		}
		return pgatendimento.Creditar(context.Background(), idFranqueado, servico, tipo, valor, idFat, obs, criadoPor)
	}
	p := map[string]any{
		"id_franqueado": idFranqueado,
		"servico":       servico,
		"tipo":          tipo,
		"valor":         valor,
		"observacao":    obs,
		"criado_por":    criadoPor,
	}
	if idFatura > 0 {
		p["id_fatura"] = idFatura
	}
	return post("/ops/atendimento/credito/creditar", p)
}

func SyncTarifa(idCentral, canal string, tentativa, minuto, unidade float64) error {
	if useLocal() {
		return pgatendimento.SaveTarifa(context.Background(), idCentral, canal, tentativa, minuto, unidade)
	}
	return put("/ops/atendimento/tarifa", map[string]any{
		"id_central":      idCentral,
		"canal":           canal,
		"valor_tentativa": tentativa,
		"valor_minuto":    minuto,
		"valor_unidade":   unidade,
	})
}

func InicializarSaldoZero(idFranqueado string) error {
	if useLocal() {
		return pgatendimento.InicializarSaldoZero(context.Background(), idFranqueado)
	}
	return post("/ops/atendimento/credito/inicializar", map[string]any{
		"id_franqueado": idFranqueado,
	})
}

type DebitarResult struct {
	JaDebitado bool `json:"ja_debitado"`
}

func Debitar(idFranqueado, servico string, valor float64, idProcesso, obs string) (*DebitarResult, error) {
	if useLocal() {
		res, err := pgatendimento.Debitar(context.Background(), idFranqueado, servico, valor, idProcesso, obs)
		if err != nil {
			return nil, err
		}
		return &DebitarResult{JaDebitado: res.JaDebitado}, nil
	}
	p := map[string]any{
		"id_franqueado": idFranqueado,
		"servico":       servico,
		"valor":         valor,
		"id_processo":   idProcesso,
		"observacao":    obs,
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL()+"/ops/atendimento/credito/debitar", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("ops HTTP %d: %s", res.StatusCode, string(raw))
	}
	out := &DebitarResult{}
	var wrap map[string]any
	if err := json.Unmarshal(raw, &wrap); err == nil {
		if v, ok := wrap["ja_debitado"].(bool); ok {
			out.JaDebitado = v
		}
		if d, ok := wrap["dados"].(map[string]any); ok {
			if v, ok := d["ja_debitado"].(bool); ok {
				out.JaDebitado = v
			}
		}
	}
	return out, nil
}

package confservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"apifunction/config"
)

type RepasseItem struct {
	IDVinculo   string  `json:"id_vinculo"`
	IDParceiro  string  `json:"id_parceiro"`
	ValorPiso   float64 `json:"valor_piso"`
	FaturaID    int     `json:"fp_fatura_id"`
	RefTipo     string  `json:"ref_tipo"`
	RefID       string  `json:"ref_id"`
}

type RepasseResumo struct {
	ID          string  `json:"id"`
	IDParceiro  string  `json:"id_parceiro"`
	NomeParceiro string `json:"nome_parceiro,omitempty"`
	FPFaturaID  int     `json:"fp_fatura_id"`
	Valor       float64 `json:"valor"`
	Status      string  `json:"status"`
	RefTipo     string  `json:"ref_tipo,omitempty"`
	RefID       string  `json:"ref_id,omitempty"`
}

func ListRepassesFatura(idFranqueado string, fpFaturaID int) ([]RepasseResumo, error) {
	base := strings.TrimRight(strings.TrimSpace(config.ConfServiceURL), "/")
	if base == "" {
		return nil, nil
	}
	q := urlQuery(idFranqueado, fpFaturaID)
	req, err := http.NewRequest(http.MethodGet, base+"/internal/repasse/listar?"+q, nil)
	if err != nil {
		return nil, err
	}
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	res, err := (&http.Client{Timeout: 12 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("confservice HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed struct {
		Dados []RepasseResumo `json:"dados"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return parsed.Dados, nil
}

func urlQuery(idFranqueado string, fpFaturaID int) string {
	v := url.Values{}
	if idFranqueado != "" {
		v.Set("id_franqueado", idFranqueado)
	}
	if fpFaturaID > 0 {
		v.Set("fp_fatura_id", fmt.Sprintf("%d", fpFaturaID))
	}
	return v.Encode()
}

func GerarRepassesParceiro(idFranqueado string, faturaID int, itens []RepasseItem) error {
	base := config.ConfServiceURL
	if base == "" {
		return nil
	}
	type itemOut struct {
		IDVinculo  string  `json:"id_vinculo"`
		IDParceiro string  `json:"id_parceiro"`
		ValorPiso  float64 `json:"valor_piso"`
		RefTipo    string  `json:"ref_tipo"`
		RefID      string  `json:"ref_id"`
	}
	outItens := make([]itemOut, 0, len(itens))
	for _, it := range itens {
		outItens = append(outItens, itemOut{
			IDVinculo:  it.IDVinculo,
			IDParceiro: it.IDParceiro,
			ValorPiso:  it.ValorPiso,
			RefTipo:    it.RefTipo,
			RefID:      it.RefID,
		})
	}
	payload := map[string]any{
		"id_franqueado": idFranqueado,
		"fp_fatura_id":  faturaID,
		"itens":         outItens,
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, base+"/internal/repasse/gerar", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return fmt.Errorf("confservice HTTP %d: %s", res.StatusCode, string(raw))
	}
	return nil
}

func CancelarRepassesFatura(idFranqueado string, fpFaturaID int) (int, error) {
	base := strings.TrimRight(strings.TrimSpace(config.ConfServiceURL), "/")
	if base == "" || fpFaturaID <= 0 {
		return 0, nil
	}
	payload, _ := json.Marshal(map[string]any{
		"id_franqueado": idFranqueado,
		"fp_fatura_id":  fpFaturaID,
	})
	req, err := http.NewRequest(http.MethodPost, base+"/internal/repasse/cancelar", bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	res, err := (&http.Client{Timeout: 12 * time.Second}).Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return 0, fmt.Errorf("confservice HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed struct {
		Afetados int `json:"afetados"`
	}
	_ = json.Unmarshal(raw, &parsed)
	return parsed.Afetados, nil
}

func DesativarVinculoPorID(idFranqueado, idCliente, idVinculo string) error {
	base := strings.TrimRight(strings.TrimSpace(config.ConfServiceURL), "/")
	if base == "" || strings.TrimSpace(idVinculo) == "" {
		return nil
	}
	payload, _ := json.Marshal(map[string]any{
		"idFranqueado": idFranqueado,
		"idCliente":    idCliente,
		"idVinculo":    idVinculo,
	})
	req, err := http.NewRequest(http.MethodPost, base+"/internal/vinculo/desativar-id", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	res, err := (&http.Client{Timeout: 12 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("confservice HTTP %d: %s", res.StatusCode, string(raw))
	}
	return nil
}

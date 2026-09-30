package xano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL   string
	workerKey string
	http      *http.Client
}

func New(baseURL, workerKey string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil
	}
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		workerKey: strings.TrimSpace(workerKey),
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != ""
}

func (c *Client) Post(path string, payload any) (map[string]any, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url := c.baseURL + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("xano HTTP %d: %s", res.StatusCode, string(raw))
	}
	var out map[string]any
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse xano: %w", err)
	}
	return out, nil
}

type ItemSync struct {
	Descricao     string  `json:"descricao"`
	Quantidade    int     `json:"quantidade"`
	ValorUnitario float64 `json:"valor_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	RefChave      string  `json:"ref_chave,omitempty"`
}

type SyncFaturaInput struct {
	IDContratoMySQL int
	IDFaturaMySQL   int
	IDFranqueado    string
	IDCentral       string
	IDRepresentante string
	TipoFatura      string
	ValorTotal      float64
	VencimentoEm    string
	CicloRef        string
	NomeContrato    string
	AdminUsuario    string
	Itens           []ItemSync
}

func (c *Client) SyncFaturaContrato(in SyncFaturaInput) (int, error) {
	payload := map[string]any{
		"worker_key":         c.workerKey,
		"id_contrato_mysql":  in.IDContratoMySQL,
		"id_fatura_mysql":    in.IDFaturaMySQL,
		"id_franqueado":      in.IDFranqueado,
		"id_central":         in.IDCentral,
		"id_representante":   in.IDRepresentante,
		"tipo_fatura":        in.TipoFatura,
		"valor_total":        in.ValorTotal,
		"vencimento_em":      in.VencimentoEm,
		"ciclo_ref":          in.CicloRef,
		"nome_contrato":      in.NomeContrato,
		"admin_usuario":      in.AdminUsuario,
	}
	if len(in.Itens) > 0 {
		items := make([]map[string]any, 0, len(in.Itens))
		for _, it := range in.Itens {
			items = append(items, map[string]any{
				"descricao":      it.Descricao,
				"quantidade":     it.Quantidade,
				"valor_unitario": it.ValorUnitario,
				"valor_total":    it.ValorTotal,
				"ref_chave":      it.RefChave,
			})
		}
		payload["itens"] = items
	}
	out, err := c.Post("/fp_fatura_contrato_sync", payload)
	if err != nil {
		return 0, err
	}
	return intFromAny(out["id_fatura_contabil"]), nil
}

func (c *Client) CancelarFaturaContrato(idContrato, idFatura, idFaturaContabil int, obs, adminUsuario string) error {
	payload := map[string]any{
		"worker_key": c.workerKey,
		"observacao": obs,
	}
	if idContrato > 0 {
		payload["id_contrato_mysql"] = idContrato
	}
	if idFatura > 0 {
		payload["id_fatura_mysql"] = idFatura
	}
	if idFaturaContabil > 0 {
		payload["id_fatura_contabil"] = idFaturaContabil
	}
	_, err := c.Post("/fp_fatura_contrato_cancelar", payload)
	return err
}

func (c *Client) PagarFaturaContrato(idFaturaContabil, idContrato, idFatura int, valor float64, adminUsuario string) error {
	payload := map[string]any{
		"worker_key":    c.workerKey,
		"metodo":        "manual",
		"admin_usuario": adminUsuario,
	}
	if idFaturaContabil > 0 {
		payload["id_fatura_contabil"] = idFaturaContabil
	}
	if idContrato > 0 {
		payload["id_contrato_mysql"] = idContrato
	}
	if idFatura > 0 {
		payload["id_fatura_mysql"] = idFatura
	}
	if valor > 0 {
		payload["valor"] = valor
	}
	_, err := c.Post("/fp_fatura_contrato_pagar", payload)
	return err
}

func (c *Client) EstornarFaturaContabil(idFaturaContabil int, motivo, adminUsuario string) error {
	payload := map[string]any{
		"worker_key":    c.workerKey,
		"fatura_id":     idFaturaContabil,
		"motivo":        motivo,
		"admin_usuario": adminUsuario,
	}
	_, err := c.Post("/fp_fatura_estornar_pagamento", payload)
	return err
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

type OrdemItemSync struct {
	Descricao     string  `json:"descricao"`
	Quantidade    int     `json:"quantidade"`
	ValorUnitario float64 `json:"valor_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	RefTipo       string  `json:"ref_tipo"`
	RefID         string  `json:"ref_id"`
	ValorPiso     float64 `json:"valor_piso"`
	MargemCentral float64 `json:"margem_central"`
	MargemRep     float64 `json:"margem_rep"`
}

type SyncOrdemInput struct {
	IDFranqueado    string
	IDCentral       string
	IDRepresentante string
	CicloRef        string
	VencimentoEm    string
	ValorTotal      float64
	Itens           []OrdemItemSync
}

func (c *Client) SyncFaturaOrdem(in SyncOrdemInput) (int, error) {
	items := make([]map[string]any, 0, len(in.Itens))
	for _, it := range in.Itens {
		items = append(items, map[string]any{
			"descricao":      it.Descricao,
			"quantidade":     it.Quantidade,
			"valor_unitario": it.ValorUnitario,
			"valor_total":    it.ValorTotal,
			"ref_tipo":       it.RefTipo,
			"ref_id":         it.RefID,
			"valor_piso":     it.ValorPiso,
			"margem_central": it.MargemCentral,
			"margem_rep":     it.MargemRep,
		})
	}
	payload := map[string]any{
		"worker_key":       c.workerKey,
		"id_franqueado":    in.IDFranqueado,
		"id_central":       in.IDCentral,
		"id_representante": in.IDRepresentante,
		"ciclo_ref":        in.CicloRef,
		"vencimento_em":    in.VencimentoEm,
		"valor_total":      in.ValorTotal,
		"itens":            items,
	}
	out, err := c.Post("/fp_fatura_ordem_sync", payload)
	if err != nil {
		return 0, err
	}
	if id, ok := out["id_fatura_contabil"].(float64); ok {
		return int(id), nil
	}
	if fatura, ok := out["fatura"].(map[string]any); ok {
		return intFromAny(fatura["id"]), nil
	}
	return intFromAny(out["id_fatura_contabil"]), nil
}

type RemoverOrdemItem struct {
	RefTipo string `json:"ref_tipo"`
	RefID   string `json:"ref_id"`
}

func (c *Client) RemoverItensOrdemConsolidada(idFranqueado string, itens []RemoverOrdemItem) (map[string]any, error) {
	if len(itens) == 0 {
		return map[string]any{}, nil
	}
	rows := make([]map[string]any, 0, len(itens))
	for _, it := range itens {
		if strings.TrimSpace(it.RefTipo) == "" || strings.TrimSpace(it.RefID) == "" {
			continue
		}
		rows = append(rows, map[string]any{
			"ref_tipo": strings.TrimSpace(it.RefTipo),
			"ref_id":   strings.TrimSpace(it.RefID),
		})
	}
	if len(rows) == 0 {
		return map[string]any{}, nil
	}
	payload := map[string]any{
		"worker_key":    c.workerKey,
		"id_franqueado": strings.TrimSpace(idFranqueado),
		"remover_itens": rows,
		"origem":        "apifunction",
	}
	return c.Post("/fp_fatura_ordem_sync", payload)
}

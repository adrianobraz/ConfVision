package xano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"webAmbiente/src/config"
	"webAmbiente/src/seguranca"
)

// CoClienteConfig — tabela Xano co_cliente_config (grupo centerOperacion)
type CoClienteConfig struct {
	Id           int    `json:"id"`
	IdCliente    string `json:"idCliente"`
	IdFranqueado string `json:"idFranqueado"`
	CorAvatar    string `json:"corAvatar"`
	Iniciais     string `json:"iniciais"`
}

func coCenterOperacionBase() string {
	return strings.TrimRight(config.XanoCenterOperacion, "/")
}

func reqCenterOperacionJSON(path string, payload any) ([]byte, error) {
	if coCenterOperacionBase() == "" {
		return nil, fmt.Errorf("XANO_CENTER_OPERACION_API nao configurado")
	}
	resp, err := seguranca.ReqXanoParaBase(coCenterOperacionBase(), path, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("centerOperacion %s: %s", path, string(raw))
	}
	return raw, nil
}

// BuscarCoClienteConfig retorna config visual do cliente (ou vazio se nao existir).
func BuscarCoClienteConfig(idCliente, idFranqueado string) (CoClienteConfig, error) {
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return CoClienteConfig{}, fmt.Errorf("idCliente obrigatorio")
	}
	raw, err := reqCenterOperacionJSON("/co_cliente_config_get", map[string]string{
		"idCliente":    idCliente,
		"idFranqueado": strings.TrimSpace(idFranqueado),
	})
	if err != nil {
		return CoClienteConfig{}, err
	}
	var cfg CoClienteConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return CoClienteConfig{}, err
	}
	return cfg, nil
}

// SalvarCoClienteConfig grava cor/iniciais do avatar.
func SalvarCoClienteConfig(idCliente, idFranqueado, corAvatar, iniciais string) (CoClienteConfig, error) {
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return CoClienteConfig{}, fmt.Errorf("idCliente obrigatorio")
	}
	raw, err := reqCenterOperacionJSON("/co_cliente_config_salvar", map[string]string{
		"idCliente":    idCliente,
		"idFranqueado": strings.TrimSpace(idFranqueado),
		"corAvatar":    strings.TrimSpace(corAvatar),
		"iniciais":     strings.TrimSpace(iniciais),
	})
	if err != nil {
		return CoClienteConfig{}, err
	}
	var cfg CoClienteConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return CoClienteConfig{}, err
	}
	return cfg, nil
}

// CoTimelineEvento — linha do tempo da ocorrência (caixa-preta)
type CoTimelineEvento struct {
	Id             int    `json:"id"`
	Chave          string `json:"chave"`
	MapaAmbienteId int    `json:"mapa_ambiente_id"`
	IdCliente      string `json:"idCliente"`
	IdFranqueado   string `json:"idFranqueado"`
	IdProcesso     string `json:"idProcesso"`
	IdSetor        string `json:"idSetor"`
	IdDispositivo  string `json:"idDispositivo"`
	IdOperador     string `json:"idOperador"`
	NomeOperador   string `json:"nomeOperador"`
	Tipo           string `json:"tipo"`
	Origem         string `json:"origem"`
	Codigo         string `json:"codigo"`
	ZonaUser       string `json:"zonaUser"`
	Particao       string `json:"particao"`
	Texto          string `json:"texto"`
	EventoTs       string `json:"evento_ts"`
}

// CoRastroDisparo — ponto do rastro na planta
type CoRastroDisparo struct {
	Id             int     `json:"id"`
	Chave          string  `json:"chave"`
	MapaAmbienteId int     `json:"mapa_ambiente_id"`
	IdCliente      string  `json:"idCliente"`
	IdFranqueado   string  `json:"idFranqueado"`
	IdProcesso     string  `json:"idProcesso"`
	IdSetor        string  `json:"idSetor"`
	LabelSetor     string  `json:"labelSetor"`
	PosX           float64 `json:"posX"`
	PosY           float64 `json:"posY"`
	Sequencia      int     `json:"sequencia"`
	PreditoIdSetor string  `json:"predito_id_setor"`
	Direcao        string  `json:"direcao"`
	EventoTs       string  `json:"evento_ts"`
}

func parseJSONArray(raw []byte) ([]json.RawMessage, error) {
	if len(raw) == 0 || string(bytes.TrimSpace(raw)) == "null" {
		return []json.RawMessage{}, nil
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var wrap struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(raw, &wrap); err == nil && wrap.Items != nil {
		return wrap.Items, nil
	}
	return nil, fmt.Errorf("resposta nao e array json")
}

func ListarCoTimeline(mapaId int, idProcesso, idCliente string, limite int) ([]CoTimelineEvento, error) {
	payload := map[string]interface{}{
		"mapa_ambiente_id": mapaId,
		"idProcesso":       strings.TrimSpace(idProcesso),
		"idCliente":        strings.TrimSpace(idCliente),
		"limite":           limite,
	}
	raw, err := reqCenterOperacionJSON("/co_timeline_listar", payload)
	if err != nil {
		return nil, err
	}
	parts, err := parseJSONArray(raw)
	if err != nil {
		return nil, err
	}
	out := make([]CoTimelineEvento, 0, len(parts))
	for _, p := range parts {
		var item CoTimelineEvento
		if err := json.Unmarshal(p, &item); err != nil {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func SalvarCoTimeline(eventos []CoTimelineEvento) (map[string]interface{}, error) {
	if len(eventos) == 0 {
		return map[string]interface{}{"status": "OK", "inseridos": 0}, nil
	}
	raw, err := reqCenterOperacionJSON("/co_timeline_salvar", map[string]interface{}{
		"eventos": eventos,
	})
	if err != nil {
		return nil, err
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func ListarCoRastro(mapaId int, idProcesso, idCliente string, limite int) ([]CoRastroDisparo, error) {
	payload := map[string]interface{}{
		"mapa_ambiente_id": mapaId,
		"idProcesso":       strings.TrimSpace(idProcesso),
		"idCliente":        strings.TrimSpace(idCliente),
		"limite":           limite,
	}
	raw, err := reqCenterOperacionJSON("/co_rastro_listar", payload)
	if err != nil {
		return nil, err
	}
	parts, err := parseJSONArray(raw)
	if err != nil {
		return nil, err
	}
	out := make([]CoRastroDisparo, 0, len(parts))
	for _, p := range parts {
		var item CoRastroDisparo
		if err := json.Unmarshal(p, &item); err != nil {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func SalvarCoRastro(pontos []CoRastroDisparo) (map[string]interface{}, error) {
	if len(pontos) == 0 {
		return map[string]interface{}{"status": "OK", "inseridos": 0}, nil
	}
	raw, err := reqCenterOperacionJSON("/co_rastro_salvar", map[string]interface{}{
		"pontos": pontos,
	})
	if err != nil {
		return nil, err
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

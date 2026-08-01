package xano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"webAmbiente/src/config"
	"webAmbiente/src/storage"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
)

// APIs Xano — grupo mapaAmbiente (base: https://xpcy-oyme-lno7.b2.xano.io/api:XUUyTjyq)
//
// mapa_ambiente:
//   POST   /mapa_ambiente_query_all   — listar mapas (idFranqueado opcional)
//   GET    /mapa_ambiente
//   POST   /mapa_ambiente
//   GET    /mapa_ambiente/{mapa_ambiente_id}
//   PUT    /mapa_ambiente/{mapa_ambiente_id}
//   DELETE /mapa_ambiente/{mapa_ambiente_id}
//
// mapa_setor:
//   GET    /mapa_setor
//   POST   /mapa_setor
//   GET    /mapa_setor/{mapa_setor_id}
//   PUT    /mapa_setor/{mapa_setor_id}
//   DELETE /mapa_setor/{mapa_setor_id}

// QueryAllMapaAmbiente chama POST /mapa_ambiente_query_all no Xano.
// Com idFranqueado retorna mapas daquele franqueado; vazio retorna todos.
func QueryAllMapaAmbiente(idFranqueado string) ([]tipos.MapaAmbiente, error) {
	if config.Xano == "" {
		return nil, fmt.Errorf("XANO_API nao configurado")
	}

	idFranqueado = strings.TrimSpace(idFranqueado)
	payload := map[string]string{}
	if idFranqueado != "" {
		payload["idFranqueado"] = idFranqueado
	}

	raw, err := postXano("/mapa_ambiente_query_all", payload)
	if err != nil {
		// fallback enquanto endpoint nao estiver publicado no Xano
		return fetchMapaAmbienteLegacy(idFranqueado)
	}

	lista, err := parseMapaAmbienteLista(raw)
	if err != nil {
		return nil, fmt.Errorf("parse mapa_ambiente_query_all: %w", err)
	}
	return normalizarMapas(lista), nil
}

func ListarMapaAmbiente(idFranqueado, idCliente string) ([]tipos.MapaAmbiente, error) {
	todos, err := QueryAllMapaAmbiente(idFranqueado)
	if err != nil {
		return nil, err
	}

	idFranqueado = normID(idFranqueado)
	idCliente = normID(idCliente)

	out := make([]tipos.MapaAmbiente, 0, len(todos))
	for _, m := range todos {
		if idFranqueado != "" && normID(m.IdFranqueado) != idFranqueado {
			continue
		}
		if idCliente != "" && normID(m.IdCliente) != idCliente {
			continue
		}
		if m.Ativo != nil && !*m.Ativo {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func GetMapaAmbiente(id int) (tipos.MapaAmbiente, error) {
	if config.Xano == "" {
		return tipos.MapaAmbiente{}, fmt.Errorf("XANO_API nao configurado")
	}

	raw, err := getXano("/mapa_ambiente/" + strconv.Itoa(id))
	if err != nil {
		return tipos.MapaAmbiente{}, err
	}

	var m tipos.MapaAmbiente
	if err := json.Unmarshal(raw, &m); err != nil {
		return tipos.MapaAmbiente{}, err
	}
	m.ImagemUrl = storage.NormalizeImagemURL(m.ImagemUrl)
	return m, nil
}

func CarregarMapaComSetores(mapaAmbienteId int) (tipos.MapaAmbiente, []tipos.MapaSetor, error) {
	mapa, err := GetMapaAmbiente(mapaAmbienteId)
	if err != nil {
		return tipos.MapaAmbiente{}, nil, err
	}

	setores, err := ListarSetoresPorMapa(mapaAmbienteId)
	if err != nil {
		return mapa, nil, err
	}
	return mapa, setores, nil
}

func ListarSetoresPorMapa(mapaAmbienteId int) ([]tipos.MapaSetor, error) {
	todos, err := fetchMapaSetorAll()
	if err != nil {
		return nil, err
	}

	out := make([]tipos.MapaSetor, 0)
	for _, s := range todos {
		if s.MapaAmbienteId == mapaAmbienteId {
			out = append(out, s)
		}
	}
	return out, nil
}

// ListarTodosSetores retorna todos os mapa_setor do Xano em uma unica chamada.
func ListarTodosSetores() ([]tipos.MapaSetor, error) {
	return fetchMapaSetorAll()
}

func GetMapaSetor(id int) (tipos.MapaSetor, error) {
	if config.Xano == "" {
		return tipos.MapaSetor{}, fmt.Errorf("XANO_API nao configurado")
	}

	raw, err := getXano("/mapa_setor/" + strconv.Itoa(id))
	if err != nil {
		return tipos.MapaSetor{}, err
	}

	var s tipos.MapaSetor
	if err := json.Unmarshal(raw, &s); err != nil {
		return tipos.MapaSetor{}, err
	}
	return s, nil
}

func CriarMapaAmbiente(m tipos.MapaAmbiente) (tipos.MapaAmbiente, error) {
	if config.Xano == "" {
		return tipos.MapaAmbiente{}, fmt.Errorf("XANO_API nao configurado")
	}

	res, err := seguranca.ReqXanoJSON("/mapa_ambiente", m)
	if err != nil {
		return tipos.MapaAmbiente{}, err
	}
	defer res.Body.Close()

	raw, err := lerCorpoResposta(res)
	if err != nil {
		return tipos.MapaAmbiente{}, err
	}

	var out tipos.MapaAmbiente
	if err := json.Unmarshal(raw, &out); err != nil {
		return tipos.MapaAmbiente{}, err
	}
	return out, nil
}

func AtualizarMapaAmbiente(id int, m tipos.MapaAmbiente) (tipos.MapaAmbiente, error) {
	if config.Xano == "" {
		return tipos.MapaAmbiente{}, fmt.Errorf("XANO_API nao configurado")
	}

	res, err := seguranca.ReqXanoPUT("/mapa_ambiente/"+strconv.Itoa(id), m)
	if err != nil {
		return tipos.MapaAmbiente{}, err
	}
	defer res.Body.Close()

	raw, err := lerCorpoResposta(res)
	if err != nil {
		return tipos.MapaAmbiente{}, err
	}

	var out tipos.MapaAmbiente
	if err := json.Unmarshal(raw, &out); err != nil {
		return tipos.MapaAmbiente{}, err
	}
	return out, nil
}

func ExcluirMapaAmbiente(id int) error {
	if config.Xano == "" {
		return fmt.Errorf("XANO_API nao configurado")
	}

	res, err := seguranca.ReqXano(http.MethodDelete, "/mapa_ambiente/"+strconv.Itoa(id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = lerCorpoResposta(res)
	return nil
}

func CriarMapaSetor(s tipos.MapaSetor) (tipos.MapaSetor, error) {
	if config.Xano == "" {
		return tipos.MapaSetor{}, fmt.Errorf("XANO_API nao configurado")
	}

	res, err := seguranca.ReqXanoJSON("/mapa_setor", s)
	if err != nil {
		return tipos.MapaSetor{}, err
	}
	defer res.Body.Close()

	raw, err := lerCorpoResposta(res)
	if err != nil {
		return tipos.MapaSetor{}, err
	}

	var out tipos.MapaSetor
	if err := json.Unmarshal(raw, &out); err != nil {
		return tipos.MapaSetor{}, err
	}
	return out, nil
}

func AtualizarMapaSetor(id int, s tipos.MapaSetor) (tipos.MapaSetor, error) {
	if config.Xano == "" {
		return tipos.MapaSetor{}, fmt.Errorf("XANO_API nao configurado")
	}

	res, err := seguranca.ReqXanoPUT("/mapa_setor/"+strconv.Itoa(id), s)
	if err != nil {
		return tipos.MapaSetor{}, err
	}
	defer res.Body.Close()

	raw, err := lerCorpoResposta(res)
	if err != nil {
		return tipos.MapaSetor{}, err
	}

	var out tipos.MapaSetor
	if err := json.Unmarshal(raw, &out); err != nil {
		return tipos.MapaSetor{}, err
	}
	return out, nil
}

func ExcluirMapaSetor(id int) error {
	if config.Xano == "" {
		return fmt.Errorf("XANO_API nao configurado")
	}

	res, err := seguranca.ReqXano(http.MethodDelete, "/mapa_setor/"+strconv.Itoa(id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = lerCorpoResposta(res)
	return nil
}

// StatusMapa delegado ao pacote monitor (MySQL processo + keep alive).
func StatusMapa(mapaAmbienteId int) ([]map[string]string, error) {
	_ = mapaAmbienteId
	return []map[string]string{}, nil
}

func fetchMapaAmbienteLegacy(idFranqueado string) ([]tipos.MapaAmbiente, error) {
	raw, err := getXano("/mapa_ambiente")
	if err != nil {
		return nil, err
	}

	lista, err := parseMapaAmbienteLista(raw)
	if err != nil {
		return nil, err
	}
	lista = normalizarMapas(lista)

	idFranqueado = normID(idFranqueado)
	if idFranqueado == "" {
		return lista, nil
	}

	out := make([]tipos.MapaAmbiente, 0)
	for _, m := range lista {
		if normID(m.IdFranqueado) == idFranqueado {
			out = append(out, m)
		}
	}
	return out, nil
}

func fetchMapaSetorAll() ([]tipos.MapaSetor, error) {
	if config.Xano == "" {
		return nil, fmt.Errorf("XANO_API nao configurado")
	}

	raw, err := getXano("/mapa_setor")
	if err != nil {
		return nil, err
	}
	return parseMapaSetorLista(raw)
}

func getXano(path string) ([]byte, error) {
	res, err := seguranca.ReqXanoGET(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return lerCorpoResposta(res)
}

func postXano(path string, payload any) ([]byte, error) {
	res, err := seguranca.ReqXanoJSON(path, payload)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return lerCorpoResposta(res)
}

func lerCorpoResposta(res *http.Response) ([]byte, error) {
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = res.Status
		}
		return nil, fmt.Errorf("xano HTTP %d: %s", res.StatusCode, msg)
	}
	return raw, nil
}

func parseMapaAmbienteLista(raw []byte) ([]tipos.MapaAmbiente, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return []tipos.MapaAmbiente{}, nil
	}

	var direct []tipos.MapaAmbiente
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, nil
	}

	var wrapped struct {
		Dados   []tipos.MapaAmbiente `json:"dados"`
		Items   []tipos.MapaAmbiente `json:"items"`
		Records []tipos.MapaAmbiente `json:"records"`
		Data    []tipos.MapaAmbiente `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return parseMapaAmbienteFlex(raw)
	}
	if len(wrapped.Dados) > 0 {
		return wrapped.Dados, nil
	}
	if len(wrapped.Items) > 0 {
		return wrapped.Items, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}
	if len(wrapped.Records) > 0 {
		return wrapped.Records, nil
	}
	return []tipos.MapaAmbiente{}, nil
}

func parseMapaAmbienteFlex(raw []byte) ([]tipos.MapaAmbiente, error) {
	var arr []map[string]interface{}
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("json mapa_ambiente: %w", err)
	}

	out := make([]tipos.MapaAmbiente, 0, len(arr))
	for _, row := range arr {
		m := tipos.MapaAmbiente{
			Id:             intVal(row, "id"),
			IdCliente:      strVal(row, "idCliente", "id_cliente"),
			IdFranqueado:   strVal(row, "idFranqueado", "id_franqueado"),
			NomeCliente:    strVal(row, "nomeCliente", "nome_cliente"),
			NomeFranqueado: strVal(row, "nomeFranqueado", "nome_franqueado"),
			Descricao:      strVal(row, "descricao"),
			ImagemUrl:      strVal(row, "imagem_url", "imagemUrl"),
			Ordem:          intVal(row, "ordem"),
		}
		if v, ok := row["ativo"]; ok && v != nil {
			b := boolVal(v)
			m.Ativo = &b
		}
		out = append(out, m)
	}
	return out, nil
}

func parseMapaSetorLista(raw []byte) ([]tipos.MapaSetor, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return []tipos.MapaSetor{}, nil
	}

	var direct []tipos.MapaSetor
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, nil
	}

	var wrapped struct {
		Dados   []tipos.MapaSetor `json:"dados"`
		Items   []tipos.MapaSetor `json:"items"`
		Records []tipos.MapaSetor `json:"records"`
		Data    []tipos.MapaSetor `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	if len(wrapped.Dados) > 0 {
		return wrapped.Dados, nil
	}
	if len(wrapped.Items) > 0 {
		return wrapped.Items, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}
	return wrapped.Records, nil
}

func normalizarMapas(lista []tipos.MapaAmbiente) []tipos.MapaAmbiente {
	for i := range lista {
		lista[i].IdCliente = normID(lista[i].IdCliente)
		lista[i].IdFranqueado = normID(lista[i].IdFranqueado)
		lista[i].ImagemUrl = storage.NormalizeImagemURL(lista[i].ImagemUrl)
	}
	return lista
}

func normID(s string) string {
	return strings.TrimSpace(s)
}

func strVal(row map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := row[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				return strings.TrimSpace(t)
			case float64:
				return strings.TrimSpace(strconv.FormatInt(int64(t), 10))
			default:
				return strings.TrimSpace(fmt.Sprint(v))
			}
		}
	}
	return ""
}

func intVal(row map[string]interface{}, key string) int {
	v, ok := row[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	default:
		return 0
	}
}

func boolVal(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true")
	case float64:
		return t != 0
	default:
		return false
	}
}

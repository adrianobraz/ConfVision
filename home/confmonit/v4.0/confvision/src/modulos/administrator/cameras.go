package administrator

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/conexao"
	"confvision/src/modulos/confvision"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type cameraAdminRow struct {
	ID             int    `json:"id"`
	Nome           string `json:"nome"`
	NomeFranqueado string `json:"nome_franqueado"`
	NomeCliente    string `json:"nome_cliente"`
	Bloqueado      bool   `json:"bloqueado"`
	Ativo          bool   `json:"ativo"`
	Plano          string `json:"plano"`
	IDFranqueado   string `json:"id_franqueado"`
	IDCliente      string `json:"id_cliente"`
}

func ApiListarCameras(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	lista, err := fetchTodasVisCameras()
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	fraIDs, cliIDs := coletarIDsVinculo(lista)
	fraMap := lookupFranqueadosBatch(fraIDs)
	cliMap := lookupClientesBatch(cliIDs)

	out := make([]cameraAdminRow, 0, len(lista))
	for _, cam := range lista {
		row := montarCameraAdminRow(cam, fraMap, cliMap)
		if q != "" && !correspondeBusca(row, q) {
			continue
		}
		out = append(out, row)
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i].Nome), strings.ToLower(out[j].Nome)
		if a == b {
			return out[i].ID < out[j].ID
		}
		return a < b
	})

	payload, _ := json.Marshal(map[string]any{
		"status": "ok",
		"total":  len(out),
		"dados":  out,
	})
	auxiliar.RespostaAPP(w, payload)
}

func fetchTodasVisCameras() ([]map[string]any, error) {
	if config.XanoBaseUrl == "" {
		return nil, fmt.Errorf("XANO_BASE_URL nao configurado")
	}
	u := strings.TrimRight(config.XanoBaseUrl, "/") + "/vis_camera"
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xano HTTP %d: %s", resp.StatusCode, string(raw))
	}
	return normalizarListaXano(raw)
}

func normalizarListaXano(raw []byte) ([]map[string]any, error) {
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	for _, key := range []string{"dados", "items", "data"} {
		if v, ok := root[key].([]any); ok {
			return mapsFromAnySlice(v), nil
		}
	}
	return nil, fmt.Errorf("formato de lista vis_camera nao reconhecido")
}

func mapsFromAnySlice(items []any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func coletarIDsVinculo(lista []map[string]any) (fraIDs, cliIDs []string) {
	fraSet := map[string]struct{}{}
	cliSet := map[string]struct{}{}
	for _, cam := range lista {
		if id := textoCampo(cam, "id_franqueado"); id != "" {
			if _, ok := fraSet[id]; !ok {
				fraSet[id] = struct{}{}
				fraIDs = append(fraIDs, id)
			}
		}
		if id := textoCampo(cam, "id_cliente"); id != "" {
			if _, ok := cliSet[id]; !ok {
				cliSet[id] = struct{}{}
				cliIDs = append(cliIDs, id)
			}
		}
	}
	return fraIDs, cliIDs
}

func montarCameraAdminRow(cam map[string]any, fraMap, cliMap map[string]string) cameraAdminRow {
	id := intDeCampo(cam, "id")
	nome := textoCampo(cam, "nome")
	if nome == "" {
		nome = fmt.Sprintf("Câmera #%d", id)
	}
	idFra := textoCampo(cam, "id_franqueado")
	idCli := textoCampo(cam, "id_cliente")

	nomeFra := fraMap[idFra]
	if nomeFra == "" && idFra != "" {
		nomeFra = idFra
	}
	if nomeFra == "" {
		nomeFra = "—"
	}

	nomeCli := cliMap[idCli]
	if nomeCli == "" && idCli != "" {
		nomeCli = idCli
	}
	if nomeCli == "" {
		nomeCli = "—"
	}

	return cameraAdminRow{
		ID:             id,
		Nome:           nome,
		NomeFranqueado: nomeFra,
		NomeCliente:    nomeCli,
		Bloqueado:      confvision.TruthyCameraBool(cam["bloqueado"]),
		Ativo:          confvision.TruthyCameraBool(cam["ativo"]),
		Plano:          textoCampo(cam, "plano"),
		IDFranqueado:   idFra,
		IDCliente:      idCli,
	}
}

func correspondeBusca(row cameraAdminRow, q string) bool {
	blob := strings.ToLower(strings.Join([]string{
		row.Nome,
		row.NomeFranqueado,
		row.NomeCliente,
		row.Plano,
		fmt.Sprintf("%d", row.ID),
	}, " "))
	return strings.Contains(blob, q)
}

func textoCampo(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", v))
	case json.Number:
		return strings.TrimSpace(v.String())
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "<nil>" {
			return ""
		}
		return s
	}
}

func intDeCampo(m map[string]any, key string) int {
	s := textoCampo(m, key)
	if s == "" {
		return 0
	}
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func lookupFranqueadosBatch(ids []string) map[string]string {
	out := map[string]string{}
	if len(ids) == 0 {
		return out
	}
	db, err := conexao.Conectar()
	if err != nil {
		return out
	}
	defer db.Close()

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = strings.TrimSuffix(placeholders, ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := db.Query(`
		SELECT ID_Franqueado, RazaoSocial
		FROM franqueado
		WHERE ID_Franqueado IN (`+placeholders+`)`, args...)
	if err != nil {
		return out
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var nome sql.NullString
		if err := rows.Scan(&id, &nome); err != nil {
			continue
		}
		k := strings.TrimSpace(id.String)
		if k != "" {
			out[k] = strings.TrimSpace(nome.String)
		}
	}
	return out
}

func lookupClientesBatch(ids []string) map[string]string {
	out := map[string]string{}
	if len(ids) == 0 {
		return out
	}
	db, err := conexao.Conectar()
	if err != nil {
		return out
	}
	defer db.Close()

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = strings.TrimSuffix(placeholders, ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := db.Query(`
		SELECT ID_Cliente, Nome
		FROM cliente
		WHERE ID_Cliente IN (`+placeholders+`)`, args...)
	if err != nil {
		return out
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var nome sql.NullString
		if err := rows.Scan(&id, &nome); err != nil {
			continue
		}
		k := strings.TrimSpace(id.String)
		if k != "" {
			out[k] = strings.TrimSpace(nome.String)
		}
	}
	return out
}

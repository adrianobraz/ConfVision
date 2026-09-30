package confvision

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/conexao"
	"confvision/src/modulos/visdata"
	"confvision/src/seguranca"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type camerasResumoPayload struct {
	Total          int `json:"total"`
	Ativas         int `json:"ativas"`
	Inativas       int `json:"inativas"`
	SemComunicacao int `json:"sem_comunicacao"`
	Desarmadas     int `json:"desarmadas"`
	Pausadas       int `json:"pausadas"`
	Bloqueadas     int `json:"bloqueadas"`
	PingStaleMin   int `json:"ping_stale_min"`
}

func ProxyCamerasResumo(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	if seguranca.EhCliente(cookie) {
		responderEscopoProibido(w, "resumo restrito ao franqueado")
		return
	}
	idFra, err := FranqueadoDaSessao(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusForbidden, err)
		return
	}

	resumo, err := calcularResumoCameras(r.Context(), idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	raw, _ := json.Marshal(map[string]any{
		"status": "ok",
		"dados":  resumo,
	})
	auxiliar.RespostaAPP(w, raw)
}

func calcularResumoCameras(ctx context.Context, idFranqueado string) (camerasResumoPayload, error) {
	out := camerasResumoPayload{PingStaleMin: cameraPingStaleMinutes}

	if config.VisPostgresEnabled {
		sqlRes, err := visdata.CountCamerasResumoSQL(ctx, idFranqueado, cameraPingStaleMinutes)
		if err != nil {
			return out, err
		}
		out.Total = sqlRes.Total
		out.Ativas = sqlRes.Ativas
		out.Inativas = sqlRes.Inativas
		out.SemComunicacao = sqlRes.SemComunicacao
		out.Pausadas = sqlRes.Pausadas
		out.Bloqueadas = sqlRes.Bloqueadas

		rows, err := visdata.ListCamerasResumoRows(ctx, idFranqueado)
		if err != nil {
			return out, err
		}
		armadoMap := lookupArmadoDispositivos(idFranqueado)
		out.Desarmadas = contarDesarmadas(rows, armadoMap)
		return out, nil
	}

	lista, err := visdata.ListCamerasByFranqueado(ctx, idFranqueado)
	if err != nil || len(lista) == 0 {
		if config.XanoBaseUrl != "" {
			return calcularResumoCamerasLegado(ctx, idFranqueado)
		}
		if err != nil {
			return out, err
		}
	}
	armadoMap := lookupArmadoDispositivos(idFranqueado)
	return agregarResumoDeMaps(lista, armadoMap), nil
}

func calcularResumoCamerasLegado(ctx context.Context, idFranqueado string) (camerasResumoPayload, error) {
	out := camerasResumoPayload{PingStaleMin: cameraPingStaleMinutes}
	path := fmt.Sprintf("/vis_camera_by_franqueado?id_franqueado=%s", idFranqueado)
	status, raw, ok := fetchXano(http.MethodGet, path)
	if !ok || status >= 400 {
		return out, fmt.Errorf("falha ao listar cameras")
	}
	lista, err := parseListaCamerasJSON(raw)
	if err != nil {
		return out, err
	}
	armadoMap := lookupArmadoDispositivos(idFranqueado)
	return agregarResumoDeMaps(lista, armadoMap), nil
}

func parseListaCamerasJSON(raw []byte) ([]map[string]any, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		var arr []map[string]any
		if err2 := json.Unmarshal(raw, &arr); err2 != nil {
			return nil, err
		}
		return arr, nil
	}
	if dados, ok := root["dados"].([]any); ok {
		return mapsFromAnySliceAdmin(dados), nil
	}
	return nil, fmt.Errorf("formato invalido")
}

func mapsFromAnySliceAdmin(items []any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func agregarResumoDeMaps(lista []map[string]any, armadoMap map[string]string) camerasResumoPayload {
	out := camerasResumoPayload{PingStaleMin: cameraPingStaleMinutes}
	limite := pingStaleLimite()

	for _, cam := range lista {
		out.Total++
		ativo := truthyCameraBool(cam["ativo"])
		bloq := truthyCameraBool(cam["bloqueado"])
		pausado := truthyCameraBool(cam["analitico_pausado"])

		if bloq {
			out.Bloqueadas++
		}
		if ativo {
			out.Ativas++
		} else {
			out.Inativas++
		}
		if ativo && pausado {
			out.Pausadas++
		}
		if ativo && !bloq && cameraSemComunicacaoMap(cam, limite) {
			out.SemComunicacao++
		}
		if ativo && cameraDesarmadaMap(cam, armadoMap) {
			out.Desarmadas++
		}
	}
	return out
}

func contarDesarmadas(rows []visdata.CameraResumoRow, armadoMap map[string]string) int {
	n := 0
	for _, row := range rows {
		if !row.Ativo {
			continue
		}
		if cameraDesarmadaRow(row, armadoMap) {
			n++
		}
	}
	return n
}

func cameraDesarmadaRow(row visdata.CameraResumoRow, armadoMap map[string]string) bool {
	if !isPlanoArmadoGo(row.Plano) && !row.SomenteArmado {
		return false
	}
	armado := strings.ToUpper(strings.TrimSpace(armadoMap[row.IDDispositivo]))
	return armado == "N"
}

func cameraDesarmadaMap(cam map[string]any, armadoMap map[string]string) bool {
	plano := planoFromCam(cam)
	if !isPlanoArmadoGo(plano) && !truthyCameraBool(cam["somente_armado"]) {
		return false
	}
	idDisp := strings.TrimSpace(fmt.Sprint(cam["id_dispositivo"]))
	if idDisp == "<nil>" {
		idDisp = ""
	}
	return strings.ToUpper(strings.TrimSpace(armadoMap[idDisp])) == "N"
}

func isPlanoArmadoGo(plano string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(plano)), "analitico_armado")
}

func lookupArmadoDispositivos(idFranqueado string) map[string]string {
	out := map[string]string{}
	db, err := conexao.Conectar()
	if err != nil {
		return out
	}
	defer db.Close()

	rows, err := db.Query(`
SELECT d.ID_Dispositivo, d.Armado
FROM dispositivo d
INNER JOIN cliente c ON c.ID_Cliente = d.ID_Cliente
WHERE c.ID_Franqueado = ?`, idFranqueado)
	if err != nil {
		return out
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var armado sql.NullString
		if err := rows.Scan(&id, &armado); err != nil {
			continue
		}
		k := strings.TrimSpace(id.String)
		if k != "" {
			out[k] = strings.TrimSpace(armado.String)
		}
	}
	return out
}

func pingStaleLimite() time.Time {
	return time.Now().UTC().Add(-time.Duration(cameraPingStaleMinutes) * time.Minute)
}

func cameraSemComunicacaoMap(cam map[string]any, limite time.Time) bool {
	if !truthyCameraBool(cam["ativo"]) || truthyCameraBool(cam["bloqueado"]) {
		return false
	}
	raw := strings.TrimSpace(fmt.Sprint(cam["ultimo_ping_em"]))
	if raw == "" || raw == "<nil>" {
		return true
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", raw)
	}
	if err != nil {
		return true
	}
	return t.Before(limite)
}

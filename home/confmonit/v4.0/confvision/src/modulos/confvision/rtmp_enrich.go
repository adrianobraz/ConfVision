package confvision

import (
	"confvision/src/conexao"
	"confvision/src/config"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type cameraContexto struct {
	NomeCamera     string `json:"nome_camera"`
	NomeFranqueado string `json:"nome_franqueado"`
	NomeCliente    string `json:"nome_cliente"`
	Bloqueado      bool   `json:"bloqueado"`
}

func enrichGuardJSON(raw []byte) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return raw, err
	}

	dados, ok := root["dados"].([]any)
	if !ok || len(dados) == 0 {
		return raw, nil
	}

	ids := coletarCameraIDs(dados)
	if len(ids) == 0 {
		return raw, nil
	}

	ctxMap := lookupCamerasContexto(ids)
	for _, item := range dados {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := cameraIDFromRow(row)
		if id < 1 {
			continue
		}
		if ctx, ok := ctxMap[id]; ok {
			aplicarContexto(row, ctx, id)
		}
	}

	out, err := json.Marshal(root)
	if err != nil {
		return raw, err
	}
	return out, nil
}

func coletarCameraIDs(dados []any) []int {
	seen := map[int]struct{}{}
	var ids []int
	for _, item := range dados {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := cameraIDFromRow(row)
		if id < 1 {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func cameraIDFromRow(row map[string]any) int {
	switch v := row["camera_id"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	}
	if path, ok := row["path"].(string); ok && path != "" {
		if id, ok := ParseChaveRtmp(path); ok {
			return id
		}
		nome := strings.Trim(path, "/")
		if i := strings.LastIndex(nome, "/"); i >= 0 {
			nome = nome[i+1:]
		}
		if n, err := strconv.Atoi(nome); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func aplicarContexto(row map[string]any, ctx cameraContexto, cameraID int) {
	if cameraID > 0 {
		row["camera_id"] = cameraID
	}
	row["nome_camera"] = ctx.NomeCamera
	row["nome_franqueado"] = ctx.NomeFranqueado
	row["nome_cliente"] = ctx.NomeCliente
	row["bloqueado"] = ctx.Bloqueado
}

func lookupCamerasContexto(ids []int) map[int]cameraContexto {
	out := make(map[int]cameraContexto, len(ids))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)
		go func(cameraID int) {
			defer wg.Done()
			ctx := buscarContextoCamera(cameraID)
			mu.Lock()
			out[cameraID] = ctx
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return out
}

func buscarContextoCamera(cameraID int) cameraContexto {
	ctx := cameraContexto{
		NomeCamera:     fmt.Sprintf("Câmera #%d", cameraID),
		NomeFranqueado: "—",
		NomeCliente:    "—",
	}

	cam, ok := fetchVisCamera(cameraID)
	if !ok {
		return ctx
	}

	if nome := strings.TrimSpace(fmt.Sprint(cam["nome"])); nome != "" && nome != "<nil>" {
		ctx.NomeCamera = nome
	}
	ctx.Bloqueado = truthyCameraBool(cam["bloqueado"])

	idFra := strings.TrimSpace(fmt.Sprint(cam["id_franqueado"]))
	if idFra != "" && idFra != "<nil>" {
		if nome := lookupNomeFranqueado(idFra); nome != "" {
			ctx.NomeFranqueado = nome
		} else {
			ctx.NomeFranqueado = "Franq. " + idFra
		}
	}

	idCli := strings.TrimSpace(fmt.Sprint(cam["id_cliente"]))
	if idCli != "" && idCli != "<nil>" {
		if nome := lookupNomeCliente(idCli); nome != "" {
			ctx.NomeCliente = nome
		} else {
			ctx.NomeCliente = "Cliente " + idCli
		}
	}

	return ctx
}

func fetchVisCamera(cameraID int) (map[string]any, bool) {
	if config.XanoBaseUrl == "" {
		return nil, false
	}
	id := strconv.Itoa(cameraID)
	u := fmt.Sprintf("%s/vis_camera/%s?vis_camera_id=%s",
		strings.TrimRight(config.XanoBaseUrl, "/"), id, id)
	resp, err := http.Get(u)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, false
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false
	}
	var cam map[string]any
	if json.Unmarshal(raw, &cam) != nil {
		return nil, false
	}
	return cam, true
}

func lookupNomeFranqueado(idFranqueado string) string {
	db, err := conexao.Conectar()
	if err != nil {
		return ""
	}
	defer db.Close()

	var nome sql.NullString
	err = db.QueryRow(`
		SELECT RazaoSocial
		FROM franqueado
		WHERE ID_Franqueado = ?
		LIMIT 1
	`, idFranqueado).Scan(&nome)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(nome.String)
}

func lookupNomeCliente(idCliente string) string {
	db, err := conexao.Conectar()
	if err != nil {
		return ""
	}
	defer db.Close()

	var nome sql.NullString
	err = db.QueryRow(`
		SELECT Nome
		FROM cliente
		WHERE ID_Cliente = ?
		LIMIT 1
	`, idCliente).Scan(&nome)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(nome.String)
}

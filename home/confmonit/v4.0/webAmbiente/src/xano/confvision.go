package xano

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"webAmbiente/src/seguranca"
)

type VisCameraResumo struct {
	Id   int    `json:"id"`
	Nome string `json:"nome"`
}

func BuscarCameraPorSetor(idSetor, idDispositivo, idFranqueado, particao, zonauser string) (*VisCameraResumo, error) {
	idSetor = strings.ToUpper(strings.TrimSpace(idSetor))
	idDispositivo = strings.TrimSpace(idDispositivo)
	idFranqueado = strings.TrimSpace(idFranqueado)
	particao = strings.TrimSpace(particao)
	zonauser = strings.TrimSpace(zonauser)

	if idFranqueado != "" {
		if cam, err := buscarCameraPorIdSetor(idFranqueado, idSetor); err == nil && cam != nil {
			return cam, nil
		}
	}

	if idDispositivo != "" && particao != "" && zonauser != "" {
		if cam, err := buscarCameraPorSetorAPI(idDispositivo, idFranqueado, particao, zonauser); err == nil && cam != nil {
			return cam, nil
		}
	}

	return nil, nil
}

func buscarCameraPorIdSetor(idFranqueado, idSetor string) (*VisCameraResumo, error) {
	path := "/vis_camera_by_franqueado?id_franqueado=" + url.QueryEscape(idFranqueado)
	raw, err := seguranca.ReqConfVisionGET(path)
	if err != nil {
		return nil, err
	}

	lista, err := parseVisCameraLista(raw)
	if err != nil {
		return nil, err
	}

	for _, cam := range lista {
		if idSetorIgual(cam, idSetor) {
			return camResumo(cam), nil
		}
	}
	return nil, nil
}

func buscarCameraPorSetorAPI(idDispositivo, idFranqueado, particao, zonauser string) (*VisCameraResumo, error) {
	q := url.Values{}
	q.Set("id_dispositivo", idDispositivo)
	q.Set("particao", particao)
	q.Set("zonauser", zonauser)
	if idFranqueado != "" {
		q.Set("id_franqueado", idFranqueado)
	}

	raw, err := seguranca.ReqConfVisionGET("/vis_camera_by_setor?" + q.Encode())
	if err != nil {
		return nil, err
	}

	cam, err := parseVisCameraUnica(raw)
	if err != nil {
		return nil, err
	}
	if cam == nil {
		return nil, nil
	}
	return camResumo(cam), nil
}

func parseVisCameraLista(raw []byte) ([]map[string]interface{}, error) {
	var direct []map[string]interface{}
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, nil
	}

	var wrapped struct {
		Dados []map[string]interface{} `json:"dados"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("parse vis_camera lista: %w", err)
	}
	return wrapped.Dados, nil
}

func parseVisCameraUnica(raw []byte) (map[string]interface{}, error) {
	var direct map[string]interface{}
	if err := json.Unmarshal(raw, &direct); err == nil {
		if id := direct["id"]; id != nil {
			return direct, nil
		}
		if dados, ok := direct["dados"].(map[string]interface{}); ok {
			return dados, nil
		}
		return nil, nil
	}
	return nil, fmt.Errorf("parse vis_camera: json invalido")
}

func idSetorIgual(cam map[string]interface{}, idSetor string) bool {
	val, ok := cam["id_setor"]
	if !ok || val == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(fmt.Sprint(val)), idSetor)
}

func camResumo(cam map[string]interface{}) *VisCameraResumo {
	id := camIntVal(cam, "id")
	if id <= 0 {
		return nil
	}
	nome := strValCam(cam, "nome")
	return &VisCameraResumo{Id: id, Nome: nome}
}

func strValCam(row map[string]interface{}, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func camIntVal(row map[string]interface{}, key string) int {
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
		n, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(v)))
		return n
	}
}

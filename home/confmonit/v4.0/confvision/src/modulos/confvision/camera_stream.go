package confvision

import (
	"fmt"
	"strings"
)

// CameraPodeStream: bloqueado=false E (plano=online OU ativo=true).
// Plano online trata ativo=false como normal (sob demanda).
func CameraPodeStream(bloqueado, ativo bool, plano string) (ok bool, motivo string) {
	if bloqueado {
		return false, "camera_bloqueada"
	}
	if strings.EqualFold(strings.TrimSpace(plano), "online") || ativo {
		return true, ""
	}
	return false, "camera_inativa"
}

func cameraPodeStreamMsg(motivo string) string {
	switch motivo {
	case "camera_bloqueada":
		return "Câmera suspensa pelo administrativo (pagamento)."
	case "camera_inativa":
		return "Câmera desligada. Ative no cadastro."
	default:
		return "Stream indisponível."
	}
}

func truthyCameraBool(v interface{}) bool {
	return TruthyCameraBool(v)
}

func TruthyCameraBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		return s == "1" || s == "true" || s == "yes" || s == "sim" || s == "s"
	default:
		return false
	}
}

func planoFromCam(cam map[string]any) string {
	if cam == nil || cam["plano"] == nil {
		return ""
	}
	if p, ok := cam["plano"].(string); ok {
		return p
	}
	return strings.TrimSpace(strings.ToLower(fmt.Sprint(cam["plano"])))
}

func cameraPodeStreamFromMap(cam map[string]any) (ok bool, motivo string) {
	if cam == nil {
		return false, "camera_invalida"
	}
	return CameraPodeStream(
		truthyCameraBool(cam["bloqueado"]),
		truthyCameraBool(cam["ativo"]),
		planoFromCam(cam),
	)
}

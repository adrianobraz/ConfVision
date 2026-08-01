package licenca

import (
	"encoding/json"
	"fmt"
	"net/http"

	"franqueadopro/src/auxiliar"
)

type Limites struct {
	UsuariosAlarmeMax int
	SetoresAlarmeMax  int
	ClientesMax       int
	ContasMax         int
}

func limitesDeEstado(est Estado) Limites {
	lim := Limites{
		UsuariosAlarmeMax: 9999,
		SetoresAlarmeMax:  9999,
		ClientesMax:       9999,
		ContasMax:         9999,
	}
	if est.Limites == nil {
		return lim
	}
	if v, ok := est.Limites["usuarios_alarme_max"]; ok && v > 0 {
		lim.UsuariosAlarmeMax = v
	}
	if v, ok := est.Limites["setores_alarme_max"]; ok && v > 0 {
		lim.SetoresAlarmeMax = v
	}
	if v, ok := est.Limites["clientes_max"]; ok && v > 0 {
		lim.ClientesMax = v
	}
	if v, ok := est.Limites["contas_max"]; ok && v > 0 {
		lim.ContasMax = v
	}
	return lim
}

func PodeInserir(est Estado, recurso string, quantidadeAtual int) (bool, string) {
	if !est.Liberado {
		return false, "Licença inativa."
	}
	lim := limitesDeEstado(est)
	switch recurso {
	case "clientes":
		if quantidadeAtual >= lim.ClientesMax {
			return false, fmt.Sprintf("Limite da cota atingido: máximo de %d clientes. Compre um Pacote de Cotas em Meu Plano ou faça upgrade.", lim.ClientesMax)
		}
	case "contas":
		if quantidadeAtual >= lim.ContasMax {
			return false, fmt.Sprintf("Limite da cota atingido: máximo de %d contas/dispositivos. Compre um Pacote de Cotas em Meu Plano.", lim.ContasMax)
		}
	case "usuarios_alarme":
		if quantidadeAtual >= lim.UsuariosAlarmeMax {
			return false, fmt.Sprintf("Limite da cota atingido: máximo de %d usuários de alarme. Compre um Pacote de Cotas em Meu Plano.", lim.UsuariosAlarmeMax)
		}
	case "setores_alarme":
		if quantidadeAtual >= lim.SetoresAlarmeMax {
			return false, fmt.Sprintf("Limite da cota atingido: máximo de %d setores de alarme. Compre um Pacote de Cotas em Meu Plano.", lim.SetoresAlarmeMax)
		}
	default:
		return true, ""
	}
	return true, ""
}

func RespostaLimiteExcedido(w http.ResponseWriter, msg string) {
	auxiliar.RespostaJSON(w, http.StatusForbidden, map[string]interface{}{
		"status": "ERRO",
		"erro":   msg,
		"codigo": "limite_plano",
	})
}

func ContarItensJSON(corpo []byte) int {
	if len(corpo) == 0 {
		return 0
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(corpo, &arr); err == nil {
		return len(arr)
	}
	var wrapped struct {
		Dados []json.RawMessage `json:"dados"`
	}
	if err := json.Unmarshal(corpo, &wrapped); err == nil && len(wrapped.Dados) > 0 {
		return len(wrapped.Dados)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(corpo, &obj); err == nil {
		if d, ok := obj["dados"]; ok {
			return ContarItensJSON(d)
		}
	}
	return 0
}

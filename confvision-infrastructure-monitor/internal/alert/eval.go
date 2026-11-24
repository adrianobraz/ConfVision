package alert

import (
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

const (
	Normal      = "normal"
	Suspeita    = "suspeita"
	Alerta      = "alerta"
	Recuperacao = "recuperacao"
)

type Thresholds struct {
	CPUSuspeita  float64
	CPUAlerta    float64
	RAMSuspeita  float64
	RAMAlerta    float64
	DiskSuspeita float64
	DiskAlerta   float64
	YoloP95      float64
}

type Open struct {
	Estado string
	Desde  time.Time
}

// Evaluate mantém uma ocorrência por chave. O mesmo estado aberto não gera outra.
func Evaluate(prev map[string]Open, now time.Time, levels map[string]string, detalhes map[string]string) ([]model.Alert, map[string]Open) {
	next := map[string]Open{}
	var out []model.Alert
	seen := map[string]bool{}
	for chave, level := range levels {
		seen[chave] = true
		estado := transition(stateOf(prev, chave), level)
		desde := now
		if old, ok := prev[chave]; ok && old.Estado == estado {
			desde = old.Desde
		}
		if estado == Normal {
			continue
		}
		next[chave] = Open{Estado: estado, Desde: desde}
		out = append(out, model.Alert{
			Chave:   chave,
			Estado:  estado,
			Nome:    chave,
			Origem:  model.OrigemCalculado,
			Detalhe: detalhes[chave],
			Desde:   desde,
		})
	}
	for chave, old := range prev {
		if seen[chave] {
			continue
		}
		if old.Estado == Normal || old.Estado == Recuperacao {
			continue
		}
		out = append(out, model.Alert{
			Chave:  chave,
			Estado: Recuperacao,
			Nome:   chave,
			Origem: model.OrigemCalculado,
			Desde:  now,
		})
	}
	return out, next
}

func transition(prev, level string) string {
	if level == "" || level == Normal {
		if prev == Alerta || prev == Suspeita {
			return Recuperacao
		}
		return Normal
	}
	return level
}

func stateOf(prev map[string]Open, chave string) string {
	if prev == nil {
		return Normal
	}
	old, ok := prev[chave]
	if !ok {
		return Normal
	}
	return old.Estado
}

func Level(value float64, suspeita, alerta float64) string {
	if value >= alerta {
		return Alerta
	}
	if value >= suspeita {
		return Suspeita
	}
	return Normal
}

func Find(readings []model.Reading, nome string) (model.Reading, bool) {
	for _, item := range readings {
		if item.Nome == nome {
			return item, true
		}
	}
	return model.Reading{}, false
}

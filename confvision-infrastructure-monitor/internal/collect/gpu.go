package collect

import (
	"os"
	"path/filepath"
	"strings"

	"confvision-infrastructure-monitor/internal/model"
)

// ReadGPU identifica a placa pelo /proc. Ausência de GPU não é falha:
// a origem fica indisponível e o cálculo de capacidade segue na CPU.
func ReadGPU(root string) []model.Reading {
	base := filepath.Join(root, "proc", "driver", "nvidia", "gpus")
	entries, err := os.ReadDir(base)
	if err != nil || len(entries) == 0 {
		return []model.Reading{
			model.Text("gpu_presente", model.OrigemMedido, "nao"),
			model.Unavailable("gpu_modelo", "gpu nao detectada"),
			model.Unavailable("gpu_util", "gpu nao detectada"),
			model.Unavailable("gpu_vram_mb", "gpu nao detectada"),
		}
	}
	modelo := ""
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(base, entry.Name(), "information"))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if strings.EqualFold(key, "Model") && value != "" {
				modelo = value
			}
		}
	}
	out := []model.Reading{model.Text("gpu_presente", model.OrigemMedido, "sim")}
	if modelo != "" {
		out = append(out, model.Text("gpu_modelo", model.OrigemMedido, modelo))
	} else {
		out = append(out, model.Unavailable("gpu_modelo", "modelo nao legivel em proc"))
	}
	out = append(out, model.Unavailable("gpu_vram_mb", "vram nao exposta em proc"))
	out = append(out, model.Unavailable("gpu_util", "utilizacao nao exposta em proc"))
	return out
}

// UsoGPUNoLog lê o device que o worker já escreveu no log.
// device=cpu significa que a GPU, mesmo presente, não entra na conta.
func UsoGPUNoLog(text string) model.Reading {
	switch {
	case strings.Contains(text, "device=cuda"), strings.Contains(text, "device=mps"):
		return model.Text("gpu_em_uso", model.OrigemMedido, "sim")
	case strings.Contains(text, "device=cpu"):
		return model.Text("gpu_em_uso", model.OrigemMedido, "nao")
	default:
		return model.Unavailable("gpu_em_uso", "worker nao informou o device")
	}
}

package collect

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"confvision-infrastructure-monitor/internal/model"
)

// ReadLinkMbps lê a velocidade anunciada pela placa em sysfs.
// Tráfego instantâneo não entra aqui. Sem velocidade positiva, o valor fica ausente.
func ReadLinkMbps(root string) model.Reading {
	base := filepath.Join(root, "sys", "class", "net")
	entries, err := os.ReadDir(base)
	if err != nil {
		return model.Unavailable("link_mbps", "sysfs de rede ausente")
	}
	melhor := 0
	for _, entry := range entries {
		nome := entry.Name()
		if ignorarIface(nome) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(base, nome, "speed"))
		if err != nil {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(string(raw)))
		if err != nil || n <= 0 {
			continue
		}
		if n > melhor {
			melhor = n
		}
	}
	if melhor == 0 {
		return model.Unavailable("link_mbps", "velocidade do link nao exposta")
	}
	return model.Measured("link_mbps", "mbps", float64(melhor))
}

func ignorarIface(nome string) bool {
	switch nome {
	case "lo", "docker0":
		return true
	}
	prefixos := []string{"veth", "br-", "virbr", "tun", "dummy", "cali", "flannel", "cni", "lxc"}
	for _, p := range prefixos {
		if strings.HasPrefix(nome, p) {
			return true
		}
	}
	return false
}

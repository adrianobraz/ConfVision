package collect

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

var (
	errDockerLeitura = errors.New("operacao docker fora da lista de leitura")
	reContainerID    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)
)

// DockerGet é o cliente HTTP já apontado para o socket local.
// A API do monitor não recebe esse cliente nem caminhos livres.
type DockerGet func(ctx context.Context, path string) ([]byte, error)

func AllowDockerPath(path string) bool {
	if strings.Contains(path, "..") || strings.Contains(path, "?") {
		return false
	}
	blocked := []string{"/start", "/stop", "/kill", "/restart", "/exec", "/attach", "/archive", "/rename", "/update", "/pause", "/unpause", "/wait", "/resize"}
	for _, item := range blocked {
		if strings.Contains(path, item) {
			return false
		}
	}
	if path == "/containers/json" {
		return true
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 3 && parts[0] == "containers" && reContainerID.MatchString(parts[1]) {
		switch parts[2] {
		case "json", "stats", "logs", "top":
			return true
		}
	}
	return false
}

func dockerQuery(path string) string {
	switch {
	case strings.HasSuffix(path, "/stats"):
		return path + "?stream=false"
	case strings.HasSuffix(path, "/logs"):
		return path + "?stdout=1&stderr=1&tail=400"
	default:
		return path
	}
}

func DoDockerGet(ctx context.Context, client *http.Client, path string) ([]byte, error) {
	if !AllowDockerPath(path) {
		return nil, errDockerLeitura
	}
	if client == nil {
		return nil, errors.New("cliente docker ausente")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker"+dockerQuery(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("docker recusou a leitura")
	}
	return body, nil
}

type dockerListItem struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
	State string   `json:"State"`
}

type dockerInspect struct {
	Name  string `json:"Name"`
	State struct {
		Status    string `json:"Status"`
		StartedAt string `json:"StartedAt"`
	} `json:"State"`
	RestartCount int `json:"RestartCount"`
	HostConfig   struct {
		Memory   int64 `json:"Memory"`
		NanoCpus int64 `json:"NanoCpus"`
	} `json:"HostConfig"`
}

type dockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

func ContainersFromDocker(ctx context.Context, get DockerGet) ([]model.Container, model.Reading, map[string]string) {
	if get == nil {
		return nil, model.Unavailable("containers", "socket docker nao configurado"), nil
	}
	raw, err := get(ctx, "/containers/json")
	if err != nil {
		return nil, model.Unavailable("containers", "leitura docker indisponivel"), nil
	}
	var items []dockerListItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, model.Unavailable("containers", "resposta docker ilegivel"), nil
	}
	logs := map[string]string{}
	out := make([]model.Container, 0, len(items))
	for _, item := range items {
		name := ""
		if len(item.Names) > 0 {
			name = strings.TrimPrefix(item.Names[0], "/")
		}
		ctr := model.Container{
			ID:           item.ID,
			Nome:         name,
			Papel:        PapelDoContainer(name),
			IncluiFilhos: true,
			Estado:       model.Text("estado", model.OrigemMedido, item.State),
			CPU:          model.Unavailable("cpu", "stats nao lidos"),
			Memoria:      model.Unavailable("memoria", "stats nao lidos"),
			LimiteCPU:    model.Unavailable("limite_cpu", "inspect nao lido"),
			LimiteMem:    model.Unavailable("limite_memoria", "inspect nao lido"),
			Uptime:       model.Unavailable("uptime", "inspect nao lido"),
			Reinicios:    model.Unavailable("reinicios", "inspect nao lido"),
			RedeRX:       model.Unavailable("rede_rx", "stats nao lidos"),
			RedeTX:       model.Unavailable("rede_tx", "stats nao lidos"),
		}
		id := item.ID
		if !reContainerID.MatchString(id) {
			out = append(out, ctr)
			continue
		}
		if body, err := get(ctx, "/containers/"+id+"/json"); err == nil {
			applyInspect(&ctr, body)
		}
		if body, err := get(ctx, "/containers/"+id+"/stats"); err == nil {
			applyStats(&ctr, body)
		}
		if ctr.Papel == "worker" {
			if body, err := get(ctx, "/containers/"+id+"/logs"); err == nil {
				logs[name] = string(body)
			}
		}
		out = append(out, ctr)
	}
	return out, model.Text("containers", model.OrigemMedido, "lido"), logs
}

func applyInspect(ctr *model.Container, raw []byte) {
	var info dockerInspect
	if err := json.Unmarshal(raw, &info); err != nil {
		return
	}
	if info.State.Status != "" {
		ctr.Estado = model.Text("estado", model.OrigemMedido, info.State.Status)
	}
	ctr.Reinicios = model.Measured("reinicios", "count", float64(info.RestartCount))
	if info.HostConfig.Memory > 0 {
		ctr.LimiteMem = model.Measured("limite_memoria", "mb", float64(info.HostConfig.Memory)/(1024*1024))
	} else {
		ctr.LimiteMem = model.Unavailable("limite_memoria", "sem limite de memoria")
	}
	if info.HostConfig.NanoCpus > 0 {
		ctr.LimiteCPU = model.Measured("limite_cpu", "nucleos", float64(info.HostConfig.NanoCpus)/1e9)
	} else {
		ctr.LimiteCPU = model.Unavailable("limite_cpu", "sem limite de cpu")
	}
	if started, err := time.Parse(time.RFC3339Nano, info.State.StartedAt); err == nil && !started.IsZero() {
		ctr.Uptime = model.Measured("uptime", "seg", time.Since(started).Seconds())
	}
}

func applyStats(ctr *model.Container, raw []byte) {
	var st dockerStats
	if err := json.Unmarshal(raw, &st); err != nil {
		return
	}
	deltaCPU := float64(st.CPUStats.CPUUsage.TotalUsage) - float64(st.PreCPUStats.CPUUsage.TotalUsage)
	deltaSys := float64(st.CPUStats.SystemCPUUsage) - float64(st.PreCPUStats.SystemCPUUsage)
	if deltaCPU > 0 && deltaSys > 0 && st.CPUStats.OnlineCPUs > 0 {
		pct := (deltaCPU / deltaSys) * float64(st.CPUStats.OnlineCPUs) * 100
		ctr.CPU = model.Measured("cpu", "pct", pct)
	}
	if st.MemoryStats.Usage > 0 {
		ctr.Memoria = model.Measured("memoria", "mb", float64(st.MemoryStats.Usage)/(1024*1024))
	}
	var rx, tx uint64
	for _, net := range st.Networks {
		rx += net.RxBytes
		tx += net.TxBytes
	}
	if rx > 0 || tx > 0 {
		ctr.RedeRX = model.Measured("rede_rx", "bytes", float64(rx))
		ctr.RedeTX = model.Measured("rede_tx", "bytes", float64(tx))
	}
}

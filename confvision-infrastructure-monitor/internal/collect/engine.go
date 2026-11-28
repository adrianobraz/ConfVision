package collect

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"confvision-infrastructure-monitor/internal/alert"
	"confvision-infrastructure-monitor/internal/config"
	"confvision-infrastructure-monitor/internal/model"
)

// Engine guarda só a última coleta em memória. Reinício perde a amostra.
type Engine struct {
	cfg    config.Config
	root   string
	docker DockerGet
	now    func() time.Time

	mu       sync.RWMutex
	host     HostState
	procs    map[int]ProcSample
	alerts   map[string]alert.Open
	last     model.Snapshot
	ready    bool
	restarts map[string]float64
	ffmpegs  int
}

func NewEngine(cfg config.Config) *Engine {
	root := strings.TrimSpace(cfg.HostRoot)
	if root == "" {
		root = "/"
	}
	eng := &Engine{
		cfg:      cfg,
		root:     root,
		now:      time.Now,
		alerts:   map[string]alert.Open{},
		restarts: map[string]float64{},
	}
	if client := NewDockerClient(cfg.DockerSocket); client != nil {
		eng.docker = func(ctx context.Context, path string) ([]byte, error) {
			return DoDockerGet(ctx, client, path)
		}
	}
	return eng
}

func (e *Engine) Run(ctx context.Context) {
	e.safeCollect(ctx)
	ticker := time.NewTicker(e.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("nivel=info evento=encerramento componente=coleta")
			return
		case <-ticker.C:
			e.safeCollect(ctx)
		}
	}
}

// Refresh executa uma coleta. Não grava arquivo nem banco.
func (e *Engine) Refresh(ctx context.Context) {
	e.safeCollect(ctx)
}

func (e *Engine) Snapshot() (model.Snapshot, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.last, e.ready
}

func (e *Engine) safeCollect(ctx context.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("nivel=erro evento=panic_coleta")
		}
	}()
	snap := e.collect(ctx)
	e.mu.Lock()
	e.last = snap
	e.ready = true
	e.mu.Unlock()
}

func (e *Engine) collect(ctx context.Context) model.Snapshot {
	now := e.now()
	e.mu.RLock()
	prevHost := e.host
	prevProc := e.procs
	prevAlerts := e.alerts
	prevRestarts := copyFloatMap(e.restarts)
	prevFFmpeg := e.ffmpegs
	e.mu.RUnlock()

	seconds := e.cfg.Interval.Seconds()
	if prevHost.Seen && !prevHost.When.IsZero() {
		seconds = now.Sub(prevHost.When).Seconds()
	}

	host, nextHost := ReadHost(e.root, prevHost, now)
	containers, containerStatus, logs := ContainersFromDocker(ctx, e.docker)
	if containerStatus.Origem == model.OrigemIndisponivel {
		if samples, cg := ReadCgroups(e.root); cg.Origem == model.OrigemMedido && len(samples) > 0 {
			containers = containersFromCgroup(samples)
			containerStatus = cg
		}
	}
	procs, nextProc, procStatus := ReadProcesses(e.root, prevProc, seconds)
	process, cameras, metricsStatus := ReadMetricsFile(e.cfg.MetricsLog)
	var logText strings.Builder
	for _, body := range logs {
		decoded := DecodeDockerLog([]byte(body))
		logText.WriteString(decoded)
		logText.WriteByte('\n')
		if metricsStatus.Origem == model.OrigemIndisponivel {
			p, c := ParseMetrics(decoded)
			if len(p) > 0 || len(c) > 0 {
				process, cameras = p, c
				metricsStatus = model.Text("metrics_log", model.OrigemMedido, "log do container")
			}
		}
	}
	worker := readingsDeMapa(process)
	worker = append(worker, UsoGPUNoLog(logText.String()))

	servicos := make([]model.Container, 0)
	for _, item := range containers {
		if item.Papel != "" {
			servicos = append(servicos, item)
		}
	}

	snap := model.Snapshot{
		ColetadoEm:          now.UTC(),
		Servidor:            model.Server{ID: e.cfg.ServerID, Nome: e.cfg.ServerNome},
		Host:                host,
		Containers:          containers,
		Servicos:            servicos,
		Processos:           procs,
		Cameras:             cameras,
		Worker:              worker,
		MediaMTX:            ReadMediaMTX(ctx, e.cfg.MediaMTXURL, e.cfg.MediaMTXUser, e.cfg.MediaMTXPass),
		Redis:               ReadRedis(e.cfg.RedisURL, e.cfg.RedisKeys),
		Capacidade:          capacity(host, e.cfg.CPUQuota, e.cfg.RAMQuotaMB),
		Coleta:              []model.Reading{containerStatus, procStatus, metricsStatus},
		HistoricoPersistido: false,
	}
	levels, detalhes, restarts, ffmpegN := e.levels(snap, prevRestarts, prevFFmpeg)
	alerts, nextAlerts := alert.Evaluate(prevAlerts, now.UTC(), levels, detalhes)
	snap.Alertas = alerts

	e.mu.Lock()
	e.host = nextHost
	e.procs = nextProc
	e.alerts = nextAlerts
	e.restarts = restarts
	e.ffmpegs = ffmpegN
	e.mu.Unlock()
	return snap
}

func (e *Engine) levels(snap model.Snapshot, prevRestarts map[string]float64, prevFFmpeg int) (map[string]string, map[string]string, map[string]float64, int) {
	levels := map[string]string{}
	detalhes := map[string]string{}
	th := alert.Thresholds{
		CPUSuspeita:  e.cfg.AlertCPUSus,
		CPUAlerta:    e.cfg.AlertCPUAl,
		RAMSuspeita:  e.cfg.AlertRAMSus,
		RAMAlerta:    e.cfg.AlertRAMAl,
		DiskSuspeita: e.cfg.AlertDiskSus,
		DiskAlerta:   e.cfg.AlertDiskAl,
		YoloP95:      e.cfg.AlertYoloP95,
	}
	if cpu, ok := alert.Find(snap.Host, "cpu_pct"); ok && cpu.Origem == model.OrigemMedido && cpu.Valor != nil {
		levels["cpu"] = alert.Level(*cpu.Valor, th.CPUSuspeita, th.CPUAlerta)
	}
	if ram, ok := percentOf(snap.Host, "ram_usada_mb", "ram_total_mb"); ok {
		levels["memoria"] = alert.Level(ram, th.RAMSuspeita, th.RAMAlerta)
	}
	if disk, ok := percentOf(snap.Host, "disco_usado_mb", "disco_total_mb"); ok {
		levels["disco"] = alert.Level(disk, th.DiskSuspeita, th.DiskAlerta)
	}
	hostDown := true
	for _, item := range snap.Host {
		if item.Origem == model.OrigemMedido {
			hostDown = false
			break
		}
	}
	if hostDown {
		levels["coleta"] = alert.Alerta
		detalhes["coleta"] = "nenhuma metrica de host medida"
	}
	restarts := map[string]float64{}
	for _, item := range snap.Containers {
		if item.Estado.Texto == "exited" || item.Estado.Texto == "dead" {
			levels["servico_"+item.Nome] = alert.Alerta
			detalhes["servico_"+item.Nome] = "container existente fora de execucao"
		}
		if item.Reinicios.Origem == model.OrigemMedido && item.Reinicios.Valor != nil {
			restarts[item.Nome] = *item.Reinicios.Valor
			if prev, ok := prevRestarts[item.Nome]; ok && *item.Reinicios.Valor-prev >= 3 {
				levels["reinicio_"+item.Nome] = alert.Alerta
				detalhes["reinicio_"+item.Nome] = "reinicios aumentaram desde a amostra anterior"
			}
		}
	}
	ffmpegN := 0
	for _, proc := range snap.Processos {
		if proc.Nome == "ffmpeg" {
			ffmpegN++
		}
	}
	if prevFFmpeg > 0 && ffmpegN-prevFFmpeg >= 5 {
		levels["ffmpeg"] = alert.Alerta
		detalhes["ffmpeg"] = "quantidade de processos ffmpeg cresceu"
	}
	var maxP95 *float64
	for _, cam := range snap.Cameras {
		if reading, ok := cam.Campos["yolo_ms_p95"]; ok && reading.Origem == model.OrigemMedido && reading.Valor != nil {
			if maxP95 == nil || *reading.Valor > *maxP95 {
				v := *reading.Valor
				maxP95 = &v
			}
		}
	}
	if maxP95 != nil && *maxP95 >= th.YoloP95 {
		levels["yolo"] = alert.Alerta
		detalhes["yolo"] = "p95 do yolo acima do limiar configurado"
	}
	return levels, detalhes, restarts, ffmpegN
}

func percentOf(readings []model.Reading, part, total string) (float64, bool) {
	a, oka := alert.Find(readings, part)
	b, okb := alert.Find(readings, total)
	if !oka || !okb || a.Valor == nil || b.Valor == nil || *b.Valor <= 0 {
		return 0, false
	}
	if a.Origem != model.OrigemMedido || b.Origem != model.OrigemMedido {
		return 0, false
	}
	return *a.Valor / *b.Valor * 100, true
}

func capacity(host []model.Reading, cpuQuota, ramQuota *float64) model.Capacity {
	cap := model.Capacity{
		CPUMedida:      pick(host, "cpu_pct", "cpu do host nao medida"),
		RAMMedida:      pick(host, "ram_usada_mb", "ram do host nao medida"),
		CPUConfigurada: model.Unavailable("cpu_configurada", "MONITOR_CPU_QUOTA nao definido"),
		RAMConfigurada: model.Unavailable("ram_configurada", "MONITOR_RAM_QUOTA_MB nao definido"),
		CPUProjetada:   model.Unavailable("cpu_projetada", "monitor sem historico para projetar"),
		RAMProjetada:   model.Unavailable("ram_projetada", "monitor sem historico para projetar"),
		MargemCPU:      model.Unavailable("margem_cpu", "sem cota configurada"),
		MargemRAM:      model.Unavailable("margem_ram", "sem cota configurada"),
	}
	if cpuQuota != nil && cap.CPUMedida.Valor != nil && cap.CPUMedida.Origem == model.OrigemMedido {
		cap.CPUConfigurada = model.Text("cpu_configurada", model.OrigemCalculado, "informada por ambiente")
		cap.CPUConfigurada.Valor = cpuQuota
		cap.CPUConfigurada.Unidade = "pct"
		cap.MargemCPU = model.Calculated("margem_cpu", "pct", *cpuQuota-*cap.CPUMedida.Valor)
	}
	if ramQuota != nil && cap.RAMMedida.Valor != nil && cap.RAMMedida.Origem == model.OrigemMedido {
		cap.RAMConfigurada = model.Measured("ram_configurada", "mb", *ramQuota)
		cap.RAMConfigurada.Origem = model.OrigemCalculado
		cap.MargemRAM = model.Calculated("margem_ram", "mb", *ramQuota-*cap.RAMMedida.Valor)
	}
	return cap
}

func pick(readings []model.Reading, nome, motivo string) model.Reading {
	if item, ok := alert.Find(readings, nome); ok {
		return item
	}
	return model.Unavailable(nome, motivo)
}

func containersFromCgroup(samples []CgroupSample) []model.Container {
	out := make([]model.Container, 0, len(samples))
	for _, sample := range samples {
		item := model.Container{
			Nome:         sample.ID,
			IncluiFilhos: true,
			Estado:       model.Unavailable("estado", "cgroup nao informa estado docker"),
			CPU:          model.Unavailable("cpu", "cgroup exige duas amostras para percentual"),
			Memoria:      model.Unavailable("memoria", "memory.current ausente"),
			LimiteCPU:    model.Unavailable("limite_cpu", "cpu.max ausente"),
			LimiteMem:    model.Unavailable("limite_memoria", "memory.max ausente"),
			Uptime:       model.Unavailable("uptime", "cgroup nao informa uptime"),
			Reinicios:    model.Unavailable("reinicios", "cgroup nao informa reinicios"),
			RedeRX:       model.Unavailable("rede_rx", "cgroup nao informa rede"),
			RedeTX:       model.Unavailable("rede_tx", "cgroup nao informa rede"),
		}
		if sample.MemBytes != nil {
			item.Memoria = model.Measured("memoria", "mb", *sample.MemBytes/(1024*1024))
		}
		if sample.MemLimit != nil {
			item.LimiteMem = model.Measured("limite_memoria", "mb", *sample.MemLimit/(1024*1024))
		}
		if sample.CPUMax != "" {
			item.LimiteCPU = model.Text("limite_cpu", model.OrigemMedido, sample.CPUMax)
		}
		out = append(out, item)
	}
	return out
}

func readingsDeMapa(m map[string]model.Reading) []model.Reading {
	out := make([]model.Reading, 0, len(m))
	for _, item := range m {
		if item.Nome == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func copyFloatMap(in map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

package collect

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

func TestCPUPercentNaoInventaZeroNaPrimeiraAmostra(t *testing.T) {
	prev := CPUSample{}
	cur := CPUSample{Total: 100, Idle: 80}
	reading, ok := CPUPercent(prev, cur)
	if ok || reading.Origem != model.OrigemIndisponivel || reading.Valor != nil {
		t.Fatalf("leitura=%+v ok=%v", reading, ok)
	}
}

func TestCPUPercentMedeDelta(t *testing.T) {
	prev := CPUSample{Total: 100, Idle: 80}
	cur := CPUSample{Total: 200, Idle: 100}
	reading, ok := CPUPercent(prev, cur)
	if !ok || reading.Origem != model.OrigemMedido || reading.Valor == nil {
		t.Fatalf("leitura=%+v", reading)
	}
	if *reading.Valor < 79 || *reading.Valor > 81 {
		t.Fatalf("pct=%v", *reading.Valor)
	}
}

func TestParseMeminfoELoad(t *testing.T) {
	mem, err := ParseMeminfo(strings.NewReader("MemTotal: 16384000 kB\nMemAvailable: 8000000 kB\nCached: 1000000 kB\nBuffers: 500000 kB\nSwapTotal: 1000 kB\nSwapFree: 1000 kB\n"))
	if err != nil || !mem.OK || mem.TotalMB < 15000 {
		t.Fatalf("mem=%+v err=%v", mem, err)
	}
	a, b, c, err := ParseLoadavg("0.10 0.20 0.30 1/2 3")
	if err != nil || a != 0.10 || b != 0.20 || c != 0.30 {
		t.Fatalf("%v %v %v %v", a, b, c, err)
	}
}

func TestReadHostSemProcFicaIndisponivel(t *testing.T) {
	dir := t.TempDir()
	readings, _ := ReadHost(dir, HostState{}, time.Now())
	for _, item := range readings {
		if item.Origem == model.OrigemMedido && item.Valor != nil && *item.Valor == 0 && item.Nome == "cpu_pct" {
			t.Fatal("cpu zero artificial")
		}
	}
	cpu := findReading(readings, "cpu_pct")
	if cpu.Origem != model.OrigemIndisponivel {
		t.Fatalf("cpu=%+v", cpu)
	}
}

func TestReadHostComFixture(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "proc", "stat"), "cpu  100 0 0 800 0 0 0 0 0 0\n")
	mustWrite(t, filepath.Join(dir, "proc", "loadavg"), "1.00 0.50 0.25 1/100 10\n")
	mustWrite(t, filepath.Join(dir, "proc", "meminfo"), "MemTotal: 1048576 kB\nMemAvailable: 524288 kB\nCached: 1024 kB\nBuffers: 1024 kB\nSwapTotal: 0 kB\nSwapFree: 0 kB\n")
	mustWrite(t, filepath.Join(dir, "proc", "uptime"), "3600.00 1000.00\n")
	mustWrite(t, filepath.Join(dir, "proc", "diskstats"), " 259 0 nvme0n1 1 0 100 0 1 0 50 0 0 0 0 0 0 0 0\n")
	mustWrite(t, filepath.Join(dir, "proc", "net", "dev"), "inter-|   receive\n face |bytes\n  lo: 10 0 0 0 0 0 0 0 10\n eth0: 1000 0 0 0 0 0 0 0 2000\n")
	first, state := ReadHost(dir, HostState{}, time.Now())
	if findReading(first, "cpu_pct").Origem != model.OrigemIndisponivel {
		t.Fatal("primeira cpu deveria ser ausencia")
	}
	if findReading(first, "load1").Origem != model.OrigemMedido {
		t.Fatal("load nao medido")
	}
	mustWrite(t, filepath.Join(dir, "proc", "stat"), "cpu  200 0 0 900 0 0 0 0 0 0\n")
	second, _ := ReadHost(dir, state, state.When.Add(time.Second))
	cpu := findReading(second, "cpu_pct")
	if cpu.Origem != model.OrigemMedido || cpu.Valor == nil {
		t.Fatalf("segunda cpu=%+v", cpu)
	}
}

func TestCgroupAusente(t *testing.T) {
	_, status := ReadCgroups(t.TempDir())
	if status.Origem != model.OrigemIndisponivel {
		t.Fatalf("%+v", status)
	}
}

func TestCgroupLeMemoria(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "sys", "fs", "cgroup", "docker-abc.scope")
	mustWrite(t, filepath.Join(base, "memory.current"), "1048576\n")
	mustWrite(t, filepath.Join(base, "memory.max"), "2097152\n")
	mustWrite(t, filepath.Join(base, "cpu.stat"), "usage_usec 10\n")
	mustWrite(t, filepath.Join(base, "cpu.max"), "100000 100000\n")
	samples, status := ReadCgroups(dir)
	if status.Origem != model.OrigemMedido || len(samples) != 1 || samples[0].MemBytes == nil {
		t.Fatalf("status=%+v samples=%+v", status, samples)
	}
}

func TestDockerSoLeitura(t *testing.T) {
	if AllowDockerPath("/containers/abc/stop") || AllowDockerPath("/containers/abc/exec") {
		t.Fatal("caminho de escrita foi aceito")
	}
	if !AllowDockerPath("/containers/json") || !AllowDockerPath("/containers/abc123/stats") {
		t.Fatal("caminho de leitura foi recusado")
	}
}

func TestContainersNaoInventamServicoAusente(t *testing.T) {
	body := []byte(`[{"Id":"abc123","Names":["/confvision-worker"],"State":"running"}]`)
	containers, status, _ := ContainersFromDocker(context.Background(), func(_ context.Context, path string) ([]byte, error) {
		if path == "/containers/json" {
			return body, nil
		}
		return nil, os.ErrNotExist
	})
	if status.Origem != model.OrigemMedido || len(containers) != 1 {
		t.Fatalf("%+v %d", status, len(containers))
	}
	if containers[0].Papel != "worker" {
		t.Fatal(containers[0].Papel)
	}
	found := map[string]bool{}
	for _, item := range containers {
		found[item.Papel] = true
	}
	if found["sensor"] {
		t.Fatal("sensor ausente foi inventado")
	}
}

func TestParseMetricsNaoZeraAusente(t *testing.T) {
	line := "[METRICS] camera=5 resolucao=1280x720 fps_recebido=15.00 fps_analisado=7.50 yolo_ms_p95=- gate=ocioso infer_descartados=2"
	_, cameras := ParseMetrics(line)
	if len(cameras) != 1 || cameras[0].VisCameraID != 5 {
		t.Fatalf("%+v", cameras)
	}
	if cameras[0].Campos["yolo_ms_p95"].Origem != model.OrigemIndisponivel || cameras[0].Campos["yolo_ms_p95"].Valor != nil {
		t.Fatalf("p95 %+v", cameras[0].Campos["yolo_ms_p95"])
	}
	if cameras[0].Campos["fps_recebido"].Origem != model.OrigemMedido || cameras[0].Campos["fps_recebido"].Valor == nil {
		t.Fatal("fps nao medido")
	}
	if cameras[0].Campos["gate"].Texto != "ocioso" {
		t.Fatal(cameras[0].Campos["gate"])
	}
}

func TestMetricsArquivoAusente(t *testing.T) {
	_, _, status := ReadMetricsFile(filepath.Join(t.TempDir(), "nao-existe"))
	if status.Origem != model.OrigemIndisponivel {
		t.Fatal(status)
	}
}

func TestMediaMTX(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/paths/list" || r.Method != http.MethodGet {
			t.Errorf("path %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"itemCount":1,"items":[{"confName":"cam","ready":true}]}`))
	}))
	defer srv.Close()
	readings := ReadMediaMTX(context.Background(), srv.URL, "", "")
	if findReading(readings, "paths").Origem != model.OrigemMedido {
		t.Fatalf("%+v", readings)
	}
}

func TestRedisSomenteLeitura(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		_, _ = conn.Read(buf)
		payload := "used_memory:100\r\nused_cpu_sys:1.5\r\nconnected_clients:2\r\n"
		_, _ = conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(payload), payload)))
		_, _ = conn.Read(buf)
		_, _ = conn.Write([]byte("+list\r\n"))
		_, _ = conn.Read(buf)
		_, _ = conn.Write([]byte(":4\r\n"))
	}()
	readings := ReadRedis("redis://"+ln.Addr().String()+"/0", []string{"confvision:eventos"})
	if findReading(readings, "used_memory").Origem != model.OrigemMedido {
		t.Fatalf("%+v", readings)
	}
	if findReading(readings, "fila_confvision:eventos").Valor == nil || *findReading(readings, "fila_confvision:eventos").Valor != 4 {
		t.Fatalf("%+v", readings)
	}
}

func TestPapelNaoConfundeTimelapseComOutro(t *testing.T) {
	if PapelDoContainer("confvision-timelapse") != "timelapse" {
		t.Fatal("timelapse")
	}
	if PapelDoContainer("confvision-dvr") != "dvr" {
		t.Fatal("dvr")
	}
	if PapelDoContainer("outro") != "" {
		t.Fatal("papel inventado")
	}
}

func TestNucleosNaoViraramZero(t *testing.T) {
	n, ok := CountCPUCores("cpu  1 0 0 1\ncpu0 1 0 0 1\ncpu1 1 0 0 1\n")
	if !ok || n != 2 {
		t.Fatalf("n=%d ok=%v", n, ok)
	}
	if _, ok := CountCPUCores("cpu  1 0 0 1\n"); ok {
		t.Fatal("linha agregada nao e nucleo")
	}
}

func TestLinkSemVelocidadeFicaAusente(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "sys", "class", "net", "lo", "speed"), "0\n")
	mustWrite(t, filepath.Join(dir, "sys", "class", "net", "eth0", "speed"), "-1\n")
	leitura := ReadLinkMbps(dir)
	if leitura.Origem != model.OrigemIndisponivel || leitura.Valor != nil {
		t.Fatalf("%+v", leitura)
	}
	mustWrite(t, filepath.Join(dir, "sys", "class", "net", "eth0", "speed"), "1000\n")
	leitura = ReadLinkMbps(dir)
	if leitura.Origem != model.OrigemMedido || leitura.Valor == nil || *leitura.Valor != 1000 {
		t.Fatalf("%+v", leitura)
	}
}

func TestGPUAusenteNaoBloqueia(t *testing.T) {
	readings := ReadGPU(t.TempDir())
	gpu := findReading(readings, "gpu_presente")
	if gpu.Texto != "nao" || gpu.Origem != model.OrigemMedido {
		t.Fatalf("%+v", gpu)
	}
	if findReading(readings, "gpu_util").Origem != model.OrigemIndisponivel {
		t.Fatal("utilizacao inventada")
	}
	if UsoGPUNoLog("[YOLO] device=cpu").Texto != "nao" {
		t.Fatal("cpu deveria ficar fora do calculo de gpu")
	}
	if UsoGPUNoLog("sem device").Origem != model.OrigemIndisponivel {
		t.Fatal("device ausente virou valor")
	}
}

func findReading(items []model.Reading, nome string) model.Reading {
	for _, item := range items {
		if item.Nome == nome {
			return item
		}
	}
	return model.Reading{}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

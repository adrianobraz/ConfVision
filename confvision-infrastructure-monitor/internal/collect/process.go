package collect

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"confvision-infrastructure-monitor/internal/model"
)

var relevantComm = map[string]bool{
	"ffmpeg":       true,
	"python":       true,
	"python3":      true,
	"mediamtx":     true,
	"redis-server": true,
}

type ProcSample struct {
	PID   int
	Ticks uint64
}

func ReadProcesses(root string, prev map[int]ProcSample, seconds float64) ([]model.Process, map[int]ProcSample, model.Reading) {
	procRoot := filepath.Join(root, "proc")
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil, prev, model.Unavailable("processos", "proc indisponivel")
	}
	next := map[int]ProcSample{}
	var out []model.Process
	clk := 100.0
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		commRaw, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "comm"))
		if err != nil {
			continue
		}
		comm := strings.TrimSpace(string(commRaw))
		if !relevantComm[comm] {
			continue
		}
		statRaw, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "stat"))
		if err != nil {
			continue
		}
		ticks := procTicks(string(statRaw))
		next[pid] = ProcSample{PID: pid, Ticks: ticks}
		cpu := model.Unavailable("cpu", "primeira amostra")
		if old, ok := prev[pid]; ok && seconds > 0 && ticks >= old.Ticks {
			cpu = model.Measured("cpu", "pct", float64(ticks-old.Ticks)/clk/seconds*100)
		}
		mem := model.Unavailable("memoria", "status ilegivel")
		if status, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "status")); err == nil {
			if kb, ok := statusKB(string(status), "VmRSS:"); ok {
				mem = model.Measured("memoria", "mb", kb/1024)
			}
		}
		container := ""
		if cg, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "cgroup")); err == nil {
			container = cgroupHint(string(cg))
		}
		out = append(out, model.Process{
			PID:                 pid,
			Nome:                comm,
			Container:           container,
			CPU:                 cpu,
			Memoria:             mem,
			IncluidoNoContainer: container != "",
		})
	}
	return out, next, model.Text("processos", model.OrigemMedido, "lido")
}

func procTicks(stat string) uint64 {
	// comm pode ter espaço e fica entre parênteses. Os campos úteis começam depois.
	end := strings.LastIndex(stat, ")")
	if end < 0 || end+2 >= len(stat) {
		return 0
	}
	fields := strings.Fields(stat[end+2:])
	if len(fields) < 13 {
		return 0
	}
	ut, _ := strconv.ParseUint(fields[11], 10, 64)
	st, _ := strconv.ParseUint(fields[12], 10, 64)
	return ut + st
}

func statusKB(status, key string) (float64, bool) {
	for _, line := range strings.Split(status, "\n") {
		if !strings.HasPrefix(line, key) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}
		n, err := strconv.ParseFloat(fields[1], 64)
		return n, err == nil
	}
	return 0, false
}

func cgroupHint(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "docker") || strings.Contains(line, "containerd") {
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}

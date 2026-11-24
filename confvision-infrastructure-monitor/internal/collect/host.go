package collect

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

type diskSpaceFunc func(path string) (total, free uint64, err error)

// DiskSpace lê o volume raiz quando o sistema expõe statfs.
var DiskSpace diskSpaceFunc

type HostState struct {
	CPU  CPUSample
	IO   IOSample
	Net  NetSample
	When time.Time
	Seen bool
}

func ReadHost(root string, prev HostState, now time.Time) ([]model.Reading, HostState) {
	next := prev
	next.When = now
	var out []model.Reading
	seconds := now.Sub(prev.When).Seconds()

	if raw, err := os.ReadFile(filepath.Join(root, "proc", "stat")); err != nil {
		out = append(out, model.Unavailable("cpu_pct", "proc/stat indisponivel"))
	} else if sample, err := ParseCPUStat(strings.NewReader(string(raw))); err != nil {
		out = append(out, model.Unavailable("cpu_pct", "proc/stat ilegivel"))
	} else {
		if prev.Seen {
			reading, _ := CPUPercent(prev.CPU, sample)
			out = append(out, reading)
		} else {
			out = append(out, model.Unavailable("cpu_pct", "primeira amostra"))
		}
		next.CPU = sample
	}

	if raw, err := os.ReadFile(filepath.Join(root, "proc", "loadavg")); err != nil {
		out = append(out, model.Unavailable("load1", "proc/loadavg indisponivel"))
	} else if a, b, c, err := ParseLoadavg(string(raw)); err != nil {
		out = append(out, model.Unavailable("load1", "proc/loadavg ilegivel"))
	} else {
		out = append(out,
			model.Measured("load1", "load", a),
			model.Measured("load5", "load", b),
			model.Measured("load15", "load", c),
		)
	}

	if raw, err := os.ReadFile(filepath.Join(root, "proc", "meminfo")); err != nil {
		out = append(out, model.Unavailable("ram_total_mb", "proc/meminfo indisponivel"))
	} else if mem, err := ParseMeminfo(strings.NewReader(string(raw))); err != nil || !mem.OK {
		out = append(out, model.Unavailable("ram_total_mb", "proc/meminfo ilegivel"))
	} else {
		out = append(out,
			model.Measured("ram_total_mb", "mb", mem.TotalMB),
			model.Measured("ram_disponivel_mb", "mb", mem.AvailableMB),
			model.Measured("ram_usada_mb", "mb", mem.TotalMB-mem.AvailableMB),
			model.Measured("ram_cache_mb", "mb", mem.CacheMB),
			model.Measured("swap_total_mb", "mb", mem.SwapTotalMB),
			model.Measured("swap_livre_mb", "mb", mem.SwapFreeMB),
		)
	}

	if raw, err := os.ReadFile(filepath.Join(root, "proc", "uptime")); err != nil {
		out = append(out, model.Unavailable("uptime_seg", "proc/uptime indisponivel"))
	} else if sec, err := ParseUptime(string(raw)); err != nil {
		out = append(out, model.Unavailable("uptime_seg", "proc/uptime ilegivel"))
	} else {
		out = append(out, model.Measured("uptime_seg", "seg", sec))
	}

	if raw, err := os.ReadFile(filepath.Join(root, "proc", "diskstats")); err != nil {
		out = append(out, model.Unavailable("disco_leitura_bps", "proc/diskstats indisponivel"))
	} else if sample, err := ParseDiskstats(strings.NewReader(string(raw))); err != nil {
		out = append(out, model.Unavailable("disco_leitura_bps", "proc/diskstats ilegivel"))
	} else {
		if prev.Seen {
			out = append(out,
				RatePerSec("disco_leitura_bps", "Bps", prev.IO.ReadBytes, sample.ReadBytes, seconds),
				RatePerSec("disco_escrita_bps", "Bps", prev.IO.WriteBytes, sample.WriteBytes, seconds),
			)
		} else {
			out = append(out,
				model.Unavailable("disco_leitura_bps", "primeira amostra"),
				model.Unavailable("disco_escrita_bps", "primeira amostra"),
			)
		}
		next.IO = sample
	}

	if raw, err := os.ReadFile(filepath.Join(root, "proc", "net", "dev")); err != nil {
		out = append(out, model.Unavailable("rede_rx_bps", "proc/net/dev indisponivel"))
	} else if sample, err := ParseNetDev(strings.NewReader(string(raw))); err != nil {
		out = append(out, model.Unavailable("rede_rx_bps", "proc/net/dev ilegivel"))
	} else {
		if prev.Seen {
			out = append(out,
				RatePerSec("rede_rx_bps", "Bps", prev.Net.RX, sample.RX, seconds),
				RatePerSec("rede_tx_bps", "Bps", prev.Net.TX, sample.TX, seconds),
			)
		} else {
			out = append(out,
				model.Unavailable("rede_rx_bps", "primeira amostra"),
				model.Unavailable("rede_tx_bps", "primeira amostra"),
			)
		}
		next.Net = sample
	}

	if DiskSpace == nil {
		out = append(out, model.Unavailable("disco_total_mb", "statfs indisponivel neste sistema"))
	} else if total, free, err := DiskSpace("/"); err != nil || total == 0 {
		out = append(out, model.Unavailable("disco_total_mb", "statfs sem leitura"))
	} else {
		totalMB := float64(total) / (1024 * 1024)
		freeMB := float64(free) / (1024 * 1024)
		out = append(out,
			model.Measured("disco_total_mb", "mb", totalMB),
			model.Measured("disco_livre_mb", "mb", freeMB),
			model.Measured("disco_usado_mb", "mb", totalMB-freeMB),
		)
	}

	tempPath := filepath.Join(root, "sys", "class", "thermal", "thermal_zone0", "temp")
	if raw, err := os.ReadFile(tempPath); err != nil {
		out = append(out, model.Unavailable("temperatura_c", "sensor termico ausente"))
	} else if n, err := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64); err != nil {
		out = append(out, model.Unavailable("temperatura_c", "sensor termico ilegivel"))
	} else {
		out = append(out, model.Measured("temperatura_c", "c", n/1000))
	}

	if _, err := os.Stat(filepath.Join(root, "proc", "driver", "nvidia")); err != nil {
		out = append(out, model.Unavailable("gpu", "gpu nao exposta em proc"))
	} else {
		out = append(out, model.Unavailable("gpu_vram", "gpu presente sem leitura de vram"))
	}

	next.Seen = true
	return out, next
}

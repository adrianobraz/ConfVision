package collect

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"confvision-infrastructure-monitor/internal/model"
)

type CPUSample struct {
	Total uint64
	Idle  uint64
}

type MemSample struct {
	TotalMB     float64
	AvailableMB float64
	CacheMB     float64
	SwapTotalMB float64
	SwapFreeMB  float64
	OK          bool
}

type IOSample struct {
	ReadBytes  uint64
	WriteBytes uint64
}

type NetSample struct {
	RX uint64
	TX uint64
}

func ParseCPUStat(r io.Reader) (CPUSample, error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return CPUSample{}, io.ErrUnexpectedEOF
		}
		var total, idle uint64
		for i := 1; i < len(fields); i++ {
			n, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				return CPUSample{}, err
			}
			total += n
			if i == 4 || i == 5 {
				idle += n
			}
		}
		return CPUSample{Total: total, Idle: idle}, nil
	}
	return CPUSample{}, io.EOF
}

// CountCPUCores conta as linhas cpu0, cpu1, ... de /proc/stat.
// A linha agregada "cpu" não entra. Zero não é um núcleo medido.
func CountCPUCores(text string) (int, bool) {
	n := 0
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		if len(name) < 4 || !strings.HasPrefix(name, "cpu") {
			continue
		}
		digit := true
		for _, r := range name[3:] {
			if r < '0' || r > '9' {
				digit = false
				break
			}
		}
		if digit {
			n++
		}
	}
	return n, n > 0
}

func CPUPercent(prev, cur CPUSample) (model.Reading, bool) {
	if prev.Total == 0 || cur.Total <= prev.Total {
		return model.Unavailable("cpu_pct", "primeira amostra ou contador sem avanco"), false
	}
	total := float64(cur.Total - prev.Total)
	idle := float64(cur.Idle - prev.Idle)
	if total <= 0 {
		return model.Unavailable("cpu_pct", "contador de cpu sem delta"), false
	}
	pct := (1 - idle/total) * 100
	if pct < 0 {
		pct = 0
	}
	return model.Measured("cpu_pct", "pct", pct), true
}

func ParseLoadavg(text string) (float64, float64, float64, error) {
	fields := strings.Fields(text)
	if len(fields) < 3 {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}
	a, err1 := strconv.ParseFloat(fields[0], 64)
	b, err2 := strconv.ParseFloat(fields[1], 64)
	c, err3 := strconv.ParseFloat(fields[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}
	return a, b, c, nil
}

func ParseMeminfo(r io.Reader) (MemSample, error) {
	vals := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		n, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}
		vals[key] = n / 1024
	}
	if vals["MemTotal"] <= 0 {
		return MemSample{}, io.EOF
	}
	return MemSample{
		TotalMB:     vals["MemTotal"],
		AvailableMB: vals["MemAvailable"],
		CacheMB:     vals["Cached"] + vals["Buffers"],
		SwapTotalMB: vals["SwapTotal"],
		SwapFreeMB:  vals["SwapFree"],
		OK:          true,
	}, sc.Err()
}

func ParseUptime(text string) (float64, error) {
	fields := strings.Fields(text)
	if len(fields) < 1 {
		return 0, io.ErrUnexpectedEOF
	}
	return strconv.ParseFloat(fields[0], 64)
}

func ParseDiskstats(r io.Reader) (IOSample, error) {
	sc := bufio.NewScanner(r)
	var readSectors, writeSectors uint64
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 14 {
			continue
		}
		name := fields[2]
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "dm-") || isPartition(name) {
			continue
		}
		rs, err1 := strconv.ParseUint(fields[5], 10, 64)
		ws, err2 := strconv.ParseUint(fields[9], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		readSectors += rs
		writeSectors += ws
	}
	return IOSample{ReadBytes: readSectors * 512, WriteBytes: writeSectors * 512}, sc.Err()
}

func isPartition(name string) bool {
	if strings.HasPrefix(name, "nvme") {
		return strings.Contains(name, "p")
	}
	if strings.HasPrefix(name, "dm-") || name == "" {
		return false
	}
	last := name[len(name)-1]
	return last >= '0' && last <= '9'
}

func ParseNetDev(r io.Reader) (NetSample, error) {
	sc := bufio.NewScanner(r)
	var rx, tx uint64
	lineNo := 0
	for sc.Scan() {
		lineNo++
		if lineNo <= 2 {
			continue
		}
		parts := strings.Split(sc.Text(), ":")
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		if name == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rxi, err1 := strconv.ParseUint(fields[0], 10, 64)
		txi, err2 := strconv.ParseUint(fields[8], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		rx += rxi
		tx += txi
	}
	return NetSample{RX: rx, TX: tx}, sc.Err()
}

func RatePerSec(nome, unidade string, prev, cur uint64, seconds float64) model.Reading {
	if seconds <= 0 || cur < prev || (prev == 0 && cur == 0) {
		return model.Unavailable(nome, "primeira amostra ou contador sem avanco")
	}
	if prev == 0 {
		return model.Unavailable(nome, "primeira amostra")
	}
	return model.Measured(nome, unidade, float64(cur-prev)/seconds)
}

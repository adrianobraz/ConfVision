package collect

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"confvision-infrastructure-monitor/internal/model"
)

type CgroupSample struct {
	ID         string
	MemBytes   *float64
	MemLimit   *float64
	CPUUsageUS *float64
	CPUMax     string
}

func ReadCgroups(root string) ([]CgroupSample, model.Reading) {
	base := filepath.Join(root, "sys", "fs", "cgroup")
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		return nil, model.Unavailable("cgroup", "sysfs de cgroup ausente")
	}
	var found []CgroupSample
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.Contains(name, "docker") && !strings.HasPrefix(name, "cri-containerd") {
			return nil
		}
		sample := CgroupSample{ID: name}
		if raw, err := os.ReadFile(filepath.Join(path, "memory.current")); err == nil {
			if n, err := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64); err == nil {
				sample.MemBytes = &n
			}
		}
		if raw, err := os.ReadFile(filepath.Join(path, "memory.max")); err == nil {
			text := strings.TrimSpace(string(raw))
			if text != "max" {
				if n, err := strconv.ParseFloat(text, 64); err == nil {
					sample.MemLimit = &n
				}
			}
		}
		if raw, err := os.ReadFile(filepath.Join(path, "cpu.stat")); err == nil {
			for _, line := range strings.Split(string(raw), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 2 && fields[0] == "usage_usec" {
					if n, err := strconv.ParseFloat(fields[1], 64); err == nil {
						sample.CPUUsageUS = &n
					}
				}
			}
		}
		if raw, err := os.ReadFile(filepath.Join(path, "cpu.max")); err == nil {
			sample.CPUMax = strings.TrimSpace(string(raw))
		}
		if sample.MemBytes != nil || sample.CPUUsageUS != nil {
			found = append(found, sample)
		}
		return nil
	})
	if len(found) == 0 {
		return nil, model.Text("cgroup", model.OrigemMedido, "nenhum container no cgroup visivel")
	}
	return found, model.Text("cgroup", model.OrigemMedido, "lido")
}

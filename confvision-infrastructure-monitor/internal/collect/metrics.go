package collect

import (
	"bufio"
	"encoding/binary"
	"os"
	"strconv"
	"strings"

	"confvision-infrastructure-monitor/internal/model"
)

// ParseMetrics lê linhas [METRICS] já emitidas pelo worker.
// Campo ausente, "-" ou "indisponivel" não vira zero.
func ParseMetrics(text string) (map[string]model.Reading, []model.Camera) {
	process := map[string]model.Reading{}
	var cameras []model.Camera
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		idx := strings.Index(line, "[METRICS]")
		if idx < 0 {
			continue
		}
		line = strings.TrimSpace(line[idx+len("[METRICS]"):])
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "processo" {
			for _, field := range fields[1:] {
				key, value, ok := splitKV(field)
				if !ok {
					continue
				}
				process[key] = readingFromToken(key, value)
			}
			continue
		}
		if !strings.HasPrefix(fields[0], "camera=") {
			continue
		}
		id, err := strconv.Atoi(strings.TrimPrefix(fields[0], "camera="))
		if err != nil || id <= 0 {
			continue
		}
		cam := model.Camera{VisCameraID: id, Campos: map[string]model.Reading{}}
		for _, field := range fields[1:] {
			key, value, ok := splitKV(field)
			if !ok {
				continue
			}
			cam.Campos[key] = readingFromToken(key, value)
		}
		cameras = append(cameras, cam)
	}
	return process, cameras
}

func ReadMetricsFile(path string) (map[string]model.Reading, []model.Camera, model.Reading) {
	if strings.TrimSpace(path) == "" {
		return nil, nil, model.Unavailable("metrics_log", "METRICS_LOG_PATH nao configurado")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, model.Unavailable("metrics_log", "arquivo de log indisponivel")
	}
	process, cameras := ParseMetrics(string(raw))
	if len(process) == 0 && len(cameras) == 0 {
		return process, cameras, model.Unavailable("metrics_log", "nenhuma linha [METRICS] no arquivo")
	}
	return process, cameras, model.Text("metrics_log", model.OrigemMedido, "lido")
}

func DecodeDockerLog(raw []byte) string {
	if len(raw) < 8 {
		return string(raw)
	}
	var b strings.Builder
	i := 0
	frames := 0
	for i+8 <= len(raw) {
		size := int(binary.BigEndian.Uint32(raw[i+4 : i+8]))
		if size < 0 || i+8+size > len(raw) {
			break
		}
		b.Write(raw[i+8 : i+8+size])
		i += 8 + size
		frames++
	}
	if frames == 0 {
		return string(raw)
	}
	return b.String()
}

func splitKV(field string) (string, string, bool) {
	key, value, ok := strings.Cut(field, "=")
	if !ok || key == "" {
		return "", "", false
	}
	return key, value, true
}

func readingFromToken(key, value string) model.Reading {
	switch value {
	case "", "-", "indisponivel", "desconhecida":
		return model.Unavailable(key, "worker nao informou o valor")
	}
	if n, err := strconv.ParseFloat(value, 64); err == nil {
		return model.Measured(key, "", n)
	}
	return model.Text(key, model.OrigemMedido, value)
}

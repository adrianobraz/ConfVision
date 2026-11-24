package collect

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

// ReadRedis só envia INFO, TYPE e a contagem do tipo encontrado.
func ReadRedis(addr string, keys []string) []model.Reading {
	if strings.TrimSpace(addr) == "" {
		return []model.Reading{model.Unavailable("redis", "REDIS_URL nao configurado")}
	}
	u, err := url.Parse(addr)
	if err != nil || u.Host == "" {
		return []model.Reading{model.Unavailable("redis", "REDIS_URL invalida")}
	}
	conn, err := net.DialTimeout("tcp", u.Host, 2*time.Second)
	if err != nil {
		return []model.Reading{model.Unavailable("redis", "redis sem resposta")}
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if u.User != nil {
		if pass, ok := u.User.Password(); ok {
			if err := redisCmd(conn, "AUTH", pass); err != nil {
				return []model.Reading{model.Unavailable("redis", "autenticacao recusada")}
			}
			_, _ = readRedis(conn)
		}
	}
	if err := redisCmd(conn, "INFO"); err != nil {
		return []model.Reading{model.Unavailable("redis", "INFO recusado")}
	}
	info, err := readRedis(conn)
	if err != nil {
		return []model.Reading{model.Unavailable("redis", "INFO ilegivel")}
	}
	out := []model.Reading{model.Text("redis", model.OrigemMedido, "disponivel")}
	out = append(out, redisInfoReadings(info)...)
	for _, key := range keys {
		if key == "" {
			continue
		}
		if err := redisCmd(conn, "TYPE", key); err != nil {
			out = append(out, model.Unavailable("fila_"+key, "TYPE recusado"))
			continue
		}
		kind, err := readRedis(conn)
		if err != nil {
			out = append(out, model.Unavailable("fila_"+key, "TYPE ilegivel"))
			continue
		}
		kind = strings.TrimSpace(kind)
		cmd := ""
		switch kind {
		case "list":
			cmd = "LLEN"
		case "zset":
			cmd = "ZCARD"
		case "set":
			cmd = "SCARD"
		case "none":
			out = append(out, model.Unavailable("fila_"+key, "chave ausente"))
			continue
		default:
			out = append(out, model.Unavailable("fila_"+key, "tipo "+kind+" sem contagem de leitura"))
			continue
		}
		if err := redisCmd(conn, cmd, key); err != nil {
			out = append(out, model.Unavailable("fila_"+key, "contagem recusada"))
			continue
		}
		raw, err := readRedis(conn)
		if err != nil {
			out = append(out, model.Unavailable("fila_"+key, "contagem ilegivel"))
			continue
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			out = append(out, model.Unavailable("fila_"+key, "contagem ilegivel"))
			continue
		}
		out = append(out, model.Measured("fila_"+key, "count", n))
	}
	return out
}

func redisInfoReadings(info string) []model.Reading {
	want := map[string]string{
		"used_memory":       "bytes",
		"used_cpu_sys":      "seg",
		"used_cpu_user":     "seg",
		"connected_clients": "count",
	}
	var out []model.Reading
	for _, line := range strings.Split(info, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		unit, known := want[key]
		if !known {
			continue
		}
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			out = append(out, model.Unavailable(key, "valor ilegivel"))
			continue
		}
		out = append(out, model.Measured(key, unit, n))
	}
	return out
}

func redisCmd(conn net.Conn, args ...string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(arg), arg)
	}
	_, err := ioWrite(conn, b.String())
	return err
}

func ioWrite(conn net.Conn, s string) (int, error) {
	return conn.Write([]byte(s))
}

func readRedis(conn net.Conn) (string, error) {
	r := bufio.NewReader(conn)
	prefix, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	switch prefix {
	case '+', ':':
		return line, nil
	case '-':
		return "", fmt.Errorf("redis: %s", line)
	case '$':
		n, err := strconv.Atoi(line)
		if err != nil {
			return "", err
		}
		if n < 0 {
			return "", nil
		}
		buf := make([]byte, n+2)
		if _, err := ioReadFull(r, buf); err != nil {
			return "", err
		}
		return string(buf[:n]), nil
	default:
		return "", fmt.Errorf("prefixo redis %q", prefix)
	}
}

func ioReadFull(r *bufio.Reader, buf []byte) (int, error) {
	n := 0
	for n < len(buf) {
		read, err := r.Read(buf[n:])
		n += read
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

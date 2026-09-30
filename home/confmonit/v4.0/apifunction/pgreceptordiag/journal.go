package pgreceptordiag

import (
	"regexp"
	"strings"
)

var (
	reJournalConta  = regexp.MustCompile(`(?i)\[CONTA:\s*([^\]]+)\]`)
	reJournalIMEI   = regexp.MustCompile(`\[([A-F0-9]{6})\]`)
	reJournalEvent  = regexp.MustCompile(`->\s*(.+)$`)
	reJournalTime   = regexp.MustCompile(`(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})`)
	reJournalDriver = regexp.MustCompile(`\[\s*([A-Za-z0-9_-]+)\s*:`)
)

func ParseJournalLines(lines []string) []JournalEvent {
	var out []JournalEvent
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		ev := JournalEvent{Linha: ln}
		if m := reJournalTime.FindStringSubmatch(ln); len(m) > 1 {
			ev.DataHora = m[1]
		}
		if m := reJournalDriver.FindStringSubmatch(ln); len(m) > 1 {
			ev.Driver = strings.ToUpper(strings.TrimSpace(m[1]))
		}
		if m := reJournalConta.FindStringSubmatch(ln); len(m) > 1 {
			ev.Conta = strings.TrimSpace(m[1])
		}
		if m := reJournalIMEI.FindStringSubmatch(ln); len(m) > 1 {
			ev.IMEI = strings.ToUpper(m[1])
		}
		if m := reJournalEvent.FindStringSubmatch(ln); len(m) > 1 {
			ev.Evento = simplifyEvento(strings.TrimSpace(m[1]))
		}
		out = append(out, ev)
	}
	return out
}

func enrichJournalDrivers(eventos []JournalEvent, modulo string) []JournalEvent {
	mod := strings.ToUpper(strings.TrimSpace(modulo))
	for i := range eventos {
		if strings.TrimSpace(eventos[i].Driver) == "" && mod != "" {
			eventos[i].Driver = mod
		}
	}
	return eventos
}

func simplifyEvento(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "hardbyte"):
		return "Keep-alive OK"
	case strings.HasPrefix(lower, "evento["):
		return "Alarme processado — " + s
	case lower == "bloqueado":
		return "Bloqueado (filtro do sistema)"
	default:
		return s
	}
}

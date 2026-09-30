package pgreceptordiag

import (
	"net"
	"strings"

	"apifunction/config"
)

func isServerIP(ip string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false
	}
	if ip == "127.0.0.1" || ip == "::1" || ip == "0.0.0.0" {
		return true
	}
	server := strings.TrimSpace(config.ReceptorDNSIP)
	return server != "" && ip == server
}

// formatEstablishedPeer extrai o IP publico da central (lado remoto da conexao TCP).
func formatEstablishedPeer(peer string) string {
	peer = strings.TrimSpace(peer)
	if peer == "" {
		return ""
	}
	if i := strings.Index(peer, "->"); i >= 0 {
		remote := strings.TrimSpace(peer[i+2:])
		host := hostFromAddr(remote)
		if host != "" && !isServerIP(host) {
			return host
		}
		local := strings.TrimSpace(peer[:i])
		host = hostFromAddr(local)
		if host != "" && !isServerIP(host) {
			return host
		}
		return ""
	}
	return SanitizeIPRemoto(peer)
}

func hostFromAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	if i := strings.LastIndex(addr, ":"); i > 0 {
		possible := addr[:i]
		if net.ParseIP(possible) != nil {
			return possible
		}
	}
	if net.ParseIP(addr) != nil {
		return addr
	}
	return ""
}

// SanitizeIPRemoto remove IP/porta do servidor receptor; retorna so o IP da central.
func SanitizeIPRemoto(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	host := hostFromAddr(s)
	if host == "" {
		return s
	}
	if isServerIP(host) {
		return ""
	}
	return host
}

func sanitizeErroConexaoRows(rows []ErroConexaoRow) []ErroConexaoRow {
	out := make([]ErroConexaoRow, len(rows))
	for i, e := range rows {
		e.IPRemoto = SanitizeIPRemoto(e.IPRemoto)
		out[i] = e
	}
	return out
}

func DiagErroConexaoLimite() int {
	return diagErroConexaoLimite
}

func DiagErroConexaoPageSize() int {
	return diagErroConexaoPageSize
}

func DiagErroConexaoMaxPage() int {
	return diagErroConexaoMaxPage
}

func SanitizeErroConexaoList(rows []ErroConexaoRow) []ErroConexaoRow {
	return sanitizeErroConexaoRows(rows)
}

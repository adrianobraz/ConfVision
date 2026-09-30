package handler

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"apifunction/config"
	"apifunction/pgreceptordiag"
)

func clientIP(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		v := strings.TrimSpace(r.Header.Get(h))
		if v == "" {
			continue
		}
		if i := strings.Index(v, ","); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		if net.ParseIP(v) != nil {
			return v
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func checkDiagLabKey(r *http.Request) bool {
	key := strings.TrimSpace(config.ReceptorDiagLabKey)
	if key == "" {
		return true
	}
	q := strings.TrimSpace(r.URL.Query().Get("key"))
	h := strings.TrimSpace(r.Header.Get("X-Diag-Lab-Key"))
	return q == key || h == key
}

func (h *Handler) ReceptorDiagFabricantes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	lista, err := pgreceptordiag.ListarFabricantesPortas()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ReceptorDiagFranqueados(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	lab := strings.EqualFold(r.URL.Query().Get("lab"), "1")
	if lab {
		if !checkDiagLabKey(r) {
			writeErr(w, http.StatusUnauthorized, "chave lab invalida")
			return
		}
	} else if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	lista, err := pgreceptordiag.ListarFranqueados(200)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ReceptorDiagVerificar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req pgreceptordiag.ReqVerificar
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	lab := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("lab")), "1") ||
		strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("lab")), "true") ||
		req.Lab
	if lab {
		if !checkDiagLabKey(r) {
			writeErr(w, http.StatusUnauthorized, "chave lab invalida")
			return
		}
		req.Lab = true
	} else if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	if strings.TrimSpace(req.IPTecnico) == "" {
		req.IPTecnico = clientIP(r)
	}
	res, err := pgreceptordiag.Verificar(req)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func queryInt(r *http.Request, key string, def, max int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	if max > 0 && n > max {
		return max
	}
	return n
}

func (h *Handler) ReceptorDiagErroConexao(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	lab := strings.EqualFold(r.URL.Query().Get("lab"), "1")
	if lab {
		if !checkDiagLabKey(r) {
			writeErr(w, http.StatusUnauthorized, "chave lab invalida")
			return
		}
	} else if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	fab := strings.TrimSpace(r.URL.Query().Get("fabricante"))
	if fab == "" {
		fab = strings.TrimSpace(r.URL.Query().Get("modulo"))
	}
	if fab == "" {
		writeErr(w, http.StatusBadRequest, "fabricante ou modulo obrigatorio")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	conta := strings.TrimSpace(r.URL.Query().Get("conta"))
	codigo := strings.TrimSpace(r.URL.Query().Get("codigo"))
	limit := pgreceptordiag.DiagErroConexaoLimite()
	offset := 0
	if lab {
		limit = queryInt(r, "limit", pgreceptordiag.DiagErroConexaoPageSize(), pgreceptordiag.DiagErroConexaoMaxPage())
		offset = queryInt(r, "offset", 0, 0)
	}
	lista, err := pgreceptordiag.ListarErroConexao(idFra, fab, conta, codigo, limit, offset)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if lista == nil {
		lista = []pgreceptordiag.ErroConexaoRow{}
	}
	lista = pgreceptordiag.SanitizeErroConexaoList(lista)
	lista = pgreceptordiag.EnrichErroConexaoMatch(idFra, fab, conta, codigo, lista)
	resp := map[string]any{"ok": true, "dados": lista}
	if lab {
		totalHoje, porFab, _ := pgreceptordiag.ContagemErroConexaoHoje(fab)
		carregados := offset + len(lista)
		meta := map[string]any{
			"fabricante":          pgreceptordiag.ModuloForFabricante(fab),
			"limit":               limit,
			"offset":              offset,
			"carregados":          carregados,
			"totalHoje":           totalHoje,
			"totalFabricanteHoje": porFab,
			"hasMore":             carregados < porFab,
		}
		resp["meta"] = meta
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ReceptorDiagSimularCadastro(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	lab := strings.EqualFold(r.URL.Query().Get("lab"), "1")
	if lab {
		if !checkDiagLabKey(r) {
			writeErr(w, http.StatusUnauthorized, "chave lab invalida")
			return
		}
	} else if !checkWorkerKey(r, "") {
		writeErr(w, http.StatusUnauthorized, "nao autorizado")
		return
	}
	fab := strings.TrimSpace(r.URL.Query().Get("fabricante"))
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	conta := strings.TrimSpace(r.URL.Query().Get("conta"))
	codigo := strings.TrimSpace(r.URL.Query().Get("codigo"))
	res, err := pgreceptordiag.SimularCadastro(idFra, fab, conta, codigo)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) ReceptorDiagPresenca(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if !checkWorkerKey(r, "") {
			writeErr(w, http.StatusUnauthorized, "nao autorizado")
			return
		}
		var req pgreceptordiag.ReqPresenca
		if err := decodeJSONPermissive(r, &req); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := pgreceptordiag.UpsertPresenca(req); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	case http.MethodGet:
		lab := strings.EqualFold(r.URL.Query().Get("lab"), "1")
		if lab {
			if !checkDiagLabKey(r) {
				writeErr(w, http.StatusUnauthorized, "chave lab invalida")
				return
			}
		} else if !checkWorkerKey(r, "") {
			writeErr(w, http.StatusUnauthorized, "nao autorizado")
			return
		}
		lista, err := pgreceptordiag.ListarPresenca(
			r.URL.Query().Get("id_franqueado"),
			r.URL.Query().Get("fabricante"),
			50,
		)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
	default:
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

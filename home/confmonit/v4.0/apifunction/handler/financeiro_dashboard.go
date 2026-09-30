package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

	"apifunction/auth"
	"apifunction/config"
	"apifunction/pgfinmirror"
	"apifunction/pggovernanca"
	"apifunction/service/finresumo"
	"apifunction/xano"
)

func (h *Handler) FinanceiroFinMirrorSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey string `json:"worker_key"`
	}
	_ = decodeJSONPermissive(r, &req)
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	res, err := pgfinmirror.SyncFromMeta(r.Context())
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) FinanceiroFinMirrorHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey string                  `json:"worker_key"`
		Entidades []pgfinmirror.HookEntity `json:"entidades"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	if len(req.Entidades) == 0 {
		writeErr(w, http.StatusBadRequest, "entidades vazio")
		return
	}
	res, err := pgfinmirror.ApplyHook(r.Context(), req.Entidades)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) FinanceiroDashboardKPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		Competencia  string `json:"competencia"`
		AdminUsuario string `json:"admin_usuario"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil && r.ContentLength > 0 {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}

	ready, motivo, _ := pgfinmirror.IsReady(r.Context())
	if !ready {
		writeErr(w, http.StatusServiceUnavailable, motivo)
		return
	}

	kpi, err := finresumo.ComputeDashboard(r.Context(), sess, req.Competencia, strings.TrimSpace(req.AdminUsuario))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	gov := governancaAlerta(sess)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"dados": mergeKPIComGovernanca(kpi, gov),
		"fonte": "postgres_mirror",
	})
}

func (h *Handler) FinanceiroDashboardKPIShadow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey    string `json:"worker_key"`
		Competencia  string `json:"competencia"`
		AdminToken   string `json:"admin_token"`
		AdminUsuario string `json:"admin_usuario"`
		IDCentral    string `json:"id_central"`
		IDRep        string `json:"id_representante"`
		UserTipo     string `json:"user_tipo"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}

	ready, motivo, _ := pgfinmirror.IsReady(r.Context())
	if !ready {
		writeErr(w, http.StatusServiceUnavailable, motivo)
		return
	}

	sess := auth.SessaoAdm{
		UserTipo:          strings.ToUpper(strings.TrimSpace(req.UserTipo)),
		IDRepresentante:   strings.TrimSpace(req.IDRep),
		IDCentralUUID:     strings.TrimSpace(req.IDCentral),
		IDCentralCatalogo: strings.TrimSpace(req.IDCentral),
	}
	if sess.UserTipo == "" {
		sess.UserTipo = "CEN"
	}

	pgKPI, err := finresumo.ComputeDashboard(r.Context(), sess, req.Competencia, strings.TrimSpace(req.AdminUsuario))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	xanoKPI, err := fetchXanoDashboard(req.AdminToken, req.Competencia)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "xano: "+err.Error())
		return
	}

	diffs := compareKPI(pgKPI, xanoKPI)
	escopoID := sess.IDCentralCatalogo
	escopoTipo := "CEN"
	if sess.UserTipo == "REP" {
		escopoTipo = "REP"
		escopoID = sess.IDRepresentante
	}
	for _, d := range diffs {
		_ = pgfinmirror.LogShadowDiff(r.Context(), escopoTipo, escopoID, req.Competencia, d.Campo, d.Postgres, d.Xano)
	}
	if len(diffs) > 0 {
		log.Printf("[fin-shadow] %d diffs escopo=%s/%s competencia=%s", len(diffs), escopoTipo, escopoID, req.Competencia)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"diffs":     diffs,
		"diff_count": len(diffs),
		"postgres":  pgKPI,
		"xano":      xanoKPI,
	})
}

type kpiDiff struct {
	Campo    string `json:"campo"`
	Postgres string `json:"postgres"`
	Xano     string `json:"xano"`
}

func compareKPI(pg finresumo.DashboardKPI, xano map[string]any) []kpiDiff {
	campos := []struct {
		key string
		pg  float64
	}{
		{"valor_em_aberto", pg.ValorEmAberto},
		{"valor_pago", pg.ValorPago},
		{"valor_vencido", pg.ValorVencido},
		{"receita_periodo", pg.ReceitaPeriodo},
		{"mrr_total", pg.MRRTotal},
		{"saldo_caixa", pg.SaldoCaixa},
		{"faturas_abertas", float64(pg.FaturasAbertas)},
		{"assinaturas_ativas", float64(pg.AssinaturasAtivas)},
		{"qtd_inadimplentes", float64(pg.QtdInadimplentes)},
	}
	var diffs []kpiDiff
	for _, c := range campos {
		xv := floatFromMap(xano, c.key)
		if math.Abs(c.pg-xv) > 0.02 {
			diffs = append(diffs, kpiDiff{
				Campo:    c.key,
				Postgres: fmt.Sprintf("%.2f", c.pg),
				Xano:     fmt.Sprintf("%.2f", xv),
			})
		}
	}
	return diffs
}

func floatFromMap(m map[string]any, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func fetchXanoDashboard(adminToken, competencia string) (map[string]any, error) {
	cli := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	if cli == nil || !cli.Enabled() {
		return nil, fmt.Errorf("XANO_API_FINANCEIRO nao configurado")
	}
	payload := map[string]any{"admin_token": strings.TrimSpace(adminToken)}
	if c := strings.TrimSpace(competencia); c != "" {
		payload["competencia"] = c
	}
	return cli.Post("/fp_fin_dashboard", payload)
}

func governancaAlerta(sess auth.SessaoAdm) map[string]any {
	rt := pggovernanca.NovoRuntime()
	idRep := strings.TrimSpace(sess.IDRepresentante)
	idCen := strings.TrimSpace(sess.IDCentralCatalogo)
	if idCen == "" {
		idCen = strings.TrimSpace(sess.IDCentralUUID)
	}
	est := rt.EstadoEntidade(sess.UserTipo, idRep, idCen, false)
	if est.Nivel == pggovernanca.NivelOK || est.Nivel == "" {
		return nil
	}
	return map[string]any{
		"pode_operar":       est.PodeOperar,
		"nivel":             est.Nivel,
		"dias_restantes":    est.DiasRestantes,
		"dias_vencido":      est.DiasVencido,
		"mensagem_publica":  est.MensagemPublica,
		"repasse_fatura_id": est.RepasseFaturaID,
	}
}

func mergeKPIComGovernanca(kpi finresumo.DashboardKPI, gov map[string]any) map[string]any {
	b, _ := json.Marshal(kpi)
	out := map[string]any{}
	_ = json.Unmarshal(b, &out)
	if gov != nil {
		out["governanca_alerta"] = gov
	}
	return out
}
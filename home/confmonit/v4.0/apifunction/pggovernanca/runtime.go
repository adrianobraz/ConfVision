package pggovernanca

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"apifunction/config"
	"apifunction/db"
	"apifunction/xano"
)

type Runtime struct {
	Xano       *xano.Client
	AdminToken string
	DiasGraca  int
}

func NovoRuntime() *Runtime {
	dias := config.GovDiasGraca
	if dias <= 0 {
		dias = 6
	}
	return &Runtime{
		Xano:       xano.New(config.XanoAPIFinanceiro, config.WorkerSecret),
		AdminToken: config.WorkerAdminToken,
		DiasGraca:  dias,
	}
}

func (rt *Runtime) EstadoEntidade(userTipo, idVinculo, idCentral string, breakglass bool) EstadoEntidade {
	def := EstadoEntidade{
		PodeOperar: true,
		Nivel:      NivelOK,
	}
	if breakglass {
		return def
	}
	userTipo = strings.ToUpper(strings.TrimSpace(userTipo))
	entTipo := ""
	entID := ""
	switch userTipo {
	case EntidadeREP:
		entTipo = EntidadeREP
		entID = strings.TrimSpace(idVinculo)
	case EntidadeCEN:
		entTipo = EntidadeCEN
		entID = strings.TrimSpace(idCentral)
		if entID == "" {
			entID = strings.TrimSpace(idVinculo)
		}
	default:
		return def
	}
	if entID == "" {
		return def
	}
	row, err := GetRestricaoAtiva(entTipo, entID)
	if err != nil || row == nil {
		return def
	}
	return estadoFromRestricao(*row)
}

func estadoFromRestricao(row RestricaoRow) EstadoEntidade {
	diasVenc := diasDesdeVencimento(row.Vencimento)
	diasRest := row.DiasRestantes
	if diasRest < 0 {
		diasRest = 0
	}
	out := EstadoEntidade{
		PodeOperar:      false,
		Nivel:           row.Nivel,
		DiasRestantes:   diasRest,
		DiasVencido:     diasVenc,
		MensagemPublica: MensagemPublica,
		RepasseFaturaID: row.FaturaID,
	}
	if row.Nivel == NivelOK {
		out.PodeOperar = true
		out.MensagemPublica = ""
	}
	return out
}

func (rt *Runtime) FranqueadoEstado(idFranqueado string) FranqueadoEstado {
	def := FranqueadoEstado{Motivo: "ok"}
	ok, err := FranqueadoBloqueadoCascata(idFranqueado)
	if err != nil || !ok {
		return def
	}
	return FranqueadoEstado{
		BloqueadoCascata: true,
		Motivo:           MotivoPublico,
		MensagemPublica:  MensagemPublica,
	}
}

func (rt *Runtime) WorkerTick() (TickResult, error) {
	res := TickResult{}
	if rt.Xano == nil || !rt.Xano.Enabled() || rt.AdminToken == "" {
		return res, fmt.Errorf("xano ou WORKER_ADMIN_TOKEN nao configurado")
	}

	repasses, err := rt.listarRepasseAbertos(TipoRepCentral)
	if err != nil {
		return res, err
	}
	bg, err := rt.listarRepasseAbertos(TipoRepBreakglass)
	if err != nil {
		return res, err
	}
	repasses = append(repasses, bg...)

	seen := map[int]bool{}
	for _, r := range repasses {
		seen[r.FaturaID] = true
		if err := UpsertRepasse(r); err != nil {
			return res, err
		}
		res.RepasseSincronizados++
		if r.Status != "aberta" || !estaVencido(r.Vencimento) {
			entTipo, entID := entidadeForRepasse(r)
			if entID != "" {
				_ = ClearRestricao(entTipo, entID)
			}
			reat, _ := rt.reativarPorRepasse(r.FaturaID)
			res.Reativacoes += reat
			continue
		}
		entTipo, entID := entidadeForRepasse(r)
		if entID == "" {
			continue
		}
		diasVenc := diasDesdeVencimento(r.Vencimento)
		nivel := NivelAlerta
		diasRest := rt.DiasGraca - diasVenc
		if diasRest < 0 {
			diasRest = 0
		}
		if diasVenc >= rt.DiasGraca {
			nivel = NivelBloqueado
			diasRest = 0
		}
		if err := UpsertRestricao(RestricaoRow{
			EntidadeTipo:  entTipo,
			EntidadeID:    entID,
			FaturaID:      r.FaturaID,
			Nivel:         nivel,
			Vencimento:    r.Vencimento,
			DiasRestantes: diasRest,
			Ativo:         true,
		}); err != nil {
			return res, err
		}
		res.RestricoesAtivas++
		if nivel == NivelBloqueado && r.Tipo == TipoRepCentral {
			n, err := rt.aplicarCascataFranqueados(r)
			if err != nil {
				return res, err
			}
			res.SuspensoesNovas += n
		}
	}

	if err := rt.limparRepasseQuitados(seen); err != nil {
		return res, err
	}
	return res, nil
}

func (rt *Runtime) limparRepasseQuitados(seen map[int]bool) error {
	rows, err := ListarRepasseAbertosVencidos()
	if err != nil {
		return err
	}
	for _, r := range rows {
		if seen[r.FaturaID] {
			continue
		}
		_ = MarkRepasseStatus(r.FaturaID, "paga")
		entTipo, entID := entidadeForRepasse(r)
		if entID != "" {
			_ = ClearRestricao(entTipo, entID)
		}
		_, _ = rt.reativarPorRepasse(r.FaturaID)
	}
	return nil
}

func (rt *Runtime) reativarPorRepasse(faturaID int) (int, error) {
	ids, err := ReativarSuspensoesPorRepasse(faturaID)
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	n := 0
	for _, assID := range ids {
		if err := rt.xanoSuspenderOuLiberar(assID, false); err != nil {
			continue
		}
		n++
	}
	return n, nil
}

func (rt *Runtime) aplicarCascataFranqueados(r RepasseRow) (int, error) {
	if r.IDRepresentante == "" {
		return 0, nil
	}
	franqueados, err := listarFranqueadosMySQL(r.IDRepresentante)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, idFra := range franqueados {
		assinaturas, err := rt.listarAssinaturasAtivas(idFra)
		if err != nil {
			continue
		}
		for _, assID := range assinaturas {
			if err := RegistrarSuspensaoCascata(idFra, assID, r.IDRepresentante, r.FaturaID, "rep_inadimplente_central"); err != nil {
				continue
			}
			if err := rt.xanoSuspenderOuLiberar(assID, true); err != nil {
				continue
			}
			n++
		}
	}
	return n, nil
}

func listarFranqueadosMySQL(idRep string) ([]string, error) {
	if db.Conn == nil {
		return nil, fmt.Errorf("mysql indisponivel")
	}
	rows, err := db.Conn.Query(`
		SELECT ID_Franqueado FROM franqueado
		WHERE ID_Representante = ?
		  AND (DataCancelamento IS NULL OR DataCancelamento = '0000-00-00 00:00:00')
	`, idRep)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, strings.TrimSpace(id))
	}
	return ids, rows.Err()
}

func (rt *Runtime) listarRepasseAbertos(tipo string) ([]RepasseRow, error) {
	out, err := rt.Xano.Post("/fp_repasse_listar", map[string]any{
		"admin_token": rt.AdminToken,
		"status":      "aberta",
		"tipo":        tipo,
		"limite":      500,
	})
	if err != nil {
		return nil, err
	}
	rawList, _ := out["dados"].([]any)
	var rows []RepasseRow
	for _, item := range rawList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := mapRepasseFatura(m, tipo)
		if r.FaturaID > 0 {
			rows = append(rows, r)
		}
	}
	return rows, nil
}

func (rt *Runtime) listarAssinaturasAtivas(idFranqueado string) ([]int, error) {
	out, err := rt.Xano.Post("/fp_assinatura_listar", map[string]any{
		"admin_token":   rt.AdminToken,
		"id_franqueado": idFranqueado,
		"status":        "ativa",
	})
	if err != nil {
		return nil, err
	}
	raw, _ := out["dados"].([]any)
	var ids []int
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := intFromAny(m["id"])
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (rt *Runtime) xanoSuspenderOuLiberar(assinaturaID int, suspender bool) error {
	if suspender {
		_, err := rt.Xano.Post("/fp_assinatura_suspender", map[string]any{
			"admin_token":   rt.AdminToken,
			"assinatura_id": assinaturaID,
			"observacao":    MotivoPublico,
		})
		return err
	}
	_, err := rt.Xano.Post("/fp_assinatura_liberar_manual", map[string]any{
		"admin_token":   rt.AdminToken,
		"assinatura_id": assinaturaID,
		"observacao":    "Reativacao automatica pos repasse",
		"admin_usuario": "governanca_worker",
	})
	return err
}

func mapRepasseFatura(m map[string]any, tipo string) RepasseRow {
	return RepasseRow{
		FaturaID:        intFromAny(m["id"]),
		Tipo:            strings.TrimSpace(firstString(m["tipo"], tipo)),
		IDRepresentante: strings.TrimSpace(anyString(m["id_representante"])),
		IDCentral:       strings.TrimSpace(anyString(m["id_central"])),
		Status:          strings.TrimSpace(firstString(m["status"], "aberta")),
		Vencimento:      parseTimeAny(m["vencimento_em"]),
		ValorTotal:      floatFromAny(m["valor_total"]),
	}
}

func entidadeForRepasse(r RepasseRow) (string, string) {
	if r.Tipo == TipoRepBreakglass {
		return EntidadeCEN, r.IDCentral
	}
	return EntidadeREP, r.IDRepresentante
}

func estaVencido(vencimento int64) bool {
	if vencimento <= 0 {
		return false
	}
	t := time.UnixMilli(vencimento)
	if vencimento < 1e12 {
		t = time.Unix(vencimento, 0)
	}
	return t.Before(time.Now())
}

func diasDesdeVencimento(vencimento int64) int {
	if vencimento <= 0 {
		return 0
	}
	t := time.UnixMilli(vencimento)
	if vencimento < 1e12 {
		t = time.Unix(vencimento, 0)
	}
	diff := time.Since(t)
	d := int(math.Floor(diff.Hours() / 24))
	if d < 0 {
		return 0
	}
	return d
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func floatFromAny(v any) float64 {
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

func anyString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func firstString(v any, def string) string {
	if s := strings.TrimSpace(anyString(v)); s != "" {
		return s
	}
	return def
}

func parseTimeAny(v any) int64 {
	switch t := v.(type) {
	case float64:
		if t > 1e12 {
			return int64(t)
		}
		return int64(t) * 1000
	case string:
		if t == "" {
			return 0
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if parsed, err := time.Parse(layout, t); err == nil {
				return parsed.UnixMilli()
			}
		}
	}
	return 0
}

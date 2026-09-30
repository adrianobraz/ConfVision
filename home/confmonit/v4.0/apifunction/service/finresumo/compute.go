package finresumo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"apifunction/auth"
	"apifunction/db"
	"apifunction/pgcredito"
	financeiro "apifunction/service/financeiro"
)

type ProdutoKPI struct {
	Produto string  `json:"produto"`
	Ativas  int     `json:"ativas"`
	MRR     float64 `json:"mrr"`
	Bundle  int     `json:"bundle"`
}

type DashboardKPI struct {
	AssinaturasAtivas      int            `json:"assinaturas_ativas"`
	AssinaturasPendentes   int            `json:"assinaturas_pendentes"`
	AssinaturasSuspensas   int            `json:"assinaturas_suspensas"`
	FaturasAbertas         int            `json:"faturas_abertas"`
	FaturasPagas           int            `json:"faturas_pagas"`
	FaturasCanceladas      int            `json:"faturas_canceladas"`
	FaturasVencidas        int            `json:"faturas_vencidas"`
	ValorEmAberto          float64        `json:"valor_em_aberto"`
	ValorPago              float64        `json:"valor_pago"`
	ValorVencido           float64        `json:"valor_vencido"`
	ValorCancelado         float64        `json:"valor_cancelado"`
	QtdInadimplentes       int            `json:"qtd_inadimplentes"`
	SaldoCaixa             float64        `json:"saldo_caixa"`
	TotalEntradasCaixa     float64        `json:"total_entradas_caixa"`
	TotalSaidasCaixa       float64        `json:"total_saidas_caixa"`
	ReceitaPeriodo         float64        `json:"receita_periodo"`
	EntradasPeriodo        float64        `json:"entradas_periodo"`
	RetiradasPeriodo       float64        `json:"retiradas_periodo"`
	DespesasPeriodo        float64        `json:"despesas_periodo"`
	RepassesPeriodo        float64        `json:"repasses_periodo"`
	ValorRepassesAbertos   float64        `json:"valor_repasses_abertos"`
	TotalDespesasPeriodo   float64        `json:"total_despesas_periodo"`
	ResultadoPeriodo       float64        `json:"resultado_periodo"`
	ContasPagarAbertas     int            `json:"contas_pagar_abertas"`
	ValorAPagar            float64        `json:"valor_a_pagar"`
	Competencia            string         `json:"competencia"`
	MRRTotal               float64        `json:"mrr_total"`
	MRRAvulso              float64        `json:"mrr_avulso"`
	QtdAssinaturasBundle   int            `json:"qtd_assinaturas_bundle"`
	ReceitaConfvision      float64        `json:"receita_confvision"`
	ReceitaAssinaturas     float64        `json:"receita_assinaturas"`
	ReceitaOutros          float64        `json:"receita_outros"`
	AssinaturasPorProduto  []ProdutoKPI   `json:"assinaturas_por_produto"`
	EscopoRep              bool           `json:"escopo_rep"`
	IDRepresentante        string         `json:"id_representante,omitempty"`
	QtdFranqueadosCarteira int            `json:"qtd_franqueados_carteira,omitempty"`
	Fonte                  string         `json:"fonte,omitempty"`
}

type faturaRow struct {
	ID              int
	Status          string
	Tipo            string
	ValorTotal      float64
	VencimentoEm    *time.Time
	PagoEm          *time.Time
	IDFranqueado    string
	IDRepresentante string
	IDCentral       string
}

type assinaturaRow struct {
	Produto    string
	Status     string
	Valor      float64
	Observacao string
}

type pagamentoRow struct {
	FPFaturaID int
	Valor      float64
	PagoEm     *time.Time
}

var produtosKeys = []string{
	"franqueadopro", "webterminal", "terminalmovel", "confvision", "webambiente", "dialyze",
}

func ComputeDashboard(ctx context.Context, sess auth.SessaoAdm, competencia, adminUsuario string) (DashboardKPI, error) {
	pg, err := pgcredito.DB()
	if err != nil {
		return DashboardKPI{}, err
	}

	escopo, err := buildEscopo(ctx, sess)
	if err != nil {
		return DashboardKPI{}, err
	}

	faturas, err := loadFaturas(ctx, pg, escopo)
	if err != nil {
		return DashboardKPI{}, err
	}

	faturaIDs := make([]int, 0, len(faturas))
	for _, f := range faturas {
		faturaIDs = append(faturaIDs, f.ID)
	}

	pagamentos, err := loadPagamentos(ctx, pg, faturaIDs)
	if err != nil {
		return DashboardKPI{}, err
	}

	assinaturas, err := loadAssinaturas(ctx, pg, escopo)
	if err != nil {
		return DashboardKPI{}, err
	}

	out := DashboardKPI{
		Competencia:     strings.TrimSpace(competencia),
		EscopoRep:       escopo.isRep,
		IDRepresentante: escopo.idRep,
		Fonte:           "postgres_mirror",
	}
	if escopo.isRep {
		out.QtdFranqueadosCarteira = len(escopo.fraIDs)
	}

	temCompetencia := len(strings.TrimSpace(competencia)) >= 7
	now := time.Now().UTC()
	inadimplentes := map[string]struct{}{}
	repasseIDs := map[int]struct{}{}

	for _, f := range faturas {
		st := strings.TrimSpace(f.Status)
		val := f.ValorTotal
		tipoF := strings.TrimSpace(f.Tipo)
		ehRepasse := ehRepasseTipo(tipoF)
		if ehRepasse {
			repasseIDs[f.ID] = struct{}{}
		}

		if st == "aberta" {
			if ehRepasse {
				out.ValorRepassesAbertos += val
			} else {
				out.FaturasAbertas++
				out.ValorEmAberto += val
				if f.VencimentoEm != nil && f.VencimentoEm.Before(now) {
					out.FaturasVencidas++
					out.ValorVencido += val
					if f.IDFranqueado != "" {
						if _, ok := inadimplentes[f.IDFranqueado]; !ok {
							inadimplentes[f.IDFranqueado] = struct{}{}
						}
					}
				}
			}
		}

		if st == "paga" {
			noPeriodo := !temCompetencia
			if temCompetencia && f.PagoEm != nil {
				noPeriodo = f.PagoEm.UTC().Format("2006-01") == strings.TrimSpace(competencia)
			}
			if ehRepasse {
				if noPeriodo {
					out.RepassesPeriodo += val
				}
			} else {
				out.FaturasPagas++
				out.ValorPago += val
				if noPeriodo {
					out.ReceitaPeriodo += val
					classificarReceitaTipo(tipoF, val, &out)
				}
			}
		}

		if st == "cancelada" {
			out.FaturasCanceladas++
			out.ValorCancelado += val
		}
	}

	out.QtdInadimplentes = len(inadimplentes)

	for _, p := range pagamentos {
		incluir := !temCompetencia
		if temCompetencia && p.PagoEm != nil {
			incluir = p.PagoEm.UTC().Format("2006-01") == strings.TrimSpace(competencia)
		}
		if !incluir {
			continue
		}
		if _, ok := repasseIDs[p.FPFaturaID]; ok {
			continue
		}
		out.EntradasPeriodo += p.Valor
	}

	if adminUsuario != "" {
		caixa, err := loadCaixa(ctx, pg, adminUsuario)
		if err != nil {
			return DashboardKPI{}, err
		}
		for _, m := range caixa {
			tipo := strings.TrimSpace(m.Tipo)
			if tipo == "entrada" {
				out.TotalEntradasCaixa += m.Valor
				out.SaldoCaixa += m.Valor
			} else if tipo == "retirada" || tipo == "saida" {
				out.TotalSaidasCaixa += m.Valor
				out.SaldoCaixa -= m.Valor
				dt := m.MovimentoEm
				if dt == nil {
					dt = m.CreatedAt
				}
				if inCompetencia(dt, competencia, temCompetencia) {
					out.RetiradasPeriodo += m.Valor
				}
			}
		}

		contas, err := loadContasPagar(ctx, pg, adminUsuario)
		if err != nil {
			return DashboardKPI{}, err
		}
		for _, cp := range contas {
			st := strings.TrimSpace(cp.Status)
			if st == "aberta" {
				out.ContasPagarAbertas++
				out.ValorAPagar += cp.Valor
			}
			if st == "paga" && inCompetencia(cp.PagoEm, competencia, temCompetencia) {
				out.DespesasPeriodo += cp.Valor
			}
		}
	}

	out.ValorAPagar += out.ValorRepassesAbertos
	out.TotalDespesasPeriodo = out.RetiradasPeriodo + out.DespesasPeriodo + out.RepassesPeriodo
	out.ResultadoPeriodo = out.ReceitaPeriodo - out.TotalDespesasPeriodo

	var ativas, pendentes, suspensas []assinaturaRow
	for _, a := range assinaturas {
		switch strings.TrimSpace(a.Status) {
		case "ativa":
			ativas = append(ativas, a)
			out.AssinaturasAtivas++
		case "pendente":
			pendentes = append(pendentes, a)
			out.AssinaturasPendentes++
		case "suspensa":
			suspensas = append(suspensas, a)
			out.AssinaturasSuspensas++
		}
	}

	out.AssinaturasPorProduto = buildProdutoKPI(ativas, &out)

	return out, nil
}

type escopoFin struct {
	isRep   bool
	idRep   string
	idCen   string
	cenKeys map[string]struct{}
	fraIDs  []string
}

func buildEscopo(ctx context.Context, sess auth.SessaoAdm) (escopoFin, error) {
	idCen := financeiro.IDCentralSessaoExport(sess)
	if idCen == "" {
		return escopoFin{}, fmt.Errorf("sessao sem idCentral")
	}
	cenKeys, err := financeiro.ChavesCentral(ctx, idCen)
	if err != nil {
		return escopoFin{}, err
	}
	e := escopoFin{cenKeys: cenKeys, idCen: idCen}
	if sess.UserTipo == "REP" {
		e.isRep = true
		e.idRep = strings.TrimSpace(sess.IDRepresentante)
		if e.idRep == "" {
			return escopoFin{}, fmt.Errorf("REP sem id_representante")
		}
		e.fraIDs, err = franqueadosDoRepresentante(ctx, e.idRep)
		if err != nil {
			return escopoFin{}, err
		}
	}
	return e, nil
}

func franqueadosDoRepresentante(ctx context.Context, idRep string) ([]string, error) {
	rows, err := db.Conn.QueryContext(ctx, `
		SELECT ID_Franqueado FROM franqueado
		WHERE TRIM(COALESCE(ID_Representante, '')) = ?
		  AND TRIM(COALESCE(ID_Franqueado, '')) != ''
	`, idRep)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	seen := map[string]struct{}{}
	for rows.Next() {
		var id sql.NullString
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		f := strings.TrimSpace(id.String)
		if f == "" {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out, rows.Err()
}

func loadFaturas(ctx context.Context, pg *sql.DB, e escopoFin) ([]faturaRow, error) {
	rows, err := pg.QueryContext(ctx, `
		SELECT id, status, tipo, valor_total, vencimento_em, pago_em,
		       id_franqueado, id_representante, id_central
		FROM fp_fin_fatura
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fraSet := toSet(e.fraIDs)
	var out []faturaRow
	for rows.Next() {
		var f faturaRow
		var venc, pago sql.NullTime
		if err := rows.Scan(&f.ID, &f.Status, &f.Tipo, &f.ValorTotal, &venc, &pago,
			&f.IDFranqueado, &f.IDRepresentante, &f.IDCentral); err != nil {
			return nil, err
		}
		if venc.Valid {
			t := venc.Time
			f.VencimentoEm = &t
		}
		if pago.Valid {
			t := pago.Time
			f.PagoEm = &t
		}
		if !faturaNoEscopo(f, e, fraSet) {
			continue
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func faturaNoEscopo(f faturaRow, e escopoFin, fraSet map[string]struct{}) bool {
	if e.isRep {
		if strings.TrimSpace(f.IDRepresentante) == e.idRep {
			return true
		}
		if _, ok := fraSet[strings.TrimSpace(f.IDFranqueado)]; ok {
			return true
		}
		return false
	}
	if _, ok := e.cenKeys[strings.TrimSpace(f.IDCentral)]; ok {
		return true
	}
	return false
}

func loadPagamentos(ctx context.Context, pg *sql.DB, faturaIDs []int) ([]pagamentoRow, error) {
	if len(faturaIDs) == 0 {
		return nil, nil
	}
	rows, err := pg.QueryContext(ctx, `SELECT fp_fatura_id, valor, pago_em FROM fp_fin_pagamento`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	idSet := map[int]struct{}{}
	for _, id := range faturaIDs {
		idSet[id] = struct{}{}
	}
	var out []pagamentoRow
	for rows.Next() {
		var p pagamentoRow
		var pago sql.NullTime
		if err := rows.Scan(&p.FPFaturaID, &p.Valor, &pago); err != nil {
			return nil, err
		}
		if _, ok := idSet[p.FPFaturaID]; !ok {
			continue
		}
		if pago.Valid {
			t := pago.Time
			p.PagoEm = &t
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func loadAssinaturas(ctx context.Context, pg *sql.DB, e escopoFin) ([]assinaturaRow, error) {
	rows, err := pg.QueryContext(ctx, `
		SELECT produto, status, valor, observacao, id_franqueado, id_representante, id_central
		FROM fp_fin_assinatura
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fraSet := toSet(e.fraIDs)
	var out []assinaturaRow
	for rows.Next() {
		var a assinaturaRow
		var idFra, idRep, idCen string
		if err := rows.Scan(&a.Produto, &a.Status, &a.Valor, &a.Observacao, &idFra, &idRep, &idCen); err != nil {
			return nil, err
		}
		if e.isRep {
			if strings.TrimSpace(idRep) != e.idRep {
				if _, ok := fraSet[strings.TrimSpace(idFra)]; !ok {
					continue
				}
			}
		} else {
			if _, ok := e.cenKeys[strings.TrimSpace(idCen)]; !ok {
				continue
			}
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type caixaRow struct {
	Tipo         string
	Valor        float64
	MovimentoEm  *time.Time
	CreatedAt    *time.Time
}

func loadCaixa(ctx context.Context, pg *sql.DB, adminUsuario string) ([]caixaRow, error) {
	rows, err := pg.QueryContext(ctx, `
		SELECT tipo, valor, movimento_em, created_at
		FROM fp_fin_caixa_movimento WHERE admin_usuario = $1
	`, adminUsuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []caixaRow
	for rows.Next() {
		var m caixaRow
		var mov, created sql.NullTime
		if err := rows.Scan(&m.Tipo, &m.Valor, &mov, &created); err != nil {
			return nil, err
		}
		if mov.Valid {
			t := mov.Time
			m.MovimentoEm = &t
		}
		if created.Valid {
			t := created.Time
			m.CreatedAt = &t
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type contaPagarRow struct {
	Valor  float64
	Status string
	PagoEm *time.Time
}

func loadContasPagar(ctx context.Context, pg *sql.DB, adminUsuario string) ([]contaPagarRow, error) {
	rows, err := pg.QueryContext(ctx, `
		SELECT valor, status, pago_em FROM fp_fin_conta_pagar WHERE admin_usuario = $1
	`, adminUsuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []contaPagarRow
	for rows.Next() {
		var cp contaPagarRow
		var pago sql.NullTime
		if err := rows.Scan(&cp.Valor, &cp.Status, &pago); err != nil {
			return nil, err
		}
		if pago.Valid {
			t := pago.Time
			cp.PagoEm = &t
		}
		out = append(out, cp)
	}
	return out, rows.Err()
}

func buildProdutoKPI(ativas []assinaturaRow, out *DashboardKPI) []ProdutoKPI {
	result := make([]ProdutoKPI, 0, len(produtosKeys))
	for _, pk := range produtosKeys {
		kpi := ProdutoKPI{Produto: pk}
		for _, aa := range ativas {
			if strings.TrimSpace(aa.Produto) != pk {
				continue
			}
			kpi.Ativas++
			kpi.MRR += aa.Valor
			out.MRRTotal += aa.Valor
			if strings.Contains(strings.TrimSpace(aa.Observacao), "bundle_fp") {
				kpi.Bundle++
				out.QtdAssinaturasBundle++
			} else {
				out.MRRAvulso += aa.Valor
			}
		}
		result = append(result, kpi)
	}
	return result
}

func ehRepasseTipo(tipo string) bool {
	tipo = strings.TrimSpace(tipo)
	return tipo == "repasse_rep_central" || tipo == "repasse_central_breakglass" || strings.Contains(tipo, "repasse")
}

func classificarReceitaTipo(tipo string, val float64, out *DashboardKPI) {
	tipo = strings.TrimSpace(tipo)
	if ehRepasseTipo(tipo) {
		return
	}
	if tipo == "confvision_venda" || tipo == "confvision_renovacao" || strings.Contains(tipo, "confvision") {
		out.ReceitaConfvision += val
		return
	}
	if tipo == "assinatura" {
		out.ReceitaAssinaturas += val
		return
	}
	out.ReceitaOutros += val
}

func inCompetencia(dt *time.Time, competencia string, temCompetencia bool) bool {
	if !temCompetencia {
		return true
	}
	if dt == nil {
		return false
	}
	return dt.UTC().Format("2006-01") == strings.TrimSpace(competencia)
}

func toSet(vals []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" {
			out[v] = struct{}{}
		}
	}
	return out
}

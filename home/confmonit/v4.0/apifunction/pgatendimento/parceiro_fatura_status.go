package pgatendimento

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"apifunction/confservice"
)

const (
	FaseAguardandoPagamento     = "AGUARDANDO_PAGAMENTO"
	FasePagaAguardandoRepasse   = "PAGA_AGUARDANDO_REPASSE_PARCEIRO"
	FaseRepasseParceiroAberto   = "REPASSE_PARCEIRO_ABERTO"
	FaseConcluido               = "CONCLUIDO"
	FaseCancelada               = "CANCELADA"
)

type FaturaParceiroStatus struct {
	FPFaturaID     int     `json:"fp_fatura_id"`
	Tipo           string  `json:"tipo"`
	CicloRef       string  `json:"ciclo_ref"`
	IDParceiro     string  `json:"id_parceiro"`
	NomeParceiro   string  `json:"nome_parceiro,omitempty"`
	StatusOps      string  `json:"status_ops"`
	Fase           string  `json:"fase"`
	ValorTotal     float64 `json:"valor_total,omitempty"`
	RepasseStatus  string  `json:"repasse_status,omitempty"`
	RepasseValor   float64 `json:"repasse_valor,omitempty"`
	PagoAte        string  `json:"pago_ate,omitempty"`
}

func GetAtivacaoPorFatura(ctx context.Context, faturaID int) (*AtivacaoRegistro, error) {
	if faturaID <= 0 {
		return nil, nil
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	return scanAtivacao(d.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_parceiro, status, qtd_clientes, preco_unitario, valor_total,
       fp_fatura_id, ciclo_ref, periodo_inicio, periodo_fim, pago_ate
FROM ops_parceiro_ativacao
WHERE fp_fatura_id = $1
ORDER BY id DESC LIMIT 1`, faturaID))
}

// MarcarAtivacaoPagaFatura marca ativação pela fatura (mesmo se status cancelada) e retorna registro.
func MarcarAtivacaoPagaFatura(ctx context.Context, idFranqueado string, faturaID int, refIDParceiro string, pagoAte time.Time) (*AtivacaoRegistro, error) {
	if faturaID <= 0 {
		return nil, errors.New("fatura_id invalido")
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	pagoAte = dateOnlyBR(pagoAte)
	rec, err := GetAtivacaoPorFatura(ctx, faturaID)
	if err != nil {
		return nil, err
	}
	if rec == nil && strings.TrimSpace(refIDParceiro) != "" {
		rec, err = scanAtivacao(d.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_parceiro, status, qtd_clientes, preco_unitario, valor_total,
       fp_fatura_id, ciclo_ref, periodo_inicio, periodo_fim, pago_ate
FROM ops_parceiro_ativacao
WHERE id_franqueado = $1 AND id_parceiro = $2 AND status IN ($3, $4)
ORDER BY id DESC LIMIT 1`, idFranqueado, refIDParceiro, StatusAtivPendente, StatusAtivCancelada))
		if err != nil {
			return nil, err
		}
	}
	// Renovação: ativação já paga vinculada a fatura anterior no mesmo ciclo.
	if rec == nil && strings.TrimSpace(refIDParceiro) != "" {
		rec, err = scanAtivacao(d.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_parceiro, status, qtd_clientes, preco_unitario, valor_total,
       fp_fatura_id, ciclo_ref, periodo_inicio, periodo_fim, pago_ate
FROM ops_parceiro_ativacao
WHERE id_franqueado = $1 AND id_parceiro = $2 AND status = $3
  AND (fp_fatura_id IS NULL OR fp_fatura_id <> $4)
ORDER BY id DESC LIMIT 1`, idFranqueado, refIDParceiro, StatusAtivPaga, faturaID))
		if err != nil {
			return nil, err
		}
	}
	if rec == nil {
		return nil, nil
	}
	_, err = d.ExecContext(ctx, `
UPDATE ops_parceiro_ativacao SET status = $2, fp_fatura_id = $3,
       pago_ate = $4, updated_at = NOW()
WHERE id = $1`, rec.ID, StatusAtivPaga, faturaID, pagoAte)
	if err != nil {
		return nil, err
	}
	rec.Status = StatusAtivPaga
	rec.FPFaturaID = faturaID
	rec.PagoAte = pagoAte.Format("2006-01-02")
	return rec, nil
}

// MetaRepasseAtivacao retorna parceiro e valor piso total (piso × qtd) após ativação paga.
func MetaRepasseAtivacao(ctx context.Context, idFranqueado string, faturaID int, refIDParceiro string, valorPisoHint float64) (idParceiro string, valorPiso float64, err error) {
	rec, err := GetAtivacaoPorFatura(ctx, faturaID)
	if err != nil {
		return "", 0, err
	}
	if rec == nil || rec.Status != StatusAtivPaga {
		if rec, err = MarcarAtivacaoPagaFatura(ctx, idFranqueado, faturaID, refIDParceiro, time.Now()); err != nil {
			return "", 0, err
		}
	}
	if rec == nil {
		return "", 0, errors.New("ativacao nao encontrada para fatura")
	}
	idParceiro = strings.TrimSpace(rec.IDParceiro)
	if idParceiro == "" {
		return "", 0, errors.New("ativacao sem parceiro")
	}
	qtd := rec.QtdClientes
	if qtd <= 0 {
		qtd = 1
	}
	pisoUnit := valorPisoHint
	if pisoUnit <= 0 {
		preco, err := lookupPrecoParceiro(ctx, rec.IDFranqueado, idParceiro)
		if err != nil {
			return "", 0, err
		}
		pisoUnit = preco.PisoParceiro
	}
	valorPiso = roundMoney(pisoUnit * float64(qtd))
	if valorPiso <= 0 {
		return "", 0, fmt.Errorf("piso ativacao invalido")
	}
	return idParceiro, valorPiso, nil
}

func ListFaturasParceiroStatus(ctx context.Context, idFranqueado string) ([]FaturaParceiroStatus, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, errors.New("id_franqueado obrigatorio")
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var out []FaturaParceiroStatus

	rows, err := d.QueryContext(ctx, `
SELECT fp_fatura_id, id_parceiro, status, valor_total, ciclo_ref, COALESCE(pago_ate::text, '')
FROM ops_parceiro_ativacao
WHERE id_franqueado = $1 AND fp_fatura_id IS NOT NULL AND fp_fatura_id > 0
ORDER BY fp_fatura_id DESC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var fpID int
		var idP, st, ciclo, pagoAte string
		var valor float64
		if err := rows.Scan(&fpID, &idP, &st, &valor, &ciclo, &pagoAte); err != nil {
			continue
		}
		item := FaturaParceiroStatus{
			FPFaturaID: fpID, Tipo: "PARCEIRO", CicloRef: ciclo,
			IDParceiro: idP, StatusOps: st, ValorTotal: valor, PagoAte: pagoAte,
		}
		item.Fase = faseFromOps(st)
		out = append(out, item)
	}

	rows2, err := d.QueryContext(ctx, `
SELECT fp_fatura_id, id_parceiro, status, valor_total, ciclo_ref, COALESCE(pago_ate::text, '')
FROM ops_parceiro_excecao
WHERE id_franqueado = $1 AND fp_fatura_id IS NOT NULL AND fp_fatura_id > 0
ORDER BY fp_fatura_id DESC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var fpID int
		var idP, st, ciclo, pagoAte string
		var valor float64
		if err := rows2.Scan(&fpID, &idP, &st, &valor, &ciclo, &pagoAte); err != nil {
			continue
		}
		item := FaturaParceiroStatus{
			FPFaturaID: fpID, Tipo: "EXCECAO", CicloRef: ciclo,
			IDParceiro: idP, StatusOps: st, ValorTotal: valor, PagoAte: pagoAte,
		}
		item.Fase = faseFromOps(st)
		out = append(out, item)
	}

	repasses, _ := confservice.ListRepassesFatura(idFranqueado, 0)
	byFatura := map[int][]confservice.RepasseResumo{}
	for _, r := range repasses {
		if r.FPFaturaID <= 0 {
			continue
		}
		byFatura[r.FPFaturaID] = append(byFatura[r.FPFaturaID], r)
	}
	nomes := map[string]string{}
	for _, r := range repasses {
		if r.NomeParceiro != "" && r.IDParceiro != "" {
			nomes[r.IDParceiro] = r.NomeParceiro
		}
	}
	for i := range out {
		out[i].NomeParceiro = nomes[out[i].IDParceiro]
		if reps := byFatura[out[i].FPFaturaID]; len(reps) > 0 {
			applyRepasseFase(&out[i], reps)
		}
	}
	return out, nil
}

func faseFromOps(st string) string {
	switch strings.ToLower(strings.TrimSpace(st)) {
	case StatusAtivPendente:
		return FaseAguardandoPagamento
	case StatusAtivPaga:
		return FasePagaAguardandoRepasse
	case StatusAtivCancelada:
		return FaseCancelada
	default:
		return st
	}
}

func applyRepasseFase(item *FaturaParceiroStatus, reps []confservice.RepasseResumo) {
	for _, r := range reps {
		if strings.EqualFold(r.IDParceiro, item.IDParceiro) {
			item.RepasseStatus = r.Status
			item.RepasseValor = r.Valor
			switch strings.ToUpper(strings.TrimSpace(r.Status)) {
			case "ABERTO":
				item.Fase = FaseRepasseParceiroAberto
			case "PAGO":
				item.Fase = FaseConcluido
			}
			return
		}
	}
}

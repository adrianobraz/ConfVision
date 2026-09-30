package pgatendimento

import (
	"context"
	"errors"
	"strings"

	"apifunction/confservice"
	"apifunction/pgcobranca"
)

type PosEstornoResult struct {
	RepassesCancelados int `json:"repasses_cancelados"`
	AtivacaoRevertida  int `json:"ativacao_revertida"`
	ExcecaoRevertida   int `json:"excecao_revertida"`
	VinculosDesativados int `json:"vinculos_desativados"`
}

func ReverterPosEstorno(ctx context.Context, idFranqueado string, faturaID int) (PosEstornoResult, error) {
	res := PosEstornoResult{}
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" || faturaID <= 0 {
		return res, errors.New("id_franqueado e fatura_id obrigatorios")
	}
	d, err := db(ctx)
	if err != nil {
		return res, err
	}

	n, _ := confservice.CancelarRepassesFatura(idFranqueado, faturaID)
	res.RepassesCancelados = n

	if rec, _ := GetAtivacaoPorFatura(ctx, faturaID); rec != nil && rec.Status == StatusAtivPaga {
		_ = pgcobranca.DesativarServicoPorRef(ctx, idFranqueado, RefTipoParceiroConfig, rec.IDParceiro)
	}
	r1, err := d.ExecContext(ctx, `
UPDATE ops_parceiro_ativacao SET status = $3, pago_ate = NULL, updated_at = NOW()
WHERE id_franqueado = $1 AND fp_fatura_id = $2 AND status = $4`,
		idFranqueado, faturaID, StatusAtivCancelada, StatusAtivPaga)
	if err == nil {
		n, _ := r1.RowsAffected()
		res.AtivacaoRevertida = int(n)
	}

	rows, err := d.QueryContext(ctx, excecaoSelectCols+`
FROM ops_parceiro_excecao
WHERE id_franqueado = $1 AND fp_fatura_id = $2 AND status = $3`, idFranqueado, faturaID, StatusAtivPaga)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		rec, err := scanExcecaoRow(rows)
		if err != nil || rec == nil {
			continue
		}
		if vinc := strings.TrimSpace(rec.IDVinculoAtivo); vinc != "" {
			_ = confservice.DesativarVinculoPorID(idFranqueado, rec.IDCliente, vinc)
			_ = pgcobranca.DesativarServicoPorRef(ctx, idFranqueado, "cs_parceiro_vinculo", vinc)
			res.VinculosDesativados++
		}
		_, _ = d.ExecContext(ctx, `
UPDATE ops_parceiro_excecao SET status = $2, pago_ate = NULL, updated_at = NOW()
WHERE id = $1`, rec.ID, StatusAtivCancelada)
		res.ExcecaoRevertida++
	}

	return res, nil
}

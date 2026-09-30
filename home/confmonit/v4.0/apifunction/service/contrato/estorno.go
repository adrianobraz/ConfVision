package contrato

import (
	"apifunction/auth"
	"apifunction/db"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EstornoResult struct {
	IDFaturaEstornada int `json:"id_fatura_estornada"`
	IDContrato        int `json:"id_contrato"`
	IDFaturaReaberta  int `json:"id_fatura_reaberta,omitempty"`
	ContratoStatus    string `json:"contrato_status"`
}

// EstornarPagamento reverte pagamento confirmado: fatura estornada, contrato suspenso,
// liberação revertida e nova fatura aberta para regularização.
func EstornarPagamento(sess auth.SessaoAdm, idFatura int, motivo string) (EstornoResult, error) {
	return estornarPagamentoInterno(sess, idFatura, motivo, false)
}

// EstornarPagamentoContabil estorna fatura MySQL pelo ID espelho Xano (Financeiro).
func EstornarPagamentoContabil(sess auth.SessaoAdm, idFaturaContabil int, motivo string, contabilJaEstornado bool) (EstornoResult, error) {
	if idFaturaContabil <= 0 {
		return EstornoResult{}, errors.New("id_fatura_contabil obrigatorio")
	}
	var idFatura int
	err := db.Conn.QueryRow(`
		SELECT ID_Fatura FROM fp_fatura_contrato WHERE ID_FaturaContabil = ?
	`, idFaturaContabil).Scan(&idFatura)
	if err == sql.ErrNoRows {
		return EstornoResult{}, errors.New("fatura contrato nao vinculada ao espelho contabil")
	}
	if err != nil {
		return EstornoResult{}, err
	}
	return estornarPagamentoInterno(sess, idFatura, motivo, contabilJaEstornado)
}

func estornarPagamentoInterno(sess auth.SessaoAdm, idFatura int, motivo string, contabilJaEstornado bool) (EstornoResult, error) {
	var out EstornoResult
	if idFatura <= 0 {
		return out, errors.New("id_fatura invalido")
	}

	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		motivo = "Estorno manual — módulo Contrato"
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return out, err
	}
	defer tx.Rollback()

	var idContrato int
	var idFra, status, tipo, referencia, ciclo string
	var valor float64
	var vencimento time.Time
	var idCentral, idRep string
	var idContabil sql.NullInt64
	var pagoEm sql.NullTime

	err = tx.QueryRow(`
		SELECT ID_Contrato, ID_Franqueado, ID_Central, COALESCE(ID_Representante,''),
		       Status, Tipo, Referencia, ValorTotal, VencimentoEm, CicloRef,
		       COALESCE(ID_FaturaContabil, 0), PagoEm
		FROM fp_fatura_contrato WHERE ID_Fatura = ? FOR UPDATE
	`, idFatura).Scan(
		&idContrato, &idFra, &idCentral, &idRep,
		&status, &tipo, &referencia, &valor, &vencimento, &ciclo,
		&idContabil, &pagoEm,
	)
	if err == sql.ErrNoRows {
		return out, errors.New("fatura nao encontrada")
	}
	if err != nil {
		return out, err
	}
	if err := AssertPodeEditarContrato(sess, idFra); err != nil {
		return out, err
	}
	if strings.ToLower(strings.TrimSpace(status)) != "paga" {
		return out, errors.New("somente faturas pagas podem ser estornadas")
	}

	idContabilVal := int(idContabil.Int64)
	if !contabilJaEstornado && idContabilVal > 0 {
		if err := erroContabilidade(estornarFaturaContabil(idContabilVal, motivo, sess)); err != nil {
			return out, err
		}
	}

	obs := motivo
	if pagoEm.Valid {
		obs = motivo + " | pago_em=" + pagoEm.Time.Format("2006-01-02 15:04:05")
	}

	_, err = tx.Exec(`
		UPDATE fp_fatura_contrato
		SET Status = 'estornada', Observacao = ?
		WHERE ID_Fatura = ?
	`, obs, idFatura)
	if err != nil {
		return out, err
	}

	var stContrato string
	if err = tx.QueryRow(`SELECT Status FROM fp_contrato WHERE ID_Contrato = ? FOR UPDATE`, idContrato).Scan(&stContrato); err != nil {
		return out, err
	}

	novoStatusContrato := stContrato
	switch strings.ToLower(strings.TrimSpace(stContrato)) {
	case "ativo", "aguardando_pagamento":
		novoStatusContrato = "suspenso"
		_, err = tx.Exec(`UPDATE fp_contrato SET Status = 'suspenso' WHERE ID_Contrato = ?`, idContrato)
		if err != nil {
			return out, err
		}
		if err = revertLiberacaoSeNecessarioTx(tx, idFra); err != nil {
			return out, err
		}
	}

	vencReabertura := time.Now().AddDate(0, 0, 7)
	if vencimento.After(vencReabertura) {
		vencReabertura = vencimento
	}
	cicloReab := fmt.Sprintf("%s-E%d", ciclo, idFatura)
	idNova, err := gerarFaturaReaberturaTx(tx, idContrato, idFra, idCentral, idRep, tipo, referencia, valor, vencReabertura, cicloReab)
	if err != nil {
		return out, err
	}

	detalheLog := fmt.Sprintf("estornada=%d nova_fatura=%d contrato=%s->%s", idFatura, idNova, stContrato, novoStatusContrato)
	_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Fatura, ID_Franqueado, Acao, Detalhe, ID_Usuario) VALUES (?,?,?,?,?,?)`,
		idContrato, idFatura, idFra, "pagamento_estornado", detalheLog, sess.IDUsuario)

	if err = tx.Commit(); err != nil {
		return out, err
	}

	out = EstornoResult{
		IDFaturaEstornada: idFatura,
		IDContrato:        idContrato,
		IDFaturaReaberta:  idNova,
		ContratoStatus:    novoStatusContrato,
	}

	var nome string
	_ = db.Conn.QueryRow(`SELECT Nome FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&nome)
	if err := espelharFaturaContabil(idContrato, idNova, nome, sess); err != nil {
		logContabilidadeOpcional("espelhar fatura reabertura pos-estorno", err)
	}

	return out, nil
}

func gerarFaturaReaberturaTx(tx *sql.Tx, idContrato int, idFra, idCentral, idRep, tipo, referencia string, valor float64, vencimento time.Time, cicloRef string) (int, error) {
	res, err := tx.Exec(`
		INSERT INTO fp_fatura_contrato
		(ID_Contrato, ID_Franqueado, ID_Central, ID_Representante, Tipo, Referencia, Status, ValorTotal, VencimentoEm, CicloRef, Observacao)
		VALUES (?,?,?,?,?,?,'aberta',?,?,?,?)
	`, idContrato, idFra, idCentral, nullStr(idRep), tipo, referencia, valor, vencimento, cicloRef, "Reaberta apos estorno")
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

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

func dataPassouOuHoje(iso string) bool {
	t, err := parseDate(iso)
	if err != nil {
		return false
	}
	now := time.Now()
	hoje := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	ref := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return !ref.After(hoje)
}

func FaturaPodeExcluir(status, vencimentoEm string) bool {
	st := strings.ToLower(strings.TrimSpace(status))
	if st == "aberta" || st == "vencida" {
		return true
	}
	if st == "paga" {
		return dataPassouOuHoje(vencimentoEm)
	}
	return false
}

func ContratoPodeExcluir(status, proximaCobrancaEm string) bool {
	st := strings.ToLower(strings.TrimSpace(status))
	if st == "rascunho" || st == "aguardando_pagamento" {
		return true
	}
	if st == "ativo" || st == "suspenso" {
		return dataPassouOuHoje(proximaCobrancaEm)
	}
	return false
}

func ExcluirFatura(sess auth.SessaoAdm, idFatura int) error {
	if idFatura <= 0 {
		return errors.New("id_fatura invalido")
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var idContrato int
	var idFra, status, vencimento string
	var idContabil sql.NullInt64
	err = tx.QueryRow(`
		SELECT ID_Contrato, ID_Franqueado, Status, DATE_FORMAT(VencimentoEm,'%Y-%m-%d'),
		       COALESCE(ID_FaturaContabil, 0)
		FROM fp_fatura_contrato WHERE ID_Fatura = ? FOR UPDATE
	`, idFatura).Scan(&idContrato, &idFra, &status, &vencimento, &idContabil)
	if err == sql.ErrNoRows {
		return errors.New("fatura nao encontrada")
	}
	if err != nil {
		return err
	}
	if err := AssertPodeEditarContrato(sess, idFra); err != nil {
		return err
	}
	if !FaturaPodeExcluir(status, vencimento) {
		if strings.ToLower(status) == "paga" {
			return errors.New("fatura paga so pode ser excluida apos o vencimento")
		}
		return errors.New("fatura nao pode ser excluida")
	}

	_, err = tx.Exec(`DELETE FROM fp_fatura_contrato WHERE ID_Fatura = ?`, idFatura)
	if err != nil {
		return err
	}

	var stContrato string
	if err = tx.QueryRow(`SELECT Status FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&stContrato); err == nil {
		if strings.ToLower(stContrato) == "aguardando_pagamento" {
			var abertas int
			_ = tx.QueryRow(`
				SELECT COUNT(*) FROM fp_fatura_contrato
				WHERE ID_Contrato = ? AND Status IN ('aberta','vencida')
			`, idContrato).Scan(&abertas)
			if abertas == 0 {
				_, _ = tx.Exec(`UPDATE fp_contrato SET Status = 'rascunho' WHERE ID_Contrato = ?`, idContrato)
			}
		}
	}

	_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Fatura, ID_Franqueado, Acao, Detalhe, ID_Usuario) VALUES (?,?,?,?,?,?)`,
		idContrato, idFatura, idFra, "fatura_excluida", fmt.Sprintf("status=%s", status), sess.IDUsuario)

	ref := faturaContabilRef{
		IDContrato:       idContrato,
		IDFatura:         idFatura,
		IDFaturaContabil: int(idContabil.Int64),
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	cancelarFaturasContabilidade([]faturaContabilRef{ref}, "Excluida — módulo Contrato", sess)
	return nil
}

func ExcluirContrato(sess auth.SessaoAdm, idContrato int) error {
	if idContrato <= 0 {
		return errors.New("id_contrato invalido")
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var idFra, status, nome string
	var prox sql.NullString
	err = tx.QueryRow(`
		SELECT ID_Franqueado, Status, Nome, DATE_FORMAT(ProximaCobrancaEm,'%Y-%m-%d')
		FROM fp_contrato WHERE ID_Contrato = ? FOR UPDATE
	`, idContrato).Scan(&idFra, &status, &nome, &prox)
	if err == sql.ErrNoRows {
		return errors.New("contrato nao encontrado")
	}
	if err != nil {
		return err
	}
	if err := AssertPodeEditarContrato(sess, idFra); err != nil {
		return err
	}

	proxStr := ""
	if prox.Valid {
		proxStr = prox.String
	}
	if !ContratoPodeExcluir(status, proxStr) {
		if strings.ToLower(status) == "ativo" || strings.ToLower(status) == "suspenso" {
			return errors.New("contrato pago so pode ser excluido apos o vencimento do periodo")
		}
		return errors.New("contrato nao pode ser excluido")
	}

	st := strings.ToLower(strings.TrimSpace(status))
	eraAtivo := st == "ativo" || st == "suspenso"

	faturasContabil, _ := listarTodasFaturasContabil(idContrato, tx)

	_, _ = tx.Exec(`DELETE FROM fp_contrato_uso WHERE ID_Contrato = ?`, idContrato)
	_, err = tx.Exec(`DELETE FROM fp_fatura_contrato WHERE ID_Contrato = ?`, idContrato)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM fp_contrato WHERE ID_Contrato = ?`, idContrato)
	if err != nil {
		return err
	}

	_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Franqueado, Acao, Detalhe, ID_Usuario) VALUES (?,?,?,?,?)`,
		idContrato, idFra, "contrato_excluido", nome, sess.IDUsuario)

	if eraAtivo {
		if err = revertLiberacaoSeNecessarioTx(tx, idFra); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	cancelarFaturasContabilidade(faturasContabil, "Contrato excluido — módulo Contrato", sess)
	return nil
}

func listarTodasFaturasContabil(idContrato int, tx *sql.Tx) ([]faturaContabilRef, error) {
	rows, err := tx.Query(`
		SELECT ID_Fatura, COALESCE(ID_FaturaContabil, 0)
		FROM fp_fatura_contrato WHERE ID_Contrato = ?
	`, idContrato)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []faturaContabilRef
	for rows.Next() {
		var f faturaContabilRef
		f.IDContrato = idContrato
		if err := rows.Scan(&f.IDFatura, &f.IDFaturaContabil); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func revertLiberacaoSeNecessarioTx(tx *sql.Tx, idFra string) error {
	var n int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM fp_contrato WHERE ID_Franqueado = ? AND Status = 'ativo'
	`, idFra).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := tx.Exec(`UPDATE franqueado SET UsaConfVision = 'N' WHERE ID_Franqueado = ?`, idFra)
	return err
}

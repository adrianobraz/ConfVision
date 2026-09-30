package contrato

import (
	"database/sql"
	"errors"
	"strings"
)

// ContratoEditavel rascunho sempre; aguardando_pagamento só com fatura aberta/vencida.
func ContratoEditavel(status string, idContrato int, q dbQuerier) (bool, error) {
	st := strings.ToLower(strings.TrimSpace(status))
	if st == "rascunho" {
		return true, nil
	}
	if st != "aguardando_pagamento" || idContrato <= 0 {
		return false, nil
	}
	var n int
	err := q.QueryRow(`
		SELECT COUNT(*) FROM fp_fatura_contrato
		WHERE ID_Contrato = ? AND Status IN ('aberta','vencida')
	`, idContrato).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func assertContratoEditavel(status string, idContrato int, q dbQuerier) error {
	ok, err := ContratoEditavel(status, idContrato, q)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("contrato nao pode ser editado")
	}
	return nil
}

func cancelarFaturasAbertasTx(tx *sql.Tx, idContrato int) (int64, error) {
	res, err := tx.Exec(`
		UPDATE fp_fatura_contrato SET Status = 'cancelada'
		WHERE ID_Contrato = ? AND Status IN ('aberta','vencida')
	`, idContrato)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type dbQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
}

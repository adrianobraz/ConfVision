package contrato

import (
	"database/sql"
	"fmt"
	"strings"
)

// contratoVigente retorna o contrato não cancelado do franqueado, se existir.
func contratoVigente(q dbQuerier, idFranqueado string) (idContrato int, status string, err error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return 0, "", nil
	}
	err = q.QueryRow(`
		SELECT ID_Contrato, Status FROM fp_contrato
		WHERE ID_Franqueado = ? AND Status != 'cancelado'
		ORDER BY FIELD(Status,'ativo','aguardando_pagamento','suspenso','rascunho'), ID_Contrato DESC
		LIMIT 1
	`, idFranqueado).Scan(&idContrato, &status)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return idContrato, status, nil
}

func assertNovoContratoPermitido(q dbQuerier, idFranqueado string) error {
	id, st, err := contratoVigente(q, idFranqueado)
	if err != nil {
		return err
	}
	if id <= 0 {
		return nil
	}
	return fmt.Errorf("franqueado ja possui contrato #%d (%s). Edite o existente ou exclua antes de criar outro", id, st)
}

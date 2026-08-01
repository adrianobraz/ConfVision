package auxiliar

import (
	"database/sql"
	"errors"
	"strings"
)

func FranqueadoDoCliente(idCliente string) (string, error) {
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return "", errors.New("idCliente obrigatorio")
	}

	db, err := Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var id sql.NullString
	err = db.QueryRow(`
		SELECT cliente.ID_Franqueado
		FROM cliente
		WHERE cliente.ID_Cliente = ?
	`, idCliente).Scan(&id)
	if err == sql.ErrNoRows {
		return "", errors.New("cliente nao encontrado")
	}
	if err != nil {
		return "", err
	}
	if !id.Valid || strings.TrimSpace(id.String) == "" {
		return "", errors.New("cliente sem franqueado")
	}
	return strings.TrimSpace(id.String), nil
}

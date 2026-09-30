package visdata

import (
	"confvision/src/conexao"
	"confvision/src/config"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// GetClienteByCodigoInterno resolve ID_Cliente pelo CodigoInterno (+ CodEmpresa opcional).
func GetClienteByCodigoInterno(ctx context.Context, idFranqueado, codigoInterno, codEmpresa string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	codigoInterno = strings.TrimSpace(codigoInterno)
	codEmpresa = strings.TrimSpace(codEmpresa)
	if idFranqueado == "" || codigoInterno == "" {
		return "", nil
	}
	if config.ConexaoMySQL == "" {
		return "", nil
	}

	db, err := conexao.Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var idCliente sql.NullString
	if codEmpresa != "" {
		err = db.QueryRowContext(ctxQ, `
SELECT ID_Cliente
FROM cliente
WHERE ID_Franqueado = ? AND UPPER(TRIM(CodigoInterno)) = UPPER(TRIM(?))
  AND TRIM(COALESCE(CodEmpresa, '')) = ?
LIMIT 1`, idFranqueado, codigoInterno, codEmpresa).Scan(&idCliente)
	} else {
		err = db.QueryRowContext(ctxQ, `
SELECT ID_Cliente
FROM cliente
WHERE ID_Franqueado = ? AND UPPER(TRIM(CodigoInterno)) = UPPER(TRIM(?))
LIMIT 1`, idFranqueado, codigoInterno).Scan(&idCliente)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if idCliente.Valid {
		return strings.TrimSpace(idCliente.String), nil
	}
	return "", nil
}

// GetClienteCodigoInterno le CodigoInterno da tabela cliente no MySQL legado.
func GetClienteCodigoInterno(ctx context.Context, idFranqueado, idCliente string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return "", nil
	}
	if config.ConexaoMySQL == "" {
		return "", nil
	}

	db, err := conexao.Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var codigo sql.NullString
	if idFranqueado != "" {
		err = db.QueryRowContext(ctxQ, `
SELECT CodigoInterno
FROM cliente
WHERE ID_Cliente = ? AND ID_Franqueado = ?
LIMIT 1`, idCliente, idFranqueado).Scan(&codigo)
	} else {
		err = db.QueryRowContext(ctxQ, `
SELECT CodigoInterno
FROM cliente
WHERE ID_Cliente = ?
LIMIT 1`, idCliente).Scan(&codigo)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if codigo.Valid {
		return strings.TrimSpace(codigo.String), nil
	}
	return "", nil
}

// GetClienteCodEmpresa le CodEmpresa da tabela cliente no MySQL legado.
func GetClienteCodEmpresa(ctx context.Context, idFranqueado, idCliente string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return "", nil
	}
	if config.ConexaoMySQL == "" {
		return "", nil
	}

	db, err := conexao.Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var codigo sql.NullString
	if idFranqueado != "" {
		err = db.QueryRowContext(ctxQ, `
SELECT CodEmpresa
FROM cliente
WHERE ID_Cliente = ? AND ID_Franqueado = ?
LIMIT 1`, idCliente, idFranqueado).Scan(&codigo)
	} else {
		err = db.QueryRowContext(ctxQ, `
SELECT CodEmpresa
FROM cliente
WHERE ID_Cliente = ?
LIMIT 1`, idCliente).Scan(&codigo)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if codigo.Valid {
		return strings.TrimSpace(codigo.String), nil
	}
	return "", nil
}

// GetClienteNome le Nome da tabela cliente no MySQL legado.
func GetClienteNome(ctx context.Context, idFranqueado, idCliente string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return "", nil
	}
	if config.ConexaoMySQL == "" {
		return "", nil
	}

	db, err := conexao.Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var nome sql.NullString
	if idFranqueado != "" {
		err = db.QueryRowContext(ctxQ, `
SELECT Nome
FROM cliente
WHERE ID_Cliente = ? AND ID_Franqueado = ?
LIMIT 1`, idCliente, idFranqueado).Scan(&nome)
	} else {
		err = db.QueryRowContext(ctxQ, `
SELECT Nome
FROM cliente
WHERE ID_Cliente = ?
LIMIT 1`, idCliente).Scan(&nome)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if nome.Valid {
		return strings.TrimSpace(nome.String), nil
	}
	return "", nil
}

// SearchClienteIDsByNome busca ID_Cliente no MySQL legado por trecho do nome (case-insensitive).
func SearchClienteIDsByNome(ctx context.Context, idFranqueado, termo string) ([]string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	termo = strings.TrimSpace(termo)
	if termo == "" {
		return nil, nil
	}
	if config.ConexaoMySQL == "" {
		return nil, nil
	}

	db, err := conexao.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	like := "%" + termo + "%"
	var rows *sql.Rows
	if idFranqueado != "" {
		rows, err = db.QueryContext(ctxQ, `
SELECT ID_Cliente
FROM cliente
WHERE ID_Franqueado = ? AND Nome LIKE ?
ORDER BY Nome
LIMIT 100`, idFranqueado, like)
	} else {
		rows, err = db.QueryContext(ctxQ, `
SELECT ID_Cliente
FROM cliente
WHERE Nome LIKE ?
ORDER BY Nome
LIMIT 100`, like)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id sql.NullString
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if id.Valid {
			if v := strings.TrimSpace(id.String); v != "" {
				ids = append(ids, v)
			}
		}
	}
	return ids, rows.Err()
}

// SetClienteCodigoInterno grava CodigoInterno na tabela cliente (MySQL).
func SetClienteCodigoInterno(ctx context.Context, idFranqueado, idCliente, codigo string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return errors.New("id_cliente obrigatorio")
	}
	if config.ConexaoMySQL == "" {
		return errors.New("mysql nao configurado")
	}

	db, err := conexao.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	var codigoVal interface{}
	if codigo == "" {
		codigoVal = nil
	} else {
		codigoVal = codigo
	}

	if idFranqueado != "" {
		_, err = db.ExecContext(ctxQ, `
UPDATE cliente
SET CodigoInterno = ?
WHERE ID_Cliente = ? AND ID_Franqueado = ?`, codigoVal, idCliente, idFranqueado)
	} else {
		_, err = db.ExecContext(ctxQ, `
UPDATE cliente
SET CodigoInterno = ?
WHERE ID_Cliente = ?`, codigoVal, idCliente)
	}
	return err
}

// SetClienteCodEmpresa grava CodEmpresa na tabela cliente (MySQL).
func SetClienteCodEmpresa(ctx context.Context, idFranqueado, idCliente, codigo string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return errors.New("id_cliente obrigatorio")
	}
	if config.ConexaoMySQL == "" {
		return errors.New("mysql nao configurado")
	}

	db, err := conexao.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	var codigoVal interface{}
	if codigo == "" {
		codigoVal = nil
	} else {
		codigoVal = codigo
	}

	if idFranqueado != "" {
		_, err = db.ExecContext(ctxQ, `
UPDATE cliente
SET CodEmpresa = ?
WHERE ID_Cliente = ? AND ID_Franqueado = ?`, codigoVal, idCliente, idFranqueado)
	} else {
		_, err = db.ExecContext(ctxQ, `
UPDATE cliente
SET CodEmpresa = ?
WHERE ID_Cliente = ?`, codigoVal, idCliente)
	}
	return err
}

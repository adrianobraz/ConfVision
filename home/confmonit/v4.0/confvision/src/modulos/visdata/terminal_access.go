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

// FranqueadoTemTerminal consulta UsuarioTeminal=S do usuario master no MySQL legado
// (mesma regra do WebTerminal; sincronizado pelo FP via setUsuarioMasterTerminal).
func FranqueadoTemTerminal(ctx context.Context, idFranqueado string) (bool, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return false, nil
	}
	if config.ConexaoMySQL == "" {
		return true, nil
	}

	db, err := conexao.Conectar()
	if err != nil {
		return false, err
	}
	defer db.Close()

	ctxQ, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var usa sql.NullString
	err = db.QueryRowContext(ctxQ, `
SELECT COALESCE(u.UsuarioTeminal, '')
FROM franqueado f
INNER JOIN usuarios u ON u.ID_Usuario = f.ID_UsuarioMaster
WHERE f.ID_Franqueado = ?
LIMIT 1`, idFranqueado).Scan(&usa)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(usa.String), "S"), nil
}

package visdata

import (
	"context"
	"database/sql"
	"strings"
)

// SnapshotURLImagemLiberada retorna snapshot_url apenas se imagem foi liberada apos envio Moni.
func SnapshotURLImagemLiberada(ctx context.Context, eventoID int, codigoPublico string) (string, error) {
	db, err := DB()
	if err != nil {
		return "", err
	}
	var snap sql.NullString
	var codigo sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT snapshot_url, codigo_imagem_publico
FROM vis_evento
WHERE id = $1
  AND imagem_liberada_em IS NOT NULL
  AND snapshot_url IS NOT NULL
  AND snapshot_url <> ''`, eventoID).Scan(&snap, &codigo)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if !snap.Valid || strings.TrimSpace(snap.String) == "" {
		return "", nil
	}
	if codigo.Valid && strings.TrimSpace(codigo.String) != "" {
		if strings.TrimSpace(codigo.String) != strings.TrimSpace(codigoPublico) {
			return "", nil
		}
	}
	return strings.TrimSpace(snap.String), nil
}

func LiberarImagemEvento(ctx context.Context, eventoID int, codigoPublico string, integracaoID int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
UPDATE vis_evento
SET codigo_imagem_publico = $2,
    imagem_liberada_em = NOW(),
    vis_integracao_imagem_id = $3
WHERE id = $1`, eventoID, codigoPublico, integracaoID)
	return err
}


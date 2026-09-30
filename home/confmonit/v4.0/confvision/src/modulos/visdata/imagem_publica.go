package visdata

import (
	"context"
	"database/sql"
	"fmt"
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

func SetClienteCodigoInterno(ctx context.Context, idFranqueado, idCliente, codigo string) error {
	_, err := PutClienteExt(ctx, map[string]any{
		"id_franqueado":  idFranqueado,
		"id_cliente":     idCliente,
		"codigo_interno": codigo,
	})
	return err
}

func GetClienteCodigoInterno(ctx context.Context, idFranqueado, idCliente string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idFranqueado == "" || idCliente == "" {
		return "", nil
	}
	db, err := DB()
	if err != nil {
		return "", err
	}
	var codigo sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT codigo_interno FROM vis_cliente_ext
WHERE id_franqueado = $1 AND id_cliente = $2`, idFranqueado, idCliente).Scan(&codigo)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if codigo.Valid {
		return strings.TrimSpace(codigo.String), nil
	}
	return "", nil
}

func GetClienteExt(ctx context.Context, idFranqueado, idCliente string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var codigo sql.NullString
	var updated sql.NullTime
	err = db.QueryRowContext(ctx, `
SELECT codigo_interno, updated_at FROM vis_cliente_ext
WHERE id_franqueado = $1 AND id_cliente = $2`, idFranqueado, idCliente).Scan(&codigo, &updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]any{
				"id_franqueado":  idFranqueado,
				"id_cliente":     idCliente,
				"codigo_interno": nil,
			}, nil
		}
		return nil, err
	}
	return map[string]any{
		"id_franqueado":  idFranqueado,
		"id_cliente":     idCliente,
		"codigo_interno": nullStr(codigo),
		"updated_at":     nullTime(updated),
	}, nil
}

func PutClienteExt(ctx context.Context, input map[string]any) (map[string]any, error) {
	idFra := strVal(input, "id_franqueado")
	idCli := strVal(input, "id_cliente")
	if idFra == "" || idCli == "" {
		return nil, fmt.Errorf("id_franqueado e id_cliente obrigatorios")
	}
	codigo := strings.TrimSpace(strVal(input, "codigo_interno"))
	db, err := DB()
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO vis_cliente_ext (id_franqueado, id_cliente, codigo_interno, updated_at)
VALUES ($1, $2, NULLIF($3, ''), NOW())
ON CONFLICT (id_franqueado, id_cliente)
DO UPDATE SET codigo_interno = EXCLUDED.codigo_interno, updated_at = NOW()`,
		idFra, idCli, codigo)
	if err != nil {
		return nil, err
	}
	return GetClienteExt(ctx, idFra, idCli)
}

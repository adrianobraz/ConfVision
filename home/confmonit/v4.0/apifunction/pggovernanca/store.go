package pggovernanca

import (
	"database/sql"
	"time"

	"apifunction/pgcredito"
)

func UpsertRepasse(r RepasseRow) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	venc := tsFromUnix(r.Vencimento)
	_, err = db.Exec(`
		INSERT INTO fp_gov_repasse (fatura_id, tipo, id_representante, id_central, status, vencimento_em, valor_total, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (fatura_id) DO UPDATE SET
			tipo = EXCLUDED.tipo,
			id_representante = EXCLUDED.id_representante,
			id_central = EXCLUDED.id_central,
			status = EXCLUDED.status,
			vencimento_em = EXCLUDED.vencimento_em,
			valor_total = EXCLUDED.valor_total,
			synced_at = NOW()
	`, r.FaturaID, r.Tipo, r.IDRepresentante, r.IDCentral, r.Status, venc, r.ValorTotal)
	return err
}

func MarkRepasseStatus(faturaID int, status string) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE fp_gov_repasse SET status = $2, synced_at = NOW() WHERE fatura_id = $1`, faturaID, status)
	return err
}

func UpsertRestricao(row RestricaoRow) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	venc := tsFromUnix(row.Vencimento)
	_, err = db.Exec(`
		INSERT INTO fp_gov_restricao (entidade_tipo, entidade_id, fatura_id, nivel, vencimento_em, dias_restantes, ativo, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (entidade_tipo, entidade_id) DO UPDATE SET
			fatura_id = EXCLUDED.fatura_id,
			nivel = EXCLUDED.nivel,
			vencimento_em = EXCLUDED.vencimento_em,
			dias_restantes = EXCLUDED.dias_restantes,
			ativo = EXCLUDED.ativo,
			updated_at = NOW()
	`, row.EntidadeTipo, row.EntidadeID, row.FaturaID, row.Nivel, venc, row.DiasRestantes, row.Ativo)
	return err
}

func ClearRestricao(entidadeTipo, entidadeID string) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		UPDATE fp_gov_restricao SET ativo = FALSE, nivel = $3, dias_restantes = 0, updated_at = NOW()
		WHERE entidade_tipo = $1 AND entidade_id = $2
	`, entidadeTipo, entidadeID, NivelOK)
	return err
}

func GetRestricaoAtiva(entidadeTipo, entidadeID string) (*RestricaoRow, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	var row RestricaoRow
	var venc sql.NullTime
	err = db.QueryRow(`
		SELECT entidade_tipo, entidade_id, fatura_id, nivel, vencimento_em, dias_restantes, ativo
		FROM fp_gov_restricao
		WHERE entidade_tipo = $1 AND entidade_id = $2 AND ativo = TRUE
		LIMIT 1
	`, entidadeTipo, entidadeID).Scan(
		&row.EntidadeTipo, &row.EntidadeID, &row.FaturaID, &row.Nivel, &venc, &row.DiasRestantes, &row.Ativo,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if venc.Valid {
		row.Vencimento = venc.Time.UnixMilli()
	}
	return &row, nil
}

func FranqueadoBloqueadoCascata(idFranqueado string) (bool, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return false, err
	}
	var n int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM fp_gov_suspensao_cascata
		WHERE id_franqueado = $1 AND ativo = TRUE
	`, idFranqueado).Scan(&n)
	return n > 0, err
}

func RegistrarSuspensaoCascata(idFranqueado string, assinaturaID int, idRep string, faturaRepasseID int, motivoInterno string) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO fp_gov_suspensao_cascata (id_franqueado, assinatura_id, id_representante, fatura_repasse_id, motivo_interno, ativo)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		ON CONFLICT (assinatura_id, fatura_repasse_id) DO NOTHING
	`, idFranqueado, assinaturaID, idRep, faturaRepasseID, motivoInterno)
	return err
}

func ReativarSuspensoesPorRepasse(faturaRepasseID int) ([]int, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT assinatura_id FROM fp_gov_suspensao_cascata
		WHERE fatura_repasse_id = $1 AND ativo = TRUE
	`, faturaRepasseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		UPDATE fp_gov_suspensao_cascata SET ativo = FALSE, reativado_em = NOW()
		WHERE fatura_repasse_id = $1 AND ativo = TRUE
	`, faturaRepasseID)
	return ids, err
}

func ListarRepasseAbertosVencidos() ([]RepasseRow, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT fatura_id, tipo, id_representante, id_central, status, vencimento_em, valor_total
		FROM fp_gov_repasse
		WHERE status = 'aberta' AND vencimento_em IS NOT NULL AND vencimento_em < NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRepasseRows(rows)
}

func scanRepasseRows(rows *sql.Rows) ([]RepasseRow, error) {
	var out []RepasseRow
	for rows.Next() {
		var r RepasseRow
		var venc sql.NullTime
		if err := rows.Scan(&r.FaturaID, &r.Tipo, &r.IDRepresentante, &r.IDCentral, &r.Status, &venc, &r.ValorTotal); err != nil {
			return nil, err
		}
		if venc.Valid {
			r.Vencimento = venc.Time.UnixMilli()
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func tsFromUnix(ms int64) sql.NullTime {
	if ms <= 0 {
		return sql.NullTime{}
	}
	if ms > 1e12 {
		return sql.NullTime{Time: time.UnixMilli(ms), Valid: true}
	}
	return sql.NullTime{Time: time.Unix(ms, 0), Valid: true}
}

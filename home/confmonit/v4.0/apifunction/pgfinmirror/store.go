package pgfinmirror

import (
	"context"
	"database/sql"
	"fmt"
)

func upsertFaturaOne(ctx context.Context, tx *sql.Tx, row map[string]any) error {
	id := intFromAny(row["id"])
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fp_fin_fatura (
			id, created_at, id_franqueado, id_representante, id_central, referencia,
			status, tipo, valor_total, vencimento_em, pago_em, ciclo_ref, observacao, synced_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
		ON CONFLICT (id) DO UPDATE SET
			created_at = EXCLUDED.created_at,
			id_franqueado = EXCLUDED.id_franqueado,
			id_representante = EXCLUDED.id_representante,
			id_central = EXCLUDED.id_central,
			referencia = EXCLUDED.referencia,
			status = EXCLUDED.status,
			tipo = EXCLUDED.tipo,
			valor_total = EXCLUDED.valor_total,
			vencimento_em = EXCLUDED.vencimento_em,
			pago_em = EXCLUDED.pago_em,
			ciclo_ref = EXCLUDED.ciclo_ref,
			observacao = EXCLUDED.observacao,
			synced_at = NOW()
	`,
		id, timeFromAny(row["created_at"]),
		stringFromAny(row["id_franqueado"]), stringFromAny(row["id_representante"]), stringFromAny(row["id_central"]),
		stringFromAny(row["referencia"]), stringFromAny(row["status"]), stringFromAny(row["tipo"]),
		floatFromAny(row["valor_total"]), timeFromAny(row["vencimento_em"]), timeFromAny(row["pago_em"]),
		stringFromAny(row["ciclo_ref"]), stringFromAny(row["observacao"]),
	)
	return err
}

func deleteFatura(ctx context.Context, tx *sql.Tx, id int) error {
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_fatura WHERE id = $1`, id)
	return err
}

func upsertPagamentoOne(ctx context.Context, tx *sql.Tx, row map[string]any) error {
	id := intFromAny(row["id"])
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fp_fin_pagamento (id, created_at, fp_fatura_id, valor, metodo, pago_em, observacao, synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		ON CONFLICT (id) DO UPDATE SET
			created_at = EXCLUDED.created_at,
			fp_fatura_id = EXCLUDED.fp_fatura_id,
			valor = EXCLUDED.valor,
			metodo = EXCLUDED.metodo,
			pago_em = EXCLUDED.pago_em,
			observacao = EXCLUDED.observacao,
			synced_at = NOW()
	`,
		id, timeFromAny(row["created_at"]), intFromAny(row["fp_fatura_id"]),
		floatFromAny(row["valor"]), stringFromAny(row["metodo"]), timeFromAny(row["pago_em"]),
		stringFromAny(row["observacao"]),
	)
	return err
}

func deletePagamento(ctx context.Context, tx *sql.Tx, id int) error {
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_pagamento WHERE id = $1`, id)
	return err
}

func upsertAssinaturaOne(ctx context.Context, tx *sql.Tx, row map[string]any) error {
	id := intFromAny(row["id"])
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fp_fin_assinatura (
			id, created_at, id_franqueado, id_representante, id_central,
			produto, plano, status, valor, observacao, synced_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
		ON CONFLICT (id) DO UPDATE SET
			created_at = EXCLUDED.created_at,
			id_franqueado = EXCLUDED.id_franqueado,
			id_representante = EXCLUDED.id_representante,
			id_central = EXCLUDED.id_central,
			produto = EXCLUDED.produto,
			plano = EXCLUDED.plano,
			status = EXCLUDED.status,
			valor = EXCLUDED.valor,
			observacao = EXCLUDED.observacao,
			synced_at = NOW()
	`,
		id, timeFromAny(row["created_at"]),
		stringFromAny(row["id_franqueado"]), stringFromAny(row["id_representante"]), stringFromAny(row["id_central"]),
		stringFromAny(row["produto"]), stringFromAny(row["plano"]), stringFromAny(row["status"]),
		floatFromAny(row["valor"]), stringFromAny(row["observacao"]),
	)
	return err
}

func deleteAssinatura(ctx context.Context, tx *sql.Tx, id int) error {
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_assinatura WHERE id = $1`, id)
	return err
}

func upsertCaixaOne(ctx context.Context, tx *sql.Tx, row map[string]any) error {
	id := intFromAny(row["id"])
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fp_fin_caixa_movimento (id, created_at, tipo, valor, descricao, movimento_em, admin_usuario, synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		ON CONFLICT (id) DO UPDATE SET
			created_at = EXCLUDED.created_at,
			tipo = EXCLUDED.tipo,
			valor = EXCLUDED.valor,
			descricao = EXCLUDED.descricao,
			movimento_em = EXCLUDED.movimento_em,
			admin_usuario = EXCLUDED.admin_usuario,
			synced_at = NOW()
	`,
		id, timeFromAny(row["created_at"]), stringFromAny(row["tipo"]),
		floatFromAny(row["valor"]), stringFromAny(row["descricao"]), timeFromAny(row["movimento_em"]),
		stringFromAny(row["admin_usuario"]),
	)
	return err
}

func upsertContaPagarOne(ctx context.Context, tx *sql.Tx, row map[string]any) error {
	id := intFromAny(row["id"])
	if id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fp_fin_conta_pagar (
			id, created_at, fornecedor, descricao, valor, status, vencimento_em, pago_em, admin_usuario, synced_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (id) DO UPDATE SET
			created_at = EXCLUDED.created_at,
			fornecedor = EXCLUDED.fornecedor,
			descricao = EXCLUDED.descricao,
			valor = EXCLUDED.valor,
			status = EXCLUDED.status,
			vencimento_em = EXCLUDED.vencimento_em,
			pago_em = EXCLUDED.pago_em,
			admin_usuario = EXCLUDED.admin_usuario,
			synced_at = NOW()
	`,
		id, timeFromAny(row["created_at"]), stringFromAny(row["fornecedor"]), stringFromAny(row["descricao"]),
		floatFromAny(row["valor"]), stringFromAny(row["status"]), timeFromAny(row["vencimento_em"]),
		timeFromAny(row["pago_em"]), stringFromAny(row["admin_usuario"]),
	)
	return err
}

func upsertFaturas(ctx context.Context, tx *sql.Tx, rows []map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_fatura`); err != nil {
		return err
	}
	stmt := `
		INSERT INTO fp_fin_fatura (
			id, created_at, id_franqueado, id_representante, id_central, referencia,
			status, tipo, valor_total, vencimento_em, pago_em, ciclo_ref, observacao, synced_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
	`
	for _, row := range rows {
		id := intFromAny(row["id"])
		if id <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt,
			id, timeFromAny(row["created_at"]),
			stringFromAny(row["id_franqueado"]), stringFromAny(row["id_representante"]), stringFromAny(row["id_central"]),
			stringFromAny(row["referencia"]), stringFromAny(row["status"]), stringFromAny(row["tipo"]),
			floatFromAny(row["valor_total"]), timeFromAny(row["vencimento_em"]), timeFromAny(row["pago_em"]),
			stringFromAny(row["ciclo_ref"]), stringFromAny(row["observacao"]),
		); err != nil {
			return err
		}
	}
	return nil
}

func upsertPagamentos(ctx context.Context, tx *sql.Tx, rows []map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_pagamento`); err != nil {
		return err
	}
	stmt := `
		INSERT INTO fp_fin_pagamento (id, created_at, fp_fatura_id, valor, metodo, pago_em, observacao, synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
	`
	for _, row := range rows {
		id := intFromAny(row["id"])
		if id <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt,
			id, timeFromAny(row["created_at"]), intFromAny(row["fp_fatura_id"]),
			floatFromAny(row["valor"]), stringFromAny(row["metodo"]), timeFromAny(row["pago_em"]),
			stringFromAny(row["observacao"]),
		); err != nil {
			return err
		}
	}
	return nil
}

func upsertAssinaturas(ctx context.Context, tx *sql.Tx, rows []map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_assinatura`); err != nil {
		return err
	}
	stmt := `
		INSERT INTO fp_fin_assinatura (
			id, created_at, id_franqueado, id_representante, id_central,
			produto, plano, status, valor, observacao, synced_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
	`
	for _, row := range rows {
		id := intFromAny(row["id"])
		if id <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt,
			id, timeFromAny(row["created_at"]),
			stringFromAny(row["id_franqueado"]), stringFromAny(row["id_representante"]), stringFromAny(row["id_central"]),
			stringFromAny(row["produto"]), stringFromAny(row["plano"]), stringFromAny(row["status"]),
			floatFromAny(row["valor"]), stringFromAny(row["observacao"]),
		); err != nil {
			return err
		}
	}
	return nil
}

func upsertCaixa(ctx context.Context, tx *sql.Tx, rows []map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_caixa_movimento`); err != nil {
		return err
	}
	stmt := `
		INSERT INTO fp_fin_caixa_movimento (id, created_at, tipo, valor, descricao, movimento_em, admin_usuario, synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
	`
	for _, row := range rows {
		id := intFromAny(row["id"])
		if id <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt,
			id, timeFromAny(row["created_at"]), stringFromAny(row["tipo"]),
			floatFromAny(row["valor"]), stringFromAny(row["descricao"]), timeFromAny(row["movimento_em"]),
			stringFromAny(row["admin_usuario"]),
		); err != nil {
			return err
		}
	}
	return nil
}

func upsertContasPagar(ctx context.Context, tx *sql.Tx, rows []map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM fp_fin_conta_pagar`); err != nil {
		return err
	}
	stmt := `
		INSERT INTO fp_fin_conta_pagar (
			id, created_at, fornecedor, descricao, valor, status, vencimento_em, pago_em, admin_usuario, synced_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
	`
	for _, row := range rows {
		id := intFromAny(row["id"])
		if id <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt,
			id, timeFromAny(row["created_at"]), stringFromAny(row["fornecedor"]), stringFromAny(row["descricao"]),
			floatFromAny(row["valor"]), stringFromAny(row["status"]), timeFromAny(row["vencimento_em"]),
			timeFromAny(row["pago_em"]), stringFromAny(row["admin_usuario"]),
		); err != nil {
			return err
		}
	}
	return nil
}

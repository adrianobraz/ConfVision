package pgcredito

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const (
	TipoMovBonificacao = "bonificacao"
	TipoMovRecarga     = "recarga"
	TipoMovDebitoUso   = "debito_uso"

	CanalCredito = "credito"
)

var errPostgresNaoConfigurado = errors.New("POSTGRES_URL nao configurado no apifunction")

type DebitarResult struct {
	JaDebitado bool
}

func movimentoCanal(servicoCanal, tipo string) string {
	servicoCanal = strings.TrimSpace(servicoCanal)
	tipo = strings.TrimSpace(tipo)
	if tipo == TipoMovRecarga || tipo == TipoMovBonificacao {
		return CanalCredito
	}
	if servicoCanal == "" || servicoCanal == CanalCredito {
		return CanalCredito
	}
	return servicoCanal
}

func CreditarSaldo(ctx context.Context, idFranqueado, servicoCanal, tipo string, valor float64, idFatura *int64, obs, criadoPor string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	tipo = strings.TrimSpace(tipo)
	if idFranqueado == "" || tipo == "" {
		return errors.New("id_franqueado e tipo obrigatorios")
	}
	if valor <= 0 {
		return errors.New("valor deve ser positivo")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	movCanal := movimentoCanal(servicoCanal, tipo)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO ops_credito_saldo (id_franqueado, canal, saldo, saldo_inicial, updated_at)
VALUES ($1,$2,$3,$3,NOW())
ON CONFLICT (id_franqueado, canal) DO UPDATE SET
  saldo = ops_credito_saldo.saldo + EXCLUDED.saldo,
  updated_at = NOW()`, idFranqueado, CanalCredito, valor)
	if err != nil {
		return err
	}

	if tipo == TipoMovRecarga || tipo == TipoMovBonificacao {
		_, err = tx.ExecContext(ctx, `
UPDATE ops_credito_saldo SET saldo_inicial = saldo, updated_at = NOW()
WHERE id_franqueado = $1 AND canal = $2`, idFranqueado, CanalCredito)
		if err != nil {
			return err
		}
	}

	var idFat any
	if idFatura != nil && *idFatura > 0 {
		idFat = *idFatura
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO ops_credito_movimento
  (id_franqueado, canal, tipo, quantidade, valor_total, id_fatura, observacao, criado_por)
VALUES ($1,$2,$3,$4,$4,$5,$6,$7)`,
		idFranqueado, movCanal, tipo, valor, idFat, obs, criadoPor)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func DebitarSaldo(ctx context.Context, idFranqueado, servicoCanal string, valor float64, idProcesso, obs string) (DebitarResult, error) {
	out := DebitarResult{}
	idFranqueado = strings.TrimSpace(idFranqueado)
	servicoCanal = strings.TrimSpace(servicoCanal)
	idProcesso = strings.TrimSpace(idProcesso)
	if idFranqueado == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	if servicoCanal == "" {
		servicoCanal = "ligacao"
	}
	if valor <= 0 {
		return out, errors.New("valor deve ser positivo")
	}
	db, err := DB()
	if err != nil {
		return out, err
	}

	if idProcesso != "" {
		var n int
		err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_credito_movimento
WHERE id_processo = $1 AND tipo = $2`, idProcesso, TipoMovDebitoUso).Scan(&n)
		if err != nil {
			return out, err
		}
		if n > 0 {
			out.JaDebitado = true
			return out, nil
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()

	var saldo float64
	err = tx.QueryRowContext(ctx, `
SELECT COALESCE(saldo,0) FROM ops_credito_saldo
WHERE id_franqueado = $1 AND canal = $2 FOR UPDATE`, idFranqueado, CanalCredito).Scan(&saldo)
	if errors.Is(err, sql.ErrNoRows) {
		return out, fmt.Errorf("sem saldo de credito")
	}
	if err != nil {
		return out, err
	}
	if saldo < valor {
		return out, fmt.Errorf("saldo insuficiente: %.4f < %.4f", saldo, valor)
	}
	_, err = tx.ExecContext(ctx, `
UPDATE ops_credito_saldo SET saldo = saldo - $3, updated_at = NOW()
WHERE id_franqueado = $1 AND canal = $2`, idFranqueado, CanalCredito, valor)
	if err != nil {
		return out, err
	}
	var idProc any
	if idProcesso != "" {
		idProc = idProcesso
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO ops_credito_movimento
  (id_franqueado, canal, tipo, quantidade, valor_total, id_processo, observacao, criado_por)
VALUES ($1,$2,$3,$4,$4,$5,$6,'sistema')`,
		idFranqueado, servicoCanal, TipoMovDebitoUso, valor, idProc, obs)
	if err != nil {
		return out, err
	}
	return out, tx.Commit()
}

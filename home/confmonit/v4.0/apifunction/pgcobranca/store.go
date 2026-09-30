package pgcobranca

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"apifunction/pgcredito"
)

func db(ctx context.Context) (*sql.DB, error) {
	return pgcredito.DB()
}

func GetConfig(ctx context.Context, idFranqueado string) (Config, error) {
	out := Config{
		IDFranqueado: idFranqueado, Modo: ModoConsolidado,
		DiaMensal: 29, DiaQuinzenal1: 14, DiaQuinzenal2: 29,
	}
	d, err := db(ctx)
	if err != nil {
		return out, err
	}
	err = d.QueryRowContext(ctx, `
SELECT modo, dia_mensal, dia_quinzenal_1, dia_quinzenal_2, id_central, id_representante
FROM fp_cobranca_config WHERE id_franqueado = $1`, idFranqueado).Scan(
		&out.Modo, &out.DiaMensal, &out.DiaQuinzenal1, &out.DiaQuinzenal2, &out.IDCentral, &out.IDRepresentante,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	return out, err
}

func SaveConfig(ctx context.Context, cfg Config) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	if cfg.Modo == "" {
		cfg.Modo = ModoConsolidado
	}
	if cfg.DiaMensal <= 0 {
		cfg.DiaMensal = 29
	}
	if cfg.DiaQuinzenal1 <= 0 {
		cfg.DiaQuinzenal1 = 14
	}
	if cfg.DiaQuinzenal2 <= 0 {
		cfg.DiaQuinzenal2 = 29
	}
	_, err = d.ExecContext(ctx, `
INSERT INTO fp_cobranca_config
(id_franqueado, modo, dia_mensal, dia_quinzenal_1, dia_quinzenal_2, id_central, id_representante, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
ON CONFLICT (id_franqueado) DO UPDATE SET
  modo=EXCLUDED.modo, dia_mensal=EXCLUDED.dia_mensal,
  dia_quinzenal_1=EXCLUDED.dia_quinzenal_1, dia_quinzenal_2=EXCLUDED.dia_quinzenal_2,
  id_central=EXCLUDED.id_central, id_representante=EXCLUDED.id_representante, updated_at=NOW()`,
		cfg.IDFranqueado, cfg.Modo, cfg.DiaMensal, cfg.DiaQuinzenal1, cfg.DiaQuinzenal2,
		cfg.IDCentral, cfg.IDRepresentante,
	)
	return err
}

func UpsertServico(ctx context.Context, s Servico) (int64, error) {
	d, err := db(ctx)
	if err != nil {
		return 0, err
	}
	if s.DiasCiclo <= 0 {
		s.DiasCiclo = 15
	}
	if s.StatusCiclo == "" {
		s.StatusCiclo = StatusCicloEmAjuste
	}
	var id int64
	err = d.QueryRowContext(ctx, `
INSERT INTO fp_servico_cobranca
(id_franqueado, ref_tipo, ref_id, descricao, data_inicio, periodicidade, valor_ciclo, dias_ciclo,
 pago_ate, saldo_ajuste, status_ciclo, valor_piso, margem_central, margem_rep, ativo, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW())
ON CONFLICT (id_franqueado, ref_tipo, ref_id) DO UPDATE SET
  descricao=EXCLUDED.descricao, periodicidade=EXCLUDED.periodicidade, valor_ciclo=EXCLUDED.valor_ciclo,
  dias_ciclo=EXCLUDED.dias_ciclo, pago_ate=COALESCE(EXCLUDED.pago_ate, fp_servico_cobranca.pago_ate),
  valor_piso=EXCLUDED.valor_piso, margem_central=EXCLUDED.margem_central, margem_rep=EXCLUDED.margem_rep,
  ativo=EXCLUDED.ativo, updated_at=NOW()
RETURNING id`,
		s.IDFranqueado, s.RefTipo, s.RefID, s.Descricao, s.DataInicio, s.Periodicidade, s.ValorCiclo, s.DiasCiclo,
		s.PagoAte, s.SaldoAjuste, s.StatusCiclo, s.ValorPiso, s.MargemCentral, s.MargemRep, s.Ativo,
	).Scan(&id)
	return id, err
}

func ListServicos(ctx context.Context, idFranqueado string) ([]Servico, error) {
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.QueryContext(ctx, `
SELECT id, id_franqueado, ref_tipo, ref_id, descricao, data_inicio, periodicidade, valor_ciclo, dias_ciclo,
       pago_ate, saldo_ajuste, status_ciclo, valor_piso, margem_central, margem_rep, ativo
FROM fp_servico_cobranca WHERE id_franqueado = $1 AND ativo = TRUE ORDER BY id`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServicos(rows)
}

func ListFranqueadosConsolidados(ctx context.Context) ([]string, error) {
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.QueryContext(ctx, `
SELECT id_franqueado FROM fp_cobranca_config
WHERE modo IN ('consolidado','alinhar')
UNION
SELECT DISTINCT id_franqueado FROM fp_servico_cobranca WHERE ativo = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func OrdemJaGerada(ctx context.Context, idFranqueado, cicloRef string) (bool, error) {
	d, err := db(ctx)
	if err != nil {
		return false, err
	}
	var n int
	err = d.QueryRowContext(ctx, `
SELECT COUNT(*) FROM fp_ordem_cobranca_log
WHERE id_franqueado = $1 AND ciclo_ref = $2`, idFranqueado, cicloRef).Scan(&n)
	return n > 0, err
}

func LogOrdem(ctx context.Context, idFranqueado, cicloRef string, vencimento time.Time, total, ajustes float64, qtdItens, fpFaturaID int) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	_, err = d.ExecContext(ctx, `
INSERT INTO fp_ordem_cobranca_log
(id_franqueado, ciclo_ref, vencimento, valor_total, valor_ajustes, qtd_itens, fp_fatura_id, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,'gerada')
ON CONFLICT (id_franqueado, ciclo_ref) DO UPDATE SET
  valor_total=EXCLUDED.valor_total, valor_ajustes=EXCLUDED.valor_ajustes,
  qtd_itens=EXCLUDED.qtd_itens, fp_fatura_id=EXCLUDED.fp_fatura_id`,
		idFranqueado, cicloRef, vencimento, total, ajustes, qtdItens, fpFaturaID,
	)
	return err
}

func MarcarServicosNormal(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	d, err := db(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		_, _ = d.ExecContext(ctx, `
UPDATE fp_servico_cobranca SET status_ciclo = 'normal', updated_at = NOW() WHERE id = $1`, id)
	}
	return nil
}

func AtualizarPagoAtePorRef(ctx context.Context, idFranqueado, refTipo, refID string, pagoAte time.Time) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	pagoAte = dateOnly(pagoAte)
	if idFranqueado != "" {
		_, err = d.ExecContext(ctx, `
UPDATE fp_servico_cobranca SET pago_ate = $4, status_ciclo = 'normal', updated_at = NOW()
WHERE id_franqueado = $1 AND ref_tipo = $2 AND ref_id = $3 AND ativo = TRUE`,
			idFranqueado, refTipo, refID, pagoAte)
		return err
	}
	_, err = d.ExecContext(ctx, `
UPDATE fp_servico_cobranca SET pago_ate = $3, status_ciclo = 'normal', updated_at = NOW()
WHERE ref_tipo = $1 AND ref_id = $2 AND ativo = TRUE`,
		refTipo, refID, pagoAte)
	return err
}

func DesativarServicoPorRef(ctx context.Context, idFranqueado, refTipo, refID string) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	_, err = d.ExecContext(ctx, `
UPDATE fp_servico_cobranca SET ativo = FALSE, updated_at = NOW()
WHERE id_franqueado = $1 AND ref_tipo = $2 AND ref_id = $3`,
		idFranqueado, refTipo, refID)
	return err
}

func GetServicoPorRef(ctx context.Context, idFranqueado, refTipo, refID string) (*Servico, error) {
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var s Servico
	var pago sql.NullTime
	err = d.QueryRowContext(ctx, `
SELECT id, id_franqueado, ref_tipo, ref_id, descricao, data_inicio, periodicidade, valor_ciclo, dias_ciclo,
       pago_ate, saldo_ajuste, status_ciclo, valor_piso, margem_central, margem_rep, ativo
FROM fp_servico_cobranca WHERE id_franqueado = $1 AND ref_tipo = $2 AND ref_id = $3 LIMIT 1`,
		idFranqueado, refTipo, refID).Scan(
		&s.ID, &s.IDFranqueado, &s.RefTipo, &s.RefID, &s.Descricao, &s.DataInicio,
		&s.Periodicidade, &s.ValorCiclo, &s.DiasCiclo, &pago, &s.SaldoAjuste, &s.StatusCiclo,
		&s.ValorPiso, &s.MargemCentral, &s.MargemRep, &s.Ativo,
	)
	if err != nil {
		return nil, err
	}
	if pago.Valid {
		t := pago.Time
		s.PagoAte = &t
	}
	return &s, nil
}

func scanServicos(rows *sql.Rows) ([]Servico, error) {
	var list []Servico
	for rows.Next() {
		var s Servico
		var pago sql.NullTime
		if err := rows.Scan(&s.ID, &s.IDFranqueado, &s.RefTipo, &s.RefID, &s.Descricao, &s.DataInicio,
			&s.Periodicidade, &s.ValorCiclo, &s.DiasCiclo, &pago, &s.SaldoAjuste, &s.StatusCiclo,
			&s.ValorPiso, &s.MargemCentral, &s.MargemRep, &s.Ativo); err != nil {
			continue
		}
		if pago.Valid {
			t := pago.Time
			s.PagoAte = &t
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

package financeiro

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	"apifunction/auth"
	"apifunction/pgcredito"
)

func listarFaturasPostgres(ctx context.Context, sess auth.SessaoAdm, f ListarFiltro) ([]Fatura, error) {
	pg, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}

	idCen := idCentralSessao(sess)
	cenKeys, err := chavesCentral(ctx, idCen)
	if err != nil {
		return nil, err
	}
	reps, err := repsDaCentral(ctx, idCen)
	if err != nil {
		return nil, err
	}
	if sess.UserTipo == "REP" {
		reps = []string{strings.TrimSpace(sess.IDRepresentante)}
	} else if f.IDRepresentante != "" {
		reps = []string{strings.TrimSpace(f.IDRepresentante)}
	}
	fraIDs, err := franqueadosDaCentral(ctx, reps)
	if err != nil {
		return nil, err
	}

	repSet := toStringSet(reps)
	fraSet := toStringSet(fraIDs)

	rows, err := pg.QueryContext(ctx, `
		SELECT id, created_at, id_franqueado, id_representante, id_central, referencia,
		       status, tipo, valor_total, vencimento_em, pago_em, ciclo_ref, observacao
		FROM fp_fin_fatura
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Fatura
	for rows.Next() {
		fat, err := scanFaturaPostgres(rows)
		if err != nil {
			return nil, err
		}
		if !faturaNoEscopo(fat, cenKeys, repSet, fraSet) {
			continue
		}
		if st := strings.TrimSpace(f.Status); st != "" && strings.TrimSpace(fat.Status) != st {
			continue
		}
		if tp := strings.TrimSpace(f.Tipo); tp != "" && strings.TrimSpace(fat.Tipo) != tp {
			continue
		}
		if idf := strings.TrimSpace(f.IDFranqueado); idf != "" && strings.TrimSpace(fat.IDFranqueado) != idf {
			continue
		}
		out = append(out, fat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		pi := statusPriority(out[i].Status)
		pj := statusPriority(out[j].Status)
		if pi != pj {
			return pi < pj
		}
		return out[i].ID > out[j].ID
	})

	return out, nil
}

func scanFaturaPostgres(rows *sql.Rows) (Fatura, error) {
	var fat Fatura
	var created, venc, pago sql.NullTime
	if err := rows.Scan(
		&fat.ID, &created,
		&fat.IDFranqueado, &fat.IDRepresentante, &fat.IDCentral, &fat.Referencia,
		&fat.Status, &fat.Tipo, &fat.ValorTotal, &venc, &pago,
		&fat.CicloRef, &fat.Observacao,
	); err != nil {
		return fat, err
	}
	if created.Valid {
		t := created.Time
		fat.CreatedAt = &t
	}
	if venc.Valid {
		t := venc.Time
		fat.VencimentoEm = &t
	}
	if pago.Valid {
		t := pago.Time
		fat.PagoEm = &t
	}
	return fat, nil
}

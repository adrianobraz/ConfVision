package pgatendimento

import (
	"context"
	"strconv"
	"strings"
)

// PosPagamentoItemRef item inferido quando fp_fatura_item no Xano vem sem ref_tipo/ref_id.
type PosPagamentoItemRef struct {
	RefTipo    string
	RefID      string
	ValorPiso  float64
	IDParceiro string
}

// FallbackItensPosPagamento busca exceção/ativação vinculadas à fatura no Postgres ops_*.
func FallbackItensPosPagamento(ctx context.Context, idFranqueado string, faturaID int) ([]PosPagamentoItemRef, error) {
	if faturaID <= 0 {
		return nil, nil
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	idFranqueado = strings.TrimSpace(idFranqueado)
	var out []PosPagamentoItemRef

	rows, err := d.QueryContext(ctx, excecaoSelectCols+`
FROM ops_parceiro_excecao
WHERE fp_fatura_id = $1
ORDER BY id DESC`, faturaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rec, err := scanExcecaoRow(rows)
		if err != nil || rec == nil {
			continue
		}
		if idFranqueado != "" && rec.IDFranqueado != idFranqueado {
			continue
		}
		piso := rec.PrecoUnitario
		if preco, err := lookupPrecoParceiro(ctx, rec.IDFranqueado, rec.IDParceiro); err == nil && preco.PisoParceiro > 0 {
			piso = preco.PisoParceiro
		}
		out = append(out, PosPagamentoItemRef{
			RefTipo:    RefTipoParceiroExcecao,
			RefID:      strconv.FormatInt(rec.ID, 10),
			ValorPiso:  piso,
			IDParceiro: rec.IDParceiro,
		})
	}

	var ativID int64
	var ativFranq, ativParceiro string
	var ativPreco float64
	err = d.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_parceiro, preco_unitario
FROM ops_parceiro_ativacao
WHERE fp_fatura_id = $1
ORDER BY id DESC LIMIT 1`, faturaID).Scan(&ativID, &ativFranq, &ativParceiro, &ativPreco)
	if err == nil {
		if idFranqueado == "" || ativFranq == idFranqueado {
			// não mistura ativação quando a fatura já é de exceção
			if len(out) > 0 {
				return out, nil
			}
			piso := ativPreco
			if preco, err := lookupPrecoParceiro(ctx, ativFranq, ativParceiro); err == nil && preco.PisoParceiro > 0 {
				piso = preco.PisoParceiro
			}
			out = append(out, PosPagamentoItemRef{
				RefTipo:    RefTipoParceiroConfig,
				RefID:      ativParceiro,
				ValorPiso:  piso,
				IDParceiro: ativParceiro,
			})
		}
	}

	return out, nil
}

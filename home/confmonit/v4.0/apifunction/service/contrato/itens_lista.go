package contrato

import (
	"apifunction/db"
	"database/sql"
)

func itensPorContratos(ids []int) (map[int][]ItemCalculado, error) {
	out := make(map[int][]ItemCalculado)
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := make([]byte, 0, len(ids)*2-1)
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args = append(args, id)
	}
	q := `SELECT ID_Contrato, Tipo, Chave, Descricao, Quantidade, QtdPorPacote, ValorUnitario, ValorTotal, RefID
		FROM fp_contrato_item WHERE ID_Contrato IN (` + string(placeholders) + `) ORDER BY ID_Contrato, ID_ContratoItem`
	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var idContrato int
		var it ItemCalculado
		var qtdPac sql.NullInt64
		var refID sql.NullInt64
		if err := rows.Scan(&idContrato, &it.Tipo, &it.Chave, &it.Descricao, &it.Quantidade, &qtdPac, &it.ValorUnitario, &it.ValorTotal, &refID); err != nil {
			return nil, err
		}
		if qtdPac.Valid {
			it.QtdPorPacote = int(qtdPac.Int64)
		}
		if refID.Valid {
			it.RefID = int(refID.Int64)
		}
		out[idContrato] = append(out[idContrato], it)
	}
	return out, rows.Err()
}

func anexarItensContratos(lista []ContratoResumo) error {
	ids := make([]int, 0, len(lista))
	for _, c := range lista {
		ids = append(ids, c.IDContrato)
	}
	m, err := itensPorContratos(ids)
	if err != nil {
		return err
	}
	for i := range lista {
		lista[i].Itens = m[lista[i].IDContrato]
	}
	return nil
}

func anexarItensFaturas(lista []FaturaResumo) error {
	ids := make([]int, 0, len(lista))
	seen := make(map[int]bool)
	for _, f := range lista {
		if f.IDContrato > 0 && !seen[f.IDContrato] {
			seen[f.IDContrato] = true
			ids = append(ids, f.IDContrato)
		}
	}
	m, err := itensPorContratos(ids)
	if err != nil {
		return err
	}
	for i := range lista {
		lista[i].Itens = m[lista[i].IDContrato]
	}
	return nil
}

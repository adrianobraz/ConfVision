package pgcatalogo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"apifunction/auth"
	"apifunction/pgcredito"
)

func open() (*sql.DB, error) {
	return pgcredito.DB()
}

type CentralPrecoCfg struct {
	ModoPreco               string
	ValorUnitarioMinimoCota float64
}

func ConfigCentralPreco(ctx context.Context, idCentral string) (CentralPrecoCfg, error) {
	cfg := CentralPrecoCfg{ModoPreco: "livre"}
	db, err := open()
	if err != nil {
		return cfg, err
	}
	var modo sql.NullString
	var unitMin sql.NullFloat64
	err = db.QueryRowContext(ctx, `
		SELECT modo_preco, COALESCE(valor_unitario_minimo_cota, 0)
		FROM fp_central_preco_config WHERE id_central = $1
	`, idCentral).Scan(&modo, &unitMin)
	if err == sql.ErrNoRows {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	m := strings.ToLower(strings.TrimSpace(modo.String))
	if m == "piso" {
		cfg.ModoPreco = "piso"
	}
	cfg.ValorUnitarioMinimoCota = unitMin.Float64
	return cfg, nil
}

func ModoPreco(ctx context.Context, idCentral string) (string, error) {
	cfg, err := ConfigCentralPreco(ctx, idCentral)
	if err != nil {
		return "livre", err
	}
	return cfg.ModoPreco, nil
}

func SalvarModoPreco(ctx context.Context, idCentral, modo, observacao string, breakglass bool) error {
	if !breakglass {
		return errors.New("somente Break-glass pode alterar modo de preco")
	}
	modo = strings.ToLower(strings.TrimSpace(modo))
	if modo != "livre" && modo != "piso" {
		return errors.New("modo_preco deve ser livre ou piso")
	}
	db, err := open()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO fp_central_preco_config (id_central, modo_preco, observacao, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (id_central) DO UPDATE SET
			modo_preco = EXCLUDED.modo_preco,
			observacao = CASE WHEN EXCLUDED.observacao = '' THEN fp_central_preco_config.observacao ELSE EXCLUDED.observacao END,
			updated_at = NOW()
	`, idCentral, modo, observacao)
	if err != nil {
		return err
	}
	if modo == "piso" {
		return syncPisoBreakglass(ctx, db, idCentral)
	}
	return nil
}

func SalvarUnitarioMinimoCota(ctx context.Context, idCentral string, valor float64, breakglass bool) error {
	if !breakglass {
		return errors.New("somente Break-glass pode alterar valor unitario minimo de cota")
	}
	if valor < 0 {
		return errors.New("valor unitario minimo invalido")
	}
	db, err := open()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO fp_central_preco_config (id_central, modo_preco, valor_unitario_minimo_cota, updated_at)
		VALUES ($1, 'livre', $2, NOW())
		ON CONFLICT (id_central) DO UPDATE SET
			valor_unitario_minimo_cota = EXCLUDED.valor_unitario_minimo_cota,
			updated_at = NOW()
	`, idCentral, valor)
	return err
}

func SyncPisoBreakglass(ctx context.Context, idCentral string) error {
	db, err := open()
	if err != nil {
		return err
	}
	return syncPisoBreakglass(ctx, db, idCentral)
}

func syncPisoBreakglass(ctx context.Context, db *sql.DB, idCentral string) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE fp_catalogo_produto
		SET valor_piso_breakglass = valor_mensal
		WHERE id_central = $1 AND id_representante = ''
	`, idCentral); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE fp_pacote_cota
		SET valor_piso_breakglass = valor
		WHERE id_central = $1
	`, idCentral); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		UPDATE vis_capacidade_config
		SET preco_piso_breakglass = preco_base_camera
		WHERE id_central = $1 AND id_representante = ''
	`, idCentral)
	return err
}

func ListarProdutos(ctx context.Context, idCentral, produto, ativo string) ([]Produto, string, error) {
	db, err := open()
	if err != nil {
		return nil, "", err
	}
	modo, _ := ModoPreco(ctx, idCentral)

	q := `
		SELECT p.id, p.produto, p.plano, p.nome_exibicao, p.valor_mensal, p.valor_piso_breakglass,
		       COALESCE(p.periodicidade_padrao,'mensal'), p.retencao_dias,
		       p.limites_json, p.modulos_json, p.fp_pacote_cota_id, p.ativo, COALESCE(p.observacao,'')
		FROM fp_catalogo_produto p
		WHERE p.id_central = $1 AND p.id_representante = ''
	`
	args := []any{idCentral}
	n := 2
	if produto = strings.TrimSpace(produto); produto != "" {
		q += fmt.Sprintf(" AND p.produto = $%d", n)
		args = append(args, produto)
		n++
	}
	if ativo = strings.TrimSpace(ativo); ativo != "" {
		q += fmt.Sprintf(" AND p.ativo = $%d", n)
		args = append(args, ativo)
	}
	q += " ORDER BY p.produto ASC, p.valor_mensal ASC"

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, modo, err
	}
	defer rows.Close()

	var out []Produto
	for rows.Next() {
		var item Produto
		var lim, mod []byte
		var cotaID sql.NullInt64
		if err := rows.Scan(
			&item.ID, &item.Produto, &item.Plano, &item.NomeExibicao,
			&item.ValorMensal, &item.ValorPisoBreakglass,
			&item.PeriodicidadePadrao, &item.RetencaoDias,
			&lim, &mod, &cotaID, &item.Ativo, &item.Observacao,
		); err != nil {
			return nil, modo, err
		}
		item.ValorPisoMinimo = item.ValorPisoBreakglass
		item.ModoPreco = modo
		if len(lim) > 0 {
			item.LimitesJSON = json.RawMessage(lim)
		}
		if len(mod) > 0 {
			item.ModulosJSON = json.RawMessage(mod)
		}
		if cotaID.Valid {
			v := int(cotaID.Int64)
			item.FPPacoteCotaID = &v
			enrichCota(ctx, db, &item)
		}
		out = append(out, item)
	}
	return out, modo, rows.Err()
}

func enrichCota(ctx context.Context, db *sql.DB, item *Produto) {
	if item.FPPacoteCotaID == nil {
		return
	}
	var nome sql.NullString
	var qtd sql.NullInt64
	var val sql.NullFloat64
	err := db.QueryRowContext(ctx, `
		SELECT nome, quantidade, valor FROM fp_pacote_cota WHERE id = $1
	`, *item.FPPacoteCotaID).Scan(&nome, &qtd, &val)
	if err != nil {
		return
	}
	item.PacoteCotaNome = nome.String
	item.PacoteCotaQuantidade = int(qtd.Int64)
	item.ValorCota = val.Float64
	item.ValorTotalMensal = item.ValorMensal + item.ValorCota
}

func SalvarProduto(ctx context.Context, sess auth.SessaoAdm, in SalvarProdutoInput) (Produto, error) {
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		return Produto{}, err
	}
	if strings.TrimSpace(in.Produto) == "" || strings.TrimSpace(in.Plano) == "" {
		return Produto{}, errors.New("produto e plano obrigatorios")
	}
	bg := isBreakglass(sess)
	modo, _ := ModoPreco(ctx, idCentral)

	db, err := open()
	if err != nil {
		return Produto{}, err
	}

	if in.Produto == "franqueadopro" && (in.FPPacoteCotaID == nil || *in.FPPacoteCotaID <= 0) {
		return Produto{}, errors.New("vincule um Pacote de Cotas a este plano FranqueadoPro")
	}

	pisoBG := 0.0
	if in.ID > 0 {
		var atual sql.NullFloat64
		var idCen string
		err = db.QueryRowContext(ctx, `
			SELECT valor_piso_breakglass, id_central FROM fp_catalogo_produto WHERE id = $1
		`, in.ID).Scan(&atual, &idCen)
		if err == sql.ErrNoRows {
			return Produto{}, errors.New("item do catalogo nao encontrado")
		}
		if err != nil {
			return Produto{}, err
		}
		if strings.TrimSpace(idCen) != idCentral {
			return Produto{}, errors.New("item do catalogo fora do escopo da Central")
		}
		pisoBG = atual.Float64
		if !bg && modo == "piso" && in.ValorMensal < pisoBG {
			return Produto{}, fmt.Errorf("Central em modo piso: valor minimo e R$ %.2f", pisoBG)
		}
		if bg {
			pisoBG = in.ValorMensal
		}
	} else {
		if bg || modo == "piso" {
			pisoBG = in.ValorMensal
		}
	}

	lim := coalesceJSON(in.LimitesJSON)
	mod := coalesceJSON(in.ModulosJSON)
	ativo := normalizeAtivo(in.Ativo)
	per := strings.TrimSpace(in.Periodicidade)
	if per == "" {
		per = "mensal"
	}
	retencao := in.RetencaoDias
	if retencao <= 0 {
		retencao = 30
	}

	var id int
	if in.ID > 0 {
		err = db.QueryRowContext(ctx, `
			UPDATE fp_catalogo_produto SET
				produto = $2, plano = $3, nome_exibicao = $4, valor_mensal = $5,
				valor_piso_breakglass = $6, periodicidade_padrao = $7, retencao_dias = $8,
				limites_json = $9, modulos_json = $10, fp_pacote_cota_id = $11,
				ativo = $12, observacao = $13
			WHERE id = $1 AND id_central = $14
			RETURNING id
		`, in.ID, in.Produto, in.Plano, in.NomeExibicao, in.ValorMensal, pisoBG, per, retencao,
			lim, mod, in.FPPacoteCotaID, ativo, in.Observacao, idCentral).Scan(&id)
	} else {
		err = db.QueryRowContext(ctx, `
			INSERT INTO fp_catalogo_produto
			(id_central, produto, plano, nome_exibicao, valor_mensal, valor_piso_breakglass,
			 periodicidade_padrao, retencao_dias, limites_json, modulos_json, fp_pacote_cota_id, ativo, observacao)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			RETURNING id
		`, idCentral, in.Produto, in.Plano, in.NomeExibicao, in.ValorMensal, pisoBG, per, retencao,
			lim, mod, in.FPPacoteCotaID, ativo, in.Observacao).Scan(&id)
	}
	if err != nil {
		return Produto{}, err
	}
	lista, _, err := ListarProdutos(ctx, idCentral, in.Produto, "")
	if err != nil {
		return Produto{}, err
	}
	for _, p := range lista {
		if p.ID == id {
			return p, nil
		}
	}
	return Produto{ID: id}, nil
}

func ListarPacotes(ctx context.Context, idCentral, idRep, ativo string) ([]PacoteCota, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	q := `SELECT id, nome, quantidade, valor, valor_piso_breakglass, ativo, COALESCE(observacao,'')
	      FROM fp_pacote_cota WHERE id_central = $1`
	args := []any{idCentral}
	if ativo = strings.TrimSpace(ativo); ativo != "" {
		q += " AND ativo = $2"
		args = append(args, ativo)
	}
	q += " ORDER BY quantidade ASC"
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PacoteCota
	for rows.Next() {
		var p PacoteCota
		var pisoBG float64
		if err := rows.Scan(&p.ID, &p.Nome, &p.Quantidade, &p.Valor, &pisoBG, &p.Ativo, &p.Observacao); err != nil {
			return nil, err
		}
		p.ValorPiso = pisoBG
		if p.ValorPiso <= 0 {
			p.ValorPiso = p.Valor
		}
		p.ValorVenda = p.ValorPiso
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	idRep = strings.TrimSpace(idRep)
	if idRep == "" || len(out) == 0 {
		return out, nil
	}
	for i := range out {
		var venda sql.NullFloat64
		var precoID sql.NullInt64
		err = db.QueryRowContext(ctx, `
			SELECT id, valor_venda FROM fp_preco_pacote_cota
			WHERE id_representante = $1 AND fp_pacote_cota_id = $2 AND ativo = 'S'
		`, idRep, out[i].ID).Scan(&precoID, &venda)
		if err == nil && venda.Valid && venda.Float64 >= out[i].ValorPiso {
			out[i].ValorVenda = venda.Float64
			if precoID.Valid {
				v := int(precoID.Int64)
				out[i].PrecoRepID = &v
			}
		}
	}
	return out, nil
}

func SalvarPacote(ctx context.Context, sess auth.SessaoAdm, in SalvarPacoteInput) (PacoteCota, error) {
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		return PacoteCota{}, err
	}
	if strings.TrimSpace(in.Nome) == "" {
		return PacoteCota{}, errors.New("nome obrigatorio")
	}
	if in.Quantidade <= 0 {
		return PacoteCota{}, errors.New("quantidade deve ser maior que zero")
	}
	bg := isBreakglass(sess)
	cfg, _ := ConfigCentralPreco(ctx, idCentral)
	modo := cfg.ModoPreco

	db, err := open()
	if err != nil {
		return PacoteCota{}, err
	}

	if in.ID <= 0 && !bg && modo == "piso" && cfg.ValorUnitarioMinimoCota > 0 {
		minTotal := float64(in.Quantidade) * cfg.ValorUnitarioMinimoCota
		if in.Valor < minTotal {
			return PacoteCota{}, fmt.Errorf(
				"Central em modo piso: valor minimo do pacote e R$ %.2f (%d x R$ %.2f)",
				minTotal, in.Quantidade, cfg.ValorUnitarioMinimoCota,
			)
		}
	}

	pisoBG := 0.0
	if in.ID > 0 {
		var atual sql.NullFloat64
		var idCen string
		err = db.QueryRowContext(ctx, `
			SELECT valor_piso_breakglass, id_central FROM fp_pacote_cota WHERE id = $1
		`, in.ID).Scan(&atual, &idCen)
		if err == sql.ErrNoRows {
			return PacoteCota{}, errors.New("pacote de cota nao encontrado")
		}
		if err != nil {
			return PacoteCota{}, err
		}
		if strings.TrimSpace(idCen) != idCentral {
			return PacoteCota{}, errors.New("pacote fora do escopo da Central")
		}
		pisoBG = atual.Float64
		if pisoBG <= 0 {
			pisoBG = in.Valor
		}
		if !bg && modo == "piso" && in.Valor < pisoBG {
			return PacoteCota{}, fmt.Errorf("Central em modo piso: valor minimo e R$ %.2f", pisoBG)
		}
		if bg {
			pisoBG = in.Valor
		}
	} else if bg || modo == "piso" {
		pisoBG = in.Valor
	}

	ativo := normalizeAtivo(in.Ativo)
	var id int
	if in.ID > 0 {
		err = db.QueryRowContext(ctx, `
			UPDATE fp_pacote_cota SET nome=$2, quantidade=$3, valor=$4, valor_piso_breakglass=$5,
				ativo=$6, observacao=$7
			WHERE id=$1 AND id_central=$8 RETURNING id
		`, in.ID, in.Nome, in.Quantidade, in.Valor, pisoBG, ativo, in.Observacao, idCentral).Scan(&id)
	} else {
		err = db.QueryRowContext(ctx, `
			INSERT INTO fp_pacote_cota (id_central, nome, quantidade, valor, valor_piso_breakglass, ativo, observacao)
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id
		`, idCentral, in.Nome, in.Quantidade, in.Valor, pisoBG, ativo, in.Observacao).Scan(&id)
	}
	if err != nil {
		return PacoteCota{}, err
	}
	lista, err := ListarPacotes(ctx, idCentral, "", "")
	if err != nil {
		return PacoteCota{}, err
	}
	for _, p := range lista {
		if p.ID == id {
			return p, nil
		}
	}
	return PacoteCota{ID: id, Nome: in.Nome, Quantidade: in.Quantidade, Valor: in.Valor, Ativo: ativo}, nil
}

func SalvarPrecoPacoteRep(ctx context.Context, sess auth.SessaoAdm, in SalvarPrecoPacoteRepInput) (PacoteCota, error) {
	if sess.UserTipo != "REP" {
		return PacoteCota{}, errors.New("somente Representante define preco de venda de cota")
	}
	idRep := strings.TrimSpace(in.IDRepresentante)
	if idRep == "" {
		idRep = strings.TrimSpace(sess.IDRepresentante)
	}
	if idRep == "" {
		return PacoteCota{}, errors.New("id_representante obrigatorio")
	}
	if in.FPPacoteCotaID <= 0 {
		return PacoteCota{}, errors.New("fp_pacote_cota_id obrigatorio")
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		return PacoteCota{}, err
	}
	db, err := open()
	if err != nil {
		return PacoteCota{}, err
	}
	var nome, ativo string
	var qtd int
	var valorPacote float64
	var idCenPacote string
	err = db.QueryRowContext(ctx, `
		SELECT nome, quantidade, valor, ativo, id_central FROM fp_pacote_cota WHERE id = $1
	`, in.FPPacoteCotaID).Scan(&nome, &qtd, &valorPacote, &ativo, &idCenPacote)
	if err == sql.ErrNoRows {
		return PacoteCota{}, errors.New("pacote de cota nao encontrado")
	}
	if err != nil {
		return PacoteCota{}, err
	}
	if ativo != "S" {
		return PacoteCota{}, errors.New("pacote de cota inativo")
	}
	if strings.TrimSpace(idCenPacote) != idCentral {
		return PacoteCota{}, errors.New("pacote fora do escopo da Central")
	}
	_ = qtd
	piso := valorPacote
	if in.ValorVenda < piso {
		return PacoteCota{}, fmt.Errorf("valor_venda nao pode ser menor que o piso da Central (R$ %.2f)", piso)
	}
	var id int
	err = db.QueryRowContext(ctx, `
		INSERT INTO fp_preco_pacote_cota (id_representante, fp_pacote_cota_id, valor_venda, ativo, observacao, atualizado_em)
		VALUES ($1,$2,$3,'S',$4,NOW())
		ON CONFLICT (id_representante, fp_pacote_cota_id) DO UPDATE SET
			valor_venda = EXCLUDED.valor_venda,
			ativo = 'S',
			observacao = EXCLUDED.observacao,
			atualizado_em = NOW()
		RETURNING id
	`, idRep, in.FPPacoteCotaID, in.ValorVenda, in.Observacao).Scan(&id)
	if err != nil {
		return PacoteCota{}, err
	}
	_ = id
	lista, err := ListarPacotes(ctx, idCentral, idRep, "")
	if err != nil {
		return PacoteCota{}, err
	}
	for _, p := range lista {
		if p.ID == in.FPPacoteCotaID {
			return p, nil
		}
	}
	return PacoteCota{}, errors.New("pacote nao encontrado apos salvar preco")
}

func ListarPrecoConfig(ctx context.Context) ([]PrecoConfig, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id_central, modo_preco, COALESCE(valor_unitario_minimo_cota, 0), COALESCE(observacao,'')
		FROM fp_central_preco_config ORDER BY id_central
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PrecoConfig
	for rows.Next() {
		var c PrecoConfig
		if err := rows.Scan(&c.IDCentral, &c.ModoPreco, &c.ValorUnitarioMinimoCota, &c.Observacao); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func isBreakglass(sess auth.SessaoAdm) bool {
	return strings.EqualFold(strings.TrimSpace(sess.IDUsuario), "BREAKGLASS")
}

func coalesceJSON(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

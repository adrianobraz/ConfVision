package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

const planoCapacidadeProcessamento = "capacidade_processamento"

type CapacidadeConfig struct {
	IDCentral         string  `json:"id_central"`
	IDRepresentante   string  `json:"id_representante"`
	QuantidadeMinima  int     `json:"quantidade_minima"`
	PrecoBaseCamera   float64 `json:"preco_base_camera"`
	PrecoVendaCamera  float64 `json:"preco_venda_camera"`
	PrecoEfetivoCamera float64 `json:"preco_efetivo_camera"`
}

type CapacidadeContrato struct {
	ID                       int        `json:"id"`
	IDFranqueado             string     `json:"id_franqueado"`
	IDCentral                string     `json:"id_central"`
	IDRepresentante          string     `json:"id_representante"`
	QuantidadeContratada     int        `json:"quantidade_contratada"`
	QuantidadeMinimaSnapshot int        `json:"quantidade_minima_snapshot"`
	PrecoPorCamera           float64    `json:"preco_por_camera"`
	ValorMensal              float64    `json:"valor_mensal"`
	Status                   string     `json:"status"`
	PagoEm                   *time.Time `json:"pago_em,omitempty"`
	ValidoAte                *time.Time `json:"valido_ate,omitempty"`
	IDFatura                 string     `json:"id_fatura,omitempty"`
	Observacao               string     `json:"observacao,omitempty"`
	DescFatura               string     `json:"desc_fatura,omitempty"`
}

func CountCamerasEmUso(ctx context.Context, idFranqueado string) (int, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return 0, fmt.Errorf("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return 0, err
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_camera
WHERE id_franqueado = $1 AND vis_licenca_id IS NOT NULL`, idFranqueado).Scan(&n)
	return n, err
}

func GetCapacidadeConfig(ctx context.Context, idCentral, idRepresentante string) (CapacidadeConfig, error) {
	idCentral = normalizeCentralID(idCentral)
	idRepresentante = strings.TrimSpace(idRepresentante)

	db, err := DB()
	if err != nil {
		return CapacidadeConfig{}, err
	}

	cfg, err := loadCapacidadeConfigRow(ctx, db, idCentral, idRepresentante)
	if err == nil {
		return cfg, nil
	}
	if !strings.EqualFold(idRepresentante, "") {
		if cfg2, err2 := loadCapacidadeConfigRow(ctx, db, idCentral, ""); err2 == nil {
			return cfg2, nil
		}
	}
	if idCentral != "*" {
		if cfg3, err3 := loadCapacidadeConfigRow(ctx, db, "*", ""); err3 == nil {
			cfg3.IDCentral = idCentral
			cfg3.IDRepresentante = idRepresentante
			return cfg3, nil
		}
	}
	return CapacidadeConfig{}, fmt.Errorf("configuracao de capacidade nao encontrada")
}

func loadCapacidadeConfigRow(ctx context.Context, db *sql.DB, idCentral, idRep string) (CapacidadeConfig, error) {
	var cfg CapacidadeConfig
	var precoVenda sql.NullFloat64
	err := db.QueryRowContext(ctx, `
SELECT id_central, id_representante, quantidade_minima, preco_base_camera, preco_venda_camera
FROM vis_capacidade_config
WHERE id_central = $1 AND id_representante = $2`,
		idCentral, idRep,
	).Scan(&cfg.IDCentral, &cfg.IDRepresentante, &cfg.QuantidadeMinima, &cfg.PrecoBaseCamera, &precoVenda)
	if err != nil {
		return cfg, err
	}
	if precoVenda.Valid && precoVenda.Float64 > 0 {
		cfg.PrecoVendaCamera = precoVenda.Float64
		cfg.PrecoEfetivoCamera = precoVenda.Float64
	} else {
		cfg.PrecoEfetivoCamera = cfg.PrecoBaseCamera
	}
	return cfg, nil
}

func normalizeCentralID(idCentral string) string {
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		return "*"
	}
	return idCentral
}

func contratoAtivo(ctx context.Context, idFranqueado string) (*CapacidadeContrato, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var c CapacidadeContrato
	var pago, valido sql.NullTime
	var idFat, obs sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_central, COALESCE(id_representante,''), quantidade_contratada,
       quantidade_minima_snapshot, preco_por_camera, valor_mensal, status,
       pago_em, valido_ate, id_fatura, COALESCE(observacao,'')
FROM vis_capacidade_contrato
WHERE id_franqueado = $1 AND status = 'ativo'
  AND (valido_ate IS NULL OR valido_ate > NOW())
ORDER BY id DESC LIMIT 1`, idFranqueado).Scan(
		&c.ID, &c.IDFranqueado, &c.IDCentral, &c.IDRepresentante,
		&c.QuantidadeContratada, &c.QuantidadeMinimaSnapshot, &c.PrecoPorCamera, &c.ValorMensal,
		&c.Status, &pago, &valido, &idFat, &obs,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if pago.Valid {
		c.PagoEm = &pago.Time
	}
	if valido.Valid {
		c.ValidoAte = &valido.Time
	}
	if idFat.Valid {
		c.IDFatura = idFat.String
	}
	if obs.Valid {
		c.Observacao = obs.String
	}
	return &c, nil
}

func contratoPendente(ctx context.Context, idFranqueado string) (*CapacidadeContrato, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var c CapacidadeContrato
	var idFat, obs sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_central, COALESCE(id_representante,''), quantidade_contratada,
       quantidade_minima_snapshot, preco_por_camera, valor_mensal, status,
       id_fatura, COALESCE(observacao,'')
FROM vis_capacidade_contrato
WHERE id_franqueado = $1 AND status = 'pendente'
ORDER BY id DESC LIMIT 1`, idFranqueado).Scan(
		&c.ID, &c.IDFranqueado, &c.IDCentral, &c.IDRepresentante,
		&c.QuantidadeContratada, &c.QuantidadeMinimaSnapshot, &c.PrecoPorCamera, &c.ValorMensal,
		&c.Status, &idFat, &obs,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if idFat.Valid {
		c.IDFatura = idFat.String
	}
	if obs.Valid {
		c.Observacao = obs.String
	}
	return &c, nil
}

func CotacaoCapacidade(ctx context.Context, idFranqueado string, quantidade int, idCentral, idRepresentante string) (map[string]any, error) {
	if quantidade <= 0 {
		return nil, fmt.Errorf("quantidade deve ser maior que zero")
	}
	cfg, err := GetCapacidadeConfig(ctx, idCentral, idRepresentante)
	if err != nil {
		return nil, err
	}
	emUso, err := CountCamerasEmUso(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	minQtd := cfg.QuantidadeMinima
	if emUso > minQtd {
		minQtd = emUso
	}
	if quantidade < minQtd {
		return nil, fmt.Errorf("Quantidade minima: %d camera(s) (em uso: %d)", minQtd, emUso)
	}
	preco := cfg.PrecoEfetivoCamera
	valor := roundMoney(preco * float64(quantidade))
	return map[string]any{
		"id_franqueado":       idFranqueado,
		"quantidade":          quantidade,
		"quantidade_minima":   cfg.QuantidadeMinima,
		"quantidade_minima_efetiva": minQtd,
		"em_uso":              emUso,
		"preco_por_camera":    preco,
		"valor_mensal":        valor,
		"plano":               planoCapacidadeProcessamento,
		"id_central":          normalizeCentralID(idCentral),
		"id_representante":    strings.TrimSpace(idRepresentante),
	}, nil
}

func ResumoCapacidade(ctx context.Context, idFranqueado, idCentral, idRepresentante string) (map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}

	cfg, err := GetCapacidadeConfig(ctx, idCentral, idRepresentante)
	if err != nil {
		return nil, err
	}
	emUso, err := CountCamerasEmUso(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}

	ativo, _ := contratoAtivo(ctx, idFranqueado)
	pendente, _ := contratoPendente(ctx, idFranqueado)

	contratada := 0
	disponivel := 0
	temAtivo := ativo != nil
	if temAtivo {
		contratada = ativo.QuantidadeContratada
		disponivel = contratada - emUso
		if disponivel < 0 {
			disponivel = 0
		}
	}

	minEfetiva := cfg.QuantidadeMinima
	if emUso > minEfetiva {
		minEfetiva = emUso
	}

	out := map[string]any{
		"config": map[string]any{
			"quantidade_minima":  cfg.QuantidadeMinima,
			"preco_por_camera":   cfg.PrecoEfetivoCamera,
			"preco_base_camera":  cfg.PrecoBaseCamera,
			"id_central":         cfg.IDCentral,
		},
		"em_uso":                  emUso,
		"contratada":              contratada,
		"disponivel":              disponivel,
		"quantidade_minima_efetiva": minEfetiva,
		"tem_contrato_ativo":      temAtivo,
		"tem_pendente":            pendente != nil,
		"pode_comprar_licenca":    temAtivo,
		"pode_cadastrar_camera":   temAtivo && disponivel > 0,
	}
	if ativo != nil {
		out["contrato_ativo"] = contratoToMap(*ativo)
	}
	if pendente != nil {
		out["contrato_pendente"] = contratoToMap(*pendente)
	}
	return out, nil
}

func contratoToMap(c CapacidadeContrato) map[string]any {
	m := map[string]any{
		"id":                          c.ID,
		"id_franqueado":               c.IDFranqueado,
		"id_central":                  c.IDCentral,
		"id_representante":            c.IDRepresentante,
		"quantidade_contratada":       c.QuantidadeContratada,
		"quantidade_minima_snapshot":  c.QuantidadeMinimaSnapshot,
		"preco_por_camera":            c.PrecoPorCamera,
		"valor_mensal":                c.ValorMensal,
		"status":                      c.Status,
		"observacao":                  c.Observacao,
	}
	if c.PagoEm != nil {
		m["pago_em"] = c.PagoEm.UTC().Format(time.RFC3339)
	}
	if c.ValidoAte != nil {
		m["valido_ate"] = c.ValidoAte.UTC().Format(time.RFC3339)
	}
	if c.IDFatura != "" {
		m["id_fatura"] = c.IDFatura
	}
	return m
}

// ReservarCapacidadePendente cria contrato pendente (chamado pelo Xano na venda).
func ReservarCapacidadePendente(ctx context.Context, input map[string]any) (map[string]any, error) {
	idFranqueado := strVal(input, "id_franqueado")
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	quantidade := intVal(input, "quantidade")
	if quantidade <= 0 {
		quantidade = intVal(input, "quantidade_contratada")
	}
	if quantidade <= 0 {
		return nil, fmt.Errorf("quantidade obrigatoria")
	}
	idCentral := strVal(input, "id_central")
	idRep := strVal(input, "id_representante")

	cot, err := CotacaoCapacidade(ctx, idFranqueado, quantidade, idCentral, idRep)
	if err != nil {
		return nil, err
	}

	pend, _ := contratoPendente(ctx, idFranqueado)
	if pend != nil {
		return nil, fmt.Errorf("Ja existe contrato de capacidade pendente (#%d). Pague ou cancele antes de contratar novamente", pend.ID)
	}

	preco := floatVal(cot, "preco_por_camera")
	valor := floatVal(cot, "valor_mensal")
	minSnap := int(floatVal(cot, "quantidade_minima"))
	idCentral = normalizeCentralID(strVal(cot, "id_central"))

	db, err := DB()
	if err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("ConfVision — Capacidade de processamento (%d cameras)", quantidade)
	obs := fmt.Sprintf("Aguardando pagamento — %d vagas de camera", quantidade)

	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_capacidade_contrato (
    id_franqueado, id_central, id_representante, quantidade_contratada,
    quantidade_minima_snapshot, preco_por_camera, valor_mensal, status, observacao
) VALUES ($1,$2,$3,$4,$5,$6,$7,'pendente',$8)
RETURNING id`,
		idFranqueado, idCentral, idRep, quantidade, minSnap, preco, valor, obs,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                    id,
		"id_franqueado":         idFranqueado,
		"quantidade_contratada": quantidade,
		"preco_por_camera":      preco,
		"valor":                 valor,
		"valor_mensal":          valor,
		"status":                "pendente",
		"desc_fatura":           desc,
		"plano":                 planoCapacidadeProcessamento,
	}, nil
}

// AtivarCapacidadePagamento ativa contrato apos pagamento da fatura.
func AtivarCapacidadePagamento(ctx context.Context, contratoID, faturaID, pagamentoID int, pagoEm time.Time) (map[string]any, error) {
	if contratoID <= 0 {
		return nil, fmt.Errorf("vis_capacidade_contrato_id obrigatorio")
	}
	if pagoEm.IsZero() {
		pagoEm = time.Now().UTC()
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var idFra string
	var qtd int
	var status string
	err = tx.QueryRowContext(ctx, `
SELECT id_franqueado, quantidade_contratada, status
FROM vis_capacidade_contrato WHERE id = $1 FOR UPDATE`, contratoID).Scan(&idFra, &qtd, &status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contrato de capacidade nao encontrado")
	}
	if err != nil {
		return nil, err
	}

	// Substitui contrato ativo anterior (upgrade/downgrade apos pagamento)
	_, _ = tx.ExecContext(ctx, `
UPDATE vis_capacidade_contrato SET status = 'substituido', updated_at = NOW()
WHERE id_franqueado = $1 AND status = 'ativo' AND id <> $2`, idFra, contratoID)

	validoAte := pagoEm.Add(30 * 24 * time.Hour)
	idFat := ""
	if faturaID > 0 {
		idFat = fmt.Sprintf("%d", faturaID)
	}
	idPag := ""
	if pagamentoID > 0 {
		idPag = fmt.Sprintf("%d", pagamentoID)
	}

	novoStatus := "ativo"
	if status == "ativo" {
		novoStatus = "ativo"
	} else if status != "pendente" && status != "ativo" {
		return nil, fmt.Errorf("contrato em status invalido: %s", status)
	}

	_, err = tx.ExecContext(ctx, `
UPDATE vis_capacidade_contrato SET
    status = $2, pago_em = $3, valido_ate = $4,
    id_fatura = COALESCE(NULLIF($5,''), id_fatura),
    id_pagamento = COALESCE(NULLIF($6,''), id_pagamento),
    updated_at = NOW()
WHERE id = $1`, contratoID, novoStatus, pagoEm, validoAte, idFat, idPag)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                contratoID,
		"id_franqueado":     idFra,
		"status":            novoStatus,
		"quantidade_contratada": qtd,
		"pago_em":           pagoEm.UTC().Format(time.RFC3339),
		"valido_ate":        validoAte.UTC().Format(time.RFC3339),
		"id_fatura":         idFat,
	}, nil
}

// EstornarCapacidadeContrato cancela contrato pendente ou expira ativo.
func EstornarCapacidadeContrato(ctx context.Context, contratoID int, observacao string) (map[string]any, error) {
	if contratoID <= 0 {
		return nil, fmt.Errorf("vis_capacidade_contrato_id obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var obsArg any
	if strings.TrimSpace(observacao) != "" {
		obsArg = observacao
	}
	var idFra, status string
	err = db.QueryRowContext(ctx, `
UPDATE vis_capacidade_contrato SET
    status = CASE WHEN status = 'pendente' THEN 'cancelado' ELSE 'expirado' END,
    valido_ate = $2, updated_at = NOW(),
    observacao = COALESCE($3, observacao)
WHERE id = $1
RETURNING id_franqueado, status`, contratoID, now, obsArg).Scan(&idFra, &status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contrato nao encontrado")
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":            contratoID,
		"id_franqueado": idFra,
		"status":        status,
	}, nil
}

// CheckCapacidadeDisponivel valida se franqueado pode comprar licenca ou cadastrar camera.
func CheckCapacidadeDisponivel(ctx context.Context, idFranqueado string, paraCadastroCamera bool) error {
	ativo, err := contratoAtivo(ctx, idFranqueado)
	if err != nil {
		return err
	}
	if ativo == nil {
		return fmt.Errorf("Contrate a capacidade de processamento antes de continuar. Acesse Minhas licencas > Processamento.")
	}
	if !paraCadastroCamera {
		return nil
	}
	emUso, err := CountCamerasEmUso(ctx, idFranqueado)
	if err != nil {
		return err
	}
	if emUso >= ativo.QuantidadeContratada {
		return fmt.Errorf("Capacidade esgotada (%d/%d cameras). Aumente a capacidade de processamento para cadastrar mais cameras.", emUso, ativo.QuantidadeContratada)
	}
	return nil
}

func TemCapacidadeAtiva(ctx context.Context, idFranqueado string) (bool, error) {
	ativo, err := contratoAtivo(ctx, idFranqueado)
	if err != nil {
		return false, err
	}
	return ativo != nil, nil
}

func UpsertCapacidadeConfig(ctx context.Context, input map[string]any) (map[string]any, error) {
	idCentral := normalizeCentralID(strVal(input, "id_central"))
	idRep := strings.TrimSpace(strVal(input, "id_representante"))
	minQtd := intVal(input, "quantidade_minima")
	if minQtd <= 0 {
		minQtd = 10
	}
	precoBase := floatVal(input, "preco_base_camera")
	if precoBase <= 0 {
		precoBase = 11.50
	}
	precoVenda := floatVal(input, "preco_venda_camera")

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var precoVendaArg any
	if precoVenda > 0 {
		precoVendaArg = precoVenda
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO vis_capacidade_config (id_central, id_representante, quantidade_minima, preco_base_camera, preco_venda_camera, updated_at)
VALUES ($1,$2,$3,$4,$5,NOW())
ON CONFLICT (id_central, id_representante) DO UPDATE SET
    quantidade_minima = EXCLUDED.quantidade_minima,
    preco_base_camera = EXCLUDED.preco_base_camera,
    preco_venda_camera = EXCLUDED.preco_venda_camera,
    updated_at = NOW()`,
		idCentral, idRep, minQtd, precoBase, precoVendaArg,
	)
	if err != nil {
		return nil, err
	}
	cfg, err := GetCapacidadeConfig(ctx, idCentral, idRep)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id_central":         cfg.IDCentral,
		"id_representante":   cfg.IDRepresentante,
		"quantidade_minima":  cfg.QuantidadeMinima,
		"preco_base_camera":  cfg.PrecoBaseCamera,
		"preco_venda_camera": cfg.PrecoVendaCamera,
		"preco_efetivo_camera": cfg.PrecoEfetivoCamera,
	}, nil
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscan(s, &f)
	return f, err
}

func ListCapacidadePendentesRenovacao(ctx context.Context, diasAntecedencia int) ([]map[string]any, error) {
	_ = ctx
	_ = diasAntecedencia
	return []map[string]any{}, nil
}

func GetContratoCapacidade(ctx context.Context, contratoID int) (map[string]any, error) {
	_ = ctx
	if contratoID <= 0 {
		return nil, fmt.Errorf("contrato_id invalido")
	}
	return map[string]any{"id": contratoID}, nil
}

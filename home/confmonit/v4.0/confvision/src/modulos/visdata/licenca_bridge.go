package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CreateLicencaLote cria licencas pendentes no Postgres (fonte unica — compra Postgres-first).
func CreateLicencaLote(ctx context.Context, records []map[string]any) ([]map[string]any, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("licencas obrigatorio")
	}
	out := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		lic, err := createLicencaPendente(ctx, rec)
		if err != nil {
			return out, err
		}
		out = append(out, lic)
	}
	return out, nil
}

func createLicencaPendente(ctx context.Context, input map[string]any) (map[string]any, error) {
	plano := strVal(input, "plano")
	if plano == "" {
		return nil, fmt.Errorf("plano obrigatorio")
	}
	idFranqueado := strVal(input, "id_franqueado")
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}

	flags := PlanoFlagsFrom(plano)
	unidade := strVal(input, "unidade")
	if unidade == "" {
		unidade = flags.Unidade
	}
	valor := floatVal(input, "valor")
	if valor <= 0 {
		valor = flags.Valor
	}
	status := strVal(input, "status")
	if status == "" {
		status = "pendente"
	}
	obs := strVal(input, "observacao")
	descFatura := strVal(input, "desc_fatura")

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var id int
	var created time.Time
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_licenca (id_franqueado, plano, unidade, valor, status, observacao)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`,
		idFranqueado, plano, unidade, valor, status, nullStrOrNil(obs),
	).Scan(&id, &created)
	if err != nil {
		return nil, err
	}

	item := map[string]any{
		"id":             id,
		"created_at":     created.UTC().Format(time.RFC3339),
		"id_franqueado":  idFranqueado,
		"plano":          plano,
		"unidade":        unidade,
		"valor":          valor,
		"status":         status,
		"observacao":     obs,
	}
	if descFatura != "" {
		item["desc_fatura"] = descFatura
	}
	return item, nil
}

// SyncLicencasFromXanoRecords espelha licencas do Xano (financeiro) no Postgres operacional.
func SyncLicencasFromXanoRecords(ctx context.Context, records []map[string]any) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}
	db, err := DB()
	if err != nil {
		return 0, err
	}

	n := 0
	for _, rec := range records {
		if err := upsertLicencaFromXano(ctx, db, rec); err != nil {
			return n, err
		}
		n++
	}
	if n > 0 {
		_, _ = db.ExecContext(ctx, `
SELECT setval(pg_get_serial_sequence('vis_licenca','id'),
  GREATEST(COALESCE((SELECT MAX(id) FROM vis_licenca), 1),
           COALESCE((SELECT last_value FROM vis_licenca_id_seq), 1)))`)
	}
	return n, nil
}

func upsertLicencaFromXano(ctx context.Context, db *sql.DB, rec map[string]any) error {
	id := intFromAny(rec["id"])
	if id <= 0 {
		return fmt.Errorf("licenca xano sem id")
	}
	idFranqueado := strVal(rec, "id_franqueado")
	plano := strVal(rec, "plano")
	if idFranqueado == "" || plano == "" {
		return fmt.Errorf("licenca %d dados incompletos", id)
	}

	flags := PlanoFlagsFrom(plano)
	unidade := strVal(rec, "unidade")
	if unidade == "" {
		unidade = flags.Unidade
	}
	valor := floatFromAny(rec["valor"])
	if valor <= 0 {
		valor = flags.Valor
	}
	status := strVal(rec, "status")
	if status == "" {
		status = "pendente"
	}
	obs := strVal(rec, "observacao")
	created := parseTimeAny(rec["created_at"])
	pago := parseTimeAny(rec["pago_em"])
	valido := parseTimeAny(rec["valido_ate"])
	idFatura := strVal(rec, "id_fatura")
	idPagamento := strVal(rec, "id_pagamento")
	var camID sql.NullInt64
	if v := intFromAny(rec["vis_camera_id"]); v > 0 {
		camID = sql.NullInt64{Int64: int64(v), Valid: true}
	}

	_, err := db.ExecContext(ctx, `
INSERT INTO vis_licenca (
    id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
    status, id_fatura, id_pagamento, observacao, vis_camera_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (id) DO UPDATE SET
    id_franqueado = EXCLUDED.id_franqueado,
    plano = EXCLUDED.plano,
    unidade = EXCLUDED.unidade,
    valor = EXCLUDED.valor,
    pago_em = EXCLUDED.pago_em,
    valido_ate = EXCLUDED.valido_ate,
    status = EXCLUDED.status,
    id_fatura = EXCLUDED.id_fatura,
    id_pagamento = EXCLUDED.id_pagamento,
    observacao = EXCLUDED.observacao,
    vis_camera_id = COALESCE(EXCLUDED.vis_camera_id, vis_licenca.vis_camera_id)`,
		id, nullTimeOrNow(created), idFranqueado, plano, unidade, valor,
		nullTimePtr(pago), nullTimePtr(valido), status,
		nullStrOrNil(idFatura), nullStrOrNil(idPagamento), nullStrOrNil(obs),
		nullInt64(camID),
	)
	return err
}

// AtivarLicencaPagamento aplica pagamento de fatura no Postgres (espelho do Xano).
func AtivarLicencaPagamento(ctx context.Context, licID, faturaID, pagamentoID int, pagoEm time.Time) (map[string]any, error) {
	if licID <= 0 {
		return nil, fmt.Errorf("vis_licenca_id obrigatorio")
	}
	if pagoEm.IsZero() {
		pagoEm = time.Now().UTC()
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var plano, unidade, status string
	var valido sql.NullTime
	err = db.QueryRowContext(ctx, `
SELECT plano, COALESCE(unidade,''), status, valido_ate
FROM vis_licenca WHERE id = $1`, licID).Scan(&plano, &unidade, &status, &valido)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("licenca nao encontrada")
	}
	if err != nil {
		return nil, err
	}

	if unidade == "gravacao" {
		_ = ensureGravacaoStorageMinimal(ctx, db, licID)
	}

	base := pagoEm
	if valido.Valid && valido.Time.After(base) {
		base = valido.Time
	}
	novoValido := base.Add(30 * 24 * time.Hour)

	novoStatus := status
	switch status {
	case "pendente", "expirada":
		novoStatus = "disponivel"
	case "disponivel", "em_uso":
		novoStatus = status
	}

	idFat := ""
	if faturaID > 0 {
		idFat = strconv.Itoa(faturaID)
	}
	idPag := ""
	if pagamentoID > 0 {
		idPag = strconv.Itoa(pagamentoID)
	}

	_, err = db.ExecContext(ctx, `
UPDATE vis_licenca SET
    valido_ate = $2,
    status = $3,
    pago_em = $4,
    id_fatura = COALESCE(NULLIF($5,''), id_fatura),
    id_pagamento = COALESCE(NULLIF($6,''), id_pagamento)
WHERE id = $1`,
		licID, novoValido, novoStatus, pagoEm, idFat, idPag,
	)
	if err != nil {
		return nil, err
	}

	var idFra string
	_ = db.QueryRowContext(ctx, `SELECT id_franqueado FROM vis_licenca WHERE id = $1`, licID).Scan(&idFra)

	return map[string]any{
		"id":             licID,
		"status":         novoStatus,
		"pago_em":        pagoEm.UTC().Format(time.RFC3339),
		"valido_ate":     novoValido.UTC().Format(time.RFC3339),
		"id_fatura":      idFat,
		"id_pagamento":   idPag,
		"id_franqueado":  idFra,
		"plano":          plano,
	}, nil
}

// EstornarLicencaFatura expira licenca no Postgres (espelho estorno/cancelamento Xano).
func EstornarLicencaFatura(ctx context.Context, licID int, observacao string) (map[string]any, error) {
	if licID <= 0 {
		return nil, fmt.Errorf("vis_licenca_id obrigatorio")
	}
	observacao = strings.TrimSpace(observacao)

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var plano, idFra string
	err = db.QueryRowContext(ctx, `SELECT plano, id_franqueado FROM vis_licenca WHERE id = $1`, licID).Scan(&plano, &idFra)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("licenca nao encontrada")
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var obsArg any
	if observacao != "" {
		obsArg = observacao
	}

	_, err = db.ExecContext(ctx, `
UPDATE vis_licenca SET
    status = 'expirada',
    valido_ate = $2,
    id_fatura = NULL,
    id_pagamento = NULL,
    observacao = COALESCE($3, observacao)
WHERE id = $1`, licID, now, obsArg)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":            licID,
		"status":        "expirada",
		"valido_ate":    now.Format(time.RFC3339),
		"id_franqueado": idFra,
		"plano":         plano,
	}, nil
}

func ensureGravacaoStorageMinimal(ctx context.Context, db *sql.DB, licID int) error {
	var idFra string
	if err := db.QueryRowContext(ctx, `SELECT id_franqueado FROM vis_licenca WHERE id = $1`, licID).Scan(&idFra); err != nil {
		return err
	}
	idFra = strings.TrimSpace(idFra)
	if idFra == "" {
		return nil
	}
	var n int
	_ = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_gravacao_storage
WHERE id_franqueado = $1 AND status = 'ativo'`, idFra).Scan(&n)
	if n > 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO vis_gravacao_storage (id_franqueado, status, provisionado_em, observacao)
VALUES ($1, 'ativo', NOW(), 'Provisionado automaticamente na ativacao da licenca')`, idFra)
	return err
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	default:
		return 0
	}
}

func floatFromAny(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		return 0
	}
}

func parseTimeAny(v any) time.Time {
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return time.Time{}
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02 15:04:05"} {
			if parsed, err := time.Parse(layout, s); err == nil {
				return parsed.UTC()
			}
		}
	case float64:
		if t > 1e12 {
			return time.UnixMilli(int64(t)).UTC()
		}
		return time.Unix(int64(t), 0).UTC()
	}
	return time.Time{}
}

func nullTimeOrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

func nullTimePtr(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func nullStrOrNil(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func nullInt64(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}

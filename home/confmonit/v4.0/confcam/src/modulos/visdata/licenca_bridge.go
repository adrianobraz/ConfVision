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

// GetLicencaByID retorna uma licenca do Postgres (fonte unica).
func GetLicencaByID(ctx context.Context, licID int) (map[string]any, error) {
	if licID <= 0 {
		return nil, fmt.Errorf("vis_licenca_id obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var created sql.NullTime
	var idFra, plano, unid, st, idDisp, idFat, idPag, obs sql.NullString
	var valor sql.NullFloat64
	var pago, valido sql.NullTime
	var camID sql.NullInt64
	err = db.QueryRowContext(ctx, `
SELECT id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
       status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
FROM vis_licenca WHERE id = $1`, licID).Scan(
		&licID, &created, &idFra, &plano, &unid, &valor, &pago, &valido,
		&st, &idDisp, &idFat, &idPag, &obs, &camID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("licenca nao encontrada")
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id": licID, "created_at": nullTime(created), "id_franqueado": nullStr(idFra),
		"plano": nullStr(plano), "unidade": nullStr(unid), "valor": nullFloat(valor),
		"pago_em": nullTime(pago), "valido_ate": nullTime(valido), "status": nullStr(st),
		"id_dispositivo": nullStr(idDisp), "id_fatura": nullStr(idFat),
		"id_pagamento": nullStr(idPag), "observacao": nullStr(obs),
		"vis_camera_id": nullInt(camID),
	}, nil
}

// ExpirarLicencasVencidas marca licencas vencidas como expirada no Postgres.
func ExpirarLicencasVencidas(ctx context.Context) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rows, err := db.QueryContext(ctx, `
SELECT id, status, vis_camera_id, id_franqueado
FROM vis_licenca
WHERE valido_ate IS NOT NULL AND valido_ate < $1
  AND status IN ('disponivel','em_uso')`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expiradas := 0
	franqueados := map[string]bool{}
	for rows.Next() {
		var id int
		var status, idFra sql.NullString
		var camID sql.NullInt64
		if err := rows.Scan(&id, &status, &camID, &idFra); err != nil {
			return nil, err
		}
		_, err = db.ExecContext(ctx, `UPDATE vis_licenca SET status = 'expirada' WHERE id = $1`, id)
		if err != nil {
			return nil, err
		}
		if status.Valid && status.String == "em_uso" && camID.Valid && camID.Int64 > 0 {
			_, _ = db.ExecContext(ctx, `UPDATE vis_camera SET ativo = false WHERE id = $1`, camID.Int64)
		}
		expiradas++
		if idFra.Valid && strings.TrimSpace(idFra.String) != "" {
			franqueados[idFra.String] = true
		}
	}
	fraList := make([]string, 0, len(franqueados))
	for k := range franqueados {
		fraList = append(fraList, k)
	}
	return map[string]any{
		"expiradas": expiradas,
		"franqueados": fraList,
	}, nil
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
	pago := parseTimeAny(rec["pago_em"])
	valido := parseTimeAny(rec["valido_ate"])
	idFatura := strVal(rec, "id_fatura")
	idPagamento := strVal(rec, "id_pagamento")
	obs := strVal(rec, "observacao")
	camID := intFromAny(rec["vis_camera_id"])

	_, err := db.ExecContext(ctx, `
INSERT INTO vis_licenca (
    id, id_franqueado, plano, unidade, valor, status, observacao,
    pago_em, valido_ate, id_fatura, id_pagamento, vis_camera_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (id) DO UPDATE SET
    id_franqueado = EXCLUDED.id_franqueado,
    plano = EXCLUDED.plano,
    unidade = EXCLUDED.unidade,
    valor = EXCLUDED.valor,
    status = EXCLUDED.status,
    observacao = EXCLUDED.observacao,
    pago_em = COALESCE(EXCLUDED.pago_em, vis_licenca.pago_em),
    valido_ate = COALESCE(EXCLUDED.valido_ate, vis_licenca.valido_ate),
    id_fatura = COALESCE(NULLIF(EXCLUDED.id_fatura,''), vis_licenca.id_fatura),
    id_pagamento = COALESCE(NULLIF(EXCLUDED.id_pagamento,''), vis_licenca.id_pagamento),
    vis_camera_id = COALESCE(EXCLUDED.vis_camera_id, vis_licenca.vis_camera_id)`,
		id, idFranqueado, plano, unidade, valor, status, nullStrOrNil(obs),
		nullTimePtr(pago), nullTimePtr(valido), nullStrOrNil(idFatura), nullStrOrNil(idPagamento), nullInt64(sql.NullInt64{Int64: int64(camID), Valid: camID > 0}),
	)
	return err
}

// AtivarLicencaPagamento aplica pagamento de fatura no Postgres (fonte unica).
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
	var licCamID sql.NullInt64
	err = db.QueryRowContext(ctx, `
SELECT plano, COALESCE(unidade,''), status, valido_ate, vis_camera_id
FROM vis_licenca WHERE id = $1`, licID).Scan(&plano, &unidade, &status, &valido, &licCamID)
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
	case "pendente":
		novoStatus = "disponivel"
	case "expirada":
		if licencaTemVinculoCamera(ctx, db, licID, licCamID) {
			novoStatus = "em_uso"
		} else {
			novoStatus = "disponivel"
		}
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

	if novoStatus == "em_uso" || (status == "em_uso" && licencaTemVinculoCamera(ctx, db, licID, licCamID)) {
		if err := reativarCameraPosRenovacaoLicenca(ctx, db, licID, plano, licCamID); err != nil {
			return nil, err
		}
		novoStatus = "em_uso"
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

	obsFinal := observacao
	if obsFinal == "" {
		obsFinal = "Estorno/cancelamento fatura"
	}

	_, err = db.ExecContext(ctx, `
UPDATE vis_licenca SET
    status = 'expirada',
    observacao = COALESCE(NULLIF($2,''), observacao)
WHERE id = $1`, licID, obsFinal)
	if err != nil {
		return nil, err
	}

	var camID sql.NullInt64
	_ = db.QueryRowContext(ctx, `SELECT vis_camera_id FROM vis_licenca WHERE id = $1`, licID).Scan(&camID)
	if camID.Valid && camID.Int64 > 0 {
		_, _ = db.ExecContext(ctx, `UPDATE vis_camera SET ativo = false WHERE id = $1`, camID.Int64)
	}

	return map[string]any{
		"id":            licID,
		"status":        "expirada",
		"id_franqueado": idFra,
		"plano":         plano,
	}, nil
}

func licencaTemVinculoCamera(ctx context.Context, db *sql.DB, licID int, licCamID sql.NullInt64) bool {
	if licCamID.Valid && licCamID.Int64 > 0 {
		return true
	}
	var id int
	err := db.QueryRowContext(ctx, `
SELECT id FROM vis_camera WHERE vis_licenca_id = $1 LIMIT 1`, licID).Scan(&id)
	return err == nil && id > 0
}

func reativarCameraPosRenovacaoLicenca(ctx context.Context, db *sql.DB, licID int, plano string, licCamID sql.NullInt64) error {
	camID := int64(0)
	if licCamID.Valid {
		camID = licCamID.Int64
	}
	if camID <= 0 {
		err := db.QueryRowContext(ctx, `
SELECT id FROM vis_camera WHERE vis_licenca_id = $1 LIMIT 1`, licID).Scan(&camID)
		if err != nil || camID <= 0 {
			return nil
		}
	}

	flags := PlanoFlagsFrom(plano)
	ativo := true
	if flags.SemAtivo || plano == "online" {
		ativo = false
	}

	_, err := db.ExecContext(ctx, `
UPDATE vis_licenca SET status = 'em_uso', vis_camera_id = $2 WHERE id = $1`, licID, camID)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET vis_licenca_id = $1, plano = $2, ativo = $3 WHERE id = $4`,
		licID, plano, ativo, camID)
	return err
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

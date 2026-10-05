package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

func ListGravacaoStorageByFranqueado(ctx context.Context, idFranqueado, status string, mask bool) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	q := `SELECT id, created_at, id_franqueado, s3_endpoint, s3_bucket, s3_tenant_id,
        s3_access_key, s3_secret_key, segmento_minutos, status, provisionado_em, cancelado_em, observacao
FROM vis_gravacao_storage WHERE id_franqueado = $1`
	args := []any{idFranqueado}
	if status != "" {
		q += " AND status = $2"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC"
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, segMin int
		var created, prov, cancel sql.NullTime
		var idFra, endpoint, bucket, tenant, access, secret, st, obs sql.NullString
		if err := rows.Scan(&id, &created, &idFra, &endpoint, &bucket, &tenant, &access, &secret, &segMin, &st, &prov, &cancel, &obs); err != nil {
			return nil, err
		}
		acc := access.String
		sec := secret.String
		if mask {
			acc = maskSecret(acc)
			sec = maskSecret(sec)
		}
		out = append(out, map[string]any{
			"id": id, "created_at": nullTime(created), "id_franqueado": nullStr(idFra),
			"s3_endpoint": nullStr(endpoint), "s3_bucket": nullStr(bucket),
			"s3_tenant_id": nullStr(tenant), "s3_access_key": acc, "s3_secret_key": sec,
			"segmento_minutos": segMin, "status": nullStr(st),
			"provisionado_em": nullTime(prov), "cancelado_em": nullTime(cancel),
			"observacao": nullStr(obs),
		})
	}
	return out, nil
}

func CreateGravacaoStorage(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	var created time.Time
	seg := intVal(input, "segmento_minutos")
	if seg <= 0 {
		seg = 5
	}
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_gravacao_storage (
    id_franqueado, s3_endpoint, s3_bucket, s3_tenant_id, s3_access_key, s3_secret_key,
    segmento_minutos, status, provisionado_em, observacao
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),$9) RETURNING id, created_at`,
		strVal(input, "id_franqueado"), strVal(input, "s3_endpoint"), strVal(input, "s3_bucket"),
		strVal(input, "s3_tenant_id"), strVal(input, "s3_access_key"), strVal(input, "s3_secret_key"),
		seg, firstNonEmpty(strVal(input, "status"), "ativo"), strVal(input, "observacao"),
	).Scan(&id, &created)
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "created_at": created.UTC().Format(time.RFC3339)}, nil
}

func UpdateGravacaoStorage(ctx context.Context, id int, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `
UPDATE vis_gravacao_storage SET
    s3_endpoint = COALESCE(NULLIF($2,''), s3_endpoint),
    s3_bucket = COALESCE(NULLIF($3,''), s3_bucket),
    s3_tenant_id = COALESCE(NULLIF($4,''), s3_tenant_id),
    s3_access_key = COALESCE(NULLIF($5,''), s3_access_key),
    s3_secret_key = COALESCE(NULLIF($6,''), s3_secret_key),
    segmento_minutos = COALESCE(NULLIF($7,0), segmento_minutos),
    status = COALESCE(NULLIF($8,''), status),
    observacao = COALESCE(NULLIF($9,''), observacao)
WHERE id = $1`,
		id, strVal(input, "s3_endpoint"), strVal(input, "s3_bucket"), strVal(input, "s3_tenant_id"),
		strVal(input, "s3_access_key"), strVal(input, "s3_secret_key"), intVal(input, "segmento_minutos"),
		strVal(input, "status"), strVal(input, "observacao"))
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

func PostGravacaoSegmento(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_gravacao_segmento (
    vis_camera_id, vis_gravacao_storage_id, id_franqueado, id_cliente,
    inicio_em, fim_em, duracao_seg, s3_key, s3_url, tamanho_bytes, status, uploaded_em, expira_em, tipo
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),$12,$13) RETURNING id`,
		intVal(input, "vis_camera_id"), intVal(input, "vis_gravacao_storage_id"),
		strVal(input, "id_franqueado"), strVal(input, "id_cliente"),
		parseTimeInput(input, "inicio_em"), parseTimeInput(input, "fim_em"),
		intVal(input, "duracao_seg"), strVal(input, "s3_key"), strVal(input, "s3_url"),
		intVal(input, "tamanho_bytes"), firstNonEmpty(strVal(input, "status"), "ok"),
		parseTimeInput(input, "expira_em"), strVal(input, "tipo"),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

func ListGravacaoSegmentos(ctx context.Context, idFranqueado string, cameraID, page int, de, ate string) (map[string]any, error) {
	if page <= 0 {
		page = 1
	}
	pageSize := 50
	offset := (page - 1) * pageSize
	db, err := DB()
	if err != nil {
		return nil, err
	}
	where := "WHERE 1=1"
	args := []any{}
	n := 1
	if cameraID > 0 {
		where += fmt.Sprintf(" AND vis_camera_id = $%d", n)
		args = append(args, cameraID)
		n++
	} else if idFranqueado != "" {
		where += fmt.Sprintf(" AND id_franqueado = $%d", n)
		args = append(args, idFranqueado)
		n++
	}
	if de != "" {
		where += fmt.Sprintf(" AND inicio_em >= $%d", n)
		args = append(args, de)
		n++
	}
	if ate != "" {
		where += fmt.Sprintf(" AND inicio_em <= $%d", n)
		args = append(args, ate)
		n++
	}
	var total int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vis_gravacao_segmento "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	q := fmt.Sprintf(`
SELECT id, created_at, vis_camera_id, id_franqueado, id_cliente, inicio_em, fim_em,
       duracao_seg, s3_key, s3_url, tamanho_bytes, status, tipo
FROM vis_gravacao_segmento %s ORDER BY inicio_em DESC LIMIT %d OFFSET %d`, where, pageSize, offset)
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, camID, dur, size int
		var created, inicio, fim sql.NullTime
		var idFra, idCli, s3key, s3url, st, tipo sql.NullString
		if err := rows.Scan(&id, &created, &camID, &idFra, &idCli, &inicio, &fim, &dur, &s3key, &s3url, &size, &st, &tipo); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id": id, "created_at": nullTime(created), "vis_camera_id": camID,
			"id_franqueado": nullStr(idFra), "id_cliente": nullStr(idCli),
			"inicio_em": nullTime(inicio), "fim_em": nullTime(fim), "duracao_seg": dur,
			"s3_key": nullStr(s3key), "s3_url": nullStr(s3url), "tamanho_bytes": size,
			"status": nullStr(st), "tipo": nullStr(tipo),
		})
	}
	return map[string]any{"dados": map[string]any{
		"items": items, "itemsReceived": len(items), "curPage": page,
		"nextPage": page + 1, "totalItems": total,
	}}, nil
}

func parseIntQuery(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

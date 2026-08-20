package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func CreateEvento(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	var created time.Time
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_evento (
    vis_camera_id, id_franqueado, id_cliente, id_dispositivo, conta, particao, canal,
    tipo_deteccao, confianca, snapshot_url, video_url, bbox_json, processado, ignorado,
    status, id_evento, id_processo, started_at, ended_at, clip_count
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
RETURNING id, created_at`,
		intVal(input, "vis_camera_id"), strVal(input, "id_franqueado"), strVal(input, "id_cliente"),
		strVal(input, "id_dispositivo"), strVal(input, "conta"), strVal(input, "particao"),
		strVal(input, "canal"), strVal(input, "tipo_deteccao"), floatInput(input, "confianca"),
		strVal(input, "snapshot_url"), strVal(input, "video_url"), strVal(input, "bbox_json"),
		boolDefault(input, "processado", false), boolDefault(input, "ignorado", false),
		strVal(input, "status"), strVal(input, "id_evento"), strVal(input, "id_processo"),
		parseTimeInput(input, "started_at"), parseTimeInput(input, "ended_at"), intVal(input, "clip_count"),
	).Scan(&id, &created)
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "created_at": created.UTC().Format(time.RFC3339)}, nil
}

func floatInput(m map[string]any, key string) any {
	v := strVal(m, key)
	if v == "" {
		if f, ok := m[key].(float64); ok {
			return f
		}
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return f
}

func boolDefault(m map[string]any, key string, def bool) bool {
	if v := boolVal(m, key); v != nil {
		return *v
	}
	return def
}

func FinalizarEvento(ctx context.Context, input map[string]any) (map[string]any, error) {
	eventoID := intVal(input, "vis_evento_id")
	if eventoID <= 0 {
		return nil, fmt.Errorf("vis_evento_id obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	status := strVal(input, "status")
	if status == "" {
		status = "pronto"
	}
	_, err = db.ExecContext(ctx, `
UPDATE vis_evento SET snapshot_url = $2, video_url = $3, status = $4,
    clip_count = $5, processado = $6 WHERE id = $1`,
		eventoID, strVal(input, "snapshot_url"), strVal(input, "video_url"), status,
		intVal(input, "clip_count"), boolDefault(input, "processado", true))
	if err != nil {
		return nil, err
	}

	var clip map[string]any
	videoURL := strVal(input, "video_url")
	if videoURL != "" {
		seq := intVal(input, "clip_seq")
		if seq <= 0 {
			seq = 1
		}
		var clipID int
		err = db.QueryRowContext(ctx, `
INSERT INTO vis_evento_clip (vis_evento_id, seq, video_url, duracao_seg, snapshot_url)
VALUES ($1,$2,$3,$4,$5) RETURNING id`,
			eventoID, seq, videoURL, intVal(input, "clip_duracao_seg"),
			firstNonEmpty(strVal(input, "clip_snapshot_url"), strVal(input, "snapshot_url")),
		).Scan(&clipID)
		if err == nil {
			clip = map[string]any{"id": clipID, "vis_evento_id": eventoID, "seq": seq, "video_url": videoURL}
		}
	}

	row := db.QueryRowContext(ctx, `SELECT id, created_at, status, snapshot_url, video_url, clip_count FROM vis_evento WHERE id = $1`, eventoID)
	var id, clipCount int
	var created time.Time
	var st, snap, vid sql.NullString
	if err := row.Scan(&id, &created, &st, &snap, &vid, &clipCount); err != nil {
		return nil, err
	}
	evento := map[string]any{
		"id": id, "created_at": created.UTC().Format(time.RFC3339),
		"status": nullStr(st), "snapshot_url": nullStr(snap), "video_url": nullStr(vid),
		"clip_count": clipCount,
	}
	maybeScheduleIntegracaoDispatch(ctx, eventoID, sqlString(snap))
	return map[string]any{"evento": evento, "clip": clip}, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func ListEventosPage(ctx context.Context, idFranqueado, idCliente string, page int, dataDe, dataAte string) (map[string]any, error) {
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
	if idFranqueado != "" {
		where += fmt.Sprintf(" AND id_franqueado = $%d", n)
		args = append(args, idFranqueado)
		n++
	}
	if idCliente != "" {
		where += fmt.Sprintf(" AND id_cliente = $%d", n)
		args = append(args, idCliente)
		n++
	}
	if dataDe != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", n)
		args = append(args, dataDe)
		n++
	}
	if dataAte != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", n)
		args = append(args, dataAte)
		n++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM vis_evento " + where
	if err := db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	listQ := fmt.Sprintf(`
SELECT id, created_at, vis_camera_id, id_franqueado, id_cliente, id_dispositivo, conta, particao, canal,
       tipo_deteccao, confianca, snapshot_url, video_url, status, clip_count, processado
FROM vis_evento %s ORDER BY created_at DESC LIMIT %d OFFSET %d`, where, pageSize, offset)
	rows, err := db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		var id, camID, clipCount int
		var created time.Time
		var idFra, idCli, idDisp, conta, part, canal, tipo, snap, vid, st sql.NullString
		var conf sql.NullFloat64
		var proc sql.NullBool
		if err := rows.Scan(&id, &created, &camID, &idFra, &idCli, &idDisp, &conta, &part, &canal,
			&tipo, &conf, &snap, &vid, &st, &clipCount, &proc); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id": id, "created_at": created.UTC().Format(time.RFC3339),
			"vis_camera_id": camID, "id_franqueado": nullStr(idFra), "id_cliente": nullStr(idCli),
			"id_dispositivo": nullStr(idDisp), "conta": nullStr(conta), "particao": nullStr(part),
			"canal": nullStr(canal), "tipo_deteccao": nullStr(tipo), "confianca": nullFloat(conf),
			"snapshot_url": nullStr(snap), "video_url": nullStr(vid), "status": nullStr(st),
			"clip_count": clipCount, "processado": nullBool(proc),
		})
	}

	nextPage := page + 1
	if offset+len(items) >= total {
		nextPage = 0
	}
	return map[string]any{
		"dados": map[string]any{
			"items":         items,
			"itemsReceived": len(items),
			"curPage":       page,
			"nextPage":      nextPage,
			"prevPage":      page - 1,
			"totalItems":    total,
		},
	}, nil
}

func GetEventoClips(ctx context.Context, eventoID int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `
SELECT id, created_at, vis_camera_id, snapshot_url, video_url, status, clip_count
FROM vis_evento WHERE id = $1`, eventoID)
	var id, camID, clipCount int
	var created time.Time
	var snap, vid, st sql.NullString
	if err := row.Scan(&id, &created, &camID, &snap, &vid, &st, &clipCount); err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, vis_evento_id, seq, video_url, duracao_seg, snapshot_url
FROM vis_evento_clip WHERE vis_evento_id = $1 ORDER BY seq ASC`, eventoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	clips := []map[string]any{}
	for rows.Next() {
		var cid, eid, seq, dur int
		var ccreated time.Time
		var vurl, ss sql.NullString
		if err := rows.Scan(&cid, &ccreated, &eid, &seq, &vurl, &dur, &ss); err != nil {
			return nil, err
		}
		clips = append(clips, map[string]any{
			"id": cid, "created_at": ccreated.UTC().Format(time.RFC3339),
			"vis_evento_id": eid, "seq": seq, "video_url": nullStr(vurl),
			"duracao_seg": dur, "snapshot_url": nullStr(ss),
		})
	}
	return map[string]any{
		"evento": map[string]any{
			"id": id, "created_at": created.UTC().Format(time.RFC3339),
			"vis_camera_id": camID, "snapshot_url": nullStr(snap), "video_url": nullStr(vid),
			"status": nullStr(st), "clip_count": clipCount,
		},
		"clips": clips,
	}, nil
}

func ListEventosSensorPendentes(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
SELECT id, created_at, vis_camera_id, id_franqueado, id_cliente, id_dispositivo, status, snapshot_url
FROM vis_evento
WHERE processado = FALSE AND tipo_deteccao = 'sensor'
ORDER BY created_at ASC LIMIT %d`, limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, camID int
		var created time.Time
		var idFra, idCli, idDisp, st, snap sql.NullString
		if err := rows.Scan(&id, &created, &camID, &idFra, &idCli, &idDisp, &st, &snap); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id": id, "created_at": created.UTC().Format(time.RFC3339),
			"vis_camera_id": camID, "id_franqueado": nullStr(idFra), "id_cliente": nullStr(idCli),
			"id_dispositivo": nullStr(idDisp), "status": nullStr(st), "snapshot_url": nullStr(snap),
		})
	}
	return out, nil
}

func UpdateEvento(ctx context.Context, id int, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	sets := []string{}
	args := []any{}
	n := 1
	for _, col := range []string{"snapshot_url", "video_url", "status", "bbox_json", "id_evento", "id_processo"} {
		if v, ok := input[col]; ok {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, trimAny(v))
			n++
		}
	}
	if v := boolVal(input, "processado"); v != nil {
		sets = append(sets, fmt.Sprintf("processado = $%d", n))
		args = append(args, *v)
		n++
	}
	if v := strVal(input, "codigo_imagem_publico"); v != "" {
		sets = append(sets, fmt.Sprintf("codigo_imagem_publico = $%d", n))
		args = append(args, v)
		n++
		sets = append(sets, "imagem_liberada_em = NOW()")
	}
	if len(sets) == 0 {
		return map[string]any{"id": id}, nil
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE vis_evento SET %s WHERE id = $%d", strings.Join(sets, ", "), n)
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	if v, ok := input["snapshot_url"]; ok && strings.TrimSpace(trimAny(v)) != "" {
		maybeScheduleIntegracaoDispatch(ctx, id, trimAny(v))
	}
	return map[string]any{"id": id}, nil
}

func CreateEventoClip(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_evento_clip (vis_evento_id, seq, video_url, duracao_seg, snapshot_url)
VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		intVal(input, "vis_evento_id"), intVal(input, "seq"), strVal(input, "video_url"),
		intVal(input, "duracao_seg"), strVal(input, "snapshot_url"),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

// UpsertEventoDemoMoni insere/atualiza evento demo Moni com id fixo (worker API).
func UpsertEventoDemoMoni(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	eventoID := intVal(input, "id")
	if eventoID < 1 {
		return nil, fmt.Errorf("id obrigatorio")
	}
	snapURL := strVal(input, "snapshot_url")
	hash := strVal(input, "codigo_imagem_publico")
	if snapURL == "" || hash == "" {
		return nil, fmt.Errorf("snapshot_url e codigo_imagem_publico obrigatorios")
	}
	idFra := strVal(input, "id_franqueado")
	if idFra == "" {
		idFra = "0"
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO vis_evento (
	id, id_franqueado, id_cliente, conta, particao, canal,
	tipo_deteccao, snapshot_url,
	codigo_imagem_publico, imagem_liberada_em
) VALUES (
	$1, $2, '0', '0000', '00', '001',
	'movimento', $3,
	$4, NOW()
)
ON CONFLICT (id) DO UPDATE SET
	snapshot_url = EXCLUDED.snapshot_url,
	codigo_imagem_publico = EXCLUDED.codigo_imagem_publico,
	imagem_liberada_em = NOW()`,
		eventoID, idFra, snapURL, hash,
	)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id": eventoID, "snapshot_url": snapURL, "codigo_imagem_publico": hash,
	}, nil
}

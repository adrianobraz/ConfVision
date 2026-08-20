package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const eventoSelectCols = `
e.id, e.created_at, e.vis_camera_id, e.id_franqueado, e.id_cliente, e.id_dispositivo,
e.conta, e.particao, e.canal, e.tipo_deteccao, e.confianca, e.snapshot_url, e.video_url,
e.bbox_json, e.processado, e.alarm_events_id, e.ignorado, e.status, e.id_evento,
e.id_processo, e.started_at, e.ended_at, e.clip_count
`

func scanEventoRows(rows *sql.Rows) ([]map[string]any, error) {
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, camID, alarmID, clipCount sql.NullInt64
		var created, started, ended sql.NullTime
		var idFra, idCli, idDisp, conta, part, canal, tipo sql.NullString
		var snap, vid, bbox, st, idEvt, idProc sql.NullString
		var conf sql.NullFloat64
		var proc, ign sql.NullBool
		if err := rows.Scan(
			&id, &created, &camID, &idFra, &idCli, &idDisp, &conta, &part, &canal,
			&tipo, &conf, &snap, &vid, &bbox, &proc, &alarmID, &ign, &st, &idEvt,
			&idProc, &started, &ended, &clipCount,
		); err != nil {
			return nil, err
		}
		out = append(out, eventoRowMap(
			id, camID, alarmID, clipCount,
			created, started, ended,
			idFra, idCli, idDisp, conta, part, canal, tipo,
			conf, snap, vid, bbox, st, idEvt, idProc,
			proc, ign,
		))
	}
	return out, rows.Err()
}

func scanEventoRow(row *sql.Row) (map[string]any, error) {
	var id, camID, alarmID, clipCount sql.NullInt64
	var created, started, ended sql.NullTime
	var idFra, idCli, idDisp, conta, part, canal, tipo sql.NullString
	var snap, vid, bbox, st, idEvt, idProc sql.NullString
	var conf sql.NullFloat64
	var proc, ign sql.NullBool
	if err := row.Scan(
		&id, &created, &camID, &idFra, &idCli, &idDisp, &conta, &part, &canal,
		&tipo, &conf, &snap, &vid, &bbox, &proc, &alarmID, &ign, &st, &idEvt,
		&idProc, &started, &ended, &clipCount,
	); err != nil {
		return nil, err
	}
	m := eventoRowMap(
		id, camID, alarmID, clipCount,
		created, started, ended,
		idFra, idCli, idDisp, conta, part, canal, tipo,
		conf, snap, vid, bbox, st, idEvt, idProc,
		proc, ign,
	)
	return m, nil
}

func eventoRowMap(
	id, camID, alarmID, clipCount sql.NullInt64,
	created, started, ended sql.NullTime,
	idFra, idCli, idDisp, conta, part, canal, tipo sql.NullString,
	conf sql.NullFloat64,
	snap, vid, bbox, st, idEvt, idProc sql.NullString,
	proc, ign sql.NullBool,
) map[string]any {
	m := map[string]any{
		"id":             nullInt(id),
		"vis_camera_id":  nullInt(camID),
		"id_franqueado":  nullStr(idFra),
		"id_cliente":     nullStr(idCli),
		"id_dispositivo": nullStr(idDisp),
		"conta":          nullStr(conta),
		"particao":       nullStr(part),
		"canal":          nullStr(canal),
		"tipo_deteccao":  nullStr(tipo),
		"confianca":      nullFloat(conf),
		"snapshot_url":   nullStr(snap),
		"video_url":      nullStr(vid),
		"bbox_json":      nullStr(bbox),
		"processado":     nullBool(proc),
		"alarm_events_id": nullInt(alarmID),
		"ignorado":       nullBool(ign),
		"status":         nullStr(st),
		"id_evento":      nullStr(idEvt),
		"id_processo":    nullStr(idProc),
		"clip_count":     nullInt(clipCount),
	}
	if created.Valid {
		m["created_at"] = created.Time.UTC().Format(time.RFC3339)
	}
	if started.Valid {
		m["started_at"] = started.Time.UTC().Format(time.RFC3339)
	}
	if ended.Valid {
		m["ended_at"] = ended.Time.UTC().Format(time.RFC3339)
	}
	return m
}

// DisparoSensorEvento cria evento de captura por alarme (receptor → licenca sensor).
func DisparoSensorEvento(ctx context.Context, input map[string]any) (map[string]any, error) {
	idDisp := strVal(input, "id_dispositivo")
	particao := strVal(input, "particao")
	zonauser := strVal(input, "zonauser")
	if idDisp == "" || particao == "" || zonauser == "" {
		return nil, fmt.Errorf("id_dispositivo, particao e zonauser obrigatorios")
	}

	camera, err := FindCameraSensor(ctx, idDisp, particao, zonauser)
	if err != nil {
		return nil, err
	}
	if camera == nil {
		return nil, fmt.Errorf("Camera sensor nao encontrada para o setor")
	}

	gravaFoto := boolDefault(camera, "evento_grava_foto", false)
	gravaVideo := boolDefault(camera, "evento_grava_video", false)
	if !gravaFoto && !gravaVideo {
		return nil, fmt.Errorf("Licenca sensor sem permissao de foto ou video")
	}

	cameraID := intVal(camera, "id")
	canal := strVal(camera, "canal")

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var alarmID any
	if v := intVal(input, "alarm_events_id"); v > 0 {
		alarmID = v
	}

	var id int
	var created time.Time
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_evento (
    vis_camera_id, id_franqueado, id_cliente, id_dispositivo, conta, particao, canal,
    tipo_deteccao, confianca, processado, ignorado, status, id_evento, id_processo,
    alarm_events_id, clip_count
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    'sensor', 1, FALSE, FALSE, 'capturando', $8, $9,
    $10, 0
) RETURNING id, created_at`,
		cameraID,
		firstNonEmpty(strVal(input, "id_franqueado"), strVal(camera, "id_franqueado")),
		firstNonEmpty(strVal(input, "id_cliente"), strVal(camera, "id_cliente")),
		idDisp,
		firstNonEmpty(strVal(input, "conta"), strVal(camera, "conta")),
		particao,
		canal,
		strVal(input, "id_evento"),
		strVal(input, "id_processo"),
		alarmID,
	).Scan(&id, &created)
	if err != nil {
		return nil, err
	}

	evento := map[string]any{
		"id": id, "created_at": created.UTC().Format(time.RFC3339),
		"vis_camera_id": cameraID, "tipo_deteccao": "sensor", "status": "capturando",
		"processado": false, "clip_count": 0,
	}
	return map[string]any{"dados": evento, "camera": camera}, nil
}

func FindCameraSensor(ctx context.Context, idDispositivo, particao, zonauser string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+cameraSelectCols+`
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.id_dispositivo = $1
  AND c.particao = $2
  AND c.zonauser = $3
  AND c.captura_sensor = TRUE
ORDER BY c.id DESC
LIMIT 1`, idDispositivo, particao, zonauser)
	if err != nil {
		return nil, err
	}
	list, err := scanCameraRows(rows)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return list[0], nil
}

func GetCameraBySetor(ctx context.Context, idDispositivo, particao, zonauser, idFranqueado string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	query := `
SELECT ` + cameraSelectCols + `
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.id_dispositivo = $1
  AND c.particao = $2
  AND c.zonauser = $3`
	args := []any{idDispositivo, particao, zonauser}
	if idFranqueado != "" {
		query += " AND c.id_franqueado = $4"
		args = append(args, idFranqueado)
	}
	query += " ORDER BY c.id DESC LIMIT 1"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	list, err := scanCameraRows(rows)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return list[0], nil
}

func ListEventosByContexto(ctx context.Context, idProcesso, idCliente, idDispositivo, idFranqueado, dataInicio string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	where := "WHERE 1=1"
	args := []any{}
	n := 1

	if idFranqueado != "" {
		where += fmt.Sprintf(" AND e.id_franqueado = $%d", n)
		args = append(args, idFranqueado)
		n++
	}

	var orParts []string
	if idProcesso != "" {
		orParts = append(orParts, fmt.Sprintf("e.id_processo = $%d", n))
		args = append(args, idProcesso)
		n++
	}
	if idCliente != "" && idDispositivo != "" && dataInicio != "" {
		orParts = append(orParts, fmt.Sprintf(
			"(e.id_cliente = $%d AND e.id_dispositivo = $%d AND e.created_at >= $%d)",
			n, n+1, n+2,
		))
		args = append(args, idCliente, idDispositivo, dataInicio)
		n += 3
	}
	if len(orParts) == 0 {
		return []map[string]any{}, nil
	}
	where += " AND (" + strings.Join(orParts, " OR ") + ")"

	q := fmt.Sprintf(`
SELECT %s FROM vis_evento e %s ORDER BY e.created_at DESC LIMIT 500`, eventoSelectCols, where)
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanEventoRows(rows)
}

func GetEventoByProcessoSetor(ctx context.Context, idProcesso, idDispositivo, particao, zonauser, idEvento string) (map[string]any, map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, nil, err
	}

	var evento map[string]any
	var camera map[string]any

	if idEvento != "" {
		row := db.QueryRowContext(ctx, `
SELECT `+eventoSelectCols+`
FROM vis_evento e
WHERE e.id_evento = $1
  AND ($2 = '' OR e.id_processo = $2)
ORDER BY e.created_at DESC
LIMIT 1`, idEvento, idProcesso)
		ev, err := scanEventoRow(row)
		if err == nil {
			evento = ev
		} else if err != sql.ErrNoRows {
			return nil, nil, err
		}
	}

	if evento == nil {
		cam, err := GetCameraBySetor(ctx, idDispositivo, particao, zonauser, "")
		if err != nil {
			return nil, nil, err
		}
		if cam == nil {
			return nil, nil, fmt.Errorf("Camera nao encontrada para o setor")
		}
		camera = cam
		cameraID := intVal(cam, "id")

		row := db.QueryRowContext(ctx, `
SELECT `+eventoSelectCols+`
FROM vis_evento e
WHERE e.id_processo = $1
  AND e.vis_camera_id = $2
  AND ($3 = '' OR e.particao = $3)
ORDER BY e.created_at DESC
LIMIT 1`, idProcesso, cameraID, particao)
		ev, err := scanEventoRow(row)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, err
		}
		evento = ev
	}

	if camera == nil && evento != nil {
		camID := intVal(evento, "vis_camera_id")
		if camID > 0 {
			cam, err := GetCameraByID(ctx, camID)
			if err == nil {
				camera = cam
			}
		}
	}

	return evento, camera, nil
}

func ListEventosUltimos25Setor(ctx context.Context, idDispositivo, particao, zonauser string) ([]map[string]any, map[string]any, error) {
	camera, err := GetCameraBySetor(ctx, idDispositivo, particao, zonauser, "")
	if err != nil {
		return nil, nil, err
	}
	if camera == nil {
		return nil, nil, fmt.Errorf("Camera nao encontrada para o setor")
	}
	cameraID := intVal(camera, "id")

	db, err := DB()
	if err != nil {
		return nil, nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+eventoSelectCols+`
FROM vis_evento e
WHERE e.vis_camera_id = $1
  AND e.id_dispositivo = $2
ORDER BY e.created_at DESC
LIMIT 25`, cameraID, idDispositivo)
	if err != nil {
		return nil, nil, err
	}
	list, err := scanEventoRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return list, camera, nil
}

func ListEventosUltimos25Dispositivo(ctx context.Context, idDispositivo string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+eventoSelectCols+`
FROM vis_evento e
WHERE e.id_dispositivo = $1
ORDER BY e.created_at DESC
LIMIT $2`, idDispositivo, limit)
	if err != nil {
		return nil, err
	}
	return scanEventoRows(rows)
}

package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func ListCamerasByFranqueado(ctx context.Context, idFranqueado string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+cameraSelectCols+`
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.id_franqueado = $1
ORDER BY c.created_at DESC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	return scanCameraRows(rows)
}

func ListAllCameras(ctx context.Context) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+cameraSelectCols+`
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
ORDER BY c.id ASC`)
	if err != nil {
		return nil, err
	}
	return scanCameraRows(rows)
}

func ListCamerasByCliente(ctx context.Context, idCliente string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+cameraSelectCols+`
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.id_cliente = $1
ORDER BY c.created_at DESC`, idCliente)
	if err != nil {
		return nil, err
	}
	return scanCameraRows(rows)
}

func GetCameraByID(ctx context.Context, id int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT `+cameraSelectCols+`
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.id = $1`, id)
	if err != nil {
		return nil, err
	}
	list, err := scanCameraRows(rows)
	if err != nil || len(list) == 0 {
		return nil, fmt.Errorf("camera nao encontrada")
	}
	return list[0], nil
}

func ListCamerasAnaliticas(ctx context.Context, workerID string, nodeID int) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	query := `
SELECT ` + cameraSelectCols + `
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.ativo = TRUE
  AND c.deteccao_humano = TRUE
  AND (c.analitico_pausado IS NOT TRUE)`
	args := []any{}
	n := 1
	if workerID != "" {
		query += fmt.Sprintf(" AND NULLIF(BTRIM(c.worker_id), '') = $%d", n)
		args = append(args, workerID)
		n++
	}
	if nodeID > 0 {
		query += fmt.Sprintf(" AND c.vis_mediamtx_node_id = $%d", n)
		args = append(args, nodeID)
	}
	query += " ORDER BY c.id ASC"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanCameraRows(rows)
}

func ListGravacaoCameras(ctx context.Context, workerID string, nodeID int) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	query := `
SELECT ` + cameraSelectCols + `
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE (c.grava_continua = TRUE OR c.grava_movimento = TRUE OR c.grava_timelapse = TRUE)
  AND (c.bloqueado IS NOT TRUE)`
	args := []any{}
	n := 1
	if workerID != "" {
		query += fmt.Sprintf(" AND c.worker_id = $%d", n)
		args = append(args, workerID)
		n++
	}
	if nodeID > 0 {
		query += fmt.Sprintf(" AND c.vis_mediamtx_node_id = $%d", n)
		args = append(args, nodeID)
	}
	query += " ORDER BY c.id ASC"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanCameraRows(rows)
}

func CreateCamera(ctx context.Context, input map[string]any) (map[string]any, error) {
	out, err := createCameraOnce(ctx, input)
	if err != nil && isBadConn(err) {
		resetDB()
		out, err = createCameraOnce(ctx, input)
	}
	if err != nil {
		return nil, humanizeDBErr(err)
	}
	return out, nil
}

func createCameraOnce(ctx context.Context, input map[string]any) (map[string]any, error) {
	licID := intVal(input, "vis_licenca_id")
	idFra := strVal(input, "id_franqueado")
	if licID <= 0 {
		return nil, fmt.Errorf("Selecione uma licenca disponivel para cadastrar a camera")
	}
	if idFra == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
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

	var lic struct {
		ID           int
		Plano        sql.NullString
		Unidade      sql.NullString
		Status       sql.NullString
		IDFranqueado sql.NullString
		ValidoAte    sql.NullTime
	}
	err = tx.QueryRowContext(ctx, `
SELECT id, plano, unidade, status, id_franqueado, valido_ate
FROM vis_licenca WHERE id = $1 FOR UPDATE`, licID).Scan(
		&lic.ID, &lic.Plano, &lic.Unidade, &lic.Status, &lic.IDFranqueado, &lic.ValidoAte)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Licenca nao encontrada")
	}
	if err != nil {
		return nil, err
	}
	if !lic.Status.Valid || lic.Status.String != "disponivel" {
		return nil, fmt.Errorf("Licenca indisponivel ou ja em uso")
	}
	if !lic.Plano.Valid || lic.Plano.String == "" {
		return nil, fmt.Errorf("Licenca sem plano definido")
	}
	if lic.IDFranqueado.String != idFra {
		return nil, fmt.Errorf("Licenca nao pertence ao franqueado")
	}
	unidade := lic.Unidade.String
	if unidade == "" {
		unidade = "camera"
	}
	if unidade != "camera" {
		return nil, fmt.Errorf("Licenca informada nao e de camera — use licenca de plano online/sensor/analitico")
	}
	if lic.ValidoAte.Valid && lic.ValidoAte.Time.Before(time.Now()) {
		return nil, fmt.Errorf("Licenca expirada")
	}

	if err := CheckCapacidadeDisponivel(ctx, idFra, true); err != nil {
		return nil, err
	}

	flags := PlanoFlagsFrom(lic.Plano.String)
	ativo := true
	if v := boolVal(input, "ativo"); v != nil {
		ativo = *v
	}
	detHumano := false
	if v := boolVal(input, "deteccao_humano"); v != nil {
		detHumano = *v
	}
	detVeiculo := false
	if v := boolVal(input, "deteccao_veiculo"); v != nil {
		detVeiculo = *v
	}

	if lic.Plano.String == "online" {
		ativo, detHumano, detVeiculo = false, false, false
	} else if flags.SemAtivo {
		ativo = false
	}
	if flags.CapturaAnalitico {
		detHumano = true
	}

	var ativadoEm any
	if ativo {
		ativadoEm = time.Now().UTC()
	}

	node, err := pickNodeTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	analiticoPausado := false
	if v := boolVal(input, "analitico_pausado"); v != nil {
		analiticoPausado = *v
	}

	var newID int
	err = tx.QueryRowContext(ctx, `
INSERT INTO vis_camera (
    ativo, bloqueado, nome, id_franqueado, id_cliente, id_dispositivo,
    conta, particao, canal, setor, protocolo, rtsp_url_sec,
    onvif_host, onvif_porta, onvif_usuario, onvif_senha, confianca_min, cooldown_seg,
    modo_deteccao, somente_armado, deteccao_humano, deteccao_veiculo, status,
    worker_id, zonauser, captura_sensor, captura_analitico, analitico_pausado,
    evento_grava_foto, evento_grava_video, id_setor, snapshot_url,
    vis_licenca_id, plano, ativado_em, vis_mediamtx_node_id
) VALUES (
    $1,false,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
    $19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35
) RETURNING id`,
		ativo, strVal(input, "nome"), idFra, strVal(input, "id_cliente"), strVal(input, "id_dispositivo"),
		strVal(input, "conta"), strVal(input, "particao"), strVal(input, "canal"), strVal(input, "setor"),
		strVal(input, "protocolo"), strVal(input, "rtsp_url_sec"),
		strVal(input, "onvif_host"), intVal(input, "onvif_porta"), strVal(input, "onvif_usuario"),
		strVal(input, "onvif_senha"), floatVal(input, "confianca_min"), intVal(input, "cooldown_seg"),
		strVal(input, "modo_deteccao"),
		flags.SomenteArmado, detHumano, detVeiculo, strVal(input, "status"),
		strVal(input, "worker_id"), strVal(input, "zonauser"),
		flags.CapturaSensor, flags.CapturaAnalitico, analiticoPausado,
		flags.EventoGravaFoto, flags.EventoGravaVideo,
		strVal(input, "id_setor"), strVal(input, "snapshot_url"),
		licID, lic.Plano.String, ativadoEm, node.ID,
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
UPDATE vis_licenca SET status = 'em_uso', vis_camera_id = $2, id_dispositivo = $3, plano = $4
WHERE id = $1`, licID, newID, strVal(input, "id_dispositivo"), lic.Plano.String)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = SyncMediamtxNode(ctx, node.ID)
	return GetCameraByID(ctx, newID)
}

func pickNodeTx(ctx context.Context, tx *sql.Tx) (*MediamtxNode, error) {
	// Uma unica query: pgx nao permite QueryRow com rows abertos na mesma Tx.
	rows, err := tx.QueryContext(ctx, `
SELECT n.id, COALESCE(n.max_cameras, 200), COUNT(c.id)::int
FROM vis_mediamtx_node n
LEFT JOIN vis_camera c ON c.vis_mediamtx_node_id = n.id
WHERE n.status = 'ativo'
GROUP BY n.id, n.max_cameras, n.ordem
ORDER BY n.ordem ASC, n.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *MediamtxNode
	bestCount := 999999
	for rows.Next() {
		var n MediamtxNode
		var total int
		if err := rows.Scan(&n.ID, &n.MaxCameras, &total); err != nil {
			return nil, err
		}
		if total < n.MaxCameras && total < bestCount {
			c := n
			best = &c
			bestCount = total
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if best == nil {
		return nil, fmt.Errorf("Todos os nos MediaMTX estao cheios — cadastre um novo servidor ou aumente max_cameras")
	}
	return best, nil
}

func UpdateCamera(ctx context.Context, id int, input map[string]any) (map[string]any, error) {
	out, err := updateCameraOnce(ctx, id, input)
	if err != nil && isBadConn(err) {
		resetDB()
		out, err = updateCameraOnce(ctx, id, input)
	}
	if err != nil {
		return nil, humanizeDBErr(err)
	}
	return out, nil
}

func updateCameraOnce(ctx context.Context, id int, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var curLicID sql.NullInt64
	var idFra, idDisp sql.NullString
	var nodeID sql.NullInt64
	var ativadoEm sql.NullTime
	var curAtivo sql.NullBool
	err = tx.QueryRowContext(ctx, `
SELECT vis_licenca_id, id_franqueado, id_dispositivo, vis_mediamtx_node_id, ativado_em, ativo
FROM vis_camera WHERE id = $1 FOR UPDATE`, id).Scan(&curLicID, &idFra, &idDisp, &nodeID, &ativadoEm, &curAtivo)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("camera nao encontrada")
	}
	if err != nil {
		return nil, err
	}

	idFranqueado := strVal(input, "id_franqueado")
	if idFranqueado == "" && idFra.Valid {
		idFranqueado = idFra.String
	}
	idDispositivo := strVal(input, "id_dispositivo")
	if idDispositivo == "" && idDisp.Valid {
		idDispositivo = idDisp.String
	}

	curLic := 0
	if curLicID.Valid {
		curLic = int(curLicID.Int64)
	}

	newLicID := 0
	if _, ok := input["vis_licenca_id"]; ok {
		newLicID = intVal(input, "vis_licenca_id")
	}

	if newLicID > 0 && newLicID != curLic {
		plano, err := trocarLicencaCameraTx(ctx, tx, id, curLic, newLicID, idFranqueado, idDispositivo, input, ativadoEm)
		if err != nil {
			return nil, err
		}
		input["plano"] = plano
		flags := PlanoFlagsFrom(plano)
		input["captura_sensor"] = flags.CapturaSensor
		input["captura_analitico"] = flags.CapturaAnalitico
		input["somente_armado"] = flags.SomenteArmado
		input["evento_grava_foto"] = flags.EventoGravaFoto
		input["evento_grava_video"] = flags.EventoGravaVideo
		ativo, detHumano, detVeiculo, newAtivado := resolveCameraAtivoDeteccao(plano, flags, input, ativadoEm)
		input["ativo"] = ativo
		input["deteccao_humano"] = detHumano
		input["deteccao_veiculo"] = detVeiculo
		if newAtivado != nil {
			input["ativado_em"] = newAtivado.UTC().Format(time.RFC3339)
		} else if !ativo {
			input["ativado_em"] = nil
		}
	}

	sets := []string{}
	args := []any{}
	n := 1
	for _, col := range []string{
		"nome", "id_cliente", "id_dispositivo", "conta", "particao", "canal", "setor",
		"protocolo", "rtsp_url_sec", "onvif_host", "onvif_usuario", "onvif_senha",
		"status", "worker_id", "zonauser", "id_setor", "snapshot_url", "modo_deteccao", "plano",
	} {
		if v, ok := input[col]; ok {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, trimAny(v))
			n++
		}
	}
	for _, col := range []string{"vis_licenca_id", "onvif_porta", "confianca_min", "cooldown_seg"} {
		if v, ok := input[col]; ok {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, intVal(map[string]any{col: v}, col))
			n++
		}
	}
	for _, col := range []string{
		"ativo", "deteccao_humano", "deteccao_veiculo", "analitico_pausado", "bloqueado",
		"captura_sensor", "captura_analitico", "somente_armado", "evento_grava_foto", "evento_grava_video",
	} {
		if v := boolVal(input, col); v != nil {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, *v)
			n++
		}
	}
	if v, ok := input["ativado_em"]; ok {
		if v == nil || trimAny(v) == "" {
			sets = append(sets, "ativado_em = NULL")
		} else {
			sets = append(sets, fmt.Sprintf("ativado_em = $%d", n))
			args = append(args, trimAny(v))
			n++
		}
	}

	if len(sets) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return GetCameraByID(ctx, id)
	}

	prevAtivo := curAtivo.Valid && curAtivo.Bool
	newAtivo := prevAtivo
	if v := boolVal(input, "ativo"); v != nil {
		newAtivo = *v
	}
	if err := bumpStreamPolicyGeneration(ctx, tx, id, prevAtivo, newAtivo); err != nil {
		return nil, err
	}

	args = append(args, id)
	q := fmt.Sprintf("UPDATE vis_camera SET %s WHERE id = $%d", strings.Join(sets, ", "), n)
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if nodeID.Valid {
		_ = SyncMediamtxNode(ctx, int(nodeID.Int64))
	}
	return GetCameraByID(ctx, id)
}

func trocarLicencaCameraTx(
	ctx context.Context,
	tx *sql.Tx,
	cameraID, licAntigaID, licNovaID int,
	idFranqueado, idDispositivo string,
	input map[string]any,
	ativadoEmAtual sql.NullTime,
) (string, error) {
	plano, err := reservarLicencaCameraTx(ctx, tx, licNovaID, idFranqueado)
	if err != nil {
		return "", err
	}
	if licAntigaID > 0 {
		if err := liberarLicencaRegistroTx(ctx, tx, licAntigaID); err != nil {
			return "", err
		}
	}
	if err := vincularLicencaCameraTx(ctx, tx, licNovaID, cameraID, idDispositivo, plano); err != nil {
		return "", err
	}
	_ = input
	_ = ativadoEmAtual
	return plano, nil
}

func reservarLicencaCameraTx(ctx context.Context, tx *sql.Tx, licID int, idFranqueado string) (string, error) {
	var lic struct {
		Plano        sql.NullString
		Unidade      sql.NullString
		Status       sql.NullString
		IDFranqueado sql.NullString
		ValidoAte    sql.NullTime
	}
	err := tx.QueryRowContext(ctx, `
SELECT plano, unidade, status, id_franqueado, valido_ate
FROM vis_licenca WHERE id = $1 FOR UPDATE`, licID).Scan(
		&lic.Plano, &lic.Unidade, &lic.Status, &lic.IDFranqueado, &lic.ValidoAte)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("Licenca selecionada nao encontrada")
	}
	if err != nil {
		return "", err
	}
	if !lic.Status.Valid || lic.Status.String != "disponivel" {
		return "", fmt.Errorf("Licenca indisponivel ou ja em uso")
	}
	if !lic.Plano.Valid || lic.Plano.String == "" {
		return "", fmt.Errorf("Licenca selecionada sem plano definido")
	}
	if idFranqueado != "" && lic.IDFranqueado.String != idFranqueado {
		return "", fmt.Errorf("Licenca nao pertence ao franqueado")
	}
	unidade := lic.Unidade.String
	if unidade == "" {
		unidade = "camera"
	}
	if unidade != "camera" {
		return "", fmt.Errorf("Licenca selecionada nao e de camera")
	}
	if lic.ValidoAte.Valid && lic.ValidoAte.Time.Before(time.Now()) {
		return "", fmt.Errorf("Licenca expirada")
	}
	return lic.Plano.String, nil
}

func liberarLicencaRegistroTx(ctx context.Context, tx *sql.Tx, licID int) error {
	var plano sql.NullString
	var valido sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT plano, valido_ate FROM vis_licenca WHERE id = $1 FOR UPDATE`, licID).
		Scan(&plano, &valido)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	status := "disponivel"
	if valido.Valid && valido.Time.Before(time.Now()) {
		status = "expirada"
	}
	_, err = tx.ExecContext(ctx, `
UPDATE vis_licenca SET status = $2, vis_camera_id = NULL, id_dispositivo = NULL, plano = $3
WHERE id = $1`, licID, status, nullStr(plano))
	return err
}

func vincularLicencaCameraTx(ctx context.Context, tx *sql.Tx, licID, cameraID int, idDispositivo, plano string) error {
	_, err := tx.ExecContext(ctx, `
UPDATE vis_licenca SET status = 'em_uso', vis_camera_id = $2, id_dispositivo = $3, plano = $4
WHERE id = $1`, licID, cameraID, idDispositivo, plano)
	return err
}

func resolveCameraAtivoDeteccao(
	plano string,
	flags PlanoFlags,
	input map[string]any,
	ativadoEmAtual sql.NullTime,
) (ativo, detHumano, detVeiculo bool, newAtivado *time.Time) {
	ativo = true
	if v := boolVal(input, "ativo"); v != nil {
		ativo = *v
	}
	detHumano = false
	if v := boolVal(input, "deteccao_humano"); v != nil {
		detHumano = *v
	}
	detVeiculo = false
	if v := boolVal(input, "deteccao_veiculo"); v != nil {
		detVeiculo = *v
	}

	if plano == "online" {
		ativo, detHumano, detVeiculo = false, false, false
	} else if flags.SemAtivo {
		ativo = false
	}
	if flags.CapturaAnalitico {
		detHumano = true
	}

	if ativo {
		if ativadoEmAtual.Valid {
			t := ativadoEmAtual.Time
			newAtivado = &t
		} else {
			t := time.Now().UTC()
			newAtivado = &t
		}
	}
	return ativo, detHumano, detVeiculo, newAtivado
}
func UpdateCameraSnapshot(ctx context.Context, id int, snapshotURL string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET snapshot_url = $2 WHERE id = $1`, id, snapshotURL)
	return err
}

func SetCameraBloqueado(ctx context.Context, id int, bloqueado bool) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET bloqueado = $2 WHERE id = $1`, id, bloqueado)
	if err != nil {
		return nil, err
	}
	return GetCameraByID(ctx, id)
}

func SetAnaliticoPausado(ctx context.Context, id int, pausado bool) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET analitico_pausado = $2 WHERE id = $1`, id, pausado)
	if err != nil {
		return nil, err
	}
	return GetCameraByID(ctx, id)
}

func SoftDeleteCamera(ctx context.Context, id int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET ativo = false WHERE id = $1`, id)
	return err
}

func LiberarLicencaCamera(ctx context.Context, id int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var licID sql.NullInt64
	var nodeID sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
SELECT vis_licenca_id, vis_mediamtx_node_id FROM vis_camera WHERE id = $1 FOR UPDATE`, id).
		Scan(&licID, &nodeID); err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
UPDATE vis_camera SET ativo = false, deteccao_humano = false, deteccao_veiculo = false,
    captura_sensor = false, captura_analitico = false, somente_armado = false,
    evento_grava_foto = false, evento_grava_video = false,
    vis_licenca_id = NULL, plano = NULL,
    grava_continua = false, grava_movimento = false, grava_timelapse = false,
    vis_licenca_gravacao_id = NULL, gravacao_status = NULL
WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}

	if licID.Valid {
		_, _ = tx.ExecContext(ctx, `
UPDATE vis_licenca SET status = 'disponivel', vis_camera_id = NULL WHERE id = $1`, licID.Int64)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if nodeID.Valid {
		_ = SyncMediamtxNode(ctx, int(nodeID.Int64))
	}
	return GetCameraByID(ctx, id)
}

func AckGravacaoFlush(ctx context.Context, id int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET gravacao_flush_pedido = false WHERE id = $1`, id)
	return err
}

func RequestGravacaoFlush(ctx context.Context, id int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET gravacao_flush_pedido = true WHERE id = $1`, id)
	return err
}

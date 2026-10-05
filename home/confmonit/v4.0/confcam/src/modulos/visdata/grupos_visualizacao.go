package visdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func ListGruposByFranqueado(ctx context.Context, idFranqueado string) ([]map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT g.id, g.created_at, g.updated_at, g.id_franqueado, g.nome, g.descricao, g.ativo, g.layout_mosaic,
       COUNT(gc.id) FILTER (WHERE gc.ativo IS TRUE) AS total_cameras
FROM vis_grupo_visualizacao g
LEFT JOIN vis_grupo_visualizacao_camera gc ON gc.grupo_id = g.id
WHERE g.id_franqueado = $1
GROUP BY g.id
ORDER BY g.nome ASC, g.id ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGrupoRows(rows)
}

func ListGruposDisponiveisCliente(ctx context.Context, idFranqueado, idCliente string) ([]map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idFranqueado == "" || idCliente == "" {
		return nil, fmt.Errorf("id_franqueado e id_cliente obrigatorios")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT g.id, g.created_at, g.updated_at, g.id_franqueado, g.nome, g.descricao, g.ativo, g.layout_mosaic,
       COUNT(gc.id) FILTER (WHERE gc.ativo IS TRUE AND gc.id_cliente = $2) AS total_cameras
FROM vis_grupo_visualizacao g
INNER JOIN vis_grupo_visualizacao_cliente gcl ON gcl.grupo_id = g.id AND gcl.id_cliente = $2
LEFT JOIN vis_grupo_visualizacao_camera gc ON gc.grupo_id = g.id
WHERE g.id_franqueado = $1 AND g.ativo IS TRUE
GROUP BY g.id
ORDER BY g.nome ASC, g.id ASC`, idFranqueado, idCliente)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGrupoRows(rows)
}

func GetGrupoVisualizacao(ctx context.Context, grupoID int, idFranqueado, idClienteFilter string) (map[string]any, error) {
	if grupoID < 1 {
		return nil, fmt.Errorf("grupo invalido")
	}
	idClienteFilter = strings.TrimSpace(idClienteFilter)
	db, err := DB()
	if err != nil {
		return nil, err
	}
	camCountFilter := ""
	args := []any{grupoID}
	if idClienteFilter != "" {
		camCountFilter = ` AND gc.id_cliente = $2`
		args = append(args, idClienteFilter)
	}
	row := db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT g.id, g.created_at, g.updated_at, g.id_franqueado, g.nome, g.descricao, g.ativo, g.layout_mosaic,
       (SELECT COUNT(*) FROM vis_grupo_visualizacao_camera gc WHERE gc.grupo_id = g.id AND gc.ativo IS TRUE%s) AS total_cameras
FROM vis_grupo_visualizacao g
WHERE g.id = $1`, camCountFilter), args...)
	grupo, err := scanGrupoRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("grupo nao encontrado")
	}
	if err != nil {
		return nil, err
	}
	idFraGrupo := strings.TrimSpace(fmt.Sprint(grupo["id_franqueado"]))
	if idFranqueado != "" && idFraGrupo != idFranqueado {
		return nil, fmt.Errorf("grupo nao autorizado")
	}
	if idClienteFilter != "" {
		if err := ClienteTemAcessoGrupo(ctx, grupoID, idFranqueado, idClienteFilter); err != nil {
			return nil, err
		}
	}
	clientes, err := listGrupoClientes(ctx, db, grupoID)
	if err != nil {
		return nil, err
	}
	if idClienteFilter != "" {
		clientes = []string{idClienteFilter}
	}
	cameras, err := listGrupoCamerasInternal(ctx, db, grupoID, idClienteFilter)
	if err != nil {
		return nil, err
	}
	grupo["clientes"] = clientes
	grupo["cameras"] = cameras
	return grupo, nil
}

func CreateGrupoVisualizacao(ctx context.Context, payload map[string]any) (map[string]any, error) {
	idFra := strings.TrimSpace(strVal(payload, "id_franqueado"))
	nome := strings.TrimSpace(strVal(payload, "nome"))
	if idFra == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	if nome == "" {
		return nil, fmt.Errorf("nome obrigatorio")
	}
	descricao := strings.TrimSpace(strVal(payload, "descricao"))
	ativo := true
	if b := boolVal(payload, "ativo"); b != nil {
		ativo = *b
	}
	layout := defaultLayoutMosaic()
	if raw, ok := payload["layout_mosaic"]; ok && raw != nil {
		if b, err := json.Marshal(raw); err == nil {
			layout = b
		}
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_grupo_visualizacao (id_franqueado, nome, descricao, ativo, layout_mosaic)
VALUES ($1, $2, $3, $4, $5::jsonb)
RETURNING id`, idFra, nome, nullIfEmpty(descricao), ativo, string(layout)).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetGrupoVisualizacao(ctx, id, idFra, "")
}

func UpdateGrupoVisualizacao(ctx context.Context, grupoID int, payload map[string]any) (map[string]any, error) {
	idFra := strings.TrimSpace(strVal(payload, "id_franqueado"))
	if idFra == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	if _, err := assertGrupoFranqueado(ctx, grupoID, idFra); err != nil {
		return nil, err
	}
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	n := 1
	if v := strings.TrimSpace(strVal(payload, "nome")); v != "" {
		sets = append(sets, fmt.Sprintf("nome = $%d", n))
		args = append(args, v)
		n++
	}
	if payload["descricao"] != nil {
		sets = append(sets, fmt.Sprintf("descricao = $%d", n))
		args = append(args, nullIfEmpty(strings.TrimSpace(strVal(payload, "descricao"))))
		n++
	}
	if b := boolVal(payload, "ativo"); b != nil {
		sets = append(sets, fmt.Sprintf("ativo = $%d", n))
		args = append(args, *b)
		n++
	}
	if raw, ok := payload["layout_mosaic"]; ok && raw != nil {
		if b, err := json.Marshal(raw); err == nil {
			sets = append(sets, fmt.Sprintf("layout_mosaic = $%d::jsonb", n))
			args = append(args, string(b))
			n++
		}
	}
	if len(sets) == 1 {
		return GetGrupoVisualizacao(ctx, grupoID, idFra, "")
	}
	args = append(args, grupoID)
	db, err := DB()
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf("UPDATE vis_grupo_visualizacao SET %s WHERE id = $%d", strings.Join(sets, ", "), n)
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return GetGrupoVisualizacao(ctx, grupoID, idFra, "")
}

func DeleteGrupoVisualizacao(ctx context.Context, grupoID int, idFranqueado string) error {
	if _, err := assertGrupoFranqueado(ctx, grupoID, idFranqueado); err != nil {
		return err
	}
	db, err := DB()
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, `DELETE FROM vis_grupo_visualizacao WHERE id = $1`, grupoID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("grupo nao encontrado")
	}
	return nil
}

func SaveGrupoComposicao(ctx context.Context, grupoID int, payload map[string]any) (map[string]any, error) {
	idFra := strings.TrimSpace(strVal(payload, "id_franqueado"))
	if idFra == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	if _, err := assertGrupoFranqueado(ctx, grupoID, idFra); err != nil {
		return nil, err
	}
	clientes := parseStringList(payload["clientes"])
	cameras := parseCameraComposicao(payload["cameras"])
	if err := validateCamerasComposicao(ctx, idFra, cameras); err != nil {
		return nil, err
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
	if _, err := tx.ExecContext(ctx, `UPDATE vis_grupo_visualizacao SET updated_at = NOW() WHERE id = $1`, grupoID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM vis_grupo_visualizacao_cliente WHERE grupo_id = $1`, grupoID); err != nil {
		return nil, err
	}
	for _, idCli := range clientes {
		idCli = strings.TrimSpace(idCli)
		if idCli == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO vis_grupo_visualizacao_cliente (grupo_id, id_cliente) VALUES ($1, $2)`,
			grupoID, idCli); err != nil {
			return nil, err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM vis_grupo_visualizacao_camera WHERE grupo_id = $1`, grupoID); err != nil {
		return nil, err
	}
	for _, cam := range cameras {
		ativo := true
		if cam.ativo != nil {
			ativo = *cam.ativo
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO vis_grupo_visualizacao_camera (grupo_id, id_cliente, vis_camera_id, ordem, ativo)
VALUES ($1, $2, $3, $4, $5)`,
			grupoID, cam.idCliente, cam.cameraID, cam.ordem, ativo); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return GetGrupoVisualizacao(ctx, grupoID, idFra, "")
}

func ReordenarGrupoCameras(ctx context.Context, grupoID int, payload map[string]any) error {
	idFra := strings.TrimSpace(strVal(payload, "id_franqueado"))
	if idFra == "" {
		return fmt.Errorf("id_franqueado obrigatorio")
	}
	if _, err := assertGrupoFranqueado(ctx, grupoID, idFra); err != nil {
		return err
	}
	itens, ok := payload["itens"].([]any)
	if !ok || len(itens) == 0 {
		return fmt.Errorf("itens obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, raw := range itens {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		slotID := intVal(m, "id")
		ordem := intVal(m, "ordem")
		if slotID < 1 {
			continue
		}
		res, err := tx.ExecContext(ctx, `
UPDATE vis_grupo_visualizacao_camera SET ordem = $1
WHERE id = $2 AND grupo_id = $3`, ordem, slotID, grupoID)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("camera do grupo invalida")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE vis_grupo_visualizacao SET updated_at = NOW() WHERE id = $1`, grupoID); err != nil {
		return err
	}
	return tx.Commit()
}

func ListGrupoCameras(ctx context.Context, grupoID int, idFranqueado, idClienteFilter string) ([]map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idClienteFilter = strings.TrimSpace(idClienteFilter)
	if grupoID < 1 {
		return nil, fmt.Errorf("grupo invalido")
	}
	if _, err := assertGrupoFranqueado(ctx, grupoID, idFranqueado); err != nil {
		return nil, err
	}
	if idClienteFilter != "" {
		if err := ClienteTemAcessoGrupo(ctx, grupoID, idFranqueado, idClienteFilter); err != nil {
			return nil, err
		}
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return listGrupoCamerasInternal(ctx, db, grupoID, idClienteFilter)
}

func assertGrupoFranqueado(ctx context.Context, grupoID int, idFranqueado string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `
SELECT id, id_franqueado FROM vis_grupo_visualizacao WHERE id = $1`, grupoID)
	var id int
	var idFra string
	if err := row.Scan(&id, &idFra); err == sql.ErrNoRows {
		return nil, fmt.Errorf("grupo nao encontrado")
	} else if err != nil {
		return nil, err
	}
	if strings.TrimSpace(idFra) != strings.TrimSpace(idFranqueado) {
		return nil, fmt.Errorf("grupo nao autorizado")
	}
	return map[string]any{"id": id, "id_franqueado": idFra}, nil
}

func assertClienteNoGrupo(ctx context.Context, grupoID int, idCliente string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_grupo_visualizacao_cliente WHERE grupo_id = $1 AND id_cliente = $2`,
		grupoID, idCliente).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("grupo nao autorizado para cliente")
	}
	return nil
}

func ClienteTemAcessoGrupo(ctx context.Context, grupoID int, idFranqueado, idCliente string) error {
	if _, err := assertGrupoFranqueado(ctx, grupoID, idFranqueado); err != nil {
		return err
	}
	if err := assertClienteNoGrupo(ctx, grupoID, idCliente); err != nil {
		return err
	}
	db, err := DB()
	if err != nil {
		return err
	}
	var ativo bool
	err = db.QueryRowContext(ctx, `SELECT ativo FROM vis_grupo_visualizacao WHERE id = $1`, grupoID).Scan(&ativo)
	if err == sql.ErrNoRows {
		return fmt.Errorf("grupo nao encontrado")
	}
	if err != nil {
		return err
	}
	if !ativo {
		return fmt.Errorf("grupo inativo")
	}
	return nil
}

type cameraComposicao struct {
	cameraID  int
	idCliente string
	ordem     int
	ativo     *bool
}

func parseCameraComposicao(raw any) []cameraComposicao {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := []cameraComposicao{}
	for i, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		camID := intVal(m, "vis_camera_id")
		if camID == 0 {
			camID = intVal(m, "camera_id")
		}
		if camID < 1 {
			continue
		}
		ordem := intVal(m, "ordem")
		if ordem == 0 {
			ordem = i + 1
		}
		out = append(out, cameraComposicao{
			cameraID:  camID,
			idCliente: strings.TrimSpace(strVal(m, "id_cliente")),
			ordem:     ordem,
			ativo:     boolVal(m, "ativo"),
		})
	}
	return out
}

func validateCamerasComposicao(ctx context.Context, idFranqueado string, cameras []cameraComposicao) error {
	for _, cam := range cameras {
		if cam.idCliente == "" {
			return fmt.Errorf("id_cliente obrigatorio para camera %d", cam.cameraID)
		}
		row, err := GetCameraByID(ctx, cam.cameraID)
		if err != nil {
			return fmt.Errorf("camera %d nao encontrada", cam.cameraID)
		}
		idFra := strings.TrimSpace(fmt.Sprint(row["id_franqueado"]))
		idCli := strings.TrimSpace(fmt.Sprint(row["id_cliente"]))
		if idFra != idFranqueado {
			return fmt.Errorf("camera %d nao pertence ao franqueado", cam.cameraID)
		}
		if idCli != cam.idCliente {
			return fmt.Errorf("camera %d nao pertence ao cliente informado", cam.cameraID)
		}
	}
	return nil
}

func listGrupoClientes(ctx context.Context, db *sql.DB, grupoID int) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id_cliente FROM vis_grupo_visualizacao_cliente WHERE grupo_id = $1 ORDER BY id_cliente`, grupoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func listGrupoCamerasInternal(ctx context.Context, db *sql.DB, grupoID int, idClienteFilter string) ([]map[string]any, error) {
	q := `
SELECT gc.id, gc.grupo_id, gc.id_cliente, gc.vis_camera_id, gc.ordem, gc.ativo, gc.created_at,
       c.nome, c.ativo AS camera_ativo, c.bloqueado, c.plano, c.ultimo_ping_em, c.protocolo
FROM vis_grupo_visualizacao_camera gc
INNER JOIN vis_camera c ON c.id = gc.vis_camera_id
WHERE gc.grupo_id = $1 AND gc.ativo IS TRUE`
	args := []any{grupoID}
	if strings.TrimSpace(idClienteFilter) != "" {
		q += ` AND gc.id_cliente = $2`
		args = append(args, idClienteFilter)
	}
	q += ` ORDER BY gc.ordem ASC, gc.id ASC`
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var slotID, gID, camID, ordem int
		var idCliente string
		var ativo, camAtivo, bloqueado sql.NullBool
		var createdAt sql.NullTime
		var nome, plano, protocolo sql.NullString
		var ultimoPing sql.NullTime
		if err := rows.Scan(&slotID, &gID, &idCliente, &camID, &ordem, &ativo, &createdAt,
			&nome, &camAtivo, &bloqueado, &plano, &ultimoPing, &protocolo); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":              slotID,
			"grupo_id":        gID,
			"id_cliente":      idCliente,
			"vis_camera_id":   camID,
			"camera_id":       camID,
			"ordem":           ordem,
			"ativo":           nullBool(ativo),
			"created_at":      nullTime(createdAt),
			"camera_nome":     nullStr(nome),
			"nome":            nullStr(nome),
			"camera_ativo":    nullBool(camAtivo),
			"bloqueado":       nullBool(bloqueado),
			"plano":           nullStr(plano),
			"ultimo_ping_em":  nullTime(ultimoPing),
			"protocolo":       nullStr(protocolo),
		})
	}
	return out, rows.Err()
}

func scanGrupoRows(rows *sql.Rows) ([]map[string]any, error) {
	out := []map[string]any{}
	for rows.Next() {
		item, err := scanGrupoRowFromRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanGrupoRow(row *sql.Row) (map[string]any, error) {
	var id, totalCameras int
	var createdAt, updatedAt time.Time
	var idFra, nome string
	var descricao sql.NullString
	var ativo bool
	var layout []byte
	if err := row.Scan(&id, &createdAt, &updatedAt, &idFra, &nome, &descricao, &ativo, &layout, &totalCameras); err != nil {
		return nil, err
	}
	return mapGrupo(id, createdAt, updatedAt, idFra, nome, descricao, ativo, layout, totalCameras), nil
}

func scanGrupoRowFromRows(rows *sql.Rows) (map[string]any, error) {
	var id, totalCameras int
	var createdAt, updatedAt time.Time
	var idFra, nome string
	var descricao sql.NullString
	var ativo bool
	var layout []byte
	if err := rows.Scan(&id, &createdAt, &updatedAt, &idFra, &nome, &descricao, &ativo, &layout, &totalCameras); err != nil {
		return nil, err
	}
	return mapGrupo(id, createdAt, updatedAt, idFra, nome, descricao, ativo, layout, totalCameras), nil
}

func mapGrupo(id int, createdAt, updatedAt time.Time, idFra, nome string, descricao sql.NullString, ativo bool, layout []byte, totalCameras int) map[string]any {
	layoutObj := any(map[string]any{"modo": "auto"})
	if len(layout) > 0 {
		var parsed any
		if json.Unmarshal(layout, &parsed) == nil {
			layoutObj = parsed
		}
	}
	return map[string]any{
		"id":             id,
		"created_at":     createdAt.UTC().Format(time.RFC3339),
		"updated_at":     updatedAt.UTC().Format(time.RFC3339),
		"id_franqueado":  idFra,
		"nome":           nome,
		"descricao":      nullStr(descricao),
		"ativo":          ativo,
		"layout_mosaic":  layoutObj,
		"total_cameras":  totalCameras,
	}
}

func defaultLayoutMosaic() []byte {
	return []byte(`{"modo":"auto"}`)
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func parseStringList(raw any) []string {
	switch t := raw.(type) {
	case []any:
		out := []string{}
		for _, v := range t {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s != "" && s != "<nil>" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		out := []string{}
		for _, s := range t {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

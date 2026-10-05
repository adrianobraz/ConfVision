package visdata

import (
	"context"
	"database/sql"
	"fmt"
)

func ListAreasByCamera(ctx context.Context, cameraID int) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, vis_camera_id, nome, ativo, poligono_json, cor
FROM vis_camera_area WHERE vis_camera_id = $1 ORDER BY id ASC`, cameraID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, camID int
		var created sql.NullTime
		var nome, poligono, cor sql.NullString
		var ativo sql.NullBool
		if err := rows.Scan(&id, &created, &camID, &nome, &ativo, &poligono, &cor); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":             id,
			"created_at":     nullTime(created),
			"vis_camera_id":  camID,
			"nome":           nullStr(nome),
			"ativo":          nullBool(ativo),
			"poligono_json":  nullStr(poligono),
			"cor":            nullStr(cor),
		})
	}
	return out, rows.Err()
}

func ListAreasAtivas(ctx context.Context) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, vis_camera_id, nome, ativo, poligono_json, cor
FROM vis_camera_area WHERE ativo = TRUE ORDER BY vis_camera_id, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, camID int
		var created sql.NullTime
		var nome, poligono, cor sql.NullString
		var ativo sql.NullBool
		if err := rows.Scan(&id, &created, &camID, &nome, &ativo, &poligono, &cor); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":             id,
			"created_at":     nullTime(created),
			"vis_camera_id":  camID,
			"nome":           nullStr(nome),
			"ativo":          nullBool(ativo),
			"poligono_json":  nullStr(poligono),
			"cor":            nullStr(cor),
		})
	}
	return out, rows.Err()
}

func CreateArea(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	camID := intVal(input, "vis_camera_id")
	ativo := true
	if v := boolVal(input, "ativo"); v != nil {
		ativo = *v
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_camera_area (vis_camera_id, nome, ativo, poligono_json, cor)
VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		camID, strVal(input, "nome"), ativo, strVal(input, "poligono_json"), strVal(input, "cor"),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	list, err := ListAreasByCamera(ctx, camID)
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		if intVal(a, "id") == id {
			return a, nil
		}
	}
	return map[string]any{"id": id}, nil
}

func UpdateArea(ctx context.Context, id int, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	ativo := boolDefault(input, "ativo", true)
	_, err = db.ExecContext(ctx, `
UPDATE vis_camera_area SET
    nome = COALESCE(NULLIF($2,''), nome),
    ativo = $3,
    poligono_json = COALESCE(NULLIF($4,''), poligono_json),
    cor = COALESCE(NULLIF($5,''), cor)
WHERE id = $1`,
		id, strVal(input, "nome"), ativo, strVal(input, "poligono_json"), strVal(input, "cor"))
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `SELECT vis_camera_id FROM vis_camera_area WHERE id = $1`, id)
	var camID int
	if err := row.Scan(&camID); err != nil {
		return nil, err
	}
	list, _ := ListAreasByCamera(ctx, camID)
	for _, a := range list {
		if intVal(a, "id") == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("area nao encontrada")
}

func DeleteArea(ctx context.Context, id int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `DELETE FROM vis_camera_area WHERE id = $1`, id)
	return err
}

package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func ListLicencasByFranqueado(ctx context.Context, idFranqueado, status, unidade string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	q := `SELECT id, created_at, id_franqueado, plano, unidade, valor, pago_em, valido_ate,
        status, id_dispositivo, id_fatura, id_pagamento, observacao, vis_camera_id
FROM vis_licenca WHERE id_franqueado = $1`
	args := []any{idFranqueado}
	n := 2
	if status != "" {
		q += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if unidade != "" {
		q += fmt.Sprintf(" AND unidade = $%d", n)
		args = append(args, unidade)
	}
	q += " ORDER BY created_at DESC"
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []map[string]any
	resumo := map[string]int{"pendente": 0, "disponivel": 0, "em_uso": 0, "expirada": 0, "total": 0}
	for rows.Next() {
		var id int
		var created sql.NullTime
		var idFra, plano, unid, st, idDisp, idFat, idPag, obs sql.NullString
		var valor sql.NullFloat64
		var pago, valido sql.NullTime
		var camID sql.NullInt64
		if err := rows.Scan(&id, &created, &idFra, &plano, &unid, &valor, &pago, &valido,
			&st, &idDisp, &idFat, &idPag, &obs, &camID); err != nil {
			return nil, err
		}
		item := map[string]any{
			"id": id, "created_at": nullTime(created), "id_franqueado": nullStr(idFra),
			"plano": nullStr(plano), "unidade": nullStr(unid), "valor": nullFloat(valor),
			"pago_em": nullTime(pago), "valido_ate": nullTime(valido), "status": nullStr(st),
			"id_dispositivo": nullStr(idDisp), "id_fatura": nullStr(idFat),
			"id_pagamento": nullStr(idPag), "observacao": nullStr(obs),
			"vis_camera_id": nullInt(camID),
		}
		lista = append(lista, item)
		if st.Valid {
			if c, ok := resumo[st.String]; ok {
				resumo[st.String] = c + 1
			}
		}
		resumo["total"]++
	}
	return map[string]any{"dados": lista, "resumo": resumo}, nil
}

func CreateLicenca(ctx context.Context, input map[string]any) (map[string]any, error) {
	plano := strVal(input, "plano")
	flags := PlanoFlagsFrom(plano)
	db, err := DB()
	if err != nil {
		return nil, err
	}
	status := strVal(input, "status")
	if status == "" {
		status = "disponivel"
	}
	unidade := strVal(input, "unidade")
	if unidade == "" {
		unidade = flags.Unidade
	}
	var id int
	var created time.Time
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_licenca (id_franqueado, plano, unidade, valor, status, id_dispositivo, observacao, pago_em, valido_ate)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at`,
		strVal(input, "id_franqueado"), plano, unidade, flags.Valor, status,
		strVal(input, "id_dispositivo"), strVal(input, "observacao"),
		parseTimeInput(input, "pago_em"), parseTimeInput(input, "valido_ate"),
	).Scan(&id, &created)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id": id, "created_at": created.UTC().Format(time.RFC3339),
		"id_franqueado": strVal(input, "id_franqueado"), "plano": plano,
		"unidade": unidade, "valor": flags.Valor, "status": status,
	}, nil
}

func parseTimeInput(m map[string]any, key string) any {
	s := strVal(m, key)
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return t.UTC()
}

func AtivarGravacaoCamera(ctx context.Context, cameraID int, input map[string]any) (map[string]any, error) {
	licID := intVal(input, "vis_licenca_gravacao_id")
	if licID == 0 {
		licID = intVal(input, "vis_licenca_id")
	}
	if licID <= 0 {
		return nil, fmt.Errorf("vis_licenca_gravacao_id obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var plano sql.NullString
	err = db.QueryRowContext(ctx, `SELECT plano FROM vis_licenca WHERE id = $1 AND status = 'disponivel'`, licID).Scan(&plano)
	if err != nil {
		return nil, fmt.Errorf("Licenca de gravacao indisponivel")
	}
	flags := PlanoFlagsFrom(plano.String)
	sets := []string{"vis_licenca_gravacao_id = $2", "gravacao_status = 'ativa'", "gravacao_ativada_em = NOW()"}
	args := []any{cameraID, licID}
	if flags.GravaContinua {
		sets = append(sets, "grava_continua = true", fmt.Sprintf("retencao_dias = %d", flags.RetencaoDias))
	}
	if flags.GravaMovimento {
		sets = append(sets, "grava_movimento = true", fmt.Sprintf("retencao_dias = %d", flags.RetencaoDias))
	}
	if flags.GravaTimelapse {
		sets = append(sets, "grava_timelapse = true", fmt.Sprintf("retencao_dias = %d", flags.RetencaoDias))
	}
	q := fmt.Sprintf("UPDATE vis_camera SET %s WHERE id = $1", strings.Join(sets, ", "))
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	_, _ = db.ExecContext(ctx, `UPDATE vis_licenca SET status = 'em_uso', vis_camera_id = $2 WHERE id = $1`, licID, cameraID)
	return GetCameraByID(ctx, cameraID)
}

func DesativarGravacaoCamera(ctx context.Context, cameraID int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var licID sql.NullInt64
	_ = db.QueryRowContext(ctx, `SELECT vis_licenca_gravacao_id FROM vis_camera WHERE id = $1`, cameraID).Scan(&licID)
	_, err = db.ExecContext(ctx, `
UPDATE vis_camera SET grava_continua = false, grava_movimento = false, grava_timelapse = false,
    vis_licenca_gravacao_id = NULL, gravacao_status = NULL WHERE id = $1`, cameraID)
	if err != nil {
		return nil, err
	}
	if licID.Valid {
		_, _ = db.ExecContext(ctx, `UPDATE vis_licenca SET status = 'disponivel', vis_camera_id = NULL WHERE id = $1`, licID.Int64)
	}
	return GetCameraByID(ctx, cameraID)
}

package visdata

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cvgAcaoDesarmar = 0
	cvgAcaoArmar    = 1
)

var cvgTokenCache struct {
	mu    sync.Mutex
	token string
	expAt time.Time
}

// CvgGradeByClienteGET retorna config e slots da grade por cliente.
func CvgGradeByClienteGET(ctx context.Context, q url.Values) (int, []byte, error) {
	idCliente := strings.TrimSpace(q.Get("id_cliente"))
	if idCliente == "" {
		return bizErrJSON(fmt.Errorf("id_cliente obrigatorio"))
	}
	filtroDisp := strings.TrimSpace(q.Get("id_dispositivo"))

	db, err := DB()
	if err != nil {
		return errJSON(err)
	}

	var configID int
	var cfgGradeAtiva sql.NullBool
	err = db.QueryRowContext(ctx, `
SELECT id, grade_ativa FROM vis_cliente_grade_config WHERE id_cliente = $1 LIMIT 1`, idCliente).
		Scan(&configID, &cfgGradeAtiva)
	if err == sql.ErrNoRows {
		configID = 0
	} else if err != nil {
		return errJSON(err)
	}

	gradeAtiva, err := cvgEscopoGradeAtiva(ctx, idCliente, filtroDisp)
	if err != nil {
		return errJSON(err)
	}

	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, id_franqueado, id_cliente, COALESCE(id_dispositivo, ''), dia_semana, hora, acao, ativo
FROM vis_cliente_grade_slot
WHERE id_cliente = $1
ORDER BY dia_semana ASC, hora ASC`, idCliente)
	if err != nil {
		return errJSON(err)
	}
	defer rows.Close()

	slots := []map[string]any{}
	for rows.Next() {
		slot, err := scanGradeSlotRow(rows)
		if err != nil {
			return errJSON(err)
		}
		disp := strVal(slot, "id_dispositivo")
		incluir := false
		if filtroDisp != "" {
			incluir = disp == filtroDisp
		} else {
			incluir = disp == ""
		}
		if incluir {
			slots = append(slots, slot)
		}
	}
	if err := rows.Err(); err != nil {
		return errJSON(err)
	}

	_ = cfgGradeAtiva
	return okJSON(map[string]any{
		"id_cliente":  idCliente,
		"grade_ativa": gradeAtiva,
		"config_id":   configID,
		"slots":       slots,
	})
}

// CvgGradeByClientePUT salva grade_ativa e slots do cliente (substitui slots do escopo).
func CvgGradeByClientePUT(ctx context.Context, payload map[string]any) (int, []byte, error) {
	idCliente := strVal(payload, "id_cliente")
	idFranqueado := strVal(payload, "id_franqueado")
	if idCliente == "" {
		return bizErrJSON(fmt.Errorf("id_cliente obrigatorio"))
	}
	if idFranqueado == "" {
		return bizErrJSON(fmt.Errorf("id_franqueado obrigatorio"))
	}

	gradeAtiva := false
	if v := boolVal(payload, "grade_ativa"); v != nil {
		gradeAtiva = *v
	}
	escopo := strVal(payload, "id_dispositivo_escopo")

	for _, sl := range parseSlotsArray(payload["slots"]) {
		if strVal(sl, "hora") == "" {
			return bizErrJSON(fmt.Errorf("hora obrigatoria em cada slot (HH:MM)"))
		}
		acaoNorm := strVal(sl, "acao")
		if acaoNorm != "ativar" && acaoNorm != "desativar" {
			return bizErrJSON(fmt.Errorf("acao deve ser ativar ou desativar"))
		}
	}

	out, err := cvgGradeByClientePUTOnce(ctx, payload, idCliente, idFranqueado, gradeAtiva, escopo)
	if err != nil && isBadConn(err) {
		resetDB()
		out, err = cvgGradeByClientePUTOnce(ctx, payload, idCliente, idFranqueado, gradeAtiva, escopo)
	}
	if err != nil {
		return errJSON(humanizeDBErr(err))
	}
	return okJSON(out)
}

func cvgGradeByClientePUTOnce(ctx context.Context, payload map[string]any, idCliente, idFranqueado string, gradeAtiva bool, escopo string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var cfgID int
	err = tx.QueryRowContext(ctx, `
SELECT id FROM vis_cliente_grade_config WHERE id_cliente = $1 LIMIT 1`, idCliente).Scan(&cfgID)
	if err == sql.ErrNoRows {
		err = tx.QueryRowContext(ctx, `
INSERT INTO vis_cliente_grade_config (id_franqueado, id_cliente, grade_ativa)
VALUES ($1, $2, false) RETURNING id`, idFranqueado, idCliente).Scan(&cfgID)
	}
	if err != nil {
		return nil, err
	}

	if err := cvgEscopoSalvarTx(ctx, tx, idFranqueado, idCliente, escopo, gradeAtiva); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
DELETE FROM vis_cliente_grade_slot
WHERE id_cliente = $1 AND COALESCE(id_dispositivo, '') = $2`, idCliente, escopo); err != nil {
		return nil, err
	}

	slotsIn := parseSlotsArray(payload["slots"])
	slotsSalvos := make([]map[string]any, 0, len(slotsIn))
	for _, sl := range slotsIn {
		horaNorm := strVal(sl, "hora")
		acaoNorm := strVal(sl, "acao")
		ativo := true
		if v := boolVal(sl, "ativo"); v != nil {
			ativo = *v
		}
		diaSemana := intVal(sl, "dia_semana")
		var slotID int
		var createdAt time.Time
		err := tx.QueryRowContext(ctx, `
INSERT INTO vis_cliente_grade_slot (
    id_franqueado, id_cliente, id_dispositivo, dia_semana, hora, acao, ativo
) VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id, created_at`,
			idFranqueado, idCliente, escopo, diaSemana, horaNorm, acaoNorm, ativo,
		).Scan(&slotID, &createdAt)
		if err != nil {
			return nil, err
		}
		slotsSalvos = append(slotsSalvos, map[string]any{
			"id":             slotID,
			"created_at":     createdAt.UTC().Format(time.RFC3339),
			"id_franqueado":  idFranqueado,
			"id_cliente":     idCliente,
			"id_dispositivo": escopo,
			"dia_semana":     diaSemana,
			"hora":           horaNorm,
			"acao":           acaoNorm,
			"ativo":          ativo,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]any{
		"success":     true,
		"id_cliente":  idCliente,
		"grade_ativa": gradeAtiva,
		"slots":       slotsSalvos,
	}, nil
}

// CvgGradeListGET lista grades por franqueado (cliente+dispositivo com slots salvos).
func CvgGradeListGET(ctx context.Context, q url.Values) (int, []byte, error) {
	idFranqueado := strings.TrimSpace(q.Get("id_franqueado"))
	if idFranqueado == "" {
		return bizErrJSON(fmt.Errorf("id_franqueado obrigatorio"))
	}

	itens, err := cvgGradeListData(ctx, idFranqueado)
	if err != nil && isBadConn(err) {
		resetDB()
		itens, err = cvgGradeListData(ctx, idFranqueado)
	}
	if err != nil {
		return errJSON(humanizeDBErr(err))
	}

	return okJSON(map[string]any{"itens": itens})
}

func cvgGradeListData(ctx context.Context, idFranqueado string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
SELECT DISTINCT id_cliente, COALESCE(id_dispositivo, '') AS id_dispositivo
FROM vis_cliente_grade_slot
WHERE id_franqueado = $1
ORDER BY id_cliente ASC, COALESCE(id_dispositivo, '') ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := []map[string]any{}
	for rows.Next() {
		var idCliente, disp string
		if err := rows.Scan(&idCliente, &disp); err != nil {
			return nil, err
		}
		ativa, err := cvgEscopoGradeAtiva(ctx, idCliente, disp)
		if err != nil {
			return nil, err
		}
		itens = append(itens, map[string]any{
			"id_cliente":     idCliente,
			"id_dispositivo": disp,
			"grade_ativa":    ativa,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return itens, nil
}

// CvgGradeEscopoAtivaPATCH ativa/desativa grade por escopo sem alterar slots.
func CvgGradeEscopoAtivaPATCH(ctx context.Context, payload map[string]any) (int, []byte, error) {
	idFranqueado := strVal(payload, "id_franqueado")
	idCliente := strVal(payload, "id_cliente")
	if idFranqueado == "" {
		return bizErrJSON(fmt.Errorf("id_franqueado obrigatorio"))
	}
	if idCliente == "" {
		return bizErrJSON(fmt.Errorf("id_cliente obrigatorio"))
	}
	disp := strVal(payload, "id_dispositivo")
	gradeAtiva := false
	if v := boolVal(payload, "grade_ativa"); v != nil {
		gradeAtiva = *v
	}

	db, err := DB()
	if err != nil {
		return errJSON(err)
	}

	var qtd int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_cliente_grade_slot
WHERE id_cliente = $1 AND COALESCE(id_dispositivo, '') = $2`, idCliente, disp).Scan(&qtd)
	if err != nil {
		return errJSON(err)
	}
	if qtd == 0 {
		return bizErrJSON(fmt.Errorf("Nenhum horario salvo para este escopo"))
	}

	if err := cvgEscopoSalvar(ctx, idFranqueado, idCliente, disp, gradeAtiva); err != nil {
		return errJSON(err)
	}

	return okJSON(map[string]any{
		"success":        true,
		"id_cliente":     idCliente,
		"id_dispositivo": disp,
		"grade_ativa":    gradeAtiva,
	})
}

// CvgWorkerTickPOST executa slots pendentes do minuto (worker).
func CvgWorkerTickPOST(ctx context.Context, payload map[string]any, workerKeyHeader string) (int, []byte, error) {
	workerKey := strings.TrimSpace(workerKeyHeader)
	if workerKey == "" {
		workerKey = strVal(payload, "worker_key")
	}
	if err := cvgWorkerValidar(workerKey); err != nil {
		return bizErrJSON(err)
	}

	diaSemana := intVal(payload, "dia_semana")
	if diaSemana < 1 || diaSemana > 7 {
		if s := strVal(payload, "dia_semana"); s != "" {
			diaSemana, _ = strconv.Atoi(s)
		}
	}
	if diaSemana < 1 || diaSemana > 7 {
		return bizErrJSON(fmt.Errorf("dia_semana obrigatorio (1=seg .. 7=dom)"))
	}
	hora := strVal(payload, "hora")
	dataRef := strVal(payload, "data_ref")
	if hora == "" {
		return bizErrJSON(fmt.Errorf("hora obrigatoria (HH:MM)"))
	}
	if dataRef == "" {
		return bizErrJSON(fmt.Errorf("data_ref obrigatorio (YYYY-MM-DD)"))
	}

	resultado, err := cvgWorkerTick(ctx, diaSemana, hora, dataRef)
	if err != nil {
		return errJSON(err)
	}
	return okJSON(resultado)
}

func cvgWorkerValidar(workerKey string) error {
	esperado := strings.TrimSpace(os.Getenv("CVG_WORKER_KEY"))
	if esperado == "" {
		esperado = strings.TrimSpace(os.Getenv("CONFVISION_GRADE_WORKER_KEY"))
	}
	if esperado == "" {
		return fmt.Errorf("cvg_worker_secret nao configurado em fp_config_financeiro")
	}
	if workerKey != esperado {
		return fmt.Errorf("Worker key invalida")
	}
	return nil
}

func cvgPlanoTipo(plano string) (tipo string, capturaAnalitico, somenteArmado bool) {
	flags := PlanoFlagsFrom(plano)
	capturaAnalitico = flags.CapturaAnalitico
	somenteArmado = flags.SomenteArmado
	tipo = "outro"
	if !capturaAnalitico {
		return tipo, capturaAnalitico, somenteArmado
	}
	if somenteArmado {
		tipo = "armado"
	} else {
		tipo = "24h"
	}
	return tipo, capturaAnalitico, somenteArmado
}

func cvgEscopoGradeAtiva(ctx context.Context, idCliente, idDispositivo string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	disp := strings.TrimSpace(idDispositivo)

	rows, err := db.QueryContext(ctx, `
SELECT id, COALESCE(id_dispositivo, ''), grade_ativa
FROM vis_cliente_grade_escopo WHERE id_cliente = $1`, idCliente)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var escopoAtiva sql.NullBool
	found := false
	for rows.Next() {
		var id int
		var ed string
		var ga sql.NullBool
		if err := rows.Scan(&id, &ed, &ga); err != nil {
			return false, err
		}
		if ed == disp {
			escopoAtiva = ga
			found = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	if found {
		return escopoAtiva.Valid && escopoAtiva.Bool, nil
	}

	var cfgAtiva sql.NullBool
	err = db.QueryRowContext(ctx, `
SELECT grade_ativa FROM vis_cliente_grade_config WHERE id_cliente = $1 LIMIT 1`, idCliente).
		Scan(&cfgAtiva)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return cfgAtiva.Valid && cfgAtiva.Bool, nil
}

func cvgEscopoSalvar(ctx context.Context, idFranqueado, idCliente, idDispositivo string, gradeAtiva bool) error {
	db, err := DB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := cvgEscopoSalvarTx(ctx, tx, idFranqueado, idCliente, idDispositivo, gradeAtiva); err != nil {
		return err
	}
	return tx.Commit()
}

func cvgEscopoSalvarTx(ctx context.Context, tx *sql.Tx, idFranqueado, idCliente, idDispositivo string, gradeAtiva bool) error {
	disp := strings.TrimSpace(idDispositivo)
	var existID int
	err := tx.QueryRowContext(ctx, `
SELECT id FROM vis_cliente_grade_escopo
WHERE id_cliente = $1 AND COALESCE(id_dispositivo, '') = $2
LIMIT 1`, idCliente, disp).Scan(&existID)
	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx, `
INSERT INTO vis_cliente_grade_escopo (id_franqueado, id_cliente, id_dispositivo, grade_ativa)
VALUES ($1,$2,$3,$4)`, idFranqueado, idCliente, disp, gradeAtiva)
		return err
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
UPDATE vis_cliente_grade_escopo SET id_franqueado = $2, grade_ativa = $3 WHERE id = $1`,
		existID, idFranqueado, gradeAtiva)
	return err
}

func cvgSlotsPendentes(ctx context.Context, diaSemana int, hora, dataRef string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, id_franqueado, id_cliente, COALESCE(id_dispositivo, ''), dia_semana, hora, acao, ativo
FROM vis_cliente_grade_slot
WHERE dia_semana = $1 AND hora = $2 AND ativo = TRUE
ORDER BY id ASC`, diaSemana, hora)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pendentes := []map[string]any{}
	for rows.Next() {
		slot, err := scanGradeSlotRow(rows)
		if err != nil {
			return nil, err
		}
		slotID := intVal(slot, "id")
		idCliente := strVal(slot, "id_cliente")
		idDisp := strVal(slot, "id_dispositivo")

		ativa, err := cvgEscopoGradeAtiva(ctx, idCliente, idDisp)
		if err != nil {
			return nil, err
		}
		if !ativa {
			continue
		}

		var execID int
		err = db.QueryRowContext(ctx, `
SELECT id FROM vis_cliente_grade_exec
WHERE slot_id = $1 AND data_ref = $2 AND hora_ref = $3 LIMIT 1`, slotID, dataRef, hora).Scan(&execID)
		if err == sql.ErrNoRows {
			pendentes = append(pendentes, slot)
		} else if err != nil {
			return nil, err
		}
	}
	return pendentes, rows.Err()
}

func cvgWorkerTick(ctx context.Context, diaSemana int, hora, dataRef string) (map[string]any, error) {
	pendentes, err := cvgSlotsPendentes(ctx, diaSemana, hora, dataRef)
	if err != nil {
		return nil, err
	}

	executados, ignorados, erros := 0, 0, 0
	detalhes := []map[string]any{}

	for _, slot := range pendentes {
		slotID := intVal(slot, "id")
		res, err := cvgExecutarSlot(ctx, slotID, dataRef, hora)
		if err != nil {
			erros++
			detalhes = append(detalhes, map[string]any{
				"slot_id": slotID,
				"erro":    "execucao_falhou",
			})
			continue
		}
		res["slot_id"] = slotID
		if res["ignorado"] == true {
			ignorados++
		} else {
			executados++
			if r := strVal(res, "resultado"); r == "erro" || r == "parcial" {
				erros++
			}
		}
		detalhes = append(detalhes, res)
	}

	return map[string]any{
		"data_ref":   dataRef,
		"hora":       hora,
		"dia_semana": diaSemana,
		"total":      len(pendentes),
		"executados": executados,
		"ignorados":  ignorados,
		"erros":      erros,
		"detalhes":   detalhes,
	}, nil
}

func cvgExecutarSlot(ctx context.Context, slotID int, dataRef, horaRef string) (map[string]any, error) {
	if dataRef == "" {
		return nil, fmt.Errorf("data_ref obrigatorio")
	}
	if horaRef == "" {
		return nil, fmt.Errorf("hora_ref obrigatorio")
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var execID int
	err = db.QueryRowContext(ctx, `
SELECT id FROM vis_cliente_grade_exec
WHERE slot_id = $1 AND data_ref = $2 AND hora_ref = $3 LIMIT 1`, slotID, dataRef, horaRef).Scan(&execID)
	if err == nil {
		return map[string]any{
			"slot_id":  slotID,
			"ignorado": true,
			"motivo":   "ja_executado",
			"exec_id":  execID,
		}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	var slot struct {
		IDCliente     string
		IDDispositivo sql.NullString
		Acao          string
		Ativo         bool
	}
	err = db.QueryRowContext(ctx, `
SELECT id_cliente, id_dispositivo, acao, ativo
FROM vis_cliente_grade_slot WHERE id = $1`, slotID).
		Scan(&slot.IDCliente, &slot.IDDispositivo, &slot.Acao, &slot.Ativo)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Slot nao encontrado")
	}
	if err != nil {
		return nil, err
	}
	if !slot.Ativo {
		return nil, fmt.Errorf("Slot inativo")
	}

	idDisp := ""
	if slot.IDDispositivo.Valid {
		idDisp = strings.TrimSpace(slot.IDDispositivo.String)
	}

	ativa, err := cvgEscopoGradeAtiva(ctx, slot.IDCliente, idDisp)
	if err != nil {
		return nil, err
	}
	if !ativa {
		return nil, fmt.Errorf("Grade desativada para este dispositivo")
	}

	detalhe := map[string]any{}
	resultado := "ok"

	switch slot.Acao {
	case "desativar":
		detalhe = cvgAplicarDesativar(ctx, slot.IDCliente, idDisp)
	case "ativar":
		detalhe = cvgAplicarAtivar(ctx, slot.IDCliente, idDisp)
	default:
		resultado = "erro"
		detalhe = map[string]any{"erro": "acao invalida"}
	}

	if erros, ok := detalhe["erros"].([]string); ok && len(erros) > 0 {
		resultado = "parcial"
	}

	detalheJSON, _ := json.Marshal(detalhe)
	var newExecID int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_cliente_grade_exec (slot_id, data_ref, hora_ref, executado_em, resultado, detalhe)
VALUES ($1,$2,$3,NOW(),$4,$5::jsonb) RETURNING id`,
		slotID, dataRef, horaRef, resultado, string(detalheJSON)).Scan(&newExecID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"slot_id":   slotID,
		"ignorado":  false,
		"resultado": resultado,
		"detalhe":   detalhe,
		"exec_id":   newExecID,
	}, nil
}

func cvgListarCamerasSlot(ctx context.Context, idCliente, idDispositivo string) ([]map[string]any, error) {
	if strings.TrimSpace(idCliente) == "" {
		return nil, fmt.Errorf("id_cliente obrigatorio")
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}
	query := `
SELECT id, id_cliente, COALESCE(id_dispositivo, ''), plano, ativo
FROM vis_camera WHERE id_cliente = $1 AND ativo = TRUE ORDER BY id ASC`
	rows, err := db.QueryContext(ctx, query, idCliente)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cameras := []map[string]any{}
	for rows.Next() {
		var id int
		var idCli, idDisp, plano string
		var ativo bool
		if err := rows.Scan(&id, &idCli, &idDisp, &plano, &ativo); err != nil {
			return nil, err
		}
		if idDispositivo != "" && idDisp != idDispositivo {
			continue
		}
		tipo, _, _ := cvgPlanoTipo(plano)
		if tipo == "outro" {
			continue
		}
		cameras = append(cameras, map[string]any{
			"id":             id,
			"id_cliente":     idCli,
			"id_dispositivo": idDisp,
			"plano":          plano,
			"plano_tipo":     tipo,
			"ativo":          ativo,
		})
	}
	return cameras, rows.Err()
}

func cvgCameraPausar(ctx context.Context, visCameraID int, pausado bool) error {
	db, err := DB()
	if err != nil {
		return err
	}
	var plano string
	err = db.QueryRowContext(ctx, `SELECT plano FROM vis_camera WHERE id = $1`, visCameraID).Scan(&plano)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Camera nao encontrada")
	}
	if err != nil {
		return err
	}
	_, captura, _ := cvgPlanoTipo(plano)
	if !captura {
		return fmt.Errorf("Plano sem deteccao analitica")
	}
	_, err = db.ExecContext(ctx, `UPDATE vis_camera SET analitico_pausado = $2 WHERE id = $1`, visCameraID, pausado)
	return err
}

func cvgAplicarDesativar(ctx context.Context, idCliente, idDispositivo string) map[string]any {
	cameras, err := cvgListarCamerasSlot(ctx, idCliente, idDispositivo)
	erros := []string{}
	camerasPausar := []int{}
	dispositivos := []map[string]any{}
	dispFeitos := ","

	if err != nil {
		erros = append(erros, "listar:"+err.Error())
		return map[string]any{
			"acao":           "desativar",
			"cameras_pausar": camerasPausar,
			"dispositivos":   dispositivos,
			"erros":          erros,
		}
	}

	for _, cam := range cameras {
		camID := intVal(cam, "id")
		if err := cvgCameraPausar(ctx, camID, true); err != nil {
			erros = append(erros, fmt.Sprintf("pausar:%d", camID))
			continue
		}
		camerasPausar = append(camerasPausar, camID)

		if strVal(cam, "plano_tipo") == "armado" {
			idDisp := strVal(cam, "id_dispositivo")
			if idDisp == "" {
				continue
			}
			chave := "," + idDisp + ","
			if strings.Contains(dispFeitos, chave) {
				continue
			}
			dispFeitos += idDisp + ","
			d, err := cvgDispositivoArmar(ctx, idDisp, cvgAcaoDesarmar)
			if err != nil {
				erros = append(erros, "desarmar:"+idDisp)
			} else {
				dispositivos = append(dispositivos, d)
			}
		}
	}

	return map[string]any{
		"acao":           "desativar",
		"cameras_pausar": camerasPausar,
		"dispositivos":   dispositivos,
		"erros":          erros,
	}
}

func cvgAplicarAtivar(ctx context.Context, idCliente, idDispositivo string) map[string]any {
	cameras, err := cvgListarCamerasSlot(ctx, idCliente, idDispositivo)
	erros := []string{}
	camerasDespausar := []int{}
	dispositivos := []map[string]any{}
	dispFeitos := ","

	if err != nil {
		erros = append(erros, "listar:"+err.Error())
		return map[string]any{
			"acao":              "ativar",
			"dispositivos":      dispositivos,
			"cameras_despausar": camerasDespausar,
			"erros":             erros,
		}
	}

	for _, cam := range cameras {
		if strVal(cam, "plano_tipo") == "armado" {
			idDisp := strVal(cam, "id_dispositivo")
			if idDisp == "" {
				continue
			}
			chave := "," + idDisp + ","
			if strings.Contains(dispFeitos, chave) {
				continue
			}
			dispFeitos += idDisp + ","
			d, err := cvgDispositivoArmar(ctx, idDisp, cvgAcaoArmar)
			if err != nil {
				erros = append(erros, "armar:"+idDisp)
			} else {
				dispositivos = append(dispositivos, d)
			}
		}
	}

	for _, cam := range cameras {
		camID := intVal(cam, "id")
		if err := cvgCameraPausar(ctx, camID, false); err != nil {
			erros = append(erros, fmt.Sprintf("despausar:%d", camID))
			continue
		}
		camerasDespausar = append(camerasDespausar, camID)
	}

	return map[string]any{
		"acao":              "ativar",
		"dispositivos":      dispositivos,
		"cameras_despausar": camerasDespausar,
		"erros":             erros,
	}
}

func cvgDispositivoArmar(ctx context.Context, idDispositivo string, acao int) (map[string]any, error) {
	idDispositivo = strings.TrimSpace(idDispositivo)
	if idDispositivo == "" {
		return nil, fmt.Errorf("id_dispositivo obrigatorio")
	}

	urlDisp := cvgEnv("CVG_DISPOSITIVO_URL", "http://185.130.61.4:2010/v4/dispositivo/getDadosById")
	urlCmd := cvgEnv("CVG_COMANDO_URL", "http://185.130.61.3:2030/armar")
	senhaWeb := cvgEnv("CVG_COMANDO_SENHA_WEB", "")

	token, err := cvgWebLogarToken(ctx)
	if err != nil {
		return nil, err
	}

	bodyDisp, status, err := cvgHTTPJSON(ctx, http.MethodPost, urlDisp, map[string]any{
		"idDispositivo": idDispositivo,
	}, map[string]string{"Authorization": "Bearer " + token})
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("getDadosById status=%d", status)
	}

	disp, err := cvgExtractDados(bodyDisp)
	if err != nil || disp == nil {
		return nil, fmt.Errorf("Dispositivo nao encontrado na central")
	}

	particao := "1"
	if p, ok := disp["particao"].(string); ok && strings.TrimSpace(p) != "" {
		particao = strings.TrimSpace(p)
	}
	particaoNum, _ := strconv.Atoi(particao)
	senha := ""
	if s, ok := disp["senha"].(string); ok {
		senha = s
	}
	armadoAntes := ""
	if a, ok := disp["armado"].(string); ok {
		armadoAntes = a
	}

	bodyCmd, status, err := cvgHTTPJSON(ctx, http.MethodPost, urlCmd, map[string]any{
		"idDispositivo": idDispositivo,
		"numero":        particaoNum,
		"usuario":       "000",
		"acao":          acao,
		"senha":         senha,
		"senhaWeb":      senhaWeb,
	}, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("comando status=%d", status)
	}

	statusCmd := cvgExtractStatus(bodyCmd)

	return map[string]any{
		"ok":            true,
		"id_dispositivo": idDispositivo,
		"acao":          acao,
		"status_cmd":    statusCmd,
		"armado_antes":  armadoAntes,
	}, nil
}

func cvgWebLogarToken(ctx context.Context) (string, error) {
	cvgTokenCache.mu.Lock()
	defer cvgTokenCache.mu.Unlock()

	if cvgTokenCache.token != "" && time.Now().Before(cvgTokenCache.expAt.Add(-5*time.Minute)) {
		return cvgTokenCache.token, nil
	}

	urlLogar := cvgEnv("CVG_WEBLOGAR_URL", "http://185.130.61.4:2010/v4/cliente/webLogar")
	senha := cvgEnv("CVG_WEBLOGAR_SENHA", "")
	if senha == "" {
		return "", fmt.Errorf("CVG_WEBLOGAR_SENHA nao configurada")
	}

	body, status, err := cvgHTTPJSON(ctx, http.MethodPost, urlLogar, map[string]string{"senha": senha}, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("webLogar status=%d", status)
	}

	var resp struct {
		Status string `json:"status"`
		Dados  any    `json:"dados"`
		Result struct {
			Dados any `json:"dados"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	token := cvgExtractToken(resp.Dados)
	if token == "" {
		token = cvgExtractToken(resp.Result.Dados)
	}
	if token == "" {
		return "", fmt.Errorf("token nao encontrado na resposta webLogar")
	}

	cvgTokenCache.token = token
	cvgTokenCache.expAt = time.Now().Add(12 * time.Hour)
	return token, nil
}

func cvgHTTPJSON(ctx context.Context, method, rawURL string, payload any, headers map[string]string) ([]byte, int, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

func cvgExtractDados(body []byte) (map[string]any, error) {
	var wrap map[string]any
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	if d, ok := wrap["dados"].(map[string]any); ok {
		return d, nil
	}
	if r, ok := wrap["result"].(map[string]any); ok {
		if d, ok := r["dados"].(map[string]any); ok {
			return d, nil
		}
	}
	return nil, fmt.Errorf("dados ausente")
}

func cvgExtractStatus(body []byte) string {
	var wrap map[string]any
	if err := json.Unmarshal(body, &wrap); err != nil {
		return ""
	}
	if s, ok := wrap["status"].(string); ok {
		return s
	}
	if r, ok := wrap["result"].(map[string]any); ok {
		if s, ok := r["status"].(string); ok {
			return s
		}
	}
	return ""
}

func cvgExtractToken(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		for _, key := range []string{"token", "Token", "access_token", "jwt"} {
			if vv, ok := x[key]; ok {
				if t := cvgExtractToken(vv); t != "" {
					return t
				}
			}
		}
	case []any:
		for _, el := range x {
			if t := cvgExtractToken(el); t != "" {
				return t
			}
		}
	}
	return ""
}

func cvgEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func scanGradeSlotRow(rows *sql.Rows) (map[string]any, error) {
	var id, diaSemana int
	var createdAt time.Time
	var idFranqueado, idCliente, idDisp, hora, acao sql.NullString
	var ativo sql.NullBool
	if err := rows.Scan(&id, &createdAt, &idFranqueado, &idCliente, &idDisp, &diaSemana, &hora, &acao, &ativo); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":             id,
		"created_at":     createdAt.UTC().Format(time.RFC3339),
		"id_franqueado":  nullStr(idFranqueado),
		"id_cliente":     nullStr(idCliente),
		"id_dispositivo": nullStr(idDisp),
		"dia_semana":     diaSemana,
		"hora":           nullStr(hora),
		"acao":           nullStr(acao),
		"ativo":          nullBool(ativo),
	}, nil
}

func parseSlotsArray(raw any) []map[string]any {
	if raw == nil {
		return nil
	}
	switch t := raw.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, el := range t {
			if m, ok := el.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return t
	default:
		return nil
	}
}

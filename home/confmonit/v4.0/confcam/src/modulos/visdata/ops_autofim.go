package visdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OpsAutofimPending lista fila pendente pronta para execucao.
func OpsAutofimPending(ctx context.Context, limit int) (map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	agoraMs := time.Now().UnixMilli()
	rows, err := db.QueryContext(ctx, `
SELECT id, id_processo, id_dispositivo, status, tentativas, alarm_events_id, rodar_em, ultimo_evento_ts, updated_at
FROM ops_bot_finalizaeventoauto
WHERE status = 'PENDENTE' AND rodar_em <= $1
ORDER BY id ASC
LIMIT $2`, agoraMs, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dados := []map[string]any{}
	for rows.Next() {
		row, err := scanAutofimQueueRow(rows)
		if err != nil {
			return nil, err
		}
		dados = append(dados, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return map[string]any{"dados": dados}, nil
}

// OpsAutofimClaim reserva registro da fila para processamento.
func OpsAutofimClaim(ctx context.Context, id int, lockToken string) (map[string]any, error) {
	if id <= 0 {
		return nil, fmt.Errorf("id obrigatório")
	}
	lockToken = strings.TrimSpace(lockToken)
	if lockToken == "" {
		return nil, fmt.Errorf("lock_token obrigatório")
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	agoraMs := time.Now().UnixMilli()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var candID int
	err = tx.QueryRowContext(ctx, `
SELECT id FROM ops_bot_finalizaeventoauto
WHERE id = $1 AND status = 'PENDENTE' AND rodar_em <= $2
FOR UPDATE`, id, agoraMs).Scan(&candID)
	if err == sql.ErrNoRows {
		return map[string]any{
			"dados": map[string]any{
				"claimed": false,
				"row":     map[string]any{},
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
UPDATE ops_bot_finalizaeventoauto SET
    status = 'PROCESSANDO', lock_token = $2, lock_at = NOW(), updated_at = NOW(), erro = ''
WHERE id = $1`, candID, lockToken)
	if err != nil {
		return nil, err
	}

	row, err := getAutofimQueueByIDTx(ctx, tx, candID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]any{
		"dados": map[string]any{
			"claimed": true,
			"row":     row,
		},
	}, nil
}

// OpsAutofimContext monta contexto de analise para um evento.
func OpsAutofimContext(ctx context.Context, idEvento int) (map[string]any, error) {
	if idEvento <= 0 {
		return nil, fmt.Errorf("idEvento obrigatório")
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	evt, err := getOpsAlarmEventByID(ctx, db, idEvento)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Evento não encontrado")
		}
		return nil, err
	}
	idProcesso := strVal(evt, "idProcesso")
	if idProcesso == "" {
		return nil, fmt.Errorf("Evento sem idProcesso")
	}
	idDispositivo := strVal(evt, "idDispositivo")

	apiProcOk := false
	dadosProcessoAPI := map[string]any{}
	dataAtenFim := ""
	processoJaFinalizado := false

	procDados, err := opsGetDadosProcessoByID(ctx, idProcesso)
	if err == nil && procDados != nil {
		apiProcOk = true
		dadosProcessoAPI = procDados
		if v, ok := procDados["dataAtenFim"].(string); ok {
			dataAtenFim = strings.TrimSpace(v)
		}
	}
	if dataAtenFim != "" && dataAtenFim != "01/01/0001 00:00:00" {
		processoJaFinalizado = true
	}

	histAgg, err := opsHistAgg(ctx, db, idDispositivo)
	if err != nil {
		return nil, err
	}

	eventosProcesso, err := listOpsAlarmEventsByProcesso(ctx, db, idProcesso)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"dados": map[string]any{
			"evento":               evt,
			"eventosProcesso":      eventosProcesso,
			"histAgg":              histAgg,
			"processoJaFinalizado": processoJaFinalizado,
			"dataAtenFim":          dataAtenFim,
			"dadosProcessoApi":     dadosProcessoAPI,
			"apiProcOk":            apiProcOk,
		},
	}, nil
}

// OpsAutofimProcessEnd finaliza processo na central ConfMonit.
func OpsAutofimProcessEnd(ctx context.Context, idProcesso, motivo, perfilLocal string) (map[string]any, error) {
	idProcesso = strings.TrimSpace(idProcesso)
	motivo = strings.TrimSpace(motivo)
	if idProcesso == "" {
		return nil, fmt.Errorf("idProcesso obrigatório")
	}
	if motivo == "" {
		return nil, fmt.Errorf("motivo obrigatório")
	}
	perfil := strings.TrimSpace(perfilLocal)
	if perfil == "" {
		perfil = "NORMAL"
	}

	retorno, acao, err := opsProcessoEnd(ctx, idProcesso, motivo, perfil)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"dados": map[string]any{
			"acao":               acao,
			"retornoProcessoEnd": retorno,
		},
	}, nil
}

// OpsAutofimLog grava log de decisao autofim.
func OpsAutofimLog(ctx context.Context, payload map[string]any) (map[string]any, error) {
	idEvento := intVal(payload, "idEvento")
	if idEvento <= 0 {
		idEvento = intVal(payload, "id_evento")
	}
	idProcesso := strVal(payload, "idProcesso")
	idDispositivo := strVal(payload, "idDispositivo")
	motivo := strVal(payload, "motivo")
	acao := strVal(payload, "acao")
	if idEvento <= 0 {
		return nil, fmt.Errorf("idEvento obrigatório")
	}
	if idProcesso == "" {
		return nil, fmt.Errorf("idProcesso obrigatório")
	}
	if idDispositivo == "" {
		return nil, fmt.Errorf("idDispositivo obrigatório")
	}
	if motivo == "" {
		return nil, fmt.Errorf("motivo obrigatório")
	}
	if acao == "" {
		return nil, fmt.Errorf("acao obrigatório")
	}

	qtdCiclos := intVal(payload, "qtdCiclos3mDiffProc")
	regraVersao := strVal(payload, "regraVersao")
	if regraVersao == "" {
		regraVersao = "vFinal_go_domain_api"
	}

	retornoJSON, _ := json.Marshal(payload["retornoProcessoEnd"])

	db, err := DB()
	if err != nil {
		return nil, err
	}

	var idLog int
	err = db.QueryRowContext(ctx, `
INSERT INTO ops_alarm_event_autofimlog (
    id_evento, id_processo, id_dispositivo, motivo, acao, qtd_ciclos_5m, bloqueio_3x5,
    tem_falhas, tem_alarme, tem_desarme, tem_restaure, tem_par_alarme_rest_50,
    retorno_processo_end, regra_versao, erro
) VALUES ($1,$2,$3,$4,$5,$6,false,$7,$8,$9,$10,$11,$12::jsonb,$13,$14)
RETURNING id`,
		idEvento, idProcesso, idDispositivo, motivo, acao, qtdCiclos,
		boolDefaultMap(payload, "temFalhas", false),
		boolDefaultMap(payload, "temAlarme", false),
		boolDefaultMap(payload, "temDesarme", false),
		boolDefaultMap(payload, "temRestaure", false),
		boolDefaultMap(payload, "temParAlarmeRest50", false),
		string(retornoJSON), regraVersao, strVal(payload, "erro"),
	).Scan(&idLog)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"dados": map[string]any{
			"ok":    true,
			"idLog": idLog,
		},
	}, nil
}

// OpsAutofimSuccess marca registro como finalizado com sucesso.
func OpsAutofimSuccess(ctx context.Context, payload map[string]any) (map[string]any, error) {
	id := intVal(payload, "id")
	lockToken := strings.TrimSpace(strVal(payload, "lock_token"))
	if id <= 0 {
		return nil, fmt.Errorf("id obrigatório")
	}
	if lockToken == "" {
		return nil, fmt.Errorf("lock_token obrigatório")
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	row, err := getAutofimQueueByID(ctx, db, id)
	if err != nil {
		return nil, fmt.Errorf("registro não encontrado")
	}

	status := strVal(row, "status")
	skipped := false
	reason := ""

	if status == "FINALIZADO" {
		skipped = true
		reason = "already_finalized"
	} else if status == "PROCESSANDO" && strVal(row, "lock_token") == lockToken {
		payloadJSON, _ := json.Marshal(payload["payload_resultado"])
		_, err = db.ExecContext(ctx, `
UPDATE ops_bot_finalizaeventoauto SET
    status = 'FINALIZADO', motivo_final = $2, acao_final = $3,
    payload_resultado = $4::jsonb, erro = '', lock_token = '', lock_at = NULL, updated_at = NOW()
WHERE id = $1`, id, strVal(payload, "motivo_final"), strVal(payload, "acao_final"), string(payloadJSON))
		if err != nil {
			return nil, err
		}
	} else {
		skipped = true
		reason = "status_or_lock_mismatch"
	}

	return map[string]any{
		"dados": map[string]any{
			"ok":      true,
			"id":      id,
			"skipped": skipped,
			"reason":  reason,
		},
	}, nil
}

// OpsAutofimFail reagenda registro apos falha.
func OpsAutofimFail(ctx context.Context, payload map[string]any) (map[string]any, error) {
	id := intVal(payload, "id")
	lockToken := strings.TrimSpace(strVal(payload, "lock_token"))
	if id <= 0 {
		return nil, fmt.Errorf("id obrigatório")
	}
	if lockToken == "" {
		return nil, fmt.Errorf("lock_token obrigatório")
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	row, err := getAutofimQueueByID(ctx, db, id)
	if err != nil {
		return nil, fmt.Errorf("registro não encontrado")
	}
	if strVal(row, "status") != "PROCESSANDO" {
		return nil, fmt.Errorf("status inválido para fail")
	}
	if strVal(row, "lock_token") != lockToken {
		return nil, fmt.Errorf("lock_token inválido")
	}

	nextTentativas := intVal(row, "tentativas") + 1
	delaySec := intVal(payload, "retry_delay_seconds")
	if delaySec <= 0 {
		switch nextTentativas {
		case 2:
			delaySec = 30
		case 3:
			delaySec = 60
		default:
			if nextTentativas >= 4 {
				delaySec = 120
			} else {
				delaySec = 15
			}
		}
	}

	rodarEm := time.Now().UnixMilli() + int64(delaySec*1000)
	_, err = db.ExecContext(ctx, `
UPDATE ops_bot_finalizaeventoauto SET
    status = 'PENDENTE', tentativas = $2, erro = $3, lock_token = '', lock_at = NULL,
    rodar_em = $4, updated_at = NOW()
WHERE id = $1`, id, nextTentativas, strVal(payload, "erro"), rodarEm)
	if err != nil {
		return nil, err
	}

	updated, err := getAutofimQueueByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"dados": map[string]any{
			"ok":         true,
			"id":         id,
			"status":     strVal(updated, "status"),
			"tentativas": intVal(updated, "tentativas"),
			"rodar_em":   intVal(updated, "rodar_em"),
		},
	}, nil
}

// OpsAlarmEventsUpsert insere evento de alarme com deduplicacao (webhook).
func OpsAlarmEventsUpsert(ctx context.Context, payload map[string]any) (int, error) {
	idEvento := strVal(payload, "idEvento")
	if idEvento == "" {
		idEvento = strVal(payload, "id_evento")
	}

	db, err := DB()
	if err != nil {
		return 0, err
	}

	if idEvento != "" {
		var existID int
		err = db.QueryRowContext(ctx, `SELECT id FROM ops_alarm_events WHERE id_evento = $1 LIMIT 1`, idEvento).Scan(&existID)
		if err == nil {
			return existID, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	idDisp := strVal(payload, "idDispositivo")
	codigo := strVal(payload, "codigo")
	particao := strVal(payload, "particao")
	zonaUser := strVal(payload, "zonaUser")
	if idDisp != "" && codigo != "" {
		var dupID int
		err = db.QueryRowContext(ctx, `
SELECT id FROM ops_alarm_events
WHERE id_dispositivo = $1 AND codigo = $2 AND COALESCE(particao,'') = $3 AND COALESCE(zona_user,'') = $4
  AND created_at > NOW() - INTERVAL '10 seconds'
LIMIT 1`, idDisp, codigo, particao, zonaUser).Scan(&dupID)
		if err == nil {
			return dupID, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	img := strVal(payload, "img")
	emailCliente := strVal(payload, "emailCliente")
	codigoBenuvem := strVal(payload, "codigoBenuvem")
	cameraAtiva := strVal(payload, "carmeraAtiva")
	if cameraAtiva == "" {
		cameraAtiva = strVal(payload, "cameraAtiva")
	}
	if cameraAtiva == "" {
		cameraAtiva = "N"
	}
	usaConfvision := strVal(payload, "usaConfvision")
	if usaConfvision == "" {
		usaConfvision = strVal(payload, "usa_confvision")
	}
	if usaConfvision == "" {
		usaConfvision = "N"
	}
	provedorVideo := strVal(payload, "provedorVideo")
	if provedorVideo == "" {
		provedorVideo = strVal(payload, "provedor_video")
	}
	if provedorVideo == "" {
		provedorVideo = "nenhum"
	}

	var newID int
	err = db.QueryRowContext(ctx, `
INSERT INTO ops_alarm_events (
    id_evento, codigo, particao, zona_user, nivel, data_entrada, img,
    id_processo, id_dispositivo, id_cliente, nome_cliente, email_cliente,
    cti_grupo, cti_descricao, id_franqueado, codigo_benuvem, conta,
    camera_ativa, usa_confvision, provedor_video
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
RETURNING id`,
		idEvento, codigo, particao, zonaUser,
		strVal(payload, "nivel"), strVal(payload, "dataEntrada"), img,
		strVal(payload, "idProcesso"), idDisp, strVal(payload, "idCliente"),
		strVal(payload, "nomeCliente"), emailCliente,
		strVal(payload, "ctiGrupo"), strVal(payload, "ctiDescricao"),
		strVal(payload, "idFranqueado"), codigoBenuvem, strVal(payload, "conta"),
		cameraAtiva, usaConfvision, provedorVideo,
	).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// OpsAutofimEnqueue enfileira processo para autofim (ignora codigos 1M01 e 3M02).
func OpsAutofimEnqueue(ctx context.Context, idProcesso, idDispositivo string, alarmEventsID int, codigo string) error {
	codigo = strings.TrimSpace(codigo)
	if codigo == "1M01" || codigo == "3M02" {
		return nil
	}
	idProcesso = strings.TrimSpace(idProcesso)
	if idProcesso == "" {
		return fmt.Errorf("idProcesso obrigatorio")
	}

	db, err := DB()
	if err != nil {
		return err
	}

	agoraMs := time.Now().UnixMilli()
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_bot_finalizaeventoauto (
    id_processo, id_dispositivo, ultimo_evento_ts, status, rodar_em, tentativas,
    lock_token, lock_at, motivo_final, acao_final, erro, payload_resultado, updated_at, alarm_events_id
) VALUES ($1,$2,$3,'PENDENTE',$3,0,'',NOW(),'','','','{}'::jsonb,NOW(),$4)
ON CONFLICT (id_processo) WHERE id_processo IS NOT NULL AND id_processo <> ''
DO UPDATE SET
    id_dispositivo = EXCLUDED.id_dispositivo,
    ultimo_evento_ts = EXCLUDED.ultimo_evento_ts,
    status = 'PENDENTE',
    rodar_em = EXCLUDED.rodar_em,
    tentativas = 0,
    lock_token = '',
    lock_at = NOW(),
    motivo_final = '',
    acao_final = '',
    erro = '',
    payload_resultado = '{}'::jsonb,
    updated_at = NOW(),
    alarm_events_id = EXCLUDED.alarm_events_id`,
		idProcesso, idDispositivo, agoraMs, alarmEventsID)
	return err
}

func opsProcessoEnd(ctx context.Context, idProcesso, motivo, perfil string) (map[string]any, string, error) {
	descricao := fmt.Sprintf("[AUTO] %s [%s]", motivo, perfil)

	procDados, err := opsGetDadosProcessoByID(ctx, idProcesso)
	if err != nil {
		return nil, "", err
	}

	dataAtenFim := ""
	if procDados != nil {
		if v, ok := procDados["dataAtenFim"].(string); ok {
			dataAtenFim = strings.TrimSpace(v)
		}
	}

	if dataAtenFim != "" && dataAtenFim != "01/01/0001 00:00:00" {
		return map[string]any{"dados": []any{}}, "JA_FINALIZADO", nil
	}

	base := opsEnv("OPS_CONFMONIT_BASE", "http://185.130.61.4:2010")
	urlFinalizar := opsEnv("OPS_FINALIZAR_PROCESSO_URL", base+"/v4/terminal/finalizarProcesso")

	token, err := opsWebLogarToken(ctx)
	if err != nil {
		return nil, "", err
	}

	body, status, err := cvgHTTPJSON(ctx, http.MethodPost, urlFinalizar, map[string]any{
		"idProcesso": idProcesso,
		"idCliente":  "0ROBOAUTO",
		"descricao":  descricao,
		"nome":       "ROBO AUTO",
	}, map[string]string{"Authorization": "Bearer " + token})
	if err != nil {
		return nil, "", err
	}
	if status < 200 || status >= 300 {
		return nil, "", fmt.Errorf("finalizarProcesso status=%d", status)
	}

	dados, _ := cvgExtractDados(body)
	ret := map[string]any{"dados": dados}
	acao := "FINALIZOU"
	if dados == nil || len(dados) == 0 {
		acao = "JA_FINALIZADO"
		ret["dados"] = []any{}
	}
	return ret, acao, nil
}

func opsGetDadosProcessoByID(ctx context.Context, idProcesso string) (map[string]any, error) {
	base := opsEnv("OPS_CONFMONIT_BASE", "http://185.130.61.4:2010")
	urlProc := opsEnv("OPS_GET_PROCESSO_URL", base+"/v4/terminal/getDadosProcessoById")

	token, err := opsWebLogarToken(ctx)
	if err != nil {
		return nil, err
	}

	body, status, err := cvgHTTPJSON(ctx, http.MethodPost, urlProc, map[string]any{
		"idProcesso": idProcesso,
	}, map[string]string{"Authorization": "Bearer " + token})
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("getDadosProcessoById status=%d", status)
	}
	return cvgExtractDados(body)
}

func opsWebLogarToken(ctx context.Context) (string, error) {
	return cvgWebLogarToken(ctx)
}

func opsEnv(key, fallback string) string {
	return cvgEnv(key, fallback)
}

func opsHistAgg(ctx context.Context, db *sql.DB, idDispositivo string) (map[string]any, error) {
	row := db.QueryRowContext(ctx, `
WITH ultimos AS (
  SELECT e.id_processo, e.cti_grupo, EXTRACT(EPOCH FROM e.created_at)::bigint AS ts_s
  FROM ops_alarm_events e
  WHERE e.id_dispositivo = $1
  ORDER BY e.created_at DESC
  LIMIT 150
),
hist AS (
  SELECT
    COALESCE(COUNT(*) FILTER (WHERE cti_grupo = 'ALARME'), 0) AS qtd_alarme_hist,
    COALESCE(COUNT(*) FILTER (WHERE cti_grupo = 'RESTAURE'), 0) AS qtd_restaure_hist
  FROM ultimos
),
janela3m AS (
  SELECT * FROM ultimos WHERE ts_s >= (EXTRACT(EPOCH FROM NOW())::bigint - 180)
),
proc_flags AS (
  SELECT id_processo,
    MAX(CASE WHEN cti_grupo = 'ALARME' THEN 1 ELSE 0 END) AS tem_alarme,
    MAX(CASE WHEN cti_grupo = 'RESTAURE' THEN 1 ELSE 0 END) AS tem_restaure
  FROM janela3m GROUP BY id_processo
),
ciclos3m AS (
  SELECT COALESCE(COUNT(*), 0)::integer AS qtd_ciclos_3m_diff_proc
  FROM proc_flags WHERE tem_alarme = 1 AND tem_restaure = 1
)
SELECT
  COALESCE(h.qtd_alarme_hist, 0),
  COALESCE(h.qtd_restaure_hist, 0),
  COALESCE(c.qtd_ciclos_3m_diff_proc, 0),
  CASE
    WHEN COALESCE(h.qtd_alarme_hist, 0) >= 40
     AND COALESCE(h.qtd_restaure_hist, 0) >= (COALESCE(h.qtd_alarme_hist, 0) * 0.7)
    THEN 'PORTAO' ELSE 'NORMAL'
  END
FROM hist h CROSS JOIN ciclos3m c`, idDispositivo)

	var qtdAlarme, qtdRestaure, qtdCiclos int
	var perfil string
	if err := row.Scan(&qtdAlarme, &qtdRestaure, &qtdCiclos, &perfil); err != nil {
		return map[string]any{
			"qtd_alarme_hist":           0,
			"qtd_restaure_hist":         0,
			"qtd_ciclos_3m_diff_proc":   0,
			"perfil_local":              "NORMAL",
		}, nil
	}
	if perfil == "" {
		perfil = "NORMAL"
	}
	return map[string]any{
		"qtd_alarme_hist":           qtdAlarme,
		"qtd_restaure_hist":         qtdRestaure,
		"qtd_ciclos_3m_diff_proc":   qtdCiclos,
		"perfil_local":              perfil,
	}, nil
}

func getOpsAlarmEventByID(ctx context.Context, db *sql.DB, id int) (map[string]any, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, created_at, id_evento, codigo, particao, zona_user, nivel, data_entrada, img,
       id_processo, id_dispositivo, id_cliente, nome_cliente, email_cliente,
       cti_grupo, cti_descricao, id_franqueado, codigo_benuvem, conta, camera_ativa,
       usa_confvision, provedor_video
FROM ops_alarm_events WHERE id = $1`, id)
	return scanOpsAlarmEventRow(row)
}

func listOpsAlarmEventsByProcesso(ctx context.Context, db *sql.DB, idProcesso string) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, id_evento, codigo, particao, zona_user, nivel, data_entrada, img,
       id_processo, id_dispositivo, id_cliente, nome_cliente, email_cliente,
       cti_grupo, cti_descricao, id_franqueado, codigo_benuvem, conta, camera_ativa,
       usa_confvision, provedor_video
FROM ops_alarm_events WHERE id_processo = $1 ORDER BY created_at ASC`, idProcesso)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		evt, err := scanOpsAlarmEventRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, evt)
	}
	return out, rows.Err()
}

func scanOpsAlarmEventRow(row *sql.Row) (map[string]any, error) {
	var id int
	var createdAt time.Time
	var idEvento, codigo, particao, zonaUser, nivel, dataEntrada, img sql.NullString
	var idProcesso, idDispositivo, idCliente, nomeCliente, emailCliente sql.NullString
	var ctiGrupo, ctiDescricao, idFranqueado, codigoBenuvem, conta sql.NullString
	var cameraAtiva, usaConfvision, provedorVideo sql.NullString

	err := row.Scan(
		&id, &createdAt, &idEvento, &codigo, &particao, &zonaUser, &nivel, &dataEntrada, &img,
		&idProcesso, &idDispositivo, &idCliente, &nomeCliente, &emailCliente,
		&ctiGrupo, &ctiDescricao, &idFranqueado, &codigoBenuvem, &conta, &cameraAtiva,
		&usaConfvision, &provedorVideo,
	)
	if err != nil {
		return nil, err
	}
	return opsAlarmEventMap(id, createdAt, idEvento, codigo, particao, zonaUser, nivel, dataEntrada, img,
		idProcesso, idDispositivo, idCliente, nomeCliente, emailCliente,
		ctiGrupo, ctiDescricao, idFranqueado, codigoBenuvem, conta, cameraAtiva, usaConfvision, provedorVideo), nil
}

func scanOpsAlarmEventRows(rows *sql.Rows) (map[string]any, error) {
	var id int
	var createdAt time.Time
	var idEvento, codigo, particao, zonaUser, nivel, dataEntrada, img sql.NullString
	var idProcesso, idDispositivo, idCliente, nomeCliente, emailCliente sql.NullString
	var ctiGrupo, ctiDescricao, idFranqueado, codigoBenuvem, conta sql.NullString
	var cameraAtiva, usaConfvision, provedorVideo sql.NullString

	err := rows.Scan(
		&id, &createdAt, &idEvento, &codigo, &particao, &zonaUser, &nivel, &dataEntrada, &img,
		&idProcesso, &idDispositivo, &idCliente, &nomeCliente, &emailCliente,
		&ctiGrupo, &ctiDescricao, &idFranqueado, &codigoBenuvem, &conta, &cameraAtiva,
		&usaConfvision, &provedorVideo,
	)
	if err != nil {
		return nil, err
	}
	return opsAlarmEventMap(id, createdAt, idEvento, codigo, particao, zonaUser, nivel, dataEntrada, img,
		idProcesso, idDispositivo, idCliente, nomeCliente, emailCliente,
		ctiGrupo, ctiDescricao, idFranqueado, codigoBenuvem, conta, cameraAtiva, usaConfvision, provedorVideo), nil
}

func opsAlarmEventMap(
	id int, createdAt time.Time,
	idEvento, codigo, particao, zonaUser, nivel, dataEntrada, img sql.NullString,
	idProcesso, idDispositivo, idCliente, nomeCliente, emailCliente sql.NullString,
	ctiGrupo, ctiDescricao, idFranqueado, codigoBenuvem, conta sql.NullString,
	cameraAtiva, usaConfvision, provedorVideo sql.NullString,
) map[string]any {
	return map[string]any{
		"id":             id,
		"created_at":     createdAt.UnixMilli(),
		"idEvento":       nullStr(idEvento),
		"codigo":         nullStr(codigo),
		"particao":       nullStr(particao),
		"zonaUser":       nullStr(zonaUser),
		"nivel":          nullStr(nivel),
		"dataEntrada":    nullStr(dataEntrada),
		"img":            nullStr(img),
		"idProcesso":     nullStr(idProcesso),
		"idDispositivo":  nullStr(idDispositivo),
		"idCliente":      nullStr(idCliente),
		"nomeCliente":    nullStr(nomeCliente),
		"emailCliente":   nullStr(emailCliente),
		"ctiGrupo":       nullStr(ctiGrupo),
		"ctiDescricao":   nullStr(ctiDescricao),
		"idFranqueado":   nullStr(idFranqueado),
		"codigoBenuvem":  nullStr(codigoBenuvem),
		"conta":          nullStr(conta),
		"carmeraAtiva":   nullStr(cameraAtiva),
		"usa_confvision": nullStr(usaConfvision),
		"provedor_video": nullStr(provedorVideo),
	}
}

func scanAutofimQueueRow(rows *sql.Rows) (map[string]any, error) {
	var id, tentativas, alarmEventsID sql.NullInt64
	var idProcesso, idDispositivo, status sql.NullString
	var rodarEm, ultimoEventoTs sql.NullInt64
	var updatedAt sql.NullTime

	if err := rows.Scan(&id, &idProcesso, &idDispositivo, &status, &tentativas, &alarmEventsID, &rodarEm, &ultimoEventoTs, &updatedAt); err != nil {
		return nil, err
	}
	return autofimQueueMap(id, tentativas, alarmEventsID, rodarEm, ultimoEventoTs, idProcesso, idDispositivo, status, updatedAt), nil
}

func autofimQueueMap(
	id, tentativas, alarmEventsID, rodarEm, ultimoEventoTs sql.NullInt64,
	idProcesso, idDispositivo, status sql.NullString,
	updatedAt sql.NullTime,
) map[string]any {
	m := map[string]any{
		"id":               nullInt(id),
		"idProcesso":       nullStr(idProcesso),
		"idDispositivo":    nullStr(idDispositivo),
		"status":           nullStr(status),
		"tentativas":       nullInt(tentativas),
		"alarm_events_id":  nullInt(alarmEventsID),
		"rodar_em":         nullInt(rodarEm),
		"ultimo_evento_ts": nullInt(ultimoEventoTs),
	}
	if updatedAt.Valid {
		m["updated_at"] = updatedAt.Time.UTC().Format(time.RFC3339)
	}
	return m
}

func getAutofimQueueByID(ctx context.Context, db *sql.DB, id int) (map[string]any, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, id_processo, id_dispositivo, status, tentativas, alarm_events_id, rodar_em, ultimo_evento_ts, updated_at, lock_token
FROM ops_bot_finalizaeventoauto WHERE id = $1`, id)
	return scanAutofimQueueRowSingle(row)
}

func getAutofimQueueByIDTx(ctx context.Context, tx *sql.Tx, id int) (map[string]any, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id, id_processo, id_dispositivo, status, tentativas, alarm_events_id, rodar_em, ultimo_evento_ts, updated_at, lock_token
FROM ops_bot_finalizaeventoauto WHERE id = $1`, id)
	return scanAutofimQueueRowSingle(row)
}

func scanAutofimQueueRowSingle(row *sql.Row) (map[string]any, error) {
	var id, tentativas, alarmEventsID sql.NullInt64
	var idProcesso, idDispositivo, status, lockToken sql.NullString
	var rodarEm, ultimoEventoTs sql.NullInt64
	var updatedAt sql.NullTime

	if err := row.Scan(&id, &idProcesso, &idDispositivo, &status, &tentativas, &alarmEventsID, &rodarEm, &ultimoEventoTs, &updatedAt, &lockToken); err != nil {
		return nil, err
	}
	m := autofimQueueMap(id, tentativas, alarmEventsID, rodarEm, ultimoEventoTs, idProcesso, idDispositivo, status, updatedAt)
	m["lock_token"] = nullStr(lockToken)
	return m, nil
}

func boolDefaultMap(m map[string]any, key string, def bool) bool {
	if v := boolVal(m, key); v != nil {
		return *v
	}
	return def
}

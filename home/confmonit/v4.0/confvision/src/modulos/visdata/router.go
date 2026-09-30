package visdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Dispatch(ctx context.Context, method, pathWithQuery string, body io.Reader) (int, []byte, error) {
	if !Ready() {
		return http.StatusServiceUnavailable, []byte(`{"erro":"postgres indisponivel"}`), nil
	}

	u, err := url.Parse(pathWithQuery)
	if err != nil {
		return http.StatusBadRequest, []byte(`{"erro":"path invalido"}`), nil
	}
	path := u.Path
	q := u.Query()

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = io.ReadAll(body)
	}
	var payload map[string]any
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}

	switch {
	case method == http.MethodGet && path == "/vis_health":
		if !Ready() {
			return http.StatusServiceUnavailable, []byte(`{"status":"postgres indisponivel"}`), nil
		}
		return okJSON(map[string]any{"status": "ok", "enabled": true, "postgres": "ok"})

	case method == http.MethodGet && path == "/vis_camera_query_ativas":
		list, err := ListCamerasAnaliticas(ctx, q.Get("worker_id"), parseIntQuery(q.Get("vis_mediamtx_node_id")))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_camera_sync_ativas":
		cams, err := ListCamerasAnaliticas(ctx, q.Get("worker_id"), parseIntQuery(q.Get("vis_mediamtx_node_id")))
		if err != nil {
			return errJSON(err)
		}
		areas, err := ListAreasAtivas(ctx)
		if err != nil {
			return errJSON(err)
		}
		configVersion := buildConfigVersion(cams, areas)
		sinceVersion := strings.TrimSpace(q.Get("since_version"))
		if sinceVersion != "" && sinceVersion == configVersion {
			return okJSON(map[string]any{
				"unchanged":      true,
				"config_version": configVersion,
			})
		}
		out := map[string]any{
			"cameras": cams, "areas": areas,
			"config_version": configVersion,
			"unchanged":      false,
			"gravacao":       []any{},
		}
		if strings.EqualFold(q.Get("include_gravacao"), "true") {
			grav, err := ListGravacaoCameras(ctx, q.Get("worker_id"), parseIntQuery(q.Get("vis_mediamtx_node_id")))
			if err != nil {
				return errJSON(err)
			}
			out["gravacao"] = grav
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/vis_camera_area_query_ativas":
		list, err := ListAreasAtivas(ctx)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_camera_query_gravacao_ativas":
		list, err := ListGravacaoCameras(ctx, q.Get("worker_id"), parseIntQuery(q.Get("vis_mediamtx_node_id")))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodPost && path == "/vis_worker_ping":
		var in WorkerPingInput
		if err := json.Unmarshal(bodyBytes, &in); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		out, err := UpsertWorkerPing(ctx, in)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_camera_stream_reactivate":
		cameraID := parseIntQuery(q.Get("camera_id"))
		if cameraID <= 0 {
			if v, ok := payload["camera_id"]; ok {
				cameraID = intVal(map[string]any{"camera_id": v}, "camera_id")
			}
		}
		if cameraID <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"camera_id obrigatorio"}`), nil
		}
		out, err := ReactivateCameraStream(ctx, cameraID)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && strings.HasPrefix(path, "/vis_camera/rtmp_auth/"):
		id := pathID(path, "/vis_camera/rtmp_auth/")
		out, err := GetCameraRTMPAuth(ctx, id)
		if err != nil {
			return http.StatusNotFound, []byte(`{"erro":"camera nao encontrada"}`), nil
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/vis_camera_rtmp_auth_sync":
		nodeID := parseIntQuery(q.Get("vis_mediamtx_node_id"))
		list, err := ListCamerasRTMPAuthSync(ctx, nodeID)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list, "total": len(list)})

	case method == http.MethodGet && path == "/vis_camera_by_franqueado":
		list, err := ListCamerasByFranqueado(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_camera_by_cliente":
		list, err := ListCamerasByCliente(ctx, q.Get("id_cliente"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_camera_by_setor":
		cam, err := GetCameraBySetor(ctx, q.Get("id_dispositivo"), q.Get("particao"), q.Get("zonauser"), q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		if cam == nil {
			return okJSON(map[string]any{"dados": nil})
		}
		return okJSON(map[string]any{"dados": cam})

	case method == http.MethodGet && path == "/vis_camera":
		// admin list all — simplified: by franqueado optional
		if idFra := q.Get("id_franqueado"); idFra != "" {
			list, err := ListCamerasByFranqueado(ctx, idFra)
			if err != nil {
				return errJSON(err)
			}
			return okJSON(list)
		}
		if q.Get("all") == "1" || q.Get("all") == "true" {
			list, err := ListAllCameras(ctx)
			if err != nil {
				return errJSON(err)
			}
			return okJSON(list)
		}
		return http.StatusBadRequest, []byte(`{"erro":"id_franqueado obrigatorio"}`), nil

	case method == http.MethodPost && path == "/vis_camera":
		out, err := CreateCamera(ctx, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && strings.HasPrefix(path, "/vis_camera/mediamtx/"):
		id := pathID(path, "/vis_camera/mediamtx/")
		atribuir := q.Get("atribuir_se_ausente") != "false"
		out, err := ResolveMediamtxForCamera(ctx, id, atribuir)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && strings.HasPrefix(path, "/vis_camera/") && !strings.Contains(path, "/gravacao/") && !strings.Contains(path, "/rtmp_auth/") && !strings.Contains(path, "/bloquear/") && !strings.Contains(path, "/analitico/") && !strings.Contains(path, "/licenca/") && !strings.Contains(path, "/mediamtx/") && !strings.Contains(path, "/snapshot/"):
		id := pathID(path, "/vis_camera/")
		out, err := GetCameraByID(ctx, id)
		if err != nil {
			return http.StatusNotFound, []byte(`{"erro":"camera nao encontrada"}`), nil
		}
		return okJSON(out)

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_camera/snapshot/"):
		id := pathID(path, "/vis_camera/snapshot/")
		if err := UpdateCameraSnapshot(ctx, id, strVal(payload, "snapshot_url")); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"id": id})

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_camera/"):
		id := pathID(path, "/vis_camera/")
		out, err := UpdateCamera(ctx, id, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodDelete && strings.HasPrefix(path, "/vis_camera/"):
		id := pathID(path, "/vis_camera/")
		if err := SoftDeleteCamera(ctx, id); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"id": id})

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/bloquear/"):
		id := pathID(path, "/vis_camera/bloquear/")
		bloq := boolDefault(payload, "bloqueado", true)
		out, err := SetCameraBloqueado(ctx, id, bloq)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/analitico/pausar/"):
		id := pathID(path, "/vis_camera/analitico/pausar/")
		// UI e API Xano enviam "pausado"; fallback legado "analitico_pausado".
		pausa := boolDefault(payload, "pausado", boolDefault(payload, "analitico_pausado", false))
		out, err := SetAnaliticoPausado(ctx, id, pausa)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/licenca/liberar/"):
		id := pathID(path, "/vis_camera/licenca/liberar/")
		out, err := LiberarLicencaCamera(ctx, id)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/gravacao/ativar/"):
		id := pathID(path, "/vis_camera/gravacao/ativar/")
		out, err := AtivarGravacaoCamera(ctx, id, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/gravacao/desativar/"):
		id := pathID(path, "/vis_camera/gravacao/desativar/")
		out, err := DesativarGravacaoCamera(ctx, id)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/gravacao/flush/ack/"):
		id := pathID(path, "/vis_camera/gravacao/flush/ack/")
		if err := AckGravacaoFlush(ctx, id); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_camera/gravacao/flush/"):
		id := pathID(path, "/vis_camera/gravacao/flush/")
		if err := RequestGravacaoFlush(ctx, id); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodGet && path == "/vis_camera_area_by_camera":
		list, err := ListAreasByCamera(ctx, parseIntQuery(q.Get("vis_camera_id")))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodPost && path == "/vis_camera_area":
		out, err := CreateArea(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_camera_area/"):
		id := pathID(path, "/vis_camera_area/")
		out, err := UpdateArea(ctx, id, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodDelete && strings.HasPrefix(path, "/vis_camera_area/"):
		id := pathID(path, "/vis_camera_area/")
		if err := DeleteArea(ctx, id); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"id": id})

	case method == http.MethodGet && path == "/vis_licenca_by_franqueado":
		out, err := ListLicencasByFranqueado(ctx, q.Get("id_franqueado"), q.Get("status"), q.Get("unidade"))
		if err != nil {
			return errJSON(err)
		}
		raw, _ := json.Marshal(out)
		return http.StatusOK, raw, nil

	case method == http.MethodPost && path == "/vis_licenca":
		out, err := CreateLicenca(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_evento":
		out, err := CreateEvento(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_evento_disparo_sensor":
		out, err := DisparoSensorEvento(ctx, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		raw, _ := json.Marshal(out)
		return http.StatusOK, raw, nil

	case method == http.MethodPost && path == "/vis_evento_demo_moni":
		out, err := UpsertEventoDemoMoni(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_evento_finalizar":
		out, err := FinalizarEvento(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		raw, _ := json.Marshal(out)
		return http.StatusOK, raw, nil

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_evento/"):
		id := pathID(path, "/vis_evento/")
		out, err := UpdateEvento(ctx, id, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_evento_clip":
		out, err := CreateEventoClip(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && strings.HasSuffix(path, "/clips"):
		id := pathID(strings.TrimSuffix(path, "/clips"), "/vis_evento/")
		out, err := GetEventoClips(ctx, id)
		if err != nil {
			return errJSON(err)
		}
		raw, _ := json.Marshal(out)
		return http.StatusOK, raw, nil

	case method == http.MethodGet && (path == "/vis_evento_by_franqueado_page" || path == "/vis_evento_by_cliente_page"):
		page := parseIntQuery(q.Get("page"))
		idFra := q.Get("id_franqueado")
		idCli := q.Get("id_cliente")
		if path == "/vis_evento_by_cliente_page" {
			idFra = ""
		}
		out, err := ListEventosPage(ctx, idFra, idCli, page, q.Get("data_de"), q.Get("data_ate"))
		if err != nil {
			return errJSON(err)
		}
		raw, _ := json.Marshal(out)
		return http.StatusOK, raw, nil

	case method == http.MethodGet && path == "/vis_evento_query_sensor_pendentes":
		limit := parseIntQuery(q.Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		list, err := ListEventosSensorPendentes(ctx, limit)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_evento_by_contexto":
		list, err := ListEventosByContexto(ctx,
			q.Get("id_processo"), q.Get("id_cliente"), q.Get("id_dispositivo"),
			q.Get("id_franqueado"), q.Get("data_inicio"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_evento_by_processo_setor":
		ev, cam, err := GetEventoByProcessoSetor(ctx,
			q.Get("id_processo"), q.Get("id_dispositivo"), q.Get("particao"),
			q.Get("zonauser"), q.Get("id_evento"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": ev, "camera": cam})

	case method == http.MethodGet && path == "/vis_evento_ultimos25_setor":
		list, cam, err := ListEventosUltimos25Setor(ctx, q.Get("id_dispositivo"), q.Get("particao"), q.Get("zonauser"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": list, "camera": cam})

	case method == http.MethodGet && path == "/vis_evento_ultimos25_dispositivo":
		limit := parseIntQuery(q.Get("limit"))
		if limit <= 0 {
			limit = 25
		}
		list, err := ListEventosUltimos25Dispositivo(ctx, q.Get("id_dispositivo"), limit)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{
			"dados":  list,
			"total":  len(list),
			"offset": 0,
			"limit":  limit,
		})

	case method == http.MethodGet && path == "/vis_gravacao_storage_by_franqueado":
		mask := !strings.EqualFold(q.Get("credenciais_completas"), "true")
		list, err := ListGravacaoStorageByFranqueado(ctx, q.Get("id_franqueado"), q.Get("status"), mask)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/vis_gravacao_storage_credenciais_by_franqueado":
		list, err := ListGravacaoStorageByFranqueado(ctx, q.Get("id_franqueado"), "ativo", false)
		if err != nil {
			return errJSON(err)
		}
		if len(list) == 0 {
			return http.StatusNotFound, []byte(`{"erro":"storage nao encontrado"}`), nil
		}
		return okJSON(list[0])

	case method == http.MethodPost && path == "/vis_gravacao_storage":
		out, err := CreateGravacaoStorage(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_gravacao_storage/"):
		id := pathID(path, "/vis_gravacao_storage/")
		out, err := UpdateGravacaoStorage(ctx, id, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_gravacao_segmento":
		out, err := PostGravacaoSegmento(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && (path == "/vis_gravacao_segmento_by_camera" || path == "/vis_gravacao_segmento_by_franqueado"):
		page := parseIntQuery(q.Get("page"))
		camID := parseIntQuery(q.Get("vis_camera_id"))
		idFra := q.Get("id_franqueado")
		if path == "/vis_gravacao_segmento_by_camera" && camID <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"vis_camera_id obrigatorio"}`), nil
		}
		out, err := ListGravacaoSegmentos(ctx, idFra, camID, page, q.Get("de"), q.Get("ate"))
		if err != nil {
			return errJSON(err)
		}
		raw, _ := json.Marshal(out)
		return http.StatusOK, raw, nil

	case method == http.MethodGet && path == "/vis_mediamtx_node":
		list, err := ListMediamtxNodes(ctx)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

	case method == http.MethodGet && path == "/cvg_grade_by_cliente":
		return CvgGradeByClienteGET(ctx, q)
	case method == http.MethodPut && path == "/cvg_grade_by_cliente":
		return CvgGradeByClientePUT(ctx, payload)
	case method == http.MethodGet && path == "/cvg_grade_list":
		return CvgGradeListGET(ctx, q)
	case method == http.MethodPatch && path == "/cvg_grade_escopo_ativa":
		return CvgGradeEscopoAtivaPATCH(ctx, payload)
	case method == http.MethodPost && path == "/cvg_worker_tick":
		workerKey := ""
		if payload != nil {
			workerKey = strVal(payload, "worker_key")
		}
		return CvgWorkerTickPOST(ctx, payload, workerKey)

	case method == http.MethodGet && path == "/vis_integracao_by_franqueado":
		out, err := ListIntegracoesByFranqueado(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/vis_integracao_franqueado":
		out, err := GetIntegracaoFranqueado(ctx, q.Get("id_franqueado"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPut && path == "/vis_integracao_franqueado":
		out, err := SaveIntegracaoFranqueado(ctx, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_integracao_franqueado/test":
		out, err := TestIntegracaoFranqueado(ctx, strVal(payload, "id_franqueado"), payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_integracao":
		out, err := CreateIntegracao(ctx, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_integracao/"):
		id := pathID(path, "/vis_integracao/")
		out, err := UpdateIntegracao(ctx, id, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodDelete && strings.HasPrefix(path, "/vis_integracao/"):
		id := pathID(path, "/vis_integracao/")
		if err := DeleteIntegracao(ctx, id); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && strings.HasPrefix(path, "/vis_integracao_test/"):
		id := pathID(path, "/vis_integracao_test/")
		out, err := TestIntegracaoMoni(ctx, id, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/vis_integracao_log":
		limit := parseIntQuery(q.Get("limit"))
		offset := parseIntQuery(q.Get("offset"))
		out, err := ListIntegracaoLog(ctx, q.Get("id_franqueado"), limit, offset,
			q.Get("data_de"), q.Get("data_ate"), q.Get("cliente_nome"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_integracao/moni/inbound":
		return MoniInboundPOST(ctx, payload)

	case method == http.MethodGet && path == "/ops/arme/agenda":
		dia := parseIntQuery(q.Get("DiaSemana"))
		if dia == 0 {
			dia = parseIntQuery(q.Get("dia_semana"))
		}
		out, err := OpsArmeAgendaGET(ctx, dia)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/ops/atendimento/politica":
		out, err := GetPoliticaAtendimento(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPut && path == "/ops/atendimento/politica":
		var p PoliticaAtendimento
		if err := json.Unmarshal(bodyBytes, &p); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		if err := SavePoliticaAtendimento(ctx, p); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodGet && path == "/ops/atendimento/resolve":
		idFra := q.Get("id_franqueado")
		idCli := q.Get("id_cliente")
		recurso := q.Get("recurso")
		ctiGrupo := q.Get("cti_grupo")
		ok, failOpen, err := ResolveRecursoAtivoComHorarioCompat(ctx, idFra, idCli, recurso, ctiGrupo)
		if err != nil {
			return errJSON(err)
		}
		bloq, _ := IsIABloqueado(ctx, idFra, idCli)
		return okJSON(map[string]any{
			"ativo": ok, "ia_bloqueado": bloq, "recurso": recurso, "fail_open": failOpen,
			"cti_grupo": ctiGrupo, "emergencia": IsGrupoEmergenciaAtendimento(ctiGrupo),
		})

	case method == http.MethodGet && path == "/ops/atendimento/grade":
		out, err := ListGradeSlots(ctx, q.Get("id_franqueado"), q.Get("id_cliente"), q.Get("responsavel"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPut && path == "/ops/atendimento/grade":
		idFra := strVal(payload, "id_franqueado")
		idCli := strVal(payload, "id_cliente")
		resp := strVal(payload, "responsavel")
		var slots []GradeSlot
		if raw, ok := payload["slots"].([]any); ok {
			for _, el := range raw {
				m, ok := el.(map[string]any)
				if !ok {
					continue
				}
				slot := GradeSlot{
					DiasSemana: intSliceFromAny(m["dias_semana"]),
					HoraInicio: strVal(m, "hora_inicio"),
					HoraFim:    strVal(m, "hora_fim"),
				}
				slots = append(slots, slot)
			}
		}
		if err := ReplaceGradeSlots(ctx, idFra, idCli, resp, slots); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && path == "/ops/atendimento/parceiro/envio":
		out, err := ProcessarEnvioParceiro(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/ops/atendimento/ia_bloqueio":
		out, err := ListIABloqueios(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPost && path == "/ops/atendimento/ia_bloqueio":
		idFra := strVal(payload, "id_franqueado")
		idCli := strVal(payload, "id_cliente")
		nome := strVal(payload, "nome_cliente")
		motivo := strVal(payload, "motivo")
		if err := AddIABloqueio(ctx, idFra, idCli, nome, motivo); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodDelete && strings.HasPrefix(path, "/ops/atendimento/ia_bloqueio/"):
		id := pathID(path, "/ops/atendimento/ia_bloqueio/")
		if err := RemoveIABloqueio(ctx, id); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodGet && path == "/ops/atendimento/creditos":
		out, err := ListCreditoSaldos(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodGet && path == "/ops/atendimento/credito/pode_ligar":
		idFra := strings.TrimSpace(q.Get("id_franqueado"))
		idCentral := strings.TrimSpace(q.Get("id_central"))
		ok, saldo, min, err := PodeUsarCanal(ctx, idFra, idCentral, CanalLigacao)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{
			"pode_ligar": ok, "saldo": saldo, "custo_minimo": min, "canal": CanalLigacao, "saldo_unico": true,
		})

	case method == http.MethodGet && path == "/ops/atendimento/credito/resumo":
		r, err := GetCreditoResumo(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": r})

	case method == http.MethodPost && path == "/ops/atendimento/credito/creditar":
		idFra := strVal(payload, "id_franqueado")
		servico := strVal(payload, "servico")
		if servico == "" {
			servico = strVal(payload, "canal")
		}
		tipo := strVal(payload, "tipo")
		if tipo == "" {
			tipo = TipoMovRecarga
		}
		valor := floatVal(payload, "valor")
		var idFat *int64
		if n := intVal(payload, "id_fatura"); n > 0 {
			v := int64(n)
			idFat = &v
		}
		if err := CreditarSaldo(ctx, idFra, servico, tipo, valor, idFat, strVal(payload, "observacao"), strVal(payload, "criado_por")); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && path == "/ops/atendimento/credito/inicializar":
		idFra := strVal(payload, "id_franqueado")
		if err := InicializarSaldoZero(ctx, idFra); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && path == "/ops/atendimento/credito/debitar":
		idFra := strVal(payload, "id_franqueado")
		servico := strVal(payload, "servico")
		if servico == "" {
			servico = strVal(payload, "canal")
		}
		valor := floatVal(payload, "valor")
		ja, err := DebitarSaldo(ctx, idFra, servico, valor, strVal(payload, "id_processo"), strVal(payload, "observacao"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true, "ja_debitado": ja})

	case method == http.MethodGet && path == "/ops/atendimento/tarifa":
		out, err := GetTarifaOperacional(ctx, q.Get("id_central"), q.Get("canal"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPut && path == "/ops/atendimento/tarifa":
		var t TarifaOperacional
		if err := json.Unmarshal(bodyBytes, &t); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		if err := SaveTarifaOperacional(ctx, t); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodGet && path == "/ops/atendimento/creditos/movimentos":
		out, err := ListCreditoMovimentos(ctx, q.Get("id_franqueado"), q.Get("canal"), q.Get("de"), q.Get("ate"), 100)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodGet && path == "/ops/atendimento/smtp":
		idFra := q.Get("id_franqueado")
		interno := q.Get("interno") == "1" || q.Get("interno") == "true"
		cfg, err := GetSMTPConfig(ctx, idFra, interno)
		if err != nil {
			return errJSON(err)
		}
		if !interno {
			cfg.SmtpSenha = ""
		}
		return okJSON(map[string]any{"dados": cfg})

	case method == http.MethodGet && path == "/ops/atendimento/smtp/status":
		okCfg, err := SMTPConfiguradoAtivo(ctx, q.Get("id_franqueado"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"configurado_ativo": okCfg})

	case method == http.MethodPut && path == "/ops/atendimento/smtp":
		var body struct {
			Config    SMTPConfigEnvio `json:"config"`
			NovaSenha string          `json:"nova_senha"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			var cfg SMTPConfigEnvio
			if err2 := json.Unmarshal(bodyBytes, &cfg); err2 != nil {
				return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
			}
			body.Config = cfg
		}
		if err := SaveSMTPConfig(ctx, body.Config, body.NovaSenha); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && path == "/ops/atendimento/smtp/testar":
		var cfg SMTPConfigEnvio
		if err := json.Unmarshal(bodyBytes, &cfg); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		if strings.TrimSpace(cfg.SmtpSenha) == "" && cfg.IDFranqueado != "" {
			stored, err := GetSMTPConfig(ctx, cfg.IDFranqueado, true)
			if err == nil && stored.SmtpSenha != "" {
				cfg.SmtpSenha = stored.SmtpSenha
			}
		}
		out := TestarSMTP(ctx, cfg)
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPost && path == "/ops/atendimento/smtp/enviar-teste":
		var body struct {
			Config       SMTPConfigEnvio `json:"config"`
			Destinatario string          `json:"destinatario"`
			NovaSenha    string          `json:"nova_senha"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		if strings.TrimSpace(body.NovaSenha) != "" {
			body.Config.SmtpSenha = body.NovaSenha
		} else if strings.TrimSpace(body.Config.SmtpSenha) == "" && body.Config.IDFranqueado != "" {
			stored, err := GetSMTPConfig(ctx, body.Config.IDFranqueado, true)
			if err == nil {
				body.Config.SmtpSenha = stored.SmtpSenha
			}
		}
		corpo := `<p>E-mail de teste enviado pela configuracao SMTP do FranqueadoPro.</p>`
		if err := EnviarEmailSMTP(ctx, body.Config, body.Destinatario, "Teste SMTP FranqueadoPro", corpo); err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPost && path == "/ops/vis_licenca/criar_lote":
		var body struct {
			Licencas []map[string]any `json:"licencas"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		list, err := CreateLicencaLote(ctx, body.Licencas)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"ok": true, "licencas": list, "criadas": len(list)})

	case method == http.MethodPost && path == "/ops/vis_licenca/sync_lote":
		var body struct {
			Licencas []map[string]any `json:"licencas"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			return http.StatusBadRequest, []byte(`{"erro":"json invalido"}`), nil
		}
		n, err := SyncLicencasFromXanoRecords(ctx, body.Licencas)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"ok": true, "sincronizadas": n})

	case method == http.MethodPost && path == "/ops/vis_licenca/sync_franqueado":
		idFra := strVal(payload, "id_franqueado")
		if idFra == "" {
			return http.StatusBadRequest, []byte(`{"erro":"id_franqueado obrigatorio"}`), nil
		}
		// licencas enviadas no body (Xano repassa apos consulta)
		if licRaw, ok := payload["licencas"].([]any); ok && len(licRaw) > 0 {
			records := make([]map[string]any, 0, len(licRaw))
			for _, el := range licRaw {
				if m, ok := el.(map[string]any); ok {
					records = append(records, m)
				}
			}
			n, err := SyncLicencasFromXanoRecords(ctx, records)
			if err != nil {
				return errJSON(err)
			}
			return okJSON(map[string]any{"ok": true, "sincronizadas": n})
		}
		return http.StatusBadRequest, []byte(`{"erro":"licencas obrigatorio"}`), nil

	case method == http.MethodPost && path == "/ops/vis_licenca/ativar_pagamento":
		licID := intVal(payload, "vis_licenca_id")
		if licID == 0 {
			licID = intVal(payload, "vis_licencaId")
		}
		if lic, ok := payload["licenca"].(map[string]any); ok && intFromAny(lic["id"]) > 0 {
			if licID == 0 {
				licID = intFromAny(lic["id"])
			}
		}
		if licID <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"vis_licenca_id obrigatorio"}`), nil
		}
		pago := parseTimeAny(payload["pago_em"])
		if pago.IsZero() {
			pago = time.Now().UTC()
		}
		out, err := AtivarLicencaPagamento(ctx, licID, intVal(payload, "fp_fatura_id"), intVal(payload, "fp_pagamento_id"), pago)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/ops/vis_licenca/estornar_pagamento":
		licID := intVal(payload, "vis_licenca_id")
		if licID == 0 {
			licID = intVal(payload, "vis_licencaId")
		}
		if lic, ok := payload["licenca"].(map[string]any); ok && intFromAny(lic["id"]) > 0 {
			if licID == 0 {
				licID = intFromAny(lic["id"])
			}
		}
		if licID <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"vis_licenca_id obrigatorio"}`), nil
		}
		out, err := EstornarLicencaFatura(ctx, licID, strVal(payload, "observacao"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/vis_capacidade_resumo":
		idFra := q.Get("id_franqueado")
		if idFra == "" {
			return http.StatusBadRequest, []byte(`{"erro":"id_franqueado obrigatorio"}`), nil
		}
		out, err := ResumoCapacidade(ctx, idFra, q.Get("id_central"), q.Get("id_representante"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/vis_capacidade_cotacao":
		qtd := intVal(payload, "quantidade")
		if qtd <= 0 {
			qtd = intVal(payload, "quantidade_contratada")
		}
		out, err := CotacaoCapacidade(ctx, strVal(payload, "id_franqueado"), qtd, strVal(payload, "id_central"), strVal(payload, "id_representante"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/ops/vis_capacidade/reservar_pendente":
		out, err := ReservarCapacidadePendente(ctx, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"ok": true, "contrato": out})

	case method == http.MethodPost && path == "/ops/vis_capacidade/ativar_pagamento":
		contratoID := intVal(payload, "vis_capacidade_contrato_id")
		if contratoID == 0 {
			contratoID = intVal(payload, "contrato_id")
		}
		if contratoID <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"vis_capacidade_contrato_id obrigatorio"}`), nil
		}
		pago := parseTimeAny(payload["pago_em"])
		if pago.IsZero() {
			pago = time.Now().UTC()
		}
		out, err := AtivarCapacidadePagamento(ctx, contratoID, intVal(payload, "fp_fatura_id"), intVal(payload, "fp_pagamento_id"), pago)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/ops/vis_capacidade/estornar_pagamento":
		contratoID := intVal(payload, "vis_capacidade_contrato_id")
		if contratoID == 0 {
			contratoID = intVal(payload, "contrato_id")
		}
		if contratoID <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"vis_capacidade_contrato_id obrigatorio"}`), nil
		}
		out, err := EstornarCapacidadeContrato(ctx, contratoID, strVal(payload, "observacao"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/ops/vis_capacidade/config_salvar":
		out, err := UpsertCapacidadeConfig(ctx, payload)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/ops/vis_capacidade/config_listar":
		out, err := GetCapacidadeConfig(ctx, q.Get("id_central"), q.Get("id_representante"))
		if err != nil {
			return errJSON(err)
		}
		return okJSON(out)

	case method == http.MethodPost && path == "/ops/vis_capacidade/listar_pendentes_renovacao":
		dias := intVal(payload, "dias_antecedencia")
		if dias <= 0 {
			dias = 5
		}
		list, err := ListCapacidadePendentesRenovacao(ctx, dias)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"contratos": list, "total": len(list)})

	case method == http.MethodGet && strings.HasPrefix(path, "/ops/vis_capacidade/contrato/"):
		id := pathID(path, "/ops/vis_capacidade/contrato/")
		if id <= 0 {
			return http.StatusBadRequest, []byte(`{"erro":"contrato_id obrigatorio"}`), nil
		}
		out, err := GetContratoCapacidade(ctx, id)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(out)

	case method == http.MethodGet && path == "/vis_grupo_visualizacao":
		out, err := ListGruposByFranqueado(ctx, q.Get("id_franqueado"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodGet && path == "/vis_grupo_visualizacao/disponiveis":
		out, err := ListGruposDisponiveisCliente(ctx, q.Get("id_franqueado"), q.Get("id_cliente"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodGet && strings.HasPrefix(path, "/vis_grupo_visualizacao/") && strings.HasSuffix(path, "/cameras"):
		grupoID := pathID(path, "/vis_grupo_visualizacao/")
		if grupoID <= 0 {
			return bizErrJSON(fmt.Errorf("grupo invalido"))
		}
		out, err := ListGrupoCameras(ctx, grupoID, q.Get("id_franqueado"), q.Get("id_cliente"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodGet && strings.HasPrefix(path, "/vis_grupo_visualizacao/"):
		grupoID := pathID(path, "/vis_grupo_visualizacao/")
		if grupoID <= 0 {
			return bizErrJSON(fmt.Errorf("grupo invalido"))
		}
		out, err := GetGrupoVisualizacao(ctx, grupoID, q.Get("id_franqueado"), q.Get("id_cliente"))
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPost && path == "/vis_grupo_visualizacao":
		out, err := CreateGrupoVisualizacao(ctx, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPut && strings.HasSuffix(path, "/composicao"):
		grupoID := pathID(strings.TrimSuffix(path, "/composicao"), "/vis_grupo_visualizacao/")
		if grupoID <= 0 {
			return bizErrJSON(fmt.Errorf("grupo invalido"))
		}
		out, err := SaveGrupoComposicao(ctx, grupoID, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodPut && strings.HasSuffix(path, "/reordenar"):
		grupoID := pathID(strings.TrimSuffix(path, "/reordenar"), "/vis_grupo_visualizacao/")
		if grupoID <= 0 {
			return bizErrJSON(fmt.Errorf("grupo invalido"))
		}
		if err := ReordenarGrupoCameras(ctx, grupoID, payload); err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"ok": true})

	case method == http.MethodPut && strings.HasPrefix(path, "/vis_grupo_visualizacao/"):
		grupoID := pathID(path, "/vis_grupo_visualizacao/")
		if grupoID <= 0 {
			return bizErrJSON(fmt.Errorf("grupo invalido"))
		}
		out, err := UpdateGrupoVisualizacao(ctx, grupoID, payload)
		if err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"dados": out})

	case method == http.MethodDelete && strings.HasPrefix(path, "/vis_grupo_visualizacao/"):
		grupoID := pathID(path, "/vis_grupo_visualizacao/")
		if grupoID <= 0 {
			return bizErrJSON(fmt.Errorf("grupo invalido"))
		}
		idFra := strVal(payload, "id_franqueado")
		if idFra == "" {
			idFra = q.Get("id_franqueado")
		}
		if err := DeleteGrupoVisualizacao(ctx, grupoID, idFra); err != nil {
			return bizErrJSON(err)
		}
		return okJSON(map[string]any{"ok": true})
	}

	return 0, nil, ErrNotHandled
}

func pathID(path, prefix string) int {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Trim(s, "/")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	if q := strings.Index(s, "?"); q >= 0 {
		s = s[:q]
	}
	return parseIntQuery(s)
}

func okJSON(v any) (int, []byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return http.StatusInternalServerError, []byte(`{"erro":"marshal falhou"}`), nil
	}
	return http.StatusOK, raw, nil
}

func errJSON(err error) (int, []byte, error) {
	raw, _ := json.Marshal(map[string]any{"erro": err.Error()})
	return http.StatusInternalServerError, raw, nil
}

func bizErrJSON(err error) (int, []byte, error) {
	raw, _ := json.Marshal(map[string]any{"status": err.Error()})
	return http.StatusBadRequest, raw, nil
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprint(err)
}

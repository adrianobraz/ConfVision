package visdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
		return okJSON(map[string]any{"status": "ok", "enabled": true})

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
		out := map[string]any{
			"cameras": cams, "areas": areas,
			"config_version": buildConfigVersion(cams, areas),
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

	case method == http.MethodGet && strings.HasPrefix(path, "/vis_camera/rtmp_auth/"):
		id := pathID(path, "/vis_camera/rtmp_auth/")
		out, err := GetCameraRTMPAuth(ctx, id)
		if err != nil {
			return http.StatusNotFound, []byte(`{"erro":"camera nao encontrada"}`), nil
		}
		return okJSON(out)

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

	case method == http.MethodGet && path == "/vis_camera":
		// admin list all — simplified: by franqueado optional
		if idFra := q.Get("id_franqueado"); idFra != "" {
			list, err := ListCamerasByFranqueado(ctx, idFra)
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
		pausa := boolDefault(payload, "analitico_pausado", true)
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
		list, err := ListEventosSensorPendentes(ctx)
		if err != nil {
			return errJSON(err)
		}
		return okJSON(map[string]any{"dados": list})

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

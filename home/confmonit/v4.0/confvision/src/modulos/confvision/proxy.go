package confvision

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/modulos/visdata"
	"confvision/src/seguranca"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func sessaoCookie(r *http.Request) map[string]string {
	cookie, _ := seguranca.LerCookies(r)
	return cookie
}

func ProxyListarClientes(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	if seguranca.EhCliente(cookie) {
		auxiliar.RespostaAPP(w, []byte(`{"dados":[]}`))
		return
	}
	idFra, err := assertFranqueadoQuery(r, "")
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusForbidden, err)
		return
	}
	raw, _ := json.Marshal(map[string]any{"idFranqueado": idFra})
	r.Body = io.NopCloser(bytes.NewReader(raw))
	u := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)
	proxyConfmonitPost(w, r, u)
}

func ProxyListarDispositivos(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	if seguranca.EhCliente(cookie) {
		raw, _ := json.Marshal(map[string]any{"idCliente": seguranca.IdClienteDoCookie(cookie)})
		r.Body = io.NopCloser(bytes.NewReader(raw))
	}
	u := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)
	proxyConfmonitPost(w, r, u)
}

func ProxyDadosDispositivo(w http.ResponseWriter, r *http.Request) {
	u := fmt.Sprintf("%s/v4/dispositivo/getDadosById", config.ApiUrl)
	proxyConfmonitPost(w, r, u)
}

func ProxyListarDispositivosFranqueado(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	body, _ := io.ReadAll(r.Body)
	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	idCli := seguranca.IdClienteDoCookie(cookie)
	if idCli == "" {
		idCli = seguranca.TextoDoPayload(payload, "idCliente", "id_cliente")
	}
	if seguranca.EhCliente(cookie) || idCli != "" {
		raw, _ := json.Marshal(map[string]any{"idCliente": idCli})
		r.Body = io.NopCloser(bytes.NewReader(raw))
		u := fmt.Sprintf("%s/v4/dispositivo/listarByIdCliente", config.ApiUrl)
		proxyConfmonitPost(w, r, u)
		return
	}
	idFra, err := assertFranqueadoQuery(r, seguranca.TextoDoPayload(payload, "idFranqueado", "id_franqueado"))
	if err != nil {
		if strings.Contains(err.Error(), "nao autorizado") {
			responderEscopoProibido(w, err.Error())
			return
		}
		auxiliar.RespostaErro(w, http.StatusForbidden, err)
		return
	}
	raw, _ := json.Marshal(map[string]any{"idFranqueado": idFra})
	r.Body = io.NopCloser(bytes.NewReader(raw))
	u := fmt.Sprintf("%s/v4/dispositivo/listarByIdFranqueado", config.ApiUrl)
	proxyConfmonitPost(w, r, u)
}

func ProxySetArmadoDispositivo(w http.ResponseWriter, r *http.Request) {
	u := fmt.Sprintf("%s/v4/dispositivo/setArmadoById", config.ApiUrl)
	proxyConfmonitPost(w, r, u)
}

func ProxyGetArmadoDispositivo(w http.ResponseWriter, r *http.Request) {
	u := fmt.Sprintf("%s/v4/dispositivo/getArmadoById", config.ApiUrl)
	proxyConfmonitPost(w, r, u)
}

func ProxyListarSetores(w http.ResponseWriter, r *http.Request) {
	u := fmt.Sprintf("%s/v4/setor/listaByIdDispositivo", config.ApiUrlSetor)
	proxyConfmonitPost(w, r, u)
}

func ProxyListarCameras(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	idCli := seguranca.IdClienteDoCookie(cookie)
	if idCli == "" {
		idCli = strings.TrimSpace(r.URL.Query().Get("id_cliente"))
	}
	// Escopo cliente: cookie CLI ou pedido explícito com sessão autenticada
	if seguranca.EhCliente(cookie) || (idCli != "" && seguranca.IdFranqueadoDoCookie(cookie) != "") {
		if !seguranca.EhCliente(cookie) && idCli != "" {
			// reforça id no cookie map local para o helper
			if cookie == nil {
				cookie = map[string]string{}
			}
			cookie["idCliente"] = idCli
			cookie["tipo"] = "CLI"
		}
		listarCamerasCliente(w, r, cookie)
		return
	}
	idFranqueado, err := assertFranqueadoQuery(r, r.URL.Query().Get("id_franqueado"))
	if err != nil {
		if strings.Contains(err.Error(), "nao autorizado") {
			responderEscopoProibido(w, err.Error())
			return
		}
		auxiliar.RespostaErro(w, http.StatusForbidden, err)
		return
	}
	path := fmt.Sprintf("/vis_camera_by_franqueado?id_franqueado=%s", url.QueryEscape(idFranqueado))
	proxyVisOrXano(w, r, http.MethodGet, path)
}

// listarCamerasCliente — usa endpoint by_cliente se existir; senão filtra by_franqueado no Go.
func listarCamerasCliente(w http.ResponseWriter, r *http.Request, cookie map[string]string) {
	idCli := seguranca.IdClienteDoCookie(cookie)
	if idCli == "" {
		http.Error(w, `{"status":"id_cliente obrigatorio — faca login novamente"}`, http.StatusBadRequest)
		return
	}

	// 1) Tenta API dedicada
	pathCli := fmt.Sprintf("/vis_camera_by_cliente?id_cliente=%s", url.QueryEscape(idCli))
	if status, raw, ok := fetchXano(http.MethodGet, pathCli); ok && status < 400 {
		auxiliar.RespostaAPP(w, raw)
		return
	}

	// 2) Fallback: lista do franqueado e filtra pelo cliente da sessão
	idFra := seguranca.IdFranqueadoDoCookie(cookie)
	if idFra == "" {
		http.Error(w, `{"status":"id_franqueado ausente na sessao"}`, http.StatusBadRequest)
		return
	}
	pathFra := fmt.Sprintf("/vis_camera_by_franqueado?id_franqueado=%s", url.QueryEscape(idFra))
	status, raw, ok := fetchXano(http.MethodGet, pathFra)
	if !ok || status >= 400 {
		http.Error(w, `{"status":"falha ao listar cameras do cliente"}`, http.StatusBadGateway)
		return
	}

	filtrado, err := filtrarListaPorCampo(raw, "id_cliente", idCli)
	if err != nil {
		http.Error(w, `{"status":"resposta invalida do xano"}`, http.StatusBadGateway)
		return
	}
	auxiliar.RespostaAPP(w, filtrado)
}

func fetchXano(metodo, path string) (int, []byte, bool) {
	if config.VisPostgresEnabled {
		status, raw, err := visdata.Dispatch(context.Background(), metodo, path, http.NoBody)
		if err == nil {
			return status, raw, true
		}
		if !errors.Is(err, visdata.ErrNotHandled) {
			return 0, nil, false
		}
	}
	if config.XanoBaseUrl == "" {
		return 0, nil, false
	}
	req, err := http.NewRequest(metodo, config.XanoBaseUrl+path, nil)
	if err != nil {
		return 0, nil, false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, false
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, false
	}
	return resp.StatusCode, raw, true
}

func filtrarListaPorCampo(raw []byte, campo, valor string) ([]byte, error) {
	valor = strings.TrimSpace(valor)
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}

	switch dados := root["dados"].(type) {
	case []any:
		root["dados"] = filtrarSlicePorCampo(dados, campo, valor)
	case map[string]any:
		// paginação Xano: { items: [...], curPage, nextPage, ... }
		if items, ok := dados["items"].([]any); ok {
			filtrados := filtrarSlicePorCampo(items, campo, valor)
			dados["items"] = filtrados
			dados["itemsReceived"] = float64(len(filtrados))
			root["dados"] = dados
		} else {
			return nil, fmt.Errorf("formato dados map sem items")
		}
	default:
		var arr []any
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, fmt.Errorf("formato dados desconhecido")
		}
		root = map[string]any{"dados": filtrarSlicePorCampo(arr, campo, valor)}
	}
	return json.Marshal(root)
}

func filtrarSlicePorCampo(lista []any, campo, valor string) []any {
	out := make([]any, 0, len(lista))
	for _, item := range lista {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(fmt.Sprint(m[campo])) == valor {
			out = append(out, item)
		}
	}
	return out
}

func ProxyListarLicencas(w http.ResponseWriter, r *http.Request) {
	idFranqueado := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFranqueado != "" && config.VisPostgresEnabled {
		_, _ = SyncLicencasFranqueadoFromXano(r.Context(), idFranqueado)
	}
	status := r.URL.Query().Get("status")
	unidade := r.URL.Query().Get("unidade")
	path := fmt.Sprintf("/vis_licenca_by_franqueado?id_franqueado=%s", idFranqueado)
	if status != "" {
		path += fmt.Sprintf("&status=%s", status)
	}
	if unidade != "" {
		path += fmt.Sprintf("&unidade=%s", unidade)
	}
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func ProxyListarGravacaoStorage(w http.ResponseWriter, r *http.Request) {
	idFranqueado := r.URL.Query().Get("id_franqueado")
	status := r.URL.Query().Get("status")
	path := fmt.Sprintf("/vis_gravacao_storage_by_franqueado?id_franqueado=%s", idFranqueado)
	if status != "" {
		path += fmt.Sprintf("&status=%s", status)
	}
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func ProxyCriarGravacaoStorage(w http.ResponseWriter, r *http.Request) {
	proxyVisOrXano(w, r, http.MethodPost, "/vis_gravacao_storage")
}

func ProxyAtualizarGravacaoStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_gravacao_storage/%s?vis_gravacao_storage_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPut, path)
}

func ProxyAtivarGravacaoCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera/gravacao/ativar/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPost, path)
}

func ProxyDesativarGravacaoCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera/gravacao/desativar/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPost, path)
}

func ProxyLiberarLicencaCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera/licenca/liberar/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPost, path)
}

func ProxyFlushGravacaoCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera/gravacao/flush/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPost, path)
}

func ProxyPausarAnaliticoCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera/analitico/pausar/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPost, path)
}

func ProxyReativarStreamCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera_stream_reactivate?camera_id=%s", id)
	proxyVisOrXano(w, r, http.MethodPost, path)
}

func ProxyGetGradeCliente(w http.ResponseWriter, r *http.Request) {
	idCliente := mux.Vars(r)["id_cliente"]
	idDisp := strings.TrimSpace(r.URL.Query().Get("id_dispositivo"))
	path := fmt.Sprintf("/cvg_grade_by_cliente?id_cliente=%s", url.QueryEscape(idCliente))
	if idDisp != "" {
		path += "&id_dispositivo=" + url.QueryEscape(idDisp)
	}
	proxyXanoCvg(w, r, http.MethodGet, path)
}

func ProxyPutGradeCliente(w http.ResponseWriter, r *http.Request) {
	idCliente := mux.Vars(r)["id_cliente"]
	cookie := sessaoCookie(r)
	body, _ := io.ReadAll(r.Body)
	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	if _, ok := payload["id_cliente"]; !ok {
		payload["id_cliente"] = idCliente
	}
	if _, ok := payload["id_franqueado"]; !ok {
		payload["id_franqueado"] = seguranca.IdFranqueadoDoCookie(cookie)
	}
	raw, _ := json.Marshal(payload)
	r.Body = io.NopCloser(bytes.NewReader(raw))
	path := fmt.Sprintf("/cvg_grade_by_cliente?id_cliente=%s", url.QueryEscape(idCliente))
	proxyXanoCvg(w, r, http.MethodPut, path)
}

func ProxyGetGradeList(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		idFra = seguranca.IdFranqueadoDoCookie(cookie)
	}
	if idFra == "" {
		http.Error(w, "id_franqueado obrigatorio", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("/cvg_grade_list?id_franqueado=%s", url.QueryEscape(idFra))
	proxyXanoCvg(w, r, http.MethodGet, path)
}

func ProxyPatchGradeEscopoAtiva(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	body, _ := io.ReadAll(r.Body)
	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	if _, ok := payload["id_franqueado"]; !ok {
		payload["id_franqueado"] = seguranca.IdFranqueadoDoCookie(cookie)
	}
	raw, _ := json.Marshal(payload)
	r.Body = io.NopCloser(bytes.NewReader(raw))
	proxyXanoCvg(w, r, http.MethodPatch, "/cvg_grade_escopo_ativa")
}

func ProxyListarGravacaoSegmentos(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	idFranqueado := r.URL.Query().Get("id_franqueado")
	cameraID := r.URL.Query().Get("vis_camera_id")
	de := r.URL.Query().Get("de")
	ate := r.URL.Query().Get("ate")
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	if seguranca.EhCliente(cookie) {
		if cameraID == "" {
			http.Error(w, "vis_camera_id obrigatorio", http.StatusBadRequest)
			return
		}
		if !cameraPertenceAoCliente(cameraID, seguranca.IdClienteDoCookie(cookie)) {
			http.Error(w, `{"status":"camera nao autorizada"}`, http.StatusForbidden)
			return
		}
	}

	var path string
	if cameraID != "" {
		path = fmt.Sprintf("/vis_gravacao_segmento_by_camera?vis_camera_id=%s&page=%s", cameraID, page)
	} else if idFranqueado != "" {
		path = fmt.Sprintf("/vis_gravacao_segmento_by_franqueado?id_franqueado=%s&page=%s", idFranqueado, page)
	} else {
		http.Error(w, "vis_camera_id ou id_franqueado obrigatorio", http.StatusBadRequest)
		return
	}
	if de != "" {
		path += "&de=" + url.QueryEscape(de)
	}
	if ate != "" {
		path += "&ate=" + url.QueryEscape(ate)
	}
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func cameraPertenceAoCliente(cameraID, idCliente string) bool {
	cameraID = strings.TrimSpace(cameraID)
	idCliente = strings.TrimSpace(idCliente)
	if cameraID == "" || idCliente == "" {
		return false
	}
	id, err := strconv.Atoi(cameraID)
	if err != nil || id < 1 {
		return false
	}
	if config.VisPostgresEnabled {
		cam, err := visdata.GetCameraByID(context.Background(), id)
		if err != nil {
			return false
		}
		idCamCli := strings.TrimSpace(fmt.Sprint(cam["id_cliente"]))
		return idCamCli != "" && idCamCli != "<nil>" && idCamCli == idCliente
	}
	if config.XanoBaseUrl == "" {
		return false
	}
	path := fmt.Sprintf("/vis_camera/%s?vis_camera_id=%s", cameraID, cameraID)
	req, err := http.NewRequest(http.MethodGet, config.XanoBaseUrl+path, nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return false
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	var out map[string]any
	if json.Unmarshal(raw, &out) != nil {
		return false
	}
	cam := out
	if d, ok := out["dados"].(map[string]any); ok {
		cam = d
	}
	idCamCli := strings.TrimSpace(fmt.Sprint(cam["id_cliente"]))
	return idCamCli != "" && idCamCli != "<nil>" && idCamCli == idCliente
}

func ProxyCriarLicenca(w http.ResponseWriter, r *http.Request) {
	proxyVisOrXano(w, r, http.MethodPost, "/vis_licenca")
}

func ProxyListarEventos(w http.ResponseWriter, r *http.Request) {
	cookie := sessaoCookie(r)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	idCli := seguranca.IdClienteDoCookie(cookie)
	if idCli == "" {
		idCli = strings.TrimSpace(r.URL.Query().Get("id_cliente"))
	}
	if seguranca.EhCliente(cookie) || (idCli != "" && seguranca.IdFranqueadoDoCookie(cookie) != "") {
		if cookie == nil {
			cookie = map[string]string{}
		}
		if seguranca.IdClienteDoCookie(cookie) == "" {
			cookie["idCliente"] = idCli
		}
		listarEventosCliente(w, r, cookie, page)
		return
	}
	idFranqueado := r.URL.Query().Get("id_franqueado")
	if idFranqueado == "" {
		idFranqueado = seguranca.IdFranqueadoDoCookie(cookie)
	}
	path := fmt.Sprintf("/vis_evento_by_franqueado_page?id_franqueado=%s&page=%s", url.QueryEscape(idFranqueado), page)
	path = appendEventosFiltroQuery(path, r)
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func appendEventosFiltroQuery(path string, r *http.Request) string {
	q := r.URL.Query()
	if idCli := strings.TrimSpace(q.Get("id_cliente")); idCli != "" {
		path += "&id_cliente=" + url.QueryEscape(idCli)
	}
	if de := strings.TrimSpace(q.Get("data_de")); de != "" {
		path += "&data_de=" + url.QueryEscape(de)
	}
	if ate := strings.TrimSpace(q.Get("data_ate")); ate != "" {
		path += "&data_ate=" + url.QueryEscape(ate)
	}
	return path
}

func listarEventosCliente(w http.ResponseWriter, r *http.Request, cookie map[string]string, page string) {
	idCli := seguranca.IdClienteDoCookie(cookie)
	if idCli == "" {
		http.Error(w, `{"status":"id_cliente obrigatorio — faca login novamente"}`, http.StatusBadRequest)
		return
	}

	pathCli := fmt.Sprintf("/vis_evento_by_cliente_page?id_cliente=%s&page=%s", url.QueryEscape(idCli), page)
	pathCli = appendEventosFiltroQuery(pathCli, r)
	if status, raw, ok := fetchXano(http.MethodGet, pathCli); ok && status < 400 {
		auxiliar.RespostaAPP(w, raw)
		return
	}

	idFra := seguranca.IdFranqueadoDoCookie(cookie)
	if idFra == "" {
		http.Error(w, `{"status":"id_franqueado ausente na sessao"}`, http.StatusBadRequest)
		return
	}
	pathFra := fmt.Sprintf("/vis_evento_by_franqueado_page?id_franqueado=%s&page=%s", url.QueryEscape(idFra), page)
	pathFra = appendEventosFiltroQuery(pathFra, r)
	if strings.TrimSpace(r.URL.Query().Get("id_cliente")) == "" {
		pathFra += "&id_cliente=" + url.QueryEscape(idCli)
	}
	status, raw, ok := fetchXano(http.MethodGet, pathFra)
	if !ok || status >= 400 {
		http.Error(w, `{"status":"falha ao listar eventos do cliente"}`, http.StatusBadGateway)
		return
	}
	filtrado, err := filtrarListaPorCampo(raw, "id_cliente", idCli)
	if err != nil {
		http.Error(w, `{"status":"resposta invalida do xano"}`, http.StatusBadGateway)
		return
	}
	auxiliar.RespostaAPP(w, filtrado)
}

func ProxyGetEventoClips(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_evento/%s/clips?id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func ProxyCriarCamera(w http.ResponseWriter, r *http.Request) {
	proxyVisOrXano(w, r, http.MethodPost, "/vis_camera")
}

func ProxyGetCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cookie := sessaoCookie(r)
	if seguranca.EhCliente(cookie) {
		if !cameraPertenceAoCliente(id, seguranca.IdClienteDoCookie(cookie)) {
			http.Error(w, `{"status":"camera nao autorizada"}`, http.StatusForbidden)
			return
		}
	} else if !seguranca.EhAdministrator(cookie) {
		if idFra := seguranca.IdFranqueadoDoCookie(cookie); idFra != "" {
			if !cameraPertenceAoFranqueado(id, idFra) {
				responderEscopoProibido(w, "camera nao autorizada")
				return
			}
		}
	}
	path := fmt.Sprintf("/vis_camera/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func ProxyExcluirCameraArea(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera_area/%s?id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodDelete, path)
}

func ProxyListarCameraAreas(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera_area_by_camera?vis_camera_id=%s", id)
	proxyVisOrXano(w, r, http.MethodGet, path)
}

func ProxyCriarCameraArea(w http.ResponseWriter, r *http.Request) {
	proxyVisOrXano(w, r, http.MethodPost, "/vis_camera_area")
}

func ProxyAtualizarCameraArea(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera_area/%s?id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPut, path)
}

func ProxyAtualizarCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	path := fmt.Sprintf("/vis_camera/%s?vis_camera_id=%s", id, id)
	proxyVisOrXano(w, r, http.MethodPut, path)
	go func() {
		inv, _ := json.Marshal(map[string]any{"vis_camera_id": id})
		if gresp, gerr := guardRequest(http.MethodPost, "/cache/invalidate", inv); gerr == nil && gresp != nil {
			_ = gresp.Body.Close()
		}
	}()
}

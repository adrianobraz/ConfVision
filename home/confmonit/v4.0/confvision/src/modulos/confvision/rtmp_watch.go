package confvision

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Timeout curto: se o guard travar, a UI recebe erro em vez de Proxy Error do nginx.
var guardHTTP = &http.Client{Timeout: 8 * time.Second}

func fetchMediamtxForCamera(cameraID int) (map[string]string, error) {
	u := fmt.Sprintf("%s/vis_camera/mediamtx/%d?vis_camera_id=%d",
		strings.TrimRight(config.XanoBaseUrl, "/"), cameraID, cameraID)
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("xano mediamtx: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xano mediamtx HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("parse mediamtx: %w", err)
	}
	out := map[string]string{
		"rtmp_public":          strings.TrimSpace(fmt.Sprint(payload["rtmp_public"])),
		"hls_public":           strings.TrimSpace(fmt.Sprint(payload["hls_public"])),
		"rtsp_internal":        strings.TrimSpace(fmt.Sprint(payload["rtsp_internal"])),
		"nome":                 strings.TrimSpace(fmt.Sprint(payload["nome"])),
		"vis_mediamtx_node_id": strings.TrimSpace(fmt.Sprint(payload["vis_mediamtx_node_id"])),
	}
	if out["rtmp_public"] == "" || out["rtmp_public"] == "<nil>" {
		return nil, fmt.Errorf("xano mediamtx sem rtmp_public para camera %d", cameraID)
	}
	return out, nil
}

func guardBase() string {
	u := strings.TrimRight(strings.TrimSpace(config.RtmpGuardURL), "/")
	if u == "" {
		u = strings.TrimRight(strings.TrimSpace(config.RtmpWatchURL), "/")
	}
	return u
}

func guardRequest(method, path string, body []byte) (*http.Response, error) {
	base := guardBase()
	if base == "" {
		return nil, fmt.Errorf("RTMP_GUARD_URL (ou RTMP_WATCH_URL) nao configurado")
	}
	req, err := http.NewRequest(method, base+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if key := strings.TrimSpace(config.RtmpGuardAdminKey); key != "" {
		req.Header.Set("X-RTMP-Guard-Key", key)
	}
	return guardHTTP.Do(req)
}

func proxyGuardJSON(w http.ResponseWriter, method, path string, body []byte) {
	proxyGuardJSONMaybeEnrich(w, method, path, body, false)
}

func proxyGuardJSONEnriched(w http.ResponseWriter, method, path string, body []byte) {
	proxyGuardJSONMaybeEnrich(w, method, path, body, true)
}

func proxyGuardJSONMaybeEnrich(w http.ResponseWriter, method, path string, body []byte, enrich bool) {
	resp, err := guardRequest(method, path, body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("rtmp-guard indisponivel: %w", err))
		return
	}
	defer resp.Body.Close()
	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	if resp.StatusCode >= 400 {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("rtmp-guard HTTP %d: %s", resp.StatusCode, string(corpo)))
		return
	}
	if enrich {
		if enriched, err := enrichGuardJSON(corpo); err == nil {
			corpo = enriched
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(corpo)
}

func CarregarRtmpFalhas(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "rtmp-falhas.html", paginaBase("Falhas RTMP", "/carregar-menu-confvision"))
}

func ProxyRtmpFalhas(w http.ResponseWriter, r *http.Request) {
	if guardBase() == "" {
		auxiliar.RespostaAPP(w, []byte(`{"status":"rtmp guard nao configurado","dados":[],"resumo":{},"aviso":"Defina RTMP_GUARD_URL no .env apontando para confvision-rtmp-guard na VPS"}`))
		return
	}
	q := r.URL.RawQuery
	path := "/falhas"
	if q != "" {
		path += "?" + q
	}
	proxyGuardJSONEnriched(w, http.MethodGet, path, nil)
}

func ProxyRtmpFalhasResumo(w http.ResponseWriter, r *http.Request) {
	if guardBase() == "" {
		auxiliar.RespostaAPP(w, []byte(`{"status":"rtmp guard nao configurado","dados":{}}`))
		return
	}
	proxyGuardJSON(w, http.MethodGet, "/resumo", nil)
}

func ProxyRtmpBans(w http.ResponseWriter, r *http.Request) {
	if guardBase() == "" {
		auxiliar.RespostaAPP(w, []byte(`{"status":"rtmp guard nao configurado","dados":[]}`))
		return
	}
	proxyGuardJSON(w, http.MethodGet, "/bans", nil)
}

func ProxyRtmpOnline(w http.ResponseWriter, r *http.Request) {
	if guardBase() == "" {
		auxiliar.RespostaAPP(w, []byte(`{"status":"rtmp guard nao configurado","dados":[],"aviso":"Defina RTMP_GUARD_URL no .env"}`))
		return
	}
	proxyGuardJSONEnriched(w, http.MethodGet, "/online", nil)
}

func ProxyRtmpPublishHealth(w http.ResponseWriter, r *http.Request) {
	if guardBase() == "" {
		auxiliar.RespostaAPP(w, []byte(`{"status":"rtmp guard nao configurado","dados":[],"resumo":{}}`))
		return
	}
	q := r.URL.RawQuery
	path := "/health/publishers"
	if q != "" {
		path += "?" + q
	}
	proxyGuardJSONEnriched(w, http.MethodGet, path, nil)
}

func ProxyRtmpUnban(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	proxyGuardJSON(w, http.MethodPost, "/unban", body)
}

func ProxyRtmpBan(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	proxyGuardJSON(w, http.MethodPost, "/ban", body)
}

// ProxyRtmpPublishURL devolve URL RTMP com Hashids no path (sem /live/, sem query).
func ProxyRtmpPublishURL(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cameraID, err := strconv.Atoi(id)
	if err != nil || cameraID < 1 {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id de camera invalido"))
		return
	}
	if strings.TrimSpace(config.RtmpPublishSecret) == "" {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("RTMP_PUBLISH_SECRET nao configurado no servidor"))
		return
	}

	cookie, _ := seguranca.LerCookies(r)
	idFraSessao := seguranca.IdFranqueadoDoCookie(cookie)

	path := fmt.Sprintf("%s/vis_camera/%s?vis_camera_id=%s",
		strings.TrimRight(config.XanoBaseUrl, "/"), id, id)
	resp, err := http.Get(path)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		auxiliar.RespostaErro(w, http.StatusNotFound, fmt.Errorf("camera nao encontrada"))
		return
	}
	if resp.StatusCode >= 400 {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("xano HTTP %d: %s", resp.StatusCode, string(raw)))
		return
	}

	var cam map[string]any
	if err := json.Unmarshal(raw, &cam); err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("parse camera: %w", err))
		return
	}
	idFra := strings.TrimSpace(fmt.Sprint(cam["id_franqueado"]))
	if idFra == "" || idFra == "<nil>" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("camera sem id_franqueado"))
		return
	}
	if idFraSessao != "" && idFraSessao != idFra {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("camera de outro franqueado"))
		return
	}

	mtx, err := fetchMediamtxForCamera(cameraID)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	base := strings.TrimSpace(mtx["rtmp_public"])
	if base == "" {
		base = config.MediamtxRtmpPublic
		if base == "" {
			base = config.MediamtxRtmpPublishBase
		}
	}
	chave := ChaveRtmp(cameraID)
	pathLive := StreamPath(cameraID)
	wifi := MontarRtmpPublishURL(base, cameraID, false)
	dvr := MontarRtmpPublishURL(base, cameraID, true)
	protocolo := strings.ToLower(strings.TrimSpace(fmt.Sprint(cam["protocolo"])))
	atual := wifi
	if protocolo == "dvr" {
		atual = dvr
	}

	podeStream, motivoStream := cameraPodeStreamFromMap(cam)

	hlsBase := strings.TrimSpace(mtx["hls_public"])
	if hlsBase == "" {
		hlsBase = config.MediamtxHlsPublic
		if hlsBase == "" {
			hlsBase = config.MediamtxHlsBase
		}
	}
	hlsBase = strings.TrimRight(hlsBase, "/")
	hls := ""
	if hlsBase != "" && pathLive != "" {
		hls = hlsBase + "/" + pathLive + "/index.m3u8"
	}

	out, _ := json.Marshal(map[string]any{
		"status":               "ok",
		"vis_camera_id":        cameraID,
		"vis_mediamtx_node_id": mtx["vis_mediamtx_node_id"],
		"mediamtx_nome":        mtx["nome"],
		"chave":                chave,
		"path":                 pathLive,
		"user":                 idFra,
		"wifi":                 wifi,
		"dvr":                  dvr,
		"url":                  atual,
		"hls":                  hls,
		"protocolo":            protocolo,
		"bloqueado":            truthyCameraBool(cam["bloqueado"]),
		"ativo":                truthyCameraBool(cam["ativo"]),
		"plano":                planoFromCam(cam),
		"pode_stream":          podeStream,
		"motivo_stream":        motivoStream,
	})
	auxiliar.RespostaAPP(w, out)
}

// ProxyBloquearCamera marca bloqueado no Xano e invalida cache do guard.
func ProxyBloquearCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	body, _ := io.ReadAll(r.Body)
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	if _, ok := payload["bloqueado"]; !ok {
		payload["bloqueado"] = true
	}
	raw, _ := json.Marshal(payload)

	u := fmt.Sprintf("%s/vis_camera/bloquear/%s?vis_camera_id=%s",
		strings.TrimRight(config.XanoBaseUrl, "/"), id, id)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	defer resp.Body.Close()
	corpo, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("xano HTTP %d: %s", resp.StatusCode, string(corpo)))
		return
	}

	// invalida cache do guard (best-effort)
	inv, _ := json.Marshal(map[string]any{"vis_camera_id": id})
	if gresp, gerr := guardRequest(http.MethodPost, "/cache/invalidate", inv); gerr == nil {
		_ = gresp.Body.Close()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(corpo)
}

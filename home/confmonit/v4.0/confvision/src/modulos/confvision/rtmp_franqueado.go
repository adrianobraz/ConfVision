package confvision

import (
	"confvision/src/auxiliar"
	"confvision/src/seguranca"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func sessaoEhAdministrator(r *http.Request) bool {
	cookie, err := seguranca.LerCookies(r)
	return err == nil && seguranca.EhAdministrator(cookie)
}

func proxyGuardJSONEnrichedFiltrado(w http.ResponseWriter, r *http.Request, method, path string, body []byte) {
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
	if enriched, err := enrichGuardJSON(corpo); err == nil {
		corpo = enriched
	}
	if !sessaoEhAdministrator(r) {
		idFra, err := FranqueadoDaSessao(r)
		if err != nil {
			responderEscopoProibido(w, err.Error())
			return
		}
		corpo, err = filtrarGuardJSONPorFranqueado(corpo, idFra)
		if err != nil {
			auxiliar.RespostaErro(w, http.StatusBadGateway, err)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(corpo)
}

func filtrarGuardJSONPorFranqueado(raw []byte, idFranqueado string) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return raw, err
	}

	if dados, ok := root["dados"].([]any); ok {
		root["dados"] = filtrarLinhasGuardPorFranqueado(dados, idFranqueado)
		if t, ok := root["total"].(float64); ok {
			root["total"] = len(root["dados"].([]any))
			_ = t
		} else {
			root["total"] = len(root["dados"].([]any))
		}
	}

	if resumo, ok := root["resumo"].(map[string]any); ok {
		root["resumo"] = resumirFalhasFiltradas(root["dados"], resumo)
	}

	out, err := json.Marshal(root)
	if err != nil {
		return raw, err
	}
	return out, nil
}

func filtrarLinhasGuardPorFranqueado(dados []any, idFranqueado string) []any {
	var out []any
	for _, item := range dados {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if linhaPertenceAoFranqueado(row, idFranqueado) {
			out = append(out, row)
		}
	}
	return out
}

func linhaPertenceAoFranqueado(row map[string]any, idFranqueado string) bool {
	if idFra := strings.TrimSpace(fmt.Sprint(row["id_franqueado"])); idFra != "" && idFra != "<nil>" {
		return idFra == idFranqueado
	}
	if id := cameraIDFromRow(row); id > 0 {
		if idFra, ok := cameraIDFranqueado(id); ok {
			return idFra == idFranqueado
		}
	}
	return false
}

func resumirFalhasFiltradas(dados any, resumo map[string]any) map[string]any {
	lista, ok := dados.([]any)
	if !ok {
		return resumo
	}
	out := map[string]any{}
	for k, v := range resumo {
		out[k] = v
	}
	out["total_eventos"] = len(lista)
	ips := map[string]struct{}{}
	for _, item := range lista {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ip := strings.TrimSpace(fmt.Sprint(row["ip"]))
		if ip != "" && ip != "<nil>" {
			ips[ip] = struct{}{}
		}
	}
	top := make([]string, 0, len(ips))
	for ip := range ips {
		top = append(top, ip)
	}
	out["top_ips"] = top
	return out
}

func ProxyRtmpBansFranqueado(w http.ResponseWriter, r *http.Request) {
	if guardBase() == "" {
		auxiliar.RespostaAPP(w, []byte(`{"status":"rtmp guard nao configurado","dados":[]}`))
		return
	}
	idFra, err := FranqueadoDaSessao(r)
	if err != nil {
		responderEscopoProibido(w, err.Error())
		return
	}

	resp, err := guardRequest(http.MethodGet, "/bans", nil)
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

	ipMap, err := mapaIPCamarasFranqueado(idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	enriquecidos, err := montarBansFranqueado(corpo, ipMap, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	auxiliar.RespostaAPP(w, enriquecidos)
}

func mapaIPCamarasFranqueado(idFranqueado string) (map[string][]map[string]any, error) {
	out := map[string][]map[string]any{}

	addRow := func(row map[string]any) {
		ip := strings.TrimSpace(fmt.Sprint(row["ip"]))
		if ip == "" || ip == "<nil>" {
			return
		}
		if !linhaPertenceAoFranqueado(row, idFranqueado) {
			return
		}
		out[ip] = append(out[ip], row)
	}

	if resp, err := guardRequest(http.MethodGet, "/online", nil); err == nil {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 400 {
			if enriched, err := enrichGuardJSON(raw); err == nil {
				raw = enriched
			}
			var root map[string]any
			if json.Unmarshal(raw, &root) == nil {
				if dados, ok := root["dados"].([]any); ok {
					for _, item := range dados {
						if row, ok := item.(map[string]any); ok {
							addRow(row)
						}
					}
				}
			}
		}
	}

	if resp, err := guardRequest(http.MethodGet, "/falhas?limit=500", nil); err == nil {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 400 {
			if enriched, err := enrichGuardJSON(raw); err == nil {
				raw = enriched
			}
			var root map[string]any
			if json.Unmarshal(raw, &root) == nil {
				if dados, ok := root["dados"].([]any); ok {
					for _, item := range dados {
						if row, ok := item.(map[string]any); ok {
							addRow(row)
						}
					}
				}
			}
		}
	}

	return out, nil
}

func montarBansFranqueado(rawBans []byte, ipMap map[string][]map[string]any, idFranqueado string) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(rawBans, &root); err != nil {
		return nil, err
	}
	bans, _ := root["dados"].([]any)
	var out []map[string]any

	for _, item := range bans {
		ban, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ip := strings.TrimSpace(fmt.Sprint(ban["ip"]))
		if ip == "" || ip == "<nil>" {
			continue
		}
		refs := ipMap[ip]
		if len(refs) == 0 {
			continue
		}
		ref := refs[0]
		cameraID := cameraIDFromRow(ref)
		ctx := buscarContextoCamera(cameraID)
		row := map[string]any{
			"ip":           ip,
			"motivo":       ban["motivo"],
			"restante_sec": ban["restante_sec"],
			"manual":       ban["manual"],
			"camera_id":    cameraID,
			"nome_camera":  ctx.NomeCamera,
			"nome_cliente": ctx.NomeCliente,
			"bloqueado":    ctx.Bloqueado,
			"id_franqueado": idFranqueado,
		}
		if cameraID > 0 {
			if cam, ok := fetchVisCamera(cameraID); ok {
				row["ativo"] = truthyCameraBool(cam["ativo"])
				row["plano"] = planoFromCam(cam)
			}
		}
		out = append(out, row)
	}

	payload, err := json.Marshal(map[string]any{
		"status": "ok",
		"total":  len(out),
		"dados":  out,
	})
	return payload, err
}

func validarUnbanIPFranqueado(r *http.Request, ip string) error {
	idFra, err := FranqueadoDaSessao(r)
	if err != nil {
		return err
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return fmt.Errorf("ip obrigatorio")
	}

	ipMap, err := mapaIPCamarasFranqueado(idFra)
	if err != nil {
		return err
	}
	if len(ipMap[ip]) == 0 {
		return fmt.Errorf("ip nao pertence ao franqueado")
	}
	return nil
}

func ProxyRtmpUnbanSeguro(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if !sessaoEhAdministrator(r) {
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		ip := strings.TrimSpace(fmt.Sprint(payload["ip"]))
		if err := validarUnbanIPFranqueado(r, ip); err != nil {
			auxiliar.RespostaErro(w, http.StatusForbidden, err)
			return
		}
	}
	proxyGuardJSON(w, http.MethodPost, "/unban", body)
}

func ProxyCamerasFranqueado(w http.ResponseWriter, r *http.Request) {
	idFra, err := FranqueadoDaSessao(r)
	if err != nil {
		responderEscopoProibido(w, err.Error())
		return
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	lista, err := fetchCamerasFranqueado(r, idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	fraMap := map[string]string{idFra: lookupNomeFranqueado(idFra)}
	cliIDs := coletarClientesIDs(lista)
	cliMap := lookupClientesBatchConfvision(cliIDs)

	type row struct {
		ID             int    `json:"id"`
		Nome           string `json:"nome"`
		NomeFranqueado string `json:"nome_franqueado"`
		NomeCliente    string `json:"nome_cliente"`
		Bloqueado      bool   `json:"bloqueado"`
		Ativo          bool   `json:"ativo"`
		Plano          string `json:"plano"`
		IDFranqueado   string `json:"id_franqueado"`
		IDCliente      string `json:"id_cliente"`
	}

	out := make([]row, 0, len(lista))
	for _, cam := range lista {
		id := intDeCampoConf(cam, "id")
		nome := textoCampoConf(cam, "nome")
		if nome == "" {
			nome = fmt.Sprintf("Câmera #%d", id)
		}
		idCli := textoCampoConf(cam, "id_cliente")
		r := row{
			ID:             id,
			Nome:           nome,
			NomeFranqueado: fraMap[idFra],
			NomeCliente:    cliMap[idCli],
			Bloqueado:      truthyCameraBool(cam["bloqueado"]),
			Ativo:          truthyCameraBool(cam["ativo"]),
			Plano:          textoCampoConf(cam, "plano"),
			IDFranqueado:   idFra,
			IDCliente:      idCli,
		}
		if r.NomeCliente == "" && idCli != "" {
			r.NomeCliente = idCli
		}
		if q != "" {
			blob := strings.ToLower(strings.Join([]string{r.Nome, r.NomeCliente, r.Plano, strconv.Itoa(r.ID)}, " "))
			if !strings.Contains(blob, q) {
				continue
			}
		}
		out = append(out, r)
	}

	raw, _ := json.Marshal(map[string]any{"status": "ok", "total": len(out), "dados": out})
	auxiliar.RespostaAPP(w, raw)
}

func fetchCamerasFranqueado(r *http.Request, idFranqueado string) ([]map[string]any, error) {
	if configVisPostgres() {
		return visdataListCamerasByFranqueado(r, idFranqueado)
	}
	path := fmt.Sprintf("/vis_camera_by_franqueado?id_franqueado=%s", idFranqueado)
	status, raw, ok := fetchXano(http.MethodGet, path)
	if !ok || status >= 400 {
		return nil, fmt.Errorf("falha ao listar cameras")
	}
	return parseListaCamerasJSON(raw)
}

func configVisPostgres() bool {
	return configEnabledPostgres()
}

// evita import cycle — wrappers mínimos
func configEnabledPostgres() bool {
	return configVisEnabled()
}

func visdataListCamerasByFranqueado(r *http.Request, idFranqueado string) ([]map[string]any, error) {
	return listCamerasByFranqueadoCtx(r.Context(), idFranqueado)
}

func coletarClientesIDs(lista []map[string]any) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, cam := range lista {
		id := textoCampoConf(cam, "id_cliente")
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func textoCampoConf(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprint(raw))
	if s == "<nil>" {
		return ""
	}
	return s
}

func intDeCampoConf(m map[string]any, key string) int {
	s := textoCampoConf(m, key)
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

func lookupClientesBatchConfvision(ids []string) map[string]string {
	out := map[string]string{}
	if len(ids) == 0 {
		return out
	}
	for _, id := range ids {
		if nome := lookupNomeCliente(id); nome != "" {
			out[id] = nome
		}
	}
	return out
}

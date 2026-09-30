package visdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"confvision/src/config"
)

type moniInboundRequest struct {
	IDFranqueado string          `json:"id_franqueado"`
	Event        moniInboundEvent `json:"event"`
	RawBody      string          `json:"raw_body"`
}

type moniInboundEvent struct {
	DateTime        string `json:"DateTime"`
	CustomerCode    int    `json:"CustomerCode"`
	SecondCode      string `json:"SecondCode"`
	Partition       string `json:"Partition"`
	Company         int    `json:"Company"`
	EventCode       int    `json:"EventCode"`
	EventSecondCode string `json:"EventSecondCode"`
	EventType       int    `json:"EventType"`
	Sector          int    `json:"Sector"`
	User            int    `json:"User"`
}

func MoniInboundPOST(ctx context.Context, payload map[string]any) (int, []byte, error) {
	r := HTTPRequestFromContext(ctx)
	if err := validateWebhookInboundRequest(r); err != nil {
		return http.StatusUnauthorized, []byte(`{"success":false,"message":"nao autorizado"}`), nil
	}

	var req moniInboundRequest
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return bizErrJSON(fmt.Errorf("payload invalido"))
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			return bizErrJSON(fmt.Errorf("payload invalido"))
		}
	}
	headerFranq := ""
	if r != nil {
		headerFranq = strings.TrimSpace(r.Header.Get("X-Franqueado-Id"))
	}
	req.IDFranqueado = firstNonEmpty(strings.TrimSpace(req.IDFranqueado), headerFranq)
	if req.IDFranqueado == "" {
		return bizErrJSON(fmt.Errorf("id_franqueado obrigatorio"))
	}

	result, err := ProcessMoniInbound(ctx, req)
	if err != nil {
		return bizErrJSON(err)
	}
	return okJSON(result)
}

func validateWebhookInboundRequest(r *http.Request) error {
	if r == nil {
		return fmt.Errorf("request ausente")
	}
	expected := strings.TrimSpace(config.WebhookInboundKey)
	if expected == "" {
		return fmt.Errorf("webhook inbound nao configurado")
	}
	got := strings.TrimSpace(r.Header.Get("X-Webhook-Inbound-Key"))
	if got == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		const prefix = "Bearer "
		if strings.HasPrefix(auth, prefix) {
			got = strings.TrimSpace(auth[len(prefix):])
		}
	}
	if got != expected {
		return fmt.Errorf("chave invalida")
	}
	if !webhookInboundIPAllowed(r) {
		return fmt.Errorf("ip nao permitido")
	}
	return nil
}

func webhookInboundIPAllowed(r *http.Request) bool {
	allowed := strings.TrimSpace(config.WebhookAllowedIPs)
	if allowed == "" {
		return true
	}
	ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if ip != "" {
		if idx := strings.Index(ip, ","); idx >= 0 {
			ip = strings.TrimSpace(ip[:idx])
		}
	} else {
		ip = strings.TrimSpace(r.RemoteAddr)
		if idx := strings.LastIndex(ip, ":"); idx >= 0 {
			ip = ip[:idx]
		}
	}
	for _, part := range strings.Split(allowed, ",") {
		if strings.TrimSpace(part) == ip {
			return true
		}
	}
	return false
}

func ProcessMoniInbound(ctx context.Context, req moniInboundRequest) (map[string]any, error) {
	cfg, found, err := loadIntegracaoFranqueadoConfig(ctx, req.IDFranqueado)
	if err != nil {
		return nil, err
	}
	if !found || cfg == nil {
		return nil, fmt.Errorf("integracao nao encontrada")
	}
	if normalizeIntegracaoSistema(cfg.Sistema) != "moni" {
		return nil, fmt.Errorf("integracao nao e moni")
	}
	if !cfg.WebhookInboundAtivo {
		return map[string]any{
			"success": true,
			"ignored": true,
			"message": "webhook inbound desativado",
		}, nil
	}

	acao, ok := resolveMoniInboundAcao(cfg, req.Event)
	if !ok {
		insertIntegracaoLog(ctx, req.IDFranqueado, cfg.ID, 0, "moni-inbound", true, http.StatusOK,
			"evento ignorado", truncateStr(moniInboundResumo(req.Event), 500))
		return map[string]any{
			"success": true,
			"ignored": true,
			"message": "evento ignorado",
		}, nil
	}

	idCliente, err := resolveMoniInboundCliente(ctx, req.IDFranqueado, req.Event, cfg)
	if err != nil {
		insertIntegracaoLog(ctx, req.IDFranqueado, cfg.ID, 0, "moni-inbound", false, 0, err.Error(), moniInboundResumo(req.Event))
		return nil, err
	}
	if idCliente == "" {
		insertIntegracaoLog(ctx, req.IDFranqueado, cfg.ID, 0, "moni-inbound", false, 404,
			"cliente nao encontrado", moniInboundResumo(req.Event))
		return nil, fmt.Errorf("cliente nao encontrado")
	}

	particao := moniParticaoValor(*cfg, req.Event.Partition)
	cameras, err := ListCamerasMoniInbound(ctx, req.IDFranqueado, idCliente, particao, req.Event.Sector)
	if err != nil {
		insertIntegracaoLog(ctx, req.IDFranqueado, cfg.ID, 0, "moni-inbound", false, 0, err.Error(), moniInboundResumo(req.Event))
		return nil, err
	}
	if len(cameras) == 0 {
		insertIntegracaoLog(ctx, req.IDFranqueado, cfg.ID, 0, "moni-inbound", false, 404,
			"camera nao encontrada", moniInboundResumo(req.Event))
		return nil, fmt.Errorf("camera nao encontrada")
	}

	if acao == "toggle" {
		if moniInboundEstaAtivo(cameras) {
			acao = "desativar"
		} else {
			acao = "ativar"
		}
	}

	resultado := applyMoniInboundAcao(ctx, acao, cameras)
	sucesso := len(resultado["erros"].([]string)) == 0
	status := http.StatusOK
	msg := "processado"
	if !sucesso {
		msg = "processado com erros"
	}
	insertIntegracaoLog(ctx, req.IDFranqueado, cfg.ID, 0, "moni-inbound", sucesso, status, msg, moniInboundResumo(req.Event))

	out := map[string]any{
		"success":    sucesso,
		"acao":       acao,
		"id_cliente": idCliente,
		"particao":   particao,
		"resultado":  resultado,
	}
	if !sucesso {
		log.Printf("moni inbound parcial id_franqueado=%s cliente=%s acao=%s erros=%v",
			req.IDFranqueado, idCliente, acao, resultado["erros"])
	}
	return out, nil
}

func resolveMoniInboundAcao(cfg *integracaoConfig, ev moniInboundEvent) (string, bool) {
	codigo := moniInboundEventCode(ev)
	codigo = normalizeIntegracaoCodigo(codigo)
	if codigo == "" {
		return "", false
	}
	armar := normalizeIntegracaoCodigo(defaultStr(cfg.CodigoEventoArmar, "130"))
	desarmar := normalizeIntegracaoCodigo(defaultStr(cfg.CodigoEventoDesarmar, "131"))
	if armar == desarmar && codigo == armar {
		return "toggle", true
	}
	switch codigo {
	case armar:
		return "ativar", true
	case desarmar:
		return "desativar", true
	default:
		return "", false
	}
}

// moniInboundEstaAtivo indica se alguma camera analitica do escopo esta retomada (nao pausada).
func moniInboundEstaAtivo(cameras []map[string]any) bool {
	for _, cam := range cameras {
		if moniInboundCameraAtiva(cam) {
			return true
		}
	}
	return false
}

func moniInboundCameraAtiva(cam map[string]any) bool {
	plano := strVal(cam, "plano")
	_, captura, _ := cvgPlanoTipo(plano)
	if !captura {
		return false
	}
	pausado := boolVal(cam, "analitico_pausado")
	if pausado == nil {
		return true
	}
	return !*pausado
}

func moniInboundEventCode(ev moniInboundEvent) string {
	if s := strings.TrimSpace(ev.EventSecondCode); s != "" {
		return s
	}
	if ev.EventCode > 0 {
		return strconv.Itoa(ev.EventCode)
	}
	if ev.EventType == 3 {
		return "130"
	}
	if ev.EventType == 2 {
		return "131"
	}
	return ""
}

func normalizeIntegracaoCodigo(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if n, err := strconv.Atoi(s); err == nil {
		return strconv.Itoa(n)
	}
	return strings.ToUpper(s)
}

func resolveMoniInboundCliente(ctx context.Context, idFranqueado string, ev moniInboundEvent, cfg *integracaoConfig) (string, error) {
	codigos := moniCustomerCodeVariants(ev.CustomerCode)
	empresas := moniCompanyCodeVariants(ev, cfg)
	for _, codigo := range codigos {
		for _, empresa := range empresas {
			idCliente, err := GetClienteByCodigoInterno(ctx, idFranqueado, codigo, empresa)
			if err != nil {
				return "", err
			}
			if idCliente != "" {
				return idCliente, nil
			}
		}
	}
	return "", nil
}

func moniCustomerCodeVariants(code int) []string {
	raw := strconv.Itoa(code)
	padded := padDigits(raw, 4)
	out := []string{}
	seen := map[string]bool{}
	for _, s := range []string{padded, raw, strings.TrimLeft(padded, "0")} {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func moniCompanyCode(ev moniInboundEvent) string {
	if ev.Company > 0 {
		return strconv.Itoa(ev.Company)
	}
	return ""
}

func moniCompanyCodeVariants(ev moniInboundEvent, cfg *integracaoConfig) []string {
	raw := moniCompanyCode(ev)
	if raw == "" && cfg != nil {
		raw = strings.TrimSpace(cfg.EmpresaCodigo)
	}
	return moniEmpresaCodeVariants(raw)
}

// moniEmpresaCodeVariants gera formas equivalentes (1, 01, 001) para bater com CodEmpresa no MySQL.
func moniEmpresaCodeVariants(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{""}
	}
	out := []string{}
	seen := map[string]bool{}
	for _, s := range []string{
		padDigits(raw, 3),
		padDigits(raw, 2),
		raw,
		strings.TrimLeft(padDigits(raw, 3), "0"),
	} {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func moniInboundResumo(ev moniInboundEvent) string {
	return fmt.Sprintf("cliente=%d empresa=%d particao=%s setor=%d codigo=%s",
		ev.CustomerCode, ev.Company, strings.TrimSpace(ev.Partition), ev.Sector, moniInboundEventCode(ev))
}

func applyMoniInboundAcao(ctx context.Context, acao string, cameras []map[string]any) map[string]any {
	erros := []string{}
	camerasOK := []int{}
	dispositivos := []map[string]any{}
	dispFeitos := ","

	for _, cam := range cameras {
		camID := intVal(cam, "id")
		plano := strVal(cam, "plano")
		tipo, captura, _ := cvgPlanoTipo(plano)
		if !captura {
			continue
		}

		if acao == "ativar" {
			if tipo == "armado" {
				idDisp := strVal(cam, "id_dispositivo")
				if idDisp != "" {
					chave := "," + idDisp + ","
					if !strings.Contains(dispFeitos, chave) {
						dispFeitos += idDisp + ","
						d, err := cvgDispositivoArmar(ctx, idDisp, cvgAcaoArmar)
						if err != nil {
							erros = append(erros, "armar:"+idDisp+":"+err.Error())
						} else {
							dispositivos = append(dispositivos, d)
						}
					}
				}
			}
			if err := cvgCameraPausar(ctx, camID, false); err != nil {
				erros = append(erros, fmt.Sprintf("despausar:%d:%s", camID, err.Error()))
				continue
			}
			camerasOK = append(camerasOK, camID)
			continue
		}

		if err := cvgCameraPausar(ctx, camID, true); err != nil {
			erros = append(erros, fmt.Sprintf("pausar:%d:%s", camID, err.Error()))
			continue
		}
		camerasOK = append(camerasOK, camID)

		if tipo == "armado" {
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
				erros = append(erros, "desarmar:"+idDisp+":"+err.Error())
			} else {
				dispositivos = append(dispositivos, d)
			}
		}
	}

	return map[string]any{
		"acao":         acao,
		"cameras":      camerasOK,
		"dispositivos": dispositivos,
		"erros":        erros,
	}
}

func ListCamerasMoniInbound(ctx context.Context, idFranqueado, idCliente, particao string, sector int) ([]map[string]any, error) {
	particao, zona := prepareSetorQueryParams(particao, "")
	if sector > 0 {
		_, zona = prepareSetorQueryParams(particao, strconv.Itoa(sector))
	}

	db, err := DB()
	if err != nil {
		return nil, err
	}

	query := `
SELECT ` + cameraSelectCols + `
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.id_cliente = $1
  AND c.ativo = TRUE
  AND LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE(c.particao, '')), '[^0-9]', '', 'g'), ''), 2, '0')
    = LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE($2, '')), '[^0-9]', '', 'g'), ''), 2, '0')`
	args := []any{idCliente, particao}
	n := 3

	if strings.TrimSpace(idFranqueado) != "" {
		query += fmt.Sprintf(" AND c.id_franqueado = $%d", n)
		args = append(args, idFranqueado)
		n++
	}

	if sector > 0 && strings.TrimSpace(zona) != "" {
		query += fmt.Sprintf(` AND (
    TRIM(COALESCE(c.zonauser, '')) = TRIM(COALESCE($%d, ''))
    OR LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE(c.zonauser, '')), '[^0-9]', '', 'g'), ''), 3, '0')
      = LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE($%d, '')), '[^0-9]', '', 'g'), ''), 3, '0')
    OR TRIM(COALESCE(c.setor, '')) = TRIM(COALESCE($%d, ''))
    OR LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE(c.setor, '')), '[^0-9]', '', 'g'), ''), 3, '0')
      = LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE($%d, '')), '[^0-9]', '', 'g'), ''), 3, '0')
  )`, n, n, n, n)
		args = append(args, zona)
	}

	query += " ORDER BY c.id ASC"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanCameraRows(rows)
}

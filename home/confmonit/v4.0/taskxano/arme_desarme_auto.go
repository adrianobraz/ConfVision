package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	acaoDesarmar = 0
	acaoArmar    = 1
)

type agendaArmeDesarmeResponse struct {
	Dados struct {
		Desarme []agendaArmeDesarmeItem `json:"desarme"`
		Arme    []agendaArmeDesarmeItem `json:"arme"`
	} `json:"dados"`
}

type agendaArmeDesarmeItem struct {
	ID              int    `json:"id"`
	WhatsappEvento  int    `json:"whatsappeventocad_id"`
	IDCliente       string `json:"IdCliente"`
	IDFranqueado    string `json:"IdFranqueado"`
	Whatsapp        string `json:"whatsapp"`
	Nome            string `json:"nome"`
	IDDispositivo   string `json:"idDispositivo"`
	NomeDispositivo string `json:"NomeDispositivo"`
}

type dispositivoGetDadosResponse struct {
	Status string                `json:"status"`
	Dados  dispositivoGetDadosVO `json:"dados"`
}

type dispositivoGetDadosVO struct {
	IDDispositivo string `json:"idDispositivo"`
	Nome          string `json:"nome"`
	Particao      string `json:"particao"`
	Senha         string `json:"senha"`
	Armado        string `json:"armado"`
}

type comandoResponse struct {
	Status string `json:"status"`
}

type webLogarResponse struct {
	Status string `json:"status"`
	Dados  any    `json:"dados"`
}

var armeDesarmeAutoTokenCache struct {
	mu       sync.Mutex
	token    string
	expireAt time.Time
}

func runArmeDesarmeAutoLoop(client *http.Client, cfg Config, shutdown <-chan struct{}) {
	ticker := time.NewTicker(cfg.ArmeDesarmeAutoInterval)
	defer ticker.Stop()

	runArmeDesarmeAutoOnce(client, cfg)
	for {
		select {
		case <-shutdown:
			log.Printf("arme_desarme_auto: loop encerrado")
			return
		case <-ticker.C:
			runArmeDesarmeAutoOnce(client, cfg)
		}
	}
}

func runArmeDesarmeAutoOnce(client *http.Client, cfg Config) {
	if !atomic.CompareAndSwapInt32(&runningArmeDesarmeAuto, 0, 1) {
		log.Printf("arme_desarme_auto: execucao anterior em andamento, pulando ciclo")
		return
	}
	defer atomic.StoreInt32(&runningArmeDesarmeAuto, 0)

	start := time.Now()
	attempts := cfg.Retries + 1
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			backoff := backoffForAttempt(attempt)
			log.Printf("arme_desarme_auto retry=%d backoff=%s", attempt-1, backoff)
			time.Sleep(backoff)
		}
		err := processArmeDesarmeAutoRoutine(client, cfg)
		if err == nil {
			log.Printf("arme_desarme_auto sucesso | attempt=%d/%d | duracao=%s", attempt, attempts, time.Since(start))
			return
		}
		lastErr = err
		log.Printf("arme_desarme_auto falha | attempt=%d/%d | erro=%v", attempt, attempts, err)
	}
	log.Printf("arme_desarme_auto erro final | duracao=%s | erro=%v", time.Since(start), lastErr)
}

func processArmeDesarmeAutoRoutine(client *http.Client, cfg Config) error {
	if strings.TrimSpace(cfg.ArmeDesarmeAutoQueryURL) == "" {
		return fmt.Errorf("ARME_DESARME_AUTO_QUERY_URL nao configurada")
	}
	day := diaSemanaAgenda(time.Now())
	agenda, err := buscarAgendaArmeDesarme(client, cfg.ArmeDesarmeAutoQueryURL, day)
	if err != nil {
		return fmt.Errorf("consulta agenda Xano: %w", err)
	}
	token, err := getArmeDesarmeAutoToken(client, cfg)
	if err != nil {
		return fmt.Errorf("webLogar token: %w", err)
	}

	log.Printf("arme_desarme_auto: agenda dia=%d | desarme=%d | arme=%d", day, len(agenda.Dados.Desarme), len(agenda.Dados.Arme))
	for _, item := range agenda.Dados.Desarme {
		processarItemArmeDesarme(client, cfg, token, item, acaoDesarmar)
	}
	for _, item := range agenda.Dados.Arme {
		processarItemArmeDesarme(client, cfg, token, item, acaoArmar)
	}
	return nil
}

func processarItemArmeDesarme(client *http.Client, cfg Config, token string, item agendaArmeDesarmeItem, acao int) {
	idDisp := normalizaID(item.IDDispositivo)
	whatsapp := normalizaNumero(item.Whatsapp)
	nomeDisp := nomeDispositivoItem(item)

	if idDisp == "" {
		log.Printf("arme_desarme_auto: item sem idDispositivo valido | item_id=%d", item.ID)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp, mensagemResultado(cfg, acao, nomeDisp, idDisp, "INDEFINIDO", "FALHA", "idDispositivo vazio"))
		}
		return
	}

	disp, err := buscarDispositivoByID(client, cfg, token, idDisp)
	if err != nil {
		log.Printf("arme_desarme_auto: erro getDadosById id=%s: %v", idDisp, err)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp, mensagemResultado(cfg, acao, nomeDisp, idDisp, "INDEFINIDO", "FALHA", "erro ao consultar dispositivo: "+err.Error()))
		}
		return
	}

	if acao == acaoDesarmar && strings.ToUpper(strings.TrimSpace(disp.Armado)) != "S" {
		log.Printf("arme_desarme_auto: dispositivo %s ja desarmado", idDisp)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp,
				mensagemResultado(cfg, acao, nomeDisp, idDisp, statusAtualTexto(disp.Armado), "NAO EXECUTADO", "dispositivo ja estava desarmado"))
		}
		return
	}
	if acao == acaoArmar && strings.ToUpper(strings.TrimSpace(disp.Armado)) != "N" {
		log.Printf("arme_desarme_auto: dispositivo %s ja armado", idDisp)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp,
				mensagemResultado(cfg, acao, nomeDisp, idDisp, statusAtualTexto(disp.Armado), "NAO EXECUTADO", "dispositivo ja estava armado"))
		}
		return
	}

	if err := enviarComandoArmeDesarme(client, cfg, disp, acao); err != nil {
		log.Printf("arme_desarme_auto: erro comando id=%s acao=%d: %v", idDisp, acao, err)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp,
				mensagemResultado(cfg, acao, nomeDisp, idDisp, statusAtualTexto(disp.Armado), "FALHA", "erro ao executar comando: "+err.Error()))
		}
		return
	}

	if cfg.ArmeDesarmeAutoStatusDelay > 0 {
		time.Sleep(cfg.ArmeDesarmeAutoStatusDelay)
	}

	after, err := buscarDispositivoByID(client, cfg, token, idDisp)
	if err != nil {
		log.Printf("arme_desarme_auto: comando enviado, sem confirmar status id=%s: %v", idDisp, err)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp,
				mensagemResultado(cfg, acao, nomeDisp, idDisp, "INDEFINIDO", "FALHA", "comando enviado, mas nao foi possivel confirmar status: "+err.Error()))
		}
		return
	}

	ok := (acao == acaoDesarmar && strings.ToUpper(strings.TrimSpace(after.Armado)) == "N") ||
		(acao == acaoArmar && strings.ToUpper(strings.TrimSpace(after.Armado)) == "S")
	if ok {
		log.Printf("arme_desarme_auto: sucesso id=%s acao=%d", idDisp, acao)
		if whatsapp != "" {
			_ = enviarWhatsappUazapi(client, cfg, whatsapp,
				mensagemResultado(cfg, acao, nomeDisp, idDisp, statusAtualTexto(after.Armado), "SUCESSO", ""))
		}
		return
	}

	log.Printf("arme_desarme_auto: comando sem efeito id=%s acao=%d status_final=%s", idDisp, acao, after.Armado)
	if whatsapp != "" {
		_ = enviarWhatsappUazapi(client, cfg, whatsapp,
			mensagemResultado(cfg, acao, nomeDisp, idDisp, statusAtualTexto(after.Armado), "FALHA", "sem confirmacao de alteracao de status apos comando"))
	}
}

func buscarAgendaArmeDesarme(client *http.Client, rawURL string, diaSemana int) (agendaArmeDesarmeResponse, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return agendaArmeDesarmeResponse{}, err
	}
	q := u.Query()
	q.Set("DiaSemana", strconv.Itoa(diaSemana))
	u.RawQuery = q.Encode()

	body, status, err := doJSONRequest(client, http.MethodGet, u.String(), nil)
	if err != nil {
		return agendaArmeDesarmeResponse{}, err
	}
	if status < 200 || status >= 300 {
		return agendaArmeDesarmeResponse{}, fmt.Errorf("status=%d body=%s", status, compactBody(body))
	}

	var out agendaArmeDesarmeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return agendaArmeDesarmeResponse{}, err
	}
	return out, nil
}

func buscarDispositivoByID(client *http.Client, cfg Config, token, idDisp string) (dispositivoGetDadosVO, error) {
	if strings.TrimSpace(cfg.ArmeDesarmeAutoDispositivoURL) == "" {
		return dispositivoGetDadosVO{}, fmt.Errorf("ARME_DESARME_AUTO_DISPOSITIVO_GET_URL nao configurada")
	}

	payload := map[string]any{
		"idDispositivo": idDisp,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + strings.TrimSpace(token),
	}
	body, status, err := doJSONRequestWithHeaders(client, http.MethodPost, cfg.ArmeDesarmeAutoDispositivoURL, payload, headers)
	if err != nil {
		return dispositivoGetDadosVO{}, err
	}
	if status < 200 || status >= 300 {
		return dispositivoGetDadosVO{}, fmt.Errorf("status=%d body=%s", status, compactBody(body))
	}

	var resp dispositivoGetDadosResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return dispositivoGetDadosVO{}, err
	}
	return resp.Dados, nil
}

func enviarComandoArmeDesarme(client *http.Client, cfg Config, disp dispositivoGetDadosVO, acao int) error {
	if strings.TrimSpace(cfg.ArmeDesarmeAutoComandoURL) == "" {
		return fmt.Errorf("ARME_DESARME_AUTO_COMANDO_URL nao configurada")
	}

	numero, _ := strconv.Atoi(strings.TrimSpace(disp.Particao))
	payload := map[string]any{
		"idDispositivo": disp.IDDispositivo,
		"numero":       numero,
		"usuario":      "000",
		"acao":         acao,
		"senha":        disp.Senha,
		"senhaWeb":     cfg.ArmeDesarmeAutoComandoSenhaWeb,
	}
	body, status, err := doJSONRequest(client, http.MethodPost, cfg.ArmeDesarmeAutoComandoURL, payload)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("status=%d body=%s", status, compactBody(body))
	}

	var resp comandoResponse
	if err := json.Unmarshal(body, &resp); err == nil {
		st := strings.TrimSpace(strings.ToUpper(resp.Status))
		if st != "" && st != "OK" {
			return fmt.Errorf("resposta comando=%s", resp.Status)
		}
	}
	return nil
}

func enviarWhatsappUazapi(client *http.Client, cfg Config, number, textMensagem string) error {
	if strings.TrimSpace(cfg.ArmeDesarmeAutoUazapiURL) == "" {
		return fmt.Errorf("ARME_DESARME_AUTO_UAZAPI_URL nao configurada")
	}
	if strings.TrimSpace(cfg.ArmeDesarmeAutoUazapiToken) == "" {
		return fmt.Errorf("ARME_DESARME_AUTO_UAZAPI_TOKEN nao configurado")
	}
	number = normalizaNumero(number)
	if number == "" {
		return nil
	}

	payload := map[string]string{
		"number": number,
		"text":   strings.TrimSpace(textMensagem),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.ArmeDesarmeAutoUazapiURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", cfg.ArmeDesarmeAutoUazapiToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}
	return nil
}

func getArmeDesarmeAutoToken(client *http.Client, cfg Config) (string, error) {
	armeDesarmeAutoTokenCache.mu.Lock()
	defer armeDesarmeAutoTokenCache.mu.Unlock()

	now := time.Now()
	if strings.TrimSpace(armeDesarmeAutoTokenCache.token) != "" && now.Before(armeDesarmeAutoTokenCache.expireAt.Add(-cfg.ArmeDesarmeAutoTokenRefresh)) {
		return armeDesarmeAutoTokenCache.token, nil
	}

	if strings.TrimSpace(cfg.ArmeDesarmeAutoWebLogarURL) == "" {
		return "", fmt.Errorf("ARME_DESARME_AUTO_WEBLOGAR_URL nao configurada")
	}
	if strings.TrimSpace(cfg.ArmeDesarmeAutoWebLogarSenha) == "" {
		return "", fmt.Errorf("ARME_DESARME_AUTO_WEBLOGAR_SENHA nao configurada")
	}

	payload := map[string]string{
		"senha": cfg.ArmeDesarmeAutoWebLogarSenha,
	}
	body, status, err := doJSONRequest(client, http.MethodPost, cfg.ArmeDesarmeAutoWebLogarURL, payload)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("status=%d body=%s", status, compactBody(body))
	}

	var resp webLogarResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	token := extractTokenFromAny(resp.Dados)
	if token == "" {
		return "", fmt.Errorf("token nao encontrado na resposta webLogar")
	}

	exp := jwtExpireAt(token)
	if exp.IsZero() {
		exp = time.Now().Add(12 * time.Hour)
	}
	armeDesarmeAutoTokenCache.token = token
	armeDesarmeAutoTokenCache.expireAt = exp
	return token, nil
}

func doJSONRequestWithHeaders(client *http.Client, method, rawURL string, payload any, headers map[string]string) ([]byte, int, error) {
	var bodyReader *bytes.Reader
	if payload == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

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

func extractTokenFromAny(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		for _, key := range []string{"token", "Token", "access_token", "jwt"} {
			if vv, ok := x[key]; ok {
				if t := extractTokenFromAny(vv); t != "" {
					return t
				}
			}
		}
	case []any:
		for _, el := range x {
			if t := extractTokenFromAny(el); t != "" {
				return t
			}
		}
	}
	return ""
}

func jwtExpireAt(token string) time.Time {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}
	}
	segment := parts[1]
	data, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return time.Time{}
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return time.Time{}
	}
	expVal, ok := payload["exp"]
	if !ok {
		return time.Time{}
	}
	switch v := expVal.(type) {
	case float64:
		return time.Unix(int64(v), 0)
	case int64:
		return time.Unix(v, 0)
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return time.Time{}
		}
		return time.Unix(n, 0)
	default:
		return time.Time{}
	}
}

func tituloAcao(acao int) string {
	if acao == acaoDesarmar {
		return "DESARME"
	}
	return "ARME"
}

func acaoTextoCapitalizado(acao int) string {
	if acao == acaoDesarmar {
		return "Desarmar"
	}
	return "Armar"
}

func mensagemResultado(cfg Config, acao int, nomeDisp, idDisp, statusAtual, resultado, retorno string) string {
	msg := fmt.Sprintf(
		"Agenda automatica - %s\nSistema tentou %s\nDispositivo: %s\nStatus Atual: %s\nResultado: %s",
		tituloAcao(acao),
		acaoTextoCapitalizado(acao),
		nomeDisp,
		strings.TrimSpace(statusAtual),
		strings.TrimSpace(resultado),
	)
	if strings.TrimSpace(retorno) != "" {
		msg += "\nRetorno: " + strings.TrimSpace(retorno)
	}
	msg += "\nConfira no link abaixo\n" + linkFinalizaEvento(cfg, idDisp)
	return msg
}

func statusAtualTexto(armado string) string {
	v := strings.ToUpper(strings.TrimSpace(armado))
	switch v {
	case "S":
		return "ARMADO"
	case "N":
		return "DESARMADO"
	default:
		return "INDEFINIDO"
	}
}

func linkFinalizaEvento(cfg Config, idDisp string) string {
	base := strings.TrimSpace(cfg.ArmeDesarmeAutoLinkBaseURL)
	if base == "" {
		base = "https://admmonitoramento.com.br/finalizaevento/?dispositivo="
	}
	return strings.TrimSpace(base) + strings.TrimSpace(idDisp)
}

func diaSemanaAgenda(t time.Time) int {
	wd := t.Weekday()
	if wd == time.Sunday {
		return 7
	}
	return int(wd)
}

func textoAcao(acao int) string {
	if acao == acaoDesarmar {
		return "desarmar"
	}
	return "armar"
}

func normalizaID(v string) string {
	s := strings.TrimSpace(v)
	if s == "" || strings.EqualFold(s, "null") || strings.EqualFold(s, "nil") {
		return ""
	}
	return s
}

func normalizaNumero(v string) string {
	var b strings.Builder
	for _, ch := range strings.TrimSpace(v) {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func nomeDispositivoItem(item agendaArmeDesarmeItem) string {
	if n := strings.TrimSpace(item.NomeDispositivo); n != "" && !strings.EqualFold(n, "null") {
		return n
	}
	if n := strings.TrimSpace(item.Nome); n != "" && !strings.EqualFold(n, "null") {
		return n
	}
	return strings.TrimSpace(item.IDDispositivo)
}

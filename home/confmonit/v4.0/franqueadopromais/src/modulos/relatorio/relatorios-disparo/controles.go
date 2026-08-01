package relatoriosDisparo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
)

var httpCD = &http.Client{Timeout: 90 * time.Second}

var Rotas = []auxiliar.Rota{
	{URI: "/carregar-relatorio-eventos-pendentes", Metodo: http.MethodGet, Funcao: page("Eventos Pendentes", "eventos-pendentes"), Aberto: false},
	{URI: "/carregar-relatorio-finalizados-robo", Metodo: http.MethodGet, Funcao: page("Eventos Finalizados por Robô", "finalizados-robo"), Aberto: false},
	{URI: "/carregar-relatorio-finalizados-bot", Metodo: http.MethodGet, Funcao: page("Finalizados Automaticamente pelo Bot", "finalizados-bot"), Aberto: false},
	{URI: "/carregar-relatorio-ligacao-historico", Metodo: http.MethodGet, Funcao: page("Ligação Histórico", "ligacao-historico"), Aberto: false},
	{URI: "/carregar-relatorio-sms-historico", Metodo: http.MethodGet, Funcao: page("SMS Histórico", "sms-historico"), Aberto: false},
	{URI: "/carregar-relatorio-eventos-falhas", Metodo: http.MethodGet, Funcao: page("Eventos com Falhas", "eventos-falhas"), Aberto: false},
	{URI: "/carregar-relatorio-whatsapp-enviados", Metodo: http.MethodGet, Funcao: page("WhatsApp Enviados", "whatsapp-enviados"), Aberto: false},
	{URI: "/carregar-relatorio-ligacoes-cd", Metodo: http.MethodGet, Funcao: page("Ligações", "ligacoes"), Aberto: false},
	{URI: "/carregar-relatorio-sms", Metodo: http.MethodGet, Funcao: page("SMS", "sms"), Aberto: false},
	{URI: "/carregar-relatorio-fila-envio", Metodo: http.MethodGet, Funcao: page("Eventos para Enviar ou Ligação", "fila-envio"), Aberto: false},
	{URI: "/carregar-relatorio-ligacao-erros", Metodo: http.MethodGet, Funcao: page("Ligações com Erros e Tentativas", "ligacao-erros"), Aberto: false},
	{URI: "/carregar-relatorio-fila-ligacao", Metodo: http.MethodGet, Funcao: page("Filas de Ligações", "fila-ligacao"), Aberto: false},
	{URI: "/carregar-relatorio-cd-unificado", Metodo: http.MethodGet, Funcao: pageUnificado, Aberto: false},

	{URI: "/cdEventosPendentesListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_eventos_pendentes"), Aberto: false},
	{URI: "/cdFinalizadosRoboListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_finalizados_robo"), Aberto: false},
	{URI: "/cdFinalizadosBotListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_finalizados_bot"), Aberto: false},
	{URI: "/cdLigacaoHistoricoListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_ligacao_historico"), Aberto: false},
	{URI: "/ligacaoHistoricoListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_ligacao_historico"), Aberto: false},
	{URI: "/cdSmsHistoricoListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_sms_historico"), Aberto: false},
	{URI: "/cdEventosFalhasListar", Metodo: http.MethodPost, Funcao: cdEventosFalhasListar, Aberto: false},
	{URI: "/cdWhatsappEnviadosListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_whatsapp_enviados"), Aberto: false},
	{URI: "/cdLigacoesListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_ligacoes"), Aberto: false},
	{URI: "/cdSmsListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_sms"), Aberto: false},
	{URI: "/cdFilaEnvioListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_fila_envio"), Aberto: false},
	{URI: "/cdLigacaoErrosListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_ligacao_erros"), Aberto: false},
	{URI: "/cdFilaLigacaoListar", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_fila_ligacao"), Aberto: false},
	{URI: "/cdUnificadoLista", Metodo: http.MethodPost, Funcao: proxyCD("/fp_cd_unificado_lista"), Aberto: false},
	{URI: "/cdUnificadoDetalhe", Metodo: http.MethodPost, Funcao: cdUnificadoDetalhe, Aberto: false},

	{URI: "/ligacaoHistoricoAudio", Metodo: http.MethodGet, Funcao: LigacaoHistoricoAudio, Aberto: false},
}

func pageUnificado(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Centro de Operações"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-central-disparos"
	auxiliar.ExecutarTemplate(w, "relatorio-cd-unificado.html", d)
}

func cdUnificadoDetalhe(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	base := config.XanoApiCentralDisparos
	if base == "" {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("XANO_API_CENTRAL_DISPAROS nao configurado"))
		return
	}

	req, erro := http.NewRequest(http.MethodPost, base+"/fp_cd_unificado_detalhe", bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, erro := httpCD.Do(req)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}
	defer resp.Body.Close()

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(corpo)
		return
	}

	var out map[string]interface{}
	if erro := json.Unmarshal(corpo, &out); erro != nil {
		auxiliar.RespostaAPP(w, corpo)
		return
	}

	if ev, ok := out["evento"].(map[string]interface{}); ok && ev != nil {
		idDisp := strings.TrimSpace(fmt.Sprint(ev["idDispositivo"]))
		if idDisp != "" && idDisp != "<nil>" {
			if nome := buscarNomeDispositivo(r, idDisp); nome != "" {
				ev["DispNome"] = nome
			}
		}
		out["evento"] = ev
	}

	respOut, erro := json.Marshal(out)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}
	auxiliar.RespostaAPP(w, respOut)
}

func page(nomeTela, reportKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var d auxiliar.Pagina
		d.TituloSite = config.TituloSite
		d.NomeTela = nomeTela
		d.LogoMarca = "logo2Id6.png"
		d.LinkRetorno = "/carregar-menu-central-disparos"
		d.LinkJs = reportKey
		auxiliar.ExecutarTemplate(w, "relatorio-cd.html", d)
	}
}

func proxyCD(xanoPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, erro := io.ReadAll(r.Body)
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		base := config.XanoApiCentralDisparos
		if base == "" {
			auxiliar.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("XANO_API_CENTRAL_DISPAROS nao configurado"))
			return
		}

		req, erro := http.NewRequest(http.MethodPost, base+xanoPath, bytes.NewBuffer(body))
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, erro := httpCD.Do(req)
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
			return
		}
		defer resp.Body.Close()

		corpo, erro := io.ReadAll(resp.Body)
		if erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
		if resp.StatusCode >= 400 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			w.Write(corpo)
			return
		}
		auxiliar.RespostaAPP(w, corpo)
	}
}

type cdListaResp struct {
	Status  string                   `json:"status"`
	Dados   []map[string]interface{} `json:"dados"`
	Total   int                      `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	HasMore bool                     `json:"hasMore"`
}

func cdEventosFalhasListar(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	base := config.XanoApiCentralDisparos
	if base == "" {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("XANO_API_CENTRAL_DISPAROS nao configurado"))
		return
	}

	req, erro := http.NewRequest(http.MethodPost, base+"/fp_cd_eventos_falhas", bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, erro := httpCD.Do(req)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}
	defer resp.Body.Close()

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(corpo)
		return
	}

	var out cdListaResp
	if erro := json.Unmarshal(corpo, &out); erro != nil {
		auxiliar.RespostaAPP(w, corpo)
		return
	}

	for _, item := range out.Dados {
		idDisp := strings.TrimSpace(fmt.Sprint(item["idDispositivo"]))
		if idDisp == "" || idDisp == "<nil>" {
			continue
		}
		if nome := buscarNomeDispositivo(r, idDisp); nome != "" {
			item["DispNome"] = nome
		} else {
			item["DispNome"] = idDisp
		}
	}

	respOut, erro := json.Marshal(out)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}
	auxiliar.RespostaAPP(w, respOut)
}

func buscarNomeDispositivo(r *http.Request, idDispositivo string) string {
	urlV4 := fmt.Sprintf("%s/v4/dispositivo/getDadosById", config.ApiUrl)
	payload, _ := json.Marshal(map[string]interface{}{
		"idDispositivo": idDispositivo,
	})

	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, urlV4, bytes.NewBuffer(payload))
	if erro != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ""
	}

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		return ""
	}

	var out struct {
		Status string `json:"status"`
		Dados  struct {
			Nome string `json:"nome"`
		} `json:"dados"`
	}
	if json.Unmarshal(corpo, &out) != nil {
		return ""
	}
	return strings.TrimSpace(out.Dados.Nome)
}

func LigacaoHistoricoAudio(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	if id <= 0 {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id invalido"))
		return
	}
	baixar := r.URL.Query().Get("dl") == "1"

	cookie, erro := seguranca.LerCookies(r)
	if erro != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	idFra := cookie["idFranqueado"]
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, fmt.Errorf("sessao sem franqueado"))
		return
	}

	if config.ElevenLabsAPIKey == "" {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("ELEVENLABS_API_KEY nao configurada"))
		return
	}

	convID, erro := obterConversationIDAudio(id, idFra)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	urlEL := "https://api.elevenlabs.io/v1/convai/conversations/" + url.PathEscape(convID) + "/audio"
	req, erro := http.NewRequest(http.MethodGet, urlEL, nil)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	req.Header.Set("xi-api-key", config.ElevenLabsAPIKey)

	resp, erro := httpCD.Do(req)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("falha ao chamar ElevenLabs: %v", erro))
		return
	}
	defer resp.Body.Close()

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}
	if resp.StatusCode >= 400 {
		msg := string(corpo)
		if len(msg) > 300 {
			msg = msg[:300]
		}
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("ElevenLabs HTTP %d: %s", resp.StatusCode, msg))
		return
	}
	if len(corpo) == 0 {
		auxiliar.RespostaErro(w, http.StatusNotFound, fmt.Errorf("audio vazio na ElevenLabs"))
		return
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || strings.Contains(ct, "json") {
		ct = "audio/mpeg"
	}
	fn := fmt.Sprintf("ligacao-%d.mp3", id)

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", strconv.Itoa(len(corpo)))
	w.Header().Set("Cache-Control", "private, max-age=300")
	if baixar {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fn+"\"")
	} else {
		w.Header().Set("Content-Disposition", "inline; filename=\""+fn+"\"")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(corpo)
}

func obterConversationIDAudio(id int, idFranqueado string) (string, error) {
	base := config.XanoApiCentralDisparos
	if base == "" {
		return "", fmt.Errorf("XANO_API_CENTRAL_DISPAROS nao configurado")
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"id":           id,
		"idFranqueado": idFranqueado,
	})

	req, erro := http.NewRequest(http.MethodPost, base+"/fp_cd_ligacao_audio", bytes.NewBuffer(payload))
	if erro != nil {
		return "", erro
	}
	req.Header.Set("Content-Type", "application/json")

	resp, erro := httpCD.Do(req)
	if erro != nil {
		return "", erro
	}
	defer resp.Body.Close()

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		return "", erro
	}
	if resp.StatusCode >= 400 {
		msg := extrairMsgXano(corpo)
		if msg == "" {
			msg = "erro ao validar audio no Xano"
		}
		return "", fmt.Errorf("%s", msg)
	}

	// Preferencia: conversation_id direto; fallback audio_base64 legado
	var out struct {
		ConversationID string `json:"conversation_id"`
		AudioBase64    string `json:"audio_base64"`
	}
	if erro := json.Unmarshal(corpo, &out); erro != nil {
		return "", fmt.Errorf("resposta Xano invalida")
	}
	if out.ConversationID != "" {
		return out.ConversationID, nil
	}

	// Fallback: se ainda vier base64 antigo, devolve marker especial? Nao — so conversation_id
	if out.AudioBase64 != "" {
		return "", fmt.Errorf("endpoint Xano ainda retorna audio_base64; faca push de fp_cd_ligacao_audio")
	}
	return "", fmt.Errorf("conversation_id ausente na resposta do Xano")
}

func extrairMsgXano(corpo []byte) string {
	var m map[string]interface{}
	if json.Unmarshal(corpo, &m) != nil {
		return string(corpo)
	}
	for _, k := range []string{"message", "erro", "error", "payload"} {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			}
		}
	}
	return ""
}

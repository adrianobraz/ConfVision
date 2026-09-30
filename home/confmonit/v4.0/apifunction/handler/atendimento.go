package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"apifunction/confservice"
	"apifunction/pgatendimento"
)

func (h *Handler) OpsAtendimentoResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	q := r.URL.Query()
	out, err := pgatendimento.Resolve(r.Context(),
		q.Get("id_franqueado"), q.Get("id_cliente"), q.Get("recurso"), q.Get("cti_grupo"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ativo":        out.Ativo,
		"ia_bloqueado": out.IABloqueado,
		"recurso":      out.Recurso,
		"fail_open":    out.FailOpen,
		"cti_grupo":    out.CtiGrupo,
		"emergencia":   out.Emergencia,
	})
}

func (h *Handler) OpsAtendimentoParceiroVinculo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	q := r.URL.Query()
	out, err := pgatendimento.ResolveParceiroVinculo(r.Context(),
		q.Get("id_franqueado"), q.Get("id_cliente"), q.Get("cti_grupo"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) OpsAtendimentoParceiroLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	limite := 50
	if q := r.URL.Query().Get("limite"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limite = n
		}
	}
	lista, err := pgatendimento.ListParceiroEnvioLog(r.Context(), idFra, limite)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) OpsAtendimentoParceiroDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var in pgatendimento.ParceiroDispatchInput
	if err := json.Unmarshal(raw, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	out, err := pgatendimento.DispatchEventoParceiro(r.Context(), in)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) OpsAtendimentoParceiroEnvio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	out, err := pgatendimento.ProcessarEnvioParceiro(context.Background(), payload)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) OpsAtendimentoParceiroEventosList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	q := r.URL.Query()
	idFra := strings.TrimSpace(q.Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	limite := 50
	if n, err := strconv.Atoi(strings.TrimSpace(q.Get("limite"))); err == nil && n > 0 {
		limite = n
	}
	out, err := confservice.ListEventosFranqueado(
		idFra,
		strings.TrimSpace(q.Get("de")),
		strings.TrimSpace(q.Get("ate")),
		strings.TrimSpace(q.Get("status")),
		strings.TrimSpace(q.Get("id_cliente")),
		limite,
	)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) OpsAtendimentoCreditoPodeLigar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	idCentral := strings.TrimSpace(r.URL.Query().Get("id_central"))
	ok, saldo, min, err := pgatendimento.PodeLigar(r.Context(), idFra, idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pode_ligar":   ok,
		"saldo":        saldo,
		"custo_minimo": min,
		"canal":        pgatendimento.CanalLigacao,
		"saldo_unico":  true,
	})
}

func (h *Handler) OpsAtendimentoParceiroAtivacaoStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	out, err := pgatendimento.GetStatusAtivacao(r.Context(), idFra)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": out})
}

func (h *Handler) OpsAtendimentoParceiroAtivacaoPropor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var in struct {
		IDFranqueado             string   `json:"id_franqueado"`
		ParceiroMonitoramento    bool     `json:"parceiro_monitoramento"`
		ParceiroModo             string   `json:"parceiro_modo"`
		IDParceiro               string   `json:"id_parceiro"`
		GruposEventoParceiro     []string `json:"grupos_evento_parceiro"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	pol := pgatendimento.Politica{
		IDFranqueado:          strings.TrimSpace(in.IDFranqueado),
		ParceiroMonitoramento: in.ParceiroMonitoramento,
		ParceiroModo:          strings.TrimSpace(in.ParceiroModo),
		IDParceiro:            strings.TrimSpace(in.IDParceiro),
		GruposEventoParceiro:  in.GruposEventoParceiro,
	}
	if pol.IDFranqueado == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	out, err := pgatendimento.ProporAtivacao(r.Context(), pol)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": out})
}

func (h *Handler) OpsAtendimentoParceiroAtivacaoCancelar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	var in struct {
		IDFranqueado string `json:"id_franqueado"`
	}
	if err := decodeJSONPermissive(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	idFra := strings.TrimSpace(in.IDFranqueado)
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	if err := pgatendimento.CancelarAtivacaoPendente(r.Context(), idFra); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) OpsAtendimentoParceiroExcecaoStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	out, err := pgatendimento.GetExcecaoStatus(r.Context(), idFra)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": out})
}

func (h *Handler) OpsAtendimentoParceiroExcecaoVinculo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	idCli := strings.TrimSpace(r.URL.Query().Get("id_cliente"))
	if idFra == "" || idCli == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado e id_cliente obrigatorios")
		return
	}
	out, err := pgatendimento.ConsultaVinculoCliente(r.Context(), idFra, idCli)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": out})
}

func (h *Handler) OpsAtendimentoParceiroExcecaoPropor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	var in struct {
		IDFranqueado string `json:"id_franqueado"`
		IDCliente    string `json:"id_cliente"`
		NomeCliente  string `json:"nome_cliente"`
		IDParceiro   string `json:"id_parceiro"`
	}
	if err := decodeJSONPermissive(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	out, err := pgatendimento.ProporExcecao(r.Context(), pgatendimento.ProporExcecaoInput{
		IDFranqueado: strings.TrimSpace(in.IDFranqueado),
		IDCliente:    strings.TrimSpace(in.IDCliente),
		NomeCliente:  strings.TrimSpace(in.NomeCliente),
		IDParceiro:   strings.TrimSpace(in.IDParceiro),
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": out})
}

func (h *Handler) OpsAtendimentoParceiroHistorico(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	limite := 50
	if v := strings.TrimSpace(r.URL.Query().Get("limite")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limite = n
		}
	}
	list, err := pgatendimento.ListHistoricoParceiro(r.Context(), idFra, limite)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []pgatendimento.HistoricoRegistro{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) OpsAtendimentoParceiroFaturasStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgatendimento.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	list, err := pgatendimento.ListFaturasParceiroStatus(r.Context(), idFra)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []pgatendimento.FaturaParceiroStatus{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) OpsFranqueadoNomes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("ids"))
	if raw == "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": map[string]string{}})
		return
	}
	parts := strings.Split(raw, ",")
	nomes := pgatendimento.LookupFranqueadoNomes(parts)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": nomes})
}

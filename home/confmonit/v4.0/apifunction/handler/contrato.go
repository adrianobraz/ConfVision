package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"apifunction/auth"
	"apifunction/service/contrato"
)

func (h *Handler) ContratoCatalogo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = contrato.EnsureCatalogo(idCentral)

	produtos, err := contrato.ListarCatalogoProdutos(idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	cotas, err := contrato.ListarPacotesCota(idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	cv, err := contrato.ListarCVLicencas(idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"dados": map[string]any{
			"produtos": produtos, "pacotes_cota": cotas, "cv_licencas": cv,
		},
	})
}

func (h *Handler) ContratoCatalogoSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idCentral, err := sess.RequireIDCentralFinanceiro()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	n, err := contrato.SeedCatalogo(idCentral)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": map[string]any{"inseridos": n}})
}

func (h *Handler) ContratoPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if _, err := auth.RequireEscopo(r); err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		Itens         []contrato.ItemInput `json:"itens"`
		DescontoTipo  string               `json:"desconto_tipo"`
		DescontoValor float64              `json:"desconto_valor"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	prev := contrato.CalcularPreview(req.Itens, req.DescontoTipo, req.DescontoValor)
	if err := contrato.ValidarDesconto(prev.ValorBase, req.DescontoTipo, req.DescontoValor); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": prev})
}

func (h *Handler) ContratoSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var in contrato.SalvarInput
	if err := decodeJSONPermissive(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := contrato.Salvar(sess, in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) ContratoObter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idContrato, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id_contrato")))
	if idContrato <= 0 {
		writeErr(w, http.StatusBadRequest, "id_contrato obrigatorio")
		return
	}
	det, err := contrato.Obter(sess, idContrato)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": det})
}

func (h *Handler) ContratoListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	lista, err := contrato.ListarContratos(sess, idFra)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ContratoFaturas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	lista, err := contrato.ListarFaturas(sess, idFra)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ContratoConfirmarPagamento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		IDFatura int `json:"id_fatura"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDFatura <= 0 {
		writeErr(w, http.StatusBadRequest, "id_fatura obrigatorio")
		return
	}
	if err := contrato.ConfirmarPagamento(sess, req.IDFatura); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ContratoEstornarPagamento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		IDFatura int    `json:"id_fatura"`
		Motivo   string `json:"motivo"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDFatura <= 0 {
		writeErr(w, http.StatusBadRequest, "id_fatura obrigatorio")
		return
	}
	res, err := contrato.EstornarPagamento(sess, req.IDFatura, req.Motivo)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) ContratoEstornarPagamentoContabil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		IDFaturaContabil    int    `json:"id_fatura_contabil"`
		Motivo              string `json:"motivo"`
		ContabilJaEstornado bool   `json:"contabil_ja_estornado"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDFaturaContabil <= 0 {
		writeErr(w, http.StatusBadRequest, "id_fatura_contabil obrigatorio")
		return
	}
	res, err := contrato.EstornarPagamentoContabil(sess, req.IDFaturaContabil, req.Motivo, req.ContabilJaEstornado)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) ContratoConfirmarPagamentoContabil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		IDFaturaContabil int `json:"id_fatura_contabil"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDFaturaContabil <= 0 {
		writeErr(w, http.StatusBadRequest, "id_fatura_contabil obrigatorio")
		return
	}
	if err := contrato.ConfirmarPagamentoContabil(sess, req.IDFaturaContabil); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ContratoExcluirFatura(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		IDFatura int `json:"id_fatura"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDFatura <= 0 {
		writeErr(w, http.StatusBadRequest, "id_fatura obrigatorio")
		return
	}
	if err := contrato.ExcluirFatura(sess, req.IDFatura); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ContratoExcluir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		IDContrato int `json:"id_contrato"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.IDContrato <= 0 {
		writeErr(w, http.StatusBadRequest, "id_contrato obrigatorio")
		return
	}
	if err := contrato.ExcluirContrato(sess, req.IDContrato); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ContratoEfetivo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	produto := strings.TrimSpace(r.URL.Query().Get("produto"))
	ef := contrato.Efetivo(idFra, produto)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": ef})
}

func (h *Handler) ContratoWorkerBilling(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey string `json:"worker_key"`
	}
	_ = decodeJSONPermissive(r, &req)
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	dias := 5
	if q := r.URL.Query().Get("dias_antecedencia"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			dias = n
		}
	}
	susp, _ := contrato.WorkerSuspenderVencidas()
	ren, _ := contrato.WorkerGerarRenovacoes(dias)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"dados": map[string]any{
			"faturas_vencidas": susp, "renovacoes_geradas": ren,
		},
	})
}

func checkWorkerKey(r *http.Request, bodyKey string) bool {
	secret := strings.TrimSpace(os.Getenv("WORKER_SECRET"))
	if secret == "" {
		return true
	}
	key := strings.TrimSpace(bodyKey)
	if key == "" {
		key = strings.TrimSpace(r.Header.Get("X-Worker-Key"))
	}
	return key == secret
}

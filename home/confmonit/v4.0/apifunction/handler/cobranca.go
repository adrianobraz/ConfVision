package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"apifunction/pgcobranca"
	"apifunction/pgcredito"
	"apifunction/confservice"
	"apifunction/pgatendimento"
)

func (h *Handler) CobrancaConfigGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	cfg, err := pgcobranca.GetConfig(r.Context(), idFra)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": cfg})
}

func (h *Handler) CobrancaConfigSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	var cfg pgcobranca.Config
	if err := decodeJSONPermissive(r, &cfg); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	cfg.IDFranqueado = strings.TrimSpace(cfg.IDFranqueado)
	if cfg.IDFranqueado == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	if err := pgcobranca.SaveConfig(r.Context(), cfg); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) CobrancaServicosListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	list, err := pgcobranca.ListServicos(r.Context(), idFra)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) CobrancaServicoSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	var req struct {
		IDFranqueado    string  `json:"id_franqueado"`
		RefTipo         string  `json:"ref_tipo"`
		RefID           string  `json:"ref_id"`
		Descricao       string  `json:"descricao"`
		DataInicio      string  `json:"data_inicio"`
		Periodicidade   string  `json:"periodicidade"`
		ValorCiclo      float64 `json:"valor_ciclo"`
		DiasCiclo       int     `json:"dias_ciclo"`
		PagoAte         string  `json:"pago_ate"`
		SaldoAjuste     float64 `json:"saldo_ajuste"`
		StatusCiclo     string  `json:"status_ciclo"`
		ValorPiso       float64 `json:"valor_piso"`
		MargemCentral   float64 `json:"margem_central"`
		MargemRep       float64 `json:"margem_rep"`
		Ativo           bool    `json:"ativo"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido: "+err.Error())
		return
	}
	s := pgcobranca.Servico{
		IDFranqueado:  strings.TrimSpace(req.IDFranqueado),
		RefTipo:       strings.TrimSpace(req.RefTipo),
		RefID:         strings.TrimSpace(req.RefID),
		Descricao:     strings.TrimSpace(req.Descricao),
		Periodicidade: strings.TrimSpace(req.Periodicidade),
		ValorCiclo:    req.ValorCiclo,
		DiasCiclo:     req.DiasCiclo,
		SaldoAjuste:   req.SaldoAjuste,
		StatusCiclo:   strings.TrimSpace(req.StatusCiclo),
		ValorPiso:     req.ValorPiso,
		MargemCentral: req.MargemCentral,
		MargemRep:     req.MargemRep,
		Ativo:         req.Ativo,
	}
	if strings.TrimSpace(req.DataInicio) != "" {
		s.DataInicio = parseDataAncora(req.DataInicio)
	} else {
		s.DataInicio = time.Now()
	}
	if pago := strings.TrimSpace(req.PagoAte); pago != "" {
		t := parseDataAncora(pago)
		s.PagoAte = &t
	}
	if s.IDFranqueado == "" || s.RefTipo == "" || s.RefID == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado, ref_tipo e ref_id obrigatorios")
		return
	}
	if s.Periodicidade == "" {
		s.Periodicidade = pgcobranca.PeriodicidadeQuinzenal
	}
	if s.DiasCiclo <= 0 {
		s.DiasCiclo = 15
	}
	if s.StatusCiclo == "" {
		s.StatusCiclo = pgcobranca.StatusCicloEmAjuste
	}
	id, err := pgcobranca.UpsertServico(r.Context(), s)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
}

func (h *Handler) CobrancaServicoSyncLote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey string `json:"worker_key"`
		Servicos  []struct {
			IDFranqueado  string  `json:"id_franqueado"`
			RefTipo       string  `json:"ref_tipo"`
			RefID         string  `json:"ref_id"`
			Descricao     string  `json:"descricao"`
			DataInicio    string  `json:"data_inicio"`
			Periodicidade string  `json:"periodicidade"`
			ValorCiclo    float64 `json:"valor_ciclo"`
			DiasCiclo     int     `json:"dias_ciclo"`
			PagoAte       string  `json:"pago_ate"`
			ValorPiso     float64 `json:"valor_piso"`
			MargemCentral float64 `json:"margem_central"`
			MargemRep     float64 `json:"margem_rep"`
			Ativo         bool    `json:"ativo"`
		} `json:"servicos"`
	}
	if err := decodeJSONPermissive(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	synced := 0
	var errs []string
	for _, row := range req.Servicos {
		s := pgcobranca.Servico{
			IDFranqueado:  strings.TrimSpace(row.IDFranqueado),
			RefTipo:       strings.TrimSpace(row.RefTipo),
			RefID:         strings.TrimSpace(row.RefID),
			Descricao:     strings.TrimSpace(row.Descricao),
			Periodicidade: strings.TrimSpace(row.Periodicidade),
			ValorCiclo:    row.ValorCiclo,
			DiasCiclo:     row.DiasCiclo,
			ValorPiso:     row.ValorPiso,
			MargemCentral: row.MargemCentral,
			MargemRep:     row.MargemRep,
			Ativo:         row.Ativo,
		}
		if strings.TrimSpace(row.DataInicio) != "" {
			s.DataInicio = parseDataAncora(row.DataInicio)
		} else {
			s.DataInicio = time.Now()
		}
		if pago := strings.TrimSpace(row.PagoAte); pago != "" {
			t := parseDataAncora(pago)
			s.PagoAte = &t
		}
		if s.IDFranqueado == "" || s.RefTipo == "" || s.RefID == "" {
			continue
		}
		if s.Periodicidade == "" {
			s.Periodicidade = pgcobranca.PeriodicidadeMensal
		}
		if s.DiasCiclo <= 0 {
			if s.Periodicidade == pgcobranca.PeriodicidadeQuinzenal {
				s.DiasCiclo = 15
			} else {
				s.DiasCiclo = 30
			}
		}
		if s.StatusCiclo == "" {
			s.StatusCiclo = pgcobranca.StatusCicloNormal
		}
		if _, err := pgcobranca.UpsertServico(r.Context(), s); err != nil {
			errs = append(errs, err.Error())
			continue
		}
		synced++
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "synced": synced, "total": len(req.Servicos), "erros": errs,
	})
}

func (h *Handler) CobrancaSimular(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	var req struct {
		IDFranqueado string `json:"id_franqueado"`
		DataAncora   string `json:"data_ancora"`
	}
	_ = decodeJSONPermissive(r, &req)
	idFra := strings.TrimSpace(req.IDFranqueado)
	if idFra == "" {
		idFra = strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	}
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	anchor := parseDataAncora(req.DataAncora)
	eng := pgcobranca.NovoEngine()
	ord, err := eng.Simular(r.Context(), idFra, anchor)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": ord})
}

func (h *Handler) CobrancaWorkerGerarOrdens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey      string `json:"worker_key"`
		IDFranqueado   string `json:"id_franqueado"`
		DataAncora     string `json:"data_ancora"`
	}
	_ = decodeJSONPermissive(r, &req)
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	anchor := parseDataAncora(req.DataAncora)
	eng := pgcobranca.NovoEngine()
	res, err := eng.GerarOrdens(context.Background(), anchor, strings.TrimSpace(req.IDFranqueado))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func parseDataAncora(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now()
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
		"02/01/2006",
	} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t
		}
	}
	return time.Now()
}

func (h *Handler) CobrancaPosPagamento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey    string `json:"worker_key"`
		IDFranqueado string `json:"id_franqueado"`
		FaturaID     int    `json:"fatura_id"`
		PagoEm       string `json:"pago_em"`
		Itens        []struct {
			RefTipo   string  `json:"ref_tipo"`
			RefID     string  `json:"ref_id"`
			ValorPiso float64 `json:"valor_piso"`
			PagoAte   string  `json:"pago_ate"`
			IDParceiro string `json:"id_parceiro"`
		} `json:"itens"`
	}
	_ = decodeJSONPermissive(r, &req)
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	pagoEm := parseDataAncora(req.PagoEm)
	type posItemIn struct {
		RefTipo    string
		RefID      string
		ValorPiso  float64
		PagoAte    string
		IDParceiro string
	}
	var entrada []posItemIn
	for _, it := range req.Itens {
		refTipo := strings.TrimSpace(it.RefTipo)
		refID := strings.TrimSpace(it.RefID)
		if refTipo == "" || refID == "" {
			continue
		}
		entrada = append(entrada, posItemIn{
			RefTipo: refTipo, RefID: refID, ValorPiso: it.ValorPiso,
			PagoAte: it.PagoAte, IDParceiro: strings.TrimSpace(it.IDParceiro),
		})
	}
	if len(entrada) == 0 && req.FaturaID > 0 {
		if fb, err := pgatendimento.FallbackItensPosPagamento(r.Context(), strings.TrimSpace(req.IDFranqueado), req.FaturaID); err == nil {
			for _, f := range fb {
				entrada = append(entrada, posItemIn{
					RefTipo: f.RefTipo, RefID: f.RefID, ValorPiso: f.ValorPiso, IDParceiro: f.IDParceiro,
				})
			}
		}
	}
	var itens []pgcobranca.PosPagamentoItem
	var repasseItens []confservice.RepasseItem
	temExcecao := false
	vinculosExcecaoAnt := map[string]struct{}{}
	for _, it := range entrada {
		if it.RefTipo == pgatendimento.RefTipoParceiroExcecao {
			temExcecao = true
			if v := pgatendimento.VinculoAnteriorExcecao(r.Context(), it.RefID); v != "" {
				vinculosExcecaoAnt[v] = struct{}{}
			}
		}
	}
	idFranq := strings.TrimSpace(req.IDFranqueado)
	for _, it := range entrada {
		item := pgcobranca.PosPagamentoItem{
			RefTipo:   it.RefTipo,
			RefID:     it.RefID,
			ValorPiso: it.ValorPiso,
		}
		if it.PagoAte != "" {
			item.PagoAte = parseDataAncora(it.PagoAte)
		}
		if it.RefTipo == pgatendimento.RefTipoParceiroConfig {
			ate := item.PagoAte
			if ate.IsZero() {
				ate = pagoEm.AddDate(0, 0, 14)
			}
			item.PagoAte = ate
			_, _ = pgatendimento.MarcarAtivacaoPagaFatura(r.Context(), idFranq, req.FaturaID, item.RefID, ate)
			if idP, piso, err := pgatendimento.MetaRepasseAtivacao(r.Context(), idFranq, req.FaturaID, item.RefID, it.ValorPiso); err == nil && piso > 0 {
				repasseItens = append(repasseItens, confservice.RepasseItem{
					IDParceiro: idP,
					ValorPiso:  piso,
					FaturaID:   req.FaturaID,
					RefTipo:    pgatendimento.RefTipoParceiroConfig,
					RefID:      idP,
				})
			}
		}
		if it.RefTipo == pgatendimento.RefTipoParceiroExcecao {
			ate := item.PagoAte
			if ate.IsZero() {
				ate = pagoEm.AddDate(0, 0, 14)
			}
			item.PagoAte = ate
			if err := pgatendimento.AtivarExcecaoPaga(r.Context(), idFranq, item.RefID, req.FaturaID, ate); err == nil {
				if idV, idP, piso, err := pgatendimento.MetaRepasseExcecao(r.Context(), item.RefID, it.ValorPiso); err == nil && piso > 0 {
					repasseItens = append(repasseItens, confservice.RepasseItem{
						IDVinculo:  idV,
						IDParceiro: idP,
						ValorPiso:  piso,
						FaturaID:   req.FaturaID,
						RefTipo:    "cs_parceiro_vinculo",
						RefID:      idV,
					})
				}
			}
		}
		itens = append(itens, item)
		if it.RefTipo == "cs_parceiro_vinculo" && it.ValorPiso > 0 {
			if temExcecao {
				if _, skip := vinculosExcecaoAnt[it.RefID]; skip {
					continue
				}
				continue
			}
			idParceiro := it.IDParceiro
			if idParceiro == "" {
				continue
			}
			repasseItens = append(repasseItens, confservice.RepasseItem{
				IDVinculo:  it.RefID,
				IDParceiro: idParceiro,
				ValorPiso:  it.ValorPiso,
				FaturaID:   req.FaturaID,
				RefTipo:    it.RefTipo,
				RefID:      it.RefID,
			})
		}
	}
	res, err := pgcobranca.PosPagamento(r.Context(), pgcobranca.PosPagamentoInput{
		IDFranqueado: strings.TrimSpace(req.IDFranqueado),
		FaturaID:     req.FaturaID,
		PagoEm:       pagoEm,
		Itens:        itens,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(repasseItens) > 0 {
		_ = confservice.GerarRepassesParceiro(req.IDFranqueado, req.FaturaID, repasseItens)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"dados":          res,
		"repasse_itens":  len(repasseItens),
		"itens_processados": len(entrada),
	})
}

func (h *Handler) CobrancaPosEstorno(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req struct {
		WorkerKey    string `json:"worker_key"`
		IDFranqueado string `json:"id_franqueado"`
		FaturaID     int    `json:"fatura_id"`
	}
	_ = decodeJSONPermissive(r, &req)
	if !checkWorkerKey(r, req.WorkerKey) {
		writeErr(w, http.StatusUnauthorized, "worker_key invalido")
		return
	}
	if !pgcredito.Configurado() {
		writeErr(w, http.StatusServiceUnavailable, "POSTGRES_URL nao configurado")
		return
	}
	if req.FaturaID <= 0 {
		writeErr(w, http.StatusBadRequest, "fatura_id obrigatorio")
		return
	}
	res, err := pgatendimento.ReverterPosEstorno(r.Context(), strings.TrimSpace(req.IDFranqueado), req.FaturaID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

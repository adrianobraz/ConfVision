package handler

import (
	"net/http"
	"strconv"
	"strings"

	"apifunction/auth"
	"apifunction/service"
)

type Handler struct{}

func Novo() *Handler {
	return &Handler{}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/transferencia/preview", h.TransferenciaPreview)
	mux.HandleFunc("/transferencia/executar", h.TransferenciaExecutar)
	mux.HandleFunc("/transferencia/historico", h.TransferenciaHistorico)
	// Contrato (admConfmonit / MySQL)
	mux.HandleFunc("/contrato/catalogo", h.ContratoCatalogo)
	mux.HandleFunc("/contrato/catalogo/seed", h.ContratoCatalogoSeed)
	mux.HandleFunc("/contrato/preview", h.ContratoPreview)
	mux.HandleFunc("/contrato/salvar", h.ContratoSalvar)
	mux.HandleFunc("/contrato/obter", h.ContratoObter)
	mux.HandleFunc("/contrato/listar", h.ContratoListar)
	mux.HandleFunc("/contrato/faturas", h.ContratoFaturas)
	mux.HandleFunc("/contrato/fatura/confirmar-pagamento", h.ContratoConfirmarPagamento)
	mux.HandleFunc("/contrato/fatura/confirmar-pagamento-contabil", h.ContratoConfirmarPagamentoContabil)
	mux.HandleFunc("/contrato/fatura/estornar-pagamento", h.ContratoEstornarPagamento)
	mux.HandleFunc("/contrato/fatura/estornar-pagamento-contabil", h.ContratoEstornarPagamentoContabil)
	mux.HandleFunc("/contrato/fatura/excluir", h.ContratoExcluirFatura)
	mux.HandleFunc("/contrato/excluir", h.ContratoExcluir)
	mux.HandleFunc("/contrato/efetivo", h.ContratoEfetivo)
	mux.HandleFunc("/contrato/worker/billing", h.ContratoWorkerBilling)
	// Servicos operacionais / creditos prepago
	mux.HandleFunc("/servico/tarifa/listar", h.ServicoTarifaListar)
	mux.HandleFunc("/servico/tarifa/salvar", h.ServicoTarifaSalvar)
	mux.HandleFunc("/servico/tarifa/rep/listar", h.ServicoTarifaRepListar)
	mux.HandleFunc("/servico/tarifa/rep/salvar", h.ServicoTarifaRepSalvar)
	mux.HandleFunc("/servico/tarifa/seed", h.ServicoTarifaSeed)
	mux.HandleFunc("/servico/tarifa/efetiva", h.ServicoTarifaEfetiva)
	mux.HandleFunc("/credito/recarga/solicitar", h.CreditoRecargaSolicitar)
	mux.HandleFunc("/credito/recarga/confirmar", h.CreditoRecargaConfirmar)
	mux.HandleFunc("/credito/recarga/listar", h.CreditoRecargaListar)
	mux.HandleFunc("/credito/debitar-uso", h.CreditoDebitarUso)
	// Atendimento operacional (runtime — migrado do ConfVision :8086)
	mux.HandleFunc("/ops/atendimento/resolve", h.OpsAtendimentoResolve)
	mux.HandleFunc("/ops/atendimento/parceiro/vinculo", h.OpsAtendimentoParceiroVinculo)
	mux.HandleFunc("/ops/atendimento/parceiro/dispatch", h.OpsAtendimentoParceiroDispatch)
	mux.HandleFunc("/ops/atendimento/parceiro/log", h.OpsAtendimentoParceiroLog)
	mux.HandleFunc("/ops/atendimento/parceiro/envio", h.OpsAtendimentoParceiroEnvio)
	mux.HandleFunc("/ops/atendimento/parceiro/eventos", h.OpsAtendimentoParceiroEventosList)
	mux.HandleFunc("/ops/atendimento/parceiro/ativacao/status", h.OpsAtendimentoParceiroAtivacaoStatus)
	mux.HandleFunc("/ops/atendimento/parceiro/ativacao/propor", h.OpsAtendimentoParceiroAtivacaoPropor)
	mux.HandleFunc("/ops/atendimento/parceiro/ativacao/cancelar", h.OpsAtendimentoParceiroAtivacaoCancelar)
	mux.HandleFunc("/ops/atendimento/parceiro/excecao/status", h.OpsAtendimentoParceiroExcecaoStatus)
	mux.HandleFunc("/ops/atendimento/parceiro/excecao/vinculo", h.OpsAtendimentoParceiroExcecaoVinculo)
	mux.HandleFunc("/ops/atendimento/parceiro/excecao/propor", h.OpsAtendimentoParceiroExcecaoPropor)
	mux.HandleFunc("/ops/atendimento/parceiro/historico", h.OpsAtendimentoParceiroHistorico)
	mux.HandleFunc("/ops/atendimento/parceiro/faturas/status", h.OpsAtendimentoParceiroFaturasStatus)
	mux.HandleFunc("/ops/atendimento/franqueado/nomes", h.OpsFranqueadoNomes)
	mux.HandleFunc("/ops/atendimento/credito/pode_ligar", h.OpsAtendimentoCreditoPodeLigar)
	// Governanca financeira cascata (Postgres + worker)
	mux.HandleFunc("/financeiro/governanca/estado", h.GovernancaEstado)
	mux.HandleFunc("/financeiro/governanca/franqueado", h.GovernancaFranqueado)
	mux.HandleFunc("/financeiro/governanca/worker/tick", h.GovernancaWorkerTick)

	mux.HandleFunc("/financeiro/cobranca/config", h.CobrancaConfigGet)
	mux.HandleFunc("/financeiro/cobranca/config/salvar", h.CobrancaConfigSave)
	mux.HandleFunc("/financeiro/cobranca/servicos", h.CobrancaServicosListar)
	mux.HandleFunc("/financeiro/cobranca/servico/salvar", h.CobrancaServicoSalvar)
	mux.HandleFunc("/financeiro/cobranca/servico/sync-lote", h.CobrancaServicoSyncLote)
	mux.HandleFunc("/financeiro/cobranca/simular", h.CobrancaSimular)
	mux.HandleFunc("/financeiro/cobranca/worker/gerar-ordens", h.CobrancaWorkerGerarOrdens)
	mux.HandleFunc("/financeiro/cobranca/pos-pagamento", h.CobrancaPosPagamento)
	mux.HandleFunc("/financeiro/cobranca/pos-estorno", h.CobrancaPosEstorno)
	mux.HandleFunc("/financeiro/fatura/listar", h.FinanceiroFaturaListar)
	mux.HandleFunc("/financeiro/fin-mirror/worker/sync", h.FinanceiroFinMirrorSync)
	mux.HandleFunc("/financeiro/fin-mirror/hook", h.FinanceiroFinMirrorHook)
	mux.HandleFunc("/financeiro/dashboard/kpi", h.FinanceiroDashboardKPI)
	mux.HandleFunc("/financeiro/dashboard/kpi-shadow", h.FinanceiroDashboardKPIShadow)
	mux.HandleFunc("/financeiro/catalogo/listar", h.FinanceiroCatalogoListar)
	mux.HandleFunc("/financeiro/catalogo/salvar", h.FinanceiroCatalogoSalvar)
	mux.HandleFunc("/financeiro/catalogo/seed", h.FinanceiroCatalogoSeed)
	mux.HandleFunc("/financeiro/catalogo/sync-xano", h.FinanceiroCatalogoSyncXano)
	mux.HandleFunc("/financeiro/pacote-cota/listar", h.FinanceiroPacoteCotaListar)
	mux.HandleFunc("/financeiro/pacote-cota/salvar", h.FinanceiroPacoteCotaSalvar)
	mux.HandleFunc("/financeiro/pacote-cota/preco-rep/salvar", h.FinanceiroPacoteCotaPrecoRepSalvar)
	mux.HandleFunc("/financeiro/pacote-cota/seed", h.FinanceiroPacoteCotaSeed)
	mux.HandleFunc("/financeiro/central-preco-config/listar", h.FinanceiroCentralPrecoConfigListar)
	mux.HandleFunc("/financeiro/central-preco-config/salvar", h.FinanceiroCentralPrecoConfigSalvar)
	mux.HandleFunc("/financeiro/central-dominio/carregar", h.FinanceiroCentralDominioCarregar)
	mux.HandleFunc("/financeiro/central-dominio/salvar", h.FinanceiroCentralDominioSalvar)
	mux.HandleFunc("/financeiro/central-dominio/remover", h.FinanceiroCentralDominioRemover)
	mux.HandleFunc("/financeiro/central-dominio/retentar-ssl", h.FinanceiroCentralDominioRetentarSSL)
	mux.HandleFunc("/financeiro/central-dominio/urls", h.FinanceiroCentralDominioURLs)
	mux.HandleFunc("/financeiro/central-whitelabel/carregar", h.FinanceiroCentralWhitelabelCarregar)
	mux.HandleFunc("/financeiro/central-whitelabel/salvar", h.FinanceiroCentralWhitelabelSalvar)
	mux.HandleFunc("/marca/resolve", h.MarcaResolve)
	// DNS receptor alarmes (FranqueadoPro)
	mux.HandleFunc("/receptor/dns/obter", h.ReceptorDNSObter)
	mux.HandleFunc("/receptor/dns/disponibilidade", h.ReceptorDNSDisponibilidade)
	mux.HandleFunc("/receptor/dns/salvar", h.ReceptorDNSSalvar)
	mux.HandleFunc("/receptor/dns/remover", h.ReceptorDNSRemover)
	mux.HandleFunc("/receptor/dns/endpoints", h.ReceptorDNSEndpoints)
	// Diagnostico receptor (lab + admin)
	mux.HandleFunc("/receptor/diag/fabricantes", h.ReceptorDiagFabricantes)
	mux.HandleFunc("/receptor/diag/franqueados", h.ReceptorDiagFranqueados)
	mux.HandleFunc("/receptor/diag/verificar", h.ReceptorDiagVerificar)
	mux.HandleFunc("/receptor/diag/erro-conexao", h.ReceptorDiagErroConexao)
	mux.HandleFunc("/receptor/diag/simular-cadastro", h.ReceptorDiagSimularCadastro)
	mux.HandleFunc("/receptor/diag/presenca", h.ReceptorDiagPresenca)
	return withCORS(mux)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"servico": "apifunction",
		"versao":  "2026.08.27-fatura-listar-mirror",
		"auth": map[string]any{
			"jwt_bearer":            true,
			"adm_session_headers":   true,
			"breakglass":            true,
		},
	})
}

func (h *Handler) TransferenciaPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if _, err := auth.RequireAdm(r); err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req service.ReqTransferencia
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	prev, err := service.Preview(req)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": prev.OK, "dados": prev})
}

func (h *Handler) TransferenciaExecutar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	usuario, err := auth.RequireAdm(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req service.ReqTransferencia
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Motivo) == "" {
		writeErr(w, http.StatusBadRequest, "motivo obrigatorio")
		return
	}
	res, err := service.Executar(req, usuario.IDUsuario)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) TransferenciaHistorico(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if _, err := auth.RequireAdm(r); err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	limite := 50
	if q := r.URL.Query().Get("limite"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limite = n
		}
	}
	itens, err := service.ListarHistorico(limite)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": itens})
}

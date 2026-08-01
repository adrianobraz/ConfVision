package handler

import "net/http"

type Handler struct{}

func Novo() *Handler {
	return &Handler{}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.Health)

	// Portal parceiro (monitoramento)
	mux.HandleFunc("/parceiro/registrar", h.ParceiroRegistrar)
	mux.HandleFunc("/parceiro/login", h.ParceiroLogin)
	mux.HandleFunc("/parceiro/me", h.ParceiroMe)
	mux.HandleFunc("/parceiro/me/atualizar", h.ParceiroAtualizar)
	mux.HandleFunc("/parceiro/me/clientes", h.MeusClientes)
	mux.HandleFunc("/parceiro/me/faturas", h.FaturasParceiroMe)
	mux.HandleFunc("/parceiro/me/fatura/fechar", h.FaturaParceiroFechar)
	mux.HandleFunc("/parceiro/me/eventos", h.EventosParceiroMe)

	// Admin (portal /admin — usuario/senha do .env, default admin/admin)
	mux.HandleFunc("/admin/login", h.AdminLogin)
	mux.HandleFunc("/admin/parceiros", h.AdminParceiros)
	mux.HandleFunc("/admin/parceiros/{id}/reset-senha", h.AdminResetSenha)

	// Marketplace (vitrine pública)
	mux.HandleFunc("/marketplace/categorias", h.MarketplaceCategorias)
	mux.HandleFunc("/marketplace/fabricantes", h.MarketplaceFabricantes)
	mux.HandleFunc("/marketplace/fabricante/registrar", h.MarketplaceFabricanteRegistrar)
	mux.HandleFunc("/marketplace/prestadores", h.MarketplacePrestadores)
	mux.HandleFunc("/marketplace/prestadores/{id}", h.MarketplacePrestadorDetalhe)
	mux.HandleFunc("/marketplace/prestador/registrar", h.MarketplacePrestadorRegistrar)
	mux.HandleFunc("/marketplace/orcamento", h.MarketplaceOrcamento)

	// Catalogo para Franqueado Pro (legado — retorna vazio; use /internal/parceiros/catalogo)
	mux.HandleFunc("/parceiros", h.ParceirosCatalogo)

	// Interno (Xano / Franqueado Pro / admConfmonit via API V4) — X-Api-Key
	mux.HandleFunc("/internal/parceiros/global", h.ParceirosGlobal)
	mux.HandleFunc("/internal/parceiros/rep", h.ParceirosRepListar)
	mux.HandleFunc("/internal/parceiros/rep/salvar", h.ParceirosRepSalvar)
	mux.HandleFunc("/internal/parceiros/franqueado", h.ParceirosFranqueadoListar)
	mux.HandleFunc("/internal/parceiros/franqueado/salvar", h.ParceirosFranqueadoSalvar)
	mux.HandleFunc("/internal/parceiros/catalogo", h.ParceirosCatalogoFranqueado)
	mux.HandleFunc("/internal/vinculo", h.VinculoCriar)
	mux.HandleFunc("/internal/vinculo/desativar", h.VinculoDesativar)
	mux.HandleFunc("/internal/vinculo/cliente", h.VinculoPorCliente)
	mux.HandleFunc("/internal/vinculos", h.VinculosPorFranqueado)
	mux.HandleFunc("/internal/evento", h.EventoEnfileirar)
	mux.HandleFunc("/internal/webhook/processar", h.WebhookProcessar)
	mux.HandleFunc("/internal/fatura/fechar", h.FaturaFechar)
	mux.HandleFunc("/internal/faturas/franqueado", h.FaturasFranqueado)
	mux.HandleFunc("/internal/eventos", h.EventosInternal)

	return withCORS(mux)
}

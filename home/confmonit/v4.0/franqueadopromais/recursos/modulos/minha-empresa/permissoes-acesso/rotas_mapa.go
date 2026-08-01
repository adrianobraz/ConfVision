package permissoesAcesso

import "strings"

// MapaRotaChave associa URI da pagina a chave de permissao
var MapaRotaChave = map[string]string{
	"/carregar-menu-configuracoes":              "configuracao",
	"/carregar-gerenciar-dispositivo":           "configuracao.dispositivo",
	"/carregar-gerenciar-usuarios-alarme":         "configuracao.usuarios-alarme",
	"/carregar-gerenciar-setores-alarme":          "configuracao.setores-alarme",
	"/carregar-gerenciar-procedimento-atendimento": "configuracao.procedimento",
	"/carregar-gerenciar-contactid-personalizado": "configuracao.contactid",
	"/carregar-gerenciar-grade":                   "configuracao.grade",
	"/gerenciar-configuracao-email-eveto":         "configuracao.email-evento",

	"/carregar-menu-atendimento":    "atendimento",
	"/CarregarGeradorEventos":       "atendimento.gerar-evento",
	"/carregar-inteligencia-artificial": "atendimento.inteligencia-artificial",

	"/carregar-menu-minha-empresa":  "minha-empresa",
	"/carregar-menu-operacional":    "minha-empresa.operacional",
	"/CarregarPaginaGerenciarTecnico": "minha-empresa.operacional.tecnico",
	"/CarregarPaginaGerenciarViatura": "minha-empresa.operacional.viatura",
	"/carregar-menu-comercial":      "minha-empresa.comercial",
	"/CarregarPaginaGerenciarCliente": "minha-empresa.comercial.cliente",
	"/carregar-parceiro-monitoramento": "minha-empresa.comercial.parceiro-monitoramento",
	"/carregar-menu-gestao":         "minha-empresa.gestao",
	"/CarregarPaginaDadosMonitoramento": "minha-empresa.gestao.dados",
	"/carregarPaginaGerenciarUsuarios": "minha-empresa.gestao.responsaveis",
	"/carregar-permissoes-acesso":   "minha-empresa.gestao.permissoes",
	"/carregar-meu-plano":           "minha-empresa.gestao.meu-plano",
	"/carregar-montar-plano":        "minha-empresa.gestao.montar-plano",

	"/carregar-whitelabel":          "whitelabel",
	"/carregar-dominio-saas":        "franqueadopro.saas",

	"/carregar-menu-relatorio":              "relatorio",
	"/carregar-relatorio-atendimento":       "relatorio.atendimento",
	"/carregar-relatorio-ligacoes":          "relatorio.ligacoes",
	"/carregar-menu-eventos":                "relatorio.eventos",
	"/carregar-relatorio-eventos":           "relatorio.eventos.lista",
	"/carregar-configurar-relatorio-cliente": "relatorio.eventos.config-grupo",
	"/carregar-menu-alarmes":                "relatorio.alarmes",
	"/carregar-listar-dispositivos-armados": "relatorio.alarmes.armados",
	"/carregar-listar-dispositivos-desarmados": "relatorio.alarmes.desarmados",
	"/carregar-menu-clientes":               "relatorio.clientes",
	"/listar-cliente-dispositivo":           "relatorio.clientes.disp",
	"/carregar-listar-clientes":             "relatorio.clientes.lista",
	"/carregar-listar-clientes-ativos":      "relatorio.clientes.ativos",
	"/carregar-listar-clientes-inativos":    "relatorio.clientes.inativos",

	"/carregar-menu-central-disparos":           "relatorio.central-disparos",
	"/carregar-relatorio-cd-unificado":          "relatorio.central-disparos.unificado",
	"/cdUnificadoLista":                         "relatorio.central-disparos.unificado",
	"/cdUnificadoDetalhe":                       "relatorio.central-disparos.unificado",
	"/carregar-relatorio-eventos-pendentes":     "relatorio.central-disparos.eventos-pendentes",
	"/carregar-relatorio-finalizados-robo":      "relatorio.central-disparos.finalizados-robo",
	"/carregar-relatorio-finalizados-bot":       "relatorio.central-disparos.finalizados-bot",
	"/carregar-relatorio-ligacao-historico":     "relatorio.central-disparos.ligacao-historico",
	"/ligacaoHistoricoListar":                   "relatorio.central-disparos.ligacao-historico",
	"/cdLigacaoHistoricoListar":                 "relatorio.central-disparos.ligacao-historico",
	"/ligacaoHistoricoAudio":                    "relatorio.central-disparos.ligacao-historico",
	"/carregar-relatorio-sms-historico":         "relatorio.central-disparos.sms-historico",
	"/carregar-relatorio-eventos-falhas":        "relatorio.central-disparos.eventos-falhas",
	"/carregar-relatorio-whatsapp-enviados":     "relatorio.central-disparos.whatsapp-enviados",
	"/cdWhatsappEnviadosListar":                 "relatorio.central-disparos.whatsapp-enviados",
	"/carregar-relatorio-ligacoes-cd":           "relatorio.central-disparos.ligacoes",
	"/cdLigacoesListar":                         "relatorio.central-disparos.ligacoes",
	"/carregar-relatorio-sms":                   "relatorio.central-disparos.sms",
	"/cdSmsListar":                              "relatorio.central-disparos.sms",
	"/carregar-relatorio-fila-envio":            "relatorio.central-disparos.fila-envio",
	"/carregar-relatorio-ligacao-erros":         "relatorio.central-disparos.ligacao-erros",
	"/carregar-relatorio-fila-ligacao":          "relatorio.central-disparos.fila-ligacao",

	"/carregar-dashboard-sem-comunicacao": "dashboard.sem-comunicacao",
	"/carregar-dashboard-eventos":         "dashboard.eventos",
}

// ChavePorRota retorna chave de permissao para a URI (case-insensitive)
func ChavePorRota(uri string) (string, bool) {
	return ChavePorRotaComTab(uri, "")
}

// ChavePorRotaComTab considera query tab=faturas em /carregar-meu-plano
func ChavePorRotaComTab(uri, tab string) (string, bool) {
	uri = strings.TrimSuffix(strings.TrimSpace(uri), "/")
	if uri == "" {
		uri = "/"
	}
	lower := strings.ToLower(uri)
	if lower == "/carregar-meu-plano" {
		if strings.EqualFold(strings.TrimSpace(tab), "faturas") {
			return "minha-empresa.gestao.faturas", true
		}
		return "minha-empresa.gestao.meu-plano", true
	}
	if chave, ok := MapaRotaChave[uri]; ok {
		return chave, true
	}
	for k, v := range MapaRotaChave {
		if strings.ToLower(k) == lower {
			return v, true
		}
	}
	return "", false
}

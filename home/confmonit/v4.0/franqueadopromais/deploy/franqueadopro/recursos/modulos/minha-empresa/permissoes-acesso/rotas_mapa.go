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
	"/carregar-menu-gestao":         "minha-empresa.gestao",
	"/CarregarPaginaDadosMonitoramento": "minha-empresa.gestao.dados",
	"/carregarPaginaGerenciarUsuarios": "minha-empresa.gestao.responsaveis",
	"/CarregarPaginaGerenciarOperador": "minha-empresa.gestao.operador",
	"/carregar-permissoes-acesso":   "minha-empresa.gestao.permissoes",
	// meu-plano: rota livre — nao mapear aqui (evita loop de redirect sem licenca)

	"/carregar-whitelabel":          "whitelabel",

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

	"/carregar-dashboard-sem-comunicacao": "dashboard.sem-comunicacao",
	"/carregar-dashboard-eventos":         "dashboard.eventos",
}

// ChavePorRota retorna chave de permissao para a URI (case-insensitive)
func ChavePorRota(uri string) (string, bool) {
	if chave, ok := MapaRotaChave[uri]; ok {
		return chave, true
	}
	lower := strings.ToLower(uri)
	for k, v := range MapaRotaChave {
		if strings.ToLower(k) == lower {
			return v, true
		}
	}
	return "", false
}

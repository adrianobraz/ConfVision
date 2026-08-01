package auxiliar

import (
	"net/http"
	"text/template"
)

var templates *template.Template

func CarregarTemplates() {
	//MODULOS
	templates = template.Must(template.ParseGlob("recursos/templates/*.html"))

	// Login
	templates = template.Must(templates.ParseGlob("recursos/modulos/login/*.html"))

	// Alterar Senha
	templates = template.Must(templates.ParseGlob("recursos/modulos/alterar-senha/*.html"))

	// Menu Principal
	templates = template.Must(templates.ParseGlob("recursos/modulos/menu-principal/*.html"))

	// Em Manutençao
	templates = template.Must(templates.ParseGlob("recursos/modulos/emManutencao/*.html"))

	// MINHA EMPRESA
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/menu-minha-empresa/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/menu-operacional/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/menu-comercial/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/menu-gestao/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/dados-monitoramento/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciar-tecnico/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciar-cliente/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/parceiro-monitoramento/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciar-viatura/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciar-operador/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciar-ticket-franqueado/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciarUsuarios/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/gerenciarPacotes/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/permissoes-acesso/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/meu-plano/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/montar-plano/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/whitelabel/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/minha-empresa/dominio-saas/*.html"))

	// ATENDIMENTO
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/geradorEventos/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/menu-atendimento/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/inteligencia-artificial/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/listar-dispositivos-armados/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/listar-dispositivos-desarmados/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/listar-clientes-ativos/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/listar-clientes-inativos/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/atendimento/gerenciar-ticket-cliente/*.html"))

	// CLIENTES
	templates = template.Must(templates.ParseGlob("recursos/modulos/clientes/menu-clientes/*.html"))

	// RELATORIO
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/menu-relatorio/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/menu-eventos/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/menu-alarmes/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/menu-central-disparos/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/relatorio-atendimento/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/relatorio-eventos/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/relatorio-ligacoes/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/relatorios-disparo/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/relatorio/listar-clientes/*.html"))

	// CONFIGURAÇÃO
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/menu-configuracao/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-dispositivo/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-usuarios-alarme/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-setores-alarme/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-procedimento-atendimento/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-contactid-personalizado/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciar-grade/*.html"))
	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/configurar-relatorio-cliente/*.html"))

	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/gerenciarEnvioEvento/*.html"))

	templates = template.Must(templates.ParseGlob("recursos/modulos/configuracao/listar-cliente-dispositivo/*.html"))

}

func ExecutarTemplate(w http.ResponseWriter, template string, dados interface{}) {
	templates.ExecuteTemplate(w, template, dados)
}

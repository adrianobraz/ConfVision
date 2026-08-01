package roteador

import (
	"net/http"
	gerenciarProcedimentoAtendimento "franqueadopro/recursos/modulos/configuracao/gerenciar-procedimento-atendimento"
	menuPrincipal "franqueadopro/recursos/modulos/menu-principal"
	dadosMonitoramento "franqueadopro/recursos/modulos/minha-empresa/dados-monitoramento"
	gerenciarCliente "franqueadopro/recursos/modulos/minha-empresa/gerenciar-cliente"
	parceiroMonitoramento "franqueadopro/recursos/modulos/minha-empresa/parceiro-monitoramento"
	gerenciarOperador "franqueadopro/recursos/modulos/minha-empresa/gerenciar-operador"
	gerenciarTecnico "franqueadopro/recursos/modulos/minha-empresa/gerenciar-tecnico"
	gerenciarTicketFranqueado "franqueadopro/recursos/modulos/minha-empresa/gerenciar-ticket-franqueado"
	gerenciarViatura "franqueadopro/recursos/modulos/minha-empresa/gerenciar-viatura"
	"franqueadopro/recursos/modulos/minha-empresa/gerenciarPacotes"
	"franqueadopro/recursos/modulos/minha-empresa/gerenciarUsuarios"
	permissoesAcesso "franqueadopro/recursos/modulos/minha-empresa/permissoes-acesso"
	meuPlano "franqueadopro/recursos/modulos/minha-empresa/meu-plano"
	whitelabel "franqueadopro/recursos/modulos/minha-empresa/whitelabel"
	dominioSaas "franqueadopro/recursos/modulos/minha-empresa/dominio-saas"
	montarPlano "franqueadopro/recursos/modulos/minha-empresa/montar-plano"

	menuComercial "franqueadopro/recursos/modulos/minha-empresa/menu-comercial"
	menuGestao "franqueadopro/recursos/modulos/minha-empresa/menu-gestao"
	menuMinhaEmpresa "franqueadopro/recursos/modulos/minha-empresa/menu-minha-empresa"
	menuOperacional "franqueadopro/recursos/modulos/minha-empresa/menu-operacional"
	alterarSenha "franqueadopro/src/modulos/alterar-senha"
	menuClientes "franqueadopro/src/modulos/clientes/menu-clientes"
	"franqueadopro/src/modulos/atendimento/geradorEventos"
	inteligenciaArtificial "franqueadopro/recursos/modulos/atendimento/inteligencia-artificial"
	gerenciarTicketCliente "franqueadopro/src/modulos/atendimento/gerenciar-ticket-cliente"
	listarClientesAtivos "franqueadopro/src/modulos/atendimento/listar-clientes-ativos"
	listarClientesInativos "franqueadopro/src/modulos/atendimento/listar-clientes-inativos"
	listarDispositivoArmados "franqueadopro/src/modulos/atendimento/listar-dispositivos-armados"
	listarDispositivoDesarmado "franqueadopro/src/modulos/atendimento/listar-dispositivos-desarmados"
	menuAtendimento "franqueadopro/src/modulos/atendimento/menu-atendimento"
	configurarRelatorioCliente "franqueadopro/src/modulos/configuracao/configurar-relatorio-cliente"
	gerenciarContactIdPersonalizado "franqueadopro/src/modulos/configuracao/gerenciar-contactid-personalizado"
	gerenciarDispositivo "franqueadopro/src/modulos/configuracao/gerenciar-dispositivo"
	gerenciarGrade "franqueadopro/src/modulos/configuracao/gerenciar-grade"
	gerenciarSetoresAlarme "franqueadopro/src/modulos/configuracao/gerenciar-setores-alarme"
	gerenciarUsuarioAlarme "franqueadopro/src/modulos/configuracao/gerenciar-usuarios-alarme"
	"franqueadopro/src/modulos/configuracao/gerenciarEnvioEvento"
	listarClienteDispositivo "franqueadopro/src/modulos/configuracao/listar-cliente-dispositivo"
	menuConfiguracao "franqueadopro/src/modulos/configuracao/menu-configuracao"
	"franqueadopro/src/modulos/login"
	"franqueadopro/src/modulos/modAuxiliar"
	"franqueadopro/src/modulos/notificacao"
	"franqueadopro/src/modulos/feedback"
	listarClientes "franqueadopro/src/modulos/relatorio/listar-clientes"
	menuAlarmes "franqueadopro/src/modulos/relatorio/menu-alarmes"
	menuCentralDisparos "franqueadopro/src/modulos/relatorio/menu-central-disparos"
	menuEventos "franqueadopro/src/modulos/relatorio/menu-eventos"
	menuRelatorio "franqueadopro/src/modulos/relatorio/menu-relatorio"
	relatorioAtendimento "franqueadopro/src/modulos/relatorio/relatorio-atendimento"
	relatorioEventos "franqueadopro/src/modulos/relatorio/relatorio-eventos"
	relatorioLigacoes "franqueadopro/src/modulos/relatorio/relatorio-ligacoes"
	relatoriosDisparo "franqueadopro/src/modulos/relatorio/relatorios-disparo"

	"github.com/gorilla/mux"
)

func ConfigurarRotas() *mux.Router {
	r := mux.NewRouter()

	// Modulos
	rotas := login.Rotas

	// Auxiliar
	rotas = append(rotas, modAuxiliar.Rotas...)
	rotas = append(rotas, notificacao.Rotas...)
	rotas = append(rotas, feedback.Rotas...)

	// Alterar Senha
	rotas = append(rotas, alterarSenha.RotasAlterarSenha...)

	// Menu principal
	rotas = append(rotas, menuPrincipal.RotasMenuPrincipal...)

	// MINHA EMPRESA
	rotas = append(rotas, menuMinhaEmpresa.RotasMenuMinhaEmpresa...)
	rotas = append(rotas, menuOperacional.RotasMenuOperacional...)
	rotas = append(rotas, menuComercial.RotasMenuComercial...)
	rotas = append(rotas, menuGestao.RotasMenuGestao...)
	rotas = append(rotas, dadosMonitoramento.Rotas...)
	rotas = append(rotas, gerenciarTecnico.RotasGerenciarTecnico...)
	rotas = append(rotas, gerenciarCliente.Rotas...)
	rotas = append(rotas, parceiroMonitoramento.Rotas...)
	rotas = append(rotas, gerenciarViatura.RotasGerenciarViatura...)
	rotas = append(rotas, gerenciarOperador.Rotas...)
	rotas = append(rotas, gerenciarTicketFranqueado.Rotas...)
	rotas = append(rotas, gerenciarUsuarios.Rotas...)
	rotas = append(rotas, gerenciarPacotes.Rotas...)
	rotas = append(rotas, permissoesAcesso.Rotas...)
	rotas = append(rotas, meuPlano.Rotas...)
	rotas = append(rotas, montarPlano.Rotas...)
	rotas = append(rotas, whitelabel.Rotas...)
	rotas = append(rotas, dominioSaas.Rotas...)

	// ATENDIMENTO
	rotas = append(rotas, geradorEventos.Rotas...)
	rotas = append(rotas, inteligenciaArtificial.Rotas...)
	rotas = append(rotas, menuAtendimento.RotasMenuAtendimento...)
	rotas = append(rotas, listarDispositivoArmados.RotasListarClienteArmado...)
	rotas = append(rotas, listarDispositivoDesarmado.RotasListarClientesDesarmados...)
	rotas = append(rotas, listarClientesAtivos.RotasListarClientesAtivos...)
	rotas = append(rotas, listarClientesInativos.RotasListarClientesAtivos...)
	rotas = append(rotas, gerenciarTicketCliente.RotasGerenciarTicketCliente...)

	// CLIENTES
	rotas = append(rotas, menuClientes.RotasMenuClientes...)

	// RELATORIO
	rotas = append(rotas, menuRelatorio.RotasMenuRelatorio...)
	rotas = append(rotas, menuEventos.RotasMenuEventos...)
	rotas = append(rotas, menuAlarmes.RotasMenuAlarmes...)
	rotas = append(rotas, menuCentralDisparos.RotasMenuCentralDisparos...)
	rotas = append(rotas, relatorioAtendimento.Rotas...)
	rotas = append(rotas, relatorioEventos.Rotas...)
	rotas = append(rotas, relatorioLigacoes.RotasRelatorioLigacoes...)
	rotas = append(rotas, relatoriosDisparo.Rotas...)
	rotas = append(rotas, listarClientes.Rotas...)

	// CONFIGURACOES
	rotas = append(rotas, menuConfiguracao.RotasMenuConfiguracoes...)
	rotas = append(rotas, gerenciarUsuarioAlarme.Rotas...)
	rotas = append(rotas, gerenciarSetoresAlarme.Rotas...)
	rotas = append(rotas, gerenciarDispositivo.Rotas...)
	rotas = append(rotas, gerenciarProcedimentoAtendimento.Rotas...)
	rotas = append(rotas, gerenciarContactIdPersonalizado.RotasContactIdPersonalizado...)
	rotas = append(rotas, gerenciarGrade.RotasGerenciarGrade...)
	rotas = append(rotas, configurarRelatorioCliente.RotasConfigurarRelatorioCliente...)
	rotas = append(rotas, gerenciarEnvioEvento.Rotas...)
	rotas = append(rotas, listarClienteDispositivo.Rotas...)

	for _, rota := range rotas {
		if rota.Aberto {
			r.HandleFunc(rota.URI,

				Logger(rota.Funcao),
			).Methods(rota.Metodo)
		} else {
			r.HandleFunc(rota.URI,
				Logger(Autenticar(rota.Funcao)),
			).Methods(rota.Metodo)
		}

	}

	fileServer := http.FileServer(http.Dir("./recursos/"))
	r.PathPrefix("/recursos/").Handler(http.StripPrefix("/recursos/", fileServer))
	return r
}

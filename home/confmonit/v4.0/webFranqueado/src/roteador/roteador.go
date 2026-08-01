package roteador

import (
	"net/http"
	gerenciarProcedimentoAtendimento "webFranqueado/recursos/modulos/configuracao/gerenciar-procedimento-atendimento"
	menuPrincipal "webFranqueado/recursos/modulos/menu-principal"
	dadosMonitoramento "webFranqueado/recursos/modulos/minha-empresa/dados-monitoramento"
	gerenciarCliente "webFranqueado/recursos/modulos/minha-empresa/gerenciar-cliente"
	gerenciarOperador "webFranqueado/recursos/modulos/minha-empresa/gerenciar-operador"
	gerenciarTecnico "webFranqueado/recursos/modulos/minha-empresa/gerenciar-tecnico"
	gerenciarTicketFranqueado "webFranqueado/recursos/modulos/minha-empresa/gerenciar-ticket-franqueado"
	gerenciarViatura "webFranqueado/recursos/modulos/minha-empresa/gerenciar-viatura"
	"webFranqueado/recursos/modulos/minha-empresa/gerenciarPacotes"
	"webFranqueado/recursos/modulos/minha-empresa/gerenciarUsuarios"

	menuMinhaEmpresa "webFranqueado/recursos/modulos/minha-empresa/menu-minha-empresa"
	alterarSenha "webFranqueado/src/modulos/alterar-senha"
	"webFranqueado/src/modulos/atendimento/geradorEventos"
	gerenciarTicketCliente "webFranqueado/src/modulos/atendimento/gerenciar-ticket-cliente"
	listarClientesAtivos "webFranqueado/src/modulos/atendimento/listar-clientes-ativos"
	listarClientesInativos "webFranqueado/src/modulos/atendimento/listar-clientes-inativos"
	listarDispositivoArmados "webFranqueado/src/modulos/atendimento/listar-dispositivos-armados"
	listarDispositivoDesarmado "webFranqueado/src/modulos/atendimento/listar-dispositivos-desarmados"
	menuAtendimento "webFranqueado/src/modulos/atendimento/menu-atendimento"
	configurarRelatorioCliente "webFranqueado/src/modulos/configuracao/configurar-relatorio-cliente"
	gerenciarContactIdPersonalizado "webFranqueado/src/modulos/configuracao/gerenciar-contactid-personalizado"
	gerenciarDispositivo "webFranqueado/src/modulos/configuracao/gerenciar-dispositivo"
	gerenciarGrade "webFranqueado/src/modulos/configuracao/gerenciar-grade"
	gerenciarSetoresAlarme "webFranqueado/src/modulos/configuracao/gerenciar-setores-alarme"
	gerenciarUsuarioAlarme "webFranqueado/src/modulos/configuracao/gerenciar-usuarios-alarme"
	"webFranqueado/src/modulos/configuracao/gerenciarEnvioEvento"
	listarClienteDispositivo "webFranqueado/src/modulos/configuracao/listar-cliente-dispositivo"
	menuConfiguracao "webFranqueado/src/modulos/configuracao/menu-configuracao"
	"webFranqueado/src/modulos/login"
	"webFranqueado/src/modulos/modAuxiliar"
	listarClientes "webFranqueado/src/modulos/relatorio/listar-clientes"
	menuRelatorio "webFranqueado/src/modulos/relatorio/menu-relatorio"
	relatorioAtendimento "webFranqueado/src/modulos/relatorio/relatorio-atendimento"
	relatorioEventos "webFranqueado/src/modulos/relatorio/relatorio-eventos"
	relatorioLigacoes "webFranqueado/src/modulos/relatorio/relatorio-ligacoes"

	"github.com/gorilla/mux"
)

func ConfigurarRotas() *mux.Router {
	r := mux.NewRouter()

	// Modulos
	rotas := login.Rotas

	// Auxiliar
	rotas = append(rotas, modAuxiliar.Rotas...)

	// Alterar Senha
	rotas = append(rotas, alterarSenha.RotasAlterarSenha...)

	// Menu principal
	rotas = append(rotas, menuPrincipal.RotasMenuPrincipal...)

	// MINHA EMPRESA
	rotas = append(rotas, menuMinhaEmpresa.RotasMenuMinhaEmpresa...)
	rotas = append(rotas, dadosMonitoramento.Rotas...)
	rotas = append(rotas, gerenciarTecnico.RotasGerenciarTecnico...)
	rotas = append(rotas, gerenciarCliente.Rotas...)
	rotas = append(rotas, gerenciarViatura.RotasGerenciarViatura...)
	rotas = append(rotas, gerenciarOperador.Rotas...)
	rotas = append(rotas, gerenciarTicketFranqueado.Rotas...)
	rotas = append(rotas, gerenciarUsuarios.Rotas...)
	rotas = append(rotas, gerenciarPacotes.Rotas...)

	// ATENDIMENTO
	rotas = append(rotas, geradorEventos.Rotas...)
	rotas = append(rotas, menuAtendimento.RotasMenuAtendimento...)
	rotas = append(rotas, listarDispositivoArmados.RotasListarClienteArmado...)
	rotas = append(rotas, listarDispositivoDesarmado.RotasListarClientesDesarmados...)
	rotas = append(rotas, listarClientesAtivos.RotasListarClientesAtivos...)
	rotas = append(rotas, listarClientesInativos.RotasListarClientesAtivos...)
	rotas = append(rotas, gerenciarTicketCliente.RotasGerenciarTicketCliente...)

	// RELATORIO
	rotas = append(rotas, menuRelatorio.RotasMenuRelatorio...)
	rotas = append(rotas, relatorioAtendimento.Rotas...)
	rotas = append(rotas, relatorioEventos.Rotas...)
	rotas = append(rotas, relatorioLigacoes.RotasRelatorioLigacoes...)
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

$(document).ready(function () {
configurarBotoes()
})

function configurarBotoes(){
    $('#BtnGerenciarCadastroAlarme').on('click', function(){ 
        window.location.href = "/carregar-gerenciar-dispositivo"
    })

    $('#BtnGerenciarUsuarioAlarme').on('click', function(){ 
        window.location.href = "/carregar-gerenciar-usuarios-alarme"
    })

    $('#BtnGerenciarSetoresAlarme').on('click', function(){ 
        window.location.href = "/carregar-gerenciar-setores-alarme"
    })

    $('#BtnGerenciarProcedimentoAtendimento').on('click', function(){ 
        window.location.href = "/carregar-gerenciar-procedimento-atendimento"
    })

    $('#BtnGerenciarContactidPersonalizado').on('click', function(){ 
        window.location.href = "/carregar-gerenciar-contactid-personalizado"
    })

    $('#BtnGerenciarGradeHorario').on('click', function(){ 
        window.location.href = "/carregar-gerenciar-grade"
    })

    $('#BtnGerenciarRelatorioCliente').on('click', function(){ 
        window.location.href = "/carregar-configurar-relatorio-cliente"
    })

    $('#BtnGerenciarEnvioEvento').on('click', function(){ 
        window.location.href = "/gerenciar-configuracao-email-eveto"
    })
    $('#BtnListarClientesDispositivos').on('click', function(){ 
        window.location.href = "/listar-cliente-dispositivo"
    })
}
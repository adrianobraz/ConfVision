$(document).ready(function () {
    $(window).trigger('resize');
    DadosMonitoramentoAlturaFixa()


    // Associa o click do botao Gravar do alterar senha
    $('#BtnGravarSenha').on('click', DadosMonitoramentoGravarSenha)

    $('#BtnAtivarEmail').on('click', DadosMonitoramentoAtivarEmail)

    DadosMonitoramentoBuscar()
})

// Redimenciona a tela 
$(window).resize(DadosMonitoramentoAlturaFixa)

function DadosMonitoramentoAlturaFixa() {
    $('.altura-fixa').height(window.innerHeight - 180)
}

function DadosMonitoramentoAtivarEmail(id) {
    const idFranqueado = localStorage.getItem('idFranqueado')
    const url = `/DadosMonitoramentoEnvioEmailAtivo`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'Post',
        data: JSON.stringify({ FraId: idFranqueado })
    }).fail(function (e) {
        boxErro('Erro ao Habilitar/Desabilitar o envio de email')
    }).done(function (r) {
        if (r.dados == "S") {
            $('#BtnAtivarEmail').removeClass("btn-secondary")
            $('#BtnAtivarEmail').addClass("btn-success")
            $('#BtnAtivarEmail').text("DESATIVAR ENVIO DE EMAIL")
        } else {
            $('#BtnAtivarEmail').removeClass("btn-success")
            $('#BtnAtivarEmail').addClass("btn-secondary")
            $('#BtnAtivarEmail').text("ATIVAR ENVIO DE EMAIL")
        }
        boxFechar()
    })
}

function DadosMonitoramentoBuscar() {
    const idFranqueado = localStorage.getItem('idFranqueado')
    const url = `/DadosMonitoramentoBuscar`
    $.ajax({
        url: url,
        method: 'Post',
        data: JSON.stringify({ fraId: idFranqueado })
    }).fail(function (e) {
        boxErro('Erro ao buscar os dados do monitoramento')
    }).done(function (r) {       
        const d = r.dados

        //######################### FRANQUEADO ########################//
        if (d.emailEnvio == 'S') {
            $('#BtnAtivarEmail').removeClass("btn-secondary")
            $('#BtnAtivarEmail').addClass("btn-success")
            $('#BtnAtivarEmail').text("DESATIVAR ENVIO DE EMAIL")
        } else {
            $('#BtnAtivarEmail').removeClass("btn-success")
            $('#BtnAtivarEmail').addClass("btn-secondary")
            $('#BtnAtivarEmail').text("ATIVAR ENVIO DE EMAIL")
        }

        //////
        $('#razao-social').val(d.fraRazao)
        $('#nome-fantasia').val(d.fraNome)
        $('#cnpj').val(d.fraCnpj)
        $('#escricao-estadual').val(d.fraInscricaoEstadual)
        $('#cep').val(d.fraCep)
        $('#uf').val(d.fraUf)
        $('#endereco').val(d.fraEndereco)
        $('#bairro').val(d.fraBairro)
        $('#cidade').val(d.fraCidade)
        $('#codigo-benuvem').val(d.fraCodBenuvem)
        $('#observacao').val(d.fraObservacao)

        //#############################################################//


        //######################## Responsavel ########################//
        const resp = d.userDados
        $('#nome-responsavel').val(resp.nome)
        $('#nick-responsavel').val(resp.nick)
        $('#user-voip-responsavel').val(resp.usuarioVoip)
        $('#email-principal-responsavel').val(resp.email1)
        $('#email-alternativo-responsavel').val(resp.email2)
        $('#telefone-principal-responsavel').val(formatarCelular(resp.telefone1))
        $('#telefone-alternativo-responsavel').val(formatarCelular(resp.telefone2))
        //#############################################################//
    })
}

function DadosMonitoramentoGravarSenha() {
    const idUsuario = localStorage.getItem('idUsuario')
    const senha = $("#nova-senha-franqueado").val()
    const repSenha = $("#repetir-nova-senha-franqueado").val()
    if (senha == "") {
        boxAdvertenciaAuto('O campo senha não pode estar vazio')
        return
    }
    if (repSenha == "") {
        boxAdvertenciaAuto('O campo repetir senha não pode estar vazio')
        return
    }
    if (senha != repSenha) {
        boxAdvertenciaAuto('As senhas são diferentes')
        return
    }
    const url = `/DadosMonitoramentoGravarSenha`
    $.ajax({
        url: url,
        method: 'POST',
        data: JSON.stringify({
            idUsuario: idUsuario,
            senha: senha
        })
    }).fail(function (e) {
        boxErro('Erro ao gravar a senha')
    }).done(function (r) {
        boxSucessoAuto('Senha auterada com sucesso')
        $("#nova-senha-franqueado").val("")
        $("#repetir-nova-senha-franqueado").val("")
    })
}


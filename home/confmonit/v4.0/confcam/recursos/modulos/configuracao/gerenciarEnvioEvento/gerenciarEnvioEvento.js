$(document).ready(function () {

    $(window).on('resize', function () {
        ajustaTabela()
    })

    localStorage.setItem('id', '0')

    // Associa o click do botao limpar
    
    $('#idCliente').on('change', emailEventoCarregarDispositivos)
    $('#idDisp').on('change', emailEventoCarregarConfiguracao)

    // Associa o click do botao gravar
    $('#btn-gravar').on('click', emailEventoGravar)
    $('#btn-limpar').on('click', emailEventoLimpar)

    ajustaTabela()
    emailEventoCarregarClientes()
})

function emailEventoCarregarClientes() {
    const idFranqueado = localStorage.getItem('idFranqueado')

    $.ajax({
        url: '/emailEventoCarregarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: idFranqueado })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        console.log(r)
        $('#idCliente').empty()
        $('#idCliente').append('<option value="0">Selecione</option>')

        r.dados.forEach(i => {
            $('#idCliente').append(`
                <option value="${i.idCliente}">${i.nome}</option>
            `)
        });

    })
}

function emailEventoCarregarDispositivos() {
    const idCliente = $('#idCliente').val()

    if (idCliente == '0'){
        emailEventoLimpar()
        $('#idDisp').val('0')
    }

    $.ajax({
        url: '/emailEventoCarregarDispositivo',
        method: 'Post',
        data: JSON.stringify({ idCliente: idCliente })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        console.log(r)
        $('#idDisp').empty()
        $('#idDisp').append('<option value="0">Selecione</option>')

        if ($('#idCliente').val() != '0'){
            r.dados.forEach(i => {
                $('#idDisp').append(`
                    <option value="${i.idDispositivo}">${i.nome}</option>
                `)
            });
        }

    })
}

function emailEventoCarregarConfiguracao() {
    const idDisp = $('#idDisp').val()

    if (idDisp == '0') {
        emailEventoLimpar()
        return
    }
    
    $.ajax({
        url: '/emailEventoCarregarConfiguracao',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: idDisp })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar configurações")
    }).done(function (r) {
        console.log('>>>',r)
        if (r.status != 'Vazio') {
            const i = r.dados
            localStorage.setItem('id', i.idSetupEnvioEvento)
            $('#EmailAlarme').val(i.emailAlarme)
            $('#EmailArme').val(i.emailArme)
            $('#EmailDesarme').val(i.emailDesarme)
            $('#EmailEmergencia').val(i.emailEmergencia)
            $('#EmailFalhas').val(i.emailFalhas)
            $('#EmailGeral').val(i.emailGeral)
            $('#EmailMedico').val(i.emailMedico)
            $('#EmailPanico').val(i.emailPanico)
            $('#EmailRestaure').val(i.emailRestaure)
            $('#EmailSetup').val(i.emailSetup)
            $('#EmailTeste').val(i.emailTeste)
            $('#EmailSemComunicar').val(i.emailSemComunicar)
        }

    })
}

function emailEventoGravar() {
    if (localStorage.getItem('id') != '0'){

        $.ajax({
            start: boxProcessando(),
            url: 'emailEventoGravar',
            method: 'Post',
            data: JSON.stringify({
                idSetupEnvioEvento: localStorage.getItem('id'),
                emailAlarme: $('#EmailAlarme').val(),
                emailArme: $('#EmailArme').val(),
                emailDesarme: $('#EmailDesarme').val(),
                emailEmergencia: $('#EmailEmergencia').val(),
                emailFalhas: $('#EmailFalhas').val(),
                emailGeral: $('#EmailGeral').val(),
                emailMedico: $('#EmailMedico').val(),
                emailPanico: $('#EmailPanico').val(),
                emailRestaure: $('#EmailRestaure').val(),
                emailSetup: $('#EmailSetup').val(),
                emailTeste: $('#EmailTeste').val(),
                emailSemComunicar: $('#EmailSemComunicar').val(),
            })
        }).fail(function (e) {
            console.log(e)
            boxErro("Erro ao gravar permissões")
        }).done(function (r) {
            boxAteradoSucesso()

            emailEventoCarregarConfiguracao()
        })
    }else {
        boxMesagemAtencaoPersonalizada('um dispositivo deve ser selecionado')
    }
}

function emailEventoLimpar() {
    $('#EmailAlarme').val('N')
    $('#EmailArme').val('N')
    $('#EmailDesarme').val('N')
    $('#EmailEmergencia').val('N')
    $('#EmailFalhas').val('N')
    $('#EmailGeral').val('N')
    $('#EmailMedico').val('N')
    $('#EmailPanico').val('N')
    $('#EmailRestaure').val('N')
    $('#EmailSetup').val('N')
    $('#EmailTeste').val('N')
    $('#EmailSemComunicar').val('N')
}
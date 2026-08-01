$(document).ready(function () {
    $('#btn-limpar').on('click', ConfigurarRelatorioClienteLimpar)
    $('#btn-gravar').on('click', ConfigurarRelatorioClienteAlterar)
    $('#id-cliente').on('change', ConfigurarRelatorioClienteBuscar)
   

    ConfigurarRelatorioClientesListar()
    localStorage.setItem('id', '0')
})

function ConfigurarRelatorioBuscarDispositivo() {
    const idCliente = $('#id-cliente').val()
    if (idCliente == '0') {
        ConfigurarRelatorioClienteLimpar()
        $('#id-dispositivo').val('0')
        return
    }

    $.ajax({
        url: '/ConfigurarRelatorioListarDispositivo',
        method: 'Post',
        data: JSON.stringify({ idCliente: idCliente })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        console.log('dispositivo', r)
        $('#id-dispositivo').empty()
        $('#id-dispositivo').append('<option value="0">SELECIONE UM DISPOSITIVO</option>')

        r.dados.forEach(i => {
            $('#id-dispositivo').append(`
                <option value="${i.idDispositivo}">${i.nome}</option>
            `)
        });

    })

}

function ConfigurarRelatorioClienteLimpar() {
    localStorage.setItem('id', '0')
    $('#formulario').each(function () {
        this.reset();
    })
}

function ConfigurarRelatorioClientesListar() {

    $.ajax({
        url: '/ConfigurarRelatorioClientesListar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        console.log(r)
        $('#id-cliente').empty()
        $('#id-cliente').append(`
            <option value="0">SELECIONE UM CLIENTE</option>
        `)

        r.dados.forEach(i => {
            $('#id-cliente').append(`
                <option value="${i.idCliente}">${i.nome}</option>
            `)
        });

    })
}

function ConfigurarRelatorioClienteBuscar() {

    const idCliente = $('#id-cliente').val()

    if (idCliente == '0') {
        ConfigurarRelatorioClienteLimpar()
        return
    }

    $.ajax({
        start: boxProcessando(),
        url: '/RelatorioEventoPermisaoBuscar',
        method: 'Post',
        data: JSON.stringify({ idCliente: idCliente })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        console.log(r)
        boxFechar()
        // const permisao = Array.from(r.dados);
        const d = r.dados
        if (r.status == 'OK') {
            localStorage.setItem('id', d.idSetupRelatorio)
            $('#chk-alarme').val(d.alarme)
            $('#chk-arme').val(d.arme)
            $('#chk-desarme').val(d.desarme)
            $('#chk-emergencia').val(d.emergencia)
            $('#chk-falhas').val(d.falhas)
            $('#chk-geral').val(d.geral)
            $('#chk-medico').val(d.medico)
            $('#chk-panico').val(d.panico)
            $('#chk-restaure').val(d.restaure)
            $('#chk-setup').val(d.setup)
            $('#chk-teste').val(d.teste)
        }
        
    })
}

function ConfigurarRelatorioClienteAlterar() {
    if (localStorage.getItem('id') != '0') {
        $.ajax({
            start: boxProcessando(),
            url: '/RelatorioEventoPermisaoAlterar',
            method: 'Post',
            data: JSON.stringify({
                idSetupRelatorio: localStorage.getItem('id'),
                alarme: $('#chk-alarme').val(),
                arme: $('#chk-arme').val(),
                desarme: $('#chk-desarme').val(),
                emergencia: $('#chk-emergencia').val(),
                falhas: $('#chk-falhas').val(),
                geral: $('#chk-geral').val(),
                medico: $('#chk-medico').val(),
                panico: $('#chk-panico').val(),
                restaure: $('#chk-restaure').val(),
                setup: $('#chk-setup').val(),
                teste: $('#chk-teste').val(),
            })
        }).fail(function (e) {
            console.log(e)
            boxErro("Erro ao carregar a tabela")
        }).done(function (r) {
            boxAteradoSucesso()
            ConfigurarRelatorioClienteLimpar()
        })
    }

}
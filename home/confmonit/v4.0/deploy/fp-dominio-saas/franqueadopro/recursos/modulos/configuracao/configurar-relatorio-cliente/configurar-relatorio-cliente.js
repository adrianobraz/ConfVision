var listaClientes = []

$(document).ready(function () {
    $('#btn-limpar').on('click', ConfigurarRelatorioClienteLimpar)
    $('#btn-gravar').on('click', ConfigurarRelatorioClienteAlterar)
    $('#id-cliente').on('change', ConfigurarRelatorioClienteBuscar)

    // Pesquisa de cliente no combo (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', ConfigurarRelatorioFiltrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') ConfigurarRelatorioFiltrarClientes()
    })
    $('#lista-clientes').on('click', '.fp-combo-item', function () {
        $('#busca-cliente').val($(this).attr('data-nome'))
        $('#id-cliente').val($(this).attr('data-id')).trigger('change')
        $('#lista-clientes').addClass('d-none')
    })
    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-cliente, #lista-clientes').length) {
            $('#lista-clientes').addClass('d-none')
        }
    })

    ConfigurarRelatorioClientesListar()
    localStorage.setItem('id', '0')
})

function clienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function ConfigurarRelatorioFiltrarClientes() {
    var raw = ($('#busca-cliente').val() || '').toLowerCase().trim()
    var termoNum = raw.replace(/[^a-z0-9]/g, '')
    var lista = $('#lista-clientes')
    lista.empty()

    if (raw == '') {
        $('#id-cliente').val('0').trigger('change')
        lista.addClass('d-none')
        return
    }

    var resultados = listaClientes.filter(function (i) {
        var texto = clienteTextoBusca(i)
        var textoNum = texto.replace(/[^a-z0-9]/g, '')
        return texto.indexOf(raw) != -1 || (termoNum != '' && textoNum.indexOf(termoNum) != -1)
    }).slice(0, 30)

    if (resultados.length == 0) {
        lista.append('<li class="fp-combo-empty">Nenhum cliente encontrado</li>')
    } else {
        resultados.forEach(function (i) {
            var doc = i.documento1 ? ' — ' + i.documento1 : ''
            $('<li class="fp-combo-item"></li>')
                .attr('data-id', i.idCliente)
                .attr('data-nome', i.nome)
                .text(i.nome + doc)
                .appendTo(lista)
        })
    }
    lista.removeClass('d-none')
}

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
    $('#busca-cliente').val('')
    $('#id-cliente').val('0')
    $('#lista-clientes').empty().addClass('d-none')
}

function ConfigurarRelatorioClientesListar() {

    $.ajax({
        url: '/ConfigurarRelatorioClientesListar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        listaClientes = (r.dados || [])
        $('#id-cliente').val('0')
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
var iaListaClientes = []

var IA_AJAX = {
    contentType: 'application/json; charset=utf-8',
    dataType: 'json'
}

$(document).ready(function () {
    $('#btn-salvar-telefones').on('click', iaSalvarTelefones)
    $('#btn-adicionar-bloqueio').on('click', iaAdicionarBloqueio)
    $('#busca-cliente-ia').on('input', iaFiltrarClientes)

    iaCarregarTelefones()
    iaCarregarBloqueios()
    iaCarregarClientes()
})

function iaIdFranqueado() {
    return localStorage.getItem('idFranqueado')
}

function iaCarregarTelefones() {
    $.ajax($.extend({}, IA_AJAX, {
        url: '/iaTelefonesCarregar',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: iaIdFranqueado() })
    })).fail(function () {
        boxErro('Erro ao carregar telefones da IA')
    }).done(function (r) {
        var d = r.dados || {}
        $('#tel-monitoramento').val(d.telMonitoramento || '')
        $('#tel-gerente').val(d.telGerente || '')
        $('#tel-viatura').val(d.telViatura || '')
        $('#hora-ini').val(iaNormalizarHora(d.horaIni))
        $('#hora-fin').val(iaNormalizarHora(d.horaFin))
    })
}

function iaNormalizarHora(val) {
    if (!val) return ''
    var s = String(val).trim()
    if (s.length >= 5) return s.substring(0, 5)
    return s
}

function iaSalvarTelefones() {
    var telViatura = ($('#tel-viatura').val() || '').trim()
    var horaIni = $('#hora-ini').val()
    var horaFin = $('#hora-fin').val()

    if (telViatura && (!horaIni || !horaFin)) {
        boxMesagemAtencaoPersonalizada('Informe hora início e fim para o telefone da viatura.')
        return
    }

    boxProcessando('Gravando telefones...')
    $.ajax($.extend({}, IA_AJAX, {
        url: '/iaTelefonesSalvar',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: iaIdFranqueado(),
            telMonitoramento: ($('#tel-monitoramento').val() || '').trim(),
            telGerente: ($('#tel-gerente').val() || '').trim(),
            telViatura: telViatura,
            horaIni: horaIni,
            horaFin: horaFin
        })
    })).fail(function (xhr) {
        Swal.close()
        iaErroMsg(xhr, 'Erro ao gravar telefones')
    }).done(function () {
        Swal.close()
        boxAteradoSucesso()
        iaCarregarTelefones()
    })
}

function iaCarregarBloqueios() {
    $.ajax($.extend({}, IA_AJAX, {
        url: '/iaBloqueioListar',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: iaIdFranqueado() })
    })).fail(function () {
        boxErro('Erro ao carregar lista de bloqueio')
    }).done(function (r) {
        var lista = Array.isArray(r.dados) ? r.dados : []
        iaRenderBloqueios(lista)
    })
}

function iaRenderBloqueios(lista) {
    var html = ''
    if (!lista.length) {
        html = '<tr><td colspan="2" class="text-center text-muted py-3">Nenhum cliente bloqueado</td></tr>'
    } else {
        lista.forEach(function (item) {
            html += '<tr data-id="' + item.id + '">' +
                '<td class="text-start">' + iaEscape(item.nomeCliente || '') +
                '<br><small class="text-muted">ID: ' + iaEscape(item.idCliente || '') + '</small></td>' +
                '<td class="text-center">' +
                '<button type="button" class="btn btn-sm btn-danger btn-remover-bloqueio" data-id="' + item.id + '">' +
                '<i class="bi bi-trash-fill"></i></button></td></tr>'
        })
    }
    $('#lista-bloqueios').html(html)
    $('#lista-bloqueios .btn-remover-bloqueio').on('click', function () {
        iaRemoverBloqueio($(this).data('id'))
    })
}

function iaAdicionarBloqueio() {
    var idCliente = $('#ia-id-cliente').val()
    var nomeCliente = $('#ia-nome-cliente').val()
    if (!idCliente || !nomeCliente) {
        boxMesagemAtencaoPersonalizada('Selecione um cliente na busca.')
        return
    }

    $.ajax($.extend({}, IA_AJAX, {
        url: '/iaBloqueioAdicionar',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: iaIdFranqueado(),
            idCliente: idCliente,
            nomeCliente: nomeCliente
        })
    })).fail(function (xhr) {
        iaErroMsg(xhr, 'Erro ao adicionar bloqueio')
    }).done(function () {
        $('#busca-cliente-ia').val('')
        $('#ia-id-cliente').val('')
        $('#ia-nome-cliente').val('')
        $('#lista-clientes-ia').addClass('d-none').empty()
        iaCarregarBloqueios()
        boxSucessoAuto('Cliente adicionado à lista de bloqueio')
    })
}

function iaRemoverBloqueio(id) {
    $.ajax($.extend({}, IA_AJAX, {
        url: '/iaBloqueioRemover',
        method: 'POST',
        data: JSON.stringify({ id: id })
    })).fail(function () {
        boxErro('Erro ao remover bloqueio')
    }).done(function () {
        iaCarregarBloqueios()
    })
}

function iaCarregarClientes() {
    $.ajax($.extend({}, IA_AJAX, {
        url: '/carregarClientes',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: iaIdFranqueado() })
    })).fail(function () {
        console.log('Erro ao carregar clientes')
    }).done(function (r) {
        iaListaClientes = r.dados || []
    })
}

function iaFiltrarClientes() {
    var raw = ($('#busca-cliente-ia').val() || '').toLowerCase().trim()
    var lista = $('#lista-clientes-ia')
    lista.empty()

    if (!raw) {
        $('#ia-id-cliente').val('')
        $('#ia-nome-cliente').val('')
        lista.addClass('d-none')
        return
    }

    var filtrados = iaListaClientes.filter(function (c) {
        var txt = [(c.nome || ''), (c.nick || ''), (c.documento1 || ''), (c.documento2 || '')]
            .join(' ').toLowerCase()
        return txt.indexOf(raw) !== -1
    }).slice(0, 15)

    if (!filtrados.length) {
        lista.removeClass('d-none')
        lista.append('<div class="fp-ia-cliente-item text-muted">Nenhum cliente encontrado</div>')
        return
    }

    filtrados.forEach(function (c) {
        var id = c.idCliente || c.id || ''
        var nome = c.nome || c.nick || id
        lista.append(
            '<div class="fp-ia-cliente-item" data-id="' + iaEscapeAttr(id) + '" data-nome="' + iaEscapeAttr(nome) + '">' +
            iaEscape(nome) + '</div>'
        )
    })

    lista.removeClass('d-none')
    lista.find('.fp-ia-cliente-item').on('click', function () {
        $('#ia-id-cliente').val($(this).data('id'))
        $('#ia-nome-cliente').val($(this).data('nome'))
        $('#busca-cliente-ia').val($(this).data('nome'))
        lista.addClass('d-none')
    })
}

function iaErroMsg(xhr, padrao) {
    var msg = padrao
    try {
        if (xhr && xhr.responseJSON && xhr.responseJSON.status) {
            msg = String(xhr.responseJSON.status).replace(/^Erro:\s*/i, '')
        }
    } catch (e) { /* ignore */ }
    boxErro(msg)
}

function iaEscape(txt) {
    return String(txt || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function iaEscapeAttr(txt) {
    return iaEscape(txt)
}

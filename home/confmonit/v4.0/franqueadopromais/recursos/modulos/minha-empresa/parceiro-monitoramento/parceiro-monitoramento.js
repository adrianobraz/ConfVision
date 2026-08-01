var pmListaClientes = []
var pmParceiros = []

var PM_AJAX = {
    contentType: 'application/json; charset=utf-8',
    dataType: 'json'
}

$(document).ready(function () {
    $('#pm-busca-cliente').on('input', pmFiltrarClientes)
    $('#pm-parceiro').on('change', pmAtualizarPrecoHint)
    $('#btn-pm-salvar').on('click', pmSalvarVinculo)

    pmCarregarClientes()
    pmCarregarParceiros()
    pmCarregarVinculos()
    pmCarregarFaturas()
})

function pmIdFranqueado() {
    return localStorage.getItem('idFranqueado')
}

function pmCarregarClientes() {
    $.ajax($.extend({}, PM_AJAX, {
        url: '/carregarClientes',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: pmIdFranqueado() })
    })).done(function (r) {
        pmListaClientes = r.dados || []
    })
}

function pmCarregarParceiros() {
    $.ajax($.extend({}, PM_AJAX, {
        url: '/pmParceirosListar',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: pmIdFranqueado() })
    })).fail(function () {
        boxErro('Erro ao carregar parceiros de monitoramento')
    }).done(function (r) {
        pmParceiros = Array.isArray(r.dados) ? r.dados : []
        var html = '<option value="">Selecione...</option>'
        if (!pmParceiros.length) {
            html += '<option value="" disabled>Nenhum parceiro liberado — fale com seu representante</option>'
        }
        pmParceiros.forEach(function (p) {
            var nome = p.nomeFantasia || p.razaoSocial || p.id
            html += '<option value="' + pmEscapeAttr(p.id) + '" data-preco="' + Number(p.precoClienteQuinzena || 0) + '">' +
                pmEscape(nome) + ' (' + pmEscape(p.software || '') + ') — R$ ' +
                Number(p.precoClienteQuinzena || 0).toFixed(2) + '/quinzena</option>'
        })
        $('#pm-parceiro').html(html)
    })
}

function pmAtualizarPrecoHint() {
    var opt = $('#pm-parceiro option:selected')
    var preco = opt.data('preco')
    if (opt.val() && preco !== undefined) {
        $('#pm-preco-hint').text('Você será cobrado R$ ' + Number(preco).toFixed(2) + ' por quinzena neste cliente.')
    } else {
        $('#pm-preco-hint').text('')
    }
}

function pmFiltrarClientes() {
    var raw = ($('#pm-busca-cliente').val() || '').toLowerCase().trim()
    var lista = $('#pm-lista-clientes')
    lista.empty()
    if (!raw) {
        $('#pm-id-cliente').val('')
        $('#pm-nome-cliente').val('')
        lista.addClass('d-none')
        return
    }
    var filtrados = pmListaClientes.filter(function (c) {
        var txt = [(c.nome || ''), (c.nick || ''), (c.documento1 || ''), (c.documento2 || '')].join(' ').toLowerCase()
        return txt.indexOf(raw) !== -1
    }).slice(0, 15)

    if (!filtrados.length) {
        lista.addClass('d-none')
        return
    }
    filtrados.forEach(function (c) {
        var nome = c.nome || c.nick || c.id
        var btn = $('<button type="button"></button>').text(nome + ' — ' + (c.id || ''))
        btn.on('click', function () {
            $('#pm-busca-cliente').val(nome)
            $('#pm-id-cliente').val(c.id)
            $('#pm-nome-cliente').val(nome)
            lista.addClass('d-none').empty()
        })
        lista.append(btn)
    })
    lista.removeClass('d-none')
}

function pmSalvarVinculo() {
    var idCliente = $('#pm-id-cliente').val()
    var nomeCliente = $('#pm-nome-cliente').val()
    var idParceiro = $('#pm-parceiro').val()
    if (!idCliente || !idParceiro) {
        boxMesagemAtencaoPersonalizada('Selecione o cliente e o parceiro.')
        return
    }
    boxProcessando('Vinculando...')
    $.ajax($.extend({}, PM_AJAX, {
        url: '/pmVinculoSalvar',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: pmIdFranqueado(),
            idCliente: idCliente,
            nomeCliente: nomeCliente,
            idParceiro: idParceiro,
            contaExterna: ($('#pm-conta-externa').val() || '').trim()
        })
    })).fail(function (xhr) {
        Swal.close()
        pmErro(xhr, 'Erro ao vincular parceiro')
    }).done(function () {
        Swal.close()
        boxAteradoSucesso()
        $('#pm-busca-cliente').val('')
        $('#pm-id-cliente').val('')
        $('#pm-nome-cliente').val('')
        $('#pm-conta-externa').val('')
        $('#pm-parceiro').val('')
        pmAtualizarPrecoHint()
        pmCarregarVinculos()
        pmCarregarFaturas()
    })
}

function pmCarregarVinculos() {
    $.ajax($.extend({}, PM_AJAX, {
        url: '/pmVinculosListar',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: pmIdFranqueado() })
    })).fail(function () {
        $('#pm-lista-vinculos').html('<tr><td colspan="4" class="text-muted text-center">Erro ao carregar</td></tr>')
    }).done(function (r) {
        var lista = Array.isArray(r.dados) ? r.dados : []
        if (!lista.length) {
            $('#pm-lista-vinculos').html('<tr><td colspan="4" class="text-muted text-center py-3">Nenhum cliente terceirizado</td></tr>')
            return
        }
        var html = ''
        lista.forEach(function (v) {
            html += '<tr>' +
                '<td>' + pmEscape(v.nomeCliente || v.idCliente) +
                '<br><small class="text-muted">' + pmEscape(v.idCliente) + '</small></td>' +
                '<td>' + pmEscape(v.parceiroNome || '') +
                '<br><small class="text-muted">' + pmEscape(v.software || '') + '</small></td>' +
                '<td>R$ ' + Number(v.precoCongelado || 0).toFixed(2) + '</td>' +
                '<td class="text-center"><button type="button" class="btn btn-sm btn-danger btn-pm-remover" data-cliente="' +
                pmEscapeAttr(v.idCliente) + '"><i class="bi bi-trash"></i></button></td></tr>'
        })
        $('#pm-lista-vinculos').html(html)
        $('#pm-lista-vinculos .btn-pm-remover').on('click', function () {
            pmRemoverVinculo($(this).data('cliente'))
        })
    })
}

function pmRemoverVinculo(idCliente) {
    $.ajax($.extend({}, PM_AJAX, {
        url: '/pmVinculoRemover',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: pmIdFranqueado(),
            idCliente: idCliente
        })
    })).fail(function () {
        boxErro('Erro ao remover vínculo')
    }).done(function () {
        pmCarregarVinculos()
        boxSucessoAuto('Vínculo removido — cliente volta ao monitoramento padrão')
    })
}

function pmCarregarFaturas() {
    $.ajax($.extend({}, PM_AJAX, {
        url: '/pmFaturasListar',
        method: 'POST',
        data: JSON.stringify({ idFranqueado: pmIdFranqueado() })
    })).fail(function () {
        $('#pm-lista-faturas').html('<tr><td colspan="3" class="text-muted text-center">—</td></tr>')
    }).done(function (r) {
        var lista = Array.isArray(r.dados) ? r.dados : []
        if (!lista.length) {
            $('#pm-lista-faturas').html('<tr><td colspan="3" class="text-muted text-center py-3">Nenhuma fatura ainda</td></tr>')
            return
        }
        var html = ''
        lista.forEach(function (f) {
            html += '<tr><td>' + pmEscape(f.periodoInicio) + ' a ' + pmEscape(f.periodoFim) +
                '</td><td>R$ ' + Number(f.valorBruto || 0).toFixed(2) +
                '</td><td>' + pmEscape(f.status || '') + '</td></tr>'
        })
        $('#pm-lista-faturas').html(html)
    })
}

function pmErro(xhr, fallback) {
    var msg = fallback
    try {
        var j = xhr.responseJSON
        if (j && (j.erro || j.message)) msg = j.erro || j.message
    } catch (e) {}
    boxErro(msg)
}

function pmEscape(s) {
    return String(s == null ? '' : s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function pmEscapeAttr(s) {
    return pmEscape(s).replace(/'/g, '&#39;')
}

let cacheFaturas = []

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return

    try {
        confVisionAuthGuard()
        confVisionCarregarCabecalho()
    } catch (e) {
        console.error(e)
    }

    $('#btn-atualizar-faturas').on('click', carregarRelatorioFaturas)
    $('#filtro-status-faturas, #filtro-tipo-faturas').on('change', renderTabelaFaturas)
    $('#busca-faturas').on('input', function () {
        clearTimeout(window._cvBuscaFat)
        window._cvBuscaFat = setTimeout(renderTabelaFaturas, 250)
    })

    carregarRelatorioFaturas()
})

function escHtmlFat(valor) {
    if (valor == null || valor === '') return '—'
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function formatMoneyFat(v) {
    const n = parseFloat(v)
    if (isNaN(n)) return '—'
    return 'R$ ' + n.toFixed(2).replace('.', ',')
}

function formatDataFat(valor) {
    if (!valor) return '—'
    if (typeof confVisionFormatarData === 'function') {
        const txt = confVisionFormatarData(valor)
        return txt.indexOf(' ') > 0 ? txt.split(' ')[0] : txt
    }
    const d = new Date(valor)
    if (isNaN(d.getTime())) return String(valor)
    return d.toLocaleDateString('pt-BR')
}

function labelTipoFat(tipo) {
    const t = String(tipo || '')
    if (t === 'confvision_venda') return 'Compra'
    if (t === 'confvision_renovacao') return 'Renovação'
    return t || '—'
}

function badgeStatusFat(status) {
    const s = String(status || '').toLowerCase()
    if (s === 'paga') {
        return '<span class="cv-badge cv-badge-on">Paga</span>'
    }
    if (s === 'aberta') {
        return '<span class="cv-badge cv-badge-warn">Em aberto</span>'
    }
    if (s === 'cancelada') {
        return '<span class="cv-badge cv-badge-off">Cancelada</span>'
    }
    return '<span class="cv-badge cv-badge-off">' + escHtmlFat(status) + '</span>'
}

function normalizarListaFat(r) {
    if (!r) return []
    if (Array.isArray(r)) return r
    if (Array.isArray(r.dados)) return r.dados
    return []
}

function ehFaturaConfVision(f) {
    const t = String((f && f.tipo) || '')
    return t === 'confvision_venda' || t === 'confvision_renovacao'
}

function atualizarResumoFat(lista) {
    let abertas = 0
    let pagas = 0
    let valorAberto = 0
    ;(lista || []).forEach(function (f) {
        const st = String(f.status || '').toLowerCase()
        if (st === 'aberta') {
            abertas++
            valorAberto += parseFloat(f.valor_total) || 0
        } else if (st === 'paga') {
            pagas++
        }
    })
    $('#lbl-fat-total').text((lista || []).length)
    $('#lbl-fat-abertas').text(abertas)
    $('#lbl-fat-pagas').text(pagas)
    $('#lbl-fat-valor-aberto').text(formatMoneyFat(valorAberto))
}

function filtrarFaturas(lista) {
    const termo = ($('#busca-faturas').val() || '').toLowerCase().trim()
    const status = ($('#filtro-status-faturas').val() || '').trim()
    const tipo = ($('#filtro-tipo-faturas').val() || '').trim()

    return (lista || []).filter(function (f) {
        if (!f) return false
        if (status && String(f.status || '') !== status) return false
        if (tipo && String(f.tipo || '') !== tipo) return false
        if (!termo) return true

        const ref = String(f.referencia || '').toLowerCase()
        const tip = labelTipoFat(f.tipo).toLowerCase()
        const obs = String(f.observacao || '').toLowerCase()
        const id = String(f.id || '').toLowerCase()

        return ref.indexOf(termo) !== -1
            || tip.indexOf(termo) !== -1
            || obs.indexOf(termo) !== -1
            || id.indexOf(termo) !== -1
    })
}

function renderTabelaFaturas() {
    const $tbody = $('#tab-faturas')
    const $total = $('#lbl-faturas-total')
    $tbody.empty()

    if (!cacheFaturas.length) {
        $tbody.append(
            '<tr><td colspan="6"><div class="cv-empty">Nenhuma fatura de licença encontrada</div></td></tr>'
        )
        $total.text('')
        return
    }

    const lista = filtrarFaturas(cacheFaturas)
    const geral = cacheFaturas.length

    if (!lista.length) {
        $tbody.append(
            '<tr><td colspan="6"><div class="cv-empty">Nenhuma fatura encontrada com esse filtro</div></td></tr>'
        )
        $total.text('0 de ' + geral)
        return
    }

    if (lista.length === geral) {
        $total.text(geral + (geral === 1 ? ' fatura' : ' faturas'))
    } else {
        $total.text(lista.length + ' de ' + geral + ' faturas')
    }

    const ordenada = lista.slice().sort(function (a, b) {
        const da = new Date(a.created_at || a.vencimento_em || 0).getTime() || 0
        const db = new Date(b.created_at || b.vencimento_em || 0).getTime() || 0
        return db - da
    })

    ordenada.forEach(function (f) {
        $tbody.append(
            '<tr>' +
            '<td data-label="Referência"><strong>' + escHtmlFat(f.referencia || ('#' + f.id)) + '</strong></td>' +
            '<td data-label="Tipo">' + escHtmlFat(labelTipoFat(f.tipo)) + '</td>' +
            '<td data-label="Status">' + badgeStatusFat(f.status) + '</td>' +
            '<td data-label="Valor">' + escHtmlFat(formatMoneyFat(f.valor_total)) + '</td>' +
            '<td data-label="Vencimento">' + escHtmlFat(formatDataFat(f.vencimento_em)) + '</td>' +
            '<td data-label="Observação">' + escHtmlFat(f.observacao || '—') + '</td>' +
            '</tr>'
        )
    })
}

function carregarRelatorioFaturas() {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    const $tbody = $('#tab-faturas')

    if (!idFranqueado) {
        $tbody.html(
            '<tr><td colspan="6"><div class="cv-empty">Sessão inválida. Faça login novamente.</div></td></tr>'
        )
        return
    }

    $tbody.html(
        '<tr><td colspan="6"><div class="cv-detail-loading">' +
        '<i class="bi bi-arrow-repeat"></i> Carregando faturas…</div></td></tr>'
    )

    $.ajax({
        url: '/cvLicencaFaturas',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: idFranqueado, status: '' })
    }).done(function (r) {
        cacheFaturas = normalizarListaFat(r).filter(ehFaturaConfVision)
        atualizarResumoFat(cacheFaturas)
        renderTabelaFaturas()
    }).fail(function (xhr) {
        console.error('Erro ao carregar faturas', xhr)
        $tbody.html(
            '<tr><td colspan="6"><div class="cv-empty">Erro ao carregar faturas</div></td></tr>'
        )
    })
}

$(window).on('load', function () {
    sessionStorage.setItem('idFat', '0')
    initFiltros()
    carregarRepresentantes()
    faturaListar()
    $('#btnFiltrar').on('click', faturaListar)
    $('#btnReceber').on('click', receberFatura)
    $('#btnLimparReceber').on('click', limparFecharFatura)
    $('#filtroAno, #filtroMes, #filtroCliente, #filtroStatus').on('change', faturaListar)
    $('#filtroVencidas').on('change', faturaListar)
})

function initFiltros() {
    const agora = new Date()
    const anoAtual = agora.getFullYear()
    const mesAtual = agora.getMonth() + 1

    const $ano = $('#filtroAno')
    $ano.empty()
    for (let a = anoAtual - 3; a <= anoAtual + 1; a++) {
        $ano.append(`<option value="${a}">${a}</option>`)
    }
    $ano.val(String(anoAtual))
    $('#filtroMes').val(String(mesAtual))
}

function carregarRepresentantes() {
    const payload = payloadTenant({})
    if (!payload) return

    $.ajax({
        url: '/representante/listar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload),
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status !== 'OK' || !r.dados) return
        const $sel = $('#filtroCliente')
        $sel.find('option:not(:first)').remove()
        r.dados
            .filter((i) => !i.dataCancelamento)
            .sort((a, b) => (a.razaoSocial || '').localeCompare(b.razaoSocial || ''))
            .forEach((i) => {
                $sel.append(`<option value="${i.idRepresentante}">${i.razaoSocial}</option>`)
            })
    })
}

function faturaListar() {
    const payload = payloadTenant({
        idOrigem: 'CENTRAL',
        ano: parseInt($('#filtroAno').val(), 10) || 0,
        mes: parseInt($('#filtroMes').val(), 10) || 0,
        idDestino: $('#filtroCliente').val() || '',
        filtroStatus: $('#filtroStatus').val() || '',
        somenteVencidas: $('#filtroVencidas').is(':checked'),
        carregarItens: false,
    })
    if (!payload) return

    $.ajax({
        url: '/financeiro/faturaListarByCentral',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload),
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar faturas')
    }).done(function (r) {
        const $tbody = $('.tech-col-list tbody')
        $tbody.empty()
        limparFecharFatura()

        if (r.status === 'OK' && Array.isArray(r.dados)) {
            r.dados.forEach((item) => {
                const vencida = isVencida(item.dataVencimento) && item.status !== 'PAGO'
                const rowClass = vencida ? 'tech-row-off' : 'tech-row-on'
                $tbody.append(`
                    <tr class="${rowClass}">
                        <td class="text-sm">${item.nomeDestino || ''}</td>
                        <td class="text-sm">${item.dataVencimento || ''}</td>
                        <td class="text-sm">R$ ${item.valor || '0,00'}</td>
                        <td class="text-sm">${item.status || ''}${vencida ? ' · VENCIDA' : ''}</td>
                        <td class="text-sm btn tech-action-edit" title="Visualizar / receber"
                            onclick="carregaDadosFatura('${item.idFatura}')">
                            <i class="bi bi-eye-fill"></i>
                        </td>
                    </tr>
                `)
            })
        } else if (r.status && r.status !== 'Vazio' && r.status !== 'OK') {
            boxErro(r.status)
        }
    })
}

function isVencida(dataBr) {
    if (!dataBr || dataBr.length < 10) return false
    const p = dataBr.split('/')
    if (p.length !== 3) return false
    const dt = new Date(parseInt(p[2], 10), parseInt(p[1], 10) - 1, parseInt(p[0], 10))
    const hoje = new Date()
    hoje.setHours(0, 0, 0, 0)
    return dt < hoje
}

function limparFecharFatura() {
    $('#nomeFranqueado').val('')
    $('#vencimento').val('')
    $('#valorFat').val('')
    $('#tipoPagamento').val('PIX')
    $('#valorPagamento').val('')
    $('#tabelaItens tbody').empty()
    sessionStorage.setItem('idFat', '0')
}

function carregaDadosFatura(idFatura) {
    if (!idFatura) return

    $.ajax({
        url: '/financeiro/faturaGetDadosById',
        method: 'POST',
        data: JSON.stringify({ idFatura: idFatura }),
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao carregar fatura')
    }).done(function (r) {
        if (r.status !== 'OK' || !r.dados) {
            if (r.status && r.status !== 'OK') boxErro(r.status)
            return
        }

        const item = r.dados
        sessionStorage.setItem('idFat', item.idFatura || '0')
        $('#nomeFranqueado').val(item.nomeDestino || '')
        $('#vencimento').val(item.dataVencimento || '')
        $('#valorFat').val(item.valor || '0,00')
        $('#valorPagamento').val(item.valor || '')

        const $itens = $('#tabelaItens tbody')
        $itens.empty()
        if (Array.isArray(item.itens) && item.itens.length > 0) {
            item.itens.forEach((i) => {
                const descr = (i.descricao && i.descricao.length > 40)
                    ? i.descricao.substring(0, 40) + '...'
                    : (i.descricao || '')
                $itens.append(`
                    <tr>
                        <td class="text-sm">${i.quantidade || ''}</td>
                        <td class="text-sm" title="${i.descricao || ''}">${descr}</td>
                        <td class="text-sm">R$ ${i.credito || '0,00'}</td>
                        <td class="text-sm">R$ ${i.debito || '0,00'}</td>
                    </tr>
                `)
            })
        } else {
            $itens.append(`
                <tr>
                    <td colspan="4" class="text-sm py-3">Sem itens nesta fatura</td>
                </tr>
            `)
        }

        if (item.status === 'PAGO') {
            boxErro('Esta fatura já está paga')
        }
    })
}

function parseMoneyBr(val) {
    if (val == null) return NaN
    let s = String(val).replace(/R\$/g, '').trim()
    if (!s) return NaN
    if (s.indexOf(',') >= 0) {
        s = s.replace(/\./g, '').replace(',', '.')
    }
    return parseFloat(s)
}

function receberFatura() {
    const idFat = sessionStorage.getItem('idFat')
    if (!idFat || idFat === '0') {
        boxErro('Selecione uma fatura na listagem')
        return
    }

    const valorFat = parseMoneyBr($('#valorFat').val())
    const valorPagamento = parseMoneyBr($('#valorPagamento').val())
    const meio = $('#tipoPagamento').val()

    if (!meio) {
        boxErro('Informe o meio de pagamento')
        return
    }
    if (isNaN(valorPagamento) || valorPagamento <= 0) {
        boxErro('Informe um valor de pagamento válido')
        return
    }
    if (!isNaN(valorFat) && valorFat > 0 && valorPagamento + 0.009 < valorFat) {
        boxErro(`Valor pago (R$ ${$('#valorPagamento').val()}) menor que o valor da fatura (R$ ${$('#valorFat').val()})`)
        return
    }

    if (!confirm(`Confirmar recebimento de R$ ${$('#valorPagamento').val()}?`)) {
        return
    }

    $.ajax({
        url: '/financeiro/recebeFaturaById',
        method: 'POST',
        data: JSON.stringify({
            idFatura: idFat,
            valorPagamento: $('#valorPagamento').val(),
            meioPagamento: meio,
        }),
    }).fail(function (e) {
        console.log(e)
        let msg = 'Falha ao receber fatura'
        try {
            const j = JSON.parse(e.responseText)
            if (j && j.status) msg = j.status
        } catch (_) {}
        boxErro(msg)
    }).done(function (r) {
        if (r.status === 'OK') {
            limparFecharFatura()
            boxSucesso('Fatura recebida com sucesso')
            faturaListar()
        } else {
            boxErro(r.status || 'Não foi possível receber a fatura')
        }
    })
}

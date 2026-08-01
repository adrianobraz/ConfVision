$(window).on('load', function () {
    carregarRepresentantes()
    listar()
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
    $('#filtroRepLista').on('change', listar)
})

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
        const opts = r.dados
            .filter((i) => !i.dataCancelamento)
            .sort((a, b) => (a.razaoSocial || '').localeCompare(b.razaoSocial || ''))
        const $form = $('#idVinculo')
        const $filtro = $('#filtroRepLista')
        opts.forEach((i) => {
            const o = `<option value="${i.idRepresentante}">${i.razaoSocial}</option>`
            $form.append(o)
            $filtro.append(o)
        })
    })
}

function listar() {
    const idVinculo = $('#filtroRepLista').val() || ''
    const payload = payloadTenant({ idVinculo: idVinculo })
    if (!payload) return

    $.ajax({
        url: '/financeiro/lancamento/listar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload),
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar lançamentos')
    }).done(function (r) {
        const $tbody = $('.tech-col-list tbody')
        $tbody.empty()
        if (r.status === 'OK' && Array.isArray(r.dados)) {
            r.dados.forEach((item) => {
                const valor = (item.debito && item.debito !== '0,00' && item.debito !== '0.00')
                    ? `D R$ ${item.debito}`
                    : `C R$ ${item.credito}`
                $tbody.append(`
                    <tr class="tech-row-on">
                        <td class="text-sm">${item.nomeVinculo || ''}</td>
                        <td class="text-sm" title="${item.dadoOperacao || ''}">${item.dadoOperacao || ''}</td>
                        <td class="text-sm">${valor}</td>
                        <td class="text-sm">${item.dataOperacao || ''}</td>
                        <td class="text-sm btn tech-action-del" title="Excluir"
                            onclick="deletar('${item.idTarifacao}')">
                            <i class="bi bi-eraser-fill"></i>
                        </td>
                    </tr>
                `)
            })
        } else if (r.status && r.status !== 'Vazio' && r.status !== 'OK') {
            boxErro(r.status)
        }
    })
}

function limpar() {
    $('#idVinculo').val('')
    $('#dadoOperacao').val('')
    $('#tipo').val('0')
    $('#valor').val('')
}

function gravar() {
    const idVinculo = $('#idVinculo').val()
    const dadoOperacao = ($('#dadoOperacao').val() || '').trim()
    const tipo = $('#tipo').val()
    const valor = ($('#valor').val() || '').trim()

    if (!idVinculo) {
        boxErro('Selecione o representante')
        return
    }
    if (!dadoOperacao) {
        boxErro('Informe a descrição')
        return
    }
    if (!valor) {
        boxErro('Informe o valor')
        return
    }

    $.ajax({
        url: '/financeiro/lancamento/inserir',
        method: 'POST',
        data: JSON.stringify({
            idVinculo: idVinculo,
            dadoOperacao: dadoOperacao,
            tipo: tipo,
            valor: valor,
        }),
    }).fail(function (e) {
        let msg = 'Falha ao gravar lançamento'
        try {
            const j = JSON.parse(e.responseText)
            if (j && j.status) msg = j.status
        } catch (_) {}
        boxErro(msg)
    }).done(function (r) {
        if (r.status === 'OK') {
            boxSucesso('Lançamento gravado')
            limpar()
            listar()
        } else {
            boxErro(r.status || 'Não foi possível gravar')
        }
    })
}

function deletar(idTarifacao) {
    if (!idTarifacao) return
    if (!confirm('Excluir este lançamento pendente?')) return
    $.ajax({
        url: '/financeiro/lancamento/deletar',
        method: 'POST',
        data: JSON.stringify({ idTarifacao: idTarifacao }),
    }).fail(function (e) {
        let msg = 'Falha ao excluir'
        try {
            const j = JSON.parse(e.responseText)
            if (j && j.status) msg = j.status
        } catch (_) {}
        boxErro(msg)
    }).done(function (r) {
        if (r.status === 'OK') {
            boxSucesso('Lançamento excluído')
            listar()
        } else {
            boxErro(r.status || 'Não foi possível excluir')
        }
    })
}

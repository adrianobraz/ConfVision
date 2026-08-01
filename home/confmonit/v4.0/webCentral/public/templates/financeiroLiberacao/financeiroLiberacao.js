$(window).on('load', function () {
    listarBloqueados()
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
})

function limpar() {
    $('#idAlvoBloqueio').val('')
    $('#dias').val('3')
}

function listarBloqueados() {
    const payload = payloadTenant({})
    if (!payload) return

    $.ajax({
        url: '/financeiro/liberacao/listaBloqueados',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload),
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar bloqueados')
    }).done(function (r) {
        const $tbody = $('.tech-col-list tbody')
        const $sel = $('#idAlvoBloqueio')
        $tbody.empty()
        $sel.find('option:not(:first)').remove()

        if (r.status === 'OK' && Array.isArray(r.dados)) {
            r.dados.forEach((i) => {
                $sel.append(`<option value="${i.idRepresentante}">${i.razaoSocial}</option>`)
                $tbody.append(`
                    <tr class="tech-row-off">
                        <td class="text-sm">${i.razaoSocial || ''}</td>
                        <td class="text-sm">${i.cnpj || ''}</td>
                        <td class="text-sm btn tech-action-edit" title="Selecionar"
                            onclick="selecionar('${i.idRepresentante}')">
                            <i class="bi bi-pencil-square"></i>
                        </td>
                    </tr>
                `)
            })
        } else if (r.status && r.status !== 'Vazio' && r.status !== 'OK') {
            boxErro(r.status)
        }
    })
}

function selecionar(id) {
    $('#idAlvoBloqueio').val(id)
}

function gravar() {
    const idAlvoBloqueio = $('#idAlvoBloqueio').val()
    const dias = parseInt($('#dias').val(), 10)

    if (!idAlvoBloqueio) {
        boxErro('Selecione o representante')
        return
    }
    if (![3, 5, 10, 30].includes(dias)) {
        boxErro('Informe um prazo válido')
        return
    }
    if (!confirm(`Liberar provisoriamente por ${dias} dias?`)) {
        return
    }

    $.ajax({
        url: '/financeiro/liberacao/gravar',
        method: 'POST',
        data: JSON.stringify({
            idAlvoBloqueio: idAlvoBloqueio,
            dias: dias,
        }),
    }).fail(function (e) {
        let msg = 'Falha ao liberar'
        try {
            const j = JSON.parse(e.responseText)
            if (j && j.status) msg = j.status
        } catch (_) {}
        boxErro(msg)
    }).done(function (r) {
        if (r.status === 'OK') {
            const ate = (r.dados && r.dados.dataRetirada) ? r.dados.dataRetirada : ''
            boxSucesso(ate ? `Liberado até ${ate}` : 'Liberação provisória aplicada')
            limpar()
            listarBloqueados()
        } else {
            boxErro(r.status || 'Não foi possível liberar')
        }
    })
}

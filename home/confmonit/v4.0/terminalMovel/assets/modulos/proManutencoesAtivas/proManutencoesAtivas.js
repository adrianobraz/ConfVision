function proManutencoesAtivas_start() {
    $('#boxDir').empty()
    const uri = '/assets/modulos/proManutencoesAtivas/proManutencoesAtivas.html'
    $('#boxDir').load(uri, function () {
        proManutencoesAtivas_carregarTabela()
        $('#proManutencoesAtivas_btnFechar').on('click', proManutencoesAtivas_btnFechar)
    })
}

function proManutencoesAtivas_btnFechar() {
    if (typeof terminalMovel_mostrarProcessoPrincipal === 'function') {
        terminalMovel_mostrarProcessoPrincipal()
    } else {
        $('#boxDir').empty()
    }
}

function proManutencoesAtivas_carregarTabela() {
    const idFranqueado = sessionStorage.getItem('proAtendimento_idFranqueado') || 'TODOS'

    $.ajax({
        url: '/proAtendimento/listarManutencoes',
        method: 'Post',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idFranqueado: idFranqueado
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const $tbody = $('#proManutencoesAtivas_tbody')
        $tbody.empty()

        if (r.status !== 'Vazio' && r.dados) {
            r.dados.forEach(item => {
                proManutencoesAtivas_montaLinha(item)
            })

            $('button[tipo=proManutencoesAtivas_btnRemover]').off('click').on('click', function () {
                const idAlvo = $(this).attr('idAlvo')
                const titulo = $(this).attr('titulo')
                proManutencoesAtivas_remover(idAlvo, titulo)
            })
        } else {
            $tbody.append(`
                <tr>
                    <td colspan="3" class="text-center">Sem manutenções ativas</td>
                </tr>
            `)
        }
    })
}

function proManutencoesAtivas_montaLinha(item) {
    const tipo = item.tipo || ''
    const cliente = item.nomeCliente || ''
    const dispositivo = item.nomeDispositivo || ''
    const setor = item.nomeSetor || '-'
    const dataBloqueio = item.dataBloqueio || ''
    const dataRetirada = item.dataRetirada || ''
    const titulo = (tipo === 'SETOR')
        ? `${dispositivo} / ${setor}`
        : `${dispositivo}`

    $('#proManutencoesAtivas_tbody').append(`
        <tr>
            <td class="proManutencoesAtivas_alvo">
                <span class="proManutencoesAtivas_alvoTitulo">${tipo}</span>
                <span class="proManutencoesAtivas_alvoInfo">${cliente}</span>
                <span class="proManutencoesAtivas_alvoInfo">${dispositivo}</span>
                <span class="proManutencoesAtivas_alvoInfo">Setor: ${setor}</span>
            </td>
            <td class="proManutencoesAtivas_periodo">
                <span class="d-block">Início: ${dataBloqueio}</span>
                <span class="d-block">Fim: ${dataRetirada}</span>
            </td>
            <td>
                <button
                    type="button"
                    class="btn btn-danger proManutencoesAtivas_btnRemover"
                    tipo="proManutencoesAtivas_btnRemover"
                    idAlvo="${item.idAlvo}"
                    titulo="${titulo}"
                >Remover</button>
            </td>
        </tr>
    `)
}

function proManutencoesAtivas_remover(idAlvo, titulo) {
    if (!idAlvo) return

    const podeRemover = confirm(`Remover manutenção de: ${titulo}?`)
    if (!podeRemover) return

    $.ajax({
        url: '/proAtendimento/removerManutencao',
        method: 'Post',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idAlvo: idAlvo
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const status = String((r && r.status) || '').trim().toUpperCase()
        if (status === 'OK') {
            msgSucesso()
            proManutencoesAtivas_carregarTabela()
        } else {
            msgErro(r.status || 'Erro ao remover manutenção')
        }
    })
}

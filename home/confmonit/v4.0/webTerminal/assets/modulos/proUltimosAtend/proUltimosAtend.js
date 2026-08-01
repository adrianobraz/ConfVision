function proUltimosAtend_start() {
    $('#boxDir').empty()
    var uri = '/assets/modulos/proUltimosAtend/proUltimosAtend.html'
    $('#boxDir').load(uri, function () {
        proUltimosAtend_carregarLista()
        $('#proUltimosAtend_btnFechar').on('click', proUltimosAtend_btnFechar)
    })
}

function proUltimosAtend_btnFechar() {
    if (typeof proFranqFiltro_start === 'function') {
        proFranqFiltro_start()
    } else {
        $('#boxDir').empty()
    }
}

function proUltimosAtend_carregarLista() {
    var idFranqueado = sessionStorage.getItem('proAtendimento_idFranqueado') || 'TODOS'

    $.ajax({
        url: 'proAtendimento/ultimosFinalizados',
        method: 'Post',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idFranqueado: idFranqueado })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        $('#proUltimosAtend_lista').empty()

        if (r.status === 'Vazio' || !r.dados || !r.dados.length) {
            $('#proUltimosAtend_lista').html('<p class="text-muted">Nenhum atendimento finalizado.</p>')
            return
        }

        r.dados.forEach(function (item) {
            proUltimosAtend_montarItem(item)
        })

        $('.proUltimosAtend-item').off('click').on('click', function () {
            var idProcesso = $(this).data('id-processo')
            if (!idProcesso) return
            try {
                sessionStorage.setItem('ateProDado_modoVisualizar', 'S')
            } catch (e) {}
            if (typeof setupTelaAtendimento === 'function') {
                setupTelaAtendimento(idProcesso)
            }
        })
    })
}

function proUltimosAtend_montarItem(item) {
    var dataFim = item.dataAtenFim || item.DataAtenFim || ''
    var cliNome = item.nomeCliente || item.Nome || ''
    var operNome = item.nomeOperador || item.NomeOperador || ''
    var idProcesso = item.idProcesso || ''

    var el = $('<div class="proUltimosAtend-item click d-flex justify-content-between" data-id-processo="' + idProcesso + '">')
        .append($('<span>').text(dataFim))
        .append($('<span>').text(cliNome))
        .append($('<span>').text(operNome))
    $('#proUltimosAtend_lista').append(el)
}

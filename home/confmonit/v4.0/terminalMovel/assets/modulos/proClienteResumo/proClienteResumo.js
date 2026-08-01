function proClienteResumo_start(idCliente) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/proClienteResumo/proClienteResumo.html'
    $('#boxDir').load(uri, () => {
        $('#proClienteResumo_btnFechar').on('click', function () {
            proFranqFiltro_start()
        })
        proClienteResumo_carregar(idCliente)
    })
}

function proClienteResumo_carregar(idCliente) {
    $.ajax({
        url: 'proAtendimento/dadosCliente',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idCliente: idCliente })
    }).fail(function (e) {
        console.log(e)
        $('#proClienteResumo_franqueado').text('—')
        $('#proClienteResumo_monitoramento').text('—')
        $('#proClienteResumo_clienteNome').text('Erro ao carregar.')
        $('#proClienteResumo_telefones').text('—')
    }).done(function (r) {
        var d = r.dados || {}
        $('#proClienteResumo_franqueado').text(d.franqueadoNome || '—')
        $('#proClienteResumo_monitoramento').text(d.monitoramentoNome || '—')
        $('#proClienteResumo_clienteNome').text(d.clienteNome || '—')
        var tels = [d.telefone1, d.telefone2].filter(Boolean)
        $('#proClienteResumo_telefones').text(tels.length ? tels.join(' / ') : '—')
    })
}

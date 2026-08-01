function ctrClienteBuscar_start() {
    $('#boxDir').empty()
    const uri = '/assets/modulos/ctrClienteBuscar/ctrClienteBuscar.html'
    $('#boxDir').load(uri, () => {
        $('#ctrClienteBuscar_btnFechar').on('click', ctrMudarSenha_btnFechar)
        $('#ctrClienteBuscar_btnFiltrar').on('click', ctrClienteBuscar_btnFiltrar)
        $('input[name=filtro]').on('click', ctrClienteBuscar_selectFiltro)
    })
}

function ctrMudarSenha_btnFechar() {
    proFranqFiltro_start()
}

function ctrClienteBuscar_selectFiltro() {
    let place
    if ($('input[name="filtro"]:checked').val() == 'nome') {
        $('#ctrClienteBuscar_filtroValor').unmask()
        place = 'Entre com o nome do cliente'
    } else if ($('input[name="filtro"]:checked').val() == 'celular') {
        $('#ctrClienteBuscar_filtroValor').mask(gMkCel)
        place = 'Entre com o numero do celular cliente'
    } else {
        $('#ctrClienteBuscar_filtroValor').unmask()
        place = 'Entre com o numero da conta do cliente'
    }

    $('#ctrClienteBuscar_filtroValor').val('').attr('placeholder', place)

}

function ctrClienteBuscar_btnFiltrar() {
    const filtro = $('input[name="filtro"]:checked').val()

    let valor
    if ($('input[name="filtro"]:checked').val() == 'celular') {
        valor = gLimpaDocumento($('#ctrClienteBuscar_filtroValor').val())
    } else {
        valor = $('#ctrClienteBuscar_filtroValor').val()
    }

    if (valor == '') {
        msgErro('Digite um valor para o filtro')
        return
    }

    $.ajax({
        url: 'ctrClienteBuscar/filtrar',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            filtro: filtro,
            valor: valor
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        $('#ctrClienteBuscar tbody').empty()
       
        if (r.status != 'Vazio') {
            if (sessionStorage.getItem('login_userVinculo') == 'CENTRAL') {
                r.dados.map(item => ctrClienteBuscar_montaLinha(item));
            } else {
                r.dados.filter(item => (item.nomeFranq == sessionStorage.getItem('login_userVinculoNome'))).map(item => {
                    ctrClienteBuscar_montaLinha(item)
                })
            }


            $('td[tipo="ctrClienteBuscar_buscarDados"]').on('click', function () {
                const idCliente = $(this).attr('idCliente')
                ctrClienteBuscar_buscarDados(idCliente)
            })
        }

    })
}

function ctrClienteBuscar_montaLinha(item) {

    $('#ctrClienteBuscar tbody').append(`
        <tr>
            <td>${item.conta}</td>
            <td>${item.nome}</td>
            <td>${formatarCelular(item.telefone1)}</td>
            <td>${item.nomeFranq}</td>
            <td 
                class="bg-success click" 
                tipo="ctrClienteBuscar_buscarDados" idCliente="${item.idCliente}">
                <i class="bi bi-check2-circle"></i>
            </td>
        </tr>
    `)
}

function ctrClienteBuscar_buscarDados(idCliente) {
    cliDados_start(idCliente)
}
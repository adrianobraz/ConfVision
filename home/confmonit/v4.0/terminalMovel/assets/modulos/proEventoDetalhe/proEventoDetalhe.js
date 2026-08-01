function proEventoDetalhe_start(idProcesso) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/proEventoDetalhe/proEventoDetalhe.html'
    $('#boxDir').load(uri, () => {
        // Carrega o id do processo em sessao para ser utilizada pelo gTime

        flags.proEventoDetalhe_idProcesso = idProcesso

        $('#proEventoDetalhe_btnFechar').attr('hidden', false)

        $('#proEventoDetalhe_btnFechar').on('click', proEventoDetalhe_btnFechar)

        // Carrega a tabela de processos
        proEventoDetalhe_carregarTabela()
    })
}

function proEventoDetalhe_carregarTabela() {
    if (flags.proEventoDetalhe_idProcesso == 'off') return

    if (document.querySelector('#proEventoDetalhe')) {
        $.ajax({
            url: 'proEventoDetalhe/carregarTabela',
            method: 'Post',
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
                "Authorization": "Bearer " + sessionStorage.getItem('token')
            },
            data: JSON.stringify({ idProcesso: flags.proEventoDetalhe_idProcesso })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            $('#proEventoDetalhe tbody').empty()
            if (r.status != 'Vazio') {
                r.dados.forEach(item => {
                    proEventoDetalhe_montaLinha(item)
                });

                $('[tipo=proEventoDetalhe_btnDesagrupado]').off('click').on('click', proEventoDetalhe_btnDesagrupado)
            }
        })
    }
}

function proEventoDetalhe_montaLinha(item) {
    const tipo = (item.codigo.substring(0, 1) == 1) ? 'E' : 'R'
    const tipoClass = (tipo === 'E') ? 'proEventoDetalhe_tipoE' : 'proEventoDetalhe_tipoR'
    const codigoPart = `<span class="proEventoDetalhe_tipoBadge ${tipoClass}">${tipo}</span> | ${item.codigo.substring(1)} | ${item.particao}`
    const linhaTexto = `${codigoPart} | ${item.zonaUser} | ${item.descricao}`

    $('#proEventoDetalhe tbody').append(`
        <tr class="proEventoDetalhe_dataRow">           
            <td 
                class="fmt-tbody proEventoDetalhe_qtdCell"
            >${item.quantidade}</td>
            
            <td class="fmt-tbody proEventoDetalhe_contentCell" title="${item.descricao}">
                <div class="proEventoDetalhe_infoLine">${linhaTexto}</div>
                <div class="proEventoDetalhe_actionCell">
                    <button
                        type="button"
                        class="btn btn-success proEventoDetalhe_eyeBtn"
                        tipo="proEventoDetalhe_btnDesagrupado"
                        idProcesso="${item.idProcesso}"
                        zonaUser="${item.zonaUser}"
                        codigo="${item.codigo}"
                        title="Visualizar evento"
                    >
                        <i class="bi bi-eye-fill"></i> Visualizar
                    </button>
                </div>
            </td>
        </tr>    
    `)
}

function proEventoDetalhe_btnFechar() {
    flags.proEventoDetalhe_idProcesso = 'off'
    if (typeof terminalMovel_mostrarProcessoPrincipal === 'function') {
        terminalMovel_mostrarProcessoPrincipal()
    } else {
        proFranqFiltro_start()
    }

}

function proEventoDetalhe_btnDesagrupado() {
    const idProcesso = $(this).attr('idProcesso')
    const codigo = $(this).attr('codigo')
    const zonaUser = $(this).attr('zonaUser')
    
    modEvtDesagrupado_start(idProcesso, codigo, zonaUser, proEventoDetalhe_start)
}
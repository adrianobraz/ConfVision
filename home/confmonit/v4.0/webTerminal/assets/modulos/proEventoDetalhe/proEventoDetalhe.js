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

                $('td[tipo=proEventoDetalhe_btnDesagrupado]').on('click', proEventoDetalhe_btnDesagrupado)
            }
        })
    }
}

function proEventoDetalhe_montaLinha(item) {
    // console.log(item)
    const tipo = (item.codigo.substring(0, 1) == 1) ? 'E' : 'R'
    const tipoCor = (item.codigo.substring(0, 1) == 1) ? 'bg-danger' : 'bg-success'

    let descricao 
    
    if (item.descricao.length >30){
        descricao = item.descricao.substring(0,30) + ' ...'
    }else {
        descricao = item.descricao
    }

    $('#proEventoDetalhe tbody').append(`
        <tr>           
            <td 
                style="width: 4%;" 
                class="click bg-success" 
                tipo="proEventoDetalhe_btnDesagrupado" 
                idProcesso="${item.idProcesso}" 
                zonaUser="${item.zonaUser}"
                codigo="${item.codigo}"
            >
                <div class="d-flex justify-content-center text-bg-success">
                    <i class="bi bi-eye-fill"></i>
                </div>
            </td>
            
            <td 
                style="width: 8%;" 
                class="fmt-tbody"
            >${item.quantidade}</td>
            
            <td 
                style="width: 4%;" 
                class="${tipoCor}"
            >${tipo}</td>
            
            <td 
                style="width: 6%;" 
                class="fmt-tbody"
            >${item.codigo.substring(1)}</td>
            
            <td 
                style="width: 6%;" 
                class="fmt-tbody"
            >${item.particao}</td>
            
            <td 
                style="width: 10%;" 
                class="fmt-tbody"
            >${item.zonaUser}</td>
            
            <td 
                style="width: 62%;" 
                class="fmt-tbody" 
                title="${item.descricao}"
            >${descricao}</td>
        </tr>    
    `)
}

function proEventoDetalhe_btnFechar() {
    flags.proEventoDetalhe_idProcesso = 'off'
    proFranqFiltro_start()

}

function proEventoDetalhe_btnDesagrupado() {
    const idProcesso = $(this).attr('idProcesso')
    const codigo = $(this).attr('codigo')
    const zonaUser = $(this).attr('zonaUser')
    
    modEvtDesagrupado_start(idProcesso, codigo, zonaUser, proEventoDetalhe_start)
}
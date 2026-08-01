$(document).ready(function () {
    ajustaTabela()
    $(window).on('resize', function () {
        ajustaTabela()
    })


    // Carrega o campo Cliente
    carregarTabela()
})


function carregarTabela() {
    $.ajax({
        url: 'carregarTabela',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        console.log(r)
        $('#tab-dispositivo tbody').empty()
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                let btnImg = '<i class="bi bi-x-circle-fill"></i>'
                let btnCor = 'btn-danger'

                if (i.ativo == "S") {
                    btnImg = '<i class="bi bi-toggle-on"></i>'
                    btnCor = 'btn-success'
                } else if (i.ativo == "N") {
                    btnImg = '<i class="bi bi-toggle-off"></i>'
                    btnCor = 'btn-danger'
                } else if (i.ativo == "SD") {
                    btnImg = '<i class="bi bi-toggle-off"></i>'
                    btnCor = 'btn-secondary'
                }

                const nomeCliente = (i.nomeCliente.length > 50) ?
                    i.nomeCliente.substring(0, 47) + "..." :
                    i.nomeCliente

                $('#tab-dispositivo tbody').append(`
                        <tr>
                        <td class="text-uppercase">${i.conta}</td>
                        <td class="text-uppercase" title="${i.nomeCliente}">${nomeCliente}</td>
                        <td class="text-uppercase">${i.nome}</td>
                        
                        <td>
                            <button class="btn btn-sm ${btnCor} py-0"
                                id=${i.idDispositivo}
                                tipo="habilitar"
                                status="${i.ativo}"
                                title="Habilita/Desabilita o
                                Dispositivo">
                                ${btnImg}
                            </button>
                        </td>
                    </tr>
                `)

            });
            // Associa os botoes

            $('button[tipo=habilitar]').on('click', function () {
                const id = this.getAttribute("id")
                const status = this.getAttribute("status")
                GerenciarDispositivoHabilitar(id, status)
            })
        }
    })
}

function GerenciarDispositivoHabilitar(id, status) {
    if (status != 'N') {
        $.ajax({
            start: boxProcessando(),
            url: '/GerenciarDispositivoHabilitar',
            method: 'Post',
            data: JSON.stringify({ idDispositivo: id })
        }).fail(function (e) {
            boxErro("Erro ao habilitar/desabilitar Dispositivo")
        }).done(function (r) {
            boxFechar()
            carregarTabela($('#id-cliente').val())
        })
    }
}





var fpPagDisp = null

$(document).ready(function () {
    ajustaTabela()
    $(window).on('resize', function () {
        ajustaTabela()
    })

    $('#pesquisar-cliente').on('input', function () {
        clearTimeout(window._fpBuscaDisp)
        window._fpBuscaDisp = setTimeout(iniciarLista, 400)
    })

    iniciarLista()
})

function iniciarLista() {
    if (fpPagDisp) fpPagDisp.destroy()
    var termo = ($('#pesquisar-cliente').val() || '').trim()

    fpPagDisp = fpScrollPaginacao({
        url: 'carregarTabela',
        tbodySel: '#tab-dispositivo tbody',
        getPayload: function () {
            return {
                idFranqueado: localStorage.getItem('idFranqueado'),
                termo: termo
            }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                var btnImg = '<i class="bi bi-x-circle-fill"></i>'
                var btnCor = 'btn-danger'

                if (i.ativo == 'S') {
                    btnImg = '<i class="bi bi-toggle-on"></i>'
                    btnCor = 'btn-success'
                } else if (i.ativo == 'N') {
                    btnImg = '<i class="bi bi-toggle-off"></i>'
                    btnCor = 'btn-danger'
                } else if (i.ativo == 'SD') {
                    btnImg = '<i class="bi bi-toggle-off"></i>'
                    btnCor = 'btn-secondary'
                }

                var nomeCliente = (i.nomeCliente.length > 50) ?
                    i.nomeCliente.substring(0, 47) + '...' :
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
                                title="Habilita/Desabilita o Dispositivo">
                                ${btnImg}
                            </button>
                        </td>
                    </tr>
                `)
            })

            $('button[tipo=habilitar]').off('click').on('click', function () {
                var id = this.getAttribute('id')
                var status = this.getAttribute('status')
                GerenciarDispositivoHabilitar(id, status)
            })
        },
        onEmpty: function () {
            $('#tab-dispositivo tbody').append(
                '<tr><td colspan="4" class="text-center">LISTA VAZIA</td></tr>'
            )
        },
        onFail: function () {
            boxErro('Erro ao carregar a tabela')
        }
    })

    fpPagDisp.reset()
}

function GerenciarDispositivoHabilitar(id, status) {
    if (status != 'N') {
        $.ajax({
            start: boxProcessando(),
            url: '/GerenciarDispositivoHabilitar',
            method: 'Post',
            data: JSON.stringify({ idDispositivo: id })
        }).fail(function () {
            boxErro('Erro ao habilitar/desabilitar Dispositivo')
        }).done(function () {
            boxFechar()
            iniciarLista()
        })
    }
}

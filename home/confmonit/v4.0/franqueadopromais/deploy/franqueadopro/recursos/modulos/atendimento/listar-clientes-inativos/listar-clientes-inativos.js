var fpPagClientes = null

$(document).ready(function () {
    $('#pesquisar-cliente').on('input', function () {
        clearTimeout(window._fpBuscaCli)
        window._fpBuscaCli = setTimeout(iniciarLista, 400)
    })
    iniciarLista()
})

function iniciarLista() {
    if (fpPagClientes) fpPagClientes.destroy()
    var termo = ($('#pesquisar-cliente').val() || '').trim()

    fpPagClientes = fpScrollPaginacao({
        url: '/ListarClientesInativos',
        tbodySel: 'tbody',
        getPayload: function () {
            return {
                idFranqueado: localStorage.getItem('idFranqueado'),
                ativo: 'N',
                termo: termo
            }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                $('tbody').append(`
                    <tr>
                        <td>${i.nome}</td>
                        <td>${i.documento1}</td>
                        <td>${formatarCelular(i.telefone1)}</td>
                        <td>${formatarCelular(i.telefone2)}</td>
                        <td>${i.email1}</td>
                    </tr>
                `)
            })
        },
        onEmpty: function () {
            $('tbody').append(`<tr><td colspan="5" class="text-center">LISTA VAZIA</td></tr>`)
        },
        onFail: function () {
            boxErro('Erro ao carregar tabela')
        }
    })

    fpPagClientes.reset()
}

var fpPagClientes = null
var fpPagClientesAtivo = ''

$(document).ready(function () {
    $('#btn-imprimir').on('click', function(){
        gerarPdf('RELAÇÃO DE CLIENTES','#tabListaCliente', 'landscape', $('#qtd-eventos').text())
    })
    var debounceBusca
    $('#pesquisar-cliente').on('input', function () {
        clearTimeout(debounceBusca)
        debounceBusca = setTimeout(function () {
            iniciarListaClientes()
        }, 400)
    })
    iniciarListaClientes()
})

function iniciarListaClientes() {
    if (fpPagClientes) fpPagClientes.destroy()
    var termo = ($('#pesquisar-cliente').val() || '').trim()

    fpPagClientes = fpScrollPaginacao({
        url: '/ListarClientesCarregarClientes',
        tbodySel: 'tbody',
        showLoading: true,
        getPayload: function () {
            return {
                idFranqueado: localStorage.getItem('idFranqueado'),
                termo: termo
            }
        },
        renderRows: function (dados, primeira) {
            if (primeira) {
                $('thead').empty()
                renderCabecalhoResumo(dados, fpPagClientes.getTotal())
            }
            dados.forEach(function (i) {
                var ativo = (i.ativo == 'S') ? 'ATIVO' : 'SUSPENSO'
                $('tbody').append(`
                    <tr>
                        <td class="text-start">${i.nome}</td>
                        <td>${formatarCelular(i.telefone1)}</td>
                        <td>${formatarCelular(i.telefone2)}</td>
                        <td class="text-start">${i.email1}</td>
                        <td>${ativo}</td>
                    </tr>
                `)
            })
        },
        onTotal: function (total) {
            if (total != null) renderCabecalhoResumo([], total)
        },
        onEmpty: function () {
            $('thead').empty()
            $('tbody').empty()
            $('thead').append('<tr><td colspan="5" class="text-center">NENHUM REGISTRO</td></tr>')
        }
    })

    fpPagClientes.reset()
}

function renderCabecalhoResumo(dadosPagina, totalGeral) {
    var ativos = 0
    var suspenso = 0
    ;(dadosPagina || []).forEach(function (i) {
        if (i.ativo == 'S') ativos++
        else suspenso++
    })
    var qtd = totalGeral != null ? totalGeral : dadosPagina.length
    $('thead').html(`
        <tr>
            <td colspan="5" class="fs-5 fw-bold fst-italic">
                Total de ${qtd} clientes (rolando carrega mais)
            </td>
        </tr>
        <tr>
            <th style="width: 33%;">Nome</th>
            <th style="width: 12%;">Telefone1</th>
            <th style="width: 12%;">Telefone2</th>
            <th style="width: 33%;">Email</th>
            <th>Status</th>
        </tr>
    `)
}

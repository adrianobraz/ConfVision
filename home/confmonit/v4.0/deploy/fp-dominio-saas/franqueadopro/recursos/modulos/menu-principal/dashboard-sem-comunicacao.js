var fpPagSemCom = null
var fpSemComFaixa = 'todos'

$(document).ready(function () {
    var params = new URLSearchParams(window.location.search)
    fpSemComFaixa = SemComunicacaoResolverFaixa(params)

    $('#fp-page-header-sub').attr('data-keep-sub', '1').text(SemComunicacaoLabelFaixa(fpSemComFaixa))
    $('#dash-faixa-sub').text(SemComunicacaoLabelFaixa(fpSemComFaixa))

    iniciarListaSemComunicacao()
})

function iniciarListaSemComunicacao() {
    if (fpPagSemCom) fpPagSemCom.destroy()

    var url = fpSemComFaixa === 'todos'
        ? '/MenuPrincipalCarregarSemComunicar'
        : '/DashboardListarSemComunicacao'

    fpPagSemCom = fpScrollPaginacao({
        url: url,
        tbodySel: '#tab-sem-comunicacao',
        getPayload: function () {
            return {
                idFranqueado: localStorage.getItem('idFranqueado'),
                faixa: fpSemComFaixa
            }
        },
        renderRows: function (dados) {
            SemComunicacaoAppendLinhas(dados, '#tab-sem-comunicacao')
        },
        onTotal: function (total) {
            $('#qtd-sem-comunicar').text(total != null ? total : 0)
        },
        onEmpty: function () {
            $('#qtd-sem-comunicar').text('0')
            $('#tab-sem-comunicacao').append(
                '<tr><td colspan="7" class="text-center">NENHUM REGISTRO</td></tr>'
            )
        },
        onFail: function (e) {
            console.log(e)
            if (fpSemComFaixa !== 'todos') {
                fpSemComFaixa = 'todos'
                iniciarListaSemComunicacao()
                return
            }
            $('#qtd-sem-comunicar').text('0')
            $('#tab-sem-comunicacao').append(
                '<tr><td colspan="7" class="text-center">NENHUM REGISTRO</td></tr>'
            )
        }
    })

    fpPagSemCom.reset()
}

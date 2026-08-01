var fpPagEventos = null
var fpPagEventosParams = {}

$(document).ready(function () {
    var params = new URLSearchParams(window.location.search)
    var periodo = params.get('periodo')
    var grupo = (params.get('grupo') || '').toUpperCase()
    var dias = parseInt(params.get('dias')) || 7

    if (periodo === 'hoje' || periodo === 'ontem') {
        var datas = EventosDashboardPeriodoDatas(periodo)
        $('#fp-page-header-text').attr('data-keep-title', '1').text(
            'Eventos — ' + (periodo === 'hoje' ? 'Hoje' : 'Ontem')
        )
        $('#fp-page-header-sub').attr('data-keep-sub', '1').text(
            moment(datas.dataInicio).format('DD/MM/YYYY') +
            ' (00:00 às 23:59)'
        )
        DashboardEventosIniciar(datas.dataInicio, datas.dataFim, '')
        return
    }

    if (grupo && GRUPOS_DASHBOARD_EVENTO.indexOf(grupo) !== -1) {
        var datasGrupo = EventosDashboardPeriodo7Dias()
        $('#fp-page-header-text').attr('data-keep-title', '1').text(
            'Eventos — ' + EventosDashboardLabelGrupo(grupo)
        )
        $('#fp-page-header-sub').attr('data-keep-sub', '1').text(
            'Últimos ' + dias + ' dias (' +
            moment(datasGrupo.dataInicio).format('DD/MM/YYYY') + ' a ' +
            moment(datasGrupo.dataFim).format('DD/MM/YYYY') + ')'
        )
        DashboardEventosIniciar(datasGrupo.dataInicio, datasGrupo.dataFim, grupo)
        return
    }

    window.location.href = '/carregar-menu-principal'
})

function DashboardEventosIniciar(dataInicio, dataFim, grupo) {
    if (fpPagEventos) fpPagEventos.destroy()
    fpPagEventosParams = { dataInicio: dataInicio, dataFim: dataFim, grupo: grupo || '' }

    fpPagEventos = fpScrollPaginacao({
        url: '/DashboardListarEventos',
        tbodySel: '#tab-eventos',
        getPayload: function (offset) {
            return {
                idFranqueado: localStorage.getItem('idFranqueado'),
                dataInicio: fpPagEventosParams.dataInicio,
                dataFim: fpPagEventosParams.dataFim,
                grupo: fpPagEventosParams.grupo
            }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                $('#tab-eventos').append(
                    '<tr style="font-size: 11px;">' +
                    '<td>' + (i.dataEntrada || '') + '</td>' +
                    '<td class="text-start">' + (i.dispConta || '') + ' - ' + (i.cliNome || '') + '</td>' +
                    '<td>' + (i.codigo || '') + '</td>' +
                    '<td class="text-start">' + (i.ctiDescricao || '') + '</td>' +
                    '<td>' + (i.zonaUser || '') + '</td>' +
                    '<td class="text-start">' + (i.zonaUserDescricao || '') + '</td>' +
                    '</tr>'
                )
            })
        },
        onEmpty: function () {
            EventosDashboardRenderTabela([], '#tab-eventos', '#qtd-eventos')
        },
        onTotal: function (total) {
            $('#qtd-eventos').text(total != null ? total : fpPagEventos.getOffset())
        },
        onFail: function () {
            EventosDashboardRenderTabela([], '#tab-eventos', '#qtd-eventos')
        }
    })

    fpPagEventos.reset()
}

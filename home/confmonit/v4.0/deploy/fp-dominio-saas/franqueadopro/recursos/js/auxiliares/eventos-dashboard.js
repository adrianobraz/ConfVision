/* Utilitarios compartilhados: dashboard de eventos */

var GRUPOS_DASHBOARD_EVENTO = [
    'ALARME', 'ARME', 'DESARME', 'EMERGENCIA', 'FALHAS', 'GERAL',
    'MEDICO', 'PANICO', 'RESTAURE', 'SETUP', 'TESTE'
]

function EventosDashboardPeriodoDatas(periodo) {
    var m = (periodo === 'ontem') ? moment().subtract(1, 'days') : moment()
    return {
        dataInicio: m.format('YYYY-MM-DD') + ' 00:00:00',
        dataFim: m.format('YYYY-MM-DD') + ' 23:59:59'
    }
}

function EventosDashboardPeriodo7Dias() {
    return {
        dataInicio: moment().subtract(6, 'days').format('YYYY-MM-DD') + ' 00:00:00',
        dataFim: moment().format('YYYY-MM-DD') + ' 23:59:59'
    }
}

function EventosDashboardLabelGrupo(grupo) {
    var labels = {
        ALARME: 'Alarme',
        ARME: 'Arme',
        DESARME: 'Desarme',
        EMERGENCIA: 'Emergência',
        FALHAS: 'Falhas',
        GERAL: 'Geral',
        MEDICO: 'Médico',
        PANICO: 'Pânico',
        RESTAURE: 'Restaure',
        SETUP: 'Setup',
        TESTE: 'Teste'
    }
    return labels[grupo] || grupo
}

function EventosDashboardRenderTabela(dados, tbodySel, qtdSel) {
    var $tbody = $(tbodySel)
    $tbody.empty()

    if (!dados || dados.length === 0) {
        if (qtdSel) $(qtdSel).text('0')
        $tbody.append('<tr><td colspan="6" class="text-center">NENHUM REGISTRO</td></tr>')
        return
    }

    if (qtdSel) $(qtdSel).text(dados.length)

    dados.forEach(function (i) {
        $tbody.append(
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
}

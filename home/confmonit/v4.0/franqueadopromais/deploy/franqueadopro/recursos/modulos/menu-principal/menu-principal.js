$(document).ready(function () {
    MenuPrincipalCarregarDadosLogin()
    MenuPrincipalCarregarInformativo()
    DashboardCarregarContadores()
    DashboardCarregarTotalClientes()
    DashboardCarregarAlarmes()
    DashboardCarregarEventosHoje()
    DashboardCarregarEventosOntem()
    DashboardCarregarEventosGrupo()
    if (typeof fpAplicarLicencaUI === 'function') {
        setTimeout(fpAplicarLicencaUI, 300)
    }
})

function MenuPrincipalCarregarDadosLogin() {
    $('#display-nome-monitoramento').text(localStorage.getItem('nomeFranqueado') || '')
    $('#display-nome-usuario').text(localStorage.getItem('nomeUsuario'))
}

function MenuPrincipalCarregarInformativo() {
    let lista
    buscarLista(l => {
        lista = l
        contador(0, 2)
    })

    function contador(cont, qtd) {
        exibeInformativo(lista[cont])
        setTimeout(() => {
            cont++
            if (cont >= qtd) cont = 0
            contador(cont, qtd)
        }, 60000);
    }

    function exibeInformativo(item) {
        $.ajax({
            async: false,
            url: '/MenuPrincipalCarregarInformativo',
            method: 'Post',
            data: JSON.stringify({ idInformativo: item.idInformativo }),
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == "OK") {
                const d = r.dados
                const mensagem = (d.messagem || '').trim()
                $('#display-informativo').val(mensagem)
                $('#informativo-container').toggleClass('d-none', mensagem === '')
            }
        })
    }

    function buscarLista(next) {
        $.ajax({
            url: 'MenuPrincipalBuscarListaInformativo',
            method: 'Post',
            data: JSON.stringify({ idAlvo: localStorage.getItem('idFranqueado') })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == "OK") {
                next(r.dados)
            }
        })
    }
}

function DashboardCarregarContadores() {
    $.ajax({
        url: '/DashboardContarSemComunicacao',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function () {
        SEM_COM_FAIXAS.forEach(function (f) {
            $('#dash-sc-' + f).text('--')
            $('#dash-sc-card-' + f).addClass('d-none')
        })
    }).done(function (r) {
        if (r.status == 'OK' && r.dados) {
            var d = r.dados
            SEM_COM_FAIXAS.forEach(function (f) {
                var val = d[f] != null ? d[f] : 0
                $('#dash-sc-' + f).text(val)
                $('#dash-sc-card-' + f).toggleClass('d-none', val <= 0)
            })
            DashboardAplicarAlertaKpi()
        }
    })
}

function DashboardAplicarAlertaKpi() {
    $('.fp-dash-kpi[data-faixa]').each(function () {
        if ($(this).attr('data-faixa') === 'todos') return
        var val = parseInt($(this).find('.fp-dash-kpi-val').text()) || 0
        $(this).removeClass('fp-dash-kpi-warn fp-dash-kpi-danger')
        if (val > 0) {
            var faixa = parseInt($(this).attr('data-faixa'))
            if (faixa >= 24) {
                $(this).addClass('fp-dash-kpi-danger')
            } else if (faixa >= 6) {
                $(this).addClass('fp-dash-kpi-warn')
            }
        }
    })
}

function DashboardCarregarTotalClientes() {
    $.ajax({
        url: '/DashboardContarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function () {
        $('#dash-total-clientes').text('--')
    }).done(function (r) {
        var total = (r.status == 'OK' && r.dados) ? r.dados.length : 0
        $('#dash-total-clientes').text(total)
    })
}

function DashboardCarregarAlarmes() {
    DashboardContarAlarmes('/dispositivoListarArmado', 'S', '#dash-alarmes-armados', 'armados')
    DashboardContarAlarmes('/DispositivoListarDesarmado', 'N', '#dash-alarmes-desarmados', 'desarmados')
}

var dashAlarmesResumo = { armados: null, desarmados: null }

function DashboardContarAlarmes(url, estadoArmado, alvo, campo) {
    $.ajax({
        url: url,
        method: 'Post',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado')
        })
    }).fail(function () {
        $(alvo).text('--')
        dashAlarmesResumo[campo] = null
        DashboardAtualizarPctArmado()
    }).done(function (r) {
        var total = 0
        if (r.status != 'Vazio' && r.dados) {
            total = r.dados.filter(function (i) {
                return i.armado == estadoArmado
            }).length
        }
        $(alvo).text(total)
        dashAlarmesResumo[campo] = total
        DashboardAtualizarPctArmado()
    })
}

function DashboardAtualizarPctArmado() {
    var arm = dashAlarmesResumo.armados
    var des = dashAlarmesResumo.desarmados
    if (arm === null || des === null) {
        if (arm === null && des === null) return
        $('#dash-pct-armado').text('--')
        return
    }

    var total = arm + des
    var pct = total > 0 ? Math.round((arm / total) * 100) : 0
    $('#dash-pct-armado').text(pct + '%')
}

function DashboardCarregarEventosHoje() {
    var datas = EventosDashboardPeriodoDatas('hoje')
    $.ajax({
        url: '/DashboardContarEventosPeriodo',
        method: 'Post',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            dataInicio: datas.dataInicio,
            dataFim: datas.dataFim
        })
    }).fail(function () {
        $('#dash-evt-hoje').text('--')
    }).done(function (r) {
        var total = (r.status == 'OK' && r.dados) ? (r.dados.total || 0) : 0
        $('#dash-evt-hoje').text(total)
    })
}

function DashboardCarregarEventosOntem() {
    var datas = EventosDashboardPeriodoDatas('ontem')
    $.ajax({
        url: '/DashboardContarEventosPeriodo',
        method: 'Post',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            dataInicio: datas.dataInicio,
            dataFim: datas.dataFim
        })
    }).fail(function () {
        $('#dash-evt-ontem').text('--')
    }).done(function (r) {
        var total = (r.status == 'OK' && r.dados) ? (r.dados.total || 0) : 0
        $('#dash-evt-ontem').text(total)
    })
}

function DashboardCarregarEventosGrupo() {
    var datas = EventosDashboardPeriodo7Dias()
    $.ajax({
        url: '/DashboardContarEventosGrupo',
        method: 'Post',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            dataInicio: datas.dataInicio,
            dataFim: datas.dataFim
        })
    }).fail(function () {
        GRUPOS_DASHBOARD_EVENTO.forEach(function (g) {
            $('#dash-evt-' + g).text('--')
        })
    }).done(function (r) {
        if (r.status == 'OK' && r.dados) {
            GRUPOS_DASHBOARD_EVENTO.forEach(function (g) {
                var val = r.dados[g] != null ? r.dados[g] : 0
                $('#dash-evt-' + g).text(val)
            })
            DashboardAplicarAlertaEventosGrupo()
        }
    })
}

function DashboardAplicarAlertaEventosGrupo() {
    $('.fp-dash-kpi[data-grupo]').each(function () {
        var val = parseInt($(this).find('.fp-dash-kpi-val').text()) || 0
        var grupo = $(this).attr('data-grupo')
        $(this).removeClass('fp-dash-kpi-warn fp-dash-kpi-danger')
        if (val > 0 && (grupo === 'PANICO' || grupo === 'EMERGENCIA' || grupo === 'ALARME')) {
            $(this).addClass('fp-dash-kpi-danger')
        } else if (val > 0 && grupo === 'FALHAS') {
            $(this).addClass('fp-dash-kpi-warn')
        }
    })
}

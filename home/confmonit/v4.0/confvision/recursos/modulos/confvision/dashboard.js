/**
 * KPIs do home ConfVision (menu-confvision.html).
 */
let homeKpiFiltro = 'total'

function initDashboardInteracao() {
    const $dash = $('#cv-home-dashboard')
    if (!$dash.length || ConfVisionUrls.ehCliente()) return

    $dash.removeClass('d-none')

    $('.cv-dashboard-kpi-btn[data-kpi]').on('click', function () {
        const kpi = String($(this).data('kpi') || 'total')
        homeKpiFiltro = kpi
        $('.cv-dashboard-kpi-btn[data-kpi]').removeClass('is-active')
        $(this).addClass('is-active')
        if (typeof renderHomeClientesCameras === 'function') {
            renderHomeClientesCameras()
        }
    })

    $('.cv-dashboard-kpi-btn[data-kpi="total"]').addClass('is-active')
}

function cameraPassaFiltroKpi(cam) {
    if (!cam || homeKpiFiltro === 'total') return true
    const A = ConfVisionArmado
    switch (homeKpiFiltro) {
        case 'ativas':
            return A.cameraAnaliticoAtiva(cam) && !A.cameraBloqueadaAdmin(cam)
        case 'inativas':
            return !A.cameraAnaliticoAtiva(cam)
        case 'desarmadas':
            return String(cam.dispositivo_armado || cam.armado || '').toUpperCase() === 'N'
        case 'sem_comunicacao':
            return A.cameraSemComunicacao(cam, 15)
        case 'pausadas':
            return A.isAnaliticoPausado(cam)
        case 'bloqueadas':
            return A.cameraBloqueadaAdmin(cam)
        default:
            return true
    }
}

function setKpiNum(kpi, n) {
    const $btn = $('.cv-dashboard-kpi-btn[data-kpi="' + kpi + '"]')
    if (!$btn.length) return
    let $num = $btn.find('.cv-armado-resumo-num')
    if (!$num.length) {
        $btn.find('.cv-kpi-visual').first().html('<span class="cv-armado-resumo-num">0</span>')
        $num = $btn.find('.cv-armado-resumo-num')
    }
    $num.text(String(n))
}

function renderDashboardChart(contagens) {
    const $bars = $('#cv-dashboard-chart-bars')
    if (!$bars.length) return
    const keys = [
        { k: 'ativas', label: 'Ativas', cls: 'cv-bar-on' },
        { k: 'inativas', label: 'Inativas', cls: 'cv-bar-muted' },
        { k: 'sem_comunicacao', label: 'Sem com.', cls: 'cv-bar-warn' },
        { k: 'pausadas', label: 'Pausadas', cls: 'cv-bar-warn' },
        { k: 'bloqueadas', label: 'Bloq.', cls: 'cv-bar-off' }
    ]
    const max = Math.max(1, ...keys.map(function (x) { return contagens[x.k] || 0 }))
    $bars.html(keys.map(function (x) {
        const v = contagens[x.k] || 0
        const h = Math.round((v / max) * 100)
        return (
            '<div class="cv-chart-bar-wrap" title="' + x.label + ': ' + v + '">' +
            '<div class="cv-chart-bar ' + x.cls + '" style="height:' + h + '%"></div>' +
            '<span class="cv-chart-bar-label">' + v + '</span></div>'
        )
    }).join(''))
}

function carregarDashboardResumo() {
    if (ConfVisionUrls.ehCliente()) return
    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!idFranqueado) return

    $.when(
        $.get('/api/cameras?id_franqueado=' + encodeURIComponent(idFranqueado)),
        $.get('/api/licencas?id_franqueado=' + encodeURIComponent(idFranqueado))
    ).done(function (camResp, licResp) {
        const cameras = ConfVisionArmado.normalizarLista(camResp[0])
        const contagens = {
            ativas: 0,
            inativas: 0,
            desarmadas: 0,
            sem_comunicacao: 0,
            pausadas: 0,
            bloqueadas: 0,
            total: cameras.length
        }
        cameras.forEach(function (cam) {
            const A = ConfVisionArmado
            if (A.cameraBloqueadaAdmin(cam)) contagens.bloqueadas++
            if (A.isAnaliticoPausado(cam)) contagens.pausadas++
            if (A.cameraSemComunicacao(cam, 15)) contagens.sem_comunicacao++
            if (!A.cameraAnaliticoAtiva(cam)) contagens.inativas++
            else contagens.ativas++
            if (String(cam.dispositivo_armado || cam.armado || '').toUpperCase() === 'N') {
                contagens.desarmadas++
            }
        })
        Object.keys(contagens).forEach(function (k) {
            if (k !== 'total') setKpiNum(k, contagens[k])
        })
        setKpiNum('total', contagens.total)
        renderDashboardChart(contagens)

        const licBody = licResp[0]
        if (licBody && licBody.resumo && typeof licBody.resumo.em_uso === 'number') {
            /* resumo licencas disponivel para futuros KPIs */
        }
    }).fail(function () {
        /* home ainda carrega painel armado via menu.js */
    })

    if ($('#cv-kpi-ips-banidos').length && typeof $.get === 'function') {
        $.get('/api/rtmp-bans/franqueado?id_franqueado=' + encodeURIComponent(idFranqueado))
            .done(function (r) {
                const n = (r && r.dados && r.dados.length) || (r && r.total) || 0
                $('#cv-kpi-ips-banidos').text(String(n))
                if (n > 0) {
                    $('#cv-kpi-ips-banidos-ver').removeClass('d-none')
                    $('#cv-kpi-ips-banidos-card').addClass('cv-ip-bans-link-alert')
                }
            })
    }
}

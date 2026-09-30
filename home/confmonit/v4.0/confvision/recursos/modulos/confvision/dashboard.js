/* Dashboard KPI — home ConfVision (franqueado) */
const CV_PING_STALE_MIN = 15
const CV_KPI_RING_SIZE = 44
const CV_KPI_DONUT_SIZE = 52

const CV_DASH_CHART_ITEMS = [
    { key: 'ativas', label: 'Ativas', color: '#34d399', cls: 'cv-dash-bar-ok' },
    { key: 'inativas', label: 'Inativas', color: '#64748b', cls: 'cv-dash-bar-muted' },
    { key: 'sem_comunicacao', label: 'Sem comun.', color: '#f59e0b', cls: 'cv-dash-bar-warn' },
    { key: 'desarmadas', label: 'Desarmadas', color: '#eab308', cls: 'cv-dash-bar-warn' },
    { key: 'pausadas', label: 'Pausadas', color: '#fb923c', cls: 'cv-dash-bar-warn' },
    { key: 'bloqueadas', label: 'Bloqueadas', color: '#f87171', cls: 'cv-dash-bar-off' }
]

let homeResumoCache = null
let homeFiltroKpi = null

function dashPct(val, denom) {
    return denom > 0 ? Math.round((val / denom) * 100) : 0
}

function dashConicGradient(items, map, denom) {
    let acc = 0
    const stops = []
    let sumRaw = 0
    items.forEach(function (item) {
        sumRaw += (map[item.key] || 0) / denom
    })
    const scale = sumRaw > 1 ? 1 / sumRaw : 1
    items.forEach(function (item) {
        const val = map[item.key] || 0
        if (val <= 0) return
        const slice = (val / denom) * 100 * scale
        const from = acc
        acc += slice
        stops.push(item.color + ' ' + from.toFixed(2) + '% ' + acc.toFixed(2) + '%')
    })
    if (acc < 99.9) {
        stops.push('rgba(255,255,255,0.08) ' + acc.toFixed(2) + '% 100%')
    }
    if (!stops.length) {
        return 'conic-gradient(rgba(255,255,255,0.08) 0% 100%)'
    }
    return 'conic-gradient(from -90deg, ' + stops.join(', ') + ')'
}

function dashRingSvg(pct, color, size, stroke) {
    size = size || CV_KPI_RING_SIZE
    stroke = stroke || (size <= 48 ? 3 : 4)
    const r = (size - stroke) / 2
    const cx = size / 2
    const cy = size / 2
    const circ = 2 * Math.PI * r
    const p = Math.min(100, Math.max(0, pct))
    const offset = circ * (1 - p / 100)
    return (
        '<svg class="cv-dash-ring-svg" width="' + size + '" height="' + size + '" viewBox="0 0 ' + size + ' ' + size + '" aria-hidden="true">' +
        '<circle class="cv-dash-ring-track" cx="' + cx + '" cy="' + cy + '" r="' + r + '" fill="none" stroke-width="' + stroke + '"/>' +
        '<circle class="cv-dash-ring-arc" cx="' + cx + '" cy="' + cy + '" r="' + r + '" fill="none" stroke="' + color + '" stroke-width="' + stroke + '" stroke-linecap="round" stroke-dasharray="' + circ.toFixed(2) + '" stroke-dashoffset="' + offset.toFixed(2) + '" transform="rotate(-90 ' + cx + ' ' + cy + ')"/>' +
        '</svg>'
    )
}

function carregarDashboardResumo() {
    if (typeof ConfVisionUrls === 'undefined' || ConfVisionUrls.ehCliente()) return
    const $box = $('#cv-home-dashboard')
    if (!$box.length) return

    $.when(
        $.get('/api/cameras/resumo'),
        $.get('/api/rtmp-bans/franqueado')
    ).done(function (resumoResp, bansResp) {
        const r = resumoResp[0]
        const bans = bansResp[0]
        homeResumoCache = (r && r.dados) || null
        if (homeResumoCache && homeResumoCache.ping_stale_min) {
            window.CV_PING_STALE_MIN = homeResumoCache.ping_stale_min
        }
        renderDashboardResumo(homeResumoCache)
        const totalBans = (bans && bans.total != null)
            ? bans.total
            : ((bans && bans.dados) || []).length
        atualizarCardIpsBanidos(totalBans, '#cv-kpi-ips-banidos-card', '#cv-kpi-ips-banidos', '#cv-kpi-ips-banidos-ver')
    }).fail(function () {
        $('#cv-dashboard-chart-bars').html(
            '<div class="cv-empty">Não foi possível carregar o resumo das câmeras.</div>'
        )
    })
}

function atualizarCardIpsBanidos(total, cardSel, numSel, verSel) {
    const n = total || 0
    $(numSel).text(n)
    const $card = $(cardSel)
    const $ver = $(verSel)
    $card.toggleClass('cv-ip-bans-link-alert', n > 0)
    $ver.toggleClass('d-none', n <= 0)
    if (n > 0) {
        $card.attr('title', 'Ver ' + n + ' IP(s) banido(s) — clique para abrir')
    } else {
        $card.attr('title', 'Nenhum IP banido no momento')
    }
}

function renderDashboardResumo(d) {
    if (!d) return
    const map = {
        total: d.total || 0,
        ativas: d.ativas || 0,
        inativas: d.inativas || 0,
        sem_comunicacao: d.sem_comunicacao || 0,
        desarmadas: d.desarmadas || 0,
        pausadas: d.pausadas || 0,
        bloqueadas: d.bloqueadas || 0
    }

    renderDashboardCards(map)
    renderDashboardChart(map)
    $('#cv-home-dashboard').removeClass('d-none')
}

function renderDashboardCards(map) {
    const items = CV_DASH_CHART_ITEMS
    const total = map.total || 0
    const denom = total > 0 ? total : 1

    items.forEach(function (item) {
        const val = map[item.key] || 0
        const pct = dashPct(val, denom)
        const $btn = $('.cv-dashboard-cards .cv-dashboard-kpi-btn[data-kpi="' + item.key + '"]')
        const $vis = $btn.find('.cv-kpi-visual')
        if (!$vis.length) return
        $vis.html(
            dashRingSvg(pct, item.color, CV_KPI_RING_SIZE, 3) +
            '<span class="cv-kpi-ring-num">' + val + '</span>'
        )
        $btn.attr('title', 'Filtrar: ' + item.label + ' (' + val + ' de ' + total + ' · ' + pct + '%)')
    })

    const gradient = dashConicGradient(items, map, denom)
    const $totalBtn = $('.cv-dashboard-cards .cv-dashboard-kpi-btn[data-kpi="total"]')
    const $totalVis = $totalBtn.find('.cv-kpi-visual')
    if ($totalVis.length) {
        $totalVis.html(
            '<span class="cv-kpi-donut-ring" style="background:' + gradient + '"></span>' +
            '<span class="cv-kpi-donut-hole"><span class="cv-kpi-ring-num">' + total + '</span></span>'
        )
    }
    $totalBtn.attr('title', total > 0
        ? 'Total: ' + total + ' câmeras — clique para listar todas'
        : 'Nenhuma câmera cadastrada')
}

function renderDashboardChart(map) {
    const $bars = $('#cv-dashboard-chart-bars').empty()
    const items = CV_DASH_CHART_ITEMS
    const total = map.total || 0
    const denom = total > 0 ? total : 1

    items.forEach(function (item) {
        const val = map[item.key] || 0
        const pct = dashPct(val, denom)
        $bars.append(
            '<button type="button" class="cv-dash-bar-row cv-dashboard-kpi-btn" data-kpi="' + item.key + '" title="Filtrar: ' + item.label + ' (' + val + ' de ' + total + ')">' +
            '<span class="cv-dash-bar-label">' + item.label + '</span>' +
            '<span class="cv-dash-bar-track"><span class="cv-dash-bar-fill ' + item.cls + '" style="width:' + pct + '%"></span></span>' +
            '<span class="cv-dash-bar-val">' + val + '</span>' +
            '</button>'
        )
    })
}

function initDashboardInteracao() {
    if (ConfVisionUrls.ehCliente()) return

    $(document).on('click', '.cv-dashboard-kpi-btn', function () {
        const kpi = String($(this).data('kpi') || '')
        if (!kpi) return
        if (homeFiltroKpi === kpi) {
            homeFiltroKpi = null
            $('.cv-dashboard-kpi-btn').removeClass('is-active')
        } else {
            homeFiltroKpi = kpi
            $('.cv-dashboard-kpi-btn').removeClass('is-active')
            $(this).addClass('is-active')
        }
        renderHomeClientesCameras()
        const alvo = document.getElementById('box-clientes-cameras')
        if (alvo) {
            alvo.scrollIntoView({ behavior: 'smooth', block: 'start' })
        }
    })

    $('#cv-home-filtros-estado').on('click', '.cv-home-filtro-chip', function () {
        homeFiltroKpi = null
        $('.cv-dashboard-kpi-btn').removeClass('is-active')
    })
}

function cameraPassaFiltroKpi(cam) {
    if (!homeFiltroKpi) return true
    const k = homeFiltroKpi
    const stale = (homeResumoCache && homeResumoCache.ping_stale_min) || CV_PING_STALE_MIN

    if (k === 'total') return true
    if (k === 'ativas') {
        return ConfVisionArmado.cameraAnaliticoAtiva(cam)
    }
    if (k === 'inativas') {
        return !ConfVisionArmado.cameraAnaliticoAtiva(cam)
    }
    if (k === 'bloqueadas') {
        return ConfVisionArmado.cameraBloqueadaAdmin(cam)
    }
    if (k === 'sem_comunicacao') {
        return ConfVisionArmado.cameraSemComunicacao(cam, stale)
    }
    if (k === 'pausadas') {
        return ConfVisionArmado.cameraAnaliticoAtiva(cam) && ConfVisionArmado.isAnaliticoPausado(cam)
    }
    if (k === 'desarmadas') {
        if (!ConfVisionArmado.cameraAnaliticoAtiva(cam)) return false
        const somenteArmado = cam._somente_armado || ConfVisionArmado.isPlanoArmado(cam.plano)
        return somenteArmado && cam._armado === 'N'
    }
    return true
}

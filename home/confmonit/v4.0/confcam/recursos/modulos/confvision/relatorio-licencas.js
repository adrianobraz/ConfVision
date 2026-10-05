let cacheLicencas = []
let cacheCamerasMap = {}

const CV_LIC_PLANOS = {
    online: 'Câmera online',
    sensor_foto: 'Sensor — foto',
    sensor_foto_video: 'Sensor — foto + vídeo',
    sensor: 'Sensor — foto + vídeo',
    analitico_armado_evento: 'Analítico armado — só evento',
    analitico_armado_foto: 'Analítico armado — foto',
    analitico_armado_foto_video: 'Analítico armado — foto + vídeo',
    analitico_armado: 'Analítico armado — foto + vídeo',
    analitico_24h_evento: 'Analítico 24h — só evento',
    analitico_24h_foto: 'Analítico 24h — foto',
    analitico_24h_foto_video: 'Analítico 24h — foto + vídeo',
    analitico_24h: 'Analítico 24h — foto + vídeo',
    gravacao_7d: 'Gravação contínua 7 dias',
    gravacao_15d: 'Gravação contínua 15 dias',
    gravacao_30d: 'Gravação contínua 30 dias',
    gravacao_movimento_7d: 'Gravação por movimento 7 dias',
    gravacao_movimento_15d: 'Gravação por movimento 15 dias',
    gravacao_movimento_30d: 'Gravação por movimento 30 dias',
    gravacao_timelapse_7d: 'Timelapse Inteligente 7 dias',
    gravacao_timelapse_15d: 'Timelapse Inteligente 15 dias',
    gravacao_timelapse_30d: 'Timelapse Inteligente 30 dias'
}

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return

    try {
        confVisionAuthGuard()
        confVisionCarregarCabecalho()
    } catch (e) {
        console.error(e)
    }

    $('#btn-atualizar-licencas').on('click', carregarRelatorioLicencas)
    $('#filtro-status-licencas, #filtro-unidade-licencas').on('change', renderTabelaLicencas)
    $('#busca-licencas').on('input', function () {
        clearTimeout(window._cvBuscaLic)
        window._cvBuscaLic = setTimeout(renderTabelaLicencas, 250)
    })

    carregarRelatorioLicencas()
})

function escHtmlLic(valor) {
    if (valor == null || valor === '') return '—'
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function labelPlanoLic(plano) {
    const p = String(plano || '').trim()
    return CV_LIC_PLANOS[p] || p || '—'
}

function labelUnidadeLic(unidade) {
    const u = String(unidade || 'camera').toLowerCase()
    if (u === 'gravacao') return 'Gravação'
    return 'Câmera'
}

function badgeStatusLic(status) {
    const s = String(status || '').toLowerCase()
    if (s === 'em_uso') {
        return '<span class="cv-badge cv-badge-on">Usada</span>'
    }
    if (s === 'disponivel') {
        return '<span class="cv-badge cv-badge-stream">Não usada</span>'
    }
    if (s === 'pendente') {
        return '<span class="cv-badge cv-badge-warn">Pendente</span>'
    }
    if (s === 'expirada') {
        return '<span class="cv-badge cv-badge-off">Expirada</span>'
    }
    return '<span class="cv-badge cv-badge-off">' + escHtmlLic(status) + '</span>'
}

function formatMoneyLic(v) {
    const n = parseFloat(v)
    if (isNaN(n)) return '—'
    return 'R$ ' + n.toFixed(2).replace('.', ',')
}

function formatDataLic(valor) {
    if (!valor) return '—'
    if (typeof confVisionFormatarData === 'function') {
        const txt = confVisionFormatarData(valor)
        return txt.indexOf(' ') > 0 ? txt.split(' ')[0] : txt
    }
    const d = new Date(valor)
    if (isNaN(d.getTime())) return String(valor)
    return d.toLocaleDateString('pt-BR')
}

function nomeCameraPorLicenca(lic) {
    const idCam = lic && lic.vis_camera_id
    if (idCam == null || idCam === '') return '—'
    const cam = cacheCamerasMap[String(idCam)]
    if (cam && cam.nome) return cam.nome + ' (#' + idCam + ')'
    return '#' + idCam
}

function normalizarListaLic(r) {
    if (!r) return []
    if (Array.isArray(r)) return r
    if (Array.isArray(r.dados)) return r.dados
    return []
}

function atualizarResumoLic(resumo) {
    const r = resumo || {}
    $('#lbl-lic-total').text(r.total != null ? r.total : 0)
    $('#lbl-lic-em-uso').text(r.em_uso != null ? r.em_uso : 0)
    $('#lbl-lic-disponivel').text(r.disponivel != null ? r.disponivel : 0)
    $('#lbl-lic-pendente').text(r.pendente != null ? r.pendente : 0)
    $('#lbl-lic-expirada').text(r.expirada != null ? r.expirada : 0)
}

function filtrarLicencas(lista) {
    const termo = ($('#busca-licencas').val() || '').toLowerCase().trim()
    const status = ($('#filtro-status-licencas').val() || '').trim()
    const unidade = ($('#filtro-unidade-licencas').val() || '').trim()

    return (lista || []).filter(function (lic) {
        if (!lic) return false
        if (status && String(lic.status || '') !== status) return false

        const uni = String(lic.unidade || 'camera').toLowerCase()
        if (unidade && uni !== unidade) return false

        if (!termo) return true

        const plano = labelPlanoLic(lic.plano).toLowerCase()
        const cam = String(nomeCameraPorLicenca(lic)).toLowerCase()
        const id = String(lic.id || '').toLowerCase()
        const obs = String(lic.observacao || '').toLowerCase()

        return plano.indexOf(termo) !== -1
            || cam.indexOf(termo) !== -1
            || id.indexOf(termo) !== -1
            || obs.indexOf(termo) !== -1
    })
}

function renderTabelaLicencas() {
    const $tbody = $('#tab-licencas')
    const $total = $('#lbl-licencas-total')
    $tbody.empty()

    if (!cacheLicencas.length) {
        $tbody.append(
            '<tr><td colspan="7"><div class="cv-empty">Nenhuma licença cadastrada</div></td></tr>'
        )
        $total.text('')
        return
    }

    const lista = filtrarLicencas(cacheLicencas)
    const geral = cacheLicencas.length

    if (!lista.length) {
        $tbody.append(
            '<tr><td colspan="7"><div class="cv-empty">Nenhuma licença encontrada com esse filtro</div></td></tr>'
        )
        $total.text('0 de ' + geral)
        return
    }

    if (lista.length === geral) {
        $total.text(geral + (geral === 1 ? ' licença' : ' licenças'))
    } else {
        $total.text(lista.length + ' de ' + geral + ' licenças')
    }

    const ordenada = lista.slice().sort(function (a, b) {
        const sa = String(a.status || '')
        const sb = String(b.status || '')
        if (sa !== sb) {
            const ordem = { em_uso: 0, disponivel: 1, pendente: 2, expirada: 3 }
            return (ordem[sa] != null ? ordem[sa] : 9) - (ordem[sb] != null ? ordem[sb] : 9)
        }
        return (parseInt(a.id, 10) || 0) - (parseInt(b.id, 10) || 0)
    })

    ordenada.forEach(function (lic) {
        $tbody.append(
            '<tr>' +
            '<td data-label="ID"><strong>#' + escHtmlLic(lic.id) + '</strong></td>' +
            '<td data-label="Plano">' + escHtmlLic(labelPlanoLic(lic.plano)) + '</td>' +
            '<td data-label="Tipo">' + escHtmlLic(labelUnidadeLic(lic.unidade)) + '</td>' +
            '<td data-label="Status">' + badgeStatusLic(lic.status) + '</td>' +
            '<td data-label="Câmera">' + escHtmlLic(nomeCameraPorLicenca(lic)) + '</td>' +
            '<td data-label="Válido até">' + escHtmlLic(formatDataLic(lic.valido_ate)) + '</td>' +
            '<td data-label="Valor">' + escHtmlLic(formatMoneyLic(lic.valor)) + '</td>' +
            '</tr>'
        )
    })
}

function carregarRelatorioLicencas() {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    const $tbody = $('#tab-licencas')

    if (!idFranqueado) {
        $tbody.html(
            '<tr><td colspan="7"><div class="cv-empty">Sessão inválida. Faça login novamente.</div></td></tr>'
        )
        return
    }

    $tbody.html(
        '<tr><td colspan="7"><div class="cv-detail-loading">' +
        '<i class="bi bi-arrow-repeat"></i> Carregando licenças…</div></td></tr>'
    )

    $.when(
        $.get('/api/licencas?id_franqueado=' + encodeURIComponent(idFranqueado)),
        $.get('/api/cameras?id_franqueado=' + encodeURIComponent(idFranqueado)).then(
            null,
            function () { return $.Deferred().resolve({ dados: [] }).promise() }
        )
    ).done(function (licResp, camResp) {
        const rLic = licResp[0]
        const rCam = camResp[0]

        cacheLicencas = normalizarListaLic(rLic)
        cacheCamerasMap = {}
        const cams = normalizarListaLic(rCam)
        cams.forEach(function (cam) {
            if (cam && cam.id != null) cacheCamerasMap[String(cam.id)] = cam
        })

        atualizarResumoLic(rLic && rLic.resumo ? rLic.resumo : {})
        renderTabelaLicencas()
    }).fail(function (xhr) {
        console.error('Erro ao carregar licenças', xhr)
        $tbody.html(
            '<tr><td colspan="7"><div class="cv-empty">Erro ao carregar licenças</div></td></tr>'
        )
    })
}

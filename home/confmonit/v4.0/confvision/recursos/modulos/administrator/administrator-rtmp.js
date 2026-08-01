let cacheFalhas = []
let cacheBans = []
let cacheOnline = []
let cacheCameras = []
let buscaCamerasTimer = null
let relPagina = 0
const REL_PAGE_SIZE = 200

$(document).ready(function () {
    administratorAuthGuard()
    initAdminTabs()
    carregarTudo()
    $('#btn-atualizar-rtmp-falhas').on('click', carregarTudo)
    $('#busca-rtmp-falhas').on('input', function () {
        relPagina = 0
        renderFalhas()
    })
    $('#busca-cameras-admin').on('input', function () {
        clearTimeout(buscaCamerasTimer)
        buscaCamerasTimer = setTimeout(carregarCameras, 350)
    })
    $('#btn-rel-prev').on('click', function () {
        if (relPagina > 0) {
            relPagina--
            renderFalhas()
        }
    })
    $('#btn-rel-next').on('click', function () {
        const totalPag = totalPaginasRelatorio(filtrarFalhas())
        if (relPagina < totalPag - 1) {
            relPagina++
            renderFalhas()
        }
    })
    $('#tab-rtmp-bans').on('click', '.btn-rtmp-unban', function () {
        const ip = $(this).data('ip')
        if (ip) desbanir(ip)
    })
    $('#tab-rtmp-falhas').on('click', '.btn-rtmp-ban', function () {
        const ip = $(this).data('ip')
        if (ip) banir(ip)
    })
    $('#tab-rtmp-online, #tab-rtmp-falhas, #tab-cameras-admin').on('click', '.btn-camera-bloquear', function () {
        const id = $(this).data('camera-id')
        const bloquear = $(this).data('bloquear') === 1 || $(this).data('bloquear') === '1'
        if (id) toggleBloqueioCamera(id, bloquear)
    })
    setInterval(carregarTudo, 15000)
})

function administratorAuthGuard() {
    if (localStorage.getItem('papel') !== 'ADM' || !localStorage.getItem('token')) {
        window.location = '/administrator'
    }
}

function initAdminTabs() {
    const salva = (function () {
        try { return localStorage.getItem('admin_tab') || 'cameras' } catch (e) { return 'cameras' }
    })()
    ativarAdminTab(salva, false)

    $('#admin-tabs').on('click', '[data-admin-tab]', function () {
        ativarAdminTab($(this).data('admin-tab'), true)
    })
}

function ativarAdminTab(tab, persistir) {
    const validas = ['cameras', 'online', 'bans', 'relatorio']
    let t = String(tab || 'cameras')
    if (validas.indexOf(t) < 0) t = 'cameras'

    $('#admin-tabs .nav-link').removeClass('active')
    $('#admin-tabs [data-admin-tab="' + t + '"]').addClass('active')

    $('[data-admin-panel]').each(function () {
        const $p = $(this)
        if ($p.data('admin-panel') === t) {
            $p.removeClass('cv-admin-panel-hidden d-none')
        } else {
            $p.addClass('cv-admin-panel-hidden')
        }
    })

    if (persistir) {
        try { localStorage.setItem('admin_tab', t) } catch (e) { /* ignore */ }
    }
}

function esc(v) {
    return $('<div>').text(v == null ? '' : String(v)).html()
}

function labelCamera(o) {
    const nome = o.nome_camera || (o.camera_id != null ? ('Câmera #' + o.camera_id) : '—')
    const id = o.camera_id != null ? (' #' + o.camera_id) : ''
    if (o.nome_camera && o.camera_id != null) {
        return esc(nome) + ' <span class="text-muted small">' + esc('#' + o.camera_id) + '</span>'
    }
    return esc(nome + id)
}

function badgeBloqueado(bloqueado) {
    if (bloqueado === true || bloqueado === 'true' || bloqueado === 1 || bloqueado === '1') {
        return '<span class="cv-badge cv-badge-off">Bloqueada</span>'
    }
    return '<span class="cv-badge cv-badge-on">Ativa</span>'
}

function botaoBloqueio(o) {
    const id = o.camera_id != null ? o.camera_id : o.id
    if (id == null || id === '') return ''
    const bloqueado = o.bloqueado === true || o.bloqueado === 'true' || o.bloqueado === 1 || o.bloqueado === '1'
    if (bloqueado) {
        return '<button type="button" class="cv-btn-primary btn-sm btn-camera-bloquear" data-camera-id="' +
            esc(id) + '" data-bloquear="0" title="Desbloquear câmera">' +
            '<i class="bi bi-unlock"></i> Desbloquear</button>'
    }
    return '<button type="button" class="cv-btn-ghost btn-sm btn-camera-bloquear" data-camera-id="' +
        esc(id) + '" data-bloquear="1" title="Bloquear câmera (administrativo)">' +
        '<i class="bi bi-slash-circle"></i> Bloquear</button>'
}

function labelPlano(plano) {
    const p = String(plano || '').trim()
    if (!p) return '—'
    return esc(p.replace(/_/g, ' '))
}

function carregarTudo() {
    carregarCameras()
    carregarOnline()
    carregarFalhas()
    carregarBans()
}

function carregarCameras() {
    const q = ($('#busca-cameras-admin').val() || '').trim()
    const params = q ? { q: q } : {}
    $.get('/api/administrator/cameras', params)
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao listar câmeras'
            $('#tab-cameras-admin').html(
                '<tr><td colspan="6"><div class="cv-empty">' + esc(msg) + '</div></td></tr>'
            )
            $('#lbl-cam-total').text('0')
            $('#lbl-cam-bloq').text('0')
        })
        .done(function (r) {
            cacheCameras = (r && r.dados) || []
            $('#lbl-cam-total').text(r && r.total != null ? r.total : cacheCameras.length)
            let bloq = 0
            cacheCameras.forEach(function (c) {
                if (c.bloqueado) bloq++
            })
            $('#lbl-cam-bloq').text(bloq)
            renderCamerasAdmin()
        })
}

function renderCamerasAdmin() {
    const $tb = $('#tab-cameras-admin').empty()
    if (!cacheCameras.length) {
        $tb.append('<tr><td colspan="6"><div class="cv-empty">Nenhuma câmera encontrada.</div></td></tr>')
        return
    }
    cacheCameras.forEach(function (c) {
        const nome = esc(c.nome || ('Câmera #' + c.id))
        $tb.append(
            '<tr class="' + (c.bloqueado ? 'cv-row-bloqueada' : '') + '">' +
            '<td data-label="Câmera"><strong>' + nome + '</strong>' +
            ' <span class="text-muted small">#' + esc(c.id) + '</span></td>' +
            '<td data-label="Franqueado">' + esc(c.nome_franqueado || '—') + '</td>' +
            '<td data-label="Cliente">' + esc(c.nome_cliente || '—') + '</td>' +
            '<td data-label="Licença">' + labelPlano(c.plano) + '</td>' +
            '<td data-label="Bloqueado">' + badgeBloqueado(c.bloqueado) + '</td>' +
            '<td data-label="Ações">' + botaoBloqueio(c) + '</td>' +
            '</tr>'
        )
    })
}

function carregarOnline() {
    $.get('/api/rtmp-online')
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao listar online'
            $('#lbl-rtmp-online').text('0')
            $('#tab-rtmp-online').html(
                '<tr><td colspan="7"><div class="cv-empty">' + esc(msg) + '</div></td></tr>'
            )
        })
        .done(function (r) {
            cacheOnline = (r && r.dados) || []
            $('#lbl-rtmp-online').text(r && r.total != null ? r.total : cacheOnline.length)
            renderOnline(r && r.aviso)
        })
}

function renderOnline(aviso) {
    const $tb = $('#tab-rtmp-online').empty()
    if (aviso) {
        $tb.append('<tr><td colspan="7"><div class="cv-empty">' + esc(aviso) + '</div></td></tr>')
        return
    }
    if (!cacheOnline.length) {
        $tb.append('<tr><td colspan="7"><div class="cv-empty">Nenhuma câmera publicando no momento.</div></td></tr>')
        return
    }
    cacheOnline.forEach(function (o) {
        $tb.append(
            '<tr>' +
            '<td data-label="Câmera"><strong>' + labelCamera(o) + '</strong></td>' +
            '<td data-label="Franqueado">' + esc(o.nome_franqueado || '—') + '</td>' +
            '<td data-label="Cliente">' + esc(o.nome_cliente || '—') + '</td>' +
            '<td data-label="Path"><code>' + esc(o.path_label || o.path || '—') + '</code></td>' +
            '<td data-label="IP"><code>' + esc(o.ip || '—') + '</code></td>' +
            '<td data-label="Status">' + badgeBloqueado(o.bloqueado) + '</td>' +
            '<td data-label="">' + botaoBloqueio(o) + '</td>' +
            '</tr>'
        )
    })
}

function carregarBans() {
    $.get('/api/rtmp-bans')
        .fail(function () {
            $('#tab-rtmp-bans').html(
                '<tr><td colspan="4"><div class="cv-empty">Não foi possível carregar bans.</div></td></tr>'
            )
            $('#lbl-rtmp-bans').text('0')
        })
        .done(function (r) {
            cacheBans = (r && r.dados) || []
            $('#lbl-rtmp-bans').text(cacheBans.length)
            renderBans()
        })
}

function renderBans() {
    const $tb = $('#tab-rtmp-bans').empty()
    if (!cacheBans.length) {
        $tb.append('<tr><td colspan="4"><div class="cv-empty">Nenhum IP banido no momento.</div></td></tr>')
        return
    }
    cacheBans.forEach(function (b) {
        const min = Math.ceil((b.restante_sec || 0) / 60)
        $tb.append(
            '<tr>' +
            '<td data-label="IP"><code>' + esc(b.ip) + '</code></td>' +
            '<td data-label="Motivo">' + esc(b.motivo || '') + (b.manual ? ' <span class="cv-badge">manual</span>' : '') + '</td>' +
            '<td data-label="Restante">' + esc(min) + ' min</td>' +
            '<td data-label="">' +
            '<button type="button" class="cv-btn-ghost btn-rtmp-unban btn-sm" data-ip="' + esc(b.ip) + '">' +
            '<i class="bi bi-unlock"></i> Desbanir</button></td>' +
            '</tr>'
        )
    })
}

function desbanir(ip) {
    $.ajax({
        url: '/api/rtmp-bans/unban',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ ip: ip })
    }).done(carregarBans).fail(function (xhr) {
        boxMesagemAtencaoPersonalizada((xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao desbanir')
    })
}

function banir(ip) {
    CvMsg.confirmar('Banir IP?', 'Banir ' + ip + ' por 1 hora?').then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({
            url: '/api/rtmp-bans/ban',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ ip: ip, motivo: 'manual-ui', ttl_sec: 3600 })
        }).done(carregarBans).fail(function (xhr) {
            boxMesagemAtencaoPersonalizada((xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao banir')
        })
    })
}

function toggleBloqueioCamera(cameraId, bloquear) {
    const acao = bloquear ? 'bloquear' : 'desbloquear'
    CvMsg.confirmar(acao.charAt(0).toUpperCase() + acao.slice(1) + ' câmera?', 'Deseja ' + acao + ' a câmera #' + cameraId + '?').then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({
            url: '/api/cameras/' + encodeURIComponent(cameraId) + '/bloquear',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ bloqueado: !!bloquear })
        }).done(function () {
            carregarTudo()
        }).fail(function (xhr) {
            boxMesagemAtencaoPersonalizada((xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao ' + acao + ' câmera')
        })
    })
}

function carregarFalhas() {
    $.get('/api/rtmp-falhas?limit=500')
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || xhr.statusText || 'Erro ao carregar'
            $('#box-rtmp-aviso').removeClass('d-none').html(
                '<i class="bi bi-exclamation-triangle-fill"></i><div><strong>Guard RTMP indisponível</strong><p>' +
                esc(msg) + '</p></div>'
            )
            $('#tab-rtmp-falhas').html(
                '<tr><td colspan="8"><div class="cv-empty">Não foi possível carregar as falhas.</div></td></tr>'
            )
            atualizarPagerRelatorio(0, 0, 0)
        })
        .done(function (r) {
            cacheFalhas = (r && r.dados) || []
            relPagina = 0
            const aviso = (r && r.aviso) || ''
            if (aviso) {
                $('#box-rtmp-aviso').removeClass('d-none').html(
                    '<i class="bi bi-info-circle-fill"></i><div><strong>Configuração</strong><p>' + esc(aviso) + '</p></div>'
                )
            } else {
                $('#box-rtmp-aviso').addClass('d-none').empty()
            }
            const resumo = (r && r.resumo) || {}
            $('#lbl-rtmp-total').text(resumo.total_eventos != null ? resumo.total_eventos : cacheFalhas.length)
            $('#lbl-rtmp-ips').text((resumo.top_ips || []).length || contarIps(cacheFalhas))
            renderFalhas()
        })
}

function contarIps(lista) {
    const s = {}
    lista.forEach(function (f) {
        if (f.ip) s[f.ip] = 1
    })
    return Object.keys(s).length
}

function filtrarFalhas() {
    const q = ($('#busca-rtmp-falhas').val() || '').trim().toLowerCase()
    if (!q) return cacheFalhas
    return cacheFalhas.filter(function (f) {
        const blob = [
            f.ip, f.path, f.path_label, f.titulo, f.dica, f.motivo_raw, f.motivo_codigo,
            f.nome_camera, f.nome_franqueado, f.nome_cliente, f.camera_id
        ].join(' ').toLowerCase()
        return blob.indexOf(q) >= 0
    })
}

function totalPaginasRelatorio(lista) {
    if (!lista.length) return 1
    return Math.ceil(lista.length / REL_PAGE_SIZE)
}

function atualizarPagerRelatorio(total, inicio, fim) {
    const totalPag = totalPaginasRelatorio(filtrarFalhas())
    if (relPagina >= totalPag) relPagina = Math.max(0, totalPag - 1)

    if (total === 0) {
        $('#lbl-rel-pager').text('Nenhum registro')
        $('#lbl-rel-pagina').text('—')
        $('#btn-rel-prev, #btn-rel-next').prop('disabled', true)
        return
    }

    $('#lbl-rel-pager').text('Mostrando ' + inicio + '–' + fim + ' de ' + total)
    $('#lbl-rel-pagina').text((relPagina + 1) + ' / ' + totalPag)
    $('#btn-rel-prev').prop('disabled', relPagina <= 0)
    $('#btn-rel-next').prop('disabled', relPagina >= totalPag - 1)
}

function renderFalhas() {
    const lista = filtrarFalhas()
    const total = lista.length
    const totalPag = totalPaginasRelatorio(lista)
    if (relPagina >= totalPag) relPagina = Math.max(0, totalPag - 1)

    const inicioIdx = relPagina * REL_PAGE_SIZE
    const fimIdx = Math.min(inicioIdx + REL_PAGE_SIZE, total)
    const pagina = lista.slice(inicioIdx, fimIdx)

    const $tb = $('#tab-rtmp-falhas').empty()
    if (!total) {
        $tb.append('<tr><td colspan="8"><div class="cv-empty">Nenhuma falha RTMP registrada ainda.</div></td></tr>')
        atualizarPagerRelatorio(0, 0, 0)
        return
    }

    atualizarPagerRelatorio(total, inicioIdx + 1, fimIdx)

    pagina.forEach(function (f) {
        const quando = f.ts_ultimo && f.ts_ultimo !== f.ts
            ? esc(f.ts) + ' → ' + esc(f.ts_ultimo)
            : esc(f.ts)
        const grav = f.gravidade === 'error' ? 'cv-badge-off' : (f.gravidade === 'warn' ? 'cv-badge-stream' : '')
        $tb.append(
            '<tr>' +
            '<td data-label="Quando">' + quando + '</td>' +
            '<td data-label="Câmera">' + labelCamera(f) + '</td>' +
            '<td data-label="Franqueado">' + esc(f.nome_franqueado || '—') + '</td>' +
            '<td data-label="Cliente">' + esc(f.nome_cliente || '—') + '</td>' +
            '<td data-label="IP"><code>' + esc(f.ip) + (f.porta ? ':' + esc(f.porta) : '') + '</code></td>' +
            '<td data-label="Problema"><span class="cv-badge ' + grav + '">' + esc(f.titulo || f.motivo_codigo) + '</span>' +
            '<div class="small text-muted mt-1">' + esc(f.motivo_raw || '') + '</div></td>' +
            '<td data-label="Vezes">' + esc(f.vezes || 1) + '</td>' +
            '<td data-label="Ações">' +
            (f.ip ? '<button type="button" class="cv-btn-ghost btn-rtmp-ban btn-sm" data-ip="' + esc(f.ip) + '" title="Banir IP">' +
                '<i class="bi bi-slash-circle"></i></button> ' : '') +
            botaoBloqueio(f) +
            '</td>' +
            '</tr>'
        )
    })
}

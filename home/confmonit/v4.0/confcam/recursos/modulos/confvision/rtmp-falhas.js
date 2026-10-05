let cacheFalhas = []
let cacheBans = []
let cacheOnline = []
let cacheHealth = []

const RTMP_TABS = ['falhas', 'online', 'publish', 'bans']

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()

    let tabSalva = 'falhas'
    try {
        tabSalva = localStorage.getItem('rtmp_tab') || 'falhas'
    } catch (e) { /* ignore */ }
    ativarRtmpTab(tabSalva, false)

    carregarTudo()
    $('#btn-atualizar-rtmp-falhas').on('click', carregarTudo)
    $('#rtmp-tabs').on('click', '[data-rtmp-tab]', function () {
        ativarRtmpTab($(this).attr('data-rtmp-tab'), true)
    })
    $('#cv-rtmp-bans-card').on('click', function (e) {
        e.preventDefault()
        ativarRtmpTab('bans', true)
    })
    $('#busca-rtmp-falhas').on('input', function () {
        renderFalhas()
    })
    $('#tab-rtmp-bans').on('click', '.btn-rtmp-unban', function () {
        const ip = $(this).data('ip')
        if (!ip) return
        desbanir(ip)
    })
    $('#tab-rtmp-falhas').on('click', '.btn-rtmp-ban', function () {
        const ip = $(this).data('ip')
        if (!ip) return
        banir(ip)
    })
    setInterval(carregarTudo, 15000)
})

function ativarRtmpTab(tab, persistir) {
    let t = String(tab || 'falhas')
    if (RTMP_TABS.indexOf(t) < 0) t = 'falhas'

    $('#rtmp-tabs .nav-link').removeClass('active')
    $('#rtmp-tabs [data-rtmp-tab="' + t + '"]').addClass('active')

    $('.cv-rtmp-tab-panel').addClass('d-none')
    $('.cv-rtmp-tab-panel[data-rtmp-panel="' + t + '"]').removeClass('d-none')

    if (persistir) {
        try { localStorage.setItem('rtmp_tab', t) } catch (e) { /* ignore */ }
    }
}

function esc(v) {
    return $('<div>').text(v == null ? '' : String(v)).html()
}

function carregarTudo() {
    carregarOnline()
    carregarHealth()
    carregarFalhas()
    carregarBans()
}

function carregarOnline() {
    $.get('/api/rtmp-online')
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao listar online'
            $('#lbl-rtmp-online').text('0')
            $('#tab-rtmp-online').html(
                '<tr><td colspan="6"><div class="cv-empty">' + esc(msg) + '</div></td></tr>'
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
        $tb.append('<tr><td colspan="6"><div class="cv-empty">' + esc(aviso) + '</div></td></tr>')
        return
    }
    if (!cacheOnline.length) {
        $tb.append('<tr><td colspan="6"><div class="cv-empty">Nenhuma câmera publicando no momento.</div></td></tr>')
        return
    }
    cacheOnline.forEach(function (o) {
        const cam = o.nome_camera || (o.camera_id != null ? ('Câmera #' + o.camera_id) : '—')
        const cliente = o.nome_cliente || '—'
        const tracks = Array.isArray(o.tracks) ? o.tracks.join(', ') : ''
        $tb.append(
            '<tr>' +
            '<td data-label="Câmera"><strong>' + esc(cam) + '</strong>' +
            (o.camera_id != null ? ' <span class="text-muted small">#' + esc(o.camera_id) + '</span>' : '') + '</td>' +
            '<td data-label="Cliente">' + esc(cliente) + '</td>' +
            '<td data-label="Path"><code>' + esc(o.path_label || o.path || '—') + '</code></td>' +
            '<td data-label="IP"><code>' + esc(o.ip || '—') + '</code></td>' +
            '<td data-label="Tracks">' + esc(tracks || '—') + '</td>' +
            '<td data-label="Leitores">' + esc(o.leitores != null ? o.leitores : 0) + '</td>' +
            '</tr>'
        )
    })
}

function carregarHealth() {
    $.get('/api/rtmp-publish-health')
        .fail(function () {
            $('#tab-rtmp-health').html(
                '<tr><td colspan="5"><div class="cv-empty">Histórico indisponível (redeploy do guard necessário).</div></td></tr>'
            )
        })
        .done(function (r) {
            cacheHealth = (r && r.dados) || []
            renderHealth()
        })
}

function renderHealth() {
    const $tb = $('#tab-rtmp-health').empty()
    if (!cacheHealth.length) {
        $tb.append(
            '<tr><td colspan="5"><div class="cv-empty">Nenhum publish registrado ainda. ' +
            'Aparece aqui quando uma câmera publicar com sucesso (sem reiniciar o servidor).</div></td></tr>'
        )
        return
    }
    cacheHealth.forEach(function (h) {
        let badge = 'cv-badge-off'
        let rotulo = 'Offline'
        if (h.status === 'online') {
            badge = 'cv-badge-on'
            rotulo = 'Online'
        } else if (h.status === 'recente_offline') {
            badge = 'cv-badge-stream'
            rotulo = 'Caiu recente'
        }
        const seg = h.segundos_desde_publish
        const quando = h.ultimo_publish_iso
            ? esc(h.ultimo_publish_iso) + (seg != null ? ' (' + esc(seg) + 's)' : '')
            : '—'
        $tb.append(
            '<tr>' +
            '<td data-label="Câmera"><strong>' + esc(h.camera_id != null ? h.camera_id : '—') + '</strong></td>' +
            '<td data-label="Path"><code>' + esc(h.path || '—') + '</code></td>' +
            '<td data-label="IP"><code>' + esc(h.ip || '—') + '</code></td>' +
            '<td data-label="Status"><span class="cv-badge ' + badge + '">' + rotulo + '</span></td>' +
            '<td data-label="Último publish">' + quando + '</td>' +
            '</tr>'
        )
    })
}

function atualizarRtmpBansCard(total) {
    const n = total || 0
    const $card = $('#cv-rtmp-bans-card')
    $card.toggleClass('cv-ip-bans-link-alert', n > 0)
    $('#cv-rtmp-bans-ver').toggleClass('d-none', n <= 0)
    if (n > 0) {
        $card.attr('title', 'Ver ' + n + ' IP(s) banido(s) — clique para abrir')
    } else {
        $card.attr('title', 'Nenhum IP banido no momento')
    }
}

function carregarBans() {
    $.get('/api/rtmp-bans/franqueado')
        .fail(function () {
            $('#tab-rtmp-bans').html(
                '<tr><td colspan="4"><div class="cv-empty">Não foi possível carregar bans (RTMP_GUARD_URL?).</div></td></tr>'
            )
            $('#lbl-rtmp-bans').text('0')
            atualizarRtmpBansCard(0)
        })
        .done(function (r) {
            cacheBans = (r && r.dados) || []
            const total = r && r.total != null ? r.total : cacheBans.length
            $('#lbl-rtmp-bans').text(total)
            atualizarRtmpBansCard(total)
            renderBans()
        })
}

function renderBans() {
    const $tb = $('#tab-rtmp-bans').empty()
    if (!cacheBans.length) {
        $tb.append('<tr><td colspan="4"><div class="cv-empty">Nenhum IP banido nos seus clientes.</div></td></tr>')
        return
    }
    cacheBans.forEach(function (b) {
        const min = Math.ceil((b.restante_sec || 0) / 60)
        const extra = b.nome_camera ? (' — ' + b.nome_camera) : ''
        $tb.append(
            '<tr>' +
            '<td data-label="IP"><code>' + esc(b.ip) + '</code>' +
            (extra ? '<div class="small text-muted">' + esc(b.nome_cliente || '') + extra + '</div>' : '') + '</td>' +
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
    CvMsg.confirmar('Desbloquear IP?', 'Deseja desbloquear o IP ' + ip + '?').then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({
            url: '/api/rtmp-bans/unban',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ ip: ip })
        })
            .done(function () {
                carregarBans()
            })
            .fail(function (xhr) {
                const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao desbanir'
                boxMesagemAtencaoPersonalizada(msg)
            })
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
        })
            .done(function () {
                carregarBans()
            })
            .fail(function (xhr) {
                const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao banir'
                boxMesagemAtencaoPersonalizada(msg)
            })
    })
}

function carregarFalhas() {
    $.get('/api/rtmp-falhas?limit=200')
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || xhr.statusText || 'Erro ao carregar'
            $('#box-rtmp-aviso').removeClass('d-none').html(
                '<i class="bi bi-exclamation-triangle-fill"></i><div><strong>Guard RTMP indisponível</strong><p>' +
                esc(msg) + '</p></div>'
            )
            $('#tab-rtmp-falhas').html(
                '<tr><td colspan="7"><div class="cv-empty">Não foi possível carregar as falhas. Verifique RTMP_GUARD_URL e o serviço confvision-rtmp-guard.</div></td></tr>'
            )
        })
        .done(function (r) {
            cacheFalhas = (r && r.dados) || []
            const aviso = (r && r.aviso) || ''
            if (aviso) {
                $('#box-rtmp-aviso').removeClass('d-none').html(
                    '<i class="bi bi-info-circle-fill"></i><div><strong>Configuração</strong><p>' +
                    esc(aviso) + '</p></div>'
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

function renderFalhas() {
    const q = ($('#busca-rtmp-falhas').val() || '').trim().toLowerCase()
    let lista = cacheFalhas
    if (q) {
        lista = lista.filter(function (f) {
            const blob = [f.ip, f.path, f.path_label, f.titulo, f.dica, f.motivo_raw, f.motivo_codigo].join(' ').toLowerCase()
            return blob.indexOf(q) >= 0
        })
    }

    const $tb = $('#tab-rtmp-falhas').empty()
    if (!lista.length) {
        $tb.append('<tr><td colspan="7"><div class="cv-empty">Nenhuma falha RTMP registrada ainda.</div></td></tr>')
        return
    }

    lista.forEach(function (f) {
        const quando = f.ts_ultimo && f.ts_ultimo !== f.ts
            ? esc(f.ts) + ' → ' + esc(f.ts_ultimo)
            : esc(f.ts)
        const grav = f.gravidade === 'error' ? 'cv-badge-off' : (f.gravidade === 'warn' ? 'cv-badge-stream' : '')
        $tb.append(
            '<tr>' +
            '<td data-label="Quando">' + quando + '</td>' +
            '<td data-label="IP"><code>' + esc(f.ip) + (f.porta ? ':' + esc(f.porta) : '') + '</code></td>' +
            '<td data-label="Path"><code>' + esc(f.path_label || f.path || '—') + '</code></td>' +
            '<td data-label="Problema"><span class="cv-badge ' + grav + '">' + esc(f.titulo || f.motivo_codigo) + '</span>' +
            '<div class="small text-muted mt-1">' + esc(f.motivo_raw || '') + '</div></td>' +
            '<td data-label="Vezes">' + esc(f.vezes || 1) + '</td>' +
            '<td data-label="Dica">' + esc(f.dica || '') + '</td>' +
            '<td data-label=""></td>' +
            '</tr>'
        )
    })
}

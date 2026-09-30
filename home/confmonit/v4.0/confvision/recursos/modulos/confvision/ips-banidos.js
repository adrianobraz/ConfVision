let cacheIpsBanidos = []

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    carregarIpsBanidos()
    $('#btn-atualizar-ips-banidos').on('click', carregarIpsBanidos)
    $('#tab-ips-banidos').on('click', '.btn-rtmp-unban', function () {
        const ip = $(this).data('ip')
        if (ip) desbanirIp(ip)
    })
    setInterval(carregarIpsBanidos, 30000)
})

function esc(v) {
    return $('<div>').text(v == null ? '' : String(v)).html()
}

function carregarIpsBanidos() {
    $.get('/api/rtmp-bans/franqueado')
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao listar IPs banidos'
            $('#tab-ips-banidos').html(
                '<tr><td colspan="7"><div class="cv-empty">' + esc(msg) + '</div></td></tr>'
            )
            $('#lbl-ips-banidos-total').text('0')
        })
        .done(function (r) {
            cacheIpsBanidos = (r && r.dados) || []
            $('#lbl-ips-banidos-total').text(r && r.total != null ? r.total : cacheIpsBanidos.length)
            renderIpsBanidos()
        })
}

function badgeStatusBan(row) {
    if (row.bloqueado) {
        return '<span class="cv-badge cv-badge-off">RTMP bloqueada</span>'
    }
    if (row.ativo === false) {
        return '<span class="cv-badge cv-badge-warn">Inativa</span>'
    }
    return '<span class="cv-badge cv-badge-on">Ativa</span>'
}

function renderIpsBanidos() {
    const $tb = $('#tab-ips-banidos').empty()
    if (!cacheIpsBanidos.length) {
        $tb.append('<tr><td colspan="7"><div class="cv-empty">Nenhum IP banido nos seus clientes no momento.</div></td></tr>')
        return
    }
    cacheIpsBanidos.forEach(function (b) {
        const min = Math.ceil((b.restante_sec || 0) / 60)
        const nomeCam = b.nome_camera || (b.camera_id != null ? ('Câmera #' + b.camera_id) : '—')
        $tb.append(
            '<tr>' +
            '<td data-label="Cliente">' + esc(b.nome_cliente || '—') + '</td>' +
            '<td data-label="Câmera"><strong>' + esc(nomeCam) + '</strong>' +
            (b.camera_id != null ? ' <span class="text-muted small">#' + esc(b.camera_id) + '</span>' : '') + '</td>' +
            '<td data-label="IP"><code>' + esc(b.ip) + '</code></td>' +
            '<td data-label="Status">' + badgeStatusBan(b) + '</td>' +
            '<td data-label="Motivo">' + esc(b.motivo || '') +
            (b.manual ? ' <span class="cv-badge">manual</span>' : '') + '</td>' +
            '<td data-label="Restante">' + esc(min) + ' min</td>' +
            '<td data-label="">' +
            '<button type="button" class="cv-btn-ghost btn-rtmp-unban btn-sm" data-ip="' + esc(b.ip) + '">' +
            '<i class="bi bi-unlock"></i> Desbloquear</button></td>' +
            '</tr>'
        )
    })
}

function desbanirIp(ip) {
    CvMsg.confirmar('Desbloquear IP?', 'Deseja desbloquear o IP ' + ip + '?').then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({
            url: '/api/rtmp-bans/unban',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ ip: ip })
        })
            .done(function () {
                CvMsg.sucesso('IP desbloqueado.')
                carregarIpsBanidos()
            })
            .fail(function (xhr) {
                const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao desbloquear IP'
                CvMsg.aviso(msg)
            })
    })
}

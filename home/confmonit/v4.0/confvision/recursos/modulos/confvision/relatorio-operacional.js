let relOpTab = 'stream'

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    $('#btn-atualizar-relatorio-op').on('click', carregarRelOp)
    $('#rel-op-tabs').on('click', '[data-rel-tab]', function () {
        relOpTab = $(this).attr('data-rel-tab')
        $('#rel-op-tabs .nav-link').removeClass('active')
        $(this).addClass('active')
        carregarRelOp()
    })
    carregarRelOp()
})

function esc(v) {
    return $('<div>').text(v == null ? '' : String(v)).html()
}

function carregarRelOp() {
    const url = '/api/relatorio-operacional/' + relOpTab + '?limit=150'
    const $tb = $('#tab-rel-op-body').html('<tr><td><div class="cv-detail-loading"><i class="bi bi-arrow-repeat"></i> Carregando…</div></td></tr>')
    $.get(url)
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao carregar'
            $tb.html('<tr><td><div class="cv-empty">' + esc(msg) + '</div></td></tr>')
        })
        .done(function (r) {
            const lista = (r && r.dados) || []
            renderRelOp(lista)
        })
}

function renderRelOp(lista) {
    const $tb = $('#tab-rel-op-body').empty()
    if (!lista.length) {
        $tb.append('<tr><td><div class="cv-empty">Nenhum registro ainda. A coleta roda a cada ~30 min após deploy do Go com Postgres.</div></td></tr>')
        return
    }
    if (relOpTab === 'stream') {
        lista.forEach(function (row) {
            $tb.append(
                '<tr><td data-label="Quando">' + esc(row.created_at) +
                '</td><td data-label="Câmera">' + esc(row.vis_camera_id) +
                '</td><td data-label="Problema"><strong>' + esc(row.titulo || row.motivo_codigo) +
                '</strong><div class="small text-muted">' + esc(row.dica || '') + '</div></td></tr>'
            )
        })
        return
    }
    if (relOpTab === 'health') {
        lista.forEach(function (row) {
            $tb.append(
                '<tr><td data-label="Quando">' + esc(row.coletado_em) +
                '</td><td data-label="Componente">' + esc(row.componente) +
                '</td><td data-label="Status">' + esc(row.status) +
                '</td><td data-label="Msg">' + esc(row.mensagem || '') + '</td></tr>'
            )
        })
        return
    }
    lista.forEach(function (row) {
        const preview = row.metricas_json ? JSON.stringify(row.metricas_json).slice(0, 200) : (row.valor_text || '')
        $tb.append(
            '<tr><td data-label="Quando">' + esc(row.coletado_em) +
            '</td><td data-label="Escopo">' + esc(row.escopo) +
            '</td><td data-label="Chave">' + esc(row.chave) +
            '</td><td data-label="Dados"><code class="small">' + esc(preview) + '</code></td></tr>'
        )
    })
}

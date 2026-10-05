/* Administrator — pool infra ConfVision por franqueado */
$(document).ready(function () {
    administratorAuthGuard()
    carregarNodesMtx()
    let buscaTimer = null

    $('#infra-busca-franqueado').on('input', function () {
        clearTimeout(buscaTimer)
        buscaTimer = setTimeout(buscarFranqueados, 300)
    })

    $('#infra-franqueado-sugestoes').on('click', '.cv-infra-sug-item', function () {
        const id = $(this).data('id')
        const label = $(this).data('label') || id
        $('#infra-id-franqueado').val(id)
        $('#infra-franqueado-selecionado').text('Selecionado: ' + label)
        $('#infra-franqueado-sugestoes').empty()
        $('#infra-busca-franqueado').val('')
        carregarPoolStatus()
    })

    $('#btn-infra-carregar').on('click', carregarPoolStatus)
    $('#form-infra-mtx').on('submit', function (e) {
        e.preventDefault()
        salvarMtxPool()
    })
    $('#form-infra-worker').on('submit', function (e) {
        e.preventDefault()
        salvarWorkerPool()
    })

    $('#admin-tabs').on('click', '[data-admin-tab="infra"]', function () {
        carregarNodesMtx()
        if ($('#infra-id-franqueado').val()) {
            carregarPoolStatus()
        }
    })
})

function carregarNodesMtx() {
    $.getJSON('/api/administrator/infra-pool/mediamtx-nodes').done(function (r) {
        const $sel = $('#infra-mtx-node').empty().append('<option value="">Nó MediaMTX…</option>')
        ;(r.dados || []).forEach(function (n) {
            const label = (n.nome || ('#' + n.id)) + ' (max ' + (n.max_cameras || '?') + ')'
            $sel.append($('<option></option>').val(n.id).text(label))
        })
    }).fail(function (xhr) {
        boxMesagemAtencaoPersonalizada(parseErro(xhr))
    })
}

function buscarFranqueados() {
    const q = ($('#infra-busca-franqueado').val() || '').trim()
    if (q.length < 2) {
        $('#infra-franqueado-sugestoes').empty()
        return
    }
    $.getJSON('/api/administrator/infra-pool/franqueados', { q: q }).done(function (r) {
        const $box = $('#infra-franqueado-sugestoes').empty()
        ;(r.dados || []).forEach(function (item) {
            const $btn = $(
                '<button type="button" class="list-group-item list-group-item-action cv-infra-sug-item"></button>'
            )
                .attr('data-id', item.id_franqueado)
                .attr('data-label', item.label || item.nome || '')
            const $main = $('<div class="cv-infra-sug-main"></div>').text(item.label || item.id_franqueado)
            $btn.append($main)
            if (item.hint) {
                $btn.append($('<div class="cv-infra-sug-sub"></div>').text(item.hint))
            }
            $box.append($btn)
        })
        if (!$box.children().length) {
            $box.append('<div class="list-group-item cv-infra-sug-empty">Nenhum resultado</div>')
        }
    }).fail(function (xhr) {
        boxMesagemAtencaoPersonalizada(parseErro(xhr))
    })
}

function carregarPoolStatus() {
    const idFra = ($('#infra-id-franqueado').val() || '').trim()
    if (!idFra) {
        boxMesagemAtencaoPersonalizada('Selecione um franqueado.')
        return
    }
    $.getJSON('/api/administrator/infra-pool/status', { id_franqueado: idFra })
        .done(function (r) {
            const d = r.dados || {}
            renderInfraBadge(d)
            renderMtxTable(d.mediamtx || [])
            renderWorkerTable(d.workers || [])
        })
        .fail(function (xhr) {
            boxMesagemAtencaoPersonalizada(parseErro(xhr))
        })
}

function renderInfraBadge(d) {
    const ok = d.infra_completa === true
    const cls = ok ? 'cv-badge cv-badge-on' : 'cv-badge cv-badge-off'
    let txt = ok ? 'Infra completa (MTX + worker)' : 'Infra incompleta — cadastro de câmera bloqueado'
    if (!ok) {
        const parts = []
        if (!d.tem_mediamtx_pool) parts.push('falta nó MediaMTX no pool')
        if (!d.tem_worker_pool) parts.push('falta worker analítico no pool')
        if (parts.length) {
            txt += ' (' + parts.join('; ') + ')'
        }
    }
    $('#infra-status-badge').html(
        '<div class="cv-form-tab-aviso cv-infra-status-aviso ' +
            (ok ? 'cv-infra-status-ok' : '') +
            '">' +
            '<i class="bi ' +
            (ok ? 'bi-check-circle-fill' : 'bi-exclamation-triangle-fill') +
            '"></i>' +
            '<div><strong>' +
            escHtml(ok ? 'Pronto para cadastrar câmeras' : 'Configure o pool') +
            '</strong><p class="mb-0">' +
            escHtml(txt) +
            '</p></div></div>'
    )
}

function renderMtxTable(rows) {
    const $tb = $('#infra-tbody-mtx').empty()
    if (!rows.length) {
        $tb.append('<tr><td colspan="4"><div class="cv-empty">Nenhum nó no pool — use Salvar acima</div></td></tr>')
        return
    }
    rows.forEach(function (row) {
        const uso =
            (row.uso_franqueado != null ? row.uso_franqueado : 0) +
            ' / global ' +
            (row.uso_global_no != null ? row.uso_global_no : '?')
        $tb.append(
            '<tr><td>' +
                escHtml((row.node_nome || '') + ' #' + row.vis_mediamtx_node_id) +
                '</td><td>' +
                escHtml(String(uso)) +
                '</td><td>' +
                escHtml(String(row.max_cameras || '')) +
                '</td><td>' +
                (row.ativo ? 'Sim' : 'Não') +
                '</td></tr>'
        )
    })
}

function renderWorkerTable(rows) {
    const $tb = $('#infra-tbody-worker').empty()
    if (!rows.length) {
        $tb.append(
            '<tr><td colspan="4"><div class="cv-empty">Nenhum worker no pool — use Salvar acima</div></td></tr>'
        )
        return
    }
    rows.forEach(function (row) {
        $tb.append(
            '<tr><td>' +
                escHtml(row.worker_id || '') +
                '</td><td>' +
                escHtml(String(row.uso_franqueado != null ? row.uso_franqueado : 0)) +
                '</td><td>' +
                escHtml(String(row.max_cameras || '')) +
                '</td><td>' +
                (row.ativo ? 'Sim' : 'Não') +
                '</td></tr>'
        )
    })
}

function salvarMtxPool() {
    const idFra = ($('#infra-id-franqueado').val() || '').trim()
    const nodeId = parseInt($('#infra-mtx-node').val(), 10)
    const max = parseInt($('#infra-mtx-max').val(), 10)
    if (!idFra || !nodeId) {
        boxMesagemAtencaoPersonalizada('Franqueado e nó MediaMTX obrigatórios.')
        return
    }
    const body = { id_franqueado: idFra, vis_mediamtx_node_id: nodeId, ativo: true }
    if (max > 0) body.max_cameras = max
    $.ajax({
        url: '/api/administrator/infra-pool/mediamtx',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(body)
    })
        .done(function () {
            boxSucessoAuto('Nó MediaMTX adicionado ao pool.')
            carregarPoolStatus()
        })
        .fail(function (xhr) {
            boxMesagemAtencaoPersonalizada(parseErro(xhr))
        })
}

function salvarWorkerPool() {
    const idFra = ($('#infra-id-franqueado').val() || '').trim()
    const wid = ($('#infra-worker-id').val() || '').trim()
    const max = parseInt($('#infra-worker-max').val(), 10) || 30
    if (!idFra || !wid) {
        boxMesagemAtencaoPersonalizada('Franqueado e worker_id obrigatórios.')
        return
    }
    $.ajax({
        url: '/api/administrator/infra-pool/worker',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            id_franqueado: idFra,
            worker_id: wid,
            worker_tipo: 'analitico',
            max_cameras: max,
            ativo: true
        })
    })
        .done(function () {
            $('#infra-worker-id').val('')
            boxSucessoAuto('Worker analítico adicionado ao pool.')
            carregarPoolStatus()
        })
        .fail(function (xhr) {
            boxMesagemAtencaoPersonalizada(parseErro(xhr))
        })
}

function parseErro(xhr) {
    try {
        const j = xhr.responseJSON
        if (j && j.status) return j.status
    } catch (e) {
        /* ignore */
    }
    return xhr.statusText || 'Erro na requisição'
}

function escHtml(v) {
    return $('<div>')
        .text(v == null ? '' : String(v))
        .html()
}

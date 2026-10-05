$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    if (!ConfVisionUrls.ehCliente()) {
        $('#link-gerenciar-grupos').removeClass('d-none')
    }
    carregarMosaicos()
})

function escHtml(v) {
    return String(v == null ? '' : v)
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

function erroApi(xhr) {
    try {
        const j = JSON.parse(xhr.responseText || '{}')
        return j.status || j.erro || j.message || xhr.statusText
    } catch (e) {
        return xhr.statusText || 'Erro'
    }
}

function carregarMosaicos() {
    const url = ConfVisionUrls.ehCliente()
        ? '/api/grupos-visualizacao/disponiveis'
        : '/api/grupos-visualizacao'
    $.get(url)
        .done(function (r) {
            renderMosaicos((r && r.dados) || [])
        })
        .fail(function (xhr) {
            $('#mosaicos-lista').html('<div class="cv-empty">' + escHtml(erroApi(xhr)) + '</div>')
        })
}

function renderMosaicos(lista) {
    const $box = $('#mosaicos-lista').empty()
    if (!lista.length) {
        $box.html('<div class="cv-empty">Nenhum mosaico disponível.</div>')
        return
    }
    lista.forEach(function (g) {
        const id = g.id
        const total = g.total_cameras || 0
        const ativo = g.ativo !== false
        $box.append(`
            <a href="/mosaicos/${id}" class="cv-mosaico-card ${ativo ? '' : 'is-inativo'}">
                <div class="cv-mosaico-card-icon"><i class="bi bi-grid-3x3-gap"></i></div>
                <div class="cv-mosaico-card-body">
                    <h3>${escHtml(g.nome || 'Grupo #' + id)}</h3>
                    <p>${escHtml(g.descricao || '')}</p>
                    <span class="cv-mosaico-card-meta">${total} câmera${total === 1 ? '' : 's'}</span>
                </div>
                <i class="bi bi-chevron-right cv-mosaico-card-arrow"></i>
            </a>
        `)
    })
}

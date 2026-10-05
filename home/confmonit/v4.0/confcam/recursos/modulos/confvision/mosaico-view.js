const MOSAICO_PAGE_SIZE = 9
const PING_STALE_MIN = 15

let grupoAtual = null
let camerasTodas = []
let paginaAtual = 0
let mapaClientes = {}
let resizeTimer = null

$(document).ready(function () {
    confVisionAuthGuard()
    const grupoId = ($('#cfg-grupo-id').val() || '').trim()
    if (!grupoId) {
        window.location = '/mosaicos'
        return
    }
    $(window).on('resize.mosaic', function () {
        clearTimeout(resizeTimer)
        resizeTimer = setTimeout(function () {
            if (!camerasTodas.length) return
            if (document.fullscreenElement) return
            const n = camerasDaPagina().length
            if (n && document.getElementById('mosaico-grid') && document.getElementById('mosaico-grid').children.length) {
                aplicarLayoutGrid(n)
                return
            }
            renderPaginaAtual()
        }, 150)
    })
    $('#mosaico-paginas').on('click', '.cv-mosaic-tab', function (e) {
        e.preventDefault()
        const p = parseInt($(this).data('page'), 10)
        if (isNaN(p) || p === paginaAtual) return
        paginaAtual = p
        renderPaginaAtual()
    })
    $('#mosaico-grid').on('dblclick', '.cv-mosaic-frame', function () {
        toggleMosaicFullscreen(this)
    })
    document.addEventListener('fullscreenchange', syncMosaicFullscreenClass)
    carregarClientesNomes(function () {
        carregarGrupo(grupoId)
    })
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

function carregarClientesNomes(cb) {
    if (ConfVisionUrls.ehCliente()) {
        if (typeof cb === 'function') cb()
        return
    }
    $.ajax({ url: '/api/clientes', method: 'POST', contentType: 'application/json', data: '{}' })
        .done(function (r) {
            const lista = (r && r.dados) || r || []
            if (Array.isArray(lista)) {
                lista.forEach(function (c) {
                    const id = String(c.id || c.idCliente || c.id_cliente || '')
                    if (id) mapaClientes[id] = c.nome || c.nick || ('Cliente ' + id)
                })
            }
        })
        .always(function () {
            if (typeof cb === 'function') cb()
        })
}

function nomeCliente(id) {
    id = String(id || '')
    if (mapaClientes[id]) return mapaClientes[id]
    return 'Cliente ' + id
}

function carregarGrupo(id) {
    $.when(
        $.get('/api/grupos-visualizacao/' + encodeURIComponent(id)),
        $.get('/api/grupos-visualizacao/' + encodeURIComponent(id) + '/cameras')
    ).done(function (rGrupo, rCams) {
        grupoAtual = (rGrupo[0] && rGrupo[0].dados) || rGrupo[0]
        camerasTodas = ((rCams[0] && rCams[0].dados) || rCams[0] || []).slice().sort(function (a, b) {
            return (a.ordem || 0) - (b.ordem || 0)
        })
        if (!grupoAtual || !grupoAtual.id) {
            $('#mosaico-nome').text('Grupo não encontrado')
            return
        }
        document.title = (grupoAtual.nome || 'Mosaico') + ' — ConfVision'
        $('#mosaico-nome').text(grupoAtual.nome || 'Mosaico')
        $('#mosaico-meta').text((camerasTodas.length || 0) + ' câmera' + (camerasTodas.length === 1 ? '' : 's'))
        if (!camerasTodas.length) {
            $('#mosaico-empty').removeClass('d-none')
            $('#mosaico-grid').empty()
            $('#mosaico-paginas').empty()
            return
        }
        $('#mosaico-empty').addClass('d-none')
        paginaAtual = 0
        renderTabs()
        renderPaginaAtual()
    }).fail(function (xhr) {
        $('#mosaico-nome').text('Erro')
        $('#mosaico-meta').text(erroApi(xhr))
    })
}

function renderTabs() {
    const total = camerasTodas.length
    const pages = Math.ceil(total / MOSAICO_PAGE_SIZE)
    const $tabs = $('#mosaico-paginas').empty()
    if (pages <= 1) return
    for (let p = 0; p < pages; p++) {
        const de = p * MOSAICO_PAGE_SIZE + 1
        const ate = Math.min((p + 1) * MOSAICO_PAGE_SIZE, total)
        $tabs.append(`
            <button type="button" class="cv-mosaic-tab ${p === paginaAtual ? 'is-active' : ''}" data-page="${p}">
                ${de}–${ate}
            </button>
        `)
    }
}

function camerasDaPagina() {
    const ini = paginaAtual * MOSAICO_PAGE_SIZE
    return camerasTodas.slice(ini, ini + MOSAICO_PAGE_SIZE)
}

function renderPaginaAtual() {
    ConfVisionMosaicCell.stopAll()
    renderTabs()
    const cams = camerasDaPagina()
    const $grid = $('#mosaico-grid').empty()
    cams.forEach(function (cam, idx) {
        const camId = cam.vis_camera_id || cam.camera_id
        const cellId = 'p' + paginaAtual + '-c' + idx + '-' + camId
        const online = cameraOnline(cam)
        const html = `
            <div class="cv-mosaic-cell" data-cell-id="${escHtml(cellId)}">
                <div class="cv-mosaic-frame is-loading" title="Duplo clique para tela cheia">
                    <video class="cv-mosaic-video" muted playsinline autoplay></video>
                    <div class="cv-mosaic-overlay">
                        <div class="cv-mosaic-cli">${escHtml(nomeCliente(cam.id_cliente).toUpperCase())}</div>
                        <div class="cv-mosaic-cam">${escHtml(cam.camera_nome || cam.nome || ('Câmera ' + camId))}</div>
                        <div class="cv-mosaic-status ${online ? 'is-online' : 'is-offline'}">${online ? 'Online' : 'Offline'}</div>
                    </div>
                </div>
            </div>`
        const $cell = $(html)
        $grid.append($cell)
        const frame = $cell.find('.cv-mosaic-frame')[0]
        const video = $cell.find('video')[0]
        ConfVisionMosaicCell.start({
            cellId: cellId,
            cameraId: camId,
            video: video,
            frame: frame
        })
    })
    aplicarLayoutGrid(cams.length)
}

function cameraOnline(cam) {
    const raw = cam.ultimo_ping_em
    if (!raw) return false
    const t = new Date(raw).getTime()
    if (isNaN(t)) return false
    return (Date.now() - t) < PING_STALE_MIN * 60 * 1000
}

function aplicarLayoutGrid(count) {
    const main = document.getElementById('mosaico-main')
    const grid = document.getElementById('mosaico-grid')
    if (!main || !grid || count < 1) return

    const gap = 8
    const aspect = 16 / 9
    const areaW = main.clientWidth - 16
    const areaH = main.clientHeight - 16

    let best = { cols: 1, cellW: 0, cellH: 0, rows: count }
    for (let cols = 1; cols <= count; cols++) {
        const rows = Math.ceil(count / cols)
        const cellW = (areaW - gap * (cols - 1)) / cols
        if (cellW <= 0) continue
        const cellH = cellW / aspect
        const totalH = rows * cellH + gap * (rows - 1)
        if (totalH <= areaH && cellW >= best.cellW) {
            best = { cols: cols, rows: rows, cellW: cellW, cellH: cellH }
        }
    }
    if (best.cellW <= 0) {
        best.cols = Math.min(count, 3)
        best.rows = Math.ceil(count / best.cols)
        best.cellW = Math.max(120, (areaW - gap * (best.cols - 1)) / best.cols)
        best.cellH = best.cellW / aspect
    }

    grid.style.gridTemplateColumns = 'repeat(' + best.cols + ', 1fr)'
    grid.style.gap = gap + 'px'
    grid.querySelectorAll('.cv-mosaic-frame').forEach(function (el) {
        el.style.aspectRatio = '16 / 9'
        el.style.width = '100%'
        el.style.maxHeight = best.cellH + 'px'
    })
}

function toggleMosaicFullscreen(frame) {
    if (!frame) return
    if (document.fullscreenElement === frame) {
        document.exitFullscreen().catch(function () {})
        return
    }
    if (document.fullscreenElement) {
        document.exitFullscreen().then(function () {
            frame.requestFullscreen().catch(function () {})
        }).catch(function () {})
        return
    }
    frame.requestFullscreen().catch(function () {})
}

function syncMosaicFullscreenClass() {
    document.querySelectorAll('.cv-mosaic-frame').forEach(function (el) {
        el.classList.toggle('is-fullscreen', document.fullscreenElement === el)
    })
}

window.addEventListener('beforeunload', function () {
    ConfVisionMosaicCell.stopAll()
})

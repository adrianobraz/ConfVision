let cacheCameras = []
let cacheSegmentos = []

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    $('#modal-gravacao-video').on('hidden.bs.modal', function () {
        const video = document.getElementById('modal-gravacao-player')
        if (video) {
            video.pause()
            video.removeAttribute('src')
            video.load()
        }
    })
    $('#btn-atualizar-timeline').on('click', iniciar)
    $('#btn-buscar-segmentos').on('click', buscarSegmentos)
    $('#sel-timeline-camera').on('change', buscarSegmentos)
    $('#tab-gravacao-segmentos').on('click', '.btn-assistir-segmento', function () {
        const idx = $(this).data('idx')
        if (cacheSegmentos[idx]) {
            abrirVideoSegmento(cacheSegmentos[idx])
        }
    })
    iniciar()
})

function iniciar() {
    carregarCameras().always(function () {
        preencherSelectCameras()
        aplicarCameraDaUrl()
        if (!cameraIdDaUrl()) {
            buscarSegmentos()
        }
    })
}

function cameraIdDaUrl() {
    const params = new URLSearchParams(window.location.search)
    return params.get('camera') || params.get('vis_camera_id') || ''
}

function aplicarCameraDaUrl() {
    const camId = cameraIdDaUrl()
    if (!camId) return
    $('#sel-timeline-camera').val(String(camId))
    buscarSegmentos()
}

function carregarCameras() {
    return $.get(`/api/cameras?${ConfVisionUrls.camerasQueryString()}`)
        .then(function (r) {
            cacheCameras = r.dados || r || []
            return cacheCameras
        })
        .fail(function () {
            cacheCameras = []
        })
}

function gravacaoAtiva(cam) {
    return String(cam.gravacao_status || '').toLowerCase() === 'ativa'
}

function camerasParaTimeline() {
    return cacheCameras.slice().sort(function (a, b) {
        const ga = gravacaoAtiva(a) ? 0 : 1
        const gb = gravacaoAtiva(b) ? 0 : 1
        if (ga !== gb) return ga - gb
        return String(a.nome || a.id).localeCompare(String(b.nome || b.id), 'pt-BR')
    })
}

function preencherSelectCameras() {
    const sel = $('#sel-timeline-camera')
    const atual = sel.val() || cameraIdDaUrl()
    sel.find('option:not(:first)').remove()
    camerasParaTimeline().forEach(function (cam) {
        const gravando = gravacaoAtiva(cam)
        const inativa = cam.ativo === false || cam.ativo === 'N' || cam.ativo === 'n'
        let sufixo = ''
        if (gravando) sufixo = ' ● gravando'
        else if (inativa) sufixo = ' (inativa)'
        sel.append(
            $('<option></option>')
                .val(cam.id)
                .text(`${cam.nome || 'Câmera'} (#${cam.id})${sufixo}`)
        )
    })
    if (atual) sel.val(String(atual))
}

function nomeCameraPorId(idCamera) {
    if (!idCamera || !cacheCameras.length) return null
    const id = String(idCamera)
    const camera = cacheCameras.find(function (c) {
        return String(c.id) === id
    })
    if (!camera) return null
    return camera.nome || null
}

function formatarBytes(bytes) {
    if (!bytes || bytes <= 0) return '—'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function escHtml(valor) {
    if (valor == null || valor === '') return '—'
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function extrairItemsResposta(r) {
    const dados = r && r.dados != null ? r.dados : r
    if (Array.isArray(dados)) return dados
    if (dados && Array.isArray(dados.items)) return dados.items
    return []
}

function timestampSegmento(val) {
    if (val == null || val === '') return null
    if (typeof val === 'number') return val
    const t = new Date(val).getTime()
    return isNaN(t) ? null : t
}

/** Converte datetime-local em ISO UTC. "Até" com 00:00 vira fim do dia local. */
function parseFiltroDataInput(val, papel) {
    if (!val || !String(val).trim()) return null
    const bruto = String(val).trim()
    const d = new Date(bruto)
    if (isNaN(d.getTime())) return null
    if (papel === 'ate' && /T00:00(:00)?$/.test(bruto)) {
        d.setHours(23, 59, 59, 999)
    }
    return d.toISOString()
}

function obterFiltroDatas() {
    const deRaw = $('#filtro-de').val()
    const ateRaw = $('#filtro-ate').val()
    const de = parseFiltroDataInput(deRaw, 'de')
    const ate = parseFiltroDataInput(ateRaw, 'ate')

    if (deRaw && !de) {
        return { erro: 'Data inicial inválida.' }
    }
    if (ateRaw && !ate) {
        return { erro: 'Data final inválida.' }
    }
    if (de && ate && new Date(de).getTime() > new Date(ate).getTime()) {
        return { erro: 'A data "De" não pode ser posterior à data "Até".' }
    }

    return {
        de: de,
        ate: ate,
        deMs: de ? new Date(de).getTime() : null,
        ateMs: ate ? new Date(ate).getTime() : null
    }
}

function montarUrlSegmentos(camId, de, ate, page) {
    const idFranq = encodeURIComponent(ConfVisionUrls.idFranqueado())
    let url = `/api/gravacao-segmentos?id_franqueado=${idFranq}&page=${page || 1}`
    if (camId) {
        url += '&vis_camera_id=' + encodeURIComponent(camId)
    }
    if (de) url += '&de=' + encodeURIComponent(de)
    if (ate) url += '&ate=' + encodeURIComponent(ate)
    return url
}

function buscarTodosSegmentosPag(camId, de, ate) {
    let pagina = 1
    let acumulado = []

    function proximaPagina() {
        return $.get(montarUrlSegmentos(camId, de, ate, pagina)).then(function (r) {
            const items = extrairItemsResposta(r)
            acumulado = acumulado.concat(items)
            const dados = r && r.dados != null ? r.dados : r
            const next = dados && dados.nextPage != null ? dados.nextPage : null
            if (next && items.length) {
                pagina = next
                return proximaPagina()
            }
            return acumulado
        })
    }

    return proximaPagina()
}

function filtrarSegmentosClientSide(items, deMs, ateMs) {
    if (deMs == null && ateMs == null) return items
    return items.filter(function (seg) {
        const t = timestampSegmento(seg.inicio_em)
        if (t == null) return false
        if (deMs != null && t < deMs) return false
        if (ateMs != null && t > ateMs) return false
        return true
    })
}

function ordenarSegmentos(items) {
    return items.slice().sort(function (a, b) {
        return (timestampSegmento(b.inicio_em) || 0) - (timestampSegmento(a.inicio_em) || 0)
    })
}

function buscarSegmentos() {
    const filtro = obterFiltroDatas()
    if (filtro.erro) {
        CvMsg.aviso(filtro.erro)
        return
    }

    const camId = $('#sel-timeline-camera').val()
    const tbody = $('#tab-gravacao-segmentos')
    tbody.html('<tr><td colspan="7" class="cv-empty">Carregando...</td></tr>')

    if (camId) {
        buscarTodosSegmentosPag(camId, filtro.de, filtro.ate)
            .then(function (items) {
                const filtrados = filtrarSegmentosClientSide(items, filtro.deMs, filtro.ateMs)
                renderSegmentos(ordenarSegmentos(filtrados), tbody)
            })
            .catch(function (xhr) {
                console.error('Erro segmentos:', xhr)
                tbody.html('<tr><td colspan="7" class="cv-empty">Erro ao carregar segmentos.</td></tr>')
            })
        return
    }

    buscarTodosSegmentosPag('', filtro.de, filtro.ate)
        .then(function (items) {
            if (items.length) {
                const filtrados = filtrarSegmentosClientSide(items, filtro.deMs, filtro.ateMs)
                renderSegmentos(ordenarSegmentos(filtrados), tbody)
                return
            }
            return buscarSegmentosTodasCameras(filtro, tbody)
        })
        .catch(function () {
            buscarSegmentosTodasCameras(filtro, tbody)
        })
}

function buscarSegmentosTodasCameras(filtro, tbody) {
    const cameras = camerasParaTimeline()
    if (!cameras.length) {
        tbody.html('<tr><td colspan="7" class="cv-empty">Nenhum segmento encontrado.</td></tr>')
        return Promise.resolve()
    }

    const reqs = cameras.map(function (cam) {
        return buscarTodosSegmentosPag(cam.id, filtro.de, filtro.ate).catch(function () {
            return []
        })
    })

    return Promise.all(reqs).then(function (listas) {
        const merged = []
        listas.forEach(function (items) {
            items.forEach(function (seg) { merged.push(seg) })
        })
        const filtrados = filtrarSegmentosClientSide(merged, filtro.deMs, filtro.ateMs)
        renderSegmentos(ordenarSegmentos(filtrados), tbody)
    }).catch(function () {
        tbody.html('<tr><td colspan="7" class="cv-empty">Erro ao carregar segmentos.</td></tr>')
    })
}

function renderSegmentos(items, tbody) {
    cacheSegmentos = items || []
    tbody.empty()
    if (!cacheSegmentos.length) {
        tbody.append('<tr><td colspan="7" class="cv-empty">Nenhum segmento encontrado.</td></tr>')
        return
    }
    cacheSegmentos.forEach(function (seg, idx) {
        const tr = $('<tr></tr>')
        const nomeCam = nomeCameraPorId(seg.vis_camera_id)
        const tipo = normalizarTipoSegmento(seg)
        tr.append(`<td data-label="Tipo" class="cv-gravacao-tipo-col">${htmlBadgeTipo(tipo)}</td>`)
        tr.append(`<td data-label="Câmera">${escHtml(nomeCam || seg.vis_camera_id)}</td>`)
        tr.append(`<td data-label="Início">${confVisionFormatarData(seg.inicio_em)}</td>`)
        tr.append(`<td data-label="Fim">${confVisionFormatarData(seg.fim_em)}</td>`)
        tr.append(`<td data-label="Duração">${seg.duracao_seg ? seg.duracao_seg + 's' : '—'}</td>`)
        tr.append(`<td data-label="Tamanho">${formatarBytes(seg.tamanho_bytes)}</td>`)
        const tdVideo = $('<td data-label="Vídeo" class="cv-cell-action"></td>')
        if (seg.s3_url) {
            tdVideo.append(
                $('<button type="button" class="cv-btn-ghost cv-btn-sm btn-assistir-segmento"><i class="bi bi-play-circle"></i> Assistir</button>')
                    .attr('data-idx', idx)
            )
            tdVideo.append(
                $('<a class="cv-btn-ghost cv-btn-sm btn-baixar-segmento ms-1"><i class="bi bi-download"></i> Baixar</a>')
                    .attr('href', videoDownloadUrlSegmento(seg))
                    .attr('download', nomeArquivoSegmento(seg))
            )
        } else {
            tdVideo.text('—')
        }
        tr.append(tdVideo)
        tbody.append(tr)
    })
}

function obterModalGravacao() {
    const el = document.getElementById('modal-gravacao-video')
    if (!el || typeof bootstrap === 'undefined') return null
    return bootstrap.Modal.getOrCreateInstance(el)
}

function videoUrlSegmento(seg) {
    const idFranq = encodeURIComponent(ConfVisionUrls.idFranqueado())
    return `/api/gravacao-segmento/video?id_franqueado=${idFranq}&url=${encodeURIComponent(seg.s3_url)}`
}

function nomeArquivoSegmento(seg) {
    const base = nomeCameraPorId(seg.vis_camera_id) || ('camera-' + (seg.vis_camera_id || ''))
    const nomeCam = String(base).replace(/[^\w-]+/g, '-').replace(/^-+|-+$/g, '').toLowerCase() || 'camera'
    const t = timestampSegmento(seg.inicio_em)
    let stamp = 'gravacao'
    if (t) {
        const d = new Date(t)
        const p = function (n) { return String(n).padStart(2, '0') }
        stamp = `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}_${p(d.getHours())}-${p(d.getMinutes())}-${p(d.getSeconds())}`
    }
    return `${nomeCam}-${stamp}.mp4`
}

function videoDownloadUrlSegmento(seg) {
    return videoUrlSegmento(seg) + '&download=1&nome=' + encodeURIComponent(nomeArquivoSegmento(seg))
}

function abrirVideoSegmento(seg) {
    if (!seg || !seg.s3_url) return
    const nomeCam = nomeCameraPorId(seg.vis_camera_id) || `Câmera #${seg.vis_camera_id}`
    const titulo = `${nomeCam} — ${confVisionFormatarData(seg.inicio_em)}`
    $('#modal-gravacao-tipo-badge').html(htmlBadgeTipo(normalizarTipoSegmento(seg)))
    $('#modal-gravacao-titulo').text(titulo)
    const dl = document.getElementById('modal-gravacao-download')
    if (dl) {
        dl.href = videoDownloadUrlSegmento(seg)
        dl.setAttribute('download', nomeArquivoSegmento(seg))
    }
    const proxyUrl = videoUrlSegmento(seg)
    const video = document.getElementById('modal-gravacao-player')
    if (video) {
        video.src = proxyUrl
        video.load()
    }
    const modal = obterModalGravacao()
    if (modal) {
        modal.show()
        return
    }
    window.open(proxyUrl, '_blank', 'noopener')
}

/* ConfVision DVR — reprodução estilo NVR */
const DVR_BLOCK_SEC = 10
const DVR_BLOCK_PX_MIN = 3
const DVR_PADDING_MS = 2 * 60 * 1000
const DVR_SPAN_MIN_MS = 20 * 60 * 1000

let cacheCameras = []
let segmentosDia = []
let diaSelecionado = ''
let posicaoAtualMs = 0
let segmentoCarregadoId = null
let carregandoSegmento = false
let modalCaptura = null
let pickerInstance = null
let diasComGravacao = new Set()
let timelineViewport = {
    inicioMs: 0,
    fimMs: 0,
    blocks: 0,
    widthPx: 0,
    blockPx: DVR_BLOCK_PX_MIN
}

const video = () => document.getElementById('dvr-player')

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()

    const elCaptura = document.getElementById('modal-dvr-captura')
    if (elCaptura && window.bootstrap) {
        modalCaptura = bootstrap.Modal.getOrCreateInstance(elCaptura)
    }

    diaSelecionado = dataDaUrl() || hojeLocal()

    $('#sel-dvr-camera').on('change', function () {
        recarregarDia()
        atualizarDiasCalendario()
    })

    initCalendario()
    $('#dvr-lista-segmentos').on('click', 'li[data-idx]', function () {
        const idx = parseInt($(this).data('idx'), 10)
        if (segmentosDia[idx]) {
            irParaTimestamp(segmentosDia[idx].inicio_em, true)
        }
    })
    $('#dvr-btn-seg-anterior').on('click', irParaSegmentoAnterior)
    $('#dvr-btn-seg-proximo').on('click', irParaSegmentoProximo)
    $('#dvr-btn-menos-10s').on('click', function () { pularSegundos(-DVR_BLOCK_SEC) })
    $('#dvr-btn-mais-10s').on('click', function () { pularSegundos(DVR_BLOCK_SEC) })
    $('#dvr-btn-play').on('click', togglePlay)
    $('#dvr-btn-stop').on('click', parar)
    $('#sel-dvr-velocidade').on('change', function () {
        const v = video()
        if (v) v.playbackRate = parseFloat($(this).val()) || 1
    })
    $('#dvr-btn-captura').on('click', abrirCaptura)
    $('#dvr-btn-download').on('click', baixarSegmentoAtual)
    $('#dvr-btn-fullscreen').on('click', toggleFullscreen)
    $('#dvr-captura-reset').on('click', resetarVistaCaptura)
    $('#dvr-captura-salvar').on('click', salvarCaptura)

    initTimelineCanvas()
    bindVideoEvents()
    document.addEventListener('fullscreenchange', atualizarIconeFullscreen)
    iniciar()
})

// ------------------------------------------------------------------ //
// Calendário Flatpickr com dias com gravação destacados               //
// ------------------------------------------------------------------ //

function formatarDataKey(d) {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
}

function initCalendario() {
    if (typeof flatpickr === 'undefined') {
        // fallback para input nativo se flatpickr não carregou
        const el = document.getElementById('dvr-data')
        if (el) {
            el.type = 'date'
            el.value = diaSelecionado
            el.addEventListener('change', function () {
                diaSelecionado = this.value
                recarregarDia()
            })
        }
        return
    }
    if (flatpickr.l10ns && flatpickr.l10ns.pt) {
        flatpickr.localize(flatpickr.l10ns.pt)
    }
    pickerInstance = flatpickr('#dvr-data', {
        dateFormat: 'Y-m-d',
        defaultDate: diaSelecionado,
        maxDate: hojeLocal(),
        onDayCreate: function (dObj, dStr, fp, dayElem) {
            const key = formatarDataKey(dayElem.dateObj)
            if (diasComGravacao.has(key)) {
                dayElem.classList.add('cv-dia-com-gravacao')
            }
        },
        onChange: function (selectedDates, dateStr) {
            if (!dateStr) return
            diaSelecionado = dateStr
            recarregarDia()
        },
        onMonthChange: function () {
            atualizarDiasCalendario()
        },
        onYearChange: function () {
            atualizarDiasCalendario()
        }
    })
}

function atualizarDiasCalendario() {
    if (!pickerInstance) return
    const camId = $('#sel-dvr-camera').val()
    if (!camId) return
    const ano = pickerInstance.currentYear
    const mes = pickerInstance.currentMonth + 1
    carregarDiasDoMes(camId, ano, mes).then(function (dias) {
        diasComGravacao = dias
        pickerInstance.redraw()
    })
}

function carregarDiasDoMes(camId, ano, mes) {
    const de = new Date(ano, mes - 1, 1, 0, 0, 0).getTime()
    const ate = new Date(ano, mes, 0, 23, 59, 59, 999).getTime()
    let pagina = 1
    const dias = new Set()

    function proximaPagina() {
        return $.get(montarUrlSegmentos(camId, de, ate, pagina)).then(function (r) {
            const items = extrairItemsResposta(r)
            items.forEach(function (seg) {
                const ts = normalizarTimestamp(seg.inicio_em)
                if (ts) {
                    dias.add(formatarDataKey(new Date(ts)))
                }
            })
            const dados = r && r.dados != null ? r.dados : r
            const next = dados && dados.nextPage != null ? dados.nextPage : null
            if (next && items.length) {
                pagina = next
                return proximaPagina()
            }
            return dias
        }).fail(function () {
            return dias
        })
    }

    return proximaPagina()
}

function atualizarIconeFullscreen() {
    const btn = $('#dvr-btn-fullscreen i')
    if (!btn.length) return
    if (document.fullscreenElement) {
        btn.removeClass('bi-fullscreen').addClass('bi-fullscreen-exit')
    } else {
        btn.removeClass('bi-fullscreen-exit').addClass('bi-fullscreen')
    }
}

function hojeLocal() {
    const d = new Date()
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
}

function dataDaUrl() {
    const p = new URLSearchParams(window.location.search)
    return p.get('data') || ''
}

function cameraIdDaUrl() {
    const p = new URLSearchParams(window.location.search)
    return p.get('camera') || p.get('vis_camera_id') || ''
}

function limitesDia(dataStr) {
    const inicio = new Date(dataStr + 'T00:00:00').getTime()
    return {
        de: inicio,
        ate: inicio + 86400000 - 1,
        inicio: inicio
    }
}

function normalizarTimestamp(valor) {
    if (valor == null || valor === '') return 0
    if (typeof valor === 'string' && /[T\-]/.test(valor)) {
        const parsed = Date.parse(valor)
        return isFinite(parsed) ? parsed : 0
    }
    let n = Number(valor)
    if (!isFinite(n)) return 0
    if (n > 0 && n < 1e12) n *= 1000
    return n
}

function normalizarSegmento(seg) {
    if (!seg) return seg
    seg.inicio_em = normalizarTimestamp(seg.inicio_em)
    if (seg.fim_em) seg.fim_em = normalizarTimestamp(seg.fim_em)
    if (seg.duracao_seg) seg.duracao_seg = Number(seg.duracao_seg) || 0
    return seg
}

function iniciar() {
    carregarCameras().always(function () {
        preencherSelectCameras()
        recarregarDia()
        atualizarDiasCalendario()
    })
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

function camerasComGravacao() {
    return cacheCameras.filter(function (c) {
        return c.grava_continua || c.grava_movimento || c.grava_timelapse || c.vis_licenca_gravacao_id
    })
}

function preencherSelectCameras() {
    const sel = $('#sel-dvr-camera')
    const atual = cameraIdDaUrl()
    sel.empty()

    // DVR mostra todas as câmeras — câmeras desativadas podem ter gravações históricas
    const lista = cacheCameras
    if (!lista.length) {
        sel.append($('<option value="">Nenhuma câmera disponível</option>'))
        return
    }
    lista.forEach(function (cam) {
        const ativa = cam.grava_continua || cam.grava_movimento || cam.grava_timelapse || cam.vis_licenca_gravacao_id
        const sufixo = ativa ? ' ● gravando' : ''
        sel.append(
            $('<option></option>')
                .val(cam.id)
                .text(`${cam.nome || 'Câmera'} (#${cam.id})${sufixo}`)
        )
    })
    if (atual) {
        sel.val(String(atual))
    } else if (lista.length) {
        sel.val(String(lista[0].id))
    }
}

function nomeCameraAtual() {
    const id = $('#sel-dvr-camera').val()
    const cam = cacheCameras.find(function (c) { return String(c.id) === String(id) })
    return cam ? (cam.nome || `Câmera #${cam.id}`) : ''
}

function extrairItemsResposta(r) {
    const dados = r && r.dados != null ? r.dados : r
    if (Array.isArray(dados)) return dados
    if (dados && Array.isArray(dados.items)) return dados.items
    return []
}

function montarUrlSegmentos(camId, de, ate, page) {
    const idFranq = encodeURIComponent(ConfVisionUrls.idFranqueado())
    let url = `/api/gravacao-segmentos?id_franqueado=${idFranq}&vis_camera_id=${encodeURIComponent(camId)}&page=${page || 1}`
    if (de) url += '&de=' + encodeURIComponent(new Date(de).toISOString())
    if (ate) url += '&ate=' + encodeURIComponent(new Date(ate).toISOString())
    return url
}

function buscarTodosSegmentos(camId, de, ate) {
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

function segmentoMaisAntigo() {
    if (!segmentosDia.length) return null
    return segmentosDia[segmentosDia.length - 1]
}

function inicioPlaybackDia() {
    const seg = segmentoMaisAntigo()
    return seg ? seg.inicio_em : null
}

function recarregarDia() {
    const camId = $('#sel-dvr-camera').val()
    diaSelecionado = $('#dvr-data').val() || diaSelecionado
    parar(true)
    segmentosDia = []
    segmentoCarregadoId = null
    renderListaSegmentos()
    desenharTimeline([])
    setStatus('Carregando gravações...')

    if (!camId) {
        setStatus('Selecione uma câmera com gravação ativa.')
        return
    }

    const limites = limitesDia(diaSelecionado)
    buscarTodosSegmentos(camId, limites.de, limites.ate)
        .then(function (items) {
            segmentosDia = (items || [])
                .map(normalizarSegmento)
                .filter(function (s) { return s.s3_url && s.inicio_em })
                .sort(function (a, b) { return b.inicio_em - a.inicio_em })
            renderListaSegmentos()
            desenharTimeline(segmentosDia)

            if (!segmentosDia.length) {
                setStatus('Nenhuma gravação neste dia.')
                mostrarOverlay(true, 'Nenhuma gravação neste dia')
                return
            }

            setStatus(`${segmentosDia.length} segmento(s) encontrado(s).`)
            mostrarOverlay(false)
            irParaTimestamp(inicioPlaybackDia(), false)
        })
        .fail(function () {
            setStatus('Erro ao carregar gravações.')
            renderListaSegmentos(true)
        })
}

function renderListaSegmentos(erro) {
    const ul = $('#dvr-lista-segmentos')
    ul.empty()
    if (erro) {
        ul.append('<li class="cv-dvr-lista-vazia">Erro ao carregar.</li>')
        return
    }
    if (!segmentosDia.length) {
        ul.append('<li class="cv-dvr-lista-vazia">Nenhum segmento.</li>')
        return
    }
    segmentosDia.forEach(function (seg, idx) {
        const hora = formatarHora(seg.inicio_em) + ' – ' + formatarHora(seg.fim_em)
        const dur = seg.duracao_seg ? seg.duracao_seg + 's' : ''
        const badge = htmlBadgeTipo(normalizarTipoSegmento(seg))
        ul.append(
            $('<li></li>')
                .attr('data-idx', idx)
                .html(
                    badge +
                    `<span class="cv-dvr-lista-hora">${hora}</span>` +
                    `<span class="cv-dvr-lista-dur">${dur}</span>`
                )
        )
    })
}

function formatarHora(ts) {
    if (!ts) return '--:--:--'
    return moment(ts).format('HH:mm:ss')
}

function setStatus(msg) {
    $('#dvr-status').text(msg || '')
}

function mostrarOverlay(show, msg) {
    const ov = $('#dvr-player-overlay')
    if (show) {
        ov.removeClass('d-none')
        if (msg) $('#dvr-player-msg').text(msg)
    } else {
        ov.addClass('d-none')
    }
}

function atualizarLabelPlayer() {
    const nome = nomeCameraAtual()
    const hora = posicaoAtualMs ? formatarHora(posicaoAtualMs) : '--:--:--'
    const seg = segmentoEmTimestamp(posicaoAtualMs) || segmentoNoIndice(segmentoCarregadoId)
    const badge = seg ? htmlBadgeTipo(normalizarTipoSegmento(seg), 'cv-tipo-badge-sm') : ''
    const texto = nome ? `${nome} — ${hora}` : hora
    $('#dvr-player-label').html(badge + escHtmlPlayer(texto))
}

function escHtmlPlayer(valor) {
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
}

/* ——— Timeline canvas ——— */

function initTimelineCanvas() {
    const scroll = document.getElementById('dvr-timeline-scroll')
    if (!scroll) return
    scroll.addEventListener('click', onTimelineClick)
    scroll.addEventListener('scroll', function () {
        atualizarPlayheadVisual(false)
    })

    let resizeTimer = null
    window.addEventListener('resize', function () {
        clearTimeout(resizeTimer)
        resizeTimer = setTimeout(function () {
            if (segmentosDia.length) {
                desenharTimeline(segmentosDia)
            }
        }, 120)
    })
}

function larguraContainerTimeline() {
    const scroll = document.getElementById('dvr-timeline-scroll')
    if (!scroll) return 800
    return scroll.clientWidth || scroll.offsetWidth || 800
}

function dimensoesTimeline(blocks) {
    const containerW = larguraContainerTimeline()
    const naturalW = blocks * DVR_BLOCK_PX_MIN
    const widthPx = Math.max(containerW, naturalW)
    const blockPx = widthPx / blocks
    return { widthPx: widthPx, blockPx: blockPx }
}

function calcularViewportTimeline(segmentos) {
    const limites = limitesDia(diaSelecionado)
    if (!segmentos.length) {
        const blocks = Math.ceil(86400000 / 1000 / DVR_BLOCK_SEC)
        const dim = dimensoesTimeline(blocks)
        return {
            inicioMs: limites.inicio,
            fimMs: limites.ate,
            blocks: blocks,
            widthPx: dim.widthPx,
            blockPx: dim.blockPx
        }
    }

    const lista = segmentos.slice().sort(function (a, b) { return a.inicio_em - b.inicio_em })
    const primeiro = lista[0].inicio_em
    let ultimo = primeiro
    lista.forEach(function (seg) {
        ultimo = Math.max(ultimo, fimSegmento(seg))
    })

    let inicioMs = Math.max(limites.inicio, primeiro - DVR_PADDING_MS)
    let fimMs = Math.min(limites.ate, ultimo + DVR_PADDING_MS)

    if (fimMs - inicioMs < DVR_SPAN_MIN_MS) {
        const meio = (primeiro + ultimo) / 2
        inicioMs = Math.max(limites.inicio, meio - DVR_SPAN_MIN_MS / 2)
        fimMs = Math.min(limites.ate, meio + DVR_SPAN_MIN_MS / 2)
    }

    const blocks = Math.max(1, Math.ceil((fimMs - inicioMs) / 1000 / DVR_BLOCK_SEC))
    const dim = dimensoesTimeline(blocks)
    return {
        inicioMs: inicioMs,
        fimMs: fimMs,
        blocks: blocks,
        widthPx: dim.widthPx,
        blockPx: dim.blockPx
    }
}

function aplicarDimensoesTimeline() {
    const canvas = document.getElementById('dvr-timeline-canvas')
    const inner = document.getElementById('dvr-timeline-inner')
    if (!canvas || !inner) return
    canvas.width = Math.round(timelineViewport.widthPx)
    canvas.height = 40
    canvas.style.width = timelineViewport.widthPx + 'px'
    inner.style.width = timelineViewport.widthPx + 'px'
}

function posicaoXTimeline(ev) {
    const scroll = document.getElementById('dvr-timeline-scroll')
    if (!scroll) return 0
    const rect = scroll.getBoundingClientRect()
    return ev.clientX - rect.left + scroll.scrollLeft
}

function blocosGravados(segmentos) {
    const mapa = new Uint8Array(timelineViewport.blocks)
    segmentos.forEach(function (seg) {
        const tipoVal = valorTimelineTipo(seg)
        const ini = Math.max(0, Math.floor((seg.inicio_em - timelineViewport.inicioMs) / 1000 / DVR_BLOCK_SEC))
        const fim = Math.min(timelineViewport.blocks, Math.ceil((fimSegmento(seg) - timelineViewport.inicioMs) / 1000 / DVR_BLOCK_SEC))
        for (let i = ini; i < fim; i++) {
            mapa[i] = tipoVal
        }
    })
    return mapa
}

function desenharTimeline(segmentos) {
    requestAnimationFrame(function () {
        timelineViewport = calcularViewportTimeline(segmentos)
        aplicarDimensoesTimeline()

        const canvas = document.getElementById('dvr-timeline-canvas')
        if (!canvas) return
        const ctx = canvas.getContext('2d')
        const mapa = blocosGravados(segmentos)
        const h = canvas.height
        const w = canvas.width

        ctx.fillStyle = '#2a3144'
        ctx.fillRect(0, 0, w, h)

        for (let i = 0; i < timelineViewport.blocks; i++) {
            const x = i * timelineViewport.blockPx
            const bw = Math.max(1, timelineViewport.blockPx - (timelineViewport.blockPx > 2 ? 0.5 : 0))
            let cor = '#2a3144'
            if (mapa[i] === 1) cor = '#16a34a'
            else if (mapa[i] === 2) cor = '#f59e0b'
            ctx.fillStyle = cor
            ctx.fillRect(x, 8, bw, h - 16)
        }

        ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)'
        ctx.lineWidth = 1
        const spanMs = timelineViewport.fimMs - timelineViewport.inicioMs
        const passoHoraMs = spanMs > 6 * 3600000 ? 3600000 : 1800000
        for (let t = timelineViewport.inicioMs; t <= timelineViewport.fimMs; t += passoHoraMs) {
            const x = msParaPosicaoTimeline(t)
            ctx.beginPath()
            ctx.moveTo(x, 0)
            ctx.lineTo(x, h)
            ctx.stroke()
        }

        atualizarPlayheadVisual(true)
        atualizarIntervaloTimeline()
        marcarHorasTimeline()
    })
}

function atualizarIntervaloTimeline() {
    const el = $('#dvr-timeline-intervalo')
    if (!el.length || !timelineViewport.blocks) {
        el.text('')
        return
    }
    el.text(
        '· ' + moment(timelineViewport.inicioMs).format('HH:mm') +
        ' – ' + moment(timelineViewport.fimMs).format('HH:mm')
    )
}

function marcarHorasTimeline() {
    const wrap = $('#dvr-timeline-marcas')
    wrap.empty()
    if (!timelineViewport.blocks) return

    const spanMs = timelineViewport.fimMs - timelineViewport.inicioMs
    const passoMs = spanMs > 6 * 3600000 ? 3600000 : 1800000
    for (let t = timelineViewport.inicioMs; t <= timelineViewport.fimMs; t += passoMs) {
        const x = msParaPosicaoTimeline(t)
        wrap.append(
            $('<span class="cv-dvr-marca-hora"></span>')
                .css('left', x + 'px')
                .text(moment(t).format('HH:mm'))
        )
    }
}

function timestampDoBloco(bloco) {
    return timelineViewport.inicioMs + bloco * DVR_BLOCK_SEC * 1000
}

function fimSegmento(seg) {
    if (seg.fim_em) return seg.fim_em
    if (seg.duracao_seg) return seg.inicio_em + seg.duracao_seg * 1000
    return seg.inicio_em + 300 * 1000
}

function resolverSegmento(ts) {
    const lista = segmentosCronologicos()
    if (!lista.length) return null
    for (let i = 0; i < lista.length; i++) {
        const seg = lista[i]
        if (ts >= seg.inicio_em && ts < fimSegmento(seg)) {
            return seg
        }
    }
    for (let i = 0; i < lista.length; i++) {
        if (ts < lista[i].inicio_em) {
            return lista[i]
        }
    }
    return lista[lista.length - 1]
}

function onTimelineClick(ev) {
    if (!diaSelecionado || !segmentosDia.length) {
        setStatus('Nenhuma gravação carregada.')
        return
    }

    const x = posicaoXTimeline(ev)
    let bloco = Math.floor(x / timelineViewport.blockPx)
    bloco = Math.max(0, Math.min(timelineViewport.blocks - 1, bloco))

    const ts = timestampDoBloco(bloco)
    const seg = resolverSegmento(ts)
    if (!seg || !seg.s3_url) {
        setStatus('Sem gravação neste horário.')
        return
    }

    const fim = fimSegmento(seg)
    const destino = Math.max(seg.inicio_em, Math.min(ts, fim - 1000))
    posicaoAtualMs = destino
    atualizarPlayheadVisual(true)
    irParaSegmento(seg, destino, true)
}

function msParaPosicaoTimeline(ms) {
    const bloco = (ms - timelineViewport.inicioMs) / 1000 / DVR_BLOCK_SEC
    return bloco * timelineViewport.blockPx
}

function atualizarPlayheadVisual(centralizar) {
    const playhead = document.getElementById('dvr-timeline-playhead')
    const scroll = document.getElementById('dvr-timeline-scroll')
    if (!playhead || !posicaoAtualMs || !timelineViewport.blocks) return
    const x = msParaPosicaoTimeline(posicaoAtualMs)
    playhead.style.left = Math.max(0, Math.min(timelineViewport.widthPx, x)) + 'px'
    $('#dvr-timeline-hora').text(formatarHora(posicaoAtualMs))

    if (scroll) {
        const viewW = scroll.clientWidth
        const scrollLeft = scroll.scrollLeft
        if (centralizar || x < scrollLeft + 40 || x > scrollLeft + viewW - 40) {
            scroll.scrollLeft = Math.max(0, x - viewW / 2)
        }
    }

    $('#dvr-lista-segmentos li').removeClass('cv-dvr-lista-ativo')
    segmentosDia.forEach(function (seg, idx) {
        if (posicaoAtualMs >= seg.inicio_em && posicaoAtualMs < fimSegmento(seg)) {
            $(`#dvr-lista-segmentos li[data-idx="${idx}"]`).addClass('cv-dvr-lista-ativo')
        }
    })
}

/* ——— Player ——— */
function bindVideoEvents() {
    const v = video()
    if (!v) return
    v.addEventListener('timeupdate', function () {
        if (carregandoSegmento) return
        const seg = segmentoNoIndice(segmentoCarregadoId)
        if (!seg) return
        posicaoAtualMs = seg.inicio_em + v.currentTime * 1000
        atualizarLabelPlayer()
        atualizarPlayheadVisual()
    })
    v.addEventListener('ended', function () {
        const seg = segmentoNoIndice(segmentoCarregadoId)
        if (!seg) return
        const proximo = posicaoAtualMs + 500
        if (irParaTimestamp(proximo, true)) {
            v.play().catch(function () {})
        }
    })
    v.addEventListener('play', function () {
        $('#dvr-icon-play').removeClass('bi-play-fill').addClass('bi-pause-fill')
    })
    v.addEventListener('pause', function () {
        $('#dvr-icon-play').removeClass('bi-pause-fill').addClass('bi-play-fill')
    })
}

function segmentoNoIndice(id) {
    return segmentosDia.find(function (s) { return s.id === id }) || null
}

function segmentoEmTimestamp(ts) {
    return segmentosDia.find(function (s) {
        return ts >= s.inicio_em && ts < fimSegmento(s)
    }) || null
}

function urlVideoSegmento(seg) {
    if (!seg || !seg.s3_url) return ''
    const idFranq = encodeURIComponent(ConfVisionUrls.idFranqueado())
    const url = encodeURIComponent(seg.s3_url)
    return `/api/gravacao-segmento/video?id_franqueado=${idFranq}&url=${url}`
}

function nomeArquivoSegmentoDvr(seg) {
    const nome = (nomeCameraAtual() || 'camera').replace(/[^\w-]+/g, '-').replace(/^-+|-+$/g, '').toLowerCase() || 'camera'
    const hora = seg && seg.inicio_em ? moment(seg.inicio_em).format('YYYY-MM-DD_HH-mm-ss') : 'gravacao'
    return `dvr-${nome}-${hora}.mp4`
}

function segmentoParaDownload() {
    return segmentoNoIndice(segmentoCarregadoId) || segmentoEmTimestamp(posicaoAtualMs)
}

function baixarSegmentoAtual() {
    const seg = segmentoParaDownload()
    if (!seg || !seg.s3_url) {
        setStatus('Nenhuma gravação carregada para baixar.')
        return
    }
    const nome = nomeArquivoSegmentoDvr(seg)
    const link = document.createElement('a')
    link.href = urlVideoSegmento(seg) + '&download=1&nome=' + encodeURIComponent(nome)
    link.download = nome
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    setStatus('Baixando ' + nome)
}

function urlBaseVideo(url) {
    if (!url) return ''
    try {
        return decodeURIComponent(String(url).split('?')[0])
    } catch (e) {
        return String(url).split('?')[0]
    }
}

function mesmoSegmentoNoPlayer(seg, v) {
    if (!seg || !v || !v.src) return false
    if (segmentoCarregadoId === seg.id) return true
    const baseSeg = urlBaseVideo(seg.s3_url)
    const baseAtual = urlBaseVideo(v.src)
    return baseSeg && baseAtual && (baseAtual.indexOf(baseSeg) !== -1 || baseSeg.indexOf(baseAtual) !== -1)
}

function irParaSegmento(seg, ts, autoplay) {
    if (!seg || !seg.s3_url) {
        setStatus('Sem gravação neste horário.')
        return false
    }

    ts = normalizarTimestamp(ts)
    const inicio = normalizarTimestamp(seg.inicio_em)
    posicaoAtualMs = ts
    const offsetSec = Math.max(0, (ts - inicio) / 1000)
    const v = video()
    if (!v) return false

    mostrarOverlay(false)
    atualizarLabelPlayer()
    atualizarPlayheadVisual()
    setStatus('Indo para ' + formatarHora(ts) + '…')

    const rate = parseFloat($('#sel-dvr-velocidade').val()) || 1

    if (mesmoSegmentoNoPlayer(seg, v)) {
        segmentoCarregadoId = seg.id
        v.currentTime = offsetSec
        v.playbackRate = rate
        if (autoplay) v.play().catch(function () {})
        setStatus('Reproduzindo ' + formatarHora(ts))
        return true
    }

    carregandoSegmento = true
    segmentoCarregadoId = seg.id

    function aoPronto() {
        v.removeEventListener('loadedmetadata', aoPronto)
        v.removeEventListener('error', aoErro)
        carregandoSegmento = false
        v.currentTime = Math.min(offsetSec, Math.max(0, v.duration - 0.1))
        v.playbackRate = rate
        if (autoplay) v.play().catch(function () {})
        setStatus('Reproduzindo ' + formatarHora(ts))
    }

    function aoErro() {
        v.removeEventListener('loadedmetadata', aoPronto)
        v.removeEventListener('error', aoErro)
        carregandoSegmento = false
        setStatus('Erro ao carregar o vídeo deste horário.')
    }

    v.pause()
    v.addEventListener('loadedmetadata', aoPronto)
    v.addEventListener('error', aoErro, { once: true })
    v.src = urlVideoSegmento(seg)
    v.load()
    return true
}

function irParaTimestamp(ts, autoplay) {
    const seg = segmentoEmTimestamp(normalizarTimestamp(ts))
    if (!seg) {
        setStatus('Sem gravação neste horário.')
        return false
    }
    return irParaSegmento(seg, ts, autoplay)
}

function togglePlay() {
    const v = video()
    if (!v || !v.src) {
        if (segmentosDia.length) irParaTimestamp(inicioPlaybackDia(), true)
        return
    }
    if (v.paused) v.play().catch(function () {})
    else v.pause()
}

function toggleFullscreen() {
    const wrap = document.querySelector('.cv-dvr-player-wrap')
    if (!wrap) return
    if (!document.fullscreenElement) {
        wrap.requestFullscreen().catch(function () {})
    } else {
        document.exitFullscreen()
    }
}

function segmentosCronologicos() {
    return segmentosDia.slice().sort(function (a, b) { return a.inicio_em - b.inicio_em })
}

function indiceSegmentoAtual() {
    const seg = segmentoCarregadoId
        ? segmentoNoIndice(segmentoCarregadoId)
        : segmentoEmTimestamp(posicaoAtualMs)
    if (!seg) return -1
    const lista = segmentosCronologicos()
    return lista.findIndex(function (s) { return s.id === seg.id })
}

function irParaSegmentoAnterior() {
    const lista = segmentosCronologicos()
    if (!lista.length) return
    const v = video()
    const playing = v && !v.paused
    const idx = indiceSegmentoAtual()
    if (idx <= 0) {
        irParaTimestamp(lista[0].inicio_em, playing)
        return
    }
    irParaTimestamp(lista[idx - 1].inicio_em, playing)
}

function irParaSegmentoProximo() {
    const lista = segmentosCronologicos()
    if (!lista.length) return
    const v = video()
    const playing = v && !v.paused
    const idx = indiceSegmentoAtual()
    if (idx < 0) {
        irParaTimestamp(lista[0].inicio_em, playing)
        return
    }
    if (idx >= lista.length - 1) {
        irParaTimestamp(lista[lista.length - 1].inicio_em, playing)
        return
    }
    irParaTimestamp(lista[idx + 1].inicio_em, playing)
}

function parar(silent) {
    const v = video()
    if (v) {
        v.pause()
        if (segmentosDia.length && !silent) {
            irParaTimestamp(inicioPlaybackDia(), false)
        }
    }
    if (!silent) setStatus('Parado.')
}

function pularSegundos(seg) {
    if (!posicaoAtualMs && segmentosDia.length) {
        posicaoAtualMs = inicioPlaybackDia()
    }
    if (!posicaoAtualMs) return
    const v = video()
    const playing = v && !v.paused
    irParaTimestamp(posicaoAtualMs + seg * 1000, playing)
}

/* ——— Captura: aproximar por seleção de área ——— */
let capturaView = null
let capturaDispW = 0
let capturaDispH = 0
let capturaSelecao = null
let capturaArrastando = false
let capturaInicio = null
let capturaVideoSource = null

function videoCaptura() {
    return capturaVideoSource || video()
}

function abrirCaptura() {
    const v = video()
    if (!v || !v.videoWidth || v.readyState < 2) {
        CvMsg.aviso('Reproduza um vídeo antes de capturar a imagem.')
        return
    }
    v.pause()
    capturaVideoSource = v
    capturaSelecao = null
    capturaView = {
        srcX: 0,
        srcY: 0,
        srcW: v.videoWidth,
        srcH: v.videoHeight
    }

    function iniciarPreviewCaptura() {
        sincronizarTamanhoCaptura()
        renderizarCapturaPreview()
        atualizarNivelCaptura()
        limparOverlayCaptura()
        bindCapturaOverlay()
    }

    const modalEl = document.getElementById('modal-dvr-captura')
    if (modalCaptura) {
        modalCaptura.show()
        if (modalEl) {
            modalEl.addEventListener('shown.bs.modal', function () {
                iniciarPreviewCaptura()
            }, { once: true })
        } else {
            requestAnimationFrame(iniciarPreviewCaptura)
        }
    } else {
        iniciarPreviewCaptura()
    }
}

function sincronizarTamanhoCaptura() {
    const src = videoCaptura()
    const canvas = document.getElementById('dvr-captura-canvas')
    const viewport = document.getElementById('dvr-captura-viewport')
    const overlay = document.getElementById('dvr-captura-overlay')
    if (!src || !canvas || !viewport || !overlay) return

    const vw = src.videoWidth
    const vh = src.videoHeight
    if (!vw || !vh) return

    const stage = document.getElementById('dvr-captura-stage')
    const wrap = document.getElementById('dvr-captura-wrap')
    let maxW = (stage && stage.clientWidth > 0) ? stage.clientWidth - 16 : 0
    if (maxW < 320 && wrap && wrap.parentElement) {
        maxW = wrap.parentElement.clientWidth - 32
    }
    if (maxW < 320) maxW = Math.min(window.innerWidth - 48, 900)

    const maxH = Math.round(window.innerHeight * 0.55)
    let dispW = maxW
    let dispH = Math.round(dispW * vh / vw)
    if (dispH > maxH) {
        dispH = maxH
        dispW = Math.round(dispH * vw / vh)
    }

    capturaDispW = dispW
    capturaDispH = dispH
    viewport.style.width = dispW + 'px'
    viewport.style.height = dispH + 'px'
    canvas.width = dispW
    canvas.height = dispH
    canvas.style.width = dispW + 'px'
    canvas.style.height = dispH + 'px'
    overlay.width = dispW
    overlay.height = dispH
    overlay.style.width = dispW + 'px'
    overlay.style.height = dispH + 'px'
}

function renderizarCapturaPreview() {
    const src = videoCaptura()
    const canvas = document.getElementById('dvr-captura-canvas')
    if (!src || !canvas || !capturaView || !src.videoWidth || !capturaDispW) return

    const ctx = canvas.getContext('2d')
    ctx.fillStyle = '#000'
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    ctx.imageSmoothingEnabled = true
    try {
        ctx.drawImage(
            src,
            capturaView.srcX, capturaView.srcY, capturaView.srcW, capturaView.srcH,
            0, 0, capturaDispW, capturaDispH
        )
    } catch (err) {
        console.error('Erro ao renderizar captura:', err)
    }
}

function telaParaVideoCoords(tx, ty) {
    const rx = tx / capturaDispW
    const ry = ty / capturaDispH
    return {
        x: capturaView.srcX + rx * capturaView.srcW,
        y: capturaView.srcY + ry * capturaView.srcH
    }
}

function capturaAproximarSelecao(rect) {
    if (!rect || rect.w < 16 || rect.h < 16 || !capturaView) return

    const p1 = telaParaVideoCoords(rect.x, rect.y)
    const p2 = telaParaVideoCoords(rect.x + rect.w, rect.y + rect.h)
    const src = videoCaptura()
    if (!src) return

    let srcX = Math.min(p1.x, p2.x)
    let srcY = Math.min(p1.y, p2.y)
    let srcW = Math.abs(p2.x - p1.x)
    let srcH = Math.abs(p2.y - p1.y)

    srcX = Math.max(0, srcX)
    srcY = Math.max(0, srcY)
    srcW = Math.max(8, Math.min(src.videoWidth - srcX, srcW))
    srcH = Math.max(8, Math.min(src.videoHeight - srcY, srcH))

    capturaView = { srcX: srcX, srcY: srcY, srcW: srcW, srcH: srcH }
    capturaSelecao = null
    limparOverlayCaptura()
    renderizarCapturaPreview()
    atualizarNivelCaptura()
}

function capturaNivelAproximacao() {
    const src = videoCaptura()
    if (!src || !capturaView || !src.videoWidth) return 1
    const zx = src.videoWidth / capturaView.srcW
    const zy = src.videoHeight / capturaView.srcH
    return Math.max(zx, zy)
}

function atualizarNivelCaptura() {
    const nivel = capturaNivelAproximacao()
    const el = $('#dvr-captura-nivel')
    if (nivel <= 1.05) {
        el.text('Vista completa')
    } else {
        el.text('Aproximação: ' + nivel.toFixed(1) + '×')
    }
}

function bindCapturaOverlay() {
    const overlay = document.getElementById('dvr-captura-overlay')
    if (!overlay || overlay._dvrBound) return
    overlay._dvrBound = true

    overlay.addEventListener('mousedown', function (e) {
        capturaArrastando = true
        const r = overlay.getBoundingClientRect()
        capturaInicio = { x: e.clientX - r.left, y: e.clientY - r.top }
        capturaSelecao = null
    })
    overlay.addEventListener('mousemove', function (e) {
        if (!capturaArrastando || !capturaInicio) return
        const r = overlay.getBoundingClientRect()
        const x = e.clientX - r.left
        const y = e.clientY - r.top
        capturaSelecao = normalizarRect(capturaInicio.x, capturaInicio.y, x, y)
        desenharSelecaoCaptura()
    })
    overlay.addEventListener('mouseup', function () {
        if (capturaArrastando && capturaSelecao) {
            capturaAproximarSelecao(capturaSelecao)
        }
        capturaArrastando = false
    })
    overlay.addEventListener('mouseleave', function () {
        capturaArrastando = false
    })
}

function normalizarRect(x1, y1, x2, y2) {
    return {
        x: Math.min(x1, x2),
        y: Math.min(y1, y2),
        w: Math.abs(x2 - x1),
        h: Math.abs(y2 - y1)
    }
}

function limparOverlayCaptura() {
    const overlay = document.getElementById('dvr-captura-overlay')
    if (!overlay) return
    overlay.getContext('2d').clearRect(0, 0, overlay.width, overlay.height)
}

function desenharSelecaoCaptura() {
    const overlay = document.getElementById('dvr-captura-overlay')
    if (!overlay || !capturaSelecao) return
    const ctx = overlay.getContext('2d')
    ctx.clearRect(0, 0, overlay.width, overlay.height)
    ctx.strokeStyle = '#38bdf8'
    ctx.lineWidth = 2
    ctx.setLineDash([6, 4])
    ctx.strokeRect(capturaSelecao.x, capturaSelecao.y, capturaSelecao.w, capturaSelecao.h)
    ctx.fillStyle = 'rgba(56, 189, 248, 0.15)'
    ctx.fillRect(capturaSelecao.x, capturaSelecao.y, capturaSelecao.w, capturaSelecao.h)
}

function resetarVistaCaptura() {
    const src = videoCaptura()
    if (!src || !src.videoWidth) return
    capturaView = {
        srcX: 0,
        srcY: 0,
        srcW: src.videoWidth,
        srcH: src.videoHeight
    }
    capturaSelecao = null
    limparOverlayCaptura()
    renderizarCapturaPreview()
    atualizarNivelCaptura()
}

function salvarCaptura() {
    const src = videoCaptura()
    if (!src || !src.videoWidth || !capturaView) {
        CvMsg.aviso('Aguarde o frame carregar antes de baixar.')
        return
    }

    const srcX = Math.max(0, Math.min(src.videoWidth - 1, capturaView.srcX))
    const srcY = Math.max(0, Math.min(src.videoHeight - 1, capturaView.srcY))
    const srcW = Math.max(1, Math.min(src.videoWidth - srcX, capturaView.srcW))
    const srcH = Math.max(1, Math.min(src.videoHeight - srcY, capturaView.srcH))

    const out = document.createElement('canvas')
    out.width = Math.round(srcW)
    out.height = Math.round(srcH)
    const ctx = out.getContext('2d')
    ctx.imageSmoothingEnabled = true
    try {
        ctx.drawImage(src, srcX, srcY, srcW, srcH, 0, 0, out.width, out.height)
    } catch (err) {
        CvMsg.erro('Não foi possível exportar a imagem. Verifique se o vídeo está carregado.')
        return
    }

    const nome = (nomeCameraAtual() || 'camera').replace(/\s+/g, '-').toLowerCase()
    const hora = posicaoAtualMs ? moment(posicaoAtualMs).format('YYYY-MM-DD_HH-mm-ss') : 'frame'
    const link = document.createElement('a')
    link.download = `dvr-${nome}-${hora}.png`
    try {
        link.href = out.toDataURL('image/png')
    } catch (err) {
        CvMsg.erro('Não foi possível exportar a imagem. Tente novamente.')
        return
    }
    link.click()

    if (modalCaptura) modalCaptura.hide()
}

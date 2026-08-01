let cacheCameras = []
let acaoGravacao = 'ativar'
let viewfinderTimers = {}
let viewfinderPopoverCamId = null
let viewfinderPopoverAnchor = null
let viewfinderHls = null

const HLS_VF_OPTS = {
    lowLatencyMode: true,
    enableWorker: true,
    backBufferLength: 0,
    liveBackBufferLength: 0,
    liveSyncDuration: 1,
    liveMaxLatencyDuration: 5,
    maxBufferLength: 8,
    manifestLoadingTimeOut: 10000,
    fragLoadingTimeOut: 15000
}

const CV_PLANOS_GRAVACAO = {
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
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    carregarTudo()
    $('#btn-atualizar-gravacoes').on('click', carregarTudo)
    $('#btn-ativar-gravacao').on('click', executarAcaoGravacao)
    $('#sel-gravacao-camera').on('change', onChangeCameraGravacao)
    $('#sel-gravacao-licenca').on('change', atualizarEstadoAtivar)

    $('#cv-vf-popover-backdrop').on('click', fecharViewfinderPopover)
    $(document).on('keydown', function (e) {
        if (e.key === 'Escape') fecharViewfinderPopover()
    })
    $('#cv-vf-popover').on('click', '.cv-vf-btn-flush', function () {
        const camId = $(this).data('cam-id')
        executarFlush(camId, $(this))
    })
})

function escHtml(valor) {
    if (valor == null || valor === '') return '—'
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function labelPlanoGravacao(plano) {
    return CV_PLANOS_GRAVACAO[plano] || plano || '—'
}

function gravacaoAtiva(cam) {
    return String(cam.gravacao_status || '').toLowerCase() === 'ativa'
}

function cameraTemGravacao(cam) {
    return gravacaoAtiva(cam)
}

function carregarTudo() {
    $.when(carregarCameras(), carregarLicencasGravacao())
        .always(function () {
            renderCamerasGravacao()
            preencherSelectsCameras()
            onChangeCameraGravacao()
            if (viewfinderPopoverCamId) {
                const cam = cacheCameras.find(function (c) {
                    return String(c.id) === String(viewfinderPopoverCamId)
                })
                if (cam && gravacaoAtiva(cam)) {
                    atualizarConteudoPopover(cam)
                } else {
                    fecharViewfinderPopover()
                }
            }
        })
}

function carregarCameras() {
    const idFranq = encodeURIComponent(ConfVisionUrls.idFranqueado())
    return $.get(`/api/cameras?id_franqueado=${idFranq}`)
        .then(function (r) {
            cacheCameras = r.dados || r || []
            return cacheCameras
        })
        .fail(function () {
            cacheCameras = []
        })
}

function carregarLicencasGravacao() {
    const idFranq = encodeURIComponent(ConfVisionUrls.idFranqueado())
    return $.get(`/api/licencas?id_franqueado=${idFranq}&status=disponivel&unidade=gravacao`)
        .then(function (r) {
            const lista = r.dados || r || []
            const sel = $('#sel-gravacao-licenca')
            sel.find('option:not(:first)').remove()
            lista.forEach(function (lic) {
                sel.append(
                    $('<option></option>')
                        .val(lic.id)
                        .text(`#${lic.id} — ${labelPlanoGravacao(lic.plano)}`)
                )
            })
            $('#lbl-gravacao-licenca-vazia').toggleClass('d-none', lista.length > 0)
            atualizarEstadoAtivar()
            return lista
        })
        .fail(function () {
            $('#lbl-gravacao-licenca-vazia').removeClass('d-none')
        })
}

function preencherSelectsCameras() {
    const sel = $('#sel-gravacao-camera')
    const atual = sel.val()
    sel.find('option:not(:first)').remove()
    cacheCameras.forEach(function (cam) {
        const suffix = gravacaoAtiva(cam) ? ' — gravação ativa' : ''
        sel.append(
            $('<option></option>')
                .val(cam.id)
                .text(`${cam.nome || 'Câmera'} (#${cam.id})${suffix}`)
        )
    })
    if (atual) sel.val(atual)
}

function camerasComGravacao() {
    return cacheCameras.filter(cameraTemGravacao)
}

function renderCamerasGravacao() {
    const tbody = $('#tab-gravacao-cameras')
    tbody.empty()
    const lista = camerasComGravacao()
    if (!lista.length) {
        tbody.append('<tr><td colspan="6" class="cv-empty">Nenhuma câmera com gravação configurada.</td></tr>')
        return
    }
    lista.forEach(function (cam) {
        const tr = $('<tr></tr>')
        const ativa = gravacaoAtiva(cam)
        const tdCam = $('<td data-label="Câmera" class="cv-gravacao-cam-cell"></td>')
        tdCam.append($('<span class="cv-gravacao-cam-nome"></span>').text(cam.nome || ('Câmera #' + cam.id)))
        tr.append(`<td data-label="Modo">${ativa ? htmlBadgeTipo(tipoFromCamera(cam)) : '—'}</td>`)
        tr.append(tdCam)
        tr.append(`<td data-label="Retenção">${cam.retencao_dias ? cam.retencao_dias + ' dias' : '—'}</td>`)
        tr.append(`<td data-label="Status">${escHtml(cam.gravacao_status || '—')}</td>`)
        tr.append(`<td data-label="Ativada em">${confVisionFormatarData(cam.gravacao_ativada_em)}</td>`)
        const tdAcoes = $('<td data-label="Ações" class="cv-gravacao-acoes"></td>')

        const linkTimeline = $('<a class="cv-btn-ghost cv-btn-sm">Timeline</a>')
            .attr('href', `/gravacoes/timeline?camera=${encodeURIComponent(cam.id)}`)
        const linkDvr = $('<a class="cv-btn-ghost cv-btn-sm">DVR</a>')
            .attr('href', `/gravacoes/dvr?camera=${encodeURIComponent(cam.id)}`)
        const btnTrocar = $('<button type="button" class="cv-btn-ghost cv-btn-sm">Trocar licença</button>')
        btnTrocar.on('click', function () { prepararTrocarLicenca(cam.id) })
        const btnDesativar = $('<button type="button" class="cv-btn-desativar-gravacao cv-btn-sm">Retirar licença</button>')
        btnDesativar.on('click', function () { desativarGravacao(cam.id) })
        tdAcoes.append(linkTimeline, linkDvr, btnTrocar, btnDesativar)

        if (ativa) {
            const btnRec = $('<button type="button" class="cv-btn-rec-ao-vivo cv-btn-sm"></button>')
                .attr('title', 'Ver transmissão ao vivo (gravando)')
                .attr('aria-label', 'Ver transmissão ao vivo, gravando')
                .attr('data-cam-id', cam.id)
                .append($('<span class="cv-btn-rec-ao-vivo-label"></span>').text('recording...'))
                .append($('<span class="cv-btn-rec-ao-vivo-dot" aria-hidden="true"></span>'))
            if (String(viewfinderPopoverCamId) === String(cam.id)) {
                btnRec.addClass('cv-btn-rec-ao-vivo-open')
            }
            btnRec.on('click', function (e) {
                e.stopPropagation()
                toggleViewfinderPopover(cam, this)
            })
            tdAcoes.append(btnRec)
        }

        tr.append(tdAcoes)
        tbody.append(tr)
    })
}

function prepararAtivarCamera(camId) {
    acaoGravacao = 'ativar'
    $('#sel-gravacao-camera').val(String(camId))
    $('#sel-gravacao-licenca').val('')
    onChangeCameraGravacao()
    document.querySelector('.cv-gravacoes-col-direita')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function prepararTrocarLicenca(camId) {
    acaoGravacao = 'trocar'
    $('#sel-gravacao-camera').val(String(camId))
    $('#sel-gravacao-licenca').val('')
    onChangeCameraGravacao()
    document.querySelector('.cv-gravacoes-col-direita')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function onChangeCameraGravacao() {
    const camId = $('#sel-gravacao-camera').val()
    const cam = cacheCameras.find(function (c) { return String(c.id) === String(camId) })
    if (cam && gravacaoAtiva(cam)) {
        acaoGravacao = 'trocar'
    } else {
        acaoGravacao = 'ativar'
    }
    atualizarEstadoAtivar()
}

function atualizarEstadoAtivar() {
    const camId = $('#sel-gravacao-camera').val()
    const licId = $('#sel-gravacao-licenca').val()
    const cam = cacheCameras.find(function (c) { return String(c.id) === String(camId) })
    const btn = $('#btn-ativar-gravacao')
    const hint = $('#lbl-gravacao-acao-hint')

    if (cam && gravacaoAtiva(cam)) {
        btn.text('Trocar licença')
        hint.text('Selecione a nova licença disponível e confirme a troca. A licença atual será liberada.')
    } else {
        btn.text('Ativar gravação')
        hint.text('Associe uma licença de gravação disponível à câmera selecionada.')
    }

    const bloqueado = !camId || !licId
    btn.prop('disabled', bloqueado)
}

function executarAcaoGravacao() {
    if (acaoGravacao === 'trocar') {
        trocarLicencaGravacao()
        return
    }
    ativarGravacao()
}

function ativarGravacao() {
    const camId = $('#sel-gravacao-camera').val()
    const licId = $('#sel-gravacao-licenca').val()
    if (!camId || !licId) return
    postAtivarGravacao(camId, licId).done(function () {
        CvMsg.sucesso('Gravação ativada.')
        acaoGravacao = 'ativar'
        carregarTudo()
    }).fail(function (xhr) {
        CvMsg.erro(erroApi(xhr) || 'Erro ao ativar gravação.')
    })
}

function trocarLicencaGravacao() {
    const camId = $('#sel-gravacao-camera').val()
    const licId = $('#sel-gravacao-licenca').val()
    if (!camId || !licId) return
    CvMsg.confirmar('Trocar licença?', 'A licença atual voltará para disponível.').then(function (r) {
        if (!r.isConfirmed) return
        postAtivarGravacao(camId, licId).done(function () {
            CvMsg.sucesso('Licença trocada com sucesso.')
            acaoGravacao = 'ativar'
            carregarTudo()
        }).fail(function (xhr) {
            CvMsg.erro(erroApi(xhr) || 'Erro ao trocar licença de gravação.')
            carregarTudo()
        })
    })
}

function postAtivarGravacao(camId, licId) {
    return $.ajax({
        url: `/api/cameras/${camId}/gravacao/ativar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ vis_licenca_gravacao_id: parseInt(licId, 10) })
    })
}

function desativarGravacao(camId, silencioso) {
    function executar() {
        return $.ajax({
            url: `/api/cameras/${camId}/gravacao/desativar`,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({})
        }).done(function () {
            if (!silencioso) {
                CvMsg.sucesso('Licença de gravação retirada.')
                fecharViewfinderPopover()
                carregarTudo()
            }
        }).fail(function (xhr) {
            if (!silencioso) {
                CvMsg.erro(erroApi(xhr) || 'Erro ao retirar licença de gravação.')
            }
        })
    }
    if (silencioso) return executar()
    return CvMsg.confirmar('Retirar licença?', 'A licença de gravação voltará para disponível.').then(function (r) {
        if (!r.isConfirmed) return $.Deferred().reject().promise()
        return executar()
    })
}

// ------------------------------------------------------------------ //
// Viewfinder popover — ao vivo HLS ao clicar no ícone REC              //
// ------------------------------------------------------------------ //

function pararTimerViewfinder(camId) {
    const t = viewfinderTimers[camId]
    if (t && t.intervalId) {
        clearInterval(t.intervalId)
    }
    delete viewfinderTimers[camId]
}

function pararTodosTimersViewfinder() {
    Object.keys(viewfinderTimers).forEach(pararTimerViewfinder)
}

function pararHlsViewfinder() {
    if (viewfinderHls) {
        viewfinderHls.destroy()
        viewfinderHls = null
    }
    const video = document.querySelector('#cv-vf-popover .cv-vf-video')
    if (video) {
        video.onplaying = null
        video.onerror = null
        video.removeAttribute('src')
        video.load()
    }
}

function mostrarLoadingViewfinder(mostrar) {
    const $loading = $('#cv-vf-popover .cv-vf-loading')
    if (mostrar) {
        $loading.removeClass('d-none')
    } else {
        $loading.addClass('d-none')
    }
}

function iniciarHlsViewfinder(camId) {
    const video = document.querySelector('#cv-vf-popover .cv-vf-video')
    if (!video) return

    pararHlsViewfinder()
    mostrarLoadingViewfinder(true)

    const url = ConfVisionUrls.hlsUrl(camId)
    video.muted = true
    video.playsInline = true
    video.onplaying = function () {
        mostrarLoadingViewfinder(false)
    }
    video.onerror = function () {
        mostrarLoadingViewfinder(true)
    }

    if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = url + '?t=' + Date.now()
        video.play().catch(function () { /* aguarda retry do usuário */ })
        return
    }

    if (typeof Hls === 'undefined' || !Hls.isSupported()) {
        return
    }

    viewfinderHls = new Hls(HLS_VF_OPTS)
    viewfinderHls.attachMedia(video)
    viewfinderHls.on(Hls.Events.MANIFEST_PARSED, function () {
        video.play().catch(function () { /* ignore */ })
    })
    viewfinderHls.on(Hls.Events.ERROR, function (_, data) {
        if (data.fatal) {
            mostrarLoadingViewfinder(true)
            if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
                viewfinderHls.recoverMediaError()
            }
        }
    })
    viewfinderHls.loadSource(url + '?t=' + Date.now())
}

function htmlViewfinderCard(cam) {
    const camId = cam.id
    const tipo = tipoFromCamera(cam)
    const podeFlush = tipo === 'continua' || tipo === 'timelapse'
    const modoLabel = labelTipoGravacao(tipo)

    return `
        <div class="cv-viewfinder-card" data-cam-id="${camId}">
            <div class="cv-vf-frame">
                <video class="cv-vf-video" muted playsinline autoplay></video>
                <div class="cv-vf-loading">
                    <i class="bi bi-arrow-repeat cv-vf-loading-spin"></i>
                    <span>Conectando ao vivo…</span>
                </div>
                <div class="cv-vf-overlay">
                    <div class="cv-vf-corner cv-vf-tl"></div>
                    <div class="cv-vf-corner cv-vf-tr"></div>
                    <div class="cv-vf-corner cv-vf-bl"></div>
                    <div class="cv-vf-corner cv-vf-br"></div>
                    <div class="cv-vf-rec"><span class="cv-vf-rec-dot"></span> REC</div>
                    <div class="cv-vf-info-tl">${escHtml(cam.nome || 'Câmera #' + camId)}</div>
                    <div class="cv-vf-info-tr">${modoLabel}</div>
                    <div class="cv-vf-timer" id="vf-timer-${camId}">00:00:00</div>
                    <div class="cv-vf-info-bl">${escHtml((typeof CvWhitelabel !== 'undefined' && CvWhitelabel.brandName) ? CvWhitelabel.brandName() : ((localStorage.getItem('nomeFranqueado') || '').trim() || 'Central de Câmeras'))}</div>
                </div>
            </div>
            ${podeFlush ? `<button type="button" class="cv-btn-primary cv-vf-btn-flush w-100 mt-2" data-cam-id="${camId}">
                <i class="bi bi-stop-circle"></i> Parar e enviar agora
            </button>` : ''}
        </div>
    `
}

function iniciarTimerViewfinder(camId) {
    pararTimerViewfinder(camId)
    const startMs = Date.now()
    const intervalId = setInterval(function () {
        if (String(viewfinderPopoverCamId) !== String(camId)) return
        const elapsed = Math.floor((Date.now() - startMs) / 1000)
        const h = String(Math.floor(elapsed / 3600)).padStart(2, '0')
        const m = String(Math.floor((elapsed % 3600) / 60)).padStart(2, '0')
        const s = String(elapsed % 60).padStart(2, '0')
        const el = document.getElementById('vf-timer-' + camId)
        if (el) el.textContent = `${h}:${m}:${s}`
    }, 1000)
    viewfinderTimers[camId] = { intervalId: intervalId, startMs: startMs }
}

function posicionarViewfinderPopover(anchor) {
    const pop = $('#cv-vf-popover')
    pop.removeClass('d-none')
    const popEl = pop[0]
    const rect = anchor.getBoundingClientRect()
    const popW = popEl.offsetWidth || 300
    const popH = popEl.offsetHeight || 220
    let top = rect.bottom + 8
    let left = rect.left

    if (left + popW > window.innerWidth - 12) {
        left = window.innerWidth - popW - 12
    }
    if (left < 12) left = 12

    if (top + popH > window.innerHeight - 12) {
        top = rect.top - popH - 8
    }
    if (top < 12) top = 12

    pop.css({ top: top + 'px', left: left + 'px' })
}

function atualizarConteudoPopover(cam) {
    const pop = $('#cv-vf-popover')
    pop.html(htmlViewfinderCard(cam))
    iniciarTimerViewfinder(cam.id)
    iniciarHlsViewfinder(cam.id)
    if (viewfinderPopoverAnchor) {
        posicionarViewfinderPopover(viewfinderPopoverAnchor)
    }
}

function abrirViewfinderPopover(cam, anchor) {
    viewfinderPopoverCamId = cam.id
    viewfinderPopoverAnchor = anchor
    pararTodosTimersViewfinder()

    $('.cv-btn-rec-ao-vivo').removeClass('cv-btn-rec-ao-vivo-open')
    $(anchor).addClass('cv-btn-rec-ao-vivo-open')

    $('#cv-vf-popover-backdrop').removeClass('d-none')
    atualizarConteudoPopover(cam)
}

function fecharViewfinderPopover() {
    viewfinderPopoverCamId = null
    viewfinderPopoverAnchor = null
    pararTodosTimersViewfinder()
    pararHlsViewfinder()
    $('#cv-vf-popover').addClass('d-none').empty()
    $('#cv-vf-popover-backdrop').addClass('d-none')
    $('.cv-btn-rec-ao-vivo').removeClass('cv-btn-rec-ao-vivo-open')
}

function toggleViewfinderPopover(cam, anchor) {
    if (String(viewfinderPopoverCamId) === String(cam.id)) {
        fecharViewfinderPopover()
        return
    }
    fecharViewfinderPopover()
    abrirViewfinderPopover(cam, anchor)
}

function executarFlush(camId, btn) {
    CvMsg.confirmar('Parar e enviar agora?', 'A gravação atual será encerrada e o segmento enviado imediatamente.').then(function (r) {
        if (!r.isConfirmed) return
        btn.prop('disabled', true).html('<i class="bi bi-hourglass-split"></i> Enviando...')
        $.ajax({
            url: `/api/cameras/${camId}/gravacao/flush`,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({})
        }).done(function () {
            btn.html('<i class="bi bi-check-circle"></i> Sinal enviado!')
            setTimeout(function () {
                btn.prop('disabled', false).html('<i class="bi bi-stop-circle"></i> Parar e enviar agora')
            }, 4000)
        }).fail(function (xhr) {
            btn.prop('disabled', false).html('<i class="bi bi-stop-circle"></i> Parar e enviar agora')
            CvMsg.erro(erroApi(xhr) || 'Erro ao enviar sinal de flush.')
        })
    })
}

function erroApi(xhr) {
    if (!xhr) return ''
    const body = xhr.responseJSON || {}
    if (body.message) return body.message
    if (body.error) return body.error
    if (typeof body === 'string') return body
    if (xhr.responseText) {
        try {
            const parsed = JSON.parse(xhr.responseText)
            return parsed.message || parsed.error || xhr.responseText
        } catch (e) {
            return xhr.responseText.substring(0, 200)
        }
    }
    return xhr.status ? ('HTTP ' + xhr.status) : ''
}

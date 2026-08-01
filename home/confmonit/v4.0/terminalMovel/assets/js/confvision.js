var confvision_hlsBase = ''

var confvision_painelState = {
    ctx: null,
    camera: null,
    lista: [],
    clipsCache: {},
    offset: 0,
    limit: 25,
    total: 0,
    focusEventoId: null,
    filtro: 'dispositivo'
}

function confvision_ensureModals(cb) {
    if (document.getElementById('confvisionModalSetor')) {
        cb()
        return
    }
    var tries = 0
    var timer = setInterval(function () {
        tries++
        if (document.getElementById('confvisionModalSetor') || tries > 40) {
            clearInterval(timer)
            cb()
        }
    }, 100)
}

function confvision_carregarConfig() {
    return $.ajax({
        url: 'ateEventosDetalhe/confVisionConfig',
        method: 'GET',
        headers: {
            Authorization: 'Bearer ' + sessionStorage.getItem('token')
        }
    }).done(function (r) {
        confvision_hlsBase = (r && r.dados && r.dados.hlsBase) ? r.dados.hlsBase.replace(/\/$/, '') : ''
    }).fail(function () {
        confvision_hlsBase = ''
    })
}

function confvision_usaNoProcesso() {
    return sessionStorage.getItem('ateProDado_proc_usaConfVision') === 'S'
}

function confvision_eConfVision(item) {
    if (!item) return false
    const provedor = String(item.provedorVideo || sessionStorage.getItem('ateProDado_proc_provedorVideo') || '').toLowerCase()
    const usa = item.usaConfVision || sessionStorage.getItem('ateProDado_proc_usaConfVision') || 'N'
    return provedor === 'confvision' && usa === 'S'
}

function confvision_hlsUrl(cameraId) {
    if (!cameraId || !confvision_hlsBase) return ''
    return confvision_hlsBase + '/live/' + cameraId + '/index.m3u8'
}

function confvision_extrairClipsPayload(r) {
    let payload = (r && r.dados) ? r.dados : r
    if (payload && payload.dados && (payload.dados.evento || payload.dados.clips)) {
        payload = payload.dados
    }
    return {
        evento: payload.evento || {},
        clips: Array.isArray(payload.clips) ? payload.clips : []
    }
}

function confvision_extrairPorSetorPayload(r) {
    let payload = (r && r.dados) ? r.dados : r
    if (payload && payload.dados && (payload.dados.dados !== undefined || payload.dados.camera !== undefined)) {
        payload = payload.dados
    }
    return {
        evento: payload.dados || payload.evento || null,
        camera: payload.camera || null
    }
}

function confvision_temMidia(ev) {
    if (!ev) return false
    return !!(ev.snapshot_url || ev.video_url || (ev.clip_count && ev.clip_count > 0))
}

function confvision_temFoto(ev) {
    if (!ev) return false
    return !!ev.snapshot_url
}

function confvision_temVideo(ev) {
    if (!ev) return false
    return !!(ev.video_url || (ev.clip_count && ev.clip_count > 0))
}

function confvision_formatarData(valor) {
    if (!valor) return '—'
    const d = new Date(valor)
    if (isNaN(d.getTime())) return String(valor)
    return d.toLocaleString('pt-BR')
}

function confvision_normalizarUltimos25Setor(r) {
    let payload = (r && r.dados) ? r.dados : r
    if (payload && payload.dados && (payload.dados.dados !== undefined || payload.dados.camera !== undefined)) {
        payload = payload.dados
    }

    let lista = payload.dados
    if (lista && lista.items && Array.isArray(lista.items)) {
        lista = lista.items
    }
    if (!Array.isArray(lista)) {
        lista = Array.isArray(payload) ? payload : []
    }

    return {
        lista: lista,
        camera: payload.camera || null,
        total: lista.length,
        offset: 0,
        limit: 25
    }
}

function confvision_normalizarUltimos25Dispositivo(r) {
    let payload = (r && r.dados) ? r.dados : r
    if (payload && payload.dados && (payload.dados.dados !== undefined || payload.dados.total !== undefined)) {
        payload = payload.dados
    }

    let lista = payload.dados
    if (lista && lista.items && Array.isArray(lista.items)) {
        lista = lista.items
    }
    if (!Array.isArray(lista)) {
        lista = []
    }

    return {
        lista: lista,
        total: payload.total != null ? payload.total : lista.length,
        offset: payload.offset != null ? payload.offset : 0,
        limit: payload.limit != null ? payload.limit : 25
    }
}

function confvision_buscarUltimos25Dispositivo(ctx, offset, limit) {
    return $.ajax({
        url: 'ateEventosDetalhe/buscarConfVisionUltimos25Dispositivo',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idDispositivo: ctx.idDispositivo || sessionStorage.getItem('ateProDado_proc_idDispositivo') || '',
            offset: offset || 0,
            limit: limit || confvision_painelState.limit || 25
        })
    })
}

function confvision_ctxTitulo(ctx) {
    ctx = ctx || {}
    const setor = ctx.zonaUser || ctx.setor || '—'
    const part = ctx.particao || '—'
    let titulo = confvision_painelState.filtro === 'setor'
        ? ('ConfVision — Setor ' + setor + ' / Part. ' + part)
        : ('ConfVision — Dispositivo · Setor ' + setor + ' / Part. ' + part)
    const cam = confvision_painelState.camera
    if (cam && cam.id) {
        titulo += ' / Câm. ' + cam.id
        if (cam.nome) titulo += ' (' + cam.nome + ')'
    }
    return titulo
}

function confvision_resolverCamera(ctx, cb) {
    confvision_buscarPorSetor(ctx).done(function (r) {
        const payload = confvision_extrairPorSetorPayload(r)
        confvision_painelState.camera = payload.camera || null
        cb(payload)
    }).fail(function () {
        cb({ evento: null, camera: null })
    })
}

function confvision_parseImg(img) {
    if (!img || img === 'SEM IMAGEM') return null
    try {
        const o = JSON.parse(img)
        if (o && o.confvision && o.vis_evento_id) {
            return { visEventoId: o.vis_evento_id, isConfvision: true }
        }
    } catch (e) {}
    return null
}

function confvision_buscarUltimos25Setor(ctx) {
    return $.ajax({
        url: 'ateEventosDetalhe/buscarConfVisionUltimos25Setor',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idDispositivo: ctx.idDispositivo || sessionStorage.getItem('ateProDado_proc_idDispositivo') || '',
            particao: ctx.particao || '',
            zonaUser: ctx.zonaUser || ctx.setor || ''
        })
    })
}

function confvision_buscarPorSetor(ctx) {
    return $.ajax({
        url: 'ateEventosDetalhe/buscarConfVisionPorSetor',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idProcesso: ctx.idProcesso || sessionStorage.getItem('ateProDado_proc_idProcesso') || '',
            idDispositivo: ctx.idDispositivo || sessionStorage.getItem('ateProDado_proc_idDispositivo') || '',
            particao: ctx.particao || '',
            zonaUser: ctx.zonaUser || ctx.setor || '',
            idEvento: ctx.idEvento || ''
        })
    })
}

function confvision_escHtml(valor) {
    return String(valor || '')
        .replace(/&/g, '&amp;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
}

function confvision_fecharModalSetor(cb) {
    const el = document.getElementById('confvisionModalSetor')
    if (!el) {
        if (cb) cb()
        return
    }
    const inst = bootstrap.Modal.getInstance(el)
    if (inst) {
        el.addEventListener('hidden.bs.modal', function onHide() {
            el.removeEventListener('hidden.bs.modal', onHide)
            if (cb) cb()
        })
        inst.hide()
    } else if (cb) {
        cb()
    }
}

function confvision_fecharModalSetorAgora() {
    const el = document.getElementById('confvisionModalSetor')
    if (!el) return
    const inst = bootstrap.Modal.getInstance(el)
    if (inst) {
        inst.hide()
    }
}

function confvision_tentarPlayVideo(video) {
    if (!video) return
    video.muted = true
    video.setAttribute('playsinline', '')
    video.setAttribute('webkit-playsinline', '')
    const prom = video.play()
    if (prom && typeof prom.catch === 'function') {
        prom.catch(function () {
            confvision_mostrarErroAoVivo('Toque em ▶ no vídeo para iniciar o ao vivo.')
        })
    }
}

function confvision_exibirSnapshotNoModal(snapshotUrl) {
    const $img = $('#confvisionModalSnapshot')
    $img.addClass('d-none').off('load error').attr('src', '')

    if (!snapshotUrl) {
        return false
    }

    $img
        .on('load', function () {
            $img.removeClass('d-none')
            $('#confvisionModalClips .cv-snapshot-hint').remove()
        })
        .on('error', function () {
            $img.addClass('d-none')
            if (!$('#confvisionModalClips .cv-snapshot-erro').length) {
                $('#confvisionModalClips').prepend(
                    `<div class="cv-snapshot-erro alert alert-warning py-2">
                        Não foi possível carregar a foto.
                        <a href="${confvision_escHtml(snapshotUrl)}" target="_blank" rel="noopener">Abrir em nova aba</a>
                    </div>`
                )
            }
        })
        .attr('src', snapshotUrl + (snapshotUrl.indexOf('?') >= 0 ? '&' : '?') + 't=' + Date.now())

    return true
}

function confvision_abrirMidia(eventoId, modo) {
    if (!eventoId) return

    modo = modo || 'tudo'
    const tituloFoto = 'ConfVision — Foto #' + eventoId
    const tituloVideo = 'ConfVision — Vídeo #' + eventoId
    $('#confvisionModalMidiaTitulo').text(modo === 'video' ? tituloVideo : tituloFoto)
    $('#confvisionModalSnapshot').addClass('d-none').attr('src', '')
    $('#confvisionModalClips').empty().append('<div class="txt-tbody py-2">Carregando...</div>')

    const modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('confvisionModalMidia'))
    modal.show()

    function renderMidia(evento, clips) {
        evento = evento || {}
        clips = clips || []
        $('#confvisionModalClips').empty()
        $('#confvisionModalSnapshot').addClass('d-none').attr('src', '')

        if (modo === 'foto') {
            if (confvision_exibirSnapshotNoModal(evento.snapshot_url)) {
                return
            }
            if (String(evento.status || '').toLowerCase() === 'capturando') {
                $('#confvisionModalClips').append('<div class="txt-tbody py-2">Captura em andamento…</div>')
            } else {
                $('#confvisionModalClips').append('<div class="txt-tbody py-2">Foto indisponível.</div>')
            }
            return
        }

        if (modo === 'video') {
            let videoUrl = evento.video_url
            if (!videoUrl && clips.length && clips[0].video_url) {
                videoUrl = clips[0].video_url
            }
            if (videoUrl) {
                $('#confvisionModalClips').append(`<video controls playsinline class="w-100" src="${confvision_escHtml(videoUrl)}"></video>`)
            } else {
                $('#confvisionModalClips').append('<div class="txt-tbody py-2">Vídeo indisponível (licença somente foto).</div>')
            }
            return
        }

        const temSnapshot = confvision_exibirSnapshotNoModal(evento.snapshot_url)
        if (clips.length) {
            clips.forEach(function (clip) {
                if (clip.video_url) {
                    $('#confvisionModalClips').append(`<video controls playsinline class="w-100 mb-2" src="${confvision_escHtml(clip.video_url)}"></video>`)
                }
            })
        } else if (evento.video_url) {
            $('#confvisionModalClips').append(`<video controls playsinline class="w-100" src="${confvision_escHtml(evento.video_url)}"></video>`)
        } else if (temSnapshot) {
            $('#confvisionModalClips').append('<div class="cv-snapshot-hint txt-tbody py-2">Somente foto (sem vídeo).</div>')
        } else if (String(evento.status || '').toLowerCase() === 'capturando') {
            $('#confvisionModalClips').append('<div class="txt-tbody py-2">Captura em andamento…</div>')
        } else {
            $('#confvisionModalClips').append('<div class="txt-tbody py-2">Sem mídia disponível.</div>')
        }
    }

    $.ajax({
        url: 'ateEventosDetalhe/buscarConfVisionClips',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ eventoId: eventoId })
    }).done(function (r) {
        const parsed = confvision_extrairClipsPayload(r)
        renderMidia(parsed.evento, parsed.clips)
    }).fail(function () {
        $('#confvisionModalClips').html('<div class="text-danger py-2">Erro ao carregar mídia</div>')
    })
}

function confvision_abrirMidiaFoto(eventoId) {
    confvision_abrirMidia(eventoId, 'foto')
}

function confvision_abrirMidiaVideo(eventoId) {
    confvision_abrirMidia(eventoId, 'video')
}

var confvision_hlsInstance = null

function confvision_pararAoVivo() {
    const video = document.getElementById('confvisionModalLiveVideo')
    if (confvision_hlsInstance) {
        confvision_hlsInstance.destroy()
        confvision_hlsInstance = null
    }
    if (video) {
        video.removeAttribute('src')
        video.load()
    }
    $('#confvisionModalLiveErro').addClass('d-none').text('')
}

function confvision_mostrarErroAoVivo(msg) {
    $('#confvisionModalLiveErro').removeClass('d-none').text(msg)
}

function confvision_abrirAoVivo(cameraId, fecharSetor) {
    function executar() {
        const hlsUrl = confvision_hlsUrl(cameraId)
        if (!hlsUrl) {
            alert('Ao vivo indisponível (HLS não configurado no servidor)')
            return
        }

        confvision_ensureModals(function () {
            const liveEl = document.getElementById('confvisionModalLive')
            const video = document.getElementById('confvisionModalLiveVideo')
            if (!liveEl || !video) {
                alert('Modal ao vivo não carregado. Recarregue a página.')
                return
            }

            $('#confvisionModalLiveTitulo').text('Ao vivo — câmera ' + cameraId)
            confvision_pararAoVivo()

            const modal = bootstrap.Modal.getOrCreateInstance(liveEl)
            modal.show()

            if (window.Hls && Hls.isSupported()) {
                confvision_hlsInstance = new Hls({ lowLatencyMode: true })
                confvision_hlsInstance.loadSource(hlsUrl)
                confvision_hlsInstance.attachMedia(video)
                confvision_hlsInstance.on(Hls.Events.MANIFEST_PARSED, function () {
                    confvision_tentarPlayVideo(video)
                })
                confvision_hlsInstance.on(Hls.Events.ERROR, function (ev, data) {
                    if (data.fatal) {
                        confvision_mostrarErroAoVivo('Não foi possível reproduzir ao vivo. Verifique se o stream HLS está ativo.')
                        console.log('HLS error', data)
                    }
                })
            } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
                video.src = hlsUrl
                video.onerror = function () {
                    confvision_mostrarErroAoVivo('Erro ao carregar stream ao vivo.')
                }
                confvision_tentarPlayVideo(video)
            } else {
                alert('Seu navegador não suporta HLS')
            }

            liveEl.addEventListener('hidden.bs.modal', confvision_pararAoVivo, { once: true })

            if (fecharSetor !== false) {
                confvision_fecharModalSetorAgora()
            }
        })
    }

    if (confvision_hlsUrl(cameraId)) {
        executar()
        return
    }

    confvision_carregarConfig().always(function () {
        executar()
    })
}

function confvision_abrirAoVivoCameraPainel() {
    function abrirComCamera(cameraId) {
        if (cameraId) {
            confvision_abrirAoVivo(cameraId, true)
        } else {
            alert('Câmera ao vivo não encontrada para este setor.')
        }
    }

    function resolverEAbrir() {
        const cam = confvision_painelState.camera
        if (cam && cam.id) {
            abrirComCamera(cam.id)
            return
        }

        const ctx = confvision_painelState.ctx || {}
        confvision_buscarPorSetor(ctx).done(function (r) {
            const payload = confvision_extrairPorSetorPayload(r)
            confvision_painelState.camera = payload.camera || null
            const cameraId = (payload.camera && payload.camera.id) ||
                (payload.evento && payload.evento.vis_camera_id)
            abrirComCamera(cameraId)
        }).fail(function () {
            alert('Não foi possível localizar a câmera ao vivo.')
        })
    }

    if (confvision_hlsBase) {
        resolverEAbrir()
        return
    }

    confvision_carregarConfig().always(resolverEAbrir)
}

function confvision_painelCarregarClips(ev, cb) {
    if (!ev || !ev.id) {
        cb(ev, [])
        return
    }
    if (confvision_painelState.clipsCache[ev.id]) {
        cb(confvision_painelState.clipsCache[ev.id].evento, confvision_painelState.clipsCache[ev.id].clips)
        return
    }
    $.ajax({
        url: 'ateEventosDetalhe/buscarConfVisionClips',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ eventoId: ev.id })
    }).done(function (r) {
        const parsed = confvision_extrairClipsPayload(r)
        const evento = parsed.evento && parsed.evento.id ? parsed.evento : ev
        confvision_painelState.clipsCache[ev.id] = { evento: evento, clips: parsed.clips }
        cb(evento, parsed.clips)
    }).fail(function () {
        cb(ev, [])
    })
}

function confvision_painelIconeFoto(ev) {
    if (!confvision_temFoto(ev)) {
        return '<span class="text-muted">—</span>'
    }
    return `<i class="bi bi-image-fill text-primary cv-ico-foto" title="Ver foto"></i>`
}

function confvision_painelIconeVideo(ev) {
    if (!confvision_temVideo(ev)) {
        return ''
    }
    return `<i class="bi bi-play-circle-fill text-info click cv-ico-video" data-evento-id="${ev.id}" title="Ver vídeo"></i>`
}

function confvision_painelIconeLive() {
    return ''
}

function confvision_painelAtualizarToolbar() {
    const isSetor = confvision_painelState.filtro === 'setor'
    $('#confvisionModalSetorBtnMais').toggleClass('d-none', isSetor)
}

function confvision_painelBindToolbar() {
    $('#confvisionModalSetorBtnLiveTop').off('click').on('click', function () {
        confvision_abrirAoVivoCameraPainel()
    })
    $('#confvisionModalSetorBtnRefresh').off('click').on('click', function () {
        confvision_painelCarregar(true)
    })
    $('#confvisionModalSetorBtnMais').off('click').on('click', function () {
        confvision_painelCarregar(false)
    })
    confvision_painelAtualizarToolbar()
}

function confvision_painelBindIcones() {
    $('#confvisionModalSetorTabela').off('click', '.cv-cell-foto').on('click', '.cv-cell-foto', function (e) {
        e.stopPropagation()
        confvision_abrirMidiaFoto($(this).data('evento-id'))
    })
    $('#confvisionModalSetorTabela').off('click', '.cv-ico-video').on('click', '.cv-ico-video', function (e) {
        e.stopPropagation()
        confvision_abrirMidiaVideo($(this).data('evento-id'))
    })
}

function confvision_painelAtualizarInfoLista() {
    const total = confvision_painelState.total || confvision_painelState.lista.length
    const carregados = confvision_painelState.lista.length
    const ctx = confvision_painelState.ctx || {}
    const setor = ctx.zonaUser || ctx.setor || '—'
    const part = ctx.particao || '—'
    let info = confvision_painelState.filtro === 'setor'
        ? ('Eventos ConfVision — dispositivo / part. ' + part + ' / setor ' + setor + ' — ' + carregados)
        : ('Eventos ConfVision do dispositivo — ' + carregados + ' de ' + total)
    if (confvision_painelState.focusEventoId) {
        info += ' · destaque no evento #' + confvision_painelState.focusEventoId
    }
    $('#confvisionModalSetorInfo').text(info)
    const podeMais = confvision_painelState.filtro !== 'setor' && carregados < total
    $('#confvisionModalSetorBtnMais').prop('disabled', !podeMais)
}

function confvision_painelRenderGrid(lista) {
    const tbody = $('#confvisionModalSetorTabela tbody')
    tbody.empty()

    let algumVideo = false
    lista.forEach(function (ev) {
        if (confvision_temVideo(ev)) algumVideo = true
    })
    if (algumVideo) {
        $('#confvisionModalSetorTabela .cv-col-video').show()
    } else {
        $('#confvisionModalSetorTabela .cv-col-video').hide()
    }

    if (!lista.length) {
        tbody.append('<tr><td colspan="7" class="text-center txt-tbody py-3">Nenhum evento ConfVision</td></tr>')
        return
    }

    const focusId = confvision_painelState.focusEventoId

    lista.forEach(function (ev, idx) {
        const videoCell = algumVideo
            ? `<td class="text-center">${confvision_painelIconeVideo(ev) || '<span class="text-muted">—</span>'}</td>`
            : ''
        const focusCls = focusId && String(ev.id) === String(focusId) ? ' cv-painel-row-focus' : ''
        const fotoCls = confvision_temFoto(ev) ? 'text-center click cv-cell-foto' : 'text-center'
        const fotoAttr = confvision_temFoto(ev) ? ` data-evento-id="${ev.id}" title="Ver foto"` : ''

        tbody.append(`
            <tr data-id="${ev.id}" class="${focusCls.trim()}">
                <td class="fmt-tbody">${idx + 1}</td>
                <td class="fmt-tbody">${confvision_formatarData(ev.created_at)}</td>
                <td class="fmt-tbody">${confvision_escHtml(ev.zonauser || ev.canal || '—')}</td>
                <td class="fmt-tbody">${confvision_escHtml(ev.particao || '—')}</td>
                <td class="fmt-tbody">${confvision_escHtml(ev.status || '—')}</td>
                <td class="${fotoCls}"${fotoAttr}>${confvision_painelIconeFoto(ev)}</td>
                ${videoCell}
            </tr>
        `)
    })

    confvision_painelBindIcones()
    confvision_painelAtualizarInfoLista()

    lista.forEach(function (ev) {
        if (!confvision_temVideo(ev)) return
        confvision_painelCarregarClips(ev, function (evento) {
            const idx = lista.findIndex(function (e) { return String(e.id) === String(evento.id) })
            if (idx >= 0) {
                lista[idx] = evento
            }
            if (confvision_temVideo(evento)) {
                const $row = tbody.find('tr[data-id="' + evento.id + '"]')
                if ($row.length && algumVideo) {
                    $row.find('td').eq(6).html(confvision_painelIconeVideo(evento))
                }
            }
            confvision_painelBindIcones()
        })
    })
}

function confvision_painelAbrir(titulo, info) {
    confvision_ensureModals(function () {
        $('#confvisionModalSetorTitulo').text(titulo)
        $('#confvisionModalSetorInfo').text(info || 'Carregando…')
        $('#confvisionModalSetorTabela tbody').html('<tr><td colspan="7" class="text-center txt-tbody py-3">Carregando...</td></tr>')

        const modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('confvisionModalSetor'))
        modal.show()

        confvision_painelState.clipsCache = {}
        confvision_painelBindToolbar()
    })
}

function confvision_painelCarregarSetor(reset) {
    const ctx = confvision_painelState.ctx || {}
    if (reset) {
        confvision_painelState.offset = 0
        confvision_painelState.lista = []
        $('#confvisionModalSetorTabela tbody').html('<tr><td colspan="7" class="text-center txt-tbody py-3">Carregando...</td></tr>')
    }

    confvision_buscarUltimos25Setor(ctx)
        .done(function (r) {
            const parsed = confvision_normalizarUltimos25Setor(r)
            confvision_painelState.lista = parsed.lista
            confvision_painelState.total = parsed.total
            confvision_painelState.camera = parsed.camera || confvision_painelState.camera
            $('#confvisionModalSetorTitulo').text(confvision_ctxTitulo(ctx))
            confvision_painelRenderGrid(confvision_painelState.lista)
        })
        .fail(function () {
            $('#confvisionModalSetorTabela tbody').html('<tr><td colspan="7" class="text-danger text-center py-3">Erro ao carregar eventos do setor</td></tr>')
        })
}

function confvision_painelCarregarDispositivo(reset) {
    const ctx = confvision_painelState.ctx || {}
    if (reset) {
        confvision_painelState.offset = 0
        confvision_painelState.lista = []
        $('#confvisionModalSetorTabela tbody').html('<tr><td colspan="7" class="text-center txt-tbody py-3">Carregando...</td></tr>')
    } else {
        confvision_painelState.offset = confvision_painelState.lista.length
    }

    confvision_buscarUltimos25Dispositivo(ctx, confvision_painelState.offset, confvision_painelState.limit)
        .done(function (r) {
            const parsed = confvision_normalizarUltimos25Dispositivo(r)
            confvision_painelState.total = parsed.total
            if (reset) {
                confvision_painelState.lista = parsed.lista
            } else {
                confvision_painelState.lista = confvision_painelState.lista.concat(parsed.lista)
            }
            confvision_painelRenderGrid(confvision_painelState.lista)
        })
        .fail(function () {
            $('#confvisionModalSetorTabela tbody').html('<tr><td colspan="7" class="text-danger text-center py-3">Erro ao carregar eventos</td></tr>')
        })
}

function confvision_painelCarregar(reset) {
    if (confvision_painelState.filtro === 'setor') {
        confvision_painelCarregarSetor(reset)
    } else {
        confvision_painelCarregarDispositivo(reset)
    }
}

function confvision_abrirPainel(ctx, opts) {
    ctx = ctx || {}
    opts = opts || {}

    confvision_painelState.ctx = ctx
    confvision_painelState.filtro = opts.filtro === 'setor' ? 'setor' : 'dispositivo'
    confvision_painelState.offset = 0
    confvision_painelState.lista = []
    confvision_painelState.total = 0
    confvision_painelState.camera = null
    confvision_painelState.focusEventoId = opts.focusEventoId || null

    const parsedImg = confvision_parseImg(ctx.img)
    if (!confvision_painelState.focusEventoId && parsedImg && parsedImg.visEventoId) {
        confvision_painelState.focusEventoId = parsedImg.visEventoId
    }

    const infoLoading = confvision_painelState.filtro === 'setor'
        ? 'Carregando eventos deste setor…'
        : 'Carregando eventos do dispositivo…'

    confvision_painelAbrir(confvision_ctxTitulo(ctx), infoLoading)

    if (confvision_painelState.filtro === 'setor') {
        confvision_painelCarregarSetor(true)
        return
    }

    confvision_resolverCamera(ctx, function () {
        $('#confvisionModalSetorTitulo').text(confvision_ctxTitulo(ctx))
        confvision_painelCarregarDispositivo(true)
    })
}

function confvision_abrirPainelSetor(ctx) {
    confvision_abrirPainel(ctx, { filtro: 'dispositivo' })
}

function confvision_abrirPainelPorSetor(ctx, opts) {
    confvision_abrirPainel(ctx, Object.assign({ filtro: 'setor' }, opts || {}))
}

function confvision_visualizarCameraBenuvem(item) {
    benuvem_visualizar(item.codBenuvem, item.particao, item.conta, item.zonaUser)
}

function confvision_itemFromCell($el) {
    return {
        idProcesso: $el.attr('idProcesso') || sessionStorage.getItem('ateProDado_proc_idProcesso') || '',
        idDispositivo: $el.attr('idDispositivo') || sessionStorage.getItem('ateProDado_proc_idDispositivo') || '',
        particao: $el.attr('particao') || '',
        zonaUser: $el.attr('setor') || $el.attr('zonaUser') || '',
        idEvento: $el.attr('idEvento') || '',
        img: $el.attr('img') || '',
        conta: $el.attr('conta') || '',
        codBenuvem: $el.attr('codBenuvem') || '',
        provedorVideo: $el.attr('provedorVideo') || sessionStorage.getItem('ateProDado_proc_provedorVideo') || 'nenhum',
        usaConfVision: $el.attr('usaConfVision') || sessionStorage.getItem('ateProDado_proc_usaConfVision') || 'N'
    }
}

if (typeof $ !== 'undefined') {
    confvision_carregarConfig()
}

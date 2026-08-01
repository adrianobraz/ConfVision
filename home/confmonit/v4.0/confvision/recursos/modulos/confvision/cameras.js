let cacheCameras = []
let cacheClientes = null
let modalDetalheCamera = null

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') {
        $('#tab-cameras').html(`
            <tr><td colspan="6">
                <div class="cv-empty">Erro ao carregar scripts. Limpe o cache (Ctrl+F5).</div>
            </td></tr>
        `)
        return
    }

    try {
        confVisionAuthGuard()
        confVisionCarregarCabecalho()
        exibirFlashCamerasSalvas()
        carregarResumoLicencas()
        carregarCameras()

        $('#busca-cameras').on('input', function () {
            clearTimeout(window._cvBuscaCam)
            window._cvBuscaCam = setTimeout(renderListaCameras, 250)
        })
        $('#filtro-ativo-cameras').on('change', renderListaCameras)

        $('#tab-cameras').on('click', '.btn-ver-camera', function () {
            const id = $(this).data('id')
            if (id) abrirDetalheCamera(id)
        })
        $('#tab-cameras').on('click', '.cv-btn-pausar-analitico', function () {
            const $btn = $(this)
            const cameraId = $btn.attr('data-pausar-camera')
            const pausado = $btn.attr('data-pausar-estado') === '1'
            if (typeof ConfVisionArmado === 'undefined') return
            ConfVisionArmado.togglePausarAnalitico(cameraId, !pausado, $btn, function (err) {
                if (!err) carregarCameras()
            })
        })
        $('#modal-detalhe-camera-body').on('click', '.btn-ajuda-encode-rtmp', function () {
            const id = $(this).data('camera-id')
            if (typeof confVisionAbrirAjudaEncodeRtmp === 'function') {
                confVisionAbrirAjudaEncodeRtmp(id)
            }
        })
    } catch (e) {
        console.error('Erro ao iniciar pagina de cameras', e)
        $('#tab-cameras').html(`
            <tr><td colspan="6">
                <div class="cv-empty">Erro ao carregar a pagina. Atualize com Ctrl+F5.</div>
            </td></tr>
        `)
    }
})

function exibirFlashCamerasSalvas() {
    let msg = ''
    try {
        msg = sessionStorage.getItem('cv_cameras_flash') || ''
        if (msg) sessionStorage.removeItem('cv_cameras_flash')
    } catch (e) { /* ignore */ }
    if (msg && typeof boxSucessoAuto === 'function') {
        boxSucessoAuto(msg)
    }
}

function getModalDetalheCamera() {
    if (modalDetalheCamera) return modalDetalheCamera
    const el = document.getElementById('modal-detalhe-camera')
    if (!el || typeof bootstrap === 'undefined') return null
    modalDetalheCamera = new bootstrap.Modal(el)
    return modalDetalheCamera
}

function normalizarListaCameras(r) {
    if (Array.isArray(r)) return r
    if (r && Array.isArray(r.dados)) return r.dados
    return []
}

function escHtml(valor) {
    if (valor == null || valor === '') return '—'
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function badgeAtivo(ativo) {
    return ativo
        ? '<span class="cv-badge cv-badge-on">Ativo</span>'
        : '<span class="cv-badge cv-badge-off">Inativo</span>'
}

function badgeBloqueado(bloqueado) {
    return bloqueado
        ? '<span class="cv-badge cv-badge-off">RTMP bloqueado</span>'
        : ''
}

function badgeStatusCamera(cam) {
    const plano = String(cam.plano || '').trim()
    let base = ''
    if (plano === 'online') {
        base = '<span class="cv-badge cv-badge-stream">Sob demanda</span>'
    } else if (plano.indexOf('sensor') === 0 || plano === 'sensor') {
        base = '<span class="cv-badge cv-badge-stream">Sensor</span>'
    } else {
        base = badgeAtivo(!!cam.ativo)
        if (typeof ConfVisionArmado !== 'undefined' && ConfVisionArmado.isAnaliticoPausado(cam)) {
            base += ' ' + ConfVisionArmado.badgeAnaliticoPausado(cam)
        }
    }
    return base + (cam.bloqueado ? ' ' + badgeBloqueado(true) : '')
}

function labelPlanoCamera(plano) {
    if (typeof labelPlano === 'function' && typeof CV_PLANOS !== 'undefined') {
        return labelPlano(plano)
    }
    const map = {
        online: 'Câmera online',
        sensor_foto: 'Sensor — foto',
        sensor_foto_video: 'Sensor — foto + vídeo',
        sensor: 'Sensor — foto + vídeo',
        analitico_armado_evento: 'Analítico armado — só evento',
        analitico_armado_foto: 'Analítico armado — foto',
        analitico_armado_foto_video: 'Analítico armado — foto + vídeo',
        analitico_armado: 'Analítico armado — foto + vídeo',
        analitico_24h_evento: 'Analítico 24h — só evento',
        analitico_24h_foto: 'Analítico 24h — foto',
        analitico_24h_foto_video: 'Analítico 24h — foto + vídeo',
        analitico_24h: 'Analítico 24h — foto + vídeo'
    }
    const p = String(plano || '').trim()
    return map[p] || p || '—'
}

function badgeStream(online, verificando) {
    if (verificando) {
        return '<span class="cv-badge cv-badge-warn">Verificando…</span>'
    }
    return online
        ? '<span class="cv-badge cv-badge-stream">Online</span>'
        : '<span class="cv-badge cv-badge-off">Offline</span>'
}

function linhaDetalhe(label, valor, mono) {
    const cls = mono ? ' cv-detail-mono' : ''
    return `
        <div class="cv-detail-row">
            <span class="cv-detail-label">${label}</span>
            <span class="cv-detail-value${cls}">${valor}</span>
        </div>
    `
}

function secaoDetalhe(titulo, conteudo) {
    return `
        <section class="cv-detail-section">
            <h6 class="cv-detail-section-title">${titulo}</h6>
            <div class="cv-detail-grid">${conteudo}</div>
        </section>
    `
}

function parseCameraResponse(r) {
    if (!r) return null
    if (r.id) return r
    if (r.dados && r.dados.id) return r.dados
    return r
}

function carregarClientesCache() {
    if (cacheClientes) {
        return $.Deferred().resolve(cacheClientes).promise()
    }
    return $.ajax({
        url: '/api/clientes',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idFranqueado: ConfVisionUrls.idFranqueado() })
    }).then(function (r) {
        cacheClientes = r.dados || r || []
        return cacheClientes
    }).fail(function () {
        cacheClientes = []
        return cacheClientes
    })
}

function nomeClientePorId(idCliente) {
    if (!idCliente || !cacheClientes) return null
    const id = String(idCliente)
    const cliente = cacheClientes.find(function (c) {
        return String(c.idCliente || c.id) === id
    })
    if (!cliente) return null
    return cliente.nome || cliente.nomeCliente || null
}

function verificarStreamOnline(cameraId) {
    return ConfVisionUrls.carregarRtmpPublish(cameraId).then(function (r) {
        const base = r.hls || ConfVisionUrls.hlsUrlFromPath(r.path)
        if (!base) return false
        const full = base + (base.indexOf('?') >= 0 ? '&' : '?') + 't=' + Date.now()
        return fetch(full, { method: 'GET', mode: 'cors' })
            .then(function (resp) { return resp.ok })
            .catch(function () { return false })
    }).catch(function () { return false })
}

function renderDetalheCamera(cam, nomeCliente, streamOnline, verificandoStream) {
    const path = '—'
    const hls = '—'

    const statusHtml = secaoDetalhe('Status', [
        linhaDetalhe('Plano', escHtml(labelPlanoCamera(cam.plano))),
        linhaDetalhe('Cadastro', badgeStatusCamera(cam)),
        linhaDetalhe('Transmissão', badgeStream(streamOnline, verificandoStream)),
        linhaDetalhe('Status worker', escHtml(cam.status || '—')),
        linhaDetalhe('Último evento', confVisionFormatarData(cam.ultimo_evento_em)),
        linhaDetalhe('Último ping', confVisionFormatarData(cam.ultimo_ping_em))
    ].join(''))

    const identHtml = secaoDetalhe('Identificação', [
        linhaDetalhe('ID', escHtml(cam.id)),
        linhaDetalhe('Nome', escHtml(cam.nome)),
        linhaDetalhe('Cadastrada em', confVisionFormatarData(cam.created_at))
    ].join(''))

    const vinculoHtml = secaoDetalhe('Vínculo', [
        linhaDetalhe('Cliente', nomeCliente ? `${escHtml(nomeCliente)} (${escHtml(cam.id_cliente)})` : escHtml(cam.id_cliente)),
        linhaDetalhe('Dispositivo', escHtml(cam.id_dispositivo)),
        linhaDetalhe('Setor', escHtml(cam.setor)),
        linhaDetalhe('Conta', escHtml(cam.conta)),
        linhaDetalhe('Partição', escHtml(cam.particao)),
        linhaDetalhe('Tipo', escHtml(ConfVisionUrls.rotuloTipoCamera(cam.protocolo))),
        linhaDetalhe('Canal', escHtml(cam.canal))
    ].join(''))

    const deteccaoHtml = String(cam.plano || '').trim() === 'online'
        ? ''
        : secaoDetalhe('Detecção', [
            linhaDetalhe('Detecção humano', confVisionSimNao(!!cam.deteccao_humano)),
            linhaDetalhe('Detecção veículo', confVisionSimNao(!!cam.deteccao_veiculo)),
            linhaDetalhe('Confiança mín.', cam.confianca_min != null ? escHtml(cam.confianca_min) : '—'),
            linhaDetalhe('Cooldown (seg)', cam.cooldown_seg != null ? escHtml(cam.cooldown_seg) : '—'),
            linhaDetalhe('Somente armado', confVisionSimNao(!!cam.somente_armado)),
            linhaDetalhe('Captura sensor', confVisionSimNao(!!cam.captura_sensor)),
            linhaDetalhe('Captura analítico', confVisionSimNao(!!cam.captura_analitico)),
            linhaDetalhe('Evento grava foto', confVisionSimNao(!!cam.evento_grava_foto)),
            linhaDetalhe('Evento grava vídeo', confVisionSimNao(!!cam.evento_grava_video))
        ].join(''))

    const ajudaBtn = `<button type="button" class="cv-btn-ajuda-rtmp btn-ajuda-encode-rtmp" data-camera-id="${escHtml(String(cam.id))}" title="Ajuda: encode e URL">
        <i class="bi bi-question-circle"></i> Ajuda encode
    </button>`
    const streamHtml = secaoDetalhe('Stream', [
        linhaDetalhe('Path', `<code id="detalhe-stream-path">${escHtml(path)}</code>`, false),
        linhaDetalhe('RTMP publicação', `<code id="detalhe-rtmp-url">Carregando…</code>`, false),
        linhaDetalhe('HLS reprodução', `<span id="detalhe-hls-url">${escHtml(hls)}</span>`, true),
        linhaDetalhe('Bloqueio RTMP', confVisionSimNao(!!cam.bloqueado)),
        `<div class="cv-detail-ajuda-rtmp">${ajudaBtn}</div>`
    ].join(''))

    setTimeout(function () {
        ConfVisionUrls.carregarRtmpPublish(cam.id).done(function (r) {
            const url = r.url || r.wifi || ''
            const el = document.getElementById('detalhe-rtmp-url')
            if (el) el.textContent = url || '(sem chave — configure RTMP_PUBLISH_SECRET)'
            const elPath = document.getElementById('detalhe-stream-path')
            if (elPath) elPath.textContent = r.path || '—'
            const elHls = document.getElementById('detalhe-hls-url')
            if (elHls) elHls.textContent = r.hls || ConfVisionUrls.hlsUrlFromPath(r.path) || '—'
        }).fail(function () {
            const el = document.getElementById('detalhe-rtmp-url')
            if (el) el.textContent = 'Erro ao gerar URL autenticada'
        })
    }, 0)

    return statusHtml + identHtml + vinculoHtml + deteccaoHtml + streamHtml
}

function abrirDetalheCamera(id) {
    const modal = getModalDetalheCamera()
    if (!modal) {
        boxMesagemAtencaoPersonalizada('Detalhes indisponíveis. Atualize a página (Ctrl+F5).')
        return
    }

    $('#modal-detalhe-camera-titulo').text(`Câmera #${id}`)
    $('#modal-detalhe-camera-body').html(`
        <div class="cv-detail-loading">
            <i class="bi bi-arrow-repeat"></i> Carregando dados…
        </div>
    `)
    $('#modal-detalhe-btn-editar').addClass('d-none').attr('href', '#')
    $('#modal-detalhe-btn-ao-vivo').addClass('d-none').attr('href', '#')
    modal.show()

    Promise.all([
        $.get(`/api/cameras/${id}`).then(parseCameraResponse),
        carregarClientesCache(),
        verificarStreamOnline(id)
    ]).then(function (results) {
        const cam = results[0]
        const streamOnline = results[2]
        if (!cam || !cam.id) {
            $('#modal-detalhe-camera-body').html('<div class="cv-empty">Câmera não encontrada</div>')
            return
        }

        const nomeCliente = nomeClientePorId(cam.id_cliente)
        const html = renderDetalheCamera(cam, nomeCliente, streamOnline, false)
        $('#modal-detalhe-camera-titulo').text(cam.nome ? `${cam.nome} (#${cam.id})` : `Câmera #${cam.id}`)
        $('#modal-detalhe-camera-body').html(html)
        $('#modal-detalhe-btn-editar').removeClass('d-none').attr('href', `/cameras/editar/${cam.id}`)
        if (typeof confVisionCameraApareceAoVivo === 'function' && confVisionCameraApareceAoVivo(cam)) {
            $('#modal-detalhe-btn-ao-vivo').removeClass('d-none').attr('href', `/ao-vivo/${cam.id}`)
        } else {
            $('#modal-detalhe-btn-ao-vivo').addClass('d-none').attr('href', '#')
        }
    }).catch(function () {
        const cam = cacheCameras.find(function (c) { return String(c.id) === String(id) })
        if (!cam) {
            $('#modal-detalhe-camera-body').html('<div class="cv-empty">Erro ao carregar dados da câmera</div>')
            return
        }

        carregarClientesCache().always(function () {
            const nomeCliente = nomeClientePorId(cam.id_cliente)
            $('#modal-detalhe-camera-body').html(renderDetalheCamera(cam, nomeCliente, false, true))
            $('#modal-detalhe-camera-titulo').text(cam.nome ? `${cam.nome} (#${cam.id})` : `Câmera #${cam.id}`)
            $('#modal-detalhe-btn-editar').removeClass('d-none').attr('href', `/cameras/editar/${cam.id}`)
            if (typeof confVisionCameraApareceAoVivo === 'function' && confVisionCameraApareceAoVivo(cam)) {
                $('#modal-detalhe-btn-ao-vivo').removeClass('d-none').attr('href', `/ao-vivo/${cam.id}`)
            } else {
                $('#modal-detalhe-btn-ao-vivo').addClass('d-none').attr('href', '#')
            }
            verificarStreamOnline(cam.id).then(function (online) {
                $('#modal-detalhe-camera-body').html(renderDetalheCamera(cam, nomeCliente, online, false))
            })
        })
    })
}

function nomeClienteCamera(cam) {
    const nome = nomeClientePorId(cam.id_cliente)
    if (nome) return nome
    return cam.id_cliente ? String(cam.id_cliente) : '—'
}

function carregarResumoLicencas() {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!idFranqueado) return

    $.get(`/api/licencas?id_franqueado=${encodeURIComponent(idFranqueado)}`)
        .done(function (r) {
            const resumo = (r && r.resumo) ? r.resumo : {}
            $('#lbl-lic-disponivel').text(resumo.disponivel != null ? resumo.disponivel : 0)
            $('#lbl-lic-em-uso').text(resumo.em_uso != null ? resumo.em_uso : 0)
            $('#panel-licencas').removeClass('d-none')

            const btnNova = $('a[href="/cameras/nova"]')
            if ((resumo.disponivel || 0) === 0) {
                btnNova.addClass('cv-btn-disabled')
                btnNova.attr('title', 'Nenhuma licença disponível — solicite à central')
            } else {
                btnNova.removeClass('cv-btn-disabled')
                btnNova.removeAttr('title')
            }
        })
        .fail(function () {
            $('#panel-licencas').addClass('d-none')
        })
}

function cameraEstaAtiva(cam) {
    return cam.ativo === true || cam.ativo === 'S' || cam.ativo === 1 || cam.ativo === '1'
}

function filtrarCameras(lista) {
    const termo = ($('#busca-cameras').val() || '').toLowerCase().trim()
    const status = ($('#filtro-ativo-cameras').val() || '').trim()

    return (lista || []).filter(function (cam) {
        if (!cam || cam.id == null) return false

        if (status === 'ativo' && !cameraEstaAtiva(cam)) return false
        if (status === 'inativo' && cameraEstaAtiva(cam)) return false

        if (!termo) return true

        const nomeCam = String(cam.nome || '').toLowerCase()
        const nomeCli = String(nomeClienteCamera(cam) || '').toLowerCase()
        const idCam = String(cam.id || '').toLowerCase()
        const setor = String(cam.setor || '').toLowerCase()

        return nomeCam.indexOf(termo) !== -1
            || nomeCli.indexOf(termo) !== -1
            || idCam.indexOf(termo) !== -1
            || setor.indexOf(termo) !== -1
    })
}

function renderListaCameras() {
    const tbody = $('#tab-cameras')
    const $total = $('#lbl-cameras-total')
    tbody.empty()

    if (!cacheCameras.length) {
        tbody.append(`
            <tr>
                <td colspan="6">
                    <div class="cv-empty">Nenhuma câmera cadastrada</div>
                </td>
            </tr>
        `)
        $total.text('')
        return
    }

    const lista = filtrarCameras(cacheCameras)
    const totalGeral = cacheCameras.length
    if (!lista.length) {
        tbody.append(`
            <tr>
                <td colspan="6">
                    <div class="cv-empty">Nenhuma câmera encontrada com esse filtro</div>
                </td>
            </tr>
        `)
        $total.text('0 de ' + totalGeral + ' câmeras')
        return
    }

    if (lista.length === totalGeral) {
        $total.text(totalGeral + (totalGeral === 1 ? ' câmera' : ' câmeras'))
    } else {
        $total.text(lista.length + ' de ' + totalGeral + ' câmeras')
    }

    try {
        lista.forEach(function (cam) {
            const status = badgeStatusCamera(cam)
            const pausarBtn = (typeof ConfVisionArmado !== 'undefined')
                ? ConfVisionArmado.htmlBotaoPausarAnalitico(cam)
                : ''
            const aoVivoBtn = (typeof confVisionCameraApareceAoVivo === 'function' && confVisionCameraApareceAoVivo(cam))
                ? `<a href="/ao-vivo/${cam.id}" class="cv-btn-icon btn-ao-vivo-camera" title="Ao vivo">
                                <i class="bi bi-play-fill"></i>
                            </a>`
                : ''
            tbody.append(`
                <tr>
                    <td data-label="ID"><strong>${escHtml(cam.id)}</strong></td>
                    <td data-label="Cliente">${escHtml(nomeClienteCamera(cam))}</td>
                    <td data-label="Câmera">${escHtml(cam.nome)}</td>
                    <td data-label="Setor">${escHtml(cam.setor)}</td>
                    <td data-label="Status">${status}</td>
                    <td class="cv-cell-action" data-label="Ações">
                        <div class="cv-action-group">
                            <a href="/cameras/editar/${cam.id}" class="cv-btn-icon btn-editar-camera" title="Editar">
                                <i class="bi bi-pencil"></i>
                            </a>
                            ${pausarBtn}
                            <button type="button" class="cv-btn-icon btn-ver-camera" data-id="${cam.id}" title="Ver detalhes">
                                <i class="bi bi-eye"></i>
                            </button>
                            ${aoVivoBtn}
                        </div>
                    </td>
                </tr>
            `)
        })
    } catch (e) {
        console.error('Erro ao renderizar câmeras', e)
        tbody.html(`
            <tr>
                <td colspan="6">
                    <div class="cv-empty">Erro ao exibir a listagem. Atualize a página (Ctrl+F5).</div>
                </td>
            </tr>
        `)
    }
}

function carregarCameras() {
    const tbody = $('#tab-cameras')
    tbody.html(`
        <tr>
            <td colspan="6">
                <div class="cv-detail-loading">
                    <i class="bi bi-arrow-repeat"></i> Carregando câmeras…
                </div>
            </td>
        </tr>
    `)

    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!idFranqueado) {
        tbody.html(`
            <tr>
                <td colspan="6">
                    <div class="cv-empty">Sessão inválida. Faça login novamente.</div>
                </td>
            </tr>
        `)
        return
    }

    $.when(
        $.get(`/api/cameras?id_franqueado=${encodeURIComponent(idFranqueado)}`),
        carregarClientesCache()
    ).done(function (camResp) {
        const r = camResp[0]
        cacheCameras = normalizarListaCameras(r)
        renderListaCameras()
    }).fail(function (xhr) {
        console.error('Erro ao carregar câmeras', xhr)
        tbody.html(`
            <tr>
                <td colspan="6">
                    <div class="cv-empty">Erro ao carregar câmeras. Tente sair e entrar novamente.</div>
                </td>
            </tr>
        `)
        boxMesagemAtencaoPersonalizada('Erro ao carregar câmeras. Faça logout e login novamente.')
    })
}

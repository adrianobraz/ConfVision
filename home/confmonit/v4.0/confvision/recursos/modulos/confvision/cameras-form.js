let dispositivoSelecionado = null
let clientesCache = []
let clienteSelecionadoId = null
let modoForm = 'create'
let cameraEditId = null
let cameraEditCache = null
let snapshotUrlAtual = null
let snapshotBusy = false
let cameraEdicaoPendente = null
let dispositivoInicialEdicao = null
let licencasDisponiveis = []
let licencaSelecionadaId = null
let licencaAtualId = null
let planoAtualCam = null
let valorAtualCam = null

let planoInicialCam = null
let valorInicialCam = null
let licencasGravacaoDisponiveis = []
let setoresOcupadosCache = new Set()

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

const CV_PLANOS = {
    online: { label: 'Câmera online', valor: 'R$ 2,99', captura_sensor: false, captura_analitico: false, somente_armado: false, evento_grava_foto: false, evento_grava_video: false, sem_ativo: true },
    sensor_foto: { label: 'Sensor — foto', valor: 'R$ 7,99', captura_sensor: true, captura_analitico: false, somente_armado: false, evento_grava_foto: true, evento_grava_video: false, sem_ativo: true },
    sensor_foto_video: { label: 'Sensor — foto + vídeo', valor: 'R$ 9,99', captura_sensor: true, captura_analitico: false, somente_armado: false, evento_grava_foto: true, evento_grava_video: true, sem_ativo: true },
    analitico_armado_evento: { label: 'Analítico armado — só evento', valor: 'R$ 11,99', captura_sensor: false, captura_analitico: true, somente_armado: true, evento_grava_foto: false, evento_grava_video: false, sem_ativo: false },
    analitico_armado_foto: { label: 'Analítico armado — foto', valor: 'R$ 13,99', captura_sensor: false, captura_analitico: true, somente_armado: true, evento_grava_foto: true, evento_grava_video: false, sem_ativo: false },
    analitico_armado_foto_video: { label: 'Analítico armado — foto + vídeo', valor: 'R$ 14,99', captura_sensor: false, captura_analitico: true, somente_armado: true, evento_grava_foto: true, evento_grava_video: true, sem_ativo: false },
    analitico_24h_evento: { label: 'Analítico 24h — só evento', valor: 'R$ 16,99', captura_sensor: false, captura_analitico: true, somente_armado: false, evento_grava_foto: false, evento_grava_video: false, sem_ativo: false },
    analitico_24h_foto: { label: 'Analítico 24h — foto', valor: 'R$ 18,99', captura_sensor: false, captura_analitico: true, somente_armado: false, evento_grava_foto: true, evento_grava_video: false, sem_ativo: false },
    analitico_24h_foto_video: { label: 'Analítico 24h — foto + vídeo', valor: 'R$ 19,99', captura_sensor: false, captura_analitico: true, somente_armado: false, evento_grava_foto: true, evento_grava_video: true, sem_ativo: false },
    sensor: { label: 'Sensor — foto + vídeo', valor: 'R$ 9,99', captura_sensor: true, captura_analitico: false, somente_armado: false, evento_grava_foto: true, evento_grava_video: true, sem_ativo: true },
    analitico_armado: { label: 'Analítico armado — foto + vídeo', valor: 'R$ 14,99', captura_sensor: false, captura_analitico: true, somente_armado: true, evento_grava_foto: true, evento_grava_video: true, sem_ativo: false },
    analitico_24h: { label: 'Analítico 24h — foto + vídeo', valor: 'R$ 19,99', captura_sensor: false, captura_analitico: true, somente_armado: false, evento_grava_foto: true, evento_grava_video: true, sem_ativo: false }
}

function planoEhOnline(plano) {
    return String(plano || '').trim() === 'online'
}

function planoSemAtivo(plano) {
    const p = CV_PLANOS[String(plano || '').trim()]
    return !!(p && p.sem_ativo)
}

function planoEhAnalitico(plano) {
    const p = String(plano || '').trim()
    return p.indexOf('analitico_') === 0 || p === 'analitico_armado' || p === 'analitico_24h'
}

function planoEhSensor(plano) {
    const p = String(plano || '').trim()
    return p.indexOf('sensor') === 0
}

function planoPermiteAreaDeteccao(plano) {
    return planoEhAnalitico(plano || planoAtualForm())
}

function normalizarPlano(plano) {
    return String(plano || '').trim()
}

function inferirPlanoLicenca(lic) {
    const p = normalizarPlano(lic && lic.plano)
    if (p && CV_PLANOS[p]) return p
    const v = parseFloat(lic && lic.valor)
    if (!isNaN(v)) {
        if (v <= 3) return 'online'
        if (v <= 8) return 'sensor_foto'
        if (v <= 10) return 'sensor_foto_video'
        if (v <= 12) return 'analitico_armado_evento'
        if (v <= 14) return 'analitico_armado_foto'
        if (v <= 15) return 'analitico_armado_foto_video'
        if (v <= 17) return 'analitico_24h_evento'
        if (v <= 19) return 'analitico_24h_foto'
        if (v <= 20) return 'analitico_24h_foto_video'
    }
    return p || null
}

function planoAtualForm() {
    if (planoAtualCam) return planoAtualCam
    if (modoForm === 'create') {
        const lic = licencasDisponiveis.find(function (l) {
            return String(l.id) === String(licencaSelecionadaId || '')
        })
        return lic ? inferirPlanoLicenca(lic) : null
    }
    return null
}

function labelPlano(plano) {
    const p = CV_PLANOS[plano]
    return p ? p.label : (plano || '—')
}

function valorPlano(plano, valorLicenca) {
    if (valorLicenca != null && valorLicenca !== '') {
        const n = parseFloat(valorLicenca)
        if (!isNaN(n)) {
            return 'R$ ' + n.toFixed(2).replace('.', ',')
        }
    }
    const p = CV_PLANOS[plano]
    return p ? p.valor : '—'
}

function ordenarLicencas(lista) {
    return lista.slice().sort(function (a, b) {
        return (parseInt(a.id, 10) || 0) - (parseInt(b.id, 10) || 0)
    })
}

function labelPlanoGravacao(plano) {
    return CV_PLANOS_GRAVACAO[plano] || plano || '—'
}

function gravacaoAtivaCam(cam) {
    return String((cam && cam.gravacao_status) || '').toLowerCase() === 'ativa'
}

function temLicencaGravacaoCam(cam) {
    if (!cam) return false
    if (cam.vis_licenca_gravacao_id != null) return true
    if (gravacaoAtivaCam(cam)) return true
    if (cam.grava_continua || cam.grava_movimento || cam.grava_timelapse) return true
    if (cam.retencao_dias != null && Number(cam.retencao_dias) > 0) return true
    return false
}

function modoGravacaoCam(cam) {
    if (!cam) return '—'
    if (cam.grava_timelapse) return 'Timelapse Inteligente'
    if (cam.grava_movimento) return 'Gravação por movimento'
    if (cam.grava_continua) return 'Gravação contínua'
    return '—'
}

function ocultarPainelGravacao() {
    $('#box-gravacao-painel').addClass('d-none')
    if (modoForm === 'edit' && cameraEditId) {
        // Em edicao a aba permanece visivel; so o conteudo fica indisponivel se falhar o load
        $('#tab-item-gravacao').removeClass('d-none')
        $('#box-gravacao-indisponivel').removeClass('d-none')
            .find('p').text('Não foi possível carregar a gravação. Atualize a página ou use /gravacoes.')
        return
    }
    $('#tab-item-gravacao').addClass('d-none')
    $('#box-gravacao-indisponivel').removeClass('d-none')
        .find('p').text('Salve a câmera primeiro para gerenciar a licença de gravação.')
}

function mostrarAbaGravacao(mostrar) {
    if (modoForm === 'edit' && cameraEditId) {
        mostrar = true
    }
    $('#tab-item-gravacao').toggleClass('d-none', !mostrar)
    if (!mostrar) {
        const tabAtiva = $('#cv-form-tabs .nav-link.active').attr('data-tab')
        if (tabAtiva === 'gravacao') {
            selecionarAbaForm('dados')
        }
        $('#box-gravacao-indisponivel').removeClass('d-none')
    } else {
        $('#box-gravacao-indisponivel').addClass('d-none')
    }
}

function selecionarAbaForm(tab) {
    const alvo = String(tab || 'dados')
    const eraAoVivo = $('#cv-form-tabs .nav-link.active').attr('data-tab') === 'ao-vivo'

    $('#cv-form-tabs .nav-link').removeClass('active')
    $('#cv-form-tabs .nav-link[data-tab="' + alvo + '"]').addClass('active')
    $('.cv-form-tab-panel').addClass('d-none')
    $('.cv-form-tab-panel[data-tab-panel="' + alvo + '"]').removeClass('d-none')

    if (alvo !== 'ao-vivo' && (eraAoVivo || typeof ConfVisionLivePlayer !== 'undefined')) {
        if (typeof ConfVisionLivePlayer !== 'undefined') {
            ConfVisionLivePlayer.stop()
        }
    }

    if (alvo === 'snapshot' || alvo === 'area') {
        if (typeof ConfVisionAreaEditor !== 'undefined') {
            window.setTimeout(function () {
                ConfVisionAreaEditor.refreshEditorVisibility()
            }, 50)
        }
    }

    if (alvo === 'ao-vivo' && cameraEditId && typeof ConfVisionLivePlayer !== 'undefined') {
        window.setTimeout(function () {
            ConfVisionLivePlayer.start({
                video: '#player-hls-form',
                status: '#cv-live-status-form',
                cameraId: cameraEditId
            })
        }, 80)
    }

    if (alvo === 'url' && cameraEditId) {
        exibirLinkStream(cameraEditId)
    }

    $('#box-form-actions').toggleClass('d-none', alvo === 'gravacao' || alvo === 'licenca')
    $('#btn-ajuda-encode-rtmp-acoes').toggleClass('d-none', alvo !== 'url')
}

function atualizarBotoesGravacaoPainel(temLicencaGrav) {
    const licAtivar = normalizarId($('#sel-gravacao-licenca').val())
    const licTrocar = normalizarId($('#sel-trocar-gravacao-licenca').val())
    $('#btn-ativar-gravacao-cam').prop('disabled', !licAtivar)
    $('#btn-trocar-gravacao-cam').prop('disabled', !licTrocar)
    $('#btn-desativar-gravacao-cam').toggleClass('d-none', !temLicencaGrav)
}

function preencherSelectLicencasGravacao(lista, temLicencaGrav) {
    licencasGravacaoDisponiveis = lista
    const selAtivar = $('#sel-gravacao-licenca')
    const selTrocar = $('#sel-trocar-gravacao-licenca')
    selAtivar.find('option:not(:first)').remove()
    selTrocar.find('option:not(:first)').remove()
    lista.forEach(function (lic) {
        const label = `#${lic.id} — ${labelPlanoGravacao(lic.plano)}`
        selAtivar.append(`<option value="${lic.id}">${label}</option>`)
        selTrocar.append(`<option value="${lic.id}">${label}</option>`)
    })
    $('#lbl-gravacao-licenca-vazia').toggleClass('d-none', lista.length > 0)
    $('#box-gravacao-ativar').toggleClass('d-none', temLicencaGrav)
    $('#box-gravacao-trocar').toggleClass('d-none', !temLicencaGrav)
    $('#box-gravacao-trocar-select').toggleClass('d-none', !lista.length)
    $('#btn-trocar-gravacao-cam').toggleClass('d-none', !lista.length)
    atualizarBotoesGravacaoPainel(temLicencaGrav)
}

function renderPainelGravacao(cam) {
    if (!cam || !cameraEditId) {
        ocultarPainelGravacao()
        return
    }

    const temLicenca = temLicencaGravacaoCam(cam)
    mostrarAbaGravacao(true)
    $('#box-gravacao-painel').removeClass('d-none')
    $('#box-gravacao-indisponivel').addClass('d-none')
    $('#box-gravacao-ativa').toggleClass('d-none', !temLicenca)
    $('#box-gravacao-inativa').toggleClass('d-none', temLicenca)

    if (temLicenca) {
        const licId = cam.vis_licenca_gravacao_id
        const modo = modoGravacaoCam(cam)
        const retencao = cam.retencao_dias ? `${cam.retencao_dias} dias` : '—'
        $('#lbl-gravacao-atual').text(licId ? `#${licId} — ${modo}` : modo)
        $('#lbl-gravacao-detalhe').text(`Retenção: ${retencao}`)
    }
}

function carregarLicencasGravacaoPainel(cam) {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!cameraEditId) {
        ocultarPainelGravacao()
        return $.Deferred().resolve().promise()
    }

    // Em edicao, aba sempre disponivel
    mostrarAbaGravacao(true)

    if (!idFranqueado) {
        ocultarPainelGravacao()
        return $.Deferred().resolve().promise()
    }

    const temLicenca = temLicencaGravacaoCam(cam)
    renderPainelGravacao(cam)

    return $.get(`/api/licencas?id_franqueado=${encodeURIComponent(idFranqueado)}&status=disponivel&unidade=gravacao`)
        .fail(function (xhr) {
            console.error(xhr)
            preencherSelectLicencasGravacao([], temLicenca)
            boxMesagemAtencaoPersonalizada(extrairErroApi(xhr) || 'Erro ao carregar licenças de gravação.')
        })
        .done(function (r) {
            const bruto = (r && r.dados) ? r.dados : (Array.isArray(r) ? r : [])
            preencherSelectLicencasGravacao(ordenarLicencas(bruto), temLicenca)
        })
}

function configurarPainelGravacao(cam) {
    if (modoForm !== 'edit' || !cameraEditId) {
        ocultarPainelGravacao()
        return
    }
    mostrarAbaGravacao(true)
    carregarLicencasGravacaoPainel(cam || cameraEditCache || {})
}

function postAtivarGravacaoCam(licId) {
    return $.ajax({
        url: `/api/cameras/${cameraEditId}/gravacao/ativar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ vis_licenca_gravacao_id: parseInt(licId, 10) })
    })
}

function postDesativarGravacaoCam() {
    return $.ajax({
        url: `/api/cameras/${cameraEditId}/gravacao/desativar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({})
    })
}

function recarregarGravacaoPainel() {
    if (!cameraEditId) return
    $.get(`/api/cameras/${cameraEditId}`)
        .done(function (r) {
            const cam = parseCameraResponse(r)
            if (!cam) return
            cameraEditCache = cam
            configurarPainelGravacao(cam)
        })
        .fail(function (xhr) {
            boxMesagemAtencaoPersonalizada(extrairErroApi(xhr) || 'Erro ao atualizar painel de gravação.')
        })
}

function ativarGravacaoCamForm() {
    const licId = normalizarId($('#sel-gravacao-licenca').val())
    if (!cameraEditId || !licId) {
        boxMesagemAtencaoPersonalizada('Selecione uma licença de gravação disponível.')
        return
    }
    const $btn = $('#btn-ativar-gravacao-cam')
    $btn.prop('disabled', true)
    postAtivarGravacaoCam(licId)
        .done(function (r) {
            const cam = parseCameraResponse(r)
            if (cam && cam.id) cameraEditCache = cam
            boxSucessoAuto('Gravação ativada (mesma função da página Gravações).')
            recarregarGravacaoPainel()
        })
        .fail(function (xhr) {
            boxMesagemAtencaoPersonalizada(extrairErroApi(xhr) || 'Erro ao ativar gravação. Confira storage Contabo e se a licença está disponível.')
        })
        .always(function () {
            atualizarBotoesGravacaoPainel(temLicencaGravacaoCam(cameraEditCache))
        })
}

function trocarGravacaoCamForm() {
    const licId = normalizarId($('#sel-trocar-gravacao-licenca').val())
    if (!cameraEditId || !licId) {
        boxMesagemAtencaoPersonalizada('Selecione a nova licença de gravação.')
        return
    }
    CvMsg.confirmar('Trocar licença?', 'A licença atual voltará para disponível.').then(function (r) {
        if (!r.isConfirmed) return
        const $btn = $('#btn-trocar-gravacao-cam')
        $btn.prop('disabled', true)
        postAtivarGravacaoCam(licId)
            .done(function (res) {
                const cam = parseCameraResponse(res)
                if (cam && cam.id) cameraEditCache = cam
                boxSucessoAuto('Licença de gravação trocada.')
                recarregarGravacaoPainel()
            })
            .fail(function (xhr) {
                boxMesagemAtencaoPersonalizada(extrairErroApi(xhr) || 'Erro ao trocar licença. Faça push do endpoint 2238 (ativar/trocar) no Xano.')
                recarregarGravacaoPainel()
            })
            .always(function () {
                atualizarBotoesGravacaoPainel(true)
            })
    })
}

function desativarGravacaoCamForm() {
    if (!cameraEditId) return
    CvMsg.confirmar('Retirar licença?', 'A licença de gravação voltará para disponível.').then(function (r) {
        if (!r.isConfirmed) return
        const $btn = $('#btn-desativar-gravacao-cam')
        $btn.prop('disabled', true)
        postDesativarGravacaoCam()
            .done(function () {
                boxSucessoAuto('Licença de gravação retirada.')
                recarregarGravacaoPainel()
            })
            .fail(function (xhr) {
                boxMesagemAtencaoPersonalizada(extrairErroApi(xhr) || 'Erro ao retirar licença de gravação.')
            })
            .always(function () {
                $btn.prop('disabled', false)
            })
    })
}

function aplicarEstadoPlano(planoOuLic, valorLicenca, opts) {
    opts = opts || {}
    let plano = planoOuLic
    if (planoOuLic && typeof planoOuLic === 'object') {
        valorLicenca = planoOuLic.valor
        plano = inferirPlanoLicenca(planoOuLic)
    } else {
        plano = normalizarPlano(plano)
    }

    planoAtualCam = plano || null

    const p = plano ? (CV_PLANOS[plano] || null) : null
    const online = planoEhOnline(plano)
    const analitico = p ? !!p.captura_analitico : planoEhAnalitico(plano)
    const temPlano = !!plano

    $('.cv-form-ativo-wrap').toggleClass('d-none', !analitico || !temPlano || (p && p.sem_ativo) || planoSemAtivo(plano))
    $('.cv-campo-flags-plano').toggleClass('d-none', online || !temPlano)
    $('.cv-campo-yolo').toggleClass('d-none', online || !temPlano)

    $('#inp-captura-sensor, #inp-captura-analitico, #inp-somente-armado, #inp-deteccao')
        .prop('disabled', false)

    if (online) {
        $('#inp-ativo, #inp-deteccao, #inp-captura-sensor, #inp-captura-analitico, #inp-somente-armado')
            .prop('checked', false)
    } else if (p) {
        $('#inp-captura-sensor').prop('checked', !!p.captura_sensor)
        $('#inp-captura-analitico').prop('checked', !!p.captura_analitico)
        $('#inp-somente-armado').prop('checked', !!p.somente_armado)
        $('#inp-deteccao').prop('checked', analitico)
        if (opts.preserveAtivo) {
            $('#inp-ativo').prop('checked', !!opts.ativo)
            if (analitico && opts.deteccao != null) {
                $('#inp-deteccao').prop('checked', !!opts.deteccao)
            }
        } else {
            $('#inp-ativo').prop('checked', false)
        }
        bloquearCamposPlano(true)
        if (!analitico) {
            $('#inp-deteccao').prop('checked', false).prop('disabled', true)
        }
    } else {
        $('#inp-ativo, #inp-deteccao, #inp-captura-sensor, #inp-captura-analitico, #inp-somente-armado')
            .prop('checked', false)
        bloquearCamposPlano(false)
    }

    if (cameraEditId) {
        configurarAbaSnapshot(true)
        configurarAbaArea(true)
        configurarAbaAoVivo(true)
        configurarAbaUrl(true)
        atualizarTextoSnapshot(plano)
        atualizarTextoArea(plano)
        exibirLinkStream(cameraEditId)
    } else {
        configurarAbaSnapshot(false)
        configurarAbaArea(false)
        configurarAbaAoVivo(false)
        configurarAbaUrl(false)
    }

    atualizarBoxPlanoInfo(plano, valorLicenca)

    if (typeof ConfVisionAreaEditor !== 'undefined') {
        ConfVisionAreaEditor.setPermiteAreaDeteccao(planoPermiteAreaDeteccao(plano))
        ConfVisionAreaEditor.refreshEditorVisibility()
    }
}

function atualizarTextoSnapshot(plano) {
    const $desc = $('#lbl-snapshot-desc')
    if (!$desc.length) return
    const permiteArea = planoPermiteAreaDeteccao(plano)
    if (permiteArea) {
        $desc.text('Capture uma imagem do stream. Depois configure a área de detecção na aba Área de detecção.')
    } else {
        $desc.text('Capture uma imagem de referência da câmera.')
    }
}

function atualizarTextoArea(plano) {
    const permiteArea = planoPermiteAreaDeteccao(plano)
    $('#box-snapshot-aviso-area').toggleClass('d-none', !permiteArea)
    const $desc = $('#lbl-area-deteccao-desc')
    if (!$desc.length) return
    if (permiteArea) {
        $desc.text('Escolha retângulo ou polígono e desenhe sobre o snapshot. O modo Detectar define se o evento dispara dentro, fora ou em qualquer lugar do frame.')
    } else {
        $desc.text('Área de detecção não se aplica a este plano.')
    }
}

function configurarAbaSnapshot(disponivel) {
    const cam = cameraEditCache
    const elegivel = !cam || confVisionCameraPodeStream(cam).ok
    const ok = !!disponivel && !!cameraEditId && elegivel
    $('#box-snapshot-conteudo').toggleClass('d-none', !ok)
    $('#box-snapshot-indisponivel').toggleClass('d-none', ok)
    const $ind = $('#box-snapshot-indisponivel p')
    if (!ok && $ind.length) {
        if (!cameraEditId) {
            $ind.text('Salve a câmera primeiro para capturar o snapshot.')
        } else if (cam && !elegivel) {
            $ind.text(confVisionMotivoStream(cam))
        }
    }
    if (ok) {
        $('#box-snapshot').removeClass('d-none')
        atualizarTextoSnapshot(planoAtualForm())
    }
}

function configurarAbaArea(disponivel) {
    const cam = cameraEditCache
    const elegivel = !cam || confVisionCameraPodeStream(cam).ok
    const ok = !!disponivel && !!cameraEditId && elegivel
    $('#box-area-conteudo').toggleClass('d-none', !ok)
    $('#box-area-indisponivel').toggleClass('d-none', ok)
    const $ind = $('#box-area-indisponivel p')
    if (!ok && $ind.length) {
        if (!cameraEditId) {
            $ind.text('Salve a câmera e capture um snapshot na aba Snapshot para desenhar a área de detecção.')
        } else if (cam && !elegivel) {
            $ind.text(confVisionMotivoStream(cam))
        }
    }
    if (ok) {
        atualizarTextoArea(planoAtualForm())
    }
}

function configurarAbaAoVivo(disponivel) {
    const cam = cameraEditCache
    const ok = !!disponivel && !!cameraEditId
    $('#box-ao-vivo-conteudo').toggleClass('d-none', !ok)
    $('#box-ao-vivo-indisponivel').toggleClass('d-none', ok)
    const $ind = $('#box-ao-vivo-indisponivel p')
    if (!ok && $ind.length && !cameraEditId) {
        $ind.text('Salve a câmera primeiro para assistir ao vivo.')
    }
    if (!ok && typeof ConfVisionLivePlayer !== 'undefined') {
        ConfVisionLivePlayer.stop()
    }
}

function configurarAbaUrl(disponivel) {
    const ok = !!disponivel && !!cameraEditId
    $('#box-url-conteudo').toggleClass('d-none', !ok)
    $('#box-url-indisponivel').toggleClass('d-none', ok)
}

function aplicarFlagsPlano(plano) {
    aplicarEstadoPlano(plano)
}

function aplicarUiPlano(plano) {
    aplicarEstadoPlano(plano)
}

function bloquearCamposPlano(travado) {
    const $campos = $('#inp-somente-armado, #inp-captura-analitico, #inp-captura-sensor')
    $campos.prop('disabled', !!travado)
    $campos.closest('.form-check').toggleClass('cv-field-licenca-locked', !!travado)
}

function atualizarBoxPlanoInfo(plano, valorLicenca) {
    if (!plano) {
        $('#box-plano-licenca').addClass('d-none')
        if (modoForm === 'create') {
            $('#box-licenca-painel').addClass('d-none')
        }
        return
    }
    const p = CV_PLANOS[plano] || null
    const $hint = $('#lbl-plano-online-hint')
    $('#lbl-plano-nome').text(labelPlano(plano))
    $('#lbl-plano-valor').text(valorPlano(plano, valorLicenca))
    if (planoEhOnline(plano)) {
        $hint.removeClass('d-none').text(
            'Visualização ao vivo sob demanda. Sem captura por sensor, analítico ou monitoramento contínuo.'
        )
    } else if (planoEhSensor(plano)) {
        let msg = 'Captura automática no alarme.'
        if (p && p.evento_grava_foto && p.evento_grava_video) {
            msg += ' Evento com foto e vídeo.'
        } else if (p && p.evento_grava_foto) {
            msg += ' Evento com foto.'
        } else if (p && p.evento_grava_video) {
            msg += ' Evento com vídeo.'
        } else {
            msg += ' Evento somente registro.'
        }
        $hint.removeClass('d-none').text(msg)
    } else {
        $hint.addClass('d-none').text('')
    }
    $('#box-plano-licenca').removeClass('d-none')
    $('#box-licenca-painel').removeClass('d-none')
}

function carregarLicencasDisponiveis() {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!idFranqueado || modoForm !== 'create') return $.Deferred().resolve().promise()

    return $.get(`/api/licencas?id_franqueado=${encodeURIComponent(idFranqueado)}&status=disponivel&unidade=camera`)
        .fail(function (xhr) {
            console.error(xhr)
            $('#lbl-licenca-vazia').removeClass('d-none').text('Erro ao carregar licenças.')
        })
        .done(function (r) {
            const bruto = (r && r.dados) ? r.dados : (Array.isArray(r) ? r : [])
            const lista = ordenarLicencas(bruto)
            licencasDisponiveis = lista
            const sel = $('#sel-licenca').empty().append('<option value="">Selecione uma licença…</option>')
            lista.forEach(function (lic) {
                const id = lic.id
                const planoLic = inferirPlanoLicenca(lic)
                const label = `#${id} — ${labelPlano(planoLic)} (${valorPlano(planoLic, lic.valor)})`
                sel.append(`<option value="${id}">${label}</option>`)
            })
            $('#box-selecionar-licenca').removeClass('d-none')
            if (!lista.length) {
                $('#lbl-licenca-vazia').removeClass('d-none')
                sel.prop('disabled', true)
            } else {
                $('#lbl-licenca-vazia').addClass('d-none')
                sel.prop('disabled', false)
                const params = new URLSearchParams(window.location.search)
                const pre = params.get('licenca')
                if (pre && sel.find(`option[value="${pre}"]`).length) {
                    sel.val(pre)
                    sel.trigger('change')
                }
            }
        })
}

function onLicencaChange() {
    const id = normalizarId($('#sel-licenca').val())
    licencaSelecionadaId = id || null
    if (!id) {
        planoAtualCam = null
        bloquearCamposPlano(false)
        atualizarBoxPlanoInfo(null)
        aplicarEstadoPlano(null)
        return
    }
    const lic = licencasDisponiveis.find(function (l) { return String(l.id) === id })
    if (!lic) return
    aplicarEstadoPlano(lic)
}

function onTrocarLicencaChange() {
    const id = normalizarId($('#sel-trocar-licenca').val())
    if (!id) {
        licencaSelecionadaId = licencaAtualId
        planoAtualCam = planoInicialCam
        aplicarEstadoPlano({ plano: planoInicialCam, valor: valorInicialCam }, valorInicialCam, {
            preserveAtivo: true,
            ativo: $('#inp-ativo').data('valor-inicial'),
            deteccao: $('#inp-deteccao').data('valor-inicial')
        })
        return
    }
    const lic = licencasDisponiveis.find(function (l) { return String(l.id) === id })
    if (!lic) return
    licencaSelecionadaId = id
    aplicarEstadoPlano(lic)
}

function carregarLicencasParaTroca(cam) {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!idFranqueado || !cam.vis_licenca_id) return $.Deferred().resolve().promise()

    return $.get(`/api/licencas?id_franqueado=${encodeURIComponent(idFranqueado)}&status=disponivel&unidade=camera`)
        .fail(function (xhr) {
            console.error(xhr)
            $('#lbl-trocar-licenca-vazia').removeClass('d-none').text('Erro ao carregar licenças para troca.')
            $('#box-trocar-licenca').removeClass('d-none')
        })
        .done(function (r) {
            const bruto = (r && r.dados) ? r.dados : (Array.isArray(r) ? r : [])
            const lista = ordenarLicencas(bruto)
            licencasDisponiveis = lista
            const sel = $('#sel-trocar-licenca').empty().append('<option value="">Manter licença atual</option>')
            lista.forEach(function (lic) {
                const id = lic.id
                const planoLic = inferirPlanoLicenca(lic)
                const label = `#${id} — ${labelPlano(planoLic)} (${valorPlano(planoLic, lic.valor)})`
                sel.append(`<option value="${id}">${label}</option>`)
            })
            $('#box-trocar-licenca').removeClass('d-none')
            if (!lista.length) {
                sel.prop('disabled', true)
                $('#lbl-trocar-licenca-vazia').removeClass('d-none')
            } else {
                sel.prop('disabled', false)
                $('#lbl-trocar-licenca-vazia').addClass('d-none')
            }
        })
}

function carregarLicencasParaVincular() {
    const idFranqueado = ConfVisionUrls.idFranqueado()
    if (!idFranqueado) return $.Deferred().resolve().promise()

    return $.get(`/api/licencas?id_franqueado=${encodeURIComponent(idFranqueado)}&status=disponivel&unidade=camera`)
        .fail(function (xhr) {
            console.error(xhr)
            $('#lbl-vincular-licenca-vazia').removeClass('d-none').text('Erro ao carregar licenças.')
            $('#box-vincular-licenca').removeClass('d-none')
        })
        .done(function (r) {
            const bruto = (r && r.dados) ? r.dados : (Array.isArray(r) ? r : [])
            const lista = ordenarLicencas(bruto)
            licencasDisponiveis = lista
            const sel = $('#sel-vincular-licenca').empty().append('<option value="">Selecione uma licença…</option>')
            lista.forEach(function (lic) {
                const id = lic.id
                const planoLic = inferirPlanoLicenca(lic)
                const label = `#${id} — ${labelPlano(planoLic)} (${valorPlano(planoLic, lic.valor)})`
                sel.append(`<option value="${id}">${label}</option>`)
            })
            $('#box-vincular-licenca').removeClass('d-none')
            if (!lista.length) {
                sel.prop('disabled', true)
                $('#lbl-vincular-licenca-vazia').removeClass('d-none')
            } else {
                sel.prop('disabled', false)
                $('#lbl-vincular-licenca-vazia').addClass('d-none')
            }
        })
}

function onVincularLicencaChange() {
    const id = normalizarId($('#sel-vincular-licenca').val())
    licencaSelecionadaId = id || null
    if (!id) {
        planoAtualCam = null
        aplicarEstadoPlano(null)
        return
    }
    const lic = licencasDisponiveis.find(function (l) { return String(l.id) === id })
    if (!lic) return
    aplicarEstadoPlano(lic)
}

function postLiberarLicencaCam() {
    return $.ajax({
        url: `/api/cameras/${cameraEditId}/licenca/liberar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({})
    })
}

function liberarLicencaCamForm() {
    if (!cameraEditId || !licencaAtualId) return
    const tinhaGravacao = cameraEditCache && (
        String(cameraEditCache.gravacao_status || '').toLowerCase() === 'ativa' ||
        cameraEditCache.vis_licenca_gravacao_id != null
    )
    let msg = 'Retirar a licença desta câmera? A licença voltará para disponível.'
    if (tinhaGravacao) {
        msg += ' Se houver gravação ativa, ela também será encerrada.'
    }
    CvMsg.confirmar('Retirar licença?', msg).then(function (r) {
        if (!r.isConfirmed) return
        const $btn = $('#btn-liberar-licenca-cam')
        $btn.prop('disabled', true)
        postLiberarLicencaCam()
            .done(function (res) {
                const cam = parseCameraResponse(res) || {}
                cameraEditCache = cam
                boxSucessoAuto('Licença retirada. A câmera ficou sem plano até vincular outra.')
                configurarPlanoEdicao(cam)
                selecionarAbaForm('licenca')
            })
            .fail(function (xhr) {
                boxMesagemAtencaoPersonalizada(extrairErroApi(xhr) || 'Erro ao retirar licença.')
            })
            .always(function () {
                $btn.prop('disabled', false)
            })
    })
}

function configurarPlanoEdicao(cam) {
    licencaAtualId = cam.vis_licenca_id || null
    licencaSelecionadaId = licencaAtualId
    planoInicialCam = inferirPlanoLicenca(cam) || normalizarPlano(cam.plano) || null
    planoAtualCam = planoInicialCam
    valorInicialCam = cam.valor != null ? cam.valor : null
    valorAtualCam = valorInicialCam

    $('#inp-ativo').data('valor-inicial', !!cam.ativo)
    $('#inp-deteccao').data('valor-inicial', !!cam.deteccao_humano)

    $('#box-licenca-painel').addClass('d-none')
    $('#box-licenca-atual').addClass('d-none')
    $('#box-trocar-licenca').addClass('d-none')
    $('#box-liberar-licenca').addClass('d-none')
    $('#box-vincular-licenca').addClass('d-none')
    $('#lbl-trocar-licenca-vazia').addClass('d-none')
    $('#lbl-vincular-licenca-vazia').addClass('d-none')
    $('#box-selecionar-licenca').addClass('d-none')

    if (licencaAtualId) {
        aplicarEstadoPlano(cam, valorInicialCam, {
            preserveAtivo: true,
            ativo: cam.ativo,
            deteccao: cam.deteccao_humano
        })
        $('#box-licenca-painel').removeClass('d-none')
        $('#lbl-licenca-atual').text(`#${licencaAtualId} — ${labelPlano(planoInicialCam)}`)
        $('#box-licenca-atual').removeClass('d-none')
        $('#box-liberar-licenca').removeClass('d-none')
        carregarLicencasParaTroca(cam)
        configurarPainelGravacao(cam)
        return
    }

    // Camera sem licenca: permitir vincular; painel de gravacao continua disponivel
    planoAtualCam = null
    licencaSelecionadaId = null
    bloquearCamposPlano(false)
    aplicarEstadoPlano(null)
    atualizarBoxPlanoInfo(null)
    $('#box-licenca-painel').removeClass('d-none')
    carregarLicencasParaVincular()
    configurarPainelGravacao(cam)
}

function extrairErroApi(xhr) {
    const j = xhr.responseJSON || (function () {
        try { return JSON.parse(xhr.responseText || '') } catch (e) { return null }
    })()
    if (!j) return null
    return j.error || j.erro || j.message || j.status || null
}

function parseCameraResponse(r) {
    if (!r) return null
    if (r.id) return r
    if (r.dados && r.dados.id) return r.dados
    if (Array.isArray(r.dados) && r.dados[0]) return r.dados[0]
    return r
}

function nomeCliente(c) {
    return c.nome || c.nomeCliente || 'Sem nome'
}

function normalizarId(valor) {
    if (valor === null || valor === undefined) return ''
    return String(valor).trim()
}

function idsIguais(a, b) {
    const x = normalizarId(a)
    const y = normalizarId(b)
    if (!x || !y) return false
    return x.toUpperCase() === y.toUpperCase()
}

function selecionarOptionPorId($sel, id) {
    const alvo = normalizarId(id).toUpperCase()
    if (!alvo) return false
    let encontrado = ''
    $sel.find('option').each(function () {
        if (normalizarId($(this).val()).toUpperCase() === alvo) {
            encontrado = $(this).val()
            return false
        }
    })
    if (encontrado) {
        $sel.val(encontrado)
        return true
    }
    return false
}

function pendenciaSetorEdicao() {
    if (cameraEdicaoPendente) return cameraEdicaoPendente

    const idDispAtual = normalizarId($('#sel-dispositivo').val())
    if (modoForm !== 'edit' || !cameraEditCache || !dispositivoInicialEdicao) return null
    if (!idsIguais(idDispAtual, dispositivoInicialEdicao)) return null

    return {
        id_setor: normalizarId(cameraEditCache.id_setor),
        setor: cameraEditCache.setor || '',
        zonauser: cameraEditCache.zonauser != null ? String(cameraEditCache.zonauser).trim() : ''
    }
}

function selecionarOptionSetor(sel, pend) {
    if (!pend || !sel.length) return false

    if (pend.id_setor) {
        sel.find('option').each(function () {
            if (idsIguais($(this).val(), pend.id_setor)) {
                sel.val($(this).val())
                return false
            }
        })
    }

    if (!sel.val() && pend.setor) {
        const nomePend = String(pend.setor).trim().toUpperCase()
        sel.find('option').each(function () {
            if ($(this).text().trim().toUpperCase() === nomePend) {
                sel.val($(this).val())
                return false
            }
        })
    }

    if (!sel.val() && pend.zonauser) {
        const zona = String(pend.zonauser).trim()
        sel.find('option').each(function () {
            const num = $(this).data('numero')
            if (num != null && String(num).trim() === zona) {
                sel.val($(this).val())
                return false
            }
        })
    }

    return !!sel.val()
}

function resolverParticaoPayload(disp) {
    const idDispAtual = normalizarId($('#sel-dispositivo').val())

    if (modoForm === 'edit' && cameraEditCache && dispositivoInicialEdicao && idsIguais(idDispAtual, dispositivoInicialEdicao)) {
        const p = cameraEditCache.particao
        if (p !== null && p !== undefined && String(p).trim() !== '') {
            return String(p).trim()
        }
    }

    if (disp && disp.particao !== null && disp.particao !== undefined) {
        return String(disp.particao).trim()
    }
    return ''
}

function resolverContaPayload(disp) {
    const idDispAtual = normalizarId($('#sel-dispositivo').val())

    if (modoForm === 'edit' && cameraEditCache && dispositivoInicialEdicao && idsIguais(idDispAtual, dispositivoInicialEdicao)) {
        const c = cameraEditCache.conta
        if (c !== null && c !== undefined && String(c).trim() !== '') {
            return String(c).trim()
        }
    }

    if (disp && disp.conta !== null && disp.conta !== undefined) {
        return String(disp.conta).trim()
    }
    return ''
}

function buscarClienteCache(idCliente) {
    const id = normalizarId(idCliente)
    if (!id) return null
    return clientesCache.find(function (c) {
        return normalizarId(c.idCliente) === id
    }) || null
}

function atualizarLabelCliente(idCliente, nomePreferido) {
    const id = normalizarId(idCliente)
    let nome = nomePreferido || ''
    if (!nome && id) {
        const cliente = buscarClienteCache(id)
        nome = cliente ? nomeCliente(cliente) : id
    }
    clienteSelecionadoId = id || null
    if (id) {
        $('#lbl-cliente-nome').text(`Cliente: ${nome}`)
    } else if (nome) {
        $('#lbl-cliente-nome').text(nome)
    } else {
        $('#lbl-cliente-nome').text('Cliente não selecionado')
    }
    $('#btn-trocar-cliente').removeClass('d-none')
    renderListaClientes($('#busca-cliente').val().trim())
}

function iniciarTrocaCliente() {
    const panel = document.querySelector('.cv-form-clientes')
    if (panel) {
        panel.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
        panel.classList.add('cv-form-clientes-highlight')
        window.setTimeout(function () {
            panel.classList.remove('cv-form-clientes-highlight')
        }, 2200)
    }
    $('#busca-cliente').val('').trigger('input')
    window.setTimeout(function () {
        $('#busca-cliente').trigger('focus')
    }, 350)
}

function mostrarFormCamera(mostrar) {
    if (mostrar) {
        $('#box-form-placeholder').addClass('d-none')
        $('#box-form-camera').removeClass('d-none')
        $('#box-form-actions').removeClass('d-none')
    } else {
        $('#box-form-placeholder').removeClass('d-none')
        $('#box-form-camera').addClass('d-none')
        $('#box-form-actions').addClass('d-none')
        $('#lbl-cliente-nome').text('')
        $('#btn-trocar-cliente').addClass('d-none')
    }
}

function tipoCameraSelecionado() {
    return ConfVisionUrls.normalizarTipoCamera($('#sel-tipo-camera').val())
}

function exibirLinkStream(cameraId) {
    if (!cameraId) {
        configurarAbaUrl(false)
        configurarAbaSnapshot(false)
        configurarAbaArea(false)
        configurarAbaAoVivo(false)
        return
    }
    const tipo = tipoCameraSelecionado()
    if (tipo === 'dvr') {
        $('#lbl-rtmp-instrucao').text('Cole no campo RTMP do Cabeado IP (com barra / no final; path Hashids, sem ?pass=):')
    } else if (tipo === 'wifi') {
        $('#lbl-rtmp-instrucao').text('Cole no campo RTMP da câmera Wifi (sem barra /; path Hashids, sem ?pass=):')
    } else {
        $('#lbl-rtmp-instrucao').text('Selecione o tipo (Cabeado IP ou Wifi) na aba Dados para montar a URL correta:')
    }
    $('#inp-rtmp-url').val('Carregando URL autenticada…')
    $('#lbl-stream-path').text('…')
    configurarAbaUrl(true)
    configurarAbaSnapshot(true)
    configurarAbaArea(true)
    configurarAbaAoVivo(true)

    ConfVisionUrls.carregarRtmpPublish(cameraId)
        .done(function (r) {
            const url = tipo === 'dvr' ? (r.dvr || r.url) : (r.wifi || r.url)
            $('#inp-rtmp-url').val(url || '')
            $('#lbl-stream-path').text(r.path || '')
            if (r.bloqueado) {
                $('#lbl-rtmp-instrucao').text('Câmera BLOQUEADA — publish RTMP será rejeitado até desbloquear.')
            }
        })
        .fail(function (xhr) {
            const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Falha ao gerar URL (RTMP_PUBLISH_SECRET?)'
            $('#inp-rtmp-url').val('')
            $('#lbl-rtmp-instrucao').text(msg)
        })
}

function atualizarBotoesSnapshot(temSnapshot) {
    $('#btn-remover-snapshot').toggleClass('d-none', !temSnapshot)
}

function setSnapshotBusy(busy, modo) {
    snapshotBusy = !!busy
    const $tirar = $('#btn-tirar-snapshot')
    const $remover = $('#btn-remover-snapshot')

    if (busy) {
        $tirar.prop('disabled', true).addClass('cv-btn-disabled')
        $remover.prop('disabled', true).addClass('cv-btn-disabled')
        if (modo === 'tirar') {
            $tirar.html('<i class="bi bi-arrow-repeat cv-snapshot-spin"></i> Capturando...')
            $('#lbl-snapshot-status').text('Capturando snapshot...')
        } else if (modo === 'remover') {
            $remover.html('<i class="bi bi-arrow-repeat cv-snapshot-spin"></i> Removendo...')
            $('#lbl-snapshot-status').text('Removendo snapshot...')
        }
        return
    }

    $tirar
        .prop('disabled', false)
        .removeClass('cv-btn-disabled')
        .html('<i class="bi bi-camera-fill"></i> Tirar snapshot')
    $remover
        .prop('disabled', false)
        .removeClass('cv-btn-disabled')
        .html('<i class="bi bi-trash"></i> Remover snapshot')
}

function exibirSnapshotPreview(url) {
    if (!url) {
        snapshotUrlAtual = null
        $('#box-snapshot-preview').addClass('d-none')
        $('#box-snapshot-preview-tab').addClass('d-none')
        $('#img-snapshot').attr('src', '')
        $('#img-snapshot-tab').attr('src', '')
        atualizarBotoesSnapshot(false)
        if (typeof ConfVisionAreaEditor !== 'undefined') {
            ConfVisionAreaEditor.onSnapshotChange(false)
        }
        return
    }
    const sep = url.includes('?') ? '&' : '?'
    const src = url + sep + 't=' + Date.now()
    $('#img-snapshot').attr('src', src)
    $('#img-snapshot-tab').attr('src', src)
    $('#box-snapshot-preview').removeClass('d-none')
    $('#box-snapshot-preview-tab').removeClass('d-none')
    atualizarBotoesSnapshot(true)
    if (typeof ConfVisionAreaEditor !== 'undefined') {
        ConfVisionAreaEditor.onSnapshotChange(true)
    }
}

function removerSnapshot() {
    if (snapshotBusy) return
    if (!cameraEditId) {
        boxMesagemAtencaoPersonalizada('Salve a câmera antes de remover o snapshot')
        return
    }
    if (!snapshotUrlAtual) return
    CvMsg.confirmar('Remover snapshot?', 'A área de detecção ficará indisponível até capturar um novo snapshot.').then(function (r) {
        if (!r.isConfirmed) return
        setSnapshotBusy(true, 'remover')

    $.ajax({
        url: `/api/cameras/${cameraEditId}/snapshot`,
        method: 'DELETE',
        contentType: 'application/json',
        data: JSON.stringify({
            id_franqueado: ConfVisionUrls.idFranqueado()
        })
    }).fail(function (xhr) {
        console.error(xhr)
        let msg = 'Erro ao remover snapshot'
        const j = xhr.responseJSON || (function () {
            try { return JSON.parse(xhr.responseText || '') } catch (e) { return null }
        })()
        if (j) {
            if (j.status) msg = j.status
            else if (j.erro) msg = j.erro
            else if (j.message) msg = j.message
        }
        boxMesagemAtencaoPersonalizada(msg)
        $('#lbl-snapshot-status').text('')
    }).done(function () {
        exibirSnapshotPreview('')
        $('#lbl-snapshot-status').text('Snapshot removido')
        boxSucessoAuto('Snapshot removido')
    }).always(function () {
        setSnapshotBusy(false)
    })
    })
}

function tirarSnapshot() {
    if (snapshotBusy) return
    if (!cameraEditId) {
        boxMesagemAtencaoPersonalizada('Salve a câmera antes de tirar o snapshot')
        return
    }
    if (cameraEditCache && !confVisionCameraPodeStream(cameraEditCache).ok) {
        boxMesagemAtencaoPersonalizada(confVisionMotivoStream(cameraEditCache))
        return
    }

    setSnapshotBusy(true, 'tirar')

    $.ajax({
        url: `/api/cameras/${cameraEditId}/snapshot`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            id_franqueado: ConfVisionUrls.idFranqueado()
        })
    }).fail(function (xhr) {
        console.error(xhr)
        let msg = 'Erro ao capturar snapshot'
        const j = xhr.responseJSON || (function () {
            try { return JSON.parse(xhr.responseText || '') } catch (e) { return null }
        })()
        if (j) {
            if (j.status) msg = j.status
            else if (j.erro) msg = j.erro
            else if (j.message) msg = j.message
        }
        boxMesagemAtencaoPersonalizada(msg)
        $('#lbl-snapshot-status').text('')
    }).done(function (r) {
        const url = (r && r.snapshot_url) || (r.dados && r.dados.snapshot_url) || ''
        snapshotUrlAtual = url
        exibirSnapshotPreview(url)
        $('#lbl-snapshot-status').text('Snapshot atualizado — configure a área na aba Área de detecção')
        boxSucessoAuto('Snapshot capturado com sucesso')
        if (planoPermiteAreaDeteccao(planoAtualForm())) {
            selecionarAbaForm('area')
        }
    }).always(function () {
        setSnapshotBusy(false)
    })
}

function copiarLinkRtmp() {
    const url = $('#inp-rtmp-url').val()
    if (!url) return

    function ok() {
        boxSucessoAuto('URL copiada!')
    }

    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(url).then(ok).catch(function () {
            copiarLinkRtmpFallback(url, ok)
        })
    } else {
        copiarLinkRtmpFallback(url, ok)
    }
}

function copiarLinkRtmpFallback(texto, ok) {
    const el = document.createElement('textarea')
    el.value = texto
    el.setAttribute('readonly', '')
    el.style.position = 'absolute'
    el.style.left = '-9999px'
    document.body.appendChild(el)
    el.select()
    try {
        document.execCommand('copy')
        ok()
    } catch (e) {
        boxMesagemAtencaoPersonalizada('Não foi possível copiar. Selecione e copie manualmente.')
    }
    document.body.removeChild(el)
}

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    mostrarFormCamera(false)

    const editId = ($('#cfg-camera-id').val() || '').trim()
    modoForm = editId ? 'edit' : 'create'
    cameraEditId = editId || null

    $('#btn-salvar-camera').on('click', salvarCamera)
    $('#btn-copiar-rtmp').on('click', copiarLinkRtmp)
    $('#btn-ajuda-encode-rtmp, #btn-ajuda-encode-rtmp-acoes').on('click', function () {
        const id = cameraEditId || ($('#cfg-camera-id').val() || '').trim()
        confVisionAbrirAjudaEncodeRtmp(id)
    })
    $('#sel-tipo-camera').on('change', function () {
        const id = cameraEditId || ($('#cfg-camera-id').val() || '').trim()
        if (id) exibirLinkStream(id)
    })
    $('#btn-tirar-snapshot').on('click', tirarSnapshot)
    $('#btn-remover-snapshot').on('click', removerSnapshot)
    $('#cv-form-tabs').on('click', '.nav-link', function () {
        const tab = $(this).attr('data-tab')
        if (!tab || $(this).closest('.nav-item').hasClass('d-none')) return
        selecionarAbaForm(tab)
    })
    $('#lbl-rtmp-instrucao').text('Cole no campo de publicação RTMP (Wifi — sem barra no final):')
    $('#sel-dispositivo').on('change', function () {
        onDispositivoChange()
    })
    $('#inp-captura-analitico').on('change', function () {
        if (typeof ConfVisionAreaEditor !== 'undefined') {
            ConfVisionAreaEditor.refreshEditorVisibility()
        }
    })
    $('#sel-licenca').on('change', onLicencaChange)
    $('#sel-trocar-licenca').on('change', onTrocarLicencaChange)
    $('#sel-vincular-licenca').on('change', onVincularLicencaChange)
    $('#btn-liberar-licenca-cam').on('click', liberarLicencaCamForm)
    $('#sel-gravacao-licenca, #sel-trocar-gravacao-licenca').on('change', function () {
        atualizarBotoesGravacaoPainel(temLicencaGravacaoCam(cameraEditCache))
    })
    $('#btn-ativar-gravacao-cam').on('click', ativarGravacaoCamForm)
    $('#btn-trocar-gravacao-cam').on('click', trocarGravacaoCamForm)
    $('#btn-desativar-gravacao-cam').on('click', desativarGravacaoCamForm)
    $('#btn-trocar-cliente').on('click', iniciarTrocaCliente)
    $('#busca-cliente').on('input', function () {
        renderListaClientes($(this).val().trim())
    })
    $('#lista-clientes').on('click', '.cv-cliente-item', function () {
        selecionarCliente($(this).attr('data-id'))
    })

    carregarClientes().done(function () {
        if (editId) {
            cameraEditId = editId
            modoForm = 'edit'
            mostrarAbaGravacao(true)
            selecionarAbaForm('dados')
            carregarCameraParaEdicao(editId)
        } else {
            selecionarAbaForm('licenca')
            carregarLicencasDisponiveis()
            ocultarPainelGravacao()
            configurarAbaUrl(false)
            configurarAbaSnapshot(false)
            configurarAbaArea(false)
            configurarAbaAoVivo(false)
        }
    })
})

function carregarClientes() {
    $('#lista-clientes').html('<div class="cv-empty py-3">Carregando clientes...</div>')

    return $.ajax({
        url: '/api/clientes',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            idFranqueado: ConfVisionUrls.idFranqueado()
        })
    }).fail(function (xhr) {
        console.error(xhr)
        $('#lista-clientes').html('<div class="cv-empty py-3 text-danger">Erro ao carregar clientes</div>')
        boxMesagemAtencaoPersonalizada('Erro ao carregar clientes do franqueado')
    }).done(function (r) {
        const bruto = r.dados || r || []
        clientesCache = Array.isArray(bruto) ? bruto : []
        renderListaClientes('')
    })
}

function renderListaClientes(filtro) {
    const termo = (filtro || '').toLowerCase()
    const box = $('#lista-clientes').empty()

    const lista = clientesCache.filter(function (c) {
        if (!termo) return true
        const nome = nomeCliente(c).toLowerCase()
        const id = String(c.idCliente || '').toLowerCase()
        return nome.includes(termo) || id.includes(termo)
    })

    if (!lista.length) {
        const msg = clientesCache.length
            ? 'Nenhum cliente encontrado na busca'
            : 'Nenhum cliente para este franqueado'
        box.append(`<div class="cv-empty py-3">${msg}</div>`)
        return
    }

    lista.forEach(function (c) {
        const id = normalizarId(c.idCliente)
        const inativo = c.ativo !== 'S' && c.ativo !== true
        const ativo = id === normalizarId(clienteSelecionadoId) ? ' cv-cliente-item-active' : ''
        const inativoCls = inativo ? ' cv-cliente-item-inativo' : ''
        box.append(`
            <button type="button" class="cv-cliente-item${ativo}${inativoCls}" data-id="${id}">
                <span class="cv-cliente-nome">${nomeCliente(c)}${inativo ? ' (inativo)' : ''}</span>
                <span class="cv-cliente-id">${id}</span>
            </button>
        `)
    })
}

function aplicarCamposExtras(cam) {
    if (!cam) return
    const tipo = ConfVisionUrls.normalizarTipoCamera(cam.protocolo)
    $('#sel-tipo-camera').val(tipo || '')
    $('#inp-canal').val(cam.canal || '')
}

function carregarCameraParaEdicao(id) {
    $.get(`/api/cameras/${id}`)
        .fail(function (xhr) {
            console.error(xhr)
            boxMesagemAtencaoPersonalizada('Erro ao carregar câmera para edição')
        })
        .done(function (r) {
            const cam = parseCameraResponse(r)
            if (!cam || !cam.id) {
                boxMesagemAtencaoPersonalizada('Câmera não encontrada')
                return
            }

            cameraEditId = cam.id
            cameraEditCache = cam
            dispositivoInicialEdicao = normalizarId(cam.id_dispositivo)
            $('#form-camera-titulo').text(`Editar câmera #${cam.id}`)
            $('#inp-nome').val(cam.nome || '')
            $('#inp-confianca').val(cam.confianca_min != null ? cam.confianca_min : 0.5)
            $('#inp-cooldown').val(cam.cooldown_seg != null ? cam.cooldown_seg : 30)
            $('#sel-modo-deteccao').val(cam.modo_deteccao || 'dentro')
            if (typeof ConfVisionAreaEditor !== 'undefined' && ConfVisionAreaEditor.atualizarModoDeteccaoHint) {
                ConfVisionAreaEditor.atualizarModoDeteccaoHint()
            }
            aplicarCamposExtras(cam)
            configurarPlanoEdicao(cam)
            cameraEdicaoPendente = {
                id_setor: normalizarId(cam.id_setor),
                setor: cam.setor || '',
                zonauser: cam.zonauser != null ? String(cam.zonauser).trim() : ''
            }
            exibirLinkStream(cam.id)
            snapshotUrlAtual = cam.snapshot_url || null
            exibirSnapshotPreview(snapshotUrlAtual || '')
            configurarAbaSnapshot(true)
            configurarAbaArea(true)
            configurarAbaAoVivo(true)
            configurarAbaUrl(true)

            if (cam.id_cliente) {
                selecionarCliente(normalizarId(cam.id_cliente), normalizarId(cam.id_dispositivo))
            } else {
                mostrarFormCamera(true)
                atualizarLabelCliente('', 'Cliente não vinculado à câmera')
            }

            if (typeof ConfVisionAreaEditor !== 'undefined') {
                try {
                    ConfVisionAreaEditor.setCameraId(cam.id)
                } catch (e) {
                    console.error('Erro ao iniciar editor de area', e)
                }
            }
        })
}

function selecionarCliente(idCliente, idDispositivoPre) {
    const id = normalizarId(idCliente)
    atualizarLabelCliente(id)
    mostrarFormCamera(true)

    $('#sel-dispositivo').empty().append('<option value="">Carregando...</option>').prop('disabled', true)
    $('#sel-setor').empty().append('<option value="">Selecione o dispositivo</option>').prop('disabled', true)
    dispositivoSelecionado = null

    $.ajax({
        url: '/api/dispositivos',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idCliente: id })
    }).fail(function (xhr) {
        console.error(xhr)
        $('#sel-dispositivo').empty().append('<option value="">Erro ao carregar</option>')
        boxMesagemAtencaoPersonalizada('Erro ao carregar dispositivos do cliente')
    }).done(function (r) {
        const lista = r.dados || []
        if (!buscarClienteCache(id) && lista.length) {
            const nomeDisp = lista[0].nomeCliente || lista[0].NomeCliente
            if (nomeDisp) atualizarLabelCliente(id, nomeDisp)
        }
        const sel = $('#sel-dispositivo').empty().append('<option value="">Selecione...</option>')
        lista.forEach(function (d) {
            const label = d.nome || d.nomeDispositivo || d.conta || d.idDispositivo
            sel.append(`<option value="${normalizarId(d.idDispositivo)}">${label}</option>`)
        })
        sel.prop('disabled', false)
        if (idDispositivoPre) {
            if (!selecionarOptionPorId(sel, idDispositivoPre)) {
                sel.val(normalizarId(idDispositivoPre))
            }
            onDispositivoChange()
        }
    })
}

function onDispositivoChange() {
    const idDispositivo = $('#sel-dispositivo').val()
    $('#sel-setor').empty().append('<option value="">Carregando...</option>').prop('disabled', true)
    dispositivoSelecionado = null
    if (!idDispositivo) return

    if (!idsIguais(idDispositivo, dispositivoInicialEdicao)) {
        cameraEdicaoPendente = null
    }

    $.ajax({
        url: '/api/dispositivo-dados',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).done(function (r) {
        dispositivoSelecionado = r.dados || r
        carregarSetores(idDispositivo)
    })
}

function normalizarListaCamerasSetor(r) {
    if (Array.isArray(r)) return r
    if (r && Array.isArray(r.dados)) return r.dados
    return []
}

function obterSetorAtualEdicao(idDispositivo) {
    if (modoForm !== 'edit' || !cameraEditCache) return ''
    if (!idsIguais(idDispositivo, dispositivoInicialEdicao)) return ''
    return normalizarId(cameraEditCache.id_setor)
}

function calcularSetoresOcupados(cameras, idDispositivo) {
    const idDisp = normalizarId(idDispositivo).toUpperCase()
    const ocupados = new Set()
    cameras.forEach(function (cam) {
        if (normalizarId(cam.id_dispositivo).toUpperCase() !== idDisp) return
        if (cameraEditId && String(cam.id) === String(cameraEditId)) return
        const idSetor = normalizarId(cam.id_setor)
        if (idSetor) ocupados.add(idSetor.toUpperCase())
    })
    return ocupados
}

function atualizarAvisoSetores(vazio) {
    $('#lbl-setor-vazio').toggleClass('d-none', !vazio)
}

function renderSetores(idDispositivo, listaSetores, cameras) {
    const ocupados = calcularSetoresOcupados(cameras, idDispositivo)
    setoresOcupadosCache = ocupados
    const setorAtualEdicao = obterSetorAtualEdicao(idDispositivo)

    const sel = $('#sel-setor').empty().append('<option value="">Selecione...</option>')
    let disponiveis = 0
    listaSetores.forEach(function (s) {
        const idSetor = normalizarId(s.idSetor || s.setorId)
        const ehAtual = setorAtualEdicao && idsIguais(idSetor, setorAtualEdicao)
        if (idSetor && ocupados.has(idSetor.toUpperCase()) && !ehAtual) return
        const nome = s.setorNome || s.nome || idSetor
        const numero = s.numero != null ? String(s.numero) : ''
        sel.append(
            `<option value="${idSetor}" data-numero="${numero.replace(/"/g, '&quot;')}">${nome}</option>`
        )
        disponiveis++
    })

    sel.prop('disabled', disponiveis === 0)
    atualizarAvisoSetores(disponiveis === 0)
    aplicarSetorEdicao()
}

function carregarSetores(idDispositivo) {
    const idFranqueado = ConfVisionUrls.idFranqueado()

    $.ajax({
        url: '/api/setores',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).done(function (r) {
        const listaSetores = (r && r.dados) || []
        $.get(`/api/cameras?id_franqueado=${encodeURIComponent(idFranqueado)}`)
            .done(function (rc) {
                renderSetores(idDispositivo, listaSetores, normalizarListaCamerasSetor(rc))
            })
            .fail(function () {
                renderSetores(idDispositivo, listaSetores, [])
            })
    })
}

function aplicarSetorEdicao() {
    const pend = pendenciaSetorEdicao()
    if (!pend) return

    const sel = $('#sel-setor')
    if (selecionarOptionSetor(sel, pend)) {
        cameraEdicaoPendente = null
    }
}

function montarPayload() {
    const idCliente = clienteSelecionadoId
    const idDispositivo = $('#sel-dispositivo').val()
    const nome = $('#inp-nome').val().trim()
    const disp = dispositivoSelecionado || {}
    const setorOpt = $('#sel-setor option:selected')
    const idSetor = normalizarId(setorOpt.val())
    const zonaNumero = setorOpt.data('numero')
    const zonauser = zonaNumero != null && String(zonaNumero) !== '' ? String(zonaNumero) : ''

    const tipoCamera = tipoCameraSelecionado()
    if (!tipoCamera) {
        return { erro: 'Selecione o tipo da câmera (Cabeado IP ou Wifi)' }
    }

    const payload = {
            ativo: planoSemAtivo(planoAtualForm()) ? false : $('#inp-ativo').is(':checked'),
            nome: nome,
            id_franqueado: ConfVisionUrls.idFranqueado(),
            id_cliente: idCliente,
            id_dispositivo: idDispositivo,
            conta: resolverContaPayload(disp),
            particao: resolverParticaoPayload(disp),
            protocolo: tipoCamera,
            canal: ($('#inp-canal').val() || '').trim(),
            setor: setorOpt.text() && idSetor ? setorOpt.text().trim() : '',
            id_setor: idSetor,
            zonauser: zonauser,
            confianca_min: parseFloat($('#inp-confianca').val()) || 0.5,
            cooldown_seg: parseInt($('#inp-cooldown').val(), 10) || 30,
            modo_deteccao: ($('#sel-modo-deteccao').val() || 'dentro'),
            somente_armado: planoEhOnline(planoAtualForm()) ? false : $('#inp-somente-armado').is(':checked'),
            deteccao_humano: planoEhOnline(planoAtualForm()) ? false : $('#inp-deteccao').is(':checked'),
            deteccao_veiculo: false,
            captura_sensor: planoEhOnline(planoAtualForm()) ? false : $('#inp-captura-sensor').is(':checked'),
            captura_analitico: planoEhOnline(planoAtualForm()) ? false : $('#inp-captura-analitico').is(':checked'),
            status: 'offline'
        }

    if (snapshotUrlAtual) {
        payload.snapshot_url = snapshotUrlAtual
    }

    if (modoForm === 'create') {
        if (!licencaSelecionadaId) {
            return { erro: 'Selecione uma licença disponível antes de salvar' }
        }
        payload.vis_licenca_id = parseInt(licencaSelecionadaId, 10)
    } else if (licencaSelecionadaId && String(licencaSelecionadaId) !== String(licencaAtualId || '')) {
        payload.vis_licenca_id = parseInt(licencaSelecionadaId, 10)
    }

    return {
        payload: payload,
        idCliente: idCliente,
        idDispositivo: idDispositivo,
        nome: nome
    }
}

function salvarCamera() {
    const $btn = $('#btn-salvar-camera')
    if ($btn.prop('disabled')) return

    let dados
    try {
        dados = montarPayload()
    } catch (err) {
        console.error('montarPayload', err)
        boxMesagemAtencaoPersonalizada('Erro ao preparar dados da câmera. Atualize a página (Ctrl+F5).')
        return
    }

    if (dados.erro) {
        boxMesagemAtencaoPersonalizada(dados.erro)
        return
    }
    if (!dados.idCliente || !dados.idDispositivo || !dados.nome) {
        boxMesagemAtencaoPersonalizada('Selecione cliente, dispositivo e informe o nome')
        return
    }
    if (!dados.payload.protocolo) {
        boxMesagemAtencaoPersonalizada('Selecione o tipo da câmera (Cabeado IP ou Wifi)')
        return
    }
    if (!dados.payload.id_setor) {
        boxMesagemAtencaoPersonalizada('Selecione o setor da câmera')
        return
    }
    const idSetorSel = normalizarId(dados.payload.id_setor).toUpperCase()
    if (idSetorSel && setoresOcupadosCache.has(idSetorSel)) {
        boxMesagemAtencaoPersonalizada('Este setor já possui uma câmera neste dispositivo')
        return
    }

    const isEdit = modoForm === 'edit' && cameraEditId
    const desativandoAnalitico = isEdit
        && !planoSemAtivo(planoAtualForm())
        && $('#inp-ativo').data('valor-inicial') === true
        && !dados.payload.ativo
    const tinhaGravacao = cameraEditCache && String(cameraEditCache.gravacao_status || '').toLowerCase() === 'ativa'
    function enviarSalvar() {
        const url = isEdit ? `/api/cameras/${cameraEditId}` : '/api/cameras'
        const method = isEdit ? 'PUT' : 'POST'

        if (isEdit) {
            dados.payload.vis_camera_id = parseInt(cameraEditId, 10)
        }

        $btn.prop('disabled', true)

        $.ajax({
            url: url,
            method: method,
            contentType: 'application/json',
            data: JSON.stringify(dados.payload)
        }).fail(function (e) {
            console.error(e)
            const msgApi = extrairErroApi(e)
            boxMesagemAtencaoPersonalizada(msgApi || (isEdit ? 'Erro ao atualizar câmera' : 'Erro ao salvar câmera'))
        }).done(function (r) {
            const cam = parseCameraResponse(r)
            const id = isEdit ? cameraEditId : (cam && cam.id)
            if (!id) {
                boxMesagemAtencaoPersonalizada('Câmera salva, mas não foi possível obter o ID')
                return
            }

            if (isEdit) {
                if (typeof ConfVisionLivePlayer !== 'undefined') {
                    ConfVisionLivePlayer.stop()
                }
                try {
                    sessionStorage.setItem('cv_cameras_flash', 'Câmera #' + id + ' atualizada com sucesso.')
                } catch (e) { /* ignore */ }
                window.location.href = '/cameras'
                return
            }

            cameraEditId = id
            cameraEditCache = cam || null
            dispositivoInicialEdicao = normalizarId((cam && cam.id_dispositivo) || dados.idDispositivo)
            modoForm = 'edit'
            $('#form-camera-titulo').text(`Editar câmera #${id}`)
            exibirLinkStream(id)
            snapshotUrlAtual = (cam && cam.snapshot_url) ? cam.snapshot_url : null
            exibirSnapshotPreview(snapshotUrlAtual || '')
            if (typeof ConfVisionAreaEditor !== 'undefined') {
                ConfVisionAreaEditor.setCameraId(id)
            }
            if (window.history && window.history.replaceState) {
                window.history.replaceState(null, '', `/cameras/editar/${id}`)
            }

            configurarPlanoEdicao(cam || {})
            boxSucessoAuto('Câmera cadastrada! Copie a URL RTMP na aba URL de Transmissão.')
            selecionarAbaForm('url')
        }).always(function () {
            $btn.prop('disabled', false)
        })
    }

    if (desativandoAnalitico && tinhaGravacao) {
        CvMsg.confirmar('Desativar câmera?', 'Isso encerra a gravação (timelapse/movimento/contínua) e libera a licença de gravação.').then(function (r) {
            if (!r.isConfirmed) return
            enviarSalvar()
        })
        return
    }

    enviarSalvar()
}

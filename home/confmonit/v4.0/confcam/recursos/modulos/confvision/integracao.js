let modalMidiaIntegracao = null

const INT_LOG_PAGE_SIZE = 15
let intLogOffset = 0
let intLogCarregando = false
let intLogTemMais = true

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return
    confVisionAuthGuard()
    confVisionCarregarCabecalho()

    const elModal = document.getElementById('modal-evento-midia')
    if (elModal && window.bootstrap) {
        modalMidiaIntegracao = new bootstrap.Modal(elModal)
    }

    $('#btn-int-log-buscar').on('click', function () { carregarLogIntegracao(true) })
    $('#btn-int-log-limpar').on('click', function () {
        limparFiltrosLogIntegracao()
        carregarLogIntegracao(true)
    })
    $('#int-log-de, #int-log-ate, #int-log-cliente').on('keydown', function (e) {
        if (e.key === 'Enter') {
            e.preventDefault()
            carregarLogIntegracao(true)
        }
    })
    $('#int-log-scroll').on('scroll', onScrollLogIntegracao)
    $('#int-log-body').on('click', '.btn-int-log-midia', function () {
        abrirMidiaEventoIntegracao($(this).data('id'))
    })

    $('#int-sistema').on('change', function () {
        $('#int-ativo').data('usuarioAlterou', false)
        atualizarPainelSistema()
        atualizarAuthCampos()
    })
    $('#int-ativo').on('change', function () {
        $('#int-ativo').data('usuarioAlterou', true)
    })
    $('#int-dupla-comunicacao').on('change', function () {
        $('#int-dupla-comunicacao').data('usuarioAlterou', true)
    })
    $('#btn-int-salvar, #btn-int-salvar-rodape').on('click', salvarIntegracao)
    $('#btn-int-testar, #btn-int-testar-rodape').on('click', testarIntegracao)
    $('#int-moni-tabs').on('click', '[data-tab]', function () {
        selecionarAbaMoni($(this).attr('data-tab'))
    })
    $('#int-auth-tipo').on('change', atualizarAuthCampos)
    $('#int-enviar-imagem').on('change', atualizarEnviarImagemCampos)
    $('#int-protocolo, #int-host, #int-porta, #int-endpoint').on('input change', atualizarUrlPreview)

    carregarIntegracao()
    carregarLogIntegracao(true)
})

let integracaoMeta = {}

function parseRespostaIntegracao(r) {
    if (!r || typeof r !== 'object') return { integracao: {}, meta: {} }
    if (r.integracao || r.meta) return { integracao: r.integracao || {}, meta: r.meta || {} }
    if (r.dados && (r.dados.integracao || r.dados.meta)) {
        return { integracao: r.dados.integracao || {}, meta: r.dados.meta || {} }
    }
    return { integracao: {}, meta: {} }
}

function sistemaPermiteDuplaComunicacao(sistema) {
    return sistema === 'moni' || sistema === 'dguard' || sistema === 'segware'
}

function carregarIntegracao() {
    $.get('/api/integracao')
        .done(function (r) {
            const parsed = parseRespostaIntegracao(r)
            preencherFormIntegracao(parsed.integracao, parsed.meta)
            atualizarPainelSistema()
            atualizarAuthCampos()
        })
        .fail(function (xhr) {
            boxErro(mensagemErroAjax(xhr, 'Erro ao carregar integração'))
        })
}

function preencherFormIntegracao(item, meta) {
    integracaoMeta = meta || {}
    $('#int-id').val(item.id || '')
    $('#int-sistema').val(item.sistema || 'nenhum')
    $('#int-ativo').prop('checked', !!item.ativo)
    $('#int-ativo').data('usuarioAlterou', false)
    $('#int-dupla-comunicacao').prop('checked', !!item.dupla_comunicacao)
    $('#int-dupla-comunicacao').data('usuarioAlterou', false)
    $('#int-nome').val(item.nome || '')

    preencherConexaoFromUrl(item.eventos_url || '')

    $('#int-auth-tipo').val(item.auth_tipo || 'none')
    $('#int-auth-user').val(item.auth_user || '')
    $('#int-auth-pass').val(item.auth_pass || item.auth_token || '')
    $('#int-empresa').val(item.empresa_codigo || '')
    $('#int-evento-codigo').val(item.evento_codigo || '')
    $('#int-webhook-inbound').prop('checked', !!item.webhook_inbound_ativo)
    $('#int-codigo-armar').val(item.codigo_evento_armar || '130')
    $('#int-codigo-desarmar').val(item.codigo_evento_desarmar || '131')
    $('#int-identificacao').val(item.identificacao_padrao || 'E')
    $('#int-setor').val(item.setor_padrao || '1')
    $('#int-particao').val(item.particao_padrao || '01')
    $('#int-enviar-imagem').prop('checked', item.enviar_imagem !== false)

    renderImagemInfo(meta)
    atualizarEnviarImagemCampos()
    atualizarWebhookTokenHint(item)
}

function atualizarWebhookTokenHint(item) {
    const id = (item && item.id_franqueado) ||
        (typeof ConfVisionUrls !== 'undefined' ? ConfVisionUrls.idFranqueado() : '') ||
        '—'
    $('#int-webhook-token').text(id)
}

function preencherConexaoFromUrl(url) {
    const conn = parseEventosUrl(url)
    $('#int-protocolo').val(conn.protocolo)
    $('#int-host').val(conn.host)
    $('#int-porta').val(conn.porta)
    $('#int-endpoint').val(conn.endpoint)
    atualizarUrlPreview()
}

function parseEventosUrl(url) {
    const def = { protocolo: 'http', host: '', porta: '', endpoint: 'GerarEvento' }
    const raw = String(url || '').trim()
    if (!raw) return def

    let protocolo = 'http'
    let rest = raw
    const protoMatch = raw.match(/^(https?|tcp):\/\//i)
    if (protoMatch) {
        protocolo = protoMatch[1].toLowerCase()
        rest = raw.slice(protoMatch[0].length)
    }

    let host = ''
    let porta = ''
    let endpoint = def.endpoint
    const slashIdx = rest.indexOf('/')
    const hostPort = slashIdx >= 0 ? rest.slice(0, slashIdx) : rest
    const path = slashIdx >= 0 ? rest.slice(slashIdx + 1) : ''
    if (path) {
        endpoint = path.replace(/^\/+/, '').split('?')[0] || def.endpoint
    }

    const colonIdx = hostPort.lastIndexOf(':')
    if (colonIdx > 0 && /^\d+$/.test(hostPort.slice(colonIdx + 1))) {
        host = hostPort.slice(0, colonIdx)
        porta = hostPort.slice(colonIdx + 1)
    } else {
        host = hostPort
    }

    return { protocolo, host, porta, endpoint }
}

function montarEventosUrl() {
    const protocolo = ($('#int-protocolo').val() || 'http').toLowerCase()
    const host = String($('#int-host').val() || '').trim()
    const porta = String($('#int-porta').val() || '').trim()
    const endpoint = String($('#int-endpoint').val() || '').trim().replace(/^\/+/, '')
    if (!host) return ''
    let url = protocolo + '://' + host
    if (porta) url += ':' + porta
    if (endpoint) url += '/' + endpoint
    return url
}

function atualizarUrlPreview() {
    const url = montarEventosUrl()
    const $box = $('#int-url-preview')
    if (!url) {
        $box.removeClass('text-muted').addClass('text-muted').text('—')
        return
    }
    $box.removeClass('text-muted').html('<code>' + escapeHtml(url) + '</code>')
}

function sistemaPrecisaCamposConfig(sistema) {
    return sistema === 'moni' || sistema === 'dguard' || sistema === 'segware'
}

function renderImagemInfo(meta) {
    const base = (meta.imagem_public_base_url || 'https://imagem.dnsid.com.br').replace(/\/$/, '')
    $('#int-imagem-info').html(
        `<div class="cv-int-url-box"><code>${escapeHtml(base)}/</code></div>`
    )
}

function atualizarEnviarImagemCampos() {
    if ($('#int-enviar-imagem').is(':checked')) {
        $('#int-imagem-url-row').removeClass('d-none')
    } else {
        $('#int-imagem-url-row').addClass('d-none')
    }
}

function atualizarConfmonitLicenca() {
    const sistema = $('#int-sistema').val() || 'nenhum'
    const temLicenca = integracaoMeta.confmonit_licenca_ok === true
    const $aviso = $('#int-confmonit-licenca-aviso')
    const bloquear = sistema === 'confmonit' && !temLicenca

    if (bloquear) {
        $aviso.removeClass('d-none')
        $('#btn-int-salvar, #btn-int-salvar-rodape').prop('disabled', true)
    } else {
        $aviso.addClass('d-none')
        $('#btn-int-salvar, #btn-int-salvar-rodape').prop('disabled', false)
    }
}

function selecionarAbaMoni(tab) {
    const alvo = String(tab || 'geral')
    $('#int-moni-tabs .nav-link').removeClass('active')
    $('#int-moni-tabs .nav-link[data-tab="' + alvo + '"]').addClass('active')
    $('#int-config-sections .cv-form-tab-panel').addClass('d-none')
    $('#int-config-sections .cv-form-tab-panel[data-tab-panel="' + alvo + '"]').removeClass('d-none')
}

function reposicionarCampoNome(sistema) {
    const $input = $('#int-nome')
    const usaPainelConfig = sistemaPrecisaCamposConfig(sistema)

    if (usaPainelConfig) {
        $('#int-nome-row-top').addClass('d-none')
        $('#int-nome-row-geral').removeClass('d-none')
        $('#int-nome-slot-geral').append($input)
        return
    }

    $('#int-nome-row-geral').addClass('d-none')
    $('#int-nome-row-top').removeClass('d-none')
    $('#int-nome-row-top').append($input)
}

function atualizarDuplaComunicacao() {
    const sistema = $('#int-sistema').val() || 'nenhum'
    const $dupla = $('#int-dupla-comunicacao')
    const permite = sistemaPermiteDuplaComunicacao(sistema)
    const temLicenca = integracaoMeta.confmonit_licenca_ok === true

    if (!permite) {
        $dupla.prop('disabled', true).prop('checked', false)
        return
    }

    $dupla.prop('disabled', !temLicenca)
    if (!temLicenca) {
        $dupla.prop('checked', false)
    }
}

function atualizarPainelSistema() {
    const sistema = $('#int-sistema').val() || 'nenhum'
    $('#int-painel-config').addClass('d-none')
    $('#int-moni-tabs').addClass('d-none')
    $('#btn-int-testar, #btn-int-testar-rodape').addClass('d-none')
    $('#int-confmonit-licenca-aviso').addClass('d-none')
    $('#int-externo-aviso').addClass('d-none')
    $('#int-moni-acoes').addClass('d-none')
    $('#btn-int-salvar-rodape').removeClass('d-none')
    $('#int-acoes-rodape').removeClass('d-none')

    reposicionarCampoNome(sistema)

    if (sistema === 'nenhum') {
        $('#int-ativo').prop('disabled', true).prop('checked', false)
        atualizarDuplaComunicacao()
        $('#btn-int-salvar-rodape').prop('disabled', false)
        return
    }

    $('#int-ativo').prop('disabled', false)
    if (!$('#int-ativo').data('usuarioAlterou')) {
        $('#int-ativo').prop('checked', true)
    }

    atualizarDuplaComunicacao()

    if (sistema === 'confmonit') {
        $('#btn-int-testar-rodape').removeClass('d-none')
        $('#int-acoes-rodape').removeClass('d-none')
        atualizarConfmonitLicenca()
        return
    }

    $('#btn-int-salvar-rodape').prop('disabled', false)

    if (!sistemaPrecisaCamposConfig(sistema)) {
        return
    }

    $('#int-painel-config').removeClass('d-none')

    if (sistema === 'moni') {
        $('#int-moni-tabs').removeClass('d-none')
        selecionarAbaMoni('geral')
        $('#btn-int-testar').removeClass('d-none')
        $('#int-moni-acoes').removeClass('d-none')
        $('#btn-int-salvar-rodape').addClass('d-none')
        $('#int-acoes-rodape').addClass('d-none')
        atualizarEnviarImagemCampos()
    } else if (sistema === 'dguard' || sistema === 'segware') {
        $('#int-externo-aviso').removeClass('d-none')
            .text('Integração ' + sistema + ' em desenvolvimento — os campos abaixo serão usados em breve.')
        $('#int-config-sections .cv-form-tab-panel').addClass('d-none')
        $('#int-config-sections .cv-form-tab-panel[data-tab-panel="geral"]').removeClass('d-none')
        $('#int-config-sections .cv-form-tab-panel[data-tab-panel="comunicacao"]').removeClass('d-none')
    }

    atualizarUrlPreview()
}

function atualizarAuthCampos() {
    const sistema = $('#int-sistema').val()
    if (!sistemaPrecisaCamposConfig(sistema)) return

    const tipo = $('#int-auth-tipo').val() || 'none'
    const $userCol = $('#int-auth-user-col')
    const $passCol = $('#int-auth-pass-col')
    const $passLabel = $('#int-auth-pass-label')

    if (tipo === 'none') {
        $userCol.addClass('d-none')
        $passCol.addClass('d-none')
        return
    }

    if (tipo === 'basic') {
        $userCol.removeClass('d-none')
        $passCol.removeClass('d-none')
        $('#int-auth-user-label').text('Usuário')
        $passLabel.text('Senha')
        return
    }

    $userCol.addClass('d-none')
    $passCol.removeClass('d-none')
    if (tipo === 'bearer') {
        $passLabel.text('Token')
    } else if (tipo === 'apikey') {
        $passLabel.text('ApiKey')
    } else {
        $passLabel.text('Senha')
    }
}

function payloadAuth() {
    const tipo = $('#int-auth-tipo').val() || 'none'
    const auth = { auth_tipo: tipo }
    if (tipo === 'basic') {
        auth.auth_user = $('#int-auth-user').val()
        auth.auth_pass = $('#int-auth-pass').val()
    } else if (tipo === 'bearer' || tipo === 'apikey') {
        auth.auth_token = $('#int-auth-pass').val()
    }
    return auth
}

function payloadIntegracao() {
    const sistema = $('#int-sistema').val() || 'nenhum'
    const payload = {
        sistema: sistema,
        nome: $('#int-nome').val(),
        ativo: sistema !== 'nenhum' && $('#int-ativo').is(':checked'),
        dupla_comunicacao: sistemaPermiteDuplaComunicacao(sistema) && $('#int-dupla-comunicacao').is(':checked')
    }

    if (sistema === 'moni') {
        Object.assign(payload, payloadAuth(), {
            eventos_url: montarEventosUrl(),
            empresa_codigo: $('#int-empresa').val(),
            evento_codigo: $('#int-evento-codigo').val(),
            identificacao_padrao: ($('#int-identificacao').val() || 'E').trim(),
            setor_padrao: ($('#int-setor').val() || '1').trim(),
            particao_padrao: ($('#int-particao').val() || '01').trim(),
            enviar_imagem: $('#int-enviar-imagem').is(':checked'),
            webhook_inbound_ativo: $('#int-webhook-inbound').is(':checked'),
            codigo_evento_armar: ($('#int-codigo-armar').val() || '130').trim(),
            codigo_evento_desarmar: ($('#int-codigo-desarmar').val() || '131').trim()
        })
    } else if (sistema === 'dguard' || sistema === 'segware') {
        Object.assign(payload, payloadAuth(), {
            eventos_url: montarEventosUrl()
        })
    }

    return payload
}

function salvarIntegracao() {
    const sistema = $('#int-sistema').val() || 'nenhum'
    if (sistema === 'confmonit' && integracaoMeta.confmonit_licenca_ok !== true) {
        boxAdvertenciaAuto('Este franqueado não possui licença de terminal ConfMonit.')
        return
    }
    if ($('#int-dupla-comunicacao').is(':checked') && integracaoMeta.confmonit_licenca_ok !== true) {
        boxAdvertenciaAuto('Dupla comunicação requer licença de terminal ConfMonit.')
        return
    }

    const payload = payloadIntegracao()
    $.ajax({
        url: '/api/integracao',
        method: 'PUT',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (r) {
        boxSucessoAuto('Integração salva')
        const parsed = parseRespostaIntegracao(r)
        preencherFormIntegracao(parsed.integracao, parsed.meta)
        atualizarPainelSistema()
        atualizarAuthCampos()
        carregarLogIntegracao(true)
    }).fail(function (xhr) {
        boxErro(mensagemErroAjax(xhr, 'Erro ao salvar integração'))
    })
}

function mensagemErroAjax(xhr, fallback) {
    if (!xhr) return fallback
    let j = xhr.responseJSON
    if (!j && xhr.responseText) {
        try { j = JSON.parse(xhr.responseText) } catch (e) { /* ignore */ }
    }
    if (j && typeof j === 'object') {
        return j.status || j.erro || j.mensagem || fallback
    }
    return fallback
}

function payloadTesteIntegracao() {
    const payload = payloadIntegracao()
    if (($('#int-sistema').val() || '') === 'moni') {
        payload.cliente_teste = ($('#int-cliente-teste').val() || '').trim()
    }
    return payload
}

function testarIntegracao() {
    $.ajax({
        url: '/api/integracao/testar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payloadTesteIntegracao())
    }).done(function (r) {
            if (r.sucesso) {
                let msg = r.mensagem || 'Teste OK'
                if (r.demo_imagem) {
                    msg += ' — foto: ' + r.demo_imagem
                }
                if (r.payload && r.payload.complemento) {
                    msg += ' — complemento: ' + r.payload.complemento
                }
                if (r.id_evento || r.id_processo) {
                    msg += ' — idEvento=' + (r.id_evento || '—') + ' processo=' + (r.id_processo || '—')
                }
                if (r.vis_evento_id) {
                    msg += ' — vis_evento=' + r.vis_evento_id
                }
                if (r.dupla_comunicacao && r.terminal) {
                    msg += ' — terminal: ' + (r.terminal.mensagem || (r.terminal.sucesso ? 'OK' : 'Falha'))
                }
                boxSucessoAuto(msg)
            } else {
                let msg = r.mensagem || 'Falha no teste'
                if (r.payload) {
                    msg += ' — payload: ' + JSON.stringify(r.payload)
                }
                boxAdvertenciaAuto(msg)
            }
            carregarLogIntegracao(true)
        })
        .fail(function () { boxErro('Erro ao testar integração') })
}

function parseFiltroDataLog(val, papel) {
    if (!val || !String(val).trim()) return null
    const bruto = String(val).trim()
    const d = new Date(bruto)
    if (isNaN(d.getTime())) return null
    if (papel === 'ate' && /T00:00(:00)?$/.test(bruto)) {
        d.setHours(23, 59, 59, 999)
    }
    return d.toISOString()
}

function obterFiltrosLogIntegracao() {
    const deRaw = $('#int-log-de').val()
    const ateRaw = $('#int-log-ate').val()
    const de = parseFiltroDataLog(deRaw, 'de')
    const ate = parseFiltroDataLog(ateRaw, 'ate')

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
        cliente_nome: ($('#int-log-cliente').val() || '').trim()
    }
}

function limparFiltrosLogIntegracao() {
    $('#int-log-de').val('')
    $('#int-log-ate').val('')
    $('#int-log-cliente').val('')
}

function setIntLogLoadingMais(loading) {
    $('#int-log-loading').toggleClass('d-none', !loading)
}

function atualizarIntLogFimLista() {
    const temLinhas = $('#int-log-body tr').length > 0 &&
        !$('#int-log-body tr:first td').attr('colspan')
    $('#int-log-fim').toggleClass('d-none', intLogTemMais || !temLinhas)
}

function renderLinhaLogIntegracao(row) {
    const ok = row.sucesso ? '<span class="text-success">OK</span>' : '<span class="text-danger">Falha</span>'
    const cliente = row.cliente_nome || row.id_cliente || '—'
    const evtId = row.vis_evento_id
    const midiaCell = evtId
        ? '<button type="button" class="btn btn-link btn-sm p-0 btn-int-log-midia" data-id="' +
            escapeHtml(String(evtId)) + '" title="Ver imagem e vídeo"><i class="bi bi-eye"></i></button>'
        : '<span class="text-muted">—</span>'
    return '<tr>' +
        '<td>' + escapeHtml(fmtData(row.created_at)) + '</td>' +
        '<td>' + escapeHtml(cliente) + '</td>' +
        '<td>' + escapeHtml(row.camera_nome || '—') + '</td>' +
        '<td>' + escapeHtml(row.sistema || '') + '</td>' +
        '<td>' + escapeHtml(String(row.vis_evento_id || '—')) + '</td>' +
        '<td>' + ok + '</td>' +
        '<td class="small">' + escapeHtml(row.mensagem || '') + '</td>' +
        '<td class="text-center">' + midiaCell + '</td>' +
        '</tr>'
}

function onScrollLogIntegracao() {
    if (intLogCarregando || !intLogTemMais) return
    const el = this
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 48) {
        carregarLogIntegracao(false)
    }
}

function carregarLogIntegracao(reset) {
    if (intLogCarregando) return
    if (!reset && !intLogTemMais) return

    const filtros = obterFiltrosLogIntegracao()
    if (filtros.erro) {
        boxAdvertenciaAuto(filtros.erro)
        return
    }

    if (reset) {
        intLogOffset = 0
        intLogTemMais = true
        $('#int-log-fim').addClass('d-none')
        $('#int-log-body').html('<tr><td colspan="8">Carregando...</td></tr>')
    } else {
        setIntLogLoadingMais(true)
    }

    intLogCarregando = true

    const params = new URLSearchParams({
        limit: String(INT_LOG_PAGE_SIZE),
        offset: String(intLogOffset)
    })
    if (filtros.de) params.set('data_de', filtros.de)
    if (filtros.ate) params.set('data_ate', filtros.ate)
    if (filtros.cliente_nome) params.set('cliente_nome', filtros.cliente_nome)

    $.get('/api/integracao/log?' + params.toString())
        .done(function (r) {
            const rows = r.dados || []
            const $tb = $('#int-log-body')

            if (reset) {
                $tb.empty()
            }

            if (!rows.length) {
                if (reset) {
                    $tb.append('<tr><td colspan="8">Sem registros</td></tr>')
                }
                intLogTemMais = false
                atualizarIntLogFimLista()
                return
            }

            rows.forEach(function (row) {
                $tb.append(renderLinhaLogIntegracao(row))
            })

            intLogOffset += rows.length
            if (typeof r.has_more === 'boolean') {
                intLogTemMais = r.has_more
            } else {
                intLogTemMais = rows.length >= INT_LOG_PAGE_SIZE
            }
            atualizarIntLogFimLista()
        })
        .fail(function (xhr) {
            if (reset) {
                $('#int-log-body').html('<tr><td colspan="8">Erro ao carregar</td></tr>')
            }
            boxErro(mensagemErroAjax(xhr, 'Erro ao carregar histórico'))
        })
        .always(function () {
            intLogCarregando = false
            setIntLogLoadingMais(false)
        })
}

function abrirMidiaEventoIntegracao(eventoId) {
    if (!eventoId) return

    $('#modal-evento-midia-titulo').text('Evento #' + eventoId)
    $('#modal-evento-loading').removeClass('d-none')
    $('#modal-evento-conteudo').addClass('d-none')
    $('#modal-evento-snapshot').attr('src', '')
    $('#modal-evento-clips').empty()
    $('#box-modal-snapshot').addClass('d-none')

    if (modalMidiaIntegracao) modalMidiaIntegracao.show()

    $.get('/api/eventos/' + eventoId + '/clips')
        .fail(function () {
            $('#modal-evento-loading').addClass('d-none')
            $('#modal-evento-conteudo').removeClass('d-none')
            $('#modal-evento-clips').html('<div class="cv-empty py-3 text-danger">Erro ao carregar mídia do evento</div>')
        })
        .done(function (r) {
            const evento = r.evento || {}
            const clips = r.clips || []

            $('#modal-evento-loading').addClass('d-none')
            $('#modal-evento-conteudo').removeClass('d-none')

            if (evento.snapshot_url) {
                const sep = String(evento.snapshot_url).includes('?') ? '&' : '?'
                $('#modal-evento-snapshot').attr('src', evento.snapshot_url + sep + 't=' + Date.now())
                $('#box-modal-snapshot').removeClass('d-none')
            }

            const box = $('#modal-evento-clips').empty()
            if (!clips.length) {
                if (evento.video_url) {
                    box.append(renderClipPlayerIntegracao(1, evento.video_url, null))
                } else {
                    box.append('<div class="cv-empty py-3">Nenhum vídeo capturado ainda</div>')
                }
                return
            }

            clips.forEach(function (clip) {
                box.append(renderClipPlayerIntegracao(clip.seq, clip.video_url, clip.duracao_seg))
            })
        })
}

function renderClipPlayerIntegracao(seq, url, duracaoSeg) {
    if (!url) return ''
    const duracao = duracaoSeg ? ' (' + duracaoSeg + 's)' : ''
    return '<div class="cv-clip-item">' +
        '<p class="cv-clip-label">Clip ' + (seq || 1) + duracao + '</p>' +
        '<video controls playsinline preload="metadata" class="cv-clip-video" src="' + escapeHtml(url) + '"></video>' +
        '</div>'
}

function fmtData(v) {
    if (!v) return '—'
    const d = new Date(v)
    return isNaN(d.getTime()) ? v : d.toLocaleString('pt-BR')
}

function escapeHtml(s) {
    return String(s || '').replace(/[&<>"']/g, function (c) {
        return ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]
    })
}

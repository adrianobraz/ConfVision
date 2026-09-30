$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return

    try {
        confVisionAuthGuard()
        confVisionCarregarCabecalho()
    } catch (e) {
        console.error(e)
    }

    carregarResumo()

    $('#cv-lic-tabs .nav-link').on('click', function () {
        cvLicAtivarTab($(this).data('tab'))
    })

    $('#cv-fatura-filtro').on('change', carregarFaturas)
    $('#cv-fatura-atualizar').on('click', carregarFaturas)

    $('#cv-cap-qty').on('input change', cvCapAtualizarCalculadora)
    $('#cv-cap-contratar-btn').on('click', cvCapContratar)

    $(document).on('click', '.cv-lic-comprar-btn', function () {
        const plano = $(this).data('plano')
        const qty = parseInt($('#cv-qty-' + plano).val(), 10) || 1
        cvLicComprar(plano, qty)
    })

    // Deep-link: /minhas-licencas?tab=comprar|faturas|resumo
    const params = new URLSearchParams(window.location.search || '')
    const tabIni = (params.get('tab') || '').toLowerCase()
    if (tabIni === 'comprar' || tabIni === 'faturas' || tabIni === 'resumo' || tabIni === 'processamento') {
        cvLicAtivarTab(tabIni)
    }
})

var cvCapState = null

function cvLicIdFranqueado() {
    return ConfVisionUrls.idFranqueado()
}

function cvLicErroAjax(xhr) {
    const j = xhr.responseJSON
    if (j) {
        if (j.message) return j.message
        if (j.status) return String(j.status).replace(/^Erro:\s*/i, '')
        if (j.code) return j.code
    }
    const txt = (xhr.responseText || '').trim()
    if (txt) return txt
    return 'tente novamente'
}

function cvLicAtivarTab(tab) {
    $('#cv-lic-tabs .nav-link').removeClass('active')
    $('#cv-lic-tabs .nav-link[data-tab="' + tab + '"]').addClass('active')

    $('#tab-resumo, #tab-comprar, #tab-faturas, #tab-processamento').addClass('d-none')
    $('#tab-' + tab).removeClass('d-none')

    if (tab === 'comprar') carregarPlanos()
    if (tab === 'faturas') carregarFaturas()
    if (tab === 'processamento') carregarCapacidade()
}

function cvLicFormatMoney(v) {
    const n = parseFloat(v)
    if (isNaN(n)) return '—'
    return 'R$ ' + n.toFixed(2).replace('.', ',')
}

function cvLicFormatData(iso) {
    if (!iso) return '—'
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    return d.toLocaleDateString('pt-BR')
}

function cvLicLabelPlano(plano) {
    if (typeof labelPlano === 'function') return labelPlano(plano)
    return plano || '—'
}

function cvLicStatusBadge(status) {
    const map = {
        pendente: 'warn',
        disponivel: 'ok',
        em_uso: 'ok',
        expirada: 'err'
    }
    const cls = map[status] || 'warn'
    return '<span class="cv-lic-badge ' + cls + '">' + (status || '—') + '</span>'
}

function carregarResumo() {
    const id = cvLicIdFranqueado()
    if (!id) {
        $('#cv-lic-resumo').html('<p class="cv-lic-muted">Franqueado não identificado.</p>')
        return
    }

    $.ajax({
        url: '/cvLicencaResumo',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: id })
    }).done(function (r) {
        cvCapState = r.capacidade || null
        cvLicRenderResumo(r)
        cvLicRenderListaLicencas(r.licencas || [])
        cvLicAplicarAlerta(r)
    }).fail(function (xhr) {
        $('#cv-lic-resumo').html('<p class="cv-lic-muted">Erro ao carregar resumo.</p>')
        console.error(xhr.responseText)
    })
}

function cvLicMotivoLabel(motivo, liberado) {
    const m = String(motivo || '').trim()
    const mapa = {
        plano_confvision_bundle_fp: 'Acesso liberado pelo plano FranqueadoPro Pro+ (módulo incluso).',
        plano_confvision_assinatura: 'Assinatura do módulo de câmeras ativa.',
        plano_confvision_ativo: 'Plano do módulo de câmeras ativo.',
        plano_confvision_nenhum: 'Plano do módulo de câmeras ativo.',
        licenca_ativa: 'Acesso liberado por licença de câmera paga.',
        sem_acesso: 'Sem plano ou licença que libere o acesso ao portal.',
        incluso_franqueadopro: 'Módulo incluso no plano FranqueadoPro.'
    }
    if (mapa[m]) return mapa[m]
    if (m.indexOf('plano_confvision_') === 0) {
        return liberado
            ? 'Acesso ao portal liberado.'
            : 'Sem plano ou licença que libere o acesso ao portal.'
    }
    if (!m) return liberado ? 'Acesso ao portal liberado.' : ''
    return m
}

function cvLicAplicarAlerta(r) {
    const el = $('#cv-lic-alerta')
    const cap = r.capacidade || cvCapState
    const semCap = cap && cap.tem_contrato_ativo === false

    if ((!r || r.liberado !== false) && !semCap) {
        el.addClass('d-none').empty()
        return
    }

    let texto = 'Seu acesso depende de licenças pagas ou FranqueadoPro Pro+ ativo.'
    if (semCap) {
        texto = 'Contrate a capacidade de processamento antes de comprar licenças ou cadastrar câmeras.'
        if ((r.faturas_abertas || []).length > 0 || (cap && cap.tem_pendente)) {
            texto += ' Há fatura de capacidade ou licença em aberto — após o pagamento, a liberação é feita pela equipe financeira.'
        }
        el.removeClass('d-none').html(
            '<strong><i class="bi bi-exclamation-triangle-fill"></i> Capacidade obrigatória</strong>' +
            '<p class="mb-0 mt-2">' + texto + '</p>' +
            '<p class="mb-0 mt-2"><button type="button" class="cv-btn-primary btn-sm" onclick="cvLicAtivarTab(\'processamento\')">Contratar processamento</button></p>'
        )
        return
    }

    if ((r.faturas_abertas || []).length > 0) {
        texto += ' Você possui fatura(s) em aberto — após o pagamento, a liberação é feita pela equipe financeira.'
        cvLicAtivarTab('faturas')
    }

    el.removeClass('d-none').html(
        '<strong><i class="bi bi-exclamation-triangle-fill"></i> Acesso limitado</strong>' +
        '<p class="mb-0 mt-2">' + texto + '</p>'
    )
}

function cvLicRenderResumo(r) {
    const resumo = r.resumo || {}
    const liberado = r.liberado
        ? '<span class="cv-lic-badge ok">Liberado</span>'
        : '<span class="cv-lic-badge err">Bloqueado</span>'
    const motivoTxt = cvLicMotivoLabel(r.motivo, !!r.liberado)
    const motivoHtml = motivoTxt
        ? '<p class="cv-lic-muted mb-0">' + motivoTxt + '</p>'
        : ''

    $('#cv-lic-resumo').html(
        '<div class="d-flex flex-wrap gap-2 align-items-center mb-2">' +
        '<h5 class="mb-0 me-auto">Situação do acesso</h5>' + liberado +
        '</div>' +
        motivoHtml +
        cvLicHtmlCapacidadeResumo(r.capacidade) +
        '<div class="cv-lic-resumo-grid">' +
        statBox('Pendentes', resumo.pendente) +
        statBox('Disponíveis', resumo.disponivel) +
        statBox('Em uso', resumo.em_uso) +
        statBox('Expiradas', resumo.expirada) +
        statBox('Total', resumo.total) +
        '</div>'
    )
}

function cvLicHtmlCapacidadeResumo(cap) {
    if (!cap) return ''
    const cfg = cap.config || {}
    const ativo = cap.tem_contrato_ativo
        ? '<span class="cv-lic-badge ok">Ativa</span>'
        : '<span class="cv-lic-badge err">Sem contrato</span>'
    const pend = cap.tem_pendente
        ? ' <span class="cv-lic-badge warn">Pendente pagamento</span>'
        : ''
    return (
        '<div class="cv-lic-cap-resumo-inline mb-3">' +
        '<div class="d-flex flex-wrap gap-2 align-items-center mb-2">' +
        '<h6 class="mb-0 me-auto">Capacidade de processamento</h6>' + ativo + pend +
        '</div>' +
        '<div class="cv-lic-resumo-grid cv-lic-cap-grid">' +
        statBox('Em uso', cap.em_uso) +
        statBox('Contratadas', cap.contratada) +
        statBox('Disponíveis', cap.disponivel) +
        statBox('Mínimo', cap.quantidade_minima_efetiva || cfg.quantidade_minima) +
        '</div>' +
        '</div>'
    )
}

function statBox(label, val) {
    return '<div class="cv-lic-stat"><strong>' + (val || 0) + '</strong><span>' + label + '</span></div>'
}

function cvLicRenderListaLicencas(lista) {
    if (!lista.length) {
        $('#cv-lic-lista').html('<p class="cv-lic-muted mb-0">Nenhuma licença cadastrada.</p>')
        return
    }

    let rows = ''
    lista.slice(0, 50).forEach(function (lic) {
        rows += '<tr>' +
            '<td>#' + lic.id + '</td>' +
            '<td>' + cvLicLabelPlano(lic.plano) + '</td>' +
            '<td>' + cvLicStatusBadge(lic.status) + '</td>' +
            '<td>' + cvLicFormatData(lic.valido_ate) + '</td>' +
            '<td>' + cvLicFormatMoney(lic.valor) + '</td>' +
            '</tr>'
    })

    $('#cv-lic-lista').html(
        '<h6 class="mb-3">Licenças</h6>' +
        '<table class="cv-lic-table"><thead><tr>' +
        '<th>ID</th><th>Plano</th><th>Status</th><th>Válido até</th><th>Valor</th>' +
        '</tr></thead><tbody>' + rows + '</tbody></table>'
    )
}

function cvLicDescricaoPlano(slug) {
    const s = String(slug || '')
    const map = {
        online:
            'Somente a câmera online. Permite visualizar ao vivo e manter a câmera cadastrada, ' +
            'sem detecção analítica e sem captura automática por sensor.',

        sensor_foto:
            'Usa o sensor da central de alarme para acionar a câmera. ' +
            'Quando o alarme dispara no setor vinculado, a câmera registra foto do evento.',

        sensor_foto_video:
            'Usa o sensor da central de alarme para acionar a câmera. ' +
            'Quando o alarme dispara no setor vinculado, a câmera registra foto e vídeo do evento.',

        sensor:
            'Usa o sensor da central de alarme para acionar a câmera. ' +
            'Quando o alarme dispara no setor vinculado, a câmera registra foto e vídeo do evento.',

        analitico_armado_evento:
            'Detecção analítica pela câmera somente com o dispositivo armado. ' +
            'Quando há movimento/pessoa na área desenhada e o alarme está armado, gera evento (sem mídia).',

        analitico_armado_foto:
            'Detecção analítica pela câmera somente com o dispositivo armado. ' +
            'Quando detecta na área e o alarme está armado, gera evento com foto.',

        analitico_armado_foto_video:
            'Detecção analítica pela câmera somente com o dispositivo armado. ' +
            'Quando detecta na área e o alarme está armado, gera evento com foto e vídeo.',

        analitico_armado:
            'Detecção analítica pela câmera somente com o dispositivo armado. ' +
            'Quando detecta na área e o alarme está armado, gera evento com foto e vídeo.',

        analitico_24h_evento:
            'Detecção analítica 24 horas. Em qualquer horário, se houver movimento/pessoa na área desenhada, ' +
            'envia evento (sem mídia), mesmo com o dispositivo desarmado.',

        analitico_24h_foto:
            'Detecção analítica 24 horas. Em qualquer horário, se houver detecção na área, ' +
            'envia evento com foto — independente do arme.',

        analitico_24h_foto_video:
            'Detecção analítica 24 horas. Em qualquer horário, se houver detecção na área, ' +
            'envia evento com foto e vídeo — independente do arme.',

        analitico_24h:
            'Detecção analítica 24 horas. Em qualquer horário, se houver detecção na área, ' +
            'envia evento com foto e vídeo — independente do arme.',

        gravacao_7d:
            'Gravação contínua 24h. Mantém a câmera gravando o tempo todo e guarda os vídeos por 7 dias.',

        gravacao_15d:
            'Gravação contínua 24h. Mantém a câmera gravando o tempo todo e guarda os vídeos por 15 dias.',

        gravacao_30d:
            'Gravação contínua 24h. Mantém a câmera gravando o tempo todo e guarda os vídeos por 30 dias.',

        gravacao_movimento_7d:
            'Gravação por movimento. Só grava quando há qualquer movimento na cena. Retenção de 7 dias.',

        gravacao_movimento_15d:
            'Gravação por movimento. Só grava quando há qualquer movimento na cena. Retenção de 15 dias.',

        gravacao_movimento_30d:
            'Gravação por movimento. Só grava quando há qualquer movimento na cena. Retenção de 30 dias.',

        gravacao_timelapse_7d:
            'Timelapse inteligente. Grava continuamente em ritmo acelerado (timelapse) e, ' +
            'quando há movimento, passa a gravar em ritmo normal. Retenção de 7 dias.',

        gravacao_timelapse_15d:
            'Timelapse inteligente. Grava continuamente em ritmo acelerado (timelapse) e, ' +
            'quando há movimento, passa a gravar em ritmo normal. Retenção de 15 dias.',

        gravacao_timelapse_30d:
            'Timelapse inteligente. Grava continuamente em ritmo acelerado (timelapse) e, ' +
            'quando há movimento, passa a gravar em ritmo normal. Retenção de 30 dias.'
    }
    return map[s] || ''
}

function cvLicGrupoPlano(p) {
    const unidade = String((p && p.unidade) || '').toLowerCase()
    const slug = String((p && (p.plano || p.slug)) || '')
    if (unidade === 'gravacao' || slug.indexOf('gravacao_') === 0) return 'gravacao'
    return 'camera'
}

function cvLicGrupoCameraSub(slug) {
    const s = String(slug || '')
    if (s === 'online') return 'online'
    if (s.indexOf('sensor') === 0) return 'sensor'
    if (s.indexOf('analitico_armado') === 0) return 'analitico_armado'
    if (s.indexOf('analitico_24h') === 0) return 'analitico_24h'
    return 'outros'
}

function cvLicGrupoGravacaoSub(slug) {
    const s = String(slug || '')
    if (s.indexOf('gravacao_movimento') === 0) return 'movimento'
    if (s.indexOf('gravacao_timelapse') === 0) return 'timelapse'
    if (s.indexOf('gravacao_') === 0) return 'continua'
    return 'outros'
}

function cvLicRenderCardPlano(p) {
    const slug = p.plano || p.slug
    const preco = p.valor_com_desconto != null ? p.valor_com_desconto : p.valor
    const desc = cvLicDescricaoPlano(slug)
    return (
        '<div class="cv-lic-plano-card">' +
        '<h6>' + (p.plano_label || cvLicLabelPlano(slug)) + '</h6>' +
        (desc ? '<p class="cv-lic-plano-desc">' + desc + '</p>' : '') +
        '<div class="cv-lic-plano-preco">' + cvLicFormatMoney(preco) + ' <small>/mês</small></div>' +
        '<div class="d-flex gap-2 align-items-center">' +
        '<input type="number" min="1" max="99" value="1" class="form-control form-control-sm cv-lic-qty" id="cv-qty-' + slug + '">' +
        '<button type="button" class="cv-btn-primary btn-sm cv-lic-comprar-btn" data-plano="' + slug + '">Comprar</button>' +
        '</div></div>'
    )
}

function cvLicRenderSecao(titulo, intro, planos, grupoCls) {
    if (!planos.length) return ''
    let cards = ''
    planos.forEach(function (p) { cards += cvLicRenderCardPlano(p) })
    const cls = grupoCls ? (' cv-lic-plano-grupo-' + grupoCls) : ''
    return (
        '<section class="cv-lic-plano-grupo' + cls + '">' +
        '<div class="cv-lic-plano-grupo-head">' +
        '<h6 class="cv-lic-plano-grupo-titulo">' + titulo + '</h6>' +
        (intro ? '<p class="cv-lic-plano-grupo-intro">' + intro + '</p>' : '') +
        '</div>' +
        '<div class="cv-lic-planos-grid">' + cards + '</div>' +
        '</section>'
    )
}

function cvLicRenderCatalogoCard(tipo, titulo, intro, iconClass, borderClass, secoesHtml) {
    if (!secoesHtml || !String(secoesHtml).trim()) {
        return '<p class="cv-lic-muted mb-0">Nenhum plano disponível nesta categoria.</p>'
    }
    return (
        '<div class="cv-lic-catalogo-card-head ' + borderClass + '">' +
        '<div class="cv-lic-catalogo-card-icon"><i class="bi ' + iconClass + '"></i></div>' +
        '<div>' +
        '<h5 class="cv-lic-catalogo-titulo">' + titulo + '</h5>' +
        '<p class="cv-lic-catalogo-intro">' + intro + '</p>' +
        '</div></div>' +
        '<div class="cv-lic-catalogo-card-body">' + secoesHtml + '</div>'
    )
}

function carregarPlanos() {
    const id = cvLicIdFranqueado()
    if (!id) return

    if (cvCapState && cvCapState.tem_contrato_ativo === false) {
        $('#cv-lic-planos-camera').html(
            '<div class="cv-lic-cap-bloqueio">' +
            '<p class="mb-2"><strong>Contrate a capacidade de processamento primeiro.</strong></p>' +
            '<p class="cv-lic-muted mb-3">Licenças de câmera só podem ser compradas após contratar vagas de processamento.</p>' +
            '<button type="button" class="cv-btn-primary btn-sm" onclick="cvLicAtivarTab(\'processamento\')">Ir para Processamento</button>' +
            '</div>'
        )
        $('#cv-lic-planos-gravacao').html('')
        return
    }

    $('#cv-lic-planos-camera').html('<p class="cv-lic-muted mb-0">Carregando planos de câmera…</p>')
    $('#cv-lic-planos-gravacao').html('<p class="cv-lic-muted mb-0">Carregando planos de gravação…</p>')

    $.ajax({
        url: '/cvLicencaPlanos',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: id })
    }).done(function (r) {
        const planos = r.planos || r.dados || []
        if (r.desconto_pro_plus) {
            $('#cv-lic-desconto-badge').removeClass('d-none')
        }

        if (!planos.length) {
            $('#cv-lic-planos-camera').html('<p class="cv-lic-muted mb-0">Nenhum plano disponível.</p>')
            $('#cv-lic-planos-gravacao').html('')
            return
        }

        const cams = planos.filter(function (p) { return cvLicGrupoPlano(p) === 'camera' })
        const gravs = planos.filter(function (p) { return cvLicGrupoPlano(p) === 'gravacao' })

        const camOnline = cams.filter(function (p) { return cvLicGrupoCameraSub(p.plano || p.slug) === 'online' })
        const camSensor = cams.filter(function (p) { return cvLicGrupoCameraSub(p.plano || p.slug) === 'sensor' })
        const camArmado = cams.filter(function (p) { return cvLicGrupoCameraSub(p.plano || p.slug) === 'analitico_armado' })
        const cam24h = cams.filter(function (p) { return cvLicGrupoCameraSub(p.plano || p.slug) === 'analitico_24h' })
        const camOutros = cams.filter(function (p) { return cvLicGrupoCameraSub(p.plano || p.slug) === 'outros' })

        const gContinua = gravs.filter(function (p) { return cvLicGrupoGravacaoSub(p.plano || p.slug) === 'continua' })
        const gMov = gravs.filter(function (p) { return cvLicGrupoGravacaoSub(p.plano || p.slug) === 'movimento' })
        const gTime = gravs.filter(function (p) { return cvLicGrupoGravacaoSub(p.plano || p.slug) === 'timelapse' })
        const gOutros = gravs.filter(function (p) { return cvLicGrupoGravacaoSub(p.plano || p.slug) === 'outros' })

        let htmlCam = ''
        htmlCam += cvLicRenderSecao(
            'Câmera online',
            'Apenas câmera online — visualização e cadastro na plataforma.',
            camOnline,
            'online'
        )
        htmlCam += cvLicRenderSecao(
            'Sensor',
            'Utiliza o sensor da central de alarme para acionar a câmera e registrar foto ou vídeo.',
            camSensor,
            'sensor'
        )
        htmlCam += cvLicRenderSecao(
            'Analítico armado',
            'Detecta pessoa na área desenhada somente quando o dispositivo estiver armado.',
            camArmado,
            'armado'
        )
        htmlCam += cvLicRenderSecao(
            'Analítico 24h',
            'Envia evento quando detecta pessoa na área — em qualquer horário, mesmo desarmado.',
            cam24h,
            '24h'
        )
        htmlCam += cvLicRenderSecao('Outros planos de câmera', '', camOutros, 'outros')

        let htmlGrav = ''
        htmlGrav += cvLicRenderSecao(
            'Gravação contínua',
            'Grava 24 horas por dia. Ideal para histórico completo do período.',
            gContinua,
            'continua'
        )
        htmlGrav += cvLicRenderSecao(
            'Gravação por movimento',
            'Só grava quando há movimento na cena — economiza armazenamento.',
            gMov,
            'movimento'
        )
        htmlGrav += cvLicRenderSecao(
            'Timelapse inteligente',
            'Grava em ritmo acelerado; quando há movimento, passa a gravar em ritmo normal.',
            gTime,
            'timelapse'
        )
        htmlGrav += cvLicRenderSecao('Outros planos de gravação', '', gOutros, 'outros')

        $('#cv-lic-planos-camera').html(
            cvLicRenderCatalogoCard(
                'camera',
                'Licenças de câmera',
                'Definem como a câmera monitora, detecta e gera eventos. Vincule uma licença por câmera.',
                'bi-camera-video',
                'cv-lic-catalogo-head-camera',
                htmlCam
            )
        )

        $('#cv-lic-planos-gravacao').html(
            cvLicRenderCatalogoCard(
                'gravacao',
                'Licenças de gravação',
                'Add-on separado — armazena vídeo na nuvem. Exige licença de câmera ativa e storage Contabo.',
                'bi-collection-play',
                'cv-lic-catalogo-head-gravacao',
                htmlGrav
            )
        )
    }).fail(function () {
        $('#cv-lic-planos-camera').html('<p class="cv-lic-muted mb-0">Erro ao carregar planos.</p>')
        $('#cv-lic-planos-gravacao').html('')
    })
}

function cvLicComprar(plano, quantidade) {
    const id = cvLicIdFranqueado()
    if (!plano || !id) return

    CvMsg.confirmar(
        'Comprar licenças?',
        'Adicionar ' + quantidade + ' licença(s) do plano ' + cvLicLabelPlano(plano) + '.\n\nSerá gerada uma fatura em aberto. As licenças ficam disponíveis após a confirmação do pagamento.'
    ).then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({
            url: '/cvLicencaComprar',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                id_franqueado: id,
                admin_usuario: localStorage.getItem('nomeUsuario') || localStorage.getItem('email') || 'franqueado',
                itens: [{ plano: plano, quantidade: quantidade }]
            })
        }).done(function (res) {
            const fatura = res.fatura || {}
            const ref = fatura.referencia || fatura.id || '—'
            CvMsg.sucesso('Fatura gerada: ' + ref + '.\nAguarde a confirmação do pagamento para usar as licenças.')
            carregarResumo()
            cvLicAtivarTab('faturas')
        }).fail(function (xhr) {
            CvMsg.erro('Erro ao comprar: ' + cvLicErroAjax(xhr))
        })
    })
}

function cvLicResumoItensFatura(itens) {
    if (!itens || !itens.length) return []
    const map = {}
    itens.forEach(function (it) {
        const key = String(it.descricao || 'Licença').trim()
        if (!map[key]) {
            map[key] = {
                descricao: key,
                quantidade: 0,
                valor_unitario: parseFloat(it.valor_unitario) || 0,
                valor_total: 0
            }
        }
        map[key].quantidade += parseInt(it.quantidade, 10) || 1
        const vt = parseFloat(it.valor_total)
        const vu = parseFloat(it.valor_unitario) || 0
        map[key].valor_total += isNaN(vt) ? vu : vt
    })
    return Object.keys(map).map(function (k) { return map[k] })
}

function cvLicHtmlItensFatura(itens) {
    const resumo = cvLicResumoItensFatura(itens)
    if (!resumo.length) return '<span class="cv-lic-muted">—</span>'
    return resumo.map(function (r) {
        return '<div class="cv-lic-fat-item">' +
            '<span class="cv-lic-fat-item-nome">' + r.descricao + '</span>' +
            '<span class="cv-lic-fat-item-qtd">' + r.quantidade + ' un.</span>' +
            '<span class="cv-lic-fat-item-val">' + cvLicFormatMoney(r.valor_unitario) + ' · ' + cvLicFormatMoney(r.valor_total) + '</span>' +
            '</div>'
    }).join('')
}

function carregarFaturas() {
    const id = cvLicIdFranqueado()
    if (!id) return

    const status = $('#cv-fatura-filtro').val()
    $('#cv-lic-faturas').html('<p class="cv-lic-muted">Carregando...</p>')

    $.ajax({
        url: '/cvLicencaFaturas',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: id, status: status })
    }).done(function (r) {
        let lista = r.dados || []
        lista = lista.filter(function (f) {
            const t = (f.tipo || '')
            return t === 'confvision_venda' || t === 'confvision_renovacao'
        })

        if (!lista.length) {
            $('#cv-lic-faturas').html('<p class="cv-lic-muted mb-0">Nenhuma fatura encontrada.</p>')
            return
        }

        let rows = ''
        lista.forEach(function (f) {
            const st = f.status === 'paga' ? 'ok' : (f.status === 'aberta' ? 'warn' : 'warn')
            const resumo = cvLicResumoItensFatura(f.itens || [])
            const licCell = cvLicHtmlItensFatura(f.itens || [])
            const qtdCell = resumo.length
                ? resumo.map(function (r) { return r.quantidade }).join('<br>')
                : '—'
            const unitCell = resumo.length
                ? resumo.map(function (r) { return cvLicFormatMoney(r.valor_unitario) }).join('<br>')
                : '—'
            rows += '<tr>' +
                '<td>' + (f.referencia || f.id) + '</td>' +
                '<td class="cv-lic-fat-lic">' + licCell + '</td>' +
                '<td class="text-end">' + qtdCell + '</td>' +
                '<td class="text-end">' + unitCell + '</td>' +
                '<td class="text-end">' + cvLicFormatMoney(f.valor_total) + '</td>' +
                '<td><span class="cv-lic-badge ' + st + '">' + (f.status || '—') + '</span></td>' +
                '<td>' + cvLicFormatData(f.vencimento_em) + '</td>' +
                '</tr>'
        })

        $('#cv-lic-faturas').html(
            '<table class="cv-lic-table"><thead><tr>' +
            '<th>Referência</th><th>Licença</th><th class="text-end">Qtd</th><th class="text-end">Unit.</th><th class="text-end">Total</th><th>Status</th><th>Vencimento</th>' +
            '</tr></thead><tbody>' + rows + '</tbody></table>' +
            '<p class="cv-lic-muted small mt-2 mb-0">Pagamento confirmado manualmente pela equipe financeira.</p>'
        )
    }).fail(function () {
        $('#cv-lic-faturas').html('<p class="cv-lic-muted">Erro ao carregar faturas.</p>')
    })
}

function carregarCapacidade() {
    const id = cvLicIdFranqueado()
    if (!id) {
        $('#cv-lic-cap-resumo').html('<p class="cv-lic-muted">Franqueado não identificado.</p>')
        return
    }

    $.ajax({
        url: '/cvCapacidadeResumo',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: id })
    }).done(function (cap) {
        cvCapState = cap
        cvCapRenderPainel(cap)
        cvCapAtualizarCalculadora()
    }).fail(function (xhr) {
        $('#cv-lic-cap-resumo').html('<p class="cv-lic-muted">Erro ao carregar capacidade.</p>')
        console.error(xhr.responseText)
    })
}

function cvCapRenderPainel(cap) {
    const cfg = cap.config || {}
    const ativo = cap.tem_contrato_ativo
        ? '<span class="cv-lic-badge ok">Contrato ativo</span>'
        : '<span class="cv-lic-badge err">Sem contrato ativo</span>'
    const pend = cap.tem_pendente
        ? '<span class="cv-lic-badge warn ms-1">Aguardando pagamento</span>'
        : ''

    let detalhe = ''
    if (cap.contrato_ativo) {
        const c = cap.contrato_ativo
        detalhe = '<p class="cv-lic-muted small mb-0">Contrato #' + c.id + ' · ' + c.quantidade_contratada + ' vagas · ' +
            cvLicFormatMoney(c.valor_mensal) + '/mês · válido até ' + cvLicFormatData(c.valido_ate) + '</p>'
    } else if (cap.contrato_pendente) {
        const p = cap.contrato_pendente
        detalhe = '<p class="cv-lic-muted small mb-0">Pendente #' + p.id + ' · ' + p.quantidade_contratada + ' vagas · ' +
            cvLicFormatMoney(p.valor_mensal) + ' — aguardando confirmação do pagamento</p>'
    }

    $('#cv-lic-cap-resumo').html(
        '<div class="d-flex flex-wrap gap-2 align-items-center mb-2">' +
        '<h5 class="mb-0 me-auto">Sua capacidade</h5>' + ativo + pend +
        '</div>' +
        '<div class="cv-lic-resumo-grid cv-lic-cap-grid mb-2">' +
        statBox('Em uso', cap.em_uso) +
        statBox('Contratadas', cap.contratada) +
        statBox('Disponíveis', cap.disponivel) +
        statBox('Mínimo exigido', cap.quantidade_minima_efetiva || cfg.quantidade_minima) +
        '</div>' +
        detalhe
    )

    const min = cap.quantidade_minima_efetiva || cfg.quantidade_minima || 10
    $('#cv-cap-qty').attr('min', min)
    if (parseInt($('#cv-cap-qty').val(), 10) < min) {
        $('#cv-cap-qty').val(min)
    }
    $('#cv-cap-min-hint').text('Mínimo: ' + min + ' (central ou câmeras em uso)')
}

function cvCapAtualizarCalculadora() {
    const id = cvLicIdFranqueado()
    const qty = parseInt($('#cv-cap-qty').val(), 10) || 0
    if (!id || qty <= 0) return

    $.ajax({
        url: '/cvCapacidadeCotacao',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: id, quantidade: qty })
    }).done(function (c) {
        $('#cv-lic-cap-calc').html(
            '<div><span class="cv-lic-muted">Total mensal</span><br><strong class="cv-lic-cap-total">' + cvLicFormatMoney(c.valor_mensal) + '</strong></div>'
        )
    }).fail(function (xhr) {
        $('#cv-lic-cap-calc').html('<span class="text-danger small">' + cvLicErroAjax(xhr) + '</span>')
    })
}

function cvCapContratar() {
    const id = cvLicIdFranqueado()
    const qty = parseInt($('#cv-cap-qty').val(), 10) || 0
    if (!id || qty <= 0) return

    CvMsg.confirmar(
        'Contratar capacidade?',
        'Contratar ' + qty + ' vaga(s) de câmera.\n\nSerá gerada uma fatura em aberto. A capacidade fica ativa após confirmação do pagamento.'
    ).then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({
            url: '/cvCapacidadeContratar',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                id_franqueado: id,
                quantidade: qty,
                admin_usuario: localStorage.getItem('nomeUsuario') || localStorage.getItem('email') || 'franqueado'
            })
        }).done(function (res) {
            const fatura = res.fatura || {}
            const ref = fatura.referencia || fatura.id || '—'
            CvMsg.sucesso('Fatura gerada: ' + ref + '.\nAguarde a confirmação do pagamento para ativar a capacidade.')
            carregarResumo()
            carregarCapacidade()
            cvLicAtivarTab('faturas')
        }).fail(function (xhr) {
            CvMsg.erro('Erro ao contratar: ' + cvLicErroAjax(xhr))
        })
    })
}

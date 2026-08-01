$(document).ready(function () {
    carregarResumo()

    $('#plano-tabs .nav-link').on('click', function () {
        if ($(this).hasClass('disabled') || $(this).attr('aria-disabled') === 'true') return
        var tab = $(this).data('tab')
        fpPlanoAtivarTab(tab)
    })

    $('#fatura-filtro').on('change', carregarFaturas)
    $('#fatura-atualizar').on('click', carregarFaturas)

    var params = new URLSearchParams(window.location.search)
    if (params.get('tab') === 'faturas') {
        fpPlanoAtivarTab('faturas')
    }

    $(document).on('click', '.fp-fatura-pagar-btn', function () {
        var idx = $(this).data('fatura-idx')
        if (idx == null || !window.__fpFaturasLista) return
        fpPlanoAbrirModalPagar(window.__fpFaturasLista[idx])
    })

    $(document).on('click', '.fp-plano-contratar-btn', function () {
        var plano = $(this).data('plano')
        if (!plano) return
        fpPlanoContratar(plano)
    })

    $(document).on('click', '#fp-plano-detalhes-btn', function () {
        fpPlanoAbrirModalModulos()
    })
})

function fpPlanoAtivarTab(tab) {
    $('#plano-tabs .nav-link').removeClass('active')
    $('#plano-tabs .nav-link[data-tab="' + tab + '"]').addClass('active')
    if (tab === 'faturas') {
        $('#tab-plano').addClass('d-none')
        $('#tab-faturas').removeClass('d-none')
        carregarFaturas()
    } else {
        $('#tab-faturas').addClass('d-none')
        $('#tab-plano').removeClass('d-none')
    }
}

function fpPlanoAplicarModoSuspenso(r) {
    if (!r || r.liberado !== false) return
    var prender = r.motivo === 'suspensa' || r.motivo === 'vencida' || r.motivo === 'pendente'
    if (!prender) return

    var titulo = r.motivo === 'suspensa'
        ? 'Sistema suspenso'
        : (r.motivo === 'pendente' ? 'Aguardando pagamento' : 'Assinatura vencida')
    var texto = r.motivo === 'suspensa'
        ? 'Sua assinatura está suspensa. O acesso ao FranqueadoPro ficará bloqueado até a regularização da fatura em aberto.'
        : (r.motivo === 'pendente'
            ? 'Sua assinatura foi gerada e aguarda pagamento. Regularize a fatura em aberto para liberar o acesso ao sistema.'
            : 'Sua assinatura venceu. Regularize a fatura em aberto para restaurar o acesso ao sistema.')

    $('#fp-plano-alerta-suspensao')
        .removeClass('d-none')
        .html(
            '<strong><i class="bi bi-exclamation-triangle-fill"></i> ' + titulo + '</strong>' +
            '<p class="mb-0 mt-2">' + texto + ' Utilize os dados abaixo para pagamento.</p>'
        )

    $('#plano-tabs .nav-link[data-tab="plano"]')
        .addClass('disabled')
        .attr('aria-disabled', 'true')
        .attr('title', 'Disponível após regularizar a fatura')

    fpPlanoAtivarTab('faturas')
    $('#fatura-filtro').val('aberta')

    document.body.classList.add('fp-lic-presa-fatura')
    if (typeof fpLicAplicarModoSuspenso === 'function') fpLicAplicarModoSuspenso()

    if (!window.__fpPlanoPollSuspenso) {
        window.__fpPlanoPollSuspenso = setInterval(function () {
            $.ajax({
                url: '/licencaResumo',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ produto: 'franqueadopro' }),
                cache: false
            }).done(function (resp) {
                if (typeof fpLicProcessarResposta === 'function') {
                    fpLicProcessarResposta(resp)
                } else if (resp.liberado) {
                    clearInterval(window.__fpPlanoPollSuspenso)
                    window.__fpPlanoPollSuspenso = null
                    window.location.href = '/carregar-menu-principal'
                }
            })
        }, 15000)
    }
}

function carregarResumo() {
    $.ajax({
        url: '/licencaResumo',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ produto: 'franqueadopro' })
    }).done(function (r) {
        var ass = r.assinatura || {}
        var liberado = r.liberado
        var motivo = r.motivo || ''
        var planoRaw = ass.plano || ''
        var plano = planoRaw ? planoRaw.toUpperCase().replace('_', '+') : '—'
        var valido = ass.valido_ate ? moment(ass.valido_ate).format('DD/MM/YYYY') : '—'
        var prox = ass.proxima_cobranca_em ? moment(ass.proxima_cobranca_em).format('DD/MM/YYYY') : '—'
        var valor = ass.valor != null ? ('R$ ' + Number(ass.valor).toFixed(2) + '/mês') : '—'
        var badge = liberado ? 'ok' : 'err'
        var statusTxt = liberado ? 'Ativo' : ('Bloqueado: ' + motivo)

        var lim = r.limites_json || ass.limites_json || {}
        var retencao = r.retencao_dias || ass.retencao_dias || '—'
        var limitesHtml =
            '<div class="mt-3"><strong>Limites do seu plano</strong></div>' +
            '<div class="row mt-2 text-muted small">' +
            '<div class="col-6 col-md-3">Clientes: até ' + fmtLimite(lim.clientes_max) + '</div>' +
            '<div class="col-6 col-md-3">Contas: até ' + fmtLimite(lim.contas_max) + '</div>' +
            '<div class="col-6 col-md-3">Usu. alarme: até ' + fmtLimite(lim.usuarios_alarme_max) + '</div>' +
            '<div class="col-6 col-md-3">Setores: até ' + fmtLimite(lim.setores_alarme_max) + '</div>' +
            '</div>' +
            '<div class="text-muted small mt-2">Retenção de histórico após vencimento: <strong>' + retencao + ' dias</strong></div>'

        $('#plano-status').html(
            '<h4>Meu Plano — FranqueadoPro</h4>' +
            '<p><span class="fp-plano-badge ' + badge + '">' + statusTxt + '</span></p>' +
            '<div class="row mt-3">' +
            '<div class="col-md-3"><strong>Plano</strong><br>' + plano + '</div>' +
            '<div class="col-md-3"><strong>Valor contratado</strong><br>' + valor + '</div>' +
            '<div class="col-md-3"><strong>Válido até</strong><br>' + valido + '</div>' +
            '<div class="col-md-3"><strong>Próxima cobrança</strong><br>' + prox + '</div>' +
            '</div>' +
            limitesHtml +
            '<div class="d-flex flex-wrap align-items-center gap-2 mt-3">' +
            '<p class="text-muted mb-0 flex-grow-1">Regularize faturas em aberto para manter o acesso. Dúvidas: central ConfMonit.</p>' +
            '<button type="button" class="btn btn-sm btn-outline-primary flex-shrink-0" id="fp-plano-detalhes-btn">' +
            '<i class="bi bi-list-check"></i> Detalhes</button>' +
            '</div>'
        )

        window.__fpPlanoModulos = r.modulos_json || ass.modulos_json || {}
        window.__fpPlanoRaw = planoRaw

        if (typeof fpRenderComparacaoPlanos === 'function') {
            var faturasAbertas = r.faturas_abertas || []
            fpRenderComparacaoPlanos('#plano-comparacao', planoRaw, {
                podeContratar: !liberado && faturasAbertas.length === 0
            })
        }

        if (typeof fpLicProcessarResposta === 'function') {
            if (fpLicProcessarResposta(r)) return
        } else if (typeof fpLicSalvarEstado === 'function') {
            fpLicSalvarEstado(r)
            if (typeof fpAplicarLicencaUI === 'function') fpAplicarLicencaUI()
        }

        fpPlanoAplicarModoSuspenso(r)
    }).fail(function () {
        $('#plano-status').html('<p class="text-danger">Não foi possível carregar o plano.</p>')
        if (typeof fpRenderComparacaoPlanos === 'function') {
            fpRenderComparacaoPlanos('#plano-comparacao', '', { podeContratar: true })
        }
    })
}

function fpPlanoGetModalModulos() {
    var modalEl = document.getElementById('fp-plano-modulos-modal')
    if (!modalEl || typeof bootstrap === 'undefined' || !bootstrap.Modal) return null
    if (modalEl.parentElement !== document.body) {
        document.body.appendChild(modalEl)
    }
    return bootstrap.Modal.getOrCreateInstance(modalEl, {
        backdrop: true,
        keyboard: true,
        focus: true
    })
}

function fpPlanoAbrirModalModulos() {
    if (typeof fpRenderModulosPlanoAtual !== 'function') return
    fpRenderModulosPlanoAtual('#fp-plano-modulos-corpo', window.__fpPlanoModulos || {}, window.__fpPlanoRaw || '', { noTitulo: true })
    var modal = fpPlanoGetModalModulos()
    if (modal) modal.show()
}

function fpPlanoContratar(plano) {
    var nome = typeof fpPlanoNomeExibicao === 'function' ? fpPlanoNomeExibicao(plano) : plano
    if (!confirm('Contratar o plano ' + nome + '? Será gerada uma fatura em aberto. O acesso será liberado após confirmação do pagamento pela central ConfMonit.')) {
        return
    }
    $.ajax({
        url: '/licencaContratar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ produto: 'franqueadopro', plano: plano })
    }).done(function () {
        alert('Assinatura e fatura geradas com sucesso. Acesse a aba Faturas e Cobrança para ver as instruções de pagamento.')
        fpPlanoAtivarTab('faturas')
        carregarResumo()
        carregarFaturas()
    }).fail(function (xhr) {
        var msg = 'Não foi possível contratar o plano.'
        try {
            var raw = xhr.responseText || ''
            var j = JSON.parse(raw)
            if (j.message) msg = j.message
            else if (j.error) msg = j.error
            else if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
            else if (j.payload && j.payload.message) msg = j.payload.message
        } catch (e) {
            if (xhr.responseText) msg = xhr.responseText.slice(0, 300)
        }
        alert(msg)
    })
}

function fmtLimite(v) {
    if (v == null || v === '') return '—'
    return Number(v) >= 9999 ? 'Ilimitado*' : String(v)
}

function carregarFaturas() {
    var status = $('#fatura-filtro').val()
    var payload = {}
    if (status) payload.status = status

    $('#plano-faturas').html('<p class="text-muted">Carregando...</p>')
    $.ajax({
        url: '/licencaMinhasFaturas',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (r) {
        var lista = (r.dados || [])
        window.__fpFaturasLista = lista
        if (!lista.length) {
            $('#plano-faturas').html(
                '<p class="text-muted mb-0">Nenhuma fatura encontrada.</p>' +
                '<p class="text-muted small mt-2 mb-0">Se você já efetuou o pagamento, aguarde a confirmação pela central ConfMonit.</p>'
            )
            return
        }
        var html = '<table class="table table-sm table-hover"><thead><tr>' +
            '<th>Referência</th><th>Tipo</th><th>Vencimento</th><th>Valor</th><th>Status</th><th></th>' +
            '</tr></thead><tbody>'
        lista.forEach(function (f, idx) {
            var st = (f.status || '').toLowerCase()
            var cls = st === 'aberta' ? 'text-warning' : (st === 'paga' ? 'text-success' : '')
            var acao = ''
            if (st === 'aberta') {
                acao = '<button type="button" class="btn btn-sm btn-primary fp-fatura-pagar-btn" data-fatura-idx="' + idx + '">' +
                    '<i class="bi bi-credit-card"></i> Pagar</button>'
            }
            html += '<tr><td>' + (f.referencia || f.id) + '</td>' +
                '<td>' + (f.tipo || 'assinatura') + '</td>' +
                '<td>' + (f.vencimento_em ? moment(f.vencimento_em).format('DD/MM/YYYY') : '—') + '</td>' +
                '<td>R$ ' + Number(f.valor_total || 0).toFixed(2) + '</td>' +
                '<td class="' + cls + '">' + (f.status || '') + '</td>' +
                '<td class="text-end">' + acao + '</td></tr>'
        })
        html += '</tbody></table>'
        $('#plano-faturas').html(html)
    }).fail(function (xhr) {
        var msg = 'Erro ao carregar faturas.'
        try {
            var j = JSON.parse(xhr.responseText || '')
            if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
            else if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#plano-faturas').html('<p class="text-danger">' + msg + '</p>')
    })
}

function fpPlanoGetModalPagar() {
    var modalEl = document.getElementById('fp-fatura-pagar-modal')
    if (!modalEl || typeof bootstrap === 'undefined' || !bootstrap.Modal) return null
    if (modalEl.parentElement !== document.body) {
        document.body.appendChild(modalEl)
    }
    return bootstrap.Modal.getOrCreateInstance(modalEl, {
        backdrop: true,
        keyboard: true,
        focus: true
    })
}

function fpPlanoAbrirModalPagar(f) {
    if (!f) return
    var ref = f.referencia || f.id || '—'
    var venc = f.vencimento_em ? moment(f.vencimento_em).format('DD/MM/YYYY') : '—'
    var valor = 'R$ ' + Number(f.valor_total || 0).toFixed(2)
    var obs = (f.observacao || '').trim()
    var corpo =
        '<p class="mb-2"><strong>Referência:</strong> ' + ref + '</p>' +
        '<p class="mb-2"><strong>Vencimento:</strong> ' + venc + '</p>' +
        '<p class="mb-3"><strong>Valor:</strong> <span class="fs-5">' + valor + '</span></p>'
    if (obs) {
        corpo += '<div class="alert alert-info mb-3"><strong>Instruções de pagamento</strong><br>' +
            obs.replace(/\n/g, '<br>') + '</div>'
    } else {
        corpo += '<div class="alert alert-warning mb-3">' +
            'Entre em contato com a <strong>central ConfMonit</strong> para obter PIX, boleto ou demais formas de pagamento desta fatura.' +
            '</div>'
    }
    corpo += '<p class="text-muted small mb-0">Após a confirmação do pagamento pela central, o sistema será liberado automaticamente.</p>'
    $('#fp-fatura-pagar-corpo').html(corpo)
    var modal = fpPlanoGetModalPagar()
    if (modal) modal.show()
}

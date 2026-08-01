/* Comparativo detalhado — usa fp-planos-dados.js */

function fpRenderComparacaoPlanos(containerSelector, planoAtual, opcoes) {
    var $el = $(containerSelector)
    if (!$el.length) return

    opcoes = opcoes || {}
    var planosLista = (opcoes.planosCatalogo && opcoes.planosCatalogo.length)
        ? opcoes.planosCatalogo
        : (typeof FP_PLANOS_PADRAO !== 'undefined' ? FP_PLANOS_PADRAO : [])
    if (!planosLista.length) return

    var podeContratar = !!opcoes.podeContratar

    var cards = ''
    planosLista.forEach(function (p) {
        var ativo = planoAtual === p.plano ? ' fp-plano-card-ativo' : ''
        var badge = planoAtual === p.plano ? '<span class="fp-plano-badge-atual">Seu plano</span>' : ''
        var btnContratar = ''
        if (podeContratar && (!planoAtual || planoAtual !== p.plano)) {
            btnContratar = '<button type="button" class="btn btn-sm btn-primary mt-2 fp-plano-contratar-btn" data-plano="' + p.plano + '">' +
                'Contratar ' + p.nome_exibicao + '</button>'
        }
        var limites = ''
        FP_LIMITES_LINHAS.forEach(function (l) {
            limites += '<li><span>' + l.label + '</span><strong>' + fpPlanoLimiteValor(p, l.chave) + '</strong></li>'
        })
        cards +=
            '<div class="fp-plano-compare-card' + ativo + '">' +
            '<div class="fp-plano-compare-head"><strong>' + p.nome_exibicao + '</strong>' + badge + '</div>' +
            '<div class="fp-plano-compare-preco">R$ ' + Number(p.valor_mensal).toFixed(2) + '<small>/mês</small></div>' +
            (p.pacote_cota_nome
                ? '<div class="fp-plano-compare-ret text-muted small">Inclui ' + p.pacote_cota_nome +
                  (p.pacote_cota_quantidade ? ' (' + p.pacote_cota_quantidade + ')' : '') + '</div>'
                : '') +
            '<div class="fp-plano-compare-ret">Retenção: <strong>' + p.retencao_dias + ' dias</strong></div>' +
            '<ul class="fp-plano-compare-limites">' + limites + '</ul>' + btnContratar + '</div>'
    })

    var grupos = {}
    FP_MODULOS_CATALOGO.forEach(function (m) {
        if (!grupos[m.grupo]) grupos[m.grupo] = []
        grupos[m.grupo].push(m)
    })

    var modBody = ''
    for (var gk in grupos) {
        if (!grupos.hasOwnProperty(gk)) continue
        modBody += '<tr class="fp-plano-grupo-row"><td colspan="' + (planosLista.length + 1) + '"><strong>' + gk + '</strong></td></tr>'
        grupos[gk].forEach(function (m) {
            var minLabel = m.minimo === 'pro_plus' ? 'Pro+' : (m.minimo === 'pro' ? 'Pro' : 'Lite')
            modBody += '<tr><td><span class="fp-plano-mod-label">' + m.label + '</span>' +
                '<small class="fp-plano-mod-desc">' + m.desc + '</small>' +
                '<small class="fp-plano-mod-min">Plano mínimo: ' + minLabel + '</small></td>'
            planosLista.forEach(function (p) {
                var ok = fpPlanoModuloLiberadoObj(p, m.chave)
                var cls = planoAtual === p.plano ? ' fp-plano-col-ativo' : ''
                modBody += '<td class="text-center' + cls + '"><span class="' + (ok ? 'fp-plano-sim' : 'fp-plano-nao') + '">' + (ok ? '✓' : '✗') + '</span></td>'
            })
            modBody += '</tr>'
        })
    }

    var modHead = '<th>Funcionalidade</th>'
    planosLista.forEach(function (p) {
        modHead += '<th class="text-center' + (planoAtual === p.plano ? ' fp-plano-col-ativo' : '') + '">' + fpPlanoNomeExibicao(p.plano) + '</th>'
    })

    var prodHead = '<th>Produto</th>'
    planosLista.forEach(function (p) {
        prodHead += '<th class="text-center' + (planoAtual === p.plano ? ' fp-plano-col-ativo' : '') + '">' + fpPlanoNomeExibicao(p.plano) + '</th>'
    })

    var prodBody = ''
    if (typeof FP_PRODUTOS_ECOSISTEMA !== 'undefined') {
        FP_PRODUTOS_ECOSISTEMA.forEach(function (pr) {
            var minLabel = pr.minimo === 'pro_plus' ? 'Pro+' : (pr.minimo === 'pro' ? 'Pro' : 'Lite')
            prodBody += '<tr><td><span class="fp-plano-mod-label">' + pr.nome + '</span>' +
                '<small class="fp-plano-mod-desc">' + pr.desc + '</small>' +
                (pr.ativacao ? '<small class="fp-plano-mod-desc">' + pr.ativacao + '</small>' : '') +
                '<small class="fp-plano-mod-min">Plano mínimo FranqueadoPro: ' + minLabel + '</small></td>'
            planosLista.forEach(function (p) {
                var ok = typeof fpPlanoProdutoLiberado === 'function'
                    ? fpPlanoProdutoLiberado(p, pr.id)
                    : false
                var cls = planoAtual === p.plano ? ' fp-plano-col-ativo' : ''
                prodBody += '<td class="text-center' + cls + '"><span class="' + (ok ? 'fp-plano-sim' : 'fp-plano-nao') + '">' + (ok ? '✓' : '✗') + '</span></td>'
            })
            prodBody += '</tr>'
        })
    }

    var produtos = ''
    if (prodBody) {
        produtos = '<div class="fp-plano-produtos mt-3"><strong>Produtos complementares</strong>' +
            '<div class="table-responsive"><table class="table table-sm fp-plano-compare-table mt-2">' +
            '<thead><tr>' + prodHead + '</tr></thead><tbody>' + prodBody + '</tbody></table></div></div>'
    }

    $el.html(
        '<h5 class="fp-plano-compare-title">Compare os planos FranqueadoPro</h5>' +
        '<p class="text-muted small">Preços e limites conforme catálogo do plano. Para montar um plano sob medida, use <a href="/carregar-montar-plano">Montar Meu Plano</a>.</p>' +
        '<div class="fp-plano-compare-cards">' + cards + '</div>' +
        '<div class="table-responsive"><table class="table table-sm fp-plano-compare-table">' +
        '<thead><tr>' + modHead + '</tr></thead><tbody>' + modBody + '</tbody></table></div>' +
        produtos +
        '<p class="text-muted small mt-2 mb-0">* Ilimitado = teto prático alto (9.999). ✓ nos produtos = incluso no plano FP (sem compra avulsa). Cada produto também pode ser contratado à parte em Meu Plano.</p>'
    )
}

function fpRenderModulosPlanoAtual(containerSelector, modulos, plano, opcoes) {
    var $el = $(containerSelector)
    if (!$el.length || typeof FP_MODULOS_CATALOGO === 'undefined') return

    opcoes = opcoes || {}

    if (!plano) {
        $el.html('<p class="text-muted small mb-0">Contrate um plano para ver os módulos e produtos incluídos.</p>')
        return
    }

    var planoNome = typeof fpPlanoNomeExibicao === 'function' ? fpPlanoNomeExibicao(plano) : plano
    var liberados = 0
    var bloqueados = 0

    var grupos = {}
    FP_MODULOS_CATALOGO.forEach(function (m) {
        if (!grupos[m.grupo]) grupos[m.grupo] = []
        grupos[m.grupo].push(m)
    })

    var html = ''
    if (!opcoes.noTitulo) {
        html +=
            '<div class="fp-plano-atual-head d-flex flex-wrap align-items-center gap-2 mb-1">' +
            '<strong>Seu plano — módulos e acesso</strong>' +
            '<span class="fp-plano-badge-atual">' + planoNome + '</span></div>'
        html += '<p class="text-muted small mb-3">Incluídos no seu plano atual. Itens com ✗ exigem upgrade (veja o plano mínimo ao lado).</p>'
    } else {
        html += '<p class="text-muted small mb-3">Plano <strong>' + planoNome + '</strong> — incluídos no contrato atual. Itens com ✗ exigem upgrade.</p>'
    }
    html += '<div class="row">'

    for (var gk in grupos) {
        if (!grupos.hasOwnProperty(gk)) continue
        html += '<div class="col-md-6 col-lg-4"><div class="fp-plano-mod-grupo"><strong>' + gk + '</strong><ul class="fp-plano-mod-list small">'
        grupos[gk].forEach(function (m) {
            var ok = fpModuloLiberadoSnapshot(modulos, plano, m.chave)
            if (ok) liberados++
            else bloqueados++
            var minLabel = typeof fpPlanoLabelMinimo === 'function' ? fpPlanoLabelMinimo(m.minimo) : m.minimo
            var extra = ok ? '' : ' <small class="fp-plano-mod-min">(plano ' + minLabel + ')</small>'
            html += '<li class="' + (ok ? 'fp-plano-sim' : 'fp-plano-nao') + '" title="' + m.desc.replace(/"/g, '') + '">' +
                (ok ? '✓' : '✗') + ' ' + m.label + extra + '</li>'
        })
        html += '</ul></div></div>'
    }
    html += '</div>'

    if (typeof FP_PRODUTOS_ECOSISTEMA !== 'undefined' && FP_PRODUTOS_ECOSISTEMA.length) {
        html += '<div class="fp-plano-inclusos"><strong>Produtos complementares</strong>' +
            '<ul class="fp-plano-mod-list small">'
        FP_PRODUTOS_ECOSISTEMA.forEach(function (pr) {
            var ok = typeof fpPlanoProdutoLiberado === 'function'
                ? fpPlanoProdutoLiberado(plano, pr.id)
                : false
            if (ok) liberados++
            else bloqueados++
            var minLabel = typeof fpPlanoLabelMinimo === 'function' ? fpPlanoLabelMinimo(pr.minimo) : pr.minimo
            var extra = ok ? '' : ' <small class="fp-plano-mod-min">(plano ' + minLabel + ')</small>'
            if (ok && pr.infoOnly && pr.ativacao) {
                extra = ' <small class="fp-plano-mod-min">' + pr.ativacao + '</small>'
            }
            html += '<li class="' + (ok ? 'fp-plano-sim' : 'fp-plano-nao') + '" title="' + pr.desc.replace(/"/g, '') + '">' +
                (ok ? '✓' : '✗') + ' ' + pr.nome + extra + '</li>'
        })
        html += '</ul></div>'
    }

    html += '<p class="text-muted small mt-2 mb-0">' + liberados + ' itens liberados · ' + bloqueados + ' exigem upgrade de plano.</p>'
    $el.html(html)
}

function fpModuloLiberadoSnapshot(modulos, plano, chave) {
    if (modulos && typeof modulos === 'object' && modulos[chave] !== undefined && modulos[chave] !== null) {
        return modulos[chave] === true
    }
    if (typeof fpModuloLiberadoEfetivo === 'function') {
        return fpModuloLiberadoEfetivo(plano, chave)
    }
    return false
}

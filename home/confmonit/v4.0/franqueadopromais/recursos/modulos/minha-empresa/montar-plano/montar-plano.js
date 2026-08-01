/* Montar plano a la carte */

var fpAlacarteState = {
    planos: [],
    modulos: [],
    planoBase: '',
    planoTravado: false,
    selecionados: {},
    addonsIniciais: [],
    addonsAtivos: [],
    addonsPendentes: [],
    simulacao: null,
    liberado: false,
    assinatura: null,
    faturasAbertas: [],
    podeContratar: true,
    mensagemStatus: '',
    cupomCodigo: '',
    cupomPreview: null,
    ecossistema: [],
    ecoSelecionados: {}
}

var FP_ECO_NOMES = {
    webterminal: 'WebTerminal',
    terminalmovel: 'Terminal Móvel',
    confvision: 'Vision',
    webambiente: 'webAmbiente',
    dialyze: 'Dialyze'
}

var FP_ECO_RANK = { lite: 1, padrao: 1, pro: 2, pro_plus: 3 }

$(document).ready(function () {
    carregarDados()
    fpAlacarteInitResumoSticky()

    $(document).on('click', '.fp-alacarte-base-card:not(.travado)', function () {
        var plano = $(this).data('plano')
        if (!plano || fpAlacarteState.planoTravado) return
        fpAlacarteState.planoBase = plano
        $('.fp-alacarte-base-card').removeClass('ativo')
        $(this).addClass('ativo')
        fpAlacarteLimparEcoInclusos()
        renderModulos()
        renderEcossistema()
        simular()
    })

    $(document).on('change', '.fp-alacarte-mod-check', function () {
        var chave = $(this).data('chave')
        if (!chave) return
        if ($(this).is(':checked')) {
            fpAlacarteState.selecionados[chave] = true
        } else {
            delete fpAlacarteState.selecionados[chave]
        }
        simular()
    })

    $(document).on('click', '.fp-alacarte-mod-remover', function (e) {
        e.preventDefault()
        e.stopPropagation()
        var chave = $(this).data('chave')
        if (!chave) return
        delete fpAlacarteState.selecionados[chave]
        renderModulos()
        simular()
    })

    $(document).on('click', '.fp-alacarte-mod-desfazer', function (e) {
        e.preventDefault()
        e.stopPropagation()
        var chave = $(this).data('chave')
        if (!chave) return
        fpAlacarteState.selecionados[chave] = true
        renderModulos()
        simular()
    })

    $(document).on('change', '.fp-alacarte-eco-check', function () {
        var produto = $(this).data('produto')
        if (!produto) return
        if ($(this).is(':checked')) {
            var plano = $('.fp-alacarte-eco-plano[data-produto="' + produto + '"]').val()
            if (!plano) return
            fpAlacarteState.ecoSelecionados[produto] = plano
        } else {
            delete fpAlacarteState.ecoSelecionados[produto]
        }
        renderEcossistema()
        simular()
    })

    $(document).on('change', '.fp-alacarte-eco-plano', function () {
        var produto = $(this).data('produto')
        if (!produto) return
        if (!fpAlacarteState.ecoSelecionados[produto]) return
        fpAlacarteState.ecoSelecionados[produto] = $(this).val()
        simular()
    })

    $('#fp-alacarte-contratar').on('click', contratarAlacarte)
    $('#fp-alacarte-cupom-validar').on('click', fpAlacarteValidarCupom)
    $('#fp-alacarte-cupom-codigo').on('change', function () {
        fpAlacarteState.cupomCodigo = String($(this).val() || '').trim().toUpperCase()
        if (!fpAlacarteState.cupomCodigo) {
            fpAlacarteState.cupomPreview = null
            $('#fp-alacarte-cupom-msg').removeClass('text-success text-danger').addClass('text-muted').text('')
            simular()
        }
    })
})

function fpAlacarteCupomCodigo() {
    return String($('#fp-alacarte-cupom-codigo').val() || fpAlacarteState.cupomCodigo || '').trim().toUpperCase()
}

function fpAlacarteValidarCupom() {
    var codigo = fpAlacarteCupomCodigo()
    var $msg = $('#fp-alacarte-cupom-msg')
    if (!codigo) {
        fpAlacarteState.cupomCodigo = ''
        fpAlacarteState.cupomPreview = null
        $msg.removeClass('text-success text-danger').addClass('text-muted').text('Informe um código.')
        simular()
        return
    }
    var valorBase = 0
    if (fpAlacarteState.simulacao) {
        valorBase = Math.max(0, calcularValorFaturaPagar(fpAlacarteState.simulacao) - valorEcoMensal())
        if (!valorBase) valorBase = Number(fpAlacarteState.simulacao.valor_total || 0)
    }
    $msg.removeClass('text-success text-danger').addClass('text-muted').text('Validando…')
    $.ajax({
        url: '/licencaCupomValidar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            codigo: codigo,
            produto: 'franqueadopro',
            valor_base: valorBase || undefined
        })
    }).done(function (r) {
        fpAlacarteState.cupomCodigo = codigo
        fpAlacarteState.cupomPreview = r
        var cupom = r.cupom || {}
        var desc = cupom.tipo === 'percentual'
            ? (Number(cupom.valor || 0) + '%')
            : ('R$ ' + Number(cupom.valor || 0).toFixed(2))
        var txt = 'Cupom válido: ' + desc
        if (r.valor_desconto != null) {
            txt += ' (− R$ ' + Number(r.valor_desconto).toFixed(2) + ')'
        }
        $msg.removeClass('text-muted text-danger').addClass('text-success').text(txt)
        simular()
    }).fail(function (xhr) {
        fpAlacarteState.cupomCodigo = ''
        fpAlacarteState.cupomPreview = null
        var msg = 'Cupom inválido.'
        try {
            var j = JSON.parse(xhr.responseText || '')
            if (j.message) msg = j.message
            else if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
        } catch (e) { /* ignore */ }
        $msg.removeClass('text-muted text-success').addClass('text-danger').text(msg)
        simular()
    })
}

function carregarDados() {
    $.when(
        $.ajax({
            url: '/licencaCatalogoPlanos',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: 'franqueadopro' })
        }),
        $.ajax({
            url: '/licencaCatalogoModulos',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: 'franqueadopro' })
        }),
        $.ajax({
            url: '/licencaResumo',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: 'franqueadopro' })
        }),
        $.ajax({
            url: '/licencaEcossistema',
            method: 'POST',
            contentType: 'application/json',
            data: '{}'
        })
    ).done(function (planosResp, modulosResp, resumoResp, ecoResp) {
        var planos = (planosResp[0].dados || []).filter(function (p) {
            return p.plano === 'lite' || p.plano === 'pro' || p.plano === 'pro_plus'
        })
        planos.sort(function (a, b) {
            var ordem = { lite: 1, pro: 2, pro_plus: 3 }
            return (ordem[a.plano] || 9) - (ordem[b.plano] || 9)
        })

        fpAlacarteState.planos = planos
        fpAlacarteState.modulos = modulosResp[0].dados || []
        fpAlacarteState.ecossistema = (ecoResp[0] && ecoResp[0].dados) || []
        fpAlacarteState.ecoSelecionados = {}

        var resumo = resumoResp[0] || {}
        var assinatura = resumo.assinatura || null

        fpAlacarteState.liberado = !!resumo.liberado
        fpAlacarteState.assinatura = assinatura
        fpAlacarteState.faturasAbertas = resumo.faturas_abertas || []
        fpAlacarteState.addonsPendentes = resumo.addons_pendentes || []

        if (typeof fpLicSalvarEstado === 'function') {
            fpLicSalvarEstado(resumo)
        }

        var contratados = resumo.addons_contratados || []
        var efetivos = resumo.addons_efetivos || []

        if (assinatura && assinatura.plano) {
            fpAlacarteState.planoBase = assinatura.plano
            fpAlacarteState.planoTravado = true
            fpAlacarteState.addonsIniciais = normalizarAddonsLista(contratados)
            fpAlacarteState.addonsAtivos = normalizarAddonsLista(efetivos)
        } else {
            fpAlacarteState.planoBase = planos[0] ? planos[0].plano : ''
            fpAlacarteState.planoTravado = false
            fpAlacarteState.addonsIniciais = []
            fpAlacarteState.addonsAtivos = []
        }

        fpAlacarteState.selecionados = {}
        fpAlacarteState.addonsIniciais.forEach(function (ch) {
            fpAlacarteState.selecionados[ch] = true
        })
        fpAlacarteState.addonsPendentes.forEach(function (ch) {
            if (addonChaveValida(ch)) fpAlacarteState.selecionados[ch] = true
        })

        fpAlacarteState.mensagemStatus = montarMensagemStatus(resumo)
        atualizarMensagemStatus()

        renderBases()
        renderModulos()
        renderEcossistema()
        simular()
        fpAlacarteRefreshResumoSticky()
    }).fail(function () {
        mostrarErro('Não foi possível carregar o catálogo. Tente novamente.')
    })
}

function fpEcoNomeProduto(pid) {
    return FP_ECO_NOMES[pid] || pid
}

function fpAlacarteEcoPlanosCompraveis(item) {
    var planos = (item && item.planos) || []
    if (!planos.length) return []
    var tierIncluso = fpAlacarteEcoTierInclusoPreview(item.produto)
    if (!tierIncluso && item.status === 'incluso' && item.plano_incluso_fp) {
        tierIncluso = item.plano_incluso_fp
    }
    if (!tierIncluso) return planos.slice()
    var minRank = FP_ECO_RANK[tierIncluso] || 0
    return planos.filter(function (x) {
        return (FP_ECO_RANK[x.plano] || 0) > minRank
    })
}

function fpAlacarteEcoTierInclusoPreview(produto) {
    if (typeof fpPlanoProdutoInclusoTier !== 'function') return ''
    return fpPlanoProdutoInclusoTier(fpAlacarteState.planoBase, produto) || ''
}

function fpAlacarteEcoInclusoNoBase(produto) {
    if (typeof fpPlanoProdutoLiberado === 'function') {
        return !!fpPlanoProdutoLiberado(fpAlacarteState.planoBase, produto)
    }
    return false
}

function fpAlacarteLimparEcoInclusos() {
    var sel = fpAlacarteState.ecoSelecionados || {}
    Object.keys(sel).forEach(function (pid) {
        if (fpAlacarteEcoInclusoNoBase(pid) && !fpAlacarteEcoPlanosCompraveis({
            produto: pid,
            planos: fpAlacarteEcoItem(pid) ? fpAlacarteEcoItem(pid).planos : [],
            status: 'incluso',
            plano_incluso_fp: fpAlacarteEcoTierInclusoPreview(pid)
        }).length) {
            delete fpAlacarteState.ecoSelecionados[pid]
        }
    })
}

function fpAlacarteEcoItem(produto) {
    var lista = fpAlacarteState.ecossistema || []
    for (var i = 0; i < lista.length; i++) {
        if (lista[i].produto === produto) return lista[i]
    }
    return null
}

function fpAlacarteEcoValorPlano(produto, plano) {
    var item = fpAlacarteEcoItem(produto)
    if (!item) return 0
    var planos = item.planos || []
    for (var i = 0; i < planos.length; i++) {
        if (planos[i].plano === plano) return Number(planos[i].valor_mensal || 0)
    }
    return 0
}

function ecoProdutosSelecionados() {
    var out = []
    var sel = fpAlacarteState.ecoSelecionados || {}
    Object.keys(sel).forEach(function (pid) {
        var plano = sel[pid]
        if (!plano) return
        if (fpAlacarteEcoInclusoNoBase(pid)) {
            var compraveis = fpAlacarteEcoPlanosCompraveis(fpAlacarteEcoItem(pid) || { produto: pid, planos: [] })
            var okUpgrade = compraveis.some(function (p) { return p.plano === plano })
            if (!okUpgrade) return
        }
        out.push({
            produto: pid,
            plano: plano,
            label: fpEcoNomeProduto(pid),
            valor_mensal: fpAlacarteEcoValorPlano(pid, plano)
        })
    })
    out.sort(function (a, b) { return a.produto < b.produto ? -1 : 1 })
    return out
}

function temEcoSelecionado() {
    return ecoProdutosSelecionados().length > 0
}

function valorEcoMensal() {
    return ecoProdutosSelecionados().reduce(function (acc, i) {
        return acc + Number(i.valor_mensal || 0)
    }, 0)
}

function renderEcossistema() {
    var $box = $('#fp-alacarte-ecossistema')
    if (!$box.length) return
    var lista = fpAlacarteState.ecossistema || []
    if (!lista.length) {
        $box.html('<p class="text-muted">Nenhum produto do ecossistema no catálogo.</p>')
        return
    }

    var html = ''
    lista.forEach(function (item) {
        var pid = item.produto
        var statusApi = item.status || 'disponivel'
        var inclusoBase = fpAlacarteEcoInclusoNoBase(pid)
        var ativoAvulso = statusApi === 'ativo'
        var inclusoEfetivo = statusApi === 'incluso' || (inclusoBase && !ativoAvulso)
        var compraveis = fpAlacarteEcoPlanosCompraveis(item)
        var selecionado = !!fpAlacarteState.ecoSelecionados[pid]
        var planoSel = fpAlacarteState.ecoSelecionados[pid] || (compraveis[0] && compraveis[0].plano) || ''
        var cls = ''
        if (inclusoEfetivo && !selecionado) cls = ' incluso'
        if (ativoAvulso) cls = ' contratado'
        if (selecionado) cls = ' novo'

        var precoHtml = ''
        var controleHtml = ''
        var extra = ''

        if (ativoAvulso && !selecionado) {
            controleHtml = '<input type="checkbox" class="form-check-input mt-1" checked disabled>'
            precoHtml = '<span class="fp-alacarte-mod-preco incluso">Ativo</span>'
            extra = '<small class="fp-alacarte-mod-badge-contratado">Assinatura avulsa ativa</small>'
        } else if (inclusoEfetivo && !compraveis.length) {
            controleHtml = '<input type="checkbox" class="form-check-input mt-1" checked disabled>'
            precoHtml = '<span class="fp-alacarte-mod-preco incluso">Incluído</span>'
            extra = '<small class="fp-alacarte-mod-badge-contratado">Incluso no plano base</small>'
            delete fpAlacarteState.ecoSelecionados[pid]
        } else if (compraveis.length) {
            var opt = compraveis.map(function (x) {
                return '<option value="' + x.plano + '"' + (x.plano === planoSel ? ' selected' : '') + '>' +
                    (x.nome_exibicao || x.plano) + ' — R$ ' + Number(x.valor_mensal || 0).toFixed(2) + '</option>'
            }).join('')
            controleHtml = '<input type="checkbox" class="form-check-input fp-alacarte-eco-check mt-1" data-produto="' + pid + '"' +
                (selecionado ? ' checked' : '') + '>'
            precoHtml =
                '<div class="fp-alacarte-eco-acoes">' +
                '<select class="form-select form-select-sm fp-alacarte-eco-plano" data-produto="' + pid + '"' +
                (selecionado ? '' : ' disabled') + '>' + opt + '</select>' +
                (selecionado
                    ? '<span class="fp-alacarte-mod-preco">+ R$ ' + fpAlacarteEcoValorPlano(pid, planoSel).toFixed(2) + '</span>'
                    : '<span class="fp-alacarte-mod-preco text-muted">avulso</span>') +
                '</div>'
            if (inclusoEfetivo) {
                extra = '<small class="text-muted">Já incluso — marque só para upgrade de plano</small>'
            }
            if (pid === 'confvision') {
                extra += (extra ? '<br>' : '') + '<small class="text-muted">Licenças por câmera são compradas à parte no Vision.</small>'
            }
        } else {
            controleHtml = '<input type="checkbox" class="form-check-input mt-1" disabled>'
            precoHtml = '<span class="fp-alacarte-mod-preco text-muted">—</span>'
        }

        var desc = ''
        if (typeof FP_PRODUTOS_ECOSISTEMA !== 'undefined') {
            for (var i = 0; i < FP_PRODUTOS_ECOSISTEMA.length; i++) {
                if (FP_PRODUTOS_ECOSISTEMA[i].id === pid || FP_PRODUTOS_ECOSISTEMA[i].produto === pid) {
                    desc = FP_PRODUTOS_ECOSISTEMA[i].desc || ''
                    break
                }
            }
        }

        html +=
            '<div class="fp-alacarte-mod-item fp-alacarte-eco-item' + cls + '">' +
            controleHtml +
            '<div class="fp-alacarte-mod-info">' +
            '<strong>' + fpEcoNomeProduto(pid) + '</strong>' +
            (desc ? '<small>' + desc + '</small>' : '') +
            extra +
            '</div>' +
            precoHtml +
            '</div>'
    })

    $box.html(html)
    fpAlacarteRefreshResumoSticky()
}

function montarMensagemStatus(resumo) {
    if (fpAlacarteState.planoTravado) {
        var nome = fpPlanoNomeExibicao(fpAlacarteState.planoBase)
        var msg = 'Seu plano base é ' + nome + '. Você pode alterar módulos e contratar produtos complementares.'
        if (fpAlacarteState.liberado && fpAlacarteState.faturasAbertas.length) {
            msg += ' Há fatura em aberto — ao contratar, a fatura anterior será cancelada e uma nova será gerada.'
        } else if (!fpAlacarteState.liberado && fpAlacarteState.faturasAbertas.length) {
            msg += ' Existe fatura em aberto. Alterar a seleção e contratar cancelará a fatura anterior.'
        } else if (fpAlacarteState.liberado) {
            msg += ' Módulos removidos continuam ativos até a próxima fatura; novos módulos e produtos liberam após pagamento.'
        }
        return msg
    }
    if (fpAlacarteState.faturasAbertas.length) {
        return 'Existe fatura em aberto. Ao contratar, a fatura anterior será cancelada e uma nova será gerada.'
    }
    return 'Escolha o plano base, os módulos e os produtos do ecossistema. A cobrança só é aplicada ao clicar em Contratar.'
}

function atualizarMensagemStatus() {
    if (fpAlacarteState.mensagemStatus) {
        $('#fp-alacarte-info').removeClass('d-none').text(fpAlacarteState.mensagemStatus)
    } else {
        $('#fp-alacarte-info').addClass('d-none').text('')
    }
}

function renderBases() {
    var html = ''
    fpAlacarteState.planos.forEach(function (p) {
        var ativo = fpAlacarteState.planoBase === p.plano ? ' ativo' : ''
        var travado = fpAlacarteState.planoTravado ? ' travado' : ''
        var planoAtual = fpAlacarteState.planoTravado && fpAlacarteState.planoBase === p.plano ? ' plano-atual' : ''
        var nome = p.nome_exibicao || fpPlanoNomeExibicao(p.plano)
        var badge = planoAtual
            ? '<small class="d-block mt-1 fw-semibold">Seu plano atual</small>'
            : ''
        html +=
            '<div class="fp-alacarte-base-card' + ativo + travado + planoAtual + '" data-plano="' + p.plano + '">' +
            '<strong>' + nome + '</strong>' +
            '<div class="fp-alacarte-base-preco">R$ ' + Number(p.valor_mensal || 0).toFixed(2) + '<small>/mês</small></div>' +
            '<small class="text-muted">Retenção ' + (p.retencao_dias || '—') + ' dias</small>' +
            badge +
            '</div>'
    })
    $('#fp-alacarte-bases').html(html || '<p class="text-muted">Nenhum plano disponível.</p>')
}

function moduloInclusoNoBase(chave, planoBase) {
    if (!planoBase || typeof fpPlanoModuloLiberadoObj !== 'function') return false
    return fpPlanoModuloLiberadoObj(planoBase, chave)
}

function moduloEstaPendente(chave) {
    return fpAlacarteState.addonsPendentes.indexOf(chave) >= 0
}

function moduloContratado(chave) {
    var ativos = fpAlacarteState.addonsAtivos || []
    return ativos.indexOf(chave) >= 0
}

function moduloAgendadoRemocao(chave) {
    if (!moduloContratado(chave)) return false
    var contratados = fpAlacarteState.addonsIniciais || []
    return contratados.indexOf(chave) < 0
}

function addonChaveValida(chave) {
    if (!chave || typeof chave !== 'string') return false
    if (chave === 'franqueadopro') return false
    if (moduloInclusoNoBase(chave, fpAlacarteState.planoBase)) return false
    var i
    for (i = 0; i < fpAlacarteState.modulos.length; i++) {
        if (fpAlacarteState.modulos[i].chave === chave) return true
    }
    return false
}

function normalizarAddonsLista(lista) {
    if (!lista || !lista.length) return []
    var out = []
    lista.forEach(function (ch) {
        if (addonChaveValida(ch) && out.indexOf(ch) < 0) out.push(ch)
    })
    out.sort()
    return out
}

function moduloLabel(chave) {
    var i
    for (i = 0; i < fpAlacarteState.modulos.length; i++) {
        if (fpAlacarteState.modulos[i].chave === chave) {
            return fpAlacarteState.modulos[i].label || chave
        }
    }
    return chave
}

function moduloPreco(chave) {
    var i
    for (i = 0; i < fpAlacarteState.modulos.length; i++) {
        if (fpAlacarteState.modulos[i].chave === chave) {
            return Number(fpAlacarteState.modulos[i].valor_mensal || 0)
        }
    }
    return 0
}

function calcularAlteracoes() {
    var atual = normalizarAddonsLista(addonsSelecionados())
    var inicial = normalizarAddonsLista(fpAlacarteState.addonsIniciais || [])
    return {
        adicionar: atual.filter(function (c) { return inicial.indexOf(c) < 0 }),
        remover: inicial.filter(function (c) { return atual.indexOf(c) < 0 })
    }
}

function fpAlacarteTextoBotaoAcao(alt) {
    alt = alt || calcularAlteracoes()
    var temEco = temEcoSelecionado()
    if (alt.remover.length && !alt.adicionar.length && !temEco) {
        return { icon: 'bi-check2-circle', texto: 'Confirmar remoção de módulos' }
    }
    if (alt.adicionar.length && !alt.remover.length && !temEco) {
        return { icon: 'bi-cart-check', texto: 'Contratar módulos selecionados' }
    }
    if (!alt.adicionar.length && !alt.remover.length && temEco) {
        return { icon: 'bi-cart-check', texto: 'Contratar produtos do ecossistema' }
    }
    if (alt.adicionar.length || alt.remover.length || temEco) {
        return { icon: 'bi-sliders', texto: 'Confirmar alteração do plano' }
    }
    return { icon: 'bi-cart-check', texto: 'Contratar plano personalizado' }
}

function fpAlacarteMensagemErroAcao(xhr, alt) {
    alt = alt || calcularAlteracoes()
    var msg = 'Não foi possível salvar a alteração do plano.'
    if (alt.remover.length && !alt.adicionar.length) {
        msg = 'Não foi possível registrar a remoção do módulo.'
    } else if (alt.adicionar.length && !alt.remover.length) {
        msg = 'Não foi possível contratar o módulo.'
    }
    try {
        var j = JSON.parse((xhr && xhr.responseText) || '')
        if (j.message) msg = j.message
        else if (j.error) msg = j.error
        else if (j.payload && j.payload.message) msg = j.payload.message
    } catch (e) {
        if (xhr && xhr.responseText) msg = xhr.responseText.slice(0, 300)
    }
    return msg
}

function renderModulos() {
    var plano = fpAlacarteState.planoBase
    if (!plano) {
        $('#fp-alacarte-modulos').html('<p class="text-muted">Selecione um plano base.</p>')
        return
    }

    var grupos = {}
    fpAlacarteState.modulos.forEach(function (m) {
        var g = m.grupo || 'Outros'
        if (!grupos[g]) grupos[g] = []
        grupos[g].push(m)
    })

    var html = ''
    for (var gk in grupos) {
        if (!grupos.hasOwnProperty(gk)) continue
        html += '<div class="fp-alacarte-grupo"><div class="fp-alacarte-grupo-title">' + gk + '</div>'
        grupos[gk].forEach(function (m) {
            var incluso = moduloInclusoNoBase(m.chave, plano)
            var agendadoRemocao = moduloAgendadoRemocao(m.chave)
            var contratado = !incluso && moduloContratado(m.chave) && !agendadoRemocao
            var selecionado = incluso || !!fpAlacarteState.selecionados[m.chave]
            var marcadoRemover = contratado && !selecionado
            var novo = selecionado && !contratado && !incluso && !agendadoRemocao
            var pendente = moduloEstaPendente(m.chave)
            var cls = incluso ? ' incluso' : ''
            if (agendadoRemocao) cls += ' removendo agendado'
            if (contratado && selecionado) cls += ' contratado'
            if (marcadoRemover) cls += ' removendo'
            if (novo) cls += ' novo'

            var precoHtml = incluso
                ? '<span class="fp-alacarte-mod-preco incluso">Incluído</span>'
                : '<span class="fp-alacarte-mod-preco">+ R$ ' + Number(m.valor_mensal || 0).toFixed(2) + '</span>'

            if (pendente && !incluso) {
                precoHtml += '<small class="text-warning d-block">Aguardando pagamento</small>'
            }
            if (marcadoRemover) {
                precoHtml += '<small class="text-danger d-block">Será removido na próxima fatura</small>'
            }
            if (agendadoRemocao) {
                precoHtml += '<small class="text-danger d-block">Remoção agendada — ativo até a próxima fatura</small>'
            }

            if (incluso) {
                fpAlacarteState.selecionados[m.chave] = true
            }

            var controleHtml = ''
            if (incluso) {
                controleHtml = '<input type="checkbox" class="form-check-input fp-alacarte-mod-check mt-1" data-chave="' + m.chave + '" checked disabled>'
            } else if (agendadoRemocao) {
                controleHtml = '<span class="fp-alacarte-mod-badge-remocao" title="Remoção já agendada"><i class="bi bi-clock-history"></i></span>'
            } else if (contratado && selecionado) {
                controleHtml =
                    '<button type="button" class="fp-alacarte-mod-remover" data-chave="' + m.chave + '" title="Remover módulo" aria-label="Remover ' + (m.label || m.chave) + '">' +
                    '<i class="bi bi-x-lg"></i></button>'
            } else if (marcadoRemover) {
                controleHtml =
                    '<button type="button" class="fp-alacarte-mod-desfazer" data-chave="' + m.chave + '" title="Desfazer remoção" aria-label="Desfazer remoção de ' + (m.label || m.chave) + '">' +
                    '<i class="bi bi-arrow-counterclockwise"></i></button>'
            } else {
                controleHtml = '<input type="checkbox" class="form-check-input fp-alacarte-mod-check mt-1" data-chave="' + m.chave + '"' + (selecionado ? ' checked' : '') + '>'
            }

            html +=
                '<div class="fp-alacarte-mod-item' + cls + '">' +
                controleHtml +
                '<div class="fp-alacarte-mod-info">' +
                '<strong>' + (m.label || m.chave) + '</strong>' +
                '<small>' + (m.descricao || '') + '</small>' +
                (contratado && selecionado ? '<small class="fp-alacarte-mod-badge-contratado">No seu plano</small>' : '') +
                '</div>' +
                precoHtml +
                '</div>'
        })
        html += '</div>'
    }

    $('#fp-alacarte-modulos').html(html || '<p class="text-muted">Catálogo de módulos vazio. Contate a central.</p>')
    fpAlacarteRefreshResumoSticky()
}

function addonsSelecionados() {
    var plano = fpAlacarteState.planoBase
    var out = []
    for (var chave in fpAlacarteState.selecionados) {
        if (!fpAlacarteState.selecionados.hasOwnProperty(chave)) continue
        if (!fpAlacarteState.selecionados[chave]) continue
        if (!moduloInclusoNoBase(chave, plano)) {
            out.push(chave)
        }
    }
    out.sort()
    return out
}

function addonsIguais(a, b) {
    if (a.length !== b.length) return false
    for (var i = 0; i < a.length; i++) {
        if (a[i] !== b[i]) return false
    }
    return true
}

function fpAlacarteFormatarData(iso) {
    if (!iso) return ''
    var d = new Date(iso)
    if (isNaN(d.getTime())) return ''
    return d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function fpAlacarteDiasAte(iso) {
    if (!iso) return null
    var alvo = new Date(iso)
    if (isNaN(alvo.getTime())) return null
    var hoje = new Date()
    hoje.setHours(0, 0, 0, 0)
    alvo.setHours(0, 0, 0, 0)
    return Math.round((alvo.getTime() - hoje.getTime()) / 86400000)
}

function fpAlacarteTextoDias(dias) {
    if (dias == null) return ''
    if (dias < 0) return 'há ' + Math.abs(dias) + ' dia' + (Math.abs(dias) === 1 ? '' : 's')
    if (dias === 0) return 'hoje'
    if (dias === 1) return 'em 1 dia'
    return 'em ' + dias + ' dias'
}

function calcularValorFaturaPagar(sim) {
    var ecoValor = valorEcoMensal()
    if (!sim) return Math.round(ecoValor * 100) / 100
    if (!fpAlacarteState.liberado || !fpAlacarteState.planoTravado) {
        return Math.round((Number(sim.valor_total || 0) + ecoValor) * 100) / 100
    }
    if (!temAlteracaoModulos() && !temEcoSelecionado()) return 0

    var atual = addonsSelecionados()
    var inicial = (fpAlacarteState.addonsIniciais || []).slice()
    var adicionou = atual.some(function (c) { return inicial.indexOf(c) < 0 })
    var removeu = inicial.some(function (c) { return atual.indexOf(c) < 0 })

    var soma = 0
    if (!(removeu && !adicionou)) {
        var itens = sim.itens_detalhe || []
        atual.forEach(function (ch) {
            if (inicial.indexOf(ch) >= 0) return
            itens.forEach(function (i) {
                if (i.chave === ch && i.cobrado) {
                    soma += Number(i.valor_mensal || 0)
                }
            })
        })
    }
    soma += ecoValor
    return Math.round(soma * 100) / 100
}

function renderAvisoCobranca(sim, valorFatura) {
    var partes = []
    var totalMensal = Number(sim ? sim.valor_total || 0 : 0) + valorEcoMensal()
    var ass = fpAlacarteState.assinatura || {}
    var faturaAberta = fpAlacarteState.faturasAbertas[0] || null

    if (valorFatura > 0) {
        if (temAlteracao()) {
            partes.push('Ao contratar, será gerada uma fatura de <strong>R$ ' + valorFatura.toFixed(2) + '</strong> para pagamento à sua central.')
        } else if (faturaAberta && Number(faturaAberta.valor_total || 0) > 0) {
            partes.push('Existe fatura em aberto de <strong>R$ ' + Number(faturaAberta.valor_total).toFixed(2) + '</strong>.')
            if (faturaAberta.vencimento_em) {
                var diasFat = fpAlacarteDiasAte(faturaAberta.vencimento_em)
                partes.push('Vencimento ' + fpAlacarteTextoDias(diasFat) + ' (' + fpAlacarteFormatarData(faturaAberta.vencimento_em) + ').')
            }
        }
    } else if (temAlteracaoModulos() && calcularAlteracoes().remover.length > 0 && !temEcoSelecionado()) {
        partes.push('Nenhuma fatura será gerada agora — a alteração vale na próxima renovação.')
    }

    var dataMensal = ass.proxima_cobranca_em || ass.valido_ate || null
    if (fpAlacarteState.liberado && dataMensal && totalMensal > 0) {
        var diasMensal = fpAlacarteDiasAte(dataMensal)
        partes.push(
            'O total mensal de <strong>R$ ' + totalMensal.toFixed(2) + '</strong> será cobrado ' +
            fpAlacarteTextoDias(diasMensal) + ', em <strong>' + fpAlacarteFormatarData(dataMensal) + '</strong>.'
        )
    } else if (!fpAlacarteState.liberado && totalMensal > 0 && valorFatura > 0) {
        partes.push(
            'Após o pagamento da fatura, o valor mensal de <strong>R$ ' + totalMensal.toFixed(2) + '</strong> ' +
            'passa a ser cobrado automaticamente a cada ciclo da assinatura.'
        )
    }

    var el = $('#fp-alacarte-aviso-cobranca')
    if (!partes.length) {
        el.addClass('d-none').html('')
        return
    }
    el.removeClass('d-none').html('<i class="bi bi-info-circle"></i>' + partes.join(' '))
}

function temAlteracaoModulos() {
    var atual = normalizarAddonsLista(addonsSelecionados())
    var inicial = normalizarAddonsLista(fpAlacarteState.addonsIniciais || [])
    return !addonsIguais(atual, inicial)
}

function temAlteracao() {
    return temAlteracaoModulos() || temEcoSelecionado()
}

function simular() {
    var plano = fpAlacarteState.planoBase
    if (!plano) return

    var addons = addonsSelecionados()
    var payload = { produto: 'franqueadopro', plano: plano, addons: addons }
    var cupom = fpAlacarteCupomCodigo()
    if (cupom) payload.cupom_codigo = cupom

    $.ajax({
        url: '/licencaAlacarteSimular',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (r) {
        fpAlacarteState.simulacao = r
        $('#fp-alacarte-erro').addClass('d-none').text('')
        renderResumo(r)
        var habilitar = temAlteracao() || (!fpAlacarteState.planoTravado && !fpAlacarteState.liberado)
        $('#fp-alacarte-contratar').prop('disabled', !habilitar)
    }).fail(function (xhr) {
        var msg = 'Erro ao calcular o plano.'
        try {
            var j = JSON.parse(xhr.responseText || '')
            if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        mostrarErro(msg)
    })
}

function renderResumo(sim) {
    if (!sim) return
    var html = ''
    html += '<div class="fp-alacarte-resumo-linha"><span>Plano base (' + fpPlanoNomeExibicao(sim.plano_base) + ')</span><span>R$ ' + Number(sim.valor_base || 0).toFixed(2) + '</span></div>'

    var itens = sim.itens_detalhe || []
    var extras = itens.filter(function (i) { return i.cobrado })
    if (extras.length) {
        extras.forEach(function (i) {
            html += '<div class="fp-alacarte-resumo-linha"><span>+ ' + (i.label || i.chave) + '</span><span>R$ ' + Number(i.valor_mensal || 0).toFixed(2) + '</span></div>'
        })
    } else {
        html += '<div class="text-muted small mt-1">Nenhum módulo extra selecionado.</div>'
    }

    var ecoItens = ecoProdutosSelecionados()
    if (ecoItens.length) {
        ecoItens.forEach(function (e) {
            html += '<div class="fp-alacarte-resumo-linha"><span>+ ' + e.label + ' (' + e.plano + ')</span><span>R$ ' + Number(e.valor_mensal || 0).toFixed(2) + '</span></div>'
        })
    }

    var alt = calcularAlteracoes()
    if (alt.adicionar.length) {
        html += '<div class="fp-alacarte-alteracoes mt-2">'
        html += '<div class="fp-alacarte-alteracoes-bloco upgrade">'
        html += '<div class="fp-alacarte-alteracoes-titulo"><i class="bi bi-plus-circle"></i> Módulos a adicionar</div>'
        html += '<ul class="fp-alacarte-alteracoes-lista">'
        alt.adicionar.forEach(function (ch) {
            html += '<li>' + moduloLabel(ch) + ' <span class="text-muted">(+ R$ ' + moduloPreco(ch).toFixed(2) + '/mês)</span></li>'
        })
        html += '</ul></div></div>'
    }
    if (alt.remover.length) {
        html += '<div class="fp-alacarte-alteracoes mt-2">'
        html += '<div class="fp-alacarte-alteracoes-bloco downgrade">'
        html += '<div class="fp-alacarte-alteracoes-titulo"><i class="bi bi-dash-circle"></i> Módulos a remover (downgrade)</div>'
        html += '<ul class="fp-alacarte-alteracoes-lista">'
        alt.remover.forEach(function (ch) {
            html += '<li>' + moduloLabel(ch) + ' <span class="text-muted">(− R$ ' + moduloPreco(ch).toFixed(2) + '/mês)</span></li>'
        })
        html += '</ul></div></div>'
    }
    if (ecoItens.length) {
        html += '<div class="fp-alacarte-alteracoes mt-2">'
        html += '<div class="fp-alacarte-alteracoes-bloco upgrade">'
        html += '<div class="fp-alacarte-alteracoes-titulo"><i class="bi bi-grid-3x3-gap"></i> Produtos do ecossistema</div>'
        html += '<ul class="fp-alacarte-alteracoes-lista">'
        ecoItens.forEach(function (e) {
            html += '<li>' + e.label + ' <span class="text-muted">(' + e.plano + ' · + R$ ' + Number(e.valor_mensal || 0).toFixed(2) + '/mês)</span></li>'
        })
        html += '</ul></div></div>'
    }

    if (fpAlacarteState.liberado && temAlteracaoModulos()) {
        var adicionou = alt.adicionar.length > 0
        var removeu = alt.remover.length > 0
        if (adicionou && !removeu) {
            html += '<div class="text-muted small mt-2">Será gerada fatura apenas dos novos módulos.</div>'
        } else if (removeu && !adicionou && !ecoItens.length) {
            html += '<div class="text-muted small mt-2">A redução vale na próxima fatura. Os módulos removidos seguem ativos até lá.</div>'
        } else if (adicionou && removeu) {
            html += '<div class="text-muted small mt-2">Fatura aberta será cancelada; nova fatura dos módulos adicionados.</div>'
        }
    }
    if (ecoItens.length) {
        html += '<div class="text-muted small mt-2">Produtos do ecossistema geram fatura própria ao contratar.</div>'
    }

    var valorFatura = calcularValorFaturaPagar(sim)
    var totalMensal = Number(sim.valor_total || 0) + valorEcoMensal()
    if (sim.valor_desconto_cupom != null && Number(sim.valor_desconto_cupom) > 0 && valorFatura > 0) {
        html += '<div class="fp-alacarte-resumo-linha text-success mt-2"><span>Cupom (FranqueadoPro)</span><span>− R$ ' + Number(sim.valor_desconto_cupom).toFixed(2) + '</span></div>'
        var faturaFp = Math.max(0, valorFatura - valorEcoMensal())
        var comCupom = Math.max(0, Math.round((faturaFp - Number(sim.valor_desconto_cupom)) * 100) / 100) + valorEcoMensal()
        valorFatura = Math.round(comCupom * 100) / 100
    }

    $('#fp-alacarte-resumo-detalhe').html(html)
    $('#fp-alacarte-total').text('R$ ' + totalMensal.toFixed(2))
    $('#fp-alacarte-total-fatura').text('R$ ' + valorFatura.toFixed(2))
    renderAvisoCobranca(sim, valorFatura)
    var btn = fpAlacarteTextoBotaoAcao(alt)
    $('#fp-alacarte-contratar').html('<i class="bi ' + btn.icon + '"></i> ' + btn.texto)
    fpAlacarteRefreshResumoSticky()
}

function contratarAlacarte() {
    if (!temAlteracao() && fpAlacarteState.planoTravado) {
        alert('Nenhuma alteração na seleção de módulos ou produtos.')
        return
    }

    var plano = fpAlacarteState.planoBase
    var addons = addonsSelecionados()
    var ecoItens = ecoProdutosSelecionados()
    var totalMensal = (
        (fpAlacarteState.simulacao ? Number(fpAlacarteState.simulacao.valor_total || 0) : 0) + valorEcoMensal()
    ).toFixed(2)
    var totalFatura = calcularValorFaturaPagar(fpAlacarteState.simulacao)
    var alt = calcularAlteracoes()
    var cupom = fpAlacarteCupomCodigo()
    if (fpAlacarteState.simulacao && fpAlacarteState.simulacao.valor_desconto_cupom != null && Number(fpAlacarteState.simulacao.valor_desconto_cupom) > 0) {
        var faturaFp = Math.max(0, totalFatura - valorEcoMensal())
        totalFatura = Math.max(0, faturaFp - Number(fpAlacarteState.simulacao.valor_desconto_cupom)) + valorEcoMensal()
    }
    totalFatura = Number(totalFatura).toFixed(2)

    var msgConfirm = 'Confirmar alteração do plano personalizado?\n\n'
    msgConfirm += 'Total mensal após alteração: R$ ' + totalMensal + '/mês.\n'
    if (Number(totalFatura) > 0) {
        msgConfirm += 'Fatura a pagar agora: R$ ' + totalFatura + '.\n'
    }
    if (cupom) {
        msgConfirm += 'Cupom (FranqueadoPro): ' + cupom + '\n'
    }
    if (alt.remover.length) {
        msgConfirm += '\nRemover:\n• ' + alt.remover.map(moduloLabel).join('\n• ')
    }
    if (alt.adicionar.length) {
        msgConfirm += '\n\nAdicionar módulos:\n• ' + alt.adicionar.map(moduloLabel).join('\n• ')
    }
    if (ecoItens.length) {
        msgConfirm += '\n\nContratar produtos:\n• ' + ecoItens.map(function (e) {
            return e.label + ' (' + e.plano + ')'
        }).join('\n• ')
    }
    if (alt.remover.length && !alt.adicionar.length && !ecoItens.length) {
        msgConfirm += '\n\nA redução entra na próxima fatura. Os módulos removidos seguem ativos até lá.'
    } else if ((alt.adicionar.length || ecoItens.length) && !alt.remover.length && Number(totalFatura) > 0) {
        msgConfirm += '\n\nSerá gerada fatura dos itens novos.'
    } else if (alt.adicionar.length && alt.remover.length) {
        msgConfirm += '\n\nFaturas em aberto serão canceladas se necessário; nova fatura só dos módulos adicionados.'
    }

    if (!fpAlacarteState.liberado) {
        if (fpAlacarteState.faturasAbertas.length) {
            msgConfirm += '\n\nA fatura em aberto será cancelada e uma nova será gerada.'
        } else if (Number(totalFatura) > 0) {
            msgConfirm += '\n\nSerá gerada uma fatura em aberto para pagamento à sua central.'
        }
    }

    if (!confirm(msgConfirm)) {
        return
    }

    $('#fp-alacarte-contratar').prop('disabled', true)

    var precisaAlacarte = temAlteracaoModulos() || (!fpAlacarteState.planoTravado && !fpAlacarteState.liberado)

    function irParaFaturas(msg) {
        if (msg) alert(msg)
        if (typeof fpLicCarregarLogado === 'function') {
            fpLicCarregarLogado(function () {
                window.location.href = '/carregar-meu-plano?tab=faturas'
            })
        } else {
            window.location.href = '/carregar-meu-plano?tab=faturas'
        }
    }

    function contratarEcoEmSerie(idx, acaoFp) {
        if (idx >= ecoItens.length) {
            if (acaoFp === 'reduzir_proxima_fatura' && !ecoItens.length) {
                alert('Alteração registrada! A redução será aplicada na próxima fatura.')
                window.location.reload()
                return
            }
            if (ecoItens.length && acaoFp === 'reduzir_proxima_fatura') {
                irParaFaturas('Módulos: redução na próxima fatura. Produtos do ecossistema: veja Faturas.')
                return
            }
            if (acaoFp === 'adicionar_modulos' || ecoItens.length) {
                irParaFaturas('Contratação iniciada! Acesse Faturas e Cobrança para pagamento.')
                return
            }
            irParaFaturas()
            return
        }
        var item = ecoItens[idx]
        $.ajax({
            url: '/licencaContratarProduto',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: item.produto, plano: item.plano })
        }).done(function () {
            contratarEcoEmSerie(idx + 1, acaoFp)
        }).fail(function (xhr) {
            $('#fp-alacarte-contratar').prop('disabled', false)
            simular()
            var msg = 'Não foi possível contratar ' + item.label + '.'
            try {
                var j = JSON.parse(xhr.responseText || '')
                if (j.message) msg = j.message
                else if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
            } catch (e) { /* ignore */ }
            mostrarErro(msg)
            alert(msg + (idx > 0 || acaoFp ? '\n\nParte da contratação pode já ter sido registrada — confira Meu Plano / Faturas.' : ''))
        })
    }

    if (!precisaAlacarte) {
        contratarEcoEmSerie(0, '')
        return
    }

    var payload = { produto: 'franqueadopro', plano: plano, addons: addons }
    if (cupom) payload.cupom_codigo = cupom
    $.ajax({
        url: '/licencaContratarAlacarte',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (resp) {
        var acao = resp && resp.acao ? resp.acao : ''
        contratarEcoEmSerie(0, acao)
    }).fail(function (xhr) {
        $('#fp-alacarte-contratar').prop('disabled', false)
        simular()
        var msg = fpAlacarteMensagemErroAcao(xhr, alt)
        mostrarErro(msg)
        alert(msg)
    })
}

function mostrarErro(msg) {
    $('#fp-alacarte-erro').removeClass('d-none').text(msg)
}

var fpAlacarteStickyTopGap = 16

function fpAlacarteInitResumoSticky() {
    $(window).on('scroll.fpAlacarteSticky resize.fpAlacarteSticky', fpAlacarteSyncResumoSticky)
    fpAlacarteSyncResumoSticky()
}

function fpAlacarteRefreshResumoSticky() {
    window.requestAnimationFrame(fpAlacarteSyncResumoSticky)
}

function fpAlacarteSyncResumoSticky() {
    var $sticky = $('#fp-alacarte-resumo-sticky')
    var $spacer = $('#fp-alacarte-resumo-spacer')
    var $col = $('.fp-alacarte-col-resumo')
    var $row = $('.fp-alacarte-layout-row')
    if (!$sticky.length || !$col.length || !$row.length) return

    if (window.innerWidth < 992) {
        $sticky.removeClass('is-fixed is-bottom').css({ top: '', left: '', width: '' })
        $spacer.height(0)
        return
    }

    var colEl = $col[0]
    var rowEl = $row[0]
    var stickyEl = $sticky[0]
    var colRect = colEl.getBoundingClientRect()
    var stickyH = stickyEl.offsetHeight
    var scrollY = window.pageYOffset || document.documentElement.scrollTop
    var colTop = colRect.top + scrollY
    var rowTop = rowEl.getBoundingClientRect().top + scrollY
    var rowHeight = rowEl.offsetHeight
    var start = colTop - fpAlacarteStickyTopGap
    var end = rowTop + rowHeight - stickyH - fpAlacarteStickyTopGap

    $sticky.removeClass('is-fixed is-bottom')

    if (scrollY <= start) {
        $sticky.css({ top: '', left: '', width: '' })
        $spacer.height(0)
        return
    }

    $spacer.height(stickyH)

    if (scrollY >= end) {
        $sticky.addClass('is-bottom').css({
            top: (rowHeight - stickyH) + 'px',
            left: '',
            width: ''
        })
        return
    }

    $sticky.addClass('is-fixed').css({
        top: fpAlacarteStickyTopGap + 'px',
        left: colRect.left + 'px',
        width: colRect.width + 'px'
    })
}

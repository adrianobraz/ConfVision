/* Hub de produtos ConfMonit na tela principal */

function fpProdutosRenderHub() {
    var $box = $('#fp-produtos-hub')
    if (!$box.length || typeof FP_PRODUTOS_ECOSISTEMA === 'undefined') return

    var html = '<div class="fp-produtos-hub-title"><i class="bi bi-grid-3x3-gap-fill"></i> Produtos ConfMonit</div>' +
        '<p class="text-muted small mb-2">Acesso conforme o plano ativo do FranqueadoPro (Lite, Pro ou Pro+).</p>' +
        '<div class="row g-3 fp-produtos-hub-cards justify-content-center">'

    FP_PRODUTOS_ECOSISTEMA.forEach(function (p) {
        var minLabel = p.minimo === 'pro_plus' ? 'Pro+' : (p.minimo === 'pro' ? 'Pro' : 'Lite')
        html += '<div class="col-6 col-md-4 col-lg-2">' +
            '<a href="#" class="fp-produto-card" data-fp-produto="' + p.id + '"' +
            (p.url ? ' data-fp-url="' + p.url + '"' : '') + '>' +
            '<span class="fp-produto-icone" aria-hidden="true"><i class="bi ' + p.icon + '"></i></span>' +
            '<span class="fp-produto-nome">' + p.nome + '</span>' +
            '<small class="fp-produto-desc">' + p.desc + '</small>' +
            '<small class="fp-produto-plano-min">Plano mín.: ' + minLabel + '</small>' +
            '</a></div>'
    })
    html += '</div>'
    $box.html(html)

    fpProdutosAplicarLicenca(function () {
        if (typeof fpLicInitTooltips === 'function') fpLicInitTooltips()
    })
}

function fpProdutosBloquearCard($card, motivo) {
    $card.addClass('fp-produto-desabilitado fp-lic-bloqueado')
    $card.attr('href', '#')
    $card.removeAttr('target').removeAttr('rel')
    $card.attr('data-fp-lic-motivo', motivo)
    $card.attr('title', motivo)
    $card.attr('data-bs-toggle', 'tooltip')
    $card.attr('data-bs-placement', 'top')
}

function fpProdutosLiberarCard($card, url) {
    $card.removeClass('fp-produto-desabilitado fp-lic-bloqueado fp-produto-incluso')
    $card.attr('href', url)
    $card.attr('target', '_blank')
    $card.attr('rel', 'noopener noreferrer')
    $card.attr('data-fp-produto-ok', '1')
    $card.removeAttr('data-fp-lic-motivo')
    $card.removeAttr('data-bs-toggle')
    $card.removeAttr('title')
}

function fpProdutosLiberarInfo($card, produto) {
    var tip = 'Incluído no seu plano FranqueadoPro.'
    if (produto && produto.ativacao) {
        tip = 'Incluído no Pro+. ' + produto.ativacao
    }
    $card.removeClass('fp-produto-desabilitado fp-lic-bloqueado')
    $card.addClass('fp-produto-incluso')
    $card.attr('href', '#')
    $card.removeAttr('target').removeAttr('rel')
    $card.attr('data-fp-produto-ok', '1')
    $card.attr('data-fp-lic-motivo', tip)
    $card.attr('title', tip)
    $card.attr('data-bs-toggle', 'tooltip')
    $card.attr('data-bs-placement', 'top')
}

function fpProdutosAplicarLicenca(next) {
    if (typeof FP_PRODUTOS_ECOSISTEMA === 'undefined') {
        if (typeof next === 'function') next()
        return
    }

    var st = typeof fpLicEstado === 'function' ? fpLicEstado() : {}
    var planoFp = st.plano || ''
    var fpAtivo = st.liberado === true

    FP_PRODUTOS_ECOSISTEMA.forEach(function (p) {
        var $card = $('.fp-produto-card[data-fp-produto="' + p.id + '"]')
        if (!$card.length) return

        if (!fpAtivo) {
            fpProdutosBloquearCard($card, 'Assinatura FranqueadoPro inativa. Regularize em Meu Plano.')
            return
        }

        var liberadoPlano = typeof fpPlanoProdutoLiberado === 'function'
            ? fpPlanoProdutoLiberado(planoFp, p.id)
            : false

        if (!liberadoPlano) {
            var motivo = typeof fpPlanoProdutoMotivo === 'function'
                ? fpPlanoProdutoMotivo(p.id)
                : ('Disponível a partir do plano ' + (p.minimo || 'pro_plus') + ' do FranqueadoPro.')
            fpProdutosBloquearCard($card, motivo)
            return
        }

        if (p.url) {
            fpProdutosLiberarCard($card, p.url)
        } else if (p.infoOnly) {
            fpProdutosLiberarInfo($card, p)
        }
    })

    if (typeof next === 'function') next()
}

$(document).ready(function () {
    fpProdutosRenderHub()
})

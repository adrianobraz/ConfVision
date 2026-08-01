/* Hub de produtos complementares na tela principal */

function fpProdutosRenderHub() {
    var $box = $('#fp-produtos-hub')
    if (!$box.length || typeof FP_PRODUTOS_ECOSISTEMA === 'undefined') return

    var html = '<div class="fp-produtos-hub-title"><i class="bi bi-grid-3x3-gap-fill"></i> Produtos complementares</div>' +
        '<p class="text-muted small mb-2 text-center">Incluídos no seu plano FranqueadoPro ou contratáveis à parte em Meu Plano.</p>' +
        '<div class="row g-3 fp-produtos-hub-cards justify-content-center">'

    FP_PRODUTOS_ECOSISTEMA.forEach(function (p) {
        var minLabel = p.minimo === 'pro_plus' ? 'Pro+' : (p.minimo === 'pro' ? 'Pro' : 'Lite')
        html += '<div class="col-6 col-md-4 col-lg-2 d-flex justify-content-center">' +
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

function fpProdutosLinkContratar($card, motivo) {
    $card.removeClass('fp-produto-desabilitado')
    $card.addClass('fp-lic-bloqueado')
    $card.attr('href', '/carregar-meu-plano')
    $card.removeAttr('target').removeAttr('rel')
    $card.attr('data-fp-lic-motivo', motivo)
    $card.attr('title', motivo)
    $card.attr('data-bs-toggle', 'tooltip')
    $card.attr('data-bs-placement', 'top')
    $card.find('.fp-produto-plano-min').text('Contratar em Meu Plano')
}

function fpProdutosMapaEcossistema(dados) {
    var mapa = {}
    ;(dados || []).forEach(function (p) {
        if (p && p.produto) mapa[p.produto] = p
    })
    return mapa
}

function fpProdutosAplicarLicenca(next) {
    if (typeof FP_PRODUTOS_ECOSISTEMA === 'undefined') {
        if (typeof next === 'function') next()
        return
    }

    var st = typeof fpLicEstado === 'function' ? fpLicEstado() : {}
    var planoFp = st.plano || ''
    var fpAtivo = st.liberado === true

    function aplicar(ecoMap) {
        FP_PRODUTOS_ECOSISTEMA.forEach(function (p) {
            var $card = $('.fp-produto-card[data-fp-produto="' + p.id + '"]')
            if (!$card.length) return

            if (p.infoOnly) {
                if (!fpAtivo || !fpPlanoProdutoLiberado(planoFp, p.id)) {
                    fpProdutosBloquearCard($card, typeof fpPlanoProdutoMotivo === 'function' ? fpPlanoProdutoMotivo(p.id) : 'Disponível no Pro+.')
                    return
                }
                fpProdutosLiberarInfo($card, p)
                return
            }

            var eco = ecoMap[p.id] || ecoMap[p.produto] || null
            var statusEco = eco && eco.status
            var liberadoPlano = fpAtivo && typeof fpPlanoProdutoLiberado === 'function'
                ? fpPlanoProdutoLiberado(planoFp, p.id)
                : false
            var liberadoEco = statusEco === 'incluso' || statusEco === 'ativo'

            if (liberadoPlano || liberadoEco) {
                if (p.url) {
                    fpProdutosLiberarCard($card, p.url)
                    if (statusEco === 'incluso' || (liberadoPlano && typeof fpPlanoProdutoInclusoTier === 'function')) {
                        var tier = (eco && eco.plano_incluso_fp) ||
                            (typeof fpPlanoProdutoInclusoTier === 'function' ? fpPlanoProdutoInclusoTier(planoFp, p.id) : '')
                        if (tier) {
                            $card.find('.fp-produto-plano-min').text('Incluso no FP (' + tier + ')')
                        } else if (statusEco === 'ativo') {
                            $card.find('.fp-produto-plano-min').text('Assinatura ativa')
                        }
                    } else if (statusEco === 'ativo') {
                        $card.find('.fp-produto-plano-min').text('Assinatura ativa')
                    }
                }
                return
            }

            var motivo = 'Não incluso no seu plano. Contrate em Meu Plano ou faça upgrade do FranqueadoPro.'
            if (typeof fpPlanoProdutoMotivo === 'function') motivo = fpPlanoProdutoMotivo(p.id)
            fpProdutosLinkContratar($card, motivo)
        })

        if (typeof next === 'function') next()
    }

    $.ajax({
        url: '/licencaEcossistema',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        aplicar(fpProdutosMapaEcossistema((r && r.dados) || []))
    }).fail(function () {
        aplicar({})
    })
}

$(document).ready(function () {
    fpProdutosRenderHub()
})

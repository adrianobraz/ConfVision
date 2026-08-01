/* Gate de licenca/plano no front — complementa middleware Go */

var FP_LIC_BTN_URL = {
    BtnGerenciarCadastroAlarme: '/carregar-gerenciar-dispositivo',
    BtnGerenciarUsuarioAlarme: '/carregar-gerenciar-usuarios-alarme',
    BtnGerenciarSetoresAlarme: '/carregar-gerenciar-setores-alarme',
    BtnGerenciarProcedimentoAtendimento: '/carregar-gerenciar-procedimento-atendimento',
    BtnGerenciarContactidPersonalizado: '/carregar-gerenciar-contactid-personalizado',
    BtnGerenciarGradeHorario: '/carregar-gerenciar-grade',
    BtnGerenciarEnvioEvento: '/gerenciar-configuracao-email-eveto'
}

function fpLicEstado() {
    try {
        return JSON.parse(localStorage.getItem('fpLicencaEstado') || '{}')
    } catch (e) {
        return {}
    }
}

function fpLicSalvarEstado(resp) {
    try {
        var ass = resp.assinatura || {}
        var plano = ass.plano || resp.plano || ''
        var modulos = typeof fpLicMergeModulosDados === 'function'
            ? fpLicMergeModulosDados(plano, resp.modulos_json || ass.modulos_json || {})
            : (resp.modulos_json || ass.modulos_json || {})

        if (resp.liberado) {
            var contratados = resp.addons_contratados || ass.addons_json || []
            var pendentes = resp.addons_pendentes || ass.addons_pendentes_json || []
            if (!Array.isArray(contratados)) contratados = []
            if (!Array.isArray(pendentes)) pendentes = []
            contratados.forEach(function (ch) {
                if (!ch || pendentes.indexOf(ch) >= 0) return
                var chave = ch === 'franqueadopro' ? 'franqueadopro.saas' : ch
                modulos[chave] = true
            })
        }

        var efetivos = resp.addons_efetivos || []
        if (resp.liberado && efetivos.length) {
            efetivos.forEach(function (ch) {
                if (!ch) return
                var chave = ch === 'franqueadopro' ? 'franqueadopro.saas' : ch
                modulos[chave] = true
            })
        }

        localStorage.setItem('fpLicencaEstado', JSON.stringify({
            liberado: resp.liberado,
            motivo: resp.motivo,
            plano: plano,
            modulos: modulos,
            limites: resp.limites_json || ass.limites_json || {},
            retencao_dias: resp.retencao_dias || 30
        }))
    } catch (e) { /* ignore */ }
}

function fpLicCarregarLogado(next) {
    $.ajax({
        url: '/licencaResumo',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ produto: 'franqueadopro' }),
        cache: false
    }).done(function (r) {
        if (!r) {
            if (typeof next === 'function') next()
            return
        }
        var redir = fpLicProcessarResposta(r)
        if (!redir && typeof next === 'function') next()
    }).fail(function () {
        if (typeof fpLicRefreshUI === 'function') fpLicRefreshUI()
        if (typeof next === 'function') next()
    })
}

function fpLicRemoverModoSuspenso() {
    document.body.classList.remove('fp-lic-presa-fatura')
    var atalhos = document.getElementById('fp-atalhos')
    if (atalhos) {
        atalhos.classList.remove('fp-lic-nav-bloqueado', 'fp-lic-bloqueado')
        atalhos.style.removeProperty('display')
    }
    document.querySelectorAll('.fp-lic-nav-bloqueado').forEach(function (el) {
        el.classList.remove('fp-lic-nav-bloqueado', 'fp-lic-bloqueado')
        if (el.classList.contains('fp-navbar-brand')) el.setAttribute('href', '/carregar-menu-principal')
        if (el.classList.contains('fp-home-btn')) el.setAttribute('href', '/carregar-menu-principal')
    })
    document.querySelectorAll('#fp-atalhos .fp-menu-card').forEach(function (el) {
        if (typeof fpLicHabilitarElemento === 'function') fpLicHabilitarElemento(el)
    })
    var combo = document.getElementById('fp-nav-toggle')
    if (combo) combo.removeAttribute('disabled')
    var back = document.querySelector('.fp-page-header-back')
    if (back) back.classList.remove('d-none')
    if (window.__fpPlanoPollSuspenso) {
        clearInterval(window.__fpPlanoPollSuspenso)
        window.__fpPlanoPollSuspenso = null
    }
}

function fpLicProcessarResposta(r) {
    if (!r) return false
    var eraPreso = fpLicDevePrenderFatura() || document.body.classList.contains('fp-lic-presa-fatura')
    fpLicSalvarEstado(r)
    if (r.liberado) {
        if (eraPreso) {
            window.location.replace('/carregar-menu-principal')
            return true
        }
        fpLicRemoverModoSuspenso()
        fpLicRefreshUI()
        return false
    }
    if (fpLicAplicarModoSuspenso()) return true
    fpLicRefreshUI()
    return false
}

function fpLicRefreshUI() {
    if (typeof fpProdutosAplicarLicenca === 'function') {
        fpProdutosAplicarLicenca(function () {
            if (typeof fpLicInitTooltips === 'function') fpLicInitTooltips()
        })
    }
    if (typeof fpBuildNav === 'function') {
        fpBuildNav()
    }
    if (typeof fpAplicarLicencaUI === 'function') fpAplicarLicencaUI()
}

function fpLicAcessoGlobalOk() {
    return fpLicEstado().liberado === true
}

function fpLicRotasLivresLicenca(url) {
    if (!url) return false
    var u = (url.split('?')[0] || '').toLowerCase().replace(/\/$/, '') || '/'
    return u === '/carregar-meu-plano'
        || u === '/carregar-alterar-senha'
        || u === '/carregar-dados'
        || u === '/logout'
}

function fpLicEhLogout(el) {
    if (!el) return false
    if (el.id === 'logout') return true
    var href = (el.getAttribute('href') || '').split('?')[0].toLowerCase().replace(/\/$/, '')
    return href === '/logout'
}

function fpLicDevePrenderFatura() {
    var st = fpLicEstado()
    if (st.liberado !== false) return false
    return st.motivo === 'suspensa' || st.motivo === 'vencida' || st.motivo === 'pendente'
}

function fpLicUrlFaturas() {
    return '/carregar-meu-plano?tab=faturas'
}

function fpLicEstaNaFaturas() {
    var path = (window.location.pathname || '').toLowerCase().replace(/\/$/, '') || '/'
    if (path !== '/carregar-meu-plano') return false
    return new URLSearchParams(window.location.search).get('tab') === 'faturas'
}

function fpLicAplicarModoSuspenso() {
    if (!fpLicDevePrenderFatura()) return false
    if (!fpLicEstaNaFaturas()) {
        window.location.replace(fpLicUrlFaturas())
        return true
    }
    document.body.classList.add('fp-lic-presa-fatura')
    document.querySelectorAll('.fp-navbar-brand, .fp-home-btn, #fp-atalhos').forEach(function (el) {
        el.classList.add('fp-lic-nav-bloqueado')
        if (el.tagName === 'A') {
            el.setAttribute('href', '#')
            el.classList.add('fp-lic-bloqueado')
        }
    })
    var back = document.querySelector('.fp-page-header-back')
    if (back) back.classList.add('d-none')
    var combo = document.getElementById('fp-nav-toggle')
    if (combo) combo.setAttribute('disabled', 'disabled')
    return true
}

function fpLicModulosEfetivos() {
    var st = fpLicEstado()
    if (typeof fpLicMergeModulosDados === 'function') {
        return fpLicMergeModulosDados(st.plano || '', st.modulos || {})
    }
    return st.modulos || {}
}

function fpLicEhMenuGrupo(chave) {
    return chave === 'configuracao' || chave === 'atendimento' ||
        chave === 'minha-empresa' || chave === 'relatorio'
}

function fpLicModuloLiberado(chave) {
    if (!chave) return true
    if (!fpLicAcessoGlobalOk()) return false
    if (fpLicEhMenuGrupo(chave)) return true
    var mod = fpLicModulosEfetivos()
    if (typeof fpLicModuloLiberadoPorMapa === 'function') {
        if (fpLicModuloLiberadoPorMapa(mod, chave)) return true
        if (chave === 'franqueadopro.saas' && mod['franqueadopro'] === true) return true
        // Central de Disparos: Pro+ (chave propria) OU addon Inteligencia Artificial
        if (chave.indexOf('relatorio.central-disparos') === 0) {
            if (fpLicModuloLiberadoPorMapa(mod, 'relatorio.central-disparos')) return true
            if (fpLicModuloLiberadoPorMapa(mod, 'atendimento.inteligencia-artificial')) return true
        }
        return false
    }
    var st = fpLicEstado()
    var min = typeof FP_PLANO_MINIMO_POR_CHAVE !== 'undefined' ? FP_PLANO_MINIMO_POR_CHAVE[chave] : null
    if (min && typeof fpPlanoLiberadoNoPlano === 'function') {
        return fpPlanoLiberadoNoPlano(min, st.plano || '')
    }
    return true
}

function fpLicMotivoModulo(chave) {
    var st = fpLicEstado()
    if (!fpLicAcessoGlobalOk()) {
        if (st.motivo === 'sem_assinatura') return 'Assinatura não encontrada. Solicite a liberação do plano na central.'
        if (st.motivo === 'pendente') return 'Sua assinatura está pendente de liberação.'
        if (st.motivo === 'vencida') return 'Sua assinatura está vencida. Regularize a fatura em aberto para restaurar o acesso.'
        if (st.motivo === 'suspensa') return 'Sistema suspenso. Regularize a fatura em aberto para liberar o acesso.'
        return 'Licença inativa. Acesse Meu Plano para mais informações.'
    }
    if (typeof fpLicMotivoPlano === 'function') return fpLicMotivoPlano(chave)
    return 'Este módulo não está incluído no seu plano atual.'
}

function fpLicUrlDoElemento(el) {
    var $el = $(el)
    var href = $el.attr('href')
    if (href && href !== '#') return href
    var fpUrl = $el.attr('data-fp-url')
    if (fpUrl) return fpUrl
    var id = el.id || $el.attr('id')
    if (id && FP_LIC_BTN_URL[id]) return FP_LIC_BTN_URL[id]
    if (typeof FP_NAV_BTN_URL !== 'undefined' && id && FP_NAV_BTN_URL[id]) return FP_NAV_BTN_URL[id]
    return null
}

function fpLicChaveDoElemento(el) {
    var ch = el.getAttribute('data-fp-lic-chave')
    if (ch) return ch
    ch = el.getAttribute('data-grupo')
    if (ch) return ch
    var url = fpLicUrlDoElemento(el)
    if (url && typeof fpPermChavePorUrl === 'function') return fpPermChavePorUrl(url)
    return null
}

function fpLicMostrarTooltip(el, motivo) {
    el.setAttribute('title', motivo)
    el.setAttribute('data-bs-toggle', 'tooltip')
    el.setAttribute('data-bs-placement', el.getAttribute('data-bs-placement') || 'top')
    if (typeof bootstrap !== 'undefined' && bootstrap.Tooltip) {
        var existente = bootstrap.Tooltip.getInstance(el)
        if (existente) existente.dispose()
        var tip = new bootstrap.Tooltip(el)
        tip.show()
        setTimeout(function () { tip.hide() }, 4500)
    } else {
        alert(motivo)
    }
}

function fpLicDesabilitarElemento(el, chave) {
    if (!el || fpLicEhLogout(el)) return
    var motivo = fpLicMotivoModulo(chave)
    var $el = $(el)

    if (el.tagName === 'A') {
        var hrefAtual = el.getAttribute('href')
        if (hrefAtual && hrefAtual !== '#' && !el.getAttribute('data-fp-lic-href')) {
            el.setAttribute('data-fp-lic-href', hrefAtual)
        }
        el.setAttribute('href', '#')
    }

    el.classList.add('fp-menu-card-disabled', 'fp-lic-bloqueado')
    el.setAttribute('aria-disabled', 'true')
    el.setAttribute('data-fp-lic-motivo', motivo)
    if (chave) el.setAttribute('data-fp-lic-chave', chave)
    el.setAttribute('title', motivo)
    el.setAttribute('data-bs-toggle', 'tooltip')

    $el.off('click.fpLic keydown.fpLic')
    $el.on('click.fpLic', function (e) {
        e.preventDefault()
        e.stopImmediatePropagation()
        fpLicMostrarTooltip(el, motivo)
        return false
    })
    $el.on('keydown.fpLic', function (e) {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            fpLicMostrarTooltip(el, motivo)
        }
    })
}

function fpLicHabilitarElemento(el) {
    if (!el || !el.classList.contains('fp-lic-bloqueado')) return
    var $el = $(el)
    var href = el.getAttribute('data-fp-lic-href')
    if (href && el.tagName === 'A') el.setAttribute('href', href)
    el.removeAttribute('data-fp-lic-href')
    el.classList.remove('fp-menu-card-disabled', 'fp-lic-bloqueado')
    el.removeAttribute('aria-disabled')
    el.removeAttribute('data-fp-lic-motivo')
    el.removeAttribute('title')
    el.removeAttribute('data-bs-toggle')
    el.removeAttribute('data-bs-placement')
    $el.off('click.fpLic keydown.fpLic')
    if (typeof bootstrap !== 'undefined' && bootstrap.Tooltip) {
        var tip = bootstrap.Tooltip.getInstance(el)
        if (tip) tip.dispose()
    }
}

function fpLicInitTooltips() {
    if (typeof bootstrap === 'undefined' || !bootstrap.Tooltip) return
    document.querySelectorAll('.fp-lic-bloqueado[data-bs-toggle="tooltip"]').forEach(function (el) {
        bootstrap.Tooltip.getOrCreateInstance(el)
    })
}

function fpLicClickBloqueado(e) {
    if (e.target.closest('#logout, a[href="/logout"]')) return
    var el = e.target.closest('.fp-lic-bloqueado')
    if (!el || fpLicEhLogout(el)) return
    var chave = fpLicChaveDoElemento(el)
    if (chave && fpLicModuloLiberado(chave)) {
        fpLicHabilitarElemento(el)
        var url = fpLicUrlDoElemento(el)
        if (url && url !== '#') {
            e.preventDefault()
            e.stopPropagation()
            window.location.href = url
        }
        return
    }
    e.preventDefault()
    e.stopPropagation()
    e.stopImmediatePropagation()
    fpLicMostrarTooltip(el, el.getAttribute('data-fp-lic-motivo') || fpLicMotivoModulo(chave))
    return false
}

function fpLicAplicarNoElemento(el) {
    if (!el) return
    if (fpLicEhLogout(el)) {
        fpLicHabilitarElemento(el)
        return
    }
    var url = fpLicUrlDoElemento(el)
    if (url && fpLicRotasLivresLicenca(url)) {
        fpLicHabilitarElemento(el)
        return
    }
    var chave = fpLicChaveDoElemento(el)
    if (!chave) {
        if (!fpLicAcessoGlobalOk() && url) fpLicDesabilitarElemento(el, '')
        else fpLicHabilitarElemento(el)
        return
    }
    if (fpLicModuloLiberado(chave)) {
        fpLicHabilitarElemento(el)
        return
    }
    fpLicDesabilitarElemento(el, chave)
}

function fpLicAplicarEmCards() {
    $('.fp-menu-card').each(function () {
        fpLicAplicarNoElemento(this)
    })
}

function fpLicAplicarEmLinks() {
    $('a[href], [role="button"].fp-menu-card, button[data-fp-url]').each(function () {
        var $a = $(this)
        if ($a.hasClass('fp-produto-card')) return
        fpLicAplicarNoElemento(this)
    })
}

function fpLicAplicarEmDash() {
    $('.fp-dash-kpi[href]').each(function () {
        fpLicAplicarNoElemento(this)
    })
}

function fpAplicarLicencaUI() {
    fpLicAplicarEmCards()
    fpLicAplicarEmDash()
    fpLicAplicarEmLinks()
    fpLicInitTooltips()
}

function fpLicRetencaoDias() {
    return fpLicEstado().retencao_dias || 30
}

if (!window.__fpLicClickRegistrado) {
    document.addEventListener('click', fpLicClickBloqueado, true)
    window.__fpLicClickRegistrado = true
}

document.addEventListener('DOMContentLoaded', function () {
    var path = (window.location.pathname || '').toLowerCase().replace(/\/$/, '') || '/'
    if (path === '/login') return
    if (typeof fpLicCarregarLogado === 'function') {
        fpLicCarregarLogado(function () {
            if (!fpLicAplicarModoSuspenso() && typeof fpAplicarLicencaUI === 'function') {
                fpAplicarLicencaUI()
            }
        })
        return
    }
    if (!fpLicAplicarModoSuspenso() && typeof fpAplicarLicencaUI === 'function') {
        fpAplicarLicencaUI()
    }
})

var WL_LOGO_DATA = ''
var WL_TEMA = null

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    wlInitTema()
    carregarWhitelabel()

    $('#wl-logo-file').on('change', function () {
        var f = this.files && this.files[0]
        if (!f) return
        if (f.size > 450000) {
            CvMsg.aviso('Arquivo muito grande. Use imagem até ~400 KB.')
            this.value = ''
            return
        }
        var reader = new FileReader()
        reader.onload = function (e) {
            WL_LOGO_DATA = e.target.result
            $('#wl-logo-preview').attr('src', WL_LOGO_DATA).removeClass('d-none')
            $('#wl-logo-vazio').addClass('d-none')
        }
        reader.readAsDataURL(f)
    })

    $('#wl-salvar').on('click', function () { salvarWhitelabel(false) })
    $('#wl-reset').on('click', wlRestaurarPadrao)
})

function wlIdFranqueado() {
    var id = (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.idFranqueado)
        ? ConfVisionUrls.idFranqueado()
        : (localStorage.getItem('idFranqueado') || '')
    if (!id) {
        $('#wl-alerta').removeClass('d-none').text('Sessão sem franqueado. Faça login novamente.')
    }
    return id
}

function wlInitTema() {
    if (typeof CvTheme === 'undefined') return
    WL_TEMA = CvTheme.normalizeTheme(CvTheme.getStoredTheme() || CvTheme.DEFAULT_THEME)
    wlBuildPresets()
    wlSyncPresets(WL_TEMA.preset)
}

function wlBuildPresets() {
    var $wrap = $('#wl-theme-presets')
    if (!$wrap.length || typeof CvTheme === 'undefined') return
    $wrap.empty()
    Object.keys(CvTheme.PRESETS).forEach(function (key) {
        var p = CvTheme.PRESETS[key]
        var $btn = $('<button type="button" class="cv-wl-preset"></button>')
        $btn.attr('data-preset', key)
        $btn.attr('title', p.label)
        $btn.attr('aria-label', 'Tema ' + p.label)
        $btn.append('<span class="cv-wl-preset-swatch" style="background:linear-gradient(135deg,' + p.primary + ' 50%,' + p.bg + ' 50%)"></span>')
        $btn.append('<span class="cv-wl-preset-label">' + p.label + '</span>')
        $btn.on('click', function () {
            WL_TEMA = CvTheme.getPreset(key) || CvTheme.normalizeTheme(p)
            CvTheme.applyTheme(WL_TEMA)
            wlSyncPresets(key)
        })
        $wrap.append($btn)
    })
}

function wlSyncPresets(activePreset) {
    $('#wl-theme-presets .cv-wl-preset').each(function () {
        var key = $(this).attr('data-preset')
        $(this).toggleClass('cv-wl-preset-active', key === activePreset)
    })
}

function wlAplicarTemaCarregado(dados, idFra) {
    if (typeof CvTheme === 'undefined' || !dados) return
    var tema = CvTheme.parseTemaJson(dados.tema_json || dados.temaJson || dados.tema)
    if (tema) {
        WL_TEMA = CvTheme.saveTheme(tema, idFra)
        wlSyncPresets(WL_TEMA.preset)
        return
    }
    WL_TEMA = CvTheme.normalizeTheme(CvTheme.DEFAULT_THEME)
    wlSyncPresets(WL_TEMA.preset)
}

function carregarWhitelabel() {
    var idFra = wlIdFranqueado()
    if (!idFra) return
    $.ajax({
        url: '/cvWhitelabelCarregar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: idFra })
    }).done(function (r) {
        var dados = r.dados || r
        if (!dados) {
            $('#wl-info').removeClass('d-none').text(
                'Nenhum registro ainda. Ao salvar, criamos o cadastro compartilhado com o FranqueadoPro.'
            )
            return
        }

        wlAplicarTemaCarregado(dados, idFra)

        var infos = []
        if (dados.logo_data) {
            WL_LOGO_DATA = dados.logo_data
            $('#wl-logo-preview').attr('src', dados.logo_data).removeClass('d-none')
            $('#wl-logo-vazio').addClass('d-none')
            if (typeof CvWhitelabel !== 'undefined') CvWhitelabel.applyLogo(dados.logo_data, idFra)
        }
        if (dados.dominio) {
            infos.push('Domínio base: <strong>' + $('<div>').text(dados.dominio).html() + '</strong>')
        }
        if (dados.fqdn_cv) {
            infos.push('Central de câmeras: <code>' + $('<div>').text(dados.fqdn_cv).html() + '</code>')
        }
        if (dados.fqdn) {
            infos.push('FranqueadoPro: <code>' + $('<div>').text(dados.fqdn).html() + '</code>')
        }
        if (infos.length) {
            $('#wl-info').removeClass('d-none').html(infos.join('<br>'))
        }
    }).fail(function (xhr) {
        var msg = 'Não foi possível carregar a marca.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
            else if (j.status) msg = j.status
        } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
    })
}

function wlRestaurarPadrao() {
    CvMsg.confirmar('Restaurar padrão?', 'Logo e tema voltam ao padrão do sistema.').then(function (r) {
        if (!r.isConfirmed) return
        WL_LOGO_DATA = ''
        $('#wl-logo-preview').addClass('d-none').attr('src', '')
        $('#wl-logo-vazio').removeClass('d-none')
        $('#wl-logo-file').val('')
        if (typeof CvTheme !== 'undefined') {
            WL_TEMA = CvTheme.resetTheme(wlIdFranqueado(), true)
            wlSyncPresets(WL_TEMA.preset)
        }
        salvarWhitelabel(true)
    })
}

function salvarWhitelabel(restorePadrao) {
    var idFra = wlIdFranqueado()
    if (!idFra) return

    var temaSalvar = null
    if (typeof CvTheme !== 'undefined') {
        if (restorePadrao === true) {
            temaSalvar = CvTheme.temaParaSalvar(CvTheme.DEFAULT_THEME)
        } else {
            temaSalvar = CvTheme.temaParaSalvar(WL_TEMA || CvTheme.getStoredTheme() || CvTheme.DEFAULT_THEME)
        }
    }

    var payload = {
        id_franqueado: idFra,
        logo_data: restorePadrao === true ? null : (WL_LOGO_DATA || null),
        tema_json: temaSalvar
    }

    $('#wl-salvar').prop('disabled', true)
    $.ajax({
        url: '/cvWhitelabelSalvar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function () {
        if (typeof CvWhitelabel !== 'undefined') {
            if (payload.logo_data) CvWhitelabel.applyLogo(payload.logo_data, idFra)
            else CvWhitelabel.resetLogo(idFra, true)
        }
        if (typeof CvTheme !== 'undefined' && temaSalvar) {
            WL_TEMA = CvTheme.saveTheme(temaSalvar, idFra)
            wlSyncPresets(WL_TEMA.preset)
        }
        if (typeof CvWhitelabel !== 'undefined' && CvWhitelabel.notifyBrandSaved) {
            CvWhitelabel.notifyBrandSaved(idFra)
        }
        CvMsg.sucesso('Personalização salva com sucesso.')
        carregarWhitelabel()
    }).fail(function (xhr) {
        var msg = 'Não foi possível salvar.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
            else if (j.status) msg = j.status
        } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
    }).always(function () {
        $('#wl-salvar').prop('disabled', false)
    })
}

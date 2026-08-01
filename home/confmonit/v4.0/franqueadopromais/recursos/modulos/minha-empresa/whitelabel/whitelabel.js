var WL_LOGO_DATA = ''

$(document).ready(function () {
    if (typeof fpGetTheme === 'function') {
        preencherCores(fpGetTheme())
    }

    carregarWhitelabel()

    $('#wl-logo-file').on('change', function () {
        var f = this.files && this.files[0]
        if (!f) return
        var reader = new FileReader()
        reader.onload = function (e) {
            WL_LOGO_DATA = e.target.result
            $('#wl-logo-preview').attr('src', WL_LOGO_DATA).removeClass('d-none')
        }
        reader.readAsDataURL(f)
    })

    $('#wl-salvar').on('click', salvarWhitelabel)

    $('#wl-reset').on('click', function () {
        if (typeof fpSaveTheme === 'function') {
            fpSaveTheme(FP_DEFAULT_THEME)
        }
        preencherCores(typeof FP_DEFAULT_THEME !== 'undefined' ? FP_DEFAULT_THEME : null)
        WL_LOGO_DATA = ''
        $('#wl-logo-preview').addClass('d-none').attr('src', '')
        localStorage.removeItem('fp_whitelabel_logo')
    })

    ;['wl-primary', 'wl-navbar', 'wl-bg', 'wl-surface', 'wl-text'].forEach(function (id) {
        $('#' + id).on('input', aplicarPreviewCores)
    })
})

function wlNormalizarHex(cor) {
    if (cor == null || cor === '') return ''
    var s = String(cor).trim()
    if (/^#[0-9a-fA-F]{6}$/.test(s)) return s.toLowerCase()
    if (/^[0-9a-fA-F]{6}$/.test(s)) return ('#' + s).toLowerCase()
    if (/^#[0-9a-fA-F]{3}$/.test(s)) {
        return ('#' + s[1] + s[1] + s[2] + s[2] + s[3] + s[3]).toLowerCase()
    }
    if (/^#[0-9a-fA-F]{8}$/.test(s)) return s.slice(0, 7).toLowerCase()
    return ''
}

function wlParseTema(raw) {
    if (!raw) return null
    var tema = raw
    if (typeof tema === 'string') {
        try {
            tema = JSON.parse(tema)
        } catch (e) {
            return null
        }
    }
    if (typeof tema !== 'object' || Array.isArray(tema)) return null

    // Algumas respostas vêm aninhadas (ex.: { value: {...} })
    if (!tema.primary && tema.value && typeof tema.value === 'object') {
        tema = tema.value
    }

    var primary = wlNormalizarHex(tema.primary || tema.Primary || tema.cor_primaria)
    if (!primary) return null

    var out = {
        preset: tema.preset || 'personalizado',
        primary: primary,
        primaryHover: wlNormalizarHex(tema.primaryHover || tema.primary_hover) ||
            (typeof fpLighten === 'function' ? fpLighten(primary, -8) : primary),
        navbar: wlNormalizarHex(tema.navbar || tema.Navbar) || primary,
        bg: wlNormalizarHex(tema.bg || tema.background) || '#161a21',
        surface: wlNormalizarHex(tema.surface) || '#222630',
        text: wlNormalizarHex(tema.text || tema.texto) || '#e8eaed'
    }
    return out
}

function aplicarPreviewCores() {
    if (typeof fpSaveTheme !== 'function') return
    var primary = wlNormalizarHex($('#wl-primary').val())
    if (!primary) return
    fpSaveTheme({
        preset: 'personalizado',
        primary: primary,
        primaryHover: typeof fpLighten === 'function' ? fpLighten(primary, -8) : primary,
        navbar: wlNormalizarHex($('#wl-navbar').val()) || primary,
        bg: wlNormalizarHex($('#wl-bg').val()) || '#161a21',
        surface: wlNormalizarHex($('#wl-surface').val()) || '#222630',
        text: wlNormalizarHex($('#wl-text').val()) || '#e8eaed'
    })
}

function preencherCores(temaRaw) {
    var tema = wlParseTema(temaRaw)
    if (!tema) return
    $('#wl-primary').val(tema.primary)
    $('#wl-navbar').val(tema.navbar)
    $('#wl-bg').val(tema.bg)
    $('#wl-surface').val(tema.surface)
    $('#wl-text').val(tema.text)
}

function carregarWhitelabel() {
    $.ajax({
        url: '/whitelabelCarregar',
        method: 'POST',
        contentType: 'application/json',
        dataType: 'json',
        data: '{}'
    }).done(function (r) {
        var dados = (r && r.dados) ? r.dados : r
        var tema = wlParseTema(dados && (dados.tema_json || dados.temaJson || dados.tema))
        if (tema) {
            preencherCores(tema)
            if (typeof fpSaveTheme === 'function') fpSaveTheme(tema)
        } else if (typeof fpGetTheme === 'function') {
            preencherCores(fpGetTheme())
        }
        if (dados && dados.logo_data) {
            WL_LOGO_DATA = dados.logo_data
            localStorage.setItem('fp_whitelabel_logo', dados.logo_data)
            $('#wl-logo-preview').attr('src', dados.logo_data).removeClass('d-none')
            if (typeof fpApplyWhitelabelLogo === 'function') fpApplyWhitelabelLogo(dados.logo_data)
        }
    }).fail(function (xhr) {
        var msg = 'White Label disponível a partir do plano Pro.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
        $('#wl-salvar').prop('disabled', true)
        if (typeof fpGetTheme === 'function') preencherCores(fpGetTheme())
    })
}

function salvarWhitelabel() {
    var primary = wlNormalizarHex($('#wl-primary').val())
    if (!primary) {
        $('#wl-alerta').removeClass('d-none').text('Informe a cor primária.')
        return
    }

    var tema = {
        preset: 'personalizado',
        primary: primary,
        primaryHover: typeof fpLighten === 'function' ? fpLighten(primary, -8) : primary,
        navbar: wlNormalizarHex($('#wl-navbar').val()) || primary,
        bg: wlNormalizarHex($('#wl-bg').val()) || '#161a21',
        surface: wlNormalizarHex($('#wl-surface').val()) || '#222630',
        text: wlNormalizarHex($('#wl-text').val()) || '#e8eaed'
    }

    var payload = {
        tema_json: tema,
        logo_data: WL_LOGO_DATA || null
    }

    $('#wl-salvar').prop('disabled', true)

    $.ajax({
        url: '/whitelabelSalvar',
        method: 'POST',
        contentType: 'application/json',
        dataType: 'json',
        data: JSON.stringify(payload)
    }).done(function () {
        if (typeof fpSaveTheme === 'function') fpSaveTheme(tema)
        if (WL_LOGO_DATA) {
            localStorage.setItem('fp_whitelabel_logo', WL_LOGO_DATA)
            if (typeof fpApplyWhitelabelLogo === 'function') fpApplyWhitelabelLogo(WL_LOGO_DATA)
        }
        $('#wl-alerta').addClass('d-none').removeClass('alert-danger').addClass('alert-warning')
        if (typeof Swal !== 'undefined') {
            Swal.fire({ icon: 'success', title: 'Salvo', text: 'Personalização salva.', timer: 3000, showConfirmButton: false })
        }
    }).fail(function (xhr) {
        var msg = 'Não foi possível salvar.'
        try { msg = JSON.parse(xhr.responseText).message || msg } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
    }).always(function () {
        $('#wl-salvar').prop('disabled', false)
    })
}

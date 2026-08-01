var WL_LOGO_DATA = ''
var WL_DNS_IP = '185.130.61.4'
var WL_DOMINIO_BLOQUEADO = false

$(document).ready(function () {
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

    $('#wl-subdominio, #wl-dominio').on('input', function () {
        if (!WL_DOMINIO_BLOQUEADO) atualizarInstrucoesDns()
    })

    $('#wl-salvar').on('click', salvarWhitelabel)
    $('#wl-dominio-remover').on('click', removerDominio)
    $('#wl-retentar-ssl').on('click', retentarSSL)

    $('#wl-reset').on('click', function () {
        if (typeof fpSaveTheme === 'function') {
            fpSaveTheme(FP_DEFAULT_THEME)
        }
        WL_LOGO_DATA = ''
        $('#wl-logo-preview').addClass('d-none').attr('src', '')
        if (!WL_DOMINIO_BLOQUEADO) {
            $('#wl-subdominio, #wl-dominio').val('')
            atualizarInstrucoesDns()
        }
        localStorage.removeItem('fp_whitelabel_logo')
    })

    ;['wl-primary', 'wl-navbar', 'wl-bg', 'wl-surface', 'wl-text'].forEach(function (id) {
        $('#' + id).on('input', aplicarPreviewCores)
    })
})

function wlNormalizarDominio(val) {
    return String(val || '')
        .trim()
        .toLowerCase()
        .replace(/^https?:\/\//, '')
        .replace(/\/.*$/, '')
        .replace(/\.$/, '')
}

function wlMontarFqdn(dados) {
    if (dados && dados.fqdn) return dados.fqdn
    var sub = wlNormalizarDominio($('#wl-subdominio').val())
    var dom = wlNormalizarDominio($('#wl-dominio').val())
    if (!dom) return ''
    if (!sub || sub === '@') return dom
    return sub + '.' + dom
}

function wlTemDominioSalvo(dados) {
    if (!dados) return false
    if (dados.fqdn) return true
    return !!(dados.subdominio && dados.dominio)
}

function wlResolverFqdn(dados) {
    if (!dados) return wlMontarFqdn()
    if (dados.fqdn) return dados.fqdn
    if (dados.subdominio && dados.dominio) {
        if (dados.subdominio === '@') return dados.dominio
        return dados.subdominio + '.' + dados.dominio
    }
    return wlMontarFqdn(dados)
}

function wlStatusLabel(status) {
    var map = {
        provisionando: { cls: 'info', text: 'Provisionando servidor...' },
        pendente_dns: { cls: 'warning', text: 'Aguardando DNS — configure o registro tipo A' },
        pendente_plano: { cls: 'warning', text: 'Domínio salvo — plano Pro+ necessário para provisionar' },
        ativo: { cls: 'success', text: 'Domínio ativo com SSL' },
        erro: { cls: 'danger', text: 'Erro no provisionamento' }
    }
    return map[status] || { cls: 'secondary', text: status || '—' }
}

function aplicarBloqueioDominio(dados) {
    var temDominio = wlTemDominioSalvo(dados)
    WL_DOMINIO_BLOQUEADO = temDominio
    $('#wl-subdominio, #wl-dominio').prop('readonly', temDominio).prop('disabled', false).toggleClass('fp-wl-readonly', temDominio)
    $('#wl-dominio-remover').toggleClass('d-none', !temDominio)
    if (dados && (dados.dominio_status === 'pendente_dns' || dados.dominio_status === 'provisionando')) {
        $('#wl-retentar-ssl').removeClass('d-none')
    } else {
        $('#wl-retentar-ssl').addClass('d-none')
    }
}

function renderStatusDominio(dados) {
    var $box = $('#wl-dominio-status')
    if (!wlTemDominioSalvo(dados)) {
        $box.addClass('d-none').empty()
        return
    }
    var st = wlStatusLabel(dados.dominio_status)
    var fqdn = wlResolverFqdn(dados)
    var html = '<span class="badge text-bg-' + st.cls + '">' + st.text + '</span>'
    html += ' <span class="small text-muted ms-2">' + fqdn + '</span>'
    if (dados.dominio_erro) {
        html += '<div class="small text-danger mt-1">' + $('<div>').text(dados.dominio_erro).html() + '</div>'
    }
    $box.html(html).removeClass('d-none')
}

function atualizarInstrucoesDns(dados) {
    var sub = wlNormalizarDominio($('#wl-subdominio').val())
    var dom = wlNormalizarDominio($('#wl-dominio').val())
    var fqdn = wlMontarFqdn(dados)
    var $box = $('#wl-dns-instrucoes')

    if (!dom && !(dados && wlTemDominioSalvo(dados))) {
        $box.addClass('d-none')
        return
    }

    if (dados && wlTemDominioSalvo(dados)) {
        sub = dados.subdominio || sub
        dom = dados.dominio || dom
        fqdn = wlResolverFqdn(dados)
    }

    var host = sub && sub !== '@' ? sub : '@'
    var url = (dados && dados.dominio_status === 'ativo' ? 'https://' : 'http://') + fqdn + '/login'
    $('#wl-dns-host').text(host)
    $('#wl-dns-ip').text(WL_DNS_IP)
    $('#wl-dns-url').text(url).attr('href', url)
    $box.removeClass('d-none')
}

function aplicarPreviewCores() {
    if (typeof fpSaveTheme !== 'function') return
    var primary = $('#wl-primary').val()
    fpSaveTheme({
        preset: 'personalizado',
        primary: primary,
        primaryHover: typeof fpLighten === 'function' ? fpLighten(primary, -8) : primary,
        navbar: $('#wl-navbar').val(),
        bg: $('#wl-bg').val(),
        surface: $('#wl-surface').val(),
        text: $('#wl-text').val()
    })
}

function preencherCores(tema) {
    if (!tema) return
    if (tema.primary) $('#wl-primary').val(tema.primary)
    if (tema.navbar) $('#wl-navbar').val(tema.navbar)
    if (tema.bg) $('#wl-bg').val(tema.bg)
    if (tema.surface) $('#wl-surface').val(tema.surface)
    if (tema.text) $('#wl-text').val(tema.text)
}

function preencherDominio(dados) {
    if (!dados) return
    if (dados.subdominio) $('#wl-subdominio').val(dados.subdominio)
    if (dados.dominio) $('#wl-dominio').val(dados.dominio)
    if (!dados.fqdn && dados.subdominio && dados.dominio) {
        dados.fqdn = wlResolverFqdn(dados)
    }
    aplicarBloqueioDominio(dados)
    renderStatusDominio(dados)
    atualizarInstrucoesDns(dados)
}

function carregarWhitelabel() {
    $.ajax({
        url: '/whitelabelCarregar',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        var dados = r.dados || r
        if (dados && dados.tema_json) {
            preencherCores(dados.tema_json)
            if (typeof fpSaveTheme === 'function') fpSaveTheme(dados.tema_json)
        } else if (typeof fpGetTheme === 'function') {
            preencherCores(fpGetTheme())
        }
        if (dados && dados.logo_data) {
            WL_LOGO_DATA = dados.logo_data
            localStorage.setItem('fp_whitelabel_logo', dados.logo_data)
            $('#wl-logo-preview').attr('src', dados.logo_data).removeClass('d-none')
            if (typeof fpApplyWhitelabelLogo === 'function') fpApplyWhitelabelLogo(dados.logo_data)
        }
        preencherDominio(dados)
    }).fail(function (xhr) {
        var msg = 'White Label disponível no plano Pro+.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
        $('#wl-salvar').prop('disabled', true)
    })
}

function salvarWhitelabel() {
    var primary = $('#wl-primary').val()
    var tema = {
        preset: 'personalizado',
        primary: primary,
        primaryHover: typeof fpLighten === 'function' ? fpLighten(primary, -8) : primary,
        navbar: $('#wl-navbar').val(),
        bg: $('#wl-bg').val(),
        surface: $('#wl-surface').val(),
        text: $('#wl-text').val()
    }

    var payload = {
        tema_json: tema,
        logo_data: WL_LOGO_DATA || null
    }

    if (!WL_DOMINIO_BLOQUEADO) {
        var subdominio = wlNormalizarDominio($('#wl-subdominio').val()) || null
        var dominio = wlNormalizarDominio($('#wl-dominio').val()) || null
        if (dominio && !/^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/.test(dominio)) {
            $('#wl-alerta').removeClass('d-none alert-warning').addClass('alert-danger').text('Domínio inválido. Use o formato exemplo.com.br')
            return
        }
        if (subdominio && subdominio !== '@' && !/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/.test(subdominio)) {
            $('#wl-alerta').removeClass('d-none alert-warning').addClass('alert-danger').text('Subdomínio inválido. Use apenas letras, números e hífen.')
            return
        }
        payload.subdominio = subdominio
        payload.dominio = dominio
    }

    $('#wl-salvar').prop('disabled', true)

    $.ajax({
        url: '/whitelabelSalvar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (r) {
        var dados = (r && r.dados && typeof r.dados === 'object') ? r.dados : {}
        if (!dados || Object.keys(dados).length === 0) dados = r || {}

        if (WL_LOGO_DATA) {
            localStorage.setItem('fp_whitelabel_logo', WL_LOGO_DATA)
            if (typeof fpApplyWhitelabelLogo === 'function') fpApplyWhitelabelLogo(WL_LOGO_DATA)
        }
        if (r.dominio_status) dados.dominio_status = r.dominio_status
        if (r.dominio_erro) dados.dominio_erro = r.dominio_erro
        if (r.fqdn) dados.fqdn = r.fqdn
        if (!dados.fqdn && dados.subdominio && dados.dominio) {
            dados.fqdn = wlResolverFqdn(dados)
        }
        preencherDominio(dados)
        $('#wl-alerta').addClass('d-none').removeClass('alert-danger').addClass('alert-warning')

        var txt = 'Personalização salva.'
        if (r.provisionar && (r.fqdn || dados.fqdn)) {
            if (r.ssl_ok) {
                txt = 'Domínio ' + (r.fqdn || dados.fqdn) + ' configurado com SSL. Acesse pelo link abaixo.'
            } else if (r.dominio_status === 'pendente_dns') {
                txt = 'Servidor configurado. Configure o DNS tipo A e clique em "Verificar DNS e emitir SSL".'
            } else if (r.dominio_status === 'pendente_plano') {
                txt = dados.dominio_erro || 'Domínio salvo. Upgrade para Pro+ para provisionamento automático.'
            } else {
                txt = 'Domínio ' + (r.fqdn || dados.fqdn) + ' em provisionamento.'
            }
        }
        if (typeof Swal !== 'undefined') {
            Swal.fire({ icon: 'success', title: 'Salvo', text: txt, timer: 4500, showConfirmButton: false })
        }
    }).fail(function (xhr) {
        var msg = 'Não foi possível salvar.'
        try { msg = JSON.parse(xhr.responseText).message || msg } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
    }).always(function () {
        $('#wl-salvar').prop('disabled', false)
    })
}

function removerDominio() {
    var fqdn = wlMontarFqdn()
    var confirmar = function () {
        $('#wl-dominio-remover').prop('disabled', true)
        $.ajax({
            url: '/whitelabelDominioRemover',
            method: 'POST',
            contentType: 'application/json',
            data: '{}'
        }).done(function () {
            WL_DOMINIO_BLOQUEADO = false
            $('#wl-subdominio, #wl-dominio').val('').prop('readonly', false).removeClass('fp-wl-readonly')
            $('#wl-dominio-remover, #wl-retentar-ssl').addClass('d-none')
            $('#wl-dominio-status, #wl-dns-instrucoes').addClass('d-none')
            if (typeof Swal !== 'undefined') {
                Swal.fire({ icon: 'success', title: 'Removido', text: 'Domínio removido do sistema e do servidor.', timer: 3000, showConfirmButton: false })
            }
        }).fail(function (xhr) {
            var msg = 'Não foi possível remover o domínio.'
            try { msg = JSON.parse(xhr.responseText).message || msg } catch (e) { /* ignore */ }
            $('#wl-alerta').removeClass('d-none').text(msg)
        }).always(function () {
            $('#wl-dominio-remover').prop('disabled', false)
        })
    }

    if (typeof Swal !== 'undefined') {
        Swal.fire({
            icon: 'warning',
            title: 'Remover domínio?',
            html: 'O endereço <strong>' + fqdn + '</strong> deixará de abrir o FranqueadoPro.<br>Para usar outro domínio, cadastre novamente após remover.',
            showCancelButton: true,
            confirmButtonText: 'Sim, remover',
            cancelButtonText: 'Cancelar',
            confirmButtonColor: '#dc3545'
        }).then(function (res) {
            if (res.isConfirmed) confirmar()
        })
    } else if (window.confirm('Remover domínio ' + fqdn + '?')) {
        confirmar()
    }
}

function retentarSSL() {
    $('#wl-retentar-ssl').prop('disabled', true)
    $.ajax({
        url: '/whitelabelDominioRetentarSSL',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        var dados = {
            fqdn: wlMontarFqdn(),
            subdominio: $('#wl-subdominio').val(),
            dominio: $('#wl-dominio').val(),
            dominio_status: r.dominio_status,
            dominio_erro: r.dominio_erro
        }
        preencherDominio(dados)
        var txt = r.ssl_ok ? 'SSL emitido com sucesso!' : (r.message || 'DNS ainda não propagou. Tente novamente em alguns minutos.')
        if (typeof Swal !== 'undefined') {
            Swal.fire({ icon: r.ssl_ok ? 'success' : 'info', title: r.ssl_ok ? 'SSL ativo' : 'Aguardando DNS', text: txt, timer: 4000, showConfirmButton: false })
        }
    }).fail(function (xhr) {
        var msg = 'Não foi possível verificar o SSL.'
        try { msg = JSON.parse(xhr.responseText).message || msg } catch (e) { /* ignore */ }
        $('#wl-alerta').removeClass('d-none').text(msg)
    }).always(function () {
        $('#wl-retentar-ssl').prop('disabled', false)
    })
}

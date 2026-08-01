var DS_DNS_IP = '185.130.61.4'
var DS_DOMINIO_BLOQUEADO = false

$(document).ready(function () {
    carregarDominioSaas()

    $('#ds-subdominio, #ds-dominio').on('input', function () {
        if (!DS_DOMINIO_BLOQUEADO) atualizarInstrucoesDns()
    })

    $('#ds-salvar').on('click', salvarDominioSaas)
    $('#ds-dominio-remover').on('click', removerDominio)
    $('#ds-retentar-ssl').on('click', retentarSSL)
})

function dsNormalizarDominio(val) {
    return String(val || '')
        .trim()
        .toLowerCase()
        .replace(/^https?:\/\//, '')
        .replace(/\/.*$/, '')
        .replace(/\.$/, '')
}

function dsMontarFqdn(dados) {
    if (dados && dados.fqdn) return dados.fqdn
    var sub = dsNormalizarDominio($('#ds-subdominio').val())
    var dom = dsNormalizarDominio($('#ds-dominio').val())
    if (!dom) return ''
    if (!sub || sub === '@') return dom
    return sub + '.' + dom
}

function dsTemDominioSalvo(dados) {
    if (!dados) return false
    if (dados.fqdn) return true
    return !!dados.dominio
}

function dsResolverFqdn(dados) {
    if (!dados) return dsMontarFqdn()
    if (dados.fqdn) return dados.fqdn
    if (dados.subdominio && dados.dominio) {
        if (dados.subdominio === '@') return dados.dominio
        return dados.subdominio + '.' + dados.dominio
    }
    return dsMontarFqdn(dados)
}

function dsStatusLabel(status) {
    var map = {
        provisionando: { cls: 'info', text: 'Provisionando servidor...' },
        pendente_dns: { cls: 'warning', text: 'Aguardando DNS — configure o registro tipo A' },
        ativo: { cls: 'success', text: 'Domínio ativo com SSL' },
        erro: { cls: 'danger', text: 'Erro no provisionamento' }
    }
    return map[status] || { cls: 'secondary', text: status || '—' }
}

function dsMostrarLoading(texto) {
    $('#ds-loading-texto').text(texto || 'Processando...')
    $('#ds-loading').removeClass('d-none').attr('aria-busy', 'true')
    $('#ds-card').attr('aria-busy', 'true')
}

function dsOcultarLoading() {
    $('#ds-loading').addClass('d-none').attr('aria-busy', 'false')
    $('#ds-card').removeAttr('aria-busy')
}

function aplicarBloqueioDominio(dados) {
    var temDominio = dsTemDominioSalvo(dados)
    DS_DOMINIO_BLOQUEADO = temDominio
    $('#ds-subdominio, #ds-dominio').prop('readonly', temDominio).prop('disabled', false).toggleClass('fp-ds-readonly', temDominio)
    $('#ds-salvar').prop('disabled', temDominio).attr('title', temDominio ? 'Domínio já cadastrado. Remova para cadastrar outro.' : '')
    $('#ds-dominio-remover').toggleClass('d-none', !temDominio)
    if (dados && (dados.dominio_status === 'pendente_dns' || dados.dominio_status === 'provisionando')) {
        $('#ds-retentar-ssl').removeClass('d-none')
    } else {
        $('#ds-retentar-ssl').addClass('d-none')
    }
}

function renderStatusDominio(dados) {
    var $box = $('#ds-dominio-status')
    if (!dsTemDominioSalvo(dados)) {
        $box.addClass('d-none').empty()
        return
    }
    var st = dsStatusLabel(dados.dominio_status)
    var fqdn = dsResolverFqdn(dados)
    var html = '<span class="badge text-bg-' + st.cls + '">' + st.text + '</span>'
    html += ' <span class="small text-muted ms-2">' + fqdn + '</span>'
    if (dados.dominio_erro) {
        html += '<div class="small text-danger mt-1">' + $('<div>').text(dados.dominio_erro).html() + '</div>'
    }
    $box.html(html).removeClass('d-none')
}

function atualizarInstrucoesDns(dados) {
    var sub = dsNormalizarDominio($('#ds-subdominio').val())
    var dom = dsNormalizarDominio($('#ds-dominio').val())
    var fqdn = dsMontarFqdn(dados)
    var $box = $('#ds-dns-instrucoes')

    if (!dom && !(dados && dsTemDominioSalvo(dados))) {
        $box.addClass('d-none')
        return
    }

    if (dados && dsTemDominioSalvo(dados)) {
        sub = dados.subdominio || sub
        dom = dados.dominio || dom
        fqdn = dsResolverFqdn(dados)
    }

    var host = sub && sub !== '@' ? sub : '@'
    var url = (dados && dados.dominio_status === 'ativo' ? 'https://' : 'http://') + fqdn + '/login'
    $('#ds-dns-host').text(host)
    $('#ds-dns-ip').text(DS_DNS_IP)
    $('#ds-dns-url').text(url).attr('href', url)
    $box.removeClass('d-none')
}

function preencherDominio(dados) {
    if (!dados) return
    if (dados.subdominio) $('#ds-subdominio').val(dados.subdominio)
    if (dados.dominio) $('#ds-dominio').val(dados.dominio)
    if (!dados.fqdn && dados.subdominio && dados.dominio) {
        dados.fqdn = dsResolverFqdn(dados)
    }
    aplicarBloqueioDominio(dados)
    renderStatusDominio(dados)
    atualizarInstrucoesDns(dados)
}

function carregarDominioSaas() {
    $.ajax({
        url: '/dominioSaasCarregar',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        var dados = r.dados || r
        preencherDominio(dados)
    }).fail(function (xhr) {
        var msg = 'Domínio personalizado disponível no plano Pro+ (FranqueadoPro SaaS).'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#ds-alerta').removeClass('d-none').text(msg)
        $('#ds-salvar').prop('disabled', true)
    })
}

function salvarDominioSaas() {
    if (DS_DOMINIO_BLOQUEADO) return

    var subdominio = dsNormalizarDominio($('#ds-subdominio').val()) || null
    var dominio = dsNormalizarDominio($('#ds-dominio').val()) || null

    if (!dominio) {
        $('#ds-alerta').removeClass('d-none alert-warning').addClass('alert-danger').text('Informe o domínio.')
        return
    }
    if (!/^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/.test(dominio)) {
        $('#ds-alerta').removeClass('d-none alert-warning').addClass('alert-danger').text('Domínio inválido. Use o formato exemplo.com.br')
        return
    }
    if (subdominio && subdominio !== '@' && !/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/.test(subdominio)) {
        $('#ds-alerta').removeClass('d-none alert-warning').addClass('alert-danger').text('Subdomínio inválido. Use apenas letras, números e hífen.')
        return
    }

    var payload = { subdominio: subdominio, dominio: dominio }
    $('#ds-salvar').prop('disabled', true)
    $('#ds-alerta').addClass('d-none').removeClass('alert-danger').addClass('alert-warning')
    dsMostrarLoading('Salvando domínio e provisionando servidor...')

    $.ajax({
        url: '/dominioSaasSalvar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (r) {
        var dados = (r && r.dados && typeof r.dados === 'object') ? r.dados : {}
        if (!dados || Object.keys(dados).length === 0) dados = r || {}

        if (r.dominio_status) dados.dominio_status = r.dominio_status
        if (r.dominio_erro) dados.dominio_erro = r.dominio_erro
        if (r.fqdn) dados.fqdn = r.fqdn
        if (!dados.fqdn && dados.subdominio && dados.dominio) {
            dados.fqdn = dsResolverFqdn(dados)
        }
        preencherDominio(dados)

        var txt = 'Domínio salvo.'
        if (r.provisionar && (r.fqdn || dados.fqdn)) {
            if (r.ssl_ok) {
                txt = 'Domínio ' + (r.fqdn || dados.fqdn) + ' configurado com SSL. Acesse pelo link abaixo.'
            } else if (r.dominio_status === 'pendente_dns') {
                txt = 'Servidor configurado. Configure o DNS tipo A e clique em "Verificar DNS e emitir SSL".'
            } else {
                txt = 'Domínio ' + (r.fqdn || dados.fqdn) + ' em provisionamento.'
            }
        }
        if (typeof Swal !== 'undefined') {
            Swal.fire({ icon: 'success', title: 'Salvo', text: txt, timer: 4500, showConfirmButton: false })
        }
    }).fail(function (xhr) {
        var msg = 'Não foi possível salvar o domínio.'
        try { msg = JSON.parse(xhr.responseText).message || msg } catch (e) { /* ignore */ }
        $('#ds-alerta').removeClass('d-none').text(msg)
    }).always(function () {
        dsOcultarLoading()
        if (!DS_DOMINIO_BLOQUEADO) {
            $('#ds-salvar').prop('disabled', false)
        }
    })
}

function removerDominio() {
    var fqdn = dsMontarFqdn()
    var confirmar = function () {
        $('#ds-dominio-remover').prop('disabled', true)
        dsMostrarLoading('Removendo domínio do servidor...')
        $.ajax({
            url: '/dominioSaasRemover',
            method: 'POST',
            contentType: 'application/json',
            data: '{}'
        }).done(function (r) {
            DS_DOMINIO_BLOQUEADO = false
            $('#ds-subdominio, #ds-dominio').val('').prop('readonly', false).removeClass('fp-ds-readonly')
            $('#ds-dominio-remover, #ds-retentar-ssl').addClass('d-none')
            $('#ds-dominio-status, #ds-dns-instrucoes').addClass('d-none')
            $('#ds-salvar').prop('disabled', false)
            var msg = 'Domínio desativado no servidor.'
            if (r && r.provision_msg) msg = r.provision_msg
            if (typeof Swal !== 'undefined') {
                Swal.fire({ icon: 'success', title: 'Removido', text: msg, timer: 4000, showConfirmButton: false })
            }
        }).fail(function (xhr) {
            var msg = 'Não foi possível remover o domínio.'
            try { msg = JSON.parse(xhr.responseText).message || msg } catch (e) { /* ignore */ }
            $('#ds-alerta').removeClass('d-none').text(msg)
        }).always(function () {
            dsOcultarLoading()
            $('#ds-dominio-remover').prop('disabled', false)
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
    $('#ds-retentar-ssl').prop('disabled', true)
    $.ajax({
        url: '/dominioSaasRetentarSSL',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        var dados = {
            fqdn: dsMontarFqdn(),
            subdominio: $('#ds-subdominio').val(),
            dominio: $('#ds-dominio').val(),
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
        $('#ds-alerta').removeClass('d-none').text(msg)
    }).always(function () {
        $('#ds-retentar-ssl').prop('disabled', false)
    })
}

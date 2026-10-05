var DS_DNS_IP = '185.130.61.4'
var DS_DOMINIO_BLOQUEADO = false
var DS_BASE_READONLY = false

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    carregarDominio()

    $('#ds-subdominio, #ds-dominio').on('input', function () {
        if (!DS_DOMINIO_BLOQUEADO) atualizarInstrucoesDns()
    })
    $('#ds-salvar').on('click', salvarDominio)
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

function dsMontarFqdn() {
    var sub = dsNormalizarDominio($('#ds-subdominio').val())
    var dom = dsNormalizarDominio($('#ds-dominio').val())
    if (!dom) return ''
    if (!sub || sub === '@') return dom
    return sub + '.' + dom
}

function dsResolverFqdn(dados) {
    if (!dados) return dsMontarFqdn()
    if (dados.fqdn_cv) return dados.fqdn_cv
    if (dados.subdominio_cv && dados.dominio) {
        if (dados.subdominio_cv === '@') return dados.dominio
        return dados.subdominio_cv + '.' + dados.dominio
    }
    return dsMontarFqdn()
}

function dsStatusLabel(status) {
    var map = {
        provisionando: { cls: 'info', text: 'Provisionando...' },
        pendente_dns: { cls: 'warning', text: 'Aguardando DNS' },
        ativo: { cls: 'success', text: 'Ativo com SSL' },
        erro: { cls: 'danger', text: 'Erro' }
    }
    return map[status] || { cls: 'secondary', text: status || '—' }
}

function dsMostrarLoading(texto) {
    $('#ds-loading-texto').text(texto || 'Processando...')
    $('#ds-loading').removeClass('d-none')
}

function dsOcultarLoading() {
    $('#ds-loading').addClass('d-none')
}

function aplicarBloqueio(dados) {
    var temCv = !!(dados && dados.fqdn_cv)
    DS_DOMINIO_BLOQUEADO = temCv
    DS_BASE_READONLY = !!(dados && dados.dominio && (dados.fqdn || dados.fqdn_cv))

    $('#ds-subdominio').prop('readonly', temCv).toggleClass('cv-ds-readonly', temCv)
    $('#ds-dominio').prop('readonly', DS_BASE_READONLY).toggleClass('cv-ds-readonly', DS_BASE_READONLY)
    $('#ds-salvar').prop('disabled', temCv)
    $('#ds-dominio-remover').toggleClass('d-none', !temCv)

    var st = dados && dados.dominio_status_cv
    if (st === 'pendente_dns' || st === 'provisionando' || st === 'erro') {
        $('#ds-retentar-ssl').removeClass('d-none')
    } else {
        $('#ds-retentar-ssl').addClass('d-none')
    }
}

function renderStatus(dados) {
    var $box = $('#ds-dominio-status')
    if (!dados || !dados.fqdn_cv) {
        $box.addClass('d-none').empty()
        return
    }
    var st = dsStatusLabel(dados.dominio_status_cv)
    var html = '<span class="badge text-bg-' + st.cls + '">' + st.text + '</span>'
    html += ' <span class="small text-muted ms-2">' + $('<div>').text(dados.fqdn_cv).html() + '</span>'
    if (dados.dominio_erro_cv) {
        html += '<div class="small text-danger mt-1">' + $('<div>').text(dados.dominio_erro_cv).html() + '</div>'
    }
    $box.html(html).removeClass('d-none')
}

function atualizarInstrucoesDns(dados) {
    var fqdn = dados ? dsResolverFqdn(dados) : dsMontarFqdn()
    var sub = dsNormalizarDominio($('#ds-subdominio').val()) || (dados && dados.subdominio_cv) || ''
    if (!fqdn) {
        $('#ds-dns-instrucoes').addClass('d-none')
        return
    }
    $('#ds-dns-host').text(sub && sub !== '@' ? sub : '@')
    $('#ds-dns-ip').text(DS_DNS_IP)
    var url = 'https://' + fqdn + '/login'
    $('#ds-dns-url').attr('href', url).text(url)
    $('#ds-dns-instrucoes').removeClass('d-none')
}

function renderInfoBase(dados) {
    if (!dados) {
        $('#ds-info-base').addClass('d-none').empty()
        return
    }
    var parts = []
    if (dados.dominio && dados.fqdn) {
        parts.push(
            'Usando o domínio base do FranqueadoPro: <strong>' +
            $('<div>').text(dados.dominio).html() +
            '</strong>. Informe apenas o subdomínio da central de câmeras.'
        )
    }
    if (dados.fqdn) {
        parts.push('FranqueadoPro: <code>' + $('<div>').text(dados.fqdn).html() + '</code>')
    }
    if (parts.length) {
        $('#ds-info-base').removeClass('d-none').html(parts.join('<br>'))
    } else {
        $('#ds-info-base').addClass('d-none').empty()
    }
}

function preencherForm(dados) {
    if (!dados) return
    if (dados.subdominio_cv) $('#ds-subdominio').val(dados.subdominio_cv)
    if (dados.dominio) $('#ds-dominio').val(dados.dominio)
}

function dsIdFranqueado() {
    var id = (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.idFranqueado)
        ? ConfVisionUrls.idFranqueado()
        : (localStorage.getItem('idFranqueado') || '')
    if (!id) {
        $('#ds-alerta').removeClass('d-none').text('Sessão sem franqueado. Faça login novamente.')
    }
    return id
}

function carregarDominio() {
    var idFra = dsIdFranqueado()
    if (!idFra) {
        dsOcultarLoading()
        return
    }
    dsMostrarLoading('Carregando...')
    $.ajax({
        url: '/cvDominioCarregar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: idFra })
    }).done(function (r) {
        var dados = r.dados || null
        preencherForm(dados)
        renderInfoBase(dados)
        aplicarBloqueio(dados)
        renderStatus(dados)
        atualizarInstrucoesDns(dados)
    }).fail(function (xhr) {
        var msg = 'Erro ao carregar domínio.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#ds-alerta').removeClass('d-none').text(msg)
    }).always(dsOcultarLoading)
}

function salvarDominio() {
    var idFra = dsIdFranqueado()
    if (!idFra) return
    var sub = dsNormalizarDominio($('#ds-subdominio').val())
    var dom = dsNormalizarDominio($('#ds-dominio').val())
    if (!dom) {
        CvMsg.aviso('Informe o domínio base.')
        return
    }
    if (!sub) {
        CvMsg.aviso('Informe o subdomínio da central (ex.: vision, cameras).')
        return
    }

    dsMostrarLoading('Salvando e provisionando...')
    $.ajax({
        url: '/cvDominioSalvar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            id_franqueado: idFra,
            subdominio: sub,
            dominio: dom
        })
    }).done(function (r) {
        var dados = r.dados || {}
        if (r.fqdn) dados.fqdn_cv = r.fqdn
        if (r.dominio_status) dados.dominio_status_cv = r.dominio_status
        if (r.dominio_erro) dados.dominio_erro_cv = r.dominio_erro
        preencherForm(dados)
        renderInfoBase(dados)
        aplicarBloqueio(dados)
        renderStatus(dados)
        atualizarInstrucoesDns(dados)
        CvMsg.sucesso(r.provision_msg || 'Domínio salvo.')
    }).fail(function (xhr) {
        var msg = 'Erro ao salvar.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
            else if (j.status) msg = j.status
        } catch (e) { /* ignore */ }
        $('#ds-alerta').removeClass('d-none').text(msg)
    }).always(dsOcultarLoading)
}

function removerDominio() {
    CvMsg.confirmar('Remover domínio?', 'O logo e o domínio do FranqueadoPro são mantidos.').then(function (r) {
        if (!r.isConfirmed) return
        var idFra = dsIdFranqueado()
        if (!idFra) return
        dsMostrarLoading('Removendo...')
        $.ajax({
            url: '/cvDominioRemover',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ id_franqueado: idFra })
        }).done(function () {
            $('#ds-subdominio').val('')
            carregarDominio()
        }).fail(function (xhr) {
            var msg = 'Erro ao remover.'
            try {
                var j = JSON.parse(xhr.responseText)
                if (j.message) msg = j.message
            } catch (e) { /* ignore */ }
            CvMsg.erro(msg)
        }).always(dsOcultarLoading)
    })
}

function retentarSSL() {
    var idFra = dsIdFranqueado()
    if (!idFra) return
    dsMostrarLoading('Verificando DNS/SSL...')
    $.ajax({
        url: '/cvDominioRetentarSSL',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id_franqueado: idFra })
    }).done(function () {
        carregarDominio()
    }).fail(function (xhr) {
        var msg = 'Erro ao verificar SSL.'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        CvMsg.erro(msg)
        dsOcultarLoading()
    })
}

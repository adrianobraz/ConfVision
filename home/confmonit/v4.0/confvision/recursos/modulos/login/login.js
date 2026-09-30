
$(document).ready(function () {
    var tenantHidden = $('#cv-tenant-franqueado').val()
    if (tenantHidden && typeof CvWhitelabel !== 'undefined') {
        CvWhitelabel.setTenant(tenantHidden)
    }

    $('#btn-logar').on('click', fazerLogin)

    $('#email').keydown(function (e) {
        if (e.keyCode == 13) {
            if ($('#email').val() != '') {
                $('#senha').focus()
            }
        }
    })

    $('#senha').keydown(function (e) {
        if (e.keyCode == 13) {
            if ($('#senha').val() != '') {
                fazerLogin()
            }
        }
    })
})

function irParaMenuConfVision(idFranqueado) {
    if (typeof CvWhitelabel !== 'undefined' && CvWhitelabel.fetchAndCache) {
        CvWhitelabel.fetchAndCache(idFranqueado).always(function () {
            window.location = '/carregar-menu-confvision'
        })
        return
    }
    window.location = '/carregar-menu-confvision'
}

function fazerLogin(evento) {
    var tenantSalvo = ''
    try {
        tenantSalvo = sessionStorage.getItem('cv_tenant_franqueado') || ''
    } catch (e) { /* ignore */ }
    if (!tenantSalvo) {
        tenantSalvo = ($('#cv-tenant-franqueado').val() || '')
    }

    document.cookie.split(";").forEach(function (c) {
        document.cookie = c.replace(/^ +/, "")
            .replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
    });
    localStorage.clear();
    sessionStorage.clear();

    try {
        if (tenantSalvo) sessionStorage.setItem('cv_tenant_franqueado', tenantSalvo)
    } catch (e) { /* ignore */ }

    var tenantDominio = ''
    if (typeof CvWhitelabel !== 'undefined') {
        tenantDominio = CvWhitelabel.getTenant() || tenantSalvo || ($('#cv-tenant-franqueado').val() || '')
    } else {
        tenantDominio = tenantSalvo || ($('#cv-tenant-franqueado').val() || '')
    }

    $.ajax({
        url: "/LoginLogar",
        method: 'Post',
        data: JSON.stringify({
            email: $('#email').val(),
            senha: $('#senha').val(),
            id_franqueado_dominio: tenantDominio || undefined
        })
    }).fail(function (e) {
        console.log(e)
        const status = (e.responseJSON && e.responseJSON.status) || ''
        if (status.indexOf('modulo ConfVision nao contratado') !== -1) {
            boxMesagemAtencaoPersonalizada('Módulo de câmeras não contratado. Entre em contato com a central.')
        } else if (status.indexOf('acesso negado') !== -1) {
            boxMesagemAtencaoPersonalizada('Acesso negado.')
        } else if (status.indexOf('usuario nao autorizado neste dominio') !== -1) {
            boxMesagemAtencaoPersonalizada('Acesso negado.')
        } else if (status.indexOf('usuario bloqueado') !== -1) {
            boxMesagemAtencaoPersonalizada('Conta suspensa, favor entrar em contato com a central')
        } else if (status === 'Erro: franqueado bloqueado') {
            boxMesagemAtencaoPersonalizada('Conta suspensa, favor entrar em contato com a central')
        } else {
            boxMesagemAtencaoPersonalizada('Usuário ou senha inválidos.')
        }
    }).done(function (r) {
        const d = r.dados
        if (!d || !d.token) {
            boxMesagemAtencaoPersonalizada('Usuário ou senha inválidos.')
            return
        }

        if (d.tipo === 'CLI') {
            var idFraCli = d.idVinculo || d.idFranqueado || ''
            localStorage.setItem('papel', 'CLI')
            localStorage.setItem('idCliente', d.idCliente || '')
            localStorage.setItem('idFranqueado', idFraCli)
            localStorage.setItem('idUsuario', d.idUsuario || '')
            localStorage.setItem('nomeUsuario', d.nome || '')
            localStorage.setItem('nickUsuario', d.nick || '')
            localStorage.setItem('email', d.email1 || '')
            localStorage.setItem('token', d.token)
            localStorage.setItem('nomeFranqueado', d.nomeFranqueado || '')
            localStorage.setItem('contadorInfomrativo', 0)
            irParaMenuConfVision(idFraCli)
            return
        }

        if (d.tipo != 'FRA') {
            boxMesagemAtencaoPersonalizada('Usuário não autorizado a usar o recurso!!!')
            return
        }

        getFranqDadosById(d.idVinculo, function (fra) {
            localStorage.setItem('papel', 'FRA')
            localStorage.setItem('idFranqueado', d.idVinculo)
            localStorage.setItem('idRepresentante', fra.repId)
            localStorage.setItem('idUsuario', d.idUsuario)
            localStorage.setItem('nomeUsuario', d.nome)
            localStorage.setItem('nickUsuario', d.nick)
            localStorage.setItem('email', d.email1)
            localStorage.setItem('token', d.token)
            localStorage.setItem('nomeFranqueado', fra.fraRazao)
            localStorage.removeItem('idCliente')
            localStorage.setItem('contadorInfomrativo', 0)
            irParaMenuConfVision(d.idVinculo)
        })
    })
}

function getFranqDadosById(id, next) {
    $.ajax({
        url: "/getFranqDadosById",
        method: 'Post',
        data: JSON.stringify({ fraId: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            next(r.dados)
        } else {
            boxErro("Erro ao consultar os dados do repesentante")
        }
    })
}

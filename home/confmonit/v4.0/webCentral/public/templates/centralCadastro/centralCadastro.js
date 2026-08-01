var mapaCentrais = {}

$(window).on('load', function () {
    if (sessionStorage.getItem('loginId') !== 'BREAKGLASS'
        || sessionStorage.getItem('loginBreakglass') !== 'S') {
        sessionStorage.removeItem('loginBreakglass')
        window.location = '/home'
        return
    }
    sessionStorage.setItem('id', '0')
    listar()
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
})

function limpar() {
    sessionStorage.setItem('id', '0')
    $('#razaoSocial, #nomeFantasia, #cnpj, #inscricao, #cep, #endereco, #bairro, #cidade, #estado').val('')
    $('#telefone1, #telefone2, #userNome, #userNick, #userEmail, #userTel').val('')
    $('#blocoUserMaster').removeClass('hidden')
    $('#tituloForm').html('<i class="bi bi-building-add"></i> Nova Central')
    $('#userNome, #userNick, #userEmail, #userTel').prop('disabled', false)
}

function listar() {
    $.ajax({
        url: '/cadastro/centrais/listar',
        method: 'POST',
        data: JSON.stringify({}),
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar centrais')
    }).done(function (r) {
        const $tbody = $('.tech-col-list tbody')
        $tbody.empty()
        mapaCentrais = {}
        if (r.status === 'OK' && Array.isArray(r.dados)) {
            r.dados.forEach((c) => {
                const id = c.idCentral || ''
                mapaCentrais[id] = c
                const idShow = id === 'CENTRAL' ? 'CENTRAL (legado)' : id
                $tbody.append(`
                    <tr class="tech-row-on">
                        <td class="text-sm">${c.razaoSocial || ''}</td>
                        <td class="text-sm">${c.nomeFantasia || ''}</td>
                        <td class="text-sm">${c.cnpj || ''}</td>
                        <td class="text-sm" title="ID_Central=${id}">${idShow}</td>
                        <td title="Editar central"
                            class="btn tech-action-edit"
                            tipo="editar"
                            idCentral="${id}">
                            <i class="bi bi-pencil-square"></i>
                        </td>
                    </tr>
                `)
            })
            $('td[tipo=editar]').on('click', function () {
                editar(this.getAttribute('idCentral'))
            })
        } else if (r.status && r.status !== 'Vazio' && r.status !== 'OK') {
            boxErro(r.status)
        }
    })
}

function editar(id) {
    const c = mapaCentrais[id]
    if (!c) {
        boxErro('Central não encontrada na lista')
        return
    }
    sessionStorage.setItem('id', c.idCentral || id)
    $('#razaoSocial').val(c.razaoSocial || '')
    $('#nomeFantasia').val(c.nomeFantasia || '')
    $('#cnpj').val(c.cnpj || '')
    $('#inscricao').val(c.inscricaoEstadual || '')
    $('#cep').val(c.cep || '')
    $('#endereco').val(c.endereco || '')
    $('#bairro').val(c.bairro || '')
    $('#cidade').val(c.cidade || '')
    $('#estado').val(c.estado || '')
    $('#telefone1').val(c.telefone1 || '')
    $('#telefone2').val(c.telefone2 || '')
    $('#userNome, #userNick, #userEmail, #userTel').val('').prop('disabled', true)
    $('#blocoUserMaster').addClass('hidden')
    $('#tituloForm').html('<i class="bi bi-pencil-square"></i> Editar Central')
}

function gravar() {
    const id = sessionStorage.getItem('id') || '0'
    if (id === '0') {
        inserir()
    } else {
        alterar()
    }
}

function inserir() {
    const razao = ($('#razaoSocial').val() || '').trim()
    const email = ($('#userEmail').val() || '').trim()
    if (!razao) {
        boxErro('Informe a razão social')
        return
    }
    if (!email) {
        boxErro('Informe o e-mail do usuário master')
        return
    }
    if (!confirm('Criar nova Central e usuário master?')) {
        return
    }

    $.ajax({
        url: '/cadastro/centrais/inserir',
        method: 'POST',
        data: JSON.stringify({
            razaoSocial: razao,
            nomeFantasia: ($('#nomeFantasia').val() || '').trim(),
            cnpj: $('#cnpj').val(),
            inscricaoEstadual: $('#inscricao').val(),
            cep: $('#cep').val(),
            endereco: $('#endereco').val(),
            bairro: $('#bairro').val(),
            cidade: $('#cidade').val(),
            estado: $('#estado').val(),
            telefone1: limpaDocumento($('#telefone1').val()),
            telefone2: limpaDocumento($('#telefone2').val()),
            userMaster: {
                nome: ($('#userNome').val() || '').trim() || razao,
                nick: ($('#userNick').val() || '').trim() || 'MASTER',
                email1: email,
                telefone1: limpaDocumento($('#userTel').val()),
            },
        }),
    }).fail(function (e) {
        let msg = 'Falha ao criar central'
        try {
            const j = JSON.parse(e.responseText)
            if (j && j.status) msg = j.status
        } catch (_) {}
        boxErro(msg)
    }).done(function (r) {
        if (r.status === 'OK') {
            boxSucesso('Central criada. Master: ' + email + ' / senha: usuario123')
            limpar()
            listar()
        } else {
            boxErro(r.status || 'Não foi possível criar')
        }
    })
}

function alterar() {
    const razao = ($('#razaoSocial').val() || '').trim()
    if (!razao) {
        boxErro('Informe a razão social')
        return
    }
    if (!confirm('Salvar alterações da Central?')) {
        return
    }

    const id = sessionStorage.getItem('id')
    $.ajax({
        url: '/cadastro/centrais/alterar',
        method: 'POST',
        data: JSON.stringify({
            idCentral: id,
            idCentralUUID: id,
            razaoSocial: razao,
            nomeFantasia: ($('#nomeFantasia').val() || '').trim(),
            cnpj: $('#cnpj').val(),
            inscricaoEstadual: $('#inscricao').val(),
            cep: $('#cep').val(),
            endereco: $('#endereco').val(),
            bairro: $('#bairro').val(),
            cidade: $('#cidade').val(),
            estado: $('#estado').val(),
            telefone1: limpaDocumento($('#telefone1').val()),
            telefone2: limpaDocumento($('#telefone2').val()),
        }),
    }).fail(function (e) {
        let msg = 'Falha ao alterar central'
        try {
            const j = JSON.parse(e.responseText)
            if (j && j.status) msg = j.status
        } catch (_) {}
        boxErro(msg)
    }).done(function (r) {
        if (r.status === 'OK') {
            boxSucesso('Central alterada com sucesso')
            limpar()
            listar()
        } else {
            boxErro(r.status || 'Não foi possível alterar')
        }
    })
}

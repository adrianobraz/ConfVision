var mapaCentrais = {}

function isBreakglass() {
    return sessionStorage.getItem('loginBreakglass') === 'S'
        && sessionStorage.getItem('loginId') === 'BREAKGLASS'
}

function garantirModoUsuario() {
    if (sessionStorage.getItem('loginId') !== 'BREAKGLASS') {
        sessionStorage.removeItem('loginBreakglass')
        $('.bg-only').addClass('hidden')
    }
}

function uuidCentralForm() {
    if (isBreakglass()) {
        return ($('#formCentral').val() || '').trim()
    }
    return sessionStorage.getItem('loginIdCentralUUID') || ''
}

function nomeCentral(uuid) {
    if (!uuid) return '-'
    return mapaCentrais[uuid] || uuid
}

$(window).on('load', function () {
    sessionStorage.setItem("id", "0")
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
    garantirModoUsuario()

    if (isBreakglass()) {
        $('.bg-only').removeClass('hidden').show()
        carregarCentrais(function () {
            listarUsuario()
        })
        $('#filtroCentral').on('change', listarUsuario)
    } else {
        $('.bg-only').addClass('hidden').hide()
        listarUsuario()
    }
})

function carregarCentrais(done) {
    $.ajax({
        url: '/cadastro/centrais/listar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({}),
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao carregar centrais')
    }).done(function (r) {
        mapaCentrais = {}
        const $filtro = $('#filtroCentral').empty()
        const $form = $('#formCentral').empty()
        $filtro.append('<option value="TODAS">Todas</option>')
        $form.append('<option value="">Selecione a Central</option>')

        if (r.status === 'OK' && Array.isArray(r.dados)) {
            r.dados.forEach(function (c) {
                // Vinculo: central.ID_Central <-> usuarios.IDCentralUUID
                const idCentral = (c.idCentral || '').trim()
                if (!idCentral) return
                const label = c.nomeFantasia || c.razaoSocial || idCentral
                mapaCentrais[idCentral] = label
                const idHex = (c.idCentralUUID || '').trim()
                if (idHex && idHex !== idCentral) {
                    mapaCentrais[idHex] = label
                }
                $filtro.append($('<option></option>').val(idCentral).text(label))
                $form.append($('<option></option>').val(idCentral).text(label))
            })
        } else if (r.status && r.status !== 'Vazio' && r.status !== 'OK') {
            boxErro(r.status)
        }
        $filtro.val('TODAS')
        if (typeof done === 'function') done()
    })
}

function listarUsuario() {
    const payload = (typeof payloadFiltroCentral === 'function')
        ? payloadFiltroCentral({ idVinculo: 'CENTRAL' })
        : payloadTenant({ idVinculo: 'CENTRAL' }, (function () {
            if (!isBreakglass()) return {}
            const v = ($('#filtroCentral').val() || '').trim()
            if (!v || v === 'TODAS') return { todas: true }
            return { uuid: v }
        })())
    if (!payload) return

    $.ajax({
        url: `/central/usuario/listar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar usuarios')
    }).done(function (r) {
        $('tbody').empty()
        if (r.status == "OK" && Array.isArray(r.dados)) {
            r.dados.map(i => {
                // Tela normal: so operadores (nao-master).
                // Break-glass: mostra todos, inclusive master da central.
                if (!isBreakglass() && i.master != 'N') {
                    return
                }

                    const stateTer = (i.usuarioTerminal == "S") ? 'tech-state-on' : 'tech-state-off'
                    const stateCen = (i.usuarioWeb == "S") ? 'tech-state-on' : 'tech-state-off'
                    const stateAtivo = (i.ativo == "S") ? 'tech-state-on' : 'tech-state-off'
                    const stateEmail = (i.enviarEmail == 'S') ? 'tech-state-on' : 'tech-state-off'
                    const rowState = (i.ativo == "S") ? 'tech-row-on' : 'tech-row-off'

                    const iconAtivar = (i.ativo == "N")
                        ? '<i class="bi bi-toggle-off"></i>'
                        : '<i class="bi bi-toggle-on"></i>'
                    const iconEmail = (i.enviarEmail == 'N')
                        ? '<i class="bi bi-envelope-slash"></i>'
                        : '<i class="bi bi-envelope-check-fill"></i>'

                    const nomeShow = i.master == 'S'
                        ? `${i.nome || ''} <small>(Master)</small>`
                        : (i.nome || '')

                    const colCentral = isBreakglass()
                        ? `<td class="text-sm">${nomeCentral(i.idCentralUUID)}</td>`
                        : ''

                    $('tbody').append(` 
                        <tr class="${rowState}">
                            ${colCentral}
                            <td title="${i.ativo == 'S' ? 'Ativo' : 'Desativado'}">${nomeShow}</td>
                            <td>${formatarCelular(i.telefone1)}</td>
                            <td class="btn tech-action-edit" 
                                tipo="editar" idUser="${i.idUsuario}"
                                title="Editar usuario">
                                <i class="bi bi-pencil-square"></i>
                            </td>
                            <td class="btn ${stateTer}" 
                                tipo="ativarTerminal" idUser="${i.idUsuario}" 
                                title="Terminal: ${i.usuarioTerminal == 'S' ? 'ATIVO' : 'DESATIVADO'}">
                                <i class="bi bi-headset"></i>
                            </td>
                            <td class="btn ${stateCen}" 
                                tipo="ativarCentral" 
                                idUser="${i.idUsuario}" 
                                title="Acesso Central: ${i.usuarioWeb == 'S' ? 'ATIVO' : 'DESATIVADO'}">
                                <i class="bi bi-house"></i>
                            </td>
                            <td class="btn tech-action-key"
                                tipo="resetSenha" 
                                idUser="${i.idUsuario}"
                                title="Resetar senha">
                                <i class="bi bi-key-fill"></i>
                            </td>
                            <td class="btn ${stateEmail}"
                                tipo="ativaEmail" 
                                idUser="${i.idUsuario}"
                                title="Email: ${i.enviarEmail == 'S' ? 'ATIVO' : 'DESATIVADO'}">
                                ${iconEmail}
                            </td>
                            <td class="btn tech-action-del" 
                                tipo="deletar" idUser="${i.idUsuario}" nome="${i.nome}"
                                title="Deletar usuario">
                                <i class="bi bi-eraser-fill"></i>
                            </td>
                            <td class="btn ${stateAtivo}" 
                                tipo="habilitar" idUser="${i.idUsuario}"
                                title="Usuario: ${i.ativo == 'S' ? 'ATIVO' : 'DESATIVADO'}">
                                ${iconAtivar}
                            </td>
                        </tr>
                    `)
            })

            $('td[tipo=editar]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                buscar(iduser)
            })

            $('td[tipo=ativarTerminal]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                ativarTerminal(iduser)
            })

            $('td[tipo=ativarCentral]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                ativarCentral(iduser)
            })

            $('td[tipo=resetSenha]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                resetarSenha(iduser)
            })

            $('td[tipo=ativaEmail]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                habilitarEmail(iduser)
            })

            $('td[tipo=deletar]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                const nome = this.getAttribute("nome")
                deletar(iduser, nome)
            })

            $('td[tipo=habilitar]').on('click', function () {
                const iduser = this.getAttribute("idUser")
                habilitar(iduser)
            })
        } else if (r.status && r.status !== 'Vazio' && r.status !== 'OK') {
            boxErro(r.status)
        }
    })
}

function buscar(id) {

    $.ajax({
        url: `/central/usuario/buscar`,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const d = r.dados

        sessionStorage.setItem("id", id)

        if (isBreakglass() && d.idCentralUUID) {
            $('#formCentral').val(d.idCentralUUID)
        }

        $('#nome').val(d.nome)
        $('#nick').val(d.nick)
        $('#usuarioVoip').val(d.usuarioVoip)
        $('#email1').val(d.email1).attr('readonly', true)
        $('#email2').val(d.email2)
        $('#telefone1').val(formatarCelular(d.telefone1))
        $('#telefone2').val(formatarCelular(d.telefone2))
    })
}

function ativarTerminal(id) {
    $.ajax({
        url: `/central/usuario/ativarTerminal`,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listarUsuario()
        } else {
            boxErro(r.status)
        }
    })
}

function ativarCentral(id) {

    $.ajax({
        url: `/central/usuario/ativarCentral`,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            listarUsuario()
        } else {
            boxErro(r.status)
        }
    })
}

function resetarSenha(id) {
    $.ajax({
        url: `/central/usuario/resetarSenha`,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            boxSenhaAlterada('usuario123')
        } else {
            boxErro(r.status)
        }
    })
}

function deletar(id, nome) {
    boxConfirmarExcluir(nome, () => {
        $.ajax({
            url: `/central/usuario/deletar`,
            method: 'POST',
            data: JSON.stringify({ idUsuario: id })
        }).fail(function (e) {
            boxErro(e)
        }).done(function (r) {
            console.log(r)
            if (r.status == "OK") {
                boxDeletadoSucesso()
                limpar()
                listarUsuario()
            } else {
                boxErro(r.status)
            }
        })
    })
}

function habilitar(id) {
    $.ajax({
        url: `/central/usuario/habilitar`,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listarUsuario()
        } else {
            boxErro(r.status)
        }
    })
}

function habilitarEmail(id) {

    $.ajax({
        url: `/central/usuario/habilitarEmail`,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listarUsuario()
        } else {
            boxErro(r.status)
        }
    })
}

function gravar() {
    if (isBreakglass() && !uuidCentralForm()) {
        boxErro('Selecione a Central')
        return
    }
    if (!$('#nome').val() || !$('#nome').val().trim()) {
        boxErro('Informe o nome')
        return
    }
    if (!$('#email1').val() || !$('#email1').val().trim()) {
        boxErro('Informe o e-mail')
        return
    }
    (sessionStorage.getItem("id") == "0") ? inserir() : alterar()
}

function inserir() {
    $.ajax({
        url: `/central/usuario/inserir`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            idVinculo: "CENTRAL",
            idCentralUUID: uuidCentralForm(),
            tipo: "CEN",
            nome: $('#nome').val(),
            nick: $('#nick').val(),
            usuarioVoip: $('#usuarioVoip').val(),
            email1: $('#email1').val(),
            email2: $('#email2').val(),
            telefone1: limpaDocumento($('#telefone1').val()),
            telefone2: limpaDocumento($('#telefone2').val()),
            usuarioTerminal: "N",
            master: "N",
        })
    }).fail(function (e) {
        let msg = 'Falha ao inserir usuario'
        try {
            if (e.responseJSON && e.responseJSON.status) msg = e.responseJSON.status
        } catch (_) {}
        if (msg.indexOf('email já em uso') >= 0) {
            boxAdvertenciaCampoAuto('Email já em uso por outro usuário', '#email1')
        } else {
            boxErro(msg)
        }
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            boxInseridoSucesso(r.dados)
            limpar()
            listarUsuario()
        } else {
            boxErro(r.status)
        }
    })
}

function alterar() {
    const payload = {
        idUsuario: sessionStorage.getItem("id"),
        nome: $('#nome').val(),
        nick: $('#nick').val(),
        usuarioVoip: $('#usuarioVoip').val(),
        email1: $('#email1').val(),
        email2: $('#email2').val(),
        telefone1: limpaDocumento($('#telefone1').val()),
        telefone2: limpaDocumento($('#telefone2').val()),
    }
    if (isBreakglass()) {
        payload.idCentralUUID = uuidCentralForm()
    }
    $.ajax({
        url: `/central/usuario/alterar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        boxErro('erro ao alterar o usuario')
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            boxAteradoSucesso()
            limpar()
            listarUsuario()
        } else {
            boxErro(r.status)
        }
    })
}

function limpar() {
    sessionStorage.setItem("id", "0")

    $('#nome').val('')
    $('#nick').val('')
    $('#usuarioVoip').val('')
    $('#email1').val('').attr('readonly', false)
    $('#email2').val('')
    $('#telefone1').val('')
    $('#telefone2').val('')
    $('#terminal').val('0')
}

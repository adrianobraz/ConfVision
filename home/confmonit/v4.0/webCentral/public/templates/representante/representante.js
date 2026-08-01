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
    sessionStorage.setItem("idUser", "0")

    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
    garantirModoUsuario()

    if (isBreakglass()) {
        $('.bg-only').removeClass('hidden')
        carregarCentrais(function () {
            listar()
            carregarPacotes()
        })
        $('#filtroCentral').on('change', listar)
        $('#formCentral').on('change', function () {
            carregarPacotes()
        })
    } else {
        $('.bg-only').addClass('hidden')
        listar()
        carregarPacotes()
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
                // Vinculo: central.ID_Central <-> representante.IDCentralUUID
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
        }
        // Filtro fica em "Todas"; formulario nao herda sessao do admin
        $filtro.val('TODAS')
        if (typeof done === 'function') done()
    })
}

function listar() {
    // Break-glass: UUID vem do combo #filtroCentral (nunca da sessao do Administrator)
    const payload = (typeof payloadFiltroCentral === 'function')
        ? payloadFiltroCentral({})
        : payloadTenant({}, (function () {
            if (!isBreakglass()) return {}
            const v = ($('#filtroCentral').val() || '').trim()
            if (!v || v === 'TODAS') return { todas: true }
            return { uuid: v }
        })())
    if (!payload) return

    $.ajax({
        url: `/representante/listar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload),
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        $('tbody').empty()
        if (r.status == "OK") {
            
            r.dados.map(i => {
                const stateAtivo = (i.ativo == "S") ? 'tech-state-on' : 'tech-state-off'
                const stateEmail = (i.emailEnvio == 'S') ? 'tech-state-on' : 'tech-state-off'
                const stateSms = (i.smsEnvio == 'S') ? 'tech-state-on' : 'tech-state-off'
                const rowState = (i.ativo == "S") ? 'tech-row-on' : 'tech-row-off'

                const iconAtivar = (i.ativo == "N")
                    ? '<i class="bi bi-toggle-off"></i>'
                    : '<i class="bi bi-toggle-on"></i>'
                const iconEmail = (i.emailEnvio == 'N')
                    ? '<i class="bi bi-envelope-slash"></i>'
                    : '<i class="bi bi-envelope-check-fill"></i>'
                const iconSms = (i.smsEnvio == 'N')
                    ? '<i class="bi bi-phone-vibrate"></i>'
                    : '<i class="bi bi-phone-vibrate-fill"></i>'

                const colCentral = isBreakglass()
                    ? `<td class="text-sm">${nomeCentral(i.idCentralUUID)}</td>`
                    : ''

                if (i.dataCancelamento == '') {
                    $('tbody').append(` 
                        <tr class="${rowState}">
                            ${colCentral}
                            <td class="text-sm" title="${i.ativo == 'S' ? 'Ativo' : 'Desativado'}">${i.razaoSocial}</td>
                            <td title="Editar representante" 
                                class="btn tech-action-edit" 
                                tipo="editar" 
                                idRep="${i.idRepresentante}">
                                <i class="bi bi-pencil-square"></i>
                            </td>
                            <td title="Resetar senha do master"
                                class="btn tech-action-key"
                                tipo="resetarSenha"
                                idUser="${i.idUsuarioMaster}">
                                <i class="bi bi-key-fill"></i>
                            </td>
                            <td title="Email: ${i.emailEnvio == 'S' ? 'ATIVO' : 'DESATIVADO'}" 
                                class="btn ${stateEmail}"
                                tipo="ativaEmail"
                                idRep="${i.idRepresentante}">
                                ${iconEmail}
                            </td>
                            <td title="SMS: ${i.smsEnvio == 'S' ? 'ATIVO' : 'DESATIVADO'}" 
                                class="btn ${stateSms}"
                                tipo="ativaSms"
                                idRep="${i.idRepresentante}">
                                ${iconSms}
                            </td>
                            <td class="btn ${stateAtivo}" 
                                tipo="habilitar" idRep="${i.idRepresentante}"
                                title="Representante: ${i.ativo == 'S' ? 'ATIVO' : 'DESATIVADO'}">
                                ${iconAtivar}
                            </td>
                        </tr>
                    `)
                } else {
                    $('tbody').append(` 
                        <tr>
                            ${colCentral}
                            <td class="text-sm">${i.razaoSocial}</td>
                            <td class="text-sm" colspan="5">CANCELADO</td>
                        </tr>
                    `)
                }
            })

            $('td[tipo=editar]').on('click', function () {
                const idRep = this.getAttribute("idRep")
                buscar(idRep)
            })
            $('td[tipo=resetarSenha]').on('click', function () {
                const idUser = this.getAttribute("idUser")
                resetarSenha(idUser)
            })

            $('td[tipo=ativaEmail]').on('click', function () {
                const idRep = this.getAttribute("idRep")
                habilitarEmail(idRep)
            })

            $('td[tipo=ativaSms]').on('click', function () {
                const idRep = this.getAttribute("idRep")
                habilitarSms(idRep)
            })

            $('td[tipo=habilitar]').on('click', function () {
                const idRep = this.getAttribute("idRep")
                habilitar(idRep)
            })

        } else {
            if (r.status != "Vazio") {
                boxErro(r.status)
            }
        }

    })
}

function carregarPacotes() {
    $('#pacote').empty().append('<option value="">Selecione</option>')
    const uuid = uuidCentralForm()
    if (isBreakglass() && !uuid) {
        return
    }
    const payload = payloadTenant({ idVinculo: 'CENTRAL' }, uuid ? { uuid: uuid } : {})
    if (!payload) return

    $.ajax({
        url: `/representante/pacoteListar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status != "Vazio" && Array.isArray(r.dados)) {
            r.dados.filter(p => p.ativo == "S").map(item => {
                $('#pacote').append(`<option value="${item.idPacote}">${item.nome}</option>`)
            })
        }
    })
}

function buscar(id) {

    $.ajax({
        url: `/representante/buscar`,
        method: 'POST',
        data: JSON.stringify({ idRepresentante: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (ret) {
        const d = ret.dados
        const u = ret.dados.userMaster

        sessionStorage.setItem("id", d.idRepresentante)
        sessionStorage.setItem("idUser", d.idUsuarioMaster)

        if (isBreakglass() && d.idCentralUUID) {
            $('#formCentral').val(d.idCentralUUID)
            carregarPacotes()
        }

        $('#razaoSocial').val(d.razaoSocial)
        $('#nomeFantasia').val(d.nomeFantasia)
        $('#cnpj').val(d.cnpj)
        $('#inscricao').val(d.inscricaoEstadual)
        $('#cep').val(d.cep)
        $('#uf').val(d.uf)
        $('#endereco').val(d.endereco)
        $('#bairro').val(d.bairro)
        $('#cidade').val(d.cidade)
        setTimeout(function () {
            $('#pacote').val(d.idPacote)
        }, 300)

        $('#nome').val(u.nome)
        $('#telefone1').val(u.telefone1)
        $('#nick').val(u.nick)
        $('#telefone2').val(u.telefone2)
        $('#email1').val(u.email1)
        $('#email2').val(u.email2)

    })
}

function resetarSenha(id) {
    $.ajax({
        url: `/representante/resetarSenha`,
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


function habilitarEmail(id) {

    $.ajax({
        url: `/representante/habilitarEmail`,
        method: 'POST',
        data: JSON.stringify({ idRepresentante: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            listar()
        } else {
            boxErro(r.status)
        }
    })
}

function habilitarSms(id) {

    $.ajax({
        url: `/representante/habilitarSms`,
        method: 'POST',
        data: JSON.stringify({ idRepresentante: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            listar()
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
    (sessionStorage.getItem("id") == "0") ? inserir() : alterar()
}

function inserir() {
    $.ajax({
        url: `/representante/inserir`,
        method: 'POST',
        data: JSON.stringify({
            idCentralUUID: uuidCentralForm(),
            cnpj: $('#cnpj').val(),
            inscricaoEstadual: $('#inscricao').val(),
            razaoSocial: $('#razaoSocial').val(),
            nomeFantasia: $('#nomeFantasia').val(),
            endereco: $('#endereco').val(),
            bairro: $('#bairro').val(),
            cidade: $('#cidade').val(),
            uf: $('#uf').val(),
            cep: $('#cep').val(),
            idPacote: $('#pacote').val(),
            userMaster: {
                idUsuario: sessionStorage.getItem("idUser"),
                nome: $('#nome').val(),
                telefone1: $('#telefone1').val(),
                nick: $('#nick').val(),
                telefone2: $('#telefone2').val(),
                email1: $('#email1').val(),
                email2: $('#email2').val(),
            }
        })
    }).fail(function (e) {
        boxErro("Erro ao inserir o representante")
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            boxInseridoSucesso(r.dados)
            listar()
            limpar()
        } else {
            boxErro(r.status)
        }
    })
}

function alterar() {
    const payload = {
        idRepresentante: sessionStorage.getItem('id'),
        idUsuario: sessionStorage.getItem("idUser"),
        cnpj: $('#cnpj').val(),
        inscricaoEstadual: $('#inscricao').val(),
        razaoSocial: $('#razaoSocial').val(),
        nomeFantasia: $('#nomeFantasia').val(),
        endereco: $('#endereco').val(),
        bairro: $('#bairro').val(),
        cidade: $('#cidade').val(),
        uf: $('#uf').val(),
        cep: $('#cep').val(),
        idPacote: $('#pacote').val(),
        userMaster: {
            idUsuario: sessionStorage.getItem("idUser"),
            nome: $('#nome').val(),
            telefone1: $('#telefone1').val(),
            nick: $('#nick').val(),
            telefone2: $('#telefone2').val(),
            email1: $('#email1').val(),
            email2: $('#email2').val(),
        }
    }
    if (isBreakglass()) {
        payload.idCentralUUID = uuidCentralForm()
    }
    $.ajax({
        url: `/representante/alterar`,
        method: 'POST',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        boxErro("Erro ao alterar o representatne")
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            boxAteradoSucesso()
            limpar()
            listar()
        } else {
            boxErro(r.status)
        }
    })
}

function habilitar(id) {
    $.ajax({
        url: `/representante/habilitar`,
        method: 'POST',
        data: JSON.stringify({ idRepresentante: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listar()
        } else {
            boxErro(r.status)
        }
    })
}


function limpar() {
    sessionStorage.setItem("id", "0")
    sessionStorage.setItem("idUser", "0")
    $('#razaoSocial').val('')
    $('#nomeFantasia').val('')
    $('#cnpj').val('')
    $('#inscricao').val('')
    $('#cep').val('')
    $('#uf').val('')
    $('#endereco').val('')
    $('#bairro').val('')
    $('#cidade').val('')
    $('#pacote').val('')

    $('#nome').val('')
    $('#telefone1').val('')
    $('#nick').val('')
    $('#telefone2').val('')
    $('#email1').val('')
    $('#email2').val('')
}

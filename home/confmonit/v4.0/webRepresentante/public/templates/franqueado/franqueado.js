$(window).on('load', function () {


    sessionStorage.setItem("id", "0")
    listarFranqueado()
    carregarPacotes()

    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
})

function listarFranqueado() {
    const idRepresentante = sessionStorage.getItem("loginRepId")

    $.ajax({
        url: `/franqueado/listar`,
        method: 'POST',
        data: JSON.stringify({ repId: idRepresentante })
    }).fail(function (e) {
        //boxErro(e.responseJSON.status)
        boxErro("Erro ao listar os franqueados")
        console.log(e)
    }).done(function (r) {

        $('tbody').empty()
        if (r.status == "OK") {
            r.dados.map(i => {

                const btnDeletar = (sessionStorage.getItem("loginMaster") == "S") ? `
                    <td 
                        title="Cancela o franqueado" 
                        class="btn bg-red-700 text-white" 
                        onClick="cancelar('${i.fraId}', '${i.fraRazao}')">
                        <i class="bi bi-eraser-fill"></i>
                    </td>
                `: ''

                const iconAtivo = (i.fraAtivo == "N") ?
                    '<i class="bi bi-toggle-off text-white"></i>' :
                    '<i class="bi bi-toggle-on text-white"></i>'
                const corAtivo = (i.fraAtivo == "N") ?
                    'bg-zinc-700' : 'bg-green-700'

                const iconEmail = (i.fraEnviaEmail == 'N') ?
                    '<i class="bi bi-envelope-slash"></i>' :
                    '<i class="bi bi-envelope-check-fill"></i>'
                const corEmail = (i.fraEnviaEmail == 'N') ?
                    'bg-zinc-700' : 'bg-cyan-700'

                const iconSms = (i.fraEnviaSms == 'N') ?
                    '<i class="bi bi-phone-vibrate"></i>' :
                    '<i class="bi bi-phone-vibrate-fill"></i>'
                const corSms = (i.fraEnviaSms == 'N') ?
                    'bg-zinc-700' : 'bg-lime-700'

                if (i.fraDataCancelamento == '') {
                    $('tbody').append(` 
                        <tr>
                            <td class="text-sm">${i.fraRazao}</td>
                          
                            <td title="Busca dados do franqueado para edição" class="btn bg-blue-700 text-white" 
                                onclick="buscar('${i.fraId}')">
                                <i class="bi bi-pencil-square"></i>
                            </td>
                          
                            <td title="Reseta a senha do usuario master do franqueado" class="btn bg-yellow-700 text-white"
                                onClick="resetarSenha('${i.userId}')">
                                <i class="bi bi-key-fill"></i>
                            </td>
            
                            <td title="Habilita/Desabilita o envio de email para o franqueado" class="btn ${corEmail} text-white"
                                tipo="habilitarEmail" idFra="${i.fraId}">
                                ${iconEmail}
                            </td>
    
                             <td title="Habilita/Desabilita o envio de sms para o franqueado" 
                                class="btn ${corSms} text-white"
                                tipo="habilitarSms" 
                                idFra="${i.fraId}">
                                ${iconSms}
                            </td>
    
                            <td title="Habilita/Desabilita o franqueado" 
                                class="btn ${corAtivo} text-white"
                                tipo="habilitar" 
                                idFra="${i.fraId}">
                                ${iconAtivo}
                            </td>
                          
                            ${btnDeletar}              
                           
                        </tr>
                    `)
                } else {
                    $('tbody').append(`
                        <tr>
                            <td class="text-sm">${i.fraRazao}</td>
                            <td class="text-sm">CANCELADO</td>
                        </tr>

                    `)
                }
            })

            // Associa os botoes          

            $('td[tipo=habilitarEmail]').on('click', function () {
                const idFra = this.getAttribute("idFra")
                habilitarEmail(idFra)
            })

            $('td[tipo=habilitarSms]').on('click', function () {
                const idFra = this.getAttribute("idFra")
                habilitarSms(idFra)
            })

            $('td[tipo=habilitar]').on('click', function () {
                const idFra = this.getAttribute("idFra")
                habilitar(idFra)
            })

        } else {
            if (r.status != "Vazio") boxErro(r.status)
        }

    })
}

function buscar(id) {
    $.ajax({
        url: `/franqueado/buscar`,
        method: 'POST',
        data: JSON.stringify({ fraId: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (ret) {
        console.log(ret)
        const d = ret.dados
        const r = ret.dados.userDados

        sessionStorage.setItem("id", d.fraId)
        sessionStorage.setItem("idUser", d.userId)

        $('#pacote').val(d.fraIdPacote)
        $('#razaoSocial').val(d.fraRazao)
        $('#nomeFantasia').val(d.fraNome)
        $('#cnpj').val(d.fraCnpj)
        $('#inscricao').val(d.fraInscricaoEstadual)
        $('#cep').val(d.fraCep)
        $('#endereco').val(d.fraEndereco)
        $('#complemento').val(d.fraComplemento)
        $('#bairro').val(d.fraBairro)
        $('#cidade').val(d.fraCidade)
        $('#uf').val(d.fraUf)
        $('#benuvem').val(d.fraCodBenuvem)
        $('#observacao').val(d.fraObservacao)
        //userMaster
        $('#nome').val(r.nome)
        $('#nick').val(r.nick)
        $('#email1').val(r.email1)
        $('#email2').val(r.email2)
        $('#telefone1').val(formatarCelular(r.telefone1))
        $('#telefone2').val(formatarCelular(r.telefone2))


        $('#email1').attr('readonly', true)
        $('#razaoSocial').focus()

    })
}

function resetarSenha(id) {
    $.ajax({
        url: `/franqueado/resetarSenha`,
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

function habilitar(id) {
    $.ajax({
        url: `/franqueado/habilitar`,
        method: 'POST',
        data: JSON.stringify({ fraId: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listarFranqueado()
        } else {
            boxErro(r.status)
        }
    })
}

function habilitarEmail(id) {
    $.ajax({
        url: `/franqueado/habilitarEmail`,
        method: 'POST',
        data: JSON.stringify({ fraId: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listarFranqueado()
        } else {
            boxErro(r.status)
        }
    })
}

function habilitarSms(id) {
    $.ajax({
        url: `/franqueado/habilitarSms`,
        method: 'POST',
        data: JSON.stringify({ fraId: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            listarFranqueado()
        } else {
            boxErro(r.status)
        }
    })
}

function gravar() {

    if (verificarCampos()) {
        return
    }

    (sessionStorage.getItem("id") == "0") ? inserir() : alterar()
}

function inserir() {

    $.ajax({
        url: `/franqueado/inserir`,
        method: 'POST',
        data: JSON.stringify({
            repId: sessionStorage.getItem("loginRepId"),
            fraIdPacote: $('#pacote').val(),
            fraRazao: $('#razaoSocial').val(),
            fraNome: $('#nomeFantasia').val(),
            fraCnpj: $('#cnpj').val(),
            fraInscricaoEstadual: $('#inscricao').val(),
            fraCep: $('#cep').val(),
            fraEndereco: $('#endereco').val(),
            fraComplemento: $('#complemento').val(),
            fraBairro: $('#bairro').val(),
            fraCidade: $('#cidade').val(),
            fraUf: $('#uf').val(),
            fraCodBenuvem: $('#benuvem').val(),
            fraObservacao: $('#observacao').val(),

            userDados: {
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                email1: $('#email1').val().trim(),
                email2: $('#email2').val().trim(),
                telefone1: limpaDocumento($('#telefone1').val()),
                telefone2: limpaDocumento($('#telefone2').val()),
            }
        })
    }).fail(function (e) {
        console.log(e)
        if (e.responseJSON.status == 'Erro: email já em uso por outro usuário') {
            boxAdvertenciaCampoAuto(
                'O email digitado já esta em uso',
                '#email1'
            )
        }else {
            boxErro('Erro ao cadastrar o franqueado')
        }
    }).done(function (r) {
        if (r.status == "OK") {
            boxInseridoSucesso(r.dados.idFranqueado)
            limpar()
            listarFranqueado()
        } else {
            boxErro(r.status)
        }
    })
}

function alterar() {
    $.ajax({
        url: `/franqueado/alterar`,
        method: 'POST',
        data: JSON.stringify({
            fraId: sessionStorage.getItem("id"),
            fraIdPacote: $('#pacote').val(),
            fraRazao: $('#razaoSocial').val(),
            fraNome: $('#nomeFantasia').val(),
            fraCnpj: $('#cnpj').val(),
            fraInscricaoEstadual: $('#inscricao').val(),
            fraCep: $('#cep').val(),
            fraEndereco: $('#endereco').val(),
            fraComplemento: $('#complemento').val(),
            fraBairro: $('#bairro').val(),
            fraCidade: $('#cidade').val(),
            fraUf: $('#uf').val(),
            fraCodBenuvem: $('#benuvem').val(),
            fraObservacao: $('#observacao').val(),

            userDados: {
                idUsuario: sessionStorage.getItem("idUser"),
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                email1: $('#email1').val(),
                email2: $('#email2').val(),
                telefone1: limpaDocumento($('#telefone1').val()),
                telefone2: limpaDocumento($('#telefone2').val()),
            }
        })
    }).fail(function (e) {
        boxErro("Erro ao alterar o franqueado")
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            boxAteradoSucesso()
            limpar()
            listarFranqueado()
        } else {
            boxErro(r.status)
        }
    })
}

function limpar() {
    sessionStorage.setItem("id", "0")
    sessionStorage.setItem("idUser", 0)

    // limpa os campos
    $('#pacote').val('')
    $('#razaoSocial').val('')
    $('#nomeFantasia').val('')
    $('#cnpj').val('')
    $('#inscricao').val('')
    $('#cep').val('')
    $('#endereco').val('')
    $('#complemento').val('')
    $('#bairro').val('')
    $('#cidade').val('')
    $('#uf').val('')
    $('#benuvem').val('')
    $('#observacao').val('')
    //userMaster
    $('#nome').val('')
    $('#nick').val('')
    $('#email1').val('')
    $('#email2').val('')
    $('#telefone1').val('')
    $('#telefone2').val('')

    $('#email1').attr('readonly', false)
}

function cancelar(id, nome) {

    boxConfirmarCancelar(nome, () => {
        $.ajax({
            url: `/franqueado/cancelar`,
            method: 'POST',
            data: JSON.stringify({
                fraId: id,
            })
        }).fail(function (e) {
            boxErro("Erro ao cancelar o franqueado")
            console.log(e)
        }).done(function (r) {
            if (r.status == "OK") {
                boxSucesso("Colocado em processo de cancelamento com sucesso")
                limpar()
                listarFranqueado()
            } else {
                boxErro(r.status)
            }
        })
    })
}

function carregarPacotes() {
    const payload = payloadTenant({
        idVinculo: sessionStorage.getItem("loginRepId")
    })

    $.ajax({
        url: `/franqueado/carregarPacotes`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
        $('#pacote').empty().append('<option value="">CADASTRE UM PACOTE</option>')
    }).done(function (r) {
        $('#pacote').empty()
        if (r.status != 'Vazio' && Array.isArray(r.dados)) {
            r.dados.forEach(i => {
                $('#pacote').append(`
                     <option value="${i.idPacote}">${i.nome}</option>
                `)
            })
        } else {
            $('#pacote').append(`
                <option value="">CADASTRE UM PACOTE</option>
           `)
        }
    })
}

function verificarCampos() {
    if ($('#razaoSocial').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Razão Social não pode ficar em branco',
            '#razaoSocial'
        )
        return true
    }

    if ($('#cnpj').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo CNPJ não pode ficar em branco',
            '#cnpj'
        )
        return true
    }

    if ($('#nomeFantasia').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Nome Fantasia não pode ficar em branco',
            '#nomeFantasia'
        )
        return true
    }

    if ($('#pacote').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Pacote tem que ser selecionado',
            '#pacote'
        )
        return true
    }

    if ($('#nome').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Nome não pode ficar em branco',
            '#nome'
        )
        return true
    }

    if ($('#nick').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Nick não pode ficar em branco',
            '#nick'
        )
        return true
    }

    if ($('#email1').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Email Principal não pode ficar em branco',
            '#email1'
        )
        return true
    }

    if ($('#telefone1').val() == "") {
        boxAdvertenciaCampoAuto(
            'O Campo Telefone Principal não pode ficar em branco',
            '#telefone1'
        )
        return true
    }

    return false

}

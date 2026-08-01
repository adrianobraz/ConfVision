$(window).on('load', function () {

    sessionStorage.setItem("id", "0")
    listarOs()
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
    $('#btnFechar').on('click', fechar)

    carregarFranqueado()
})

function listarOs() {
    $.ajax({
        url: `/franqueado/os/listar`,
        method: 'POST',
        data: JSON.stringify({ idMaster: sessionStorage.getItem("loginRepId") })
    }).fail(function (e) {
        boxErro("Erro ao listar os tickets")
    }).done(function (r) {
        $('tbody').empty()

        if (r.status != 'Vazio') {
            r.dados.map(item => {
                let btnCor
                if (item.status == 'NOVO') {
                    btnCor = 'bg-red-700'
                } else if (item.status == 'FECHADO') {
                    btnCor = 'bg-zinc-700'
                } else if (item.status == 'AGUARDANDO RESPOSTA') {
                    btnCor = 'bg-green-700'
                } else if (item.status == 'RESPONDIDO') {
                    btnCor = 'bg-yellow-700'
                }

                // Botao apagar que so sera mostrado para o usuario master =========
                const botao = `
                <td 
                    tipo="btnDeletar" 
                    idTicket="${item.idTicket}"  
                    assunto="${item.assunto}" 
                    class="btn bg-purple-700 text-white">
                    <i class="bi bi-eraser-fill"></i>
                </td>`

                btnExcluir = (sessionStorage.getItem("loginMaster") == 'N') ? '' : botao
                //==================================================================

                $('tbody').append(`
                    <tr>
                        <td class="text-sm">${item.nome}</td>
                        <td class="text-sm">${item.assunto}</td>
                        <td class="text-sm">${item.status}</td>

                         <td  
                            tipo="btnAtender" 
                            idTicket="${item.idTicket}"
                            status="${item.status}"
                            assunto="${item.assunto}"
                            historico="${item.descricao}"
                            class="btn ${btnCor} text-white">
                            <i class="bi bi-headset"></i>
                        </td>
                        ${btnExcluir}
                    </tr>                        
                `)
            })

            $('td[tipo=btnAtender]').on('click', function () {
                const idTicket = $(this).attr('idTicket')
                const status = $(this).attr('status')
                const assunto = $(this).attr('assunto')
                const historico = $(this).attr('historico')
                if (status == 'FECHADO') {
                    reabrir(idTicket, assunto, historico)
                } else {
                    buscar(idTicket)
                }
            })

            $('td[tipo=btnDeletar]').on('click', function () {
                const idTicket = $(this).attr('idTicket')
                const assunto = $(this).attr('assunto')

                deletar(idTicket, assunto)
            })
        }
    })
}

function reabrir(idTicket, assunto, historico) {

    Swal.fire({
        //position: 'top',
        title: `Quer Realmente reabrir o ticket: ${assunto}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#15803d',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode reabrir!'
    }).then((result) => {
        if (result.isConfirmed) {
            const descricao = geraTag(historico, 'Reabriu o ticket')
            $.ajax({
                url: `/franqueado/os/atualizar`,
                method: 'POST',
                data: JSON.stringify({
                    idTicket: idTicket,
                    status: "RESPONDIDO",
                    assunto: assunto,
                    descricao: descricao,
                })
            }).fail(function (e) {
                boxErro("Erro ao reabrir o ticket")
            }).done(function (r) {
                if (r.status == 'OK') {
                    boxSucesso('Ticket reaberto com sucesso')
                    limpar()
                    buscar(idTicket)
                    listarOs()
                } else {
                    boxErro(r.status)
                }
            })
        }
    })
}

function buscar(idTicket) {
    $.ajax({
        url: `/franqueado/os/buscar`,
        method: 'POST',
        data: JSON.stringify({ idTicket: idTicket })
    }).fail(function (e) {
        boxErro("Erro ao buscar os dados do ticket")
    }).done(function (r) {
        if (r.status == "OK") {
            const d = r.dados
            sessionStorage.setItem("id", d.idTicket)
            $('#franqueado').val(d.idSlave)
            $('#assunto').val(d.assunto)
            $('#historico').val(d.descricao)

        } else {
            boxErro(r.status)
        }
    })
}

function limpar() {
    sessionStorage.setItem("id", "0")
    $('#franqueado').val('0')
    $('#assunto').val('')
    $('#historico').val('')
    $('#descricao').val('')
}

function gravar() {
    if ($('#franqueado').val() == '0'){
        boxAdvertenciaCampoAuto(
            'Selecione um franqueado primeiro',
            '#franqueado'
        )
        return
    }

    if ($('#assunto').val() == ''){
        boxAdvertenciaCampoAuto(
            'Informe um assunto primeiro',
            '#assunto'
        )
        return
    }

    if ($('#descricao').val() == ""){
        boxAdvertenciaCampoAuto(
            'Informe uma descricao primeiro',
            '#descricao'
        )
        return
    }

    if (sessionStorage.getItem("id") != '0') {
        atualizar()
    } else {
        inserir()
    }
}

function fechar() {
    if (sessionStorage.getItem("id") != '0') {
        if ($('#descricao').val() == '') {
            boxAdvertenciaCampoAuto('Informe uma descrição', '#descricao')
            return
        }

        const descricao = geraTag($('#historico').val(), $('#descricao').val())
        $.ajax({
            url: `/franqueado/os/atualizar`,
            method: 'POST',
            data: JSON.stringify({
                idTicket: sessionStorage.getItem("id"),
                status: "FECHADO",
                assunto: $('#assunto').val(),
                descricao: descricao,
            })
        }).fail(function (e) {
            boxErro("Erro ao fechar o ticket")
        }).done(function (r) {
            if (r.status == 'OK') {
                boxSucesso('Ticket fechado com sucesso')
                limpar()
                listarOs()
            } else {
                boxErro(r.status)
            }
        })
    } else {
        boxErro('Um ticket deve ser selecionado primeiro')
    }
}

function atualizar() {
    if (sessionStorage.getItem("id") != '0') {
        const descricao = geraTag($('#historico').val(), $('#descricao').val())
        $.ajax({
            url: `/franqueado/os/atualizar`,
            method: 'POST',
            data: JSON.stringify({
                idTicket: sessionStorage.getItem("id"),
                status: "RESPONDIDO",
                assunto: $('#assunto').val(),
                descricao: descricao,
            })
        }).fail(function (e) {
            boxSucesso("Erro ao atualizar o ticket")
        }).done(function (r) {
            if (r.status == 'OK') {
                boxSucesso('Ticket atualizado com sucesso')
                limpar()
                listarOs()
            } else {
                boxErro(r.status)
            }
        })
    }
}

function inserir() {
    const descricao = geraTag('', $('#descricao').val())
    $.ajax({
        url: `/franqueado/os/inserir`,
        method: 'POST',
        data: JSON.stringify({
            idMaster: sessionStorage.getItem("loginRepId"),
            idSlave: $('#franqueado').val(),
            tipoSlave: "FRA",
            assunto: $('#assunto').val(),
            descricao: descricao,
        })
    }).fail(function (e) {
        boxErro("Erro ao inserir o ticket")
    }).done(function (r) {
        if (r.status == 'OK') {
            boxSucesso('Ticket inserido com sucesso')
            limpar()
            listarOs()
        } else {
            boxErro(r.status)
        }

    })
}

function geraTag(historico, descricao) {
    const usuario = sessionStorage.getItem('loginNome')
    const data = new Date()
    const dia = String(data.getDate()).padStart(2, '0')
    const mes = String(data.getMonth() + 1).padStart(2, '0')
    const ano = data.getFullYear();
    const hora = String(data.getHours()).padStart(2, '0')
    const minuto = String(data.getMinutes()).padStart(2, '0')
    const segundo = String(data.getSeconds()).padStart(2, '0')
    const final = dia + '/' + mes + '/' + ano + ' - ' + hora + ':' + minuto + ':' + segundo
    const tag = `[${final} - ${usuario}]`
    return `${historico}\n\n${tag}\n${descricao}`
}

function deletar(id, assunto) {
    boxConfirmarExcluir(assunto, () => {
        $.ajax({
            url: `/franqueado/os/deletar`,
            method: 'POST',
            data: JSON.stringify({ idTicket: id })
        }).fail(function (e) {
            boxErro("Erro ao deletar o ticket")
        }).done(function (r) {
            if (r.status == 'OK') {
                boxSucesso('Ticket removido com sucesso')
                limpar()
                listarOs()
            } else {
                boxErro(r.status)
            }
        })
    })
}

function carregarFranqueado() {
    $.ajax({
        url: `/franqueado/os/franqueadoListar`,
        method: 'POST',
        data: JSON.stringify({ repId: sessionStorage.getItem("loginRepId") })
    }).fail(function (e) {
        boxErro("Erro ao carregar os franqueados")
        console.log(e)
    }).done(function (r) {
        $('#franqueado').empty().append(
            `<option value="0">SELECIONE FRANQUEADO</option>`
        )

        $('#franqueado').append(r.dados.map(
            item => `
                <option value="${item.fraId}">
                ${item.fraRazao}
                </option>
            `
        ))
    })
}
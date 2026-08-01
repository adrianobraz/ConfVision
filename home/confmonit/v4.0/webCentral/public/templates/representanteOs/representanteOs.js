$(window).on('load', function () {

    sessionStorage.setItem("id", "0")
    listarOs()
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
    $('#btnFechar').on('click', fechar)

    carregarRepresentante()
})

//ok
function carregarRepresentante() {
    $.ajax({
        url: `/representante/os/carregarRepresentante`,
        method: 'POST',
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        
        $('#representante').empty().append(
            `<option value="0">SELECIONE REPRESENTANTE</option>`
        )

        $('#representante').append(r.dados.map(
            item => `
                <option value="${item.idRepresentante}">
                ${item.razaoSocial}
                </option>
            `
        ))
    })
}

//ok
function listarOs() {
    $.ajax({
        url: `/representante/os/listar`,
        method: 'POST',
        data: JSON.stringify({ idMaster: "1" })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()

        if (r.status != 'Vazio') {
            let novo, fechado, agurandoResposta, respondido

            r.dados.map(item => {

                let corAtendimento
                if (item.status == 'NOVO') {
                    novo = novo + linha(item, 'bg-red-700')
                } else if (item.status == 'FECHADO') {
                    fechado = fechado + linha(item, 'bg-zinc-700')
                } else if (item.status == 'AGUARDANDO RESPOSTA') {
                    agurandoResposta = agurandoResposta + linha(item, 'bg-green-700')
                } else if (item.status == 'RESPONDIDO') {
                    respondido = respondido + linha(item, 'bg-yellow-700')
                }
            })
            $('tbody').append(novo)
            $('tbody').append(agurandoResposta)
            $('tbody').append(respondido)
            $('tbody').append(fechado)


            function linha(item, btnCor) {
                let btnExcluir
                
                if (sessionStorage.getItem("loginMaster") == 'N') {
                    btnExcluir = '' 
                }else {
                    btnExcluir = `
                        <td  
                            tipo="btnDeletar" 
                            idTicket="${item.idTicket}" 
                            assunto="${item.assunto}"
                            class="btn bg-purple-700 text-white">
                                <i class="bi bi-eraser-fill"></i>
                        </td>
                    `
                }
                               
                return `
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
                `
            }
           

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

            
            $('td[tipo=btnDeletar]').on('click', function () {
                const idTicket = $(this).attr('idTicket')
                const assunto = $(this).attr('assunto')
                deletar(idTicket, assunto)
            })

        }
    })
}

//ok
function reabrir(idTicket, assunto, historico) {
    const descricao = geraTag(historico, 'Reabriu o ticket')
    $.ajax({
        url: `/representante/os/atualizar`,
        method: 'POST',
        data: JSON.stringify({
            idTicket: idTicket,
            status: "RESPONDIDO",
            assunto: assunto,
            descricao: descricao,
        })
    }).fail(function (e) {
        console.log(e)
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

//ok
function buscar(idTicket) {
    $.ajax({
        url: `/representante/os/buscar`,
        method: 'POST',
        data: JSON.stringify({ idTicket: idTicket })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == "OK") {
            const d = r.dados
            sessionStorage.setItem("id", d.idTicket)
            $('#representante').val(d.idSlave)
            $('#assunto').val(d.assunto)
            $('#historico').val(d.descricao)

        } else {
            console.log(r.status)
        }
    })
}

//ok
function limpar() {
    sessionStorage.setItem("id", "0")
    $('#representante').val('0')
    $('#assunto').val('')
    $('#historico').val('')
    $('#descricao').val('')
}

//ok
function gravar() {
    if (sessionStorage.getItem("id") != '0') {
        atualizar()
    } else {
        inserir()
    }
}

//ok
function fechar() {
    if (sessionStorage.getItem("id") != '0') {
        if ($('#descricao').val() == '') {
            boxAdvertenciaCampoAuto('Informe uma descrição', '#descricao')
            return
        }

        const descricao = geraTag($('#historico').val(), $('#descricao').val())
        $.ajax({
            url: `/representante/os/atualizar`,
            method: 'POST',
            data: JSON.stringify({
                idTicket: sessionStorage.getItem("id"),
                status: "FECHADO",
                assunto: $('#assunto').val(),
                descricao: descricao,
            })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == 'OK') {
                boxSucesso('Ticket fechado com sucesso')
                limpar()
                listarOs()
            } else {
                boxErro(r.status)
            }
        })
    }else {
        boxErro("Selecione um ticket primeiro")
    }
}

//ok
function atualizar() {
    if (sessionStorage.getItem("id") != '0') {
        const descricao = geraTag($('#historico').val(), $('#descricao').val())
        $.ajax({
            url: `/representante/os/atualizar`,
            method: 'POST',
            data: JSON.stringify({
                idTicket: sessionStorage.getItem("id"),
                status: "RESPONDIDO",
                assunto: $('#assunto').val(),
                descricao: descricao,
            })
        }).fail(function (e) {
            console.log(e)
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

//ok
function inserir() {
    const descricao = geraTag('', $('#descricao').val())
    $.ajax({
        url: `/representante/os/inserir`,
        method: 'POST',
        data: JSON.stringify({
            idMaster: '1',
            idSlave: $('#representante').val(),
            tipoSlave: "REP",
            assunto: $('#assunto').val(),
            descricao: descricao,
            status: "NOVO",
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == 'OK') {
            boxSucesso('Ticket inserido com sucesso')
            limpar()
            listarOs()
        } else {
            boxErro(r.status)
        }

    })
}

//ok
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
    if (historico == ''){
        return `${tag}\n${descricao}`
    }else{
        return `${historico}\n\n${tag}\n${descricao}`
    }
}

//
function deletar(id, assunto) {
    boxConfirmarExcluir(assunto, () => {
        $.ajax({
            url: `/representante/os/deletar`,
            method: 'POST',
            data: JSON.stringify({ idTicket: id })
        }).fail(function (e) {
            console.log(e)
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
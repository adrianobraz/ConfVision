$(document).ready(function () {
    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-gravar').on('click', gravarDados)
    $('#btn-fechar-reabrir').on('click', fechar)

    localStorage.setItem('idTicket', '')

    carregarTabela()
})

function limparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    localStorage.setItem('idTicket', '')
    $("#assunto-ticket").prop("disabled", false)
    $("#descricao-ticket").prop("disabled", false)
    $("#btn-fechar-reabrir").text("Fechar")
}

function gravarDados() {
    const id = localStorage.getItem('idTicket')

    if (id == "") {
        inserir()
    } else {
        atualizar(id)
    }
}

function carregarTabela() {
    
    $.ajax({
        url: 'TicketsFranqueadoCarregarTodos',
        method: 'Post',
        data: JSON.stringify({ idSlave: localStorage.getItem("idFranqueado") })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        $('tbody').empty()

        if (r.status != 'Vazio') {
            let novo, respodido, aguardando, iniciado, fechado
            
            r.dados.forEach(i => {
                if (i.status == 'NOVO'){
                    novo += montaLinha(i)
                }else if (i.status == 'RESPONDIDO'){
                    respodido += montaLinha(i)
                }else if (i.status == 'AGUARDANDO RESPOSTA'){
                    aguardando += montaLinha(i)
                }else if (i.status == 'INICIADO'){
                    iniciado += montaLinha(i)
                }else if (i.status == 'FECHADO'){
                    fechado += montaLinha(i)    
                }
            });
            $('tbody').append(novo)
            $('tbody').append(respodido)
            $('tbody').append(aguardando)
            $('tbody').append(iniciado)
            $('tbody').append(fechado)

            associaBotoesTabela()
        }
    })
}

function montaLinha(i) {
    console.log(i)
    let cor
    if (i.status == 'NOVO') {
        cor = 'bg-warning'
    } else if (i.status == 'RESPONDIDO') {
        cor = 'bg-danger'
    } else if (i.status == 'AGUARDANDO RESPOSTA') {
        cor = 'bg-success'
    } else if (i.status == 'INICIADO') {
        cor = 'bg-warning'
    } else if (i.status == 'FECHADO') {
        cor = 'bg-secondary'
    }

    return `    
        <tr>
        <td class="text-uppercase">${i.assunto}</td>
        <td class="text-uppercase">${i.status}</td>
        <td class="d-grid gap-2">
            <button class="btn btn-sm ${cor} py-0"
                id=${i.idTicket}
                tipo="interagir"
                status="${i.status}"
                title="Interage com o ticket">
                <i class="bi bi-headset"></i>
            </button>
        </td>
      
       
    </tr>
    `
}

function associaBotoesTabela() {
    $('button[tipo=interagir]').on('click', function () {
        const id = this.getAttribute("id")
        const status = this.getAttribute("status")
        buscar(id, status)
    })

}

function geraTag() {
    const usuario = $('#nome-usuario').val()
    const data = new Date()
    const dia = String(data.getDate()).padStart(2, '0')
    const mes = String(data.getMonth() + 1).padStart(2, '0')
    const ano = data.getFullYear();
    const hora = String(data.getHours()).padStart(2, '0')
    const minuto = String(data.getMinutes()).padStart(2, '0')
    const segundo = String(data.getSeconds()).padStart(2, '0')
    const final = dia + '/' + mes + '/' + ano + ' - ' + hora + ':' + minuto + ':' + segundo
    return `[${final} - ${usuario}]`
}

function buscar(id, status) {
    if (status == 'FECHADO') {
        boxAdvertenciaAuto('Antes de interagir com ticket, deve reabrir o mesmo')
        $("#assunto-ticket").prop("disabled", true)
        $("#descricao-ticket").prop("disabled", true)
        $("#btn-fechar-reabrir").text("Reabrir")

    } else {
        $("#assunto-ticket").prop("disabled", false)
        $("#descricao-ticket").prop("disabled", false)
        $("#btn-fechar-reabrir").text("Fechar")
    }

    $.ajax({
        start: boxProcessando(),
        url: 'TicketsFranqueadoBuscar',
        method: 'Post',
        data: JSON.stringify({idTicket: id})
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao buscar dados do Base")
    }).done(function (r) {
        boxFechar()
        const d = r.dados

        localStorage.setItem('idTicket', d.idTicket)
        $('#assunto-ticket').val(d.assunto)
        $('#historico-ticket').val(d.descricao)
    })
}


function atualizar(id) {
    if (validarCamposBranco()) return
    const historico = $('#historico-ticket').val()
    const tag = geraTag()
    const messagem = $('#descricao-ticket').val()

    const descricao = `${historico}\n\n${tag}\n${messagem}`

    $.ajax({
        start: boxProcessando(),
        url: 'TicketsFranqueadoAtualizar',
        method: 'Post',
        data: JSON.stringify(
            {
                idTicket: id,
                assunto: $('#assunto-ticket').val(),
                descricao: descricao,
                status: "AGUARDANDO RESPOSTA"
            }
        )
    }).fail(function (e) {
        boxErro("Erro ao alterar dados do Base")
    }).done(function (r) {
        boxAteradoSucesso()
        limparFormulario()
        carregarTabela()
    })

}

function fechar() {
    const id = localStorage.getItem('idTicket')
    
    if (id == "") {
        boxAdvertenciaAuto('Primeiro deve selecionaru um ticket')
        return
    }

    if ($("#btn-fechar-reabrir").text() == "Reabrir") {
        Swal.fire({
            //position: 'top',
            title: `Quer Realmente reabrir este ticket?`,
            text: "Você podera fechar ele no botão Fechar",
            icon: 'warning',
            showCancelButton: true,
            confirmButtonColor: '#3085d6',
            cancelButtonColor: '#d33',
            confirmButtonText: 'Sim, pode reabrir!'
        }).then((result) => {
            if (result.isConfirmed) {
                
                $.ajax({
                    start: boxProcessando(),
                    url: 'TicketsFranqueadoAtualizar',
                    method: 'Post',
                    data: JSON.stringify(
                        {
                            idTicket: id,
                            assunto: $('#assunto-ticket').val().toUpperCase(),
                            descricao: $('#historico-ticket').val(),
                            status: "AGUARDANDO RESPOSTA"
                        }
                    )
                }).fail(function (e) {
                    boxErro("Erro ao reabrir o ticket")
                }).done(function (r) {

                    boxAteradoSucesso()
                    limparFormulario()
                    carregarTabela()
                })
            }
        })
    } else {

        if (validarCamposBranco()) return

        const historico = $('#historico-ticket').val()
        const tag = geraTag()
        const messagem = $('#descricao-ticket').val()

        const descricao = `${historico}\n\n${tag}\n${messagem}`
        Swal.fire({
            //position: 'top',
            title: `Quer Realmente fechar este ticket?`,
            text: "Você podera reabrir ele no botão Reabrir",
            icon: 'warning',
            showCancelButton: true,
            confirmButtonColor: '#3085d6',
            cancelButtonColor: '#d33',
            confirmButtonText: 'Sim, pode fechar!'
        }).then((result) => {
            if (result.isConfirmed) {
                
                $.ajax({
                    start: boxProcessando(),
                    url: 'TicketsFranqueadoAtualizar',
                    method: 'Post',
                    data: JSON.stringify(
                        {
                            idTicket: id,
                            assunto: $('#assunto-ticket').val().toUpperCase(),
                            descricao: descricao,
                            status: "FECHADO"
                        }
                    )
                }).fail(function (e) {
                    boxErro("Erro ao fechar o ticket")
                }).done(function (r) {
                    boxAteradoSucesso()
                    limparFormulario()
                    carregarTabela()
                })
            }
        })
    }
}


function inserir() {

    if (validarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: 'TicketsFranqueadoInserir',
        method: 'POST',
        data: JSON.stringify(
            {
                idMaster: localStorage.getItem('idRepresentante'),
                idSlave: localStorage.getItem('idFranqueado'),
                assunto: $('#assunto-ticket').val().toUpperCase(),
                descricao: $('#descricao-ticket').val().toUpperCase(),
                status: "NOVO"
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o Ticket")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        limparFormulario()
        carregarTabela()
    })
}

function validarCamposBranco() {
    if ($('#assunto-ticket').val() == "") {
        boxAdvertenciaAuto("O campo Assunto , não poder ficar em branco")
        $('#assunto-ticket').focus()
        return true
    }
    if ($('#descricao-ticket').val() == "") {
        boxAdvertenciaAuto("O campo Descrição, não poder ficar em branco")
        $('#descricao-ticket').focus()
        return true
    }
}

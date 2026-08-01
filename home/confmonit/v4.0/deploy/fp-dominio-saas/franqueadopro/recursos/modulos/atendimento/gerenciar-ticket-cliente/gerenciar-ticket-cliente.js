$(document).ready(function () {
    $('#btn-limpar').on('click', ticketClienteLimparFormulario)
    $('#btn-gravar').on('click', ticketClienteGravarDados)
    $('#btn-fechar-reabrir').on('click', ticketClienteFechar)
    $('#btnImprimir').on('click', ticketClienteImprimirOS)

    localStorage.setItem('idTicket', '')

    ticketClienteClienteListar()
    ticketClienteTicketListar()
})

function ticketClienteClienteListar() {

    $.ajax({
        url: '/ticketClienteClienteListar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        // console.log(e)
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        console.log(r)
        $('#id-cliente').empty().append(
            '<option value="0">SELECIONE</option>'
        )

        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                $('#id-cliente').append(
                    `<option 
                        value="${i.idCliente}"
                        endereco="${i.endereco}"
                        bairro="${i.bairro}"
                        complemento="${i.complemento}"
                        cidade="${i.cidade}"
                        telefone1="${i.telefone1}"

                    >${i.nome}</option>`
                )
            });
        }
    })
}

function ticketClienteTicketListar() {

    $.ajax({
        url: '/ticketClienteTicketListar',
        method: 'Post',
        data: JSON.stringify({ idMaster: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        // console.log(e)
        boxErro("Erro ao carregar os tickets")
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()
        if (r.status != 'Vazio') {
            r.dados.forEach(i => { ticketClienteMontaLinha(i) });

            $('button[tipo=interagir]').on('click', function () {
                const id = this.getAttribute("id")
                const status = this.getAttribute("status")
                ticketClienteBuscar(id, status)
            })
            // $('button[tipo=imprimir]').on('click', function () {
            //     const id = this.getAttribute("id")
            //     ticketClienteImprimirOS(id)
            // })
        }
    })
}

function ticketClienteMontaLinha(i) {
    //console.log(i)
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

    $('tbody').append(`
        <tr>
            <td class="text-uppercase">${i.nome}</td>
            <td class="text-uppercase">${i.assunto}</td>
            <td class="text-uppercase">${i.status}</td>
            <td>
                <button class="btn btn-sm ${cor} py-0"
                    id=${i.idTicket}
                    tipo="interagir"
                    status="${i.status}"
                    title="Interage com o ticket">
                    <i class="bi bi-headset"></i>
                </button>
            </td>        
        </tr>
    `)
    // <td>
    //         <button class="btn btn-sm btn-primary py-0"
    //             id=${i.idTicket}
    //             tipo="imprimir"
    //             title="imprime o ticket">
    //             <i class="bi bi-printer-fill"></i>
    //         </button>
    //     </td>
}

function ticketClienteLimparFormulario() {
    localStorage.setItem('idTicket', '')

    $('#formulario').each(function () {
        this.reset();
    })

    $("#assunto-ticket").prop("disabled", false)
    $("#descricao-ticket").prop("disabled", false)
    $("#btn-fechar-reabrir").text("Fechar")
}

function ticketClienteGravarDados() {
    const id = localStorage.getItem('idTicket')

    if (id == "") {
        ticketClienteInserir()
    } else {
        ticketClienteAtualizar()
    }
}

function ticketClienteGeraTag() {
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

function ticketClienteBuscar(id, status) {
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
        url: '/ticketClienteBuscar',
        method: 'Post',
        data: JSON.stringify({ idTicket: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Base")
    }).done(function (r) {
        console.log(r)
        boxFechar()
        const d = r.dados

        localStorage.setItem('idTicket', d.idTicket)
        $('#id-cliente').val(d.idSlave)
        $('#assunto-ticket').val(d.assunto)
        $('#historico-ticket').val(d.descricao)

        //scroll normal
        //$('html,body').scrollTop(0);

        //scroll suave
        $('html, body').animate({ scrollTop: 0 }, 'medium'); //slow, medium, fast

    })
}

function ticketClienteImprimirOS() {
        const id = localStorage.getItem('idTicket')
        const endereco = $('#id-cliente').find(':selected').attr('endereco')
        const bairro = $('#id-cliente').find(':selected').attr('bairro')
        const complemento = $('#id-cliente').find(':selected').attr('complemento')
        const cidade = $('#id-cliente').find(':selected').attr('cidade')
        const telefone1 = $('#id-cliente').find(':selected').attr('telefone1')
        
    $.ajax({
        url: '/ticketClienteBuscar',
        method: 'Post',
        data: JSON.stringify({ idTicket: id })
    }).fail(function (e) {
        boxErro("Erro ao carregar os dados do ticket")
    }).done(function (r) {
       
        const d = r.dados
                
        
        const celular = ('celularCliente' in d) ?
            formatarCelular(d.celularCliente) : ''
        // importa biblioteca
        var jsPDF = window.jspdf.jsPDF
        // instancia a biblioteca
        var doc = new jsPDF({
            orientation: 'portrait',
            unit: 'pt',
        })
        // Header
        doc.setTextColor(40)
        doc.setLineWidth(1)
        doc.line(40, 20, 555, 20)
        doc.line(40, 20, 40, 820) // vertical line
        doc.line(555, 20, 555, 820) // vertical line
        doc.line(40, 820, 555, 820)

        //Logo
        const img = `/recursos/img/franqueado/logoId${r.dados.idFranqueado}.png`

        //doc.addImage(img, 'PNG', 45, 22, 180, 55)
        doc.line(230, 20, 230, 80) // vertical line
        //Titulo
        doc.setFont(undefined, 'bold')
        doc.setFontSize(10)
        doc.text(`Numero OS: ${d.idTicket}`, 392, 35, { maxWidth: 500, align: 'center' })
        doc.setFontSize(30)
        doc.text('Ordem de Serviços', 392, 70, { maxWidth: 500, align: 'center' })
        doc.setFont(undefined, 'normal')
        doc.line(40, 80, 555, 80)
        //dados     
        doc.setFontSize(10)

        doc.setFont(undefined, 'bold')
        doc.text('Nome: ', 45, 93)
        doc.setFont(undefined, 'normal')
        doc.text(`${d.nome}`, 80, 93)

        doc.line(430, 80, 430, 100) // vertical line
        doc.setFont(undefined, 'bold')
        doc.text('Celular: ', 435, 93)
        doc.setFont(undefined, 'normal')
        doc.text(`${formatarCelular(telefone1)}`, 475, 93)
        doc.text(``, 85, 93)

        doc.line(40, 100, 555, 100)
        doc.setFont(undefined, 'bold')
        doc.text('Endereco: ', 45, 113)
        doc.setFont(undefined, 'normal')
        doc.text(`${endereco}`, 100, 113)

        doc.line(40, 120, 555, 120)
        doc.setFont(undefined, 'bold')
        doc.text('Bairro: ', 45, 133)
        doc.setFont(undefined, 'normal')
        doc.text(`${bairro}`, 80, 133)
        doc.line(275, 120, 275, 140) // vertical line
        doc.setFont(undefined, 'bold')
        doc.text('Cidade: ', 280, 133)
        doc.setFont(undefined, 'normal')
        doc.text(`${cidade}`, 320, 133)

        doc.line(40, 140, 555, 140)

        doc.line(40, 170, 555, 170)
        doc.setFillColor(120)
        doc.rect(40, 170, 515, 30, 'FD')
        doc.setFontSize(15)
        doc.setTextColor(255)
        doc.setFont(undefined, 'bold')
        doc.text('Descrição do Serviço', 297, 189, { maxWidth: 500, align: 'center' })
        doc.setFont(undefined, 'normal')
        doc.setTextColor(40)

        doc.setFontSize(12)
        var descricao = doc.splitTextToSize(d.descricao, 500);
        doc.text(descricao, 45, 220)
        doc.text(d.descricao, 45, 220, { maxWidth: 500 })

        /*
                doc.line(40, 200, 555, 200)
                doc.line(40, 230, 555, 230)
                doc.line(40, 260, 555, 260)
                doc.line(40, 290, 555, 290)
                doc.line(40, 320, 555, 320)
                doc.line(40, 350, 555, 350)
                doc.line(40, 380, 555, 380)
                doc.line(40, 410, 555, 410)
                doc.line(40, 440, 555, 440)
                doc.line(40, 470, 555, 470)
        */
        doc.setFontSize(12)
        doc.setFont(undefined, 'bold')
        doc.text(`${d.cliCidade}, _____ de ________________________ de ______`,
            297, 700, { maxWidth: 510, align: 'center' })
        doc.setFont(undefined, 'normal')

        doc.setFontSize(12)
        doc.line(70, 780, 270, 780)
        doc.setFont(undefined, 'bold')
        doc.text('ASSINATURA TÉCNICO', 170, 793, { maxWidth: 200, align: 'center' })
        doc.setFont(undefined, 'normal')
        doc.line(325, 780, 525, 780)
        doc.setFont(undefined, 'bold')
        doc.text(d.nome, 425, 793, { maxWidth: 200, align: 'center' })
        doc.setFont(undefined, 'normal')

        //Abre a janela de visualização e impressao
        window.open(doc.output('bloburl'), '_blank')
        doc.autoPrint()
        doc.output("dataurlnewwindows")

    })
}

function ticketClienteAtualizar() {
    const id = localStorage.getItem('idTicket')

    if (validarCamposBranco()) return
    const historico = $('#historico-ticket').val()
    const tag = ticketClienteGeraTag()
    const messagem = $('#descricao-ticket').val()

    const descricao = `${historico}\n\n${tag}\n${messagem}`

    $.ajax({
        start: boxProcessando(),
        url: '/ticketClienteAtualizar',
        method: 'Post',
        data: JSON.stringify(
            {
                idTicket: id,
                assunto: $('#assunto-ticket').val().toUpperCase(),
                descricao: descricao,
                status: "AGUARDANDO RESPOSTA"
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao alterar dados do Base")
    }).done(function (r) {
        boxAteradoSucesso()
        ticketClienteLimparFormulario()
        ticketClienteTicketListar()
    })

}

function ticketClienteFechar() {
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
                    url: '/ticketClienteAtualizar',
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
                    ticketClienteLimparFormulario()
                    ticketClienteTicketListar()
                })
            }
        })
    } else {

        if (validarCamposBranco()) return

        const historico = $('#historico-ticket').val()
        const tag = ticketClienteGeraTag()
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
                    url: '/TicketsClienteAtualizar',
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
                    ticketClienteLimparFormulario()
                    ticketClienteTicketListar()
                })
            }
        })
    }
}

function ticketClienteInserir() {

    if (validarCamposBranco()) return

    if ($('#id-cliente').val() == "0") {
        boxAdvertenciaAuto("Deve selecionar um cliene antes de inserir")
        $('#id-cliente').focus()
        return
    }

    $.ajax({
        start: boxProcessando(),
        url: '/ticketClienteInserir',
        method: 'Post',
        data: JSON.stringify(
            {
                idMaster: localStorage.getItem('idFranqueado'),
                idSlave: $('#id-cliente').val(),
                assunto: $('#assunto-ticket').val().toUpperCase(),
                descricao: $('#descricao-ticket').val().toUpperCase(),
                status: "NOVO",
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o Ticket")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        ticketClienteLimparFormulario()
        ticketClienteTicketListar()
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

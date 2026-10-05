$(document).ready(function () {
        //Carrega o logo
        const id =  localStorage.getItem('idFranqueado')
        console.log(`logoId${id}`)
       
        $('#logo').append(`
            <img src="/recursos/img/franqueado/logoId${id}.png"
                alt="LogoMarca: logoId${id}.png"
                class="img-fluid">
        `)

    MenuPrincipalCarregarDadosLogin()
    MenuPrincipalCarregarInformativo()
    MenuPrincipalCarregarSemComunicar()
    

    // carregarTicket()   



})

function MenuPrincipalCarregarDadosLogin() {
    $('#display-nome-monitoramento').val(localStorage.getItem('nomeFranqueado'))
    //$('#display-nome-monitoramento').val(geradorIdCurto())
    $('#display-nome-usuario').text(localStorage.getItem('nomeUsuario'))
}

// Carrega os dado do informativo
function MenuPrincipalCarregarInformativo() {
    let lista
    buscarLista(l => {
        lista = l
        // qtd = l.length

        contador(0, 2)

    })

    function contador(cont, qtd) {
        exibeInformativo(lista[cont])
        setTimeout(() => {
            cont++
            if (cont >= qtd) cont = 0


            contador(cont, qtd)
        }, 60000);
    }



    function exibeInformativo(item) {
        $.ajax({
            async: false,
            url: '/MenuPrincipalCarregarInformativo',
            method: 'Post',
            data: JSON.stringify({ idInformativo: item.idInformativo }),
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            // console.log(r)
            if (r.status == "OK") {
                const d = r.dados
                $('#display-informativo').val(d.messagem)
                // tempo = parseInt(d.tempo) * 1000
            }
        })
    }

    function buscarLista(next) {
        $.ajax({
            url: 'MenuPrincipalBuscarListaInformativo',
            method: 'Post',
            data: JSON.stringify({ idAlvo: localStorage.getItem('idFranqueado') })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == "OK") {
                next(r.dados)
            }

        })
    }

}

// Carrega os dados da tabela de sem comunicação
function MenuPrincipalCarregarSemComunicar() {

    $.ajax({
        url: '/MenuPrincipalCarregarSemComunicar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status != 'Vazio') {
            if (r.dados.length > 0) {
                $('#qtd-sem-comunicar').text(r.dados.length)
                $('#tab-sem-comunicacao').empty()
                r.dados.forEach(i => {
                    let horas
                    let ultimoEvt

                    if (i.dataUltimoEvento == '01/01/0001 00:00:00') {
                        ultimoEvt = 'Nuca Conectou'
                        horas = '0'
                    } else {
                        const now = new Date()
                        ultimoEvt = i.dataUltimoEvento
                        const past = new Date(horaBrToUs(i.dataUltimoEvento))
                        const diff = Math.abs(now.getTime() - past.getTime()); // Subtrai uma data pela outra
                        horas = Math.ceil(diff / (1000 * 60 * 60)); // Divide o total pelo total de milisegundos correspondentes a 1 dia. (1000 milisegundos = 1 segundo).                        
                    }
                    const nomeCliente = (i.conta == '0000') ? i.idModelo : i.nomeCliente

                    const descricaoEvento = (i.descricaoUltimoEvento.length > 40) ?
                        i.descricaoUltimoEvento.substring(0, 37) + '...' :
                        i.descricaoUltimoEvento



                    $('#tab-sem-comunicacao').append(`
                        <tr style="font-size: 11px;">
                            <td>${i.conta}</td>
                            <td>${i.nome}</td>
                            <td>${i.nomeCliente}</td>
                            <td>${i.codigoUltimoEvento}</td>
                            <td>${descricaoEvento}</td>
                            <td>${ultimoEvt}</td>
                            <td>${horas}</td>
                        </tr>
                    `)
                });

            }
        }
    })
}

// Carrega os tickets do vinculado ao franqueado
function MenuPrincipalCarregarTicket() {
    const idFranqueado = $("#id-franqueado").val()
    const url = `/carregar-tickets/${idFranqueado}`
    $.ajax({
        url: url,
        method: 'GET',
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status != "Vazio") {
            $('#tab-ticket').empty()
            r.dados.forEach(item => {
                const html = `
                    <tr>
                        <td>${item.idTicket}</td>
                        <td>${item.assunto}</td>
                        <td>${item.status}</td>
                        <td>

                            <a
                                href="#">
                                <button class="btn btn-xs btn-success">
                                    <span class="glyphicon
                                        glyphicon-eye-open">
                                    </span>

                                </button>
                            </a>

                        </td>
                        <td>
                            <a
                                href="/excluir-tickets/${idTicket}">
                                <button class="btn btn-xs btn-danger">
                                    <span class="glyphicon
                                        glyphicon-remove-circle">
                                    </span></button>
                            </a>
                        </td>
                    </tr>
                `
                $('#tab-ticket').append()
            })
            return
        }
        $('#tab-ticket').empty()
    })
}

$(document).ready(function () {
    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-filtrar').on('click', filtrar)
    $('#btn-imprimir').on('click', function () {
        gerarPdf('RELATÓRIO DE LIGAÇÕES', '#tabLigacoes', 'portrait')
    })

    carregarClientes()
})

function carregarClientes() {
   
    $.ajax({
        url: '/carregarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        console.log(r)
        $('#id-cliente').empty()
        $('#id-cliente').append('<option value="0">SELECIONE UM CLIENTE</option>')
        
        r.dados.forEach(i => {
            $('#id-cliente').append(`<option value="${i.idCliente}">${i.nome}</option>`)
        });
       
    })
}

function filtrar() {
    const idCliente = $('#id-cliente').val()
    var start = $('#start').val()
    var end = $('#end').val()

    if (start != "") {
        start = start.replace('T', ' ') + ":00"
    } else {
        boxMesagemAtencaoPersonalizada('Data inicial deve ser informada')
        return
    }

    if (end != "") {
        end = end.replace('T', ' ') + ":00"
    } else {
        boxMesagemAtencaoPersonalizada('Data final deve ser informada')
        return
    }

    carregarTabela(start, end, idCliente)

}

function carregarTabela(start, end, idCliente) {
    
    
    const s = moment(start); // data atual
    if (s.isAfter(moment())){
        boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data atual')
        return
    }

    const f = moment(end);
    if (s.isAfter(f)){
        boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data final')
        return
    }
    
    const horas = f.diff(s, 'hours');

    if (horas > 744){
        boxMesagemAtencaoPersonalizada('A data inicial para data final não pode ser maior que 31 dias')
        return
    } 
    
    let filtro = `
        WHERE custoAtendimento.ID_Franqueado = ${localStorage.getItem('idFranqueado')} 
        AND custoAtendimento.TipoOperacao = 'LIGAÇÃO' `

    if (start != "0") {
        filtro += `AND custoAtendimento.DataOperacao >= '${start}' `
    }

    if (end != "0") {
        filtro += `AND custoAtendimento.DataOperacao <= '${end}' `
    }

    if (idCliente != "0") {
        filtro += `AND custoAtendimento.ID_Cliente = ${idCliente} `
    }

    $.ajax({
        start: boxProcessando(),
        url: '/custo-atendimento-listar',
        method: 'POST',
        data: JSON.stringify({
            filtro: filtro
        })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        boxFechar()
        $('tbody').empty()
        
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                console.log(i)
                const playAudio = (i.linkAudio != '') ? `
                    <button class="btn btn-success" tipo="play" link="${i.linkAudio}">
                        <i class="bi bi-play"></i>
                    </button>
                `: `
                    <button class="btn btn-secondary">
                        <i class="bi bi-play"></i>
                    </button>
                `

                $('tbody').append(`
                    <tr>
                        <td>${i.nomeCliente}</td>
                        <td>${i.dataOperacao}</td>
                        <td>${i.dadoOperacao}</td>
                        <td>${i.nomeOperador}</td>
                        <td>${playAudio}</td>
                    </tr>
                `)
            });

            $('button[tipo=play]').on('click', function () {
                const horizontal = "left=" + (window.innerWidth - 400) / 2
    
                window.open(
                    $(this).attr('link'),
                    "play-audio",
                    "height=80,width=400, " + horizontal)
            })
        }       
    })
}


function limparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    $('thead').empty()

}
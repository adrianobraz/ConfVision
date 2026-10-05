$(document).ready(function () {

    $('#btn-limpar').on('click', limparFormulario)
   
    $('#btn-filtrar').on('click', filtrar)
   
    $('#btn-imprimir').on('click', function () {
        gerarPdf('RELATÓRIO DE EVENTOS', '#tabEventos', 'landscape')
    })

    $('#id-cliente').on('change', function () {
        if ($('#id-cliente').val() == '0') {
            $('#id-dispositivos').empty()
            
            $('#id-dispositivos').append(
                '<option value="0">SELECIONE UM DISPOSITIVO</option>'
            )

        } else {            
            carregarDispositivos($('#id-cliente').val())

        }
    })

    carregarClientes()

})

function filtrar() {

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

    carregarTabela(start, end)

}

function carregarTabela(start, end) {

    const s = moment(start); // data atual
    if (s.isAfter(moment())) {
        boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data atual')
        return
    }

    const f = moment(end);
    if (s.isAfter(f)) {
        boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data final')
        return
    }

    const horas = f.diff(s, 'hours');

    if (horas > 744) {
        boxMesagemAtencaoPersonalizada('A data inicial para data final não pode ser maior que 31 dias')
        return
    }
    if ($('#id-cliente').val() != '0') {
        if ($('#id-cliente').val() == '0') {
            boxMesagemAtencaoPersonalizada('Selecione um cliente')
            return
        }

        if ($('#id-dispositivos').val() == '0') {
            boxMesagemAtencaoPersonalizada('Selecione um Dispositivo')
            return
        }
    }  


    // '2022082403304555848127799'
    const filtro = `
        WHERE processo.ID_Dispositivo = '${$('#id-dispositivos').val()}'
        AND processo.DataAtenFim >= '${start}'
        AND processo.DataAtenFim <= '${end}'
        AND processo.Nivel > '0'
    `
    
    $.ajax({
        start: boxProcessando(),
        url: 'relatorioAtendimentoListar',
        method: 'POST',
        data: JSON.stringify({ filtro: filtro })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        boxFechar()
        $('tbody').empty()

        if (r.status != 'Vazio') {
            $('tbody').append(`
                <tr>
                    <th class="bg-dark text-white text-center">${r.dados[0].cliNome}</th>
                    <th colspan="2" class="bg-dark text-white text-center">${r.dados[0].dispNome}</th>
                </tr>
            `)
            
            r.dados.map(i => {

                // const nick = (i.idOperador == '0') ? 'ATENDIMENTO AUTOMATICO' : i.nickAtendente

                const descricao = i.descricao.replaceAll('[', '<br>[')
                $('tbody').append(`
                    <tr>
                        <th style="width: 56%;" class="bg-dark text-white">Operador</th>
                        <th style="width: 22%;" class="bg-dark text-white">DataInicio</th>
                        <th style="width: 22%;" class="bg-dark text-white">DataFinalização</th>
                    </tr>
                    
                    <tr>
                        <td>${i.nick}</td>
                        <td>${i.dataAtenInicio}</td>
                        <td>${i.dataAtenFim}</td>
                    </tr>

                    <tr class="bg-light"> 
                    <td colspan="4">Descrição do atendimento </td> 
                    </tr>
                    
                    <tr>
                        <td colspan="4" class="text-start">${descricao}</td>
                    </tr>
                `)
            })

        }
    })
}

function carregarClientes() {
    
    //GerenciarDispositivoCarregarCliente
    $.ajax({
        url: '/carregarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        $('#id-cliente').empty()
        $('#id-cliente').append('<option value="0">SELECIONE UM CLIENTE</option>')
        
        r.dados.forEach(i => {
            $('#id-cliente').append(`<option value="${i.idCliente}">${i.nome}</option>`)
        });
       
    })
}

function carregarDispositivos(idCliente) {

    $.ajax({
        url: '/carregarDispositivos',
        method: 'Post',
        data: JSON.stringify({ idCliente: idCliente })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os dipositivos")
    }).done(function (r) {
        console.log(r)
        $('#id-dispositivos').empty().append('<option value="0">SELECIONE UM DISPOSITIVO</option>')
        if (r.status != 'Vazio') {

            r.dados.forEach(i => {
                $('#id-dispositivos').append(`<option value="${i.idDispositivo}">${i.nome}</option>`)
            });
        }
    })
}

function limparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    $('tbody').empty()
}
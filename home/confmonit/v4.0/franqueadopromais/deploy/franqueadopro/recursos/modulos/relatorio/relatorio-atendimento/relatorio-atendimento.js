var listaClientes = []
var fpStart, fpEnd
var fpPagAtendimento = null
var fpPagAtendFiltro = ''

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

    // Pesquisa de cliente no combo (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', filtrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') filtrarClientes()
    })
    $('#lista-clientes').on('click', '.fp-combo-item', function () {
        $('#busca-cliente').val($(this).attr('data-nome'))
        $('#id-cliente').val($(this).attr('data-id')).trigger('change')
        $('#lista-clientes').addClass('d-none')
    })
    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-cliente, #lista-clientes').length) {
            $('#lista-clientes').addClass('d-none')
        }
    })

    // Calendário + hora (flatpickr)
    if (window.flatpickr) {
        if (flatpickr.l10ns && flatpickr.l10ns.pt) {
            flatpickr.localize(flatpickr.l10ns.pt)
        }
        var cfg = {
            enableTime: true,
            time_24hr: true,
            dateFormat: 'Y-m-d H:i',
            altInput: true,
            altFormat: 'd/m/Y H:i'
        }
        fpStart = flatpickr('#start', cfg)
        fpEnd = flatpickr('#end', cfg)
    }

    carregarClientes()

})

function clienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function filtrarClientes() {
    var raw = ($('#busca-cliente').val() || '').toLowerCase().trim()
    var termoNum = raw.replace(/[^a-z0-9]/g, '')
    var lista = $('#lista-clientes')
    lista.empty()

    if (raw == '') {
        $('#id-cliente').val('0').trigger('change')
        lista.addClass('d-none')
        return
    }

    var resultados = listaClientes.filter(function (i) {
        var texto = clienteTextoBusca(i)
        var textoNum = texto.replace(/[^a-z0-9]/g, '')
        return texto.indexOf(raw) != -1 || (termoNum != '' && textoNum.indexOf(termoNum) != -1)
    }).slice(0, 30)

    if (resultados.length == 0) {
        lista.append('<li class="fp-combo-empty">Nenhum cliente encontrado</li>')
    } else {
        resultados.forEach(function (i) {
            var doc = i.documento1 ? ' — ' + i.documento1 : ''
            $('<li class="fp-combo-item"></li>')
                .attr('data-id', i.idCliente)
                .attr('data-nome', i.nome)
                .text(i.nome + doc)
                .appendTo(lista)
        })
    }
    lista.removeClass('d-none')
}

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

    if (fpPagAtendimento) fpPagAtendimento.destroy()
    fpPagAtendFiltro = filtro
    var cabecalhoInserido = false

    fpPagAtendimento = fpScrollPaginacao({
        url: 'relatorioAtendimentoListar',
        tbodySel: 'tbody',
        getPayload: function () {
            return { filtro: fpPagAtendFiltro }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                if (!cabecalhoInserido) {
                    cabecalhoInserido = true
                    $('tbody').append(`
                        <tr>
                            <th class="bg-dark text-white text-center">${i.cliNome}</th>
                            <th colspan="2" class="bg-dark text-white text-center">${i.dispNome}</th>
                        </tr>
                    `)
                }
                var descricao = (i.descricao || '').replaceAll('[', '<br>[')
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
                    <tr class="bg-dark text-white"><td colspan="4">Descrição do atendimento</td></tr>
                    <tr><td colspan="4" class="text-start">${descricao}</td></tr>
                `)
            })
        },
        onFail: function () {
            boxErro("Erro ao carregar a tabela")
        }
    })

    fpPagAtendimento.reset()
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
        listaClientes = (r.dados || [])
        $('#id-cliente').val('0')
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
    if (fpPagAtendimento) fpPagAtendimento.destroy()
    $('#formulario').each(function () {
        this.reset();
    })
    $('#busca-cliente').val('')
    $('#id-cliente').val('0')
    $('#id-dispositivos').empty().append('<option value="0">SELECIONE UM DISPOSITIVO</option>')
    $('#lista-clientes').addClass('d-none')
    if (fpStart) fpStart.clear()
    if (fpEnd) fpEnd.clear()
    $('tbody').empty()
}
var listaClientes = []
var fpStart, fpEnd
var fpPagLigacoes = null
var fpLigFiltroBase = ''

$(document).ready(function () {
    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-filtrar').on('click', filtrar)
    $('#btn-imprimir').on('click', function () {
        gerarPdf('RELATÓRIO DE LIGAÇÕES', '#tabLigacoes', 'portrait')
    })

    // Pesquisa de cliente no combo (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', filtrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') filtrarClientes()
    })
    $('#lista-clientes').on('click', '.fp-combo-item', function () {
        $('#busca-cliente').val($(this).attr('data-nome'))
        $('#id-cliente').val($(this).attr('data-id'))
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

function carregarClientes() {
   
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
        $('#id-cliente').val('0')
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
    const s = moment(start)
    if (s.isAfter(moment())) {
        boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data atual')
        return
    }

    const f = moment(end)
    if (s.isAfter(f)) {
        boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data final')
        return
    }

    const horas = f.diff(s, 'hours')
    if (horas > 744) {
        boxMesagemAtencaoPersonalizada('A data inicial para data final não pode ser maior que 31 dias')
        return
    }

    let filtro = `
        WHERE custoAtendimento.ID_Franqueado = ${localStorage.getItem('idFranqueado')} 
        AND custoAtendimento.TipoOperacao = 'LIGAÇÃO' `

    if (start != '0') {
        filtro += `AND custoAtendimento.DataOperacao >= '${start}' `
    }
    if (end != '0') {
        filtro += `AND custoAtendimento.DataOperacao <= '${end}' `
    }
    if (idCliente != '0') {
        filtro += `AND custoAtendimento.ID_Cliente = ${idCliente} `
    }

    fpLigFiltroBase = filtro

    if (fpPagLigacoes) fpPagLigacoes.destroy()

    fpPagLigacoes = fpScrollPaginacao({
        url: '/custo-atendimento-listar',
        tbodySel: 'tbody',
        getPayload: function (offset) {
            return {
                filtro: fpLigFiltroBase +
                    ' ORDER BY custoAtendimento.DataOperacao DESC LIMIT 100 OFFSET ' + offset
            }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                const playAudio = (i.linkAudio != '') ? `
                    <button class="btn btn-success" tipo="play" link="${i.linkAudio}">
                        <i class="bi bi-play"></i>
                    </button>
                ` : `
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
            })

            $('button[tipo=play]').off('click').on('click', function () {
                const horizontal = 'left=' + (window.innerWidth - 400) / 2
                window.open(
                    $(this).attr('link'),
                    'play-audio',
                    'height=80,width=400, ' + horizontal)
            })
        },
        onEmpty: function () {
            $('tbody').append('<tr><td colspan="5" class="text-center">NENHUM REGISTRO</td></tr>')
        },
        onFail: function () {
            boxErro('Erro ao carregar a tabela')
        }
    })

    fpPagLigacoes.reset()
}


function limparFormulario() {
    $('#formulario').each(function () {
        this.reset()
    })
    $('#busca-cliente').val('')
    $('#id-cliente').val('0')
    $('#lista-clientes').addClass('d-none')
    if (fpStart) fpStart.clear()
    if (fpEnd) fpEnd.clear()
    if (fpPagLigacoes) fpPagLigacoes.destroy()
    fpPagLigacoes = null
    $('tbody').empty()
}
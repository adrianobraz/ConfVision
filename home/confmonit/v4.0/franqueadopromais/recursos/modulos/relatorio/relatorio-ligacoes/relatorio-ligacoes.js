var listaClientes = []
var fpStart, fpEnd
var fpPagLigacoes = null
var fpLigFiltroBase = null

$(document).ready(function () {
    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-filtrar').on('click', filtrar)
    $('#btn-imprimir').on('click', function () {
        gerarPdf('RELATÓRIO DE LIGAÇÕES', '#tabLigacoes', 'portrait')
    })

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

    carregarClientes(function () {
        carregarPadraoUltimos100()
    })
})

function limparDatasFiltro() {
    if (fpStart) fpStart.clear()
    if (fpEnd) fpEnd.clear()
    $('#start').val('')
    $('#end').val('')
}

function carregarPadraoUltimos100() {
    limparDatasFiltro()
    // Sem filtro de data: backend devolve as mais recentes (100 + scroll)
    carregarTabela('', '', '0')
}

function carregarClientes(aoPronto) {
    $.ajax({
        url: '/carregarClientes',
        method: 'POST',
        contentType: 'application/json; charset=utf-8',
        dataType: 'json',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
        boxErro('Erro ao carregar os clientes')
        listaClientes = []
        if (typeof aoPronto === 'function') aoPronto()
    }).done(function (r) {
        listaClientes = (r && r.dados) ? r.dados : []
        $('#id-cliente').val('0')
        if (typeof aoPronto === 'function') aoPronto()
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

    if (!listaClientes || listaClientes.length === 0) {
        lista.append('<li class="fp-combo-empty">Lista de clientes vazia — recarregue a página</li>')
        lista.removeClass('d-none')
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
                .text((i.nome || '') + doc)
                .appendTo(lista)
        })
    }
    lista.removeClass('d-none')
}

function filtrar() {
    const idCliente = $('#id-cliente').val() || '0'
    var start = lerDataInput('start')
    var end = lerDataInput('end')

    if (!start) {
        boxMesagemAtencaoPersonalizada('Data inicial deve ser informada')
        return
    }
    if (!end) {
        boxMesagemAtencaoPersonalizada('Data final deve ser informada')
        return
    }

    carregarTabela(start, end, idCliente)
}

function lerDataInput(id) {
    if (id === 'start' && fpStart && fpStart.selectedDates && fpStart.selectedDates[0]) {
        return moment(fpStart.selectedDates[0]).format('YYYY-MM-DD HH:mm:ss')
    }
    if (id === 'end' && fpEnd && fpEnd.selectedDates && fpEnd.selectedDates[0]) {
        return moment(fpEnd.selectedDates[0]).format('YYYY-MM-DD HH:mm:ss')
    }
    return normalizarDataFiltro(($('#' + id).val() || '').trim())
}

function normalizarDataFiltro(v) {
    v = String(v || '').replace('T', ' ').trim()
    if (!v) return ''
    // dd/mm/yyyy HH:mm (altInput)
    var mBr = v.match(/^(\d{2})\/(\d{2})\/(\d{4})\s+(\d{2}):(\d{2})(?::(\d{2}))?$/)
    if (mBr) {
        return mBr[3] + '-' + mBr[2] + '-' + mBr[1] + ' ' + mBr[4] + ':' + mBr[5] + ':' + (mBr[6] || '00')
    }
    if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/.test(v)) return v + ':00'
    if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(v)) return v
    while ((v.match(/:/g) || []).length > 2) {
        v = v.substring(0, v.lastIndexOf(':'))
    }
    if ((v.match(/:/g) || []).length === 1) v += ':00'
    return v
}

function carregarTabela(start, end, idCliente) {
    var semData = (!start || start === '0') && (!end || end === '0')

    if (!semData) {
        const s = moment(start)
        if (!s.isValid()) {
            boxMesagemAtencaoPersonalizada('Data inicial inválida')
            return
        }
        if (s.isAfter(moment())) {
            boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data atual')
            return
        }

        const f = moment(end)
        if (!f.isValid()) {
            boxMesagemAtencaoPersonalizada('Data final inválida')
            return
        }
        if (s.isAfter(f)) {
            boxMesagemAtencaoPersonalizada('A data inicial não pode ser maior que a data final')
            return
        }

        const horas = f.diff(s, 'hours')
        if (horas > 744) {
            boxMesagemAtencaoPersonalizada('A data inicial para data final não pode ser maior que 31 dias')
            return
        }
    }

    var idFra = localStorage.getItem('idFranqueado') || ''
    if (!idFra) {
        boxMesagemAtencaoPersonalizada('Sessão sem franqueado. Faça login novamente.')
        return
    }

    fpLigFiltroBase = {
        idFranqueado: idFra,
        idCliente: (idCliente && idCliente != '0') ? String(idCliente) : '0',
        dataInicio: semData ? '' : start,
        dataFim: semData ? '' : end
    }

    if (fpPagLigacoes) fpPagLigacoes.destroy()

    fpPagLigacoes = fpScrollPaginacao({
        url: '/custo-atendimento-listar',
        tbodySel: '#tabLigacoes tbody',
        getPayload: function () {
            return {
                idFranqueado: fpLigFiltroBase.idFranqueado,
                idCliente: fpLigFiltroBase.idCliente,
                dataInicio: fpLigFiltroBase.dataInicio,
                dataFim: fpLigFiltroBase.dataFim
            }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                const hasAudio = i.linkAudio != null && String(i.linkAudio).trim() != ''
                const playAudio = hasAudio
                    ? '<button class="btn btn-success" tipo="play" link="' + String(i.linkAudio).replace(/"/g, '&quot;') + '"><i class="bi bi-play"></i></button>'
                    : '<button class="btn btn-secondary"><i class="bi bi-play"></i></button>'

                $('#tabLigacoes tbody').append(
                    '<tr>' +
                    '<td>' + (i.nomeCliente || '') + '</td>' +
                    '<td>' + (i.dataOperacao || '') + '</td>' +
                    '<td>' + (i.dadoOperacao || '') + '</td>' +
                    '<td>' + (i.nomeOperador || '') + '</td>' +
                    '<td>' + playAudio + '</td>' +
                    '</tr>'
                )
            })

            $('#tabLigacoes button[tipo=play]').off('click').on('click', function () {
                const horizontal = 'left=' + (window.innerWidth - 400) / 2
                window.open(
                    $(this).attr('link'),
                    'play-audio',
                    'height=80,width=400, ' + horizontal)
            })
        },
        onEmpty: function () {
            $('#tabLigacoes tbody').append(
                '<tr><td colspan="5" class="text-center">NENHUM REGISTRO</td></tr>'
            )
        },
        onFail: function (xhr) {
            var msg = 'Erro ao carregar a tabela'
            try {
                var j = JSON.parse(xhr.responseText || '{}')
                if (j && (j.status || j.erro || j.message)) {
                    msg = j.status || j.erro || j.message
                }
            } catch (e) { /* ignore */ }
            boxErro(msg)
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
    if (fpPagLigacoes) fpPagLigacoes.destroy()
    fpPagLigacoes = null
    $('#tabLigacoes tbody').empty()
    carregarPadraoUltimos100()
}

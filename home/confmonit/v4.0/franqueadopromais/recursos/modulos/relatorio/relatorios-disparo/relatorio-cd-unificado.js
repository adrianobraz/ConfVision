/* Central de Disparos — Visao Unificada (lista + detalhe) */
var FP_CD_UNI_PAGE = 100
var listaClientes = []
var fpStart, fpEnd
var uniOffset = 0
var uniLoading = false
var uniFim = false
var uniTotal = 0
var uniSelecionado = null
var uniMap = {}

var UNI_PAINEIS = [
    { key: 'pendente', titulo: 'Eventos Pendentes', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'grupo', l: 'Grupo' },
        { k: 'DispNome', l: 'Dispositivo', alt: 'idDispositivo' },
        { k: 'zonaUser', l: 'Zona' },
        { k: 'agendadoPara', l: 'Agendado', fmt: 'dt' },
        { k: '_st', l: 'Status', fmt: 'status_pendente' }
    ]},
    { key: 'finalizadoRobo', titulo: 'Finalizados por Robô', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'Usuario', l: 'Usuário' },
        { k: 'mensagem', l: 'Mensagem' },
        { k: 'telefone', l: 'Telefone' }
    ]},
    { key: 'finalizadoBot', titulo: 'Finalizados pelo Bot', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'status', l: 'Status' },
        { k: 'motivo_final', l: 'Motivo' },
        { k: 'acao_final', l: 'Ação' },
        { k: 'erro', l: 'Erro' }
    ]},
    { key: 'whatsappEnviados', titulo: 'WhatsApp Enviados', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'whats_sender', l: 'Destino' },
        { k: 'ctiGrupo', l: 'Grupo' },
        { k: 'DispNome', l: 'Dispositivo' },
        { k: 'whats_text', l: 'Mensagem' }
    ]},
    { key: 'ligacoes', titulo: 'Ligações', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'whats_sender', l: 'Destino' },
        { k: 'ctiGrupo', l: 'Grupo' },
        { k: 'DispNome', l: 'Dispositivo' }
    ]},
    { key: 'sms', titulo: 'SMS', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'whats_sender', l: 'Destino' },
        { k: 'ctiGrupo', l: 'Grupo' },
        { k: 'DispNome', l: 'Dispositivo' },
        { k: 'whats_text', l: 'Mensagem' }
    ]},
    { key: 'smsHistorico', titulo: 'SMS Histórico', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'whats_sender', l: 'Destino' },
        { k: 'message_status', l: 'Status' },
        { k: 'DispNome', l: 'Dispositivo' },
        { k: 'whats_text', l: 'Texto' }
    ]},
    { key: 'ligacaoHistorico', titulo: 'Ligação Histórico', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'receiver', l: 'Telefone', alt: 'telefone' },
        { k: 'duration', l: 'Duração', fmt: 'sec' },
        { k: 'status', l: 'Status' },
        { k: 'tipoevento', l: 'Evento' },
        { k: 'analysis_transcript_summary', l: 'Transcrição', fmt: 'transcript' },
        { k: '_audio', l: 'Áudio', fmt: 'audio' }
    ]},
    { key: 'ligacaoErros', titulo: 'Erros e Tentativas', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'telefone', l: 'Telefone' },
        { k: 'tipoevento', l: 'Evento' },
        { k: 'falha', l: 'Falha', fmt: 'bool' },
        { k: 'tentativas', l: 'Tentativas' },
        { k: 'atendido', l: 'Atendido', fmt: 'bool' },
        { k: 'DispNome', l: 'Dispositivo' }
    ]},
    { key: 'filaEnvio', titulo: 'Eventos para Enviar ou Ligação', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'numerowhatsapp', l: 'WhatsApp' },
        { k: 'tipoevento', l: 'Evento' },
        { k: 'enviartexto', l: 'Texto', fmt: 'bool' },
        { k: 'ligar', l: 'Ligar', fmt: 'bool' },
        { k: 'disparo', l: 'Disparado', fmt: 'bool' },
        { k: 'DispNome', l: 'Dispositivo' }
    ]},
    { key: 'filaLigacao', titulo: 'Filas de Ligações', cols: [
        { k: 'created_at', l: 'Data', fmt: 'dt' },
        { k: 'telefone', l: 'Telefone' },
        { k: 'tipoevento', l: 'Evento' },
        { k: 'tentativas', l: 'Tentativas' },
        { k: 'falha', l: 'Falha', fmt: 'bool' },
        { k: 'DispNome', l: 'Dispositivo' }
    ]}
]

$(document).ready(function () {
    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-filtrar').on('click', function () { recarregarLista() })

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

    $('#cd-uni-lista').on('scroll', onScrollLista)
    $('#cd-uni-lista').on('click', '.cd-uni-item', function () {
        var id = $(this).attr('data-id')
        var item = uniMap[id] || $(this).data('item')
        selecionarEvento(item)
    })

    $('#cd-uni-detalhe').on('click', '.btn-cd-transcript', function () {
        var key = $(this).attr('data-tr-key')
        abrirPopupTranscricao(window._cdTranscriptMap && window._cdTranscriptMap[key])
    })
    $('#cd-uni-detalhe').on('click', '.cd-tr-preview', function () {
        var key = $(this).closest('.cd-tr-cell').find('.btn-cd-transcript').attr('data-tr-key')
        abrirPopupTranscricao(window._cdTranscriptMap && window._cdTranscriptMap[key])
    })

    if (window.flatpickr) {
        if (flatpickr.l10ns && flatpickr.l10ns.pt) flatpickr.localize(flatpickr.l10ns.pt)
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
    recarregarLista()
})

function carregarClientes() {
    $.ajax({
        url: '/carregarClientes',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function () {
        listaClientes = []
    }).done(function (r) {
        listaClientes = (r && r.dados) ? r.dados : []
        $('#id-cliente').val('0')
    })
}

function clienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function filtrarClientes() {
    var raw = ($('#busca-cliente').val() || '').toLowerCase().trim()
    var lista = $('#lista-clientes')
    lista.empty()
    if (raw == '') {
        $('#id-cliente').val('0')
        lista.addClass('d-none')
        return
    }
    var resultados = listaClientes.filter(function (i) {
        return clienteTextoBusca(i).indexOf(raw) >= 0
    }).slice(0, 30)
    if (!resultados.length) {
        lista.addClass('d-none')
        return
    }
    resultados.forEach(function (i) {
        var idCli = i.idCliente || i.id || ''
        lista.append(
            '<li class="fp-combo-item" data-id="' + escAttr(idCli) + '" data-nome="' +
            escAttr(i.nome || '') + '">' + escHtml(i.nome || '') + '</li>'
        )
    })
    lista.removeClass('d-none')
}

function toMs(val) {
    if (!val) return 0
    var d = new Date(String(val).replace(' ', 'T'))
    var t = d.getTime()
    return isNaN(t) ? 0 : t
}

function lerDataMs(campo) {
    if (campo === 'start' && fpStart && fpStart.selectedDates && fpStart.selectedDates[0]) {
        return fpStart.selectedDates[0].getTime()
    }
    if (campo === 'end' && fpEnd && fpEnd.selectedDates && fpEnd.selectedDates[0]) {
        return fpEnd.selectedDates[0].getTime()
    }
    return toMs($('#' + campo).val())
}

function limparFormulario() {
    $('#busca-cliente').val('')
    $('#id-cliente').val('0')
    if (fpStart) fpStart.clear()
    if (fpEnd) fpEnd.clear()
    recarregarLista()
}

function montarPayloadLista() {
    return {
        idFranqueado: localStorage.getItem('idFranqueado'),
        idCliente: $('#id-cliente').val() || '0',
        dataInicioMs: lerDataMs('start'),
        dataFimMs: lerDataMs('end'),
        limit: FP_CD_UNI_PAGE,
        offset: uniOffset
    }
}

function recarregarLista() {
    uniOffset = 0
    uniFim = false
    uniTotal = 0
    uniSelecionado = null
    uniMap = {}
    $('#cd-uni-lista').empty().append('<div class="cd-uni-empty" id="cd-uni-lista-empty">Carregando eventos...</div>')
    $('#cd-uni-fim').addClass('d-none')
    $('#qtd-registros').text('Total: 0')
    $('#cd-uni-detalhe').html('<div class="cd-uni-empty">Selecione um evento à esquerda para ver os detalhes.</div>')
    carregarMaisLista()
}

function onScrollLista() {
    var el = document.getElementById('cd-uni-lista')
    if (!el || uniLoading || uniFim) return
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 80) {
        carregarMaisLista()
    }
}

function carregarMaisLista() {
    if (uniLoading || uniFim) return
    uniLoading = true
    $('#cd-uni-loading').removeClass('d-none')

    $.ajax({
        url: '/cdUnificadoLista',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(montarPayloadLista())
    }).fail(function (xhr) {
        uniLoading = false
        $('#cd-uni-loading').addClass('d-none')
        var msg = 'Erro ao carregar eventos'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j && (j.erro || j.message || j.status)) msg = j.erro || j.message || j.status
        } catch (e) { /* ignore */ }
        boxErro(msg)
        if (uniOffset === 0) {
            $('#cd-uni-lista').html('<div class="cd-uni-empty">Não foi possível carregar os eventos.</div>')
        }
    }).done(function (r) {
        uniLoading = false
        $('#cd-uni-loading').addClass('d-none')
        var lista = (r && r.dados) ? r.dados : []
        $('#cd-uni-lista-empty').remove()
        if (uniOffset === 0 && !lista.length) {
            $('#cd-uni-lista').html('<div class="cd-uni-empty">Nenhum evento encontrado.</div>')
        } else {
            appendLista(lista)
        }
        uniOffset += lista.length
        uniTotal += lista.length
        $('#qtd-registros').text('Total: ' + uniTotal + (uniFim || lista.length < FP_CD_UNI_PAGE ? '' : '+'))
        if (lista.length < FP_CD_UNI_PAGE) {
            uniFim = true
            if (uniTotal > 0) $('#cd-uni-fim').removeClass('d-none')
        }
    })
}

function appendLista(lista) {
    var $box = $('#cd-uni-lista')
    lista.forEach(function (item) {
        var idKey = String(item.id)
        uniMap[idKey] = item
        var titulo = item.ctiDescricao || item.codigo || 'Evento #' + item.id
        var meta = [
            fmtData(item.created_at),
            item.nomeCliente || '',
            item.ctiGrupo || '',
            item.idDispositivo || ''
        ].filter(Boolean).join(' · ')
        var $btn = $('<button type="button" class="cd-uni-item"></button>')
        $btn.attr('data-id', idKey)
        $btn.html(
            '<div class="titulo">' + escHtml(titulo) + '</div>' +
            '<div class="meta">' + escHtml(meta) + '</div>'
        )
        $box.append($btn)
    })
}

function selecionarEvento(item) {
    if (!item || item.id == null) {
        $('#cd-uni-detalhe').html('<div class="cd-uni-empty">Não foi possível identificar o evento.</div>')
        return
    }
    uniSelecionado = item
    var idKey = String(item.id)
    $('#cd-uni-lista .cd-uni-item').removeClass('ativo')
    $('#cd-uni-lista .cd-uni-item[data-id="' + idKey.replace(/"/g, '') + '"]').addClass('ativo')

    // Mostra imediatamente os dados da lista (sempre tem informação)
    renderDetalhe({ evento: item }, true)

    $.ajax({
        url: '/cdUnificadoDetalhe',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            alarmEventsId: Number(item.id) || 0,
            idProcesso: item.idProcesso || ''
        })
    }).fail(function (xhr) {
        var msg = 'Erro ao carregar vínculos do evento'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j && (j.erro || j.message || j.status)) msg = j.erro || j.message || j.status
        } catch (e) { /* ignore */ }
        renderDetalhe({ evento: uniSelecionado }, false, msg)
    }).done(function (r) {
        var out = r || {}
        if (!out.evento) out.evento = uniSelecionado
        renderDetalhe(out, false)
    })
}

function totalVinculos(r) {
    var n = 0
    UNI_PAINEIS.forEach(function (p) {
        if (Array.isArray(r[p.key])) n += r[p.key].length
    })
    return n
}

function renderDetalhe(r, carregando, erroMsg) {
    var ev = r.evento || uniSelecionado || {}
    var html = ''

    html += '<div class="cd-uni-resumo text-start">'
    html += '<strong>' + escHtml(ev.ctiDescricao || ev.codigo || ('Evento #' + (ev.id || ''))) + '</strong>'
    html += '<div class="small text-muted">'
    html += escHtml([
        fmtData(ev.created_at),
        ev.nomeCliente,
        ev.ctiGrupo,
        (ev.DispNome || ev.dispNome) ? 'Disp: ' + (ev.DispNome || ev.dispNome) : '',
        ev.zonaUser ? 'Zona: ' + ev.zonaUser : '',
        ev.particao ? 'Part: ' + ev.particao : ''
    ].filter(Boolean).join(' · '))
    html += '</div>'
    if (carregando) {
        html += '<div class="small mt-2 text-muted"><span class="spinner-border spinner-border-sm me-1"></span>Buscando vínculos (WhatsApp, SMS, ligações...)</div>'
    } else if (erroMsg) {
        html += '<div class="small mt-2 text-warning">' + escHtml(erroMsg) + '</div>'
    }
    html += '</div>'

    // Painel fixo: sempre mostra os dados do evento
    html += '<div class="accordion cd-uni-painel text-start" id="cdUniAccordion">'
    html += renderPainelEvento(ev)

    var vinculos = totalVinculos(r)
    if (!carregando && vinculos === 0) {
        html += '<div class="cd-uni-aviso mb-2">Este evento não possui disparos vinculados (WhatsApp, SMS, ligação, fila etc.). Dados do evento acima.</div>'
    }

    UNI_PAINEIS.forEach(function (p, idx) {
        var rows = Array.isArray(r[p.key]) ? r[p.key] : []
        var qtd = rows.length
        var id = 'cd-uni-acc-' + idx
        var aberto = qtd > 0
        html += '<div class="accordion-item">'
        html += '<h2 class="accordion-header">'
        html += '<button class="accordion-button' + (aberto ? '' : ' collapsed') + '" type="button"'
        html += ' data-bs-toggle="collapse" data-bs-target="#' + id + '">'
        html += escHtml(p.titulo)
        html += ' <span class="badge ' + (qtd ? 'bg-primary' : 'bg-secondary') + ' ms-2">' + qtd + '</span>'
        html += '</button></h2>'
        html += '<div id="' + id + '" class="accordion-collapse collapse' + (aberto ? ' show' : '') + '">'
        html += '<div class="accordion-body p-2">'
        if (!qtd) {
            html += '<div class="text-muted small">Nenhum registro vinculado a este evento</div>'
        } else {
            html += renderTabela(rows, p.cols)
        }
        html += '</div></div></div>'
    })
    html += '</div>'

    $('#cd-uni-detalhe').html(html)
}

function renderPainelEvento(ev) {
    var id = 'cd-uni-acc-evento'
    var dispNome = (ev.DispNome || ev.dispNome || '').trim()
    var dispTxt = dispNome || ''
    var grupo = ev.ctiGrupo || ''
    var desc = ev.ctiDescricao || ''
    var codigo = ev.codigo || ''

    function cell(label, valueHtml) {
        if (!valueHtml) return ''
        return '<div class="cd-uni-ev-cell"><span class="cd-uni-ev-lab">' + escHtml(label) +
            '</span><span class="cd-uni-ev-val">' + valueHtml + '</span></div>'
    }

    var html = '<div class="accordion-item">'
    html += '<h2 class="accordion-header">'
    html += '<button class="accordion-button" type="button" data-bs-toggle="collapse" data-bs-target="#' + id + '">'
    html += 'Dados do Evento'
    html += '</button></h2>'
    html += '<div id="' + id + '" class="accordion-collapse collapse show">'
    html += '<div class="accordion-body p-2">'
    html += '<div class="cd-uni-ev-grid">'

    html += cell('Data', escHtml(fmtData(ev.created_at)))
    html += cell('Cliente', escHtml(ev.nomeCliente || ''))

    if (codigo || grupo || desc) {
        html += '<div class="cd-uni-ev-cell cd-uni-ev-span2">'
        html += '<div class="cd-uni-ev-inline">'
        if (codigo) {
            html += '<div><span class="cd-uni-ev-lab">Código</span><span class="cd-uni-ev-val">' + escHtml(codigo) + '</span></div>'
        }
        if (grupo) {
            html += '<div><span class="cd-uni-ev-lab">Grupo</span><span class="cd-uni-ev-val">' + escHtml(grupo) + '</span></div>'
        }
        if (desc) {
            html += '<div><span class="cd-uni-ev-lab">Descrição</span><span class="cd-uni-ev-val">' + escHtml(desc) + '</span></div>'
        }
        html += '</div></div>'
    }

    var zona = ev.zonaUser || ''
    var part = ev.particao || ''
    var conta = ev.conta || ''
    if (dispTxt || ev.idDispositivo || zona || part || conta) {
        html += '<div class="cd-uni-ev-cell cd-uni-ev-span2">'
        html += '<div class="cd-uni-ev-inline">'
        if (dispTxt) {
            html += '<div><span class="cd-uni-ev-lab">Dispositivo</span><span class="cd-uni-ev-val">' + escHtml(dispTxt) + '</span></div>'
        } else if (ev.idDispositivo) {
            html += '<div><span class="cd-uni-ev-lab">Dispositivo</span><span class="cd-uni-ev-val text-muted">Carregando nome...</span></div>'
        }
        if (zona) html += '<div><span class="cd-uni-ev-lab">Zona</span><span class="cd-uni-ev-val">' + escHtml(zona) + '</span></div>'
        if (part) html += '<div><span class="cd-uni-ev-lab">Partição</span><span class="cd-uni-ev-val">' + escHtml(part) + '</span></div>'
        if (conta) html += '<div><span class="cd-uni-ev-lab">Conta</span><span class="cd-uni-ev-val">' + escHtml(conta) + '</span></div>'
        html += '</div></div>'
    }

    html += '</div></div></div></div>'
    return html
}

function renderTabela(rows, cols) {
    var h = '<div class="table-responsive"><table class="table table-sm table-bordered cd-uni-table">'
    h += '<thead><tr>'
    cols.forEach(function (c) { h += '<th>' + escHtml(c.l) + '</th>' })
    h += '</tr></thead><tbody>'
    rows.forEach(function (row) {
        h += '<tr>'
        cols.forEach(function (c) {
            var tdClass = ''
            if (c.fmt === 'audio') tdClass = ' class="cd-td-audio"'
            else if (c.fmt === 'transcript') tdClass = ' class="cd-td-transcript"'
            h += '<td' + tdClass + '>' + fmtCelula(row, c) + '</td>'
        })
        h += '</tr>'
    })
    h += '</tbody></table></div>'
    return h
}

function isTruthyFlag(v) {
    return v === true || v === 'true' || v === 1 || v === '1'
}

function fmtCelula(row, col) {
    if (col.fmt === 'audio') {
        if (!temAudioLigacao(row)) return '—'
        var url = '/ligacaoHistoricoAudio?id=' + encodeURIComponent(row.id)
        return '<div class="cd-audio-cell">' +
            '<audio controls preload="none" class="cd-audio-player" src="' + escAttr(url) + '"></audio></div>'
    }
    if (col.fmt === 'transcript') {
        return fmtTranscriptCell(row)
    }
    if (col.fmt === 'status_pendente') {
        if (isTruthyFlag(row.cancelado)) return 'Cancelado'
        if (isTruthyFlag(row.enviado)) return 'Enviado'
        return 'Pendente'
    }
    var v = row[col.k]
    if ((v == null || v === '') && col.alt) v = row[col.alt]
    if (col.fmt === 'bool') {
        if (isTruthyFlag(v)) return 'Sim'
        if (v === false || v === 'false' || v === 0 || v === '0') return 'Não'
        return v == null ? '' : escHtml(String(v))
    }
    if (col.fmt === 'sec') {
        if (v == null || v === '') return ''
        return escHtml(String(v)) + 's'
    }
    if (col.fmt === 'dt') return escHtml(fmtData(v))
    if (v == null) return ''
    return escHtml(String(v))
}

function fmtTranscriptCell(item) {
    var tr = String(item.analysis_transcript_summary || '').trim()
    if (!tr) return '—'
    if (!window._cdTranscriptMap) window._cdTranscriptMap = {}
    var key = 'tr-' + String(item.id || Math.random().toString(36).slice(2))
    window._cdTranscriptMap[key] = tr
    var preview = tr.replace(/\s+/g, ' ')
    return '<div class="cd-tr-cell">' +
        '<span class="cd-tr-preview" title="Clique para ver a transcrição completa">' + escHtml(preview) + '</span>' +
        '<button type="button" class="btn btn-sm btn-outline-primary btn-cd-transcript" data-tr-key="' +
        escAttr(key) + '" title="Ver transcrição"><i class="bi bi-arrows-fullscreen"></i></button></div>'
}

function abrirPopupTranscricao(texto) {
    if (!texto) return
    var $m = $('#cdTranscriptModal')
    if (!$m.length) {
        $('body').append(
            '<div class="modal fade" id="cdTranscriptModal" tabindex="-1">' +
            '<div class="modal-dialog modal-lg modal-dialog-scrollable">' +
            '<div class="modal-content bg-dark text-light">' +
            '<div class="modal-header border-secondary">' +
            '<h5 class="modal-title">Transcrição</h5>' +
            '<button type="button" class="btn-close btn-close-white" data-bs-dismiss="modal"></button>' +
            '</div>' +
            '<div class="modal-body"><pre id="cdTranscriptModalBody" class="cd-tr-modal-body mb-0"></pre></div>' +
            '<div class="modal-footer border-secondary">' +
            '<button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Fechar</button>' +
            '</div></div></div></div>'
        )
        $m = $('#cdTranscriptModal')
    }
    $('#cdTranscriptModalBody').text(texto)
    if (window.bootstrap && bootstrap.Modal) {
        bootstrap.Modal.getOrCreateInstance($m[0]).show()
    } else {
        $m.modal('show')
    }
}

function temAudioLigacao(item) {
    if (!item || !item.id || !item.conversation_id) return false
    var dur = Number(item.duration)
    if (!isFinite(dur) || dur <= 0) return false
    var st = String(item.status || '').toLowerCase().trim()
    if (st !== 'done') return false
    var tr = String(item.analysis_transcript_summary || '').trim()
    return !!tr
}

function fmtData(v) {
    if (v == null || v === '') return ''
    if (typeof v === 'number') {
        var d = new Date(v > 1e12 ? v : v * 1000)
        if (!isNaN(d.getTime())) return d.toLocaleString('pt-BR')
    }
    return String(v)
}

function escHtml(s) {
    return String(s == null ? '' : s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function escAttr(s) {
    return escHtml(s).replace(/'/g, '&#39;')
}

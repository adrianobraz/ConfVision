/* Relatorios Central de Disparos — motor compartilhado (auto-load + scroll infinito) */
var FP_CD_PAGE_SIZE = 100

var FP_CD_DEFS = {
    'eventos-pendentes': {
        endpoint: '/cdEventosPendentesListar',
        pdfTitle: 'EVENTOS PENDENTES',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'grupo', label: 'Grupo' },
            { key: 'ctiDescricao', label: 'Descrição' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'zonaUser', label: 'Zona' },
            { key: 'particao', label: 'Partição' },
            { key: 'agendadoPara', label: 'Agendado', fmt: 'dt' },
            { key: '_status', label: 'Status', fmt: 'status_pendente' }
        ]
    },
    'finalizados-robo': {
        endpoint: '/cdFinalizadosRoboListar',
        pdfTitle: 'FINALIZADOS POR ROBÔ',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'Usuario', label: 'Usuário' },
            { key: 'mensagem', label: 'Mensagem' },
            { key: 'ctiDescricao', label: 'Evento' }
        ]
    },
    'finalizados-bot': {
        endpoint: '/cdFinalizadosBotListar',
        pdfTitle: 'FINALIZADOS PELO BOT',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'status', label: 'Status' },
            { key: 'motivo_final', label: 'Motivo' },
            { key: 'acao_final', label: 'Ação' },
            { key: 'log_motivo', label: 'Log Motivo' },
            { key: 'erro', label: 'Erro' }
        ]
    },
    'ligacao-historico': {
        endpoint: '/cdLigacaoHistoricoListar',
        pdfTitle: 'LIGAÇÃO HISTÓRICO',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomecliente', label: 'Cliente' },
            { key: 'receiver', label: 'Telefone' },
            { key: 'duration', label: 'Duração', fmt: 'sec' },
            { key: 'status', label: 'Status' },
            { key: 'tipoevento', label: 'Evento' },
            { key: 'analysis_transcript_summary', label: 'Transcrição', fmt: 'transcript' },
            { key: '_audio', label: 'Áudio', fmt: 'audio' }
        ]
    },
    'sms-historico': {
        endpoint: '/cdSmsHistoricoListar',
        pdfTitle: 'SMS HISTÓRICO',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'whats_sender', label: 'Destino' },
            { key: 'message_status', label: 'Status', fmt: 'status_sms' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'ctiGrupo', label: 'Grupo' },
            { key: 'whats_text', label: 'Texto' }
        ]
    },
    'eventos-falhas': {
        endpoint: '/cdEventosFalhasListar',
        pdfTitle: 'EVENTOS COM FALHAS',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'ctiDescricao', label: 'Descrição' },
            { key: 'codigo', label: 'Código' },
            { key: 'idDispositivo', label: 'Dispositivo' },
            { key: 'particao', label: 'Partição' },
            { key: 'zonaUser', label: 'Zona' },
            { key: 'dataEntrada', label: 'Entrada' },
            { key: 'conta', label: 'Conta' }
        ]
    },
    'whatsapp-enviados': {
        endpoint: '/cdWhatsappEnviadosListar',
        pdfTitle: 'WHATSAPP ENVIADOS',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'whats_sender', label: 'Destino' },
            { key: 'ctiGrupo', label: 'Grupo' },
            { key: 'ctiDescricao', label: 'Evento' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'whats_text', label: 'Mensagem' }
        ]
    },
    'ligacoes': {
        endpoint: '/cdLigacoesListar',
        pdfTitle: 'LIGAÇÕES',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'ctiGrupo', label: 'Grupo' },
            { key: 'ctiDescricao', label: 'Evento' },
            { key: 'DispNome', label: 'Dispositivo' }
        ]
    },
    'sms': {
        endpoint: '/cdSmsListar',
        pdfTitle: 'SMS',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomeCliente', label: 'Cliente' },
            { key: 'whats_sender', label: 'Destino' },
            { key: 'ctiGrupo', label: 'Grupo' },
            { key: 'ctiDescricao', label: 'Evento' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'whats_text', label: 'Mensagem' }
        ]
    },
    'fila-envio': {
        endpoint: '/cdFilaEnvioListar',
        pdfTitle: 'EVENTOS PARA ENVIAR OU LIGAÇÃO',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomecliente', label: 'Cliente' },
            { key: 'numerowhatsapp', label: 'WhatsApp' },
            { key: 'tipoevento', label: 'Evento' },
            { key: 'enviartexto', label: 'Texto', fmt: 'bool' },
            { key: 'ligar', label: 'Ligar', fmt: 'bool' },
            { key: 'disparo', label: 'Disparado', fmt: 'bool' },
            { key: 'analise', label: 'Análise', fmt: 'bool' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'zona', label: 'Zona' }
        ]
    },
    'ligacao-erros': {
        endpoint: '/cdLigacaoErrosListar',
        pdfTitle: 'LIGAÇÕES COM ERROS E TENTATIVAS',
        columns: [
            { key: 'created_at', label: 'Data', fmt: 'dt' },
            { key: 'nomecliente', label: 'Cliente' },
            { key: 'telefone', label: 'Telefone' },
            { key: 'tipoevento', label: 'Evento' },
            { key: 'falha', label: 'Falha', fmt: 'bool' },
            { key: 'tentativas', label: 'Tentativas' },
            { key: 'atendido', label: 'Atendido', fmt: 'bool' },
            { key: 'exec', label: 'Exec', fmt: 'bool' },
            { key: 'dtUltimaTentativa', label: 'Última Tentativa', fmt: 'dt' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'nroErr', label: 'Nro Err' }
        ]
    },
    'fila-ligacao': {
        endpoint: '/cdFilaLigacaoListar',
        pdfTitle: 'FILAS DE LIGAÇÕES',
        hideCliente: true,
        columns: [
            { key: 'created_at', label: 'Data Fila', fmt: 'dt' },
            { key: 'nomecliente', label: 'Cliente' },
            { key: 'telefone', label: 'Telefone' },
            { key: 'tipoevento', label: 'Evento' },
            { key: 'tentativas', label: 'Tentativas' },
            { key: 'falha', label: 'Falha', fmt: 'bool' },
            { key: 'exec', label: 'Exec', fmt: 'bool' },
            { key: 'atendido', label: 'Atendido', fmt: 'bool' },
            { key: 'DispNome', label: 'Dispositivo' },
            { key: 'zona', label: 'Zona' }
        ]
    }
}

var listaClientes = []
var fpStart, fpEnd
var cdDef = null
var cdOffset = 0
var cdLoading = false
var cdFim = false
var cdTotal = 0

$(document).ready(function () {
    var key = window.FP_CD_REPORT || ''
    cdDef = FP_CD_DEFS[key]
    if (!cdDef) {
        boxErro('Relatório não configurado: ' + key)
        return
    }

    montarCabecalho()
    if (cdDef.hideCliente) $('#filtro-cliente-wrap').hide()

    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-filtrar').on('click', function () { recarregar(true) })
    $('#btn-imprimir').on('click', function () {
        gerarPdf(cdDef.pdfTitle || 'RELATÓRIO', '#tabCdRelatorio', 'landscape')
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

    $('#cd-table-wrap').on('scroll', onScrollTabela)

    // Download MP3 (evita baixar JSON de erro como arquivo)
    $('#tabCdBody').on('click', '.btn-cd-audio', function () {
        baixarAudioLigacao($(this).attr('data-id'), $(this))
    })

    $('#tabCdBody').on('click', '.btn-cd-transcript', function () {
        var key = $(this).attr('data-tr-key')
        abrirPopupTranscricao(window._cdTranscriptMap && window._cdTranscriptMap[key])
    })
    $('#tabCdBody').on('click', '.cd-tr-preview', function () {
        var key = $(this).closest('.cd-tr-cell').find('.btn-cd-transcript').attr('data-tr-key')
        abrirPopupTranscricao(window._cdTranscriptMap && window._cdTranscriptMap[key])
    })

    // Um player por vez + erro de reprodução
    document.addEventListener('play', function (e) {
        if (!e.target || !e.target.classList || !e.target.classList.contains('cd-audio-player')) return
        document.querySelectorAll('audio.cd-audio-player').forEach(function (a) {
            if (a !== e.target) a.pause()
        })
    }, true)
    document.addEventListener('error', function (e) {
        if (!e.target || !e.target.classList || !e.target.classList.contains('cd-audio-player')) return
        boxErro('Não foi possível reproduzir o áudio. Tente Baixar MP3.')
    }, true)

    if (!cdDef.hideCliente) carregarClientes()

    // Carrega automaticamente os últimos 100
    recarregar(false)
})

function baixarAudioLigacao(id, $btn) {
    if (!id) return
    var original = $btn.html()
    $btn.prop('disabled', true).html('<i class="bi bi-hourglass-split"></i>')
    fetch('/ligacaoHistoricoAudio?id=' + encodeURIComponent(id) + '&dl=1', {
        method: 'GET',
        credentials: 'same-origin'
    }).then(function (resp) {
        var ct = (resp.headers.get('content-type') || '').toLowerCase()
        if (!resp.ok || ct.indexOf('application/json') >= 0) {
            return resp.text().then(function (txt) {
                var msg = 'Erro ao baixar o áudio'
                try {
                    var j = JSON.parse(txt)
                    if (j && (j.erro || j.message || j.status)) {
                        msg = j.erro || j.message || j.status
                    }
                } catch (e) {
                    if (txt) msg = String(txt).slice(0, 200)
                }
                throw new Error(msg)
            })
        }
        var disposition = resp.headers.get('content-disposition') || ''
        var filename = 'ligacao-' + id + '.mp3'
        var m = /filename=\"?([^\";]+)\"?/i.exec(disposition)
        if (m && m[1]) filename = m[1]
        return resp.blob().then(function (blob) {
            var url = URL.createObjectURL(blob)
            var a = document.createElement('a')
            a.href = url
            a.download = filename
            document.body.appendChild(a)
            a.click()
            a.remove()
            setTimeout(function () { URL.revokeObjectURL(url) }, 2000)
        })
    }).catch(function (err) {
        boxErro(err && err.message ? err.message : 'Erro ao baixar o áudio')
    }).finally(function () {
        $btn.prop('disabled', false).html(original)
    })
}

function montarCabecalho() {
    var tr = $('<tr></tr>')
    cdDef.columns.forEach(function (c) {
        tr.append('<th>' + escHtml(c.label) + '</th>')
    })
    $('#tabCdHead').empty().append(tr)
}

function carregarClientes() {
    $.ajax({
        url: '/carregarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function () {
        boxErro('Erro ao carregar os clientes')
    }).done(function (r) {
        listaClientes = r.dados || []
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
            '<li class="fp-combo-item" data-id="' + idCli + '" data-nome="' +
            escAttr(i.nome || '') + '">' + escHtml(i.nome || '') + '</li>'
        )
    })
    lista.removeClass('d-none')
}

function toMs(val) {
    if (!val) return 0
    // flatpickr: prioriza selectedDates
    if (val === 'start' && fpStart && fpStart.selectedDates && fpStart.selectedDates[0]) {
        return fpStart.selectedDates[0].getTime()
    }
    if (val === 'end' && fpEnd && fpEnd.selectedDates && fpEnd.selectedDates[0]) {
        return fpEnd.selectedDates[0].getTime()
    }
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
    recarregar(false)
}

function montarPayload() {
    var idCli = $('#id-cliente').val() || '0'
    var nomeCli = ''
    if (idCli && idCli !== '0') {
        nomeCli = ($('#busca-cliente').val() || '').trim()
    }
    return {
        idFranqueado: localStorage.getItem('idFranqueado'),
        idCliente: idCli,
        nomeCliente: nomeCli,
        dataInicioMs: lerDataMs('start'),
        dataFimMs: lerDataMs('end'),
        limit: FP_CD_PAGE_SIZE,
        offset: cdOffset
    }
}

function recarregar(usarFiltro) {
    // usarFiltro: true = usuário clicou Filtrar; false = carga inicial / limpar
    if (!usarFiltro) {
        // mantém filtros se já preenchidos, mas reinicia lista
    }
    cdOffset = 0
    cdFim = false
    cdTotal = 0
    $('#tabCdBody').empty()
    $('#cd-fim').addClass('d-none')
    $('#qtd-registros').text('Total: 0')
    carregarMais()
}

function onScrollTabela() {
    var el = document.getElementById('cd-table-wrap')
    if (!el || cdLoading || cdFim) return
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 80) {
        carregarMais()
    }
}

function carregarMais() {
    if (cdLoading || cdFim || !cdDef) return
    cdLoading = true
    $('#cd-loading').removeClass('d-none')

    var payload = montarPayload()
    $.ajax({
        url: cdDef.endpoint,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (xhr) {
        cdLoading = false
        $('#cd-loading').addClass('d-none')
        var msg = 'Erro ao carregar o relatório'
        try {
            var j = JSON.parse(xhr.responseText)
            if (j && (j.erro || j.message)) msg = j.erro || j.message
        } catch (e) { /* ignore */ }
        boxErro(msg)
    }).done(function (r) {
        cdLoading = false
        $('#cd-loading').addClass('d-none')
        var lista = r.dados || []
        appendTabela(lista)
        cdOffset += lista.length
        cdTotal += lista.length
        $('#qtd-registros').text('Total: ' + cdTotal)
        if (lista.length < FP_CD_PAGE_SIZE) {
            cdFim = true
            if (cdTotal > 0) $('#cd-fim').removeClass('d-none')
        }
    })
}

function isTruthyFlag(v) {
    return v === true || v === 'true' || v === 1 || v === '1'
}

function statusSmsPt(status) {
    var s = String(status || '').trim()
    if (!s) return ''
    var map = {
        queued: 'Na fila',
        sent: 'Enviado',
        delivered: 'Entregue',
        failed: 'Falhou',
        undeliverable: 'Não entregue',
        rejected: 'Rejeitado',
        processed: 'Em processamento',
        error: 'Erro'
    }
    var key = s.toLowerCase()
    return map[key] || s
}

function temAudioLigacao(item) {
    if (!item || !item.id || !item.conversation_id) return false
    var dur = Number(item.duration)
    if (!isFinite(dur) || dur <= 0) return false
    var st = String(item.status || '').toLowerCase().trim()
    if (st !== 'done') return false
    var tr = String(item.analysis_transcript_summary || '').trim()
    if (!tr) return false
    return true
}

function fmtValor(item, col) {
    if (col.fmt === 'audio') {
        if (!temAudioLigacao(item)) return '—'
        var id = escAttr(item.id)
        var url = '/ligacaoHistoricoAudio?id=' + encodeURIComponent(item.id)
        return '<div class="cd-audio-cell">' +
            '<audio controls preload="none" class="cd-audio-player" src="' + url + '"></audio>' +
            '<button type="button" class="btn btn-sm btn-outline-primary btn-cd-audio" data-id="' + id + '">' +
            '<i class="bi bi-download"></i> MP3</button></div>'
    }
    if (col.fmt === 'transcript') {
        return fmtTranscriptCell(item)
    }
    if (col.fmt === 'status_pendente') {
        if (isTruthyFlag(item.cancelado)) return 'Cancelado'
        if (isTruthyFlag(item.enviado)) return 'Enviado'
        return 'Pendente'
    }
    if (col.fmt === 'status_sms') {
        return statusSmsPt(item[col.key] || item.message_status)
    }
    var v = item[col.key]
    if (col.fmt === 'bool') {
        if (isTruthyFlag(v)) return 'Sim'
        if (v === false || v === 'false' || v === 0 || v === '0') return 'Não'
        return v == null ? '' : String(v)
    }
    if (col.fmt === 'sec') {
        if (v == null || v === '') return ''
        return v + 's'
    }
    if (col.fmt === 'dt') return fmtData(v)
    if (v == null) return ''
    return String(v)
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

function fmtData(v) {
    if (v == null || v === '') return ''
    if (typeof v === 'number') {
        var d = new Date(v > 1e12 ? v : v * 1000)
        if (!isNaN(d.getTime())) {
            return d.toLocaleString('pt-BR')
        }
    }
    return String(v)
}

function appendTabela(lista) {
    var tbody = $('#tabCdBody')
    lista.forEach(function (item) {
        var tr = $('<tr></tr>')
        cdDef.columns.forEach(function (col) {
            var raw = fmtValor(item, col)
            if (col.fmt === 'audio' || col.fmt === 'transcript') {
                var tdClass = col.fmt === 'audio' ? 'cd-td-audio' : 'cd-td-transcript'
                tr.append('<td class="' + tdClass + '">' + raw + '</td>')
            } else {
                tr.append('<td class="text-start" style="max-width:280px;white-space:normal;">' + escHtml(raw) + '</td>')
            }
        })
        tbody.append(tr)
    })
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

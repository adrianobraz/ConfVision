const CVG_DIAS = [
    { n: 1, label: 'Segunda' },
    { n: 2, label: 'Terça' },
    { n: 3, label: 'Quarta' },
    { n: 4, label: 'Quinta' },
    { n: 5, label: 'Sexta' },
    { n: 6, label: 'Sábado' },
    { n: 7, label: 'Domingo' }
]

let gradeListaItens = []
let gradeClientesCache = []
let gradeDispositivosCache = {}
let gradeModoEdicao = false
let gradeClienteNomeSelecionado = ''

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    montarGradeVazia()
    bindGradeEventos()
    carregarClientesGrade().then(function () {
        carregarListaGrades()
        if (ConfVisionUrls.ehCliente()) {
            prepararClienteLogado()
        } else {
            limparFormularioGrade(true)
        }
    })
})

function bindGradeEventos() {
    $('#grade-lista-filtro').on('input', renderizarListaGrades)
    $('#btn-grade-nova').on('click', cancelarGrade)
    $('#btn-grade-cancelar').on('click', cancelarGrade)
    $('#btn-grade-salvar').on('click', salvarGrade)
    $('#btn-grade-horario-comercial').on('click', preencherHorarioComercial)
    $('#sel-grade-dispositivo').on('change', onDispositivoGradeChange)

    $('#busca-cliente-grade').on('input', function () {
        $('#id-cliente-grade').val('')
        gradeClienteNomeSelecionado = ''
        filtrarClientesGrade()
    })
    $('#busca-cliente-grade').on('focus', function () {
        const id = $('#id-cliente-grade').val()
        const val = $(this).val().trim()
        if (id && val === gradeClienteNomeSelecionado) return
        if (val) filtrarClientesGrade()
    })
    $('#lista-clientes-grade').on('mousedown', '.cv-combo-item', function (e) {
        e.preventDefault()
    })
    $('#lista-clientes-grade').on('click', '.cv-combo-item', function () {
        const id = $(this).attr('data-id')
        const nome = $(this).attr('data-nome')
        $('#id-cliente-grade').val(id)
        $('#busca-cliente-grade').val(nome)
        gradeClienteNomeSelecionado = nome
        fecharListaClientesGrade()
        $('#busca-cliente-grade').blur()
        onClienteGradeSelecionado(id)
    })
    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-cliente-grade, #lista-clientes-grade').length) {
            fecharListaClientesGrade()
        }
    })

    $('#grade-lista-body').on('change', '.cv-grade-toggle-ativa', onToggleGradeLista)
    $('#grade-lista-body').on('click', '.btn-grade-editar', onEditarGradeLista)
}

function montarGradeVazia() {
    const $body = $('#tab-grade-body')
    $body.empty()
    CVG_DIAS.forEach(function (d) {
        $body.append(`
            <tr data-dia="${d.n}">
                <td class="cv-grade-dia">${d.label}</td>
                <td><input type="time" class="form-control form-control-sm cv-grade-time" data-dia="${d.n}" data-acao="desativar" /></td>
                <td><input type="time" class="form-control form-control-sm cv-grade-time" data-dia="${d.n}" data-acao="ativar" /></td>
            </tr>
        `)
    })
}

function carregarClientesGrade() {
    gradeClientesCache = []

    if (ConfVisionUrls.ehCliente()) {
        const id = ConfVisionUrls.idCliente()
        const nome = localStorage.getItem('nomeUsuario') || 'Meu local'
        if (id) gradeClientesCache.push({ id: id, nome: nome })
        return $.when()
    }

    const idFra = ConfVisionUrls.idFranqueado()
    if (!idFra) {
        boxErro('idFranqueado ausente — faça login novamente')
        return $.when()
    }

    return $.ajax({
        url: '/api/clientes',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idFranqueado: idFra })
    }).fail(function (xhr) {
        boxErro('Erro ao carregar clientes')
        console.error('[grade] clientes', xhr && xhr.status, xhr && xhr.responseText)
    }).done(function (r) {
        const lista = (r && r.dados) ? r.dados : (Array.isArray(r) ? r : [])
        lista.forEach(function (c) {
            const id = String(c.idCliente || c.id_cliente || c.id || '')
            const nome = String(c.nome || c.nomeCliente || c.Nome || id)
            if (id) gradeClientesCache.push({ id: id, nome: nome })
        })
    })
}

function prepararClienteLogado() {
    const id = ConfVisionUrls.idCliente()
    const cli = gradeClientesCache.find(function (c) { return c.id === id })
    if (!cli) return
    gradeClienteNomeSelecionado = cli.nome
    $('#busca-cliente-grade').val(cli.nome).prop('readonly', true)
    $('#id-cliente-grade').val(cli.id)
    onClienteGradeSelecionado(cli.id)
}

function nomeClienteGrade(id) {
    const c = gradeClientesCache.find(function (x) { return x.id === String(id) })
    return c ? c.nome : String(id)
}

function fecharListaClientesGrade() {
    $('#lista-clientes-grade').addClass('d-none').empty()
}

function filtrarClientesGrade() {
    const termo = $('#busca-cliente-grade').val().trim().toLowerCase()
    const $lista = $('#lista-clientes-grade')
    $lista.empty()

    if (!termo) {
        $lista.addClass('d-none')
        return
    }

    const filtrados = gradeClientesCache.filter(function (c) {
        return c.nome.toLowerCase().indexOf(termo) >= 0 || c.id.indexOf(termo) >= 0
    })

    if (!filtrados.length) {
        $lista.append('<li class="cv-combo-empty">Nenhum cliente encontrado</li>')
    } else {
        filtrados.slice(0, 40).forEach(function (c) {
            $('<li class="cv-combo-item"></li>')
                .attr('data-id', c.id)
                .attr('data-nome', c.nome)
                .text(c.nome)
                .appendTo($lista)
        })
    }
    $lista.removeClass('d-none')
}

function onClienteGradeSelecionado(idCli) {
    $('#grade-cadastro-dispositivo').removeClass('d-none')
    $('#grade-cadastro-horarios').addClass('d-none')
    $('#sel-grade-dispositivo').prop('disabled', !idCli).empty()
        .append('<option value="">Todos os dispositivos</option>')

    if (!idCli) return

    if (gradeDispositivosCache[idCli]) {
        preencherComboDispositivos(idCli, gradeDispositivosCache[idCli])
        return
    }

    $.ajax({
        url: '/api/dispositivos',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idCliente: idCli })
    }).fail(function () {
        boxErro('Erro ao carregar dispositivos')
    }).done(function (r) {
        const lista = (r && r.dados) ? r.dados : []
        const normalizada = []
        lista.forEach(function (d) {
            const id = String(d.idDispositivo || d.id_dispositivo || '')
            const nome = String(d.nome || d.Nome || id)
            if (id) normalizada.push({ id: id, nome: nome })
        })
        gradeDispositivosCache[idCli] = normalizada
        preencherComboDispositivos(idCli, normalizada)
    })
}

function preencherComboDispositivos(idCli, lista) {
    const $sel = $('#sel-grade-dispositivo')
    $sel.empty().append('<option value="">Todos os dispositivos</option>')
    lista.forEach(function (d) {
        $sel.append(`<option value="${escHtml(d.id)}">${escHtml(d.nome)}</option>`)
    })
    if (gradeModoEdicao && $sel.data('edit-disp') !== undefined) {
        $sel.val(String($sel.data('edit-disp') || ''))
        $sel.removeData('edit-disp')
        onDispositivoGradeChange()
    }
}

function nomeDispositivoGrade(idCli, idDisp) {
    if (!idDisp) return 'Todos os dispositivos'
    const lista = gradeDispositivosCache[idCli] || []
    const d = lista.find(function (x) { return x.id === String(idDisp) })
    return d ? d.nome : String(idDisp)
}

function onDispositivoGradeChange() {
    const idCli = $('#id-cliente-grade').val()
    if (!idCli) return
    $('#grade-cadastro-horarios').removeClass('d-none')
    carregarGradeFormulario(idCli, $('#sel-grade-dispositivo').val() || '')
}

function cancelarGrade() {
    limparFormularioGrade(true)
    $('#busca-cliente-grade').focus()
}

function limparFormularioGrade(nova) {
    gradeModoEdicao = !nova
    gradeClienteNomeSelecionado = ''
    fecharListaClientesGrade()
    $('#grade-form-titulo').text(nova ? 'Nova grade' : 'Editar grade')
    $('#busca-cliente-grade').val('').prop('readonly', false)
    $('#id-cliente-grade').val('')
    $('#sel-grade-dispositivo').empty().append('<option value="">Todos os dispositivos</option>')
    $('#grade-cadastro-dispositivo').addClass('d-none')
    $('#grade-cadastro-horarios').addClass('d-none')
    $('.cv-grade-time').val('')
}

function gradeAtivaDoEscopo(idCli, idDisp) {
    const item = gradeListaItens.find(function (it) {
        return String(it.id_cliente) === String(idCli) &&
            String(it.id_dispositivo || '') === String(idDisp || '')
    })
    return item ? !!item.grade_ativa : true
}

function carregarListaGrades() {
    const $body = $('#grade-lista-body')
    $body.html('<tr class="cv-grade-list-empty"><td colspan="4">Carregando...</td></tr>')

    let url = '/api/grades'
    const idFra = ConfVisionUrls.idFranqueado()
    if (idFra) url += '?id_franqueado=' + encodeURIComponent(idFra)

    $.ajax({ url: url, method: 'GET' }).fail(function (xhr) {
        let msgLista = 'Erro ao carregar lista'
        if (xhr && xhr.status === 404) {
            msgLista = 'Lista indisponível — faça push da API cvg_grade_list no Xano e rebuild do ConfVision'
            boxAdvertenciaAuto('API GET /api/grades (cvg_grade_list) não encontrada. Push Xano + rebuild ConfVision.')
        } else if (xhr && xhr.status === 502) {
            msgLista = 'XANO_CVG_BASE_URL não configurado'
            boxAdvertenciaAuto('XANO_CVG_BASE_URL não configurado no ConfVision.')
        }
        $body.html('<tr class="cv-grade-list-empty"><td colspan="4">' + escHtml(msgLista) + '</td></tr>')
        console.error('[grade] lista', xhr && xhr.status, xhr && xhr.responseText)
    }).done(function (r) {
        const dados = r.dados || r
        gradeListaItens = Array.isArray(dados.itens) ? dados.itens : []
        if (ConfVisionUrls.ehCliente()) {
            const idCli = ConfVisionUrls.idCliente()
            gradeListaItens = gradeListaItens.filter(function (it) {
                return String(it.id_cliente) === String(idCli)
            })
        }
        precarregarDispositivosLista().always(renderizarListaGrades)
    })
}

function precarregarDispositivosLista() {
    const ids = []
    gradeListaItens.forEach(function (it) {
        const id = String(it.id_cliente || '')
        if (id && !gradeDispositivosCache[id] && ids.indexOf(id) < 0) ids.push(id)
    })
    if (!ids.length) return $.when()

    const reqs = ids.map(function (idCli) {
        return $.ajax({
            url: '/api/dispositivos',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ idCliente: idCli })
        }).done(function (r) {
            const lista = (r && r.dados) ? r.dados : []
            const normalizada = []
            lista.forEach(function (d) {
                const id = String(d.idDispositivo || d.id_dispositivo || '')
                const nome = String(d.nome || d.Nome || id)
                if (id) normalizada.push({ id: id, nome: nome })
            })
            gradeDispositivosCache[idCli] = normalizada
        })
    })
    return $.when.apply($, reqs)
}

function renderizarListaGrades() {
    const filtro = $('#grade-lista-filtro').val().trim().toLowerCase()
    const $body = $('#grade-lista-body')
    $body.empty()

    const visiveis = gradeListaItens.filter(function (it) {
        if (!filtro) return true
        const nc = nomeClienteGrade(it.id_cliente).toLowerCase()
        const nd = nomeDispositivoGrade(it.id_cliente, it.id_dispositivo).toLowerCase()
        return nc.indexOf(filtro) >= 0 || nd.indexOf(filtro) >= 0 ||
            String(it.id_cliente || '').indexOf(filtro) >= 0 ||
            String(it.id_dispositivo || '').indexOf(filtro) >= 0
    })

    if (!visiveis.length) {
        $body.append('<tr class="cv-grade-list-empty"><td colspan="4">Nenhuma grade configurada</td></tr>')
        return
    }

    visiveis.forEach(function (it) {
        const idCli = String(it.id_cliente || '')
        const idDisp = String(it.id_dispositivo || '')
        const ativa = !!it.grade_ativa
        const key = escHtml(idCli + '|' + idDisp)
        $body.append(`
            <tr data-key="${key}" data-cliente="${escHtml(idCli)}" data-dispositivo="${escHtml(idDisp)}">
                <td>${escHtml(nomeClienteGrade(idCli))}</td>
                <td>${escHtml(nomeDispositivoGrade(idCli, idDisp))}</td>
                <td class="text-center">
                    <input type="checkbox" class="form-check-input cv-grade-toggle-ativa" ${ativa ? 'checked' : ''}
                        aria-label="Grade ativa" />
                </td>
                <td class="text-end">
                    <button type="button" class="cv-btn-icon btn-grade-editar" title="Editar">
                        <i class="bi bi-pencil"></i>
                    </button>
                </td>
            </tr>
        `)
    })
}

function onToggleGradeLista() {
    const $chk = $(this)
    const $tr = $chk.closest('tr')
    const idCli = $tr.attr('data-cliente')
    const idDisp = $tr.attr('data-dispositivo') || ''
    const gradeAtiva = $chk.is(':checked')

    $chk.prop('disabled', true)
    $.ajax({
        url: '/api/grade-escopo-ativa',
        method: 'PATCH',
        contentType: 'application/json',
        data: JSON.stringify({
            id_cliente: idCli,
            id_dispositivo: idDisp,
            grade_ativa: gradeAtiva
        })
    }).fail(function (xhr) {
        $chk.prop('checked', !gradeAtiva)
        boxErro('Erro ao alterar status da grade')
        console.error('[grade] toggle', xhr && xhr.status, xhr && xhr.responseText)
    }).done(function () {
        const item = gradeListaItens.find(function (it) {
            return String(it.id_cliente) === String(idCli) &&
                String(it.id_dispositivo || '') === String(idDisp)
        })
        if (item) item.grade_ativa = gradeAtiva
    }).always(function () {
        $chk.prop('disabled', false)
    })
}

function onEditarGradeLista() {
    const $tr = $(this).closest('tr')
    const idCli = $tr.attr('data-cliente')
    const idDisp = $tr.attr('data-dispositivo') || ''

    gradeModoEdicao = true
    $('#grade-form-titulo').text('Editar grade')
    $('#id-cliente-grade').val(idCli)
    gradeClienteNomeSelecionado = nomeClienteGrade(idCli)
    $('#busca-cliente-grade').val(gradeClienteNomeSelecionado)
    fecharListaClientesGrade()
    if (ConfVisionUrls.ehCliente()) {
        $('#busca-cliente-grade').prop('readonly', true)
    }

    $('#grade-cadastro-dispositivo').removeClass('d-none')
    $('#sel-grade-dispositivo').data('edit-disp', idDisp)
    onClienteGradeSelecionado(idCli)
}

function carregarGradeFormulario(idCli, idDisp) {
    let url = '/api/grade-cliente/' + encodeURIComponent(idCli)
    if (idDisp) url += '?id_dispositivo=' + encodeURIComponent(idDisp)

    $.ajax({ url: url, method: 'GET' }).fail(function (xhr) {
        if (xhr && xhr.status === 404) {
            limparCamposHorarios()
            return
        }
        boxErro('Erro ao carregar grade')
        console.error('[grade] carregar', xhr && xhr.status, xhr && xhr.responseText)
    }).done(function (r) {
        const dados = r.dados || r
        aplicarSlotsNaTabela(Array.isArray(dados.slots) ? dados.slots : [], idDisp)
    })
}

function limparCamposHorarios() {
    $('.cv-grade-time').val('')
}

function preencherHorarioComercial() {
    CVG_DIAS.forEach(function (d) {
        const comercial = d.n <= 5
        $(`.cv-grade-time[data-dia="${d.n}"][data-acao="desativar"]`).val(comercial ? '08:00' : '')
        $(`.cv-grade-time[data-dia="${d.n}"][data-acao="ativar"]`).val(comercial ? '18:00' : '')
    })
}

function aplicarSlotsNaTabela(slots, idDisp) {
    $('.cv-grade-time').val('')
    slots.forEach(function (s) {
        if (s.ativo === false) return
        const sd = String(s.id_dispositivo || '')
        if (idDisp && sd && sd !== idDisp) return
        if (!idDisp && sd) return
        const dia = Number(s.dia_semana)
        const acao = String(s.acao || '')
        const hora = String(s.hora || '').substring(0, 5)
        if (!dia || !hora) return
        $(`.cv-grade-time[data-dia="${dia}"][data-acao="${acao}"]`).val(hora)
    })
}

function coletarSlotsDoFormulario() {
    const slots = []
    $('.cv-grade-time').each(function () {
        const $el = $(this)
        const hora = String($el.val() || '').trim()
        if (!hora) return
        slots.push({
            dia_semana: Number($el.data('dia')),
            hora: hora,
            acao: String($el.data('acao')),
            ativo: true
        })
    })
    return slots
}

function gradeMsgErroAjax(xhr, padrao) {
    let msg = padrao
    if (xhr && xhr.responseText) {
        try {
            const j = JSON.parse(xhr.responseText)
            msg = j.message || j.error || j.msg || j.erro || xhr.responseText
        } catch (e) {
            if (xhr.responseText.length < 200) msg = xhr.responseText
        }
    }
    if (xhr && xhr.status) msg += ' (' + xhr.status + ')'
    return msg
}

function salvarGrade() {
    const idCli = $('#id-cliente-grade').val()
    if (!idCli) {
        boxAdvertenciaAuto('Selecione um cliente na lista')
        $('#busca-cliente-grade').focus()
        return
    }

    const idDisp = $('#sel-grade-dispositivo').val() || ''
    const slots = coletarSlotsDoFormulario()
    if (!slots.length) {
        boxAdvertenciaAuto('Informe ao menos um horário')
        return
    }

    const payload = {
        grade_ativa: gradeAtivaDoEscopo(idCli, idDisp),
        id_dispositivo_escopo: idDisp,
        slots: slots
    }

    const idFra = ConfVisionUrls.idFranqueado()
    if (idFra) payload.id_franqueado = idFra

    $('#btn-grade-salvar, #btn-grade-cancelar').prop('disabled', true)
    $.ajax({
        url: '/api/grade-cliente/' + encodeURIComponent(idCli),
        method: 'PUT',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (xhr) {
        boxErro(gradeMsgErroAjax(xhr, 'Erro ao salvar grade'))
        console.error('[grade] salvar', xhr && xhr.status, xhr && xhr.responseText)
    }).done(function () {
        boxSucessoAuto('Grade salva com sucesso')
        carregarListaGrades()
        carregarGradeFormulario(idCli, idDisp)
    }).always(function () {
        $('#btn-grade-salvar, #btn-grade-cancelar').prop('disabled', false)
    })
}

function escHtml(s) {
    return String(s || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

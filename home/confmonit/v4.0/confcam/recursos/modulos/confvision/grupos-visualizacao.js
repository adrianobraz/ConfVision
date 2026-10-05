const GRUPO_CLI_PAGE = 50
const GRUPO_CAM_MAX_RENDER = 120
const GRUPO_BUSCA_MS = 300

const GRUPO_TABS = ['clientes', 'cameras', 'ordem']
let ordemCameras = []
let clientesAcl = {}
let cacheCameras = []
let mapaNomesCliente = {}
let mapaClientesCarregado = false
let camerasCarregadas = false
let buscaClientesTimer = null
let buscaCamerasTimer = null
let clientesBuscaOffset = 0
let clientesBuscaTermo = ''
let clientesBuscaHasMore = false
let clientesBuscaCarregando = false

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    if (ConfVisionUrls.ehCliente()) {
        window.location = '/mosaicos'
        return
    }
    initEditorGrupo()
    $('#btn-novo-grupo').on('click', abrirNovoGrupo)
    $('#btn-salvar-grupo, #btn-salvar-grupo-mobile').on('click', salvarGrupo)
    $('#btn-cancelar-grupo, #btn-cancelar-grupo-mobile, #btn-grupo-voltar').on('click', fecharEditorGrupo)
    $('#grupo-tabs').on('click', '[data-grupo-tab]', function () {
        ativarGrupoTab($(this).attr('data-grupo-tab'))
    })
    $(document).on('keydown.grupoEditor', function (e) {
        if (e.key === 'Escape' && !$('#grupo-editor').hasClass('d-none')) {
            fecharEditorGrupo()
        }
    })
    $('#grupo-busca-clientes').on('input', onBuscaClientesInput)
    $('#grupo-busca-cameras').on('input', onBuscaCamerasInput)
    $('#grupo-clientes-mais').on('click', function () {
        buscarClientes(false)
    })
    $('#grupo-clientes-resultados').on('change', 'input[type=checkbox]', function () {
        const id = String($(this).val())
        const nome = String($(this).data('nome') || '')
        if ($(this).is(':checked')) {
            clientesAcl[id] = { id: id, nome: nome || ('Cliente ' + id) }
            mapaNomesCliente[id] = clientesAcl[id].nome
        } else {
            delete clientesAcl[id]
        }
        renderChipsClientes()
    })
    $('#grupo-clientes-selecionados').on('click', '.cv-grupo-chip-remove', function (e) {
        e.preventDefault()
        const id = String($(this).closest('.cv-grupo-chip').data('id') || '')
        delete clientesAcl[id]
        renderChipsClientes()
        renderResultadosClientesUltimos()
    })
    $('#grupo-cameras').on('change', 'input[type=checkbox]', function () {
        atualizarContadorCameras()
        renderPainelOrdem()
    })
    $('#grupo-ordem').on('click', '.cv-grupo-ordem-btn', moverCameraOrdem)
    carregarGrupos()
})

function initEditorGrupo() {
    ativarGrupoTab('clientes')
}

function abrirEditorGrupo() {
    $('#grupo-editor').removeClass('d-none').attr('aria-hidden', 'false')
    $('body').addClass('cv-grupo-editor-open')
    ativarGrupoTab('clientes')
    setTimeout(function () {
        const $nome = $('#grupo-nome')
        if ($nome.length) $nome.trigger('focus')
    }, 50)
}

function fecharEditorGrupo() {
    $('#grupo-editor').addClass('d-none').attr('aria-hidden', 'true')
    $('body').removeClass('cv-grupo-editor-open')
}

function ativarGrupoTab(tab) {
    let t = String(tab || 'clientes')
    if (GRUPO_TABS.indexOf(t) < 0) t = 'clientes'

    $('#grupo-tabs .nav-link').removeClass('active')
    $('#grupo-tabs [data-grupo-tab="' + t + '"]').addClass('active')

    $('.cv-grupo-tab-panel').addClass('d-none')
    $('.cv-grupo-tab-panel[data-grupo-panel="' + t + '"]').removeClass('d-none')
}

function atualizarBadgeClientes() {
    $('#grupo-badge-clientes').text(Object.keys(clientesAcl).length)
}

function atualizarBadgeCameras() {
    const set = camerasSelecionadasSet()
    const n = Object.keys(set).length
    $('#grupo-badge-cameras').text(n)
    $('#grupo-badge-ordem').text(n)
    $('#grupo-contador-cameras').text(n)
}

function carregarMapaClientes() {
    if (mapaClientesCarregado) {
        sincronizarNomesAcl()
        return $.Deferred().resolve().promise()
    }
    return $.ajax({
        url: '/api/clientes',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idFranqueado: ConfVisionUrls.idFranqueado() })
    }).then(function (r) {
        const lista = (r && r.dados) || r || []
        if (Array.isArray(lista)) {
            lista.forEach(function (c) {
                const id = String(c.id || c.idCliente || c.id_cliente || '')
                const nome = c.nome || c.nomeCliente || c.nick || c.razaoSocial || c.fantasia || ''
                if (id && nome) mapaNomesCliente[id] = nome
            })
        }
        mapaClientesCarregado = true
        sincronizarNomesAcl()
    }, function () {
        mapaClientesCarregado = true
    })
}

function sincronizarNomesAcl() {
    Object.keys(clientesAcl).forEach(function (id) {
        if (mapaNomesCliente[id]) clientesAcl[id].nome = mapaNomesCliente[id]
    })
}

function carregarCamerasPromise() {
    const d = $.Deferred()
    carregarCamerasSeNecessario(function () { d.resolve() })
    return d.promise()
}

function prepararEditorGrupo(cb) {
    $.when(carregarMapaClientes(), carregarCamerasPromise()).done(function () {
        if (typeof cb === 'function') cb()
    })
}

function escHtml(v) {
    return String(v == null ? '' : v)
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

function erroApi(xhr) {
    try {
        const j = JSON.parse(xhr.responseText || '{}')
        return j.status || j.erro || j.message || xhr.statusText
    } catch (e) {
        return xhr.statusText || 'Erro'
    }
}

function nomeCliente(id) {
    id = String(id || '')
    return mapaNomesCliente[id] || ('Cliente ' + id)
}

function carregarGrupos() {
    $.get('/api/grupos-visualizacao')
        .done(function (r) {
            renderListaGrupos((r && r.dados) || [])
        })
        .fail(function (xhr) {
            $('#grupos-lista').html('<div class="cv-empty">' + escHtml(erroApi(xhr)) + '</div>')
        })
}

function renderListaGrupos(lista) {
    const $box = $('#grupos-lista').empty()
    if (!lista.length) {
        $box.html('<div class="cv-empty">Nenhum grupo cadastrado. Clique em <strong>Novo grupo</strong>.</div>')
        return
    }
    lista.forEach(function (g) {
        const id = g.id
        const ativo = g.ativo !== false
        const total = g.total_cameras || 0
        $box.append(`
            <article class="cv-grupo-card ${ativo ? '' : 'is-inativo'}" data-id="${id}">
                <div class="cv-grupo-card-body">
                    <h3 class="cv-grupo-card-title">${escHtml(g.nome || 'Grupo #' + id)}</h3>
                    <p class="cv-grupo-card-desc">${escHtml(g.descricao || '')}</p>
                    <div class="cv-grupo-card-meta">
                        <span><i class="bi bi-camera-video"></i> ${total} câmera${total === 1 ? '' : 's'}</span>
                        <span class="cv-badge ${ativo ? 'cv-badge-ok' : 'cv-badge-muted'}">${ativo ? 'Ativo' : 'Inativo'}</span>
                    </div>
                </div>
                <div class="cv-grupo-card-actions">
                    <a href="/mosaicos/${id}" class="cv-btn-ghost cv-btn-sm" title="Abrir mosaico"><i class="bi bi-grid-3x3-gap"></i></a>
                    <button type="button" class="cv-btn-ghost cv-btn-sm btn-editar-grupo" data-id="${id}" title="Editar"><i class="bi bi-pencil"></i></button>
                    <button type="button" class="cv-btn-ghost cv-btn-sm btn-excluir-grupo" data-id="${id}" title="Excluir"><i class="bi bi-trash"></i></button>
                </div>
            </article>
        `)
    })
    $box.find('.btn-editar-grupo').on('click', function () {
        editarGrupo($(this).data('id'))
    })
    $box.find('.btn-excluir-grupo').on('click', function () {
        excluirGrupo($(this).data('id'))
    })
}

function resetModalForm() {
    clientesAcl = {}
    ordemCameras = []
    clientesBuscaOffset = 0
    clientesBuscaTermo = ''
    clientesBuscaHasMore = false
    $('#grupo-busca-clientes').val('')
    $('#grupo-busca-cameras').val('')
    renderChipsClientes()
    $('#grupo-clientes-resultados').html('<p class="cv-grupo-hint-empty">Digite ao menos 2 caracteres para buscar clientes.</p>')
    atualizarBadgeClientes()
    atualizarBadgeCameras()
    $('#grupo-clientes-status').text('')
    $('#grupo-clientes-mais').addClass('d-none')
    $('#grupo-cameras-status').text('')
    $('#grupo-cameras').html('<p class="cv-grupo-hint-empty">Carregando câmeras…</p>')
    $('#grupo-ordem').html('<p class="cv-grupo-hint-empty">Selecione câmeras na aba anterior para definir a ordem.</p>')
}

function abrirNovoGrupo() {
    $('#grupo-id').val('')
    $('#modal-grupo-titulo').text('Novo grupo')
    $('#grupo-nome').val('')
    $('#grupo-descricao').val('')
    $('#grupo-ativo').prop('checked', true)
    resetModalForm()
    prepararEditorGrupo(function () {
        renderChipsClientes()
        renderPickCameras()
        renderPainelOrdem()
        abrirEditorGrupo()
    })
}

function editarGrupo(id) {
    $.get('/api/grupos-visualizacao/' + encodeURIComponent(id))
        .done(function (r) {
            const g = (r && r.dados) || r
            if (!g || !g.id) return
            $('#grupo-id').val(g.id)
            $('#modal-grupo-titulo').text('Editar grupo')
            $('#grupo-nome').val(g.nome || '')
            $('#grupo-descricao').val(g.descricao || '')
            $('#grupo-ativo').prop('checked', g.ativo !== false)
            clientesAcl = {}
            ;(g.clientes || []).forEach(function (idCli) {
                idCli = String(idCli)
                clientesAcl[idCli] = { id: idCli, nome: 'Cliente ' + idCli }
            })
            ordemCameras = (g.cameras || []).slice().sort(function (a, b) {
                return (a.ordem || 0) - (b.ordem || 0)
            }).map(function (c) {
                return {
                    vis_camera_id: c.vis_camera_id || c.camera_id,
                    id_cliente: String(c.id_cliente || '')
                }
            })
            $('#grupo-busca-clientes').val('')
            $('#grupo-busca-cameras').val('')
            clientesBuscaOffset = 0
            clientesBuscaTermo = ''
            $('#grupo-clientes-resultados').html('<p class="cv-grupo-hint-empty">Digite ao menos 2 caracteres para buscar clientes.</p>')
            prepararEditorGrupo(function () {
                renderChipsClientes()
                renderPickCameras()
                renderPainelOrdem()
                abrirEditorGrupo()
            })
        })
        .fail(function (xhr) {
            CvMsg.erro(erroApi(xhr))
        })
}

function normalizarClientesListar(r) {
    let raw = (r && r.dados) || []
    if (!Array.isArray(raw)) raw = []
    return raw.map(function (c) {
        const id = String(c.id || c.idCliente || c.id_cliente || '')
        const nome = c.nome || c.nick || c.razaoSocial || c.fantasia || ('Cliente ' + id)
        return { id: id, nome: nome }
    }).filter(function (c) { return c.id })
}

function onBuscaClientesInput() {
    clearTimeout(buscaClientesTimer)
    buscaClientesTimer = setTimeout(function () {
        buscarClientes(true)
    }, GRUPO_BUSCA_MS)
}

function buscarClientes(reset) {
    const termo = ($('#grupo-busca-clientes').val() || '').trim()
    if (termo.length < 2) {
        clientesBuscaOffset = 0
        clientesBuscaTermo = ''
        $('#grupo-clientes-resultados').html('<p class="cv-grupo-hint-empty">Digite ao menos 2 caracteres para buscar clientes.</p>')
        $('#grupo-clientes-status').text('')
        $('#grupo-clientes-mais').addClass('d-none')
        return
    }
    if (reset || termo !== clientesBuscaTermo) {
        clientesBuscaOffset = 0
        clientesBuscaTermo = termo
        ultimosResultadosClientes = []
        $('#grupo-clientes-resultados').empty()
    }
    if (clientesBuscaCarregando) return
    clientesBuscaCarregando = true
    $('#grupo-clientes-status').text('Buscando…')
    $.ajax({
        url: '/ClienteListar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            idFranqueado: ConfVisionUrls.idFranqueado(),
            termo: termo,
            limit: GRUPO_CLI_PAGE,
            offset: clientesBuscaOffset
        })
    }).done(function (r) {
        const lista = normalizarClientesListar(r)
        lista.forEach(function (c) {
            mapaNomesCliente[c.id] = c.nome
        })
        if (clientesBuscaOffset === 0 && !lista.length) {
            $('#grupo-clientes-resultados').html('<p class="cv-grupo-hint-empty">Nenhum cliente encontrado.</p>')
        } else {
            appendResultadosClientes(lista)
        }
        const total = r && r.total != null ? r.total : null
        clientesBuscaHasMore = !!(r && r.hasMore)
        clientesBuscaOffset += lista.length
        if (total != null) {
            $('#grupo-clientes-status').text('Exibindo ' + Math.min(clientesBuscaOffset, total) + ' de ' + total)
        } else {
            $('#grupo-clientes-status').text(lista.length ? '' : '')
        }
        if (clientesBuscaHasMore) {
            $('#grupo-clientes-mais').removeClass('d-none')
        } else {
            $('#grupo-clientes-mais').addClass('d-none')
        }
    }).fail(function (xhr) {
        $('#grupo-clientes-status').text('')
        $('#grupo-clientes-resultados').html('<p class="text-danger small mb-0">' + escHtml(erroApi(xhr)) + '</p>')
    }).always(function () {
        clientesBuscaCarregando = false
    })
}

let ultimosResultadosClientes = []

function appendResultadosClientes(lista) {
    ultimosResultadosClientes = ultimosResultadosClientes.concat(lista)
    const $box = $('#grupo-clientes-resultados')
    if (clientesBuscaOffset === 0) $box.empty()
    lista.forEach(function (cli) {
        const checked = clientesAcl[cli.id] ? ' checked' : ''
        $box.append(`
            <label class="cv-check-row">
                <input type="checkbox" value="${escHtml(cli.id)}" data-nome="${escHtml(cli.nome)}"${checked}>
                <span>${escHtml(cli.nome)} <span class="cv-grupo-meta-id">#${escHtml(cli.id)}</span></span>
            </label>
        `)
    })
}

function renderResultadosClientesUltimos() {
    if (!ultimosResultadosClientes.length) return
    const $box = $('#grupo-clientes-resultados').empty()
    ultimosResultadosClientes.forEach(function (cli) {
        const checked = clientesAcl[cli.id] ? ' checked' : ''
        $box.append(`
            <label class="cv-check-row">
                <input type="checkbox" value="${escHtml(cli.id)}" data-nome="${escHtml(cli.nome)}"${checked}>
                <span>${escHtml(cli.nome)} <span class="cv-grupo-meta-id">#${escHtml(cli.id)}</span></span>
            </label>
        `)
    })
}

function renderChipsClientes() {
    const ids = Object.keys(clientesAcl)
    const $box = $('#grupo-clientes-selecionados').empty()
    if (!ids.length) {
        $box.html('<span class="cv-grupo-hint-empty">Nenhum cliente autorizado ainda.</span>')
        atualizarBadgeClientes()
        return
    }
    ids.forEach(function (id) {
        const c = clientesAcl[id]
        $box.append(`
            <span class="cv-grupo-chip" data-id="${escHtml(id)}">
                ${escHtml(nomeCliente(id))}
                <button type="button" class="cv-grupo-chip-remove" title="Remover">&times;</button>
            </span>
        `)
    })
    atualizarBadgeClientes()
}

function clientesSelecionados() {
    return Object.keys(clientesAcl)
}

function carregarCamerasSeNecessario(cb) {
    if (camerasCarregadas) {
        if (typeof cb === 'function') cb()
        return
    }
    $.get('/api/cameras?' + ConfVisionUrls.camerasQueryString())
        .done(function (r) {
            const raw = (r && r.dados) || r || []
            cacheCameras = Array.isArray(raw) ? raw : []
            camerasCarregadas = true
            if (typeof cb === 'function') cb()
        })
        .fail(function (xhr) {
            $('#grupo-cameras').html('<p class="text-danger small">' + escHtml(erroApi(xhr)) + '</p>')
        })
}

function onBuscaCamerasInput() {
    clearTimeout(buscaCamerasTimer)
    buscaCamerasTimer = setTimeout(renderPickCameras, GRUPO_BUSCA_MS)
}

function camerasSelecionadasSet() {
    const set = {}
    ordemCameras.forEach(function (c) {
        set[String(c.vis_camera_id)] = c
    })
    $('#grupo-cameras input[type=checkbox]:checked').each(function () {
        const camId = String($(this).val())
        const idCli = String($(this).data('cliente') || '')
        if (!set[camId]) {
            set[camId] = { vis_camera_id: camId, id_cliente: idCli }
        }
    })
    return set
}

function renderPickCameras() {
    const $box = $('#grupo-cameras').empty()
    if (!camerasCarregadas) {
        $box.html('<p class="cv-grupo-hint-empty">Carregando câmeras…</p>')
        return
    }
    if (!cacheCameras.length) {
        $box.html('<p class="cv-grupo-hint-empty">Nenhuma câmera cadastrada.</p>')
        $('#grupo-cameras-status').text('')
        atualizarContadorCameras()
        return
    }

    const termo = ($('#grupo-busca-cameras').val() || '').trim().toLowerCase()
    const selSet = camerasSelecionadasSet()
    const porCliente = {}

    cacheCameras.forEach(function (cam) {
        const camId = String(cam.id)
        const idCli = String(cam.id_cliente || '')
        const nomeCam = String(cam.nome || ('Câmera ' + camId))
        const nomeCli = nomeCliente(idCli).toLowerCase()
        const hay = !termo ||
            nomeCam.toLowerCase().indexOf(termo) >= 0 ||
            nomeCli.indexOf(termo) >= 0 ||
            camId.indexOf(termo) >= 0 ||
            idCli.indexOf(termo) >= 0
        if (!hay) return
        if (!porCliente[idCli]) porCliente[idCli] = []
        porCliente[idCli].push(cam)
    })

    const idsCli = Object.keys(porCliente).sort(function (a, b) {
        return nomeCliente(a).localeCompare(nomeCliente(b), 'pt-BR')
    })

    let totalFiltradas = 0
    idsCli.forEach(function (id) { totalFiltradas += porCliente[id].length })

    let renderizadas = 0
    let omitidas = 0

    idsCli.forEach(function (idCli) {
        if (renderizadas >= GRUPO_CAM_MAX_RENDER) {
            omitidas += porCliente[idCli].length
            return
        }
        const cams = porCliente[idCli]
        const tituloCli = nomeCliente(idCli)
        let html = `<div class="cv-grupo-cli-block" data-cliente="${escHtml(idCli)}">
            <h6 class="cv-grupo-cli-title">${escHtml(tituloCli)}${idCli ? ' <span class="cv-grupo-meta-id">#' + escHtml(idCli) + '</span>' : ''}</h6><div class="cv-grupo-cli-cams">`
        cams.forEach(function (cam) {
            if (renderizadas >= GRUPO_CAM_MAX_RENDER) {
                omitidas++
                return
            }
            const camId = String(cam.id)
            const checked = selSet[camId] ? ' checked' : ''
            html += `
                <label class="cv-check-row">
                    <input type="checkbox" value="${escHtml(camId)}" data-cliente="${escHtml(idCli)}"${checked}>
                    <span>${escHtml(cam.nome || 'Câmera #' + camId)}</span>
                </label>`
            renderizadas++
        })
        html += '</div></div>'
        $box.append(html)
    })

    if (!renderizadas) {
        $box.html('<p class="cv-grupo-hint-empty">Nenhuma câmera encontrada para esta busca.</p>')
    }

    let status = totalFiltradas + ' câmera' + (totalFiltradas === 1 ? '' : 's') + ' encontrada' + (totalFiltradas === 1 ? '' : 's')
    if (omitidas > 0) {
        status += ' — mostrando ' + renderizadas + ', refine a busca para ver as demais'
    }
    $('#grupo-cameras-status').text(status)
    atualizarContadorCameras()
}

function sincronizarOrdemCameras(selSet) {
    const idsAtivos = Object.keys(selSet)
    ordemCameras = ordemCameras.filter(function (c) {
        return idsAtivos.indexOf(String(c.vis_camera_id)) >= 0
    })
    idsAtivos.forEach(function (id) {
        if (!ordemCameras.some(function (c) { return String(c.vis_camera_id) === id })) {
            ordemCameras.push(selSet[id])
        }
    })
}

function renderPainelOrdem() {
    const $box = $('#grupo-ordem').empty()
    const selSet = camerasSelecionadasSet()
    sincronizarOrdemCameras(selSet)
    atualizarBadgeCameras()

    if (!ordemCameras.length) {
        $box.html('<p class="cv-grupo-hint-empty">Selecione câmeras na aba <strong>Câmeras do mosaico</strong> para definir a ordem.</p>')
        return
    }

    let html = '<ul class="cv-grupo-ordem-list">'
    ordemCameras.forEach(function (c, idx) {
        const cam = cacheCameras.find(function (x) { return String(x.id) === String(c.vis_camera_id) })
        html += `<li data-cam-id="${escHtml(c.vis_camera_id)}">
            <span class="cv-grupo-ordem-num">${idx + 1}</span>
            <span class="cv-grupo-ordem-label">
                <span class="cv-grupo-ordem-cliente">${escHtml(nomeCliente(c.id_cliente))}</span>
                <span class="cv-grupo-ordem-camera">${escHtml(cam ? cam.nome : ('Câmera ' + c.vis_camera_id))}</span>
            </span>
            <span class="cv-grupo-ordem-btns">
                <button type="button" class="cv-btn-ghost cv-btn-sm cv-grupo-ordem-btn" data-dir="up" data-idx="${idx}" title="Subir"><i class="bi bi-arrow-up"></i></button>
                <button type="button" class="cv-btn-ghost cv-btn-sm cv-grupo-ordem-btn" data-dir="down" data-idx="${idx}" title="Descer"><i class="bi bi-arrow-down"></i></button>
            </span>
        </li>`
    })
    html += '</ul>'
    $box.html(html)
}

function moverCameraOrdem(e) {
    e.preventDefault()
    const idx = parseInt($(this).data('idx'), 10)
    const dir = $(this).data('dir')
    if (isNaN(idx)) return
    const novo = dir === 'up' ? idx - 1 : idx + 1
    if (novo < 0 || novo >= ordemCameras.length) return
    const tmp = ordemCameras[idx]
    ordemCameras[idx] = ordemCameras[novo]
    ordemCameras[novo] = tmp
    renderPainelOrdem()
}

function atualizarContadorCameras() {
    const set = camerasSelecionadasSet()
    sincronizarOrdemCameras(set)
    atualizarBadgeCameras()
}

function montarPayloadCameras() {
    const set = camerasSelecionadasSet()
    const ids = ordemCameras.map(function (c) { return String(c.vis_camera_id) })
    Object.keys(set).forEach(function (id) {
        if (ids.indexOf(id) < 0) ids.push(id)
    })
    return ids.map(function (id, i) {
        const c = set[id] || { vis_camera_id: id, id_cliente: '' }
        return {
            vis_camera_id: parseInt(id, 10),
            id_cliente: String(c.id_cliente || ''),
            ordem: i + 1,
            ativo: true
        }
    })
}

function salvarGrupo() {
    const nome = ($('#grupo-nome').val() || '').trim()
    if (!nome) {
        CvMsg.aviso('Informe o nome do grupo.')
        return
    }
    const id = ($('#grupo-id').val() || '').trim()
    const payloadBase = {
        nome: nome,
        descricao: ($('#grupo-descricao').val() || '').trim(),
        ativo: $('#grupo-ativo').is(':checked')
    }
    const clientes = clientesSelecionados()
    const cameras = montarPayloadCameras()
    const reqMeta = id
        ? $.ajax({ url: '/api/grupos-visualizacao/' + encodeURIComponent(id), method: 'PUT', contentType: 'application/json', data: JSON.stringify(payloadBase) })
        : $.ajax({ url: '/api/grupos-visualizacao', method: 'POST', contentType: 'application/json', data: JSON.stringify(payloadBase) })

    reqMeta.done(function (r) {
        const g = (r && r.dados) || r
        const grupoId = id || (g && g.id)
        if (!grupoId) {
            CvMsg.erro('Resposta inválida ao salvar.')
            return
        }
        $.ajax({
            url: '/api/grupos-visualizacao/' + encodeURIComponent(grupoId) + '/composicao',
            method: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify({ clientes: clientes, cameras: cameras })
        }).done(function () {
            fecharEditorGrupo()
            CvMsg.sucesso('Grupo salvo.')
            carregarGrupos()
        }).fail(function (xhr) {
            CvMsg.erro(erroApi(xhr))
        })
    }).fail(function (xhr) {
        CvMsg.erro(erroApi(xhr))
    })
}

function excluirGrupo(id) {
    CvMsg.confirmar('Excluir grupo?', 'Esta ação não pode ser desfeita.').then(function (r) {
        if (!r.isConfirmed) return
        $.ajax({ url: '/api/grupos-visualizacao/' + encodeURIComponent(id), method: 'DELETE' })
            .done(function () {
                CvMsg.sucesso('Grupo excluído.')
                carregarGrupos()
            })
            .fail(function (xhr) {
                CvMsg.erro(erroApi(xhr))
            })
    })
}

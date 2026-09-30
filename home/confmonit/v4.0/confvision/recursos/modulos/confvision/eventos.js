let modalMidia = null
let carregandoEventos = false
let paginaAtualEventos = 1
let temMaisEventos = true
let ultimoScrollTop = 0
let cacheClientes = null
let cacheCameras = null

const EVENTOS_POR_PAGINA = 50

$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    const el = document.getElementById('modal-evento-midia')
    if (el && window.bootstrap) {
        modalMidia = new bootstrap.Modal(el)
    }
    $.when(carregarClientesCache(), carregarCamerasCache()).always(function () {
        preencherFiltroClientesEventos()
        aplicarVisibilidadeFiltroCliente()
        carregarEventos(true)
    })
    $('#btn-atualizar-eventos').on('click', function () {
        carregarEventos(true)
    })
    $('#btn-buscar-eventos').on('click', function () {
        carregarEventos(true)
    })
    $('#btn-limpar-eventos').on('click', function () {
        limparFiltrosEventos()
        carregarEventos(true)
    })
    $('#sel-evento-cliente, #filtro-evento-de, #filtro-evento-ate').on('keydown', function (e) {
        if (e.key === 'Enter') {
            e.preventDefault()
            carregarEventos(true)
        }
    })
    $('#eventos-scroll').on('scroll', onScrollEventos)
})

function normalizarRespostaEventos(resposta, paginaSolicitada) {
    const dados = resposta && resposta.dados != null ? resposta.dados : resposta

    if (Array.isArray(dados)) {
        const inicio = (paginaSolicitada - 1) * EVENTOS_POR_PAGINA
        const fim = inicio + EVENTOS_POR_PAGINA
        const items = dados.slice(inicio, fim)
        return {
            items: items,
            curPage: paginaSolicitada,
            nextPage: fim < dados.length ? paginaSolicitada + 1 : null,
            itemsReceived: items.length
        }
    }

    if (dados && Array.isArray(dados.items)) {
        return {
            items: dados.items,
            curPage: dados.curPage != null ? dados.curPage : paginaSolicitada,
            nextPage: dados.nextPage != null ? dados.nextPage : null,
            itemsReceived: dados.itemsReceived != null ? dados.itemsReceived : dados.items.length
        }
    }

    return {
        items: [],
        curPage: paginaSolicitada,
        nextPage: null,
        itemsReceived: 0
    }
}

function escHtml(valor) {
    if (valor == null || valor === '') return '—'
    return String(valor)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function labelNomeCliente(c) {
    if (!c) return null
    return c.nome || c.nomeCliente || c.nick || c.razaoSocial || null
}

function parseFiltroDataEvento(val, papel) {
    if (!val || !String(val).trim()) return null
    const bruto = String(val).trim()
    const d = new Date(bruto)
    if (isNaN(d.getTime())) return null
    if (papel === 'ate' && /T00:00(:00)?$/.test(bruto)) {
        d.setHours(23, 59, 59, 999)
    }
    return d.toISOString()
}

function obterFiltrosEventos() {
    const deRaw = $('#filtro-evento-de').val()
    const ateRaw = $('#filtro-evento-ate').val()
    const de = parseFiltroDataEvento(deRaw, 'de')
    const ate = parseFiltroDataEvento(ateRaw, 'ate')

    if (deRaw && !de) {
        return { erro: 'Data inicial inválida.' }
    }
    if (ateRaw && !ate) {
        return { erro: 'Data final inválida.' }
    }
    if (de && ate && new Date(de).getTime() > new Date(ate).getTime()) {
        return { erro: 'A data "De" não pode ser posterior à data "Até".' }
    }

    const idCliente = ($('#sel-evento-cliente').val() || '').trim()
    return { de: de, ate: ate, id_cliente: idCliente }
}

function limparFiltrosEventos() {
    $('#sel-evento-cliente').val('')
    $('#filtro-evento-de').val('')
    $('#filtro-evento-ate').val('')
    $('#lbl-eventos-filtro').text('')
}

function aplicarVisibilidadeFiltroCliente() {
    const ehCli = typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.ehCliente()
    $('#wrap-filtro-cliente-eventos').toggleClass('d-none', !!ehCli)
}

function preencherFiltroClientesEventos() {
    const $sel = $('#sel-evento-cliente')
    if (!$sel.length || !cacheClientes) return

    const ordenados = cacheClientes.slice().sort(function (a, b) {
        const na = String(labelNomeCliente(a) || '')
        const nb = String(labelNomeCliente(b) || '')
        return na.localeCompare(nb, 'pt-BR')
    })

    $sel.find('option:not(:first)').remove()
    ordenados.forEach(function (c) {
        const id = String(c.idCliente || c.id_cliente || c.id || '').trim()
        if (!id) return
        const nome = labelNomeCliente(c) || ('Cliente #' + id)
        $sel.append(
            $('<option></option>').attr('value', id).text(nome)
        )
    })
}

function montarQueryEventos(pagina) {
    const filtros = obterFiltrosEventos()
    if (filtros.erro) {
        CvMsg.aviso(filtros.erro)
        return null
    }

    const params = new URLSearchParams()
    params.set('page', String(pagina || 1))

    if (ConfVisionUrls.ehCliente()) {
        params.set('id_cliente', ConfVisionUrls.idCliente())
        if (ConfVisionUrls.idFranqueado()) {
            params.set('id_franqueado', ConfVisionUrls.idFranqueado())
        }
    } else {
        params.set('id_franqueado', ConfVisionUrls.idFranqueado() || '')
        if (filtros.id_cliente) {
            params.set('id_cliente', filtros.id_cliente)
        }
    }

    if (filtros.de) params.set('data_de', filtros.de)
    if (filtros.ate) params.set('data_ate', filtros.ate)

    return params.toString()
}

function atualizarLabelFiltroEventos(totalVisivel, reset) {
    const $lbl = $('#lbl-eventos-filtro')
    if (!reset) return

    const filtros = obterFiltrosEventos()
    if (filtros && filtros.erro) {
        $lbl.text('')
        return
    }

    const partes = []
    if (filtros.id_cliente) {
        partes.push('cliente filtrado')
    }
    if (filtros.de || filtros.ate) {
        partes.push('período aplicado')
    }
    if (!partes.length) {
        $lbl.text(totalVisivel != null ? (totalVisivel + ' evento(s) nesta página') : '')
        return
    }
    $lbl.text(partes.join(' · ') + (totalVisivel != null ? ' · ' + totalVisivel + ' nesta página' : ''))
}

function carregarClientesCache() {
    if (cacheClientes) {
        return $.Deferred().resolve(cacheClientes).promise()
    }
    // Portal CLI: /api/clientes não lista clientes; usa nome/nick da sessão
    if (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.ehCliente()) {
        cacheClientes = [{
            idCliente: ConfVisionUrls.idCliente(),
            nome: localStorage.getItem('nomeUsuario') || '',
            nick: localStorage.getItem('nickUsuario') || ''
        }]
        return $.Deferred().resolve(cacheClientes).promise()
    }
    return $.ajax({
        url: '/api/clientes',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ idFranqueado: ConfVisionUrls.idFranqueado() })
    }).then(function (r) {
        cacheClientes = r.dados || r || []
        return cacheClientes
    }, function () {
        cacheClientes = []
        return cacheClientes
    })
}

function carregarCamerasCache() {
    if (cacheCameras) {
        return $.Deferred().resolve(cacheCameras).promise()
    }
    return $.get(`/api/cameras?${ConfVisionUrls.camerasQueryString()}`)
        .then(function (r) {
            cacheCameras = r.dados || r || []
            return cacheCameras
        })
        .fail(function () {
            cacheCameras = []
            return cacheCameras
        })
}

function nomeCameraPorId(idCamera) {
    if (!idCamera || !cacheCameras) return null
    const id = String(idCamera)
    const camera = cacheCameras.find(function (c) {
        return String(c.id) === id
    })
    if (!camera) return null
    return camera.nome || null
}

function nomeCameraEvento(ev) {
    const nome = nomeCameraPorId(ev.vis_camera_id)
    return nome || ev.vis_camera_id || '—'
}

function nomeClientePorId(idCliente) {
    if (!idCliente || !cacheClientes) return null
    const id = String(idCliente)
    const cliente = cacheClientes.find(function (c) {
        return String(c.idCliente || c.id_cliente || c.id || '') === id
    })
    return labelNomeCliente(cliente)
}

function nomeClienteEvento(ev) {
    const nome = nomeClientePorId(ev.id_cliente)
    if (nome) return nome
    if (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.ehCliente()
        && String(ConfVisionUrls.idCliente()) === String(ev.id_cliente || '')) {
        return localStorage.getItem('nomeUsuario')
            || localStorage.getItem('nickUsuario')
            || ev.id_cliente
            || '—'
    }
    return ev.id_cliente || '—'
}

function labelStatusEvento(status) {
    const s = (status || '').toLowerCase()
    if (s === 'pronto') return '<span class="cv-badge cv-badge-ok">Pronto</span>'
    if (s === 'capturando') return '<span class="cv-badge cv-badge-warn">Capturando</span>'
    if (s === 'erro') return '<span class="cv-badge cv-badge-danger">Erro</span>'
    return '<span class="cv-badge cv-badge-muted">—</span>'
}

function temMidia(ev) {
    const s = (ev.status || '').toLowerCase()
    return !!(ev.snapshot_url || ev.video_url || (ev.clip_count && ev.clip_count > 0) || s === 'capturando' || s === 'pronto' || s === 'erro')
}

function setRefreshEventosLoading(loading) {
    $('#btn-atualizar-eventos').prop('disabled', loading)
    $('#icon-atualizar-eventos').toggleClass('cv-icon-spin', loading)
}

function setLoadMoreEventosLoading(loading) {
    $('#eventos-load-more').toggleClass('d-none', !loading)
}

function atualizarFimListaEventos() {
    const temLinhas = $('#tab-eventos tr').length > 0 && !$('#tab-eventos .cv-empty').length
    $('#eventos-fim-lista').toggleClass('d-none', temMaisEventos || !temLinhas)
}

function onScrollEventos() {
    if (carregandoEventos || !temMaisEventos) return

    const el = document.getElementById('eventos-scroll')
    if (!el) return

    const scrollTop = el.scrollTop
    if (scrollTop <= ultimoScrollTop) {
        ultimoScrollTop = scrollTop
        return
    }
    ultimoScrollTop = scrollTop

    if (scrollTop + el.clientHeight >= el.scrollHeight - 48) {
        carregarEventos(false)
    }
}

function renderLinhaEvento(ev) {
    const id = ev.id
    const midiaBtn = temMidia(ev)
        ? `<button type="button" class="cv-btn-ghost cv-btn-sm btn-ver-midia" data-id="${id}"><i class="bi bi-play-circle"></i> Ver</button>`
        : '<span class="cv-text-muted">—</span>'
    return `
        <tr>
            <td data-label="Data">${confVisionFormatarData(ev.created_at)}</td>
            <td data-label="Cliente">${escHtml(nomeClienteEvento(ev))}</td>
            <td data-label="Câmera">${escHtml(nomeCameraEvento(ev))}</td>
            <td data-label="Tipo" class="text-center">${ev.tipo_deteccao || '—'}</td>
            <td data-label="Confiança" class="text-center">${ev.confianca != null ? Number(ev.confianca).toFixed(2) : '—'}</td>
            <td data-label="Status" class="text-center">${labelStatusEvento(ev.status)}</td>
            <td data-label="Mídia" class="text-center cv-cell-action">${midiaBtn}</td>
        </tr>
    `
}

function vincularBotoesMidia(container) {
    container.find('.btn-ver-midia').off('click').on('click', function () {
        abrirMidiaEvento($(this).data('id'))
    })
}

function carregarEventos(reset) {
    if (carregandoEventos) return
    if (!reset && !temMaisEventos) return

    const pagina = reset ? 1 : paginaAtualEventos
    if (!reset && !pagina) return

    carregandoEventos = true
    if (reset) {
        paginaAtualEventos = 1
        temMaisEventos = true
        ultimoScrollTop = 0
        setRefreshEventosLoading(true)
        $('#eventos-fim-lista').addClass('d-none')
        if (document.getElementById('eventos-scroll')) {
            document.getElementById('eventos-scroll').scrollTop = 0
        }
    } else {
        setLoadMoreEventosLoading(true)
    }

    const qs = montarQueryEventos(pagina)
    if (!qs) {
        carregandoEventos = false
        setRefreshEventosLoading(false)
        setLoadMoreEventosLoading(false)
        return
    }

    $.get(`/api/eventos?${qs}`)
        .fail(function (e) {
            console.error(e)
            if (reset) {
                boxMesagemAtencaoPersonalizada('Erro ao carregar eventos')
            }
        })
        .done(function (r) {
            const meta = normalizarRespostaEventos(r, pagina)
            const lista = meta.items
            const tbody = $('#tab-eventos')

            if (reset) {
                tbody.empty()
            }

            if (!lista.length) {
                if (reset) {
                    tbody.append('<tr><td colspan="7"><div class="cv-empty">Nenhum evento registrado</div></td></tr>')
                }
                temMaisEventos = false
                atualizarFimListaEventos()
                return
            }

            lista.forEach(function (ev) {
                tbody.append(renderLinhaEvento(ev))
            })
            vincularBotoesMidia(tbody)
            if (reset) {
                atualizarLabelFiltroEventos(lista.length, true)
            }

            const recebidos = meta.itemsReceived != null ? meta.itemsReceived : lista.length
            if (meta.nextPage != null) {
                paginaAtualEventos = meta.nextPage
                temMaisEventos = true
            } else {
                temMaisEventos = recebidos >= EVENTOS_POR_PAGINA
                paginaAtualEventos = pagina + 1
            }

            atualizarFimListaEventos()
        })
        .always(function () {
            carregandoEventos = false
            setRefreshEventosLoading(false)
            setLoadMoreEventosLoading(false)
        })
}

function abrirMidiaEvento(eventoId) {
    if (!eventoId) return

    $('#modal-evento-midia-titulo').text(`Evento #${eventoId}`)
    $('#modal-evento-loading').removeClass('d-none')
    $('#modal-evento-conteudo').addClass('d-none')
    $('#modal-evento-snapshot').attr('src', '')
    $('#modal-evento-clips').empty()
    $('#box-modal-snapshot').addClass('d-none')

    if (modalMidia) modalMidia.show()

    $.get(`/api/eventos/${eventoId}/clips`)
        .fail(function (xhr) {
            console.error(xhr)
            $('#modal-evento-loading').addClass('d-none')
            $('#modal-evento-conteudo').removeClass('d-none')
            $('#modal-evento-clips').html('<div class="cv-empty py-3 text-danger">Erro ao carregar mídia do evento</div>')
        })
        .done(function (r) {
            const evento = r.evento || {}
            const clips = r.clips || []

            $('#modal-evento-loading').addClass('d-none')
            $('#modal-evento-conteudo').removeClass('d-none')

            if (evento.snapshot_url) {
                const sep = String(evento.snapshot_url).includes('?') ? '&' : '?'
                $('#modal-evento-snapshot').attr('src', evento.snapshot_url + sep + 't=' + Date.now())
                $('#box-modal-snapshot').removeClass('d-none')
            }

            const box = $('#modal-evento-clips').empty()
            if (!clips.length) {
                if (evento.video_url) {
                    box.append(renderClipPlayer(1, evento.video_url, null))
                } else {
                    box.append('<div class="cv-empty py-3">Nenhum vídeo capturado ainda</div>')
                }
                return
            }

            clips.forEach(function (clip) {
                box.append(renderClipPlayer(clip.seq, clip.video_url, clip.duracao_seg))
            })
        })
}

function renderClipPlayer(seq, url, duracaoSeg) {
    if (!url) return ''
    const duracao = duracaoSeg ? ` (${duracaoSeg}s)` : ''
    return `
        <div class="cv-clip-item">
            <p class="cv-clip-label">Clip ${seq || 1}${duracao}</p>
            <video controls playsinline preload="metadata" class="cv-clip-video" src="${url}"></video>
        </div>
    `
}

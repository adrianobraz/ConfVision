/**
 * Editor de area perimetral sobre o snapshot da camera (coordenadas 0-100%).
 * Modos: retangulo (arrastar) ou poligono (cliques).
 */
const ConfVisionAreaEditor = (function () {
    let cameraId = null
    let areas = []
    let permiteAreaDeteccao = true
    let drawMode = 'rect'
    let draftRect = null
    let draftPoints = []
    let polygonClosed = false
    let drawing = false
    let startPt = null
    let areaEmEdicao = null
    let pendingClickTimer = null

    const COR_PADRAO = '#38bdf8'

    function qs(sel) {
        return document.querySelector(sel)
    }

    function setInputVal(id, val) {
        const el = qs(id)
        if (el) el.value = val
    }

    function getInputVal(id) {
        const el = qs(id)
        return el ? el.value : ''
    }

    function imgEl() {
        return qs('#img-snapshot')
    }

    function canvasEl() {
        return qs('#canvas-area')
    }

    function stageEl() {
        return qs('#area-editor-stage')
    }

    function normalizarLista(r) {
        if (Array.isArray(r)) return r
        if (r && Array.isArray(r.dados)) return r.dados
        return []
    }

    function getImageDisplayRect() {
        const img = imgEl()
        if (!img) return null

        const natW = img.naturalWidth
        const natH = img.naturalHeight
        const boxW = img.clientWidth
        const boxH = img.clientHeight

        if (!boxW || !boxH) return null

        if (!natW || !natH) {
            return { offsetX: 0, offsetY: 0, dispW: boxW, dispH: boxH }
        }

        const scale = Math.min(boxW / natW, boxH / natH)
        const dispW = natW * scale
        const dispH = natH * scale

        return {
            offsetX: (boxW - dispW) / 2,
            offsetY: (boxH - dispH) / 2,
            dispW: dispW,
            dispH: dispH
        }
    }

    function syncCanvasSize() {
        const img = imgEl()
        const canvas = canvasEl()
        const stage = stageEl()
        if (!img || !canvas || !stage || !img.clientWidth) return

        canvas.width = img.clientWidth
        canvas.height = img.clientHeight
        canvas.style.width = img.clientWidth + 'px'
        canvas.style.height = img.clientHeight + 'px'
        stage.style.width = img.clientWidth + 'px'
        redraw()
    }

    function clientToPct(clientX, clientY) {
        const canvas = canvasEl()
        if (!canvas) return { x: 0, y: 0 }

        const rect = canvas.getBoundingClientRect()
        const disp = getImageDisplayRect()
        if (!disp || !disp.dispW || !disp.dispH) return { x: 0, y: 0 }

        const localX = clientX - rect.left - disp.offsetX
        const localY = clientY - rect.top - disp.offsetY
        const x = (localX / disp.dispW) * 100
        const y = (localY / disp.dispH) * 100

        return {
            x: Math.max(0, Math.min(100, x)),
            y: Math.max(0, Math.min(100, y))
        }
    }

    function pctToCanvas(xPct, yPct) {
        const disp = getImageDisplayRect()
        if (!disp) return { x: 0, y: 0 }

        return {
            x: disp.offsetX + (xPct / 100) * disp.dispW,
            y: disp.offsetY + (yPct / 100) * disp.dispH
        }
    }

    function rectToPontos(r) {
        const x1 = Math.min(r.x1, r.x2)
        const x2 = Math.max(r.x1, r.x2)
        const y1 = Math.min(r.y1, r.y2)
        const y2 = Math.max(r.y1, r.y2)
        return [
            { x: x1, y: y1 },
            { x: x2, y: y1 },
            { x: x2, y: y2 },
            { x: x1, y: y2 }
        ]
    }

    function parsePoligonoData(json) {
        if (!json) return null
        try {
            return typeof json === 'string' ? JSON.parse(json) : json
        } catch (e) {
            return null
        }
    }

    function parsePoligono(json) {
        const data = parsePoligonoData(json)
        if (!data) return null
        const pts = data.pontos || data.points || []
        return pts.length >= 3 ? pts : null
    }

    function isRectPoints(pontos) {
        if (!pontos || pontos.length !== 4) return false
        const xs = pontos.map(function (p) { return p.x })
        const ys = pontos.map(function (p) { return p.y })
        const uniqX = xs.filter(function (v, i, a) { return a.indexOf(v) === i })
        const uniqY = ys.filter(function (v, i, a) { return a.indexOf(v) === i })
        return uniqX.length === 2 && uniqY.length === 2
    }

    function drawPolygon(ctx, pontos, stroke, fill, closed) {
        if (!pontos || pontos.length < 2) return
        const fechado = closed !== false && pontos.length >= 3
        ctx.beginPath()
        pontos.forEach(function (p, i) {
            const c = pctToCanvas(p.x, p.y)
            if (i === 0) ctx.moveTo(c.x, c.y)
            else ctx.lineTo(c.x, c.y)
        })
        if (fechado) ctx.closePath()
        if (fill && fechado) {
            ctx.fillStyle = fill
            ctx.fill()
        }
        ctx.strokeStyle = stroke
        ctx.lineWidth = 2
        ctx.stroke()

        if (!fechado || drawMode === 'polygon') {
            pontos.forEach(function (p) {
                const c = pctToCanvas(p.x, p.y)
                ctx.beginPath()
                ctx.arc(c.x, c.y, 4, 0, Math.PI * 2)
                ctx.fillStyle = stroke
                ctx.fill()
            })
        }
    }

    function redraw() {
        const canvas = canvasEl()
        if (!canvas) return
        const ctx = canvas.getContext('2d')
        ctx.clearRect(0, 0, canvas.width, canvas.height)
        if (!shouldShowEditor()) return

        areas.forEach(function (area) {
            if (areaEmEdicao && areaEmEdicao.id === area.id) return
            const pts = parsePoligono(area.poligono_json)
            if (!pts) return
            const cor = area.cor || COR_PADRAO
            drawPolygon(ctx, pts, cor, cor + '33', true)
        })

        if (drawMode === 'rect' && draftRect) {
            const pts = rectToPontos(draftRect)
            drawPolygon(ctx, pts, '#fbbf24', 'rgba(251, 191, 36, 0.25)', true)
        }

        if (drawMode === 'polygon' && draftPoints.length) {
            const fechado = polygonClosed && draftPoints.length >= 3
            drawPolygon(
                ctx,
                draftPoints,
                '#fbbf24',
                fechado ? 'rgba(251, 191, 36, 0.25)' : null,
                fechado
            )
        }

        updateModoUi()
    }

    function clearDraft() {
        draftRect = null
        draftPoints = []
        polygonClosed = false
        if (pendingClickTimer) {
            clearTimeout(pendingClickTimer)
            pendingClickTimer = null
        }
    }

    function setDraftFromPoints(pontos, tipo) {
        clearDraft()
        if (!pontos || pontos.length < 3) return

        if (tipo === 'rect' || (tipo !== 'polygon' && isRectPoints(pontos))) {
            drawMode = 'rect'
            const xs = pontos.map(function (p) { return p.x })
            const ys = pontos.map(function (p) { return p.y })
            draftRect = {
                x1: Math.min.apply(null, xs),
                y1: Math.min.apply(null, ys),
                x2: Math.max.apply(null, xs),
                y2: Math.max.apply(null, ys)
            }
        } else {
            drawMode = 'polygon'
            draftPoints = pontos.map(function (p) { return { x: p.x, y: p.y } })
            polygonClosed = true
        }
    }

    function getDraftPontos() {
        if (drawMode === 'rect' && draftRect) {
            return rectToPontos(draftRect)
        }
        if (drawMode === 'polygon' && polygonClosed && draftPoints.length >= 3) {
            return draftPoints
        }
        return null
    }

    function updateModoUi() {
        const btnRect = qs('#btn-area-modo-rect')
        const btnPoly = qs('#btn-area-modo-polygon')
        const hint = qs('#lbl-area-modo-hint')
        const btnFechar = qs('#btn-area-fechar-poligono')

        if (btnRect) btnRect.classList.toggle('cv-area-mode-active', drawMode === 'rect')
        if (btnPoly) btnPoly.classList.toggle('cv-area-mode-active', drawMode === 'polygon')

        if (hint) {
            if (drawMode === 'rect') {
                hint.textContent = 'Arraste sobre a imagem para desenhar um retângulo.'
            } else if (polygonClosed) {
                hint.textContent = 'Polígono fechado. Clique em Salvar área ou Limpar para recomeçar.'
            } else {
                hint.textContent = 'Clique para adicionar pontos (mín. 3). Finalizar: dois cliques.'
            }
        }

        if (btnFechar) {
            const mostrar = drawMode === 'polygon' && !polygonClosed && draftPoints.length >= 3
            btnFechar.classList.toggle('d-none', !mostrar)
        }

        const canvas = canvasEl()
        if (canvas) {
            canvas.style.cursor = drawMode === 'rect' ? 'crosshair' : 'pointer'
        }
    }

    function setDrawMode(mode) {
        if (mode !== 'rect' && mode !== 'polygon') return
        if (drawMode === mode) return
        drawMode = mode
        clearDraft()
        areaEmEdicao = null
        setInputVal('#inp-area-nome', '')
        redraw()
    }

    function renderListaAreas() {
        const box = qs('#lista-areas')
        if (!box) return
        box.innerHTML = ''

        if (!areas.length) {
            box.innerHTML = '<div class="cv-area-list-empty">Nenhuma area cadastrada</div>'
            toggleAviso(true)
            return
        }

        toggleAviso(false)
        areas.forEach(function (area) {
            const nome = area.nome || ('Area #' + area.id)
            const ativo = area.ativo !== false
            const data = parsePoligonoData(area.poligono_json)
            const tipo = data && data.tipo === 'polygon' ? 'Polígono' : 'Retângulo'
            const item = document.createElement('div')
            item.className = 'cv-area-list-item'
            item.innerHTML =
                '<span class="cv-area-list-nome">' + nome +
                ' <span class="cv-area-list-tipo">(' + tipo + ')</span>' +
                (ativo ? '' : ' (inativa)') + '</span>' +
                '<span class="cv-area-list-actions">' +
                '<button type="button" class="cv-btn-ghost cv-btn-sm btn-editar-area" data-id="' + area.id + '">Editar</button>' +
                '<button type="button" class="cv-btn-ghost cv-btn-sm btn-excluir-area" data-id="' + area.id + '">Excluir</button>' +
                '</span>'
            box.appendChild(item)
        })

        box.querySelectorAll('.btn-editar-area').forEach(function (btn) {
            btn.addEventListener('click', function () {
                editarArea(parseInt(btn.getAttribute('data-id'), 10))
            })
        })
        box.querySelectorAll('.btn-excluir-area').forEach(function (btn) {
            btn.addEventListener('click', function () {
                excluirArea(parseInt(btn.getAttribute('data-id'), 10))
            })
        })
    }

    function toggleAviso(mostrar) {
        const aviso = qs('#box-area-aviso')
        const editor = qs('#box-area-editor')
        const modo = getModoDeteccao()
        const precisaArea = modo !== 'ambos'
        const visivel = !!mostrar && precisaArea && shouldShowEditor()
        if (aviso) aviso.classList.toggle('d-none', !visivel)
        if (!cameraId) {
            if (editor) editor.classList.add('d-none')
            const drawPanel = qs('#box-area-draw-panel')
            const drawUi = qs('#box-area-draw-ui')
            const drawFooter = qs('#box-area-draw-footer')
            if (drawPanel) drawPanel.classList.add('d-none')
            if (drawUi) drawUi.classList.add('d-none')
            if (drawFooter) drawFooter.classList.add('d-none')
        }
        atualizarModoDeteccaoHint()
    }

    function getModoDeteccao() {
        const el = qs('#sel-modo-deteccao')
        const v = el && el.value ? String(el.value).toLowerCase() : 'dentro'
        if (v === 'fora' || v === 'ambos') return v
        return 'dentro'
    }

    function atualizarModoDeteccaoHint() {
        const hint = qs('#lbl-modo-deteccao-hint')
        const avisoTxt = qs('#lbl-area-aviso-texto')
        const modo = getModoDeteccao()
        if (hint) {
            if (modo === 'fora') {
                hint.innerHTML = 'Gera evento quando a pessoa está <strong>fora</strong> da área desenhada.'
            } else if (modo === 'ambos') {
                hint.innerHTML = 'Gera evento com qualquer pessoa no frame — <strong>dentro ou fora</strong> da área (a área fica opcional).'
            } else {
                hint.innerHTML = 'Gera evento quando a pessoa está <strong>dentro</strong> da área desenhada.'
            }
        }
        if (avisoTxt) {
            if (modo === 'fora') {
                avisoTxt.innerHTML = 'Sem área cadastrada no modo <strong>Fora</strong>, esta câmera <strong>não gera eventos</strong>.'
            } else if (modo === 'ambos') {
                avisoTxt.innerHTML = 'No modo <strong>Ambos</strong> a área é opcional.'
            } else {
                avisoTxt.innerHTML = 'Sem área cadastrada, esta câmera <strong>não gera eventos</strong>.'
            }
        }
        // Atualiza aviso conforme modo + lista atual
        const semArea = !areas.length
        const aviso = qs('#box-area-aviso')
        if (aviso && shouldShowEditor()) {
            aviso.classList.toggle('d-none', !(semArea && modo !== 'ambos'))
        }
    }

    function toggleEditor(mostrar) {
        const editor = qs('#box-area-editor')
        const drawPanel = qs('#box-area-draw-panel')
        const drawUi = qs('#box-area-draw-ui')
        const drawFooter = qs('#box-area-draw-footer')
        if (editor) editor.classList.toggle('d-none', !mostrar)
        if (drawPanel) drawPanel.classList.toggle('d-none', !mostrar)
        if (drawUi) drawUi.classList.toggle('d-none', !mostrar)
        if (drawFooter) drawFooter.classList.toggle('d-none', !mostrar)
    }

    function loadAreas() {
        if (!cameraId) {
            areas = []
            renderListaAreas()
            redraw()
            return $.Deferred().resolve().promise()
        }

        return $.get('/api/cameras/' + cameraId + '/areas')
            .fail(function () {
                areas = []
                renderListaAreas()
                redraw()
            })
            .done(function (r) {
                areas = normalizarLista(r)
                renderListaAreas()
                redraw()
            })
    }

    function editarArea(id) {
        const area = areas.find(function (a) { return a.id === id })
        if (!area) return
        areaEmEdicao = area
        setInputVal('#inp-area-nome', area.nome || '')
        const data = parsePoligonoData(area.poligono_json)
        const pts = parsePoligono(area.poligono_json)
        setDraftFromPoints(pts, data ? data.tipo : null)
        redraw()
    }

    function excluirArea(id) {
        CvMsg.confirmar('Excluir área?', 'Esta área de detecção será removida.').then(function (r) {
            if (!r.isConfirmed) return
            $.ajax({
                url: '/api/camera-areas/' + id,
                method: 'DELETE'
            }).fail(function () {
                boxMesagemAtencaoPersonalizada('Erro ao excluir area')
            }).done(function () {
                if (areaEmEdicao && areaEmEdicao.id === id) {
                    areaEmEdicao = null
                    clearDraft()
                    setInputVal('#inp-area-nome', '')
                }
                boxSucessoAuto('Area excluida')
                loadAreas()
            })
        })
    }

    function limparDesenho() {
        clearDraft()
        areaEmEdicao = null
        setInputVal('#inp-area-nome', '')
        redraw()
    }

    function fecharPoligono() {
        if (drawMode !== 'polygon' || draftPoints.length < 3) return
        polygonClosed = true
        redraw()
    }

    function salvarArea() {
        if (!cameraId) {
            boxMesagemAtencaoPersonalizada('Salve a camera antes de cadastrar a area')
            return
        }

        const pontos = getDraftPontos()
        if (!pontos) {
            if (drawMode === 'polygon' && draftPoints.length >= 3 && !polygonClosed) {
                boxMesagemAtencaoPersonalizada('Feche o poligono antes de salvar')
            } else {
                boxMesagemAtencaoPersonalizada('Desenhe uma area sobre a imagem')
            }
            return
        }

        const nome = (getInputVal('#inp-area-nome') || '').trim() || 'Area principal'
        const poligono = {
            tipo: drawMode === 'rect' ? 'rect' : 'polygon',
            pontos: pontos
        }
        const payload = {
            vis_camera_id: parseInt(cameraId, 10),
            nome: nome,
            ativo: true,
            poligono_json: JSON.stringify(poligono),
            cor: COR_PADRAO
        }

        const isEdit = areaEmEdicao && areaEmEdicao.id
        const url = isEdit ? '/api/camera-areas/' + areaEmEdicao.id : '/api/cameras/' + cameraId + '/areas'
        const method = isEdit ? 'PUT' : 'POST'

        const btnSalvarEl = qs('#btn-area-salvar')
        if (btnSalvarEl) btnSalvarEl.disabled = true
        $.ajax({
            url: url,
            method: method,
            contentType: 'application/json',
            data: JSON.stringify(payload)
        }).fail(function () {
            boxMesagemAtencaoPersonalizada('Erro ao salvar area')
        }).done(function () {
            boxSucessoAuto(isEdit ? 'Area atualizada' : 'Area cadastrada')
            areaEmEdicao = null
            clearDraft()
            setInputVal('#inp-area-nome', '')
            loadAreas()
        }).always(function () {
            const btn = qs('#btn-area-salvar')
            if (btn) btn.disabled = false
        })
    }

    function onMouseDown(e) {
        if (!cameraId || !canvasEl() || drawMode !== 'rect') return
        if (e.button !== 0) return
        drawing = true
        startPt = clientToPct(e.clientX, e.clientY)
        draftRect = {
            x1: startPt.x,
            y1: startPt.y,
            x2: startPt.x,
            y2: startPt.y
        }
        areaEmEdicao = null
        redraw()
    }

    function onMouseMove(e) {
        if (drawMode !== 'rect' || !drawing || !startPt) return
        const pt = clientToPct(e.clientX, e.clientY)
        draftRect = {
            x1: startPt.x,
            y1: startPt.y,
            x2: pt.x,
            y2: pt.y
        }
        redraw()
    }

    function onMouseUp() {
        drawing = false
    }

    function onCanvasClick(e) {
        if (!cameraId || drawMode !== 'polygon' || polygonClosed) return
        if (drawing) return

        const pt = clientToPct(e.clientX, e.clientY)
        if (pendingClickTimer) clearTimeout(pendingClickTimer)

        pendingClickTimer = setTimeout(function () {
            pendingClickTimer = null
            if (!cameraId || drawMode !== 'polygon' || polygonClosed) return
            draftPoints.push({ x: pt.x, y: pt.y })
            areaEmEdicao = null
            redraw()
        }, 220)
    }

    function onCanvasDoubleClick(e) {
        if (!cameraId || drawMode !== 'polygon' || polygonClosed) return
        e.preventDefault()
        if (pendingClickTimer) {
            clearTimeout(pendingClickTimer)
            pendingClickTimer = null
        }
        if (draftPoints.length >= 3) {
            fecharPoligono()
        }
    }

    function bindEvents() {
        const canvas = canvasEl()
        if (!canvas || canvas._cvAreaBound) return
        canvas._cvAreaBound = true

        canvas.addEventListener('mousedown', onMouseDown)
        canvas.addEventListener('click', onCanvasClick)
        canvas.addEventListener('dblclick', onCanvasDoubleClick)
        window.addEventListener('mousemove', onMouseMove)
        window.addEventListener('mouseup', onMouseUp)

        const img = imgEl()
        if (img) img.addEventListener('load', syncCanvasSize)
        window.addEventListener('resize', syncCanvasSize)

        const btnSalvar = qs('#btn-area-salvar')
        const btnLimpar = qs('#btn-area-limpar')
        const btnFechar = qs('#btn-area-fechar-poligono')
        const btnRect = qs('#btn-area-modo-rect')
        const btnPoly = qs('#btn-area-modo-polygon')

        if (btnSalvar) btnSalvar.addEventListener('click', salvarArea)
        if (btnLimpar) btnLimpar.addEventListener('click', limparDesenho)
        if (btnFechar) btnFechar.addEventListener('click', fecharPoligono)
        if (btnRect) btnRect.addEventListener('click', function () { setDrawMode('rect') })
        if (btnPoly) btnPoly.addEventListener('click', function () { setDrawMode('polygon') })

        const selModo = qs('#sel-modo-deteccao')
        if (selModo && !selModo._cvModoBound) {
            selModo._cvModoBound = true
            selModo.addEventListener('change', atualizarModoDeteccaoHint)
        }
    }

    function hasSnapshotVisible() {
        const preview = qs('#box-snapshot-preview')
        return !!(preview && !preview.classList.contains('d-none'))
    }

    function isCapturaAnaliticoAtiva() {
        if (!permiteAreaDeteccao) return false
        const el = qs('#inp-captura-analitico')
        if (!el) return false
        return el.checked
    }

    function shouldShowEditor() {
        return permiteAreaDeteccao && hasSnapshotVisible() && isCapturaAnaliticoAtiva()
    }

    function setPermiteAreaDeteccao(permite) {
        permiteAreaDeteccao = !!permite
    }

    function refreshEditorVisibility() {
        const show = shouldShowEditor()
        toggleEditor(show)
        const canvas = canvasEl()
        if (canvas) {
            canvas.classList.toggle('d-none', !show)
            canvas.style.pointerEvents = show ? '' : 'none'
        }
        renderListaAreas()
        if (show) {
            setTimeout(syncCanvasSize, 100)
            setTimeout(syncCanvasSize, 400)
        } else {
            redraw()
        }
    }

    function setCameraId(id) {
        cameraId = id ? String(id) : null
        areaEmEdicao = null
        drawMode = 'rect'
        clearDraft()
        setInputVal('#inp-area-nome', '')

        if (!cameraId) {
            areas = []
            toggleEditor(false)
            toggleAviso(false)
            redraw()
            return
        }

        bindEvents()
        toggleEditor(shouldShowEditor())
        loadAreas().always(function () {
            setTimeout(syncCanvasSize, 50)
            setTimeout(syncCanvasSize, 300)
        })
    }

    function onSnapshotChange(hasImage) {
        if (!hasImage) {
            toggleEditor(false)
            redraw()
            return
        }
        if (cameraId) {
            refreshEditorVisibility()
        }
    }

    return {
        setCameraId: setCameraId,
        onSnapshotChange: onSnapshotChange,
        refreshEditorVisibility: refreshEditorVisibility,
        setPermiteAreaDeteccao: setPermiteAreaDeteccao,
        atualizarModoDeteccaoHint: atualizarModoDeteccaoHint,
        refresh: loadAreas
    }
})()

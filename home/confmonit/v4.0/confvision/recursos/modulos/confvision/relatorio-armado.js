let cacheCamerasArmado = []

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    carregarRelatorioArmado()

    $('#btn-atualizar-armado').on('click', carregarRelatorioArmado)
    $('#sel-filtro-armado').on('change', renderTabelaArmado)

    $('#tab-armado').on('click', '.cv-btn-armar', function (e) {
        e.preventDefault()
        const $btn = $(this)
        const idDisp = $btn.attr('data-armar-dispositivo')
        const status = $btn.attr('data-armar-status')
        ConfVisionArmado.toggleArmado($btn, idDisp, status, function () {
            carregarRelatorioArmado()
        })
    })

    $('#tab-armado').on('click', '.cv-btn-pausar-analitico', function (e) {
        e.preventDefault()
        const $btn = $(this)
        const cameraId = $btn.attr('data-pausar-camera')
        const pausado = $btn.attr('data-pausar-estado') === '1'
        ConfVisionArmado.togglePausarAnalitico(cameraId, !pausado, $btn, function (err) {
            if (!err) carregarRelatorioArmado()
        })
    })
})

function carregarRelatorioArmado() {
    const $tbody = $('#tab-armado')
    $tbody.html('<tr><td colspan="6"><div class="cv-empty">Carregando…</div></td></tr>')

    $.when(
        ConfVisionArmado.carregarCameras(),
        ConfVisionArmado.carregarClientes(),
        ConfVisionArmado.carregarDispositivosFranqueado()
    ).done(function (cameras, clientes, dispositivos) {
        const clientesMap = ConfVisionArmado.mapaClientes(clientes)
        const dispMap = ConfVisionArmado.mapaDispositivos(dispositivos)
        cacheCamerasArmado = ConfVisionArmado.enriquecerCameras(cameras, dispMap, clientesMap)
        atualizarResumoArmado(cacheCamerasArmado)
        renderTabelaArmado()
    }).fail(function (xhr) {
        console.error(xhr)
        $tbody.html('<tr><td colspan="6"><div class="cv-empty">Erro ao carregar relatório.</div></td></tr>')
    })
}

function atualizarResumoArmado(cameras) {
    const analiticas = cameras.filter(function (c) { return c._somente_armado })
    let armadas = 0
    let desarmadas = 0
    let fabCamera = 0
    let viaCentral = 0

    analiticas.forEach(function (c) {
        if (c._armado === 'S') armadas++
        else if (c._armado === 'N') desarmadas++
        if (c._is_camera_fab) fabCamera++
        else if (c._id_dispositivo) viaCentral++
    })

    $('#lbl-arm-total').text(analiticas.length)
    $('#lbl-arm-armadas').text(armadas)
    $('#lbl-arm-desarmadas').text(desarmadas)
    $('#lbl-arm-camera').text(fabCamera)
    $('#lbl-arm-central').text(viaCentral)
}

function filtrarCamerasArmado(cameras) {
    const filtro = $('#sel-filtro-armado').val() || 'todos'
    return cameras.filter(function (c) {
        if (filtro === 'analitico') return c._somente_armado
        if (filtro === 'armadas') return c._somente_armado && c._armado === 'S'
        if (filtro === 'desarmadas') return c._somente_armado && c._armado === 'N'
        if (filtro === 'camera') return c._is_camera_fab
        if (filtro === 'central') return c._somente_armado && !c._is_camera_fab && !!c._id_dispositivo
        if (filtro === 'pausadas') {
            return ConfVisionArmado.isPlanoAnalitico(c.plano, c) &&
                ConfVisionArmado.cameraAnaliticoAtiva(c) &&
                ConfVisionArmado.isAnaliticoPausado(c)
        }
        if (filtro === 'ativas') {
            return ConfVisionArmado.isPlanoAnalitico(c.plano, c) &&
                ConfVisionArmado.cameraAnaliticoAtiva(c) &&
                !ConfVisionArmado.isAnaliticoPausado(c)
        }
        return true
    })
}

function renderTabelaArmado() {
    const lista = filtrarCamerasArmado(cacheCamerasArmado)
    const $tbody = $('#tab-armado')

    if (!lista.length) {
        $tbody.html('<tr><td colspan="6"><div class="cv-empty">Nenhuma câmera neste filtro.</div></td></tr>')
        return
    }

    const ordenada = lista.slice().sort(function (a, b) {
        const na = String(a._nome_cliente || '')
        const nb = String(b._nome_cliente || '')
        const cmp = na.localeCompare(nb, 'pt-BR')
        if (cmp !== 0) return cmp
        return String(a.nome || '').localeCompare(String(b.nome || ''), 'pt-BR')
    })

    $tbody.html(ordenada.map(function (cam) {
        const origem = !cam._somente_armado
            ? 'Plano 24h'
            : (cam._is_camera_fab ? 'CAMERA' : 'Central / alarme')
        const status = cam._somente_armado
            ? ConfVisionArmado.badgeArmado(cam._armado)
            : '<span class="cv-badge cv-badge-stream">Sempre ativo</span>'
        const pausa = ConfVisionArmado.badgeAnaliticoPausado(cam)
        const inativa = ConfVisionArmado.badgeCameraInativaHome(cam)

        const btnEditar = ConfVisionArmado.cameraAnaliticoAtiva(cam)
            ? '<a href="/cameras/editar/' + encodeURIComponent(cam.id) +
              '" class="cv-btn-ghost cv-btn-sm" title="Editar"><i class="bi bi-pencil"></i></a>'
            : ''

        return (
            '<tr>' +
            '<td data-label="Cliente">' + ConfVisionArmado.escHtml(cam._nome_cliente) + '</td>' +
            '<td data-label="Câmera">' + ConfVisionArmado.escHtml(cam.nome || ('#' + cam.id)) + inativa + '</td>' +
            '<td data-label="Plano">' + ConfVisionArmado.escHtml(cam.plano || '—') + '</td>' +
            '<td data-label="Origem">' + origem + '</td>' +
            '<td data-label="Status">' + status + pausa + '</td>' +
            '<td data-label="Ações" class="cv-td-acoes cv-action-group">' +
            ConfVisionArmado.htmlBotaoArmar(cam) +
            ConfVisionArmado.htmlBotaoPausarAnalitico(cam) +
            btnEditar +
            '</td>' +
            '</tr>'
        )
    }).join(''))
}

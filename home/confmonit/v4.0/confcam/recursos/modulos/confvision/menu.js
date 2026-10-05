let homeClientesAbertos = {}
let homePainelCache = null
const homeFiltrosEstado = {
    armado: true,
    desarmado: true,
    pausado: true,
    despausado: true
}

const CV_HOME_MENUS = {
    cadastros: {
        titulo: 'Cadastros',
        itens: [
            { href: '/CarregarPaginaGerenciarCliente', icon: 'bi-people', label: 'Clientes', hint: 'Cadastrar e gerenciar' },
            { href: '/carregar-gerenciar-dispositivo', icon: 'bi-hdd-network', label: 'Dispositivos', hint: 'Centrais e alarmes' },
            { href: '/carregar-gerenciar-setores-alarme', icon: 'bi-bounding-box', label: 'Setores', hint: 'Zonas do alarme' },
            { href: '/cameras', icon: 'bi-camera-video', label: 'Câmera', hint: 'Cadastrar e gerenciar câmeras' },
            { href: '/grade-horario', icon: 'bi-calendar-week', label: 'Grade horária', hint: 'Armar/desarmar e pausar por horário' },
            { href: '/grupos-visualizacao', icon: 'bi-collection', label: 'Grupos de visualização', hint: 'Mosaicos multi-cliente' }
        ]
    },
    relatorios: {
        titulo: 'Relatórios',
        itens: [
            { href: '/eventos', icon: 'bi-activity', label: 'Eventos', hint: 'Histórico de alertas' },
            { href: '/relatorio-armado', icon: 'bi-shield-lock', label: 'Armado', hint: 'Câmeras armadas e desarmadas' },
            { href: '/relatorio-licencas', icon: 'bi-key', label: 'Licenças', hint: 'Usadas e não usadas' },
            { href: '/relatorio-faturas', icon: 'bi-receipt', label: 'Faturas', hint: 'Compra de licenças' },
            { href: '/rtmp-falhas', icon: 'bi-broadcast-pin', label: 'RTMP', hint: 'Online, falhas e publish' },
            { href: '/ips-banidos', icon: 'bi-slash-circle', label: 'IPs banidos', hint: 'Desbloquear IPs dos clientes' }
        ]
    },
    configuracoes: {
        titulo: 'Configurações',
        itens: [
            { href: '/cameras', icon: 'bi-camera-video', label: 'Câmera/Licença', hint: 'Cadastrar, editar, trocar licença' },
            { href: '/integracao-eventos', icon: 'bi-diagram-3', label: 'Integração', hint: 'Moni, ConfMonit e fotos' },
            { href: '/carregar-whitelabel', icon: 'bi-palette', label: 'Minha marca', hint: 'Logo da empresa' },
            { href: '/carregar-dominio', icon: 'bi-globe2', label: 'Domínio personalizado', hint: 'Subdomínio da central' },
            { href: '/minhas-licencas?tab=comprar', icon: 'bi-cart-plus', label: 'Comprar licenças', hint: 'Fatura no financeiro · auditoria' },
            { href: '/gravacoes', icon: 'bi-record-circle', label: 'Ativar gravação', hint: 'Licença add-on (DVR, movimento, timelapse)' }
        ]
    },
    monitoramento: {
        titulo: 'Monitoramento',
        itens: [
            { href: '/ao-vivo', icon: 'bi-broadcast', label: 'Ao vivo', hint: 'Assistir em tempo real' },
            { href: '/mosaicos', icon: 'bi-grid-3x3-gap', label: 'Mosaicos', hint: 'Video wall por grupo' },
            { href: '/gravacoes/dvr', icon: 'bi-play-btn', label: 'DVR', hint: 'Player e linha do tempo' },
            { href: '/gravacoes/timeline', icon: 'bi-collection-play', label: 'Timeline', hint: 'Buscar e assistir gravações' }
        ]
    }
}

$(document).ready(function () {
    if (typeof ConfVisionUrls === 'undefined') return
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    aplicarHomeCliente()
    initDashboardInteracao()
    carregarDashboardResumo()
    carregarPainelArmadoHome()
    initHomeDrawer()

    $('#box-clientes-cameras').on('click', '.cv-home-cliente-toggle', function (e) {
        e.preventDefault()
        const $card = $(this).closest('.cv-home-cliente-card')
        const id = String($card.data('cliente-id') || '')
        const abrir = !$card.hasClass('is-open')

        if (abrir) {
            $card.siblings('.cv-home-cliente-card.is-open').each(function () {
                const $outro = $(this)
                const outroId = String($outro.data('cliente-id') || '')
                $outro.removeClass('is-open')
                $outro.find('.cv-home-cliente-toggle').attr('aria-expanded', 'false')
                if (outroId) homeClientesAbertos[outroId] = false
            })
        }

        $card.toggleClass('is-open', abrir)
        $(this).attr('aria-expanded', abrir ? 'true' : 'false')
        if (id) homeClientesAbertos[id] = abrir

        if (abrir && $card[0]) {
            window.setTimeout(function () {
                $card[0].scrollIntoView({ block: 'nearest', behavior: 'smooth' })
            }, 50)
        }
    })

    $('#box-clientes-cameras').on('click', '.cv-btn-armar', function (e) {
        e.preventDefault()
        e.stopPropagation()
        const $btn = $(this)
        const idDisp = $btn.attr('data-armar-dispositivo')
        const status = $btn.attr('data-armar-status')
        ConfVisionArmado.toggleArmado($btn, idDisp, status, function () {
            carregarPainelArmadoHome()
        })
    })

    $('#box-clientes-cameras').on('click', '.cv-btn-pausar-analitico', function (e) {
        e.preventDefault()
        e.stopPropagation()
        const $btn = $(this)
        const cameraId = $btn.attr('data-pausar-camera')
        const pausado = $btn.attr('data-pausar-estado') === '1'
        ConfVisionArmado.togglePausarAnalitico(cameraId, !pausado, $btn, function (err) {
            if (!err) carregarPainelArmadoHome()
        })
    })

    $('#cv-home-filtros-estado').on('click', '.cv-home-filtro-chip', function () {
        const key = String($(this).data('filtro') || '')
        if (!key || homeFiltrosEstado[key] == null) return
        homeFiltrosEstado[key] = !homeFiltrosEstado[key]
        $(this).toggleClass('is-active', homeFiltrosEstado[key])
        renderHomeClientesCameras()
    })
})

function aplicarHomeCliente() {
    if (!ConfVisionUrls.ehCliente()) return
    $('#cv-home-titulo').text('Minhas câmeras')
    $('#cv-home-subtitulo').text('Acompanhe ao vivo, eventos, DVR e o status de armação')
    $('#cv-home-clientes-titulo').text('Suas câmeras')
    $('#cv-home-clientes-desc').text('Armar/desarmar dispositivos e assistir gravações')
    $('#cv-home-cards-fra').addClass('d-none')
    $('#cv-home-cards-cli').removeClass('d-none')
    $('#box-clientes-cameras-empty').text('Nenhuma câmera vinculada à sua conta.')
}

function initHomeDrawer() {
    if (ConfVisionUrls.ehCliente()) return
    const $drawer = $('#cv-home-drawer')
    if (!$drawer.length) return

    $('.cv-home-menu-btn').on('click', function () {
        // No celular os atalhos usam o sanduíche; drawer lateral só no desktop
        if (window.innerWidth <= 992) return
        abrirHomeDrawer($(this).data('home-menu'))
    })

    $('#cv-home-drawer-close').on('click', fecharHomeDrawer)
    $('#cv-home-drawer-backdrop').on('click', fecharHomeDrawer)

    $(document).on('keydown.homeDrawer', function (e) {
        if (e.key === 'Escape') fecharHomeDrawer()
    })

    $(window).on('resize.homeDrawer', function () {
        if (window.innerWidth <= 992) fecharHomeDrawer()
    })
}

function abrirHomeDrawer(key) {
    const menu = CV_HOME_MENUS[key]
    if (!menu) return

    const $drawer = $('#cv-home-drawer')
    const $backdrop = $('#cv-home-drawer-backdrop')

    $('#cv-home-drawer-title').text(menu.titulo)
    $('#cv-home-drawer-links').html(menu.itens.map(function (item) {
        if (item.separator) {
            return '<div class="cv-home-drawer-divider" role="separator"></div>'
        }
        return (
            '<a href="' + item.href + '" class="cv-home-drawer-item">' +
            '<i class="bi ' + item.icon + '"></i>' +
            '<span class="cv-home-drawer-item-text">' +
            '<span class="cv-home-drawer-item-label">' + item.label + '</span>' +
            '<span class="cv-home-drawer-item-hint">' + item.hint + '</span>' +
            '</span>' +
            '<i class="bi bi-chevron-right cv-home-drawer-item-go"></i>' +
            '</a>'
        )
    }).join(''))

    $drawer.prop('hidden', false).attr('aria-hidden', 'false')
    $backdrop.prop('hidden', false)
    requestAnimationFrame(function () {
        document.body.classList.add('cv-home-drawer-open')
    })
}

function fecharHomeDrawer() {
    document.body.classList.remove('cv-home-drawer-open')
    const $drawer = $('#cv-home-drawer')
    const $backdrop = $('#cv-home-drawer-backdrop')
    if (!$drawer.length) return

    window.setTimeout(function () {
        if (!document.body.classList.contains('cv-home-drawer-open')) {
            $drawer.prop('hidden', true).attr('aria-hidden', 'true')
            $backdrop.prop('hidden', true)
        }
    }, 220)
}

function carregarPainelArmadoHome() {
    const $box = $('#box-clientes-cameras')
    const $loading = $('#box-clientes-cameras-loading')
    const $empty = $('#box-clientes-cameras-empty')
    const $filtroEmpty = $('#box-clientes-cameras-filtro-empty')
    if (!$box.length) return

    $loading.removeClass('d-none')
    $empty.addClass('d-none')
    $filtroEmpty.addClass('d-none')
    $box.empty()

    $.when(
        ConfVisionArmado.carregarCameras(),
        ConfVisionArmado.carregarClientes(),
        ConfVisionArmado.carregarDispositivosFranqueado()
    ).done(function (cameras, clientes, dispositivos) {
        $loading.addClass('d-none')
        const clientesMap = ConfVisionArmado.mapaClientes(clientes)
        const dispMap = ConfVisionArmado.mapaDispositivos(dispositivos)
        const enriquecidas = ConfVisionArmado.enriquecerCameras(cameras, dispMap, clientesMap)
        homePainelCache = ConfVisionArmado.agruparPorCliente(enriquecidas)

        if (!homePainelCache.length) {
            $empty.removeClass('d-none')
            return
        }

        renderHomeClientesCameras()
    }).fail(function (xhr) {
        console.error(xhr)
        $loading.addClass('d-none')
        homePainelCache = null
        $box.html('<div class="cv-empty">Não foi possível carregar as câmeras.</div>')
    })
}

function renderHomeClientesCameras() {
    const $box = $('#box-clientes-cameras')
    const $empty = $('#box-clientes-cameras-empty')
    const $filtroEmpty = $('#box-clientes-cameras-filtro-empty')
    if (!$box.length || !homePainelCache) return

    $box.empty()
    $empty.addClass('d-none')
    $filtroEmpty.addClass('d-none')

    const grupos = homePainelCache
        .slice()
        .sort(ConfVisionArmado.compararClientesPorEstadoHome)

    let totalVisiveis = 0

    grupos.forEach(function (grupo) {
        const idCliente = String(grupo.id_cliente || '')
        const aberto = !!homeClientesAbertos[idCliente]

        const dispositivosFiltrados = grupo.listaDispositivos
            .map(function (disp) {
                const cameras = (disp.cameras || [])
                    .filter(function (cam) {
                        if (typeof cameraPassaFiltroKpi === 'function' && !cameraPassaFiltroKpi(cam)) {
                            return false
                        }
                        return ConfVisionArmado.cameraPassaFiltroHome(cam, homeFiltrosEstado)
                    })
                    .slice()
                    .sort(ConfVisionArmado.compararCamerasPorEstadoHome)
                if (!cameras.length) return null
                return Object.assign({}, disp, { cameras: cameras })
            })
            .filter(Boolean)
            .sort(ConfVisionArmado.compararDispositivosPorEstadoHome)

        if (!dispositivosFiltrados.length) return

        totalVisiveis += dispositivosFiltrados.reduce(function (n, d) {
            return n + d.cameras.length
        }, 0)

        const estadosCliente = dispositivosFiltrados.map(function (disp) {
            return ConfVisionArmado.estadoDispositivoHome(disp)
        })
        const classeBordaCliente = ConfVisionArmado.classeBordaClienteHome(estadosCliente)
        const nCam = dispositivosFiltrados.reduce(function (n, d) { return n + d.cameras.length }, 0)
        const nDisp = dispositivosFiltrados.length

        let corposDisp = ''
        dispositivosFiltrados.forEach(function (disp) {
            const statusCls = ConfVisionArmado.classeStatusDispositivo(disp)
            const nomeDisp = ConfVisionArmado.escHtml(
                ConfVisionArmado.textoOuVazio(disp.nome, 'Dispositivo')
            )
            const rows = disp.cameras.map(function (cam) {
                const nomeCam = ConfVisionArmado.escHtml(
                    ConfVisionArmado.textoOuVazio(cam.nome, 'Câmera #' + cam.id)
                )
                const licenca = ConfVisionArmado.escHtml(
                    ConfVisionArmado.labelLicenca(cam.plano)
                )
                const estadoCam = ConfVisionArmado.estadoPrincipalCameraHome(cam)
                const rowCls = ConfVisionArmado.classeBordaPorEstadoHome(estadoCam)
                const ehCliente = typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.ehCliente()
                return (
                    '<tr class="' + rowCls + '">' +
                    '<td class="cv-col-camera">' + nomeCam +
                    ConfVisionArmado.badgeCameraInativaHome(cam) +
                    (ConfVisionArmado.cameraAnaliticoAtiva(cam) && ConfVisionArmado.isAnaliticoPausado(cam)
                        ? ' ' + ConfVisionArmado.badgeAnaliticoPausado(cam) : '') +
                    '</td>' +
                    '<td class="cv-col-licenca">' + licenca + '</td>' +
                    '<td class="cv-col-acao cv-home-cam-acoes">' +
                    '<div class="cv-action-group">' +
                    ConfVisionArmado.htmlAcoesCameraHome(cam, ehCliente) +
                    '</div>' +
                    '</td>' +
                    '</tr>'
                )
            }).join('')

            corposDisp += (
                '<div class="cv-home-disp ' + statusCls + '">' +
                '<div class="cv-home-disp-head">' +
                '<div class="cv-home-disp-titulo">' +
                '<strong>' + nomeDisp + '</strong>' +
                (disp.exige_armado ? ConfVisionArmado.badgeArmado(disp.armado) : '') +
                '</div>' +
                '<div class="cv-home-disp-acoes">' +
                ConfVisionArmado.htmlBotaoArmarDispositivoHome(disp) +
                '</div>' +
                '</div>' +
                '<table class="cv-home-cam-table">' +
                '<thead><tr>' +
                '<th>Câmera</th><th>Licença</th><th></th>' +
                '</tr></thead>' +
                '<tbody>' + rows + '</tbody>' +
                '</table>' +
                '</div>'
            )
        })

        $box.append(
            '<section class="cv-home-cliente-card' + (aberto ? ' is-open' : '') +
            (classeBordaCliente ? ' ' + classeBordaCliente : '') + '" ' +
            'data-cliente-id="' + ConfVisionArmado.escHtml(idCliente) + '">' +
            '<button type="button" class="cv-home-cliente-toggle" ' +
            'aria-expanded="' + (aberto ? 'true' : 'false') + '">' +
            '<span class="cv-home-cliente-nome">' +
            '<i class="bi bi-chevron-right cv-home-chevron"></i>' +
            ConfVisionArmado.escHtml(grupo.nome) +
            '</span>' +
            '<span class="cv-home-cliente-meta">' + nCam + ' câmera(s) · ' +
            nDisp + ' dispositivo(s)</span>' +
            '</button>' +
            '<div class="cv-home-cliente-body">' + corposDisp + '</div>' +
            '</section>'
        )
    })

    if (!totalVisiveis) {
        $filtroEmpty.removeClass('d-none')
    }
}

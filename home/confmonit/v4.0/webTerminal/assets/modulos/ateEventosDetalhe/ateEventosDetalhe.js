// Cache para não recarregar ao voltar da Grade/Contatos/Procedimentos etc.
var ateEventosDetalhe_cache = {
    idProcesso: '',
    html: '',
    idDispositivo: '',
    dadosEventos: [],
    dadosUltimos25: []
}
// Evita aplicar resultado de requisição antiga ao trocar de processo rápido
var ateEventosDetalhe_reqIdProcesso = ''

// Atualização automática somente dos "eventos pendentes"
var ateEventosDetalhe_intervalPendentes = null
var ateEventosDetalhe_refreshPendentesLock = false
var ateEventosDetalhe_refreshPendentesReqId = ''

function ateEventosDetalhe_setupRefreshPendentes() {
    if (ateEventosDetalhe_intervalPendentes) return

    const $btn = $('#ateEventosDetalhe_btnRefreshPendentes')
    if ($btn && $btn.length > 0) {
        $btn.off('click').on('click', function () {
            ateEventosDetalhe_refreshPendentes()
        })
    }

    ateEventosDetalhe_intervalPendentes = setInterval(function () {
        ateEventosDetalhe_refreshPendentes()
    }, 30000)
}

function ateEventosDetalhe_bindCameraHandlers() {
    $('[tipo=btnConfVisionPainelSetor]').off('click').on('click', function (e) {
        e.stopPropagation()
        confvision_abrirPainelSetor(confvision_itemFromCell($(this)))
    })

    $('[tipo=btnConfVisionPainelUlt25]').off('click').on('click', function (e) {
        e.stopPropagation()
        confvision_abrirPainelPorSetor(confvision_itemFromCell($(this)))
    })

    $('td[tipo=btnVisualizarCamera]').off('click').on('click', function () {
        confvision_visualizarCameraBenuvem(confvision_itemFromCell($(this)))
    })

    $('td[tipo=btnBenuvemMidiaUlt25]').off('click').on('click', function (e) {
        e.stopPropagation()
        const img = $(this).attr('img') || ''
        if (!img || img === 'SEM IMAGEM') return
        try {
            const d = JSON.parse(img.replaceAll(/'/g, '"'))
            benuvem_visualizar(d.company_code, d.partition, d.client_code, d.channels, d.date)
        } catch (err) {
            console.log(err)
        }
    })
}

function ateEventosDetalhe_htmlCameraCelula(item) {
    const attrs = `
        idProcesso="${item.idProcesso || ''}"
        idDispositivo="${item.idDispositivo || ''}"
        particao="${item.particao || ''}"
        setor="${item.zonaUser || ''}"
        idEvento="${item.idEvento || ''}"
        img="${(item.img || '').replace(/"/g, '&quot;')}"
        conta="${item.conta || ''}"
        codBenuvem="${item.codBenuvem || ''}"
        provedorVideo="${item.provedorVideo || ''}"
        usaConfVision="${item.usaConfVision || ''}"
    `

    if (item.cameraOn !== 'S') {
        return `<td class="bg-secondary text-center cv-cam-sem" style="width: 6%;" title="Setor sem câmera"><i class="bi bi-x-lg txt-tbody"></i></td>`
    }

    if (confvision_eConfVision(item)) {
        return `
            <td
                tipo="btnConfVisionPainelSetor"
                class="bg-success click text-center"
                style="width: 6%;"
                ${attrs}
                title="ConfVision — eventos, foto e ao vivo"
            ><i class="bi bi-camera-video-fill txt-tbody"></i></td>
        `
    }

    return `
        <td
            tipo="btnVisualizarCamera"
            class="bg-success click"
            ${attrs}
            title="Exibição em tempo real (Benuvem)"
            style="width: 6%;"
        ><i class="bi bi-eye-fill txt-tbody"></i></td>
    `
}

function ateEventosDetalhe_htmlCameraUltimos25(item) {
    const temCamera = item.cameraOn === 'S' || (item.img && item.img !== 'SEM IMAGEM') || confvision_parseImg(item.img)
    if (!temCamera) {
        return `<td class="text-center cv-cam-sem" title="Sem câmera"><i class="bi bi-x-lg text-muted"></i></td>`
    }

    const attrs = `
        idProcesso="${item.idProcesso || ''}"
        idDispositivo="${item.idDispositivo || ''}"
        particao="${item.particao || ''}"
        zonaUser="${item.zonaUser || ''}"
        idEvento="${item.idEvento || ''}"
        img="${(item.img || '').replace(/"/g, '&quot;')}"
        conta="${item.conta || ''}"
        codBenuvem="${item.codBenuvem || ''}"
        provedorVideo="${item.provedorVideo || ''}"
        usaConfVision="${item.usaConfVision || ''}"
    `

    if (confvision_eConfVision(item) || confvision_parseImg(item.img)) {
        return `
            <td
                class="text-center click bg-success"
                tipo="btnConfVisionPainelUlt25"
                ${attrs}
                title="ConfVision — eventos, foto e ao vivo"
            ><i class="bi bi-camera-video-fill txt-tbody"></i></td>
        `
    }

    return `
        <td class="text-center click" img="${(item.img || '').replace(/"/g, '&quot;')}" tipo="btnBenuvemMidiaUlt25" title="Ver gravação">
            <i class="bi bi-play-circle-fill text-primary"></i>
        </td>
    `
}

function ateEventosDetalhe_refreshPendentes() {
    if (!document.getElementById('ateEventosDetalhe')) return

    var idProcessoAtual = sessionStorage.getItem('ateProDado_proc_idProcesso')
    if (!idProcessoAtual) return
    if (ateEventosDetalhe_refreshPendentesLock) return

    ateEventosDetalhe_refreshPendentesLock = true
    ateEventosDetalhe_refreshPendentesReqId = idProcessoAtual

    $.ajax({
        url: 'ateEventosDetalhe/buscarEventos',
        method: 'Post',
        cache: false,
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idProcesso: idProcessoAtual })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (sessionStorage.getItem('ateProDado_proc_idProcesso') !== ateEventosDetalhe_refreshPendentesReqId) return
        if (!r.dados) r.dados = []

        ateEventosDetalhe_cache.dadosEventos = r.dados

        $('#ateEventosDetalhe_tabEventos tbody').empty()
        r.dados.forEach(item => {
            ateEventosDetalhe_buscarEventos_montarLinha(item)
        })

        $('td[tipo=btnEvento]').off('click').on('click', function () {
            modEvtDesagrupado_start(
                $(this).attr('idProcesso'),
                $(this).attr('codigoFull'),
                $(this).attr('zonaUser'),
                ateEventosDetalhe_start
            )
        })

        ateEventosDetalhe_bindCameraHandlers()

        $('td[tipo=btnSetorManutecao]').off('click').on('click', function () {
            const idSetor = $(this).attr('idSetor')
            ateEventosDetalhe_setorManutencao(idSetor)
        })
    }).always(function () {
        ateEventosDetalhe_refreshPendentesLock = false
    })
}

function ateEventosDetalhe_start() {
    var idProcessoAtual = sessionStorage.getItem('ateProDado_proc_idProcesso')
    var box = document.getElementById('boxDir')

    ateEventosDetalhe_setupRefreshPendentes()

    if (ateEventosDetalhe_cache.idProcesso === idProcessoAtual && ateEventosDetalhe_cache.html) {
        box.innerHTML = ateEventosDetalhe_cache.html
        idProcesso = idProcessoAtual
        ateEventosDetalhe_reqIdProcesso = idProcessoAtual
        ateEventosDetalhe_renderFromCache()
        ateEventosDetalhe_refreshPendentes()
        return
    }

    fetch(`/assets/modulos/ateEventosDetalhe/ateEventosDetalhe.html`, { cache: 'no-store' }).then((res) => res.text()).then((html) => {
        box.innerHTML = html
        ateEventosDetalhe_cache.html = html
        ateEventosDetalhe_cache.idProcesso = idProcessoAtual
        idProcesso = idProcessoAtual
        ateEventosDetalhe_reqIdProcesso = idProcessoAtual

        ateEventosDetalhe_setupRefreshPendentes()
        confvision_carregarConfig().always(function () {
            ateEventosDetalhe_buscarEventos(idProcesso, (idDispositivo) => {
                if (sessionStorage.getItem('ateProDado_proc_idProcesso') !== ateEventosDetalhe_reqIdProcesso) return
                ateEventosDetalhe_buscarUltimos25(idDispositivo)
            })
        })
    })
}


function ateEventosDetalhe_renderFromCache() {
    var cache = ateEventosDetalhe_cache
    $('#ateEventosDetalhe_tabEventos tbody').empty()
    if (cache.dadosEventos && cache.dadosEventos.length > 0) {
        cache.dadosEventos.forEach(function (item) {
            ateEventosDetalhe_buscarEventos_montarLinha(item)
        })
    }

    if (cache.dadosEventos && cache.dadosEventos.length > 0) {
        $('td[tipo=btnEvento]').on('click', function () {
            modEvtDesagrupado_start(
                $(this).attr('idProcesso'),
                $(this).attr('codigoFull'),
                $(this).attr('zonaUser'),
                ateEventosDetalhe_start
            )
        })
        ateEventosDetalhe_bindCameraHandlers()
        $('td[tipo=btnSetorManutecao]').on('click', function () {
            var idSetor = $(this).attr('idSetor')
            ateEventosDetalhe_setorManutencao(idSetor)
        })
    }

    $('#ateEventosDetalhe_tabUltimos25 tbody').empty()
    if (cache.dadosUltimos25 && cache.dadosUltimos25.length > 0) {
        var qtd = 1
        cache.dadosUltimos25.forEach(function (item) {
            ateEventosDetalhe_buscarUltimos25_montarLinha(item, qtd)
            qtd++
        })
        ateEventosDetalhe_bindCameraHandlers()
        $('tr[tipo=btnUltimos25]').on('click', function () {
            var modal = new bootstrap.Modal(document.getElementById('descricaoAtendimento'), { keyboard: false })
            $('#descricaoAtendimentoMsg').val($(this).attr('descAtendimento'))
            modal.show()
        })
    }
}


function ateEventosDetalhe_buscarEventos(idProcesso, next) {
    $.ajax({
        url: 'ateEventosDetalhe/buscarEventos',
        method: 'Post',
        cache: false,
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idProcesso: idProcesso })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (sessionStorage.getItem('ateProDado_proc_idProcesso') !== ateEventosDetalhe_reqIdProcesso) return
        if (!r.dados) r.dados = []
        ateEventosDetalhe_cache.dadosEventos = r.dados
        ateEventosDetalhe_cache.idDispositivo = (r.dados[0] && r.dados[0].idDispositivo) ? r.dados[0].idDispositivo : ''
        $('#ateEventosDetalhe_tabEventos tbody').empty()

        r.dados.forEach(item => {
            ateEventosDetalhe_buscarEventos_montarLinha(item)
        })

        $('td[tipo=btnEvento]').on('click', function () {
            modEvtDesagrupado_start(
                $(this).attr('idProcesso'),
                $(this).attr('codigoFull'),
                $(this).attr('zonaUser'),
                ateEventosDetalhe_start
            )
        })

        ateEventosDetalhe_bindCameraHandlers()

        $('td[tipo=btnSetorManutecao]').on('click', function () {
            const idSetor = $(this).attr('idSetor')
            ateEventosDetalhe_setorManutencao(idSetor)
        })

        var idDisp = (r.dados.length > 0 && r.dados[0].idDispositivo) ? r.dados[0].idDispositivo : (sessionStorage.getItem('ateProDado_proc_idDispositivo') || '')
        if (idDisp) {
            next(idDisp)
        } else {
            ateEventosDetalhe_cache.dadosUltimos25 = []
            $('#ateEventosDetalhe_tabUltimos25 tbody').empty()
        }
    })
}

function ateEventosDetalhe_buscarEventos_montarLinha(item) {
    let setorCor
    let setorTipo
    if (item.idSetor != 'USER' && item.setorManu == "S") {
        setorCor = 'bg-warning'
        setorTipo = 'btnSetorManutecao'
    } else {
        setorCor = 'bg-secondary'
        setorTipo = ''
    }

    const cameraCelula = ateEventosDetalhe_htmlCameraCelula(item)

    $('#ateEventosDetalhe_tabEventos tbody').append(`
        <tr class="table-line">
            <td 
                codigofull="${item.codigoFull}"
                tipo="btnEvento" 
                class="click fmt-tbody"
                idProcesso="${item.idProcesso}"
                zonaUser="${item.zonaUser}"
                style="width: 8%;"
            >${item.quantidade}</td>
             
            <td 
                tipo="${setorTipo}" 
                class="${setorCor} click" 
                idSetor="${item.idSetor}"        
                title="Coloca setor em manutenção"
                style="width: 6%;"
            ><i class="bi bi-tools txt-tbody"></i></td>

            ${cameraCelula}
            
            <td 
                codigofull="${item.codigoFull}"
                tipo="btnEvento" 
                class='click fmt-tbody' 
                idProcesso="${item.idProcesso}"
                zonaUser="${item.zonaUser}"
                style="width: 7%;"
            >${item.tipo + item.codigo}</td>
            
            <td 
                codigofull="${item.codigoFull}"
                tipo="btnEvento" 
                class='click fmt-tbody' 
                idProcesso="${item.idProcesso}"
                zonaUser="${item.zonaUser}"
                style="width: 39%;"
            >${item.descricao}</td>
                        
            <td 
                codigofull="${item.codigoFull}"
                tipo="btnEvento" 
                class='click fmt-tbody' 
                idProcesso="${item.idProcesso}"
                zonaUser="${item.zonaUser}"
                style="width: 7%;"
            >${item.zonaUser}</td>
            
            <td 
                codigofull="${item.codigoFull}"
                tipo="btnEvento" 
                class='click fmt-tbody' 
                idProcesso="${item.idProcesso}"
                zonaUser="${item.zonaUser}"
                style="width: 39%;"
            >${item.descZonUse}</td>

            <td 
                style="width: 8%;"
                class="fmt-tbody"
            >${item.hora || ''}</td>

        </tr>
    `)
}

function ateEventosDetalhe_buscarUltimos25(idDispositivo) {

    $.ajax({
        url: 'ateEventosDetalhe/buscarUltimos25',
        method: 'Post',
        cache: false,
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (sessionStorage.getItem('ateProDado_proc_idProcesso') !== ateEventosDetalhe_reqIdProcesso) return

        $('#ateEventosDetalhe_tabUltimos25 tbody').empty()
        ateEventosDetalhe_cache.dadosUltimos25 = (r.status != 'Vazio' && r.dados) ? r.dados : []

        if (r.status != 'Vazio') {
            qtd = 1
            r.dados.forEach(item => {
                ateEventosDetalhe_buscarUltimos25_montarLinha(item, qtd)
                qtd++
            });

            ateEventosDetalhe_bindCameraHandlers()

            $('tr[tipo=btnUltimos25]').on('click', function () {
                var modal = new bootstrap.Modal(document.getElementById('descricaoAtendimento'), {
                    keyboard: false
                })
                $('#descricaoAtendimentoMsg').val($(this).attr('descAtendimento'))
                modal.show()
            })

        }
    })

}

function ateEventosDetalhe_buscarUltimos25_montarLinha(item, numero) {
    const camCol = ateEventosDetalhe_htmlCameraUltimos25(item)

    $('#ateEventosDetalhe_tabUltimos25 tbody').append(`
        <tr tipo="btnUltimos25" class="click" descAtendimento="${item.descrAtendimento}">
            <td 
                style="width: 7%;"
                class="fmt-tbody"            
            >${numero}</td>

            <td 
                style="width: 7%;"
                class="fmt-tbody"            
            >${item.codigo}</td>

            <td 
                style="width: 30%;"
                class="fmt-tbody"
            >${item.descricao}</td>

            <td 
                style="width: 26%;"
                class="fmt-tbody"
            >${item.descZonUse}</td>

            <td 
                style="width: 17%;"
                class="fmt-tbody"
            >${item.entrada}</td>

            ${camCol}
        </tr>   
    `)
}

function ateEventosDetalhe_setorManutencao(idSetor) {
    const nick = sessionStorage.getItem('login_userNick')

    if (idSetor != 'USER') {
        $.ajax({
            url: 'ateEventosDetalhe/colocaSetorManutecao',
            method: 'Post',
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
                "Authorization": "Bearer " + sessionStorage.getItem('token')
            },
            data: JSON.stringify({
                idSetor: idSetor,
                operador: nick
            })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            msgSucesso()
            ateEventosDetalhe_buscarEventos(idProcesso, (idDispositivo) => {
                ateEventosDetalhe_buscarUltimos25(idDispositivo)
            })
        })
    }
}

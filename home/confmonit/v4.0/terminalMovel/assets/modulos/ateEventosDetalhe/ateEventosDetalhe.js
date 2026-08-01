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
    // Botão manual (se já existir no DOM)
    const $btn = $('#ateEventosDetalhe_btnRefreshPendentes')
    if ($btn && $btn.length > 0) {
        $btn.off('click').on('click', function () {
            ateEventosDetalhe_refreshPendentes()
        })
    }

    const $btnFechar = $('#ateEventosDetalhe_btnFechar')
    if ($btnFechar && $btnFechar.length > 0) {
        $btnFechar.off('click').on('click', function () {
            ateEventosDetalhe_fechar()
        })
    }

    // Evita criar múltiplos intervals ao voltar pra tela
    if (ateEventosDetalhe_intervalPendentes) return

    // Atualização automática
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

    $('[tipo=btnVisualizarCamera]').off('click').on('click', function (e) {
        e.stopPropagation()
        confvision_visualizarCameraBenuvem(confvision_itemFromCell($(this)))
    })

    $('[tipo=btnBenuvemMidiaUlt25]').off('click').on('click', function (e) {
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

function ateEventosDetalhe_htmlCameraBtn(item) {
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
        return `<button type="button" class="btn btn-sm btn-secondary ateEvt-actionBtn" disabled title="Sem câmera"><i class="bi bi-x-lg"></i></button>`
    }

    if (confvision_eConfVision(item)) {
        return `
            <button
                type="button"
                tipo="btnConfVisionPainelSetor"
                class="btn btn-sm btn-success click ateEvt-actionBtn"
                ${attrs}
                title="ConfVision — eventos, foto e ao vivo"
            ><i class="bi bi-camera-video-fill"></i> Câmera</button>
        `
    }

    return `
        <button
            type="button"
            tipo="btnVisualizarCamera"
            class="btn btn-sm btn-success click ateEvt-actionBtn"
            ${attrs}
            title="Exibição em tempo real (Benuvem)"
        ><i class="bi bi-eye-fill"></i> Câmera</button>
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
            ><i class="bi bi-camera-video-fill"></i></td>
        `
    }

    return `
        <td class="text-center click" img="${(item.img || '').replace(/"/g, '&quot;')}" tipo="btnBenuvemMidiaUlt25" title="Ver gravação">
            <i class="bi bi-play-circle-fill text-primary"></i>
        </td>
    `
}

function ateEventosDetalhe_fechar() {
    if (typeof ateControles_voltarPrincipalAtendimento === 'function') {
        ateControles_voltarPrincipalAtendimento()
        return
    }
    if (typeof terminalMovel_mostrarAtendimentoPrincipal === 'function') {
        terminalMovel_mostrarAtendimentoPrincipal()
    }
}

function ateEventosDetalhe_refreshPendentes() {
    // Se a tela não estiver montada (troca de tela), não dispara requests
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

        // Atualiza cache para renderização imediata quando necessário
        ateEventosDetalhe_cache.dadosEventos = r.dados

        // Atualiza SOMENTE a tabela de pendentes (sem mexer em "Últimos 25")
        $('#ateEventosDetalhe_tabEventos tbody').empty()
        r.dados.forEach(item => {
            ateEventosDetalhe_buscarEventos_montarLinha(item)
        })

        $('[tipo=btnEvento], [tipo=btnEventoDesagrupado]').off('click').on('click', function () {
            modEvtDesagrupado_start(
                $(this).attr('idProcesso'),
                $(this).attr('codigoFull'),
                $(this).attr('zonaUser'),
                ateEventosDetalhe_start
            )
        })

        $('[tipo=btnSetorManutecao]').off('click').on('click', function () {
            const idSetor = $(this).attr('idSetor')
            ateEventosDetalhe_setorManutencao(idSetor)
        })

        ateEventosDetalhe_bindCameraHandlers()
    }).always(function () {
        ateEventosDetalhe_refreshPendentesLock = false
    })
}

function ateEventosDetalhe_start() {
    var idProcessoAtual = sessionStorage.getItem('ateProDado_proc_idProcesso')
    var box = document.getElementById('boxDir')

    // Garante refresh manual/automático da tabela de pendentes
    ateEventosDetalhe_setupRefreshPendentes()

    if (ateEventosDetalhe_cache.idProcesso === idProcessoAtual && ateEventosDetalhe_cache.html) {
        box.innerHTML = ateEventosDetalhe_cache.html
        idProcesso = idProcessoAtual
        ateEventosDetalhe_reqIdProcesso = idProcessoAtual
        ateEventosDetalhe_setupRefreshPendentes()
        ateEventosDetalhe_renderFromCache()
        confvision_carregarConfig()
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
    // Sempre renderiza o que tiver no cache; se eventos vazios, ainda mostra últimos 25
    $('#ateEventosDetalhe_tabEventos tbody').empty()
    if (cache.dadosEventos && cache.dadosEventos.length > 0) {
        cache.dadosEventos.forEach(function (item) {
            ateEventosDetalhe_buscarEventos_montarLinha(item)
        })
    }

    if (cache.dadosEventos && cache.dadosEventos.length > 0) {
        $('[tipo=btnEvento], [tipo=btnEventoDesagrupado]').off('click').on('click', function () {
            modEvtDesagrupado_start(
                $(this).attr('idProcesso'),
                $(this).attr('codigoFull'),
                $(this).attr('zonaUser'),
                ateEventosDetalhe_start
            )
        })
        $('[tipo=btnSetorManutecao]').off('click').on('click', function () {
            var idSetor = $(this).attr('idSetor')
            ateEventosDetalhe_setorManutencao(idSetor)
        })
        ateEventosDetalhe_bindCameraHandlers()
    }

    $('#ateEventosDetalhe_tabUltimos25 tbody').empty()
    if (cache.dadosUltimos25 && cache.dadosUltimos25.length > 0) {
        var qtd = 1
        cache.dadosUltimos25.forEach(function (item) {
            ateEventosDetalhe_buscarUltimos25_montarLinha(item, qtd)
            qtd++
        })
        $('tr[tipo=btnUltimos25]').on('click', function () {
            var modal = new bootstrap.Modal(document.getElementById('descricaoAtendimento'), { keyboard: false })
            $('#descricaoAtendimentoMsg').val($(this).attr('descAtendimento'))
            modal.show()
        })
        ateEventosDetalhe_bindCameraHandlers()
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

        $('[tipo=btnEvento], [tipo=btnEventoDesagrupado]').off('click').on('click', function () {
            modEvtDesagrupado_start(
                $(this).attr('idProcesso'),
                $(this).attr('codigoFull'),
                $(this).attr('zonaUser'),
                ateEventosDetalhe_start
            )
        })

        $('[tipo=btnSetorManutecao]').off('click').on('click', function () {
            const idSetor = $(this).attr('idSetor')
            ateEventosDetalhe_setorManutencao(idSetor)
        })

        ateEventosDetalhe_bindCameraHandlers()

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

    const cameraBtn = ateEventosDetalhe_htmlCameraBtn(item)

    $('#ateEventosDetalhe_tabEventos tbody').append(`
        <tr class="table-line ateEvt-itemRow">
            <td
                codigofull="${item.codigoFull}"
                tipo="btnEvento"
                class="click fmt-tbody ateEvt-qtdCell"
                idProcesso="${item.idProcesso}"
                zonaUser="${item.zonaUser}"
            >${item.quantidade}</td>
            <td class="fmt-tbody ateEvt-mainCell">
                <div
                    codigofull="${item.codigoFull}"
                    tipo="btnEvento"
                    class="click ateEvt-mainInfo"
                    idProcesso="${item.idProcesso}"
                    zonaUser="${item.zonaUser}"
                >${item.tipo + item.codigo} - ${item.descricao} - ${item.zonaUser} - ${item.descZonUse} - ${item.hora || ''}</div>
                <div class="ateEvt-actionCell">
                    <button
                        type="button"
                        tipo="btnEventoDesagrupado"
                        class="bg-primary btn btn-sm click ateEvt-actionBtn"
                        codigofull="${item.codigoFull}"
                        idProcesso="${item.idProcesso}"
                        zonaUser="${item.zonaUser}"
                        title="Ver eventos desagrupados"
                    ><i class="bi bi-list-ul"></i> Desagrupados</button>
                    <button
                        type="button"
                        tipo="${setorTipo}"
                        class="${setorCor} btn btn-sm click ateEvt-actionBtn"
                        idSetor="${item.idSetor}"
                        title="Coloca setor em manutenção"
                    ><i class="bi bi-tools"></i> Manutenção</button>
                    ${cameraBtn}
                </div>
            </td>
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
            // let qtd = r.dados.length
            qtd = 1
            r.dados.forEach(item => {
                ateEventosDetalhe_buscarUltimos25_montarLinha(item, qtd)
                qtd++
            });

            $('tr[tipo=btnUltimos25]').on('click', function () {
                var modal = new bootstrap.Modal(document.getElementById('descricaoAtendimento'), {
                    keyboard: false
                })
                $('#descricaoAtendimentoMsg').val($(this).attr('descAtendimento'))
                modal.show()
            })

            ateEventosDetalhe_bindCameraHandlers()

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
                style="width: 28%;"
                class="fmt-tbody"
            >${item.descricao}</td>

            <td 
                style="width: 24%;"
                class="fmt-tbody"
            >${item.descZonUse}</td>

            ${camCol}

            <td 
                style="width: 19%;"
                class="fmt-tbody"
            >${item.entrada}</td>
        </tr>   
    `)
}

function ateEventosDetalhe_visualizarCamera() {
    particao = "${item.particao}"
    setor = "${item.zonaUser}"
    conta = "${item.conta}"
    codBenuvem = "${item.codBenuvem}"
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
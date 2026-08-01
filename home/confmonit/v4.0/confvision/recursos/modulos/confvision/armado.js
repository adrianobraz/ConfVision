/* Armado ConfVision — por dispositivo (não por câmera).
 *
 * analitico_armado_*: evento só se dispositivo.Armado == S.
 *
 * Fabricante CAMERA (id 7):
 *   → simula arme/desarme LOCAL (grava Armado=S/N na tabela)
 *
 * Fabricante != CAMERA:
 *   → envia comando de armar/desarmar para a CENTRAL
 *     (mesmo fluxo /GerenciarClienteArmar), conforme o fabricante
 */
const CV_FABRICANTE_CAMERA = '7'

const CV_LABELS_PLANO = {
    online: 'Câmera online',
    sensor_foto: 'Sensor — foto',
    sensor_foto_video: 'Sensor — foto + vídeo',
    sensor: 'Sensor — foto + vídeo',
    analitico_armado_evento: 'Analítico armado — evento',
    analitico_armado_foto: 'Analítico armado — foto',
    analitico_armado_foto_video: 'Analítico armado — foto + vídeo',
    analitico_armado: 'Analítico armado — foto + vídeo',
    analitico_24h_evento: 'Analítico 24h — evento',
    analitico_24h_foto: 'Analítico 24h — foto',
    analitico_24h_foto_video: 'Analítico 24h — foto + vídeo',
    analitico_24h: 'Analítico 24h — foto + vídeo'
}

const ConfVisionArmado = {
    isFabricanteCamera(idFabricante) {
        return String(idFabricante || '').trim() === CV_FABRICANTE_CAMERA
    },

    isPlanoArmado(plano) {
        const p = String(plano || '').trim().toLowerCase()
        return p.indexOf('analitico_armado') === 0
    },

    isPlanoAnalitico(plano, cam) {
        if (cam && cam.captura_analitico) return true
        const p = String(plano || '').trim().toLowerCase()
        return p.indexOf('analitico_') === 0 || p === 'analitico_armado' || p === 'analitico_24h'
    },

    isAnaliticoPausado(cam) {
        if (!cam) return false
        return cam.analitico_pausado === true || cam.analitico_pausado === 'S' || cam.analitico_pausado === 1
    },

    cameraAnaliticoAtiva(cam) {
        if (!cam) return false
        return cam.ativo === true || cam.ativo === 'S' || cam.ativo === 1 || cam.ativo === '1'
    },

    badgeAnaliticoPausado(cam) {
        if (!ConfVisionArmado.isAnaliticoPausado(cam)) return ''
        return '<span class="cv-badge cv-badge-warn"><i class="bi bi-pause-fill"></i> Detecção pausada</span>'
    },

    labelLicenca(plano) {
        const p = String(plano || '').trim()
        if (!p) return 'Sem licença'
        return CV_LABELS_PLANO[p] || p
    },

    normalizarLista(r) {
        if (Array.isArray(r)) return r
        if (r && Array.isArray(r.dados)) return r.dados
        return []
    },

    escHtml(valor) {
        if (valor == null) return ''
        return String(valor)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
    },

    textoOuVazio(valor, fallback) {
        const s = String(valor == null ? '' : valor).trim()
        return s || (fallback || '')
    },

    badgeArmado(armado) {
        if (armado === 'S') {
            return '<span class="cv-badge cv-badge-on"><i class="bi bi-lock-fill"></i> Armado</span>'
        }
        if (armado === 'N') {
            return '<span class="cv-badge cv-badge-off"><i class="bi bi-unlock-fill"></i> Desarmado</span>'
        }
        return '<span class="cv-badge cv-badge-warn">Sem status</span>'
    },

    carregarCameras() {
        const data = {}
        if (ConfVisionUrls.ehCliente()) {
            data.id_cliente = ConfVisionUrls.idCliente()
            data.id_franqueado = ConfVisionUrls.idFranqueado()
        } else {
            data.id_franqueado = ConfVisionUrls.idFranqueado()
        }
        return $.ajax({
            url: '/api/cameras',
            method: 'GET',
            data: data
        }).then(function (r) {
            return ConfVisionArmado.normalizarLista(r)
        })
    },

    carregarClientes() {
        if (ConfVisionUrls.ehCliente()) {
            return $.Deferred().resolve([{
                idCliente: ConfVisionUrls.idCliente(),
                nome: localStorage.getItem('nomeUsuario') || 'Meu local'
            }]).promise()
        }
        return $.ajax({
            url: '/api/clientes',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ idFranqueado: ConfVisionUrls.idFranqueado() })
        }).then(
            function (r) { return ConfVisionArmado.normalizarLista(r) },
            function () { return [] }
        )
    },

    carregarDispositivosFranqueado() {
        if (ConfVisionUrls.ehCliente()) {
            return $.ajax({
                url: '/api/dispositivos',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ idCliente: ConfVisionUrls.idCliente() })
            }).then(
                function (r) { return ConfVisionArmado.normalizarLista(r) },
                function () { return [] }
            )
        }
        return $.ajax({
            url: '/api/dispositivos/por-franqueado',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ idFranqueado: ConfVisionUrls.idFranqueado() })
        }).then(
            function (r) { return ConfVisionArmado.normalizarLista(r) },
            function () { return [] }
        )
    },

    mapaDispositivos(lista) {
        const map = {}
        ;(lista || []).forEach(function (d) {
            const id = String(d.idDispositivo || '').trim()
            if (id) map[id] = d
        })
        return map
    },

    mapaClientes(lista) {
        const map = {}
        ;(lista || []).forEach(function (c) {
            const id = String(c.idCliente || c.id_cliente || c.id || '').trim()
            if (id) {
                map[id] = c.nome || c.nomeCliente || c.razaoSocial || ('Cliente #' + id)
            }
        })
        return map
    },

    enriquecerCameras(cameras, dispositivosMap, clientesMap) {
        return (cameras || []).map(function (cam) {
            const idDisp = String(cam.id_dispositivo || '').trim()
            const disp = dispositivosMap[idDisp] || null
            const idCliente = String(cam.id_cliente || '').trim()
            const armado = disp ? String(disp.armado || '').trim().toUpperCase() : ''
            const idFabricante = disp ? String(disp.idFabricante || '').trim() : ''
            return Object.assign({}, cam, {
                _id_dispositivo: idDisp,
                _dispositivo: disp,
                _armado: armado,
                _id_fabricante: idFabricante,
                _is_camera_fab: ConfVisionArmado.isFabricanteCamera(idFabricante),
                _somente_armado: !!cam.somente_armado || ConfVisionArmado.isPlanoArmado(cam.plano),
                _particao: disp ? String(disp.particao || '1').trim() : '1',
                _senha: disp ? String(disp.senha || '').trim() : '',
                _nome_cliente: clientesMap[idCliente] || (disp && disp.nomeCliente) || ('Cliente #' + (idCliente || '?')),
                _nome_dispositivo: (disp && (disp.nome || disp.conta)) || (idDisp ? ('Disp. ' + idDisp) : 'Sem dispositivo')
            })
        })
    },

    agruparPorCliente(cameras) {
        const grupos = {}
        ;(cameras || []).forEach(function (cam) {
            const idCliente = String(cam.id_cliente || '').trim() || '_sem'
            if (!grupos[idCliente]) {
                grupos[idCliente] = {
                    id_cliente: idCliente,
                    nome: cam._nome_cliente || ('Cliente #' + idCliente),
                    cameras: [],
                    dispositivos: {}
                }
            }
            grupos[idCliente].cameras.push(cam)

            const idDisp = cam._id_dispositivo || '_sem_disp'
            if (!grupos[idCliente].dispositivos[idDisp]) {
                grupos[idCliente].dispositivos[idDisp] = {
                    id_dispositivo: cam._id_dispositivo,
                    nome: cam._nome_dispositivo,
                    armado: cam._armado,
                    is_camera_fab: cam._is_camera_fab,
                    particao: cam._particao || '1',
                    senha: cam._senha || '',
                    exige_armado: false,
                    cameras: []
                }
            }
            const bloco = grupos[idCliente].dispositivos[idDisp]
            bloco.cameras.push(cam)
            if (cam._somente_armado) bloco.exige_armado = true
            if (cam._armado) bloco.armado = cam._armado
            if (cam._is_camera_fab) bloco.is_camera_fab = true
            if (cam._particao) bloco.particao = cam._particao
            if (cam._senha) bloco.senha = cam._senha
        })

        return Object.keys(grupos)
            .map(function (k) {
                const g = grupos[k]
                g.listaDispositivos = Object.keys(g.dispositivos)
                    .map(function (dk) { return g.dispositivos[dk] })
                    .sort(function (a, b) {
                        return String(a.nome).localeCompare(String(b.nome), 'pt-BR')
                    })
                return g
            })
            .sort(function (a, b) {
                return String(a.nome).localeCompare(String(b.nome), 'pt-BR')
            })
    },

    classeStatusDispositivo(disp) {
        return ConfVisionArmado.classeBordaPorEstadoHome(ConfVisionArmado.estadoDispositivoHome(disp))
    },

    /** Prioridade de exibição: desarmado → pausado → armado → despausado → inativo */
    prioridadeEstadoHome(estado) {
        const map = { desarmado: 0, pausado: 1, armado: 2, despausado: 3, inativo: 4, neutro: 5 }
        return map[String(estado || 'neutro')] != null ? map[String(estado || 'neutro')] : 5
    },

    estadoPrincipalCameraHome(cam) {
        if (!ConfVisionArmado.isPlanoAnalitico(cam && cam.plano, cam)) return 'neutro'
        if (!ConfVisionArmado.cameraAnaliticoAtiva(cam)) return 'inativo'

        const somenteArmado = !!cam._somente_armado
        if (somenteArmado && cam._armado === 'N') return 'desarmado'
        if (ConfVisionArmado.isAnaliticoPausado(cam)) return 'pausado'
        if (somenteArmado && cam._armado === 'S') return 'armado'
        return 'despausado'
    },

    estadoDispositivoHome(disp) {
        const estados = (disp && disp.cameras || []).map(function (c) {
            return ConfVisionArmado.estadoPrincipalCameraHome(c)
        })
        const operacionais = estados.filter(function (e) {
            return e !== 'inativo' && e !== 'neutro'
        })
        if (!operacionais.length) {
            if (estados.indexOf('inativo') >= 0) return 'inativo'
            return 'neutro'
        }
        let pior = ConfVisionArmado.piorEstadoHome(operacionais)
        if (disp && disp.exige_armado && disp.armado === 'N') {
            const temAtivaArmada = (disp.cameras || []).some(function (c) {
                return ConfVisionArmado.cameraAnaliticoAtiva(c) && c._somente_armado
            })
            if (temAtivaArmada) return 'desarmado'
        }
        if (disp && disp.exige_armado && disp.armado === 'S' &&
            ConfVisionArmado.prioridadeEstadoHome(pior) > ConfVisionArmado.prioridadeEstadoHome('armado')) {
            return 'armado'
        }
        return pior
    },

    piorEstadoHome(estados) {
        let pior = 'neutro'
        let minP = 5
        ;(estados || []).forEach(function (e) {
            const p = ConfVisionArmado.prioridadeEstadoHome(e)
            if (p < minP) {
                minP = p
                pior = e
            }
        })
        return pior
    },

    classeBordaPorEstadoHome(estado) {
        if (estado === 'inativo') return 'cv-home-disp-inativo'
        if (estado === 'desarmado' || estado === 'pausado') return 'cv-home-disp-alerta'
        if (estado === 'armado' || estado === 'despausado') return 'cv-home-disp-ok'
        return 'cv-home-disp-neutro'
    },

    classeBordaClienteHome(estados) {
        const pior = ConfVisionArmado.piorEstadoHome(estados)
        if (pior === 'inativo') return 'cv-home-cliente-inativo'
        if (pior === 'desarmado' || pior === 'pausado') return 'cv-home-cliente-alerta'
        if (pior === 'armado' || pior === 'despausado') return 'cv-home-cliente-ok'
        return ''
    },

    cameraPassaFiltroHome(cam, filtros) {
        const estado = ConfVisionArmado.estadoPrincipalCameraHome(cam)
        if (estado === 'neutro' || estado === 'inativo') return true
        return !!filtros[estado]
    },

    compararCamerasPorEstadoHome(a, b) {
        const pa = ConfVisionArmado.prioridadeEstadoHome(ConfVisionArmado.estadoPrincipalCameraHome(a))
        const pb = ConfVisionArmado.prioridadeEstadoHome(ConfVisionArmado.estadoPrincipalCameraHome(b))
        if (pa !== pb) return pa - pb
        return String(a.nome || a.id).localeCompare(String(b.nome || b.id), 'pt-BR')
    },

    compararDispositivosPorEstadoHome(a, b) {
        const pa = ConfVisionArmado.prioridadeEstadoHome(ConfVisionArmado.estadoDispositivoHome(a))
        const pb = ConfVisionArmado.prioridadeEstadoHome(ConfVisionArmado.estadoDispositivoHome(b))
        if (pa !== pb) return pa - pb
        return String(a.nome).localeCompare(String(b.nome), 'pt-BR')
    },

    compararClientesPorEstadoHome(a, b) {
        const estadosA = []
        const estadosB = []
        ;(a.listaDispositivos || []).forEach(function (d) {
            estadosA.push(ConfVisionArmado.estadoDispositivoHome(d))
        })
        ;(b.listaDispositivos || []).forEach(function (d) {
            estadosB.push(ConfVisionArmado.estadoDispositivoHome(d))
        })
        const pa = ConfVisionArmado.prioridadeEstadoHome(ConfVisionArmado.piorEstadoHome(estadosA))
        const pb = ConfVisionArmado.prioridadeEstadoHome(ConfVisionArmado.piorEstadoHome(estadosB))
        if (pa !== pb) return pa - pb
        return String(a.nome).localeCompare(String(b.nome), 'pt-BR')
    },

    _mostrarErro(msg) {
        CvMsg.aviso(msg)
    },

    _mostrarSucesso(msg) {
        CvMsg.sucesso(msg)
    },

    /** CAMERA: simula arme/desarme local na tabela dispositivo.Armado */
    setArmadoLocal(idDispositivo, armado) {
        const status = String(armado || '').toUpperCase() === 'S' ? 'S' : 'N'
        return $.ajax({
            url: '/api/dispositivos/set-armado',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                idDispositivo: String(idDispositivo),
                armado: status
            })
        })
    },

    /** != CAMERA: comando para a central (mesmo endpoint do Gerenciar Cliente) */
    setArmadoCentral(idDispositivo, particao, senha, armadoAtual) {
        const status = String(armadoAtual || '').toUpperCase()
        // acao 1 = armar, 0 = desarmar (inverte o status atual)
        const acao = status === 'S' ? 0 : 1
        return $.ajax({
            url: '/GerenciarClienteArmar',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                idDispositivo: String(idDispositivo),
                numero: Number(particao || 1),
                acao: Number(acao),
                usuario: '000',
                senha: String(senha || ''),
                senhaWeb: 'WHdQkY&RX%W%4RArwm1Q'
            })
        })
    },

    statusAposComandoCentral(response, particao) {
        try {
            const parts = (response && response.dados && response.dados.particao) || []
            const part = parts.filter(function (d) {
                return String(d.numero) === String(particao)
            })[0]
            if (part && part.status) {
                const st = String(part.status).toUpperCase()
                if (st.indexOf('DESARM') >= 0) return 'N'
                if (st.indexOf('ARM') >= 0) return 'S'
            }
        } catch (e) { /* ignore */ }
        return null
    },

    toggleArmado($btn, idDispositivo, armadoAtual, onDone) {
        if (!$btn || $btn.prop('disabled')) return

        const idDisp = String(idDispositivo || $btn.attr('data-armar-dispositivo') || '').trim()
        if (!idDisp) {
            this._mostrarErro('Dispositivo não identificado.')
            return
        }

        const modo = String($btn.attr('data-armar-modo') || 'local').trim().toLowerCase()
        const atual = String(
            armadoAtual || $btn.attr('data-armar-status') || ''
        ).trim().toUpperCase()
        const novo = atual === 'S' ? 'N' : 'S'
        const label = novo === 'S' ? 'Armar' : 'Desarmar'

        const msgConfirm = modo === 'central'
            ? label + ' este dispositivo na central?'
            : label + ' este dispositivo? (simulação local CAMERA)'
        CvMsg.confirmar(label + '?', msgConfirm).then(function (r) {
        if (!r.isConfirmed) return

        const htmlOriginal = $btn.html()
        $btn.prop('disabled', true).addClass('cv-btn-disabled')
            .html('<i class="bi bi-arrow-repeat cv-snapshot-spin"></i>')

        const restaurar = function () {
            $btn.prop('disabled', false).removeClass('cv-btn-disabled').html(htmlOriginal)
        }

        if (modo === 'central') {
            const particao = $btn.attr('data-armar-particao') || '1'
            const senha = $btn.attr('data-armar-senha') || ''
            if (typeof boxProcessando === 'function') {
                boxProcessando('Enviando comando à central…')
            }

            ConfVisionArmado.setArmadoCentral(idDisp, particao, senha, atual)
                .done(function (r) {
                    if (typeof boxFechar === 'function') boxFechar()
                    let statusFinal = ConfVisionArmado.statusAposComandoCentral(r, particao)
                    if (!statusFinal) statusFinal = novo
                    ConfVisionArmado._mostrarSucesso(
                        statusFinal === 'S' ? 'Dispositivo armado na central' : 'Dispositivo desarmado na central'
                    )
                    if (typeof onDone === 'function') onDone(statusFinal)
                })
                .fail(function (xhr) {
                    if (typeof boxFechar === 'function') boxFechar()
                    console.error('[ARMADO] comando central falhou', xhr)
                    let msg = 'Erro ao armar/desarmar na central'
                    try {
                        const st = (xhr.responseJSON && xhr.responseJSON.status) || ''
                        if (String(st).indexOf('nao conectado') >= 0) {
                            msg = 'Dispositivo desconectado do sistema'
                        } else if (String(st).indexOf('não implementado') >= 0 || String(st).indexOf('nao implementado') >= 0) {
                            msg = 'Comando não implementado para este fabricante'
                        } else if (st) {
                            msg = st
                        }
                    } catch (e) { /* ignore */ }
                    ConfVisionArmado._mostrarErro(msg)
                    restaurar()
                })
            return
        }

        // CAMERA — simulação local
        ConfVisionArmado.setArmadoLocal(idDisp, novo)
            .done(function () {
                ConfVisionArmado._mostrarSucesso(
                    novo === 'S' ? 'Dispositivo armado (local)' : 'Dispositivo desarmado (local)'
                )
                if (typeof onDone === 'function') onDone(novo)
            })
            .fail(function (xhr) {
                console.error('[ARMADO] set-armado local falhou', xhr && xhr.status, xhr)
                let msg = 'Erro ao alterar armado local'
                if (xhr && xhr.status === 404) {
                    msg = 'API de armado não encontrada. Faça o deploy (rebuild) do servidor.'
                } else {
                    try {
                        const j = xhr.responseJSON
                        if (j && (j.status || j.erro || j.message)) {
                            msg = j.status || j.erro || j.message
                        }
                    } catch (e) { /* ignore */ }
                }
                ConfVisionArmado._mostrarErro(msg)
                restaurar()
            })
        })
    },

    htmlBotaoArmarDispositivo(disp) {
        if (!disp.exige_armado) {
            return '<span class="cv-armado-hint">Licença 24h</span>'
        }
        if (!disp.id_dispositivo) {
            return '<span class="cv-armado-hint">Sem dispositivo</span>'
        }

        const armado = disp.armado === 'S' ? 'S' : 'N'
        const cls = 'cv-btn-armar'
        const icon = armado === 'S' ? 'bi-lock-fill' : 'bi-unlock-fill'
        const label = armado === 'S' ? 'Desarmar' : 'Armar'
        const modo = disp.is_camera_fab ? 'local' : 'central'
        const titulo = modo === 'local'
            ? label + ' (simulação CAMERA)'
            : label + ' (comando central)'

        let extra = ''
        if (modo === 'central') {
            extra =
                ' data-armar-particao="' + ConfVisionArmado.escHtml(disp.particao || '1') + '"' +
                ' data-armar-senha="' + ConfVisionArmado.escHtml(disp.senha || '') + '"'
        }

        const hint = modo === 'local'
            ? ''
            : '<span class="cv-armado-hint">Via central</span>'

        return (
            '<div class="cv-armado-central">' +
            '<button type="button" class="' + cls + '" ' +
            'data-armar-dispositivo="' + ConfVisionArmado.escHtml(disp.id_dispositivo) + '" ' +
            'data-armar-status="' + armado + '" ' +
            'data-armar-modo="' + modo + '"' +
            extra +
            ' title="' + titulo + '">' +
            '<i class="bi ' + icon + '"></i> ' + label +
            '</button>' +
            hint +
            '</div>'
        )
    },

    htmlBotaoArmar(cam) {
        return ConfVisionArmado.htmlBotaoArmarDispositivo({
            id_dispositivo: cam._id_dispositivo,
            armado: cam._armado,
            is_camera_fab: cam._is_camera_fab,
            exige_armado: cam._somente_armado,
            particao: cam._particao,
            senha: cam._senha
        })
    },

    dispositivoComCamerasAtivas(disp) {
        return (disp && disp.cameras || []).some(function (c) {
            return ConfVisionArmado.cameraAnaliticoAtiva(c)
        })
    },

    badgeCameraInativaHome(cam) {
        if (ConfVisionArmado.cameraAnaliticoAtiva(cam)) return ''
        return ' <span class="cv-badge cv-badge-muted"><i class="bi bi-slash-circle"></i> Inativa</span>'
    },

    htmlBotaoArmarDispositivoHome(disp) {
        if (!ConfVisionArmado.dispositivoComCamerasAtivas(disp)) {
            return '<span class="cv-armado-hint cv-home-desabilitado">Câmeras inativas</span>'
        }
        return ConfVisionArmado.htmlBotaoArmarDispositivo(disp)
    },

    htmlAcoesCameraHome(cam, ehCliente) {
        if (!ConfVisionArmado.cameraAnaliticoAtiva(cam)) {
            return '<span class="cv-armado-hint cv-home-desabilitado">Sem ações</span>'
        }
        if (ehCliente) {
            return (
                '<a href="/ao-vivo/' + encodeURIComponent(cam.id) +
                '" class="cv-btn-ghost cv-btn-sm" title="Ao vivo">' +
                '<i class="bi bi-broadcast"></i> Ao vivo</a>'
            )
        }
        return (
            '<a href="/cameras/editar/' + encodeURIComponent(cam.id) +
            '" class="cv-btn-ghost cv-btn-sm" title="Editar">' +
            '<i class="bi bi-pencil"></i> Editar</a>' +
            ConfVisionArmado.htmlBotaoPausarAnalitico(cam)
        )
    },

    htmlBotaoPausarAnalitico(cam) {
        if (!ConfVisionArmado.isPlanoAnalitico(cam.plano, cam)) return ''
        if (!ConfVisionArmado.cameraAnaliticoAtiva(cam)) {
            return '<span class="cv-armado-hint">Inativa</span>'
        }
        const pausado = ConfVisionArmado.isAnaliticoPausado(cam)
        const icon = pausado ? 'bi-play-fill' : 'bi-pause-fill'
        const label = pausado ? 'Retomar' : 'Pausar'
        const titulo = pausado
            ? 'Retomar detecção analítica (worker)'
            : 'Pausar detecção analítica (para o worker)'

        return (
            '<button type="button" class="cv-btn-pausar-analitico" ' +
            'data-pausar-camera="' + ConfVisionArmado.escHtml(cam.id) + '" ' +
            'data-pausar-estado="' + (pausado ? '1' : '0') + '" ' +
            'title="' + titulo + '">' +
            '<i class="bi ' + icon + '"></i> ' + label +
            '</button>'
        )
    },

    togglePausarAnalitico(cameraId, pausar, $btn, onDone) {
        const id = String(cameraId || '').trim()
        if (!id) return

        if ($btn && $btn.length) {
            $btn.prop('disabled', true)
        }

        $.ajax({
            url: '/api/cameras/' + encodeURIComponent(id) + '/analitico/pausar',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ pausado: !!pausar })
        }).done(function (r) {
            const msg = (r && r.message) ? r.message : (pausar ? 'Detecção pausada' : 'Detecção retomada')
            if (typeof boxSucessoAuto === 'function') {
                boxSucessoAuto(msg)
            }
            if (typeof onDone === 'function') onDone(null, r)
        }).fail(function (xhr) {
            let msg = 'Não foi possível alterar a detecção.'
            try {
                const j = xhr.responseJSON || JSON.parse(xhr.responseText)
                if (j && (j.message || j.status || j.erro)) {
                    msg = j.message || j.status || j.erro
                }
            } catch (e) { /* ignore */ }
            if (typeof CvMsg !== 'undefined') {
                CvMsg.aviso(msg)
            } else {
                boxMesagemAtencaoPersonalizada(msg)
            }
            if (typeof onDone === 'function') onDone(xhr)
        }).always(function () {
            if ($btn && $btn.length) {
                $btn.prop('disabled', false)
            }
        })
    }
}

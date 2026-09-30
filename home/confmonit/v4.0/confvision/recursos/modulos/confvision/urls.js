const ConfVisionUrls = {
    hlsBase() {
        return ($('#cfg-hls-base').val() || '').replace(/\/$/, '')
    },
    rtmpBase() {
        return ($('#cfg-rtmp-base').val() || '').replace(/\/$/, '')
    },
    hlsPublicBase() {
        const cfg = ($('#cfg-hls-public').val() || '').trim()
        const pub = (window.ConfVisionStreamPublic && window.ConfVisionStreamPublic.hls || '').trim()
        return (pub || cfg || this.hlsBase()).replace(/\/$/, '')
    },
    rtmpPublicBase() {
        const cfg = ($('#cfg-rtmp-public').val() || '').trim()
        const pub = (window.ConfVisionStreamPublic && window.ConfVisionStreamPublic.rtmp || '').trim()
        return (pub || cfg || this.rtmpBase()).replace(/\/$/, '')
    },
    streamPath(cameraId) {
        // Fallback display; preferir path da API (/rtmp-publish).
        return ''
    },
    streamPathFromChave(chave) {
        let c = String(chave || '').trim().replace(/^\/+|\/+$/g, '')
        if (!c) return ''
        c = c.replace(/^live\//, '')
        if (c.startsWith('cam/')) return c
        return `cam/${c}`
    },
    hlsUrl(cameraId) {
        const p = this.streamPath(cameraId)
        if (!p) return ''
        return `${this.hlsPublicBase()}/${p}/index.m3u8`
    },
    hlsUrlFromPath(path) {
        const p = this.streamPathFromChave(path)
        if (!p) return ''
        return `${this.hlsPublicBase()}/${p}/index.m3u8`
    },
    hlsPublicUrl(cameraId) {
        return this.hlsUrl(cameraId)
    },
    rtmpPublishUrl(cameraId) {
        const p = this.streamPath(cameraId)
        if (!p) return this.rtmpPublicBase()
        return `${this.rtmpPublicBase()}/${p}`
    },
    // Monta URL a partir da resposta /rtmp-publish (cam/{hash}, sem query).
    rtmpPublishUrlAuth(cameraId, idFranqueado, token, tipo) {
        const base = this.rtmpPublicBase()
        const path = this.streamPathFromChave(token)
        if (!base || !path) return ''
        const t = this.normalizarTipoCamera(tipo)
        return t === 'dvr'
            ? `${base}/${path}/`
            : `${base}/${path}`
    },
    carregarRtmpPublish(cameraId) {
        const id = String(cameraId || '').trim()
        if (!id) return $.Deferred().reject('sem id').promise()
        return $.get(`/api/cameras/${encodeURIComponent(id)}/rtmp-publish`)
    },
    // Alguns equipamentos Cabeado IP precisam da barra no final no campo do aparelho (enviam sem / ao MediaMTX).
    rtmpPublishUrlDvr(cameraId) {
        return `${this.rtmpPublishUrl(cameraId)}/`
    },
    // tipo: wifi | dvr (gravado em vis_camera.protocolo — rótulos: Wifi | Cabeado IP)
    rtmpPublishUrlPorTipo(cameraId, tipo) {
        const t = String(tipo || '').toLowerCase()
        if (t === 'dvr') return this.rtmpPublishUrlDvr(cameraId)
        return this.rtmpPublishUrl(cameraId)
    },
    normalizarTipoCamera(valor) {
        const t = String(valor || '').toLowerCase().trim()
        if (t === 'dvr' || t === 'wifi') return t
        return ''
    },
    rotuloTipoCamera(valor) {
        const t = this.normalizarTipoCamera(valor)
        if (t === 'dvr') return 'Cabeado IP'
        if (t === 'wifi') return 'Wifi'
        return '—'
    },
    idFranqueado() {
        var id = localStorage.getItem('idFranqueado') || ''
        if (!id) {
            var el = document.getElementById('id-franqueado')
            if (el && el.value) id = el.value
        }
        return id || ''
    },
    idCliente() {
        return localStorage.getItem('idCliente') || ''
    },
    papel() {
        return (localStorage.getItem('papel') || 'FRA').toUpperCase()
    },
    ehCliente() {
        return this.papel() === 'CLI'
    },
    camerasQuery() {
        if (this.ehCliente()) {
            return {
                id_cliente: this.idCliente(),
                id_franqueado: this.idFranqueado()
            }
        }
        return { id_franqueado: this.idFranqueado() }
    },
    camerasQueryString() {
        const q = this.camerasQuery()
        return Object.keys(q).map(function (k) {
            return encodeURIComponent(k) + '=' + encodeURIComponent(q[k] || '')
        }).join('&')
    }
}

function confVisionCameraPodeStream(cam) {
    if (!cam) return { ok: false, motivo: 'camera_invalida' }
    if (cam.bloqueado === true || cam.bloqueado === 'S' || cam.bloqueado === 1 || cam.bloqueado === '1') {
        return { ok: false, motivo: 'camera_bloqueada' }
    }
    const plano = String(cam.plano || '').trim().toLowerCase()
    const ativo = cam.ativo === true || cam.ativo === 'S' || cam.ativo === 1 || cam.ativo === '1'
    if (plano === 'online' || ativo) return { ok: true, motivo: '' }
    return { ok: false, motivo: 'camera_inativa' }
}

function confVisionMotivoStream(cam) {
    const r = confVisionCameraPodeStream(cam)
    if (r.ok) return ''
    if (r.motivo === 'camera_bloqueada') {
        return 'Câmera suspensa pelo administrativo (pagamento).'
    }
    if (r.motivo === 'camera_inativa') {
        return 'Câmera desligada. Ative no cadastro.'
    }
    return 'Stream indisponível.'
}

function confVisionCameraApareceAoVivo(cam) {
    if (!cam || !cam.id) return false
    return confVisionCameraPodeStream(cam).ok
}

function confVisionAuthGuard() {
    if (!localStorage.getItem('token')) {
        // Sem token no localStorage: ir ao logout (limpa cookie) evita loop
        // login←→home quando o cookie de sessão ainda existe.
        var path = window.location.pathname || ''
        if (path === '/login' || path === '/logout') return
        window.location = '/logout'
        return
    }
    if (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.ehCliente()) {
        const path = window.location.pathname
        const permitidos = [
            '/carregar-menu-confvision',
            '/eventos',
            '/ao-vivo',
            '/mosaicos',
            '/gravacoes/timeline',
            '/gravacoes/dvr',
            '/relatorio-armado'
        ]
        const ok = permitidos.some(function (p) {
            return path === p || path.indexOf(p + '/') === 0
        })
        if (!ok && path !== '/carregar-menu-confvision') {
            window.location = '/carregar-menu-confvision'
        }
    }
}

function confVisionAuthHeaders() {
    const token = localStorage.getItem('token') || ''
    if (!token) return {}
    return { Authorization: 'Bearer ' + token }
}

function confVisionInstalarAjaxAuth() {
    if (typeof $ === 'undefined' || $.ajaxSetup._confmonitAuth) return
    $.ajaxSetup({
        beforeSend: function (xhr) {
            const token = localStorage.getItem('token')
            if (token) {
                xhr.setRequestHeader('Authorization', 'Bearer ' + token)
            }
        }
    })
    $.ajaxSetup._confmonitAuth = true
}

function confVisionCarregarCabecalho() {
    $('#display-nome-usuario').text(localStorage.getItem('nomeUsuario') || '')
    var nome = (localStorage.getItem('nomeFranqueado') || '').trim()
    $('#display-nome-franqueado').each(function () {
        var $el = $(this)
        var suffix = String($el.data('eyebrowSuffix') || $el.attr('data-eyebrow-suffix') || '').trim()
        if (suffix && nome) {
            $el.text(nome + ' · ' + suffix)
        } else if (suffix) {
            $el.text(suffix)
        } else {
            $el.text(nome)
        }
    })
    $('#id-franqueado').val(ConfVisionUrls.idFranqueado())
    if (typeof CvWhitelabel !== 'undefined' && CvWhitelabel.applyBrandToDocument) {
        CvWhitelabel.applyBrandToDocument()
    }
}

function confVisionFormatarData(valor) {
    if (!valor) return '—'
    return moment(valor).format('DD/MM/YYYY HH:mm:ss')
}

function confVisionSimNao(valor) {
    return valor ? 'Sim' : 'Não'
}

function confVisionGarantirModalAjudaEncodeRtmp() {
    if (document.getElementById('modal-ajuda-encode-rtmp')) return
    const html = `
    <div class="modal fade" id="modal-ajuda-encode-rtmp" tabindex="-1" aria-labelledby="modal-ajuda-encode-rtmp-titulo" aria-hidden="true">
        <div class="modal-dialog modal-lg modal-dialog-scrollable modal-dialog-centered">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="modal-ajuda-encode-rtmp-titulo">
                        <i class="bi bi-question-circle"></i> Configuração recomendada (RTMP)
                    </h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Fechar"></button>
                </div>
                <div class="modal-body cv-ajuda-encode-body">
                    <section class="cv-ajuda-encode-sec">
                        <h6>URL de publicação</h6>
                        <p class="cv-ajuda-encode-note">No cadastro, o campo <strong>Tipo</strong> (Wifi ou Cabeado IP) é obrigatório e define a URL.</p>
                        <ul class="cv-ajuda-encode-list">
                            <li><strong>Wifi:</strong> sem barra (<code>/</code>) no final<br>
                                <code id="ajuda-rtmp-url-ip" class="cv-ajuda-encode-code"></code>
                            </li>
                            <li><strong>Cabeado IP:</strong> com barra (<code>/</code>) no final (alguns aparelhos só publicam assim)<br>
                                <code id="ajuda-rtmp-url-dvr" class="cv-ajuda-encode-code"></code>
                            </li>
                        </ul>
                    </section>

                    <section class="cv-ajuda-encode-sec">
                        <h6>Codec</h6>
                        <p class="mb-1">Use <strong>H.264</strong> (mais estável no Ao Vivo / HLS).</p>
                        <p class="cv-ajuda-encode-note mb-0">
                            H.265 pode funcionar em alguns equipamentos, mas costuma gerar mais queda e incompatibilidade no player.
                        </p>
                    </section>

                    <section class="cv-ajuda-encode-sec">
                        <h6>Encode (Stream Principal)</h6>
                        <p class="cv-ajuda-encode-note">Menu típico: Vídeo / Encode / Compressão / Stream Principal</p>
                        <div class="table-responsive">
                            <table class="table table-sm cv-ajuda-encode-table mb-0">
                                <thead>
                                    <tr><th>Campo</th><th>Valor recomendado</th></tr>
                                </thead>
                                <tbody>
                                    <tr><td>Codec</td><td>H.264 (não H.265)</td></tr>
                                    <tr><td>Resolução</td><td>1280×720 (ou 1920×1080 se a internet for boa)</td></tr>
                                    <tr><td>FPS</td><td>15 (ou 25, mas fixo)</td></tr>
                                    <tr><td>Bitrate</td><td>1024–2048 Kbps (1–2 Mbps)</td></tr>
                                    <tr><td>Controle de bitrate</td><td>CBR (ou VBR com máximo = 2 Mbps)</td></tr>
                                    <tr><td>Keyframe / I-frame / GOP</td><td>1 ou 2 segundos</td></tr>
                                    <tr><td>Profile</td><td>Main ou Baseline</td></tr>
                                    <tr><td>Áudio</td><td>AAC ou desligado</td></tr>
                                </tbody>
                            </table>
                        </div>
                    </section>

                    <section class="cv-ajuda-encode-sec mb-0">
                        <h6>Como achar Keyframe / GOP</h6>
                        <p class="mb-1">Pode aparecer como:</p>
                        <ul class="cv-ajuda-encode-list">
                            <li>I Frame Interval</li>
                            <li>Keyframe Interval</li>
                            <li>GOP</li>
                            <li>Intervalo de quadro-chave</li>
                        </ul>
                        <p class="cv-ajuda-encode-note mb-0">
                            Fórmula: <code>I-frame ≈ FPS × segundos</code>
                            (ex.: FPS 15 → I-frame 15 = 1s; I-frame 30 = 2s)
                        </p>
                    </section>
                </div>
                <div class="modal-footer">
                    <button type="button" class="cv-btn-primary" data-bs-dismiss="modal">Entendi</button>
                </div>
            </div>
        </div>
    </div>`
    document.body.insertAdjacentHTML('beforeend', html)
}

function confVisionAbrirAjudaEncodeRtmp(cameraId) {
    if (typeof bootstrap === 'undefined') return
    confVisionGarantirModalAjudaEncodeRtmp()

    const fromInput = (($('#inp-rtmp-url').val() || '').replace(/\/$/, '').split('/').pop() || '').trim()
    const id = String(cameraId || ($('#cfg-camera-id').val() || '').trim() || fromInput || '').trim()
    const elIp = document.getElementById('ajuda-rtmp-url-ip')
    const elDvr = document.getElementById('ajuda-rtmp-url-dvr')
    const placeholder = 'rtmp://…/cam/{hash12}'
    if (elIp) elIp.textContent = placeholder
    if (elDvr) elDvr.textContent = 'rtmp://…/cam/{hash12}/'

    const el = document.getElementById('modal-ajuda-encode-rtmp')
    bootstrap.Modal.getOrCreateInstance(el).show()

    if (id && typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.carregarRtmpPublish) {
        ConfVisionUrls.carregarRtmpPublish(id).done(function (r) {
            if (elIp) elIp.textContent = (r && r.wifi) || placeholder
            if (elDvr) elDvr.textContent = (r && r.dvr) || placeholder
        })
    }
}

function confVisionNavAtiva() {
    const path = window.location.pathname
    const cadastroPaths = [
        '/CarregarPaginaGerenciarCliente',
        '/carregar-gerenciar-dispositivo',
        '/carregar-gerenciar-setores-alarme',
        '/grade-horario'
    ]
    const configuracaoPaths = [
        '/integracao-eventos',
        '/carregar-whitelabel',
        '/carregar-dominio',
        '/minhas-licencas',
        '/gravacoes'
    ]
    const monitoramentoPaths = [
        '/ao-vivo',
        '/gravacoes/dvr',
        '/gravacoes/timeline'
    ]
    const relatorioPaths = [
        '/eventos',
        '/relatorio-armado',
        '/relatorio-licencas',
        '/relatorio-faturas',
        '/rtmp-falhas',
        '/ips-banidos'
    ]

    function ativarDropdownItem($item) {
        $item.addClass('cv-nav-active')
        $item.closest('.cv-nav-dropdown').find('.cv-nav-dropdown-toggle').addClass('cv-nav-active')
    }

    $('.cv-nav-link').not('.cv-nav-dropdown-toggle').each(function () {
        const href = $(this).attr('href')
        const ativo = path === href ||
            (href === '/carregar-menu-confvision' && path === '/carregar-menu-confvision')
        if (ativo) $(this).addClass('cv-nav-active')
    })

    if (cadastroPaths.some(function (p) { return path === p || path.startsWith(p + '/') })) {
        $('[data-nav-group="cadastros"] .cv-nav-dropdown-toggle').addClass('cv-nav-active')
    }

    if (configuracaoPaths.some(function (p) { return path === p || path.startsWith(p + '/') || path.startsWith(p) })) {
        $('[data-nav-group="configuracoes"] .cv-nav-dropdown-toggle').addClass('cv-nav-active')
    }

    if (monitoramentoPaths.some(function (p) { return path === p || path.startsWith(p) })) {
        $('[data-nav-group="monitoramento"] .cv-nav-dropdown-toggle').addClass('cv-nav-active')
    }

    if (relatorioPaths.some(function (p) { return path === p || path.startsWith(p) })) {
        $('[data-nav-group="relatorios"] .cv-nav-dropdown-toggle').addClass('cv-nav-active')
    }

    $('.cv-nav-dropdown-item').each(function () {
        const href = $(this).attr('href')
        if (!href) return

        const hrefPath = href.split('?')[0]
        let ativo = path === hrefPath

        if (hrefPath === '/cameras') {
            ativo = path === '/cameras' || path.startsWith('/cameras/')
        } else if (hrefPath === '/grade-horario') {
            ativo = path === '/grade-horario' || path.startsWith('/grade-horario/')
        } else if (hrefPath === '/gravacoes') {
            ativo = path === '/gravacoes'
        } else if (hrefPath === '/gravacoes/timeline') {
            ativo = path.startsWith('/gravacoes/timeline')
        } else if (hrefPath === '/gravacoes/dvr') {
            ativo = path.startsWith('/gravacoes/dvr')
        } else if (hrefPath === '/eventos') {
            ativo = path === '/eventos' || path.startsWith('/eventos/')
        } else if (hrefPath === '/ao-vivo') {
            ativo = path === '/ao-vivo' || path.startsWith('/ao-vivo/')
        } else if (hrefPath === '/relatorio-armado') {
            ativo = path === '/relatorio-armado' || path.startsWith('/relatorio-armado')
        } else if (hrefPath === '/relatorio-licencas') {
            ativo = path === '/relatorio-licencas' || path.startsWith('/relatorio-licencas')
        } else if (hrefPath === '/relatorio-faturas') {
            ativo = path === '/relatorio-faturas' || path.startsWith('/relatorio-faturas')
        } else if (hrefPath === '/rtmp-falhas') {
            ativo = path === '/rtmp-falhas' || path.startsWith('/rtmp-falhas')
        } else if (hrefPath === '/ips-banidos') {
            ativo = path === '/ips-banidos' || path.startsWith('/ips-banidos')
        } else if (hrefPath === '/minhas-licencas' || href.indexOf('/minhas-licencas') === 0) {
            ativo = path === '/minhas-licencas' || path.startsWith('/minhas-licencas')
        } else if (hrefPath === '/carregar-whitelabel') {
            ativo = path.startsWith('/carregar-whitelabel')
        } else if (hrefPath === '/carregar-dominio') {
            ativo = path.startsWith('/carregar-dominio')
        } else if (hrefPath === '/integracao-eventos') {
            ativo = path === '/integracao-eventos' || path.startsWith('/integracao-eventos/')
        }

        if (ativo) ativarDropdownItem($(this))
    })
}

$(document).ready(confVisionNavAtiva)

function confVisionNavMobile() {
    const toggle = document.getElementById('cv-nav-toggle')
    const backdrop = document.getElementById('cv-nav-backdrop')
    const nav = document.getElementById('cv-nav-main')
    if (!toggle || !nav) return

    function ehMobileNav() {
        return window.matchMedia('(max-width: 992px)').matches
    }

    function fecharSubmenus() {
        nav.querySelectorAll('.cv-nav-dropdown').forEach(function (dd) {
            dd.classList.remove('cv-nav-dropdown-open')
            const btn = dd.querySelector('.cv-nav-dropdown-toggle')
            if (btn) btn.setAttribute('aria-expanded', 'false')
        })
    }

    function fechar() {
        document.body.classList.remove('cv-nav-open')
        toggle.setAttribute('aria-expanded', 'false')
        toggle.setAttribute('aria-label', 'Abrir menu')
        if (backdrop) backdrop.hidden = true
        document.body.style.overflow = ''
        fecharSubmenus()
    }

    function abrir() {
        if (!ehMobileNav()) return
        fecharSubmenus()
        document.body.classList.add('cv-nav-open')
        toggle.setAttribute('aria-expanded', 'true')
        toggle.setAttribute('aria-label', 'Fechar menu')
        if (backdrop) backdrop.hidden = false
        document.body.style.overflow = 'hidden'
    }

    function alternarSubmenu(btn) {
        const dd = btn.closest('.cv-nav-dropdown')
        if (!dd) return
        const jaAberto = dd.classList.contains('cv-nav-dropdown-open')
        nav.querySelectorAll('.cv-nav-dropdown').forEach(function (outro) {
            if (outro === dd) return
            outro.classList.remove('cv-nav-dropdown-open')
            const ob = outro.querySelector('.cv-nav-dropdown-toggle')
            if (ob) ob.setAttribute('aria-expanded', 'false')
        })
        if (jaAberto) {
            dd.classList.remove('cv-nav-dropdown-open')
            btn.setAttribute('aria-expanded', 'false')
        } else {
            dd.classList.add('cv-nav-dropdown-open')
            btn.setAttribute('aria-expanded', 'true')
        }
    }

    toggle.addEventListener('click', function (e) {
        e.preventDefault()
        e.stopPropagation()
        if (document.body.classList.contains('cv-nav-open')) fechar()
        else abrir()
    })

    if (backdrop) {
        backdrop.addEventListener('click', function (e) {
            e.preventDefault()
            fechar()
        })
    }

    // Delegação: submenu NÃO fecha o drawer; só links de navegação fecham
    nav.addEventListener('click', function (e) {
        const btnSub = e.target.closest('.cv-nav-dropdown-toggle')
        if (btnSub && nav.contains(btnSub)) {
            if (!ehMobileNav()) return
            e.preventDefault()
            e.stopPropagation()
            alternarSubmenu(btnSub)
            return
        }

        const link = e.target.closest('a.cv-nav-link, a.cv-nav-dropdown-item')
        if (link && nav.contains(link)) {
            fechar()
        }
    })

    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape') fechar()
    })

    window.addEventListener('resize', function () {
        if (!ehMobileNav()) fechar()
    })
}

document.addEventListener('DOMContentLoaded', function () {
    confVisionInstalarAjaxAuth()
    confVisionNavMobile()
    confVisionAplicarNavCliente()
})

function confVisionAplicarNavCliente() {
    if (typeof ConfVisionUrls === 'undefined' || !ConfVisionUrls.ehCliente()) return
    const nav = document.getElementById('cv-nav-main')
    if (!nav) return
    nav.innerHTML = [
        '<a href="/carregar-menu-confvision" class="cv-nav-link"><i class="bi bi-grid-1x2"></i> Início</a>',
        '<a href="/ao-vivo" class="cv-nav-link"><i class="bi bi-broadcast"></i> Ao vivo</a>',
        '<a href="/eventos" class="cv-nav-link"><i class="bi bi-activity"></i> Eventos</a>',
        '<a href="/gravacoes/dvr" class="cv-nav-link"><i class="bi bi-play-btn"></i> DVR</a>',
        '<a href="/gravacoes/timeline" class="cv-nav-link"><i class="bi bi-collection-play"></i> Timeline</a>',
        '<a href="/relatorio-armado" class="cv-nav-link"><i class="bi bi-shield-lock"></i> Armado</a>'
    ].join('')
    confVisionNavAtiva()
}

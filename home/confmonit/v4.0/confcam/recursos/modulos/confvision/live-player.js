/**
 * Player HLS reutilizável (página Ao vivo e aba na edição de câmera).
 */
const ConfVisionLivePlayer = (function () {
    let hlsInstance = null
    let liveSyncTimer = null
    let liveRecheckTimer = null
    let retryTimer = null
    let currentCameraId = null
    let retryCount = 0
    let videoEl = null
    let statusEl = null

    const RETRY_MS = 2500
    const MAX_RETRIES = 40
    const LIVE_RECHECK_MS = 60000

    const HLS_LOW_LATENCY = {
        lowLatencyMode: true,
        enableWorker: true,
        backBufferLength: 0,
        liveBackBufferLength: 0,
        liveSyncDuration: 1,
        liveMaxLatencyDuration: 5,
        liveSyncDurationCount: 2,
        maxLiveSyncPlaybackRate: 1.2,
        maxBufferLength: 8,
        maxMaxBufferLength: 12,
        manifestLoadingTimeOut: 10000,
        manifestLoadingMaxRetry: 8,
        levelLoadingTimeOut: 10000,
        fragLoadingTimeOut: 15000
    }

    function resolveEl(ref) {
        if (!ref) return null
        if (typeof ref === 'string') return document.querySelector(ref)
        return ref
    }

    function setStatus(msg) {
        if (statusEl) statusEl.textContent = msg || ''
    }

    function parar() {
        currentCameraId = null
        retryCount = 0
        setStatus('')
        if (liveRecheckTimer) {
            clearInterval(liveRecheckTimer)
            liveRecheckTimer = null
        }
        if (retryTimer) {
            clearTimeout(retryTimer)
            retryTimer = null
        }
        if (liveSyncTimer) {
            clearInterval(liveSyncTimer)
            liveSyncTimer = null
        }
        if (hlsInstance) {
            hlsInstance.destroy()
            hlsInstance = null
        }
        if (videoEl) {
            videoEl.onplaying = null
            videoEl.onerror = null
            videoEl.removeAttribute('src')
            videoEl.load()
        }
    }

    function agendarRetry(cameraId) {
        if (currentCameraId !== cameraId) return
        if (retryCount >= MAX_RETRIES) {
            setStatus('Stream indisponível. Verifique se a câmera está transmitindo.')
            return
        }
        retryCount++
        setStatus('Conectando… tentativa ' + retryCount)
        clearTimeout(retryTimer)
        retryTimer = setTimeout(function () {
            if (currentCameraId === cameraId) {
                iniciar(cameraId, { isRetry: true })
            }
        }, RETRY_MS)
    }

    function iniciarSyncAoVivo(video) {
        if (liveSyncTimer) clearInterval(liveSyncTimer)
        liveSyncTimer = setInterval(function () {
            if (!hlsInstance || video.paused) return
            const pos = hlsInstance.liveSyncPosition
            if (pos != null && video.currentTime < pos - 2) {
                video.currentTime = Math.max(0, pos - 0.5)
            }
        }, 2000)
    }

    function onStreamOk() {
        retryCount = 0
        setStatus('')
    }

    function verificarElegibilidade(cameraId, r) {
        const cam = {
            id: cameraId,
            bloqueado: r.bloqueado,
            ativo: r.ativo,
            plano: r.plano
        }
        if (r.pode_stream === false || !confVisionCameraPodeStream(cam).ok) {
            parar()
            currentCameraId = cameraId
            setStatus(confVisionMotivoStream(cam) || 'Stream indisponível.')
            return false
        }
        return true
    }

    function iniciarRecheck(cameraId) {
        if (liveRecheckTimer) clearInterval(liveRecheckTimer)
        liveRecheckTimer = setInterval(function () {
            if (currentCameraId !== cameraId) return
            ConfVisionUrls.carregarRtmpPublish(cameraId)
                .done(function (r) {
                    if (currentCameraId !== cameraId) return
                    verificarElegibilidade(cameraId, r)
                })
        }, LIVE_RECHECK_MS)
    }

    function iniciarPlayerComUrl(cameraId, url, isRetry) {
        const video = videoEl
        if (!video) return

        video.muted = true
        video.playsInline = true
        video.onplaying = onStreamOk
        video.onerror = function () {
            agendarRetry(cameraId)
        }

        if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = url + (isRetry ? '?t=' + Date.now() : '')
            video.play().catch(function () {
                agendarRetry(cameraId)
            })
            return
        }

        if (typeof Hls === 'undefined' || !Hls.isSupported()) {
            setStatus('HLS não suportado neste navegador.')
            return
        }

        if (!hlsInstance) {
            hlsInstance = new Hls(HLS_LOW_LATENCY)
            hlsInstance.attachMedia(video)

            hlsInstance.on(Hls.Events.MANIFEST_PARSED, function () {
                video.play().catch(function () {
                    agendarRetry(cameraId)
                })
                iniciarSyncAoVivo(video)
            })

            hlsInstance.on(Hls.Events.ERROR, function (_, data) {
                const manifest404 = data.response && data.response.code === 404
                const semStream = data.details === 'manifestLoadError' ||
                    data.details === 'levelLoadError' ||
                    data.details === 'fragLoadError' ||
                    manifest404

                if (semStream) {
                    agendarRetry(cameraId)
                    return
                }

                if (data.fatal) {
                    if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
                        agendarRetry(cameraId)
                    } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
                        hlsInstance.recoverMediaError()
                    } else {
                        agendarRetry(cameraId)
                    }
                }
            })
        }

        hlsInstance.loadSource(url + '?t=' + Date.now())
    }

    function iniciar(cameraId, opts) {
        opts = opts || {}
        const isRetry = !!opts.isRetry
        if (!cameraId || !videoEl) return

        if (!isRetry) {
            currentCameraId = cameraId
            retryCount = 0
            setStatus('Conectando…')
            if (hlsInstance) {
                hlsInstance.destroy()
                hlsInstance = null
            }
        }

        ConfVisionUrls.carregarRtmpPublish(cameraId)
            .done(function (r) {
                if (currentCameraId !== cameraId) return
                if (!verificarElegibilidade(cameraId, r)) return
                const url = r.hls || ConfVisionUrls.hlsUrlFromPath(r.path)
                if (!url) {
                    setStatus('Sem path HLS (RTMP_PUBLISH_SECRET?)')
                    return
                }
                if (!isRetry) iniciarRecheck(cameraId)
                iniciarPlayerComUrl(cameraId, url, isRetry)
            })
            .fail(function () {
                if (currentCameraId !== cameraId) return
                agendarRetry(cameraId)
            })
    }

    function start(opts) {
        opts = opts || {}
        videoEl = resolveEl(opts.video)
        statusEl = resolveEl(opts.status)
        if (!opts.cameraId || !videoEl) return
        iniciar(String(opts.cameraId), {})
    }

    return {
        start: start,
        stop: parar,
        getCameraId: function () { return currentCameraId }
    }
})()

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
    let burstExhausted = false
    let stallTimer = null
    let videoEl = null
    let statusEl = null

    const RETRY_MS = 2500
    const MAX_BURST_RETRIES = 40
    const LIVE_RECHECK_MS = 60000
    const RECOVERY_AFTER_EXHAUST_MS = 30000
    const STALL_WATCH_MS = 12000

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
        if (stallTimer) {
            clearTimeout(stallTimer)
            stallTimer = null
        }
        burstExhausted = false
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

    function destruirHls() {
        if (hlsInstance) {
            hlsInstance.destroy()
            hlsInstance = null
        }
    }

    function agendarRecuperacaoLenta(cameraId) {
        if (currentCameraId !== cameraId) return
        burstExhausted = true
        setStatus('Stream indisponível. Nova tentativa em breve…')
        clearTimeout(retryTimer)
        retryTimer = setTimeout(function () {
            retryTimer = null
            if (currentCameraId !== cameraId) return
            retryCount = 0
            burstExhausted = false
            iniciar(cameraId, { isRetry: true, forceReset: true })
        }, RECOVERY_AFTER_EXHAUST_MS)
    }

    function agendarRetry(cameraId) {
        if (currentCameraId !== cameraId) return
        if (retryCount >= MAX_BURST_RETRIES) {
            agendarRecuperacaoLenta(cameraId)
            return
        }
        retryCount++
        setStatus('Conectando… tentativa ' + retryCount)
        clearTimeout(retryTimer)
        retryTimer = setTimeout(function () {
            if (currentCameraId === cameraId) {
                iniciar(cameraId, { isRetry: true, forceReset: true })
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
        burstExhausted = false
        setStatus('')
        if (stallTimer) {
            clearTimeout(stallTimer)
            stallTimer = null
        }
    }

    function agendarStallWatch(cameraId) {
        if (stallTimer) clearTimeout(stallTimer)
        stallTimer = setTimeout(function () {
            stallTimer = null
            if (currentCameraId !== cameraId || !videoEl) return
            if (videoEl.readyState >= 3 && !videoEl.paused && !videoEl.seeking) return
            agendarRetry(cameraId)
        }, STALL_WATCH_MS)
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
                    if (!verificarElegibilidade(cameraId, r)) return
                    const video = videoEl
                    const precisaReconectar = burstExhausted ||
                        (video && (video.error || (video.readyState < 2 && retryCount > 0)))
                    if (precisaReconectar) {
                        retryCount = 0
                        burstExhausted = false
                        iniciar(cameraId, { isRetry: true, forceReset: true })
                    }
                })
        }, LIVE_RECHECK_MS)
    }

    function iniciarPlayerComUrl(cameraId, url, isRetry, forceReset) {
        const video = videoEl
        if (!video) return

        const bust = '?t=' + Date.now()
        const srcUrl = url + bust

        video.muted = true
        video.playsInline = true
        video.onplaying = onStreamOk
        video.onwaiting = function () {
            agendarStallWatch(cameraId)
        }
        video.onerror = function () {
            agendarRetry(cameraId)
        }

        if (video.canPlayType('application/vnd.apple.mpegurl')) {
            destruirHls()
            video.removeAttribute('src')
            video.load()
            video.src = srcUrl
            video.play().catch(function () {
                agendarRetry(cameraId)
            })
            agendarStallWatch(cameraId)
            return
        }

        if (typeof Hls === 'undefined' || !Hls.isSupported()) {
            setStatus('HLS não suportado neste navegador.')
            return
        }

        if (forceReset || isRetry || hlsInstance) {
            destruirHls()
        }

        hlsInstance = new Hls(HLS_LOW_LATENCY)
        hlsInstance.attachMedia(video)

        hlsInstance.on(Hls.Events.MANIFEST_PARSED, function () {
            video.play().catch(function () {
                agendarRetry(cameraId)
            })
            iniciarSyncAoVivo(video)
            agendarStallWatch(cameraId)
        })

        hlsInstance.on(Hls.Events.ERROR, function (_, data) {
            const manifest404 = data.response && data.response.code === 404
            const semStream = data.details === 'manifestLoadError' ||
                data.details === 'levelLoadError' ||
                data.details === 'fragLoadError' ||
                manifest404

            if (semStream) {
                destruirHls()
                agendarRetry(cameraId)
                return
            }

            if (data.fatal) {
                if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
                    destruirHls()
                    agendarRetry(cameraId)
                } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
                    try {
                        hlsInstance.recoverMediaError()
                    } catch (e) {
                        destruirHls()
                        agendarRetry(cameraId)
                    }
                } else {
                    destruirHls()
                    agendarRetry(cameraId)
                }
            }
        })

        hlsInstance.loadSource(srcUrl)
        agendarStallWatch(cameraId)
    }

    function iniciar(cameraId, opts) {
        opts = opts || {}
        const isRetry = !!opts.isRetry
        if (!cameraId || !videoEl) return

        const forceReset = !!opts.forceReset

        if (!isRetry) {
            currentCameraId = cameraId
            retryCount = 0
            burstExhausted = false
            setStatus('Conectando…')
            destruirHls()
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
                iniciarPlayerComUrl(cameraId, url, isRetry, forceReset)
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

/**
 * Player HLS por célula do mosaico (multi-instância).
 */
const ConfVisionMosaicCell = (function () {
    const instancias = {}
    const RETRY_MS = 2500
    const MAX_RETRIES = 20

    const HLS_OPTS = {
        lowLatencyMode: true,
        enableWorker: true,
        backBufferLength: 0,
        maxBufferLength: 6,
        maxMaxBufferLength: 10,
        manifestLoadingTimeOut: 10000,
        fragLoadingTimeOut: 15000
    }

    function parar(id) {
        const inst = instancias[id]
        if (!inst) return
        if (inst.retryTimer) clearTimeout(inst.retryTimer)
        if (inst.hls) {
            inst.hls.destroy()
            inst.hls = null
        }
        if (inst.video) {
            inst.video.onplaying = null
            inst.video.onerror = null
            inst.video.removeAttribute('src')
            inst.video.load()
        }
        delete instancias[id]
    }

    function pararTodas() {
        Object.keys(instancias).forEach(parar)
    }

    function setLoading(inst, on) {
        if (!inst || !inst.frame) return
        inst.frame.classList.toggle('is-loading', !!on)
    }

    function agendarRetry(inst) {
        if (!inst || inst.retryCount >= MAX_RETRIES) {
            setLoading(inst, true)
            return
        }
        inst.retryCount++
        clearTimeout(inst.retryTimer)
        inst.retryTimer = setTimeout(function () {
            iniciar(inst.id, inst.opts, true)
        }, RETRY_MS)
    }

    function iniciarPlayer(inst, url, isRetry) {
        const video = inst.video
        if (!video) return
        video.muted = true
        video.playsInline = true
        video.onplaying = function () {
            setLoading(inst, false)
            inst.retryCount = 0
        }
        video.onerror = function () {
            agendarRetry(inst)
        }

        if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = url + (isRetry ? '?t=' + Date.now() : '')
            video.play().catch(function () { agendarRetry(inst) })
            return
        }

        if (typeof Hls === 'undefined' || !Hls.isSupported()) {
            setLoading(inst, false)
            return
        }

        if (!inst.hls) {
            inst.hls = new Hls(HLS_OPTS)
            inst.hls.attachMedia(video)
            inst.hls.on(Hls.Events.MANIFEST_PARSED, function () {
                video.play().catch(function () { agendarRetry(inst) })
            })
            inst.hls.on(Hls.Events.ERROR, function (_, data) {
                if (data.fatal) agendarRetry(inst)
            })
        }
        inst.hls.loadSource(url + '?t=' + Date.now())
    }

    function iniciar(id, opts, isRetry) {
        opts = opts || {}
        isRetry = !!isRetry
        let inst = instancias[id]
        if (!inst) {
            inst = {
                id: id,
                opts: opts,
                video: opts.video,
                frame: opts.frame,
                retryCount: 0,
                retryTimer: null,
                hls: null
            }
            instancias[id] = inst
        }
        if (!isRetry) {
            inst.retryCount = 0
            setLoading(inst, true)
            if (inst.hls) {
                inst.hls.destroy()
                inst.hls = null
            }
        }
        const cameraId = String(opts.cameraId || id)
        ConfVisionUrls.carregarRtmpPublish(cameraId)
            .done(function (r) {
                if (!instancias[id]) return
                const cam = { id: cameraId, bloqueado: r.bloqueado, ativo: r.ativo, plano: r.plano }
                if (r.pode_stream === false || !confVisionCameraPodeStream(cam).ok) {
                    setLoading(inst, false)
                    if (inst.frame) inst.frame.classList.add('is-offline')
                    return
                }
                const url = r.hls || ConfVisionUrls.hlsUrlFromPath(r.path)
                if (!url) {
                    agendarRetry(inst)
                    return
                }
                iniciarPlayer(inst, url, isRetry)
            })
            .fail(function () {
                if (instancias[id]) agendarRetry(inst)
            })
    }

    function start(opts) {
        if (!opts || !opts.cameraId || !opts.video) return
        const id = String(opts.cellId || opts.cameraId)
        iniciar(id, opts, false)
    }

    return {
        start: start,
        stop: parar,
        stopAll: pararTodas
    }
})()

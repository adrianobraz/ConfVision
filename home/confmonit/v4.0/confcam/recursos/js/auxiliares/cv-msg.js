/**
 * ConfVision — mensagens modais (SweetAlert2 com visual do tema).
 */
var CvMsg = (function () {
    function swalReady() {
        return typeof Swal !== 'undefined' && Swal && typeof Swal.fire === 'function'
    }

    function fallbackAlert(msg) {
        if (typeof window !== 'undefined' && window.alert) window.alert(msg)
    }

    function baseOpts(extra) {
        var o = {
            customClass: {
                popup: 'cv-msg-popup',
                title: 'cv-msg-title',
                htmlContainer: 'cv-msg-text',
                actions: 'cv-msg-actions',
                confirmButton: 'cv-msg-btn cv-msg-btn-primary',
                cancelButton: 'cv-msg-btn cv-msg-btn-ghost',
                denyButton: 'cv-msg-btn cv-msg-btn-ghost'
            },
            buttonsStyling: false,
            heightAuto: false
        }
        if (extra) {
            Object.keys(extra).forEach(function (k) { o[k] = extra[k] })
        }
        return o
    }

    function normalizeText(msg) {
        if (msg == null) return ''
        return String(msg).trim()
    }

    function fire(opts) {
        if (!swalReady()) {
            fallbackAlert(opts.text || opts.title || '')
            return $.Deferred().resolve({ isConfirmed: true }).promise()
        }
        return Swal.fire(baseOpts(opts))
    }

    function sucesso(mensagem, titulo) {
        return fire({
            icon: 'success',
            title: titulo || 'Sucesso',
            text: normalizeText(mensagem),
            timer: 2200,
            showConfirmButton: false
        })
    }

    function aviso(mensagem, titulo) {
        return fire({
            icon: 'warning',
            title: titulo || 'Atenção',
            text: normalizeText(mensagem),
            confirmButtonText: 'OK'
        })
    }

    function avisoCampo(mensagem, seletorCampo, titulo) {
        return fire({
            icon: 'warning',
            title: titulo || 'Atenção',
            text: normalizeText(mensagem),
            confirmButtonText: 'OK',
            timer: 3200,
            didClose: function () {
                if (seletorCampo && typeof $ !== 'undefined') $(seletorCampo).focus()
            }
        })
    }

    function erro(mensagem, titulo) {
        return fire({
            icon: 'error',
            title: titulo || 'Erro',
            text: normalizeText(mensagem),
            confirmButtonText: 'OK'
        })
    }

    function info(mensagem, titulo) {
        return fire({
            icon: 'info',
            title: titulo || 'Informação',
            text: normalizeText(mensagem),
            confirmButtonText: 'OK'
        })
    }

    function confirmar(titulo, mensagem, opcoes) {
        var extra = opcoes || {}
        return fire({
            icon: extra.icon || 'question',
            title: titulo || 'Confirmar',
            text: normalizeText(mensagem),
            showCancelButton: true,
            confirmButtonText: extra.confirmText || 'Sim',
            cancelButtonText: extra.cancelText || 'Cancelar',
            reverseButtons: true,
            focusCancel: !!extra.focusCancel
        })
    }

    function processando(mensagem) {
        return fire({
            title: normalizeText(mensagem) || 'Processando...',
            allowOutsideClick: false,
            allowEscapeKey: false,
            showConfirmButton: false,
            didOpen: function () {
                if (Swal.showLoading) Swal.showLoading()
            }
        })
    }

    function fechar() {
        if (swalReady()) Swal.close()
    }

    return {
        sucesso: sucesso,
        aviso: aviso,
        avisoCampo: avisoCampo,
        erro: erro,
        info: info,
        confirmar: confirmar,
        processando: processando,
        fechar: fechar
    }
})()

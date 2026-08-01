/**
 * Whitelabel — logo (fp_whitelabel.logo_data) e nome da marca (nomeFranqueado).
 * Tema/logo: cache local (localStorage) aplicado na hora; Xano só 1x por sessão ou no login.
 */
var CvWhitelabel = (function () {
    var STORAGE_PREFIX = 'cv_whitelabel_logo_'
    var LEGACY_STORAGE_KEY = 'cv_whitelabel_logo'
    var TENANT_KEY = 'cv_tenant_franqueado'
    var SESSION_SYNC_PREFIX = 'cv_whitelabel_synced_'
    var DEFAULT_HORIZONTAL = '/recursos/img/brand/logo-confvision-horizontal-white.png'
    var DEFAULT_VERTICAL = '/recursos/img/brand/logo-confvision-vertical-white.png'
    var DEFAULT_BRAND_LABEL = 'Central de Câmeras'

    function storageKey(idFranqueado) {
        var id = String(idFranqueado || '').trim()
        return id ? STORAGE_PREFIX + id : ''
    }

    function sessionSyncKey(idFranqueado) {
        var id = String(idFranqueado || '').trim()
        return id ? SESSION_SYNC_PREFIX + id : ''
    }

    function removeLegacyStorage() {
        try {
            localStorage.removeItem(LEGACY_STORAGE_KEY)
        } catch (e) { /* ignore */ }
    }

    function clearTenantLogo(idFranqueado) {
        var key = storageKey(idFranqueado)
        if (!key) return
        try {
            localStorage.removeItem(key)
        } catch (e) { /* ignore */ }
    }

    function clearSessionSyncFlags() {
        try {
            var keys = []
            for (var i = 0; i < sessionStorage.length; i++) {
                var k = sessionStorage.key(i)
                if (k && k.indexOf(SESSION_SYNC_PREFIX) === 0) keys.push(k)
            }
            keys.forEach(function (k) { sessionStorage.removeItem(k) })
        } catch (e) { /* ignore */ }
    }

    function resolveIdFranqueado(idFranqueado) {
        if (idFranqueado) return String(idFranqueado).trim()
        try {
            var idLs = String(localStorage.getItem('idFranqueado') || '').trim()
            if (idLs) return idLs
        } catch (e) { /* ignore */ }
        if (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.idFranqueado) {
            return String(ConfVisionUrls.idFranqueado() || '').trim()
        }
        return getTenant()
    }

    function tenantFromDom() {
        var el = document.getElementById('cv-tenant-franqueado')
        return el && el.value ? String(el.value).trim() : ''
    }

    function brandName() {
        try {
            var nome = (localStorage.getItem('nomeFranqueado') || '').trim()
            if (nome) return nome
        } catch (e) { /* ignore */ }
        return DEFAULT_BRAND_LABEL
    }

    function applyBrandToDocument() {
        var brand = brandName()

        var meta = document.querySelector('meta[name="apple-mobile-web-app-title"]')
        if (meta) meta.setAttribute('content', brand)

        document.querySelectorAll('.cv-brand-footer').forEach(function (el) {
            el.textContent = brand
        })

        document.querySelectorAll('[data-cv-brand]').forEach(function (el) {
            var tpl = el.getAttribute('data-cv-brand') || '{brand}'
            el.textContent = tpl.replace(/\{brand\}/g, brand)
        })

        if (document.title && document.title.indexOf('ConfVision') !== -1) {
            document.title = document.title.replace(/ConfVision/g, brand)
        }
    }

    function paintLogoImages(dataUrl) {
        document.querySelectorAll('.cv-brand-logo, .login-logo').forEach(function (img) {
            img.src = dataUrl
            img.alt = brandName()
        })
    }

    function applyLogo(dataUrl, idFranqueado) {
        if (!dataUrl) return
        var id = resolveIdFranqueado(idFranqueado)
        try {
            removeLegacyStorage()
            if (id) {
                localStorage.setItem(storageKey(id), dataUrl)
                setTenant(id)
            }
        } catch (e) { /* ignore quota */ }
        paintLogoImages(dataUrl)
        applyBrandToDocument()
    }

    function resetLogo(idFranqueado, clearStorage) {
        var id = resolveIdFranqueado(idFranqueado)
        removeLegacyStorage()
        if (clearStorage && id) clearTenantLogo(id)

        document.querySelectorAll('.cv-brand-logo').forEach(function (img) {
            img.src = DEFAULT_HORIZONTAL
            img.alt = brandName()
        })
        document.querySelectorAll('.login-logo').forEach(function (img) {
            img.src = DEFAULT_VERTICAL
            img.alt = brandName()
        })
        applyBrandToDocument()
    }

    function applyFromStorage(idFranqueado) {
        var id = resolveIdFranqueado(idFranqueado)
        if (!id) return false
        try {
            removeLegacyStorage()
            var logo = localStorage.getItem(storageKey(id))
            if (!logo) return false
            paintLogoImages(logo)
            applyBrandToDocument()
            return true
        } catch (e) {
            return false
        }
    }

    function applyCachedSessionBrand(idFranqueado) {
        var id = resolveIdFranqueado(idFranqueado)
        if (id) setTenant(id)

        var hasLogo = applyFromStorage(id)
        if (typeof CvTheme !== 'undefined') {
            if (!CvTheme.applyFromStorage(id)) {
                CvTheme.applyTheme(CvTheme.DEFAULT_THEME)
            }
        }
        if (!hasLogo) {
            resetLogo(null, false)
        }
        applyBrandToDocument()
        return hasLogo
    }

    function setTenant(idFranqueado) {
        if (!idFranqueado) return
        try {
            sessionStorage.setItem(TENANT_KEY, String(idFranqueado))
        } catch (e) { /* ignore */ }
    }

    function getTenant() {
        try {
            return sessionStorage.getItem(TENANT_KEY) || ''
        } catch (e) {
            return ''
        }
    }

    function markSessionSynced(idFranqueado) {
        var key = sessionSyncKey(idFranqueado)
        if (!key) return
        try {
            sessionStorage.setItem(key, String(Date.now()))
        } catch (e) { /* ignore */ }
    }

    function isSessionSynced(idFranqueado) {
        var key = sessionSyncKey(idFranqueado)
        if (!key) return false
        try {
            return !!sessionStorage.getItem(key)
        } catch (e) {
            return false
        }
    }

    function clearSessionBrand() {
        removeLegacyStorage()
        clearSessionSyncFlags()
        try {
            sessionStorage.removeItem(TENANT_KEY)
        } catch (e) { /* ignore */ }
        document.querySelectorAll('.cv-brand-logo').forEach(function (img) {
            img.src = DEFAULT_HORIZONTAL
            img.alt = DEFAULT_BRAND_LABEL
        })
        document.querySelectorAll('.login-logo').forEach(function (img) {
            img.src = DEFAULT_VERTICAL
            img.alt = DEFAULT_BRAND_LABEL
        })
        if (typeof CvTheme !== 'undefined') {
            CvTheme.resetTheme(null, true)
        }
    }

    function applyThemeFromDados(dados, idFranqueado) {
        if (typeof CvTheme === 'undefined' || !dados) return
        var id = idFranqueado || dados.id_franqueado
        var tema = CvTheme.parseTemaJson(dados.tema_json || dados.temaJson || dados.tema)
        if (tema) {
            CvTheme.saveTheme(tema, id)
            return
        }
        if (!CvTheme.applyFromStorage(id)) {
            CvTheme.resetTheme(id, false)
        }
    }

    function ingestDadosWhitelabel(dados, idFranqueado) {
        if (!dados) return
        var id = resolveIdFranqueado(idFranqueado || dados.id_franqueado)
        if (dados.id_franqueado) setTenant(dados.id_franqueado)
        if (dados.logo_data) {
            applyLogo(dados.logo_data, id)
        } else if (id) {
            resetLogo(id, true)
        }
        applyThemeFromDados(dados, id)
        if (id) markSessionSynced(id)
    }

    function syncFromServer(idFranqueado, force) {
        var idFra = resolveIdFranqueado(idFranqueado)
        if (!idFra) {
            return $.Deferred().reject().promise()
        }

        setTenant(idFra)

        if (!force && isSessionSynced(idFra)) {
            return $.Deferred().resolve().promise()
        }

        return $.ajax({
            url: '/cvWhitelabelCarregar',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ id_franqueado: idFra })
        }).done(function (r) {
            ingestDadosWhitelabel(r && r.dados, idFra)
        }).fail(function () {
            applyCachedSessionBrand(idFra)
        }).always(function () {
            applyBrandToDocument()
        })
    }

    /** Após login: busca whitelabel uma vez e grava cache antes de redirecionar. */
    function fetchAndCache(idFranqueado) {
        var idFra = resolveIdFranqueado(idFranqueado)
        if (!idFra) {
            return $.Deferred().reject().promise()
        }
        applyCachedSessionBrand(idFra)
        return syncFromServer(idFra, true)
    }

    function invalidateSessionSync(idFranqueado) {
        var key = sessionSyncKey(idFranqueado)
        if (!key) return
        try {
            sessionStorage.removeItem(key)
        } catch (e) { /* ignore */ }
    }

    /** Resolve logo/tema pelo Host (tela de login, antes do auth). */
    function resolveByHost() {
        var tenantId = tenantFromDom() || getTenant()
        if (tenantId) {
            setTenant(tenantId)
            applyCachedSessionBrand(tenantId)
        } else {
            resetLogo(null, false)
            if (typeof CvTheme !== 'undefined' && !CvTheme.applyFromStorage()) {
                CvTheme.resetTheme(null, false)
            }
        }

        var host = (window.location.hostname || '').toLowerCase()
        if (!host || host === 'localhost' || host === '127.0.0.1') {
            applyBrandToDocument()
            return $.Deferred().resolve().promise()
        }

        if (tenantId && isSessionSynced(tenantId)) {
            applyBrandToDocument()
            return $.Deferred().resolve().promise()
        }

        return $.ajax({
            url: '/cvWhitelabelByFqdn',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ fqdn: host })
        }).done(function (r) {
            ingestDadosWhitelabel(r && r.dados)
        }).always(function () {
            applyBrandToDocument()
        })
    }

    /** Páginas autenticadas: cache imediato; Xano no máximo 1x por sessão. */
    function loadForSession() {
        var idFra = resolveIdFranqueado()
        if (!idFra) {
            return resolveByHost()
        }

        applyCachedSessionBrand(idFra)

        if (isSessionSynced(idFra)) {
            applyBrandToDocument()
            return $.Deferred().resolve().promise()
        }

        return syncFromServer(idFra, false)
    }

    function notifyBrandSaved(idFranqueado) {
        markSessionSynced(resolveIdFranqueado(idFranqueado))
    }

    return {
        brandName: brandName,
        applyBrandToDocument: applyBrandToDocument,
        applyLogo: applyLogo,
        resetLogo: resetLogo,
        applyFromStorage: applyFromStorage,
        applyCachedSessionBrand: applyCachedSessionBrand,
        clearTenantLogo: clearTenantLogo,
        clearSessionBrand: clearSessionBrand,
        resolveByHost: resolveByHost,
        loadForSession: loadForSession,
        fetchAndCache: fetchAndCache,
        invalidateSessionSync: invalidateSessionSync,
        notifyBrandSaved: notifyBrandSaved,
        setTenant: setTenant,
        getTenant: getTenant,
        DEFAULT_BRAND_LABEL: DEFAULT_BRAND_LABEL
    }
})()

$(function () {
    $(document).on('click', 'a[href="/logout"], a[href="/logout/"]', function () {
        if (typeof CvWhitelabel !== 'undefined') {
            CvWhitelabel.clearSessionBrand()
        }
    })

    if ($('body').hasClass('login-page')) {
        CvWhitelabel.resolveByHost()
    } else if ($('.cv-brand-logo').length) {
        CvWhitelabel.loadForSession()
    } else {
        var idFra = ''
        try {
            idFra = localStorage.getItem('idFranqueado') || ''
        } catch (e) { /* ignore */ }
        if (idFra && typeof CvWhitelabel !== 'undefined') {
            CvWhitelabel.applyCachedSessionBrand(idFra)
        } else if (typeof CvWhitelabel !== 'undefined') {
            CvWhitelabel.applyBrandToDocument()
        }
    }
})

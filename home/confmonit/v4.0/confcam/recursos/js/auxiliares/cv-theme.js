/**
 * ConfVision — tema whitelabel (fp_whitelabel.tema_json compartilhado com FranqueadoPro).
 * Aplica presets/cores como variáveis CSS --cv-*.
 */
var CvTheme = (function () {
    var STORAGE_PREFIX = 'cv_theme_'

    var DEFAULT_THEME = {
        preset: 'escuro_ciano',
        primary: '#38bdf8',
        primaryHover: '#0ea5e9',
        navbar: '#0d1424',
        bg: '#070b14',
        surface: '#0d1424',
        text: '#e2e8f0'
    }

    var PRESETS = {
        escuro_ciano: {
            label: 'Azul claro',
            preset: 'escuro_ciano',
            primary: '#38bdf8',
            primaryHover: '#0ea5e9',
            navbar: '#0d1424',
            bg: '#070b14',
            surface: '#0d1424',
            text: '#e2e8f0'
        },
        escuro_azul: {
            label: 'Azul escuro',
            preset: 'escuro_azul',
            primary: '#2563eb',
            primaryHover: '#1d4ed8',
            navbar: '#0a1020',
            bg: '#04060d',
            surface: '#0c1222',
            text: '#e2e8f0'
        },
        escuro_verde: {
            label: 'Verde musgo',
            preset: 'escuro_verde',
            primary: '#7d9b76',
            primaryHover: '#6b8574',
            navbar: '#101612',
            bg: '#0a0f0b',
            surface: '#121a13',
            text: '#e8ede6'
        },
        escuro_verde_claro: {
            label: 'Verde claro',
            preset: 'escuro_verde_claro',
            primary: '#4ade80',
            primaryHover: '#22c55e',
            navbar: '#0c1610',
            bg: '#061008',
            surface: '#0f1a12',
            text: '#e8f5ec'
        },
        escuro_roxo: {
            label: 'Roxo',
            preset: 'escuro_roxo',
            primary: '#a855f7',
            primaryHover: '#9333ea',
            navbar: '#150f1f',
            bg: '#0a0812',
            surface: '#181320',
            text: '#ece7f3'
        },
        escuro_ambar: {
            label: 'Laranja',
            preset: 'escuro_ambar',
            primary: '#f97316',
            primaryHover: '#ea580c',
            navbar: '#1a1208',
            bg: '#100c08',
            surface: '#1a1410',
            text: '#f5ebe0'
        },
        escuro_vermelho: {
            label: 'Vermelho',
            preset: 'escuro_vermelho',
            primary: '#dc2626',
            primaryHover: '#b91c1c',
            navbar: '#180808',
            bg: '#0c0404',
            surface: '#180c0c',
            text: '#f5e8e8'
        },
        escuro_amarelo: {
            label: 'Amarelo',
            preset: 'escuro_amarelo',
            primary: '#facc15',
            primaryHover: '#eab308',
            navbar: '#1a1608',
            bg: '#100e06',
            surface: '#1a1810',
            text: '#f5f0dc'
        },
        escuro_amarelo_vivo: {
            label: 'Amarelo vivo',
            preset: 'escuro_amarelo_vivo',
            primary: '#fde047',
            primaryHover: '#facc15',
            navbar: '#1a1806',
            bg: '#121006',
            surface: '#1c1a0a',
            text: '#faf6dc'
        },
        escuro_verde_amarelado: {
            label: 'Verde amarelado',
            preset: 'escuro_verde_amarelado',
            primary: '#a3e635',
            primaryHover: '#84cc16',
            navbar: '#141808',
            bg: '#0c1006',
            surface: '#161a10',
            text: '#eef2e4'
        },
        escuro_marrom: {
            label: 'Marrom',
            preset: 'escuro_marrom',
            primary: '#a8714a',
            primaryHover: '#8b5e34',
            navbar: '#181008',
            bg: '#100a06',
            surface: '#1c140e',
            text: '#f0e4d6'
        },
        escuro_cinza: {
            label: 'Cinza',
            preset: 'escuro_cinza',
            primary: '#a3a3a3',
            primaryHover: '#737373',
            navbar: '#121212',
            bg: '#0a0a0a',
            surface: '#161616',
            text: '#e5e5e5'
        }
    }

    var PRESET_ALIASES = {
        escuro_teal: 'escuro_ciano',
        escuro_magenta: 'escuro_vermelho'
    }

    function storageKey(idFranqueado) {
        var id = String(idFranqueado || '').trim()
        return id ? STORAGE_PREFIX + id : ''
    }

    function resolveIdFranqueado(idFranqueado) {
        if (idFranqueado) return String(idFranqueado).trim()
        if (typeof CvWhitelabel !== 'undefined' && CvWhitelabel.getTenant) {
            var t = CvWhitelabel.getTenant()
            if (t) return t
        }
        if (typeof ConfVisionUrls !== 'undefined' && ConfVisionUrls.idFranqueado) {
            return String(ConfVisionUrls.idFranqueado() || '').trim()
        }
        try {
            return String(localStorage.getItem('idFranqueado') || '').trim()
        } catch (e) {
            return ''
        }
    }

    function hexToRgb(hex) {
        var m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex || '')
        if (!m) return null
        return parseInt(m[1], 16) + ', ' + parseInt(m[2], 16) + ', ' + parseInt(m[3], 16)
    }

    function lighten(hex, percent) {
        var m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex || '')
        if (!m) return hex
        var amt = Math.round(2.55 * percent)
        var clamp = function (v) { return Math.max(0, Math.min(255, v)) }
        var r = clamp(parseInt(m[1], 16) + amt)
        var g = clamp(parseInt(m[2], 16) + amt)
        var b = clamp(parseInt(m[3], 16) + amt)
        return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)
    }

    function hexToRgba(hex, alpha) {
        var rgb = hexToRgb(hex)
        if (!rgb) return 'rgba(15, 23, 42, ' + alpha + ')'
        return 'rgba(' + rgb + ', ' + alpha + ')'
    }

    function normalizarHex(cor) {
        if (cor == null || cor === '') return ''
        var s = String(cor).trim()
        if (/^#[0-9a-fA-F]{6}$/.test(s)) return s.toLowerCase()
        if (/^[0-9a-fA-F]{6}$/.test(s)) return ('#' + s).toLowerCase()
        return ''
    }

    function normalizeTheme(raw) {
        var tema = parseTemaJson(raw)
        if (!tema) return cloneTheme(DEFAULT_THEME)
        return tema
    }

    function cloneTheme(t) {
        return {
            preset: t.preset || 'escuro_ciano',
            primary: t.primary,
            primaryHover: t.primaryHover,
            navbar: t.navbar,
            bg: t.bg,
            surface: t.surface,
            text: t.text
        }
    }

    function parseTemaJson(raw) {
        if (!raw) return null
        var tema = raw
        if (typeof tema === 'string') {
            try { tema = JSON.parse(tema) } catch (e) { return null }
        }
        if (typeof tema !== 'object' || Array.isArray(tema)) return null
        if (!tema.primary && tema.value && typeof tema.value === 'object') tema = tema.value

        var presetKey = tema.preset || ''
        if (PRESET_ALIASES[presetKey]) {
            return cloneTheme(PRESETS[PRESET_ALIASES[presetKey]])
        }
        if (presetKey && PRESETS[presetKey] && !tema.primary) {
            return cloneTheme(PRESETS[presetKey])
        }

        var primary = normalizarHex(tema.primary || tema.Primary || tema.cor_primaria)
        if (!primary && presetKey && PRESETS[presetKey]) {
            return cloneTheme(PRESETS[presetKey])
        }
        if (!primary) return null

        return {
            preset: presetKey || 'personalizado',
            primary: primary,
            primaryHover: normalizarHex(tema.primaryHover || tema.primary_hover) || lighten(primary, -8),
            navbar: normalizarHex(tema.navbar || tema.Navbar) || normalizarHex(tema.surface) || DEFAULT_THEME.navbar,
            bg: normalizarHex(tema.bg || tema.background) || DEFAULT_THEME.bg,
            surface: normalizarHex(tema.surface) || normalizarHex(tema.navbar) || DEFAULT_THEME.surface,
            text: normalizarHex(tema.text || tema.texto) || DEFAULT_THEME.text
        }
    }

    function buildThemeVars(t) {
        var primary = t.primary
        var primaryHover = t.primaryHover || lighten(primary, -8)
        var bg = t.bg || DEFAULT_THEME.bg
        var surface = t.surface || t.navbar || DEFAULT_THEME.surface
        var text = t.text || DEFAULT_THEME.text
        var rgb = hexToRgb(primary)
        var vars = {
            '--cv-bg': bg,
            '--cv-bg-elevated': surface,
            '--cv-bg-card': hexToRgba(surface, 0.72),
            '--cv-bg-input': hexToRgba(surface, 0.85),
            '--cv-accent': primary,
            '--cv-accent-dim': primaryHover,
            '--cv-text': text,
            '--cv-text-muted': lighten(text, -35),
            '--cv-bg-glow-2': hexToRgba(primary, 0.06),
            '--login-bg': bg,
            '--login-card': hexToRgba(surface, 0.82),
            '--login-text': text,
            '--login-muted': lighten(text, -35),
            '--login-accent': primary,
            '--login-accent-hover': primaryHover,
            '--login-input-bg': hexToRgba(bg, 0.65),
            '--login-bg-glow-1': hexToRgba(primary, 0.18),
            '--login-bg-glow-2': hexToRgba(primary, 0.08)
        }
        if (rgb) {
            vars['--cv-border'] = 'rgba(' + rgb + ', 0.18)'
            vars['--cv-border-hover'] = 'rgba(' + rgb + ', 0.45)'
            vars['--cv-accent-glow'] = 'rgba(' + rgb + ', 0.25)'
            vars['--cv-bg-glow-1'] = 'rgba(' + rgb + ', 0.12)'
            vars['--cv-accent-soft'] = 'rgba(' + rgb + ', 0.10)'
            vars['--cv-accent-medium'] = 'rgba(' + rgb + ', 0.12)'
            vars['--cv-accent-strong'] = 'rgba(' + rgb + ', 0.18)'
            vars['--cv-accent-faint'] = 'rgba(' + rgb + ', 0.06)'
            vars['--cv-accent-focus'] = 'rgba(' + rgb + ', 0.15)'
            vars['--cv-accent-border-soft'] = 'rgba(' + rgb + ', 0.35)'
            vars['--cv-accent-border'] = 'rgba(' + rgb + ', 0.40)'
            vars['--cv-table-hover-bg'] = 'rgba(' + rgb + ', 0.08)'
            vars['--login-border'] = 'rgba(' + rgb + ', 0.22)'
            vars['--login-input-hover-border'] = 'rgba(' + rgb + ', 0.28)'
            vars['--login-btn-shadow'] = 'rgba(' + rgb + ', 0.28)'
            vars['--login-btn-shadow-hover'] = 'rgba(' + rgb + ', 0.38)'
        }
        return vars
    }

    function applyVarsToTarget(target, vars) {
        if (!target) return
        Object.keys(vars).forEach(function (key) {
            target.setProperty(key, vars[key])
        })
    }

    function applyVarsEverywhere(vars) {
        applyVarsToTarget(document.documentElement.style, vars)
        if (document.body) applyVarsToTarget(document.body.style, vars)
        document.querySelectorAll('.cv-cadastro-root').forEach(function (el) {
            applyVarsToTarget(el.style, vars)
        })
    }

    function applyTheme(theme) {
        var t = normalizeTheme(theme)
        var vars = buildThemeVars(t)
        applyVarsEverywhere(vars)

        var meta = document.querySelector('meta[name="theme-color"]')
        if (meta) meta.setAttribute('content', vars['--cv-bg'])

        return t
    }

    function saveTheme(theme, idFranqueado) {
        var t = applyTheme(theme)
        var id = resolveIdFranqueado(idFranqueado)
        if (!id) return t
        try {
            localStorage.setItem(storageKey(id), JSON.stringify(t))
        } catch (e) { /* ignore quota */ }
        return t
    }

    function getStoredTheme(idFranqueado) {
        var id = resolveIdFranqueado(idFranqueado)
        if (!id) return null
        try {
            var raw = localStorage.getItem(storageKey(id))
            if (!raw) return null
            return parseTemaJson(JSON.parse(raw))
        } catch (e) {
            return null
        }
    }

    function applyFromStorage(idFranqueado) {
        var tema = getStoredTheme(idFranqueado)
        if (!tema) return false
        applyTheme(tema)
        return true
    }

    function clearStoredTheme(idFranqueado) {
        var id = resolveIdFranqueado(idFranqueado)
        if (!id) return
        try {
            localStorage.removeItem(storageKey(id))
        } catch (e) { /* ignore */ }
    }

    function resetTheme(idFranqueado, clearStorage) {
        if (clearStorage) clearStoredTheme(idFranqueado)
        return applyTheme(DEFAULT_THEME)
    }

    function getPreset(key) {
        var k = PRESET_ALIASES[key] || key
        return PRESETS[k] ? cloneTheme(PRESETS[k]) : null
    }

    function temaParaSalvar(theme) {
        return cloneTheme(normalizeTheme(theme))
    }

    return {
        DEFAULT_THEME: DEFAULT_THEME,
        PRESETS: PRESETS,
        parseTemaJson: parseTemaJson,
        normalizeTheme: normalizeTheme,
        applyTheme: applyTheme,
        saveTheme: saveTheme,
        resetTheme: resetTheme,
        applyFromStorage: applyFromStorage,
        clearStoredTheme: clearStoredTheme,
        getStoredTheme: getStoredTheme,
        getPreset: getPreset,
        temaParaSalvar: temaParaSalvar
    }
})()

;(function bootstrapCvTheme() {
    if (typeof CvTheme === 'undefined') return
    function run() {
        if (CvTheme.applyFromStorage()) return
        CvTheme.applyTheme(CvTheme.DEFAULT_THEME)
    }
    run()
    document.addEventListener('DOMContentLoaded', run)
})()

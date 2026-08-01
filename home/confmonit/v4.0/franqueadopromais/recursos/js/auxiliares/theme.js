/* ============================================================
   FranqueadoPro - Controle de tema (sem banco)
   Persiste a escolha do usuario em localStorage (chave fp_theme).
   As cores sao aplicadas como variaveis CSS em :root.
   ============================================================ */

var FP_THEME_KEY = 'fp_theme';

var FP_PRESETS = {
    escuro_azul: {
        preset: 'escuro_azul',
        primary: '#3b82f6',
        primaryHover: '#2563eb',
        navbar: '#0f1117',
        bg: '#161a21',
        surface: '#222630',
        text: '#e8eaed'
    },
    escuro_verde: {
        preset: 'escuro_verde',
        primary: '#22c55e',
        primaryHover: '#16a34a',
        navbar: '#0b1a11',
        bg: '#111814',
        surface: '#1a241d',
        text: '#e6efe9'
    },
    escuro_roxo: {
        preset: 'escuro_roxo',
        primary: '#a855f7',
        primaryHover: '#9333ea',
        navbar: '#150f1f',
        bg: '#181320',
        surface: '#221b2e',
        text: '#ece7f3'
    },
    escuro_ambar: {
        preset: 'escuro_ambar',
        primary: '#f59e0b',
        primaryHover: '#d97706',
        navbar: '#1a1409',
        bg: '#191510',
        surface: '#241f16',
        text: '#f0e8dc'
    },
    escuro_teal: {
        preset: 'escuro_teal',
        primary: '#14b8a6',
        primaryHover: '#0d9488',
        navbar: '#0a1817',
        bg: '#10191a',
        surface: '#182626',
        text: '#e2eeed'
    },
    escuro_cinza: {
        preset: 'escuro_cinza',
        primary: '#94a3b8',
        primaryHover: '#64748b',
        navbar: '#101010',
        bg: '#1a1a1a',
        surface: '#252525',
        text: '#e8e8e8'
    }
};

var FP_DEFAULT_THEME = FP_PRESETS.escuro_azul;

/* Logos padrao (whitelabel entra na Fase 2 via banco) */
var FP_LOGO = {
    darkBg: '/recursos/img/logo/franqueado_black.png',  /* logo claro, para fundo escuro */
    lightBg: '/recursos/img/logo/franqueado_white.png'  /* logo escuro, para fundo claro */
};

function fpGetTheme() {
    try {
        var saved = JSON.parse(localStorage.getItem(FP_THEME_KEY));
        if (saved && saved.primary) {
            return saved;
        }
    } catch (e) { }
    return FP_DEFAULT_THEME;
}

/* Converte um hex (#rrggbb) para "r, g, b" (para uso em rgba(...)) */
function fpHexToRgb(hex) {
    var m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex || '');
    if (!m) { return null; }
    return parseInt(m[1], 16) + ', ' + parseInt(m[2], 16) + ', ' + parseInt(m[3], 16);
}

function fpApplyTheme(theme) {
    var r = document.documentElement.style;
    if (theme.primary) {
        r.setProperty('--fp-primary', theme.primary);
        var rgb = fpHexToRgb(theme.primary);
        if (rgb) { r.setProperty('--fp-primary-rgb', rgb); }
    }
    if (theme.primaryHover) { r.setProperty('--fp-primary-hover', theme.primaryHover); }
    if (theme.navbar) { r.setProperty('--fp-navbar', theme.navbar); }
    if (theme.bg) { r.setProperty('--fp-bg', theme.bg); }
    if (theme.surface) {
        r.setProperty('--fp-surface', theme.surface);
        r.setProperty('--fp-surface-2', fpLighten(theme.surface, 8));
    }
    if (theme.text) { r.setProperty('--fp-text', theme.text); }
    fpUpdateLogos();
}

/* Retorna true se a cor for escura (para escolher o logo adequado) */
function fpIsDarkColor(hex) {
    var m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex || '');
    if (!m) { return true; }
    var r = parseInt(m[1], 16);
    var g = parseInt(m[2], 16);
    var b = parseInt(m[3], 16);
    var lum = (0.299 * r + 0.587 * g + 0.114 * b);
    return lum < 140;
}

/* Le a cor de fundo relevante (navbar/surface/bg) do tema atual */
function fpBgColorFor(target) {
    var name = '--fp-surface';
    if (target === 'navbar') { name = '--fp-navbar'; }
    else if (target === 'bg') { name = '--fp-bg'; }
    var val = getComputedStyle(document.documentElement).getPropertyValue(name);
    return (val || '').trim();
}

/* Escolhe o arquivo de logo conforme o fundo onde ele aparece */
function fpLogoFor(target) {
    return fpIsDarkColor(fpBgColorFor(target)) ? FP_LOGO.darkBg : FP_LOGO.lightBg;
}

/* Atualiza todas as imagens .fp-logo-auto conforme o tema atual */
function fpApplyWhitelabelLogo(dataUrl) {
    if (!dataUrl) return
    FP_LOGO.darkBg = dataUrl
    FP_LOGO.lightBg = dataUrl
    fpUpdateLogos()
}

function fpLoadWhitelabelLogoFromStorage() {
    try {
        var logo = localStorage.getItem('fp_whitelabel_logo')
        if (logo) fpApplyWhitelabelLogo(logo)
    } catch (e) { /* ignore */ }
}

function fpParseTemaJson(raw) {
    if (!raw) return null
    var tema = raw
    if (typeof tema === 'string') {
        try { tema = JSON.parse(tema) } catch (e) { return null }
    }
    if (typeof tema !== 'object' || Array.isArray(tema)) return null
    if (!tema.primary && tema.value && typeof tema.value === 'object') tema = tema.value
    if (!tema.primary) return null
    var hex = function (c) {
        if (!c) return ''
        var s = String(c).trim()
        if (/^#[0-9a-fA-F]{6}$/.test(s)) return s.toLowerCase()
        if (/^[0-9a-fA-F]{6}$/.test(s)) return ('#' + s).toLowerCase()
        return ''
    }
    var primary = hex(tema.primary)
    if (!primary) return null
    return {
        preset: tema.preset || 'personalizado',
        primary: primary,
        primaryHover: hex(tema.primaryHover || tema.primary_hover) || fpLighten(primary, -8),
        navbar: hex(tema.navbar) || primary,
        bg: hex(tema.bg) || FP_DEFAULT_THEME.bg,
        surface: hex(tema.surface) || FP_DEFAULT_THEME.surface,
        text: hex(tema.text) || FP_DEFAULT_THEME.text
    }
}

/** Sincroniza tema/logo salvos no banco (White Label) para localStorage e CSS. */
function fpSyncWhitelabelDoServidor() {
    if (typeof $ === 'undefined' || !localStorage.getItem('idFranqueado')) return
    $.ajax({
        url: '/whitelabelCarregar',
        method: 'POST',
        contentType: 'application/json',
        dataType: 'json',
        data: '{}'
    }).done(function (r) {
        var dados = (r && r.dados) ? r.dados : null
        if (!dados) return
        var tema = fpParseTemaJson(dados.tema_json || dados.temaJson)
        if (tema) fpSaveTheme(tema)
        if (dados.logo_data) {
            localStorage.setItem('fp_whitelabel_logo', dados.logo_data)
            fpApplyWhitelabelLogo(dados.logo_data)
        }
        fpSyncPanel()
    })
}

function fpUpdateLogos() {
    if (typeof document === 'undefined' || !document.querySelectorAll) { return; }
    var imgs = document.querySelectorAll('.fp-logo-auto');
    for (var i = 0; i < imgs.length; i++) {
        var target = imgs[i].getAttribute('data-logo-on') || 'navbar';
        imgs[i].src = fpLogoFor(target);
    }
}

/* Clareia/escurece um hex por uma porcentagem (para derivar tons) */
function fpLighten(hex, percent) {
    var m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex || '');
    if (!m) { return hex; }
    var amt = Math.round(2.55 * percent);
    var clamp = function (v) { return Math.max(0, Math.min(255, v)); };
    var r = clamp(parseInt(m[1], 16) + amt);
    var g = clamp(parseInt(m[2], 16) + amt);
    var b = clamp(parseInt(m[3], 16) + amt);
    return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
}

function fpSaveTheme(theme) {
    fpApplyTheme(theme);
    try {
        localStorage.setItem(FP_THEME_KEY, JSON.stringify(theme));
    } catch (e) { }
}

/* ---- Painel de selecao ------------------------------------ */

function fpBuildPresets() {
    var wrap = document.getElementById('fp-theme-presets');
    if (!wrap) { return; }
    wrap.innerHTML = '';
    Object.keys(FP_PRESETS).forEach(function (key) {
        var p = FP_PRESETS[key];
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'fp-preset';
        btn.title = key.replace('_', ' ');
        btn.setAttribute('data-preset', key);
        btn.style.background = 'linear-gradient(135deg, ' + p.primary + ' 50%, ' + p.navbar + ' 50%)';
        btn.addEventListener('click', function () {
            fpSaveTheme(p);
            fpSyncPanel();
        });
        wrap.appendChild(btn);
    });
}

function fpSyncPanel() {
    var theme = fpGetTheme();

    var map = {
        'fp-color-primary': theme.primary || FP_DEFAULT_THEME.primary,
        'fp-color-navbar': theme.navbar || FP_DEFAULT_THEME.navbar,
        'fp-color-bg': theme.bg || FP_DEFAULT_THEME.bg,
        'fp-color-surface': theme.surface || FP_DEFAULT_THEME.surface,
        'fp-color-text': theme.text || FP_DEFAULT_THEME.text
    };
    Object.keys(map).forEach(function (id) {
        var el = document.getElementById(id);
        if (el) { el.value = map[id]; }
    });

    var presets = document.querySelectorAll('.fp-preset');
    for (var i = 0; i < presets.length; i++) {
        var key = presets[i].getAttribute('data-preset');
        if (theme.preset && theme.preset === key) {
            presets[i].classList.add('fp-active');
        } else {
            presets[i].classList.remove('fp-active');
        }
    }
}

function fpReadCustomFromInputs() {
    var base = fpGetTheme();
    var val = function (id, fallback) {
        var el = document.getElementById(id);
        return el ? el.value : fallback;
    };
    var primary = val('fp-color-primary', base.primary);

    return {
        preset: 'personalizado',
        primary: primary,
        primaryHover: fpLighten(primary, -8),
        navbar: val('fp-color-navbar', base.navbar),
        bg: val('fp-color-bg', base.bg),
        surface: val('fp-color-surface', base.surface),
        text: val('fp-color-text', base.text)
    };
}

function fpInitThemePanel() {
    var toggle = document.getElementById('fp-theme-toggle');
    var panel = document.getElementById('fp-theme-panel');
    if (!toggle || !panel) { return; }

    fpBuildPresets();
    fpSyncPanel();

    toggle.addEventListener('click', function (e) {
        e.preventDefault();
        panel.classList.toggle('fp-open');
        fpSyncPanel();
    });

    document.addEventListener('click', function (e) {
        if (!panel.contains(e.target) && !toggle.contains(e.target)) {
            panel.classList.remove('fp-open');
        }
    });

    ['fp-color-primary', 'fp-color-navbar', 'fp-color-bg', 'fp-color-surface', 'fp-color-text'].forEach(function (id) {
        var el = document.getElementById(id);
        if (el) {
            el.addEventListener('input', function () {
                fpSaveTheme(fpReadCustomFromInputs());
                fpSyncPanel();
            });
        }
    });

    var reset = document.getElementById('fp-theme-reset');
    if (reset) {
        reset.addEventListener('click', function () {
            fpSaveTheme(FP_DEFAULT_THEME);
            fpSyncPanel();
        });
    }
}

/* Aplica o tema salvo o quanto antes e liga o painel quando o DOM estiver pronto */
fpApplyTheme(fpGetTheme());

function fpOnReady() {
    fpLoadWhitelabelLogoFromStorage();
    fpUpdateLogos();
    fpInitThemePanel();
    fpSyncWhitelabelDoServidor();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', fpOnReady);
} else {
    fpOnReady();
}

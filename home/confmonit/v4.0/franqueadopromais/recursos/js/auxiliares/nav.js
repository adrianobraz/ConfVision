/* ============================================================
   FranqueadoPro - Navegacao (combo contextual + titulo/subtitulo)
   Define a arvore de navegacao do sistema e monta:
   - um combo (dropdown) que muda conforme a pagina atual;
   - a barra de atalhos (menu-atalhos) no topo das paginas internas;
   - o titulo + subtitulo central da navbar (nome da opcao + o que faz).
   ============================================================ */

var FP_NAV = {
    // Menus principais (topo)
    top: [
        { label: "Parâmetros Técnicos", desc: "Parâmetros e cadastros de configuração do sistema", sub: "Parâmetros do sistema", icon: "bi-gear-fill", url: "/carregar-menu-configuracoes", group: "configuracao" },
        { label: "Atendimento", desc: "Operação diária: eventos, alarmes e ordens de serviço", sub: "Suporte e chamados", icon: "bi-headset", url: "/carregar-menu-atendimento", group: "atendimento" },
        { label: "Minha Empresa de Monitoramento", desc: "Cadastros e gestão da sua central de monitoramento", sub: "Cadastros e gerenciamento", icon: "bi-briefcase-fill", url: "/carregar-menu-minha-empresa", group: "minha-empresa" },
        { label: "Relatórios", desc: "Relatórios de atendimentos, eventos e clientes", sub: "Consultas e indicadores", icon: "bi-journal-text", url: "/carregar-menu-relatorio", group: "relatorio" }
    ],

    // Opcoes internas de cada grupo
    groups: {
        "minha-empresa": {
            label: "Minha Empresa de Monitoramento",
            menuUrl: "/carregar-menu-minha-empresa",
            items: [
                { label: "Operacional / Campo", desc: "Técnicos e viaturas", icon: "bi-tools", url: "/carregar-menu-operacional" },
                { label: "Comercial", desc: "Cadastro de clientes", icon: "bi-cart-check-fill", url: "/carregar-menu-comercial" },
                { label: "Gestão Interna", desc: "Dados do monitoramento, plano e permissões", icon: "bi-diagram-3-fill", url: "/carregar-menu-gestao" }
            ]
        },
        "operacional": {
            label: "Operacional / Campo",
            menuUrl: "/carregar-menu-operacional",
            atalho: "minha-empresa",
            items: [
                { label: "Cadastrar Técnico", desc: "Cadastro e gestão dos técnicos", icon: "bi-person-gear", url: "/CarregarPaginaGerenciarTecnico" },
                { label: "Cadastro de Viatura", desc: "Cadastro e gestão das viaturas", icon: "bi-truck", url: "/CarregarPaginaGerenciarViatura" }
            ]
        },
        "comercial": {
            label: "Comercial",
            menuUrl: "/carregar-menu-comercial",
            atalho: "minha-empresa",
            items: [
                { label: "Cadastro Cliente", desc: "Cadastro e gestão dos clientes", icon: "bi-person-plus-fill", url: "/CarregarPaginaGerenciarCliente" }
            ]
        },
        "gestao": {
            label: "Gestão Interna",
            menuUrl: "/carregar-menu-gestao",
            atalho: "minha-empresa",
            items: [
                { label: "Dados do Monitoramento", desc: "Informações e parâmetros da sua empresa", icon: "bi-clipboard-data", url: "/CarregarPaginaDadosMonitoramento" },
                { label: "Meu Plano", desc: "Plano, licença e faturas", icon: "bi-credit-card-fill", url: "/carregar-meu-plano" },
                { label: "Montar Meu Plano", desc: "Contratação à la carte", icon: "bi-sliders", url: "/carregar-montar-plano" },
                { label: "Faturas e Cobrança", desc: "Histórico e faturas em aberto", icon: "bi-receipt", url: "/carregar-meu-plano?tab=faturas" },
                { label: "White Label", desc: "Cores e logotipo da sua marca", icon: "bi-palette-fill", url: "/carregar-whitelabel" },
                { label: "Domínio Personalizado", desc: "Endereço do sistema na nuvem (Pro+)", icon: "bi-globe", url: "/carregar-dominio-saas" },
                { label: "Permissões de Acesso", desc: "Controle de acesso por menu", icon: "bi-shield-lock-fill", url: "/carregar-permissoes-acesso" }
            ]
        },
        "atendimento": {
            label: "Atendimento",
            menuUrl: "/carregar-menu-atendimento",
            items: [
                { label: "Gerar Evento", desc: "Gera eventos manuais para os clientes", icon: "bi-broadcast-pin", url: "/CarregarGeradorEventos" },
                { label: "Inteligência Artificial", desc: "Telefones de alerta e bloqueio de clientes da IA", icon: "bi-robot", url: "/carregar-inteligencia-artificial" },
                { label: "Gerenciar Responsáveis", desc: "Cadastro e permissões dos responsáveis", icon: "bi-people-fill", url: "/carregarPaginaGerenciarUsuarios" },
                { label: "Parceiro Monitoramento", desc: "Terceirizar cliente para parceiro externo", icon: "bi-building-check", url: "/carregar-parceiro-monitoramento" }
            ]
        },
        "relatorio": {
            label: "Relatórios",
            menuUrl: "/carregar-menu-relatorio",
            items: [
                { label: "Relatório de Atendimento", desc: "Relatório dos atendimentos realizados", icon: "bi-clipboard-check-fill", url: "/carregar-relatorio-atendimento" },
                { label: "Lista de Ligações", desc: "Histórico de ligações", icon: "bi-telephone-fill", url: "/carregar-relatorio-ligacoes" },
                { label: "Eventos", desc: "Relatórios de eventos", icon: "bi-card-list", url: "/carregar-menu-eventos" },
                { label: "Alarmes", desc: "Relatórios de alarmes armados e desarmados", icon: "bi-shield-fill", url: "/carregar-menu-alarmes" },
                { label: "Clientes", desc: "Relatórios de clientes e dispositivos", icon: "bi-people-fill", url: "/carregar-menu-clientes" },
                { label: "Central de Disparos", desc: "WhatsApp, SMS, ligações e disparos automáticos", icon: "bi-send-fill", url: "/carregar-menu-central-disparos" }
            ]
        },
        "central-disparos": {
            label: "Central de Disparos",
            menuUrl: "/carregar-menu-central-disparos",
            atalho: "relatorio",
            items: [
                { label: "Centro de Operações", desc: "Do evento ao resultado — tudo em um só comando", icon: "bi-radar", url: "/carregar-relatorio-cd-unificado" },
                { label: "Eventos", desc: "Relatório de eventos", icon: "bi-card-list", url: "/carregar-relatorio-eventos?from=cd" },
                { label: "Eventos Pendentes", desc: "Eventos aguardando disparo", icon: "bi-hourglass-split", url: "/carregar-relatorio-eventos-pendentes" },
                { label: "Log de Atendimentos Finalizados", desc: "Atendimentos finalizados", icon: "bi-clipboard-check-fill", url: "/carregar-relatorio-atendimento?from=cd" },
                { label: "Eventos Finalizados por Robô", desc: "Finalizados via robô de atendimento", icon: "bi-robot", url: "/carregar-relatorio-finalizados-robo" },
                { label: "Finalizados Automaticamente pelo Bot", desc: "Finalização automática do bot", icon: "bi-cpu-fill", url: "/carregar-relatorio-finalizados-bot" },
                { label: "Ligação Histórico", desc: "Histórico com transcrição e download de áudio", icon: "bi-telephone-inbound-fill", url: "/carregar-relatorio-ligacao-historico" },
                { label: "SMS Histórico", desc: "Histórico de SMS enviados", icon: "bi-chat-left-text-fill", url: "/carregar-relatorio-sms-historico" },
                { label: "Eventos com Falhas", desc: "Lista de eventos com falhas", icon: "bi-exclamation-triangle-fill", url: "/carregar-relatorio-eventos-falhas" },
                { label: "WhatsApp Enviados", desc: "Mensagens WhatsApp enviadas", icon: "bi-whatsapp", url: "/carregar-relatorio-whatsapp-enviados" },
                { label: "Ligações", desc: "Ligações disparadas (WhatsAppEnviados)", icon: "bi-telephone-fill", url: "/carregar-relatorio-ligacoes-cd" },
                { label: "SMS", desc: "SMS enviados", icon: "bi-phone-fill", url: "/carregar-relatorio-sms" },
                { label: "Eventos para Enviar ou Ligação", desc: "Fila de envio e ligações pendentes", icon: "bi-send-fill", url: "/carregar-relatorio-fila-envio" },
                { label: "Ligações com Erros e Tentativas", desc: "Erros e tentativas de ligação", icon: "bi-telephone-x-fill", url: "/carregar-relatorio-ligacao-erros" },
                { label: "Filas de Ligações", desc: "Fila operacional de ligações", icon: "bi-list-ol", url: "/carregar-relatorio-fila-ligacao" }
            ]
        },
        "eventos": {
            label: "Eventos",
            menuUrl: "/carregar-menu-eventos",
            atalho: "relatorio",
            items: [
                { label: "Relatórios de Eventos", desc: "Relatório de eventos recebidos", icon: "bi-card-list", url: "/carregar-relatorio-eventos" },
                { label: "Relatório de Eventos por Grupo", desc: "Relatório de eventos agrupados por grupo", icon: "bi-diagram-3-fill", url: "/carregar-configurar-relatorio-cliente" }
            ]
        },
        "alarmes": {
            label: "Alarmes",
            menuUrl: "/carregar-menu-alarmes",
            atalho: "relatorio",
            items: [
                { label: "Alarmes Armados", desc: "Lista dispositivos com alarme armado", icon: "bi-shield-lock-fill", url: "/carregar-listar-dispositivos-armados" },
                { label: "Alarmes Desarmados", desc: "Lista dispositivos com alarme desarmado", icon: "bi-shield-slash-fill", url: "/carregar-listar-dispositivos-desarmados" }
            ]
        },
        "clientes": {
            label: "Clientes",
            menuUrl: "/carregar-menu-clientes",
            atalho: "relatorio",
            items: [
                { label: "Listar Clientes e Dispositivos", desc: "Relação de clientes e seus dispositivos", icon: "bi-card-list", url: "/listar-cliente-dispositivo" },
                { label: "Lista de Clientes", desc: "Relação de clientes cadastrados", icon: "bi-people-fill", url: "/carregar-listar-clientes" },
                { label: "Clientes Ativos", desc: "Lista os clientes ativos", icon: "bi-person-check-fill", url: "/carregar-listar-clientes-ativos" },
                { label: "Clientes Inativos", desc: "Lista os clientes inativos", icon: "bi-person-x-fill", url: "/carregar-listar-clientes-inativos" }
            ]
        },
        "configuracao": {
            label: "Parâmetros Técnicos",
            menuUrl: "/carregar-menu-configuracoes",
            items: [
                { label: "Gerenciar Cadastro Alarme", desc: "Cadastro dos dispositivos de alarme", icon: "bi-phone-vibrate", url: "/carregar-gerenciar-dispositivo" },
                { label: "Gerenciar Cadastro Usuário Alarme", desc: "Usuários vinculados aos alarmes", icon: "bi-person-vcard-fill", url: "/carregar-gerenciar-usuarios-alarme" },
                { label: "Gerenciar Cadastro Setores Alarme", desc: "Setores/partições dos alarmes", icon: "bi-bounding-box", url: "/carregar-gerenciar-setores-alarme" },
                { label: "Gerenciar Proced. Atendimento", desc: "Procedimentos padrão de atendimento", icon: "bi-headset", url: "/carregar-gerenciar-procedimento-atendimento" },
                { label: "Gerenciar Contactid Personalizado", desc: "Códigos Contact ID personalizados", icon: "bi-flag-fill", url: "/carregar-gerenciar-contactid-personalizado" },
                { label: "Gerenciar Grade Horário", desc: "Grades de horário de funcionamento", icon: "bi-alarm-fill", url: "/carregar-gerenciar-grade" },
                { label: "Gerenciar Email Evento", desc: "Configuração de envio de e-mail por evento", icon: "bi-envelope-arrow-up-fill", url: "/gerenciar-configuracao-email-eveto" }
            ]
        }
    }
};

function fpNavNormalize(path) {
    if (!path) return "/";
    var p = path.split("?")[0].split("#")[0];
    if (p.length > 1 && p.charAt(p.length - 1) === "/") {
        p = p.slice(0, -1);
    }
    return p.toLowerCase();
}

function fpNavFromCd() {
    try {
        return new URLSearchParams(window.location.search).get("from") === "cd";
    } catch (e) {
        return false;
    }
}

/* Descobre o contexto da pagina atual a partir do pathname.
   Retorna { mode, groupKey, label, desc } onde mode = 'menu' | 'item' | 'home' | null */
function fpNavResolve(path) {
    var cur = fpNavNormalize(path);
    var fromCd = fpNavFromCd();

    if (cur === "/carregar-menu-principal" || cur === "/" || cur === "") {
        return { mode: "home", groupKey: null, label: "Menu Principal", desc: "Escolha uma área para começar", icon: "bi-house-door-fill" };
    }

    for (var i = 0; i < FP_NAV.top.length; i++) {
        if (fpNavNormalize(FP_NAV.top[i].url) === cur) {
            return { mode: "menu", groupKey: FP_NAV.top[i].group, atalhoKey: FP_NAV.top[i].group, label: FP_NAV.top[i].label, desc: FP_NAV.top[i].desc, icon: FP_NAV.top[i].icon };
        }
    }

    // Submenus (grupos sem entrada no topo, ex.: Clientes dentro de Relatórios)
    for (var mk in FP_NAV.groups) {
        if (!FP_NAV.groups.hasOwnProperty(mk)) continue;
        var grpM = FP_NAV.groups[mk];
        if (fpNavNormalize(grpM.menuUrl) === cur) {
            var jaTop = false;
            for (var t = 0; t < FP_NAV.top.length; t++) {
                if (FP_NAV.top[t].group === mk) { jaTop = true; break; }
            }
            if (!jaTop) {
                return { mode: "item", groupKey: mk, atalhoKey: grpM.atalho || mk, label: grpM.label, desc: grpM.label, icon: "bi-grid-fill" };
            }
        }
    }

    // Com ?from=cd, prioriza Central de Disparos (mesma rota existe tambem em Relatorios)
    var groupKeys = Object.keys(FP_NAV.groups);
    if (fromCd) {
        groupKeys = groupKeys.filter(function (k) { return k === "central-disparos"; })
            .concat(groupKeys.filter(function (k) { return k !== "central-disparos"; }));
    }

    for (var gi = 0; gi < groupKeys.length; gi++) {
        var gk = groupKeys[gi];
        var grp = FP_NAV.groups[gk];
        var items = grp.items;
        for (var j = 0; j < items.length; j++) {
            if (fpNavNormalize(items[j].url) === cur) {
                return { mode: "item", groupKey: gk, atalhoKey: grp.atalho || gk, label: items[j].label, desc: items[j].desc, icon: items[j].icon };
            }
        }
    }

    return { mode: null, groupKey: null, atalhoKey: null, label: null, desc: null, icon: null };
}

function fpNavItemHtml(item, active, disabled, tooltip) {
    if (disabled) {
        var tip = tooltip || 'Módulo não disponível no seu plano.'
        var ch = typeof fpPermChavePorUrl === 'function' ? fpPermChavePorUrl(item.url) : ''
        return '<li><span class="dropdown-item fp-nav-item fp-nav-disabled fp-lic-bloqueado text-muted" role="button" tabindex="0"' +
            ' title="' + tip.replace(/"/g, '&quot;') + '"' +
            ' data-fp-lic-motivo="' + tip.replace(/"/g, '&quot;') + '"' +
            (ch ? ' data-fp-lic-chave="' + ch + '"' : '') +
            ' data-bs-toggle="tooltip" data-bs-placement="left">' +
            '<i class="bi ' + item.icon + '"></i><span>' + item.label + '</span></span></li>'
    }
    var cls = "dropdown-item fp-nav-item" + (active ? " fp-nav-active" : "");
    return '<li><a class="' + cls + '" href="' + item.url + '">' +
        '<i class="bi ' + item.icon + '"></i><span>' + item.label + '</span></a></li>';
}

function fpNavItemLicenciado(url) {
    if (typeof fpLicModuloLiberado !== 'function') return true
    var chave = typeof fpPermChavePorUrl === 'function' ? fpPermChavePorUrl(url) : null
    if (!chave) return true
    return fpLicModuloLiberado(chave)
}

function fpNavDivider() {
    return '<li><hr class="dropdown-divider"></li>';
}

function fpBuildNav() {
    var toggle = document.getElementById("fp-nav-toggle");
    var list = document.getElementById("fp-nav-items");
    var current = document.getElementById("fp-nav-current");
    if (!toggle || !list || !current) return;

    var ctx = fpNavResolve(window.location.pathname);
    var html = "";
    var label = ctx.label;

    if (typeof fpLicAcessoGlobalOk === 'function' && fpLicAcessoGlobalOk()) {
        // licenca ativa — nao esconder atalhos por estado stale de suspensao
    } else if (typeof fpLicDevePrenderFatura === 'function' && fpLicDevePrenderFatura()) {
        list.innerHTML = '<li><span class="dropdown-item text-muted small">Acesso suspenso — regularize em Faturas</span></li>'
        if (!label) label = 'Faturas e Cobrança'
        current.textContent = label
        fpBuildTitulo({ mode: 'item', label: 'Faturas e Cobrança', desc: 'Regularize sua assinatura', icon: 'bi-receipt' })
        var atalhos = document.getElementById('fp-atalhos')
        if (atalhos) atalhos.classList.add('d-none')
        return
    }

    if (ctx.mode === "menu") {
        for (var i = 0; i < FP_NAV.top.length; i++) {
            var t = FP_NAV.top[i];
            if (!fpPermNavItemVisivel(t.url, t.group)) continue;
            var licOk = fpNavItemLicenciado(t.url);
            var tip = licOk ? '' : (typeof fpLicMotivoModulo === 'function' ? fpLicMotivoModulo(t.group) : '');
            html += fpNavItemHtml(t, t.group === ctx.groupKey, !licOk, tip);
        }
    } else if (ctx.mode === "item") {
        var grp = FP_NAV.groups[ctx.groupKey];
        if (fpPermNavItemVisivel(grp.menuUrl, ctx.groupKey)) {
            var licMenu = fpNavItemLicenciado(grp.menuUrl);
            html += fpNavItemHtml({ label: "Menu " + grp.label, icon: "bi-grid-fill", url: grp.menuUrl }, false, !licMenu, '');
            html += fpNavDivider();
        }
        var cur = fpNavNormalize(window.location.pathname);
        for (var j = 0; j < grp.items.length; j++) {
            var it = grp.items[j];
            if (!fpPermNavItemVisivel(it.url)) continue;
            var licItem = fpNavItemLicenciado(it.url);
            var ch = typeof fpPermChavePorUrl === 'function' ? fpPermChavePorUrl(it.url) : '';
            var tipItem = licItem ? '' : (typeof fpLicMotivoModulo === 'function' ? fpLicMotivoModulo(ch) : '');
            var navDisabled = !licItem;
            html += fpNavItemHtml(it, fpNavNormalize(it.url) === cur, navDisabled, tipItem);
        }
    } else {
        for (var k = 0; k < FP_NAV.top.length; k++) {
            var topItem = FP_NAV.top[k];
            if (!fpPermNavItemVisivel(topItem.url, topItem.group)) continue;
            var licTop = fpNavItemLicenciado(topItem.url);
            html += fpNavItemHtml(topItem, false, !licTop, '');
        }
    }

    list.innerHTML = html;

    // Rotulo do combo: usa o contexto; se nao mapeado, mantem o NomeTela do backend.
    if (!label) {
        label = current.getAttribute("data-fallback") || "Menu";
    }
    current.textContent = label;

    fpBuildTitulo(ctx);
    fpBuildAtalhos(ctx);
    fpFixLinkRetorno(ctx);
    fpAplicarSubtitulosCards();
    fpAplicarPermissoesUI();
    if (typeof fpAplicarLicencaUI === 'function') fpAplicarLicencaUI();
}

/* Corrige o href do botão Voltar conforme a árvore FP_NAV (submenu imediato). */
function fpFixLinkRetorno(ctx) {
    var back = document.querySelector(".fp-page-header-back");
    if (!back) return;

    // Relatorios abertos pela Central de Disparos (?from=cd)
    if (fpNavFromCd()) {
        back.href = "/carregar-menu-central-disparos";
        return;
    }

    var cur = fpNavNormalize(window.location.pathname);

    if (!ctx.mode || ctx.mode === "home") {
        return;
    }

    if (ctx.mode === "menu") {
        back.href = "/carregar-menu-principal";
        return;
    }

    if (ctx.mode !== "item" || !ctx.groupKey) {
        return;
    }

    var grp = FP_NAV.groups[ctx.groupKey];
    if (!grp) return;

    // Página de submenu (ex.: /carregar-menu-operacional): volta ao menu pai
    if (fpNavNormalize(grp.menuUrl) === cur) {
        if (grp.atalho && FP_NAV.groups[grp.atalho]) {
            back.href = FP_NAV.groups[grp.atalho].menuUrl;
            return;
        }
        for (var i = 0; i < FP_NAV.top.length; i++) {
            if (FP_NAV.top[i].group === ctx.groupKey) {
                back.href = "/carregar-menu-principal";
                return;
            }
        }
        return;
    }

    // Página interna: volta ao submenu do grupo
    back.href = grp.menuUrl;
}

/* Titulo + subtitulo central e cabecalho da pagina. */
function fpBuildTitulo(ctx) {
    var elTitulo = document.getElementById("fp-screen-title");
    var elSub = document.getElementById("fp-screen-sub");
    if (!elTitulo) return;

    var titulo = ctx.label;
    if (!titulo) {
        titulo = elTitulo.getAttribute("data-fallback") || "";
    }
    elTitulo.textContent = titulo;
    if (elSub) {
        elSub.textContent = ctx.desc || "";
    }

    var elPageText = document.getElementById("fp-page-header-text");
    var elPageSub = document.getElementById("fp-page-header-sub");
    var elPageIcon = document.getElementById("fp-page-header-icon");

    if (elPageText) {
        if (ctx.label && !elPageText.getAttribute("data-keep-title")) {
            elPageText.textContent = ctx.label;
        } else if (!elPageText.textContent.trim()) {
            elPageText.textContent = elPageText.getAttribute("data-fallback") || titulo;
        }
    }

    if (elPageSub && ctx.desc && !elPageSub.getAttribute("data-keep-sub")) {
        if (!elPageSub.textContent.trim()) {
            elPageSub.textContent = ctx.desc;
        }
    }

    if (elPageIcon && ctx.icon) {
        elPageIcon.className = "bi " + ctx.icon;
    }
}

/* Barra de atalhos: aparece em todas as paginas menos a inicial
   (a inicial ja mostra os cards grandes). Destaca o grupo atual. */
function fpBuildAtalhos(ctx) {
    var atalhos = document.getElementById("fp-atalhos");
    if (!atalhos) return;

    if (ctx.mode === "home") {
        atalhos.classList.add("d-none");
        return;
    }

    atalhos.classList.remove("d-none");

    var alvo = ctx.atalhoKey || ctx.groupKey;
    var cards = atalhos.querySelectorAll(".fp-menu-card");
    for (var i = 0; i < cards.length; i++) {
        var g = cards[i].getAttribute("data-grupo");
        if (alvo && g === alvo) {
            cards[i].classList.add("fp-atalho-ativo");
        } else {
            cards[i].classList.remove("fp-atalho-ativo");
        }
    }
}

/* Botoes de configuracao (divs com role=button) mapeados para URL do FP_NAV */
var FP_NAV_BTN_URL = {
    BtnGerenciarCadastroAlarme: "/carregar-gerenciar-dispositivo",
    BtnGerenciarUsuarioAlarme: "/carregar-gerenciar-usuarios-alarme",
    BtnGerenciarSetoresAlarme: "/carregar-gerenciar-setores-alarme",
    BtnGerenciarProcedimentoAtendimento: "/carregar-gerenciar-procedimento-atendimento",
    BtnGerenciarContactidPersonalizado: "/carregar-gerenciar-contactid-personalizado",
    BtnGerenciarGradeHorario: "/carregar-gerenciar-grade",
    BtnGerenciarEnvioEvento: "/gerenciar-configuracao-email-eveto"
};

/* Monta mapa url -> descricao a partir do FP_NAV */
function fpNavMapaDesc() {
    var mapa = {};
    function registrar(item) {
        if (!item || !item.url) return;
        var texto = item.sub || item.desc;
        if (texto) mapa[fpNavNormalize(item.url)] = texto;
    }
    var i, gk, grp, j;
    for (i = 0; i < FP_NAV.top.length; i++) {
        registrar(FP_NAV.top[i]);
    }
    for (gk in FP_NAV.groups) {
        if (!FP_NAV.groups.hasOwnProperty(gk)) continue;
        grp = FP_NAV.groups[gk];
        for (j = 0; j < (grp.items || []).length; j++) {
            registrar(grp.items[j]);
        }
    }
    return mapa;
}

/* Injeta subtitulo em todos os cards de menu a partir do FP_NAV */
function fpAplicarSubtitulosCards() {
    var mapa = fpNavMapaDesc();
    var cards = document.querySelectorAll(".fp-menu-card");
    for (var i = 0; i < cards.length; i++) {
        var card = cards[i];
        if (card.querySelector(".fp-card-sub")) continue;

        var url = card.getAttribute("href");
        if (!url && card.id && FP_NAV_BTN_URL[card.id]) {
            url = FP_NAV_BTN_URL[card.id];
        }
        if (!url) continue;

        var desc = mapa[fpNavNormalize(url)];
        if (!desc) continue;

        var titulo = card.querySelector("span:not(.fp-card-count):not(.fp-card-sub)");
        if (!titulo) continue;

        titulo.classList.add("fp-card-title");

        var sub = document.createElement("small");
        sub.className = "fp-card-sub";
        sub.textContent = desc;

        var contador = card.querySelector(".fp-card-count");
        if (contador) {
            card.insertBefore(sub, contador);
        } else {
            card.appendChild(sub);
        }

        card.classList.add("fp-menu-card-sub");
    }
}

document.addEventListener("DOMContentLoaded", function () {
    if (typeof fpSyncMaster === 'function') {
        fpSyncMaster(function () { fpBuildNav(); });
    } else {
        fpBuildNav();
    }
});

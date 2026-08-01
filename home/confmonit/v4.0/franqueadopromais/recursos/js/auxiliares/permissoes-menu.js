/* Catalogo de permissoes — espelha FP_NAV com chaves estaveis */
var FP_PERM_CATALOGO = [
    {
        key: 'configuracao', label: 'Parâmetros Técnicos', children: [
            { key: 'configuracao.dispositivo', label: 'Gerenciar Cadastro Alarme' },
            { key: 'configuracao.usuarios-alarme', label: 'Gerenciar Usuário Alarme' },
            { key: 'configuracao.setores-alarme', label: 'Gerenciar Setores Alarme' },
            { key: 'configuracao.procedimento', label: 'Gerenciar Proced. Atendimento' },
            { key: 'configuracao.contactid', label: 'Gerenciar Contact ID Personalizado' },
            { key: 'configuracao.grade', label: 'Gerenciar Grade Horário' },
            { key: 'configuracao.email-evento', label: 'Gerenciar Email Evento' }
        ]
    },
    {
        key: 'atendimento', label: 'Atendimento', children: [
            { key: 'atendimento.gerar-evento', label: 'Gerar Evento' },
            { key: 'atendimento.inteligencia-artificial', label: 'Inteligência Artificial' },
            { key: 'minha-empresa.gestao.responsaveis', label: 'Gerenciar Responsáveis' },
            { key: 'minha-empresa.comercial.parceiro-monitoramento', label: 'Parceiro Monitoramento' }
        ]
    },
    {
        key: 'minha-empresa', label: 'Minha Empresa', children: [
            {
                key: 'minha-empresa.operacional', label: 'Operacional / Campo', children: [
                    { key: 'minha-empresa.operacional.tecnico', label: 'Cadastrar Técnico' },
                    { key: 'minha-empresa.operacional.viatura', label: 'Cadastro de Viatura' }
                ]
            },
            {
                key: 'minha-empresa.comercial', label: 'Comercial', children: [
                    { key: 'minha-empresa.comercial.cliente', label: 'Cadastro Cliente' }
                ]
            },
            {
                key: 'minha-empresa.gestao', label: 'Gestão Interna', children: [
                    { key: 'minha-empresa.gestao.dados', label: 'Dados do Monitoramento' },
                    { key: 'minha-empresa.gestao.meu-plano', label: 'Meu Plano' },
                    { key: 'minha-empresa.gestao.montar-plano', label: 'Montar Meu Plano' },
                    { key: 'minha-empresa.gestao.faturas', label: 'Faturas e Cobrança' },
                    { key: 'minha-empresa.gestao.permissoes', label: 'Permissões de Acesso' },
                    { key: 'whitelabel', label: 'White Label' },
                    { key: 'franqueadopro.saas', label: 'Domínio Personalizado' }
                ]
            }
        ]
    },
    {
        key: 'relatorio', label: 'Relatórios', children: [
            { key: 'relatorio.atendimento', label: 'Relatório de Atendimento' },
            { key: 'relatorio.ligacoes', label: 'Lista de Ligações' },
            {
                key: 'relatorio.eventos', label: 'Eventos', children: [
                    { key: 'relatorio.eventos.lista', label: 'Relatórios de Eventos' },
                    { key: 'relatorio.eventos.config-grupo', label: 'Relatório por Grupo' }
                ]
            },
            {
                key: 'relatorio.alarmes', label: 'Alarmes', children: [
                    { key: 'relatorio.alarmes.armados', label: 'Alarmes Armados' },
                    { key: 'relatorio.alarmes.desarmados', label: 'Alarmes Desarmados' }
                ]
            },
            {
                key: 'relatorio.clientes', label: 'Clientes', children: [
                    { key: 'relatorio.clientes.disp', label: 'Clientes e Dispositivos' },
                    { key: 'relatorio.clientes.lista', label: 'Lista de Clientes' },
                    { key: 'relatorio.clientes.ativos', label: 'Clientes Ativos' },
                    { key: 'relatorio.clientes.inativos', label: 'Clientes Inativos' }
                ]
            },
            {
                key: 'relatorio.central-disparos', label: 'Central de Disparos', children: [
                    { key: 'relatorio.central-disparos.unificado', label: 'Centro de Operações' },
                    { key: 'relatorio.central-disparos.eventos-pendentes', label: 'Eventos Pendentes' },
                    { key: 'relatorio.central-disparos.finalizados-robo', label: 'Finalizados por Robô' },
                    { key: 'relatorio.central-disparos.finalizados-bot', label: 'Finalizados pelo Bot' },
                    { key: 'relatorio.central-disparos.ligacao-historico', label: 'Ligação Histórico' },
                    { key: 'relatorio.central-disparos.sms-historico', label: 'SMS Histórico' },
                    { key: 'relatorio.central-disparos.eventos-falhas', label: 'Eventos com Falhas' },
                    { key: 'relatorio.central-disparos.whatsapp-enviados', label: 'WhatsApp Enviados' },
                    { key: 'relatorio.central-disparos.ligacoes', label: 'Ligações' },
                    { key: 'relatorio.central-disparos.sms', label: 'SMS' },
                    { key: 'relatorio.central-disparos.fila-envio', label: 'Eventos para Enviar ou Ligação' },
                    { key: 'relatorio.central-disparos.ligacao-erros', label: 'Ligações com Erros' },
                    { key: 'relatorio.central-disparos.fila-ligacao', label: 'Filas de Ligações' }
                ]
            }
        ]
    },
    {
        key: 'dashboard', label: 'Dashboard (Home)', children: [
            { key: 'dashboard.sem-comunicacao', label: 'Sem Comunicação' },
            { key: 'dashboard.eventos', label: 'Eventos / 7 dias' }
        ]
    }
]

function fpPermParseStorage() {
    try {
        var raw = localStorage.getItem('permissoesMenu')
        if (!raw) return null
        var arr = JSON.parse(raw)
        if (!arr || !arr.length) return null
        var map = {}
        arr.forEach(function (k) { map[k] = true })
        return map
    } catch (e) {
        return null
    }
}

function fpEhMaster() {
    return localStorage.getItem('loginMaster') === 'S'
}

function fpSyncMaster(next) {
    if (fpEhMaster()) {
        if (typeof next === 'function') next('S')
        return
    }
    $.ajax({
        url: '/sessaoMaster',
        method: 'POST',
        data: JSON.stringify({})
    }).fail(function () {
        if (typeof next === 'function') next('N')
    }).done(function (r) {
        var master = (r && r.dados && r.dados.master) || r.master || 'N'
        localStorage.setItem('loginMaster', master === 'S' ? 'S' : 'N')
        if (typeof next === 'function') next(master)
    })
}

function fpPermLiberado(chave) {
    if (fpEhMaster()) return true
    var chaves = fpPermParseStorage()
    if (!chaves) return true
    if (chaves[chave]) return true
    var prefix = chave + '.'
    for (var k in chaves) {
        if (chaves.hasOwnProperty(k) && k.indexOf(prefix) === 0) return true
    }
    return false
}

function fpPermItemVisivel(item) {
    if (!item || !item.key) return true
    if (fpEhMaster()) return true
    var chaves = fpPermParseStorage()
    if (!chaves) return true
    if (chaves[item.key]) return true
    if (fpPermTemFilhoLiberado(item)) return true
    return false
}

function fpPermTemFilhoLiberado(node) {
    if (!node.children || !node.children.length) return false
    for (var i = 0; i < node.children.length; i++) {
        if (fpPermItemVisivel(node.children[i])) return true
    }
    return false
}

function fpPermFiltrarCatalogo(nodes) {
    if (!nodes) return []
    var out = []
    nodes.forEach(function (n) {
        if (!fpPermItemVisivel(n)) return
        var copy = { key: n.key, label: n.label }
        if (n.children) {
            copy.children = fpPermFiltrarCatalogo(n.children)
        }
        out.push(copy)
    })
    return out
}

function fpPermColetarChaves(nodes, lista) {
    if (!lista) lista = []
    ;(nodes || []).forEach(function (n) {
        lista.push(n.key)
        if (n.children) fpPermColetarChaves(n.children, lista)
    })
    return lista
}

function fpPermMapaUrlChave() {
    return {
        '/carregar-menu-configuracoes': 'configuracao',
        '/carregar-gerenciar-dispositivo': 'configuracao.dispositivo',
        '/carregar-gerenciar-usuarios-alarme': 'configuracao.usuarios-alarme',
        '/carregar-gerenciar-setores-alarme': 'configuracao.setores-alarme',
        '/carregar-gerenciar-procedimento-atendimento': 'configuracao.procedimento',
        '/carregar-gerenciar-contactid-personalizado': 'configuracao.contactid',
        '/carregar-gerenciar-grade': 'configuracao.grade',
        '/gerenciar-configuracao-email-eveto': 'configuracao.email-evento',
        '/carregar-menu-atendimento': 'atendimento',
        '/CarregarGeradorEventos': 'atendimento.gerar-evento',
        '/carregar-inteligencia-artificial': 'atendimento.inteligencia-artificial',
        '/carregar-menu-minha-empresa': 'minha-empresa',
        '/carregar-menu-operacional': 'minha-empresa.operacional',
        '/CarregarPaginaGerenciarTecnico': 'minha-empresa.operacional.tecnico',
        '/CarregarPaginaGerenciarViatura': 'minha-empresa.operacional.viatura',
        '/carregar-menu-comercial': 'minha-empresa.comercial',
        '/CarregarPaginaGerenciarCliente': 'minha-empresa.comercial.cliente',
        '/carregar-parceiro-monitoramento': 'minha-empresa.comercial.parceiro-monitoramento',
        '/carregar-menu-gestao': 'minha-empresa.gestao',
        '/CarregarPaginaDadosMonitoramento': 'minha-empresa.gestao.dados',
        '/carregarPaginaGerenciarUsuarios': 'minha-empresa.gestao.responsaveis',
        '/carregar-permissoes-acesso': 'minha-empresa.gestao.permissoes',
        '/carregar-meu-plano': 'minha-empresa.gestao.meu-plano',
        '/carregar-montar-plano': 'minha-empresa.gestao.montar-plano',
        '/carregar-whitelabel': 'whitelabel',
        '/carregar-dominio-saas': 'franqueadopro.saas',
        '/carregar-menu-relatorio': 'relatorio',
        '/carregar-relatorio-atendimento': 'relatorio.atendimento',
        '/carregar-relatorio-ligacoes': 'relatorio.ligacoes',
        '/carregar-menu-eventos': 'relatorio.eventos',
        '/carregar-relatorio-eventos': 'relatorio.eventos.lista',
        '/carregar-configurar-relatorio-cliente': 'relatorio.eventos.config-grupo',
        '/carregar-menu-alarmes': 'relatorio.alarmes',
        '/carregar-listar-dispositivos-armados': 'relatorio.alarmes.armados',
        '/carregar-listar-dispositivos-desarmados': 'relatorio.alarmes.desarmados',
        '/carregar-menu-clientes': 'relatorio.clientes',
        '/listar-cliente-dispositivo': 'relatorio.clientes.disp',
        '/carregar-listar-clientes': 'relatorio.clientes.lista',
        '/carregar-listar-clientes-ativos': 'relatorio.clientes.ativos',
        '/carregar-listar-clientes-inativos': 'relatorio.clientes.inativos',
        '/carregar-menu-central-disparos': 'relatorio.central-disparos',
        '/carregar-relatorio-cd-unificado': 'relatorio.central-disparos.unificado',
        '/cdUnificadoLista': 'relatorio.central-disparos.unificado',
        '/cdUnificadoDetalhe': 'relatorio.central-disparos.unificado',
        '/carregar-relatorio-eventos-pendentes': 'relatorio.central-disparos.eventos-pendentes',
        '/carregar-relatorio-finalizados-robo': 'relatorio.central-disparos.finalizados-robo',
        '/carregar-relatorio-finalizados-bot': 'relatorio.central-disparos.finalizados-bot',
        '/carregar-relatorio-ligacao-historico': 'relatorio.central-disparos.ligacao-historico',
        '/ligacaoHistoricoListar': 'relatorio.central-disparos.ligacao-historico',
        '/cdLigacaoHistoricoListar': 'relatorio.central-disparos.ligacao-historico',
        '/ligacaoHistoricoAudio': 'relatorio.central-disparos.ligacao-historico',
        '/carregar-relatorio-sms-historico': 'relatorio.central-disparos.sms-historico',
        '/carregar-relatorio-eventos-falhas': 'relatorio.central-disparos.eventos-falhas',
        '/carregar-relatorio-whatsapp-enviados': 'relatorio.central-disparos.whatsapp-enviados',
        '/carregar-relatorio-ligacoes-cd': 'relatorio.central-disparos.ligacoes',
        '/cdLigacoesListar': 'relatorio.central-disparos.ligacoes',
        '/carregar-relatorio-sms': 'relatorio.central-disparos.sms',
        '/cdSmsListar': 'relatorio.central-disparos.sms',
        '/cdWhatsappEnviadosListar': 'relatorio.central-disparos.whatsapp-enviados',
        '/carregar-relatorio-fila-envio': 'relatorio.central-disparos.fila-envio',
        '/carregar-relatorio-ligacao-erros': 'relatorio.central-disparos.ligacao-erros',
        '/carregar-relatorio-fila-ligacao': 'relatorio.central-disparos.fila-ligacao',
        '/carregar-dashboard-sem-comunicacao': 'dashboard.sem-comunicacao',
        '/carregar-dashboard-eventos': 'dashboard.eventos'
    }
}

function fpPermChavePorUrl(url) {
    var full = url || ''
    var path = full.split('?')[0]
    if (path.length > 1 && path.charAt(path.length - 1) === '/') {
        path = path.slice(0, -1)
    }
    var lower = path.toLowerCase()
    if (lower === '/carregar-meu-plano') {
        var qs = full.indexOf('?') >= 0 ? full.split('?')[1] : ''
        if (qs) {
            try {
                if (new URLSearchParams(qs).get('tab') === 'faturas') {
                    return 'minha-empresa.gestao.faturas'
                }
            } catch (e) { /* ignore */ }
        }
        return 'minha-empresa.gestao.meu-plano'
    }
    var map = fpPermMapaUrlChave()
    if (map[path]) return map[path]
    for (var k in map) {
        if (map.hasOwnProperty(k) && k.toLowerCase() === lower) return map[k]
    }
    return null
}

function fpPermNavItemVisivel(url, grupoKey) {
    if (fpEhMaster()) return true
    var chave = fpPermChavePorUrl(url)
    if (!chave && grupoKey) chave = grupoKey
    if (!chave) return true
    if (chave === 'minha-empresa.gestao.permissoes') return false
    return fpPermLiberado(chave)
}

function fpAplicarPermissoesUI() {
    if (fpEhMaster()) {
        if (typeof fpAplicarLicencaUI === 'function') fpAplicarLicencaUI()
        return
    }
    $('a.fp-menu-card[href]').each(function () {
        var chave = fpPermChavePorUrl($(this).attr('href'))
        if (chave && !fpPermLiberado(chave)) {
            $(this).closest('[class*="col-"]').hide()
        }
    })
    $('#fp-atalhos .fp-menu-card').each(function () {
        var g = $(this).attr('data-grupo')
        if (g && !fpPermLiberado(g)) {
            $(this).closest('[class*="col-"]').hide()
        }
    })
    if (typeof fpAplicarLicencaUI === 'function') fpAplicarLicencaUI()
}

function fpPermCarregarLogado(next) {
    var concluido = false
    var concluir = function () {
        if (concluido) return
        concluido = true
        if (typeof next === 'function') next()
    }

    if (fpEhMaster()) {
        localStorage.removeItem('permissoesMenu')
        concluir()
        return
    }

    var timeoutId = setTimeout(function () {
        localStorage.removeItem('permissoesMenu')
        concluir()
    }, 8000)

    $.ajax({
        url: '/permissoesCarregarLogado',
        method: 'POST',
        timeout: 7000,
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            idUsuario: localStorage.getItem('idUsuario')
        })
    }).always(function () {
        clearTimeout(timeoutId)
    }).fail(function () {
        localStorage.removeItem('permissoesMenu')
        concluir()
    }).done(function (r) {
        var lista = Array.isArray(r) ? r : (r.dados || r.items || [])
        if (!lista.length) {
            localStorage.removeItem('permissoesMenu')
        } else {
            var chaves = []
            lista.forEach(function (p) {
                var k = p.chave_menu || p.chaveMenu
                if (k && (p.liberado === 'S' || p.liberado === true)) chaves.push(k)
            })
            localStorage.setItem('permissoesMenu', JSON.stringify(chaves))
        }
        concluir()
    })
}

function fpFlagAtiva(val) {
    var v = String(val == null ? '' : val).toUpperCase()
    return v === '1' || v === 'S' || v === 'SIM' || v === 'TRUE'
}

function fpSalvarFlagsPacote(pacote) {
    var p = pacote || {}
    localStorage.setItem('terminalAtivo', fpFlagAtiva(p.terminal) ? '1' : '0')
    localStorage.setItem('gradeAtivo', fpFlagAtiva(p.grade) ? '1' : '0')
}

function fpTerminalAtivo() {
    return fpFlagAtiva(localStorage.getItem('terminalAtivo'))
}

function fpGradeAtivo() {
    return fpFlagAtiva(localStorage.getItem('gradeAtivo'))
}

function fpCarregarFlagsPacote(idPacote, next) {
    if (!idPacote) {
        fpSalvarFlagsPacote({})
        if (typeof next === 'function') next()
        return
    }
    $.ajax({
        url: '/pacote/buscar',
        method: 'POST',
        data: JSON.stringify({ idPacote: idPacote })
    }).fail(function () {
        fpSalvarFlagsPacote({})
    }).done(function (r) {
        var p = (r && r.dados) || r || {}
        fpSalvarFlagsPacote(p)
    }).always(function () {
        if (typeof next === 'function') next()
    })
}


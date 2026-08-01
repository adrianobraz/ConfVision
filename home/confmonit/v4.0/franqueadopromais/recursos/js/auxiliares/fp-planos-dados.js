/* Fonte canonica JS — modulos, limites e produtos (espelha planos.go) */

var FP_MODULOS_CATALOGO = [
    { chave: 'configuracao.dispositivo', label: 'Cadastro de Alarmes', grupo: 'Parâmetros Técnicos', desc: 'Cadastro e gestão de dispositivos/contas de alarme.', minimo: 'lite' },
    { chave: 'configuracao.usuarios-alarme', label: 'Usuários de Alarme', grupo: 'Parâmetros Técnicos', desc: 'Usuários vinculados aos painéis de alarme.', minimo: 'lite' },
    { chave: 'configuracao.setores-alarme', label: 'Setores de Alarme', grupo: 'Parâmetros Técnicos', desc: 'Partições e setores dos alarmes.', minimo: 'lite' },
    { chave: 'configuracao.grade', label: 'Grade Horário', grupo: 'Parâmetros Técnicos', desc: 'Grades de horário de funcionamento e armado.', minimo: 'pro' },
    { chave: 'configuracao.procedimento', label: 'Procedimento de Atendimento', grupo: 'Parâmetros Técnicos', desc: 'Roteiros padrão para operadores no atendimento.', minimo: 'pro_plus' },
    { chave: 'configuracao.contactid', label: 'Contact ID Personalizado', grupo: 'Parâmetros Técnicos', desc: 'Códigos Contact ID customizados por central.', minimo: 'pro_plus' },
    { chave: 'configuracao.email-evento', label: 'E-mail por Evento', grupo: 'Parâmetros Técnicos', desc: 'Envio automático de e-mail quando eventos ocorrem.', minimo: 'pro' },
    { chave: 'atendimento.gerar-evento', label: 'Gerar Evento', grupo: 'Atendimento', desc: 'Geração manual de eventos para clientes.', minimo: 'lite' },
    { chave: 'atendimento.inteligencia-artificial', label: 'Inteligência Artificial', grupo: 'Atendimento', desc: 'Alertas por telefone, SMS e bloqueios inteligentes.', minimo: 'pro_plus' },
    { chave: 'minha-empresa.comercial.cliente', label: 'Cadastro de Clientes', grupo: 'Comercial', desc: 'Cadastro e gestão de clientes da central.', minimo: 'lite' },
    { chave: 'minha-empresa.operacional', label: 'Operacional / Campo', grupo: 'Operacional', desc: 'Área de técnicos e viaturas em campo.', minimo: 'pro' },
    { chave: 'minha-empresa.operacional.tecnico', label: 'Cadastro de Técnico', grupo: 'Operacional', desc: 'Cadastro de técnicos que atendem em campo.', minimo: 'pro' },
    { chave: 'minha-empresa.operacional.viatura', label: 'Cadastro de Viatura', grupo: 'Operacional', desc: 'Veículos utilizados nas ordens de serviço.', minimo: 'pro' },
    { chave: 'minha-empresa.gestao.dados', label: 'Dados do Monitoramento', grupo: 'Gestão Interna', desc: 'Informações e parâmetros da empresa.', minimo: 'lite' },
    { chave: 'minha-empresa.gestao.responsaveis', label: 'Gerenciar Responsáveis', grupo: 'Gestão Interna', desc: 'Responsáveis legais e contatos da central.', minimo: 'lite' },
    { chave: 'minha-empresa.gestao.meu-plano', label: 'Meu Plano', grupo: 'Gestão Interna', desc: 'Plano contratado, licença e resumo.', minimo: 'lite' },
    { chave: 'minha-empresa.gestao.montar-plano', label: 'Montar Meu Plano', grupo: 'Gestão Interna', desc: 'Contratação à la carte de módulos.', minimo: 'lite' },
    { chave: 'minha-empresa.gestao.faturas', label: 'Faturas e Cobrança', grupo: 'Gestão Interna', desc: 'Histórico e faturas em aberto.', minimo: 'lite' },
    { chave: 'minha-empresa.gestao.permissoes', label: 'Permissões de Acesso', grupo: 'Gestão Interna', desc: 'Controle fino de menus por usuário.', minimo: 'pro_plus' },
    { chave: 'whitelabel', label: 'White Label', grupo: 'Gestão Interna', desc: 'Personalização de cores e logotipo da marca.', minimo: 'pro' },
    { chave: 'relatorio.atendimento', label: 'Relatório de Atendimento', grupo: 'Relatórios', desc: 'Histórico detalhado de atendimentos realizados.', minimo: 'pro' },
    { chave: 'relatorio.ligacoes', label: 'Lista de Ligações', grupo: 'Relatórios', desc: 'Relatório de ligações telefônicas.', minimo: 'pro' },
    { chave: 'relatorio.eventos.lista', label: 'Relatório de Eventos', grupo: 'Relatórios', desc: 'Lista de eventos recebidos no período.', minimo: 'lite' },
    { chave: 'relatorio.eventos.config-grupo', label: 'Eventos por Grupo', grupo: 'Relatórios', desc: 'Relatório avançado agrupado por grupo de eventos.', minimo: 'pro' },
    { chave: 'relatorio.alarmes.armados', label: 'Alarmes Armados', grupo: 'Relatórios', desc: 'Dispositivos com alarme armado.', minimo: 'lite' },
    { chave: 'relatorio.alarmes.desarmados', label: 'Alarmes Desarmados', grupo: 'Relatórios', desc: 'Dispositivos com alarme desarmado.', minimo: 'lite' },
    { chave: 'relatorio.clientes.disp', label: 'Clientes e Dispositivos', grupo: 'Relatórios', desc: 'Relação de clientes e suas contas.', minimo: 'lite' },
    { chave: 'relatorio.clientes.lista', label: 'Lista de Clientes', grupo: 'Relatórios', desc: 'Todos os clientes cadastrados.', minimo: 'lite' },
    { chave: 'relatorio.clientes.ativos', label: 'Clientes Ativos', grupo: 'Relatórios', desc: 'Clientes com situação ativa.', minimo: 'lite' },
    { chave: 'relatorio.clientes.inativos', label: 'Clientes Inativos', grupo: 'Relatórios', desc: 'Clientes inativos ou cancelados.', minimo: 'lite' },
    { chave: 'relatorio.central-disparos', label: 'Central de Disparos', grupo: 'Relatórios', desc: 'Relatórios de WhatsApp, SMS, ligações e disparos automáticos (incluso no Pro+ ou com módulo Inteligência Artificial).', minimo: 'pro_plus' },
    { chave: 'dashboard.sem-comunicacao', label: 'Dashboard Sem Comunicação', grupo: 'Dashboard', desc: 'Painel de contas sem comunicação por faixa de tempo.', minimo: 'lite' },
    { chave: 'dashboard.eventos', label: 'Dashboard Eventos (7 dias)', grupo: 'Dashboard', desc: 'KPIs de eventos por grupo nos últimos 7 dias.', minimo: 'pro' },
    { chave: 'franqueadopro.saas', label: 'Domínio Personalizado', grupo: 'Gestão Interna', desc: 'Configure o endereço do FranqueadoPro na nuvem com Apache e SSL automáticos.', minimo: 'pro_plus' }
]

var FP_MODULOS_CHAVES = FP_MODULOS_CATALOGO.map(function (m) { return m.chave })

var FP_PLANO_MINIMO_POR_CHAVE = {}
FP_MODULOS_CATALOGO.forEach(function (m) {
    FP_PLANO_MINIMO_POR_CHAVE[m.chave] = m.minimo
})

var FP_LIMITES_LINHAS = [
    { chave: 'clientes_max', label: 'Clientes cadastrados' },
    { chave: 'contas_max', label: 'Contas / dispositivos de alarme' },
    { chave: 'usuarios_alarme_max', label: 'Usuários de alarme' },
    { chave: 'setores_alarme_max', label: 'Setores de alarme' }
]

var FP_PRODUTOS_ECOSISTEMA = [
    {
        id: 'webterminal',
        nome: 'WebTerminal',
        desc: 'Terminal de atendimento web para operadores: fila de eventos, atendimento e comandos.',
        icon: 'bi-display',
        url: 'https://terminal.confmonit2.com.br',
        produto: 'webterminal',
        minimo: 'pro',
        planos: ['lite', 'pro', 'pro_plus'],
        inclusoEm: { pro: 'lite', pro_plus: 'pro' }
    },
    {
        id: 'terminalmovel',
        nome: 'Terminal Móvel',
        desc: 'App de atendimento em campo para operadores e técnicos no celular ou tablet.',
        icon: 'bi-phone-fill',
        url: 'https://terminalmovel.confmonit2.com.br',
        produto: 'terminalmovel',
        minimo: 'pro',
        planos: ['padrao'],
        inclusoEm: { pro: 'padrao', pro_plus: 'padrao' }
    },
    {
        id: 'confvision',
        nome: 'Vision',
        desc: 'Videomonitoramento: plano mensal do software + licença por câmera.',
        icon: 'bi-camera-video-fill',
        url: 'https://vision.confmonit2.com.br',
        produto: 'confvision',
        minimo: 'pro_plus',
        planos: ['padrao'],
        inclusoEm: { pro_plus: 'padrao' }
    },
    {
        id: 'webambiente',
        nome: 'webAmbiente',
        desc: 'Portal web para o cliente final consultar histórico, câmeras e dados da conta.',
        icon: 'bi-globe2',
        url: 'https://ambiente.confmonit2.com.br',
        produto: 'webambiente',
        minimo: 'pro_plus',
        planos: ['pro', 'pro_plus'],
        inclusoEm: { pro_plus: 'pro' }
    },
    {
        id: 'dialyze',
        nome: 'Dialyze',
        desc: 'CRM e atendimento omnichannel: WhatsApp, inbox, kanban, automações e IA para a central.',
        icon: 'bi-chat-left-text-fill',
        url: 'https://crm.confmonit.com.br',
        produto: 'dialyze',
        minimo: 'pro_plus',
        planos: ['padrao'],
        inclusoEm: { pro_plus: 'padrao' }
    },
    {
        id: 'saas',
        nome: 'FranqueadoPro SaaS',
        desc: 'Hospedagem na nuvem, sem servidor próprio. Lite e Pro: somente instalação on-premise.',
        icon: 'bi-cloud-check-fill',
        minimo: 'pro_plus',
        infoOnly: true,
        ativacao: 'Ativação pela sua central ou pelo representante da rede.'
    }
]

function fpModulosPorPlano(plano) {
    var out = {}
    FP_MODULOS_CATALOGO.forEach(function (m) {
        out[m.chave] = fpPlanoLiberadoNoPlano(m.minimo, plano)
    })
    return out
}

function fpModulosFalseLite() {
    var o = {}
    FP_MODULOS_CATALOGO.forEach(function (m) {
        if (m.minimo === 'pro' || m.minimo === 'pro_plus') o[m.chave] = false
    })
    return o
}

function fpModulosFalsePro() {
    var o = {}
    FP_MODULOS_CATALOGO.forEach(function (m) {
        if (m.minimo === 'pro_plus') o[m.chave] = false
    })
    return o
}

function fpModulosTrueProPlus() {
    var o = {}
    FP_MODULOS_CHAVES.forEach(function (k) { o[k] = true })
    return o
}

var FP_PLANOS_PADRAO = [
    {
        plano: 'lite',
        nome_exibicao: 'FranqueadoPro Lite',
        valor_mensal: 99,
        retencao_dias: 30,
        limites_json: { clientes_max: 50, contas_max: 100, usuarios_alarme_max: 5, setores_alarme_max: 4 },
        modulos_json: fpModulosPorPlano('lite')
    },
    {
        plano: 'pro',
        nome_exibicao: 'FranqueadoPro Pro',
        valor_mensal: 199,
        retencao_dias: 90,
        limites_json: { clientes_max: 500, contas_max: 1000, usuarios_alarme_max: 20, setores_alarme_max: 15 },
        modulos_json: fpModulosPorPlano('pro')
    },
    {
        plano: 'pro_plus',
        nome_exibicao: 'FranqueadoPro Pro+',
        valor_mensal: 299,
        retencao_dias: 180,
        limites_json: { clientes_max: 9999, contas_max: 9999, usuarios_alarme_max: 999, setores_alarme_max: 999 },
        modulos_json: fpModulosPorPlano('pro_plus')
    }
]

function fpPlanoNomeExibicao(plano) {
    if (plano === 'pro_plus') return 'Pro+'
    if (plano === 'pro') return 'Pro'
    if (plano === 'lite') return 'Lite'
    return (plano || '').replace('_', '+')
}

function fpPlanoLimiteValor(plano, chave) {
    var v = (plano.limites_json || {})[chave]
    if (v == null) return '—'
    return v >= 9999 ? 'Ilimitado*' : String(v)
}

function fpPlanoLabelMinimo(minimo) {
    if (minimo === 'pro_plus') return 'Pro+'
    if (minimo === 'pro') return 'Pro'
    return 'Lite'
}

function fpPlanoModuloLiberadoObj(plano, chave) {
    var planoKey = typeof plano === 'string' ? plano : (plano && plano.plano)
    var min = FP_PLANO_MINIMO_POR_CHAVE[chave] || 'lite'
    return fpPlanoLiberadoNoPlano(min, planoKey)
}

function fpModuloLiberadoEfetivo(plano, chave) {
    var min = FP_PLANO_MINIMO_POR_CHAVE[chave] || 'lite'
    return fpPlanoLiberadoNoPlano(min, plano)
}

function fpPlanoLiberadoNoPlano(minimo, plano) {
    if (!plano) return false
    if (plano === 'pro_plus') return true
    if (plano === 'pro') return minimo === 'lite' || minimo === 'pro'
    if (plano === 'lite') return minimo === 'lite'
    return false
}

function fpPlanoProdutoLiberado(plano, produtoId) {
    var pr = null
    for (var i = 0; i < FP_PRODUTOS_ECOSISTEMA.length; i++) {
        if (FP_PRODUTOS_ECOSISTEMA[i].id === produtoId) { pr = FP_PRODUTOS_ECOSISTEMA[i]; break }
    }
    if (!pr) return false
    if (pr.infoOnly) {
        var planoKeyInfo = typeof plano === 'string' ? plano : (plano && plano.plano)
        return fpPlanoLiberadoNoPlano(pr.minimo || 'pro_plus', planoKeyInfo)
    }
    var planoKey = typeof plano === 'string' ? plano : (plano && plano.plano)
    if (!planoKey) return false
    if (pr.inclusoEm && pr.inclusoEm[planoKey]) return true
    return fpPlanoLiberadoNoPlano(pr.minimo || 'pro_plus', planoKey)
}

function fpPlanoProdutoInclusoTier(plano, produtoId) {
    var pr = null
    for (var i = 0; i < FP_PRODUTOS_ECOSISTEMA.length; i++) {
        if (FP_PRODUTOS_ECOSISTEMA[i].id === produtoId) { pr = FP_PRODUTOS_ECOSISTEMA[i]; break }
    }
    if (!pr || !pr.inclusoEm) return ''
    var planoKey = typeof plano === 'string' ? plano : (plano && plano.plano)
    return pr.inclusoEm[planoKey] || ''
}

function fpPlanoProdutoMotivo(produtoId) {
    var pr = null
    for (var i = 0; i < FP_PRODUTOS_ECOSISTEMA.length; i++) {
        if (FP_PRODUTOS_ECOSISTEMA[i].id === produtoId) { pr = FP_PRODUTOS_ECOSISTEMA[i]; break }
    }
    if (!pr) return 'Produto não disponível no seu plano.'
    if (pr.id === 'saas') return 'Modo SaaS (nuvem) disponível apenas no plano Pro+. Lite e Pro utilizam instalação própria.'
    var min = pr.minimo || 'pro_plus'
    if (min === 'pro_plus') {
        return pr.nome + ' está incluso no Pro+ do FranqueadoPro, ou pode ser contratado à parte em Meu Plano.'
    }
    if (min === 'pro') {
        return pr.nome + ' está incluso a partir do Pro do FranqueadoPro, ou pode ser contratado à parte em Meu Plano.'
    }
    return pr.nome + ' pode ser contratado à parte em Meu Plano.'
}

function fpLicMergeModulosDados(plano, api) {
    var base = fpModulosPorPlano(plano || '')
    var flat = fpLicFlattenModulos(api)
    if (!Object.keys(flat).length) return base
    var out = {}
    FP_MODULOS_CHAVES.forEach(function (k) {
        if (flat[k] === true) out[k] = true
        else if (flat[k] === false) out[k] = false
        else out[k] = base[k]
    })
    return out
}

function fpLicFlattenModulos(api) {
    var out = {}
    if (!api || typeof api !== 'object') return out
    Object.keys(api).forEach(function (k) {
        var v = api[k]
        if (v === true || v === false) {
            if (k === 'franqueadopro') out['franqueadopro.saas'] = v
            else out[k] = v
            return
        }
        if (v && typeof v === 'object' && !Array.isArray(v)) {
            Object.keys(v).forEach(function (sk) {
                if (v[sk] === true || v[sk] === false) {
                    out[k + '.' + sk] = v[sk]
                }
            })
            return
        }
        if (typeof v === 'number') {
            if (k === 'franqueadopro') out['franqueadopro.saas'] = v !== 0
            else out[k] = v !== 0
        }
    })
    return out
}

function fpLicModuloLiberadoPorMapa(mod, chave) {
    if (!chave) return true
    if (mod[chave] === true) return true
    if (mod[chave] === false) return false
    var parts = chave.split('.')
    for (var i = parts.length - 1; i >= 1; i--) {
        var parent = parts.slice(0, i).join('.')
        if (mod[parent] === false) return false
    }
    var j
    for (j = 0; j < FP_MODULOS_CATALOGO.length; j++) {
        if (FP_MODULOS_CATALOGO[j].chave === chave) return false
    }
    return true
}

function fpLicMotivoPlano(chave) {
    var nome = chave
    for (var i = 0; i < FP_MODULOS_CATALOGO.length; i++) {
        if (FP_MODULOS_CATALOGO[i].chave === chave) { nome = FP_MODULOS_CATALOGO[i].label; break }
    }
    if (chave && chave.indexOf('relatorio.central-disparos') === 0) {
        return 'Para habilitar Central de Disparos, assine o plano Pro+ ou contrate o módulo Inteligência Artificial em Meu Plano.'
    }
    var min = FP_PLANO_MINIMO_POR_CHAVE[chave]
    if (min === 'pro_plus') return 'Para habilitar ' + nome + ', assine o plano Pro+.'
    if (min === 'pro') return 'Para habilitar ' + nome + ', assine o plano Pro ou Pro+.'
    return nome + ' não está incluído no seu plano. Acesse Meu Plano para fazer upgrade.'
}

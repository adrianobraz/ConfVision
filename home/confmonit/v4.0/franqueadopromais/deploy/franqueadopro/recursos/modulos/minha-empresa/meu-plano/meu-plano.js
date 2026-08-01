$(document).ready(function () {
    carregarResumo()
    fpPlanoCarregarPacotesCota()

    $('#plano-tabs .nav-link').on('click', function () {
        if ($(this).hasClass('disabled') || $(this).attr('aria-disabled') === 'true') return
        var tab = $(this).data('tab')
        fpPlanoAtivarTab(tab)
    })

    $('#fatura-filtro').on('change', carregarFaturas)
    $('#fatura-atualizar').on('click', carregarFaturas)
    $('#fatura-aplicar-cupom').on('click', fpPlanoAbrirModalCupomToolbar)

    var params = new URLSearchParams(window.location.search)
    if (params.get('tab') === 'faturas') {
        fpPlanoAtivarTab('faturas')
    }

    $(document).on('click', '.fp-fatura-pagar-btn', function () {
        var idx = $(this).data('fatura-idx')
        if (idx == null || !window.__fpFaturasLista) return
        fpPlanoAbrirModalPagar(window.__fpFaturasLista[idx])
    })

    $(document).on('change', '#fp-fatura-cupom-fatura', function () {
        fpPlanoAtualizarInfoCupomFatura()
    })

    $(document).on('click', '.fp-plano-contratar-btn', function () {
        var plano = $(this).data('plano')
        if (!plano) return
        fpPlanoContratar(plano)
    })

    $(document).on('click', '#fp-plano-detalhes-btn', function () {
        fpPlanoAbrirModalModulos()
    })

    $('#fp-plano-cupom-validar').on('click', fpPlanoValidarCupom)
    $('#fp-fatura-cupom-aplicar').on('click', fpPlanoAplicarCupomFatura)
})

window.__fpCupomCodigo = ''
window.__fpCupomPreview = null
window.__fpFaturaCupomAtual = null
window.__fpFaturasCupomOpts = []

function fpPlanoCupomCodigoAtual() {
    return String($('#fp-plano-cupom-codigo').val() || window.__fpCupomCodigo || '').trim().toUpperCase()
}

function fpPlanoValidarCupom() {
    var codigo = fpPlanoCupomCodigoAtual()
    var $msg = $('#fp-plano-cupom-msg')
    if (!codigo) {
        window.__fpCupomCodigo = ''
        window.__fpCupomPreview = null
        $msg.removeClass('text-success text-danger').addClass('text-muted').text('Informe um código de cupom.')
        return
    }
    $msg.removeClass('text-success text-danger').addClass('text-muted').text('Validando…')
    $.ajax({
        url: '/licencaCupomValidar',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ codigo: codigo, produto: 'franqueadopro' })
    }).done(function (r) {
        window.__fpCupomCodigo = codigo
        window.__fpCupomPreview = r
        var cupom = r.cupom || {}
        var desc = cupom.tipo === 'percentual'
            ? (Number(cupom.valor || 0) + '%')
            : ('R$ ' + Number(cupom.valor || 0).toFixed(2))
        $msg.removeClass('text-muted text-danger').addClass('text-success')
            .text('Cupom válido: ' + desc + ' de desconto na fatura da contratação.')
    }).fail(function (xhr) {
        window.__fpCupomCodigo = ''
        window.__fpCupomPreview = null
        var msg = 'Cupom inválido.'
        try {
            var j = JSON.parse(xhr.responseText || '')
            if (j.message) msg = j.message
            else if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
        } catch (e) { /* ignore */ }
        $msg.removeClass('text-muted text-success').addClass('text-danger').text(msg)
    })
}

function fpPlanoAtivarTab(tab) {
    $('#plano-tabs .nav-link').removeClass('active')
    $('#plano-tabs .nav-link[data-tab="' + tab + '"]').addClass('active')
    if (tab === 'faturas') {
        $('#tab-plano').addClass('d-none')
        $('#tab-faturas').removeClass('d-none')
        carregarFaturas()
    } else {
        $('#tab-faturas').addClass('d-none')
        $('#tab-plano').removeClass('d-none')
    }
}

function fpPlanoAplicarModoSuspenso(r) {
    if (!r || r.liberado !== false) {
        if (typeof fpLicRemoverModoSuspenso === 'function') fpLicRemoverModoSuspenso()
        return
    }
    var prender = r.motivo === 'suspensa' || r.motivo === 'vencida' || r.motivo === 'pendente'
    if (!prender) return

    var titulo = r.motivo === 'suspensa'
        ? 'Sistema suspenso'
        : (r.motivo === 'pendente' ? 'Aguardando pagamento' : 'Assinatura vencida')
    var texto = r.motivo === 'suspensa'
        ? 'Sua assinatura está suspensa. O acesso ao FranqueadoPro ficará bloqueado até a regularização da fatura em aberto.'
        : (r.motivo === 'pendente'
            ? 'Sua assinatura foi gerada e aguarda pagamento. Regularize a fatura em aberto para liberar o acesso ao sistema.'
            : 'Sua assinatura venceu. Regularize a fatura em aberto para restaurar o acesso ao sistema.')

    $('#fp-plano-alerta-suspensao')
        .removeClass('d-none')
        .html(
            '<strong><i class="bi bi-exclamation-triangle-fill"></i> ' + titulo + '</strong>' +
            '<p class="mb-0 mt-2">' + texto + ' Utilize os dados abaixo para pagamento.</p>'
        )

    $('#plano-tabs .nav-link[data-tab="plano"]')
        .addClass('disabled')
        .attr('aria-disabled', 'true')
        .attr('title', 'Disponível após regularizar a fatura')

    fpPlanoAtivarTab('faturas')
    $('#fatura-filtro').val('aberta')

    document.body.classList.add('fp-lic-presa-fatura')
    if (typeof fpLicAplicarModoSuspenso === 'function') fpLicAplicarModoSuspenso()

    if (!window.__fpPlanoPollSuspenso) {
        window.__fpPlanoPollSuspenso = setInterval(function () {
            $.ajax({
                url: '/licencaResumo',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ produto: 'franqueadopro' }),
                cache: false
            }).done(function (resp) {
                if (typeof fpLicProcessarResposta === 'function') {
                    fpLicProcessarResposta(resp)
                } else if (resp.liberado) {
                    clearInterval(window.__fpPlanoPollSuspenso)
                    window.__fpPlanoPollSuspenso = null
                    window.location.href = '/carregar-menu-principal'
                }
            })
        }, 15000)
    }
}

function carregarResumo() {
    $.when(
        $.ajax({
            url: '/licencaResumo',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: 'franqueadopro' }),
            timeout: 30000
        }),
        $.ajax({
            url: '/licencaCatalogoPlanos',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: 'franqueadopro' }),
            timeout: 30000
        })
    ).done(function (resumoResp, catalogoResp) {
        try {
            var r = (resumoResp && resumoResp[0]) || {}
            var catBody = (catalogoResp && catalogoResp[0]) || {}
            var catalogoLista = Array.isArray(catBody.dados) ? catBody.dados : []
            aplicarResumo(r, catalogoLista)
        } catch (e) {
            console.error('aplicarResumo', e)
            $('#plano-status').html('<p class="text-danger">Erro ao exibir o plano. Recarregue a página.</p>')
        }
        carregarEcossistema()
    }).fail(function () {
        $('#plano-status').html('<p class="text-danger">Não foi possível carregar o plano.</p>')
        if (typeof fpRenderComparacaoPlanos === 'function') {
            fpRenderComparacaoPlanos('#plano-comparacao', '', { podeContratar: true })
        }
        carregarEcossistema()
    })
}

function fpEcoNomeProduto(pid) {
    var map = {
        webterminal: 'WebTerminal',
        terminalmovel: 'Terminal Móvel',
        confvision: 'Vision',
        webambiente: 'webAmbiente',
        dialyze: 'Dialyze'
    }
    return map[pid] || pid
}

function carregarEcossistema() {
    var $box = $('#fp-ecossistema-lista')
    if (!$box.length) return
    $box.html('<p class="text-muted small">Carregando…</p>')
    $.ajax({
        url: '/licencaEcossistema',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        var dados = (r && r.dados) || []
        if (!dados.length) {
            $box.html('<p class="text-muted small">Nenhum produto no catálogo.</p>')
            return
        }
        var html = '<div class="table-responsive"><table class="table table-sm align-middle mb-0"><thead><tr>' +
            '<th>Produto</th><th>Status</th><th style="min-width:280px">Plano / Valor</th></tr></thead><tbody>'
        dados.forEach(function (p) {
            var status = p.status || 'disponivel'
            var badge = status === 'incluso' ? 'success' : (status === 'ativo' ? 'primary' : 'secondary')
            var statusLabel = status === 'incluso' ? 'Incluso no FP' : (status === 'ativo' ? 'Ativo (avulso)' : 'Disponível')
            var planos = p.planos || []
            var acao = ''
            if (status === 'incluso' && !p.pode_contratar) {
                acao = '<span class="text-muted small">Já incluso — não precisa comprar</span>'
            } else if (p.pode_contratar && planos.length) {
                var opt = planos.map(function (x) {
                    return '<option value="' + x.plano + '">' + (x.nome_exibicao || x.plano) +
                        ' — R$ ' + Number(x.valor_mensal || 0).toFixed(2) + '</option>'
                }).join('')
                // Se incluso com upgrade (ex. WT pro incluso → so mostrar tiers acima)
                if (status === 'incluso' && p.plano_incluso_fp) {
                    var rank = { lite: 1, padrao: 1, pro: 2, pro_plus: 3 }
                    var minRank = rank[p.plano_incluso_fp] || 0
                    opt = planos.filter(function (x) {
                        return (rank[x.plano] || 0) > minRank
                    }).map(function (x) {
                        return '<option value="' + x.plano + '">' + (x.nome_exibicao || x.plano) +
                            ' — R$ ' + Number(x.valor_mensal || 0).toFixed(2) + '</option>'
                    }).join('')
                }
                if (opt) {
                    acao = '<div class="d-flex gap-2 align-items-center flex-wrap">' +
                        '<select class="form-select form-select-sm fp-eco-plano flex-grow-1" data-produto="' + p.produto + '" style="min-width:260px">' +
                        opt + '</select>' +
                        '<button type="button" class="btn btn-sm btn-success fp-eco-contratar" data-produto="' + p.produto + '">Contratar</button>' +
                        '</div>'
                } else {
                    acao = '<span class="text-muted small">Já incluso — não precisa comprar</span>'
                }
            } else if (status === 'ativo') {
                var planoAtivo = p.plano_efetivo || p.plano || '—'
                acao = '<span class="text-muted small">Assinatura ativa' +
                    (planoAtivo && planoAtivo !== '—' ? ' (' + planoAtivo + ')' : '') + '</span>'
            }
            if (p.produto === 'confvision') {
                acao += '<div class="text-muted small mt-1">Licenças por câmera são compradas à parte no Vision.</div>'
            }
            html += '<tr>' +
                '<td><strong>' + fpEcoNomeProduto(p.produto) + '</strong></td>' +
                '<td><span class="badge text-bg-' + badge + '">' + statusLabel + '</span></td>' +
                '<td>' + acao + '</td></tr>'
        })
        html += '</tbody></table></div>'
        $box.html(html)
    }).fail(function () {
        $box.html('<p class="text-danger small">Não foi possível carregar o ecossistema.</p>')
    })
}

$(document).on('click', '.fp-eco-contratar', function () {
    var produto = $(this).data('produto')
    var $sel = $('.fp-eco-plano[data-produto="' + produto + '"]')
    var plano = $sel.val()
    if (!produto || !plano) return
    var $btn = $(this)
    fpConfirmar(
        'Contratar ' + fpEcoNomeProduto(produto) + ' (' + plano + ')?\nSerá gerada uma fatura.',
        { title: 'Contratar', confirmText: 'Contratar' }
    ).then(function (ok) {
        if (!ok) return
        $btn.prop('disabled', true)
        $.ajax({
            url: '/licencaContratarProduto',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ produto: produto, plano: plano })
        }).done(function () {
            fpMsgSucesso('Contratação iniciada. Veja a aba Faturas.')
            carregarEcossistema()
            fpPlanoAtivarTab('faturas')
        }).fail(function (xhr) {
            fpMsgErro(xhr)
        }).always(function () {
            $btn.prop('disabled', false)
        })
    })
})

function fpLimitesDeCotaQtd(qtd) {
    var n = Number(qtd || 0)
    if (!(n > 0)) return null
    // Mesma matemática do adm Pacotes de Cotas
    return {
        clientes_max: n,
        contas_max: n * 2,
        usuarios_alarme_max: n * 4,
        setores_alarme_max: n * 10,
        cota_quantidade: n
    }
}

function fpPlanosFromCatalogo(itens) {
    if (!itens || !itens.length) return null
    var ordem = ['lite', 'pro', 'pro_plus']
    var mapa = {}
    itens.forEach(function (c) {
        if (c.plano !== 'lite' && c.plano !== 'pro' && c.plano !== 'pro_plus') return
        var limCota = fpLimitesDeCotaQtd(c.pacote_cota_quantidade)
        var limCat = c.limites_json && typeof c.limites_json === 'object' ? c.limites_json : null
        var limTemValor = limCat && (
            limCat.clientes_max != null || limCat.contas_max != null ||
            limCat.usuarios_alarme_max != null || limCat.setores_alarme_max != null
        )
        var limPadrao = (typeof FP_PLANOS_PADRAO !== 'undefined'
            ? (FP_PLANOS_PADRAO.find(function (p) { return p.plano === c.plano }) || {}).limites_json
            : null) || {}
        var valorLic = Number(c.valor_mensal || 0)
        var valorTotal = c.valor_total_mensal != null
            ? Number(c.valor_total_mensal)
            : (valorLic + Number(c.valor_cota || 0))
        mapa[c.plano] = {
            plano: c.plano,
            nome_exibicao: c.nome_exibicao || (typeof fpPlanoNomeExibicao === 'function' ? fpPlanoNomeExibicao(c.plano) : c.plano),
            valor_mensal: valorTotal > 0 ? valorTotal : valorLic,
            retencao_dias: c.retencao_dias,
            limites_json: limCota || (limTemValor ? limCat : limPadrao),
            modulos_json: c.modulos_json,
            pacote_cota_nome: c.pacote_cota_nome || '',
            pacote_cota_quantidade: c.pacote_cota_quantidade || null,
            fp_pacote_cota_id: c.fp_pacote_cota_id || null
        }
    })
    return ordem.map(function (p) { return mapa[p] }).filter(Boolean)
}

function aplicarResumo(r, catalogoLista) {
    window.__fpCatalogoPlanos = Array.isArray(catalogoLista) ? catalogoLista : []
    r = r || {}
    var ass = r.assinatura || {}
    var liberado = r.liberado
    var motivo = r.motivo || ''
    var planoRaw = ass.plano || ''
    var plano = planoRaw ? planoRaw.toUpperCase().replace('_', '+') : '—'
    if (ass.tipo_contratacao === 'alacarte') {
        plano = plano + ' (personalizado)'
    }
    var valido = ass.valido_ate ? moment(ass.valido_ate).format('DD/MM/YYYY') : '—'
    var prox = ass.proxima_cobranca_em ? moment(ass.proxima_cobranca_em).format('DD/MM/YYYY') : '—'
    var valor = ass.valor != null ? ('R$ ' + Number(ass.valor).toFixed(2) + '/mês') : '—'
    var badge = liberado ? 'ok' : 'err'
    var statusTxt = liberado ? 'Ativo' : ('Bloqueado: ' + motivo)

    var lim = r.limites_json || ass.limites_json || {}
    var retencao = r.retencao_dias || ass.retencao_dias || '—'
    var cotas = Array.isArray(ass.cotas_json) ? ass.cotas_json : []
    var cotasTxt = '—'
    if (cotas.length) {
        cotasTxt = cotas.map(function (c) {
            return (c.nome || 'Cota') + ' (' + (c.quantidade || 0) + ')'
        }).join(' + ')
    }
    var cotaQtd = lim.cota_quantidade != null
        ? lim.cota_quantidade
        : (cotas.reduce(function (s, c) {
            return s + Number(c.quantidade || 0)
        }, 0) || null)
        var limitesHtml =
            '<div class="mt-3"><strong>Limites da sua cota</strong>' +
            (cotaQtd != null ? ' <span class="text-muted">(lote total: ' + cotaQtd + ')</span>' : '') +
            '</div>' +
            '<div class="row mt-2 text-muted small">' +
            '<div class="col-6 col-md-3">Clientes: até ' + fmtLimite(lim.clientes_max) + '</div>' +
            '<div class="col-6 col-md-3">Dispositivos: até ' + fmtLimite(lim.contas_max) + '</div>' +
            '<div class="col-6 col-md-3">Usu. alarme: até ' + fmtLimite(lim.usuarios_alarme_max) + '</div>' +
            '<div class="col-6 col-md-3">Setores: até ' + fmtLimite(lim.setores_alarme_max) + '</div>' +
            '</div>' +
            '<div class="text-muted small mt-2">Pacotes: <strong>' + cotasTxt + '</strong></div>' +
            '<div class="text-muted small mt-1">Retenção de histórico após vencimento: <strong>' + retencao + ' dias</strong></div>'

        var btnCota = ''
        if (ass && ass.id && (r.faturas_abertas || []).length === 0) {
            btnCota = '<button type="button" class="btn btn-sm btn-success flex-shrink-0" id="fp-plano-comprar-cota-btn">' +
                '<i class="bi bi-plus-circle"></i> Comprar mais cota</button>'
        }

        $('#plano-status').html(
            '<h4>Meu Plano — FranqueadoPro</h4>' +
            '<p><span class="fp-plano-badge ' + badge + '">' + statusTxt + '</span></p>' +
            '<div class="row mt-3">' +
            '<div class="col-md-3"><strong>Plano</strong><br>' + plano + '</div>' +
            '<div class="col-md-3"><strong>Valor contratado</strong><br>' + valor + '</div>' +
            '<div class="col-md-3"><strong>Válido até</strong><br>' + valido + '</div>' +
            '<div class="col-md-3"><strong>Próxima cobrança</strong><br>' + prox + '</div>' +
            '</div>' +
            limitesHtml +
            '<div class="d-flex flex-wrap align-items-center gap-2 mt-3">' +
            '<p class="text-muted mb-0 flex-grow-1">Estourou a cota? Compre outro pacote em lote ou faça upgrade de plano.</p>' +
            btnCota +
            '<button type="button" class="btn btn-sm btn-outline-primary flex-shrink-0" id="fp-plano-detalhes-btn">' +
            '<i class="bi bi-list-check"></i> Detalhes</button>' +
            '</div>'
        )

        $('#fp-plano-comprar-cota-btn').off('click').on('click', fpPlanoComprarCotaExtra)

        window.__fpPlanoModulos = r.modulos_json || ass.modulos_json || {}
        window.__fpPlanoRaw = planoRaw

        if (typeof fpRenderComparacaoPlanos === 'function') {
            var faturasAbertas = r.faturas_abertas || []
            var planosCatalogo = fpPlanosFromCatalogo(catalogoLista)
            var podeContratar = !liberado && faturasAbertas.length === 0
            fpRenderComparacaoPlanos('#plano-comparacao', planoRaw, {
                podeContratar: podeContratar,
                planosCatalogo: planosCatalogo
            })
            if (podeContratar) {
                $('#fp-plano-cupom-box').removeClass('d-none')
            } else {
                $('#fp-plano-cupom-box').addClass('d-none')
            }
        }

        if (typeof fpLicProcessarResposta === 'function') {
            if (fpLicProcessarResposta(r)) return
        } else if (typeof fpLicSalvarEstado === 'function') {
            fpLicSalvarEstado(r)
            if (typeof fpAplicarLicencaUI === 'function') fpAplicarLicencaUI()
        }

        fpPlanoAplicarModoSuspenso(r)
}

function fpPlanoGetModalModulos() {
    var modalEl = document.getElementById('fp-plano-modulos-modal')
    if (!modalEl || typeof bootstrap === 'undefined' || !bootstrap.Modal) return null
    if (modalEl.parentElement !== document.body) {
        document.body.appendChild(modalEl)
    }
    return bootstrap.Modal.getOrCreateInstance(modalEl, {
        backdrop: true,
        keyboard: true,
        focus: true
    })
}

function fpPlanoAbrirModalModulos() {
    if (typeof fpRenderModulosPlanoAtual !== 'function') return
    fpRenderModulosPlanoAtual('#fp-plano-modulos-corpo', window.__fpPlanoModulos || {}, window.__fpPlanoRaw || '', { noTitulo: true })
    var modal = fpPlanoGetModalModulos()
    if (modal) modal.show()
}

function fpPlanoCarregarPacotesCota(next) {
    $.ajax({
        url: '/licencaPacotesCota',
        method: 'POST',
        contentType: 'application/json',
        data: '{}'
    }).done(function (r) {
        window.__fpPacotesCota = (r && r.dados) || []
        if (typeof next === 'function') next(window.__fpPacotesCota)
    }).fail(function () {
        window.__fpPacotesCota = []
        if (typeof next === 'function') next([])
    })
}

function fpPlanoEscolherCota(titulo, next) {
    fpPlanoCarregarPacotesCota(function (lista) {
        if (!lista.length) {
            fpMsgAviso('Nenhum Pacote de Cotas disponível. Peça à Central cadastrar os pacotes (Cota 50/200/800).')
            return
        }
        var opcoes = lista.map(function (c) {
            return {
                value: c,
                label: (c.nome || 'Cota') +
                    ' — ' + (c.quantidade || 0) + ' clientes' +
                    ' / ' + (c.dispositivos || (c.quantidade || 0) * 2) + ' dispositivos' +
                    ' — R$ ' + Number(c.valor_venda || c.valor || 0).toFixed(2) + '/mês'
            }
        })
        fpEscolherOpcao(titulo || 'Escolha o Pacote de Cotas', opcoes, next)
    })
}

function fpPlanoInfoCatalogo(plano) {
    var lista = window.__fpCatalogoPlanos || []
    for (var i = 0; i < lista.length; i++) {
        if (String(lista[i].plano || '') === String(plano || '')) return lista[i]
    }
    return null
}

function fpPlanoContratar(plano) {
    var nome = typeof fpPlanoNomeExibicao === 'function' ? fpPlanoNomeExibicao(plano) : plano
    var cupom = fpPlanoCupomCodigoAtual()
    var cat = fpPlanoInfoCatalogo(plano)
    var cotaNome = (cat && cat.pacote_cota_nome) || 'cota vinculada'
    var cotaQtd = cat && cat.pacote_cota_quantidade != null ? cat.pacote_cota_quantidade : ''
    var total = cat && cat.valor_total_mensal != null
        ? Number(cat.valor_total_mensal)
        : null
    var msg = 'Contratar ' + nome + ' + ' + cotaNome +
        (cotaQtd !== '' ? ' (' + cotaQtd + ' clientes)' : '') + '?'
    if (total != null) {
        msg += '\nTotal mensal: R$ ' + total.toFixed(2) + ' (licença + cota do plano).'
    } else {
        msg += '\nTotal = licença do plano + pacote de cota vinculado pela Central.'
    }
    msg += '\nSerá gerada fatura em aberto.'
    if (cupom) msg += '\n\nCupom: ' + cupom
    fpConfirmar(msg, { title: 'Contratar plano', confirmText: 'Contratar' }).then(function (ok) {
        if (!ok) return
        var payload = {
            produto: 'franqueadopro',
            plano: plano
        }
        if (cat && cat.fp_pacote_cota_id) payload.fp_pacote_cota_id = cat.fp_pacote_cota_id
        if (cupom) payload.cupom_codigo = cupom
        $.ajax({
            url: '/licencaContratar',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(payload)
        }).done(function (resp) {
            var extra = ''
            if (resp && resp.cupom && resp.cupom.valor_desconto != null) {
                extra = '\nDesconto do cupom: R$ ' + Number(resp.cupom.valor_desconto).toFixed(2)
            }
            fpMsgSucesso('Assinatura e fatura geradas com sucesso. Acesse a aba Faturas e Cobrança.' + extra)
            window.__fpCupomCodigo = ''
            window.__fpCupomPreview = null
            $('#fp-plano-cupom-codigo').val('')
            $('#fp-plano-cupom-msg').text('')
            fpPlanoAtivarTab('faturas')
            carregarResumo()
            carregarFaturas()
        }).fail(function (xhr) {
            fpMsgErro(xhr)
        })
    })
}

function fpPlanoComprarCotaExtra() {
    fpPlanoEscolherCota('Comprar Pacote de Cotas extra (soma capacidade)', function (cota) {
        fpConfirmar(
            'Adicionar ' + (cota.nome || 'Cota') + ' (+' + (cota.quantidade || 0) + ' clientes)?\n' +
            'Valor: R$ ' + Number(cota.valor_venda || cota.valor || 0).toFixed(2) + '\n' +
            'Será gerada fatura e a capacidade será somada à sua cota atual.',
            { title: 'Cota extra', confirmText: 'Comprar' }
        ).then(function (ok) {
            if (!ok) return
            $.ajax({
                url: '/licencaComprarCota',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({
                    produto: 'franqueadopro',
                    fp_pacote_cota_id: cota.id
                })
            }).done(function () {
                fpMsgSucesso('Cota adicionada. Veja a fatura na aba Faturas.')
                fpPlanoAtivarTab('faturas')
                carregarResumo()
                carregarFaturas()
            }).fail(function (xhr) {
                fpMsgErro(xhr)
            })
        })
    })
}

function fmtLimite(v) {
    if (v == null || v === '') return '—'
    return Number(v) >= 9999 ? 'Ilimitado*' : String(v)
}

function carregarFaturas() {
    var status = $('#fatura-filtro').val()
    var payload = {}
    if (status) payload.status = status

    $('#plano-faturas').html('<p class="text-muted">Carregando...</p>')
    $.ajax({
        url: '/licencaMinhasFaturas',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).done(function (r) {
        var lista = (r.dados || [])
        window.__fpFaturasLista = lista
        if (!lista.length) {
            $('#plano-faturas').html(
                '<p class="text-muted mb-0">Nenhuma fatura encontrada.</p>' +
                '<p class="text-muted small mt-2 mb-0">Se você já efetuou o pagamento, aguarde a confirmação pela sua central.</p>'
            )
            return
        }
        var html = '<table class="table table-sm table-hover"><thead><tr>' +
            '<th>Referência</th><th>Tipo</th><th>Vencimento</th><th>Valor</th><th>Status</th><th></th>' +
            '</tr></thead><tbody>'
        lista.forEach(function (f, idx) {
            var st = (f.status || '').toLowerCase()
            var cls = st === 'aberta' ? 'text-warning' : (st === 'paga' ? 'text-success' : '')
            var acao = ''
            if (st === 'aberta') {
                acao = '<button type="button" class="btn btn-sm btn-primary fp-fatura-pagar-btn" data-fatura-idx="' + idx + '">' +
                    '<i class="bi bi-credit-card"></i> Pagar</button>'
            }
            html += '<tr><td>' + (f.referencia || f.id) + '</td>' +
                '<td>' + (f.tipo || 'assinatura') + '</td>' +
                '<td>' + (f.vencimento_em ? moment(f.vencimento_em).format('DD/MM/YYYY') : '—') + '</td>' +
                '<td>R$ ' + Number(f.valor_total || 0).toFixed(2) + '</td>' +
                '<td class="' + cls + '">' + (f.status || '') + '</td>' +
                '<td class="text-end text-nowrap">' + acao + '</td></tr>'
        })
        html += '</tbody></table>'
        $('#plano-faturas').html(html)
    }).fail(function (xhr) {
        var msg = 'Erro ao carregar faturas.'
        try {
            var j = JSON.parse(xhr.responseText || '')
            if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
            else if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#plano-faturas').html('<p class="text-danger">' + msg + '</p>')
    })
}

function fpPlanoGetModalPagar() {
    var modalEl = document.getElementById('fp-fatura-pagar-modal')
    if (!modalEl || typeof bootstrap === 'undefined' || !bootstrap.Modal) return null
    if (modalEl.parentElement !== document.body) {
        document.body.appendChild(modalEl)
    }
    return bootstrap.Modal.getOrCreateInstance(modalEl, {
        backdrop: true,
        keyboard: true,
        focus: true
    })
}

function fpPlanoAbrirModalPagar(f) {
    if (!f) return
    var ref = f.referencia || f.id || '—'
    var venc = f.vencimento_em ? moment(f.vencimento_em).format('DD/MM/YYYY') : '—'
    var valor = 'R$ ' + Number(f.valor_total || 0).toFixed(2)
    var obs = (f.observacao || '').trim()
    var corpo =
        '<p class="mb-2"><strong>Referência:</strong> ' + ref + '</p>' +
        '<p class="mb-2"><strong>Vencimento:</strong> ' + venc + '</p>' +
        '<p class="mb-3"><strong>Valor:</strong> <span class="fs-5">' + valor + '</span></p>'
    if (obs) {
        corpo += '<div class="alert alert-info mb-3"><strong>Instruções de pagamento</strong><br>' +
            obs.replace(/\n/g, '<br>') + '</div>'
    } else {
        corpo += '<div class="alert alert-warning mb-3">' +
            'Entre em contato com a <strong>sua central</strong> para obter PIX, boleto ou demais formas de pagamento desta fatura.' +
            '</div>'
    }
    corpo += '<p class="text-muted small mb-0">Após a confirmação do pagamento pela central, o sistema será liberado automaticamente.</p>'
    $('#fp-fatura-pagar-corpo').html(corpo)
    var modal = fpPlanoGetModalPagar()
    if (modal) modal.show()
}

function fpPlanoFaturaElegivelCupom(f) {
    if (!f) return false
    var st = String(f.status || '').toLowerCase()
    var tipo = String(f.tipo || '').toLowerCase()
    return (st === 'aberta' || st === 'paga') && tipo !== 'confvision'
}

function fpPlanoAtualizarInfoCupomFatura() {
    var idx = Number($('#fp-fatura-cupom-fatura').val())
    var f = window.__fpFaturasCupomOpts[idx]
    window.__fpFaturaCupomAtual = f || null
    if (!f) {
        $('#fp-fatura-cupom-info').text('')
        return
    }
    var st = String(f.status || '').toLowerCase()
    var info = 'Fatura ' + (f.referencia || f.id) + ' — R$ ' + Number(f.valor_total || 0).toFixed(2) + '.'
    if (st === 'paga') {
        info += ' Como a fatura já está paga, o desconto vira crédito na próxima cobrança.'
    } else {
        info += ' O valor da fatura em aberto será reduzido.'
    }
    $('#fp-fatura-cupom-info').text(info)
}

function fpPlanoAbrirModalCupomToolbar() {
    $('#fp-fatura-cupom-codigo').val('')
    $('#fp-fatura-cupom-msg').removeClass('text-danger text-success').addClass('text-muted').text('Carregando faturas…')
    $('#fp-fatura-cupom-fatura').html('')
    $('#fp-fatura-cupom-info').text('')
    window.__fpFaturaCupomAtual = null

    var modalEl = document.getElementById('fp-fatura-cupom-modal')
    if (!modalEl || typeof bootstrap === 'undefined' || !bootstrap.Modal) return
    if (modalEl.parentElement !== document.body) document.body.appendChild(modalEl)
    bootstrap.Modal.getOrCreateInstance(modalEl).show()

    $.ajax({
        url: '/licencaMinhasFaturas',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({})
    }).done(function (r) {
        var lista = (r.dados || []).filter(fpPlanoFaturaElegivelCupom)
        window.__fpFaturasCupomOpts = lista
        if (!lista.length) {
            $('#fp-fatura-cupom-msg').removeClass('text-muted text-success').addClass('text-danger')
                .text('Nenhuma fatura elegível para cupom (abertas ou pagas, exceto Vision).')
            return
        }
        var opts = lista.map(function (f, idx) {
            var st = String(f.status || '').toLowerCase()
            var label = (f.referencia || f.id) +
                ' — R$ ' + Number(f.valor_total || 0).toFixed(2) +
                ' (' + (st === 'aberta' ? 'em aberto' : (st === 'paga' ? 'paga' : st)) + ')'
            return '<option value="' + idx + '">' + label + '</option>'
        }).join('')
        $('#fp-fatura-cupom-fatura').html(opts)
        $('#fp-fatura-cupom-msg').removeClass('text-danger text-success').addClass('text-muted').text('')
        fpPlanoAtualizarInfoCupomFatura()
    }).fail(function (xhr) {
        var msg = 'Erro ao carregar faturas.'
        try {
            var j = JSON.parse(xhr.responseText || '')
            if (j.status) msg = String(j.status).replace(/^Erro:\s*/i, '')
            else if (j.message) msg = j.message
        } catch (e) { /* ignore */ }
        $('#fp-fatura-cupom-msg').removeClass('text-muted text-success').addClass('text-danger').text(msg)
    })
}

function fpPlanoAplicarCupomFatura() {
    fpPlanoAtualizarInfoCupomFatura()
    var f = window.__fpFaturaCupomAtual
    if (!f) {
        $('#fp-fatura-cupom-msg').removeClass('text-success').addClass('text-danger').text('Selecione uma fatura.')
        return
    }
    var codigo = String($('#fp-fatura-cupom-codigo').val() || '').trim().toUpperCase()
    if (!codigo) {
        $('#fp-fatura-cupom-msg').removeClass('text-success').addClass('text-danger').text('Informe o código do cupom.')
        return
    }
    $('#fp-fatura-cupom-aplicar').prop('disabled', true)
    $('#fp-fatura-cupom-msg').removeClass('text-danger text-success').addClass('text-muted').text('Aplicando…')
    $.ajax({
        url: '/licencaCupomAplicarFatura',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ codigo: codigo, fatura_id: f.id })
    }).done(function (resp) {
        var msg = 'Cupom aplicado.'
        if (resp && resp.modo === 'credito_proxima_fatura') {
            msg = 'Cupom aplicado. Crédito de R$ ' + Number(resp.valor_desconto || 0).toFixed(2) + ' na próxima fatura.'
        } else if (resp && resp.valor_final != null) {
            msg = 'Cupom aplicado. Novo valor da fatura: R$ ' + Number(resp.valor_final).toFixed(2)
        }
        fpMsgSucesso(msg)
        var modalEl = document.getElementById('fp-fatura-cupom-modal')
        if (modalEl && bootstrap && bootstrap.Modal) {
            bootstrap.Modal.getOrCreateInstance(modalEl).hide()
        }
        carregarFaturas()
        carregarResumo()
    }).fail(function (xhr) {
        var msg = fpExtrairMsgErro(xhr)
        $('#fp-fatura-cupom-msg').removeClass('text-muted text-success').addClass('text-danger').text(msg)
        fpMsgErro(msg)
    }).always(function () {
        $('#fp-fatura-cupom-aplicar').prop('disabled', false)
    })
}

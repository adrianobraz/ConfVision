function proAtendimento_start() {

    $('#boxEsq').empty()
    const uri = '/assets/modulos/proAtendimento/proAtendimento.html'
    $('#boxEsq').load(uri, () => {

        // Configuração inicial
        proAtendimento_setup(() => {
            // Verifica se o usuario esta vinculado a central
            if (sessionStorage.getItem('login_userVinculo') == "CENTRAL") {
                // Acerta o filtro para mostrar todos os processos
                sessionStorage.setItem(
                    'proAtendimento_idFranqueado',
                    'TODOS'
                )
            } else {
                // Acerta o filtro para mostrar apenas os processos do franqueado
                sessionStorage.setItem(
                    'proAtendimento_idFranqueado',
                    sessionStorage.getItem('login_userVinculo')
                )
            }
            // Carrega a tabela de processos
            proAtendimento_carregarTabela()

            $('#proAtendimento_modalManutencao_btnGravar').on('click', proAtendimento_modalManutencao_btnGravar)
        })

    })
}

function proAtendimento_setup(next) {
    // Inicializa o setup caso a session no exista
    if (sessionStorage.getItem('proAtendimento_idFranqueado') === null) {

        // Controla o filtro de processo por franqueado TODOS/idFranqueado/off
        sessionStorage.setItem('proAtendimento_idFranqueado', 'off')

        sessionStorage.setItem('proAtendimento_modalManutencao', '')

    }
    next()
}

function proAtendimento_filtraFranqueado(idFranqueado) {
    // Acerta o filtro pra mostrar processos de um franqueado especifico
    sessionStorage.setItem('proAtendimento_idFranqueado', idFranqueado)

    proAtendimento_carregarTabela()
}

// Observação temporária por processo (até 10 caracteres, localStorage)
var proAtendimento_obsStorageKey = 'proAtendimento_obs'

function proAtendimento_obsKey(idProcesso) {
    return proAtendimento_obsStorageKey + '_' + (idProcesso || '')
}

function proAtendimento_obsGet(key) {
    try { return localStorage.getItem(key) || '' } catch (e) { return '' }
}

function proAtendimento_obsSet(key, value) {
    try { localStorage.setItem(key, (value || '').slice(0, 10)) } catch (e) { }
}

function proAtendimento_carregarTabela() {
    // Clia um objto audio
    const audio = document.querySelector('audio')

    // Pega o filtro pra filtrar os processos
    const idFranqueado = sessionStorage.getItem('proAtendimento_idFranqueado')

    if (idFranqueado != 'off') {

        $.ajax({
            url: 'proAtendimento/carregarTabela',
            method: 'Post',
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
                "Authorization": "Bearer " + sessionStorage.getItem('token')
            },
            data: JSON.stringify({ idFranqueado: idFranqueado })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            // Limpa a tabela
            $('#proAtendimento tbody').empty()
            if (r.status != 'Vazio') {
                // Verifica se o audio esta ligado
                if (sessionStorage.getItem('proControles_audio') == 'ON') audio.play()


                // Carrega quantidade de processo no painel
                proControles_qtdProcessos(r.dados.length)

                r.dados.forEach(item => {
                    proAtendimento_montaLinha(item)
                });

                $('td[tipo=proAtendimento_btnAtender]').on('click', proAtendimento_btnAtender)
                $('td[tipo=proAtendimento_btnVisualizar]').on('click', proAtendimento_btnVisualizar)
                $('td[tipo=proAtendimento_btnManutencao]').on('click', proAtendimento_btnManutencao)
                $('td[tipo=proAtendimento_btnDadosCliente]').on('click', proAtendimento_btnDadosCliente)
            } else {
                audio.pause()
                $('#proAtendimento tbody').append(`
                    <tr>
                        <td colspan="8" class="proAtendimento-semProcesso"> 
                            Sem processo para atendimento 
                        </td>
                    </tr>
                `)
                // Carrega quantidade de processo no painel
                proControles_qtdProcessos("0")
            }

        })
    }
}

function proAtendimento_abrirModalUltimos() {
    const modalEl = document.getElementById('proAtendimento_modalUltimos')
    if (!modalEl) return

    const modal = bootstrap.Modal.getOrCreateInstance(modalEl)
    proAtendimento_carregarUltimosFinalizados(modal)
    modal.show()
}

function proAtendimento_carregarUltimosFinalizados(modalInstance) {
    const idFranqueado = sessionStorage.getItem('proAtendimento_idFranqueado') || 'TODOS'

    $.ajax({
        url: 'proAtendimento/ultimosFinalizados',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idFranqueado: idFranqueado })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        $('#proAtendimento_ultimos_lista_modal').empty()

        if (r.status != 'Vazio') {
            r.dados.forEach(item => {
                proAtendimento_montaLinhaHistorico(item)
            })

            $('div[tipo=proAtendimento_btnVisualizarHistorico]').off('click').on('click', function () {
                const idProcesso = $(this).attr('idProcesso')
                if (!idProcesso) return

                try {
                    sessionStorage.setItem('ateProDado_modoVisualizar', 'S')
                } catch (e) { }

                if (typeof setupTelaAtendimento === 'function') {
                    setupTelaAtendimento(idProcesso)
                }

                try {
                    if (modalInstance && typeof modalInstance.hide === 'function') {
                        modalInstance.hide()
                    } else {
                        const modalEl = document.getElementById('proAtendimento_modalUltimos')
                        if (modalEl) {
                            const m = bootstrap.Modal.getOrCreateInstance(modalEl)
                            m.hide()
                        }
                    }
                } catch (e) { }
            })
        } else {
            $('#proAtendimento_ultimos_lista_modal').append(`
                <div class="text-center text-muted">
                    Sem atendimentos finalizados recentes
                </div>
            `)
        }
    })
}

function proAtendimento_montaLinha(item) {
    const iniciado = (item.iniciado == "N") ? '' : 'bg-warning'
    const panico = (item.nivel == 6) ? 'bg-danger' : ''
    const obsKey = proAtendimento_obsKey(item.idProcesso)
    const obsVal = proAtendimento_obsGet(obsKey)
    const obsEsc = (obsVal || '').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

    $('#proAtendimento tbody').append(`
        <tr class="${corNivel(item.nivel)}">

<td tipo="proAtendimento_btnAtender" 
    idProcesso="${item.idProcesso}"                
    style="width: 5%;" 
    class="click bg-success">                        
    <i class="bi bi-headset"></i>
</td>

<td tipo="proAtendimento_btnVisualizar" 
    idProcesso="${item.idProcesso}"
    style="width: 5%;" 
    class="click bg-primary">
    <i class="bi bi-eye-fill"></i>
</td>


            <td tipo="proAtendimento_btnManutencao" 
                style="width: 5%;"
                idProcesso="${item.idProcesso}" 
                idDispositivo="${item.idDispositivo}" 
                nomeDispositivo="${item.nomeDispositivo}"
                class="click bg-warning">
                <div class="d-flex justify-content-center text-bg-warning">       
                    <i class="bi bi-tools"></i>
                </div>
            </td>
            <td tipo="proAtendimento_btnDadosCliente"
                idCliente="${item.idCliente || ''}"
                style="width: 52%; color: #000;"
                class="click"
                title="Ver dados do franqueado e cliente">
                <span class="proAtendimento-dot ${corNivelDot(item.nivel)}"></span>
                ${item.nomeCliente}
            </td>


<td style="width: 5%;" class="${panico}"><i class="bi bi-radioactive"></i></td>
<td style="width: 5%;" class="${iniciado}"><i class="bi bi-taxi-front-fill"></i></td>


            <td style="width: 11%; font-size: 10px; color: #000;" class="">${item.nomeOperador}</td>
            <td class="p-0 align-middle" style="width: 12%; min-width: 70px;">
                <input type="text" class="form-control form-control-sm proAtendimento-inputObs" 
                    maxlength="10" placeholder="ex: 2hrs" data-obs-key="${obsKey}" value="${obsEsc}" 
                    title="Observação temporária (até 10 caracteres)" />
            </td>
        </tr>

    `)
    $('#proAtendimento tbody tr:last input.proAtendimento-inputObs')
        .on('input', function () {
            var key = $(this).data('obs-key')
            if (key) proAtendimento_obsSet(key, $(this).val())
        })
        .on('click', function (e) { e.stopPropagation() })
}

function proAtendimento_montaLinhaHistorico(item) {
    const dataFim = item.dataAtenFim || item.DataAtenFim || ''
    const cliNome = item.nomeCliente || item.Nome || ''
    const operNome = item.nomeOperador || item.NomeOperador || ''

    $('#proAtendimento_ultimos_lista_modal').append(`
        <div 
            class="px-1 click d-flex justify-content-between"
            tipo="proAtendimento_btnVisualizarHistorico"
            idProcesso="${item.idProcesso}"
        >
            <span>${dataFim}</span>
            <span>${cliNome}</span>
            <span>${operNome}</span>
        </div>
    `)
}

function proAtendimento_btnAtender() {
    // Seta o filtro para desligado impedindo que atualize os dados do processo
    sessionStorage.setItem('proAtendimento_idFranqueado', 'off')

    // Modo normal de atendimento
    try {
        sessionStorage.setItem('ateProDado_modoVisualizar', 'N')
    } catch (e) { }

    // Recupera o id do processo do botao
    const idProcesso = $(this).attr('idProcesso')

    //Bloqueia o processo para não haver atendimento duplo
    proAtendimento_bloquearProcesso(idProcesso, () => {
        // Carrega a tela atendimento
        setupTelaAtendimento(idProcesso)
    })

}

function proAtendimento_btnVisualizar() {
    const idProcesso = $(this).attr('idProcesso')
    proEventoDetalhe_start(idProcesso)
}

function proAtendimento_btnDadosCliente() {
    const idCliente = $(this).attr('idCliente')
    if (idCliente) proClienteResumo_start(idCliente)
}

function proAtendimento_btnManutencao() {
    const idProcesso = $(this).attr('idProcesso')
    const idDispositivo = $(this).attr('idDispositivo')
    const nomeDispositivo = $(this).attr('nomeDispositivo')

    // Cria um modal
    gModais = new bootstrap.Modal('#proAtendimento_modalManutencao', {
        keyboard: false,
        backdrop: 'static'
    })

    // Carrega os dados no modal
    $('#proAtendimento_modalManutencao_cpNomeDisp').html('Dispositivo: ' + nomeDispositivo)

    $('#proAtendimento_modalManutencao_cpTempo')
        .attr('idDispositivo', idDispositivo)
        .attr('idProcesso', idProcesso)

    gModais.show()
}

function proAtendimento_modalManutencao_btnGravar() {

    $.ajax({
        url: '/proAtendimento/gravarManutencao',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({

            idDispositivo: $('#proAtendimento_modalManutencao_cpTempo').attr('idDispositivo'),
            tempo: $('#proAtendimento_modalManutencao_cpTempo').val(),
            idProcesso: $('#proAtendimento_modalManutencao_cpTempo').attr('idProcesso'),

            idOperador: sessionStorage.getItem('login_userIdOperador'),
            nick: sessionStorage.getItem('login_userNick')
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        msgSucesso()

        gModais.hide()


    })
}

function proAtendimento_abrirModalManutencoesAtivas() {
    const modalEl = document.getElementById('proAtendimento_modalManutencoesAtivas')
    if (!modalEl) return

    const modal = bootstrap.Modal.getOrCreateInstance(modalEl)
    proAtendimento_carregarManutencoesAtivas()
    modal.show()
}

function proAtendimento_carregarManutencoesAtivas() {
    const idFranqueado = sessionStorage.getItem('proAtendimento_idFranqueado') || 'TODOS'

    $.ajax({
        url: '/proAtendimento/listarManutencoes',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idFranqueado: idFranqueado
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const $tbody = $('#proAtendimento_modalManutencoesAtivas_tbody')
        $tbody.empty()

        if (r.status != 'Vazio' && r.dados) {
            r.dados.forEach(item => {
                proAtendimento_montaLinhaManutencaoAtiva(item)
            })

            $('button[tipo=proAtendimento_btnRemoverManutencao]').off('click').on('click', function () {
                const idAlvo = $(this).attr('idAlvo')
                const titulo = $(this).attr('titulo')
                proAtendimento_removerManutencao(idAlvo, titulo)
            })
        } else {
            $tbody.append(`
                <tr>
                    <td colspan="7" class="text-center">Sem manutenções ativas</td>
                </tr>
            `)
        }
    })
}

function proAtendimento_montaLinhaManutencaoAtiva(item) {
    const tipo = item.tipo || ''
    const cliente = item.nomeCliente || ''
    const dispositivo = item.nomeDispositivo || ''
    const setor = item.nomeSetor || '-'
    const dataBloqueio = item.dataBloqueio || ''
    const dataRetirada = item.dataRetirada || ''
    const titulo = (tipo === 'SETOR')
        ? `${dispositivo} / ${setor}`
        : `${dispositivo}`

    $('#proAtendimento_modalManutencoesAtivas_tbody').append(`
        <tr>
            <td>${tipo}</td>
            <td>${cliente}</td>
            <td>${dispositivo}</td>
            <td>${setor}</td>
            <td>${dataBloqueio}</td>
            <td>${dataRetirada}</td>
            <td>
                <button
                    type="button"
                    class="btn btn-sm btn-danger"
                    tipo="proAtendimento_btnRemoverManutencao"
                    idAlvo="${item.idAlvo}"
                    titulo="${titulo}"
                >Remover</button>
            </td>
        </tr>
    `)
}

function proAtendimento_removerManutencao(idAlvo, titulo) {
    if (!idAlvo) return

    const podeRemover = confirm(`Remover manutenção de: ${titulo}?`)
    if (!podeRemover) return

    $.ajax({
        url: '/proAtendimento/removerManutencao',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idAlvo: idAlvo
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function () {
        msgSucesso()
        proAtendimento_carregarManutencoesAtivas()
    })
}

function proAtendimento_bloquearProcesso(idProcesso, next) {
    $.ajax({
        url: '/proAtendimento/bloquearProcesso',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idProcesso: idProcesso,
            idOperador: sessionStorage.getItem('login_userIdOperador')
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.dados === "ok") {
            next()
        } else if (r.dados === sessionStorage.getItem('login_userNick')) {
            next()
        } else {
            alert(`Já esta em atendimento por: ${r.dados}`)
        }
    })
}

function corNivel(nivel) {
    var n = String(nivel || '1')
    if (n === '1') return 'table-secondary'  /* nível 1: cinza neutro */
    if (n === '2') return 'table-success'
    if (n === '3') return 'table-primary'
    if (n === '4') return 'table-warning'
    if (n === '5' || n === '6') return 'table-danger'
    return 'table-secondary'
}

function corNivelDot(nivel) {
    const cor = corNivel(nivel)
    return cor.replace('table', 'dot')
}

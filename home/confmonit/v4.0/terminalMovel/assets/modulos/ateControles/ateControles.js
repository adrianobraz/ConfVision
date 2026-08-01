function ateControles_start() {
    fetch(`/assets/modulos/ateControles/ateControles.html`).then((res) => res.text()).then((html) => {
        const box = document.getElementById('boxTop')
        box.innerHTML = html
        ateControles_setup(() => {
            ateControles_renderHeader()
            ateControles_atualizarLayout()
            $(window).off('resize.ateControlesLayout').on('resize.ateControlesLayout', ateControles_atualizarLayout)
        })
    })
}

function ateControles_setup(next){
    next()
}

function ateControles_setMenuVisivel(visivel) {
    const menuEl = document.getElementById('ateControles_boxBotoes')
    const topRowEl = document.getElementById('ateControles_topRow')
    if (menuEl) {
        menuEl.classList.toggle('d-none', !visivel)
    }
    if (topRowEl) {
        topRowEl.classList.toggle('d-none', !visivel)
    }
    ateControles_atualizarLayout()
}

function ateControles_atualizarLayout() {
    const topEl = document.getElementById('boxTop')
    if (!topEl) return

    const topHeight = topEl.offsetHeight || 100
    document.documentElement.style.setProperty('--terminalmovel-top-height', topHeight + 'px')
    document.documentElement.style.setProperty('--terminalmovel-footer-height', '0px')
}

function ateControles_abrirOrdemServico() {
    const idFranqueado = sessionStorage.getItem('ateProDado_proc_idFranqueado')
    const nomeFranqueado = sessionStorage.getItem('ateProDado_proc_franqNome') || ''
    const idCliente = sessionStorage.getItem('ateProDado_proc_idCliente')
    const nomeCliente = sessionStorage.getItem('ateProDado_proc_cliNome') || ''
    if (!idFranqueado || !idCliente) {
        if (typeof msgErro === 'function') msgErro('Processo sem empresa ou cliente vinculado.')
        else alert('Processo sem empresa ou cliente vinculado.')
        return
    }
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        ctrOrdemServico_start({
            origem: 'atendimento',
            idFranqueado: idFranqueado,
            nomeFranqueado: nomeFranqueado,
            idCliente: idCliente,
            nomeCliente: nomeCliente
        })
    })
}

function ateControles_abrirEventos() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        ateEventosDetalhe_start()
    })
}

function ateControles_voltarPrincipalAtendimento() {
    if (typeof terminalMovel_mostrarAtendimentoPrincipal === 'function') {
        terminalMovel_mostrarAtendimentoPrincipal()
    }
}

function ateControles_abrirSetores() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modSetores_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_abrirUsuarios() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modUsuarios_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_abrirGrade() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modGrade_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_abrirProcedimento() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modProcedimento_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_abrirTecnico() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modTecnico_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_abrirViatura() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modViatura_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_abrirDiscador() {
    terminalMovel_abrirTelaAtendimentoSecundaria(function () {
        modDiscador_start(ateControles_voltarPrincipalAtendimento)
    })
}

function ateControles_timer(start) {
    if (sessionStorage.getItem('gSetupTela') == 'atendimento') {

        // Em modo visualização não conta tempo nem atualiza o cronômetro
        if (sessionStorage.getItem('ateProDado_modoVisualizar') === 'S') {
            return
        }

        if (flags.ateControles_gTimer < 0) {
            flags.ateControles_gTimer = start
        }

        // Se o tempo não for zerado
        if (flags.ateControles_gTimer > 0) {
            flags.ateControles_gTimer--
            // Imprime a hora no display do terminal
            ateControles_imprimeHora(flags.ateControles_gTimer)

        } else { // Quando o contador chegar a zero faz esta ação

            Swal.fire({
                position: 'top',
                icon: 'warning',
                title: 'Atenção !!!',
                text: 'Tempo de atendimento acima do ideal, penalizações podem ser aplicadas',
                confirmButtonColor: '#242343;',
            })

            // Resta a variavel para um novo ciclo
            flags.ateControles_gTimer = start

            // Aplica penalidade no atendente e volta para tela PROCESSO
            ateControles_punir()
            // setupTelaProcesso()
            if (sessionStorage.getItem('gTempoAtendimentoFechar') == 'ON') {
                ateProDado_aguardar(
                    sessionStorage.getItem('ateProDado_proc_idProcesso')
                )
            }
        }

    }
}

function ateControles_renderHeader() {
    const modoVisualizar = sessionStorage.getItem('ateProDado_modoVisualizar') || 'N'

    if (modoVisualizar === 'S') {
        // Substitui o cronômetro por botão "Voltar para processos"
        const $timer = $('#ateControles_cpTimer')
        if ($timer.length) {
            $timer.html(`
                <button 
                    id="ateControles_btnVoltarProcesso"
                    class="w-100 h-100 btn btn-secondary border-0"
                    type="button"
                    onclick="ateControles_voltarProcesso()"
                >
                    <i class="bi bi-arrow-left-circle"></i>
                    &nbsp; Voltar
                </button>
            `)
        }
    } else {
        // Modo normal: limpa o conteúdo; o cronômetro será impresso pelo timer
        $('#ateControles_cpTimer').empty()
    }
}

function ateControles_voltarProcesso() {
    try {
        sessionStorage.setItem('ateProDado_modoVisualizar', 'N')
    } catch (e) { }
    if (typeof setupTelaProcesso === 'function') {
        setupTelaProcesso()
    }
}

function ateControles_imprimeHora(tempo) {
    // Pega a parte inteira dos minutos
    var min = parseInt(tempo / 60);

    // Calcula os segundos restantes
    var seg = tempo % 60;

    // Formata o número menor que dez, ex: 08, 07, ...
    if (min < 10) {
        min = "0" + min;
        min = min.substr(0, 2)
    }

    if (seg <= 9) seg = "0" + seg

    // Cria a variável para formatar no estilo hora/cronômetro
    $("#ateControles_cpTimer").html(min + ':' + seg);

}

function ateControles_punir() {

}


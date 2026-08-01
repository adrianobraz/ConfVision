function ateProDados_start(idProcesso, next) {
    fetch(`/assets/modulos/ateProDados/ateProDados.html`).then((res) => res.text()).then((html) => {
        const box = document.getElementById('boxEsq')
        box.innerHTML = html
         // finaliza o eveto na benuvem
         ateProDado_fecharEventosBenuvem(idProcesso)

        ateProDado_setup(idProcesso, () => {
            ateProDado_buscarDados(idProcesso, next)

        })
    })
}

function ateProDado_setup(idProcesso, next) {
    const modoVisualizar = sessionStorage.getItem('ateProDado_modoVisualizar') || 'N'

    if (modoVisualizar === 'S') {
        // Modo somente visualização: desabilita botões de ação
        $('#atenProDados_btnAguardar').prop('disabled', true).addClass('disabled')
        $('#atenProDados_btnFinalizar').prop('disabled', true).addClass('disabled')
        $('#atenProDados_btnAguardar').closest('.atenProDados_bntControle').hide()
        $('#atenProDados_btnFinalizar').closest('.atenProDados_bntControle').hide()


        // Em modo visualização, oculta completamente o bloco de botões do rodapé;
        // o botão principal de "Voltar" fica na barra superior (ateControles).
        $('#atenProDados_boxBtn').hide()
    } else {
        $('#atenProDados_boxBtn').show()
        $('#atenProDados_btnAguardar').closest('.atenProDados_bntControle').show()
        $('#atenProDados_btnFinalizar').closest('.atenProDados_bntControle').show()    
        $('#atenProDados_btnAguardar').on('click', () => { ateProDado_aguardar(idProcesso) })
        $('#atenProDados_btnFinalizar').on('click', () => { ateProDado_finalizar(idProcesso) })
    }

    // Adiciona menssagem pre definidas no campo atenProDados_cpDescricaoAtendimento 
    $('#atenProDados_cpMenssagemPreFormatada').on('change', () => {
        $('#atenProDados_cpDescricaoAtendimento').val(
            $('#atenProDados_cpMenssagemPreFormatada').val()
        ).trigger('input')
    })

    const $btnIA = $('#atenProDados_btnSugerirIA')
    const $campoDesc = $('#atenProDados_cpDescricaoAtendimento')

    // uma sugestão por processo: se já usou IA neste processo, deixa botão desabilitado
    const jaUsouIA = sessionStorage.getItem('ateProDado_iaUsado_' + idProcesso)
    if (jaUsouIA) {
        $btnIA.prop('disabled', true).html('<i class="bi bi-check"></i> Já utilizado').off('click')
    } else {
        $btnIA.prop('disabled', ($campoDesc.val() || '').trim() === '')
    }

    // habilita/desabilita conforme o usuário digita (só se ainda não usou IA)
    $campoDesc.on('input', function () {
        if (sessionStorage.getItem('ateProDado_iaUsado_' + idProcesso)) return
        const temTexto = ($(this).val() || '').trim() !== ''
        $btnIA.prop('disabled', !temTexto)
    })

    $btnIA.on('click', ateProDado_sugerirDescricaoIA)

    next()
}

function ateProDado_sugerirDescricaoIA() {
    const idProcesso = sessionStorage.getItem('ateProDado_proc_idProcesso')
    if (!idProcesso) {
        alert('Nenhum processo em atendimento.')
        return
    }
    const rascunho = $('#atenProDados_cpDescricaoAtendimento').val() || ''
    const $btn = $('#atenProDados_btnSugerirIA')
    $btn.prop('disabled', true).html('<span class="spinner-border spinner-border-sm"></span>')

    $.ajax({
        url: '/ateProDados/sugerirDescricao',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': 'Bearer ' + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idProcesso: idProcesso, rascunho: rascunho })
    }).fail(function (xhr) {
        const msg = (xhr.responseJSON && xhr.responseJSON.status) ? xhr.responseJSON.status : (xhr.responseText || 'Erro ao sugerir descrição.')
        alert(msg)
    }).done(function (r) {
        if (r.dados && r.dados.textoSugerido) {
            $('#atenProDados_cpDescricaoAtendimento').val(r.dados.textoSugerido)
            sessionStorage.setItem('ateProDado_iaUsado_' + idProcesso, '1')
        }
    }).always(function () {
        if (sessionStorage.getItem('ateProDado_iaUsado_' + idProcesso)) {
            $btn.prop('disabled', true).html('<i class="bi bi-check"></i> Já utilizado')
        } else {
            const temTexto = ($('#atenProDados_cpDescricaoAtendimento').val() || '').trim() !== ''
            $btn.prop('disabled', !temTexto).html('<i class="bi bi-stars"></i> Sugerir descrição com IA')
        }
    })
}

function ateProDado_buscarDados(idProcesso, next) {

    $.ajax({
        url: '/ateProDados/buscarDados',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idProcesso: idProcesso })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const d = r.dados

        sessionStorage.setItem('ateProDado_proc_idProcesso', d.idProcesso)
        sessionStorage.setItem('ateProDado_proc_descricao', d.descricao)

        sessionStorage.setItem('ateProDado_proc_idFranqueado', d.idFranqueado)
        sessionStorage.setItem('ateProDado_proc_franqNome', d.franqNome)
        sessionStorage.setItem('ateProDado_proc_franqCodBenuvem', d.franqCodBenuvem)

        sessionStorage.setItem('ateProDado_proc_idCliente', d.idCliente)
        sessionStorage.setItem('ateProDado_proc_cliNome', d.cliNome)

        sessionStorage.setItem('ateProDado_proc_idDispositivo', d.idDispositivo)
        sessionStorage.setItem('ateProDado_proc_dispNome', d.dispNome)
        sessionStorage.setItem('ateProDado_proc_dispConta', d.dispConta)
        sessionStorage.setItem('ateProDado_proc_dispParticao', d.dispParticao)
        sessionStorage.setItem('ateProDado_proc_dispSenha', d.dispSenha)
        sessionStorage.setItem('ateProDado_proc_dispArmado', d.dispArmado)
        sessionStorage.setItem('ateProDado_proc_dispMsgAtendente', d.dispMsgAtend)

        sessionStorage.setItem('ateProDado_proc_usaConfVision', d.usaConfVision || 'N')
        sessionStorage.setItem('ateProDado_proc_provedorVideo', d.provedorVideo || 'nenhum')
        sessionStorage.setItem('ateProDado_proc_dataCriacao', d.dataCriacao || '')

        ateProDado_montarTela(d)

        next()

    })
}

function ateProDado_montarTela(i) {
    
    if (i.dispArmado == "S") {
        $('#atenProDados_armado').html(
            `<i armado="S" class="bi bi-lock-fill bg-success 
            h-100 w-100 text-center rounded-3"></i>`
        )
    } else {
        $('#atenProDados_armado').html(
            `<i armado="N" class="bi bi-unlock-fill bg-danger 
            h-100 w-100 text-center rounded-3"></i>`
        )
    }

    $('#atenProDados_cpFranqNome').html(i.franqNome)
    $('#atenProDados_cpCliNome').val(i.cliNome)
    $('#atenProDados_cpDispNome').val(i.dispNome)
    $('#atenProDados_cpAtendimentoAnterior').val(i.descricao)
    $('#atenProDados_cpMensagemAtendente').val(i.dispMsgAtend)

    // Exibe um caixa dialogo caso tenha alguma messagem para o atendete
    if (i.dispMsgAtend != '') msgBox(i.dispMsgAtend)
}

// ateProDado_aguardar libera o processo e volta para tela processo
function ateProDado_aguardar(idProcesso) {
    const modoVisualizar = sessionStorage.getItem('ateProDado_modoVisualizar') || 'N'
    if (modoVisualizar === 'S') {
        if (typeof window.setupTelaProcesso === 'function') window.setupTelaProcesso()
        return
    }
    // Desliga o contador de atendimento
    flags.ateControles_gTimer = -1


    const descricaoAnterior = $('#atenProDados_cpAtendimentoAnterior').val()
    const tag = gGeraTag(sessionStorage.getItem('login_userNick'))
    const descricaoAtual = $('#atenProDados_cpDescricaoAtendimento').val().toUpperCase()

    let descricao
    if (descricaoAtual == '') {
        if (descricaoAnterior == '') {
            descricao = tag + '\n   VISUALIZOU O PROCESSO'
        } else {
            descricao = descricaoAnterior + '\n\n' + tag + '\n   VISUALIZOU O PROCESSO'
        }
    } else {
        if (descricaoAnterior == '') {
            descricao = tag + '\n   ' + descricaoAtual
        } else {
            descricao = descricaoAnterior + '\n\n' + tag + '\n   ' + descricaoAtual
        }
    }

    $.ajax({
        url: '/ateProDados/aguardar',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idProcesso: idProcesso,
            descricao: descricao
        })
    }).fail(function (e) {
        console.log(e)
        if (typeof window.setupTelaProcesso === 'function') window.setupTelaProcesso()
    }).done(function (r) {
        if (typeof window.setupTelaProcesso === 'function') window.setupTelaProcesso()
    })
}

function ateProDado_finalizar(idProcesso) {
    const modoVisualizar = sessionStorage.getItem('ateProDado_modoVisualizar') || 'N'
    if (modoVisualizar === 'S') {
        if (typeof window.setupTelaProcesso === 'function') window.setupTelaProcesso()
        return
    }
    // Desliga o contador de atendimento
    flags.ateControles_gTimer = -1

    if ($('#atenProDados_cpDescricaoAtendimento').val() == "") {
        msgErro("O campo descrição não pode estar vazio")
        return
    }
    const descricaoAnterior = $('#atenProDados_cpAtendimentoAnterior').val()

    const tag = gGeraTag(sessionStorage.getItem('login_userNick'))

    const descricaoAtual = $('#atenProDados_cpDescricaoAtendimento').val().toUpperCase()
    const descricao = (descricaoAtual == '') ?
        ((descricaoAnterior == '') ?
            tag + '\n   VISUALIZOU O PROCESSO' :
            descricaoAnterior + '\n\n' + tag + '\n   VISUALIZOU O PROCESSO') :
        ((descricaoAnterior == '') ?
            tag + '\n   ' + descricaoAtual :
            descricaoAnterior + '\n\n' + tag + '\n   ' + descricaoAtual)


    $.ajax({
        url: '/ateProDados/finalizar',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idProcesso: idProcesso,
            idCliente: sessionStorage.getItem('ateProDado_proc_idCliente'),
            descricao: descricao,
            idOperador: sessionStorage.getItem('login_userIdOperador'),
            nomeOperador: sessionStorage.getItem('login_userNome')
        })
    }).fail(function (e) {
        console.log(e)
        if (typeof window.setupTelaProcesso === 'function') window.setupTelaProcesso()
    }).done(function (r) {
        if (typeof msgSucesso === 'function') msgSucesso('Atendimento finalizado.')
        if (typeof window.setupTelaProcesso === 'function') window.setupTelaProcesso()
    })
}

function ateProDado_setStatusPendente(texto) {
    $('#atenProDados_armado').html(
        `<i armado="P" class="bi bi-hourglass-split bg-warning 
        h-100 w-100 text-center rounded-3" title="${texto}"></i>`
    )
}

function ateProDado_revalidarStatusDispositivo(statusEsperado, tentativa = 1, maxTentativas = 8) {
    const idProcesso = sessionStorage.getItem('ateProDado_proc_idProcesso')
    if (!idProcesso) {
        msgErro('Processo não encontrado para revalidar status')
        return
    }

    ateProDado_buscarDados(idProcesso, () => {
        const statusAtual = $('#atenProDados_armado i').attr('armado')

        if (statusAtual === statusEsperado) {
            msgSucesso('Status confirmado pela central')
            return
        }

        if (tentativa < maxTentativas) {
            setTimeout(() => {
                ateProDado_revalidarStatusDispositivo(statusEsperado, tentativa + 1, maxTentativas)
            }, 1500)
            return
        }

        msgErro('Comando enviado, mas sem confirmação de status da central')
    })
}

function ateProDado_armarCentral() {
    const idDisp = sessionStorage.getItem('ateProDado_proc_idDispositivo')
    const senha = sessionStorage.getItem('ateProDado_proc_dispSenha')
    const numero = sessionStorage.getItem('ateProDado_proc_dispParticao')    
    const armado = $('#atenProDados_armado i').attr('armado')
    if (armado == 'N') {
        ateProDado_setStatusPendente('PENDENTE CONFIRMACAO DE ARME')
        armarCentral(idDisp, senha, numero, () => {
            ateProDado_revalidarStatusDispositivo('S')
        })
    } else if (armado == 'S') {
        msgSucesso('A central já se encontra armada')
    } else if (armado == 'P') {
        msgErro('Aguardando confirmação do último comando')
    } else {
        msgErro('Status desconhecido, atualize os dados do processo')
    }
}

function ateProDado_fecharEventosBenuvem(idProcesso) {

    $.ajax({
        url: '/ateProDados/eventoProcesso',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idProcesso: idProcesso })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == 'OK') {

            let fechar = '000'
            r.dados.setor.forEach(i => {

                if (i.img != "" && i.img != "SEM IMAGEM") {
                    const d = JSON.parse(i.img)
                    if (d.channels != fechar) {
                        fechar = d.channels
                        benuvem_finalizarEvento(
                            d.company_code,
                            d.partition,
                            d.client_code,
                            d.channels,
                        )
                    }
                }
            });
        }
    })
}


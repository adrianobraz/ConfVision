function modDiscador_start(retorno) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/modDiscador/modDiscador.html'
    $('#boxDir').load(uri, () => {
        modDiscador_setup(retorno, () => {
            modDiscador_carregarDados()
        })
    })
}

function modDiscador_setup(retorno, next) {
    // Seta default pra on permitindo fazer ligações
    sessionStorage.setItem('modDiscador_ligar', 'on')

    // Adiciona mascara no campo modDiscador_cpAternativo
    $('#modDiscador_cpAternativo').on('keydown', function (e) {
        $('#modDiscador_cpAternativo').mask(gMkCel)
    })

    // Botao Fechar
    $('#modDiscador_btnFechar').on('click', retorno)

    next()
}


function modDiscador_carregarDados() {
    const idDispositivo = sessionStorage.getItem('ateProDado_proc_idDispositivo')

    $.ajax({
        url: '/modDiscador/buscarDados',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const d = r.dados
        console.log(d)
        $('#modDiscador_cpSenhaVerbal').val(d.senhaVerbal)
        $('#modDiscador_cpContraSenha').val(d.contraSenha)

        $('#modDiscador_cpTelefone1').val(formatarCelular(d.cliTelefone1))
        $('#modDiscador_cpTelefone2').val(formatarCelular(d.cliTelefone2))

        $('#modDiscador_cpNomeSupervisor').val(d.supervisor)

        $('#modDiscador_cpTelefone1Sup').val(formatarCelular(d.supTelefone1))

        $('#modDiscador_cpTelefone2Sup').val(formatarCelular(d.supTelefone2))


        //################# Configura os botoes #################\\
        $('#modDiscador_btnTelefone1').on('click', function () {
            //nvoip_realizarChamada(d.cliTelefone, (d) => { })

            modDiscador_ligar(d.cliTelefone1, '#modDiscador_btnTelefone1')
        })

        $('#modDiscador_btnTelefone2').on('click', function () {

            //nvoip_realizarChamada(d.cliCelular, (d) => { })
            modDiscador_ligar(d.cliTelefone2, '#modDiscador_btnTelefone2')

        })

        $('#modDiscador_btnTelefoneSup').on('click', function () {
            // nvoip_realizarChamada(d.supTelefone, (d) => { })
            modDiscador_ligar(d.supTelefone, '#modDiscador_btnTelefoneSup')
        })

        $('#modDiscador_btnCelularSup').on('click', function () {
            // nvoip_realizarChamada(d.supCelular, (d) => { })
            modDiscador_ligar(d.supCelular, '#modDiscador_btnCelularSup')
        })



        $('#modDiscador_btnAlternativo').on('click', function () {
            numero = gLimpaDocumento($('#modDiscador_cpAternativo').val())
            // nvoip_realizarChamada(numero, (d) => { })
            modDiscador_ligar(numero, '#modDiscador_btnAlternativo')
        })
    })
}


function modDiscador_ligar(destino, botao) {
    if (destino != '' && sessionStorage.getItem('login_voipAtivo') == 'S') {
        // Verifica se as ligações não estao bloqueadas
        if (sessionStorage.getItem('modDiscador_ligar') === 'on') {

            // Bloqueia novas ligações
            sessionStorage.setItem('modDiscador_ligar', 'off')

            // Modifica a cor do botao para vermelho
            $(botao).removeClass('btn-success').addClass('btn-danger')

            // Efetua uma ligacao           
         
            const src =  sessionStorage.getItem('login_userCelular')
            const dst = destino
            vono_ligar(
                src, 
                dst, 
                sessionStorage.getItem('ateProDado_proc_idCliente'),
                (resp) => {
                // Retorna a cor do botao para verde
                $(botao).removeClass('btn-danger').addClass('btn-success')

                // Libera novas ligações
                sessionStorage.setItem('modDiscador_ligar', 'on')

                if (!resp){
                    msgErro('Ligação falhou')
                }
            })

        } else {
            msgErro('Já existe um ligação em andamento')
        }

    }
}




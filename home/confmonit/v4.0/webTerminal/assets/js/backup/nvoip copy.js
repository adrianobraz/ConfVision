function nvoip_start() {
    sessionStorage.setItem('nvoip_idChamada', 'off')
    sessionStorage.setItem('nvoip_estadoLigacao', '')
    sessionStorage.setItem('nvoip_atendeu', 'off')
    nvoip_autenticar((d) => {
        nvoip_timer()
    })
}

function nvoip_timer() {
    if (sessionStorage.getItem('nvoip_estadoLigacao') != '') {
        console.log(sessionStorage.getItem('nvoip_estadoLigacao'))
    }
    setTimeout(() => {

        if (sessionStorage.getItem('nvoip_idChamada') == 'off') {

            if (sessionStorage.getItem('nvoip_estadoLigacao') != '') {
                if (sessionStorage.getItem('nvoip_atendeu') == 'on') {
                    msgSucesso(sessionStorage.getItem('nvoip_estadoLigacao'))
                    // Desliga a flag atendeu
                    sessionStorage.setItem('nvoip_atendeu', 'off')
                } else {
                    msgErro(sessionStorage.getItem('nvoip_estadoLigacao'))
                }
            }

            // Limpa a flag estadoLigacao
            sessionStorage.setItem('nvoip_estadoLigacao', '')

        } else {
            nvoip_consultarChamada((d) => {

            })
        }
        nvoip_timer()
    }, 3000)
}

function nvoip_autenticar(next) {
    var details = {
        'username': '102798001',
        'password': '51b6efaf-8327-11ee-aa50-02a0b41d24f8',
        'grant_type': 'password'
    };
    var formBody = [];
    for (var property in details) {
        var encodedKey = encodeURIComponent(property);
        var encodedValue = encodeURIComponent(details[property]);
        formBody.push(encodedKey + "=" + encodedValue);
    }
    formBody = formBody.join("&");

    try {
        fetch("https://api.nvoip.com.br/v2/oauth/token", {
            method: 'post',

            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'Authorization': 'Basic TnZvaXBBcGlWMjpUblp2YVhCQmNHbFdNakl3TWpFPQ=='
            },
            body: formBody
        }).then((resp) => resp.json()).then((d) => {
            sessionStorage.setItem('nvoip_access_token', d.access_token)
            sessionStorage.setItem('nvoip_expires_in', d.expires_in)
            sessionStorage.setItem('nvoip_refresh_token', d.refresh_token)
            sessionStorage.setItem('nvoip_scope', d.scope)
            sessionStorage.setItem('nvoip_token_type', d.token_type)
            next(d)
        })

    } catch (error) {
        msgErro(error)
    }
}

function nvoip_dadosToken(next) {
    if (sessionStorage.getItem('access_token') == null) {
        next({
            status: 'erro',
            erro: 'não esta logado'
        })
    } else {
        const body = encodeURIComponent('token') + "=" + encodeURIComponent(sessionStorage.getItem('nvoip_access_token'))

        try {
            fetch("https://api.nvoip.com.br/v2/oauth/check_token", {
                method: 'post',

                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    'Authorization': 'Basic TnZvaXBBcGlWMjpUblp2YVhCQmNHbFdNakl3TWpFPQ=='
                },
                body: body
            }).then((resp) => resp.json()).then((d) => next({ status: 'ok' }))

        } catch (error) {
            msgErro(error)
        }
    }
}

function nvoip_atualizarToken(next) {

    if (sessionStorage.getItem('nvoip_refresh_token') == null) {
        nvoip_autenticar(() => next({ status: 'ok' }))
    } else {
        var details = {
            'refresh_token': sessionStorage.getItem('refresh_token'),
            'grant_type': 'refresh_token'
        };

        var formBody = [];

        for (var property in details) {
            var encodedKey = encodeURIComponent(property);
            var encodedValue = encodeURIComponent(details[property]);
            formBody.push(encodedKey + "=" + encodedValue);
        }
        formBody = formBody.join("&");

        try {
            fetch("https://api.nvoip.com.br/v2/oauth/token", {
                method: 'post',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    'Authorization': 'Basic TnZvaXBBcGlWMjpUblp2YVhCQmNHbFdNakl3TWpFPQ=='
                },
                body: formBody
            }).then((resp) => resp.json()).then((d) => {
                sessionStorage.setItem('nvoip_access_token', d.access_token)
                sessionStorage.setItem('nvoip_expires_in', d.expires_in)
                sessionStorage.setItem('nvoip_refresh_token', d.refresh_token)
                sessionStorage.setItem('nvoip_scope', d.scope)
                sessionStorage.setItem('nvoip_token_type', d.token_type)
                next({ status: "ok" })
            })
        } catch (error) {
            msgErro(error)
        }
    }
}

function nvoip_realizarChamada(destino, next) {
    /*
    { 
        timestamp: "2024-07-04T15:44:02.211+00:00", 
        status: 401, 
        error: "Unauthorized", message: "", 
        path: "/v2/oauth/token" }
     */
    sessionStorage.setItem('nvoip_numeroDestino', destino)
    const tipo = sessionStorage.getItem('nvoip_token_type')
    const valor = sessionStorage.getItem('nvoip_access_token')
    const token = `${tipo} ${valor}`

    try {
        fetch("https://api.nvoip.com.br/v2/calls/", {
            method: 'post',

            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify({
                "caller": sessionStorage.getItem('login_userVoip'),
                "called": gLimpaDocumento(destino)
            })

        }).then((resp) => resp.json()).then((d) => {

            if (d.state == "success") {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Conexão nvoip estabelecida')
                if (d.callId.charAt(0) != ' ') {

                    sessionStorage.setItem('nvoip_idChamada', d.callId)
                    nvoip_consultarChamada((d) => {
                        next({
                            status: "ok",
                            idChamada: d.callId
                        })
                    })//state: "calling_origin", 

                } else {
                    next({
                        status: "erro",
                        erro: d.callId
                    })
                }
            } else {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Operador não atendeu')
                next({
                    status: "erro",
                    erro: 'Operador não atendeu'
                })
            }
        })

    } catch (error) {
        msgErro(error)
    }
}

function nvoip_consultarChamada(next) {
    /*
    { 
        state: "calling_destination", 
        linkAudio: "https://s3-sa-east-1.amazonaws.com/gravacoes-nvoip/2024/07/04/ac3146d8-717f-4aa8-8957-c08d7a2b154b.mp3", 
        talkingDurationSeconds: 0, 
        totalDurationSeconds: "0", 
        caller: "102798001" 
    }
    { 
        state: "established", 
        linkAudio: "https://s3-sa-east-1.amazonaws.com/gravacoes-nvoip/2024/07/04/ac3146d8-717f-4aa8-8957-c08d7a2b154b.mp3", 
        talkingDurationSeconds: 0, 
        totalDurationSeconds: "0", 
        caller: "102798001" 
    }
    { 
        state: "finished", 
        linkAudio: "https://s3-sa-east-1.amazonaws.com/gravacoes-nvoip/2024/07/04/ac3146d8-717f-4aa8-8957-c08d7a2b154b.mp3", 
        talkingDurationSeconds: 6, 
        totalDurationSeconds: "6", 
        caller: "102798002" 
    }

     */
    const callId = sessionStorage.getItem('nvoip_idChamada')
    try {
        fetch(`https://api.nvoip.com.br/v2/calls?callId=${callId}`, {
            method: 'get',

            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${sessionStorage.getItem('access_token')}`
            }

        }).then((resp) => resp.json()).then((d) => {
            if (d.state == 'calling_origin') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Chamando operador')


            } else if (d.state == 'calling_destination') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Chamando destino')



            } else if (d.state == 'established') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Chamada estabelecida')
                // Liga a flag atendeu
                sessionStorage.setItem('nvoip_atendeu', 'on')



            } else if (d.state == 'noanswer') {
                sessionStorage.setItem('nvoip_idChamada', 'off')
                sessionStorage.setItem('nvoip_estadoLigacao', 'Sem resposta')



            } else if (d.state == 'busy') {
                sessionStorage.setItem('nvoip_idChamada', 'off')
                sessionStorage.setItem('nvoip_estadoLigacao', 'Ocupado')



            } else if (d.state == 'finished') {

                if (sessionStorage.getItem('nvoip_atendeu') == "on") {
                    nvoip_gravaCusto(d.linkAudio, () => { })

                    // Seta a flag estadoLigacao
                    sessionStorage.setItem('nvoip_estadoLigacao', 'Finalizada com sucesso')
                } else {
                    // Seta a flag estadoLigacao
                    sessionStorage.setItem('nvoip_estadoLigacao', 'Finalizada sem sucesso')
                }

                // Desliga a flag idChamada
                sessionStorage.setItem('nvoip_idChamada', 'off')

            } else if (d.state == 'failed') {
                sessionStorage.setItem('nvoip_idChamada', 'off')
                sessionStorage.setItem('nvoip_estadoLigacao', 'Falhou')
            }
            next(d)
        })

    } catch (error) {
        msgErro(error)
    }
}

function nvoip_encerrarChamada() {
    const callId = sessionStorage.getItem('idChamada')
    try {
        fetch(`https://api.nvoip.com.br/v2/endcall?callId=${callId}`, {
            method: 'get',

            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${sessionStorage.getItem('access_token')}`
            }

        }).then((resp) => resp.json()).then((d) => {
            sessionStorage.setItem('nvoip_idChamada', 'off')
        })

    } catch (error) {
        msgErro(error)
    }
}

function nvoip_gravaCusto(link, next) {
    
    $.ajax({
        url: '/auxiliar/gravaCusto',
        method: 'POST',
        data: JSON.stringify({
            idProcesso: sessionStorage.getItem('ateProDado_proc_idProcesso'),
            idCentral: sessionStorage.getItem('ateProDado_proc_idCentral'),
            idFranqueado: sessionStorage.getItem('ateProDado_proc_idFranqueado'),
            idCliente: sessionStorage.getItem('ateProDado_proc_idCliente'),
            nomeCliente: sessionStorage.getItem('ateProDado_proc_cliNome'),

            idOperador: sessionStorage.getItem('login_userIdOperador'),
            // nomeOperador: sessionStorage.getItem('login_userNick'),
            
            tipoOperacao: 'LIGAÇÃO',
            dadoOperacao: sessionStorage.setItem('nvoip_numeroDestino'),
            linkAudio: link
        })
    }).done(function (r) {
        msgSucesso("Ligação registrada com sucesso")
        next()
    }).fail(function (e) {
        boxErro('erro ao tentar gravar ligaçao')
    })
}


function nvoip_consultarHistorico() { }

function nvoip_enviarSms() { }

function nvoip_enviarSms() { }

function nvoip_enviarTorpedoVoz() { }

function nvoip_agendarTorpedoVoz() { }

function nvoip_atualizarTorpedoVozAgendado() { }

function nvoip_buscarTorpedoVozAgendado() { }

function nvoip_excluirTorpedoVozAgendado() { }


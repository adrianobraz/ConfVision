class classNvoip {
    #login //Recebe o login do usuario usado na altenticação
    #senha
    #operador
    #token
    #tokenReflesh
    #destino
    #atendeu
    #idChamada
    constructor(login, senha, operador) {

        this.#login = login //'102798001'
        this.#senha = senha  //'51b6efaf-8327-11ee-aa50-02a0b41d24f8'
        this.#operador = operador //'102798012'

        this.#token = ''

    }

    efetuarChamada(destino, callback) {
        this.#destino = destino
        this.#atendeu = false
        this.#logar(resp => {
            if (resp.status === 'ok') { // Caso exista um token

                this.#ligar(resp => { // Efetua a ligação

                    if (resp.status === 'ok') {

                        // monitorar chamda
                        this.#monitorar(0, resp => { // monitora a ligação

                            if (resp.status === 'ok') {

                                if (this.#atendeu) {
                                    // gravar custo
                                    this.#gravaCusto(resp => {
                                        if (resp.status === 'ok') {
                                            msgSucesso(resp.message)
                                        } else {
                                            msgErro(resp.message)
                                        }
                                    })
                                } else {
                                    msgErro('Destino não atendeu')
                                }
                                callback(resp)

                            } else {
                                msgErro(resp.message)
                                callback(resp)
                            }

                        })

                    } else {
                        msgErro(resp.message)
                        callback(resp)
                    }
                })

            } else {
                msgErro(resp.message)
                callback(resp)
            }
        })





    }

    #logar(next) {

        if (this.#token == '') { // token vazio faz o login
            var details = {
                'username': this.#login,
                'password': this.#senha,
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
                    // sessionStorage.setItem('nvoip_access_token', d.access_token)
                    // sessionStorage.setItem('nvoip_expires_in', d.expires_in)
                    // sessionStorage.setItem('nvoip_refresh_token', d.refresh_token)
                    // sessionStorage.setItem('nvoip_scope', d.scope)
                    // sessionStorage.setItem('nvoip_token_type', d.token_type)

                    this.#token = d.access_token
                    this.#tokenReflesh = d.refresh_token

                    next({
                        status: 'ok',
                        message: 'Logado com sucesso'
                    })
                })

            } catch (erro) {
                next({
                    status: 'erro',
                    message: erro
                })

            }

        } else { // Caso exista um token atualiza o mesmo
            this.#token = '' //limpa o token para receber o novo

            var details = {
                'refresh_token': this.#tokenReflesh,
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
                    this.#token = d.access_token
                    this.#tokenReflesh = d.refresh_token

                    next({
                        status: 'ok',
                        message: 'Token atualizado com sucesso'
                    })
                })
            } catch (erro) {
                next({
                    status: 'erro',
                    message: erro
                })
            }
        }
    }

    #ligar(next) {
        try {
            fetch("https://api.nvoip.com.br/v2/calls/", {
                method: 'post',

                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `bearer ${this.#token}`
                },
                body: JSON.stringify({
                    "caller": this.#operador,
                    "called": this.#destino
                })

            }).then((resp) => resp.json()).then((d) => {
                // console.log(this.#operador, this.#destino, d)                
                // NORMAL_UNSPECIFIED
                // RECOVERY_ON_TIMER_EXPIRE
                if (d.state == "success") {
                    const resp = d.callId.replace(" ", '') // Retira os espaços em branco
                    console.log(resp)
                    if (resp == 'UNALLOCATED_NUMBER') { // Ramal do operador offline                        
                        console.log('Ramal do operador offline', '-', resp)
                        next({
                            status: 'erro',
                            message: 'Ramal do operador offline'
                        })

                    } else if (resp == 'ALLOTTED_TIMEOUT') {
                        console.log('Operador não atendeu', '-', resp)
                        next({
                            status: 'erro',
                            message: 'Operador não atendeu'
                        })

                    } else if (resp == 'NO_USER_RESPONSE') {
                        console.log('Operador rejeitou a ligação', '-', resp)
                        next({
                            status: 'erro',
                            message: 'Operador rejeitou a ligação'
                        })

                    } else if (resp == 'RECOVERY_ON_TIMER_EXPIRE') {
                        console.log('Operador rejeitou a ligação', '-', resp)
                        next({
                            status: 'erro',
                            message: 'Operador rejeitou a ligação'
                        })
                    } else {
                        this.#idChamada = d.callId
                        console.log('Operador atendeu ligação', '-', resp)
                        next({
                            status: 'ok',
                            message: 'Operador atendeu ligação'
                        })
                    }

                } else {
                    console.log('Ligação falhou', '-', d.state)
                    next({
                        status: 'erro',
                        message: 'Ligação falhou'
                    })
                }
            })

        } catch (erro) {
            console.log(erro)
            callback({
                status: 'erro',
                message: erro
            })
        }
    }

    #monitorar(cont, next) {
        try {
            fetch(`https://api.nvoip.com.br/v2/calls?callId=${this.#idChamada}`, {
                method: 'get',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.#token}`
                }

            }).then((resp) => resp.json()).then((d) => {

                if (d.state == 'calling_destination') {// não retorna
                    console.log('Chamando destino', '-', d.state)

                } else if (d.state == 'finished') { // return
                    console.log('Encerrada', '-', d.state)
                    next({
                        status: 'ok',
                        message: 'Ligação encerrada'
                    })
                    return

                } else if (d.state == 'established') {// nao retorna
                    console.log('Chamada estabelecida', '-', d.state)
                    // Liga a flag atendeu
                    this.#atendeu = true

                } else if (d.state == 'noanswer') { // retorna
                    console.log('Sem resposta', '-', d.state)
                    next({
                        status: 'erro',
                        message: 'Sem resposta'
                    })
                    return

                } else if (d.state == 'busy') { // retorna                    
                    console.log('Ocupado', '-', d.state)
                    next({
                        status: 'erro',
                        message: 'Ocupado'
                    })
                    return

                } else if (d.state == 'failed') { // retorna
                    console.log('Ligação falhou', '-', d.state)
                    next({
                        status: 'erro',
                        message: 'Ligação falhou'
                    })
                    return
                } else { // retorna
                    console.log(d)
                    console.log('Ligação falhou', '-', d.state)
                    next({
                        status: 'erro',
                        message: 'Ligação falhou'
                    })
                    return
                }

                setTimeout(() => {
                    if (cont > 180) {
                        console.log('Salvando link')
                    } else {
                        // console.log('looping:', cont)
                        cont++

                        this.#monitorar(cont, next)

                    }
                    // this.#consultar(cont, next)

                }, 5000)
            })


        } catch (erro) {
            console.log(erro)
            callback({
                status: 'erro',
                message: erro
            })
        }
    }

    #gravaCusto(next) {
        console.log('gravando')

        try {
            fetch(`https://api.nvoip.com.br/v2/calls?callId=${this.#idChamada}`, {
                method: 'get',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.#token}`
                }

            }).then((resp) => resp.json()).then((d) => {
                if (Number(d.talkingDurationSeconds) >= 1) {
                    try {
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
                                dadoOperacao: this.#destino,
                                linkAudio: d.linkAudio
                            })
                        }).done(function (r) {
                            console.log("Ligação registrada com sucesso")
                            next({
                                status: 'ok',
                                message: 'Ligação registrada com sucesso'
                            })
                        })
                    } catch (erro) {
                        console.log(erro)
                        callback({
                            status: 'erro',
                            message: erro
                        })
                    }
                } else {
                    next({
                        status: 'ok',
                        message: 'Destino não atendeu'
                    })
                }
            })
        } catch (erro) {
            console.log(erro)
            callback({
                status: 'erro',
                message: erro
            })
        }

    } 
}

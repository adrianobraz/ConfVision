

function nvoip_realizarChamada(origem, destino, next) {
    
    // Solicita um token a nvoip
    nvoip_solicitarToken((nvst) => {

        try {
            fetch("https://api.nvoip.com.br/v2/calls/", {
                method: 'post',
                headers: {
                    
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${nvst.access_token}`
                },     
                body: JSON.stringify({
                    "caller": origem,
                    "called": nvoip_limpaNumero(destino)
                })

            }).then((resp) => resp.json()).then((nvrc) => {
                
                if (nvrc.state == "success") {

                    function monitorar(flag) {
                        nvoip_consultarChamada(flag, nvst.access_token, nvrc.callId, (nvcc) => {

                            flag = nvcc.flag
                            if (nvcc.state) {
                                setTimeout(() => {
                                    monitorar(flag)
                                }, 3000)
                            } else {
                                next({
                                    state: nvrc.callId,
                                    msg: nvrc.msg
                                })
                            }
                        })
                    }
                    monitorar(false)

                } else if (nvrc.state == "calling_origin") {
                    sessionStorage.setItem('nvoip_estadoLigacao', 'Operador não atendeu')

                    next({
                        state: nvrc.callId,
                        msg: 'Operador não atendeu'
                    })
                }

            })

        } catch (err) {
            console.log(err)
        }
    })
}

function nvoip_encerrarChamada(callId, next) {
    // Solicita um token a nvoip
    nvoip_solicitarToken((nvst) => {

        try {
            fetch(`https://api.nvoip.com.br/v2/endcall?callId=${callId}`, {
                method: 'get',

                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${nvst.access_token}`
                }

            }).then((resp) => resp.json()).then((nvec) => {
                next({
                    state: nvec.state,
                    linkAudio: nvec.linkAudio
                })
            })

        } catch (error) {
            msgErro(error)
        }
    })
}

function nvoip_consultarChamada(flag, token, callId, next) {

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
    // const callId = sessionStorage.getItem('nvoip_idChamada')
    try {
        fetch(`https://api.nvoip.com.br/v2/calls?callId=${callId}`, {
            method: 'get',

            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            }

        }).then((resp) => resp.json()).then((d) => {

            console.log("Consultando: ", d.state)

            if (d.state == 'calling_origin') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Chamando operador')
                next({
                    state: true,
                    flag: flag,
                    msg: 'Chamando operador',
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })


            } else if (d.state == 'calling_destination') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Chamando destino')
                next({
                    state: true,
                    flag: flag,
                    msg: 'Chamando destino',
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })


            } else if (d.state == 'established') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Chamada estabelecida')
                next({
                    state: true,
                    flag: true,
                    msg: 'Chamada estabelecida',
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })

            } else if (d.state == 'noanswer') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Sem resposta')
                next({
                    state: false,
                    flag: flag,
                    msg: 'Sem resposta',
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })


            } else if (d.state == 'busy') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Ocupado')
                next({
                    state: false,
                    flag: flag,
                    msg: 'Ocupado',
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })
            } else if (d.state == 'finished') {
                let msg
                if (flag == true) {
                    // nvoip_gravaCusto(d.linkAudio, () => { })

                    sessionStorage.setItem('nvoip_estadoLigacao', 'Finalizada com sucesso')
                    msg = 'Finalizada com sucesso'
                } else {
                    sessionStorage.setItem('nvoip_estadoLigacao', 'Finalizada sem sucesso')
                    msg = 'Finalizada sem sucesso'
                }

                next({
                    state: false,
                    flag: flag,
                    msg: msg,
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })

            } else if (d.state == 'failed') {
                sessionStorage.setItem('nvoip_estadoLigacao', 'Falhou')

                next({
                    state: true,
                    flag: flag,
                    msg: 'Falhou',
                    linkAudio: d.linkAudio,
                    ligacaoTempo: talkingDurationSeconds,
                    totalTempo: totalDurationSeconds
                })
            }

        })

    } catch (error) {
        msgErro(error)
    }
}

function nvoip_limpaNumero(texto) {

    if (texto != undefined) {
        return texto
            .replace(/\./g, "")
            .replace(/\-/g, "")
            .replace(/\//g, "")
            .replace(/\(/g, "")
            .replace(/\)/g, "")
            .replace(/\-/g, "")
            .replace(/ /g, "")
            .replace(/\:/g, "")
    }

}

function nvoip_solicitarToken(next) {
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

            next(d)
        })

    } catch (error) {
        msgErro(error)
    }
}
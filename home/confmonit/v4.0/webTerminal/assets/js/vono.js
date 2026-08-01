
function vono_ligar(src, dst, idTarifar, nextLg) {
    if (src.length == 10 || src.length == 11) {
        src = '55' + src
    }

    if (dst.length == 10 || dst.length == 11) {
        dst = '55' + dst
    }

    const token = sessionStorage.getItem('login_voipToken')
    const key = sessionStorage.getItem('login_voipKey')
    const device_id = sessionStorage.getItem('login_voipDeviceId')

    $.ajax({
        url: `https://vono3.me/api/click2Call/${token}/${key}`,
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        data: JSON.stringify({
            "device_id": device_id,
            "src": src,
            "dst": dst
        })

    }).fail((e) => {
        console.log(e)
    }).done((rvlig) => {

        if (rvlig.error > 0) {
            console.log(rvlig.message)
            nextLg(false)
        } else {
            vono_consultar(src, dst, idTarifar, (rvconsut) => {
                nextLg(rvconsut)
            })
        }
    })
}

function vono_consultar(src, dst, idTarifar, nextCs) {
    const tempo = 2000
    function monitorar(src, dst, status, idChamada) {

        // Consulta a ligação destino
        vono_listar(src, dst, status, idChamada, (rvlist, idChamada) => {

            console.log(rvlist, idChamada)

            if (rvlist < 30) {
                setTimeout(() => { monitorar(src, dst, rvlist, idChamada) }, tempo)

            } else {
                /*
                    32 operador atendeu cliete atendeu (linha)
                    33 operador atendeu cliente atendeu (ramal)
                */
                if (rvlist == 32 || rvlist == 33) {
                    vono_gravaCusto(idTarifar, dst, idChamada)

                    nextCs(true)
                } else {
                    nextCs(false)
                }
            }
        })
    }
    monitorar(src, dst, 0, 0)
}

function vono_gravaCusto(idTarifar, numero, idChamada) {
    // alert("Gravando custo")
    $.ajax({
        url: `/auxiliar/gravaCusto`,
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idVinculo: idTarifar,
            tipoOperacao: "LIGAÇÃO",
            dadoOperacao: numero,
            aux: idChamada
        })

    }).fail((e) => {
        console.log(e)
    }).done((r) => {


    })
}

function vono_listar(src, dst, status, idChamada, next) {
    /*
    30 operador nao atendeu
    31 operador atendeu cliente nao
    32 operador atendeu cliete atendeu (linha)
    33 operador atendeu cliente atendeu (ramal)
    34 erro
    */
    // console.log('status', status)
    const token = sessionStorage.getItem('login_voipToken')
    const key = sessionStorage.getItem('login_voipKey')

    $.ajax({
        url: `https://vono3.me/api/onlineCalls/${token}/${key}`,
        method: "GET"
    }).fail((e) => {
        console.log(e)
    }).done((r) => {
        if (r.error > 0) {
            console.log(r.message)
            next(34, idChamada)
        } else {
            const retScr = r.data.filter(i => (i.destination == src))

            if (retScr.length > 0) {
                // Faz a consulta chamada origem            
                if (status == 0) {
                    retScr.map((d => {
                        idChamada = d.id
                        status = d.status
                    }))
                } else {
                    if (status == 1) {
                        console.log('tam', src.length)
                        if (src.length == 4) {
                            //Casso src seja ramal nao executa chamada destino
                            status = status + 2
                        } else {

                            // Caso seja numero faz a consulta da chada destino
                            const retDst = r.data.filter(
                                i => (i.destination == dst)
                            )

                            if (retDst.length > 0) {
                                retDst.map((d => status = status + d.status))
                            }
                        }
                    }
                }
                next(status, idChamada)
            } else {
                // Encerra a ligacao se operador desligar
                next(30 + status, idChamada)
            }

        }
    })
}


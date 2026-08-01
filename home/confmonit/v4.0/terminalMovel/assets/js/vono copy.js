// const token = '7486f90f-3048-5467-8345-40caff16aa7d-7047'
// const key = '0e5af12d-c0c8-58e8-b35e-71013ad16229-4728'

// vono_ligar(token, key, '9070', '5519998659753', '5519996751689', (r) => {
//     console.log(r)


// }) 

function vono_ligar(src, dst, next) {      
   
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
    }).done((rvl) => {

        if (rvl.error > 0) {
            console.log(rvl.message)
        } else {
            vono_consultar(src, dst, (rvc) => { next(rvc) })
        }
    })
}

function vono_listar(src, dst, status, tempo, next) {
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
            next(0)
        } else {
            let preCodigo = 10

            const org = r.data.filter(itemOrg => (itemOrg.destination == src))
            const des = r.data.filter(itemDes => (itemDes.destination == dst))

            // console.log("status ", status, "org ", org.length, "des ", des.length)

            if (org.length == 0 && des.length == 0) {
                // Monitora OFF Cliente OFF

                preCodigo = (status <= 1) ? 1 : status + 10

                tempo = 4000

            } else if (org.length > 0 && des.length <= 0) {
                // Monitora ON Cliente OFF

                preCodigo = (status <= 2) ? 2 : status + 10

                tempo = 4000

            } else if (org.length <= 0 && des.length > 0) {
                // Monitora OFF Cliente ON

                preCodigo = (status <= 3) ? 3 : status + 10

                tempo = 4000

            } else if (org.length > 0 && des.length > 0) {
                // Monitora OFF Cliente ON

                // console.log(org[0].status, des[0].status)

                if (org[0].status == 0 && des[0].status == 0) {
                    // Monitora CHAM Cliente CHAM

                    preCodigo = 4

                    tempo = 4000

                } else if (org[0].status == 1 && des[0].status == 0) {
                    // Monitora ATEN Cliente CHAM

                    preCodigo = 5

                    tempo = 6000

                } else if (org[0].status == 0 && des[0].status == 1) {
                    // Monitora CHAM Cliente ATEN

                    preCodigo = 6

                    tempo = 4000

                } else if (org[0].status == 1 && des[0].status == 1) {
                    // Monitora ATEN Cliente ATEN

                    preCodigo = 7

                    tempo = 4000
                }

            }

            next({
                status: (status < preCodigo) ? preCodigo : status,
                tempo: tempo
            })

        }
    })
}

function vono_consultar(src, dst, next) {

    function monitorar(src, dst, status, tempo) {

        // Consulta a ligação destino
        vono_listar(src, dst, status, tempo, (rvl) => {

            if (rvl.status <= 7) {
                status = rvl.status
                tempo = rvl.tempo
                // console.log("Codigo loop", status, tempo)
                setTimeout(() => { monitorar(src, dst, status, tempo) }, tempo)
            } else {
                status = rvl.status
                
                if (rvl.status == 17) {                   
                    next(true)
                }  else {                   
                    next(false)
                }
            }
        })
    }
    monitorar(src, dst, 0, 3000)
}

function vono_gravaCusto() {
    alert("Gravando custo")
}



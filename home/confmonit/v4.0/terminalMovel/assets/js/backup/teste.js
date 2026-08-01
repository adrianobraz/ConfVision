class api4com {
    #token; #operador; #destino; #idChamada; #status

    #statusCode; name; message;


    constructor(token, operador) {
        this.#token = token
        this.#operador = operador //1000
    }

    efetuarChamada(destino) {
        this.#destino = destino

        this.#ligar(resp => {
            console.log(resp)
        })
    }
    // { statusCode: 424, name: "Error", message: "call from 1000 to 19998659753 has been failed" }

    #ligar(next) {
        try {
            fetch(`https://api.api4com.com/api/v1/calls?access_token=${this.#token}`, {
                method: 'post',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    "caller": this.#operador, //sessionStorage.getItem('login_userVoip'),
                    "called": this.#destino,
                    "extension": this.#operador, //sessionStorage.getItem('login_userVoip'),
                    "metadata": {
                        "firstParam": "123456",
                        "secondParam": "XPTO"
                    }
                })

            }).then((resp) => resp.json()).then((d) => {

                if ('id' in d) {
                    this.#status = 'Conexão nvoip estabelecida'
                    this.#idChamada = d.id
                    console.log(d.id)
                } else {
                    this.#status = 'Operador não atendeu'
                }

                next(d)
            })

        } catch (error) {
            msgErro(error)
        }
    }
}

// let voip = new api4com('AkDcXi5HA6VVnnU6rimYSiiAQ5I9ooIUNDyxGJdcfpdjokaC41D7UBiB2hEf1mdf', '1000')

//voip.efetuarChamada('19996751689')


//

// const tokenApi4com = 'pvG0Hns2uuxma5yReTgXas9q6DYfGr0Ap1rUrBguQiWZbdvr5BzaM13UawRcyP8C'
//7d42702e-f09d-47dd-a964-f7fd1dca0543
//3e6370fe-7bf7-4ca2-8fd7-15c972bc6789
// function api4com_realizarChamada(destino) {

//     const extension = 1000

// function api4com_consultarChamada(id) {
//     try {
//         fetch(`https://api.api4com.com/api/v1/calls/${id}/hangup`, {
//             method: 'post',
//             headers: {
//                 'Content-Type': 'application/json',
//             }
//         }).then((resp) => resp.json()).then((d) => {
//             console.log(d)

//             if (d.status == '200') {
//                 sessionStorage.setItem('api4com_estadoLigacao', 'Finalizada')
//                 sessionStorage.setItem('api4com_idChamada', d.id)
//             } else {
//                 sessionStorage.setItem('api4com_estadoLigacao', 'Operador não atendeu')
//                 console.log(d.message)
//             }

//         })

//     } catch (error) {
//         msgErro(error)
//     }
// }


// function api4com_consultarChamada(id) {
//     try {
//         fetch(`https://api.api4com.com/api/v1/calls?page=1&access_token=${tokenApi4com}`, {
//             method: 'post',
//             headers: {
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({
//                 "caller": extension, //sessionStorage.getItem('login_userVoip'),
//                 "called": gLimpaDocumento(destino),
//                 "extension": extension, //sessionStorage.getItem('login_userVoip'),
//                 "metadata": {
//                     "firstParam": "123456",
//                     "secondParam": "XPTO"
//                 }
//             })

//         }).then((resp) => resp.json()).then((d) => {

//             if (d.status == '200') {
//                 sessionStorage.setItem('api4com_estadoLigacao', 'Conexão nvoip estabelecida')
//                 sessionStorage.setItem('api4com_idChamada', d.id)
//             } else {
//                 sessionStorage.setItem('api4com_estadoLigacao', 'Operador não atendeu')
//                 console.log(d.message)
//             }
//         })

//     } catch (error) {
//         msgErro(error)
//     }
// }   


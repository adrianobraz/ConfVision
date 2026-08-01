function login_logar() {
    fetch('login/logar', {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            email: document.querySelector('#email').value,
            senha: document.querySelector('#senha').value

        })
    }).then(resp => resp.json()).then(ret => {
        const d = ret.dados
        console.log(ret )
        if (ret.status == 'OK') {
            // Id do vinculo do usuario
            sessionStorage.setItem('login_userVinculo', d.idVinculo)
            sessionStorage.setItem('login_userVinculoNome', d.nomeVinculo)
            sessionStorage.setItem('login_userIdOperador', d.idOperador)
            sessionStorage.setItem('login_userCelular', d.userCelular)
            sessionStorage.setItem('login_userEmail', d.userEmail)
            sessionStorage.setItem('login_userMaster', d.userMaster)
            sessionStorage.setItem('login_userNome', d.userNome) // Nome do usuario
            sessionStorage.setItem('login_userNick', d.userNick) // Nick do usuario
            sessionStorage.setItem('login_benuvemFraCod', d.userBenuvem)

            // Credenciais Vono
            sessionStorage.setItem('login_voipToken', d.voipToken) // Token Voip
            sessionStorage.setItem('login_voipKey', d.voipKey) // Key Voip
            sessionStorage.setItem('login_voipDeviceId', d.voipDeviceId) // Key Voip
            sessionStorage.setItem('login_voipAtivo', d.voipAtivo) // Voip Ativo

            // Credenciais benuvem
            sessionStorage.setItem('login_benuvemLogin', d.benuvemEmail)
            sessionStorage.setItem('login_benuvemSenha', d.benuvemSenha)
            sessionStorage.setItem('login_benuvemAtivo', d.benuvemAtivo)

            window.location.href = "/home"

        } else {
            console.log(ret.status)
            msgErro("Usuario o senha invalidas")
        }
    }).catch(erro => {
        console.log(erro)
    })
}


// sessionStorage.setItem('login_userVoip', '102798002') // usuario do master

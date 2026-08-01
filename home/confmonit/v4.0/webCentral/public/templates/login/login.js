$(window).on("load", function () {
    sessionStorage.clear()
    document.addEventListener('keypress', function (event) {
        if (event.key === 'Enter') {
            logar()
        }
    });
})

function logar() {
    if ($('#email').val() == '') {
        boxErro("Digite um email")
        return
    } else if ($('#senha').val() == '') {
        boxErro("Digite uma senha")
        return
    }

    $.ajax({
        url: '/logar',
        method: 'POST',
        data: JSON.stringify({
            email: $('#email').val(),
            senha: $('#senha').val(),
        })

    }).fail(function (e) {
        console.log(e)
        boxErro("Usuario o senha invalidos")
    }).done(function (r) {
        // console.log(r)
        const d = r.dados
        // console.log(r)
        if (d.idVinculo == "CENTRAL") { // Verifica se e um usario da central
            if (d.usuarioWeb == "S") {
                if (d.ativo == "S") { // Verifica se usuario esta ativo
                   processa(d)
                } else {
                    boxErro('Usuario Bloqueado no sistema')
                }
            } else {
                boxErro('Usuario Bloqueado para acesso web')
            }
        } else {
            boxErro('Usuario não pode acessar esse recurso')
        }

    })
}

function processa (d) {
    const idCentralUUID = d.idCentralUUID || d.IDCentralUUID || ""
    if (d.idVinculo == "CENTRAL" && !idCentralUUID) {
        boxErro("Usuario sem Central vinculada (IDCentralUUID). Atualize a API e rode o SQL multi-tenant.")
        return
    }

    sessionStorage.setItem("loginEmail", d.email1)
    sessionStorage.setItem("loginId", d.idUsuario)
    sessionStorage.setItem("loginNome", d.nome)
    sessionStorage.setItem("loginNick", d.nick)
    sessionStorage.setItem("loginToken", d.token)
    sessionStorage.setItem("loginMaster", d.master)
    sessionStorage.setItem("loginVinculo", d.idVinculo)
    sessionStorage.setItem("loginIdCentralUUID", idCentralUUID)
    sessionStorage.removeItem("loginBreakglass")

    if ($('#senha').val() == 'usuario123') {
        window.location = '/carregar-alterar-senha'
    } else {
        window.location = '/home'
    }
}
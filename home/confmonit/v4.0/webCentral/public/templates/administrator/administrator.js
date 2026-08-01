$(window).on("load", function () {
    sessionStorage.clear()
    document.addEventListener('keypress', function (event) {
        if (event.key === 'Enter') {
            logar()
        }
    });
})

function logar() {
    if ($('#usuario').val() == '') {
        boxErro("Digite o usuario")
        return
    } else if ($('#senha').val() == '') {
        boxErro("Digite a senha")
        return
    }

    $.ajax({
        url: '/administrator/logar',
        method: 'POST',
        data: JSON.stringify({
            usuario: $('#usuario').val(),
            senha: $('#senha').val(),
        })
    }).fail(function (e) {
        console.log(e)
        boxErro("Usuario ou senha invalidos")
    }).done(function (r) {
        const d = r.dados
        if (d && d.idVinculo == "CENTRAL" && d.usuarioWeb == "S" && d.ativo == "S") {
            processa(d)
        } else {
            boxErro('Usuario nao pode acessar esse recurso')
        }
    })
}

function processa(d) {
    const idCentralUUID = d.idCentralUUID || d.IDCentralUUID || ""
    sessionStorage.setItem("loginEmail", d.email1)
    sessionStorage.setItem("loginId", d.idUsuario)
    sessionStorage.setItem("loginNome", d.nome)
    sessionStorage.setItem("loginNick", d.nick)
    sessionStorage.setItem("loginToken", d.token)
    sessionStorage.setItem("loginMaster", d.master)
    sessionStorage.setItem("loginVinculo", d.idVinculo)
    sessionStorage.setItem("loginIdCentralUUID", idCentralUUID)
    sessionStorage.setItem("loginBreakglass", "S")
    window.location = '/home'
}

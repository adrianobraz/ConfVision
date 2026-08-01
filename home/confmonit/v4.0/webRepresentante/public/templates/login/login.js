$(window).on("load", function () {
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
        boxErro("Erro ao logar")
        console.log(e)
    }).done(function (r) {
        console.log(r)
        const d = r.dados
        const idCentralUUID = d.idCentralUUID || d.IDCentralUUID || ""
        if (!idCentralUUID) {
            boxErro("Usuario sem Central vinculada (IDCentralUUID). Faca logout/login apos atualizar a API.")
            return
        }

        sessionStorage.setItem("loginEmail", d.email1)
        sessionStorage.setItem("loginId", d.idUsuario)
        sessionStorage.setItem("loginNome", d.nome)
        sessionStorage.setItem("loginToken", d.token)
        sessionStorage.setItem("loginMaster", d.master)
        sessionStorage.setItem("loginRepId", d.idVinculo)
        sessionStorage.setItem("loginIdCentralUUID", idCentralUUID)

        if (d.ativo == "S") {
            if (d.tipo == "REP") {
                window.location = '/home'
            } else {
                boxErro("Usuario não pode acessar esse recurso")
            }
        } else {
            boxErro("Usuario bloqueado")
        }
    })
}

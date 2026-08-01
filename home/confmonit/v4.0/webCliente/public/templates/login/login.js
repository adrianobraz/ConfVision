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
            email1: $('#email').val(),
            senha: $('#senha').val(),
        })

    }).fail(function (e) {
        boxErro("Erro ao logar")
        console.log(e)
    }).done(function (r) {
        console.log(r)
        sessionStorage.setItem("loginId", r.dados.idCliente)
        sessionStorage.setItem("loginNome", r.dados.nome)
        sessionStorage.setItem("loginNick", r.dados.nick)
        sessionStorage.setItem("loginEmail", r.dados.email1)
        sessionStorage.setItem("loginEmail1", r.dados.email1)
        sessionStorage.setItem("loginEmail2", r.dados.email2)
        sessionStorage.setItem("loginInformativo", r.dados.informativo)
        sessionStorage.setItem("loginIdFranq", r.dados.idFranqueado)
        sessionStorage.setItem("loginNomeFranq", r.dados.fraRazao)
        sessionStorage.setItem("loginDispAppConta", r.dados.dispAppConta)


        carregaDadosReceptor()



        if (r.dados.ativo == "S") {
            window.location = '/home'
        } else {
            boxErro("Usuario bloqueado")
        }

    })
}

function carregaDadosReceptor() {
    $.ajax({
        url: '/carregaDadosReceptor',
        method: 'POST',

    }).fail(function (e) {
        boxErro("Erro ao logar")
        console.log(e)
    }).done(function (r) {
        // console.log(r)
        r.dados.forEach(i => {
            if (i.modulo == "REC_COMANDO") {                
                sessionStorage.setItem('loginComandoSenha', i.senha)
            } else if (i.modulo == "REC_WEB") {
                sessionStorage.setItem('loginWebEventoSenha', i.senha)
            }

        });

    })
}
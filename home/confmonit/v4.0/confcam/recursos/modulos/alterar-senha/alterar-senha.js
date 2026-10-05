$(window).on("load", function () {
    $('#alterar-senha').on('click', alterarSenha)
});

function alterarSenha() {
    const senha = $('#nova-senha').val()
    const confirma = $('#confirma-senha').val()
    if (senha != '' && confirma != '') {
        if (senha == confirma) {
            
            $.ajax({
                url: '/alterar-senha',
                method: 'POST',
                data: JSON.stringify({
                    idUsuario: localStorage.getItem('idUsuario'),
                    senha: senha
                })
            }).fail(function (e) {
                console.log(e)
                boxErro("Erro ao alterar senha")
            }).done(function (r) {
                boxSenhaAlterada('Senha alterada com sucesso')
                window.location = '/carregar-menu-principal'
            })
        } else {
            boxAdvertenciaAuto('As senhas não coencidem')
        }

    } else {
        boxAdvertenciaAuto('Uma senha deve ser informada')
    }
}
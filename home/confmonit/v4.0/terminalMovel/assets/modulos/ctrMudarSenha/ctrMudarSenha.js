function ctrMudarSenha_start() {
    $('#boxDir').empty()
    const uri = '/assets/modulos/ctrMudarSenha/ctrMudarSenha.html'
    $('#boxDir').load(uri, () => {
        $('#ctrMudarSenha_btnGravar').on('click', ctrMudarSenha_btnGravar)
        $('#ctrMudarSenha_btnFechar').on('click', ctrMudarSenha_btnFechar)
    })
}

function ctrMudarSenha_btnFechar() {
    if (typeof terminalMovel_mostrarProcessoPrincipal === 'function') {
        terminalMovel_mostrarProcessoPrincipal()
    } else {
        proFranqFiltro_start()
    }
}

function ctrMudarSenha_btnGravar(){   
    
    if ($('#ctrMudarSenha_cpSenha').val() === $('#ctrMudarSenha_cpSenhaRepitir').val()){

        $.ajax({
            url: '/ctrMudarSenha/gravar',
            method: 'Post',
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
                "Authorization": "Bearer " + sessionStorage.getItem('token')
            },
            data: JSON.stringify({ 
                idOperador: sessionStorage.getItem('login_userIdOperador'),
                senha: $('#ctrMudarSenha_cpSenha').val()
            })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            ctrMudarSenha_btnFechar()
        })    
    }else {
        msgErro("As senhas não coencidem")
    }

}
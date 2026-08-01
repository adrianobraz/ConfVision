$(document).ready(function () {
    $('#btn-administrator-logar').on('click', fazerLoginAdministrator)
    $('#senha').keydown(function (e) {
        if (e.keyCode === 13) {
            fazerLoginAdministrator()
        }
    })
})

function fazerLoginAdministrator() {
    const usuario = ($('#usuario').val() || '').trim()
    const senha = ($('#senha').val() || '').trim()
    if (!usuario || !senha) {
        boxMesagemAtencaoPersonalizada('Informe usuário e senha.')
        return
    }

    $.ajax({
        url: '/administrator/login',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ usuario: usuario, senha: senha })
    }).fail(function (xhr) {
        const msg = (xhr.responseJSON && xhr.responseJSON.status) || 'Usuário ou senha inválidos.'
        boxMesagemAtencaoPersonalizada(msg)
    }).done(function (r) {
        localStorage.clear()
        sessionStorage.clear()
        localStorage.setItem('papel', 'ADM')
        localStorage.setItem('nomeUsuario', 'Administrator')
        localStorage.setItem('token', 'administrator-rtmp')
        const dest = (r.dados && r.dados.redirect) || '/administrator/rtmp-falhas'
        window.location = dest
    })
}

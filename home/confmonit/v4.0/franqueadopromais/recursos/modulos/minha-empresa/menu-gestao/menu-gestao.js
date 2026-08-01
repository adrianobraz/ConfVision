$(document).ready(function () {
    function exibirPermissoes() {
        $('#botao-permissoes-acesso').removeAttr('hidden').show()
    }

    var concluirMenu = function () {
        if (typeof fpAplicarPermissoesUI === 'function') {
            fpAplicarPermissoesUI()
        }
    }

    if ($('#botao-permissoes-acesso').data('fp-master') === 'S') {
        localStorage.setItem('loginMaster', 'S')
        exibirPermissoes()
    } else if (localStorage.getItem('loginMaster') === 'S') {
        exibirPermissoes()
    } else if (typeof fpSyncMaster === 'function') {
        fpSyncMaster(function (master) {
            if (master === 'S') {
                exibirPermissoes()
            }
        })
    }

    var idFranqueado = localStorage.getItem('idFranqueado')
    var idPacote = localStorage.getItem('fraIdPacote')

    if (!idPacote && idFranqueado) {
        $.ajax({
            url: '/getFranqDadosById',
            method: 'POST',
            data: JSON.stringify({ fraId: idFranqueado })
        }).fail(concluirMenu).done(function (r) {
            if (r.status === 'OK' && r.dados) {
                idPacote = r.dados.fraIdPacote || ''
                localStorage.setItem('fraIdPacote', idPacote)
            }
            if (idPacote && typeof fpCarregarFlagsPacote === 'function') {
                fpCarregarFlagsPacote(idPacote, concluirMenu)
            } else {
                concluirMenu()
            }
        })
        return
    }

    if (idPacote && typeof fpCarregarFlagsPacote === 'function') {
        fpCarregarFlagsPacote(idPacote, concluirMenu)
    } else {
        concluirMenu()
    }
})

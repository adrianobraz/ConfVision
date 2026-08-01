$(document).ready(function () {
    if (localStorage.getItem('terminalAtivo') == 1) {
        $('#botao-cadastro-operador').removeAttr('hidden').show();
    } else {
        $('#botao-cadastro-operador').attr('hidden', true).hide();
    }

    function exibirPermissoes() {
        $('#botao-permissoes-acesso').removeAttr('hidden').show();
    }

    if ($('#botao-permissoes-acesso').data('fp-master') === 'S') {
        localStorage.setItem('loginMaster', 'S');
        exibirPermissoes();
    } else if (localStorage.getItem('loginMaster') === 'S') {
        exibirPermissoes();
    } else if (typeof fpSyncMaster === 'function') {
        fpSyncMaster(function (master) {
            if (master === 'S') {
                exibirPermissoes();
            }
        });
    }

    if (typeof fpAplicarPermissoesUI === 'function') {
        fpAplicarPermissoesUI();
    }
});

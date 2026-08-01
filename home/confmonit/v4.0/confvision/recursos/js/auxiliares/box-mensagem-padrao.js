function boxSenhaAlterada(texto) {
    CvMsg.sucesso('Alterada para: ' + texto, 'Senha resetada com sucesso')
}

function boxInseridoSucesso(idAtribuido) {
    CvMsg.sucesso('ID atribuído: ' + idAtribuido, 'Inserido com sucesso')
}

function boxConfirmarExcluir(item, url, id) {
    CvMsg.confirmar('Apagar ' + item + '?', 'Não poderá reverter essa ação.', {
        icon: 'warning',
        confirmText: 'Sim, pode apagar!'
    }).then(function (result) {
        if (!result.isConfirmed) return
        $.ajax({
            url: url,
            method: 'POST',
            data: JSON.stringify({ id: id })
        }).fail(function (erro) {
            boxErro(erro)
        }).done(function () {
            boxDeletadoSucesso()
        })
    })
}

function boxDeletadoSucesso() {
    CvMsg.sucesso('', 'Apagado com sucesso')
}

function boxAteradoSucesso() {
    CvMsg.sucesso('', 'Alterado com sucesso')
}

function boxEmCostrucao(link) {
    CvMsg.aviso('Em breve novidades', 'Em construção').then(function () {
        window.location.href = '/' + link
    })
}

function boxMesagemAtencaoPersonalizada(mensagem) {
    CvMsg.aviso(mensagem)
}

function boxProcessando(texto) {
    CvMsg.processando(texto || 'Processando...')
}

function boxErro(erro) {
    var msg = erro
    if (erro && erro.responseText) {
        try {
            var j = JSON.parse(erro.responseText)
            msg = j.message || j.status || erro.responseText
        } catch (e) {
            msg = erro.responseText
        }
    } else if (erro && erro.message) {
        msg = erro.message
    }
    CvMsg.erro(String(msg || 'Erro'))
}

function boxMensagemAuto(mensagem) {
    CvMsg.info(mensagem)
}

function boxSucessoAuto(mensagem) {
    CvMsg.sucesso(mensagem)
}

function boxAdvertenciaAuto(mensagem) {
    CvMsg.aviso(mensagem)
}

function boxAdvertenciaCampoAuto(mensagem, idCampo) {
    CvMsg.avisoCampo(mensagem, idCampo)
}

function boxFechar() {
    CvMsg.fechar()
}

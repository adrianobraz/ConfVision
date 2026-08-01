function boxSenhaAlterada(texto) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Senhar resetada com sucesso',
        text: `Aterada para: ${texto}`,
        confirmButtonColor: '#3085d6',
    })
}

function boxInseridoSucesso(idAtribuido) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Inserido com Sucesso',
        text: `ID atribuido: ${idAtribuido}`,
        confirmButtonColor: '#3085d6',
    })
}

function boxConfirmarExcluir(item, url, id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar ${item}?`,
        text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode apagar!'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                start: boxProcessando(),
                url: url,
                method: 'POST',
                data: JSON.stringify(
                    {
                        id: id
                    }
                )
            }).fail(function (erro) {
                boxErro(e)
            }).done(function (r) {
                    boxDeletadoSucesso()
            })
        }
    })
}

function boxDeletadoSucesso() {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Apagado com sucesso',
        //showConfirmButton: false,
        confirmButtonColor: '#3085d6',
        timer: 3000
    })
}

function boxAteradoSucesso() {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Alterado com Sucesso',
        confirmButtonColor: '#3085d6',
        timer: 3000
    })
}

function boxEmCostrucao(link) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Em Cosntrução',
        text: `Em breve novidades`,
        confirmButtonColor: '#3085d6',
        didClose: () => {
            window.location.href = `/${link}`;
        }
    })
}

function boxMesagemAtencaoPersonalizada(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
    })
}

function boxProcessando(texto = "") {
    const txt = (texto == "") ? 'Processando...' : texto
    Swal.fire({
        //position: 'top-end',
        //icon: 'success',
        title: txt,
        showConfirmButton: false,
        didOpen: () => {
            Swal.showLoading()
        }
    })
}

function boxErro(erro) {
    Swal.fire({
        position: 'top',
        icon: 'error',
        title: erro,
        showConfirmButton: false,
        timer: 2000
    })
}

function boxMensagemAuto(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 2000
    })
}

function boxSucessoAuto(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Sucesso !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 2000
    })
}

function boxAdvertenciaAuto(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 3000,
    })
}

function boxAdvertenciaCampoAuto(mensagem, idCampo) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 3000,
        didClose: () => {
            $(idCampo).focus()
        }

    })
}

function boxFechar() {
    Swal.close()
}
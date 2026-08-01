
function boxInseridoSucesso(idAtribuido) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Inserido com Sucesso',
        text: `ID atribuido: ${idAtribuido}`,
        confirmButtonColor: '#3085d6',
    })
}

function boxSucesso(msg) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Sucesso',
        text: msg,
        confirmButtonColor: '#3085d6',
        timer: 3000
    })
}

function boxSenhaAlterada(texto) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Senhar resetada com sucesso',
        text: `Aterada para: ${texto}`,
        confirmButtonColor: '#3085d6',
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

function boxErro(erro) {
    Swal.fire({
        position: 'top',
        icon: 'error',
        title: erro,
        showConfirmButton: false,
        timer: 2000
    })
}

function boxConfirmarExcluir(item, executar) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar: ${item}?`,
        text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode apagar!'
    }).then((result) => { if (result.isConfirmed) executar() })
}

function boxConfirmarCancelar(item, executar) {
    Swal.fire({
        //position: 'top',
        title: `Quer realmente cancelar: ${item}?`,
        // text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode cancelar!'
    }).then((result) => { if (result.isConfirmed) executar() })
}

function boxConfirmarReverterCancelar(item, executar) {
    Swal.fire({
        //position: 'top',
        title: `Quer realmente reverter o cancelamento: ${item}?`,
        // text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode reverter!'
    }).then((result) => { if (result.isConfirmed) executar() })
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
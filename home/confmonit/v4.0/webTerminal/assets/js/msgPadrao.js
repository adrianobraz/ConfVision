function msgSucesso(msg = 'Operação realizada com Sucesso') {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: msg,
        confirmButtonColor: '#242343',
        timer: 3000
    })
}

function msgErro(erro) {
    Swal.fire({
        position: 'top',
        icon: 'error',
        title: erro,
        showConfirmButton: false,
        timer: 2000,
        //backdrop: false
    })
}


function msgBox(texto) {
    Swal.fire({
        position: 'top',
        //icon: 'info',
        text:  texto    ,
        showConfirmButton: true,
        backdrop: false,
        width: '80%',
    })
}


function msgAtendimento(mensagem){
    Swal.fire({
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#8B0000',
        width: 800,
       //height: 600,
        padding: "3em",
        color: "#8B0000",
        toast: true,

      });
}

function msgProcessando(texto = "") {
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

function msgFechar() {
    Swal.close()
}
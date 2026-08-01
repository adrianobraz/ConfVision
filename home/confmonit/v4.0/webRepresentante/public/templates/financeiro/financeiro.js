$(window).on('load', function () {
    sessionStorage.setItem("deletar", "N")
    document.addEventListener('keydown', function (event) {
        if(event.ctrlKey && event.key == "q") {
            sessionStorage.setItem("deletar", "S")
        }
    });
    document.addEventListener('keyup', function (event) {
        if(event.ctrlKey && event.key == "q") {
            sessionStorage.setItem("deletar", "N")
        }
    });
    listarFranqueadoCancelado()
})

function listarFranqueadoCancelado() {
    $.ajax({
        url: `/financeiro/listarFranq`,
        method: 'POST',
        data: JSON.stringify({ repId: sessionStorage.getItem("loginRepId") })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()

        r.dados.filter((i) => i.fraDataCancelamento != "").map(item => {

            $('tbody').append(`
                <tr>
                    <td class="text-sm">${item.fraCnpj}</td>
                    <td class="text-sm">${item.fraRazao}</td>
                    <td tipo="btnCancelar" 
                        title="Reverte o cancelamento do franquedo"
                        idFra="${item.fraId}" 
                        razao="${item.fraRazao}"
                        class="text-sm bg-green-700  btn">
                       <i class="bi bi-check-lg text-white"></i>
                    </td>
                </tr>    
            `)
        })

        // Associa os botoes          
        $('td[tipo=btnCancelar]').on('click', function () {
            const idFra = this.getAttribute("idFra")
            const nome = this.getAttribute("razao")
            reverterCancelarFranq(idFra, nome)
        })

    })
}

function reverterCancelarFranq(idFra, nome) {
    if (sessionStorage.getItem("deletar") == 'S') {
        boxConfirmarExcluir(nome, () => {
            $.ajax({
                url: `/financeiro/deletarFranq`,
                method: 'POST',
                data: JSON.stringify({ fraId: idFra })
            }).fail(function (e) {
                console.log(e)
            }).done(function (r) {
                boxSucesso('Deletado com sucesso')
                listarFranqueadoCancelado()
            })
        })
    } else {
        boxConfirmarReverterCancelar(nome, () => {
            $.ajax({
                url: `/financeiro/reverterCancelarFranq`,
                method: 'POST',
                data: JSON.stringify({ fraId: idFra })
            }).fail(function (e) {
                console.log(e)
            }).done(function (r) {
                boxSucesso('Revertido com sucesso')
                listarFranqueadoCancelado()
            })
    
        })
    }



}

// function habilitar(id, status) {
//     $.ajax({
//         url: `/financeiro/habilitarFranq`,
//         method: 'POST',
//         data: JSON.stringify({
//             idFranqueado: id,
//             ativo: status
//         })
//     }).fail(function (e) {
//         console.log(e)
//     }).done(function (r) {
//         if (r.status == "OK") {
//             listarFranqueado()
//         } else {
//             boxErro(r.status)
//         }
//     })
// }
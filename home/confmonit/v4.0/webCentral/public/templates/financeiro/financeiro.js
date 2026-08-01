$(window).on('load', function () {
    listarFranqueado()
})

function listarFranqueado() {
    const payload = payloadTenant({})
    if (!payload) return

    $.ajax({
        url: `/financeiro/listarRep`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar representantes')
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()

        r.dados.map(item => {

            let cnpj
            let botao
            if (item.dataCancelamento == '') {
                cnpj = item.cnpj
                botao = `
                    <td title="Cancela o representante" 
                        class=" btn bg-red-700 text-white" 
                        tipo="cancela" 
                        idRep="${item.idRepresentante}"
                        nome="${item.razaoSocial}">
                        <i class="bi bi-eraser-fill"></i>
                    </td>
                `
            } else {
                cnpj = 'CANCELADO'
                botao = `
                    <td title="Reverte cancelamento o representante" 
                        class=" btn bg-zinc-600 text-white"
                        tipo="revertCancelamento" 
                        idRep="${item.idRepresentante}"
                        nome="${item.razaoSocial}">
                        <i class="bi bi-eraser-fill"></i>
                    </td>
                `
            }
            
            $('tbody').append(`
                <tr>
                    <td class="text-sm">${cnpj}</td>
                    <td class="text-sm">${item.razaoSocial}</td>
                    
                    ${botao}
                </tr>    
            `)
        })

        $('td[tipo=cancela]').on('click', function () {
            const idRep = this.getAttribute("idRep")
            const nome = this.getAttribute("nome")

            cancelaRep(idRep, nome)

        })

        $('td[tipo=revertCancelamento]').on('click', function () {
            const idRep = this.getAttribute("idRep")
            const nome = this.getAttribute("nome")

            reverteCancelaRep(idRep, nome)

        })
    })
}

function cancelaRep(id, nome) {
    
    boxConfirmarCancelar(nome, () => {
        $.ajax({
            url: `/financeiro/cancelaRep`,
            method: 'POST',
            data: JSON.stringify({ idRepresentante: id })

        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {

            if (r.status == "OK") {
                listarFranqueado()
            } else {
                boxErro(r.status)
            }
        })
    })
}

function reverteCancelaRep(id, nome) {
    
    boxConfirmarReverteCancelar(nome, () => {
        $.ajax({
            url: `/financeiro/reverteCancelaRep`,
            method: 'POST',
            data: JSON.stringify({ idRepresentante: id })

        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {

            if (r.status == "OK") {
                listarFranqueado()
            } else {
                boxErro(r.status)
            }
        })
    })
}
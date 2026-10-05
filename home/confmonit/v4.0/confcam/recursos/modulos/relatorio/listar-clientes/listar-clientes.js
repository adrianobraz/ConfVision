$(document).ready(function () {
    $('#btn-imprimir').on('click', function(){
        gerarPdf('RELAÇÃO DE CLIENTES','#tabListaCliente', 'landscape', $('#qtd-eventos').text())
    })
    carregarTabela()

})

function carregarTabela() {
    
    $.ajax({
        url: '/ListarClientesCarregarClientes',
        method: 'Post',
        data: JSON.stringify({idFranqueado: localStorage.getItem('idFranqueado')})
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        const qtd = r.dados.length
        $('tbody').empty()
        $('thead').empty()

        let ativos = 0
        let suspenso = 0
        r.dados.forEach(i => {
            if (i.ativo == 'S') {
                ativos++
            } else {
                suspenso++
            }
        })
        $('thead').append(`
            <tr>
                <td colspan="5" class="fs-5 fw-bold fst-italic"> 
                    ${ativos} Ativos e ${suspenso} Suspenso. Total de ${qtd} Clientes 
                </td>
            </tr>
            <tr>
                
                <th style="width: 33%;">Nome</th>
                <th style="width: 12%;">Telefone1</th>
                <th style="width: 12%;">Telefone2</th>
                <th style="width: 33%;">Email</th>
                <th>Status</th>
            </tr>
        `)
        r.dados.forEach(i => {
            const ativo = (i.ativo == 'S') ? 'ATIVO' : 'SUSPENSO'
            $('tbody').append(`
                <tr>
                    <td class="text-start">${i.nome}</td>
                    <td>${formatarCelular(i.telefone1)}</td>
                    <td>${formatarCelular(i.telefone2)}</td>
                    <td class="text-start">${i.email1}</td>
                    <td>${ativo}</td>
                </tr>
            `)
        });
    })
}

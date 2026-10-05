$(document).ready(function () {
    carregarTabela()
})

function carregarTabela() {

    $('.carregarBases ').empty()

    $.ajax({
        url: '/ListarClientesAtivos',
        method: 'Post',
        data: JSON.stringify({idFranqueado: localStorage.getItem('idFranqueado')})
    }).fail(function (e) {
        boxErro("Erro ao carregar tabela")
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()
        if (r.status != "Vazio") {
            r.dados.forEach(i => {
                if (i.ativo == 'S'){
                    $('tbody').append(`
                        <tr>                        
                            <td>${i.nome}</td>
                            <td>${i.documento1}</td>
                            <td>${formatarCelular(i.telefone1)}</td>
                            <td>${formatarCelular(i.telefone2)}</td>
                            <td>${i.email1}</td>        
                        </tr>
                    `)
                }
            });
        }else{
            $('tbody').append(`<tr><td colspan="5" class="text-center">LISTA VAZIA</td></tr>`)    
        }
    })
}

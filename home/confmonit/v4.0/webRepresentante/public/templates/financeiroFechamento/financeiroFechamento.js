$(window).on("load", function () {
    carregarFranqueado()
    faturaListar()
})

function faturaListar() {
    $.ajax({
        url: `/faturaListarByCentral`,
        method: 'POST',
        data: JSON.stringify({ idCentral: "1" })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        
        if (r.status == "OK") {
            r.dados.map(item => {
                $("#tabelaA tbody").append(`                
                    <tr>
                        <td>${item.destinoNome}</td>   
                        <td>${item.vencimento}</td>   
                        <td>R$ ${item.valor}</td>   
                        
                        <td onclick="visualizar('${item.idFatura}')" class="btn w-10 bg-green-700 text-white" title="Click aqui para visualizar a fatura">
                            <i class="bi bi-eye-fill"></i>
                        </td>   
                    </tr>   
                `)
            })



        }
    })
}

function limparInsLancamento(){
    $('#inserirLancamento #franqueado').val('0')
    $('#inserirLancamento #descricao').val('')
    $('#inserirLancamento #tipo').val('0')
    $('#inserirLancamento #valor').val('')
    
}
function limparFecharFatura(){
    $('#fecharFatura #nomeFranqueado').val('')
    $('#fecharFatura #vencimento').val('')
    $('#fecharFatura #valor').val('')
    $('#fecharFatura #dataPagamento').val('')
    $('#fecharFatura #tipoPagamento').val('0')
    $('#fecharFatura #valorPagamento').val('')
    }
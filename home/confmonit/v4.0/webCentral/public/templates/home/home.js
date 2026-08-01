$(window).on("load", function () {

    getRepBloqueado()
    //faturaVencidas()
})

function getRepBloqueado() {
    $.ajax({
        url: '/getRepBloqueado',
        method: 'POST'
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        //console.log(r) 
        $('#tabFranqBloqueado tbody').empty()
        if (r.status == "OK"){
            r.dados.map(item => {
                $('#tabFranqBloqueado tbody').append(`
                    <tr>
                        <td>${item.razaoSocial}</td>
                        <td onclick="repHabilitar('${item.idRepresentante}')" class="btn w-10 bg-red-700 text-white" title="Click aqui para habilitar o franqueado">
                            <i class="bi bi-toggle-off"></i>
                        </td>
                    </tr>    
                `)
            })
        }else {
            if (r.status != "Vazio"){
                boxErro(r.status)
            }
        }
    })
}

function repHabilitar(idRepresentante) {
    $.ajax({
        url: '/repHabilitar',
        method: 'POST',
        data: JSON.stringify({
            idRepresentante: idRepresentante,
            ativo: "S",
        })

    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        getRepBloqueado()
    })
}

function faturaVencidas() {
    $.ajax({
        url: `/faturaListarByCentral`,
        method: 'POST',
        data: JSON.stringify({ idCentral: "1" })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {

        if (r.status == "OK") {
            const tmp = new Date()
            r.dados.filter(item => {
                const dtCompara = new Date(tmp.toLocaleDateString('en-US'))
                const dtFatura = new Date(item.vencimento)
                return dtCompara.getTime() > dtFatura.getTime() && item.status == "PENDENTE"
            }).map(item => {
                $("#tabFaturasVencidas tbody").append(`                
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
 
function gravarSenha(id) {
    const senha = $('#novaSenha').val()
    const confirma = $('#confirmaNovaSenha').val()
    if (senha === confirma) {

        $.ajax({
            url: `/usuarioCentralAlterarSenha`,
            method: 'POST',
            data: JSON.stringify({
                idUsuario: sessionStorage.getItem("loginId"),
                senha: senha,
            })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            console.log(r)
            if (r.status == "OK") {
                boxAteradoSucesso()
                $('#novaSenha').val('')
                $('#confirmaNovaSenha').val('')
            } else {
                boxErro(r.status)
            }
        })
    } else {
        alert("Senhas não coecidem")
    }

}
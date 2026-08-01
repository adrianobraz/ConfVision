$(window).on("load", function () {
    //getFranqBloqueado()
    //faturaVencidas()

    $("#btnGravarSenha").on('click', gravarSenha)
})



function gravarSenha(id) {
    const senha = $('#novaSenha').val()
    const confirma = $('#confirmaNovaSenha').val()
    if (senha === confirma) {

        $.ajax({
            url: `/alterarSenha`,
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
            } else {
                boxErro(r.status)
            }
        })
    } else {
        alert("Senhas não coecidem")
    }

}


//////////////////////////////////////////////////////////////////////

function getFranqBloqueado() {

    $.ajax({
        url: '/getFranqBloqueado',
        method: 'POST',
        data: JSON.stringify({ repId: sessionStorage.getItem("loginRepId") })

    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        //console.log(">>>>", r)
        $('#tabFranqBloqueado tbody').empty()
        r.dados.filter(f => f.fraAtivo == 'N').map((m) => {
            $('#tabFranqBloqueado tbody').append(`
                <tr>
                    <td>${m.fraRazao}</td>
                    <td idFra="${m.fraId}" tipo="habilitarFranqueado" class="btn w-10 bg-red-700 text-white" title="Click aqui para habilitar o franqueado">
                        <i class="bi bi-toggle-off"></i>
                    </td>
                </tr>    
            `)
        })

        $('td[tipo=habilitarFranqueado]').on('click', function () {
            const idFra = this.getAttribute("idFra")
            franqHabilitar(idFra)
        })
    })
}

function franqHabilitar(idFranqueado) {
    $.ajax({
        url: '/FranqHabilitar',
        method: 'POST',
        data: JSON.stringify({ fraId: idFranqueado })

    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        getFranqBloqueado()
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


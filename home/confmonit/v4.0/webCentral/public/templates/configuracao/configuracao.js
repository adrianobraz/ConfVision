$(window).on("load", function () {


    sessionStorage.setItem("idContactid", '0')

    $('#ciBtnGravar').on('click', ciGravar)
    $('#ciBtnLimpar').on('click', ciLimpar)
    ciListar()
})

// Funcoes para contacid ======================================================

function ciListar() {
    $.ajax({
        url: `/configuracao/contacidListar`,
        method: 'POST',
        data: JSON.stringify({
            idVinculo: 'CENTRAL',
            idCentralUUID: sessionStorage.getItem('loginIdCentralUUID') || ''
        }),
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        // console.log(r)
        $('tbody').empty()
        if (r.status == 'OK') {
            r.dados.map(item => {
                $('tbody').append(`
                    <tr>
                        <td>${item.codigo}</td>
                        <td>${item.grupo}</td>
                        <td>${item.descricao}</td>
                        <td>${item.nivel}</td>
                        
                        <td 
                            class="bg-blue-700" 
                            title="Edita o contactid"   
                            tipo="ciEditar" 
                            idContactid="${item.idContactid}"
                        >
                            <i class="bi bi-pencil-square text-white"></i>
                        </td>

                        <td 
                            class="bg-red-700" 
                            title="Deleta o contactid" 
                            tipo="ciDeletar" 
                            idContactid="${item.idContactid}" 
                            nome="${item.descricao}"
                        >
                            <i class="bi bi-eraser-fill text-white"></i>
                        </td>
                    </tr>    
                `)
            })
            // Associa os botoes
            $('td[tipo=ciEditar]').on('click', function () {
                const idContactid = this.getAttribute("idContactid")
                ciEditar(idContactid)
            })

            $('td[tipo=ciDeletar]').on('click', function () {
                const idContactid = this.getAttribute("idContactid")
                const nome = this.getAttribute("nome")
                ciDeletar(idContactid, nome)
            })
        } else {
            if (r.status != 'Vazio') {
                boxErro(r.status)
            }
        }
    })
}

function ciEditar(codigo) {
    $.ajax({
        url: `/configuracao/contacidEditar`,
        method: 'POST',
        data: JSON.stringify({ idContactid: codigo })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        // console.log(r)
        if (r.status == 'OK') {
            const d = r.dados
            sessionStorage.setItem("idContactid", codigo)
            $('#ciCodigo').attr('readonly', true)

            $('#ciCodigo').val(d.codigo)
            $('#ciGrupo').val(d.grupo)
            $('#ciNivel').val(d.nivel)
            $('#ciDescricao').val(d.descricao)
        } else {
            boxErro(r.status)
        }
    })
}

function ciDeletar(codigo, nome) {
    boxConfirmarExcluir(nome, () => {
        $.ajax({
            url: `/configuracao/contacidDeletar`,
            method: 'POST',
            data: JSON.stringify({ idContactid: codigo })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            // console.log(r)
            if (r.status == 'OK') {
                const d = r.dados
                ciLimpar()
                ciListar()
                boxSucesso()
            }
        })
    })
}

function ciGravar() {
    if ($('#ciCodigo').val().length < 4) {
        boxAdvertenciaCampoAuto(
            `Código deve ter 4 digitos`,
            '#ciCodigo'
        )
        return
    } else if ($('#ciDescricao').val().length < 4) {
        boxAdvertenciaCampoAuto(
            `Descrição deve ter no minimo 4 digitos`,
            '#ciCodigo'
        )
        return
    }

    (sessionStorage.getItem("idContactid") == "0") ? ciInsere() : ciAltera()
}

function ciInsere() {
    $.ajax({
        url: `/configuracao/contacidInserir`,
        method: 'POST',
        data: JSON.stringify({
            idVinculo: 'CENTRAL',
            codigo: $('#ciCodigo').val(),
            grupo: $('#ciGrupo').val(),
            descricao: $('#ciDescricao').val(),
            nivel: $('#ciNivel').val()
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        // console.log(r)
        if (r.status == 'OK') {
            ciLimpar()
            ciListar()
            boxInseridoSucesso(r.dados)
        } else {
            boxErro(r.status)
        }
    })
}

function ciAltera() {
    
    $.ajax({
        url: `/configuracao/contacidAlterar`,
        method: 'POST',
        data: JSON.stringify({
            idContactid: sessionStorage.getItem("idContactid"),
            codigo: $('#ciCodigo').val(),
            grupo: $('#ciGrupo').val(),
            descricao: $('#ciDescricao').val(),
            nivel: $('#ciNivel').val()
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        // console.log(r)
        if (r.status == 'OK') {
            ciLimpar()
            ciListar()
            boxAteradoSucesso()
        } else {
            boxErro(r.status)
        }
    })
}

function ciLimpar() {
    sessionStorage.setItem("idContactid", 0)
    $('#ciCodigo').attr('readonly', false)

    $('#ciCodigo').val('')
    $('#ciGrupo').val('ALARME')
    $('#ciNivel').val(0)
    $('#ciDescricao').val('')
}
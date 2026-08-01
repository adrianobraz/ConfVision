$(window).on("load", function () {
    $('#btnFabGravar').on('click', fabGravar)
    $('#btnFabLimpar').on('click', fabLimpar)
    sessionStorage.setItem("idFabricante", '0')
    fabListar()
})

function fabListar() {
    $.ajax({
        url: `/configuracao/centralAlarme/fabListar`,
        method: 'POST',
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        $('#tabFabricante tbody').empty()
        if (r.status == "OK" && Array.isArray(r.dados)) {
            r.dados.map(i => {
                const keep = (i.keepVisivel == 'S') ? 'SIM' : 'NÃO'
                const stateAtivo = (i.ativo == "S") ? 'tech-state-on' : 'tech-state-off'
                const iconAtivo = (i.ativo == "N")
                    ? '<i class="bi bi-toggle-off"></i>'
                    : '<i class="bi bi-toggle-on"></i>'

                $('#tabFabricante tbody').append(`
                    <tr class="${i.ativo == 'S' ? 'tech-row-on' : 'tech-row-off'}">
                        <td>${keep}</td>
                        <td title="${i.ativo == 'S' ? 'Ativo' : 'Desativado'}">${i.nome}</td>
                        <td class="btn tech-action-edit" title="Edita o fabricante" tipo="fabEditar" idFab="${i.idFabricante}">
                            <i class="bi bi-pencil-square"></i>
                        </td>
                        <td class="btn tech-action-del" title="Apaga o fabricante" tipo="fabDeletar" idFab="${i.idFabricante}" nome="${i.nome}">
                            <i class="bi bi-eraser-fill"></i>
                        </td>
                        <td class="btn ${stateAtivo}" title="Fabricante: ${i.ativo == 'S' ? 'ATIVO' : 'DESATIVADO'}" tipo="fabAtivar" idFab="${i.idFabricante}">
                            ${iconAtivo}
                        </td>
                    </tr>
                `)
            })

            $('td[tipo=fabEditar]').on('click', function () {
                fabEditar(this.getAttribute("idFab"))
            })
            $('td[tipo=fabDeletar]').on('click', function () {
                fabDeletar(this.getAttribute("idFab"), this.getAttribute("nome"))
            })
            $('td[tipo=fabAtivar]').on('click', function () {
                fabAtivar(this.getAttribute("idFab"))
            })
        } else if (r.status && r.status != "Vazio" && r.status != "OK") {
            boxErro(r.status)
        }
    })
}

function fabEditar(idFab) {
    $.ajax({
        url: `/configuracao/centralAlarme/fabGetDados`,
        method: 'POST',
        data: JSON.stringify({ idFabricante: idFab })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        const d = r.dados
        if (r.status == "OK") {
            sessionStorage.setItem("idFabricante", d.idFabricante)
            $('#fabKeepVisivel').val(d.keepVisivel)
            $('#fabNome').val(d.nome)
        } else {
            boxErro(r.status)
        }
    })
}

function fabDeletar(idFab, nome) {
    boxConfirmarExcluir(nome, () => {
        $.ajax({
            url: `/configuracao/centralAlarme/fabDeletar`,
            method: 'POST',
            data: JSON.stringify({ idFabricante: idFab })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == "OK") {
                fabLimpar()
                fabListar()
                boxDeletadoSucesso()
            } else {
                boxErro(r.status)
            }
        })
    })
}

function fabAtivar(idFab) {
    $.ajax({
        url: `/configuracao/centralAlarme/fabAtivar`,
        method: 'POST',
        data: JSON.stringify({ idFabricante: idFab })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            fabLimpar()
            fabListar()
        } else {
            boxErro(r.status)
        }
    })
}

function fabGravar() {
    if ($('#fabNome').val() != "") {
        (sessionStorage.getItem("idFabricante") == '0') ? fabInserir() : fabAlterar()
    } else {
        boxAdvertenciaCampoAuto('Um nome para o fabricante deve ser informado', '#fabNome')
    }
}

function fabInserir() {
    $.ajax({
        url: `/configuracao/centralAlarme/fabInsere`,
        method: 'POST',
        data: JSON.stringify({
            keepVisivel: $('#fabKeepVisivel').val(),
            nome: $('#fabNome').val(),
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            fabLimpar()
            fabListar()
            boxInseridoSucesso(r.dados)
        } else {
            boxErro(r.status)
        }
    })
}

function fabAlterar() {
    $.ajax({
        url: `/configuracao/centralAlarme/fabAltera`,
        method: 'POST',
        data: JSON.stringify({
            idFabricante: sessionStorage.getItem("idFabricante"),
            keepVisivel: $('#fabKeepVisivel').val(),
            nome: $('#fabNome').val(),
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            fabLimpar()
            fabListar()
            boxAteradoSucesso()
        } else {
            boxErro(r.status)
        }
    })
}

function fabLimpar() {
    sessionStorage.setItem("idFabricante", '0')
    $('#fabKeepVisivel').val('N')
    $('#fabNome').val('')
}

$(window).on("load", function () {

    $('#btnFabGravar').on('click', fabGravar)
    $('#btnFabLimpar').on('click', fabLimpar)
    $('#btnModGravar').on('click', modGravar)
    $('#btnModLimpar').on('click', modLimpar)
    sessionStorage.setItem("idFabricante", '0')
    sessionStorage.setItem("idModelo", '0')
    fabListar()
    modListarId()
})

// Funcoes pra gerenciar fabricante =====================================================
function fabListar() {

    $.ajax({
        url: `/configuracao/centralAlarme/fabListar`,
        method: 'POST',
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        $('#tabFabricante tbody').empty()
        if (r.status = "OK") {
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

            // Associa os botoes
            $('td[tipo=fabEditar]').on('click', function () {
                const idFab = this.getAttribute("idFab")
                fabEditar(idFab)
            })

            $('td[tipo=fabDeletar]').on('click', function () {
                const idFab = this.getAttribute("idFab")
                const nome = this.getAttribute("nome")
                fabDeletar(idFab, nome)
            })

            $('td[tipo=fabAtivar]').on('click', function () {                
                const idFab = this.getAttribute("idFab")
                fabAtivar(idFab)
            })

        } else {
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
        console.log(r)
        const d = r.dados
        if (r.status = "OK") {
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
            data: JSON.stringify({ idFabricante: idFab, })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            console.log(r)
            const d = r.dados
            if (r.status = "OK") {
                fabLimpar()
                fabListar()
                modListarId()
                modListar('0')
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
        data: JSON.stringify({ idFabricante: idFab})
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        // console.log(r)
        const d = r.dados
        if (r.status = "OK") {
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
        boxAdvertenciaCampoAuto(
            'Um nome para o fabricante deve ser informado',
            '#fabNome'
        )
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
        console.log(r)
        const d = r.dados
        if (r.status = "OK") {
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
        console.log(r)
        const d = r.dados
        if (r.status = "OK") {
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

    $('#fabKeepVisivel').val('')
    $('#fabNome').val('')
}

// Funcoes para gerenciar modelos =======================================================


function modListarId() {

    $.ajax({
        url: `/configuracao/centralAlarme/fabListar`,
        method: 'POST',
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)

        if (r.status == "OK") {
            $('#modIdFabricante').empty().append(`<option value="0">SELECIONE</option>`)
            if (r.status = "OK") {

                r.dados.map(i => {
                    $('#modIdFabricante').append(`
                        <option value="${i.idFabricante}">${i.nome}</option>
                    `)
                })

                $('#modIdFabricante').on('change', () => {
                    modListar($('#modIdFabricante').val())
                })
            }

        } else {
            boxErro(r.status)
        }

    })
}

function modListar(idFab) {
    modLimpar()
    $.ajax({
        url: `/configuracao/centralAlarme/modListar`,
        method: 'POST',
        data: JSON.stringify({ idFabricante: idFab })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        // console.log(r)

        $('#tabModelo tbody').empty()
        if (r.status == "OK") {
            r.dados.map(item => {
                const eletr = (item.eletrificador == 'S') ? "SIM" : "NÃO"
                const stateAtivo = (item.ativo == "S") ? 'tech-state-on' : 'tech-state-off'
                const iconAtivo = (item.ativo == "N")
                    ? '<i class="bi bi-toggle-off"></i>'
                    : '<i class="bi bi-toggle-on"></i>'

                $('#tabModelo tbody').append(`
                    <tr class="${item.ativo == 'S' ? 'tech-row-on' : 'tech-row-off'}">
                        <td title="${item.ativo == 'S' ? 'Ativo' : 'Desativado'}">${item.nome}</td>
                        <td>${item.qtdParticoes}</td>
                        <td>${item.qtdZona}</td>
                        <td>${item.qtdPgm}</td>
                        <td>${eletr}</td>
                        <td class="btn tech-action-edit" title="Edita o modelo" tipo="modEditar" idMod="${item.idModelo}">
                            <i class="bi bi-pencil-square"></i>
                        </td>
                        <td class="btn tech-action-del" title="Apaga o modelo" tipo="modDeletar" idMod="${item.idModelo}" nome="${item.nome}">
                            <i class="bi bi-eraser-fill"></i>
                        </td>
                        <td class="btn ${stateAtivo}" title="Modelo: ${item.ativo == 'S' ? 'ATIVO' : 'DESATIVADO'}" tipo="modAtivar" idMod="${item.idModelo}">
                            ${iconAtivo}
                        </td>
                    </tr>      
                `)
            })
            // Associa os botoes
            $('td[tipo=modEditar]').on('click', function () {
                const idMod = this.getAttribute("idMod")
                modEditar(idMod)
            })

            $('td[tipo=modDeletar]').on('click', function () {
                const idMod = this.getAttribute("idMod")
                const nome = this.getAttribute("nome")
                modDeletar(idMod, nome)
            })

            $('td[tipo=modAtivar]').on('click', function () {
                const idMod = this.getAttribute("idMod")
                modAtivar(idMod)
            })
        } else {
            if (r.status != 'Vazio') boxErro(r.status)
        }

    })
}

function modEditar(idMod) {
    $.ajax({
        url: `/configuracao/centralAlarme/modGetDados`,
        method: 'POST',
        data: JSON.stringify({ idModelo: idMod })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        const d = r.dados
        if (r.status = "OK") {
            const d = r.dados
            sessionStorage.setItem("idModelo", d.idModelo)

            $('#modIdFabricante').val(d.idFabricante)
            $('#modNome').val(d.nome)
            $('#modQtdParticoes').val(d.qtdParticoes)
            $('#modQtdZonas').val(d.qtdZona)
            $('#modQtdPgm').val(d.qtdPgm)
            $('#modEletrificador').val(d.eletrificador)
        } else {
            boxErro(r.status)
        }
    })
}

function modDeletar(idMod, nome) {
    boxConfirmarExcluir(nome, () => {
        $.ajax({
            url: `/configuracao/centralAlarme/modDeletar`,
            method: 'POST',
            data: JSON.stringify({ idModelo: idMod })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == "OK") {
                modLimpar()
                modListar($('#modIdFabricante').val())
                boxSucesso(r.dados)
            } else {
                boxErro(r.status)
            }
        })
    })
}

function modAtivar(idMod) {
    $.ajax({
        url: `/configuracao/centralAlarme/modAtivar`,
        method: 'POST',
        data: JSON.stringify({ idModelo: idMod })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            modLimpar()
            modListar($('#modIdFabricante').val())
        } else {
            boxErro(r.status)
        }
    })
}

function modGravar() {
    if (modValidaCampos()) {
        (sessionStorage.getItem("idModelo") != '0') ? modAlterar() : modInserir()
    }
}

function modAlterar() {
    $.ajax({
        url: `/configuracao/centralAlarme/modAlterar`,
        method: 'POST',
        data: JSON.stringify({
            idModelo: sessionStorage.getItem("idModelo"),
            nome: $('#modNome').val(),
            qtdParticoes: $('#modQtdParticoes').val(),
            qtdZona: $('#modQtdZonas').val(),
            qtdPgm: $('#modQtdPgm').val(),
            eletrificador: $('#modEletrificador').val(),
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            modLimpar()
            modListar($('#modIdFabricante').val())
            boxAteradoSucesso(r.dados)
        } else {
            boxErro(r.status)
        }
    })
}

function modInserir() {

    $.ajax({
        url: `/configuracao/centralAlarme/modInserir`,
        method: 'POST',
        data: JSON.stringify({
            idFabricante: $('#modIdFabricante').val(),
            nome: $('#modNome').val(),
            qtdParticoes: $('#modQtdParticoes').val(),
            qtdZona: $('#modQtdZonas').val(),
            qtdPgm: $('#modQtdPgm').val(),
            eletrificador: $('#modEletrificador').val(),
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == "OK") {
            modLimpar()
            modListar($('#modIdFabricante').val())
            boxInseridoSucesso(r.dados)
        } else {
            boxErro(r.status)
        }
    })
}

function modLimpar() {
    sessionStorage.setItem("idModelo", '0')

    $('#modNome').val('')
    $('#modQtdParticoes').val('')
    $('#modQtdZonas').val('')
    $('#modQtdPgm').val('')
    $('#modEletrificador').val('N')
}

function modValidaCampos() {
    if ($('#modIdFabricante').val() == '0') {
        boxAdvertenciaCampoAuto(
            'Um fabricante deve ser selecionado',
            '#modIdFabricante'
        )
        return false
    } else if ($('#modNome').val() == "") {
        boxAdvertenciaCampoAuto(
            'Um nome deve ser informado',
            '#modNome'
        )
        return false
    } else if ($('#modQtdParticoes').val() == "") {
        boxAdvertenciaCampoAuto(
            'Uma quantidade de partições deve ser informado',
            '#modQtdParticoes'
        )
        return false
    } else if ($('#modQtdZonas').val() == "") {
        boxAdvertenciaCampoAuto(
            'Uma quantidade de zonas deve ser informado',
            '#modQtdZonas'
        )
        return false
    } else if ($('#modQtdPgm').val() == "") {
        boxAdvertenciaCampoAuto(
            'Uma quantidade de pgm deve ser informado',
            '#modQtdPgm'
        )
        return false
    }
    return true
}
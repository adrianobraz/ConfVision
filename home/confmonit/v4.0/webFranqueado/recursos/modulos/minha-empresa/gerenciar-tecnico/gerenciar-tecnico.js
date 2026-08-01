$(document).ready(function () {

    // Regulariza a altura da tela
    $(window).trigger('resize');
    GerenciarTecnicoAlturaFixa()

    $('#btn-limpar').on('click', GerenciarTecnicoLimparFormulario)
    $('#btn-gravar').on('click', GerenciarTecnicoGravarDados)

    // Carrega os tecnicos cadastrado
    GerenciarTecnicoCarregarTabela()

    // Cria uma variavel 
    localStorage.setItem('idTecnico', '')
})

// Redimenciona a tela 
$(window).resize(GerenciarTecnicoAlturaFixa)

function GerenciarTecnicoAlturaFixa() {
    $('.altura-fixa').height(window.innerHeight - 172)
}

function GerenciarTecnicoLimparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    localStorage.setItem('idTecnico', '')
}

function GerenciarTecnicoGravarDados() {
    if (localStorage.getItem('idTecnico') == "") {
        GerenciarTecnicoInserir()
    } else {
        GerenciarTecnicoAlterar()
    }
}

function GerenciarTecnicoCarregarTabela() {
    const idFranqueado = localStorage.getItem('idFranqueado')
    const url = `/TecnicoCarregarTabela`
    $.ajax({
        url: url,
        method: 'Post',
        data: JSON.stringify({ idVinculo: idFranqueado })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        $('tbody').empty()
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                GerenciarTecnicoMontaLinha(i)
            });
            GerenciarTecnicoAssociaBotoesTabela()
        }
    })
}

function GerenciarTecnicoMontaLinha(i) {
    const cel = formatarCelular(i.telefone1)

    let btnAtivoImg
    let btnAtivoCor
    let btnAtivoTipo
    if (i.ativo == 'S') {
        btnAtivoImg = '<i class="bi bi-toggle-on"></i>'
        btnAtivoCor = 'btn-success'
        btnAtivoTipo = 'habilitar'
    } else {
        btnAtivoImg = '<i class="bi bi-toggle-off"></i>'
         btnAtivoCor = 'btn-secondary'
        btnAtivoTipo = 'habilitar'
    }



    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.nome}</td>
        <td>${cel}</td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idTecnico}
                tipo="editar"
                title="Editar dados da Técnico">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idTecnico}
                tipo="resetar"
                title="Resetar senha do Técnico">
                <i class="bi bi-key-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idTecnico}
                tipo="excluir"
                title="Exclui o Técnico">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm ${btnAtivoCor} py-0"
                id=${i.idTecnico}
                tipo="${btnAtivoTipo}"
                title="Habilita/Desabilita o
                Cliente">
                ${btnAtivoImg}
            </button>
        </td>
    </tr>
    `)
}

function GerenciarTecnicoAssociaBotoesTabela() {
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarTecnicoBuscar(id)
    })

    $('button[tipo=resetar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarTecnicoResetarSenha(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarTecnicoExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarTecnicoHabilitar(id)
    })
}

function GerenciarTecnicoResetarSenha(id) {
    const url = `/TecnicoResetarSenha`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({
            idTecnico: id
        })
    }).fail(function (e) {
        boxErro("Erro ao resetar a senha do tecnico")
    }).done(function (r) {
        boxFechar()
        boxSenhaAlterada('usuario123')

    })
}

function GerenciarTecnicoHabilitar(idTecnico) {
    const url = `/TecnicoHabilitar`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'Post',
        data: JSON.stringify({ idTecnico: idTecnico })
    }).fail(function (e) {
        console.log(e)
        boxErro('Erro ao habilitar/Desabilitar o Técnico')
    }).done(function (r) {
        boxFechar()
        GerenciarTecnicoCarregarTabela()
    })
}

function GerenciarTecnicoInserir() {

    if (GerenciarTecnicoValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        stop: boxFechar(),
        url: `/TecnicoInserir`,
        method: 'Post',
        data: JSON.stringify(
            {
                idVinculo: localStorage.getItem('idFranqueado'),
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                cpf: limpaDocumento($('#cpf').val()),
                rg: $('#rg').val(),
                cep: $('#cep').val(),
                uf: $('#uf').val(),
                endereco: $('#endereco').val(),
                complemento: $('#complemento').val(),
                telefone1: limpaDocumento($('#telefone1').val()),
                bairro: $('#bairro').val(),
                telefone2: limpaDocumento($('#telefone2').val()),
                cidade: $('#cidade').val(),
                email: $('#email').val()
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o Técnico")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarTecnicoLimparFormulario()
        GerenciarTecnicoCarregarTabela()
    })
}

function GerenciarTecnicoBuscar(idTecnico) {
    const url = `/TecnicoBuscar`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'Post',
        data: JSON.stringify({
            idTecnico: idTecnico,
        })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao buscar dados do Técnico")
    }).done(function (r) {
        console.log(r)
        boxFechar()
        const d = r.dados

        const cpf = ('cpf' in d) ? formatarDocumento(d.cpf) : ""
        const telefone1 = ('telefone1' in d) ? formatarCelular(d.telefone1) : ""
        const telefone2 = ('telefone2' in d) ? formatarCelular(d.telefone2) : ""


        localStorage.setItem('idTecnico', d.idTecnico)
        $('#nome').val(d.nome)
        $('#nick').val(d.nick)
        $('#cpf').val(cpf)
        $('#rg').val(d.rg)
        $('#cep').val(d.cep)
        $('#uf').val(d.uf)
        $('#endereco').val(d.endereco)
        $('#complemento').val(d.complemento)
        $('#telefone1').val(telefone1)
        $('#bairro').val(d.bairro)
        $('#telefone2').val(telefone2)
        $('#cidade').val(d.cidade)
        $('#email').val(d.email)
    })
}

function GerenciarTecnicoAlterar() {
    if (GerenciarTecnicoValidarCamposBranco()) return
    const idTecnico = localStorage.getItem('idTecnico')
    const url = `/TecnicoAlterar`

    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'Post',
        data: JSON.stringify(
            {
                idTecnico: idTecnico,
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                cpf: limpaDocumento($('#cpf').val()),
                rg: $('#rg').val(),
                cep: $('#cep').val(),
                uf: $('#uf').val(),
                endereco: $('#endereco').val(),
                complemento: $('#complemento').val(),
                telefone1: limpaDocumento($('#telefone1').val()),
                bairro: $('#bairro').val(),
                telefone2: limpaDocumento($('#telefone2').val()),
                cidade: $('#cidade').val(),
                email: $('#email').val(),

            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao alterar dados do Técnico")
    }).done(function (r) {
        boxAteradoSucesso()
        GerenciarTecnicoLimparFormulario()
        GerenciarTecnicoCarregarTabela()
    })

}

function GerenciarTecnicoExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Técnico?`,
        text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode apagar!'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                start: boxProcessando(),
                url: `/TecnicoExcluir`,
                method: 'Post',
                data: JSON.stringify({ idTecnico: id }
                )
            }).fail(function (e) {
                console.log(e)
                boxErro('Erro ao excluir o técnico')
            }).done(function (r) {
                GerenciarTecnicoLimparFormulario()
                GerenciarTecnicoCarregarTabela()
                boxDeletadoSucesso()
            })
        }
    })
}

function GerenciarTecnicoValidarCamposBranco() {
    if ($('#nome').val() == "") {
        boxAdvertenciaAuto("O campo Nome, não poder ficar em branco")
        $('#nome').focus()
        return true
    }
    if ($('#nick').val() == "") {
        boxAdvertenciaAuto("O campo Nick, não poder ficar em branco")
        $('#nick').focus()
        return true
    }
    if ($('#telefone1').val() == "") {
        boxAdvertenciaAuto("O campo Celular, não poder ficar em branco")
        $('#telefone1').focus()
        return true
    }
    if ($('#email').val() == "") {
        boxAdvertenciaAuto("O campo Email, não poder ficar em branco")
        $('#email').focus()
        return true
    }
    return false
}

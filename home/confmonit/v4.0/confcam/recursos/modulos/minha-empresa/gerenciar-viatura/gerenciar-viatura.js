$(document).ready(function () {

    // Regulariza a altura da tela
    $(window).trigger('resize')
    GerenciarViaturaAlturaFixa()

    localStorage.setItem('idViatura', '')

    $('#btn-limpar').on('click', GerenciarViaturaLimparFormulario)
    $('#btn-gravar').on('click', GerenciarViaturaGravarDados)

    GerenciarViaturaCarregarTabela()
})

// Redimenciona a tela 
$(window).resize(GerenciarViaturaAlturaFixa)

function GerenciarViaturaAlturaFixa() {
    $('.altura-fixa').height(window.innerHeight - 172)
}

// GerenciarViaturaLimparFormulario limpa o formulario e zera a variavel idViatura
function GerenciarViaturaLimparFormulario() {

    $('#formulario').each(function () {
        this.reset();
    })
    localStorage.setItem('idViatura', '')
}

function GerenciarViaturaGravarDados() {

    if (localStorage.getItem('idViatura') == "") {
        GerenciarViaturaInserir()
    } else {
        GerenciarViaturaAlterar()
    }
}

// GerenciarViaturaCarregarTabela carrega a tabela com os dados da viaturas
function GerenciarViaturaCarregarTabela() {

    $.ajax({
        url: `/ViaturaCarregarTabela`,
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        $('tbody').empty()
        if (r.status != "Vazio") {
            r.dados.forEach(i => {
                GerenciarViaturaMontaLinha(i)
            })
            GerenciarViaturaAssociaBotoesTabela()
        }
    })

}

// GerenciarViaturaMontaLinha monta a linhas da tabela
function GerenciarViaturaMontaLinha(i) {


    const btnImg = (i.ativo == "S") ?
        '<i class="bi bi-toggle-on"></i>' :
        '<i class="bi bi-toggle-off"></i>'

    const btnCor = (i.ativo == "S") ? 'btn-success' : 'btn-secondary'

    $('tbody').append(`
        <tr>
        <td class="align-middle text-uppercase">${i.nick}</td>
        <td class="align-middle">${i.modelo}</td>
        <td class="align-middle">${i.placa}</td>
        <td class="align-middle">${i.cor}</td>
        <td class="align-middle">${formatarCelular(i.telefone1)}</td>

        <td class="align-middle">
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idViatura}
                tipo="editar"
                title="Editar dados da Técnico">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>

        <!--td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idViatura}
                tipo="resetar"
                title="Resetar senha do Técnico">
                <i class="bi bi-key-fill"></i>
            </button>
        </td-->

        <td class="align-middle">
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idViatura}
                tipo="excluir"
                title="Exclui o Técnico">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>

        <td class="align-middle">
            <button class="btn btn-sm ${btnCor} py-0"
                id=${i.idViatura}
                tipo="habilitar"
                title="Habilita/Desabilita o
                Cliente">
                ${btnImg}
            </button>
        </td>

    </tr>
    `)
}

// GerenciarViaturaAssociaBotoesTabela associa os botoes da tabela
function GerenciarViaturaAssociaBotoesTabela() {

    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarViaturaBuscar(id)
    })

    $('button[tipo=resetar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarViaturaResetarSenha(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarViaturaExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarViaturaHabilitar(id)
    })
}

// GerenciarViaturaHabilitar habilita ou desabilita uma viatura
function GerenciarViaturaHabilitar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/ViaturaHabilitar',
        method: 'Post',
        data: JSON.stringify({ idViatura: id })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao habilitar/desabilitar a viatura")
    }).done(function (r) {
        boxFechar()
        GerenciarViaturaCarregarTabela()
    })
}

// GerenciarViaturaInserir insere uma nova viatura
function GerenciarViaturaInserir() {

    if (GerenciarViaturaValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/ViaturaInserir',
        method: 'Post',
        data: JSON.stringify(
            {
                idFranqueado: localStorage.getItem('idFranqueado'),
                nome: $('#nome').val().toUpperCase(),
                nick: $('#nick').val(),
                cpf: limpaDocumento($('#cpf').val()),
                rg: $('#rg').val(),
                cep: $('#cep').val(),
                uf: $('#uf').val(),
                endereco: $('#endereco').val(),
                complemento: $('#complemento').val(),
                bairro: $('#bairro').val(),
                cidade: $('#cidade').val(),
                modelo: $('#modelo').val(),
                placa: $('#placa').val(),
                cor: $('#cor').val(),
                telefone1: formatarCelular($('#telefone1').val()),
                telefone2: formatarCelular($('#telefone2').val()),
                telefone3: formatarCelular($('#telefone3').val()),
                email: $('#email').val().toLowerCase(),
                observacao: $('#observacao').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir a Viatura")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarViaturaLimparFormulario()
        GerenciarViaturaCarregarTabela()
    })
}

function GerenciarViaturaAlterar() {
    if (GerenciarViaturaValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/ViaturaAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idViatura: localStorage.getItem('idViatura'),
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                cpf: limpaDocumento($('#cpf').val()),
                rg: $('#rg').val(),
                cep: $('#cep').val(),
                uf: $('#uf').val().toUpperCase(),
                endereco: $('#endereco').val(),
                complemento: $('#complemento').val(),
                bairro: $('#bairro').val(),
                cidade: $('#cidade').val(),
                modelo: $('#modelo').val(),
                placa: $('#placa').val(),
                cor: $('#cor').val(),
                telefone1: formatarCelular($('#telefone1').val()),
                telefone2: formatarCelular($('#telefone2').val()),
                telefone3: formatarCelular($('#telefone3').val()),
                email: $('#email').val(),
                observacao: $('#observacao').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro a Alterar os dados da Viatura")
    }).done(function (r) {
        boxAteradoSucesso()
        GerenciarViaturaLimparFormulario()
        GerenciarViaturaCarregarTabela()
    })
}

function GerenciarViaturaBuscar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/ViaturaBuscar',
        method: 'Post',
        data: JSON.stringify({ idViatura: id }
        )
    }).fail(function (e) {
        boxErro('Erro ao buscar dados da viatura')
    }).done(function (r) {
        boxFechar()
        const d = r.dados
        console.log('>>>', d)
        localStorage.setItem('idViatura', d.idViatura)

        $('#nome').val(d.nome)
        $('#nick').val(d.nick)
        $('#cpf').val(formatarDocumento(d.cpf))
        $('#rg').val(d.rg)
        $('#cep').val(d.cep)
        $('#uf').val(d.uf)
        $('#endereco').val(d.endereco)
        $('#complemento').val(d.complemento),
            $('#bairro').val(d.bairro)
        $('#cidade').val(d.cidade)
        $('#modelo').val(d.modelo)
        $('#placa').val(d.placa)
        $('#cor').val(d.cor)
        $('#telefone1').val(formatarCelular(d.telefone1))
        $('#telefone2').val(formatarCelular(d.telefone2))
        $('#telefone3').val(formatarCelular(d.telefone3))
        $('#email').val(d.email)
        $('#observacao').val(d.observacao)

    })
}

//
function GerenciarViaturaExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar esta Viatura?`,
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
                url: `/ViaturaExcluir`,
                method: 'Post',
                data: JSON.stringify({ idViatura: id }
                )
            }).fail(function (e) {
                boxErro(e)
            }).done(function (r) {
                GerenciarViaturaLimparFormulario()
                GerenciarViaturaCarregarTabela()
                boxDeletadoSucesso()
            })
        }
    })
}

function GerenciarViaturaValidarCamposBranco() {
    if ($('#nome').val() == "") {
        boxAdvertenciaCampoAuto("O campo Nome, não poder ficar em branco", '#nome')
        return true
    }
    if ($('#nick-nick').val() == "") {
        boxAdvertencboxAdvertenciaCampoAutoiaAuto("O campo Nick, não poder ficar em branco", '#nick-nick')
        return true
    }
    if ($('#telefone1').val() == "") {
        boxAdvertenciaAuto("O campo telefone 1, não poder ficar em branco", '#telefone1')
        return true
    }
    if ($('#email').val() == "") {
        boxAdvertenciaCampoAuto("O campo Email, não poder ficar em branco", '#email')
        return true
    }
    if ($('#placa').val() == "") {
        boxAdvertenciaCampoAuto("O campo Placa, não poder ficar em branco", '#placa')
        return true
    }
    return false
}


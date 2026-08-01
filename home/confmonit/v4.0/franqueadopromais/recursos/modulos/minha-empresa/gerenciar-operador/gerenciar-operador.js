$(document).ready(function () {

    $('#btn-limpar').on('click', GerenciarOperadorLimparFormulario)
    $('#btn-gravar').on('click', GerenciarOperadorGravarDados)

    localStorage.setItem('idOperador', '')

    GerenciarOperadorCarregarTabela()
})

function GerenciarOperadorLimparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    localStorage.setItem('idOperador', '')
}

function GerenciarOperadorGravarDados() {
    const id = localStorage.getItem('idOperador')

    if (id == "") {
        GerenciarOperadorInserir()
    } else {
        GerenciarOperadorAlterar(id)
    }
}

function GerenciarOperadorCarregarTabela() {

    $.ajax({
        url: 'OperadorCarregarTabela',
        method: 'Post',
        data: JSON.stringify({ idVinculo: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro('Erro ao carregar tabela')
    }).done(function (r) {
        $('tbody').empty()
        if (r.status != 'Vazio') {

            r.dados.forEach(i => {
                GerenciarOperadorMontaLinha(i)
            });
            GerenciarOperadorAssociaBotoesTabela()
        }
    })
}

function GerenciarOperadorMontaLinha(i) {
    const cel = formatarCelular(i.celular)
    const btnImg = (i.ativo == 1) ?
        '<i class="bi bi-toggle-on"></i>' :
        '<i class="bi bi-toggle-off"></i>'

    const btnCor = (i.ativo == 1) ? 'btn-success' : 'btn-secondary'
    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.nome}</td>
        <td>${cel}</td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idOperador}
                tipo="editar"
                title="Editar dados da Operador">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idOperador}
                tipo="resetar"
                title="Resetar senha do Operador"
            >
                <i class="bi bi-key-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idOperador}
                tipo="excluir"
                title="Exclui o Operador"
            >
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm ${btnCor} py-0"
                id=${i.idOperador}
                tipo="habilitar"
                title="Habilita/Desabilita o Operador"
            >
                ${btnImg}
            </button>
        </td>
    </tr>
    `)
}

function GerenciarOperadorAssociaBotoesTabela() {
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarOperadorBuscar(id)
    })

    $('button[tipo=resetar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarOperadorResetarSenha(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarOperadorExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarOperadorHabilitar(id)
    })
}

function GerenciarOperadorResetarSenha(id) {

    $.ajax({
        start: boxProcessando(),
        url: 'OperadorResetarSenha',
        method: 'POST',
        data: JSON.stringify({ idOperador: id })
    }).fail(function (e) {
        boxErro("Erro ao resetar a senha do Operador")
    }).done(function (r) {
        boxFechar()
        boxSenhaAlterada('usuario123')

    })
}

function GerenciarOperadorHabilitar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/OperadorHabilitar',
        method: 'Post',
        data: JSON.stringify({ idOperador: id })
    }).fail(function (e) {
        boxErro("Erro ao Habilitar/Desabilitar Operador")
    }).done(function (r) {
        boxFechar()
        GerenciarOperadorCarregarTabela()
    })
}

function GerenciarOperadorInserir() {

    if (GerenciarOperadorValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/OperadorInserir',
        method: 'Post',
        data: JSON.stringify(
            {
                idVinculo: localStorage.getItem('idFranqueado'),
                nome: $('#nome-operador').val(),
                usuarioVoip: $('#user-voip-operador').val(),
                nick: $('#nick-operador').val(),
                cpf: formatarDocumento($('#cpf-operador').val()),
                rg: $('#rg-operador').val(),
                cep: $('#cep-operador').val(),
                uf: $('#uf-operador').val(),
                endereco: $('#endereco-operador').val(),
                telefone: formatarCelular($('#telefone-operador').val()),
                bairro: $('#bairro-operador').val(),
                celular: formatarCelular($('#celular-operador').val()),
                cidade: $('#cidade-operador').val(),
                email: $('#email-operador').val(),
                master: '0'
            }
        )
    }).fail(function (e) {
        boxErro("Erro ao inserir o Operador")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarOperadorLimparFormulario()
        GerenciarOperadorCarregarTabela()
    })
}

function GerenciarOperadorBuscar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/OperadorBuscar',
        method: 'Post',
        data: JSON.stringify({ idOperador: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Operador")
    }).done(function (r) {
        boxFechar()
        const d = r.dados
        localStorage.setItem('idOperador', d.idOperador)

        cpf = ('cpf' in d) ? formatarDocumento(d.cpf) : ''
        telefone = ('telefone' in d) ? formatarCelular(d.telefone) : ''
        celular = ('celular' in d) ? formatarCelular(d.celular) : ''
        
        $('#nome-operador').val(d.nome)
        $('#user-voip-operador').val(d.usuarioVoip)
        $('#nick-operador').val(d.nick)
        $('#cpf-operador').val(cpf)
        $('#rg-operador').val(d.rg)
        $('#cep-operador').val(d.cep)
        $('#uf-operador').val(d.uf)
        $('#endereco-operador').val(d.endereco)
        $('#telefone-operador').val(telefone)
        $('#bairro-operador').val(d.bairro)
        $('#celular-operador').val(celular)
        $('#cidade-operador').val(d.cidade)
        $('#email-operador').val(d.email)
    })
}

function GerenciarOperadorAlterar(id) {
    if (GerenciarOperadorValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/OperadorAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idOperador: id,
                nome: $('#nome-operador').val(),
                usuarioVoip: $('#user-voip-operador').val(),
                nick: $('#nick-operador').val(),
                cpf: formatarDocumento($('#cpf-operador').val()),
                rg: $('#rg-operador').val(),
                cep: $('#cep-operador').val(),
                uf: $('#uf-operador').val(),
                endereco: $('#endereco-operador').val(),
                telefone: formatarCelular($('#telefone-operador').val()),
                bairro: $('#bairro-operador').val(),
                celular: formatarCelular($('#celular-operador').val()),
                cidade: $('#cidade-operador').val(),
                email: $('#email-operador').val(),

            }
        )
    }).fail(function (e) {
        boxErro("Erro ao alterar dados do Operador")
    }).done(function (r) {
        boxAteradoSucesso()
        GerenciarOperadorLimparFormulario()
        GerenciarOperadorCarregarTabela()
    })

}

function GerenciarOperadorExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Operador?`,
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
                url: `/OperadorExcluir`,
                method: 'Post',
                data: JSON.stringify({ idOperador: id }
                )
            }).fail(function (e) {
                boxErro('Erro ao excluir o Operador')
            }).done(function (r) {
                GerenciarOperadorLimparFormulario()
                GerenciarOperadorCarregarTabela()
                boxDeletadoSucesso()
            })
        }
    })
}

function GerenciarOperadorValidarCamposBranco() {
    if ($('#nome-operador').val() == "") {
        boxAdvertenciaAuto("O campo Nome, não poder ficar em branco")
        $('#nome-operador').focus()
        return true
    }
    if ($('#nick-operador').val() == "") {
        boxAdvertenciaAuto("O campo Nick, não poder ficar em branco")
        $('#nick-operador').focus()
        return true
    }
    if ($('#celular-operador').val() == "") {
        boxAdvertenciaAuto("O campo Celular, não poder ficar em branco")
        $('#celular-operador').focus()
        return true
    }
    if ($('#email-operador').val() == "") {
        boxAdvertenciaAuto("O campo Email, não poder ficar em branco")
        $('#email-operador').focus()
        return true
    }
    return false
}

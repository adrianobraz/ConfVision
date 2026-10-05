$(document).ready(function () {
    localStorage.setItem('idProcedimento', '')
    ajustaTabela()
    $(window).on('resize', function () {
        ajustaTabela()
    })

    $('#btn-limpar').on('click', GerenciarProcedimentoAtendimentoLimparFormulario)
    $('#btn-gravar').on('click', GerenciarProcedimentoAtendimentoGravarDados)
    $('#grupo-evento-procedimento').on('change', () => {
        $('#descricao-procedimento').val('')
    })

    GerenciarProcedimentoAtendimentoListar()
    GerenciarProcedimentoAtendimentoClientesListar()
    GerenciarProcedimentoAtendimentoGrupoEventoListar()
})

// OK
function GerenciarProcedimentoAtendimentoLimparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    localStorage.setItem('idProcedimento', '')
}

// OK
function GerenciarProcedimentoAtendimentoGravarDados() {
    const idProcedimento = localStorage.getItem('idProcedimento')
    if (idProcedimento == "") {
        GerenciarProcedimentoAtendimentoInserir()
    } else {
        GerenciarProcedimentoAtendimentoAlterar(idProcedimento)
    }
}

// OK
function GerenciarProcedimentoAtendimentoClientesListar() {

    $('#modalidade-evento').empty()

    $.ajax({
        url: '/GerenciarProcedimentoAtendimentoClienteListar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro("Erro ao carregar clientes")
    }).done(function (r) {
        $('#id-cliente').append(`
            <option value="TODOS">PARA TODOS OS CLIENTE</option>
        `)
        r.dados.forEach(i => {
            $('#id-cliente').append(`
                <option value="${i.idCliente}">${i.nome}</option>
            `)
        });
    })
}

// OK
function GerenciarProcedimentoAtendimentoListar() {

    $.ajax({
        url: '/GerenciarProcedimentoAtendimentoListar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        console.log(">>>>",r)
        $('tbody').empty()
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                
                GerenciarProcedimentoAtendimentoMontaLinha(i)
            });
            GerenciarProcedimentoAtendimentoAssociaBotoesTabela()
        }
    })
}

// OK
function GerenciarProcedimentoAtendimentoMontaLinha(i) {


    let btnAtivoImg
    let btnAtivoCor
    if (i.ativo == "S") {
        btnAtivoImg = '<i class="bi bi-toggle-on"></i>'
        btnAtivoCor = 'btn-success'
    } else {
        btnAtivoImg = '<i class="bi bi-toggle-off"></i>'
        btnAtivoCor = 'btn-secondary'
    }

    let procedimento
    if ('descricao' in i) {
        procedimento = (i.descricao.length > 35) ?
            i.descricao.substring(0, 32) + "..." :
            i.descricao
    } else {
        procedimento = 'Erro cadastro'
    }


    let nomeCliente
    if ('cliNome' in i) {
        nomeCliente = (i.cliNome.length > 25) ?
            i.cliNome.substring(0, 22) + "..." :
            i.cliNome

    } else {
        nomeCliente = 'Erro cadastro'
    }


    $('tbody').append(`
        <tr >
        <td class="text-uppercase">${i.grupo}</td>
        <td class="text-uppercase" title="${i.descricao}">${procedimento}</td>
        <td class="text-uppercase" title="${i.nomeCliente}">${nomeCliente}</td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idProcedimento}
                tipo="editar"
                title="Edita dados do procedimento">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
      
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idProcedimento}
                tipo="excluir"
                title="Exclui o procedimento">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm ${btnAtivoCor} py-0"
                id=${i.idProcedimento}
                tipo="habilitar"
                status="${i.ativo}"
                title="Habilita/Desabilita o procedimento">
                ${btnAtivoImg}
            </button>
        </td>
    </tr>
    `)
}

// OK
function GerenciarProcedimentoAtendimentoAssociaBotoesTabela() {
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarProcedimentoAtendimentoBuscar(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarProcedimentoAtendimentoExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        const status = this.getAttribute("status")
        GerenciarProcedimentoAtendimentoHabilitar(id, status)
    })
}

// OK
function GerenciarProcedimentoAtendimentoGrupoEventoListar() {

    $('#grupo-evento-procedimento').empty()

    $.ajax({
        url: '/GerenciarProcedimentoAtendimentoGruposEventosListar',
        method: 'Post',
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar grupos de evento")
    }).done(function (r) {
        console.log(r)
        $('#grupo-evento-procedimento').append(`
            <option value="SELECIONAR">SELECIONE UM GRUPO</option>
        `)
        r.dados.forEach(i => {
            $('#grupo-evento-procedimento').append(`
                <option value="${i.grupo}">${i.grupo}</option>
            `)
        });
    })
}

// OK
function GerenciarProcedimentoAtendimentoHabilitar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarProcedimentoAtendimentoHabilitar',
        method: 'Post',
        data: JSON.stringify({ idProcedimento: id })
    }).fail(function (e) {
        boxErro("Erro ao habilitar/desabilitar Procedimento")
    }).done(function (r) {
        boxFechar()
        GerenciarProcedimentoAtendimentoListar()
    })
}

// OK
function GerenciarProcedimentoAtendimentoBuscar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarProcedimentoAtendimentoBuscar',
        method: 'Post',
        data: JSON.stringify({ idProcedimento: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Procedimento")
    }).done(function (r) {
        boxFechar()
        const d = r.dados
        localStorage.setItem('idProcedimento', d.idProcedimento)
        $('#id-cliente').val(d.idCliente)
        $('#grupo-evento-procedimento').val(d.grupo.toUpperCase())
        $('#descricao-procedimento').val(d.descricao.toUpperCase())
    })
}

// OK
function GerenciarProcedimentoAtendimentoExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Procedimento?`,
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
                url: '/GerenciarProcedimentoAtendimentoDeletar',
                method: 'Post',
                data: JSON.stringify({ idProcedimento: id })
            }).fail(function (e) {
                boxErro('Erro ao excluir o procedimento')
            }).done(function (r) {
                GerenciarProcedimentoAtendimentoLimparFormulario()
                boxDeletadoSucesso()
                GerenciarProcedimentoAtendimentoListar()
            })
        }
    })
}

// OK
function GerenciarProcedimentoAtendimentoAlterar(id) {

    if (GerenciarProcedimentoAtendimentoValidarCamposBranco()) return


    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarProcedimentoAtendimentoAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idProcedimento: localStorage.getItem('idProcedimento'),
                idCliente: $('#id-cliente').val(),
                grupo: $('#grupo-evento-procedimento').val(),
                descricao: $('#descricao-procedimento').val(),
            }
        )
    }).fail(function (e) {
        boxErro("Erro ao alterar dados do Procedimento")
    }).done(function (r) {
        boxAteradoSucesso()
        GerenciarProcedimentoAtendimentoLimparFormulario()
        GerenciarProcedimentoAtendimentoListar()
    })

}

// OK
function GerenciarProcedimentoAtendimentoInserir() {
   
    if (GerenciarProcedimentoAtendimentoValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarProcedimentoAtendimentoInserir',
        method: 'POST',
        data: JSON.stringify(
            {
                idFranqueado: localStorage.getItem('idFranqueado'),
                idCliente: $('#id-cliente').val(),
                grupo: $('#grupo-evento-procedimento').val(),
                descricao: $('#descricao-procedimento').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o Procedimento")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarProcedimentoAtendimentoLimparFormulario()
        GerenciarProcedimentoAtendimentoListar()
    })
}

// OK
function GerenciarProcedimentoAtendimentoValidarCamposBranco() {
    if ($('#descricao-procedimento').val() == "") {
        boxAdvertenciaAuto("O campo Decricao, não poder ficar em branco")
        $('#descricao-procedimento').focus()
        return true
    }
}



$(document).ready(function () {

    $(window).on('resize', function () {
        ajustaTabela()
        gerenciarUsuariosAlturaFixa()
    })

    ajustaTabela()
    gerenciarUsuariosAlturaFixa()
    // Variavel auxiliar para chavear entre alterar e inserir
    localStorage.setItem('idUsuario', '')


    $('#btn-limpar').on('click', gerenciarUsuariosLimparFormulario)
    $('#btn-gravar').on('click', gerenciarUsuariosGravarDados)

    $('#pesquisar-responsavel').on('input', gerenciarUsuariosFiltrarTabela)

    gerenciarUsuariosCarregarTabela()
})

// Filtra as linhas da grid por nome ou celular
function gerenciarUsuariosFiltrarTabela() {
    const termo = ($('#pesquisar-responsavel').val() || '').toLowerCase().trim()
    $('tbody tr').each(function () {
        const nome = $(this).find('td').eq(0).text().toLowerCase()
        const celular = $(this).find('td').eq(1).text().toLowerCase()
        const mostrar = nome.indexOf(termo) !== -1 || celular.indexOf(termo) !== -1
        $(this).toggle(mostrar)
    })
}

// Redimenciona a tela
function gerenciarUsuariosAlturaFixa() {
    $('.altura-fixa').height(window.innerHeight - 172)
}

function gerenciarUsuariosLimparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    localStorage.setItem('idUsuario', '')
}

function gerenciarUsuariosGravarDados() {
    const id = localStorage.getItem('idUsuario')

    if (id == "") {
        gerenciarUsuariosInserir()
    } else {
        gerenciarUsuariosAlterar(id)
    }
}

//ok
function gerenciarUsuariosCarregarTabela() {
    const idFranqueado = localStorage.getItem('idFranqueado')
    const idCentralUUID = localStorage.getItem('idCentralUUID') || ''
    const url = `/usuariosCarregarTabela`
    $.ajax({
        url: url,
        method: 'POST',
        data: JSON.stringify({
            idVinculo: idFranqueado,
            idCentralUUID: idCentralUUID
        })
    }).fail(function (e) {
        // console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        // console.log(r)
        $('tbody').empty()
        if (r.status != 'Vazio') {

            r.dados.forEach(i => {
                gerenciarUsuariosMontaLinha(i)
            });
            gerenciarUsuariosAssociaBotoesTabela()
            gerenciarUsuariosFiltrarTabela()
        }
    })
}

function gerenciarUsuariosMontaLinha(i) {

    // Configurar botao enviar email ==========================================
    let btnEnvioEmailImg
    let btnEnvioEmailCor
    if (i.enviarEmail == 'S') {
        btnEnvioEmailImg = '<i class="bi bi-envelope-check"></i>'
        btnEnvioEmailCor = 'btn-success'
    } else {
        btnEnvioEmailImg = '<i class="bi bi-envelope"></i>'
        btnEnvioEmailCor = 'btn-secondary'
    }

    // Configurar botao habilitar terminal ==========================
    let btnAtivarTerminalCor
    if (i.usuarioTerminal == 'S') {
        btnAtivarTerminalCor = 'btn-success'
    } else {
        btnAtivarTerminalCor = 'btn-secondary'
    }

    // Configurar botao habilitar site ==============================
    let btnAtivarSiteCor
    if (i.usuarioWeb == 'S') {
        btnAtivarSiteCor = 'btn-success'
    } else {
        btnAtivarSiteCor = 'btn-secondary'
    }

    const cel = formatarCelular(i.telefone1)

    // Configura o botao Ativo ================================================
    let btnAtivoCor
    let btnAtivoImg
    let btnAtivoTipo
    
    if (i.ativo == "S") {
        btnAtivoCor = 'btn-success'
        btnAtivoImg = '<i class="bi bi-toggle-on"></i>'
    } else {
        btnAtivoCor = 'btn-secondary'
        btnAtivoImg = '<i class="bi bi-toggle-off"></i>'
    }


    const botoes = (i.master == "N") ?
        `
        <!-- Botão editar -->
        <td class="p-0">
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idUsuario}
                tipo="editar"
                title="Edita os dados do usuário">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
        
        <!-- Botão resetar senha -->
        <td  class="p-0">
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idUsuario}
                tipo="resetar"
                title="Resetar senha do Usuário">
                <i class="bi bi-key-fill"></i>
            </button>
        </td>

         <!-- Botão excluir -->
        <td  class="p-0">
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idUsuario}
                tipo="excluir"
                title="Exclui o Usuário">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>

        <!-- Botão ativar o envio de email -->
        <td  class="p-0">
            <button class="btn btn-sm ${btnEnvioEmailCor} py-0"
                id=${i.idUsuario}
                tipo="habilitar-enviar-email"
                title="Habilita/Desabilita o envio de email">
                ${btnEnvioEmailImg}
                
            </button>
        </td>

         <!-- Botão ativar o uso terminal -->
        <td  class="p-0">
            <button class="btn btn-sm ${btnAtivarTerminalCor} py-0"
                id=${i.idUsuario}
                tipo="habilitarUsoTerminal"
                title="Habilita/Desabilita o uso do terminal">
                <i class="bi bi-headset"></i>
            </button>
        </td>

        <!-- Botão ativar acesso a central -->
        <td  class="p-0">
            <button class="btn btn-sm ${btnAtivarSiteCor} py-0"
                id=${i.idUsuario}
                tipo="habilitarUsoSite"
                title="Habilita/Desabilita o uso do website">
                <i class="bi bi-house"></i>
            </button>
        </td>

        <!-- Botão ativar -->
        <td  class="p-0 ">
            <button class="btn btn-sm ${btnAtivoCor} py-0"
                id=${i.idUsuario}
                tipo="habilitar"
                title="Habilita/Desabilita o Usuário">
                ${btnAtivoImg}
            </button>
        </td>

    ` : `<td colspan="7">MASTER</td>`

    // Inser a linha na tabela ================================================
    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.nome}</td>
        <td>${cel}</td>
        ${botoes}
    </tr>
    `)
}

function gerenciarUsuariosAssociaBotoesTabela() {
    
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        gerenciarUsuariosBuscar(id)
    })

    $('button[tipo=resetar]').on('click', function () {
        const id = this.getAttribute("id")
        gerenciarUsuarioslResetarSenha(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        gerenciarUsuariosExcluir(id)
    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")        
        gerenciarUsuariosHabilitar(id)
    })

    $('button[tipo=habilitarUsoTerminal]').on('click', function () {
        const id = this.getAttribute("id")
        gerenciarUsuariosHabilitarTerminal(id)
    })

    $('button[tipo=habilitarUsoSite]').on('click', function () {
        const id = this.getAttribute("id")
        gerenciarUsuariosHabilitarSite(id)
    })

    $('button[tipo=habilitar-enviar-email]').on('click', function () {
        const id = this.getAttribute("id")
        gerenciarUsuariosEmailAtivo(id)
    })
}

function gerenciarUsuarioslResetarSenha(id) {
    const url = `/usuariosResetarSenha`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        // console.log(e)
        boxErro("Erro ao resetar a senha do usuário")
    }).done(function (r) {
        boxFechar()
        boxSenhaAlterada('usuario123')

    })
}

function gerenciarUsuariosHabilitar(id) {

    const url = `/usuariosHabilitar`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        // console.log(e)
        boxErro('Erro ao habilitar/Desabilitar o usuário')
    }).done(function (r) {
        // console.log(r)
        boxFechar()
        gerenciarUsuariosCarregarTabela()
    })
}

function gerenciarUsuariosHabilitarTerminal(id) {
    const url = `/usuariosHabilitarTerminal`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        // console.log(e)
        boxErro('Erro ao habilitar/Desabilitar o usuário')
    }).done(function (r) {
        boxFechar()
        gerenciarUsuariosCarregarTabela()
    })
}

function gerenciarUsuariosHabilitarSite(id) {
    const url = `/usuariosHabilitarSite`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        // console.log(e)
        boxErro('Erro ao habilitar/Desabilitar o uso do site')
    }).done(function (r) {
        boxFechar()
        gerenciarUsuariosCarregarTabela()
    })
}

function gerenciarUsuariosEmailAtivo(id) {
    const url = `/usuariosEmailAtivo`

    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        boxErro('Erro ao habilitar/Desabilitar o Responsavel')
    }).done(function (r) {

        boxFechar()
        gerenciarUsuariosCarregarTabela()
    })
}

function gerenciarUsuariosInserir() {

    if (gerenciarUsuariosValidarCamposBranco()) return

    const url = `/usuariosInserir`

    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify(
            {
                idVinculo: localStorage.getItem('idFranqueado'),
                idCentralUUID: localStorage.getItem('idCentralUUID') || '',
                nome: $('#nome-responsavel').val(),
                nick: $('#nick-responsavel').val(),
                tipo: "FRA",
                telefone1: limpaDocumento($('#telefone_1-responsavel').val()),
                telefone2: limpaDocumento($('#telefone_2-responsavel').val()),
                email1: $('#email_1-responsavel').val(),
                email2: $('#email_2-responsavel').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o Usuário")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        gerenciarUsuariosLimparFormulario()
        gerenciarUsuariosCarregarTabela()
    })
}

function gerenciarUsuariosBuscar(id) {

    const url = `/usuariosBuscar`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Técnico")
        consoUsuario
    }).done(function (r) {
        // console.log(r)
        boxFechar()
        const d = r.dados
        // Carrega o id do responsavel na storage para informar que e uma edição
        // e não inclusoa
        localStorage.setItem('idUsuario', d.idUsuario)


        $('#nome-responsavel').val(d.nome)
        $('#nick-responsavel').val(d.nick)
        $('#email_1-responsavel').val(d.email1)
        $('#email_2-responsavel').val(d.email1)
        $('#telefone_1-responsavel').val(formatarCelular(d.telefone1))
        $('#telefone_2-responsavel').val(formatarCelular(d.telefone2))
    })
}

function gerenciarUsuariosAlterar(id) {
    if (gerenciarUsuariosValidarCamposBranco()) return

    const url = `/usuariosAlterar`
    $.ajax({
        start: boxProcessando(),
        url: url,
        method: 'POST',
        data: JSON.stringify(
            {
                idUsuario: localStorage.getItem('idUsuario'),
                nome: $('#nome-responsavel').val(),
                nick: $('#nick-responsavel').val(),
                telefone1: limpaDocumento($('#telefone_1-responsavel').val()),
                telefone2: limpaDocumento($('#telefone_2-responsavel').val()),
                email1: $('#email_1-responsavel').val(),
                email2: $('#email_2-responsavel').val(),
            }
        )
    }).fail(function (e) {
        // console.log(e)
        boxErro("Erro ao alterar dados do Responsavel")
    }).done(function (r) {
        boxAteradoSucesso()
        gerenciarUsuariosLimparFormulario()
        gerenciarUsuariosCarregarTabela()
    })

}

function gerenciarUsuariosExcluir(id) {
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
                url: `/usuariosApagar`,
                method: 'POST',
                data: JSON.stringify({ idUsuario: id })
            }).fail(function (e) {
                console.log(e)
                boxErro(e)
            }).done(function (r) {
                gerenciarUsuariosLimparFormulario()
                gerenciarUsuariosCarregarTabela()
                boxDeletadoSucesso()
            })
        }
    })
}

function gerenciarUsuariosValidarCamposBranco() {
    if ($('#nome-responsavel').val() == "") {
        boxAdvertenciaCampoAuto("O campo Nome, não poder ficar em branco", '#nome-responsavel')
        return true
    }
    if ($('#nick-responsavel').val() == "") {
        boxAdvertenciaCampoAuto("O campo Nick, não poder ficar em branco", '#nick-responsavel')
        return true
    }

    if ($('#email_1-responsavel').val() == "") {
        boxAdvertenciaCampoAuto("O campo Email, não poder ficar em branco", '#email_1-responsavel')
        return true
    }

    if ($('#telefone_1-responsavel').val() == "") {
        boxAdvertenciaCampoAuto("O campo Telefone 1, não poder ficar em branco", '#telefone_1-responsavel')
        return true
    }

    return false
}

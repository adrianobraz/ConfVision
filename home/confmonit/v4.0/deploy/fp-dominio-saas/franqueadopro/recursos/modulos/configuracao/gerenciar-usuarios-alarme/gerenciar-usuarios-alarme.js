var listaClientes = []

$(document).ready(function () {

    $(window).on('resize', function () {
        ajustaTabela()
    })

    localStorage.setItem('codigoUsuarioAlarme', "")

    // Associa o click do botao limpar
    $('#btn-limpar').on('click', GerenciarUsuariosAlarmeLimparFormulario)

    // Associa o click do botao gravar
    $('#btn-gravar').on('click', GerenciarUsuariosAlarmeGravarDados)

    // Associa e evento de perda focus do campo numero setor alarme
    $("#codigo-usuario-alarme").on('blur', function () {
        if ($("#codigo-usuario-alarme").val() != "") {
            this.value = ("000" + this.value).slice(-3)
            GerenciarUsuariosAlarmeVerificarCodigoUsuarioAlarmeLivre()
        }
    });

    // Associa o eveto de selecionar do idCliente
    $('#id-cliente').on('change', function () {
        if ($('#id-cliente').val() != '0') {
            GerenciarUsuariosAlarmeDesbloquearIdDispositivo()
            GerenciarUsuariosAlarmeCarregarTabela()
        } else {
            GerenciarUsuariosAlarmeBloquearIdDispositivo()
            GerenciarUsuariosAlarmeLimparCampos()
            GerenciarUsuariosAlarmeCarregarTabela()
        }
    })

    // Pesquisa de cliente no combo (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', GerenciarUsuariosAlarmeFiltrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') GerenciarUsuariosAlarmeFiltrarClientes()
    })
    $('#lista-clientes').on('click', '.fp-combo-item', function () {
        $('#busca-cliente').val($(this).attr('data-nome'))
        $('#id-cliente').val($(this).attr('data-id')).trigger('change')
        $('#lista-clientes').addClass('d-none')
    })
    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-cliente, #lista-clientes').length) {
            $('#lista-clientes').addClass('d-none')
        }
    })

    ajustaTabela()
    GerenciarUsuariosAlarmeCarregarClientes()
    GerenciarUsuariosAlarmeBloquearIdDispositivo()
})

function clienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function GerenciarUsuariosAlarmeFiltrarClientes() {
    var raw = ($('#busca-cliente').val() || '').toLowerCase().trim()
    var termoNum = raw.replace(/[^a-z0-9]/g, '')
    var lista = $('#lista-clientes')
    lista.empty()

    if (raw == '') {
        $('#id-cliente').val('0').trigger('change')
        lista.addClass('d-none')
        return
    }

    var resultados = listaClientes.filter(function (i) {
        var texto = clienteTextoBusca(i)
        var textoNum = texto.replace(/[^a-z0-9]/g, '')
        return texto.indexOf(raw) != -1 || (termoNum != '' && textoNum.indexOf(termoNum) != -1)
    }).slice(0, 30)

    if (resultados.length == 0) {
        lista.append('<li class="fp-combo-empty">Nenhum cliente encontrado</li>')
    } else {
        resultados.forEach(function (i) {
            var doc = i.documento1 ? ' — ' + i.documento1 : ''
            $('<li class="fp-combo-item"></li>')
                .attr('data-id', i.idCliente)
                .attr('data-nome', i.nome)
                .text(i.nome + doc)
                .appendTo(lista)
        })
    }
    lista.removeClass('d-none')
}

function GerenciarUsuariosAlarmeLimparFormulario() {

    localStorage.setItem('codigoUsuarioAlarme', "")

    $('#formulario').each(function () {
        this.reset();
    })
    $('#busca-cliente').val('')
    $('#lista-clientes').empty().addClass('d-none')
    $('#id-cliente').val('0')
    $('#id-dispositivo').val('0')
    $('#id-usuario-alarme').val("")
    $('#numero-setor-alarme').val("")
    $('tbody').empty()
}


//######################## RELACIONADO A TABELA ########################//

// Guarda conteudo da tabela para uso na função verificarCodigoUsuarioAlarmeLivre

function GerenciarUsuariosAlarmeCarregarTabela() {
    const idDispositivo = $("#id-dispositivo").val()
    // Caso idDispositivo for igual a 0 siginifica que nenhum cliente
    // foi selecionado entao limpa a tabela e retorna 
    if (idDispositivo == 0) {
        $('tbody').empty()
        return
    }

    $.ajax({
        url: '/UsuarioAlarmeListrar',
        method: 'Post',
        data: JSON.stringify({
            idDispositivo: idDispositivo
        })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela usuarios")
    }).done(function (r) {
        console.log(r)
        let tabela = new StorageObjeto('tabela')
        tabela.clear()
        $('tbody').empty()
        if (r.status != 'Vazio') {
            tabela.save(r.dados)
            r.dados.forEach(i => {
                GerenciarUsuariosAlarmeMontaLinha(i)
            })
            GerenciarUsuariosAlarmeAssociaBotoesTabela()
        }
    })
}

// montaLinha monta as linhas da tabela
function GerenciarUsuariosAlarmeMontaLinha(i) {
    const btnImg = (i.ativo == 'S') ?
        '<i class="bi bi-toggle-on"></i>' :
        '<i class="bi bi-toggle-off"></i>'

    // Acerta a cor do botao habilitar depedendo se esta ativo ou nao
    const btnCor = (i.ativo == 'S') ? 'btn-success' : 'btn-secondary'
    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.codigo}</td> 
        <td class="text-uppercase">${i.nome}</td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idUsuario}
                tipo="editar"
                title="Editar dados do Usuário do Alarme">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idUsuario}
                tipo="excluir"
                title="Exclui o Usuário do Alarme">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm ${btnCor} py-0"
                id=${i.idUsuario}
                tipo="habilitar"
                status="${i.ativo}"
                title="Habilita/Desabilita o acesso do Usuário do Alarme">
                ${btnImg}
            </button>
        </td>
    </tr>
    `)
}

// associaBotoesTabel associa o click dos botoes da tabela
function GerenciarUsuariosAlarmeAssociaBotoesTabela() {
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarUsuariosAlarmeBuscar(id)
    })

    $('button[tipo=resetar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarUsuariosAlarmeResetarSenha(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarUsuariosAlarmeExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        const status = this.getAttribute("status")
        GerenciarUsuariosAlarmeHabilitar(id, status)
    })
}
//######################################################################//

//#################### RELACIONADO AO CAMPO CLIENTE ####################//
function GerenciarUsuariosAlarmeCarregarClientes() {
    const idFranqueado = localStorage.getItem('idFranqueado')

    $.ajax({
        url: 'UsuariosAlarmeClienteListar',
        method: 'Post',
        data: JSON.stringify({
            idFranqueado: idFranqueado
        })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        listaClientes = (r.dados || [])
        $('#id-cliente').val('0')
    })
}
//######################################################################//

//################## RELACIONADO AO CAMPO DISPOSITIVO ##################//
function GerenciarUsuariosAlarmeBloquearIdDispositivo() {
    $('#id-dispositivo').attr('disabled', true)
    $('#id-dispositivo').empty()
    $('#id-dispositivo').append('<option value="0">Selecione</option>')
}

function GerenciarUsuariosAlarmeDesbloquearIdDispositivo() {
    const idCliente = $("#id-cliente").val()

    $.ajax({
        url: 'UsuariosAlarmeDispositivoListar',
        method: 'Post',
        data: JSON.stringify({
            idCliente: idCliente
        })
    }).fail(function (e) {
        boxErro("Erro ao carregar os dispositivos")
    }).done(function (r) {
        $('#id-dispositivo').empty()
        $('#id-dispositivo').append('<option value="0">Selecione</option>')

        r.dados.forEach(i => {
            $('#id-dispositivo').append(`
                <option value="${i.idDispositivo}">${i.nome}</option>
            `)
        });


        $('#id-dispositivo').on('change', function () {
            if ($('#id-dispositivo').val() != '0') {
                GerenciarUsuariosAlarmeLimparCampos()
                GerenciarUsuariosAlarmeDesbloquearCampos()
                GerenciarUsuariosAlarmeCarregarTabela()
            } else {
                limparCampos()
                bloquearCampos()
                carregarTabela()
            }
        })
        // Desbloquea o dispositivo
        $('#id-dispositivo').attr('disabled', false)

    })
}



//######################################################################//

//############## RELACIONADO AOS CAMPOS DE PREENCHIMENTO ###############//
function GerenciarUsuariosAlarmeBloquearCampos() {
    $('#codigo-usuario-alarme').attr('disabled', true)
    $('#nome-usuario-alarme').attr('disabled', true)
    $('#celular-usuario-alarme').attr('disabled', true)
    $('#email-usuario-alarme').attr('disabled', true)
    $('#observacao-usuario-alarme').attr('disabled', true)
}

function GerenciarUsuariosAlarmeDesbloquearCampos() {
    $('#codigo-usuario-alarme').attr('disabled', false).focus()
    $('#nome-usuario-alarme').attr('disabled', false)
    $('#celular-usuario-alarme').attr('disabled', false)
    $('#email-usuario-alarme').attr('disabled', false)
    $('#cep-usuario-alarme').attr('disabled', false)
    $('#observacao-usuario-alarme').attr('disabled', false)

}

function GerenciarUsuariosAlarmeLimparCampos() {

    
    $('#formulario').each(function () {
        this.reset();
    })
    $('#id-usuario-alarme').val("")
    $('#numero-setor-alarme').val("")
  

}
//######################################################################//

function GerenciarUsuariosAlarmeVerificarCodigoUsuarioAlarmeLivre() {
    const codigo = $('#codigo-usuario-alarme').val()

    if (codigo != "") {
        // pega numero anterior
        if (codigo == localStorage.getItem('codigoUsuarioAlarme')) return

        let tabela = new StorageObjeto('tabela')
        if (tabela.verificar()) {
            tabela.get().forEach(i => {
                if (i.codigo == codigo) {
                    boxAdvertenciaCampoAuto(
                        `Este numero já se encontra em uso por ${i.nome}, escolha outro`,
                        '#codigo-usuario-alarme'
                    )
                }
            })
        }
    }
}

function GerenciarUsuariosAlarmeHabilitar(id) {
    $.ajax({
        start: boxProcessando(),
        url: 'UsuariosAlarmeHabilitar',
        method: 'Post',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao habilitar/desabilitar Setor")
    }).done(function (r) {
        console.log(r)
        boxFechar()
        GerenciarUsuariosAlarmeCarregarTabela()
    })
}


//############# RELACIONADO A INSERCAO E ALTERAÇÃO DE DADOS ############//
function GerenciarUsuariosAlarmeGravarDados() {
    const id = $('#id-usuario-alarme').val()

    if (id == "") {
        GerenciarUsuariosAlarmeInserir()
    } else {
        GerenciarUsuariosAlarmeAlterar(id)
    }
}

// inserir insere um novo usuario no dispositivo
function GerenciarUsuariosAlarmeInserir() {
    if (GerenciarUsuariosAlarmeValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: 'UsuariosAlarmeInserir',
        method: 'Post',
        data: JSON.stringify(
            {
                idDispositivo: $('#id-dispositivo').val(),
                codigo: $('#codigo-usuario-alarme').val(),
                nome: $('#nome-usuario-alarme').val(),
                celular: limpaDocumento($('#celular-usuario-alarme').val()),
                email: $('#email-usuario-alarme').val(),
                observacao: $('#observacao-usuario-alarme').val(),
            }
        )

    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o Setor")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarUsuariosAlarmeLimparCampos()
        GerenciarUsuariosAlarmeCarregarTabela()
    })
}

function GerenciarUsuariosAlarmeAlterar(id) {
    if (GerenciarUsuariosAlarmeValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: 'UsuariosAlarmeAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idUsuario: $('#id-usuario-alarme').val(),
                codigo: $('#codigo-usuario-alarme').val(),
                nome: $('#nome-usuario-alarme').val(),
                celular: limpaDocumento($('#celular-usuario-alarme').val()),
                email: $('#email-usuario-alarme').val(),
                observacao: $('#observacao-usuario-alarme').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao alterar dados do Setor")
    }).done(function (r) {
        boxAteradoSucesso()
        GerenciarUsuariosAlarmeLimparCampos()
        GerenciarUsuariosAlarmeCarregarTabela()
    })
}
//######################################################################//


function GerenciarUsuariosAlarmeBuscar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/UsuariosAlarmeBuscar',
        method: 'Post',
        data: JSON.stringify({ idUsuario: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Usuario")
    }).done(function (r) {
        console.log(r)
        boxFechar()
        const d = r.dados

        localStorage.setItem('codigoUsuarioAlarme', d.codigo)

        const celular = ('celular' in d) ? formatarCelular(d.celular) : ''

        $('#id-usuario-alarme').val(d.idUsuario)
        $('#codigo-usuario-alarme').val(localStorage.getItem('codigoUsuarioAlarme'))
        $('#nome-usuario-alarme').val(d.nome)
        $('#celular-usuario-alarme').val(celular)
        $('#email-usuario-alarme').val(d.email)
        $('#observacao-usuario-alarme').val(d.observacao)

    })
}

function GerenciarUsuariosAlarmeExcluir(id) {


    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Usuario?`,
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
                url: `/UsuariosAlarmeExcluir`,
                method: 'Post',
                data: JSON.stringify({idUsuario: id})
            }).fail(function (e) {
                console.log(e)
                boxErro('Erro ao excluir o Usuario')
            }).done(function (r) {
                boxDeletadoSucesso()
                GerenciarUsuariosAlarmeCarregarTabela()
                GerenciarUsuariosAlarmeLimparCampos()
            })
        }
    })
}

function GerenciarUsuariosAlarmeValidarCamposBranco() {
    if ($('#id-cliente').val() == "0") {
        boxAdvertenciaAuto("Selecione um cliente primeiro")
        $('#busca-cliente').focus()
        return true
    }
    if ($('#codigo-usuario-alarme').val() == "") {
        boxAdvertenciaAuto("O campo Codigo, não poder ficar em branco")
        $('#codigo-usuario-alarme').focus()
        return true
    }

    if ($('#nome-usuario-alarme').val() == "") {
        boxAdvertenciaAuto("O campo Nome, não poder ficar em branco")
        $('#nome-usuario-alarme').focus()
        return true
    }

    if ($('#nick-usuario-alarme').val() == "") {
        boxAdvertenciaAuto("O campo Nick, não poder ficar em branco")
        $('#nick-usuario-alarme').focus()
        return true
    }
    /*
        if ($('#telefone-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Telefone, não poder ficar em branco")
            $('#telefone-usuario-alarme').focus()
            return true
        }
    
        if ($('#celular-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Celular, não poder ficar em branco")
            $('#celular-usuario-alarme').focus()
            return true
        }
       
        if ($('#email-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Email, não poder ficar em branco")
            $('#email-usuario-alarme').focus()
            return true
        }
        
        if ($('#cep-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Cep, não poder ficar em branco")
            $('#cep-usuario-alarme').focus()
            return true
        }
    
        if ($('#endereco-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Endereço, não poder ficar em branco")
            $('#endereco-usuario-alarme').focus()
            return true
        }
    
        if ($('#bairro-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Bairro, não poder ficar em branco")
            $('#bairro-usuario-alarme').focus()
            return true
        }
    
        if ($('#cidade-usuario-alarme').val() == "") {
            boxAdvertenciaAuto("O campo Cidade, não poder ficar em branco")
            $('#cidade-usuario-alarme').focus()
            return true
        }
        */
}




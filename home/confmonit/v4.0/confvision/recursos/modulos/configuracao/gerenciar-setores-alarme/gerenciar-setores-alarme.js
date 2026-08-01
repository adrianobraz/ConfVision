var listaClientes = []

$(document).ready(function () {
    document.body.classList.add('cv-cadastro')

    $(window).on('resize', function () {
        ajustaTabela()
    })

    localStorage.setItem('idSetorAlarme', '')

    // Associa o click do botao limpar
    $('#btn-limpar').on('click', GerenciarSetoresAlarmeLimparCampos)//GerenciarSetoresAlarmeLimparFormulario

    // Associa o click do botao gravar
    $('#btn-gravar').on('click', GerenciarSetoresAlarmeGravarDados)

    // Associa e evento de perda focus do campo numero setor alarme
    $("#numero-setor-alarme").on('blur', function () {
        this.value = ("000" + this.value).slice(-3)
        if ($("#numero-setor-alarme").val() != "") {
            GerenciarSetoresAlarmeVerificarNumeroSetorLivre()
        }
    })

    // Associa o eveto de selecionar do idCliente
    $('#id-cliente').on('change', function () {
        if ($('#id-cliente').val() == '0') {
            GerenciarSetoresAlarmeBloquearIdDispositivo()
        } else {
            GerenciarSetoresAlarmeDesbloquearIdDispositivo()
        }
    })

    // Pesquisa de cliente no combo (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', GerenciarSetoresAlarmeFiltrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') GerenciarSetoresAlarmeFiltrarClientes()
    })
    $('#lista-clientes').on('click', '.cv-combo-item', function () {
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
    GerenciarSetoresAlarmeDefaultsConfVision()
    GerenciarSetoresAlarmeClientesListar()
    GerenciarSetoresAlarmeBloquearIdDispositivo()
})


function GerenciarSetoresAlarmeMsgErro(e, padrao) {
    if (e && e.responseJSON && e.responseJSON.status) {
        return e.responseJSON.status
    }
    return padrao
}

function GerenciarSetoresAlarmeEhSensor(item) {
    return String(item && item.tipoSetor || '').trim().toUpperCase() === 'SENSOR'
}

function GerenciarSetoresAlarmeBuscarSetorTabela(id) {
    const tabela = new StorageObjeto('tabela')
    if (!tabela.verificar()) return null
    return tabela.get().find(function (i) {
        return String(i.idSetor) === String(id)
    }) || null
}

function GerenciarSetoresAlarmeEhSensorPorId(id) {
    return GerenciarSetoresAlarmeEhSensor(GerenciarSetoresAlarmeBuscarSetorTabela(id))
}

function GerenciarSetoresAlarmeMsgSetorSensor() {
    boxAdvertenciaAuto('Setor do tipo SENSOR não pode ser alterado, excluído ou habilitado/desabilitado.')
}


//######################## RELACIONADO A TABELA ########################//

function GerenciarSetoresAlarmeCarregarTabela() {
    const idDispositivo = $("#id-dispositivo").val()
    // Caso idDispositivo for igual a 0 siginifica que nenhum cliente
    // foi selecionado entao limpa a tabela e retorna 
    if (idDispositivo == 0) {
        $('tbody').empty()
        GerenciarSetoresAlarmeBloquearCampos()
        return
    }

    $.ajax({
        url: '/GerenciarSetoresAlarmeListar',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        console.log(r)
        let tabela = new StorageObjeto('tabela')
        tabela.save(r.dados)
        $('tbody').empty()

        if (r.status != 'Vazio') {
            gGerenciarSetoresAlarmeTabela = r.dados
            r.dados.forEach(i => {
                GerenciarSetoresAlarmeMontaLinha(i)
            });
            GerenciarSetoresAlarmeDesbloquearCampos()
            GerenciarSetoresAlarmeAssociaBotoesTabela()
        }
    })
}

// montaLinha monta as linhas da tabela
function GerenciarSetoresAlarmeMontaLinha(i) {


    let btnAtivoImg
    let btnAtivoCor
    if (i.ativo == "S") {
        btnAtivoImg = '<i class="bi bi-toggle-on"></i>'
        btnAtivoCor = 'btn-success'
    } else {
        btnAtivoImg = '<i class="bi bi-toggle-off"></i>'
        btnAtivoCor = 'btn-secondary'
    }

    let btnEditarHtml
    let btnExcluirHtml
    let btnHabilitarHtml
    const tituloSensor = 'Setor SENSOR — somente leitura'

    if (GerenciarSetoresAlarmeEhSensor(i)) {
        btnEditarHtml = `
            <button class="btn btn-sm btn-secondary py-0" disabled
                title="${tituloSensor}">
                <i class="bi bi-pencil-square"></i>
            </button>`
        btnExcluirHtml = `
            <button class="btn btn-sm btn-secondary py-0" disabled
                title="${tituloSensor}">
                <i class="bi bi-eraser-fill"></i>
            </button>`
        btnHabilitarHtml = `
            <button class="btn btn-sm btn-secondary py-0" disabled
                title="${tituloSensor}">
                ${btnAtivoImg}
            </button>`
    } else {
        btnEditarHtml = `
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idSetor}
                tipo="editar"
                title="Editar dados do Setor do Alarme ">
                <i class="bi bi-pencil-square"></i>
            </button>`
        btnExcluirHtml = `
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idSetor}
                tipo="excluir"
                title="Exclui o Setor do Alarme">
                <i class="bi bi-eraser-fill"></i>
            </button>`
        btnHabilitarHtml = `
            <button class="btn btn-sm ${btnAtivoCor} py-0"
                id=${i.idSetor}
                tipo="habilitar"
                status="${i.ativo}"
                title="Habilita/Desabilita o Setor do Alarme">
                ${btnAtivoImg}
            </button>`
    }

    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.numero}</td> 
        <td class="text-uppercase">${i.particao}</td> 
        <td class="text-uppercase">${i.nome}</td>
        <td class="text-uppercase">${i.tipo}</td>
        
        <td>
            ${btnEditarHtml}
        </td>
      
        <td>
            ${btnExcluirHtml}
        </td>
        <td>
            ${btnHabilitarHtml}
        </td>
    </tr>
    `)
}

// associaBotoesTabel associa o click dos botoes da tabela
function GerenciarSetoresAlarmeAssociaBotoesTabela() {
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarSetoresAlarmeBuscar(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarSetoresAlarmeExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        const status = this.getAttribute("status")
        GerenciarSetoresAlarmeHabilitar(id, status)
    })
}
//######################################################################//

//#################### RELACIONADO AO CAMPO CLIENTE ####################//
function GerenciarSetoresAlarmeClienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function GerenciarSetoresAlarmeFiltrarClientes() {
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
        var texto = GerenciarSetoresAlarmeClienteTextoBusca(i)
        var textoNum = texto.replace(/[^a-z0-9]/g, '')
        return texto.indexOf(raw) != -1 || (termoNum != '' && textoNum.indexOf(termoNum) != -1)
    }).slice(0, 30)

    if (resultados.length == 0) {
        lista.append('<li class="cv-combo-empty">Nenhum cliente encontrado</li>')
    } else {
        resultados.forEach(function (i) {
            var doc = i.documento1 ? ' — ' + i.documento1 : ''
            $('<li class="cv-combo-item"></li>')
                .attr('data-id', i.idCliente)
                .attr('data-nome', i.nome)
                .text(i.nome + doc)
                .appendTo(lista)
        })
    }
    lista.removeClass('d-none')
}

function GerenciarSetoresAlarmeClientesListar() {

    $.ajax({
        url: '/GerenciarSetoresAlarmeListarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        listaClientes = r.dados || []
        $('#id-cliente').val('0')
        $('#busca-cliente').val('')
        $('#lista-clientes').empty().addClass('d-none')
    })
}
//######################################################################//

//################## RELACIONADO AO CAMPO DISPOSITIVO ##################//
function GerenciarSetoresAlarmeBloquearIdDispositivo() {
    $('#id-dispositivo').attr('disabled', true)
    $('#id-dispositivo').empty()
    $('#id-dispositivo').append('<option value="0">SELECIONE UM DISPOSITIVO</option>')
    // limpa a tabela
    GerenciarSetoresAlarmeCarregarTabela()
}

function GerenciarSetoresAlarmeDesbloquearIdDispositivo() {

    $('tbody').empty()

    $.ajax({
        url: '/SetoresAlarmeDispositivoListarPorCliente',
        method: 'Post',
        data: JSON.stringify({ idCliente: $("#id-cliente").val() })
    }).fail(function (e) {
        boxErro("Erro ao carregar os dispositivos")
    }).done(function (r) {
        console.log("disp", r)
        $('#id-dispositivo').empty()
        $('#id-dispositivo').append('<option value="0">SELECIONE</option>')

        r.dados.forEach(i => {
            $('#id-dispositivo').append(`
                <option value="${i.idDispositivo}" particao="${i.particao}">${i.nome}</option>
            `)
        })

        $('#id-dispositivo').on('change', function () {

            if ($('#id-dispositivo').val() != '0') {
                GerenciarSetoresAlarmeDesbloquearCampos()
            } else {
                GerenciarSetoresAlarmeBloquearCampos()
            }
            GerenciarSetoresAlarmeCarregarTabela()
        })

        // Desbloquea o dispositivo
        $('#id-dispositivo').attr('disabled', false)
    })
}
//######################################################################//



//############## RELACIONADO AOS CAMPOS DE PREENCHIMENTO ###############//
function GerenciarSetoresAlarmeBloquearCampos() {

    $('#numero-setor-alarme').attr('disabled', true)
    $('#tipo-area-setor-alarme').attr('disabled', true)
    $('#nome-setor-alarme').attr('disabled', true)
    $('#descricao-setor-alarme').attr('disabled', true)
}

function GerenciarSetoresAlarmeDesbloquearCampos() {

    $('#numero-setor-alarme').attr('disabled', false)
    $('#tipo-area-setor-alarme').attr('disabled', false)
    $('#nome-setor-alarme').attr('disabled', false)
    $('#descricao-setor-alarme').attr('disabled', false)
}

function GerenciarSetoresAlarmeDefaultsConfVision() {
    $('#setor-com-camera').val('S')
}
//######################################################################//

function GerenciarSetoresAlarmeVerificarNumeroSetorLivre() {
    const numero = $('#numero-setor-alarme').val()
    if (numero != "") {
        let tabela = new StorageObjeto('tabela')
        // pega numero anterior
        if (numero == localStorage.getItem('setorAtual')) return

        if (tabela.verificar()) {
            tabela.get().forEach(i => {
                if (i.numero == numero) {
                    boxAdvertenciaCampoAuto(
                        `Este numero já se encontra em uso por ${i.nome}, escolha outro`,
                        '#numero-setor-alarme'
                    )
                }
            });

        }
    }
}

function GerenciarSetoresAlarmeLimparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })
    $('#busca-cliente').val('')
    $('#id-cliente').val('0')
    $('#lista-clientes').empty().addClass('d-none')
    localStorage.setItem('idSetorAlarme', '')
    $('#numero-setor-alarme').val("")
    GerenciarSetoresAlarmeBloquearIdDispositivo()
}

function GerenciarSetoresAlarmeLimparCampos() {
    localStorage.setItem('idSetorAlarme', '')

    $('#numero-setor-alarme').val('')
    GerenciarSetoresAlarmeDefaultsConfVision()
    $('#tipo-area-setor-alarme').val('ÁREA INTERNA')
    $('#nome-setor-alarme').val('')
    $('#descricao-setor-alarme').val('')
    GerenciarSetoresAlarmeCarregarTabela()
}

function GerenciarSetoresAlarmeHabilitar(id) {
    if (GerenciarSetoresAlarmeEhSensorPorId(id)) {
        GerenciarSetoresAlarmeMsgSetorSensor()
        return
    }

    $.ajax({
        start: boxProcessando(),
        url: `/GerenciarSetoresAlarmeHabilitar`,
        method: 'Post',
        data: JSON.stringify({ idSetor: id })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao habilitar/desabilitar Setor")
    }).done(function (r) {
        boxFechar()
        GerenciarSetoresAlarmeCarregarTabela()
    })
}

//############# RELACIONADO A INSERCAO E ALTERAÇÃO DE DADOS ############//
function GerenciarSetoresAlarmeGravarDados() {
    const id = localStorage.getItem('idSetorAlarme')

    if (id == "") {
        GerenciarSetoresAlarmeInserir()
    } else {
        GerenciarSetoresAlarmeAlterar(id)
    }
}

function GerenciarSetoresAlarmeInserir() {

    if (GerenciarSetoresAlarmeValidarCamposBranco()) return
    const particao = $('#id-dispositivo :selected').attr('particao')
    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarSetoresAlarmeInserir',
        method: 'POST',
        contentType: 'application/json',
        dataType: 'json',
        data: JSON.stringify(
            {
                idDispositivo: $('#id-dispositivo').val(),
                particao: addZeroEsquerda(particao, 2),
                numero: $('#numero-setor-alarme').val(),
                camera: 'S',
                tipo: $('#tipo-area-setor-alarme').val(),
                nome: $('#nome-setor-alarme').val(),
                descricao: $('#descricao-setor-alarme').val(),
            }
        )
    }).fail(function (e) {
        boxFechar()
        boxErro(GerenciarSetoresAlarmeMsgErro(e, "Erro ao inserir o Setor"))
    }).done(function (r) {
        boxFechar()
        boxInseridoSucesso(r.dados)
        GerenciarSetoresAlarmeLimparCampos()
    })
}

function GerenciarSetoresAlarmeAlterar(id) {
   
    if (GerenciarSetoresAlarmeValidarCamposBranco()) return

    const tabela = new StorageObjeto('tabela')
    if (tabela.verificar()) {
        const setorAtual = tabela.get().find(function (i) {
            return String(i.idSetor) === String(id)
        })
        if (GerenciarSetoresAlarmeEhSensor(setorAtual)) {
            GerenciarSetoresAlarmeMsgSetorSensor()
            return
        }
    }

    const particao = $('#id-dispositivo :selected').attr('particao')
    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarSetoresAlarmeAlterar',
        method: 'Post',
        contentType: 'application/json',
        dataType: 'json',
        data: JSON.stringify(
            {
                idSetor: localStorage.getItem('idSetorAlarme'),
                particao: addZeroEsquerda(particao, 2),
                numero: $('#numero-setor-alarme').val(),
                camera: 'S',
                tipo: $('#tipo-area-setor-alarme').val(),
                nome: $('#nome-setor-alarme').val().toUpperCase(),
                descricao: $('#descricao-setor-alarme').val().toUpperCase(),
            }
        )
    }).fail(function (e) {
        boxFechar()
        boxErro(GerenciarSetoresAlarmeMsgErro(e, "Erro ao alterar dados do Setor"))
    }).done(function (r) {
        boxFechar()
        boxAteradoSucesso()
        GerenciarSetoresAlarmeLimparCampos()
    })
}

function GerenciarSetoresAlarmeBuscar(id) {
    localStorage.setItem('setorAtual', "")

    $.ajax({
        start: boxProcessando(),
        url: 'GerenciarSetoresAlarmeBuscar',
        method: 'Post',
        data: JSON.stringify({ idSetor: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Setor")
    }).done(function (r) {
        console.log(r)
        boxFechar()
        const d = r.dados

        if (GerenciarSetoresAlarmeEhSensor(d)) {
            localStorage.setItem('idSetorAlarme', '')
            localStorage.setItem('setorAtual', '')
            GerenciarSetoresAlarmeMsgSetorSensor()
            return
        }

        localStorage.setItem('setorAtual', d.numero)
        localStorage.setItem('idSetorAlarme', d.idSetor)
       
        $('#numero-setor-alarme').val(d.numero)
        GerenciarSetoresAlarmeDefaultsConfVision()
        $('#tipo-area-setor-alarme').val(d.tipo)
        $('#nome-setor-alarme').val(d.nome)
        $('#descricao-setor-alarme').val(d.descricao)
    })
}

function GerenciarSetoresAlarmeExcluir(id) {
    if (GerenciarSetoresAlarmeEhSensorPorId(id)) {
        GerenciarSetoresAlarmeMsgSetorSensor()
        return
    }

    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Setor?`,
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
                url: `GerenciarSetoresAlarmeExcluir`,
                method: 'Post',
                data: JSON.stringify({ idSetor: id }
                )
            }).fail(function (e) {
                boxErro('Erro ao excluir o Setor')
            }).done(function (r) {
                boxDeletadoSucesso()
                GerenciarSetoresAlarmeLimparCampos()
            })
        }
    })
}

function GerenciarSetoresAlarmeValidarCamposBranco() {//
    if ($('#id-cliente').val() == "0") {
        boxAdvertenciaAuto("Selecione um cliente primeiro")
        $('#busca-cliente').focus()
        return true
    }

    if ($('#id-dispositivo').val() == "0") {
        boxAdvertenciaAuto("Selecione um dispositivo primeiro")
        $('#id-dispositivo').focus()
        return true
    }

    if ($('#numero-setor-alarme').val() == "") {
        boxAdvertenciaAuto("O campo Setor, não poder ficar em branco")
        $('#numero-setor-alarme').focus()
        return true
    }

    if ($('#nome-setor-alarme').val() == "") {
        boxAdvertenciaAuto("O campo Nome, não poder ficar em branco")
        $('#nome-setor-alarme').focus()
        return true
    }

    if ($('#descricao-setor-alarme').val() == "") {
        boxAdvertenciaAuto("O campo Descrição, não poder ficar em branco")
        $('#descricao-setor-alarme').focus()
        return true
    }
}




$(document).ready(function () {

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



    ajustaTabela()
    GerenciarSetoresAlarmeClientesListar()
    GerenciarSetoresAlarmeBloquearIdDispositivo()
})


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

    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.numero}</td> 
        <td class="text-uppercase">${i.particao}</td> 
        <td class="text-uppercase">${i.nome}</td>
        <td class="text-uppercase">${i.tipo}</td>
        
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idSetor}
                tipo="editar"
                title="Editar dados do Setor do Alarme ">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
      
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idSetor}
                tipo="excluir"
                title="Exclui o Setor do Alarme">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>
        <td>
            <button class="btn btn-sm ${btnAtivoCor} py-0"
                id=${i.idSetor}
                tipo="habilitar"
                status="${i.ativo}"
                title="Habilita/Desabilita o Setor do Alarme">
                ${btnAtivoImg}
            </button>
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
function GerenciarSetoresAlarmeClientesListar() {

    $.ajax({
        url: '/GerenciarSetoresAlarmeListarClientes',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {

        $('#id-cliente').empty()
        $('#id-cliente').append('<option value="0">SELECIONE UM CLIENTE</option>')

        r.dados.forEach(i => {
            $('#id-cliente').append(`
                <option value="${i.idCliente}">${i.nome}</option>
            `)
        });

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
    $('#setor-com-camera').attr('disabled', true)
    $('#tipo-area-setor-alarme').attr('disabled', true)
    $('#nome-setor-alarme').attr('disabled', true)
    $('#descricao-setor-alarme').attr('disabled', true)
}

function GerenciarSetoresAlarmeDesbloquearCampos() {

    $('#numero-setor-alarme').attr('disabled', false)
    $('#setor-com-camera').attr('disabled', false)
    $('#tipo-area-setor-alarme').attr('disabled', false)
    $('#nome-setor-alarme').attr('disabled', false)
    $('#descricao-setor-alarme').attr('disabled', false)
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
    localStorage.setItem('idSetorAlarme', '')
    $('#numero-setor-alarme').val("")
    GerenciarSetoresAlarmeBloquearIdDispositivo()
}

function GerenciarSetoresAlarmeLimparCampos() {
    localStorage.setItem('idSetorAlarme', '')

    $('#numero-setor-alarme').val('')
    $('#setor-com-camera').val('N')
    $('#tipo-area-setor-alarme').val('ÁREA INTERNA')
    $('#nome-setor-alarme').val('')
    $('#descricao-setor-alarme').val('')
    GerenciarSetoresAlarmeCarregarTabela()
}

function GerenciarSetoresAlarmeHabilitar(id) {

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
        data: JSON.stringify(
            {
                idDispositivo: $('#id-dispositivo').val(),
                particao: addZeroEsquerda(particao, 2),
                numero: $('#numero-setor-alarme').val(),
                camera: $('#setor-com-camera').val(),
                tipo: $('#tipo-area-setor-alarme').val(),
                nome: $('#nome-setor-alarme').val(),
                descricao: $('#descricao-setor-alarme').val(),
            }
        )
    }).fail(function (e) {
        boxErro("Erro ao inserir o Setor")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarSetoresAlarmeLimparCampos()
    })
}

function GerenciarSetoresAlarmeAlterar(id) {
   
    if (GerenciarSetoresAlarmeValidarCamposBranco()) return
    const particao = $('#id-dispositivo :selected').attr('particao')
    $.ajax({
        start: boxProcessando(),
        url: 'GerenciarSetoresAlarmeAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idSetor: localStorage.getItem('idSetorAlarme'),
                particao: addZeroEsquerda(particao, 2),
                numero: $('#numero-setor-alarme').val(),
                camera: $('#setor-com-camera').val(),
                tipo: $('#tipo-area-setor-alarme').val().toUpperCase(),
                nome: $('#nome-setor-alarme').val().toUpperCase(),
                descricao: $('#descricao-setor-alarme').val().toUpperCase(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao alterar dados do Setor")
    }).done(function (r) {
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
        localStorage.setItem('setorAtual', d.numero)
        localStorage.setItem('idSetorAlarme', d.idSetor)
       
        $('#numero-setor-alarme').val(d.numero)
        $('#setor-com-camera').val(d.camera)
        $('#tipo-area-setor-alarme').val(d.tipo)
        $('#nome-setor-alarme').val(d.nome)
        $('#descricao-setor-alarme').val(d.descricao)
    })
}

function GerenciarSetoresAlarmeExcluir(id) {
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
        $('#id-cliente').focus()
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




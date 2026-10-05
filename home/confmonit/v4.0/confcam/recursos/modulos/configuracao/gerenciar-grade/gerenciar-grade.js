$(document).ready(function () {

    localStorage.setItem('idGrade', '')
    ajustaTabela()
    $(window).on('resize', function () {
        ajustaTabela()
    })

    $('#btn-limpar').on('click', GerenciarGradeLimparFormulario)
    $('#btn-gravar').on('click', GerenciarGradeGravarDados)
    $("#numero-setor-alarme").on('blur', function () {
        this.value = ("000" + this.value).slice(-3)
        if ($("#numero-setor-alarme").val() != "") {
            GerenciarGradeVerificarNumeroSetorLivre()
        }
    });
    $('#id-cliente').on('change', listarDispositivos)

    $('#idDispositivo').on('change', GerenciarGradeListar)

    GerenciarGradeClienteListar()
})

function listarDispositivos() {
    $.ajax({
        url: '/listarDispositivoByIdCliente',
        method: 'Post',
        data: JSON.stringify({ idCliente: $("#id-cliente").val() })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar os dispositivos")
    }).done(function (r) {
        console.log("disp", r)
        $('#idDispositivo').empty()
        $('#idDispositivo').append('<option value="0">SELECIONE</option>')

        r.dados.forEach(i => {
            $('#idDispositivo').append(`
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

// OK
function GerenciarGradeLimparFormulario() {

    localStorage.setItem('idGrade', '')

    $('#formulario').each(function () {
        this.reset();
    })
    $('#id-setor-alarme').val("")
    $('#numero-setor-alarme').val("")
    GerenciarGradeListar()
}

function GerenciarGradeGravarDados() {
    if ($('#nome-grade').val() == '') {
        boxAdvertenciaCampoAuto("Um Nome de ser informado para grade", '#nome-grade')
        return
    }
    const id = localStorage.getItem('idGrade')
    if (id == "") {
        GerenciarGradeInserir()
    } else {
        GerenciarGradeAlterar(id)
    }
}

// OK
function GerenciarGradeListar() {
    const idCliente = $("#id-cliente").val()
    const idDispositivo = $("#idDispositivo").val()

    // Caso idCliente for igual a 0 siginifica que nenhum cliente
    // foi selecionado entao limpa a tabela e retorna 
    if (idCliente == 0) {
        $('tbody').empty()
        GerenciarGradeBloquearCampos()
        return
    }
    GerenciarGradeDesbloquearCampos()

    $.ajax({
        url: '/GerenciarGradeListar',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        // console.log(r)
        gGerenciarSetoresAlarmeTabela = ""
        $('tbody').empty()
        if (r.status != 'Vazio') {
            gGerenciarSetoresAlarmeTabela = r.dados
            r.dados.forEach(i => {
                GerenciarGradeMontaLinha(i)
            });
            GerenciarGradeAssociaBotoesTabela()
        }
    })
}

// OK
function GerenciarGradeMontaLinha(i) {

    let btnAtivoImg
    let btnAtivoCor
    if (i.ativo == 'S') {
        btnAtivoImg = '<i class="bi bi-toggle-on"></i>'
        btnAtivoCor = 'btn-success'
    } else {
        btnAtivoImg = '<i class="bi bi-toggle-off"></i>'
        btnAtivoCor = 'btn-secondary'
    }

    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.nome}</td>
        <td class="text-uppercase">${i.tolerancia} MINUTOS</td>

        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idGrade}
                tipo="editar"
                title="Editar dados do Setor do Alarme ">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
      
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idGrade}
                tipo="excluir"
                title="Exclui o Setor do Alarme">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>

        <td>
            <button class="btn btn-sm ${btnAtivoCor} py-0"
                id=${i.idGrade}
                tipo="habilitar"
                status="${i.ativo}"
                title="Habilita/Desabilita o Setor do Alarme">
                ${btnAtivoImg}
            </button>
        </td>
    </tr>
    `)
}

// OK
function GerenciarGradeAssociaBotoesTabela() {
    $('button[tipo=editar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarGradeBuscar(id)
    })

    $('button[tipo=excluir]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarGradeExcluir(id)

    })

    $('button[tipo=habilitar]').on('click', function () {
        const id = this.getAttribute("id")
        GerenciarGradeHabilitar(id)
    })
}

// OK
function GerenciarGradeHabilitar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarGradeHabilitar',
        method: 'Post',
        data: JSON.stringify({ idGrade: id })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao habilitar/desabilitar uma grade")
    }).done(function (r) {
        boxFechar()
        GerenciarGradeListar()
    })
}


function GerenciarGradeInserir() {

    if (GerenciarGradeValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarGradeInserir',
        method: 'Post',
        data: JSON.stringify(
            {
                idDispositivo: $('#idDispositivo').val(),
                tolerancia: $('#tolerancia-grade').val(),
                nome: $('#nome-grade').val(),
                domEam: $('#dom-eam-grade').val(),
                domSam: $('#dom-sam-grade').val(),
                domEpm: $('#dom-epm-grade').val(),
                domSpm: $('#dom-spm-grade').val(),
                segEam: $('#seg-eam-grade').val(),
                segSam: $('#seg-sam-grade').val(),
                segEpm: $('#seg-epm-grade').val(),
                segSpm: $('#seg-spm-grade').val(),
                terEam: $('#ter-eam-grade').val(),
                terSam: $('#ter-sam-grade').val(),
                terEpm: $('#ter-epm-grade').val(),
                terSpm: $('#ter-spm-grade').val(),
                quaEam: $('#qua-eam-grade').val(),
                quaSam: $('#qua-sam-grade').val(),
                quaEpm: $('#qua-epm-grade').val(),
                quaSpm: $('#qua-spm-grade').val(),
                quiEam: $('#qui-eam-grade').val(),
                quiSam: $('#qui-sam-grade').val(),
                quiEpm: $('#qui-epm-grade').val(),
                quiSpm: $('#qui-spm-grade').val(),
                sexEam: $('#sex-eam-grade').val(),
                sexSam: $('#sex-sam-grade').val(),
                sexEpm: $('#sex-epm-grade').val(),
                sexSpm: $('#sex-spm-grade').val(),
                sabEam: $('#sab-eam-grade').val(),
                sabSam: $('#sab-sam-grade').val(),
                sabEpm: $('#sab-epm-grade').val(),
                sabSpm: $('#sab-spm-grade').val(),

            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o grade")
    }).done(function (r) {
        boxInseridoSucesso(r.dados)
        GerenciarGradeLimparFormulario()
    })
}

function GerenciarGradeBuscar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarGradeBuscar',
        method: 'Post',
        data: JSON.stringify({ idGrade: id })
    }).fail(function (e) {
        //console.log(e)
        boxErro("Erro ao buscar dados do Grade")
    }).done(function (r) {
        boxFechar()
        const d = r.dados
        gNumeroSetorAlarme = d.numero
        //console.log(d)
        localStorage.setItem('idGrade', d.idGrade)
        $('#tolerancia-grade').val(d.tolerancia)
        $('#nome-grade').val(d.nome)
        $('#dom-eam-grade').val((d.domEam == "00:00") ? "" : d.domEam)
        $('#dom-sam-grade').val((d.domSam == "00:00") ? "" : d.domSam)
        $('#dom-epm-grade').val((d.domEpm == "00:00") ? "" : d.domEpm)
        $('#dom-spm-grade').val((d.domSpm == "00:00") ? "" : d.domSpm)
        $('#seg-eam-grade').val((d.segEam == "00:00") ? "" : d.segEam)
        $('#seg-sam-grade').val((d.segSam == "00:00") ? "" : d.segSam)
        $('#seg-epm-grade').val((d.segEpm == "00:00") ? "" : d.segEpm)
        $('#seg-spm-grade').val((d.segSpm == "00:00") ? "" : d.segSpm)
        $('#ter-eam-grade').val((d.terEam == "00:00") ? "" : d.terEam)
        $('#ter-sam-grade').val((d.terSam == "00:00") ? "" : d.terSam)
        $('#ter-epm-grade').val((d.terEpm == "00:00") ? "" : d.terEpm)
        $('#ter-spm-grade').val((d.terSpm == "00:00") ? "" : d.terSpm)
        $('#qua-eam-grade').val((d.quaEam == "00:00") ? "" : d.quaEam)
        $('#qua-sam-grade').val((d.quaSam == "00:00") ? "" : d.quaSam)
        $('#qua-epm-grade').val((d.quaEpm == "00:00") ? "" : d.quaEpm)
        $('#qua-spm-grade').val((d.quaSpm == "00:00") ? "" : d.quaSpm)
        $('#qui-eam-grade').val((d.quiEam == "00:00") ? "" : d.quiEam)
        $('#qui-sam-grade').val((d.quiSam == "00:00") ? "" : d.quiSam)
        $('#qui-epm-grade').val((d.quiEpm == "00:00") ? "" : d.quiEpm)
        $('#qui-spm-grade').val((d.quiSpm == "00:00") ? "" : d.quiSpm)
        $('#sex-eam-grade').val((d.sexEam == "00:00") ? "" : d.sexEam)
        $('#sex-sam-grade').val((d.sexSam == "00:00") ? "" : d.sexSam)
        $('#sex-epm-grade').val((d.sexEpm == "00:00") ? "" : d.sexEpm)
        $('#sex-spm-grade').val((d.sexSpm == "00:00") ? "" : d.sexSpm)
        $('#sab-eam-grade').val((d.sabEam == "00:00") ? "" : d.sabEam)
        $('#sab-sam-grade').val((d.sabSam == "00:00") ? "" : d.sabSam)
        $('#sab-epm-grade').val((d.sabEpm == "00:00") ? "" : d.sabEpm)
        $('#sab-spm-grade').val((d.sabSpm == "00:00") ? "" : d.sabSpm)
    })
}

function GerenciarGradeAlterar(id) {
    if (GerenciarGradeValidarCamposBranco()) return

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarGradeAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idGrade: id,
                tolerancia: $('#tolerancia-grade').val(),
                nome: $('#nome-grade').val(),
                domEam: $('#dom-eam-grade').val(),
                domSam: $('#dom-sam-grade').val(),
                domEpm: $('#dom-epm-grade').val(),
                domSpm: $('#dom-spm-grade').val(),
                segEam: $('#seg-eam-grade').val(),
                segSam: $('#seg-sam-grade').val(),
                segEpm: $('#seg-epm-grade').val(),
                segSpm: $('#seg-spm-grade').val(),
                terEam: $('#ter-eam-grade').val(),
                terSam: $('#ter-sam-grade').val(),
                terEpm: $('#ter-epm-grade').val(),
                terSpm: $('#ter-spm-grade').val(),
                quaEam: $('#qua-eam-grade').val(),
                quaSam: $('#qua-sam-grade').val(),
                quaEpm: $('#qua-epm-grade').val(),
                quaSpm: $('#qua-spm-grade').val(),
                quiEam: $('#qui-eam-grade').val(),
                quiSam: $('#qui-sam-grade').val(),
                quiEpm: $('#qui-epm-grade').val(),
                quiSpm: $('#qui-spm-grade').val(),
                sexEam: $('#sex-eam-grade').val(),
                sexSam: $('#sex-sam-grade').val(),
                sexEpm: $('#sex-epm-grade').val(),
                sexSpm: $('#sex-spm-grade').val(),
                sabEam: $('#sab-eam-grade').val(),
                sabSam: $('#sab-sam-grade').val(),
                sabEpm: $('#sab-epm-grade').val(),
                sabSpm: $('#sab-spm-grade').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao alterar dados do Grade")
    }).done(function (r) {
        boxAteradoSucesso()
        GerenciarGradeLimparFormulario()
    })
}

function GerenciarGradeExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar esta Grade?`,
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
                url: `/GerenciarGradeExcluir`,
                method: 'Post',
                data: JSON.stringify({ idGrade: id })
            }).fail(function (e) {
                console.log(e)
                boxErro('Erro ao excluir o Setor')
            }).done(function (r) {
                GerenciarGradeLimparFormulario()
                boxDeletadoSucesso()
            })
        }
    })
}

function GerenciarGradeClienteListar() {
    $.ajax({
        url: '/GerenciarGradeClienteListar',
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

function GerenciarGradeValidarCamposBranco() {
    // if ($('#id-cliente').val() == "0") {
    //     boxAdvertenciaAuto("Selecione um cliente primeiro")
    //     $('#id-cliente').focus()
    //     return true
    // }

}

function GerenciarGradeBloquearCampos() {
    $('#tolerancia-grade').attr('disabled', true)
    $('#nome-grade').attr('disabled', true)
    $('#dom-eam-grade').attr('disabled', true)
    $('#dom-sam-grade').attr('disabled', true)
    $('#dom-epm-grade').attr('disabled', true)
    $('#dom-spm-grade').attr('disabled', true)
    $('#seg-eam-grade').attr('disabled', true)
    $('#seg-sam-grade').attr('disabled', true)
    $('#seg-epm-grade').attr('disabled', true)
    $('#seg-spm-grade').attr('disabled', true)
    $('#ter-eam-grade').attr('disabled', true)
    $('#ter-sam-grade').attr('disabled', true)
    $('#ter-epm-grade').attr('disabled', true)
    $('#ter-spm-grade').attr('disabled', true)
    $('#qua-eam-grade').attr('disabled', true)
    $('#qua-sam-grade').attr('disabled', true)
    $('#qua-epm-grade').attr('disabled', true)
    $('#qua-spm-grade').attr('disabled', true)
    $('#qui-eam-grade').attr('disabled', true)
    $('#qui-sam-grade').attr('disabled', true)
    $('#qui-epm-grade').attr('disabled', true)
    $('#qui-spm-grade').attr('disabled', true)
    $('#sex-eam-grade').attr('disabled', true)
    $('#sex-sam-grade').attr('disabled', true)
    $('#sex-epm-grade').attr('disabled', true)
    $('#sex-spm-grade').attr('disabled', true)
    $('#sab-eam-grade').attr('disabled', true)
    $('#sab-sam-grade').attr('disabled', true)
    $('#sab-epm-grade').attr('disabled', true)
    $('#sab-spm-grade').attr('disabled', true)
}

function GerenciarGradeDesbloquearCampos() {
    $('#tolerancia-grade').attr('disabled', false)
    $('#nome-grade').attr('disabled', false)
    $('#dom-eam-grade').attr('disabled', false)
    $('#dom-sam-grade').attr('disabled', false)
    $('#dom-epm-grade').attr('disabled', false)
    $('#dom-spm-grade').attr('disabled', false)
    $('#seg-eam-grade').attr('disabled', false)
    $('#seg-sam-grade').attr('disabled', false)
    $('#seg-epm-grade').attr('disabled', false)
    $('#seg-spm-grade').attr('disabled', false)
    $('#ter-eam-grade').attr('disabled', false)
    $('#ter-sam-grade').attr('disabled', false)
    $('#ter-epm-grade').attr('disabled', false)
    $('#ter-spm-grade').attr('disabled', false)
    $('#qua-eam-grade').attr('disabled', false)
    $('#qua-sam-grade').attr('disabled', false)
    $('#qua-epm-grade').attr('disabled', false)
    $('#qua-spm-grade').attr('disabled', false)
    $('#qui-eam-grade').attr('disabled', false)
    $('#qui-sam-grade').attr('disabled', false)
    $('#qui-epm-grade').attr('disabled', false)
    $('#qui-spm-grade').attr('disabled', false)
    $('#sex-eam-grade').attr('disabled', false)
    $('#sex-sam-grade').attr('disabled', false)
    $('#sex-epm-grade').attr('disabled', false)
    $('#sex-spm-grade').attr('disabled', false)
    $('#sab-eam-grade').attr('disabled', false)
    $('#sab-sam-grade').attr('disabled', false)
    $('#sab-epm-grade').attr('disabled', false)
    $('#sab-spm-grade').attr('disabled', false)
}

function GerenciarGradeCorrigeCampo(campo) {
    if (campo == "") {
        return '00:00'
    }
    return campo
}



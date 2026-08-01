var listaModalidades = []
var LABEL_NOVO_EVENTO = 'CRIAR UM NOVO EVENTO'

$(document).ready(function () {

    localStorage.setItem('idContactid', '')

    ajustaTabela()
    $(window).on('resize', ajustaTabela)


    $('#btn-limpar').on('click', limparFormulario)
    $('#btn-gravar').on('click', gravarDados)

    // Pesquisa da modalidade do evento
    $('#busca-modalidade').on('input', filtrarModalidades)
    $('#busca-modalidade').on('focus', function () {
        if (!$(this).prop('disabled')) filtrarModalidades()
    })
    $('#lista-modalidades').on('click', '.fp-combo-item', function () {
        $('#busca-modalidade').val($(this).attr('data-nome'))
        $('#idContactidPadrao').val($(this).attr('data-id')).trigger('change')
        $('#lista-modalidades').addClass('d-none')
    })
    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-modalidade, #lista-modalidades').length) {
            $('#lista-modalidades').addClass('d-none')
        }
    })


    $('#idContactidPadrao').on('change', function () {
        $('#nivel').val('0')
        $('#grupo').val('ALARME')
        $('#tipo').val('')
        $('#codigo').val('')
        $('#descricao').val('')

        if ($('#idContactidPadrao').val() == 'NOVO') {
            $('#grupo').attr('disabled', false)
            $('#tipo').attr('disabled', false)
            $('#codigo').attr('disabled', false)
            $('#descricao').attr('disabled', false)
        } else {
            $('#grupo').attr('disabled', true)
            $('#tipo').attr('disabled', true)
            $('#codigo').attr('disabled', true)
            $('#descricao').attr('disabled', true)

            $.ajax({
                start: boxProcessando(),
                url: '/buscar',
                method: 'Post',
                data: JSON.stringify({ idContactid: $('#idContactidPadrao').val() })
            }).fail(function (e) {
                console.log(e)
                boxErro("Erro ao buscar dados do Base")
            }).done(function (r) {
                boxFechar()
                const d = r.dados

                $('#tipo').val(d.codigo.substring(0, 1))
                $('#codigo').val(d.codigo.substring(1))
                $('#grupo').val(d.grupo)
                $('#descricao').val(d.descricao)
                $('#nivel').val(d.nivel)
            })

        }
    })

    carregarGrupos()
    listarContacid()
    modalidadeEventoListar()
})

function carregarGrupos() {
    $.ajax({
        url: '/carregarGrupo',
        method: 'Post',
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar grupos contact id")
    }).done(function (r) {
        // console.log(r)
        r.dados.forEach(i => {
            $('#grupo').append(`
                <option value="${i.grupo}">${i.grupo}</option>
            `)
        });

    }) 
}

function filtrarModalidades() {
    var raw = ($('#busca-modalidade').val() || '').toLowerCase().trim()
    var lista = $('#lista-modalidades')
    lista.empty()

    // Opção "CRIAR UM NOVO EVENTO" sempre disponível
    $('<li class="fp-combo-item"></li>')
        .attr('data-id', 'NOVO')
        .attr('data-nome', LABEL_NOVO_EVENTO)
        .html('<strong>' + LABEL_NOVO_EVENTO + '</strong>')
        .appendTo(lista)

    var base
    if (raw == '' || raw == LABEL_NOVO_EVENTO.toLowerCase()) {
        base = listaModalidades.slice(0, 30)
    } else {
        base = listaModalidades.filter(function (i) {
            return (i.descricao || '').toLowerCase().indexOf(raw) != -1
        }).slice(0, 40)
    }

    base.forEach(function (i) {
        $('<li class="fp-combo-item"></li>')
            .attr('data-id', i.idContactid)
            .attr('data-nome', i.descricao)
            .text(i.descricao)
            .appendTo(lista)
    })

    lista.removeClass('d-none')
}

function setModalidadeDisplay(id) {
    if (id == 'NOVO' || !id) {
        $('#idContactidPadrao').val('NOVO')
        $('#busca-modalidade').val(LABEL_NOVO_EVENTO)
        return
    }
    var m = listaModalidades.filter(function (i) { return String(i.idContactid) == String(id) })[0]
    $('#idContactidPadrao').val(id)
    $('#busca-modalidade').val(m ? m.descricao : '')
}

function limparFormulario() {
    $('#formulario').each(function () {
        this.reset();
    })

    localStorage.setItem('idContactid', '')

    setModalidadeDisplay('NOVO')
    $('#busca-modalidade').prop('disabled', false)
    $('#lista-modalidades').empty().addClass('d-none')
    $('#idContactidPadrao').attr('disabled', false)
    $('#grupo').attr('disabled', false)
    $('#tipo').attr('disabled', false)
    $('#codigo').attr('disabled', false)
    $('#descricao').attr('disabled', false)
}

function gravarDados() {
    const id = localStorage.getItem('idContactid')
    if (id == "") {
        inserir()
    } else {
        alterar(id)
    }
}

function listarContacid() {

    $.ajax({
        url: '/listarContacid',
        method: 'Post',
        data: JSON.stringify({ idVinculo: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        $('tbody').empty()
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                montaLinha(i)
            });

            $('button[tipo=editar]').on('click', function () {
                const id = this.getAttribute("id")
                buscar(id)
            })

            $('button[tipo=excluir]').on('click', function () {
                const id = this.getAttribute("id")
                excluir(id)

            })
        }
    })
}

function montaLinha(i) {
    let descricao

    if (i.descricao.length > 60) {
        descricao = i.descricao.substring(0, 57) + "..."
    } else {
        descricao = i.descricao
    }



    $('tbody').append(`
        <tr>
        <td class="text-uppercase">${i.codigo}</td>
        <td class="text-uppercase" title="${i.descricao.toUpperCase()}">${descricao}</td>
        <td class="text-uppercase">${i.nivel}</td>
        <td>
            <button class="btn btn-sm btn-primary py-0"
                id=${i.idContactid}
                tipo="editar"
                title="Editar dados da contactid">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>
      
        <td>
            <button class="btn btn-sm btn-danger py-0"
                id=${i.idContactid}
                tipo="excluir"
                title="Exclui o contactid">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>
        
    </tr>
    `)
}

function inserir() {
    
    $.ajax({
        start: boxProcessando(),
        url: '/inserir',
        method: 'Post',
        data: JSON.stringify(
            {
                codigo: $('#tipo').val() + $('#codigo').val(),
                grupo: $('#grupo').val(),
                descricao: $('#descricao').val(),
                nivel: $('#nivel').val(),
                idVinculo: localStorage.getItem('idFranqueado')
            }
        )

    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao inserir o procedimento")
    }).done(function (r) {
        limparFormulario()
        listarContacid()
        boxInseridoSucesso(r.dados)
    })
}

function buscar(id) {
    $('#idContactidPadrao').attr('disabled', true)
    $('#busca-modalidade').prop('disabled', true)
    $('#lista-modalidades').addClass('d-none')
    $('#grupo').attr('disabled', true)
    $('#tipo').attr('disabled', true)
    $('#codigo').attr('disabled', true)

    $.ajax({
        start: boxProcessando(),
        url: '/buscar',
        method: 'Post',
        data: JSON.stringify({ idContactid: id })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao buscar dados do Base")
    }).done(function (r) {
        boxFechar()
        const d = r.dados
        console.log(d)

        localStorage.setItem('idContactid', d.idContactid)
        $('#idContactidPadrao').val(d.idContactid)
        $('#busca-modalidade').val(d.descricao)
        $('#tipo').val(d.codigo.substring(0, 1))
        $('#codigo').val(d.codigo.substring(1))
        $('#grupo').val(d.grupo)
        $('#descricao').val(d.descricao)
        $('#nivel').val(d.nivel)
    })
}

function alterar(id) {

    if (validarCamposBranco()) return  
    
    $.ajax({
        start: boxProcessando(),
        url: '/alterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idContactid: localStorage.getItem('idContactid'),
                codigo: $('#tipo').val() + $('#codigo').val(),
                grupo: $('#grupo').val(),
                descricao: $('#descricao').val(),
                nivel: $('#nivel').val(),
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao alterar dados do procedimento")
    }).done(function (r) {
        limparFormulario()
        listarContacid()
        boxAteradoSucesso()
    })

}

function excluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Contactid?`,
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
                url: `/excluir`,
                method: 'Post',
                data: JSON.stringify({ idContactid: id })
            }).fail(function (e) {
                console.log(e)
                boxErro('Erro ao excluir o procedimento')
            }).done(function (r) {
                limparFormulario()
                boxDeletadoSucesso()
                listarContacid()
            })
        }
    })
}

function validarCamposBranco() {
    if ($('#codigo').val() == "0") {
        boxAdvertenciaAuto("Um codigo de evento deve ser informado")
        $('#codigo').focus()
        return true
    }
    if ($('#descricao').val() == "") {
        boxAdvertenciaAuto("O campo Decricao para o evento deve ser informada")
        $('#descricao').focus()
        return true
    }

}

function modalidadeEventoListar() {

    $.ajax({
        url: 'modalidadeEventoListar',
        method: 'Post',
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar contact id")
    }).done(function (r) {
        listaModalidades = (r.status == 'OK' && r.dados) ? r.dados : []
        if (!localStorage.getItem('idContactid')) {
            setModalidadeDisplay('NOVO')
        }
    })
}


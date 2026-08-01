var listaClientes = []
var listaEventos = []
var eventoSelecionado = null

$(document).ready(function () {
    $('#btnGerar').on('click', geradorEventoEnviar)
    $('#btnLimpar').on('click', geradorEventoLimpar)
    geradorEventoClientesListar()

    // Associa o eveto de selecionar do idCliente
    $('#cliente').on('change', geradorEventoClienteChange)
    $('#dispositivo').on('change', geradorEventoDispositivoChange)
    $('#evento').on('change', geradorEventoEventoChange)
    $('#zoneUser').on('change', geradorEventoZoneUserChange)

    // Pesquisa de cliente (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', geradorFiltrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') geradorFiltrarClientes()
    })
    $('#lista-clientes').on('click', '.fp-combo-item', function () {
        $('#busca-cliente').val($(this).attr('data-nome'))
        $('#cliente').val($(this).attr('data-id')).trigger('change')
        $('#lista-clientes').addClass('d-none')
    })

    // Pesquisa de evento (descricao, codigo)
    $('#busca-evento').on('input', geradorFiltrarEventos)
    $('#busca-evento').on('focus', function () {
        if ($(this).val().trim() != '') geradorFiltrarEventos()
    })
    $('#lista-eventos').on('click', '.fp-combo-item', function () {
        var ev = $(this).data('evento')
        if (!ev) return
        eventoSelecionado = ev
        $('#busca-evento').val(ev.codigo + ' — ' + ev.descricao)
        $('#evento').val(ev.codigo).trigger('change')
        $('#lista-eventos').addClass('d-none')
    })

    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-cliente, #lista-clientes').length) {
            $('#lista-clientes').addClass('d-none')
        }
        if (!$(e.target).closest('#busca-evento, #lista-eventos').length) {
            $('#lista-eventos').addClass('d-none')
        }
    })
});

function clienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function geradorFiltrarClientes() {
    var raw = ($('#busca-cliente').val() || '').toLowerCase().trim()
    var termoNum = raw.replace(/[^a-z0-9]/g, '')
    var lista = $('#lista-clientes')
    lista.empty()

    if (raw == '') {
        $('#cliente').val('0').trigger('change')
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

function geradorFiltrarEventos() {
    var raw = ($('#busca-evento').val() || '').toLowerCase().trim()
    var lista = $('#lista-eventos')
    lista.empty()

    if (raw == '') {
        lista.addClass('d-none')
        return
    }

    var resultados = listaEventos.filter(function (i) {
        var texto = ((i.descricao || '') + ' ' + (i.codigo || '')).toLowerCase()
        return texto.indexOf(raw) != -1
    }).slice(0, 40)

    if (resultados.length == 0) {
        lista.append('<li class="fp-combo-empty">Nenhum evento encontrado</li>')
    } else {
        resultados.forEach(function (i) {
            $('<li class="fp-combo-item"></li>')
                .text(i.codigo + ' — ' + i.descricao)
                .data('evento', i)
                .appendTo(lista)
        })
    }
    lista.removeClass('d-none')
}

function resetEvento() {
    listaEventos = []
    eventoSelecionado = null
    $('#evento').val('0')
    $('#busca-evento').val('')
    $('#lista-eventos').empty().addClass('d-none')
}

function geradorEventoEnviar() {

    if ($('#cliente').val() == "0") {
        alert('Um cliente deve ser selecionado')
        return
    }

    if ($('#dispositivo').val() == "0") {
        alert('Um dispositivo deve ser selecionado')
        return
    }

    if ($('#evento').val() == "0"){
        alert('Um evento deve ser selecionado')
        return
    }

    if ($('#zoneUser').val() == "0"){
        alert('Uma setor ou usuario deve ser selecionado')
        return
    }
    
    const idFranq = localStorage.getItem('idFranqueado')
    const conta = $('#dispositivo').find(':selected').attr('conta')
    const codigo = $('#evento').val()
    const particao = addZeroEsquerda($('#dispositivo').find(':selected').attr('particao'), 2)
    const zonaUser = $('#zoneUser').val()
    const senha = "WHdQkY&RX%W%4RArwm1Q"

    if ($('#evento').val() == "") {
        alert('Um evento deve ser informado')
        return
    }
    // <!-- CCCCENNNPPZZZ -->
    const evento = conta + codigo + particao + zonaUser  
    
    $.ajax({
        start: boxProcessando(),
        url: 'geradorEventoEnviar',
        method: 'Post',
        data: JSON.stringify(
            {
                idFranqueado: idFranq,
                evento: evento,
                senha: senha
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro('Erro ao gerar o evento')
    }).done(function (r) {
        boxMensagemAuto('Evento gerado com sucesso')
    })

}

function geradorEventoLimpar() {
    const selecione = '<option value="0">SELECIONE</option>'
    $('#busca-cliente').val('')
    $('#cliente').val('0')
    $('#lista-clientes').empty().addClass('d-none')
    $('#dispositivo').empty().append(selecione)
    resetEvento()
    $('#zoneUser').empty().append(selecione)
}

function geradorEventoClienteChange() {
    const idCliente = $('#cliente').val()

    const selecione = '<option value="0">SELECIONE</option>'
    $('#dispositivo').empty().append(selecione)
    resetEvento()
    $('#zoneUser').empty().append(selecione)

    if (idCliente != '0') {
        geradorEventoDispositivoListar()
    }
}

function geradorEventoDispositivoChange() {
    const idDisp = $('#dispositivo').val()

    const selecione = '<option value="0">SELECIONE</option>'
    resetEvento()
    $('#zoneUser').empty().append(selecione)

    if (idDisp != '0') {
        geradorEventoContacidListar()
    }
}

function geradorEventoEventoChange() {
    const evento = $('#evento').val()

    const selecione = '<option value="0">SELECIONE</option>'
    $('#zoneUser').empty().append(selecione)

    if (evento != '0' && eventoSelecionado) {
        const grupo = eventoSelecionado.grupo
        if (grupo == "ARME" || grupo == "DESARME") {
            geradorEventoUsuarioListar()
        } else {
            geradorEventoSetorListar()
        }
    }
}

function geradorEventoZoneUserChange() {

}

function geradorEventoClientesListar() {

    $.ajax({
        url: '/geradorEventoClientesListar',
        method: 'Post',
        data: JSON.stringify({ idFranqueado: localStorage.getItem('idFranqueado') })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        listaClientes = (r.dados || [])
        $('#cliente').val('0')
    })
}

function geradorEventoDispositivoListar() {
    $.ajax({
        url: '/geradorEventoDispositivoListar',
        method: 'Post',
        data: JSON.stringify({ idCliente: $("#cliente").val() })
    }).fail(function (e) {
        boxErro("Erro ao carregar os dispositivos")
    }).done(function (r) {
        console.log(r)
        $('#dispositivo').empty()
        $('#dispositivo').append('<option value="0">SELECIONE</option>')

        r.dados.forEach(i => {
            if (i.ativo == "S") {
                let nome, idDisp
                if (i.manutencao == "") {
                    idDisp = i.idDispositivo
                    nome = i.nome
                } else {
                    idDisp = '0'
                    nome = i.nome + ' (MANUTEÇÃO)'
                }

                $('#dispositivo').append(`
                    <option 
                        value="${idDisp}" 
                        idFranqueado="${i.idFranqueado}"
                        particao="${i.particao}"
                        conta="${i.conta}"
                        mac="${i.idFisico1}"
                        senha="${i.senha}"
                        tipo="${i.tipo}"
                    >${nome}</option>
                `)
            }
        })


        // Desbloquea o dispositivo
        $('#dispositivo').attr('disabled', false)
    })
}

function geradorEventoContacidListar() {
    $.ajax({
        url: '/geradorEventoContacidListar',
        method: 'Post',
        data: JSON.stringify({ idCliente: $("#cliente").val() })
    }).fail(function (e) {
        boxErro("Erro ao carregar os dispositivos")
    }).done(function (r) {
        listaEventos = (r.dados || [])
        eventoSelecionado = null
        $('#evento').val('0')
        $('#busca-evento').val('')
        $('#lista-eventos').empty().addClass('d-none')

        // Desbloquea o dispositivo
        $('#dispositivo').attr('disabled', false)
    })
}

function geradorEventoUsuarioListar() {
    $.ajax({
        url: '/geradorEventoUsuarioListar',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: $("#dispositivo").val() })
    }).fail(function (e) {
        boxErro("Erro ao carregar os dispositivos")
    }).done(function (r) {
        console.log(r)
        $('#zoneUser').empty()
        $('#zoneUser').append('<option value="0">SELECIONE</option>')

        if (r.status == 'OK') {
            r.dados.forEach(i => {
                if (i.ativo == "S") {
                    $('#zoneUser').append(`
                        <option 
                            value="${i.codigo}"
                            
                        >${i.nome}</option>
                    `)
                }
            })
        }
    })
}

function geradorEventoSetorListar() {

    $.ajax({
        url: '/geradorEventoSetorListar',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: $("#dispositivo").val() })
    }).fail(function (e) {
        boxErro("Erro ao carregar os setores")
    }).done(function (r) {
        console.log(r)
        $('#zoneUser').empty().append('<option value="0">SELECIONE</option>')

        if (r.status == 'OK') {
            r.dados.forEach(i => {

                $('#zoneUser').append(`
                    <option value="${i.numero}">${i.nome}</option>
                `)

            })
        }
    })
}
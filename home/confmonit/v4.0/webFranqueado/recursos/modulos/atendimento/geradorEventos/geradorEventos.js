$(document).ready(function () {
    $('#btnGerar').on('click', geradorEventoEnviar)
    $('#btnLimpar').on('click', geradorEventoLimpar)
    geradorEventoClientesListar()

    // Associa o eveto de selecionar do idCliente
    $('#cliente').on('change', geradorEventoClienteChange)
    $('#dispositivo').on('change', geradorEventoDispositivoChange)
    $('#evento').on('change', geradorEventoEventoChange)
    $('#zoneUser').on('change', geradorEventoZoneUserChange)
});

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
    $('#cliente').val('0')
    $('#dispositivo').empty().append(selecione)
    $('#evento').empty().append(selecione)
    $('#zoneUser').empty().append(selecione)
}

function geradorEventoClienteChange() {
    const idCliente = $('#cliente').val()

    const selecione = '<option value="0">SELECIONE</option>'
    $('#dispositivo').empty().append(selecione)
    $('#evento').empty().append(selecione)
    $('#zoneUser').empty().append(selecione)

    if (idCliente != '0') {
        geradorEventoDispositivoListar()
    }
}

function geradorEventoDispositivoChange() {
    const idDisp = $('#dispositivo').val()

    const selecione = '<option value="0">SELECIONE</option>'
    $('#evento').empty().append(selecione)
    $('#zoneUser').empty().append(selecione)

    if (idDisp != '0') {
        geradorEventoContacidListar()
    }
}

function geradorEventoEventoChange() {
    const evento = $('#evento').val()

    const selecione = '<option value="0">SELECIONE</option>'
    $('#zoneUser').empty().append(selecione)

    if (evento != '0') {
        const grupo = $('#evento').find(':selected').attr('grupo')
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
        console.log(r)
        $('#cliente').empty().append('<option value="0">SELECIONE</option>')

        r.dados.forEach(i => {
            $('#cliente').append(`<option value="${i.idCliente}">${i.nome}</option>`)
        });

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
        console.log(r)
        $('#evento').empty()
        $('#evento').append('<option value="0">SELECIONE</option>')

        r.dados.forEach(i => {

            $('#evento').append(`
                    <option 
                        value="${i.codigo}"
                        idContactid="${i.idContactid}"
                        nivel="${i.nivel}"
                        grupo="${i.grupo}"
                        descricao="${i.descricao}"
                    >${i.descricao}</option>
                `)

        })


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
function modCliente_start(retorno, idDispositivo) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/modCliente/modCliente.html'
    $('#boxDir').load(uri, () => {
        $('#modCliente_btnFechar').on('click', retorno)

        modCliente_carregarDados(idDispositivo)
    })
}


function modCliente_carregarDados(idDispositivo) {

    $.ajax({
        url: 'modCliente/buscarDados',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idDispositivo: idDispositivo })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        modCliente_montarLinha(r.dados)
    })
}

function modCliente_montarLinha(item) {
    $('#cliDados_cpConta').val(item.conta)
    $('#cliDados_cpNome').val(item.nome)
    $('#cliDados_cpTelefone').val(item.telefone)
    $('#cliDados_cpCelular').val(item.celular)
    $('#cliDados_cpEndereco').val(item.endereco)
    $('#cliDados_cpBairoo').val(item.bairro)
    $('#cliDados_cpCidade').val(item.cidade)
    $('#cliDados_cpDispositivo').val(item.nomeDisp)

    
    if (item.armado == '1') {
        $('#cliDados_cpArmado').html('<i class="bi bi-lock"></i>')
        $('#cliDados_cpArmado').addClass('bg-success')
        $('#cliDados_btnArmar').addClass('bg-secondary')
        $('#cliDados_btnArmar').attr('idDispositivo', item.idDispositivo)
    } else {
        $('#cliDados_cpArmado').html('<i class="bi bi-unlock"></i>')
        $('#cliDados_cpArmado').addClass('bg-danger')
        $('#cliDados_btnArmar').addClass('btn-success')
        $('#cliDados_btnArmar').attr('bg-secondary', '0')
    }

    $('#cliDados_tabUsuario tbody').empty()
    if (item.usuarios != null) {
        item.usuarios.forEach(i => {
            $('#cliDados_tabUsuario tbody').append(`
                <tr>
                    <td>${i.codigo}</td>
                    <td>${i.nome}</td>
                </tr>
            `)
        });
    }

    $('#cliDados_tabSetores tbody').empty()
    if (item.setores != null) {
        item.setores.forEach(i => {
            $('#cliDados_tabSetores tbody').append(`
                <tr>
                    <td>${i.setor}</td>
                    <td>${i.nome}</td>
                </tr>
            `)
        });
    }
}



function modCliente_btnArmar(){
    
    alert('armando')
    // $.ajax({
    //     url: 'cliDados/buscarDados',
    //     method: 'Post',
    //     headers: {
    //         "Content-Type": "application/json",
    //         "Accept": "application/json",
    //         "Authorization": "Bearer " + sessionStorage.getItem('token')
    //     },
    //     data: JSON.stringify({ idDispositivo: idDispositivo })
    // }).fail(function (e) {
    //     console.log(e)
    // }).done(function (r) {
    //     cliDados_montarLinha(r.dados)
    // })
}

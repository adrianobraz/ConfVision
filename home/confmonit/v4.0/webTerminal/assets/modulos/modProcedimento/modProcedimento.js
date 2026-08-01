function modProcedimento_start(retorno) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/modProcedimento/modProcedimento.html'
    $('#boxDir').load(uri, () => {

        $('#modProcedimento_btnFechar').on('click', retorno)

        $('#modProcedimento_cpGrupo').on('change', () => {
            $('#modProcedimento_cpProcedimento').val($('#modProcedimento_cpGrupo').val())
        })

        modProcedimento_buscarDados()        
    })
}

function modProcedimento_buscarDados(){

    $.ajax({
        url: '/modProcedimento/buscarDados',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: sessionStorage.getItem('ateProDado_proc_idFranqueado'),
            idCliente: sessionStorage.getItem('ateProDado_proc_idCliente')
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status != 'Vazio') {
            $('#modProcedimento_cpGrupo').empty().append('<option value="">SELECIONE UM GRUPO</option>')
            r.dados.forEach(i => {
                $('#modProcedimento_cpGrupo').append(`<option value="${i.procedimento}">${i.grupo}</option>`)                
            })

           
        }else{
            $('#modProcedimento_cpGrupo').empty()   
        }
    })

}
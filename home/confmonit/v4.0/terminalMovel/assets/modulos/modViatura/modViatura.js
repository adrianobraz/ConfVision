function modViatura_start(retorno) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/modViatura/modViatura.html'
    $('#boxDir').load(uri, () => {
        $('#modViatura_btnFechar').on('click', retorno)

        const idFranqueado = sessionStorage.getItem('ateProDado_proc_idFranqueado')
        modViatura_carregarDados(idFranqueado)
    })
}

function modViatura_carregarDados(idFranqueado) {
    $.ajax({
        url: '/modViatura/buscarDados',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({ idFranqueado: idFranqueado })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {
                modViatura_montarLinha(i)
            });
        }

        $(`button[tipo=modViatura_btnLigar]`).on('click', function () {

            if (essionStorage.getItem('login_voipAtivo') == 'S') {
                modViatura_ligar($(this).attr('numero'))
            }

        })


    })
}

function modViatura_montarLinha(item) {
    $('#modViatura_responsivo').append(`
         <div class="modViatura_boxViatura">
            <div class="col-12" >
                <h5 class="bg-success text-bg-success">${item.nome}</h5>
                
            </div>

            <div class="col-12 col-sm-6 my-1 px-1">
                <div class="input-group input-group-sm">

                    <span class="input-group-text bg-success text-bg-success">Modelo:</span>
                
                    <input type="text" class="form-control form-control-sm bg-dark text-bg-dark" disabled value="${item.modelo}">
                </div>
            </div>

            <div class="col-12 col-sm-6 my-1 px-1">
                <div class="input-group input-group-sm">
                    <span class="input-group-text bg-success text-bg-success">Telefone:</span>
                
                    <input type="text" class="form-control form-control-sm bg-dark text-bg-dark" disabled value="${formatarCelular(item.telefone)}">
                
                    <button class="btn btn-success" type="button" tipo="modViatura_btnLigar" numero="${item.telefone}">
                        <i class="bi bi-telephone-outbound-fill"></i> &nbsp;
                        Ligar
                    </button>
                </div>
            </div>

            <div class="col-12 col-sm-6 my-1 px-1">
                <div class="input-group input-group-sm">
                    
                    <span class="input-group-text bg-success text-bg-success">Cor:</span>
                    
                    <input type="text" class="form-control form-control-sm bg-dark text-bg-dark" disabled value="${item.cor}">

                </div>
            </div>

            <div class="col-12 col-sm-6 my-1 px-1">
                <div class="input-group input-group-sm">
                    
                    <span class="input-group-text bg-success text-bg-success">Celular 1:</span>

                    <input type="text" class="form-control form-control-sm bg-dark text-bg-dark" disabled value="${formatarCelular(item.celular1)}">
                    
                    <button class="btn btn-success" type="button" tipo="modViatura_btnLigar" numero="${item.celular1}">
                        <i class="bi bi-telephone-outbound-fill"></i> &nbsp;
                        Ligar
                    </button>
                </div>
            </div>

            <div class="col-12 col-sm-6 my-1 px-1">
                <div class="input-group input-group-sm">

                    <span class="input-group-text bg-success text-bg-success">Placa:</span>
                
                    <input type="text" class="form-control form-control-sm bg-dark text-bg-dark" disabled value="${item.placa}">
                </div>
            </div>

            <div class="col-12 col-sm-6 my-1 px-1">
                <div class="input-group input-group-sm">
                    <span class="input-group-text bg-success text-bg-success">Celular 2:</span>
                    
                    <input type="text" class="form-control form-control-sm bg-dark text-bg-dark" disabled value"${formatarCelular(item.celular2)}">

                    <button class="btn btn-success" type="button" tipo="modViatura_btnLigar" numero="${item.celular2}">
                        <i class="bi bi-telephone-outbound-fill"></i> &nbsp;
                        Ligar
                    </button>
                </div>
            </div>
 
            </div> <!-- /modViatura_boxViatura -->    
    `)



}

function modViatura_ligar(numero) {
    // Verifica novas ligações travada
    if (sessionStorage.getItem('modViatura_ligar') === 'off') {
        msgErro('Já existe um ligação em andamento')
    } else {
        if (numero.length >= 10) {
            $(this).removeClass('btn-success').addClass('btn-danger')

            // Trava novas ligações
            sessionStorage.setItem('modViatura_ligar', 'off')

            const src = sessionStorage.getItem('login_userCelular')
            const dst = numero
            vono_ligar(
                src, 
                dst, 
                sessionStorage.getItem('ateProDado_proc_idCliente'), 
                (resp) => {
                // Retorna a cor do botao para verde
                $(this).removeClass('btn-danger').addClass('btn-success')

                // Libera novas ligações
                sessionStorage.setItem('modViatura_ligar', 'on')

                if (!resp) {
                    msgErro('Ligação falhou')
                }
            })
        }
    }
}

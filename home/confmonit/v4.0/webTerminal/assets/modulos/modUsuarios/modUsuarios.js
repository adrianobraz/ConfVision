function modUsuarios_start(retorno) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/modUsuarios/modUsuarios.html'
    $('#boxDir').load(uri, () => {
        $('#modUsuarios_btnFechar').on('click', () => { retorno() })
        modUsuarios_buscaDados()
    })
}

function modUsuarios_buscaDados() {
    const idDispositivo = sessionStorage.getItem('ateProDado_proc_idDispositivo')
    $.ajax({
        url: '/modUsuarios/buscarDados',
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
        if (r.status != 'Vazio') {
            $('#modUsuarios tbody').empty()
            r.dados.forEach(i => {
                modUsuarios_montaLinha(i)
            });
        }

        $(`td[tipo=modUsuarios_btnLigar]`).on('click', function () {
            if (sessionStorage.getItem('login_voipAtivo') == 'S') {
                modUsuarios_ligar($(this).attr('numero'), this)
            }
        })


    })
}

function modUsuarios_montaLinha(item) {
    const ligaCor = (item.celular == "") ? '' : 'bg-success'
    const click = (item.celular == "") ? '' : 'click'
    $('#modUsuarios tbody').append(`        
        <tr>
            <td class="fmt-tbody">${item.codigo}</td>

            <td class="fmt-tbody">${item.nome}</td>
            
            <td 
                tipo="modUsuarios_btnLigar" 
                numero="${item.celular}" 
                class="${click} ${ligaCor} txt-tbody"
            >${formatarCelular(item.celular)}</td>
        </tr>                
    `)

}


function modUsuarios_ligar(numero, botao) {
   
    // Verifica novas ligações travada
    if (sessionStorage.getItem('modUsuarios_ligar') === 'off') {
        msgErro('Já existe um ligação em andamento')
    } else {
        if (numero.length >= 10) {
            $(botao).removeClass('bg-success').addClass('bg-danger')

            // Trava novas ligações
            sessionStorage.setItem('modUsuarios_ligar', 'off')

            const src = sessionStorage.getItem('login_userCelular')
            const dst = numero

            vono_ligar(
                src, 
                dst, 
                sessionStorage.getItem('ateProDado_proc_idCliente'), 
                (resp) => {
                // Retorna a cor do botao para verde
                $(botao).removeClass('bg-danger').addClass('bg-success')

                // Libera novas ligações
                sessionStorage.setItem('modUsuarios_ligar', 'on')

                if (!resp) {
                    msgErro('Ligação falhou')
                }
            })
        }
    }
}

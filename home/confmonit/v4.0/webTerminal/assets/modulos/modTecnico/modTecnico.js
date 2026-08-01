function modTecnico_start(retorno) {
    $('#boxDir').empty()
    const uri = '/assets/modulos/modTecnico/modTecnico.html'
    $('#boxDir').load(uri, () => {
        $('#modTecnico_btnFechar').on('click', retorno)

        const idFranqueado = sessionStorage.getItem('ateProDado_proc_idFranqueado')
        modTecnico_carregarDados(idFranqueado)
    })
}

function modTecnico_carregarDados(idFranqueado) {
    $.ajax({
        url: '/modTecnico/buscarDados',
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
                modTecnico_montarLinha(i)
            });
        }

        $(`td[tipo=modTecnico_btnLigar]`).on('click', function () {
            if (essionStorage.getItem('login_voipAtivo') == 'S') {
                modTecnico_ligar($(this).attr('numero'))
            }
        })
    })
}

function modTecnico_montarLinha(item) {
    $('#modTecnico tbody').append(`
        <tr>
            <td 
                class="bg-tbody txt-tbody"
            >${item.nome}</td>
            
            <td 
                class="click text-success fw-bold" 
                tipo="modTecnico_btnLigar" 
                numero="${item.celular}"
            >
                <i class="bi bi-telephone-outbound-fill"></i> &nbsp;
                ${formatarCelular(item.celular)}
            </td>
            
            <td 
                class="click text-success fw-bold" 
                tipo="modTecnico_btnLigar" 
                numero="${item.telefone}">
                <i class="bi bi-telephone-outbound-fill"></i> &nbsp;
                ${formatarCelular(item.telefone)}
            </td>
        </tr>    
    `)

}

function modTecnico_ligar(numero) {
    // Verifica novas ligações travada
    if (sessionStorage.getItem('modTecnico_ligar') === 'off') {
        msgErro('Já existe um ligação em andamento')
    } else {
        if (numero.length >= 10) {
            $(this).removeClass('text-success').addClass('text-danger')

            // Trava novas ligações
            sessionStorage.setItem('modTecnico_ligar', 'off')
           
            const src = sessionStorage.getItem('login_userCelular')
            const dst = numero
            vono_ligar(
                src, 
                dst,
                sessionStorage.getItem('ateProDado_proc_idCliente'), 
                (resp) => {
                // Retorna a cor do botao para verde
                $(this).removeClass('text-danger').addClass('text-success')

                // Libera novas ligações
                sessionStorage.setItem('modTecnico_ligar', 'on')

                if (!resp) {
                    msgErro('Ligação falhou')
                }
            })
        }
    }
}


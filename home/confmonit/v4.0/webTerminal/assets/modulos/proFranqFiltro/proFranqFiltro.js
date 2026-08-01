function proFranqFiltro_start() {

    if (sessionStorage.getItem('login_userVinculo') == "CENTRAL") {
        $('#boxDir').empty()
        $('#boxDir').load('/assets/modulos/proFranqFiltro/proFranqFiltro.html', () => {

            $('#proFranqFiltro_btnLimparFiltro').on(
                'click',
                proFranqFiltro_btnLimparFiltro
            )
            proFranqFiltro_carregarTabela()

        })
    } else {
        $('#boxDir').append(`
            <div class="text-center mt-5"> 
                <i 
                    class="bi bi-emoji-laughing" 
                    style="color: green; font-size: 250px;"></i>
            </div>            
            `)
    }
}

function proFranqFiltro_carregarTabela() {
    if (document.querySelector('#proFranqFiltro')) {
        $.ajax({
            url: 'proFranqFiltro/carregarTabela',
            method: 'Post',
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
                "Authorization": "Bearer " + sessionStorage.getItem('token')
            }

        }).fail(function (e) {
            alert("erro ao carregar filtro")
            console.log(e)
        }).done(function (r) {

            $('#proFranqFiltro tbody').empty()
            $('#proFranqFiltro thead').empty()

            if (r.status != 'Vazio') {

                $('#proFranqFiltro thead').append(`
                    <tr >
                        <th style="width: 10%;" 
                            class="fmt-thead">
                            Filtrar
                        </th>

                        <th style="width: 10%;" 
                            class="fmt-thead">
                            Qtd.
                        </th>

                        <th style="width: 80%;" 
                            class="fmt-thead">
                            Franqueado
                        </th>
                    </tr>                
                `)

                r.dados.forEach(item => {
                    proFranqFiltro_montaLinha(item)
                });

                $('td[tipo=proFranqFiltro_btnFiltrar]').on('click', proFranqFiltro_btnFiltrar)

            } else {

                $('#proFranqFiltro tbody').append(
                    `<tr >
                        <td 
                            colspan=7 
                            class="sem-atendimento fmt-table"> 
                        <i class="bi bi-emoji-laughing"></i>  <br />
                        <h1> Sem atendimentos pendentes </h1>
                    </td></tr>`
                )

            }
        })
    }
}

function proFranqFiltro_montaLinha(item) {
    $('#proFranqFiltro tbody').append(`
        <tr class="bg-table">
            <td 
                tipo="proFranqFiltro_btnFiltrar"
                idFranqueado="${item.idFranqueado}"                  
                class="click bg-success"
            >                        
                <div 
                    class="d-flex justify-content-center text-bg-success"
                >
                    <i class="bi bi-headset"></i>
                </div>                        
            </td>

            <td 
                class="fmt-tbody"
            >${item.quantidade}</td>
               
            <td 
                class="fmt-tbody"
            >${item.nomeFranqueado}</td>
        </tr>    
    `)
}

function proFranqFiltro_btnFiltrar() {
    // Pega o id do franqueado dos atributos do botao
    const idFranqueado = $(this).attr('idFranqueado')
    // Filtra pelo id
    proAtendimento_filtraFranqueado(idFranqueado)

    // Exibe o botao Limpar
    $('#proFranqFiltro_btnLimparFiltro').attr("hidden", false)

    // Liga temporizador autolimpar com o tempo da sessao gTempoFraqAutoLimpeza
    flags.proFranqFiltro_autoLimpar = sessionStorage.getItem('gTempoFraqAutoLimpeza')
}

function proFranqFiltro_autoLimpar() {
    // Verifica se o timer esta desligado -1
    if (flags.proFranqFiltro_autoLimpar != -1) {

        // Verifica se contador chegou a zero e se sim ele desliga o filtro
        if (flags.proFranqFiltro_autoLimpar == 0) {

            // Desliga o timer
            flags.proFranqFiltro_autoLimpar = -1

            // Limpa o filtro
            proFranqFiltro_btnLimparFiltro()
        } else {
            flags.proFranqFiltro_autoLimpar--
        }
    }
}

function proFranqFiltro_btnLimparFiltro() {
    // Desbilita o filtro
    proAtendimento_filtraFranqueado('TODOS')

    // Esconde o botão
    $('#proFranqFiltro_btnLimparFiltro').attr("hidden", true)

    // Destiva o timer
    flags.proFranqFiltro_autoLimpar = -1
}
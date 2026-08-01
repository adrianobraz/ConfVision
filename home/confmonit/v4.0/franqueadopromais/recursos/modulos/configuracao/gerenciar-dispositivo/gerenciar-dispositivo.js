var listaClientes = []

$(document).ready(function () {
    ajustaTabela()
    $(window).on('resize', function () {
        ajustaTabela()
    })

    $('#btn-limpar').on('click', GerenciarDispositivoLimpar)
    
    $('#btn-gravar').on('click', GerenciarDispositivoGravarDados)
    
    $('#btn-gerar-conta-alarme').on('click', GerenciarDispositivoGerarContaAlarme)

    $("#conta-alarme").on('blur', function () {
        this.value = ("0000" + this.value).slice(-4)
    })

    // Associa o eveto de selecionar do idCliente
    $('#id-cliente').on('change', function () {
        localStorage.setItem('idCliente', $('#id-cliente').val())

        if ($('#id-cliente').val() == '0') {
            GerenciarDispositivoBloquearCadastro()
        } else {
            // Carrega a tabela
            GerenciarDispositivoCarregarTabela($('#id-cliente').val(), () => {
                GerenciarDispositivoDesbloquearCadastro()
            })
        }
    })

    // Pesquisa de cliente no combo (nome, CPF, CNPJ, nick)
    $('#busca-cliente').on('input', GerenciarDispositivoFiltrarClientes)
    $('#busca-cliente').on('focus', function () {
        if ($(this).val().trim() != '') GerenciarDispositivoFiltrarClientes()
    })
    $('#lista-clientes').on('click', '.fp-combo-item', function () {
        $('#busca-cliente').val($(this).attr('data-nome'))
        $('#id-cliente').val($(this).attr('data-id')).trigger('change')
        $('#lista-clientes').addClass('d-none')
    })
    $(document).on('click', function (e) {
        if (!$(e.target).closest('#busca-cliente, #lista-clientes').length) {
            $('#lista-clientes').addClass('d-none')
        }
    })


    // Associa o evento de selecionar do campo dispositivo-id-fabricante
    $('#dispositivo-id-fabricante').on('change', function () {
        GerenciarDispositivoConfiguraFrabricante()
        GerenciarDispositivoCarregarModelosCentrais()
    })

    // Associa o evento de selicionar do campo dispositivo-modelo-central
    $('#dispositivo-modelo-central').on('change', function () {
        GerenciarDispositivoConfiguraModelo()
        GerenciarDispositivoCarregarParicao()
    })

    $('#DispositivoTipo').on('change', GerenciarDispositivoConfiguraTipo)

    // Zero o variavel idDispositivo
    localStorage.setItem('idDispositivo', '')

    // Esconde o grupo keep 
    $('#grupo-keep').hide()
    $('#grupo-imei').hide()


    // Carrega o campo Cliente
    GerenciarDispositivoCarregarClientes()
    GerenciarDispositivoCarregarFabricantes()
    GerenciarDispositivoBloquearCadastro() 



})

//#################### RELACIONADO AO CAMPO CLIENTE ####################//
//ok
function GerenciarDispositivoCarregarClientes() {

    $.ajax({
        url: 'GerenciarDispositivoCarregarCliente',
        method: 'Post',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado')
        })
    }).fail(function (e) {
        boxErro("Erro ao carregar os clientes")
    }).done(function (r) {
        listaClientes = (r.dados || []).filter(i => i.ativo == "S")
        $('#id-cliente').val('0')
    })
}

function clienteTextoBusca(i) {
    return [(i.nome || ''), (i.nick || ''), (i.documento1 || ''), (i.documento2 || '')]
        .join(' ').toLowerCase()
}

function GerenciarDispositivoFiltrarClientes() {
    var raw = ($('#busca-cliente').val() || '').toLowerCase().trim()
    var termoNum = raw.replace(/[^a-z0-9]/g, '')
    var lista = $('#lista-clientes')
    lista.empty()

    if (raw == '') {
        $('#id-cliente').val('0').trigger('change')
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

function GerenciarDispositivoBloquearCadastro() {
    // Limpa a tabela
    $('#tab-dispositivo tbody').empty()

    // Bloqueia os campos
    $('#dispositivo-id-fabricante').attr('disabled', true)
    $('#dispositivo-modelo-central').attr('disabled', true)
    $('#dispositivo-nome').attr('disabled', true)
    $('#dispositivo-particao-alarme').attr('disabled', true)
    $('#dispositivo-conta-alarme').attr('disabled', true)
    $('#dispositivo-mac-alarme').attr('disabled', true)
    $('#DispositivoImeiAlarme').attr('disabled', true)
    $('#dispositivo-keep-alive-alarme').attr('disabled', true)
    $('#DispositivoTipo').attr('disabled', true)
    $('#msgAtendente').attr('disabled', true)
    $('#senhaVerbal').attr('disabled', true)
    $('#contraSenhaVerbal').attr('disabled', true)
    $('#senha').attr('disabled', true)
}

function GerenciarDispositivoDesbloquearCadastro() {

    // Desbloqueia os campos
    $('#dispositivo-id-fabricante').attr('disabled', false)
    $('#dispositivo-modelo-central').attr('disabled', false)
    $('#dispositivo-nome').attr('disabled', false)
    $('#dispositivo-particao-alarme').attr('disabled', false)
    $('#dispositivo-conta-alarme').attr('disabled', false)
    $('#dispositivo-mac-alarme').attr('disabled', false)
    $('#DispositivoImeiAlarme').attr('disabled', false)
    $('#dispositivo-keep-alive-alarme').attr('disabled', false)
    $('#DispositivoTipo').attr('disabled', false)
    $('#msgAtendente').attr('disabled', false)
    $('#senhaVerbal').attr('disabled', false)
    $('#contraSenhaVerbal').attr('disabled', false)
    $('#senha').attr('disabled', false)
}

function GerenciarDispositivoCarregarTabela(idCliente, next) {
    if (idCliente == '0') {
        $('#tab-dispositivo tbody').empty()

    } else {
        $.ajax({
            url: 'GerenciarDispositivoListarPorCliente',
            method: 'Post',
            data: JSON.stringify({ idCliente: idCliente })
        }).fail(function (e) {
            console.log(e)
            boxErro("Erro ao carregar a tabela")
        }).done(function (r) {
            $('#tab-dispositivo tbody').empty()
            $('#DispositivoTipo').empty().append(`<option value="MASTER">MASTER</option>`)

            if (r.status != 'Vazio') {
                let tmpDispositivo = ""
                r.dados.forEach(i => {

                    // Carrega o campo tipo somente com os dispositivos master
                    if (i.tipo == "MASTER") {
                        $('#DispositivoTipo').append(
                            `<option value="${i.idDispositivo}">${i.nome}</option>`
                        )
                    }

                    tmpDispositivo = tmpDispositivo + i.tipo + ':'

                    // Configura imagem botao habilitar
                    let btnHabilitarImg
                    let btnHabilitarCor
                    if (i.ativo == "S") {
                        btnHabilitarImg = '<i class="bi bi-toggle-on"></i>'
                        btnHabilitarCor = 'btn-success'
                    } else {
                        btnHabilitarImg = '<i class="bi bi-toggle-off"></i>'
                        btnHabilitarCor = 'btn-secondary'
                    }


                    // Abrevia nome do cliente
                    const nomeCliente = (i.nomeCliente.length > 50) ?
                        i.nomeCliente.substring(0, 47) + "..." :
                        i.nomeCliente

                    // Seleciona cor da linha
                    let corTexto = ''
                    if (i.tipo == "MASTER") {
                        corTexto = 'text-success'
                    } else if (i.mac == "ACTIVE CENTER") {
                        corTexto = 'text-danger'
                    } else {
                        corTexto = 'text-primary'
                    }


                    // Insere a linha na tabela
                    $('#tab-dispositivo tbody').append(`
                        <tr class='${corTexto}'>
                        <td class="text-uppercase">${i.conta}</td>
                        <td class="text-uppercase">${i.particao}</td>
                        <td class="text-uppercase" title="${i.idFisico1}">${i.idFisico1}</td>
                        <td class="text-uppercase" title="${i.nomeCliente}">${i.nomeCliente}</td>
                        <td class="text-uppercase">${i.nome}</td>
                        <td>
                            <button class="btn btn-sm btn-primary py-0"
                                id=${i.idDispositivo}
                                tipo="editar"
                                title="Editar dados da Dispositivo">
                                <i class="bi bi-pencil-square"></i>
                            </button>
                        </td>
                    
                        <td>
                            <button class="btn btn-sm btn-danger py-0"
                                id=${i.idDispositivo}
                                tipo="excluir"
                                title="Exclui o Dispositivo">
                                <i class="bi bi-eraser-fill"></i>
                            </button>
                            </td>
                            <td>
                                <button class="btn btn-sm ${btnHabilitarCor} py-0"
                                    id=${i.idDispositivo}
                                    tipo="habilitar"
                                    status="${i.ativo}"
                                    title="Habilita/Desabilita o
                                    Dispositivo">
                                    ${btnHabilitarImg}
                                </button>
                            </td>
                        </tr>
                    `)
                });


                // Grava tabela dispositivo na variavel
                localStorage.setItem("tabDispositivo", tmpDispositivo.substring(0, (tmpDispositivo.length - 1)))

                // Associa os botoes
                $('button[tipo=editar]').on('click', function () {
                    const id = this.getAttribute("id")
                    GerenciarDispositivoBuscar(id)
                    $('.grupo-cliente').hide()
                })

                $('button[tipo=excluir]').on('click', function () {
                    const id = this.getAttribute("id")
                    GerenciarDispositivoExcluir(id)

                })

                $('button[tipo=habilitar]').on('click', function () {
                    const id = this.getAttribute("id")
                    const status = this.getAttribute("status")
                    GerenciarDispositivoHabilitar(id, status)
                })
            }

            if (typeof next === 'function') { next() }
        })
    }
}

//ok
function GerenciarDispositivoCarregarFabricantes() {

    $.ajax({
        url: 'GerenciarDispositivosFabricanteListar',
        method: 'Post',
    }).fail(function (e) {
        boxErro("Erro ao carregar os fabricantes")
    }).done(function (r) {
        console.log(r.dados)
        $('#dispositivo-id-fabricante').empty()
        $('#dispositivo-id-fabricante').append('<option value="0">SELECIONE</option>')

        r.dados.forEach(i => {
            if (i.ativo == "S") {
                $('#dispositivo-id-fabricante').append(`                 
                    <option value="${i.idFabricante}" keepVisivel="${i.keepVisivel}">
                        ${i.nome}
                    </option>
                `)
            }
        });

    })
}

function GerenciarDispositivoCarregarModelosCentrais(callback) {
    const fabricante = $('#dispositivo-id-fabricante').val()

    // Caso o valor seja igual a 0 ele reseta o campo colocando somente a opcao de selecionar
    if (fabricante == '0') {
        $('#dispositivo-modelo-central').empty().append('<option value="0">Selecione</option>')
    } else {// Caso algum fabricante seja selecionado ele ira buscar os modelos do fabricante

        $.ajax({
            url: '/GerenciarDispositivosModelosCentraisListarFabricante',
            method: 'Post',
            data: JSON.stringify({ idFabricante: fabricante })
        }).fail(function (e) {
            boxErro("Erro ao carregar os modelos de centrais")
            $('#dispositivo-modelo-central').empty().append('<option value="0">Selecione</option>')
        }).done(function (r) {
            console.log("modelo", r)
            // Adiciona os modelos retornados
            $('#dispositivo-modelo-central').empty().append('<option value="0">Selecione</option>')
            if (r.status != 'Vazio') {
                r.dados.forEach(i => {
                    $('#dispositivo-modelo-central').append(`
                    <option value="${i.idModelo}" qtdparticoes="${i.qtdParticoes}">${i.nome}</option>
                `)
                });
            }
            if (typeof callback != 'undefined') {
                callback()
            }
        })

    }
}


function GerenciarDispositivoCarregarParicao(next) {

    const qtdparticoes = $('#dispositivo-modelo-central :selected').attr('qtdparticoes')

    $('#dispositivo-particao-alarme').empty()

    $('#dispositivo-particao-alarme').append(`<option value="0">SELECIONE</option>`)

    for (i = 1; i <= qtdparticoes; i++) {
        $('#dispositivo-particao-alarme').append(`<option value="${i}">Particao ${i}</option>`)
    }

    if (typeof next != 'undefined') {
        next()
    }

}

function GerenciarDispositivoHabilitar(id) {

    $.ajax({
        start: boxProcessando(),
        url: '/GerenciarDispositivoHabilitar',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: id })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao habilitar/desabilitar Dispositivo")
    }).done(function (r) {
        boxFechar()
        GerenciarDispositivoCarregarTabela($('#id-cliente').val())
    })
}

function GerenciarDispositivoBuscar(id) {
    $.ajax({
        start: boxProcessando(),
        url: 'GerenciarDispositivoBuscar',
        method: 'Post',
        data: JSON.stringify({ idDispositivo: id })
    }).fail(function (e) {
        boxErro("Erro ao buscar dados do Dispositivo")
    }).done(function (r) {
        boxFechar()
        GerenciarDispositivoDesbloquearCadastro()
        const d = r.dados
        // Carrega os dados temporario da conta do cliente
        localStorage.setItem('idDispositivo', d.idDispositivo)
        localStorage.setItem('contaAlarme', addZeroEsquerda(d.conta), 4)
        localStorage.setItem('idCliente', d.idCliente)

        //////////////////////////////////////////////////////////////////////////////
        $('#dispositivo-id-fabricante').val(d.idFabricante)

        // Exibe o grupo imei caso seja fabricante JFL
        if ($('#dispositivo-id-fabricante').val() == '1') {
            $('#grupo-imei').show()
        } else {
            $('#grupo-imei').hide()
        }


        if ($('#dispositivo-id-fabricante :selected').attr('keepVisivel') == '1') {
            $('#grupo-keep').show()
        } else {
            $('#grupo-keep').hide()
        }

        GerenciarDispositivoCarregarModelosCentrais(function () {
            $('#dispositivo-modelo-central').val(d.idModelo)
            GerenciarDispositivoCarregarParicao(function () {
                $('#dispositivo-particao-alarme').val(d.particao)
            })
        })

        ////////////////////////////////////////////////////////////////////////////// 

        const keepCorrigido = (d.keepAlive > 20) ? 20 : d.keepAlive


        $('#dispositivo-nome').val(d.nome)
        $('#dispositivo-conta-alarme').val(addZeroEsquerda(d.conta), 4)
        $('#dispositivo-keep-alive-alarme').val(d.keepAlive)
        $('#senha').val(d.senha),
            $('#senhaVerbal').val(d.senhaVerbal),
            $('#contraSenhaVerbal').val(d.contraSenhaVerbal),
            $('#msgAtendente').val(d.msgAtendente),


            $('#dispositivo-mac-alarme').val(d.idFisico1)
        $('#DispositivoImeiAlarme').val(d.idFisico2)

        $('#dispositivo-mac-alarme').attr('disabled', true)
        $('#DispositivoImeiAlarme').attr('disabled', true)
        $('#DispositivoTipo').attr('disabled', true)

        if (d.tipo == "ACTIVE CENTER") {
            $("#DispositivoTipo").html('<option value="ACTIVE CENTER">ACTIVE CENTER</option>');
            $('#DispositivoTipo').val(d.tipo)

        } else {
            GerenciarDispositivoCarregarTabela($('#id-cliente').val(), () => {
                $('#DispositivoTipo').val(d.tipo)
            })
        }

        GerenciarDispositivoConfiguraTipo()

        GerenciarDispositivoConfiguraFrabricante()


    })
}

function GerenciarDispositivoConfiguraFrabricante() {

    if ($('#dispositivo-id-fabricante').val() == '1') {
        $('#grupo-imei').show()
        $('#grupo-keep').show()
        //$('#DispositivoTipo').append(`<option value="ACTIVE">ACTIVE</option>`)

    } else {
        $('#grupo-imei').hide()
        $('#grupo-keep').hide()
        //$("#DispositivoTipo option[value='ACTIVE']").remove();
    }
}

function GerenciarDispositivoConfiguraModelo() {
    if ($('#dispositivo-modelo-central').val() == '7') {
        $("#DispositivoTipo").html('<option value="ACTIVE CENTER">ACTIVE CENTER</option>');
        $('#DispositivoTipo').val('ACTIVE CENTER')
        $('#DispositivoTipo').attr('disabled', true)
        $('#dispositivo-mac-alarme').attr('disabled', true)
        $('#DispositivoImeiAlarme').attr('disabled', true)
        $('#dispositivo-keep-alive-alarme').attr('disabled', true)
        $('#msgAtendente').attr('disabled', true)
        $('#senhaVerbal').attr('disabled', true)
        $('#contraSenhaVerbal').attr('disabled', true)
        $('#senha').attr('disabled', true)

    } else {
        GerenciarDispositivoCarregarTabela($('#id-cliente').val())
        $('#DispositivoTipo').val('MASTER')
        $('#DispositivoTipo').attr('disabled', false)
        $('#dispositivo-mac-alarme').attr('disabled', false)
        $('#DispositivoImeiAlarme').attr('disabled', false)
        $('#dispositivo-keep-alive-alarme').attr('disabled', false)
        $('#msgAtendente').attr('disabled', false)
        $('#senhaVerbal').attr('disabled', false)
        $('#contraSenhaVerbal').attr('disabled', false)
        $('#senha').attr('disabled', false)

    }
}

function GerenciarDispositivoConfiguraTipo() {
    if ($('#DispositivoTipo').val() == "MASTER") {

        $('#dispositivo-mac-alarme').attr('disabled', false)
        $('#DispositivoImeiAlarme').attr('disabled', false)
        $('#dispositivo-keep-alive-alarme').attr('disabled', false)

    } else if ($('#DispositivoTipo').val() == "ACTIVE CENTER") {
        $('#dispositivo-mac-alarme').val('ACTIVE CENTER')
        $('#dispositivo-mac-alarme').attr('disabled', true)
        $('#DispositivoImeiAlarme').attr('disabled', true)
        $('#dispositivo-keep-alive-alarme').attr('disabled', true)

    } else {
        let tmp = ''
        if ($('#DispositivoTipo :selected').html().length > 29) {
            tmp = $('#DispositivoTipo :selected').html().substring(0, 29)
        } else {
            tmp = $('#DispositivoTipo :selected').html()
        }

        $('#dispositivo-mac-alarme').val(tmp)
        $('#dispositivo-mac-alarme').attr('disabled', true)
        $('#DispositivoImeiAlarme').attr('disabled', true)
        $('#dispositivo-keep-alive-alarme').attr('disabled', true)
    }
}

function GerenciarDispositivoLimparFormulario() {
    const tmp = $('#id-cliente').val()
    $('#formulario').each(function () {
        this.reset();
    })
    $('#id-cliente').val(tmp)
    localStorage.setItem("idDispositivo", "")
    localStorage.setItem('contaAlarme', "")

    $('.grupo-cliente').show()
    //GerenciarDispositivoBloquearCadastro()
    // Esconde ou grupo keep 
    $('#grupo-keep').hide()
    $('#grupo-imei').hide()
    $('#grupo-mac').show()


    $('#DispositivoTipo').empty().append(`<option value="MASTER">MASTER</option>`)
}

function GerenciarDispositivoLimpar() {

    $('#formulario').each(function () {
        this.reset();
    })

    $('#busca-cliente').val('')
    $('#id-cliente').val('0')
    $('#lista-clientes').empty().addClass('d-none')

    localStorage.setItem("idDispositivo", "")
    localStorage.setItem('contaAlarme', "")
    localStorage.setItem('idCliente', "")

    $('.grupo-cliente').show()
    GerenciarDispositivoBloquearCadastro()
    // Esconde ou grupo keep 
    $('#grupo-keep').hide()
    $('#grupo-imei').hide()
    $('#grupo-mac').show()


    $('#DispositivoTipo').empty().append(`<option value="MASTER">MASTER</option>`)
}

function GerenciarDispositivoGravarDados() {
    const id = localStorage.getItem('idDispositivo')
    if (id == "") {
        GerenciarDispositivoInserir()
    } else {
        GerenciarDispositivoAlterar(id)
    }
}

function GerenciarDispositivoGerarContaAlarme() {

    if ($('#id-cliente').val() != '0') {

        $.ajax({
            start: boxProcessando(),
            url: '/GerenciarDispositivoGerarContaAlarme',
            method: 'Post',
            data: JSON.stringify({
                idFranqueado: localStorage.getItem('idFranqueado')
            })
        }).fail(function (e) {
            boxErro("Erro ao gerar a conta")
        }).done(function (r) {
            if (r.dados == 'Limite de conta exedido') {
                boxErro(r.dados)
            } else {
                boxFechar()
            }
            $('#dispositivo-conta-alarme').val(r.dados)
        })
    }

}

function GerenciarDispositivoInserir() {

    GerenciarDispositivoValidarCamposBranco(true, function () {

        const keep = $('#dispositivo-keep-alive-alarme').val()

        const keepCorrigido = (keep > 20) ? 20 : keep
        let mac, imei, keepAlive

        if ($('#DispositivoTipo').val() == "MASTER") {
            mac = limpaDocumento($('#dispositivo-mac-alarme').val())
            imei = $('#DispositivoImeiAlarme').val()
            keepAlive = keepCorrigido
        } else if ($('#DispositivoTipo').val() == 'ACTIVE CENTER') {
            mac = 'ACTIVE CENTER'
            imei = 'ACTIVE CENTER'
            keepAlive = '0'

        } else {
            mac = $('#dispositivo-mac-alarme').val()
            imei = "SLAVE"
            keepAlive = '0'

        }


        $.ajax({
            start: boxProcessando(),
            url: '/GerenciarDispositivoInserir',
            method: 'Post',
            data: JSON.stringify(
                {
                    idCliente: $('#id-cliente').val(),
                    IdFabricante: $('#dispositivo-id-fabricante').val(),
                    idModelo: $('#dispositivo-modelo-central').val(),
                    nome: $('#dispositivo-nome').val(),
                    particao: $('#dispositivo-particao-alarme').val(),
                    conta: addZeroEsquerda($('#dispositivo-conta-alarme').val(), 4),
                    idFisico1: mac,
                    idFisico2: imei,
                    keepAlive: keepAlive,
                    tipo: $('#DispositivoTipo').val(),
                    senhaVerbal: $('#senhaVerbal').val(),
                    contraSenhaVerbal: $('#contraSenhaVerbal').val(),
                    msgAtendente: $('#msgAtendente').val(),
                    senha: $('#senha').val(),

                }
            )
        }).fail(function (e) {
            boxErro("Erro ao inserir o Dispositivo")
        }).done(function (r) {
            boxInseridoSucesso(r.dados)
            GerenciarDispositivoCarregarTabela($('#id-cliente').val())
            GerenciarDispositivoLimparFormulario()

        })
    })
}

function GerenciarDispositivoAlterar(id) {

    $('#id-cliente').val(localStorage.getItem('idCliente'))
    // verifica se a conta continua igual se sim ele não vai verificar
    const validarConta = (
        localStorage.getItem('contaAlarme') == $('#dispositivo-conta-alarme').val()
    ) ? false : true

    GerenciarDispositivoValidarCamposBranco(validarConta, function () {

        const keep = $('#dispositivo-keep-alive-alarme').val()

        const keepCorrigido = (keep > 20) ? 20 : keep

        let mac, imei, keepAlive

        if ($('#dispositivo-id-fabricante').val() == "1") {
            if ($('#DispositivoTipo').val() == "MASTER") {
                mac = limpaDocumento($('#dispositivo-mac-alarme').val())
                imei = $('#DispositivoImeiAlarme').val()
                keepAlive = keepCorrigido
            } else if ($('#DispositivoTipo').val() == "ACTIVE CENTER") {
                mac = 'ACTIVE CENTER'
                imei = 'ACTIVE CENTER'
                keepAlive = '0'
            } else {
                mac = $('#dispositivo-mac-alarme').val()
                imei = "SLAVE"
                keepAlive = '0'
            }
        } else {
            if ($('#DispositivoTipo').val() == "MASTER") {
                mac = limpaDocumento($('#dispositivo-mac-alarme').val())
                imei = "N/A"
                keepAlive = '0'
            } else {
                mac = $('#dispositivo-mac-alarme').val()
                imei = "N/A"
                keepAlive = '0'
            }
        }

        $.ajax({
            start: boxProcessando(),
            url: '/GerenciarDispositivoAlterar',
            method: 'Post',
            data: JSON.stringify(
                {
                    idDispositivo: localStorage.getItem('idDispositivo'),
                    IdFabricante: $('#dispositivo-id-fabricante').val(),
                    idModelo: $('#dispositivo-modelo-central').val(),
                    nome: $('#dispositivo-nome').val().toUpperCase(),
                    particao: $('#dispositivo-particao-alarme').val(),
                    conta: addZeroEsquerda($('#dispositivo-conta-alarme').val(), 4),
                    idFisico1: mac,
                    idFisico2: imei,
                    keepAlive: keepAlive,
                    tipo: $('#DispositivoTipo').val(),
                    senhaVerbal: $('#senhaVerbal').val(),
                    contraSenhaVerbal: $('#contraSenhaVerbal').val(),
                    msgAtendente: $('#msgAtendente').val(),
                    senha: $('#senha').val(),
                }
            )
        }).fail(function (e) {
            boxErro("Erro ao alterar dados do Dispositivo")
        }).done(function (r) {
            boxAteradoSucesso()
            GerenciarDispositivoCarregarTabela(localStorage.getItem('idCliente'))
            GerenciarDispositivoLimparFormulario()

        })
    })
}

function GerenciarDispositivoExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar este Dispositivo?`,
        text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode apagar!'
    }).then((result) => {
        if (result.isConfirmed) {

            let chave = true

            const tabDispositivo = localStorage.getItem('tabDispositivo').split(":");

            tabDispositivo.forEach(i => {
                if (i == id) {
                    boxAdvertenciaCampoAuto("Dispositivo com slave associado")
                    chave = false
                }
            })

            if (chave) {
                $.ajax({
                    start: boxProcessando(),
                    url: `/GerenciarDispositivoExcluir`,
                    method: 'Post',
                    data: JSON.stringify({
                        idDispositivo: id
                    })
                }).fail(function (e) {
                    boxErro('Erro ao excluir o Dispositivo')
                }).done(function (r) {
                    boxDeletadoSucesso()
                    GerenciarDispositivoCarregarTabela(localStorage.getItem('idCliente'))
                    GerenciarDispositivoLimparFormulario()

                })
            }
        }
    })
}

function GerenciarDispositivoValidarCamposBranco(verificaConta, acao) {

    if ($('#id-cliente').val() == "vazio") {
        boxAdvertenciaCampoAuto("Todos os clientes estão com dispositivos cadastrados",
            '#busca-cliente')
        return
    }

    const campos = function () {
        if ($('#id-cliente').val() == "0") {
            boxAdvertenciaCampoAuto("Selecione um cliente primeiro",
                '#busca-cliente')
            return true
        }
        if ($('#imei-alarme').val() == "") {
            boxAdvertenciaCampoAuto("O campo Imei, não poder ficar em branco",
                '#imei-alarme')
            return true
        }

        if ($('#mac-alarme').val() == "") {
            boxAdvertenciaCampoAuto("O campo Mac, não poder ficar em branco",
                '#mac-alarme')
            return true
        }

        if ($('#keep-alive-alarme').val() == "") {
            boxAdvertenciaCampoAuto("O campo Keepalive, não poder ficar em branco",
                '#keep-alive-alarme')
            return true
        }

        if ($('#modelo-alarme').val() == "") {
            boxAdvertenciaCampoAuto("O campo Modelo, não poder ficar em branco",
                '#modelo-alarme')
            return true
        }

        if ($('#dispositivo-particao-alarme').val() == "0") {
            boxAdvertenciaCampoAuto("Selecione uma partição",
                '#dispositivo-particao-alarme')
            return true
        }

        if ($('#dispositivo-id-fabricante').val() == "0") {
            boxAdvertenciaCampoAuto("Selecione um fabricante",
                '#dispositivo-id-fabricante')
            return true
        }
    }

    if ($('#dispositivo-conta-alarme').val() == "") {
        boxAdvertenciaCampoAuto("O campo Conta Alarme deve ser gerado, não poder ficar em branco",
            '#conta-alarme')
        return
    } else {
        if (verificaConta) {
            $.ajax({
                url: '/GerenciarDispositivoVerificarContaAlarme',
                method: 'Post',
                data: JSON.stringify({
                    idFranqueado: localStorage.getItem('idFranqueado'),
                    conta: $('#dispositivo-conta-alarme').val()
                })
            }).fail(function (e) {
                boxErro("Erro ao gerar a conta")
            }).done(function (r) {

                if (r.dados != "LIVRE") {
                    boxAdvertenciaCampoAuto(
                        `Está conta já esta em uso por: ${r.dados}, escolha outra ou gere uma atomaticamente `,
                        '#conta-alarme'
                    )
                    return
                } else
                    if (!campos()) {
                        acao()
                    }
            })
        } else {
            if (!campos()) {
                acao()
            }
        }

    }
}



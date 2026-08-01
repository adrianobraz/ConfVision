var cvPagClientes = null
var cvBuscaClientesTimer = null

$(document).ready(function () {
    document.body.classList.add('cv-cadastro')
    $(window).trigger('resize');

    GerenciarClienteAlturaFixa()
    GerenciarClienteAssociaBotoesTabela()

    $('#btn-limpar').on('click', GerenciarClienteLimparFormulario)
    $('#btn-gravar').on('click', GerenciarClienteGravarDados)
    $('#tipo').on('change', () => {
        GerenciarClienteInicializarTipo($('#tipo').val())
    })

    $('#pesquisar-cliente').on('input', function () {
        clearTimeout(cvBuscaClientesTimer)
        cvBuscaClientesTimer = setTimeout(GerenciarClienteCarregarTabela, 400)
    })
    $('#filtro-status-cliente').on('change', GerenciarClienteCarregarTabela)

    GerenciarClienteInicializarTipo(11)
    GerenciarClienteCarregarPacotes()
    GerenciarClienteCarregarTabela()
    localStorage.setItem('idCamera', 'NOVO')
    localStorage.setItem('idCliente', '')
    localStorage.setItem('log', 'ON')


})

// Redimenciona a tela 
$(window).resize(GerenciarClienteAlturaFixa)

function GerenciarClienteAlturaFixa() {
    var h = Math.max(280, window.innerHeight - 240)
    $('#lista-clientes-scroll').height(h)
}

function GerenciarClienteLimparFormulario(evt) {
    $('#formulario').each(function () {
        this.reset();
    })

    localStorage.setItem('idCliente', '')
    $('#email1').prop('readonly', false)
}

function GerenciarClienteGravarDados() {

    // Verifica se a variavel idCliente contem um id,
    // caso contenha ele altera o caso não contenha 
    // insere
    if (localStorage.getItem('idCliente') == "") {
        GerenciarClienteEmailLivre($('#email1').val(), GerenciarClienteInserir)
    } else {
        GerenciarClienteAlterar()
    }

}

function GerenciarClienteInicializarTipo(tam) {
    $('#tipo').val(tam)
    if (tam == 11) {
        $('.fisica').show()
        $('.juridica').hide()
    } else {
        $('.fisica').hide()
        $('.juridica').show()
    }
}

function GerenciarClienteCarregarDispositivo(idCliente) {
    $.ajax({
        url: '/ClienteDispositivoListarPorCliente',
        method: 'Post',
        data: JSON.stringify({ idCliente: idCliente })
    }).fail(function (e) {
        console.log(e)
        boxErro("Erro ao carregar a tabela")
    }).done(function (r) {
        $('#tab-dispositivo-armar tbody').empty()
        if (r.status != 'Vazio') {

            r.dados.forEach(i => {
                // console.log(i)
                let btnImg
                let btnCor

                if (i.armado == 'S') {
                    btnCor = 'btn-success'
                    btnImg = '<i class="bi bi-lock-fill"></i>'
                } else {
                    btnCor = 'btn-danger'
                    btnImg = '<i class="bi bi-unlock-fill"></i>'
                }

                $('#tab-dispositivo-armar tbody').append(`
                    <tr class="text-center">
                        <td class="text-uppercase">${i.conta}</td>
                        <td class="text-uppercase">${i.particao}</td>
                        <td class="text-uppercase">${i.nome}</td>

                        <td>
                            <button class="btn btn-sm py-0 ${btnCor} "
                                id="${i.idDispositivo}"
                                tipo="armar"
                                particao="${i.particao}"
                                senhaAlarme="${i.senha}"
                                status="${i.armado}"
                                title="Arma ou desarma Dispositivo">
                                ${btnImg}
                            </button>
                        </td>
                    </tr>
                `)
            })
            $('button[tipo=armar]').on('click', function () {
                const idDispositivo = this.getAttribute("id")
                const particao = this.getAttribute("particao")
                const senhaAlarme = this.getAttribute("senhaAlarme")
                const status = this.getAttribute("status")

                GerenciarClienteArmar(idDispositivo, particao, senhaAlarme, status, this)

            })
        }
    })
}

function GerenciarClienteArmar(idDispositivo, particao, senhaAlarme, status, btn) {

    // Pega o id do dispositivo selecionado pra acerta os campos apos alteração    
    const url = '/GerenciarClienteArmar'
    acao = (status == 'S') ? 0 : 1
    $.ajax({
        start: boxProcessando("Processando arme"),
        url: url,
        method: 'Post',
        data: JSON.stringify({
            idDispositivo: idDispositivo,
            numero: Number(particao),
            acao: Number(acao),
            usuario: "000",
            senha: senhaAlarme,
            senhaWeb: 'WHdQkY&RX%W%4RArwm1Q'
        })
    }).fail(function (e) {
        if (e.responseJSON.status.includes('nao conectado')) {
            boxAdvertenciaAuto('Dispositivo desconectado do sistema')
        } else if (e.responseJSON.status.includes("não implementado")){
            boxErro('Comando não implementado para o alarme')
        } else {
            console.log(e)
            boxErro('Erro  ao tentar Armar/Desarmar o alarme')

        }
    }).done(function (r) {
        // console.log(r)
        boxFechar()
        const part = r.dados.particao.filter(d => (d.numero == particao))
        // console.log(part[0].status)
        if (part[0].status.includes("DESARMADA")) {
            $(btn).html('<i class="bi bi-unlock-fill"></i>')
            $(btn).addClass("btn-danger").removeClass("btn-success")
            $(btn).attr('status', 'N')
        } else {
            $(btn).html('<i class="bi bi-lock-fill"></i>')
            $(btn).removeClass("btn-danger").addClass("btn-success")
            $(btn).attr('status', 'S')
        }



        // if (part[p].status == 'ARMADA') {
        //     $(btn).removeClass('btn-danger')
        //     $(btn).addClass('btn-success')
        //     $(btn).html('<i class="bi bi-lock-fill"></i>')
        //     $(btn).attr("status", "1");
        // }

    })

}

function GerenciarClienteCarregarTabela() {
    if (typeof cvScrollPaginacao !== 'function') {
        console.error('cv-scroll-paginacao.js nao carregado')
        return
    }
    if (cvPagClientes) cvPagClientes.destroy()

    var termo = ($('#pesquisar-cliente').val() || '').trim()
    var filtroStatus = ($('#filtro-status-cliente').val() || '').trim()
    var $scroll = document.getElementById('lista-clientes-scroll')
    var $mais = $('#lista-clientes-mais')
    var $total = $('#lista-clientes-total')

    cvPagClientes = cvScrollPaginacao({
        url: '/ClienteListar',
        tbodySel: '#tab-clientes tbody',
        scrollTarget: $scroll,
        showLoading: true,
        getPayload: function () {
            return {
                idFranqueado: localStorage.getItem('idFranqueado'),
                termo: termo,
                filtroStatus: filtroStatus
            }
        },
        renderRows: function (dados) {
            dados.forEach(function (i) {
                GerenciarClienteMontaLinha(i)
            })
        },
        onEmpty: function () {
            $('#tab-clientes tbody').append(
                '<tr><td colspan="4" class="text-center text-muted py-3">Nenhum cliente encontrado</td></tr>'
            )
            $total.text('')
        },
        onTotal: function (total) {
            if (total == null) {
                $total.text('')
                return
            }
            var texto = total + (total === 1 ? ' cliente' : ' clientes')
            if (filtroStatus || termo) texto += ' (filtro aplicado)'
            $total.text(texto)
        },
        onLoadingMore: function () {
            $mais.removeClass('d-none')
        },
        onDone: function () {
            $mais.addClass('d-none')
            var total = cvPagClientes ? cvPagClientes.getTotal() : null
            var offset = cvPagClientes ? cvPagClientes.getOffset() : 0
            if (total != null && offset < total) {
                var base = $total.text().replace(/\s*—.*$/, '').trim()
                $total.text((base || (total + ' clientes')) + ' — role para carregar mais')
            }
        },
        onFail: function (e) {
            $mais.addClass('d-none')
            log(localStorage.getItem('log'), e)
            boxErro('Erro ao carregar clientes')
        }
    })

    cvPagClientes.reset()
}

function GerenciarClienteCarregarPacotes() {
    const id = localStorage.getItem('idFranqueado')
    $.ajax({
        start: boxProcessando(),
        url: `/clienteCarregarPacotes`,
        method: 'Post',
        data: JSON.stringify({ idVinculo: id })
    }).fail(function (e) {        
        log(localStorage.getItem('log'), e)
    }).done(function (r) {
        
        boxFechar()
        // log(localStorage.getItem('log'), r)
        $('#pacote').empty()
        if (r.status != 'Vazio') {
            r.dados.forEach(i => {  
                $("#pacote").append(
                    `<option value="${i.idPacote}">${i.nome}</option>`
                )
            });
        } else {
            $("#pacote").append(
                `<option value="">FAVOR CADASTRAR UM PACOTE</option>`
            )
        }
    })
}

function GerenciarClienteMontaLinha(i) {

    let btnAtivoImg
    let btnAtivoCor
    let btnAtivoTipo
    if (i.ativo == "S") {
        btnAtivoImg = '<i class="bi bi-toggle-on"></i>'
        btnAtivoCor = 'btn-success'
        btnAtivoTipo = 'habilitar'
    } else {
        btnAtivoImg = '<i class="bi bi-toggle-off"></i>'
        btnAtivoCor = 'btn-secondary'
        btnAtivoTipo = 'habilitar'
    }


    let botoes
    if (i.dataCancelamento == '') {
        botoes = `
        <!-- Editar Cliente -->
        <td class="px-0">
            <button class="btn btn-sm btn-primary"
                id=${i.idCliente}
                tipo="editar"
                title="Editar dados da Cliente">
                <i class="bi bi-pencil-square"></i>
            </button>
        </td>

        <!-- Excluir o cliente -->        
        <td class="px-0">
            <button class="btn btn-sm btn-danger"
                id=${i.idCliente}
                tipo="excluir"
                title="Exclui o Cliente">
                <i class="bi bi-eraser-fill"></i>
            </button>
        </td>

        <!-- Ativar ou desativar o cliente -->
        <td class="px-0">
            <button class="btn btn-sm ${btnAtivoCor}"
                id=${i.idCliente}
                tipo="${btnAtivoTipo}"
                status="${i.ativo}"
                title="Habilita/Desabilita o
                Cliente">
                ${btnAtivoImg}
            </button>
        </td>
    `
    } else {
        botoes = `<td 
                    id=${i.idCliente} 
                    tipo="restaurar"  
                    colspan="3" 
                    title="Click para restaurar o cliente"
                    class="px-0 text-center" 
                    style="color: red; cursor: pointer;"                    
                >CANCELADO(${i.dataCancelamento})</td>`
    }

    $('#tab-clientes tbody').append(`
        <tr>
            <td class="text-uppercase">${i.nome}</td>
            ${botoes}
        </tr>
    `)
}

function GerenciarClienteAssociaBotoesTabela() {
    var $tb = $('#tab-clientes tbody')
    $tb.off('click.cvCli')

    $tb.on('click.cvCli', 'button[tipo=editar]', function () {
        GerenciarClienteBuscar(this.getAttribute('id'))
    })

    $tb.on('click.cvCli', 'button[tipo=excluir]', function () {
        GerenciarClientePreExcluir(this.getAttribute('id'))
    })

    $tb.on('click.cvCli', 'td[tipo=restaurar]', function () {
        GerenciarClienteRestauraPreExcluir(this.getAttribute('id'))
    })

    $tb.on('click.cvCli', 'button[tipo=habilitar]', function () {
        GerenciarClienteHabilitar(this.getAttribute('id'))
    })
}

function GerenciarClienteModalAlterarEmail(id) {
    const myModal = new bootstrap.Modal(document.getElementById('modal-email'), {
        keyboard: false
    })
    $('#novoEmail').val("")
    myModal.show()
    $('#btnModalEnviar').on('click', function () {
        const email = $('#novoEmail').val()
        if (email != '') {
            GerenciarClienteEmailLivre(email, function () {
                const url = `ClienteAlterarEmailPrincipal`
                $.ajax({
                    start: boxProcessando(),
                    url: url,
                    method: 'Post',
                    data: JSON.stringify({
                        idCliente: id,
                        email1: email
                    })
                }).fail(function (e) {
                    console.log(e)
                    boxErro('Erro ao alterar o email')
                }).done(function (r) {
                    boxFechar()
                    boxAteradoSucesso()
                })
            })

            myModal.hide()
        } else {
            boxAdvertenciaCampoAuto('O Campo email não pode ficar em branco', '#novoEmail')
        }

        return
    })
}

function GerenciarClienteModalArmarAlarme(id) {
    GerenciarClienteCarregarDispositivo(id)
    const myModal = new bootstrap.Modal(document.getElementById('modal-armar-alarme'), {
        keyboard: false
    })

    myModal.show()
}

function GerenciarClienteResetarSenha(id) {
    $.ajax({
        start: boxProcessando(),
        url: `/ClienteResetarSenha`,
        method: 'POST',
        data: JSON.stringify({ idCliente: id })
    }).fail(function (e) {
        console.log(e)
        log(localStorage.getItem('log'), e)
        boxErro('Erro ao resetar a senha')
    }).done(function (r) {
        log(localStorage.getItem('log'), r)
        boxFechar()
        boxSenhaAlterada('usuario123')
    })
}

function GerenciarClienteHabilitar(id) {

    $.ajax({
        url: `ClienteHabilitar`,
        method: 'Post',
        data: JSON.stringify({ idCliente: id })
    }).fail(function (e) {
        log(localStorage.getItem('log'), e)
        boxErro('Erro ao Habilitar/Desabilitar o cliente')
    }).done(function (r) {
        log(localStorage.getItem('log'), r)
        GerenciarClienteCarregarTabela()
    })
}

function GerenciarClienteInserir() {
    if (GerenciarClienteValidarCamposBranco()) return

    const tipo = $('#tipo').val()

    let documento1
    let documento2
    if (tipo == 11) {
        documento1 = limpaDocumento($('#cpf').val())

        if (documento1.length != 11) {
            boxAdvertenciaAuto("Cpf com tamanho invalido")
            $('#cpf').focus()
            return
        }

        documento2 = $('#rg').val()

    } else {
        documento1 = limpaDocumento($('#cnpj').val())

        if (documento1.length != 14) {
            boxAdvertenciaAuto("Cnpj com tamanho invalido")
            $('#cnpj').focus()
            return
        }

        documento2 = $('#inscricao').val()
    }

    $.ajax({
        start: boxProcessando(),
        url: '/ClienteIncluir',
        method: 'Post',
        data: JSON.stringify(
            {
                idFranqueado: localStorage.getItem('idFranqueado'),
                idPacote: $('#pacote').val(),
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                documento1: limpaDocumento(documento1),
                documento2: documento2,
                cep: $('#cep').val(),
                endereco: $('#endereco').val().toUpperCase(),
                bairro: $('#bairro').val().toUpperCase(),
                cidade: $('#cidade').val().toUpperCase(),
                uf: $('#uf').val().toUpperCase(),
                telefone1: formatarCelular($('#telefone1').val()),
                telefone2: formatarCelular($('#telefone2').val()),
                email1: $('#email1').val().toLowerCase(),
                email2: $('#email2').val().toLowerCase(),
                
            }

        )
    }).fail(function (e) {
        console.log(e)
        boxErro('Erro ao Inserir o Cliente')
    }).done(function (r) {
        boxFechar()
        boxInseridoSucesso(r.dados)
        GerenciarClienteLimparFormulario()
        GerenciarClienteCarregarTabela()
    })
}

function GerenciarClienteBuscar(id) {
    $.ajax({
        start: boxProcessando(),
        url: `/ClienteBuscarDados`,
        method: 'Post',
        data: JSON.stringify({ idCliente: id })
    }).fail(function (e) {
        log(localStorage.getItem('log'), e)
        boxErro("erro ao carregar os dados do Cliente")
    }).done(function (r) {
        // console.log(r)
        log(localStorage.getItem('log'), r)
        boxFechar()

        if (r.status == 'OK') {

            const d = r.dados

            if (d.documento1.length == 11) {
                GerenciarClienteInicializarTipo(11)
            } else {
                GerenciarClienteInicializarTipo(14)
            }
            //GerenciarClienteInicializarTipo(14)

            localStorage.setItem('idCamera', d.codigoCamera)
            localStorage.setItem('idCliente', d.idCliente)

            $('#nome').val(d.nome)
            $('#nick').val(d.nick)

            
            if (d.documento1.length == 14) {
                $('#tipo').val(14)

                $('#cnpj').val(formatarDocumento(d.documento1))
                
                $('#inscricao').val(d.documento2)
            } else {
                $('#tipo').val(11)

                $('#cpf').val(formatarDocumento(d.documento1))

                $('#rg').val(d.documento2)

            }

            $('#pacote').val(d.idPacote)

            $('#uf').val(d.uf)

            $('#endereco').val(d.endereco)

            $('#telefone1').val(formatarCelular(d.telefone1))


            $('#bairro').val(d.bairro)

            $('#telefone2').val(formatarCelular(d.telefone2))

            $('#cidade').val(d.cidade)

            $('#email1').val(d.email1)
            $('#email1').prop('readonly', true)

            $('#email2').val(d.email2)
            $('#cep').val(d.cep)

        }
    })
}

function GerenciarClienteAlterar() {
    
    if (GerenciarClienteValidarCamposBranco()) return
    const tipo = $('#tipo').val()

    let documento1
    let documento2
    
    if (tipo == 11) {
        documento1 = limpaDocumento($('#cpf').val())

        if (documento1.length != 11) {
            boxAdvertenciaAuto("Cpf com tamanho invalido")
            $('#cpf').focus()
            return
        }

        documento2 = $('#rg').val()

    } else {
        documento1 = limpaDocumento($('#cnpj').val())

        if (documento1.length != 14) {
            boxAdvertenciaAuto("Cnpj com tamanho invalido")
            $('#cnpj').focus()
            return
        }

        documento2 = $('#inscricao').val()
    }
    
    $.ajax({
        start: boxProcessando(),
        url: 'ClienteAlterar',
        method: 'Post',
        data: JSON.stringify(
            {
                idCliente: localStorage.getItem('idCliente'),
                idPacote: $('#pacote').val(),
                nome: $('#nome').val(),
                nick: $('#nick').val(),
                documento1: documento1,
                documento2: documento2,
                uf: $('#uf').val(),
                endereco: $('#endereco').val(),
                telefone1: limpaDocumento($('#telefone1').val()),
                bairro: $('#bairro').val(),
                telefone2: limpaDocumento($('#telefone2').val()),
                cidade: $('#cidade').val(),
                email1: $('#email1').val(),
                email2: $('#email2').val(),
                cep: $('#cep').val()
            }
        )
    }).fail(function (e) {
        console.log(e)
        boxErro("erro ao alterar o cliente")
    }).done(function (r) {
        
        
        boxAteradoSucesso()
        GerenciarClienteLimparFormulario()
        GerenciarClienteCarregarTabela()
    })

}

function GerenciarClientePreExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer realmente cancelar este Cliente?`,
        text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode cancelar!'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                start: boxProcessando(),
                url: `/ClientePreExcluir`,
                method: 'Post',
                data: JSON.stringify({ idCliente: id }
                )
            }).fail(function (e) {
                console.log(e)
                boxErro("Erro ao deletar o cliente")
            }).done(function (r) {
                GerenciarClienteLimparFormulario()
                GerenciarClienteCarregarTabela()
                boxDeletadoSucesso()
            })
        }
    })
}

function GerenciarClienteRestauraPreExcluir(id) {
    Swal.fire({
        //position: 'top',
        title: `Quer realmente restaurar este Cliente?`,
        text: "Poderar cancelar ele novamente caso necessario",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode restaurar!'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                start: boxProcessando(),
                url: `/ClienteRestauraPreExcluir`,
                method: 'Post',
                data: JSON.stringify({ idCliente: id }
                )
            }).fail(function (e) {
                console.log(e)
                boxErro("Erro ao deletar o cliente")
            }).done(function (r) {
                GerenciarClienteLimparFormulario()
                GerenciarClienteCarregarTabela()
                boxDeletadoSucesso()
            })
        }
    })
}

function GerenciarClienteValidarCamposBranco() {
    if ($('#nome').val() == "") {
        boxAdvertenciaCampoAuto("O campo Nome, não poder ficar em branco", '#nome')
        return true
    }
    if ($('#nick').val() == "") {
        boxAdvertenciaCampoAuto("O campo Nick, não poder ficar em branco", '#nick')        
        return true
    }

    if ($('#telefone1').val() == "") {
        boxAdvertenciaCampoAuto("O campo Telefone 1, não poder ficar em branco", '#telefone1')
        return true
    }
    
    if ($('#email1').val() == "") {
        boxAdvertenciaCampoAuto("O campo Email, não poder ficar em branco", '#email1')
        return true
    }

    if ($('#pacote').val() == "") {
        boxAdvertenciaCampoAuto("O campo pacote, não poder ficar em branco", '#pacote')
        
        return true
    }

    return false
}

function GerenciarClienteAtivarEnvioEmail(id) {

    $.ajax({
        start: boxProcessando(),
        url: `/ClienteHabilitarEmail`,
        method: 'Post',
        data: JSON.stringify({ idCliente: id })
    }).fail(function (e) {
        log(localStorage.getItem('log'), e)
        boxErro('Erro ao Habilitar/Desabilitar o envio de email cliente')
    }).done(function (r) {
        log(localStorage.getItem('log'), r)
        boxFechar()
        GerenciarClienteCarregarTabela()
    })
}

function GerenciarClienteEmailLivre(email, executar) {
    if (email == "") {
        boxAdvertenciaAuto("O campo Email, não poder ficar em branco")
        $('#email1').focus()
        return
    }

    $.ajax({
        start: boxProcessando(),
        stop: boxFechar(),
        url: 'ClienteEmailLivre',
        method: 'Post',
        data: JSON.stringify({ email1: email })
    }).fail(function (e) {
        console(e)
        boxErro('Erro ao verificar email livre')
    }).done(function (r) {

        if (r.dados != 'LIVRE') {
            boxAdvertenciaCampoAuto(
                `Este email já esta sendo usado por: ${r.dados}`,
                '#email-principal-cliente'
            )
            return
        }
        executar()

    })
}

function GerenciarClienteVerificarCodigoCamera() {
    return false
}
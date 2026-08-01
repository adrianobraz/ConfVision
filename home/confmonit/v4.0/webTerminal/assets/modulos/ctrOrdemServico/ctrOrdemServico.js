// ctrOrdemServico_start inicializa o modulo
// opts (opcional): { origem: 'terminal' | 'atendimento', idFranqueado, nomeFranqueado, idCliente, nomeCliente }
function ctrOrdemServico_start(opts) {
    var origem = (opts && opts.origem) ? opts.origem : 'terminal'

    $('#boxDir').empty()
    const uri = '/assets/modulos/ctrOrdemServico/ctrOrdemServico.html'
    $('#boxDir').load(uri, () => {
        $('#ctrOrdemServico_btnFechar').on('click', ctrOrdemServico_btnFechar)
        $('#ctrOrdemServico_btnLimpar').on('click', ctrOrdemServico_btnLimpar)
        $('#ctrOrdemServico_btnGravar').on('click', ctrOrdemServico_btnGravar)

        if (origem === 'atendimento') {
            $('#ctrOrdemServico').attr('data-origem', 'atendimento')

            // Sempre puxar empresa e cliente do atendimento atual (sessionStorage)
            var idFranqueado = sessionStorage.getItem('ateProDado_proc_idFranqueado') || (opts && opts.idFranqueado) || ''
            var nomeFranqueado = sessionStorage.getItem('ateProDado_proc_franqNome') || (opts && opts.nomeFranqueado) || ''
            var idCliente = sessionStorage.getItem('ateProDado_proc_idCliente') || (opts && opts.idCliente) || ''
            var nomeCliente = sessionStorage.getItem('ateProDado_proc_cliNome') || (opts && opts.nomeCliente) || ''

            if (!idFranqueado || !idCliente) {
                if (typeof msgErro === 'function') msgErro('Empresa ou cliente não disponível. Abra a ordem de serviço a partir do ícone na tela de atendimento.')
                else alert('Empresa ou cliente não disponível. Abra a ordem de serviço a partir do ícone na tela de atendimento.')
            }

            // Substituir os combos por inputs fixos: localizar por estrutura (1º input-group = Empresa, 2º = Cliente)
            var $responsivo = $('#ctrOrdemServico_responsivo')
            var $franqSelect = $responsivo.find('.input-group').eq(0).find('select')
            var $cliSelect = $responsivo.find('.input-group').eq(1).find('select')

            if ($franqSelect.length) {
                var $hFranq = $('<input type="hidden" id="ctrOrdemServico_cpFranqueado">').val(idFranqueado)
                var $tFranq = $('<input type="text" id="ctrOrdemServico_cpFranqueadoFixo" class="form-control text-center fmt-campo ctrOrdemServico-campoFixo" readonly>').val(nomeFranqueado)
                $franqSelect.after($tFranq).replaceWith($hFranq)
            }
            if ($cliSelect.length) {
                var $hCli = $('<input type="hidden" id="ctrOrdemServico_cpCliente">').val(idCliente)
                var $tCli = $('<input type="text" id="ctrOrdemServico_cpClienteFixo" class="form-control text-center fmt-campo ctrOrdemServico-campoFixo" readonly>').val(nomeCliente)
                $cliSelect.after($tCli).replaceWith($hCli)
            }
        } else {
            $('#ctrOrdemServico').attr('data-origem', 'terminal')
            // Origem terminal: combos para escolher empresa e cliente
            var $selFranq = $('#ctrOrdemServico_responsivo').find('.input-group').eq(0).find('select')
            var $selCli = $('#ctrOrdemServico_responsivo').find('.input-group').eq(1).find('select')
            $selFranq.prop('disabled', false)
            $selCli.prop('disabled', false)
            $selFranq.off('change').on('change', ctrOrdemServico_buscarCliente)

            var idVinculo = sessionStorage.getItem('login_userVinculo')
            var nomeVinculo = sessionStorage.getItem('login_userVinculoNome') || ''

            if (idVinculo == 'CENTRAL') {
                ctrOrdemServico_buscarFranqueado()
            } else if (idVinculo) {
                $selFranq.empty().append($('<option>').val(idVinculo).text(nomeVinculo || idVinculo))
                ctrOrdemServico_buscarCliente()
            } else {
                ctrOrdemServico_buscarFranqueado()
            }
        }
    })
}

// ctrOrdemServico_btnFechar fecha a janela e volta para a tela anterior
function ctrOrdemServico_btnFechar() {
    var origem = $('#ctrOrdemServico').attr('data-origem')
    if (origem === 'atendimento') {
        ateEventosDetalhe_start()
    } else {
        proFranqFiltro_start()
    }
}

// ctrOrdemServico_btnLimpar limpa os campos editaveis
function ctrOrdemServico_btnLimpar() {
    $('#ctrOrdemServico_cpAssunto').val('')
    $('#ctrOrdemServico_cpDescricao').val('')
}

// ctrOrdemServico_btnGravar grava um novo ticket
function ctrOrdemServico_btnGravar() {
    const assunto = $('#ctrOrdemServico_cpAssunto').val()
    const descricao = $('#ctrOrdemServico_cpDescricao').val()

    if (assunto == "") {
        msgErro("O campo assunto não pode ficar em branco")
        return
    }

    if (descricao == "") {
        msgErro("o campo descrição não pode ficar em branco")
        return
    }

    $.ajax({
        url: '/ctrOrdemServico/gravar',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idMaster: $('#ctrOrdemServico_cpFranqueado').val(),
            idSlave: $('#ctrOrdemServico_cpCliente').val(),
            assunto: assunto,
            descricao: descricao
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        ctrOrdemServico_btnLimpar()
        msgSucesso('Aberto com sucesso')
    })
}

// ctrOrdemServico_buscarFranqueado carrega o campo de selecionar franqueado
function ctrOrdemServico_buscarFranqueado() {
    var $sel = $('#ctrOrdemServico_responsivo').find('.input-group').eq(0).find('select')
    if (!$sel.length) $sel = $('#ctrOrdemServico_cpFranqueado')

    $.ajax({
        url: '/ctrOrdemServico/buscarFranqueado',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        }
    }).fail(function (e) {
        console.log(e)
        $sel.empty().append($('<option value="">').text('Erro ao carregar empresas'))
        if (typeof msgErro === 'function') msgErro('Não foi possível carregar a lista de empresas.')
    }).done(function (r) {
        $sel.empty()
        if (r.status != 'Vazio' && r.dados && r.dados.length) {
            r.dados.forEach(function (i) {
                var razao = (i.razaoSocial && i.razaoSocial.trim() !== '') ? i.razaoSocial.trim() : ''
                var fantasia = (i.nomeFantasia && i.nomeFantasia.trim() !== '') ? i.nomeFantasia.trim() : ''
                var texto = razao || fantasia || i.idFranqueado
                if (fantasia && fantasia !== razao) texto += ' (' + fantasia + ')'
                $sel.append($('<option>').val(i.idFranqueado).text(texto))
            })
            ctrOrdemServico_buscarCliente()
        } else {
            $sel.append($('<option value="">').text('Nenhuma empresa cadastrada'))
            if (typeof msgErro === 'function') msgErro('Nenhuma empresa cadastrada.')
        }
    })
}

function ctrOrdemServico_buscarCliente() {
    var $selFranq = $('#ctrOrdemServico_responsivo').find('.input-group').eq(0).find('select')
    if (!$selFranq.length) $selFranq = $('#ctrOrdemServico_cpFranqueado')
    var $selCli = $('#ctrOrdemServico_responsivo').find('.input-group').eq(1).find('select')
    if (!$selCli.length) $selCli = $('#ctrOrdemServico_cpCliente')

    var idFranqueado = $selFranq.val()
    if (!idFranqueado) {
        $selCli.empty().append($('<option value="0">').text('SELECIONE UM CLIENTE'))
        return
    }

    $.ajax({
        url: '/ctrOrdemServico/buscarCliente',
        method: 'Post',
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + sessionStorage.getItem('token')
        },
        data: JSON.stringify({
            idFranqueado: idFranqueado
        })
    }).fail(function (e) {
        console.log(e)
        $selCli.empty().append($('<option value="0">').text('Erro ao carregar clientes'))
        if (typeof msgErro === 'function') msgErro('Não foi possível carregar os clientes da empresa.')
    }).done(function (r) {
        $selCli.empty()
        $selCli.append($('<option value="0">').text('SELECIONE UM CLIENTE'))
        if (r.status != 'Vazio' && r.dados && r.dados.length) {
            r.dados.forEach(function (i) {
                $selCli.append($('<option>').val(i.idCliente).text(i.nome || i.idCliente))
            })
        }
    })
}


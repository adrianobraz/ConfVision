$(window).on('load', function () {
    sessionStorage.setItem("id", "0")
    listar()
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
})

//ok
function listar() {
    const payload = payloadTenant({ idVinculo: "CENTRAL" })
    if (!payload) return

    $.ajax({
        url: `/financeiro/pacote/listar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao listar pacotes')
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()
        if (r.status == 'OK') {
            r.dados.map(item => {
                const stateAtivo = (item.ativo == "S") ? 'tech-state-on' : 'tech-state-off'
                const rowState = (item.ativo == "S") ? 'tech-row-on' : 'tech-row-off'
                const iconAtivar = (item.ativo == "N")
                    ? '<i class="bi bi-toggle-off"></i>'
                    : '<i class="bi bi-toggle-on"></i>'

                const btnDeletar = (sessionStorage.getItem("loginMaster") == "1")
                    ? `<td onclick="deletar('${item.idPacote}', '${item.nome}')" class="text-sm btn tech-action-del"><i class="bi bi-eraser-fill"></i></td>`
                    : ''

                $('tbody').append(`
                    <tr class="${rowState}">
                        <td class="text-sm" title="${item.ativo == 'S' ? 'Ativo' : 'Desativado'}">${item.nome}</td>
                        <td onclick="buscar('${item.idPacote}')" class="text-sm btn tech-action-edit"><i class="bi bi-pencil-square"></i></td>
                        ${btnDeletar}
                        <td onclick="habilitar('${item.idPacote}')" class="text-sm btn ${stateAtivo}" title="Pacote: ${item.ativo == 'S' ? 'ATIVO' : 'DESATIVADO'}">${iconAtivar}</td>
                    </tr>    
                `)
            })
        } else {
            if (r.status != 'Vazio'){
                boxErro(r.status)
            }
        }
    })
}

//ok
function gravar() {

    (sessionStorage.getItem("id") == '0') ? inserir() : alterar()
}

//ok
function inserir() {
    const payload = payloadTenant({
        idVinculo: "CENTRAL",
        nome: $('#pacoteNome').val(),
        peridiocidade: $('#periodo').val(),
        valor: $('#pacoteValor').val(),
        contasQtd: $('#pacoteQtdConta').val(),
        contasExedente: $('#pacoteAdicional').val(),
        comissao: $('#pacoteComissao').val(),
        terminal: $('#adicionaisTerminal').val(),
        grade: $('#adicionaisGrade').val(),
        gradeValor: $('#adicionaisGradeValor').val(),
        atendimentoQtd: $('#atendimentoQtd').val(),
        atendimentoExedente: $('#atendimentoAdicional').val(),
        smsQtd: $('#smsQtd').val(),
        smsExedente: $('#smsAdicional').val(),
        smsBoquear: $('#smsBloquear').val(),
        ligacoesQtd: $('#ligacaoQtd').val(),
        ligacoesExedente: $('#ligacaoAdicional').val(),
        ligacoesBloquear: $('#ligacaoBloquear').val(),
        emailQtd: $('#emailQtd').val(),
        emailExedente: $('#emailAdicional').val(),
        emailBloquear: $('#emailBloquear').val(),
    })
    if (!payload) return

    $.ajax({
        url: `/financeiro/pacote/inserir`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
        boxErro('Falha ao inserir pacote')
    }).done(function (r) {
        console.log(r)
        if (r.status == 'OK') {
            limpar()
            listar()
            boxInseridoSucesso(r.dados)
        } else {
            boxErro(r.status)
        }
    })
}

//ok
function alterar() {
    $.ajax({
        url: `/financeiro/pacote/alterar`,
        method: 'POST',
        data: JSON.stringify({
            idPacote: sessionStorage.getItem('id'),

            nome: $('#pacoteNome').val(),
            peridiocidade: $('#periodo').val(),

            valor: $('#pacoteValor').val(),
            contasQtd: $('#pacoteQtdConta').val(),
            contasExedente: $('#pacoteAdicional').val(),
            comissao: $('#pacoteComissao').val(),

            terminal: $('#adicionaisTerminal').val(),
            grade: $('#adicionaisGrade').val(),
            gradeValor: $('#adicionaisGradeValor').val(),

            atendimentoQtd: $('#atendimentoQtd').val(),
            atendimentoExedente: $('#atendimentoAdicional').val(),

            smsQtd: $('#smsQtd').val(),
            smsExedente: $('#smsAdicional').val(),
            smsBoquear: $('#smsBloquear').val(),

            ligacoesQtd: $('#ligacaoQtd').val(),
            ligacoesExedente: $('#ligacaoAdicional').val(),
            ligacoesBloquear: $('#ligacaoBloquear').val(),

            emailQtd: $('#emailQtd').val(),
            emailExedente: $('#emailAdicional').val(),
            emailBloquear: $('#emailBloquear').val(),
        })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == 'OK') {
            limpar()
            listar()
            boxAteradoSucesso('Pacote alterado com sucesso')
        } else {
            boxErro(r.status)
        }
    })
}

//ok
function habilitar(id) {

    $.ajax({
        url: `/financeiro/pacote/habilitar`,
        method: 'POST',
        data: JSON.stringify({ idPacote: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == 'OK') {
            listar()
        } else {
            r.status
        }
    })
}

//ok
function buscar(id) {
    $.ajax({
        url: `/financeiro/pacote/buscar`,
        method: 'POST',
        data: JSON.stringify({ idPacote: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == 'OK') {
            const d = r.dados

            sessionStorage.setItem("id", d.idPacote)

            $('#pacoteNome').val(d.nome)
            $('#periodo').val(d.peridiocidade)

            $('#pacoteValor').val(d.valor)
            $('#pacoteQtdConta').val(d.contasQtd)
            $('#pacoteAdicional').val(d.contasExedente)
            $('#pacoteComissao').val(d.comissao),

            $('#adicionaisTerminal').val(d.terminal)
            $('#adicionaisGrade').val(d.grade)
            $('#adicionaisGradeValor').val(d.gradeValor)

            $('#atendimentoQtd').val(d.atendimentoQtd)
            $('#atendimentoAdicional').val(d.atendimentoExedente)

            $('#smsQtd').val(d.smsQtd)
            $('#smsAdicional').val(d.smsExedente)
            $('#smsBloquear').val(d.smsBoquear)

            $('#ligacaoQtd').val(d.ligacoesQtd)
            $('#ligacaoAdicional').val(d.ligacoesExedente)
            $('#ligacaoBloquear').val(d.ligacoesBloquear)

            $('#emailQtd').val(d.emailQtd)
            $('#emailAdicional').val(d.emailExedente)
            $('#emailBloquear').val(d.emailBloquear)
        } else {
            boxErro(r.status)
        }
    })
}

//ok
function limpar() {
    sessionStorage.setItem("id", "0")

    $('#pacoteNome').val('')
    $('#periodo').val('')

    $('#pacoteValor').val('')
    $('#pacoteQtdConta').val('')
    $('#pacoteAdicional').val('')
    $('#pacoteComissao').val('')
    

    $('#adicionaisTerminal').val('N')
    $('#adicionaisGrade').val('N')
    $('#adicionaisGradeValor').val('')

    $('#atendimentoQtd').val('')
    $('#atendimentoAdicional').val('')

    $('#smsQtd').val('')
    $('#smsAdicional').val('')
    $('#smsBloquear').val('N')

    $('#ligacaoQtd').val('')
    $('#ligacaoAdicional').val('')
    $('#ligacaoBloquear').val('N')

    $('#emailQtd').val('')
    $('#emailAdicional').val('')
    $('#emailBloquear').val('N')
}



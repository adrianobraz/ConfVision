$(window).on('load', function () {
    sessionStorage.setItem("id", "0")
    listar()
   
    
    $('#btnLimpar').on('click', limpar)
    $('#btnGravar').on('click', gravar)
})

function listar() {
    const payload = payloadTenant({ idVinculo: sessionStorage.getItem("loginRepId") })

    $.ajax({
        url: `/pacote/listar`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        $('tbody').empty()
        if (r.status == 'OK') {
            r.dados.map(item => {
                let iconAtivar
                let corAtivo

                if (item.ativo == "N") {
                    iconAtivar = '<i class="bi bi-toggle-off text-white"></i>'
                    corAtivo = 'bg-zinc-700'
                } else {
                    iconAtivar = '<i class="bi bi-toggle-on text-white"></i>'
                    corAtivo = 'bg-green-700'
                }

                const btnDeletar = (sessionStorage.getItem("loginMaster") == "1") ? `<td onclick="deletar('${item.idPacote}', '${item.nome}')" class="text-sm bg-red-600 btn"><i class="bi bi-eraser-fill"></i></td>` : ''

                $('tbody').append(`
                    <tr>
                        <td class="text-sm">${item.nome}</td>

                        <td onclick="buscar('${item.idPacote}')" class="text-sm bg-blue-600 btn"> <i class="bi bi-pencil-square"></i></td>
                        
                        ${btnDeletar}
                        
                        <td onclick="habilitar('${item.idPacote}')" class="text-sm ${corAtivo} btn">${iconAtivar}</td>
                    </tr>    
                `)
            })
        } 
    })
}

function gravar() {

    (sessionStorage.getItem("id") == '0') ? inserir() : alterar()
}

function inserir() {
    const payload = payloadTenant({
            idVinculo: sessionStorage.getItem("loginRepId"),

            nome: $('#pacoteNome').val(),
            peridiocidade: $('#periodo').val(),

            valor: $('#pacoteValor').val(),
            contasQtd: $('#pacoteQtdConta').val(),
            contasExedente: $('#pacoteAdicional').val(),

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

    $.ajax({
        url: `/pacote/inserir`,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload)
    }).fail(function (e) {
        console.log(e)
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

function alterar() {
    $.ajax({
        url: `/pacote/alterar`,
        method: 'POST',
        data: JSON.stringify({
            idPacote: sessionStorage.getItem('id'),

            nome: $('#pacoteNome').val(),
            peridiocidade: $('#periodo').val(),

            valor: $('#pacoteValor').val(),
            contasQtd: $('#pacoteQtdConta').val(),
            contasExedente: $('#pacoteAdicional').val(),

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

function habilitar(id) {

    $.ajax({
        url: `/pacote/habilitar`,
        method: 'POST',
        data: JSON.stringify({ idPacote: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        if (r.status == 'OK') {
            listar()
            // boxSucesso('Operação realizada com sucesso')
        } else {
            r.status
        }
    })
}

function buscar(id) {
    $.ajax({
        url: `/pacote/buscar`,
        method: 'POST',
        data: JSON.stringify({ idPacote: id })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
        console.log(r)
        if (r.status == 'OK') {
            const d = r.dados
            console.log(d)
            sessionStorage.setItem("id", d.idPacote)

            $('#pacoteNome').val(d.nome)
            $('#periodo').val(d.peridiocidade)

            $('#pacoteValor').val(d.valor)
            $('#pacoteQtdConta').val(d.contasQtd)
            $('#pacoteAdicional').val(d.contasExedente)

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

function limpar() {
    sessionStorage.setItem("id", "0")

    $('#pacoteNome').val('')
    $('#periodo').val('15')

    $('#pacoteValor').val('')
    $('#pacoteQtdConta').val('')
    $('#pacoteAdicional').val('')

    $('#adicionaisTerminal').val('0')
    $('#adicionaisGrade').val('0')
    $('#adicionaisGradeValor').val('')

    $('#atendimentoQtd').val('')
    $('#atendimentoAdicional').val('')

    $('#smsQtd').val('')
    $('#smsAdicional').val('')
    $('#smsBloquear').val('0')

    $('#ligacaoQtd').val('')
    $('#ligacaoAdicional').val('')
    $('#ligacaoBloquear').val('0')

    $('#emailQtd').val('')
    $('#emailAdicional').val('')
    $('#emailBloquear').val('0')
}

// essa função so pode ser usada pelo master
function deletar(id, nome) {
    boxConfirmarExcluir(nome, () => {
        $.ajax({
            url: `/pacote/deletar`,
            method: 'POST',
            data: JSON.stringify({ idPacote: id })
        }).fail(function (e) {
            console.log(e)
        }).done(function (r) {
            if (r.status == 'OK') {
                listar()
                boxSucesso('Pacote excluido com sucesso')
            } else {
                boxErro(r.status)
            }
        })
    })
}


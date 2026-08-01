/* Utilitarios compartilhados: tabela de dispositivos sem comunicacao */

var SEM_COM_FAIXAS = ['1', '3', '6', '12', '24']

var SEM_COM_FAIXA_LABELS = {
    '1': '+1h (1h a 2h59)',
    '3': '+3h (3h a 5h59)',
    '6': '+6h (6h a 11h59)',
    '12': '+12h (12h a 23h59)',
    '24': '+24h (24h a 36h)',
    'todos': 'Todos acima de 24h'
}

function SemComunicacaoResolverFaixa(params) {
    var faixa = (params.get('faixa') || '').toLowerCase()
    if (faixa === 'todos' || SEM_COM_FAIXAS.indexOf(faixa) !== -1) {
        return faixa
    }
    var horas = parseInt(params.get('horas'))
    if (SEM_COM_FAIXAS.indexOf(String(horas)) !== -1) {
        return String(horas)
    }
    return 'todos'
}

function SemComunicacaoLabelFaixa(faixa) {
    return SEM_COM_FAIXA_LABELS[faixa] || faixa
}

function SemComunicacaoTimestampEvento(dataStr) {
    if (!dataStr || dataStr == '01/01/0001 00:00:00') {
        return 0
    }
    return new Date(horaBrToUs(dataStr)).getTime()
}

function SemComunicacaoFormatoTempoDetalhado(horas) {
    var resto = horas
    var anos = Math.floor(resto / (24 * 365))
    resto = resto % (24 * 365)
    var meses = Math.floor(resto / (24 * 30))
    resto = resto % (24 * 30)
    var dias = Math.floor(resto / 24)
    var hrs = resto % 24

    var partes = []
    if (anos > 0) partes.push(anos + 'a')
    if (meses > 0) partes.push(meses + 'm')
    if (dias > 0) partes.push(dias + 'd')
    if (hrs > 0) partes.push(hrs + 'hr')

    return partes.length > 0 ? partes.join('/') : '0hr'
}

function SemComunicacaoTagGravidade(horas, nuncaConectou) {
    horas = parseInt(horas) || 0

    var texto
    if (nuncaConectou) {
        texto = 'Nunca conectou'
    } else {
        texto = SemComunicacaoFormatoTempoDetalhado(horas)
    }

    var classe
    var icone
    if (nuncaConectou || horas >= 72) {
        classe = 'bg-danger'
        icone = 'bi-exclamation-octagon-fill'
    } else {
        classe = 'bg-warning text-dark'
        icone = 'bi-exclamation-triangle-fill'
    }

    return '<span class="badge fp-tempo-badge ' + classe + '"><i class="bi ' + icone + '"></i> ' + texto + '</span>'
}

function SemComunicacaoAppendLinhas(dados, tbodySel) {
    if (!dados || dados.length === 0) return

    var $tbody = $(tbodySel)
    dados.forEach(function (i) {
        var horas
        var ultimoEvt
        var nuncaConectou = false

        if (i.dataUltimoEvento == '01/01/0001 00:00:00') {
            ultimoEvt = 'Nuca Conectou'
            horas = 0
            nuncaConectou = true
        } else {
            var now = new Date()
            ultimoEvt = i.dataUltimoEvento
            var past = new Date(horaBrToUs(i.dataUltimoEvento))
            var diff = Math.abs(now.getTime() - past.getTime())
            horas = Math.ceil(diff / (1000 * 60 * 60))
        }

        var tagGravidade = SemComunicacaoTagGravidade(horas, nuncaConectou)
        var descricao = i.descricaoUltimoEvento || ''
        var descricaoEvento = (descricao.length > 40) ?
            descricao.substring(0, 37) + '...' : descricao

        $tbody.append(
            '<tr style="font-size: 11px;">' +
            '<td>' + i.conta + '</td>' +
            '<td>' + i.nome + '</td>' +
            '<td>' + i.nomeCliente + '</td>' +
            '<td>' + i.codigoUltimoEvento + '</td>' +
            '<td>' + descricaoEvento + '</td>' +
            '<td>' + ultimoEvt + '</td>' +
            '<td class="fp-col-tempo text-center">' + tagGravidade + '</td>' +
            '</tr>'
        )
    })
}

function SemComunicacaoRenderTabela(dados, tbodySel, qtdSel) {
    var $tbody = $(tbodySel)
    $tbody.empty()

    if (!dados || dados.length === 0) {
        if (qtdSel) $(qtdSel).text('0')
        $tbody.append('<tr><td colspan="7" class="text-center">NENHUM REGISTRO</td></tr>')
        return
    }

    if (qtdSel) $(qtdSel).text(dados.length)

    var dadosOrdenados = dados.slice().sort(function (a, b) {
        return SemComunicacaoTimestampEvento(b.dataUltimoEvento) -
            SemComunicacaoTimestampEvento(a.dataUltimoEvento)
    })

    dadosOrdenados.forEach(function (i) {
        var horas
        var ultimoEvt
        var nuncaConectou = false

        if (i.dataUltimoEvento == '01/01/0001 00:00:00') {
            ultimoEvt = 'Nuca Conectou'
            horas = 0
            nuncaConectou = true
        } else {
            var now = new Date()
            ultimoEvt = i.dataUltimoEvento
            var past = new Date(horaBrToUs(i.dataUltimoEvento))
            var diff = Math.abs(now.getTime() - past.getTime())
            horas = Math.ceil(diff / (1000 * 60 * 60))
        }

        var tagGravidade = SemComunicacaoTagGravidade(horas, nuncaConectou)
        var descricao = i.descricaoUltimoEvento || ''
        var descricaoEvento = (descricao.length > 40) ?
            descricao.substring(0, 37) + '...' : descricao

        $tbody.append(
            '<tr style="font-size: 11px;">' +
            '<td>' + i.conta + '</td>' +
            '<td>' + i.nome + '</td>' +
            '<td>' + i.nomeCliente + '</td>' +
            '<td>' + i.codigoUltimoEvento + '</td>' +
            '<td>' + descricaoEvento + '</td>' +
            '<td>' + ultimoEvt + '</td>' +
            '<td class="fp-col-tempo text-center">' + tagGravidade + '</td>' +
            '</tr>'
        )
    })
}

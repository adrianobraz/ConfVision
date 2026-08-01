$(document).ready(function () {
    mascaraConfig()
})

function mascaraConfig() {
    $('.mskNum').mask('000')
    $('.mskReal').mask('000.000,00', {reverse: true})
    $('.mskCid').mask('AAAA')
    $('.mskCep').mask('00.000-000')
    $('.mskCel').mask('(00) 00000-0000')
    $('.mskTel').mask('(00) 0000-0000')
    $('.mskCpf').mask('000.000.000-00')
    $('.mskCnpj').mask('00.000.000/0000-00')
}


function formatarCelular(numero) {

    if (numero != undefined) {
        if (numero.length == 11) {
            return numero.replace(/(\d{2})?(\d{5})?(\d{4})/, "($1) $2-$3")
        } else if (numero.length == 10) {
            return numero.replace(/(\w{2})?(\d{4})?(\d{4})/, "($1) $2-$3")
        }

        return numero
            .replace(/\(/g, "")
            .replace(/\)/g, "")
            .replace(/\-/g, "")
            .replace(/ /g, "")
    } else {
        return ""
    }
}

function limpaDocumento(texto) {
    if (texto != undefined){
        return texto
            .replace(/\./g, "")
            .replace(/\-/g, "")
            .replace(/\//g, "")
            .replace(/\(/g, "")
            .replace(/\)/g, "")
            .replace(/\-/g, "")
            .replace(/ /g, "")
            .replace(/\:/g, "")
    }else {
        return ""
    }
}

function isBreakglass() {
    return sessionStorage.getItem('loginBreakglass') === 'S'
        && sessionStorage.getItem('loginId') === 'BREAKGLASS'
}

// Monta filtro de tenant. Retorna null se sessao invalida (usuario normal sem UUID).
// opts.todas = true: break-glass lista todas as centrais
// opts.uuid: UUID explicito (select de filtro/formulario) — NUNCA cai na sessao
function payloadTenant(extra, opts) {
    extra = extra || {}
    opts = opts || {}
    if (isBreakglass() && opts.todas) {
        extra.listarTodasCentrais = true
        extra.idCentralUUID = ''
        return extra
    }
    // UUID explicito do combo/form: nao usa loginIdCentralUUID da sessao
    if (opts.uuid !== undefined && opts.uuid !== null) {
        const uuidOpt = String(opts.uuid).trim()
        if (!uuidOpt || uuidOpt === 'TODAS') {
            if (isBreakglass()) {
                extra.listarTodasCentrais = true
                extra.idCentralUUID = ''
                return extra
            }
        } else {
            extra.idCentralUUID = uuidOpt
            extra.listarTodasCentrais = false
            return extra
        }
    }
    const uuid = sessionStorage.getItem('loginIdCentralUUID') || ''
    if (!uuid) {
        // Break-glass sem UUID na sessao: permite listar todas
        if (isBreakglass()) {
            extra.listarTodasCentrais = true
            extra.idCentralUUID = ''
            return extra
        }
        if (typeof boxErro === 'function') {
            boxErro('Sessao sem Central. Faca logout e login novamente.')
        }
        return null
    }
    extra.idCentralUUID = uuid
    extra.listarTodasCentrais = false
    return extra
}

// Break-glass: le #filtroCentral e monta payload (nao usa UUID da sessao).
// Usuario normal: payloadTenant padrao da sessao.
function payloadFiltroCentral(extra) {
    extra = extra || {}
    if (!isBreakglass()) {
        return payloadTenant(extra)
    }
    const v = ($('#filtroCentral').val() || '').trim()
    if (!v || v === 'TODAS') {
        extra.listarTodasCentrais = true
        extra.idCentralUUID = ''
        return extra
    }
    extra.listarTodasCentrais = false
    extra.idCentralUUID = v
    return extra
}

function formatarCep(cep) {
    if (cep.length == 8) {
        return cep.replace(/(\d{2})?(\d{3})?(\d{3})/, "$1.$2-$3")
    }
    return cep.replace(/\(/g, "").replace(/\)/g, "").replace(/\-/g, "").replace(/ /g, "")
}

function formatarDocumento(texto) {
    if (texto.length == 11) {
        return texto.replace(/(\d{3})?(\d{3})?(\d{3})?(\d{2})/, "$1.$2.$3-$4")
    } else if (texto.length == 14) {
        if (!texto.includes('-')) {
            return texto.replace(/(\d{2})?(\d{3})?(\d{3})?(\d{4})?(\d{2})/, "$1.$2.$3/$4-$5")
        }
    }
    return texto.replace(/\./g, "").replace(/\-/g, "").replace(/\//g, "")
}



function dataExtenso(cidade){
    const data = new Date();
    const dia = String(data.getDate()).padStart(2, '0');
    //const mes = String(data.getMonth() + 1).padStart(2, '0');
    const mes = ["Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"][data.getMonth()];
    const ano = data.getFullYear();
return `${cidade}, ${dia} de ${mes} de ${ano}`
}
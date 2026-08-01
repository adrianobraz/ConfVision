$(document).ready(function () {
    mascaraConfig()
})

// Voltar para a origem: usa ?origem= explicito, senao o historico do navegador,
// senao o fallback (href do proprio link ou /home).
function voltarOrigem(fallback) {
    try {
        var origem = new URLSearchParams(window.location.search).get("origem")
        if (origem) {
            window.location.href = decodeURIComponent(origem)
            return
        }
    } catch (e) { }

    if (window.history.length > 1) {
        window.history.back()
        return
    }

    window.location.href = fallback || "/home"
}

$(document).on("click", "a.js-voltar, button.js-voltar", function (e) {
    e.preventDefault()
    var fb = $(this).attr("href") || $(this).data("fallback") || "/home"
    voltarOrigem(fb)
})

function mascaraConfig() {
    $('.mskCep').mask('00.000-000')
    $('.mskCel').mask('(00) 00000-0000')
    $('.mskTel').mask('(00) 0000-0000')
    $('.mskCpf').mask('000.000.000-00')
    $('.mskCnpj').mask('00.000.000/0000-00') 
    $('.mskReal').mask("#.##0,00" , { reverse:true});
}


function formatarCelular(numero) {

    if (numero != undefined) {
        if (numero.length == 11) {
           return numero.replace(/(\d{2})?(\d{5})?(\d{4})/, "($1) $2-$3")           
        } else if (numero.length == 10) {
            return numero.replace(/(\w{2})?(\d{4})?(\d{4})/, "($1) $2-$3")            
        }
    }
      return numero
    
}

function limpaDocumento(texto) {

    return texto
        .replace(/\./g, "")
        .replace(/\-/g, "")
        .replace(/\//g, "")
        .replace(/\(/g, "")
        .replace(/\)/g, "")
        .replace(/\-/g, "")
        .replace(/ /g, "")
        .replace(/\:/g, "")
        alert(texto)
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



function dataExtenso(cidade) {
    const data = new Date();
    const dia = String(data.getDate()).padStart(2, '0');
    //const mes = String(data.getMonth() + 1).padStart(2, '0');
    const mes = ["Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"][data.getMonth()];
    const ano = data.getFullYear();
    return `${cidade}, ${dia} de ${mes} de ${ano}`
}
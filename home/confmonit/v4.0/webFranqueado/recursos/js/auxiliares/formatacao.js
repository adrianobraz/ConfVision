function capitalize(texto) {
    texto = texto.toLowerCase().replace(/(?:^|\s)\S/g, function (capitalize) {
        return capitalize.toUpperCase();
    });
    //preposição digitada
    var PreposM = ["Da", "Do", "Das", "Dos", "A", "E", "De", "DE", "II"];
    //preposição substituta
    var prepos = ["da", "do", "das", "dos", "a", "e", "de", "de", "II"];

    for (var i = PreposM.length - 1; i >= 0; i--) {
        texto = texto.replace(RegExp("\\b" +
            PreposM[i].replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&') + "\\b", "g"), prepos[i]);
    }

    return texto;
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

function limpaDocumento(texto) {
    if (texto != undefined) {
        return texto.replace(/\./g, "").replace(/\-/g, "").replace(/\//g, "").replace(/\(/g, "")
            .replace(/\)/g, "").replace(/\-/g, "").replace(/ /g, "").replace(/\:/g, "")
    } else {
        return ""
    }
}

function formatarCelular(numero) {
    if (numero != undefined) {
        if (numero.length == 11) {
            return numero.replace(/(\d{2})?(\d{5})?(\d{4})/, "($1) $2-$3")
        } else if (numero.length == 10) {
            return numero.replace(/(\w{2})?(\d{4})?(\d{4})/, "($1) $2-$3")
        }
        return numero.replace(/\(/g, "").replace(/\)/g, "").replace(/\-/g, "").replace(/ /g, "")
    } else {
        return ""
    }


}

function formataMac(mac) {
    if (mac == undefined) return "00:00:00:00:00:00"
    return mac.replace(/(\w{2})?(\w{2})?(\w{2})?(\w{2})?(\w{2})?(\w{2})/, "$1:$2:$3:$4:$5:$6")
}


function formatarCep(cep) {
    if (cep.length == 8) {
        return cep.replace(/(\d{2})?(\d{3})?(\d{3})/, "$1.$2-$3")
    }
    return cep.replace(/\(/g, "").replace(/\)/g, "").replace(/\-/g, "").replace(/ /g, "")
}



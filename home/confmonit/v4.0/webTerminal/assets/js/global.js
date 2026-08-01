//#################### VARIAVEIS E CONSTANTES GLOBAIS ####################\\
let gModais
let voip



let flags = {
    // Vesao do terminal
    proControles_versao: 'Versao: 4.1b',

    // Usado para controlar o timer de tempo de atendimento de evento
    ateControles_gTimer: -1,

    proEventoDetalhe_idProcesso: 'off',

    ateProDados_idProcesso: '0',

    proAtendimento_idFranqueado: 'off',
    proAtendimento_modalManutencao: '',

    proFranqFiltro_autoLimpar: -1,

}

// Constante do jquery mask
const gMkData = '00/00/0000'
const gMkHora = '00:00:00'
const gMkCep = '00000-000'
const gMkTel = '(00) 0000-0000'
const gMkCel = '(00) 00000-0000'
const gMkCpf = '000.000.000-00'
const gMkReal = '000.000.000.000,00'

//########################################################################\\

$(window).on("load", function () {
    // Carrega o setup inicial
    gSetup()

    // Carrega o js dos modulos do diretorio assests/modulos/
    moduloJs(() => {
        // Incializa o gTime para sincornismo dos modulos
        gTimer()

        // Carrega a tela processo
        setupTelaProcesso()
        //setupTelaAtendimento()
    })

    
});

// gSetup contem as configurações iniciais do do sistema
function gSetup() {
    // Configura o tela processo como defaut
    sessionStorage.setItem('gSetupTela', 'processo')

    // Tempo para o sistema atualizar as informaçoes
    sessionStorage.setItem('gAtualizaTela', '5') //em Segudos

    // Tempo para o operador atender antes de penalizar
    sessionStorage.setItem('gTempoAtendimento', '300') //em segudos

    // Caso seja ON ele fecha o atendimento após o fim do tempo
    sessionStorage.setItem('gTempoAtendimentoFechar', 'OFF')

    // Tempo para auto limpeza do filtro de franqueado
    sessionStorage.setItem('gTempoFraqAutoLimpeza', '60000') //em segudos

    
    // Receptor
    sessionStorage.setItem('gUrlComando', 'https://terminal.confmonit2.com.br/api-comando/armar')
    sessionStorage.setItem('gSenhaWeb', 'WHdQkY&RX%W%4RArwm1Q')

}

// gTimer Usado para sincronizar os modulos
function gTimer(timer = Number(sessionStorage.getItem('gAtualizaTela')) - 1) {

    // Controla o relogio do painel
    proControles_relogio()

    // Monitora tempo de atendimento
    ateControles_timer(Number(sessionStorage.getItem('gTempoAtendimento')))

    // Monitora o fechamento do filtro apos tempo determinado na variavel
    proFranqFiltro_autoLimpar()
 
    setTimeout(() => {
        if (timer == 0) {
            //Atualiza os modulos
            gModulos()
            timer = Number(sessionStorage.getItem('gAtualizaTela')) - 1
        } else {
            timer--
        }

        // Reinicia o sicro
        gTimer(timer)
        //console.log('Atualizando')
    }, 1000)
}

// gModulos é usado pelo gTimer para gerenciar os modulos
function gModulos() {
    // Tela Processo
    if (sessionStorage.getItem('gSetupTela') == 'processo') {
        // Atualiza a tabela de processo
        proAtendimento_carregarTabela()
        // Atualiza a tabela filtro de franqueado
        proFranqFiltro_carregarTabela()

        // Atuliza a tabela do proEventoDetalhe
        proEventoDetalhe_carregarTabela(sessionStorage.getItem('gTime_idProcesso'))
    }

    // Tela atendimento
    else if (sessionStorage.getItem('gSetupTela') == 'atendimento') {

    }
}

// Carrega o set de tela processo
function setupTelaProcesso() {
    sessionStorage.setItem('gSetupTela', 'processo')
    proControles_start()
    
    proAtendimento_start()

    proFranqFiltro_start()

}

// Carrega o set de tela atendimento
function setupTelaAtendimento(idProcesso) {
    sessionStorage.setItem('gSetupTela', 'atendimento')
    ateControles_start()
    ateProDados_start(idProcesso, () => {
        // Excuta apos carregar os dados do processo
        ateEventosDetalhe_start()
    })
}


async function moduloJs(next) {
    // Tela Processo
    gCarJs('proControles')
    gCarJs('proAtendimento')
    gCarJs('proUltimosAtend')
    gCarJs('proManutencoesAtivas')
    gCarJs('proFranqFiltro')
    gCarJs('proEventoDetalhe')
    gCarJs('proClienteResumo')

    // Tela Atendimento
    gCarJs('ateControles')
    gCarJs('ateProDados')
    gCarJs('ateEventosDetalhe')

    // Botoes
    gCarJs('ctrMudarSenha')
    gCarJs('ctrClienteBuscar')
    gCarJs('cliDados')
    gCarJs('ctrOrdemServico')
    gCarJs('modDiscador')
    gCarJs('modCliente')
    gCarJs('modViatura')
    gCarJs('modTecnico')
    gCarJs('modProcedimento')
    gCarJs('modGrade')
    gCarJs('modUsuarios')
    gCarJs('modSetores')
    gCarJs('modEvtDesagrupado')

    next()
}


function gLimpaDocumento(texto) {
    if (texto != undefined) {
        return texto.replace(/\./g, "").replace(/\-/g, "").replace(/\//g, "").replace(/\(/g, "")
            .replace(/\)/g, "").replace(/\-/g, "").replace(/ /g, "").replace(/\:/g, "")
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

function gGeraTag(menssagem, data = new Date()) {
    const dia = String(data.getDate()).padStart(2, '0')
    const mes = String(data.getMonth() + 1).padStart(2, '0')
    const ano = data.getFullYear();
    const hora = String(data.getHours()).padStart(2, '0')
    const minuto = String(data.getMinutes()).padStart(2, '0')
    const segundo = String(data.getSeconds()).padStart(2, '0')
    const final = dia + '/' + mes + '/' + ano + ' - ' + hora + ':' + minuto + ':' + segundo
    return `[${final} - ${menssagem}]`
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


function armarCentral(idDisp, senha, numero, next, onFail) {
    msgProcessando("")

    $.ajax({
        url: sessionStorage.getItem('gUrlComando'),
        method: 'Post',
        contentType: 'application/json',
        data: JSON.stringify({
            "idDispositivo": idDisp,
            "numero": Number(numero),
            "usuario": "000",
            "acao": 1,
            "senha": senha,
            "senhaWeb": sessionStorage.getItem('gSenhaWeb')
        })
    }).fail(function (e) {
        console.log(e)
        const status = (e && e.responseJSON && e.responseJSON.status) ? e.responseJSON.status : ''
        if (status == 'Erro: não conectado') {
            msgErro('Dispositivo esta Offline')
        } else if (status != '') {
            msgErro(status)
        } else {
            msgErro('Erro ao enviar comando de arme')
        }
        if (typeof onFail == 'function') {
            onFail(e)
        }
    }).done(function (r) {
        // Aqui é apenas "comando enviado", não status final confirmado.
        if (typeof next == 'function') {
            next()
        }
    }).always(function () {
        msgFechar()
    })

}

function desarmarCentral(idDisp, senha, numero, next, onFail) {
    msgProcessando("")

    $.ajax({
        url: sessionStorage.getItem('gUrlComando'),
        method: 'Post',
        contentType: 'application/json',
        data: JSON.stringify({
            "idDispositivo": idDisp,
            "numero": Number(numero),
            "usuario": "000",
            "acao": 0,
            "senha": senha,
            "senhaWeb": sessionStorage.getItem('gSenhaWeb')
        })
    }).fail(function (e) {
        console.log(e)
        const status = (e && e.responseJSON && e.responseJSON.status) ? e.responseJSON.status : ''
        if (status == 'Erro: não conectado') {
            msgErro('Dispositivo esta Offline')
        } else if (status != '') {
            msgErro(status)
        } else {
            msgErro('Erro ao enviar comando de desarme')
        }
        if (typeof onFail == 'function') {
            onFail(e)
        }
    }).done(function (r) {
        // Aqui é apenas "comando enviado", não status final confirmado.
        if (typeof next == 'function') {
            next()
        }
    }).always(function () {
        msgFechar()
    })

}
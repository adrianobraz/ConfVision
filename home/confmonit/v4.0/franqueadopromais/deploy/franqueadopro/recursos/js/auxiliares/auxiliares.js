$(document).ready(function () {
    carregaCorCarro()
    carregarUf()
    fpNavPreencherUsuario()
    $('#logout').on('click', logout)
})

/** Nome do franqueado + e-mail do usuário na navbar (discreto). */
function fpNavPreencherUsuario() {
    var nome = (localStorage.getItem('nomeFranqueado') || '').trim()
    var email = (localStorage.getItem('email') || '').trim()
    var $chip = $('#fp-user-chip')
    var $nome = $('#fp-nav-nome-franqueado')
    var $email = $('#fp-nav-email')
    if (!$chip.length) return
    if (!nome && !email) {
        $chip.removeClass('is-visible').attr('hidden', true)
        return
    }
    $nome.text(nome || (localStorage.getItem('nomeUsuario') || ''))
    $email.text(email)
    $chip.addClass('is-visible').removeAttr('hidden')
}

function carregaCorCarro() {
    $('.carregaCorCarro').empty()

    $('.carregaCorCarro').append(`
        <option value="AMARELO">Amarelo</option>
        <option value="AZUL">Azul</option>
        <option value="BRANCO">Branco</option>
        <option value="CINZA">Cinza</option>
        <option value="PRATA">Prata</option>
        <option value="BEGE">Bege</option>
        <option value="PRETO">Preto</option>
        <option value="VERDE">Verde</option>
        <option value="VERMELHO">Vermelho</option>
    `)
}


function carregarUf() {
    $('.carregarUf').empty()

    $('.carregarUf').append(`
    <option value="AC">Acre</option>
    <option value="AL">Alagoas</option>
    <option value="AP">Amapá</option>
    <option value="AM">Amazonas</option>
    <option value="BA">Bahia</option>
    <option value="CE">Ceará</option>
    <option value="DF">Distrito Federal</option>
    <option value="ES">Espírito Santo</option>
    <option value="GO">Goiás</option>
    <option value="MA">Maranhão</option>
    <option value="MT">Mato Grosso</option>
    <option value="MS">Mato Grosso do Sul</option>
    <option value="MG">Minas Gerais</option>
    <option value="PA">Pará</option>
    <option value="PB">Paraíba</option>
    <option value="PR">Paraná</option>
    <option value="PE">Pernambuco</option>
    <option value="PI">Piauí</option>
    <option value="RJ">Rio de Janeiro</option>
    <option value="RN">Rio Grande do Norte</option>
    <option value="RS">Rio Grande do Sul</option>
    <option value="RO">Rondônia</option>
    <option value="RR">Roraima</option>
    <option value="SC">Santa Catarina</option>
    <option value="SP">São Paulo</option>
    <option value="SE">Sergipe</option>
    <option value="TO">Tocantins</option>
    <option value="EX">Estrangeiro</option>
    `)
}

function logout() {
    localStorage.clear()
}

// geradorIdCurto gera um id com 10 posicoes

function geradorIdCurto() {
    const agora = new Date()
    const ano = agora.getUTCFullYear().toString()
    const mes = addZeroEsquerda(agora.getMonth() + 1, 2)
    const dia = addZeroEsquerda(agora.getDate(), 2)
    const randon = addZeroEsquerda(Math.floor(Math.random() * 10000), 4)

    return ano.substring(2) + mes + dia + randon

}


// addZeroEsquerda adiciona zeros a esquerda conforme quatidade de len
function addZeroEsquerda(num, len) {
    var numberWithZeroes = String(num);
    var counter = numberWithZeroes.length;

    while (counter < len) {

        numberWithZeroes = "0" + numberWithZeroes

        counter++

    }

    return numberWithZeroes
}

function ajustaTabela() {
    const altura = window.innerHeight
        || document.documentElement.clientHeight
        || document.body.clientHeight

    var extra = 0
    var pageHdr = document.querySelector('.fp-page-header-wrap')
    if (pageHdr) {
        extra = pageHdr.offsetHeight + 10
    }

    $('.page-height').animate({ height: (altura - 190 - extra) + 'px' })
    $('.form-height').animate({ height: (altura - 380 - extra) + 'px' })
    $('.tab-height').animate({ height: (altura - 190 - extra) + 'px' })
}


function log(status, dado){
    if (status == 'ON' ) {
        console.log(dado)
    }
}
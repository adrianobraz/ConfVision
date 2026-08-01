function proControles_start() {
    $('#boxTop').empty()
    $('#boxTop').load('/assets/modulos/proControles/proControles.html', () => {
        // Cofiguração inicial
        proControles_setup(()=>{

            $('#proControles_cpVersao').html(flags.proControles_versao)
    
            $('#proControles_cpAtendente').html(sessionStorage.getItem('login_userNick'))
    
            $('#proControles_btnAlterarSenha').on('click', proControles_btnAlterarSenha)

            $('#proControles_btnManutencoesAtivas').on('click', proControles_btnManutencoesAtivas)
    
            $('#proControles_btnDadosCliente').on('click', proControles_btnDadosCliente)
    
            $('#proControles_btnOs').on('click', proControles_btnOs)

            $('#proControles_btnUltimosAtendimentos').on('click', proControles_btnUltimosAtendimentos)
    
            if (sessionStorage.getItem('proControles_audio') == "ON"){
                $('#proControles_ledA').removeClass('ledOff')
            }else{
                $('#proControles_ledA').addClass('ledOff')
            }
            
        })
    })
}

function proControles_setup(next){
    if (sessionStorage.getItem('proControles_audio') === null){
        sessionStorage.setItem('proControles_audio', 'OFF')
    }
    
    next()
}

function proControles_btnAdio() {
    if (sessionStorage.getItem('proControles_audio') == "ON"){ 
        sessionStorage.setItem('proControles_audio', 'OFF')
        $('#proControles_ledA').addClass('ledOff')
    }else{
        sessionStorage.setItem('proControles_audio', 'ON')
        $('#proControles_ledA').removeClass('ledOff')
    }
}

function proControles_btnAlterarSenha() {
    ctrMudarSenha_start()
}

function proControles_btnDadosCliente() {
    ctrClienteBuscar_start()
}

function proControles_btnManutencoesAtivas() {
    if (typeof proManutencoesAtivas_start === 'function') {
        proManutencoesAtivas_start()
    }
}

function proControles_btnOs() {
    ctrOrdemServico_start()
}

function proControles_btnUltimosAtendimentos() {
    if (typeof proUltimosAtend_start === 'function') {
        proUltimosAtend_start()
        return
    }
    if (typeof setupTelaProcesso === 'function') {
        setupTelaProcesso()
        setTimeout(function () {
            if (typeof proUltimosAtend_start === 'function') {
                proUltimosAtend_start()
            }
        }, 500)
    }
}

function proControles_btnLogout() { 
    sessionStorage.clear()
    window.location.href = '/'
}

function proControles_relogio() {
    
        // Obtém a data/hora atual
        var Agora = new Date();
        var Data = '';
        Data += ("00" + Agora.getDate()).slice(-2);
        Data += '/';
        Data += ("00" + (Agora.getMonth() + 1)).slice(-2);
        Data += '/';
        Data += Agora.getFullYear();

        var Hora = '';
        Hora += ("00" + Agora.getHours()).slice(-2);
        Hora += ':';
        Hora += ("00" + Agora.getMinutes()).slice(-2);
        Hora += ':';
        Hora += ("00" + Agora.getSeconds()).slice(-2);

        $('#proControles_cpRelogio').html(Data + ' - ' + Hora)
}

function proControles_qtdProcessos(valor){
    $('#proControles_cpQtdEventos').html(valor)
}


$( document ).ready(function() {   
    configurarBotoes()    
    $('#botao-dados-monitoramento').attr('hidden', false)
    $('#botao-gerenciar-responsavel').attr('hidden', false)
    $('#botao-gerenciar-tecnico').attr('hidden', false)
    $('#botao-gerenciar-cliente').attr('hidden', false)
    $('#botao-gerenciar-restricoes-tecnico').attr('hidden', true)
    $('#botao-gerenciar-viatura').attr('hidden', false)
    
    if (localStorage.getItem('terminalAtivo') == 1){
        $('#botao-cadastro-operador').attr('hidden', false)
    }else{
        $('#botao-cadastro-operador').attr('hidden', true)
    }
    
    $('#botao-gerenciar-ticket').attr('hidden', false)
});

function configurarBotoes(){
    $('#btnGerenciarPacote').on('click', function(){ 
        window.location.href = "/gerenciarPacotesPage"
    })
}
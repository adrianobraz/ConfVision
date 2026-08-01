function  gCarJs(modulo) {
    $.ajax({
        async: false,
        url: `/assets/modulos/${modulo}/${modulo}.js`,
        dataType: "script"
    }).done((script) => {
        $('#scriptJs').append(script)
    })
}

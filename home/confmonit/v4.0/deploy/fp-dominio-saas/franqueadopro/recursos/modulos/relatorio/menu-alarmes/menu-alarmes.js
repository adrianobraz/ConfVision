$(document).ready(function () {
    MenuAlarmesCarregarContadores();
});

function MenuAlarmesCarregarContadores() {
    MenuAlarmesContar("/dispositivoListarArmado", "S", "#count-armados");
    MenuAlarmesContar("/DispositivoListarDesarmado", "N", "#count-desarmados");
}

// Consulta o endpoint, conta os dispositivos com o "armado" desejado e atualiza o card.
function MenuAlarmesContar(url, estadoArmado, alvo) {
    $.ajax({
        url: url,
        method: "Post",
        data: JSON.stringify({
            idFranqueado: localStorage.getItem("idFranqueado"),
        }),
    })
        .fail(function (e) {
            console.log(e);
            $(alvo).text("--");
        })
        .done(function (r) {
            var total = 0;
            if (r.status != "Vazio" && r.dados) {
                total = r.dados.filter(function (i) {
                    return i.armado == estadoArmado;
                }).length;
            }
            $(alvo).text(total);
        });
}

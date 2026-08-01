$(window).on("load", function () {
  $("#btnLimpar").on("click", limpar);

  $("#btnFiltrar").on("click", filtrar);

  $("#btnImprimir").on("click", function () {
    gerarPdf(
      "RELATÓRIO DE EVENTOS",
      "#tabela",
      "landscape",
      "TOTAL DE EVENTOS: " + $("#qtdEventos").html()
    );
  });

  carregarDispitivo();
});

function limpar() {}

function carregarDispitivo() {
  $.ajax({
    url: `/relatorio/evento/carregarDispositivo`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      $("#dispositivo").empty();
      if (r.status != "Vazio") {
        r.dados.map((i) => {
          $("#dispositivo").append(`
                    <option value="${i.idDispositivo}">${i.nome}</option>
                `);
        });
      }
    });
}

function filtrar() {
  if ($("#dtInicial").val() == "") {
    boxErro("A data inicial não pode ficar em branco");
    return;
  }

  if ($("#dtFinal").val() == "") {
    boxErro("A data final não pode ficar em branco");
    return;
  }

  const s = moment($("#dtInicial").val()); // data atual
  if (s.isAfter(moment())) {
    boxErro("A data inicial não pode ser maior que a data atual");
    return;
  }

  const f = moment($("#dtFinal").val());
  if (s.isAfter(f)) {
    boxErro("A data inicial não pode ser maior que a data final");
    return;
  }

  const horas = f.diff(s, "hours");

  if (horas > 744) {
    boxErro("A data inicial para data final não pode ser maior que 31 dias");
    return;
  }

  let evt = "";
  if ($("#alarme").is(":checked")) evt = '"ALARME",';
  if ($("#arme").is(":checked")) evt = evt + '"ARME",';
  if ($("#desarme").is(":checked")) evt = evt + '"DESARME",';
  if ($("#panico").is(":checked")) evt = evt + '"PANICO",';
  if ($("#geral").is(":checked")) evt = evt + '"GERAL",';
  if ($("#medico").is(":checked")) evt = evt + '"MEDICO",';
  if ($("#falhas").is(":checked")) evt = evt + '"FALHAS",';
  if ($("#setup").is(":checked")) evt = evt + '"SETUP",';
  if ($("#teste").is(":checked")) evt = evt + '"TESTE",';
  if ($("#emergencia").is(":checked")) evt = evt + '"EMERGENCIA",';

  if (evt != "") {
    evt = evt.slice(0, -1);

    $.ajax({
      start: boxProcessando(),
      url: `/relatorio/evento/filtrar`,
      method: "POST",
      data: JSON.stringify({
        dispId: $("#dispositivo").val(),
        dataInicio: $("#dtInicial").val(),
        dataFim: $("#dtFinal").val(),
        grupos: evt,
      }),
    })
      .fail(function (e) {
        console.log(e);
      })
      .done(function (r) {
        boxFechar();
        $("tbody").empty();
        if (r.status != "Vazio") {
          $("#qtdEventos").html(r.dados.length);

          r.dados.map((i) => {
            $("tbody").append(`
                        <tr>
                            <td>${i.dataEntrada}</td>
                            <td>${i.dispConta}</td>
                            <td>${i.codigo}</td>
                            <td>${i.ctiDescricao}</td>
                            <td>${i.zonaUser}</td>
                            <td>${i.zonaUserDescricao}</td>
                        </tr>
                    `);
          });
        }
      });
  } else {
    boxErro("Pelomenos um grupo de evento deve ser selecionado");
  }
}

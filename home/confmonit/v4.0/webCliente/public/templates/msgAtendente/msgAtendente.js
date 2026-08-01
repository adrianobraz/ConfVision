$(window).on("load", function () {
  carregarDispitivo();

  $("#btnLimpar").on("click", () => {
    $("#msg").val("");
  });

  $("#btnGravar").on("click", gravar);
  $("#dispositivo").on("change", carregarMenssagem);
});

function carregarDispitivo() {
  $.ajax({
    url: `/msgAtendente/carregarDispositivo`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      //   console.log(r);
      $("#dispositivo").empty();
      if (r.status != "Vazio") {
        r.dados.map((i) => {
          $("#dispositivo").append(`
                    <option value="${i.idDispositivo}">${i.nome}</option>
                `);
        });
        carregarMenssagem();
      } else {
        $("#dispositivo").append(`
                <option value="0">SELECIONE</option>
            `);
      }
    });
}

function carregarMenssagem() {
  if ($("#dispositivo").val() == "0") {
    return;
  }

  $.ajax({
    url: `/msgAtendente/carregarMensagem`,
    method: "POST",
    data: JSON.stringify({
      idDispositivo: $("#dispositivo").val(),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      //   console.log(r);
      $("#msg").val("");
      if (r.status != "Vazio") {
        $("#msg").val(r.dados.msgAtendente);
      }
    });
}

function gravar() {
  $.ajax({
    url: `/msgAtendente/gravar`,
    method: "POST",
    data: JSON.stringify({
      idDispositivo: $("#dispositivo").val(),
      msgAtendente: $("#msg").val(),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      boxSucesso("Menssagem gravada com sucesso");
      carregarMenssagem();
    });
}

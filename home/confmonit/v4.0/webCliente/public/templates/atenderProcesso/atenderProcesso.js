$(window).on("load", function () {
  getDadosProcesso(sessionStorage.getItem("idProcesso"));
  listarEventos(sessionStorage.getItem("idProcesso"));
  $("#btnVoltar").on("click", () => {
    window.location = "/atendimento/page";
  });

  $("#btnFinalizar").on("click", finalizarProcesso);
});

function getDadosProcesso(id) {
  $.ajax({
    url: `/atenderProcesso/getProcessoById`,
    method: "POST",
    data: JSON.stringify({
      idProcesso: id,
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r);
      if (r.status == "OK") {
        const d = r.dados;
        $("#nome").val(d.dispNome);
      }
    });
}

function listarEventos(id) {
  $.ajax({
    url: `/atenderProcesso/listarEventosByProcesso`,
    method: "POST",
    data: JSON.stringify({
      idProcesso: id,
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r);
      $("tbody").empty();
      if (r.status == "OK") {
        r.dados.map((i) => {
          $("tbody").append(`
            <tr>
              <td>${i.codigo}</td>
              <td>${i.userSet}</td>
              <td>${i.descricao}</td>
              <td>${i.dataEntrada}</td>
              
            </tr>
            `);
        });
      }
    });
}

function finalizarProcesso() {
  if ($("#descricao").val() == "") {
    boxAdvertenciaCampoAuto(
      "O campo descrição não pode ficar em branco",
      "#descricao"
    );
    return;
  }
  $.ajax({
    url: `/atenderProcesso/finalizarProcesso`,
    method: "POST",
    data: JSON.stringify({
      idProcesso: sessionStorage.getItem("idProcesso"),
      idCliente: sessionStorage.getItem("loginId"),
      descricao: $("#descricao").val(),
      nome: sessionStorage.getItem("loginNome"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r);
      window.location = "/atendimento/page";
    });
}

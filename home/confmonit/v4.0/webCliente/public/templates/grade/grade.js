$(window).on("load", function () {
  carregarDispitivo();
  $("#dispositivo").on("change", carregarTabela);
});

function carregarDispitivo() {
  $.ajax({
    url: `/grade/carregarDispositivo`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r)
      $("#dispositivo").empty();
      if (r.status != "Vazio") {
        r.dados.map((i) => {
          $("#dispositivo").append(`
                    <option value="${i.idDispositivo}">${i.nome}</option>
                `);
        });
        carregarTabela();
      }
    });
}

function carregarTabela() {
  $.ajax({
    url: `/grade/carregarTabela`,
    method: "POST",
    data: JSON.stringify({
      idDispositivo: $("#dispositivo").val(),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r)
      $("tbody").empty();
      if (r.status != "Vazio") {
        r.dados.map((i) => {
          let btnCor;
          let btnIcon;

          if (i.ativo == "S") {
            btnCor = "bg-green-800";
            btnIcon = '<i class="bi bi-toggle-on"></i>';
          } else {
            btnCor = "bg-slate-700";
            btnIcon = '<i class="bi bi-toggle-off"></i>';
          }

          $("tbody").append(`
                    <tr>
                        <td>${i.nome}</td>
                        <td class="${btnCor}" tipo="btnHabilitar" idGrade="${i.idGrade}">${btnIcon}</td>
                        <td class="bg-blue-800" tipo="btnVisualizar" idGrade="${i.idGrade}"><i class="bi bi-eye-fill"></i></td>
                    </tr>    
                `);
        });
        $("td[tipo=btnHabilitar").on("click", function () {
          const idGrade = this.getAttribute("idGrade");
          habilitarGrade(idGrade);
        });

        $("td[tipo=btnVisualizar").on("click", function () {
          const idGrade = this.getAttribute("idGrade");
          visualizarGrade(idGrade);
        });
      }
    });
}

function visualizarGrade(idGrade) {
  $.ajax({
    url: `/grade/visualizarGrade`,
    method: "POST",
    data: JSON.stringify({ idGrade: idGrade }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r)
      if (r.status == "OK") {
        const d = r.dados;
        $("#nome").val(d.nome);
        $("#tolerancia").val(d.tolerancia);
        $("#domEam").val(d.domEam);
        $("#domSam").val(d.domSam);
        $("#domEpm").val(d.domEpm);
        $("#domSpm").val(d.domSpm);
        $("#segEam").val(d.segEam);
        $("#segSam").val(d.segSam);
        $("#segEpm").val(d.segEpm);
        $("#segSpm").val(d.segSpm);
        $("#terEam").val(d.terEam);
        $("#terSam").val(d.terSam);
        $("#terEpm").val(d.terEpm);
        $("#terSpm").val(d.terSpm);
        $("#quaEam").val(d.quaEam);
        $("#quaSam").val(d.quaSam);
        $("#quaEpm").val(d.quaEpm);
        $("#quaSpm").val(d.quaSpm);
        $("#quiEam").val(d.quiEam);
        $("#quiSam").val(d.quiSam);
        $("#quiEpm").val(d.quiEpm);
        $("#quiSpm").val(d.quiSpm);
        $("#sexEam").val(d.sexEam);
        $("#sexSam").val(d.sexSam);
        $("#sexEpm").val(d.sexEpm);
        $("#sexSpm").val(d.sexSpm);
        $("#sabEam").val(d.sabEam);
        $("#sabSam").val(d.sabSam);
        $("#sabEpm").val(d.sabEpm);
        $("#sabSpm").val(d.sabSpm);
      }
    });
}

function habilitarGrade(idGrade) {
  $.ajax({
    url: `/grade/habilitarGrade`,
    method: "POST",
    data: JSON.stringify({ idGrade: idGrade }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      carregarTabela();
    });
}

$(document).ready(function () {
  carregarTabela();
});

function carregarTabela() {
  $(".carregarBases ").empty();

  $.ajax({
    url: "/DispositivoListarDesarmado",
    method: "Post",
    data: JSON.stringify({
      idFranqueado: localStorage.getItem("idFranqueado"),
    }),
  })
    .fail(function (e) {
      boxErro("Erro ao carregar tabela");
    })
    .done(function (r) {
      // console.log(r)
      $("tbody").empty();
      if (r.status != "Vazio") {
        r.dados.forEach((i) => {
          if (i.armado == "N") {
            const dataUltimoEvento =
              i.dataUltimoEvento == "" ? "NUNCA CONECTOU" : i.dataUltimoEvento;
            $("tbody").append(`
                       <tr>
                            <td>${i.conta}</td>
                            <td>${i.nomeCliente}</td>
                            <td>${i.nome}</td>
                            <td>${i.dataArmado}</td>
                            <td>${i.codigoUltimoEvento}</td>
                            <td>${dataUltimoEvento}</td> 
                        </tr>
                    `);
          }
        });
      } else {
        $("tbody").append(
          `<tr><td colspan="5" class="text-center">LISTA VAZIA</td></tr>`
        );
      }
    });
}

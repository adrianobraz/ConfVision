$(document).ready(function () {
  $("#btn-limpar").on("click", limparFormulario);
  $("#btn-filtrar").on("click", filtrar);
  $("#btn-imprimir").on("click", function () {
    gerarPdf(
      "RELATÓRIO DE EVENTOS",
      "#tabEventos",
      "landscape",
      $("#qtd-eventos").text()
    );
  });

  $("#id-cliente").on("change", function () {
    $("tbody").empty();
    $("#qtd-eventos").text("Total de Eventos: 0");

    if ($('#id-cliente').val() == 'TODOS'){
       $("#idDispositivo")
          .empty()
          .append(`<option value="TODOS">SELECIONE</option>`);
    }else{
      carregarDispositivo();
    }
  });

  carregarClientes();
});

function filtrar() {
  if ($('#idDispositivo').val() == ''){
    boxMesagemAtencaoPersonalizada("Um dispositivo deve ser informado");
  }
  const idDispositivo = $("#idDispositivo").val();
  let start = $("#start").val();
  let end = $("#end").val();

  if (start != "") {
    start = start.replace("T", " ") + ":00";
  } else {
    boxMesagemAtencaoPersonalizada("Data inicial deve ser informada");
    return;
  }

  if (end != "") {
    end = end.replace("T", " ") + ":00";
  } else {
    boxMesagemAtencaoPersonalizada("Data final deve ser informada");
    return;
  }

  const s = moment(start); // data atual
  if (s.isAfter(moment())) {
    boxMesagemAtencaoPersonalizada(
      "A data inicial não pode ser maior que a data atual"
    );
    return;
  }

  const f = moment(end);
  if (s.isAfter(f)) {
    boxMesagemAtencaoPersonalizada(
      "A data inicial não pode ser maior que a data final"
    );
    return;
  }

  const horas = f.diff(s, "hours");
 
  if ($('#id-cliente').val() == 'TODOS' )
  {
     if (horas > 168) {
      boxMesagemAtencaoPersonalizada(
        "Na modalidade todos data inicial para data final não pode ser maior que 7 dias"
      );
      return;
    }
  }else {
    if (horas > 744) {
      boxMesagemAtencaoPersonalizada(
        "A data inicial para data final não pode ser maior que 31 dias"
      );
      return;
    }
  }

  carregarTabela(start, end);
}

function carregarTabela(dataInicio, dataFim) {
  $.ajax({
    start: boxProcessando(),
    url: "/relatorioEventosListar",
    method: "Post",
    data: JSON.stringify({
      dispId: $("#idDispositivo").val(),
      dataInicio: dataInicio,
      dataFim: dataFim,
    }),
  })
    .fail(function (e) {
      console.log(e);
      boxErro("Erro ao carregar a tabela");
    })
    .done(function (r) {
      console.log(r);
      boxFechar();
      $("tbody").empty();
      if (r.status != "Vazio") {
        const qtd = r.dados.length;
        $("#qtd-eventos").text("Total de Eventos: " + qtd);

        r.dados.forEach((i) => {
          $("tbody").append(`
                    <tr>
                        <td>${i.dataEntrada}</td>
                        <td class="text-start">${i.dispConta} - ${i.cliNome}</td>
                        <td>${i.codigo}</td>
                        <td class="text-start">${i.ctiDescricao}</td>
                        <td>${i.zonaUser}</td>
                        <td class="text-start">${i.zonaUserDescricao}</td>
                    </tr>
                `);
        });
      }
    });
}

function carregarClientes() {
  $.ajax({
    url: "/RelatorioEventosCarregarClientes",
    method: "Post",
    data: JSON.stringify({
      idFranqueado: localStorage.getItem("idFranqueado"),
    }),
  })
    .fail(function (e) {
      console.log(e);
      boxErro("Erro ao carregar os clientes");
    })
    .done(function (r) {
      $("#id-cliente").empty().append(`<option value="">SELECIONE</option>`).append(`<option value="TODOS">TODOS</option>`);
      if (r.status != "Vazio") {
        r.dados.forEach((i) => {
          $("#id-cliente").append(
            `<option value="${i.idCliente}">${i.nome}</option>`
          );
        });
      }
    });
}

function carregarDispositivo() {
  const idCliente = $("#id-cliente").val();

  if (idCliente != "") {
    $.ajax({
      url: "/relatorioEventosCarregarDispositivos",
      method: "Post",
      data: JSON.stringify({
        idCliente: idCliente,
      }),
    })
      .fail(function (e) {
        console.log(e);
        boxErro("Erro ao carregar os dispositivos");
      })
      .done(function (r) {
        console.log(r);
        $("#idDispositivo")
          .empty()
          .append(`<option value="">SELECIONE</option>`);

        if (r.status != "Vazio") {
          r.dados.forEach((i) => {
            $("#idDispositivo").append(
              `<option value="${i.idDispositivo}">${i.nome}</option>`
            );
          });
        }
      });
  }
}

function limparFormulario() {
  $("#formulario").each(function () {
    this.reset();
  });
  $("tbody").empty();
}

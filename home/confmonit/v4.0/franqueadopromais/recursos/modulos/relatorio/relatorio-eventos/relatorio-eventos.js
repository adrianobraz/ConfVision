var listaClientes = [];
var fpStart, fpEnd;
var fpPagRelEventos = null;
var fpPagRelParams = {};

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

    if ($("#id-cliente").val() == "TODOS") {
      $("#idDispositivo")
        .empty()
        .append(`<option value="TODOS" selected>TODOS OS DISPOSITIVOS</option>`);
    } else {
      carregarDispositivo();
    }
  });

  $("#busca-cliente").on("input", filtrarClientes);
  $("#busca-cliente").on("focus", filtrarClientes);
  $("#lista-clientes").on("click", ".fp-combo-item", function () {
    $("#busca-cliente").val($(this).attr("data-nome"));
    $("#id-cliente").val($(this).attr("data-id")).trigger("change");
    $("#lista-clientes").addClass("d-none");
  });
  $(document).on("click", function (e) {
    if (!$(e.target).closest("#busca-cliente, #lista-clientes").length) {
      $("#lista-clientes").addClass("d-none");
    }
  });

  if (window.flatpickr) {
    if (flatpickr.l10ns && flatpickr.l10ns.pt) {
      flatpickr.localize(flatpickr.l10ns.pt);
    }
    var cfg = {
      enableTime: true,
      time_24hr: true,
      dateFormat: "Y-m-d H:i",
      altInput: true,
      altFormat: "d/m/Y H:i",
    };
    fpStart = flatpickr("#start", cfg);
    fpEnd = flatpickr("#end", cfg);
  }

  carregarClientes(function () {
    carregarPadraoUltimos100();
  });
});

function datasPadrao7Dias() {
  var fim = moment();
  var ini = moment().subtract(7, "days");
  if (fpStart) fpStart.setDate(ini.toDate(), false);
  if (fpEnd) fpEnd.setDate(fim.toDate(), false);
  return {
    start: ini.format("YYYY-MM-DD HH:mm") + ":00",
    end: fim.format("YYYY-MM-DD HH:mm") + ":00",
  };
}

function carregarPadraoUltimos100() {
  if (!idFranqueadoAtual()) return;
  $("#busca-cliente").val("TODOS OS CLIENTES");
  $("#id-cliente").val("TODOS");
  $("#idDispositivo")
    .empty()
    .append(
      `<option value="TODOS" selected>TODOS OS DISPOSITIVOS</option>`
    );
  var d = datasPadrao7Dias();
  carregarTabela(d.start, d.end, "TODOS");
}

function clienteTextoBusca(i) {
  return [(i.nome || ""), (i.nick || ""), (i.documento1 || ""), (i.documento2 || "")]
    .join(" ")
    .toLowerCase();
}

function filtrarClientes() {
  var raw = ($("#busca-cliente").val() || "").toLowerCase().trim();
  var termoNum = raw.replace(/[^a-z0-9]/g, "");
  var lista = $("#lista-clientes");
  lista.empty();

  $('<li class="fp-combo-item"></li>')
    .attr("data-id", "TODOS")
    .attr("data-nome", "TODOS OS CLIENTES")
    .html("<strong>TODOS OS CLIENTES</strong>")
    .appendTo(lista);

  var base;
  if (raw == "") {
    base = listaClientes.slice(0, 20);
  } else {
    base = listaClientes
      .filter(function (i) {
        var texto = clienteTextoBusca(i);
        var textoNum = texto.replace(/[^a-z0-9]/g, "");
        return (
          texto.indexOf(raw) != -1 ||
          (termoNum != "" && textoNum.indexOf(termoNum) != -1)
        );
      })
      .slice(0, 30);
  }

  base.forEach(function (i) {
    var doc = i.documento1 ? " — " + i.documento1 : "";
    $('<li class="fp-combo-item"></li>')
      .attr("data-id", i.idCliente)
      .attr("data-nome", i.nome)
      .text(i.nome + doc)
      .appendTo(lista);
  });

  lista.removeClass("d-none");
}

function idFranqueadoAtual() {
  return (
    localStorage.getItem("idFranqueado") ||
    $("#id-franqueado").val() ||
    ""
  ).trim();
}

function filtrar() {
  var idCliente = String($("#id-cliente").val() || "").trim();
  if (!idCliente) {
    boxMesagemAtencaoPersonalizada("Um cliente deve ser informado");
    return;
  }

  // Modalidade todos: forca dispId=TODOS (nao depende do select)
  if (idCliente.toUpperCase() === "TODOS") {
    $("#idDispositivo")
      .empty()
      .append(`<option value="TODOS" selected>TODOS OS DISPOSITIVOS</option>`);
  }

  var dispId = String($("#idDispositivo").val() || "").trim();
  if (!dispId) {
    boxMesagemAtencaoPersonalizada("Um dispositivo deve ser informado");
    return;
  }

  let start = $("#start").val();
  let end = $("#end").val();

  if (start != "") {
    start = start.replace("T", " ");
    if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/.test(start)) {
      start = start + ":00";
    }
  } else {
    boxMesagemAtencaoPersonalizada("Data inicial deve ser informada");
    return;
  }

  if (end != "") {
    end = end.replace("T", " ");
    if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/.test(end)) {
      end = end + ":00";
    }
  } else {
    boxMesagemAtencaoPersonalizada("Data final deve ser informada");
    return;
  }

  const s = moment(start);
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

  if (idCliente.toUpperCase() === "TODOS") {
    if (horas > 168) {
      boxMesagemAtencaoPersonalizada(
        "Na modalidade todos data inicial para data final não pode ser maior que 7 dias"
      );
      return;
    }
    if (!idFranqueadoAtual()) {
      boxMesagemAtencaoPersonalizada(
        "Sessão sem franqueado. Faça login novamente."
      );
      return;
    }
  } else {
    if (horas > 744) {
      boxMesagemAtencaoPersonalizada(
        "A data inicial para data final não pode ser maior que 31 dias"
      );
      return;
    }
  }

  carregarTabela(start, end, dispId);
}

function carregarTabela(dataInicio, dataFim, dispId) {
  if (fpPagRelEventos) fpPagRelEventos.destroy();
  fpPagRelParams = {
    dispId: dispId || $("#idDispositivo").val(),
    dataInicio: dataInicio,
    dataFim: dataFim,
    idFranqueado: idFranqueadoAtual(),
  };

  $("#fp-rel-eventos-hint").text("");
  var $scroll = $("#fp-rel-eventos-scroll");
  if ($scroll.length) $scroll.scrollTop(0);

  fpPagRelEventos = fpScrollPaginacao({
    url: "/relatorioEventosListar",
    tbodySel: "#tabEventos tbody",
    scrollTarget: $scroll.length ? $scroll.get(0) : window,
    getPayload: function () {
      return {
        dispId: fpPagRelParams.dispId,
        dataInicio: fpPagRelParams.dataInicio,
        dataFim: fpPagRelParams.dataFim,
        idFranqueado: fpPagRelParams.idFranqueado,
      };
    },
    renderRows: function (dados) {
      dados.forEach(function (i) {
        $("#tabEventos tbody").append(`
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
    },
    onEmpty: function () {
      $("#qtd-eventos").text("Total de Eventos: 0");
      $("#fp-rel-eventos-hint").text("");
    },
    onTotal: function (total) {
      var carregados = fpPagRelEventos.getOffset();
      var n = total != null ? total : carregados;
      $("#qtd-eventos").text("Total de Eventos: " + n);
      if (total != null && total > 0) {
        $("#fp-rel-eventos-hint").text(
          "Exibindo " + carregados + " de " + total + " — role para carregar mais"
        );
      } else {
        $("#fp-rel-eventos-hint").text("");
      }
    },
    onDone: function (primeira, r) {
      var total = fpPagRelEventos.getTotal();
      var carregados = fpPagRelEventos.getOffset();
      if (total != null && carregados >= total && total > 0) {
        $("#fp-rel-eventos-hint").text("Exibindo todos os " + total + " eventos");
      }
    },
    onFail: function (e) {
      var msg = "Erro ao carregar a tabela";
      try {
        if (e && e.responseJSON && e.responseJSON.status) {
          msg = String(e.responseJSON.status);
        }
      } catch (err) {}
      boxErro(msg);
    },
  });

  fpPagRelEventos.reset();
}

function carregarClientes(aoPronto) {
  $.ajax({
    url: "/RelatorioEventosCarregarClientes",
    method: "Post",
    data: JSON.stringify({
      idFranqueado: idFranqueadoAtual(),
    }),
  })
    .fail(function (e) {
      console.log(e);
      boxErro("Erro ao carregar os clientes");
      if (typeof aoPronto === "function") aoPronto();
    })
    .done(function (r) {
      listaClientes = (r.status != "Vazio" && r.dados) ? r.dados : [];
      $("#id-cliente").val("");
      if (typeof aoPronto === "function") aoPronto();
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
  if (fpPagRelEventos) fpPagRelEventos.destroy();
  $("#formulario").each(function () {
    this.reset();
  });
  $("#lista-clientes").addClass("d-none");
  carregarPadraoUltimos100();
}

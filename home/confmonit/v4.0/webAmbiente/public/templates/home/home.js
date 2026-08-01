var estadoHome = {
  tipo: "",
  master: false,
  precisaFranqueado: false,
  idFranqueado: "",
  nomeFranqueado: "",
  clienteAtual: null,
  mapasCliente: [],
  mapaAtualId: null,
  listaFranqueados: [],
  listaClientes: [],
  listaMapas: [],
  buscaFranqueado: "",
  buscaCliente: "",
  buscaMapa: "",
  telaCheia: false,
  viewAtiva: "mapas",
  timerStatusLista: null,
};

$(window).on("load", function () {
  sincronizarContextoOperador();

  var nome = sessionStorage.getItem("login_userNome") || "Operador";
  var vinculo = sessionStorage.getItem("login_userVinculoNome") || "";
  estadoHome.tipo = sessionStorage.getItem("login_userTipo") || "";
  estadoHome.master = sessionStorage.getItem("login_userMaster") === "S";
  estadoHome.precisaFranqueado =
    estadoHome.tipo === "CEN" || estadoHome.tipo === "REP";

  $("#navUserNome").text(nome);

  var info = vinculo ? vinculo + " (" + estadoHome.tipo + ")" : estadoHome.tipo;
  if (estadoHome.master) info += " · Master";
  $("#navUserVinculo").text(info);

  MapaMonitor.init({
    container: "#mapaContainer",
    imagem: "#mapaImagem",
    camada: "#mapaCamada",
  });

  $("#btnFecharMapa").on("click", fecharMapa);
  $("#btnTelaCheia").on("click", toggleTelaCheia);
  $("#buscaFranqueado").on("input", function () {
    estadoHome.buscaFranqueado = $(this).val().trim().toLowerCase();
    renderGridFranqueados();
  });
  $("#buscaCliente").on("input", function () {
    estadoHome.buscaCliente = $(this).val().trim().toLowerCase();
    renderGridClientes();
  });
  $("#buscaMapa").on("input", function () {
    estadoHome.buscaMapa = $(this).val().trim().toLowerCase();
    renderGridMapas();
  });
  $(".home-nav-aba").on("click", function () {
    trocarViewHome($(this).data("view"));
  });

  if (estadoHome.precisaFranqueado) {
    carregarFranqueados();
  } else {
    estadoHome.idFranqueado =
      sessionStorage.getItem("login_idFranqueadoSelecionado") ||
      sessionStorage.getItem("login_userVinculo") ||
      "";
  }

  var viewParam = "";
  try {
    viewParam = new URLSearchParams(window.location.search).get("view") || "";
  } catch (e) {}
  var viewInicial =
    viewParam || sessionStorage.getItem("home_view") || "mapas";
  trocarViewHome(viewInicial);
});

function trocarViewHome(view) {
  if (view !== "clientes" && view !== "mapas") view = "mapas";
  estadoHome.viewAtiva = view;
  sessionStorage.setItem("home_view", view);

  $(".home-nav-aba").removeClass("home-nav-aba-ativa");
  $('.home-nav-aba[data-view="' + view + '"]').addClass("home-nav-aba-ativa");

  $("#painelClientes").toggleClass("hidden", view !== "clientes");
  $("#painelMapas").toggleClass("hidden", view !== "mapas");

  if (view === "mapas") {
    $("#painelFranqueados").addClass("hidden");
    $("#homePanels").removeClass("layout-central").addClass("layout-franqueado");
    carregarMapas();
  } else {
    pararPollingStatusLista();
    if (estadoHome.precisaFranqueado) {
      $("#painelFranqueados").removeClass("hidden");
      $("#homePanels").removeClass("layout-franqueado").addClass("layout-central");
    }
    carregarClientes();
  }
}

function sincronizarContextoOperador() {
  var ctx = $("#homeContexto");
  if (!ctx.length) return;

  var tipo = ctx.data("tipo") || "";
  var master = ctx.data("master") || "N";
  var nome = ctx.data("nome") || "";
  var vinculo = ctx.data("vinculo") || "";
  var idVinculo = ctx.data("idVinculo") || "";

  if (tipo) sessionStorage.setItem("login_userTipo", tipo);
  if (master) sessionStorage.setItem("login_userMaster", master);
  if (nome) sessionStorage.setItem("login_userNome", nome);
  if (vinculo) sessionStorage.setItem("login_userVinculoNome", vinculo);
  if (idVinculo) sessionStorage.setItem("login_userVinculo", idVinculo);
  if (tipo === "FRA" && idVinculo) {
    sessionStorage.setItem("login_idFranqueadoSelecionado", idVinculo);
  }
}

function carregarFranqueados() {
  msgGrid("#gridFranqueados", "Carregando...");

  $.ajax({
    url: "/home/franqueados",
    method: "POST",
    contentType: "application/json",
    data: "{}",
  })
    .fail(function () {
      msgGrid("#gridFranqueados", "Erro ao carregar franqueados.", true);
    })
    .done(function (r) {
      estadoHome.listaFranqueados = r.dados || [];

      if (r.status === "Vazio" || !estadoHome.listaFranqueados.length) {
        msgGrid(
          "#gridFranqueados",
          estadoHome.master
            ? "Nenhum franqueado encontrado."
            : "Nenhum franqueado com mapa.",
          true
        );
        return;
      }

      renderGridFranqueados();
      restaurarFranqueadoSalvo();
    });
}

function restaurarFranqueadoSalvo() {
  var idSalvo = sessionStorage.getItem("login_idFranqueadoSelecionado");
  if (!idSalvo) return;

  var fra = estadoHome.listaFranqueados.find(function (f) {
    return f.idFranqueado === idSalvo;
  });
  if (fra) {
    selecionarFranqueado(fra.idFranqueado, fra.nomeFranqueado, true);
  }
}

function renderGridFranqueados() {
  var lista = estadoHome.listaFranqueados;
  var termo = estadoHome.buscaFranqueado;
  $("#gridFranqueados").empty();

  if (!lista.length) return;

  var filtrada = lista.filter(function (f) {
    if (!termo) return true;
    return (f.nomeFranqueado || "").toLowerCase().indexOf(termo) >= 0;
  });

  if (!filtrada.length) {
    msgGrid("#gridFranqueados", "Nenhum franqueado encontrado.", true);
    return;
  }

  filtrada.forEach(function (f) {
    var sel =
      f.idFranqueado === estadoHome.idFranqueado ? " home-tile-ativo" : "";
    var tile = $(
      '<button type="button" class="home-tile' +
        sel +
        '" data-id="' +
        escAttr(f.idFranqueado) +
        '">' +
        '<span class="home-tile-icone"><i class="bi bi-shop"></i></span>' +
        '<span class="home-tile-nome">' +
        esc(f.nomeFranqueado) +
        "</span></button>"
    );
    tile.on("click", function () {
      selecionarFranqueado(f.idFranqueado, f.nomeFranqueado);
    });
    $("#gridFranqueados").append(tile);
  });
}

function selecionarFranqueado(id, nome, silencioso) {
  estadoHome.idFranqueado = id;
  estadoHome.nomeFranqueado = nome;
  sessionStorage.setItem("login_idFranqueadoSelecionado", id);

  if (!silencioso) {
    fecharMapa();
    estadoHome.buscaCliente = "";
    $("#buscaCliente").val("");
  }

  renderGridFranqueados();

  $("#franqueadoSelecionado")
    .removeClass("hidden")
    .text(nome);

  if (estadoHome.viewAtiva === "mapas") {
    carregarMapas();
  } else {
    carregarClientes();
  }
}

function carregarClientes() {
  if (estadoHome.precisaFranqueado && !estadoHome.idFranqueado) {
    msgGrid("#gridClientes", "Selecione um franqueado à esquerda.");
    return;
  }

  var payload = {};
  if (estadoHome.precisaFranqueado) {
    payload.idFranqueado = estadoHome.idFranqueado;
  }

  var url = estadoHome.master
    ? "/home/clientesFranqueado"
    : "/home/clientesComMapa";

  msgGrid("#gridClientes", "Carregando clientes...");

  $.ajax({
    url: url,
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(payload),
  })
    .fail(function (e) {
      var msg = "Erro ao carregar clientes.";
      if (e.responseJSON && e.responseJSON.status) {
        msg = e.responseJSON.status.replace(/^Erro:\s*/i, "");
      }
      msgGrid("#gridClientes", msg, true);
    })
    .done(function (r) {
      estadoHome.listaClientes = r.dados || [];

      if (r.status === "Vazio" || !estadoHome.listaClientes.length) {
        msgGrid(
          "#gridClientes",
          estadoHome.master
            ? "Nenhum cliente neste franqueado."
            : "Nenhum cliente com mapa.",
          true
        );
        return;
      }

      renderGridClientes();
    });
}

function renderGridClientes() {
  var lista = estadoHome.listaClientes;
  var termo = estadoHome.buscaCliente;
  $("#gridClientes").empty();

  if (!lista.length) return;

  var filtrada = lista.filter(function (c) {
    if (!termo) return true;
    return (c.nomeCliente || "").toLowerCase().indexOf(termo) >= 0;
  });

  if (!filtrada.length) {
    msgGrid("#gridClientes", "Nenhum cliente encontrado.", true);
    return;
  }

  filtrada.forEach(function (c) {
    renderTileCliente(c);
  });
}

function renderTileCliente(c) {
  var qtd = c.qtdMapas || 0;
  var sel =
    estadoHome.clienteAtual &&
    estadoHome.clienteAtual.idCliente === c.idCliente
      ? " home-tile-ativo"
      : "";
  var badge =
    qtd > 0
      ? '<span class="home-tile-badge">' + qtd + " mapa(s)</span>"
      : '<span class="home-tile-badge home-tile-badge-muted">Sem mapa</span>';

  var tile = $(
    '<button type="button" class="home-tile home-tile-cliente' +
      sel +
      '" data-id="' +
      escAttr(c.idCliente) +
      '">' +
      '<span class="home-tile-icone"><i class="bi bi-person-badge"></i></span>' +
      '<span class="home-tile-nome">' +
      esc(c.nomeCliente) +
      "</span>" +
      badge +
      "</button>"
  );

  tile.on("click", function () {
    selecionarCliente(c);
  });

  $("#gridClientes").append(tile);
}

function selecionarCliente(c) {
  estadoHome.clienteAtual = c;
  renderGridClientes();

  $("#nomeClienteMapa").text(c.nomeCliente || "");
  abrirPainelMapa();
  atualizarLinksCadastro(c);
  carregarMapasCliente(c);
}

function carregarMapas() {
  msgGrid("#gridMapas", "Carregando mapas...");

  $.ajax({
    url: "/home/mapas",
    method: "POST",
    contentType: "application/json",
    data: "{}",
  })
    .fail(function (e) {
      pararPollingStatusLista();
      var msg = "Erro ao carregar mapas.";
      if (e.responseJSON && e.responseJSON.status) {
        msg = e.responseJSON.status.replace(/^Erro:\s*/i, "");
      }
      msgGrid("#gridMapas", msg, true);
    })
    .done(function (r) {
      estadoHome.listaMapas = r.dados || [];

      if (r.status === "Vazio" || !estadoHome.listaMapas.length) {
        pararPollingStatusLista();
        msgGrid("#gridMapas", "Nenhum mapa cadastrado.", true);
        return;
      }

      renderGridMapas();
      iniciarPollingStatusLista();
    });
}

function renderGridMapas() {
  var lista = estadoHome.listaMapas;
  var termo = estadoHome.buscaMapa;
  $("#gridMapas").empty();

  if (!lista.length) return;

  var filtrada = lista.filter(function (m) {
    if (!termo) return true;
    var alvo =
      (m.descricao || "") + " " + (m.nomeCliente || "") + " " + (m.nomeFranqueado || "");
    return alvo.toLowerCase().indexOf(termo) >= 0;
  });

  if (!filtrada.length) {
    msgGrid("#gridMapas", "Nenhum mapa encontrado.", true);
    return;
  }

  filtrada.forEach(function (m) {
    renderTileMapa(m);
  });
}

function renderTileMapa(m) {
  var sel = estadoHome.mapaAtualId === m.id ? " home-tile-ativo" : "";
  var status = m.statusLista || "normal";
  var clsStatus =
    status === "alarme"
      ? " home-tile-alarme"
      : status === "falha"
      ? " home-tile-falha"
      : "";

  var franq =
    estadoHome.precisaFranqueado && m.nomeFranqueado
      ? '<span class="home-tile-franq">' + esc(m.nomeFranqueado) + "</span>"
      : "";

  var tile = $(
    '<button type="button" class="home-tile home-tile-mapa' +
      sel +
      clsStatus +
      '" data-id="' +
      m.id +
      '">' +
      '<span class="home-tile-icone"><i class="bi bi-map"></i></span>' +
      '<span class="home-tile-nome">' +
      esc(mapaNomeAmbiente(m)) +
      "</span>" +
      '<span class="home-tile-badge">' +
      esc(m.nomeCliente) +
      "</span>" +
      franq +
      "</button>"
  );

  tile.on("click", function () {
    selecionarMapa(m);
  });

  $("#gridMapas").append(tile);
}

function selecionarMapa(m) {
  estadoHome.clienteAtual = {
    idCliente: m.idCliente,
    nomeCliente: m.nomeCliente,
  };
  if (m.idFranqueado) estadoHome.idFranqueado = m.idFranqueado;
  estadoHome.mapaAtualId = m.id;

  marcarMapaAtivoNaLista(m.id);

  $("#nomeClienteMapa").text(m.nomeCliente);
  abrirPainelMapa();
  atualizarLinksCadastro(estadoHome.clienteAtual);
  carregarMapasCliente(estadoHome.clienteAtual, m.id);
}

function marcarMapaAtivoNaLista(id) {
  $("#gridMapas .home-tile-mapa").removeClass("home-tile-ativo");
  if (id) {
    $('#gridMapas .home-tile-mapa[data-id="' + id + '"]').addClass("home-tile-ativo");
  }
}

function iniciarPollingStatusLista() {
  pararPollingStatusLista();
  atualizarStatusCardsMapas();
  estadoHome.timerStatusLista = setInterval(atualizarStatusCardsMapas, 3000);
}

function pararPollingStatusLista() {
  if (estadoHome.timerStatusLista) {
    clearInterval(estadoHome.timerStatusLista);
    estadoHome.timerStatusLista = null;
  }
}

function atualizarStatusCardsMapas() {
  var ids = estadoHome.listaMapas.map(function (m) {
    return m.id;
  });
  if (!ids.length) return;

  $.ajax({
    url: "/home/mapasStatus",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ ids: ids }),
  }).done(function (r) {
    var map = (r && r.dados) || r || {};
    estadoHome.listaMapas.forEach(function (m) {
      var st = map[m.id] || map[String(m.id)] || "normal";
      m.statusLista = st;
      var $tile = $('#gridMapas .home-tile-mapa[data-id="' + m.id + '"]');
      $tile.removeClass("home-tile-alarme home-tile-falha");
      if (st === "alarme") {
        $tile.addClass("home-tile-alarme");
      } else if (st === "falha") {
        $tile.addClass("home-tile-falha");
      }
    });
  });
}

function carregarMapasCliente(c, mapaIdPreferido) {
  var payload = { idCliente: c.idCliente };
  if (estadoHome.idFranqueado) {
    payload.idFranqueado = estadoHome.idFranqueado;
  }

  $("#nomeMapaAtivo").text("Carregando mapas...");
  $("#mapaVazio").addClass("hidden");
  $("#mapaContainer").removeClass("hidden");

  $.ajax({
    url: "/mapas/listar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(payload),
  })
    .fail(function () {
      $("#nomeMapaAtivo").text("Erro ao carregar mapas.");
    })
    .done(function (r) {
      var mapas = r.dados || r.items || [];
      estadoHome.mapasCliente = mapas;

      if (r.status === "Vazio" || !mapas.length) {
        MapaMonitor.parar();
        $("#mapaContainer").addClass("hidden");
        $("#mapaVazio").removeClass("hidden");
        $("#tabsMapas").addClass("hidden").empty();
        $("#btnTelaCheia").addClass("hidden");
        $("#linkMonitorar").addClass("hidden");
        $("#linkCentroOperacional").addClass("hidden");
        ocultarLinksMasterMapa();
        $("#nomeMapaAtivo").text("");
        return;
      }

      $("#mapaVazio").addClass("hidden");
      $("#mapaContainer").removeClass("hidden");
      renderTabsMapas(mapas);

      var alvo = null;
      if (mapaIdPreferido) {
        alvo = mapas.find(function (x) {
          return parseInt(x.id || x.mapa_ambiente_id, 10) === parseInt(mapaIdPreferido, 10);
        });
      }
      ativarMapa(alvo || mapas[0]);
      redimensionarMapaHome();
    });
}

function montarQueryMapa(m) {
  var fraQ = estadoHome.idFranqueado
    ? "&idFranqueado=" + encodeURIComponent(estadoHome.idFranqueado)
    : "";
  var c = estadoHome.clienteAtual;
  return (
    "mapa_ambiente_id=" +
    m.id +
    "&idCliente=" +
    c.idCliente +
    "&nome=" +
    encodeURIComponent(mapaNomeAmbiente(m)) +
    "&cliente=" +
    encodeURIComponent(c.nomeCliente) +
    fraQ
  );
}

function atualizarLinksMapa(m) {
  var q = montarQueryMapa(m);
  $("#linkMonitorar").attr("href", "/monitor/page?" + q).removeClass("hidden");
  var c = estadoHome.clienteAtual;
  var coQ =
    "idCliente=" +
    encodeURIComponent(c.idCliente) +
    "&nome=" +
    encodeURIComponent(c.nomeCliente) +
    (estadoHome.idFranqueado
      ? "&idFranqueado=" + encodeURIComponent(estadoHome.idFranqueado)
      : "");
  $("#linkCentroOperacional").attr("href", "/centroOperacional/page?" + coQ).removeClass("hidden");
  $("#btnTelaCheia").removeClass("hidden");

  if (estadoHome.master) {
    var fraQ = estadoHome.idFranqueado
      ? "&idFranqueado=" + encodeURIComponent(estadoHome.idFranqueado)
      : "";
    var c = estadoHome.clienteAtual;
    var urlCadastro =
      "/ambiente/page?idCliente=" +
      c.idCliente +
      "&nome=" +
      encodeURIComponent(c.nomeCliente) +
      fraQ;

    $("#linkCadastroMapa").attr("href", urlCadastro).removeClass("hidden");
    $("#linkSetores").attr("href", "/editor/page?" + q).removeClass("hidden");
    $("#linkEditar").attr("href", "/ambiente/page?" + q).removeClass("hidden");
  } else {
    ocultarLinksMasterMapa();
  }
}

function ocultarLinksMasterMapa() {
  $("#linkCadastroMapa, #linkSetores, #linkEditar").addClass("hidden");
}

function renderTabsMapas(mapas) {
  var $tabs = $("#tabsMapas");
  $tabs.empty();

  if (mapas.length <= 1) {
    $tabs.addClass("hidden");
    return;
  }

  $tabs.removeClass("hidden");
  mapas.forEach(function (m) {
    var mid = m.id || m.mapa_ambiente_id;
    var tab = $(
      '<button type="button" class="home-tab" data-id="' +
        mid +
        '">' +
        esc(mapaNomeAmbiente(m)) +
        "</button>"
    );
    tab.on("click", function (e) {
      e.preventDefault();
      ativarMapa(m);
    });
    $tabs.append(tab);
  });
}

function ativarMapa(m) {
  if (!m) return;
  var mapaId = parseInt(m.id || m.mapa_ambiente_id, 10);
  if (!mapaId) {
    $("#nomeMapaAtivo").text("Mapa inválido.");
    return;
  }

  estadoHome.mapaAtualId = mapaId;
  $("#nomeMapaAtivo").text(mapaNomeAmbiente(m));
  $("#tabsMapas .home-tab").removeClass("home-tab-ativo");
  $('#tabsMapas .home-tab[data-id="' + mapaId + '"]').addClass("home-tab-ativo");
  marcarMapaAtivoNaLista(mapaId);

  atualizarLinksMapa(m);

  var promessa = MapaMonitor.abrir(mapaId);
  if (promessa && promessa.fail) {
    promessa.fail(function () {
      $("#nomeMapaAtivo").text("Erro ao carregar mapa.");
    });
  }

  redimensionarMapaHome();
}

function toggleTelaCheia() {
  if (!$("#painelMapa").is(":visible") || !estadoHome.mapaAtualId) return;

  estadoHome.telaCheia = !estadoHome.telaCheia;
  $("#homeShell").toggleClass("home-tela-cheia", estadoHome.telaCheia);
  $("#btnTelaCheia").text(estadoHome.telaCheia ? "Recolher" : "Tela Cheia");

  setTimeout(function () {
    MapaMonitor.redimensionar();
  }, 50);
}

function sairTelaCheia() {
  if (!estadoHome.telaCheia) return;
  estadoHome.telaCheia = false;
  $("#homeShell").removeClass("home-tela-cheia");
  $("#btnTelaCheia").text("Tela Cheia");
}

function abrirPainelMapa() {
  $("#painelMapa").removeClass("hidden");
  $("#homePanels").addClass("mapa-aberto");
  if (typeof MapaMonitor.reinstalarLayout === "function") {
    MapaMonitor.reinstalarLayout();
  }
  requestAnimationFrame(function () {
    MapaMonitor.redimensionar();
  });
}

function redimensionarMapaHome() {
  [0, 50, 200, 500, 1000].forEach(function (ms) {
    setTimeout(function () {
      MapaMonitor.redimensionar();
    }, ms);
  });
}

function fecharMapa() {
  sairTelaCheia();
  MapaMonitor.parar();
  estadoHome.clienteAtual = null;
  estadoHome.mapasCliente = [];
  estadoHome.mapaAtualId = null;

  $("#painelMapa").addClass("hidden");
  $("#homePanels").removeClass("mapa-aberto");
  renderGridClientes();
  marcarMapaAtivoNaLista(null);
  $("#tabsMapas").addClass("hidden").empty();
  $("#btnTelaCheia").addClass("hidden");
  $("#linkMonitorar").addClass("hidden");
  $("#linkCentroOperacional").addClass("hidden");
  ocultarLinksMasterMapa();
}

function atualizarLinksCadastro(c) {
  var fraQ = estadoHome.idFranqueado
    ? "&idFranqueado=" + encodeURIComponent(estadoHome.idFranqueado)
    : "";
  var url =
    "/ambiente/page?idCliente=" +
    c.idCliente +
    "&nome=" +
    encodeURIComponent(c.nomeCliente) +
    fraQ;

  if (estadoHome.master) {
    $("#linkCadastroVazio").attr("href", url);
  } else {
    $("#linkCadastroVazio").addClass("hidden");
  }
}

function msgGrid(sel, texto, erro) {
  $(sel).html(
    '<p class="home-grid-msg' +
      (erro ? " home-grid-msg-erro" : "") +
      '">' +
      esc(texto) +
      "</p>"
  );
}

function esc(s) {
  return $("<span>").text(s || "").html();
}

function escAttr(s) {
  return String(s || "")
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;");
}

var estadoCO = {
  modoGlobal: false,
  idCliente: "",
  idFranqueado: "",
  nomeCliente: "",
  corAvatar: "#2563eb",
  iniciais: "",

  mapas: [],
  mapaAtivoId: null,
  mapaFixadoListaId: null,
  mapaFixadoAbaId: null,
  ordemDisparoMapas: {},
  alertasRastroMapa: {},
  statusMapas: {},
  monitores: {},
  dispositivosMapa: {},
  mapasCarregados: {},
  contadoresMapa: {},

  processos: [],
  processoAtivo: null,
  processoModal: null,
  processosVistos: {},
  processosOrdemChegada: {},
  processosTrailSync: {},
  processosTimelineSync: {},
  filtroAtivo: "todos",
  rastroAnaliseIdx: 0,
  colMapasView: "mapas",

  setorParaMapa: {},
  dispZonaParaMapa: {},

  somAtivo: true,
  audioAtual: 1,
  progressoTimer: null,
  progressoInicio: 0,

  timers: { status: null, eventos: null, relogio: null, contadores: null },
  mapasProntos: false,
  setoresMapasProntos: false,
  pollingIniciado: false,
};

var CO_CORES = [
  "#2563eb", "#7c3aed", "#db2777", "#dc2626", "#ea580c",
  "#ca8a04", "#16a34a", "#0891b2", "#4b5563", "#1e40af",
];

var GRUPOS_SIRENE = ["ALARME", "EMERGENCIA", "FALHA", "MEDICO", "PANICO"];

// Grupos oficiais do CTI → filtro da fila PROCESSOS (alarmes | controle | falhas | outros → só em Todos)
var GRUPOS_CATEGORIA = {
  ALARME: "alarmes",
  EMERGENCIA: "alarmes",
  MEDICO: "alarmes",
  PANICO: "alarmes",
  ARME: "controle",
  DESARME: "controle",
  RESTAURE: "controle",
  FALHAS: "falhas",
  GERAL: "geral",
  SETUP: "setup",
  TESTE: "teste",
};

$(window).on("load", function () {
  var shell = $("#coShell");
  estadoCO.modoGlobal = shell.attr("data-modo-global") === "S";
  estadoCO.idCliente = shell.data("idCliente") || "";
  estadoCO.nomeCliente = shell.data("nomeCliente") || "";
  estadoCO.idFranqueado = resolverFranqueadoCO(shell);

  if (!estadoCO.modoGlobal && !estadoCO.idCliente) {
    window.location.href = "/home";
    return;
  }

  $("#coNomeCliente").text(estadoCO.nomeCliente || "Cliente");
  if (estadoCO.modoGlobal) {
    $("#coSubtitulo").text("Monitoramento geral — todos os clientes do seu perfil");
    estadoCO.iniciais = extrairIniciais(sessionStorage.getItem("login_userNome") || "CO");
  } else {
    estadoCO.iniciais = extrairIniciais(estadoCO.nomeCliente);
  }
  $("#coAvatarIniciais").text(estadoCO.iniciais);

  if (typeof MapaCameraLive !== "undefined") MapaCameraLive.init();
  initRastroCO();

  initSom();
  initRelogio();
  initHeader();
  initPaleta();
  initToolbar();
  initModal();
  initTabsBar();
  initRastroUI();
  initTimelineUI();
  initNotificacoesCO();

  carregarConfigCliente();
  carregarMapas();
});

function initNotificacoesCO() {
  if (typeof CoNotificacoes === "undefined") return;
  CoNotificacoes.init({
    onAbrir: function (proc) {
      var mapaId = resolverMapaProcesso(proc);
      if (mapaId) focarMapaProcesso(mapaId);
      abrirModalProcesso(proc);
    },
    onFinalizar: function (proc) {
      var mapaId = resolverMapaProcesso(proc);
      if (mapaId) focarMapaProcesso(mapaId);
      abrirModalProcesso(proc);
    },
    onCamera: function (proc) {
      var mapaId = resolverMapaProcesso(proc);
      if (mapaId) focarMapaProcesso(mapaId);
      abrirCameraSetor({
        idSetor: proc.idSetor,
        idDispositivo: proc.idDispositivo,
        idFranqueado: proc.idFranqueado,
        nomeSetor: proc.nomeSetor,
        label: proc.descricaoGrupo || proc.codigo,
        idProcesso: proc.idProcesso,
        mapaId: mapaId,
      });
    },
    onMapa: function (proc) {
      var mapaId = resolverMapaProcesso(proc);
      if (mapaId) {
        focarMapaProcesso(mapaId);
        ativarMapa(mapaId, true);
      }
    },
  });
}

function initRastroCO() {
  var ctx = {
    idCliente: estadoCO.idCliente,
    idFranqueado: estadoCO.idFranqueado,
    idOperador: sessionStorage.getItem("login_userIdOperador") || "",
  };
  if (typeof MapaRastro !== "undefined") {
    MapaRastro.init({
      idCliente: ctx.idCliente,
      idFranqueado: ctx.idFranqueado,
      idOperador: ctx.idOperador,
      onAlerta: function (mapaId, pred) {
        if (!pred || !pred.mensagem) return;
        mostrarAlertaRastro(mapaId, pred.mensagem);
        if (typeof CoTimeline !== "undefined") {
          CoTimeline.registrarRastroAlerta({
            mapaId: mapaId,
            mensagem: pred.mensagem,
            direcao: pred.direcao,
            setor: pred.setor,
            idProcesso: (estadoCO.processoAtivo && estadoCO.processoAtivo.idProcesso) || "",
          });
          renderTimelineMapa(mapaId);
        }
      },
      onChange: function (mapaId, st, meta) {
        if (mapaId === estadoCO.mapaAtivoId) {
          atualizarUiRastro(mapaId, meta);
        }
        sincronizarRastroNaTimeline(mapaId, st);
      },
    });
  }
  if (typeof CoTimeline !== "undefined") {
    CoTimeline.configurarContexto(ctx);
  }
}

function carregarOcorrenciaMapa(mapaId) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;
  var ctx = {
    idCliente: estadoCO.idCliente,
    idFranqueado: estadoCO.idFranqueado,
    idOperador: sessionStorage.getItem("login_userIdOperador") || "",
  };
  if (typeof CoTimeline !== "undefined") {
    CoTimeline.configurarContexto(ctx);
    CoTimeline.carregarServidor(mapaId, function () {
      renderTimelineMapa(mapaId);
    });
  }
  if (typeof MapaRastro !== "undefined") {
    MapaRastro.configurarContexto(ctx);
    MapaRastro.carregarServidor(mapaId, function () {
      MapaRastro.desenhar(mapaId);
      atualizarUiRastro(mapaId);
      var pred = MapaRastro.getPredicao(mapaId) || {};
      if (pred.mensagem && MapaRastro.getTrail(mapaId).length >= 2) {
        mostrarAlertaRastro(mapaId, pred.mensagem);
      }
      atualizarAlertaRastroMapaAtivo();
      sincronizarRastroNaTimeline(mapaId, {
        disparos: MapaRastro.getTrail(mapaId),
        direcao: pred.direcao || "",
        preditoId: pred.preditoId || "",
        ultimoAlerta: pred.mensagem || "",
      });
    });
  }
}

function initRastroUI() {
  $(document).on("click", ".co-alerta-rastro-mapa .co-alerta-rastro-fechar", function (e) {
    e.stopPropagation();
    var mapaId = parseInt($(this).closest(".co-mapa-pane").data("id"), 10);
    ocultarAlertaRastro(mapaId, true);
  });
  $("#coAnalisePlay").on("click", function () {
    if (typeof MapaRastro === "undefined" || !estadoCO.mapaAtivoId) return;
    MapaRastro.reproduzirAnalise(estadoCO.mapaAtivoId);
  });
  $("#coAnalisePrev").on("click", function () {
    if (typeof MapaRastro === "undefined" || !estadoCO.mapaAtivoId) return;
    var idx = Math.max(0, (estadoCO.rastroAnaliseIdx || 0) - 1);
    MapaRastro.irParaPasso(estadoCO.mapaAtivoId, idx);
    estadoCO.rastroAnaliseIdx = idx;
    atualizarUiRastro(estadoCO.mapaAtivoId, { analiseIdx: idx });
  });
  $("#coAnaliseNext").on("click", function () {
    if (typeof MapaRastro === "undefined" || !estadoCO.mapaAtivoId) return;
    var trail = MapaRastro.getTrail(estadoCO.mapaAtivoId);
    var idx = Math.min(trail.length - 1, (estadoCO.rastroAnaliseIdx || 0) + 1);
    MapaRastro.irParaPasso(estadoCO.mapaAtivoId, idx);
    estadoCO.rastroAnaliseIdx = idx;
    atualizarUiRastro(estadoCO.mapaAtivoId, { analiseIdx: idx });
  });
  $("#coAnaliseSair").on("click", sairAnaliseCalor);
  $("#coAnaliseSlider").on("input", function () {
    if (typeof MapaRastro === "undefined" || !estadoCO.mapaAtivoId) return;
    var idx = parseInt($(this).val(), 10) || 0;
    MapaRastro.irParaPasso(estadoCO.mapaAtivoId, idx);
    estadoCO.rastroAnaliseIdx = idx;
    atualizarUiRastro(estadoCO.mapaAtivoId, { analiseIdx: idx });
  });
}

function mapaIdPaneVisivel() {
  var $pane = $(".co-mapa-pane.co-pane-ativo");
  return $pane.length ? parseInt($pane.data("id"), 10) : null;
}

function ensureAlertaRastroMapa(mapaId) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return $();
  ensurePainelDOM(mapaId);
  var $wrap = $("#coPane" + mapaId + " .co-mapa-scroll-wrap");
  var $alerta = $wrap.children(".co-alerta-rastro-mapa");
  if (!$alerta.length) {
    $alerta = $(
      '<div class="co-alerta-rastro co-alerta-rastro-mapa hidden" role="status">' +
        '<i class="bi bi-exclamation-triangle-fill"></i>' +
        '<span class="co-alerta-rastro-texto"></span>' +
        '<button type="button" class="co-alerta-rastro-fechar" title="Fechar">&times;</button>' +
      "</div>"
    );
    $wrap.prepend($alerta);
  }
  return $alerta;
}

function mostrarAlertaRastro(mapaId, msg) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId || !msg) return;
  estadoCO.alertasRastroMapa[mapaId] = msg;
  var $alerta = ensureAlertaRastroMapa(mapaId);
  $alerta.find(".co-alerta-rastro-texto").text(msg);
  if (mapaId === mapaIdPaneVisivel()) {
    $alerta.removeClass("hidden");
  }
}

function ocultarAlertaRastro(mapaId, dispensar) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;
  if (dispensar) delete estadoCO.alertasRastroMapa[mapaId];
  $("#coPane" + mapaId + " .co-alerta-rastro-mapa").addClass("hidden");
}

function atualizarAlertaRastroMapaAtivo() {
  $(".co-alerta-rastro-mapa").addClass("hidden");
  var mapaId = mapaIdPaneVisivel();
  if (!mapaId) return;
  var msg = estadoCO.alertasRastroMapa[mapaId];
  if (!msg && typeof MapaRastro !== "undefined") {
    var pred = MapaRastro.getPredicao(mapaId);
    if (pred && pred.mensagem && MapaRastro.getTrail(mapaId).length >= 2) {
      msg = pred.mensagem;
      estadoCO.alertasRastroMapa[mapaId] = msg;
    }
  }
  if (!msg) return;
  var $alerta = ensureAlertaRastroMapa(mapaId);
  $alerta.find(".co-alerta-rastro-texto").text(msg);
  $alerta.removeClass("hidden");
}

function sincronizarRastroNaTimeline(mapaId, st) {
  if (typeof CoTimeline === "undefined" || !mapaId) return;
  var disparos = (st && st.disparos) || (typeof MapaRastro !== "undefined" ? MapaRastro.getTrail(mapaId) : []);
  if (!disparos || !disparos.length) return;
  disparos.forEach(function (d) {
    if (!d || !d.idSetor) return;
    CoTimeline.registrarRastroDisparo({
      mapaId: mapaId,
      idSetor: d.idSetor,
      label: d.label || d.idSetor,
      idProcesso: d.idProcesso || "",
      ts: d.ts,
    });
  });
  if (st && st.ultimoAlerta) {
    CoTimeline.registrarRastroAlerta({
      mapaId: mapaId,
      mensagem: st.ultimoAlerta,
      direcao: st.direcao || "",
      idSetor: st.preditoId || "",
      proximaZona: st.preditoId || "",
      idProcesso: st.processoId || "",
    });
  }
  if (mapaId === estadoCO.mapaAtivoId || estadoCO.colMapasView === "timeline") {
    renderTimelineMapa(mapaId);
  }
}

function atualizarUiRastro(mapaId, meta) {
  if (typeof MapaRastro === "undefined") return;
  mapaId = parseInt(mapaId || estadoCO.mapaAtivoId, 10);
  var trail = mapaId ? MapaRastro.getTrail(mapaId) : [];
  var tem = trail.length > 0;
  var $item = mapaId ? $('.co-item-mapa[data-id="' + mapaId + '"]') : $();
  $item.find(".co-item-btn-analise").prop("disabled", !tem);
  $item.find(".co-item-btn-limpar").prop("disabled", !tem);

  var $slider = $("#coAnaliseSlider");
  $slider.attr("max", Math.max(0, trail.length - 1));
  if (meta && meta.analiseIdx != null) {
    estadoCO.rastroAnaliseIdx = meta.analiseIdx;
    $slider.val(meta.analiseIdx);
  }

  var pred = mapaId ? MapaRastro.getPredicao(mapaId) : null;
  if (MapaRastro.isModoAnalise() && trail.length) {
    var idx = estadoCO.rastroAnaliseIdx || 0;
    var d = trail[idx];
    var txt = d
      ? "Passo " + (idx + 1) + "/" + trail.length + " · " + (d.label || d.idSetor)
      : "Trajetória";
    $("#coAnaliseStatus").text(txt);
    $item.find(".co-item-btn-analise").addClass("co-analise-ativo");
  } else if (pred && pred.direcao) {
    $("#coAnaliseStatus").text("Direção " + pred.direcao + " · " + trail.length + " disparo(s)");
    $item.find(".co-item-btn-analise").removeClass("co-analise-ativo");
  } else {
    $("#coAnaliseStatus").text(tem ? trail.length + " disparo(s) no rastro" : "Sem trajetória");
    $item.find(".co-item-btn-analise").removeClass("co-analise-ativo");
  }
}

function toggleAnaliseCalor() {
  if (typeof MapaRastro === "undefined" || !estadoCO.mapaAtivoId) return;
  if (MapaRastro.isModoAnalise()) {
    sairAnaliseCalor();
    return;
  }
  var trail = MapaRastro.getTrail(estadoCO.mapaAtivoId);
  if (!trail.length) return;
  MapaRastro.setModoAnalise(true);
  estadoCO.rastroAnaliseIdx = trail.length - 1;
  $("#coPainelAnalise").removeClass("hidden");
  $('.co-item-mapa[data-id="' + estadoCO.mapaAtivoId + '"] .co-item-btn-analise').addClass("co-analise-ativo");
  MapaRastro.irParaPasso(estadoCO.mapaAtivoId, trail.length - 1);
  atualizarUiRastro(estadoCO.mapaAtivoId, { analiseIdx: trail.length - 1 });
}

function sairAnaliseCalor() {
  if (typeof MapaRastro === "undefined") return;
  MapaRastro.pararAnalise();
  MapaRastro.setModoAnalise(false);
  $("#coPainelAnalise").addClass("hidden");
  $(".co-item-btn-analise").removeClass("co-analise-ativo");
  if (estadoCO.mapaAtivoId) {
    MapaRastro.desenhar(estadoCO.mapaAtivoId);
    atualizarUiRastro(estadoCO.mapaAtivoId);
  }
}

function sincronizarRastroProcesso(proc) {
  if (typeof MapaRastro === "undefined" || !proc || !proc.idProcesso) return;
  var mapaId = resolverMapaProcesso(proc);
  if (!mapaId) return;

  var chave = proc.idProcesso + ":" + (proc.quantidade || 0) + ":" + (proc.idSetor || "");
  if (estadoCO.processosTrailSync[chave]) return;
  estadoCO.processosTrailSync[chave] = true;

  if (proc.idSetor) {
    var mon = estadoCO.monitores[mapaId];
    var setores = mon && typeof mon.getSetores === "function" ? mon.getSetores() : [];
    MapaRastro.registrarDisparo({
      mapaId: mapaId,
      idSetor: proc.idSetor,
      label: proc.nomeSetor || proc.idSetor,
      idProcesso: proc.idProcesso,
      setores: setores,
    });
  }

  $.ajax({
    url: "/centroOperacional/processoDetalhe",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      idProcesso: proc.idProcesso,
      idFranqueado: proc.idFranqueado || estadoCO.idFranqueado,
    }),
  }).done(function (r) {
    var lista = r.status === "Vazio" ? [] : r.dados || [];
    if (!lista.length) return;
    ensurePainelMapa(mapaId);
    var mon2 = estadoCO.monitores[mapaId];
    function aplicar() {
      var setores2 = mon2 && typeof mon2.getSetores === "function" ? mon2.getSetores() : [];
      if (!setores2.length && mapaId === estadoCO.mapaAtivoId) {
        setTimeout(aplicar, 400);
        return;
      }
      MapaRastro.sincronizarProcesso(mapaId, proc.idProcesso, lista, setores2);
      if (mapaId === estadoCO.mapaAtivoId) atualizarUiRastro(mapaId);
      sincronizarTimelineProcesso(mapaId, proc, lista);
    }
    if (estadoCO.mapasCarregados[mapaId]) aplicar();
    else {
      carregarMapaSeNecessario(mapaId);
      setTimeout(aplicar, 600);
    }
  });
}

function initTimelineUI() {
  $("#coViewMapasLista, #coViewTimeline").on("click", function () {
    var view = $(this).data("view") || "mapas";
    setColMapasView(view);
  });
  $("#coTimelineCopiar").on("click", function () {
    if (typeof CoTimeline === "undefined" || !estadoCO.mapaAtivoId) return;
    var txt = CoTimeline.exportarTexto(estadoCO.mapaAtivoId);
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(txt).then(function () {
        if (typeof boxOk === "function") boxOk("Linha do tempo copiada.");
      });
    } else {
      window.prompt("Copie a linha do tempo:", txt);
    }
  });
  $("#coTimelineExportar").on("click", function () {
    if (typeof CoTimeline === "undefined" || !estadoCO.mapaAtivoId) return;
    var txt = CoTimeline.exportarTexto(estadoCO.mapaAtivoId);
    var blob = new Blob([txt], { type: "text/plain;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = "linha-do-tempo-ocorrencia-" + estadoCO.mapaAtivoId + ".txt";
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  });
}

function setColMapasView(view) {
  estadoCO.colMapasView = view === "timeline" ? "timeline" : "mapas";
  $(".co-col-mapas-view-tab").removeClass("co-col-view-ativo");
  $('.co-col-mapas-view-tab[data-view="' + estadoCO.colMapasView + '"]').addClass("co-col-view-ativo");
  $("#coListaMapas").toggleClass("hidden", estadoCO.colMapasView !== "mapas");
  $("#coListaMapasFixados").toggleClass("hidden", estadoCO.colMapasView !== "mapas");
  $("#coPainelTimeline").toggleClass("hidden", estadoCO.colMapasView !== "timeline");
  if (estadoCO.colMapasView === "timeline") renderTimelineMapa(estadoCO.mapaAtivoId);
}

function sincronizarTimelineProcesso(mapaId, proc, eventos) {
  if (typeof CoTimeline === "undefined" || !mapaId) return;
  var pred =
    typeof MapaRastro !== "undefined" ? MapaRastro.getPredicao(mapaId) : null;
  var ctx = {
    deslocamento: !!(pred && pred.direcao),
    direcao: (pred && pred.direcao) || "",
  };
  if (eventos && eventos.length) {
    CoTimeline.sincronizarEventosProcesso(mapaId, proc.idProcesso, eventos, ctx);
  } else {
    CoTimeline.sincronizarProcessoResumo(mapaId, proc, ctx);
  }
  if (mapaId === estadoCO.mapaAtivoId) renderTimelineMapa(mapaId);
}

function carregarTimelineProcesso(proc) {
  if (typeof CoTimeline === "undefined" || !proc || !proc.idProcesso) return;
  var mapaId = resolverMapaProcesso(proc);
  if (!mapaId) return;

  var chave = "tl:" + proc.idProcesso + ":" + (proc.quantidade || 0) + ":" + (proc.dataHora || "");
  if (estadoCO.processosTimelineSync[chave]) {
    CoTimeline.sincronizarProcessoResumo(mapaId, proc, {});
    if (mapaId === estadoCO.mapaAtivoId) renderTimelineMapa(mapaId);
    return;
  }
  estadoCO.processosTimelineSync[chave] = true;

  CoTimeline.sincronizarProcessoResumo(mapaId, proc, {});
  $.ajax({
    url: "/centroOperacional/processoDetalhe",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      idProcesso: proc.idProcesso,
      idFranqueado: proc.idFranqueado || estadoCO.idFranqueado,
    }),
  }).done(function (r) {
    var lista = r.status === "Vazio" ? [] : r.dados || [];
    sincronizarTimelineProcesso(mapaId, proc, lista);
  });
}

function renderTimelineMapa(mapaId) {
  mapaId = parseInt(mapaId || estadoCO.mapaAtivoId, 10);
  var $lista = $("#coTimelineLista");
  var $badge = $("#coTimelineBadge");
  if (typeof CoTimeline === "undefined" || !mapaId) {
    $lista.html('<p class="co-msg">Selecione um mapa para ver a linha do tempo.</p>');
    $badge.addClass("hidden").text("0");
    return;
  }
  var itens = CoTimeline.obter(mapaId);
  if (!itens.length) {
    $lista.html('<p class="co-msg">Nenhum evento registrado nesta sessão para o mapa ativo.</p>');
    $badge.addClass("hidden").text("0");
    return;
  }
  $badge.removeClass("hidden").text(String(itens.length));
  $lista.empty();
  itens.forEach(function (it) {
    var meta = [];
    if (it.origem === "contact_id" && it.codigo) meta.push("CID " + it.codigo);
    if (it.origem === "rastro") meta.push("Rastro");
    if (it.direcao) meta.push(it.direcao);
    if (it.zonaUser) meta.push("Z-" + it.zonaUser);
    if (it.operador) meta.push(it.operador);
    if (it.idProcesso) meta.push("Proc " + String(it.idProcesso).slice(-6));
    var el = $(
      '<div class="co-timeline-item co-tl-' +
        escAttr(it.tipo || "info") +
        '">' +
        '<div class="co-timeline-hora">' +
        esc(it.hora) +
        "</div>" +
        '<div class="co-timeline-texto">' +
        esc(it.texto) +
        (meta.length
          ? '<span class="co-timeline-meta">' + esc(meta.join(" · ")) + "</span>"
          : "") +
        "</div></div>"
    );
    $lista.append(el);
  });
  if (estadoCO.colMapasView === "timeline") {
    $lista.scrollTop($lista[0].scrollHeight);
  }
}

function nomeOperadorCO() {
  return (
    sessionStorage.getItem("login_userNick") ||
    sessionStorage.getItem("login_userNome") ||
    "Operador"
  );
}

function resolverFranqueadoCO(shell) {
  if (estadoCO.modoGlobal) return "";
  var fromUrl = String(shell.data("idFranqueado") || "").trim();
  if (fromUrl) return fromUrl;

  var params = new URLSearchParams(window.location.search);
  var q = String(params.get("idFranqueado") || "").trim();
  if (q) return q;

  var tipo = sessionStorage.getItem("login_userTipo") || "";
  if (tipo === "FRA") {
    return sessionStorage.getItem("login_userVinculo") || "";
  }

  // CEN/REP: deixa vazio — o servidor resolve pelo idCliente (evita franqueado errado no sessionStorage)
  return "";
}

function mapaIdFrom(m) {
  return parseInt(m.id || m.mapa_ambiente_id || m.Id, 10);
}

function normalizarListaMapas(mapas) {
  return (mapas || [])
    .map(function (m) {
      var id = mapaIdFrom(m);
      if (!id) return null;
      m.id = id;
      return m;
    })
    .filter(Boolean);
}

function initTabsBar() {
  initHorizontalScrollBars();
}

function initHorizontalScrollBars() {
  $(".co-hscroll-bar").each(function () {
    var $bar = $(this);
    if ($bar.data("coHscroll")) return;
    $bar.data("coHscroll", true);

    var $vp = $bar.find(".co-hscroll-viewport").first();
    var $prev = $bar.find(".co-hscroll-prev").first();
    var $next = $bar.find(".co-hscroll-next").first();
    if (!$vp.length || !$prev.length || !$next.length) return;

    function atualizar() {
      atualizarHscrollBotoes($vp[0], $prev, $next);
    }

    $vp.on("scroll.coHscroll", atualizar);
    $prev.on("click", function () {
      if (this.disabled) return;
      $vp.scrollLeft(function (_, v) { return Math.max(0, v - 160); });
    });
    $next.on("click", function () {
      if (this.disabled) return;
      $vp.scrollLeft(function (_, v) { return v + 160; });
    });

    atualizar();
    $bar.data("coHscrollAtualizar", atualizar);
  });

  $(window).off("resize.coHscroll").on("resize.coHscroll", function () {
    $(".co-hscroll-bar").each(function () {
      var fn = $(this).data("coHscrollAtualizar");
      if (typeof fn === "function") fn();
    });
  });
}

function atualizarHscrollBotoes(el, $prev, $next) {
  if (!el || !$prev.length || !$next.length) return;

  var temOverflow = el.scrollWidth > el.clientWidth + 2;
  if (!temOverflow) {
    $prev.addClass("hidden").prop("disabled", true);
    $next.addClass("hidden").prop("disabled", true);
    return;
  }

  $prev.removeClass("hidden").prop("disabled", el.scrollLeft <= 2);
  $next.removeClass("hidden").prop("disabled", el.scrollLeft + el.clientWidth >= el.scrollWidth - 2);
}

function atualizarBotoesTabsScroll() {
  var fn = $(".co-tabs-bar").data("coHscrollAtualizar");
  if (typeof fn === "function") fn();
}

function scrollTabAtivoParaVista() {
  var $tab = $(".co-tab.co-tab-ativo");
  if (!$tab.length) return;
  scrollTabParaVista(parseInt($tab.data("id"), 10));
}

function scrollTabParaVista(mapaId) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;
  var $tab = $('.co-tab[data-id="' + mapaId + '"]');
  if (!$tab.length) return;
  if ($tab.closest("#coTabsFixadas").length) {
    requestAnimationFrame(atualizarBotoesTabsScroll);
    return;
  }
  $tab[0].scrollIntoView({ behavior: "smooth", block: "nearest", inline: "nearest" });
  setTimeout(atualizarBotoesTabsScroll, 300);
}

function initToolbar() {
  // botões Armar/Evento/Análise ficam dentro de cada card da lista
}

function initModal() {
  $(".co-modal-fechar").on("click", fecharModalProcesso);
  $("#coModalProcesso").on("click", function (e) {
    if ($(e.target).is("#coModalProcesso")) fecharModalProcesso();
  });
  $("#coBtnFinalizar").on("click", finalizarProcessoAtual);
}

function initSom() {
  var off = sessionStorage.getItem("co_somAtivo") === "OFF";
  estadoCO.somAtivo = !off;
  atualizarUiSom();

  $("#coBtnSirene").on("click", function () {
    estadoCO.somAtivo = !estadoCO.somAtivo;
    sessionStorage.setItem("co_somAtivo", estadoCO.somAtivo ? "ON" : "OFF");
    atualizarUiSom();
    if (estadoCO.somAtivo) tocarSom(true);
    else pararSom();
  });

  $("#coAlertaSomAtivar").on("click", function () {
    estadoCO.somAtivo = true;
    sessionStorage.setItem("co_somAtivo", "ON");
    atualizarUiSom();
    tocarSom(true);
  });
}

function atualizarUiSom() {
  var off = !estadoCO.somAtivo;
  $("#coBtnSirene").toggleClass("co-som-off", off);
  $("#coAlertaSomOff").toggleClass("hidden", !off);
  $("#coBtnSirene i").attr("class", off ? "bi bi-volume-mute-fill" : "bi bi-volume-up-fill");
}

function tocarSom(teste) {
  if (!estadoCO.somAtivo && !teste) return;
  estadoCO.audioAtual = estadoCO.audioAtual === 1 ? 2 : 1;
  var outro = estadoCO.audioAtual === 1 ? 2 : 1;
  document.getElementById("coAudio" + outro).pause();
  var a = document.getElementById("coAudio" + estadoCO.audioAtual);
  if (!a) return;
  a.currentTime = 0;
  a.play().catch(function () {});
}

function pararSom() {
  document.getElementById("coAudio1").pause();
  document.getElementById("coAudio2").pause();
}

function initRelogio() {
  function tick() {
    var agora = new Date();
    $("#coRelogioHora").text(
      agora.toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit", second: "2-digit" })
    );
    var data = agora.toLocaleDateString("pt-BR", {
      weekday: "long", day: "numeric", month: "long", year: "numeric",
    });
    $("#coRelogioData").text(data.charAt(0).toUpperCase() + data.slice(1));
  }
  tick();
  estadoCO.timers.relogio = setInterval(tick, 1000);
}

function initHeader() {
  $("#coBtnExportar").on("click", exportarProcessos);
  $("#coAvatar").on("click", function (e) {
    e.stopPropagation();
    var pop = $("#coAvatarPopover");
    var rect = this.getBoundingClientRect();
    pop.css({ top: rect.bottom + 6, left: rect.left });
    pop.toggleClass("hidden");
  });
  $(document).on("click", function () { $("#coAvatarPopover").addClass("hidden"); });
  $("#coAvatarPopover").on("click", function (e) { e.stopPropagation(); });
  $("#coCorPicker").on("input", function () {
    aplicarCorAvatar($(this).val());
    salvarConfigCliente();
  });
  $(".co-filtro").on("click", function () {
    $(".co-filtro").removeClass("co-filtro-ativo");
    $(this).addClass("co-filtro-ativo");
    estadoCO.filtroAtivo = $(this).data("filtro");
    renderFilaProcessos();
  });
}

function initPaleta() {
  var $p = $("#coPaleta");
  CO_CORES.forEach(function (cor) {
    var btn = $('<button type="button"></button>').css("background", cor);
    btn.on("click", function () {
      aplicarCorAvatar(cor);
      $("#coCorPicker").val(cor);
      salvarConfigCliente();
    });
    $p.append(btn);
  });
}

function aplicarCorAvatar(cor) {
  estadoCO.corAvatar = cor || "#2563eb";
  $("#coAvatar").css("background", estadoCO.corAvatar);
}

function extrairIniciais(nome) {
  var partes = (nome || "").trim().split(/\s+/).filter(Boolean);
  if (!partes.length) return "?";
  if (partes.length === 1) return partes[0].slice(0, 2).toUpperCase();
  return (partes[0][0] + partes[1][0]).toUpperCase();
}

function carregarConfigCliente() {
  if (estadoCO.modoGlobal) {
    aplicarCorAvatar(estadoCO.corAvatar);
    return;
  }
  $.ajax({
    url: "/centroOperacional/configCliente",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ idCliente: estadoCO.idCliente, idFranqueado: estadoCO.idFranqueado }),
  }).done(function (r) {
    if (r.status === "OK" && r.dados) {
      if (r.dados.corAvatar) aplicarCorAvatar(r.dados.corAvatar);
      if (r.dados.iniciais) {
        estadoCO.iniciais = r.dados.iniciais;
        $("#coAvatarIniciais").text(estadoCO.iniciais);
      }
    } else aplicarCorAvatar(estadoCO.corAvatar);
  });
}

function salvarConfigCliente() {
  $.ajax({
    url: "/centroOperacional/configCliente/salvar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      idCliente: estadoCO.idCliente,
      idFranqueado: estadoCO.idFranqueado,
      corAvatar: estadoCO.corAvatar,
      iniciais: estadoCO.iniciais,
    }),
  });
}

function carregarMapas() {
  $("#coListaMapas").html('<p class="co-msg">Carregando mapas...</p>');
  $("#coMapaVazio").removeClass("hidden").text("Carregando mapas...");

  var url = estadoCO.modoGlobal ? "/centroOperacional/mapas" : "/mapas/listar";
  var payload = estadoCO.modoGlobal
    ? {}
    : { idCliente: estadoCO.idCliente, idFranqueado: estadoCO.idFranqueado };

  $.ajax({
    url: url,
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(payload),
  })
    .fail(function (xhr) {
      var msg = (xhr.responseJSON && xhr.responseJSON.status) || "Erro ao carregar mapas.";
      $("#coListaMapas").html('<p class="co-msg">' + esc(String(msg).replace(/^Erro:\s*/i, "")) + "</p>");
      $("#coMapaVazio").removeClass("hidden").text("Não foi possível carregar os mapas.");
    })
    .done(function (r) {
      var mapas = normalizarListaMapas(r.dados || r.items || []);
      if (r.status === "Vazio" || !mapas.length) {
        var msgVazio = estadoCO.modoGlobal
          ? "Nenhum mapa cadastrado no seu perfil."
          : "Nenhum mapa cadastrado para este cliente.";
        $("#coListaMapas").html('<p class="co-msg">' + msgVazio + "</p>");
        $("#coMapaVazio").removeClass("hidden").text(msgVazio);
        return;
      }
      estadoCO.mapas = mapas;
      estadoCO.mapasProntos = false;
      estadoCO.setoresMapasProntos = false;
      montarMapasUI();
      estadoCO.mapasProntos = true;
      ativarMapa(mapas[0].id, true);
      carregarSetoresMapas(function () {
        if (!estadoCO.pollingIniciado) {
          iniciarPolling();
          estadoCO.pollingIniciado = true;
        }
      });
      $(window).off("resize.coMapa").on("resize.coMapa", function () {
        if (estadoCO.mapaAtivoId) redimensionarMapaAtivo(estadoCO.mapaAtivoId);
      });
    });
}

function montarMapasUI() {
  $("#coListaMapas").empty();
  $("#coListaMapasFixados").empty();
  $("#coTabs").empty();
  $("#coTabsFixadas").empty();
  $("#coPainelMapas").children(".co-mapa-pane").remove();
  estadoCO.monitores = {};
  estadoCO.mapasCarregados = {};

  estadoCO.mapas.forEach(function (m) {
    var id = mapaIdFrom(m);
    if (!id) return;
    var nome = mapaNomeAmbiente(m);
    var clienteBadge = "";
    if (estadoCO.modoGlobal && m.nomeCliente) {
      clienteBadge =
        '<span class="co-item-mapa-cliente">' + esc(m.nomeCliente) + "</span>";
    }

    var item = $(
      '<div class="co-item-mapa" data-id="' + id + '" role="button" tabindex="0">' +
        '<div class="co-item-mapa-topo">' +
          '<div class="co-item-mapa-titulo">' +
            '<span class="co-item-mapa-nome">' + esc(nome) + "</span>" +
            '<span class="co-item-mapa-keep co-keep co-keep-offline">offline</span>' +
          "</div>" +
          '<button type="button" class="co-fixar-lista" title="Fixar na lista"><i class="bi bi-pin-angle"></i></button>' +
        "</div>" +
        clienteBadge +
        '<span class="co-item-mapa-status st-normal">Aguardando...</span>' +
        '<div class="co-item-mapa-contadores">' +
          "<span>Zonas <b class=\"co-item-cnt-zonas\">0</b></span>" +
          "<span>Alarmes <b class=\"co-item-cnt-alarmes\">0</b></span>" +
          "<span>Sem com. <b class=\"co-item-cnt-semcom\">0</b></span>" +
        "</div>" +
        '<div class="co-item-mapa-progress hidden">' +
          '<span class="co-progress-label">Atendimento</span>' +
          '<div class="co-progress-track"><div class="co-progress-fill"></div></div>' +
        "</div>" +
        '<div class="co-item-mapa-acoes">' +
          '<button type="button" class="co-btn-armar-toolbar co-item-btn-armar desarmado">Armar</button>' +
          '<button type="button" class="co-btn-evento co-item-btn-evento" disabled><i class="bi bi-bell"></i> Evento</button>' +
          '<button type="button" class="co-btn-analise co-item-btn-analise" title="Análise de calor" disabled><i class="bi bi-thermometer-half"></i> Análise</button>' +
          '<button type="button" class="co-btn-limpar-rastro co-item-btn-limpar" title="Limpar rastro" disabled><i class="bi bi-eraser"></i></button>' +
        "</div>" +
      "</div>"
    );

    item.on("click", function (e) {
      if ($(e.target).closest("button").length) return;
      ativarMapa(id, true);
    });
    item.on("keydown", function (e) {
      if (e.key === "Enter" || e.key === " ") { e.preventDefault(); ativarMapa(id, true); }
    });
    item.find(".co-fixar-lista").on("click", function (e) {
      e.stopPropagation();
      toggleFixarLista(id);
    });
    item.find(".co-item-btn-armar").on("click", function (e) {
      e.stopPropagation();
      ativarMapa(id, true);
      comandoArmarMapa(id);
    });
    item.find(".co-item-btn-evento").on("click", function (e) {
      e.stopPropagation();
      var proc = processoDoMapa(id);
      if (!proc) return;
      ativarMapa(id, true);
      abrirModalProcesso(proc);
    });
    item.find(".co-item-btn-analise").on("click", function (e) {
      e.stopPropagation();
      if ($(this).prop("disabled")) return;
      ativarMapa(id, true);
      toggleAnaliseCalor();
    });
    item.find(".co-item-btn-limpar").on("click", function (e) {
      e.stopPropagation();
      if ($(this).prop("disabled")) return;
      ativarMapa(id, true);
      if (typeof MapaRastro === "undefined") return;
      MapaRastro.limparMapa(id);
      ocultarAlertaRastro(id, true);
      atualizarUiRastro(id);
    });
    $("#coListaMapas").append(item);

    if (estadoCO.contadoresMapa[id]) {
      aplicarContadoresNoCard(id, estadoCO.contadoresMapa[id]);
    } else {
      atualizarContadoresMapa(id);
    }

    var tabLabel = nome;
    if (estadoCO.modoGlobal && m.nomeCliente) {
      tabLabel = m.nomeCliente + " · " + nome;
    }
    var tab = criarAbaMapa(id, tabLabel);
    $("#coTabs").append(tab);
  });
  atualizarFixarUI();
  requestAnimationFrame(atualizarBotoesTabsScroll);
}

function criarAbaMapa(id, tabLabel) {
  var tab = $(
    '<div class="co-tab" data-id="' + id + '">' +
      '<button type="button" class="co-tab-fixar" title="Fixar aba"><i class="bi bi-pin-angle"></i></button>' +
      '<span class="co-tab-label">' + esc(tabLabel) + "</span></div>"
  );
  tab.on("click", function (e) {
    if ($(e.target).closest(".co-tab-fixar").length) {
      e.stopPropagation();
      toggleFixarAba(id);
      return;
    }
    ativarMapa(id, true);
  });
  tab.find(".co-tab-fixar").on("click", function (e) {
    e.stopPropagation();
    toggleFixarAba(id);
  });
  return tab;
}

function posicionarAbaFixada() {
  var fixId = estadoCO.mapaFixadoAbaId;
  var $fixadas = $("#coTabsFixadas");
  var $rolaveis = $("#coTabs");

  $fixadas.children(".co-tab").each(function () {
    var $t = $(this);
    var tid = parseInt($t.data("id"), 10);
    if (!fixId || tid !== fixId) {
      $rolaveis.prepend($t.removeClass("co-tab-fixado"));
    }
  });

  if (!fixId) {
    requestAnimationFrame(atualizarBotoesTabsScroll);
    return;
  }

  var $tab = $('.co-tab[data-id="' + fixId + '"]');
  if (!$tab.length) {
    requestAnimationFrame(atualizarBotoesTabsScroll);
    return;
  }
  $tab.addClass("co-tab-fixado");
  $fixadas.append($tab);
  requestAnimationFrame(atualizarBotoesTabsScroll);
}

function ensurePainelDOM(id) {
  id = parseInt(id, 10);
  if (!id) return null;
  var paneId = "coPane" + id;
  if ($("#" + paneId).length) return paneId;

  var pane = $(
    '<div class="co-mapa-pane" id="' + paneId + '" data-id="' + id + '">' +
      '<div class="co-mapa-scroll-wrap">' +
      '<div class="co-alerta-rastro co-alerta-rastro-mapa hidden" role="status">' +
      '<i class="bi bi-exclamation-triangle-fill"></i>' +
      '<span class="co-alerta-rastro-texto"></span>' +
      '<button type="button" class="co-alerta-rastro-fechar" title="Fechar">&times;</button>' +
      "</div>" +
      '<div class="co-mapa-scroll" data-id="' + id + '">' +
      '<div class="mapa-container co-mapa-box monitor-page">' +
      '<img class="mapa-imagem co-img-' + id + '" alt="Planta" />' +
      '<div class="mapa-camada co-camada-' + id + '"></div></div></div>' +
      "</div></div>"
  );
  $("#coPainelMapas").append(pane);
  return paneId;
}

function ensurePainelMapa(id) {
  id = parseInt(id, 10);
  if (!id) return null;
  var paneId = ensurePainelDOM(id);
  if (!paneId) return null;
  if (estadoCO.monitores[id]) return estadoCO.monitores[id];

  estadoCO.monitores[id] = MapaMonitor.criar({
    container: "#" + paneId + " .mapa-container",
    imagem: "#" + paneId + " .mapa-imagem",
    camada: "#" + paneId + " .mapa-camada",
    popup: true,
    abrirCameraAutomatica: false,
    resolverProcesso: resolverProcessoDoSetor,
    onFinalizarProcesso: function (proc) {
      abrirModalProcesso(proc);
    },
    onAbrirCamera: function (dados) {
      abrirCameraSetor({
        idSetor: dados && dados.idSetor,
        idDispositivo: dados && dados.idDispositivo,
        idFranqueado: dados && dados.idFranqueado,
        setorNome: (dados && (dados.setorNome || dados.label)) || "",
        label: (dados && dados.label) || "",
        idProcesso: (estadoCO.processoAtivo && estadoCO.processoAtivo.idProcesso) || "",
        mapaId: id,
      });
    },
    onStatus: function (map, setoresStatus) {
      var temAlarme = false;
      Object.keys(map || {}).forEach(function (k) {
        if (map[k] === "alarme") temAlarme = true;
      });

      if (!mapaTemProcessoAberto(id)) {
        estadoCO.statusMapas[id] = "normal";
        $("#coPane" + id + " .mapa-ponto-monitor.mapa-ponto-alarme")
          .removeClass(
            "mapa-ponto-alarme mapa-ponto-rastro-antigo mapa-ponto-rastro-atual mapa-ponto-rastro-predito"
          )
          .addClass("mapa-ponto-normal");
        if (typeof MapaRastro !== "undefined" && MapaRastro.getTrail(id).length) {
          MapaRastro.limparMapa(id);
          if (id === estadoCO.mapaAtivoId) MapaRastro.desenhar(id);
          ocultarAlertaRastro(id, true);
          atualizarUiRastro(id);
        }
        atualizarStatusLista();
        return;
      }

      if (temAlarme) estadoCO.statusMapas[id] = "alarme";
      atualizarStatusLista();
      if (typeof MapaRastro !== "undefined") {
        var mon = estadoCO.monitores[id];
        var setores = mon && typeof mon.getSetores === "function" ? mon.getSetores() : [];
        MapaRastro.observarStatus(id, map, setores);
        if (id === estadoCO.mapaAtivoId) atualizarUiRastro(id);
      }
    },
    onCarregado: function (mapa, setores) {
      estadoCO.mapasCarregados[id] = true;
      if (setores && setores.length && setores[0].idDispositivo) {
        estadoCO.dispositivosMapa[id] = setores[0].idDispositivo;
      }
      redimensionarMapaAtivo(id);
      if (id === estadoCO.mapaAtivoId) {
        atualizarContadoresMapa(id);
        if (typeof MapaRastro !== "undefined") {
          MapaRastro.desenhar(id);
          atualizarUiRastro(id);
        }
      }
    },
    onPontosRenderizados: function () {
      if (typeof MapaRastro !== "undefined") MapaRastro.desenhar(id);
    },
  });

  return estadoCO.monitores[id];
}

function carregarSetoresMapas(cb) {
  var ids = estadoCO.mapas.map(function (m) { return mapaIdFrom(m); }).filter(Boolean);
  $.ajax({
    url: "/centroOperacional/setoresMapas",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ ids: ids }),
  })
    .done(function (r) {
      var dados = (r && r.dados) || {};
      estadoCO.setorParaMapa = {};
      estadoCO.dispZonaParaMapa = {};
      estadoCO.dispositivosMapa = {};
      Object.keys(dados).forEach(function (mapaId) {
        (dados[mapaId] || []).forEach(function (s) {
          if (s.idSetor) estadoCO.setorParaMapa[s.idSetor] = parseInt(mapaId, 10);
          var chave = (s.idDispositivo || "") + "|" + (s.zonaUser || "") + "|" + (s.particao || "");
          estadoCO.dispZonaParaMapa[chave] = parseInt(mapaId, 10);
          if (s.idDispositivo && !estadoCO.dispositivosMapa[mapaId]) {
            estadoCO.dispositivosMapa[mapaId] = s.idDispositivo;
          }
        });
      });
      estadoCO.setoresMapasProntos = true;
      if (estadoCO.mapaAtivoId) atualizarBotaoArmarToolbar(estadoCO.mapaAtivoId);
    })
    .always(function () {
      estadoCO.setoresMapasProntos = true;
      if (typeof cb === "function") cb();
    });
}

function processoTemMapa(proc) {
  return !!resolverMapaProcesso(proc);
}

function filtrarProcessosComMapa(lista) {
  return (lista || []).filter(function (p) {
    return processoTemMapa(p);
  });
}

function resolverMapaProcesso(proc) {
  if (!proc) return null;
  if (proc.idSetor && estadoCO.setorParaMapa[proc.idSetor]) {
    return estadoCO.setorParaMapa[proc.idSetor];
  }
  var chave = (proc.idDispositivo || "") + "|" + (proc.zonaUser || "") + "|" + (proc.particao || "");
  if (estadoCO.dispZonaParaMapa[chave]) return estadoCO.dispZonaParaMapa[chave];
  // Só considera mapa se a zona/partição (ou setor) estiver no mapa — não basta o dispositivo existir em algum mapa
  return null;
}

function processosDoMapa(mapaId) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return [];
  return estadoCO.processos.filter(function (p) {
    return resolverMapaProcesso(p) === mapaId;
  });
}

function processoDoMapa(mapaId) {
  var lista = processosDoMapa(mapaId);
  if (!lista.length) return null;
  var melhor = lista[0];
  lista.forEach(function (p) {
    if (parseInt(p.nivel, 10) > parseInt(melhor.nivel, 10)) melhor = p;
  });
  return melhor;
}

function mapaTemProcessoAberto(mapaId) {
  return processosDoMapa(mapaId).length > 0;
}

function statusExibidoMapa(mapaId) {
  var st = estadoCO.statusMapas[mapaId] || "normal";
  if (st === "alarme" && !mapaTemProcessoAberto(mapaId)) {
    return "normal";
  }
  return st;
}

function aplicarStatusCardMapa(mapaId, st) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;
  var $item = $('.co-item-mapa[data-id="' + mapaId + '"]');
  if (!$item.length) return;
  var $st = $item.find(".co-item-mapa-status");
  var $tab = $('.co-tab[data-id="' + mapaId + '"]');
  var fixLista = estadoCO.mapaFixadoListaId;
  var fixAba = estadoCO.mapaFixadoAbaId;

  $st.removeClass("st-alarme st-falha st-normal");
  $item.removeClass("co-mapa-alarme");
  $tab.removeClass("co-tab-alarme");

  if (st === "alarme") {
    $st.addClass("st-alarme").text("DISPARO");
    if (fixLista && fixLista !== mapaId) {
      $item.addClass("co-mapa-alarme");
    } else if (!fixLista && mapaId !== estadoCO.mapaAtivoId) {
      $item.addClass("co-mapa-alarme");
    }
    if (fixAba && fixAba !== mapaId) {
      $tab.addClass("co-tab-alarme");
    } else if (!fixAba && mapaId !== estadoCO.mapaAtivoId) {
      $tab.addClass("co-tab-alarme");
    }
  } else if (st === "falha") {
    $st.addClass("st-falha").text("Sem comunicação");
  } else {
    $st.addClass("st-normal").text("Normal");
  }
}

function redimensionarMapaAtivo(id) {
  var mon = estadoCO.monitores[id];
  if (!mon) return;
  ajustarAlturaMapaCO(id);
  if (mon.reinstalarLayout) mon.reinstalarLayout();
  function tick() {
    ajustarAlturaMapaCO(id);
    mon.redimensionar();
    if (typeof MapaRastro !== "undefined") MapaRastro.desenhar(id);
  }
  [0, 50, 150, 400, 800, 1200].forEach(function (ms) {
    setTimeout(tick, ms);
  });
}

function ajustarAlturaMapaCO(id) {
  id = parseInt(id, 10);
  if (!id) return;
  var paneId = "coPane" + id;
  var $scroll = $("#" + paneId + " .co-mapa-scroll");
  var $container = $("#" + paneId + " .mapa-container");
  if (!$scroll.length || !$container.length) return;

  var scrollEl = $scroll[0];
  var scrollH = scrollEl.clientHeight;
  var scrollW = scrollEl.clientWidth;
  if (scrollH < 1 || scrollW < 1) return;

  $container.css({
    width: scrollW + "px",
    height: scrollH + "px",
    "min-height": scrollH + "px",
    "max-height": scrollH + "px",
  });
}

function carregarMapaSeNecessario(id) {
  id = parseInt(id, 10);
  if (!id) return;
  var mon = ensurePainelMapa(id);
  if (!mon) return;

  $("#coMapaVazio").addClass("hidden");

  if (estadoCO.mapasCarregados[id]) {
    requestAnimationFrame(function () {
      requestAnimationFrame(function () {
        redimensionarMapaAtivo(id);
      });
    });
    return;
  }

  var $st = $('.co-item-mapa[data-id="' + id + '"] .co-item-mapa-status');
  $st.text("Carregando...");

  function abrirAgora() {
    if (mon.reinstalarLayout) mon.reinstalarLayout();
    mon.abrir(id)
      .done(function () {
        redimensionarMapaAtivo(id);
      })
      .fail(function () {
        $st.text("Erro ao carregar");
        $("#coMapaVazio")
          .removeClass("hidden")
          .text("Erro ao carregar a planta deste mapa.");
      });
  }

  requestAnimationFrame(function () {
    requestAnimationFrame(abrirAgora);
  });
}

function ativarMapa(id, manual) {
  id = parseInt(id, 10);
  if (!id) {
    $("#coMapaVazio").removeClass("hidden").text("Mapa inválido.");
    return;
  }
  estadoCO.mapaAtivoId = id;

  if (typeof MapaSetorPopup !== "undefined") MapaSetorPopup.fechar();
  if (typeof MapaCameraLive !== "undefined") MapaCameraLive.fechar();

  if (!estadoCO.mapaFixadoListaId || manual || estadoCO.mapaFixadoListaId === id) {
    $(".co-item-mapa").removeClass("co-mapa-ativo");
    $('.co-item-mapa[data-id="' + id + '"]')
      .addClass("co-mapa-ativo")
      .removeClass("co-mapa-alarme");
  }

  if (!estadoCO.mapaFixadoAbaId || manual || estadoCO.mapaFixadoAbaId === id) {
    $(".co-tab").removeClass("co-tab-ativo");
    $('.co-tab[data-id="' + id + '"]')
      .addClass("co-tab-ativo")
      .removeClass("co-tab-alarme");
    $(".co-mapa-pane").removeClass("co-pane-ativo");
    ensurePainelDOM(id);
    $("#coPane" + id).addClass("co-pane-ativo");
    ensurePainelMapa(id);
    carregarMapaSeNecessario(id);
    scrollTabAtivoParaVista();
  }

  limparAlarmeVisual(id);
  atualizarContadoresMapa(id);
  atualizarBotaoArmarToolbar(id);
  atualizarProcessoAtivo();
  if (typeof MapaRastro !== "undefined") {
    requestAnimationFrame(function () {
      MapaRastro.desenhar(id);
      atualizarUiRastro(id);
      var pred = MapaRastro.getPredicao(id);
      if (pred && pred.mensagem && MapaRastro.getTrail(id).length >= 2) {
        mostrarAlertaRastro(id, pred.mensagem);
      }
      atualizarAlertaRastroMapaAtivo();
    });
  }
  renderTimelineMapa(id);
  carregarOcorrenciaMapa(id);
  atualizarAlertaRastroMapaAtivo();
  if (manual) atualizarFixarUI();
}

function toggleFixarLista(id) {
  id = parseInt(id, 10);
  estadoCO.mapaFixadoListaId = estadoCO.mapaFixadoListaId === id ? null : id;
  atualizarFixarUI();
}

function toggleFixarAba(id) {
  id = parseInt(id, 10);
  estadoCO.mapaFixadoAbaId = estadoCO.mapaFixadoAbaId === id ? null : id;
  if (estadoCO.mapaFixadoAbaId) ativarMapa(id, false);
  atualizarFixarUI();
}

function posicionarMapaFixadoLista() {
  var fixId = estadoCO.mapaFixadoListaId;
  var $fixados = $("#coListaMapasFixados");
  var $rolaveis = $("#coListaMapas");

  $fixados.children(".co-item-mapa").each(function () {
    var $item = $(this);
    var tid = parseInt($item.data("id"), 10);
    if (!fixId || tid !== fixId) {
      $item.removeClass("co-mapa-fixado-lista");
      $rolaveis.prepend($item);
    }
  });

  if (!fixId) return;

  var $card = $('.co-item-mapa[data-id="' + fixId + '"]');
  if (!$card.length) return;
  $card.addClass("co-mapa-fixado-lista");
  $fixados.append($card);
}

function indiceMapaNoArray(mapaId) {
  var idx = -1;
  estadoCO.mapas.forEach(function (m, i) {
    if (mapaIdFrom(m) === mapaId) idx = i;
  });
  return idx;
}

function registrarOrdemDisparo(mapaId) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;
  estadoCO.ordemDisparoMapas[mapaId] = Date.now();

  var idx = indiceMapaNoArray(mapaId);
  if (idx > 0) {
    var item = estadoCO.mapas.splice(idx, 1)[0];
    estadoCO.mapas.unshift(item);
  }

  reordenarListaMapasPorDisparo();
  reordenarAbasPorDisparo();
}

function compararOrdemDisparo(idA, idB) {
  var ta = estadoCO.ordemDisparoMapas[idA] || 0;
  var tb = estadoCO.ordemDisparoMapas[idB] || 0;
  if (tb !== ta) return tb - ta;
  return indiceMapaNoArray(idA) - indiceMapaNoArray(idB);
}

function reordenarListaMapasPorDisparo() {
  var $rolaveis = $("#coListaMapas");
  var cards = [];
  $rolaveis.children(".co-item-mapa").each(function () {
    cards.push($(this));
  });
  cards.sort(function ($a, $b) {
    return compararOrdemDisparo(parseInt($a.data("id"), 10), parseInt($b.data("id"), 10));
  });
  cards.forEach(function ($c) {
    $rolaveis.append($c);
  });
  posicionarMapaFixadoLista();
}

function reordenarAbasPorDisparo() {
  var $rolaveis = $("#coTabs");
  var tabs = [];
  $rolaveis.children(".co-tab").each(function () {
    tabs.push($(this));
  });
  tabs.sort(function ($a, $b) {
    return compararOrdemDisparo(parseInt($a.data("id"), 10), parseInt($b.data("id"), 10));
  });
  tabs.forEach(function ($t) {
    $rolaveis.append($t);
  });
  posicionarAbaFixada();
}

function priorizarMapaNaLista(mapaId) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;
  if (!$('.co-item-mapa[data-id="' + mapaId + '"]').length) return;
  registrarOrdemDisparo(mapaId);
  $("#coListaMapas").scrollTop(0);
}

function deveExibirToastProcesso(proc) {
  var fixAba = estadoCO.mapaFixadoAbaId;
  if (!fixAba) return true;
  return resolverMapaProcesso(proc) === fixAba;
}

function aplicarFiltroToastsMapaFixado() {
  if (typeof CoNotificacoes === "undefined") return;
  var fixAba = estadoCO.mapaFixadoAbaId;
  if (!fixAba) return;
  estadoCO.processos.forEach(function (p) {
    if (!deveExibirToastProcesso(p)) {
      CoNotificacoes.removerProcesso(p.idProcesso);
    }
  });
}

function atualizarFixarUI() {
  $(".co-item-mapa").removeClass("co-mapa-fixado-lista");
  $(".co-tab").removeClass("co-tab-fixado");
  $(".co-fixar-lista, .co-tab-fixar").removeClass("co-fixado");
  posicionarMapaFixadoLista();
  posicionarAbaFixada();
  if (estadoCO.mapaFixadoListaId) {
    $('.co-item-mapa[data-id="' + estadoCO.mapaFixadoListaId + '"] .co-fixar-lista').addClass("co-fixado");
  }
  if (estadoCO.mapaFixadoAbaId) {
    $('.co-tab[data-id="' + estadoCO.mapaFixadoAbaId + '"] .co-tab-fixar').addClass("co-fixado");
  }
  aplicarFiltroToastsMapaFixado();
}

function limparAlarmeVisual(mapaId) {
  $('.co-item-mapa[data-id="' + mapaId + '"]').removeClass("co-mapa-alarme");
  $('.co-tab[data-id="' + mapaId + '"]').removeClass("co-tab-alarme");
}

function focarMapaProcesso(mapaId) {
  if (!mapaId || !estadoCO.mapasProntos) return;

  priorizarMapaNaLista(mapaId);

  var fixLista = estadoCO.mapaFixadoListaId;
  var fixAba = estadoCO.mapaFixadoAbaId;

  if (fixLista && fixLista !== mapaId) {
    $('.co-item-mapa[data-id="' + mapaId + '"]').addClass("co-mapa-alarme");
  } else {
    $('.co-item-mapa[data-id="' + mapaId + '"]').removeClass("co-mapa-alarme");
    $(".co-item-mapa").removeClass("co-mapa-ativo");
    $('.co-item-mapa[data-id="' + mapaId + '"]').addClass("co-mapa-ativo");
  }

  if (fixAba && fixAba !== mapaId) {
    $('.co-tab[data-id="' + mapaId + '"]').addClass("co-tab-alarme");
    scrollTabParaVista(mapaId);
    return;
  }

  if (estadoCO.mapaAtivoId === mapaId) return;

  estadoCO.mapaAtivoId = mapaId;
  $(".co-tab").removeClass("co-tab-ativo");
  $('.co-tab[data-id="' + mapaId + '"]')
    .addClass("co-tab-ativo")
    .removeClass("co-tab-alarme");
  $(".co-mapa-pane").removeClass("co-pane-ativo");
  ensurePainelDOM(mapaId);
  $("#coPane" + mapaId).addClass("co-pane-ativo");
  ensurePainelMapa(mapaId);
  carregarMapaSeNecessario(mapaId);
  scrollTabAtivoParaVista();
  limparAlarmeVisual(mapaId);
  atualizarContadoresMapa(mapaId);
  atualizarBotaoArmarToolbar(mapaId);
  atualizarProcessoAtivo();
  renderTimelineMapa(mapaId);
  atualizarAlertaRastroMapaAtivo();
}

function iniciarPolling() {
  atualizarStatusMapas();
  carregarProcessos();
  estadoCO.timers.status = setInterval(atualizarStatusMapas, 3000);
  estadoCO.timers.eventos = setInterval(carregarProcessos, 4000);
  estadoCO.timers.contadores = setInterval(function () {
    estadoCO.mapas.forEach(function (m) {
      var id = mapaIdFrom(m);
      if (id) atualizarContadoresMapa(id);
    });
  }, 5000);
}

function atualizarStatusMapas() {
  var ids = estadoCO.mapas.map(function (m) { return m.id; });
  if (!ids.length) return;
  $.ajax({
    url: "/centroOperacional/mapasStatus",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ ids: ids }),
  }).done(function (r) {
    estadoCO.statusMapas = (r && r.dados) || r || {};
    atualizarStatusLista();
  });
}

function atualizarStatusLista() {
  estadoCO.mapas.forEach(function (m) {
    aplicarStatusCardMapa(m.id, statusExibidoMapa(m.id));
  });
}

function atualizarContadoresMapa(mapaId) {
  $.ajax({
    url: "/centroOperacional/contadoresMapa",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ mapaId: mapaId }),
  }).done(function (r) {
    if (r.status !== "OK" || !r.dados) return;
    estadoCO.contadoresMapa[mapaId] = r.dados;
    aplicarContadoresNoCard(mapaId, r.dados);
  });
}

function aplicarContadoresNoCard(mapaId, d) {
  var $item = $('.co-item-mapa[data-id="' + mapaId + '"]');
  if (!$item.length || !d) return;
  $item.find(".co-item-cnt-zonas").text(d.totalZonas || 0);
  $item.find(".co-item-cnt-alarmes").text(d.totalAlarmes || 0);
  $item.find(".co-item-cnt-semcom").text(d.totalSemComunicacao || 0);
  var $keep = $item.find(".co-item-mapa-keep");
  $keep.removeClass("co-keep-online co-keep-offline");
  if (d.online) {
    $keep.addClass("co-keep-online").text("online");
  } else {
    $keep.addClass("co-keep-offline").text("offline");
  }
  atualizarBotaoArmarToolbar(mapaId, d.armado);
}

function atualizarBotaoArmarToolbar(mapaId, armadoStr) {
  var idDisp = estadoCO.dispositivosMapa[mapaId];
  if (!idDisp) return;
  var $btn = $('.co-item-mapa[data-id="' + mapaId + '"] .co-item-btn-armar');
  if (!$btn.length) return;

  function aplicar(armado, senha, particao) {
    $btn.toggleClass("armado", armado).toggleClass("desarmado", !armado)
      .text(armado ? "Armado" : "Armar")
      .data("armado", armado ? "S" : "N")
      .data("senha", senha || "")
      .data("particao", particao || "01")
      .data("disp", idDisp);
  }

  if (armadoStr) {
    aplicar(armadoStr === "S", "", "01");
    return;
  }

  $.ajax({
    url: "/centroOperacional/dispositivoStatus",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ idDispositivo: idDisp }),
  }).done(function (r) {
    if (r.status !== "OK" || !r.dados) return;
    aplicar(r.dados.armado === "S", r.dados.senha, r.dados.particao);
  });
}

function comandoArmarMapa(mapaId) {
  mapaId = parseInt(mapaId || estadoCO.mapaAtivoId, 10);
  if (!mapaId) return;
  var $btn = $('.co-item-mapa[data-id="' + mapaId + '"] .co-item-btn-armar');
  if (!$btn.length) return;
  var idDisp = $btn.data("disp") || estadoCO.dispositivosMapa[mapaId];
  if (!idDisp) {
    if (typeof boxErro === "function") boxErro("Dispositivo não identificado neste mapa.");
    return;
  }
  var armado = $btn.data("armado") === "S";
  $.ajax({
    url: "/centroOperacional/comando",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      idDispositivo: idDisp,
      acao: armado ? "desarmar" : "armar",
      senha: $btn.data("senha") || "",
      particao: $btn.data("particao") || "01",
    }),
  })
    .fail(function (e) {
      var msg = (e.responseJSON && e.responseJSON.status) || "Erro ao enviar comando.";
      if (typeof boxErro === "function") boxErro(String(msg).replace(/^Erro:\s*/i, ""));
    })
    .done(function () {
      if (typeof CoTimeline !== "undefined") {
        CoTimeline.registrarArme({
          mapaId: mapaId,
          armado: !armado,
        });
        renderTimelineMapa(mapaId);
      }
      setTimeout(function () {
        atualizarBotaoArmarToolbar(mapaId);
        atualizarContadoresMapa(mapaId);
      }, 1500);
    });
}

function carregarProcessos() {
  if (!estadoCO.setoresMapasProntos) return;
  var payload = estadoCO.modoGlobal
    ? { global: true }
    : { idCliente: estadoCO.idCliente, idFranqueado: estadoCO.idFranqueado };

  $.ajax({
    url: "/centroOperacional/eventos",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(payload),
  }).done(function (r) {
    var lista = r.status === "Vazio" ? [] : r.dados || [];
    processarProcessos(filtrarProcessosComMapa(lista));
  });
}

function mapasComProcessosAbertos(lista) {
  var mapas = {};
  (lista || []).forEach(function (p) {
    var mapaId = resolverMapaProcesso(p);
    if (mapaId) mapas[mapaId] = true;
  });
  return mapas;
}

function sincronizarFinalizacaoRemota(processosAnteriores, novaLista) {
  var mapasAntes = mapasComProcessosAbertos(processosAnteriores);
  var mapasAgora = mapasComProcessosAbertos(novaLista);
  Object.keys(mapasAntes).forEach(function (mapaIdStr) {
    var mapaId = parseInt(mapaIdStr, 10);
    if (!mapasAgora[mapaId]) {
      limparVisualAposFinalizar(mapaId, null, true);
    }
  });
}

function processarProcessos(novaLista) {
  var novosSirene = false;
  var novosIds = [];
  var processosAnteriores = estadoCO.processos || [];
  novaLista = filtrarProcessosComMapa(novaLista);

  novaLista.forEach(function (proc) {
    var chave = proc.idProcesso;
    var mapaId = resolverMapaProcesso(proc);
    if (!mapaId) return;

    if (!estadoCO.processosVistos[chave]) {
      estadoCO.processosVistos[chave] = true;
      estadoCO.processosOrdemChegada[chave] = Date.now();
      novosIds.push(chave);
      if (proc.somSirene) novosSirene = true;
      focarMapaProcesso(mapaId);
      iniciarProgressoProcesso(proc);
      if (deveExibirToastProcesso(proc) && typeof CoNotificacoes !== "undefined") {
        CoNotificacoes.abrir(proc);
      }
    } else if (!estadoCO.processosOrdemChegada[chave]) {
      estadoCO.processosOrdemChegada[chave] = Date.now();
    }
    var cat = categoriaDoGrupo(proc.grupo);
    if (processoEhAlarme(proc)) {
      sincronizarRastroProcesso(proc);
    }
    carregarTimelineProcesso(proc);
  });

  if (novosSirene && estadoCO.somAtivo) tocarSom(false);
  else if (!temProcessoSireneAtivo()) pararSom();

  var idsAtuais = novaLista.map(function (p) { return p.idProcesso; });
  if (estadoCO.processoAtivo && idsAtuais.indexOf(estadoCO.processoAtivo.idProcesso) < 0) {
    pararProgresso();
  }
  if (typeof CoNotificacoes !== "undefined") {
    Object.keys(estadoCO.processosVistos).forEach(function (id) {
      if (idsAtuais.indexOf(id) < 0) {
        CoNotificacoes.removerProcesso(id);
        delete estadoCO.processosOrdemChegada[id];
        delete estadoCO.processosVistos[id];
        delete estadoCO.processosTrailSync[id];
        delete estadoCO.processosTimelineSync[id];
      }
    });
  }

  estadoCO.processos = ordenarProcessosPorChegada(novaLista);
  sincronizarFinalizacaoRemota(processosAnteriores, novaLista);
  atualizarProcessoAtivo();
  atualizarStatusLista();
  renderFilaProcessos(novosIds.length > 0);
}

function ordenarProcessosPorChegada(lista) {
  return (lista || []).slice().sort(function (a, b) {
    var ta = estadoCO.processosOrdemChegada[a.idProcesso] || 0;
    var tb = estadoCO.processosOrdemChegada[b.idProcesso] || 0;
    if (tb !== ta) return tb - ta;
    var na = parseInt(a.nivel, 10) || 0;
    var nb = parseInt(b.nivel, 10) || 0;
    if (nb !== na) return nb - na;
    return String(b.dataHora || "").localeCompare(String(a.dataHora || ""));
  });
}

function temProcessoSireneAtivo() {
  return estadoCO.processos.some(function (p) { return p.somSirene; });
}

function atualizarProcessoAtivo() {
  estadoCO.processoAtivo = estadoCO.mapaAtivoId ? processoDoMapa(estadoCO.mapaAtivoId) : null;
  if (!estadoCO.processoAtivo && estadoCO.processos.length) {
    estadoCO.processoAtivo = estadoCO.processos[0];
  }
  atualizarBotoesEventoMapas();
}

function atualizarBotoesEventoMapas() {
  $(".co-item-btn-evento").prop("disabled", true);
  estadoCO.mapas.forEach(function (m) {
    var id = mapaIdFrom(m);
    if (id && processoDoMapa(id)) {
      $('.co-item-mapa[data-id="' + id + '"] .co-item-btn-evento').prop("disabled", false);
    }
  });
}

function iniciarProgressoProcesso(proc) {
  pararProgresso();
  estadoCO.progressoInicio = Date.now();
  var mapaId = resolverMapaProcesso(proc) || estadoCO.mapaAtivoId;
  var $prog = mapaId
    ? $('.co-item-mapa[data-id="' + mapaId + '"] .co-item-mapa-progress')
    : $();
  if (!$prog.length) return;
  $prog.removeClass("hidden co-progress-expirado");
  var $fill = $prog.find(".co-progress-fill");
  $fill.css({ width: "100%", transition: "width 30s linear" });
  requestAnimationFrame(function () {
    $fill.css("width", "0%");
  });
  estadoCO.progressoTimer = setTimeout(function () {
    $prog.addClass("co-progress-expirado");
  }, 30000);
}

function pararProgresso() {
  if (estadoCO.progressoTimer) clearTimeout(estadoCO.progressoTimer);
  estadoCO.progressoTimer = null;
  $(".co-item-mapa-progress").addClass("hidden").removeClass("co-progress-expirado");
  $(".co-item-mapa-progress .co-progress-fill").css({ width: "100%", transition: "none" });
}

function categoriaDoGrupo(grupo) {
  var g = normalizarGrupo(grupo);
  if (GRUPOS_CATEGORIA[g]) return GRUPOS_CATEGORIA[g];
  // Fallback para grupos não cadastrados ou variações de escrita
  if (g.indexOf("FALHA") >= 0) return "falhas";
  if (g.indexOf("DESARME") >= 0) return "controle";
  if (g.indexOf("RESTAUR") >= 0) return "controle";
  if (g === "ARME" || g.indexOf("ARME") >= 0) return "controle";
  if (GRUPOS_SIRENE.some(function (k) { return g.indexOf(k) >= 0; })) return "alarmes";
  return "geral";
}

function processoEhAlarme(proc) {
  return categoriaDoGrupo(proc && proc.grupo) === "alarmes";
}

function normalizarGrupo(g) {
  return String(g || "")
    .toUpperCase()
    .trim()
    .replace(/Ã/g, "A")
    .replace(/É/g, "E")
    .replace(/Ê/g, "E")
    .replace(/Í/g, "I")
    .replace(/Ó/g, "O");
}

function processoTemCamera(p) {
  return (
    String((p && p.cameraOn) || "")
      .trim()
      .toUpperCase() === "S" &&
    !!(p.idSetor || p.idDispositivo)
  );
}

function processoAberto(p) {
  return p && String(p.statusAtendimento || "").toUpperCase() !== "ATENDIDO";
}

function resolverProcessoDoSetor(dados) {
  if (!dados || !estadoCO.processos.length) return null;

  var idSetor = String(dados.idSetor || "").trim().toUpperCase();
  if (idSetor) {
    var porId = estadoCO.processos.find(function (p) {
      return processoAberto(p) && String(p.idSetor || "").trim().toUpperCase() === idSetor;
    });
    if (porId) return porId;
  }

  var idDisp = String(dados.idDispositivo || "").trim();
  var zona = pad3(dados.numero || "");
  var part = pad2(dados.particao || "");

  return (
    estadoCO.processos.find(function (p) {
      if (!processoAberto(p)) return false;
      if (idDisp && String(p.idDispositivo || "").trim() !== idDisp) return false;
      if (idSetor && String(p.idSetor || "").trim().toUpperCase() === idSetor) return true;
      return pad3(p.zonaUser) === zona && pad2(p.particao) === part;
    }) || null
  );
}

function abrirCameraSetor(opts) {
  if (typeof MapaCameraLive === "undefined") return;
  var dados = {
    idSetor: (opts && opts.idSetor) || "",
    idDispositivo: (opts && opts.idDispositivo) || "",
    idFranqueado:
      (opts && opts.idFranqueado) ||
      estadoCO.idFranqueado ||
      "",
    camera: "S",
    setorNome: (opts && (opts.setorNome || opts.nomeSetor)) || "",
    label: (opts && opts.label) || "",
  };
  if (!dados.idSetor && !dados.idDispositivo) {
    if (typeof boxErro === "function") boxErro("Setor não identificado para câmera ao vivo.");
    return;
  }
  MapaCameraLive.abrirAoVivo(dados);
  if (typeof CoTimeline !== "undefined") {
    var mapaId =
      parseInt(opts && opts.mapaId, 10) ||
      (dados.idSetor && estadoCO.setorParaMapa[dados.idSetor]) ||
      estadoCO.mapaAtivoId;
    if (mapaId) {
      CoTimeline.registrarCamera({
        mapaId: mapaId,
        idProcesso: (opts && opts.idProcesso) || (estadoCO.processoAtivo && estadoCO.processoAtivo.idProcesso) || "",
        idSetor: dados.idSetor,
        setorNome: dados.setorNome || dados.label,
        label: dados.label,
      });
      if (mapaId === estadoCO.mapaAtivoId) renderTimelineMapa(mapaId);
    }
  }
}

function renderFilaProcessos(scrollTopo) {
  var $fila = $("#coFilaEventos");
  $fila.empty();
  var filtro = estadoCO.filtroAtivo;
  var lista = ordenarProcessosPorChegada(
    estadoCO.processos.filter(function (p) {
      if (filtro === "todos") return true;
      return categoriaDoGrupo(p.grupo) === filtro;
    })
  );

  if (!lista.length) {
    $fila.html('<p class="co-msg">Nenhum processo aberto.</p>');
    return;
  }

  lista.forEach(function (p) {
    var statusCls = p.statusAtendimento === "ATENDIDO" ? "co-status-atendido" : "co-status-aberto";
    var statusTxt = p.statusAtendimento === "ATENDIDO" ? "Atendido" : "Aberto";
    var zonaPart = "Z-" + pad3(p.zonaUser) + " / P" + pad2(p.particao);
    var temCam = processoTemCamera(p);
    var btnAoVivo = temCam
      ? '<button type="button" class="co-btn-aovivo-card" title="Abrir câmera ao vivo">' +
        '<i class="bi bi-camera-video-fill"></i> Ao vivo</button>'
      : "";
    var card = $(
      '<div class="co-evento-card" data-id="' + escAttr(p.idProcesso) + '">' +
        '<div class="co-evento-top">' +
        '<span class="co-evento-titulo">' + esc(p.descricaoGrupo || p.codigo) + "</span>" +
        '<span class="co-status-badge ' + statusCls + '">' + statusTxt + "</span></div>" +
        '<div class="co-evento-grupo">' + esc(p.grupo) + " · " + esc(p.nomeCliente) + "</div>" +
        '<div class="co-evento-detalhe">' +
        esc(zonaPart) +
        (p.nomeSetor ? "<br>" + esc(p.nomeSetor) : "") +
        "<br>" +
        esc(p.nomeDispositivo) +
        "<br>" +
        esc(p.dataHora) +
        (p.nomeOperador ? " · " + esc(p.nomeOperador) : "") +
        "</div>" +
        '<div class="co-evento-acoes">' +
        btnAoVivo +
        '<button type="button" class="co-btn-finalizar-card">Finalizar</button></div></div>'
    );
    card.on("click", function (e) {
      if ($(e.target).closest(".co-btn-finalizar-card, .co-btn-aovivo-card").length) return;
      abrirModalProcesso(p);
    });
    card.find(".co-btn-aovivo-card").on("click", function (e) {
      e.stopPropagation();
      var mapaId = resolverMapaProcesso(p);
      if (mapaId) focarMapaProcesso(mapaId);
      abrirCameraSetor({
        idSetor: p.idSetor,
        idDispositivo: p.idDispositivo,
        idFranqueado: p.idFranqueado,
        nomeSetor: p.nomeSetor,
        label: p.descricaoGrupo || p.codigo,
      });
    });
    card.find(".co-btn-finalizar-card").on("click", function (e) {
      e.stopPropagation();
      abrirModalProcesso(p);
    });
    $fila.append(card);
  });

  if (scrollTopo) {
    $fila.scrollTop(0);
  }
}

function abrirModalProcesso(proc) {
  estadoCO.processoModal = proc;
  $("#coModalTitulo").text("Processo — " + (proc.descricaoGrupo || proc.codigo));
  $("#coFinalizarDesc").val("");
  $("#coModalResumo").html(
    "<strong>" + esc(proc.nomeCliente) + "</strong> · " + esc(proc.nomeDispositivo) +
    "<br>" + esc(proc.grupo) + " · " + esc(proc.dataHora) +
    "<br>Zona " + esc(proc.zonaUser) + " · Partição " + esc(proc.particao) +
    (proc.nomeSetor ? "<br>" + esc(proc.nomeSetor) : "")
  );
  var $btnCam = $("#coBtnModalAoVivo");
  if (processoTemCamera(proc)) {
    $btnCam.removeClass("hidden").off("click").on("click", function () {
      var mapaId = resolverMapaProcesso(proc);
      if (mapaId) focarMapaProcesso(mapaId);
      abrirCameraSetor({
        idSetor: proc.idSetor,
        idDispositivo: proc.idDispositivo,
        idFranqueado: proc.idFranqueado,
        nomeSetor: proc.nomeSetor,
        label: proc.descricaoGrupo || proc.codigo,
      });
    });
  } else {
    $btnCam.addClass("hidden");
  }
  $("#coTabelaEventos").html('<p class="co-msg">Carregando eventos...</p>');
  $("#coModalProcesso").removeClass("hidden");
  carregarDetalheProcesso(proc.idProcesso, proc.idFranqueado);
}

function fecharModalProcesso() {
  $("#coModalProcesso").addClass("hidden");
  estadoCO.processoModal = null;
}

function carregarDetalheProcesso(idProcesso, idFranqueadoProc) {
  $.ajax({
    url: "/centroOperacional/processoDetalhe",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      idProcesso: idProcesso,
      idFranqueado: idFranqueadoProc || estadoCO.idFranqueado,
    }),
  }).done(function (r) {
    var lista = r.status === "Vazio" ? [] : r.dados || [];
    var $t = $("#coTabelaEventos");
    $t.empty();
    if (!lista.length) {
      $t.html('<p class="co-msg">Sem eventos agrupados.</p>');
      return;
    }
    lista.forEach(function (ev) {
      var temCam = ev.cameraOn === "S";
      var linha = $(
        '<div class="co-linha-evento">' +
        '<span class="co-linha-evento-qtd">' + ev.quantidade + "x</span>" +
        "<div><strong>" + esc(ev.descricao) + "</strong><br>" +
        esc(ev.grupo) + " · Z-" + esc(ev.zonaUser) + " P" + esc(ev.particao) +
        "<br><small>" + esc(ev.hora) + "</small></div>" +
        '<button type="button" class="co-btn-camera' + (temCam ? "" : " disabled") + '" title="Câmera">' +
        '<i class="bi bi-camera-video-fill"></i></button></div>'
      );
      if (temCam) {
        linha.find(".co-btn-camera").on("click", function (e) {
          e.stopPropagation();
          var proc = estadoCO.processoModal || {};
          abrirCameraSetor({
            idSetor: ev.idSetor,
            idDispositivo: ev.idDispositivo || proc.idDispositivo,
            idFranqueado: proc.idFranqueado || estadoCO.idFranqueado,
            nomeSetor: ev.descZona,
            label: ev.descricao,
          });
        });
      }
      $t.append(linha);
    });
  });
}

function finalizarProcessoAtual() {
  var proc = estadoCO.processoModal;
  if (!proc) return;
  var desc = $("#coFinalizarDesc").val().trim();
  if (!desc) {
    if (typeof boxErro === "function") boxErro("Informe a descrição do atendimento.");
    else alert("Informe a descrição do atendimento.");
    return;
  }
  $.ajax({
    url: "/centroOperacional/finalizarProcesso",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      idProcesso: proc.idProcesso,
      idCliente: proc.idCliente || estadoCO.idCliente,
      descricao: desc,
      idOperador: sessionStorage.getItem("login_userIdOperador") || "",
      nomeOperador: sessionStorage.getItem("login_userNick") || sessionStorage.getItem("login_userNome") || "OP",
    }),
  })
    .fail(function (e) {
      var msg = (e.responseJSON && e.responseJSON.status) || "Erro ao finalizar.";
      if (typeof boxErro === "function") boxErro(String(msg).replace(/^Erro:\s*/i, ""));
    })
    .done(function () {
      var mapaId = resolverMapaProcesso(proc) || estadoCO.mapaAtivoId;
      if (typeof CoTimeline !== "undefined" && mapaId) {
        CoTimeline.registrarFinalizacao({
          mapaId: mapaId,
          idProcesso: proc.idProcesso,
          descricao: desc,
        });
        renderTimelineMapa(mapaId);
      }
      limparVisualAposFinalizar(mapaId, proc.idProcesso);
      fecharModalProcesso();
      pararSom();
      if (typeof CoNotificacoes !== "undefined") {
        CoNotificacoes.removerProcesso(proc.idProcesso);
      }
      delete estadoCO.processosVistos[proc.idProcesso];
      delete estadoCO.processosOrdemChegada[proc.idProcesso];
      carregarProcessos();
      if (estadoCO.mapaAtivoId) atualizarContadoresMapa(estadoCO.mapaAtivoId);
      if (mapaId) atualizarContadoresMapa(mapaId);
      if (typeof boxSucesso === "function") boxSucesso("Processo finalizado.");
    });
}

function limparVisualAposFinalizar(mapaId, idProcesso, remoto) {
  mapaId = parseInt(mapaId, 10);
  if (!mapaId) return;

  if (idProcesso) {
    estadoCO.processos = estadoCO.processos.filter(function (p) {
      return p.idProcesso !== idProcesso;
    });
    delete estadoCO.processosTrailSync[idProcesso];
    delete estadoCO.processosTimelineSync[idProcesso];
  }

  if (typeof MapaRastro !== "undefined" && MapaRastro.isModoAnalise() && mapaId === estadoCO.mapaAtivoId) {
    MapaRastro.setModoAnalise(false);
  }
  ocultarAlertaRastro(mapaId, true);
  pararProgresso();

  if (!mapaTemProcessoAberto(mapaId)) {
    estadoCO.statusMapas[mapaId] = "normal";
    delete estadoCO.ordemDisparoMapas[mapaId];
    aplicarStatusCardMapa(mapaId, "normal");
    ocultarAlertaRastro(mapaId, true);
  }

  if (typeof MapaRastro !== "undefined") {
    MapaRastro.limparMapa(mapaId);
    if (mapaId === estadoCO.mapaAtivoId) {
      MapaRastro.desenhar(mapaId);
    }
    atualizarUiRastro(mapaId);
  }

  $("#coPane" + mapaId + " .mapa-ponto-monitor")
    .removeClass(
      "mapa-ponto-alarme mapa-ponto-rastro-antigo mapa-ponto-rastro-atual mapa-ponto-rastro-predito"
    )
    .addClass("mapa-ponto-normal");

  var mon = estadoCO.monitores[mapaId];
  if (mon && typeof mon.atualizarStatus === "function") {
    mon.atualizarStatus();
    setTimeout(function () {
      mon.atualizarStatus();
    }, 350);
    setTimeout(function () {
      mon.atualizarStatus();
    }, 1200);
  }

  if (!remoto) {
    atualizarProcessoAtivo();
  }
  if (mapaId === estadoCO.mapaAtivoId || estadoCO.colMapasView === "timeline") {
    renderTimelineMapa(mapaId);
  }
  atualizarStatusMapas();
}

function exportarProcessos() {
  var lista = estadoCO.processos;
  if (!lista.length) {
    if (typeof boxErro === "function") boxErro("Nenhum processo para exportar.");
    return;
  }
  var cols = ["dataHora", "descricaoGrupo", "grupo", "nomeCliente", "nomeDispositivo", "zonaUser", "particao", "statusAtendimento", "idProcesso"];
  var csv = cols.join(";") + "\n";
  lista.forEach(function (p) {
    csv += cols.map(function (c) {
      return '"' + String(p[c] || "").replace(/"/g, '""') + '"';
    }).join(";") + "\n";
  });
  var blob = new Blob(["\ufeff" + csv], { type: "text/csv;charset=utf-8" });
  var a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = "processos_" + estadoCO.idCliente + "_" + new Date().toISOString().slice(0, 10) + ".csv";
  a.click();
}

function pad3(v) {
  v = String(v || "").trim();
  while (v.length < 3) v = "0" + v;
  return v;
}

function pad2(v) {
  v = String(v || "").trim();
  while (v.length < 2) v = "0" + v;
  return v;
}

function esc(s) {
  return $("<span>").text(s || "").html();
}

function escAttr(s) {
  return String(s || "").replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;");
}

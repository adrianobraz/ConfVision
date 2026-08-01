// Notificações estilo Windows Toast — Centro Operacional
var CoNotificacoes = (function () {
  var TTL_MS = 60000;
  var PROGRESSO_MS = 30000;
  var MAX_VISIVEIS = 5;
  var fila = [];
  var hooks = {
    onAbrir: null,
    onFinalizar: null,
    onCamera: null,
    onMapa: null,
  };

  function esc(s) {
    return String(s == null ? "" : s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function garantirHost() {
    var $host = $("#coToastHost");
    if ($host.length) return $host;
    $host = $('<div id="coToastHost" class="co-toast-host" aria-live="polite"></div>');
    $("body").append($host);
    return $host;
  }

  function init(opts) {
    opts = opts || {};
    hooks.onAbrir = opts.onAbrir || null;
    hooks.onFinalizar = opts.onFinalizar || null;
    hooks.onCamera = opts.onCamera || null;
    hooks.onMapa = opts.onMapa || null;
    garantirHost();
  }

  function pad3(v) {
    var s = String(v == null ? "" : v).replace(/\D/g, "");
    while (s.length < 3) s = "0" + s;
    return s || "000";
  }

  function pad2(v) {
    var s = String(v == null ? "" : v).replace(/\D/g, "");
    while (s.length < 2) s = "0" + s;
    return s || "00";
  }

  function temCamera(p) {
    return (
      String((p && p.cameraOn) || "")
        .trim()
        .toUpperCase() === "S" && !!(p.idSetor || p.idDispositivo)
    );
  }

  function categoryClass(grupo) {
    var g = String(grupo || "")
      .toUpperCase()
      .trim();
    if (
      g === "ALARME" ||
      g === "PANICO" ||
      g === "EMERGENCIA" ||
      g === "MEDICO" ||
      g.indexOf("ALARME") >= 0 ||
      g.indexOf("PANICO") >= 0 ||
      g.indexOf("EMERGENCIA") >= 0 ||
      g.indexOf("MEDICO") >= 0
    ) {
      return "co-toast-alarme";
    }
    if (g.indexOf("FALHA") >= 0) return "co-toast-falha";
    if (
      g === "ARME" ||
      g === "DESARME" ||
      g === "RESTAURE" ||
      g.indexOf("DESARME") >= 0 ||
      g.indexOf("RESTAUR") >= 0 ||
      (g.indexOf("ARME") >= 0 && g.indexOf("DESARME") < 0)
    ) {
      return "co-toast-controle";
    }
    return "co-toast-info";
  }

  function fechar(idProcesso, motivo) {
    var idx = fila.findIndex(function (t) {
      return t.idProcesso === idProcesso;
    });
    if (idx < 0) return;
    var item = fila[idx];
    if (item.timerTtl) clearTimeout(item.timerTtl);
    if (item.timerProg) clearTimeout(item.timerProg);
    if (item.$el && item.$el.length) {
      item.$el.addClass("co-toast-saindo");
      setTimeout(function () {
        item.$el.remove();
      }, 220);
    }
    fila.splice(idx, 1);
    trimFila();
  }

  function trimFila() {
    while (fila.length > MAX_VISIVEIS) {
      var antigo = fila[fila.length - 1];
      fechar(antigo.idProcesso, "limite");
    }
  }

  function abrir(proc) {
    if (!proc || !proc.idProcesso) return null;
    var existente = fila.find(function (t) {
      return t.idProcesso === proc.idProcesso;
    });
    if (existente) {
      existente.proc = proc;
      existente.$el.find(".co-toast-body").html(montarBodyHtml(proc));
      existente.$el.removeClass("co-toast-saindo").addClass("co-toast-pulse-in");
      return existente;
    }

    var $host = garantirHost();
    var inicio = Date.now();
    var $el = $(
      '<div class="co-toast ' +
        categoryClass(proc.grupo) +
        '" data-id="' +
        esc(proc.idProcesso) +
        '" role="alert">' +
        '<div class="co-toast-head">' +
        '<div class="co-toast-titulo">' +
        '<i class="bi bi-bell-fill"></i>' +
        "<div>" +
        "<strong>" +
        esc(proc.descricaoGrupo || proc.codigo || "Evento") +
        "</strong>" +
        '<span class="co-toast-sub">' +
        esc(proc.grupo || "") +
        (proc.nomeCliente ? " · " + esc(proc.nomeCliente) : "") +
        "</span>" +
        "</div></div>" +
        '<button type="button" class="co-toast-fechar" title="Fechar">&times;</button>' +
        "</div>" +
        '<div class="co-toast-body">' +
        montarBodyHtml(proc) +
        "</div>" +
        '<div class="co-toast-progress">' +
        '<div class="co-toast-progress-track"><div class="co-toast-progress-fill"></div></div>' +
        '<span class="co-toast-progress-label">Atendimento 30s</span>' +
        "</div>" +
        '<div class="co-toast-acoes">' +
        '<button type="button" class="co-toast-btn co-toast-btn-mapa"><i class="bi bi-map"></i> Mapa</button>' +
        '<button type="button" class="co-toast-btn co-toast-btn-abrir"><i class="bi bi-eye"></i> Abrir</button>' +
        (temCamera(proc)
          ? '<button type="button" class="co-toast-btn co-toast-btn-cam"><i class="bi bi-camera-video-fill"></i> Ao vivo</button>'
          : "") +
        '<button type="button" class="co-toast-btn co-toast-btn-fin"><i class="bi bi-check2-circle"></i> Finalizar</button>' +
        "</div>" +
        '<div class="co-toast-ttl"><div class="co-toast-ttl-fill"></div></div>' +
        "</div>"
    );

    $host.append($el);

    requestAnimationFrame(function () {
      $el.find(".co-toast-progress-fill").css({
        width: "0%",
        transition: "width " + PROGRESSO_MS + "ms linear",
      });
      $el.find(".co-toast-ttl-fill").css({
        width: "0%",
        transition: "width " + TTL_MS + "ms linear",
      });
    });

    var item = {
      idProcesso: proc.idProcesso,
      proc: proc,
      $el: $el,
      inicio: inicio,
      timerTtl: null,
      timerProg: null,
    };

    $el.find(".co-toast-fechar").on("click", function (e) {
      e.stopPropagation();
      fechar(proc.idProcesso, "manual");
    });
    $el.find(".co-toast-btn-abrir").on("click", function (e) {
      e.stopPropagation();
      if (typeof hooks.onAbrir === "function") hooks.onAbrir(item.proc);
    });
    $el.find(".co-toast-btn-fin").on("click", function (e) {
      e.stopPropagation();
      if (typeof hooks.onFinalizar === "function") hooks.onFinalizar(item.proc);
    });
    $el.find(".co-toast-btn-cam").on("click", function (e) {
      e.stopPropagation();
      if (typeof hooks.onCamera === "function") hooks.onCamera(item.proc);
    });
    $el.find(".co-toast-btn-mapa").on("click", function (e) {
      e.stopPropagation();
      if (typeof hooks.onMapa === "function") hooks.onMapa(item.proc);
    });
    $el.on("click", function (e) {
      if ($(e.target).closest("button").length) return;
      if (typeof hooks.onAbrir === "function") hooks.onAbrir(item.proc);
    });

    item.timerProg = setTimeout(function () {
      $el.addClass("co-toast-expirado");
      $el.find(".co-toast-progress-label").text("Atendimento expirado");
    }, PROGRESSO_MS);

    item.timerTtl = setTimeout(function () {
      fechar(proc.idProcesso, "ttl");
    }, TTL_MS);

    fila.unshift(item);
    trimFila();
    return item;
  }

  function montarBodyHtml(proc) {
    var zonaPart = "Z-" + pad3(proc.zonaUser) + " / P" + pad2(proc.particao);
    return (
      '<div class="co-toast-linha">' +
      esc(zonaPart) +
      (proc.nomeSetor ? " · " + esc(proc.nomeSetor) : "") +
      "</div>" +
      '<div class="co-toast-linha">' +
      esc(proc.nomeDispositivo || "") +
      "</div>" +
      '<div class="co-toast-linha co-toast-hora">' +
      esc(proc.dataHora || "") +
      (proc.quantidade ? " · " + esc(proc.quantidade) + "x" : "") +
      "</div>"
    );
  }

  function removerProcesso(idProcesso) {
    fechar(idProcesso, "processo");
  }

  function limpar() {
    fila.slice().forEach(function (t) {
      fechar(t.idProcesso, "limpar");
    });
  }

  return {
    init: init,
    abrir: abrir,
    fechar: fechar,
    removerProcesso: removerProcesso,
    limpar: limpar,
  };
})();

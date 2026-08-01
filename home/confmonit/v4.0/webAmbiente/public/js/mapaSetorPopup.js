// Popup somente leitura ao clicar em setor no monitoramento
var MapaSetorPopup = (function () {
  var porContainer = {};
  var docClickBound = false;

  function containerKey($container) {
    var id = $container.attr("id");
    if (id) return id;
    var key = $container.data("popupKey");
    if (!key) {
      key = "mapa-popup-" + String(Date.now()) + "-" + String(Math.random()).slice(2, 8);
      $container.data("popupKey", key);
    }
    return key;
  }

  function criarPopupHtml() {
    return $(`
      <div class="editor-popup monitor-setor-popup hidden">
        <button type="button" class="editor-popup-fechar monitor-popup-fechar" title="Fechar">&times;</button>
        <h4 class="editor-popup-titulo monitor-popup-titulo"></h4>
        <dl class="editor-popup-dados">
          <dt>Status</dt><dd class="monitor-popup-status"></dd>
          <dt>Setor</dt><dd class="monitor-popup-setor"></dd>
          <dt>Dispositivo</dt><dd class="monitor-popup-dispositivo"></dd>
          <dt>Cliente</dt><dd class="monitor-popup-cliente"></dd>
          <dt>Descrição</dt><dd class="monitor-popup-descricao"></dd>
        </dl>
        <button type="button" class="btn btn-info btn-sm btn-block monitor-popup-live hidden">
          <i class="bi bi-broadcast"></i> Ao vivo
        </button>
        <button type="button" class="btn btn-danger btn-sm btn-block monitor-popup-finalizar hidden">
          <i class="bi bi-check-circle"></i> Finalizar processo
        </button>
      </div>
    `);
  }

  function ensurePopup($container, options) {
    if (!$container || !$container.length) return null;

    var key = containerKey($container);
    if (porContainer[key]) {
      if (options && options.abrirCameraAutomatica) {
        porContainer[key].abrirCameraAutomatica = true;
      }
      if (options && options.resolverProcesso) {
        porContainer[key].resolverProcesso = options.resolverProcesso;
      }
      if (options && options.onFinalizarProcesso) {
        porContainer[key].onFinalizarProcesso = options.onFinalizarProcesso;
      }
      if (options && options.onAbrirCamera) {
        porContainer[key].onAbrirCamera = options.onAbrirCamera;
      }
      return porContainer[key];
    }

    var $popup = $container.find(".monitor-setor-popup");
    if (!$popup.length) {
      $popup = criarPopupHtml();
      $container.append($popup);
    }

    $popup.find(".monitor-popup-fechar").off("click.monitorPopup").on("click.monitorPopup", function () {
      fecharInstancia(porContainer[key]);
    });
    $popup.find(".monitor-popup-live").off("click.monitorPopup").on("click.monitorPopup", function (e) {
      e.stopPropagation();
      var dados = $popup.data("setorAtual");
      if (!dados) return;
      var instAtual = porContainer[key];
      if (instAtual && typeof instAtual.onAbrirCamera === "function") {
        instAtual.onAbrirCamera(dados);
        return;
      }
      if (typeof MapaCameraLive !== "undefined") {
        MapaCameraLive.abrirAoVivo(dados);
      }
    });
    $popup.find(".monitor-popup-finalizar").off("click.monitorPopup").on("click.monitorPopup", function (e) {
      e.stopPropagation();
      var dados = $popup.data("setorAtual");
      var proc = $popup.data("processoAtual");
      var instAtual = porContainer[key];
      if (proc && instAtual && typeof instAtual.onFinalizarProcesso === "function") {
        instAtual.onFinalizarProcesso(proc, dados);
      }
      fecharInstancia(instAtual);
    });

    var inst = {
      key: key,
      $container: $container,
      $popup: $popup,
      abrirCameraAutomatica: !!(options && options.abrirCameraAutomatica),
      resolverProcesso: options && options.resolverProcesso,
      onFinalizarProcesso: options && options.onFinalizarProcesso,
      onAbrirCamera: options && options.onAbrirCamera,
    };
    porContainer[key] = inst;

    if (!docClickBound) {
      $(document).on("click.monitorPopup", function (e) {
        if (
          !$(e.target).closest(
            ".monitor-setor-popup, .mapa-ponto-monitor, #mapaLiveOverlay"
          ).length
        ) {
          fechar();
        }
      });
      docClickBound = true;
    }

    return inst;
  }

  function init(mapContainerSelector, options) {
    ensurePopup($(mapContainerSelector), options || {});
  }

  function normalizar(raw) {
    var p = mapaNormalizarSetor(raw);
    return {
      idSetor: p.idSetor,
      label: p.label || raw.label || "",
      setorNome: raw.setornome || raw.setorNome || raw.label || p.label || "",
      dispositivoNome: raw.dispositivoNome || "",
      idDispositivo: raw.idDispositivo || raw.id_dispositivo || "",
      idFranqueado: raw.idFranqueado || raw.id_franqueado || "",
      clienteNome: raw.clienteNome || "",
      descricao: raw.descricao || "",
      camera: raw.camera || "",
      numero: raw.numero || "",
      particao: raw.particao || "",
      tipoSetor: p.tipoSetor || raw.tipoSetor || "",
      posX: p.posX,
      posY: p.posY,
    };
  }

  function statusDoElemento($el) {
    if ($el.hasClass("mapa-ponto-alarme"))
      return { texto: "Disparo", cls: "status-alarme" };
    if ($el.hasClass("mapa-ponto-falha"))
      return { texto: "Sem comunicação", cls: "status-falha" };
    return { texto: "Normal", cls: "status-normal" };
  }

  function fecharInstancia(inst) {
    if (!inst || !inst.$popup) return;
    inst.$popup.addClass("hidden");
    inst.$popup.removeData("setorAtual");
    inst.$popup.removeData("processoAtual");
  }

  function abrir(dados, $el, ev, inst) {
    if (!inst || !inst.$popup || !inst.$container) return;

    var $popup = inst.$popup;
    $popup.data("setorAtual", dados);

    var st = statusDoElemento($el);
    var titulo = dados.label || dados.setorNome || dados.idSetor || "Setor";

    $popup.find(".monitor-popup-titulo").text(titulo);
    $popup
      .find(".monitor-popup-status")
      .text(st.texto)
      .removeClass("status-normal status-alarme status-falha")
      .addClass(st.cls);
    $popup.find(".monitor-popup-setor").text(dados.setorNome || titulo);
    $popup.find(".monitor-popup-dispositivo").text(
      dados.dispositivoNome
        ? dados.dispositivoNome +
            (dados.idDispositivo ? " (" + dados.idDispositivo + ")" : "")
        : dados.idDispositivo || "—"
    );
    $popup.find(".monitor-popup-cliente").text(dados.clienteNome || "—");
    $popup.find(".monitor-popup-descricao").text(dados.descricao || "—");

    var temCamera =
      typeof MapaCameraLive !== "undefined" &&
      MapaCameraLive.setorTemCamera(dados);
    $popup.find(".monitor-popup-live").toggleClass("hidden", !temCamera);

    var proc = null;
    if (typeof inst.resolverProcesso === "function") {
      proc = inst.resolverProcesso(dados);
    }
    $popup.data("processoAtual", proc || null);
    $popup.find(".monitor-popup-finalizar").toggleClass("hidden", !proc);

    $popup.removeClass("hidden");

    var rect = inst.$container[0].getBoundingClientRect();
    var pw = $popup.outerWidth();
    var ph = $popup.outerHeight();
    var left = ev.clientX - rect.left + 12;
    var top = ev.clientY - rect.top + 12;

    if (left + pw > rect.width - 8) left = rect.width - pw - 8;
    if (top + ph > rect.height - 8) top = rect.height - ph - 8;
    if (left < 8) left = 8;
    if (top < 8) top = 8;

    $popup.css({ left: left + "px", top: top + "px" });

    if (temCamera && inst.abrirCameraAutomatica && typeof MapaCameraLive !== "undefined") {
      MapaCameraLive.abrirAoVivo(dados);
    }
  }

  function fechar(mapContainerSelector) {
    if (mapContainerSelector) {
      var $container = $(mapContainerSelector);
      var key = containerKey($container);
      if (porContainer[key]) fecharInstancia(porContainer[key]);
      return;
    }
    Object.keys(porContainer).forEach(function (key) {
      fecharInstancia(porContainer[key]);
    });
  }

  function vincularPontos($camada, setoresRaw) {
    var $container = $camada.closest(".mapa-container");
    var inst = ensurePopup($container);
    if (!inst) return;

    var map = {};
    (setoresRaw || []).forEach(function (raw) {
      var id = raw.idSetor || raw.id_setor;
      if (id) map[id] = normalizar(raw);
    });

    $camada.find(".mapa-ponto-monitor").off("click.monitorPopup");
    $camada.find(".mapa-ponto-monitor").on("click.monitorPopup", function (e) {
      e.stopPropagation();
      e.preventDefault();
      var id = $(this).attr("data-id-setor");
      if (map[id]) abrir(map[id], $(this), e, inst);
    });
  }

  return {
    init: init,
    ensure: ensurePopup,
    fechar: fechar,
    vincularPontos: vincularPontos,
  };
})();

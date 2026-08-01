// Widget de monitoramento embutido (home, centro operacional, etc.)
var MapaMonitor = (function () {
  function criarInstancia(options) {
    var timer = null;
    var mapaId = null;
    var cfg = options || {};
    cfg.container = options.container;
    cfg.imagem = options.imagem;
    cfg.camada = options.camada;
    cfg.intervalo = options.intervalo || 3000;

    var layoutCtrl = null;
    var layoutInstalado = false;
    var setoresAtuais = [];
    var reqSeq = 0;

    function initLayout() {
      if (layoutCtrl) layoutCtrl.parar();
      layoutCtrl = MapaLayout.instalar({
        container: cfg.container,
        imagem: cfg.imagem,
        camada: cfg.camada,
      });
      layoutInstalado = true;
    }

    function garantirLayout() {
      if (!layoutInstalado) initLayout();
      else if (layoutCtrl) layoutCtrl.atualizar();
    }

    if (cfg.popup !== false && typeof MapaSetorPopup !== "undefined") {
      MapaSetorPopup.init(cfg.container, {
        abrirCameraAutomatica: !!cfg.abrirCameraAutomatica,
        resolverProcesso: cfg.resolverProcesso,
        onFinalizarProcesso: cfg.onFinalizarProcesso,
        onAbrirCamera: cfg.onAbrirCamera,
      });
    }

    function parar() {
      if (timer) {
        clearInterval(timer);
        timer = null;
      }
      mapaId = null;
      setoresAtuais = [];
      reqSeq++;
    }

    function renderPontos(setores, statusPorSetor) {
      var $camada = $(cfg.camada);
      var $container = $(cfg.container);
      var $img = $(cfg.imagem);
      garantirLayout();
      $camada.empty();
      setoresAtuais = setores || [];

      setoresAtuais.forEach(function (raw) {
        var p = mapaNormalizarSetor(raw);
        var st = (statusPorSetor && statusPorSetor[p.idSetor]) || "normal";
        var cls =
          st === "alarme"
            ? "mapa-ponto-alarme"
            : st === "falha"
            ? "mapa-ponto-falha"
            : "mapa-ponto-normal";
        var icone = mapaIconeClasse(p);
        var el = $(
          '<div class="mapa-ponto ' +
            cls +
            ' mapa-ponto-monitor">' +
            '<i class="' +
            icone +
            '"></i>' +
            "</div>"
        );
        el.attr("data-id-setor", p.idSetor);
        MapaLayout.aplicarPosicao(el, p.posX, p.posY, $container, $img);
        $camada.append(el);
      });

      if (cfg.popup !== false) {
        MapaSetorPopup.vincularPontos($camada, setoresAtuais);
      }

      if (layoutCtrl) layoutCtrl.atualizar();
      if (cfg.onPontosRenderizados) cfg.onPontosRenderizados(setoresAtuais);
    }

    function atualizarStatus() {
      if (!mapaId) return;
      $.ajax({
        url: "/monitor/status",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({ mapa_ambiente_id: mapaId }),
      }).done(function (r) {
        var setores = r.setores || (r.dados && r.dados.setores) || [];
        var map = {};
        setores.forEach(function (s) {
          map[s.idSetor || s.id_setor] = s.status || "normal";
        });
        $(cfg.camada + " .mapa-ponto-monitor").each(function () {
          var id = $(this).attr("data-id-setor");
          var st = map[id] || "normal";
          $(this)
            .removeClass("mapa-ponto-normal mapa-ponto-alarme mapa-ponto-falha")
            .addClass(
              st === "alarme"
                ? "mapa-ponto-alarme"
                : st === "falha"
                ? "mapa-ponto-falha"
                : "mapa-ponto-normal"
            );
        });
        if (cfg.onStatus) cfg.onStatus(map, setores);
      });
    }

    function aposImagemPronta($img, fn) {
      if (!$img.length) {
        fn();
        return;
      }
      var feito = false;
      function once() {
        if (feito) return;
        feito = true;
        fn();
      }
      var el = $img[0];
      if (el.complete && el.naturalWidth > 0) {
        setTimeout(once, 10);
        return;
      }
      $img.one("load.mapaMonitorReady error.mapaMonitorReady", once);
      setTimeout(once, 2500);
    }

    function definirSrcImagem($img, url) {
      if (!url) return;
      if ($img.attr("src") === url) {
        $img.attr("src", "");
      }
      $img.attr("src", url).show();
    }

    function finalizarAbertura(mapa, setores) {
      renderPontos(setores, {});
      if (layoutCtrl) layoutCtrl.atualizar();
      atualizarStatus();
      timer = setInterval(atualizarStatus, cfg.intervalo);
      if (cfg.onCarregado) cfg.onCarregado(mapa, setores);
    }

    function abrir(id) {
      parar();
      mapaId = parseInt(id, 10);
      if (!mapaId) return $.Deferred().reject().promise();

      var seq = reqSeq;
      var $container = $(cfg.container);
      var $img = $(cfg.imagem);
      $container.removeClass("mapa-sem-imagem");

      return $.ajax({
        url: "/monitor/carregar",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({ mapa_ambiente_id: mapaId }),
      })
        .done(function (r) {
          if (seq !== reqSeq || mapaId !== parseInt(id, 10)) return;

          var dados = r.dados || r;
          var mapa = dados.mapa || {};
          var setores = dados.setores || [];

          function concluir() {
            if (seq !== reqSeq) return;
            finalizarAbertura(mapa, setores);
          }

          if (mapa.imagem_url) {
            definirSrcImagem($img, mapa.imagem_url);
            aposImagemPronta($img, concluir);
          } else {
            $img.hide().attr("src", "");
            $container.addClass("mapa-sem-imagem");
            concluir();
          }
        })
        .fail(function () {
          if (seq !== reqSeq) return;
          $container.addClass("mapa-sem-imagem");
        });
    }

    return {
      abrir: abrir,
      parar: parar,
      atualizarStatus: atualizarStatus,
      redimensionar: function () {
        garantirLayout();
      },
      reinstalarLayout: function () {
        layoutInstalado = false;
        initLayout();
      },
      getMapaId: function () {
        return mapaId;
      },
      getSetores: function () {
        return setoresAtuais;
      },
    };
  }

  // Singleton legado (home)
  var singleton = null;

  function init(options) {
    singleton = criarInstancia(options);
    MapaSetorPopup.init(options.container);
    if (typeof MapaCameraLive !== "undefined") MapaCameraLive.init();
  }

  function parar() {
    if (singleton) singleton.parar();
    MapaSetorPopup.fechar();
    if (typeof MapaCameraLive !== "undefined") MapaCameraLive.fechar();
  }

  function abrir(id) {
    if (!singleton) return $.Deferred().reject().promise();
    return singleton.abrir(id);
  }

  function redimensionar() {
    if (singleton) singleton.redimensionar();
  }

  function reinstalarLayout() {
    if (singleton && singleton.reinstalarLayout) singleton.reinstalarLayout();
  }

  function atualizarStatus() {
    if (singleton) singleton.atualizarStatus();
  }

  return {
    init: init,
    criar: criarInstancia,
    abrir: abrir,
    parar: parar,
    atualizarStatus: atualizarStatus,
    redimensionar: redimensionar,
    reinstalarLayout: reinstalarLayout,
  };
})();

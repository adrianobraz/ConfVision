// Coordenadas dos setores relativas à área visível da planta (object-fit: contain)
var MapaLayout = (function () {
  var observers = {};

  function dimensoesValidas($container) {
    var el = $container && $container[0];
    return !!(el && el.clientWidth >= 2 && el.clientHeight >= 2);
  }

  function obterArea($container, $img) {
    var el = $container[0];
    var cw = el.clientWidth || 0;
    var ch = el.clientHeight || 0;
    if (cw < 2 || ch < 2) return null;

    if (!$img.length || !$img.is(":visible")) {
      return {
        offsetX: 0,
        offsetY: 0,
        width: cw,
        height: ch,
        containerWidth: cw,
        containerHeight: ch,
      };
    }

    var iw = $img[0].naturalWidth;
    var ih = $img[0].naturalHeight;
    if (!iw || !ih) {
      return {
        offsetX: 0,
        offsetY: 0,
        width: cw,
        height: ch,
        containerWidth: cw,
        containerHeight: ch,
      };
    }

    var scale = Math.min(cw / iw, ch / ih);
    var dw = iw * scale;
    var dh = ih * scale;
    return {
      offsetX: (cw - dw) / 2,
      offsetY: (ch - dh) / 2,
      width: dw,
      height: dh,
      containerWidth: cw,
      containerHeight: ch,
    };
  }

  function clienteParaPct(clientX, clientY, $container, $img) {
    var rect = $container[0].getBoundingClientRect();
    var area = obterArea($container, $img);
    if (!area) return { posX: 0, posY: 0 };
    var x = clientX - rect.left - area.offsetX;
    var y = clientY - rect.top - area.offsetY;
    return {
      posX: clampPct((x / area.width) * 100),
      posY: clampPct((y / area.height) * 100),
    };
  }

  function pctParaCss(posX, posY, $container, $img) {
    var area = obterArea($container, $img);
    if (!area) return null;
    var px = area.offsetX + (posX / 100) * area.width;
    var py = area.offsetY + (posY / 100) * area.height;
    return {
      left: (px / area.containerWidth) * 100 + "%",
      top: (py / area.containerHeight) * 100 + "%",
    };
  }

  function clampPct(v) {
    return Math.round(Math.max(0, Math.min(100, v)) * 100) / 100;
  }

  function aplicarPosicao($el, posX, posY, $container, $img) {
    $el.attr("data-pos-x", posX);
    $el.attr("data-pos-y", posY);
    var css = pctParaCss(posX, posY, $container, $img);
    if (!css) return;
    $el.css(css);
  }

  function reposicionarCamada($camada, $container, $img) {
    $camada.find(".mapa-ponto").each(function () {
      var $p = $(this);
      var posX = parseFloat($p.attr("data-pos-x"));
      var posY = parseFloat($p.attr("data-pos-y"));
      if (isNaN(posX) || isNaN(posY)) return;
      aplicarPosicao($p, posX, posY, $container, $img);
    });
  }

  function instalar(opts) {
    var key = opts.container;
    if (observers[key]) {
      observers[key].disconnect();
    }

    var $container = $(opts.container);
    var $img = $(opts.imagem);
    var $camada = $(opts.camada);

    function atualizar() {
      if (!dimensoesValidas($container)) return;
      if ($img.length && $img.is(":visible") && (!$img[0].naturalWidth || !$img[0].naturalHeight)) {
        return;
      }
      reposicionarCamada($camada, $container, $img);
      if (opts.onResize) opts.onResize();
    }

    $img.off("load.mapaLayout").on("load.mapaLayout", atualizar);

    if (typeof ResizeObserver !== "undefined") {
      var ro = new ResizeObserver(function () {
        atualizar();
      });
      ro.observe($container[0]);
      observers[key] = ro;
    } else {
      $(window).on("resize.mapaLayout" + key.replace(/#/g, ""), atualizar);
    }

    atualizar();
    return { atualizar: atualizar, parar: function () {
      if (observers[key]) observers[key].disconnect();
      $img.off("load.mapaLayout");
    }};
  }

  return {
    dimensoesValidas: dimensoesValidas,
    obterArea: obterArea,
    clienteParaPct: clienteParaPct,
    pctParaCss: pctParaCss,
    aplicarPosicao: aplicarPosicao,
    reposicionarCamada: reposicionarCamada,
    instalar: instalar,
    clampPct: clampPct,
  };
})();

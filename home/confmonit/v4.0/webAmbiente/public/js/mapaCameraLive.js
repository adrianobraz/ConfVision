// Player ao vivo ConfVision ao clicar em setor com camera=S no mapa
var MapaCameraLive = (function () {
  var hlsBase = "";
  var hlsInstance = null;
  var $overlay = null;
  var configCarregada = false;

  function ensureModal() {
    if ($("#mapaLiveOverlay").length) {
      $overlay = $("#mapaLiveOverlay");
      return;
    }

    $overlay = $(`
      <div id="mapaLiveOverlay" class="mapa-live-overlay hidden">
        <div class="mapa-live-dialog">
          <div class="mapa-live-header">
            <h4 id="mapaLiveTitulo" class="mapa-live-titulo">Ao vivo</h4>
            <button type="button" class="mapa-live-fechar" title="Fechar">&times;</button>
          </div>
          <div id="mapaLiveErro" class="mapa-live-erro hidden"></div>
          <video id="mapaLiveVideo" class="mapa-live-video" controls playsinline autoplay muted></video>
        </div>
      </div>
    `);

    $("body").append($overlay);
    $overlay.find(".mapa-live-fechar").on("click", fechar);
    $overlay.on("click", function (e) {
      if ($(e.target).is("#mapaLiveOverlay")) fechar();
    });
  }

  function carregarConfig(cb) {
    if (configCarregada) {
      if (cb) cb();
      return;
    }
    $.ajax({
      url: "/monitor/confvision/config",
      method: "GET",
    })
      .always(function () {
        configCarregada = true;
        if (cb) cb();
      })
      .done(function (r) {
        var dados = r.dados || r;
        hlsBase = String(dados.hlsBase || "").replace(/\/$/, "");
      });
  }

  function hlsUrl(cameraId, urlServidor) {
    if (urlServidor) return urlServidor;
    if (!cameraId || !hlsBase) return "";
    return hlsBase + "/live/" + cameraId + "/index.m3u8";
  }

  function pararPlayer() {
    if (hlsInstance) {
      hlsInstance.destroy();
      hlsInstance = null;
    }
    var video = document.getElementById("mapaLiveVideo");
    if (video) {
      video.removeAttribute("src");
      video.load();
    }
    $("#mapaLiveErro").addClass("hidden").text("");
  }

  function mostrarErro(msg) {
    $("#mapaLiveErro").removeClass("hidden").text(msg || "Erro ao reproduzir ao vivo.");
  }

  function iniciarPlayer(camera) {
    var cameraId = camera && camera.id;
    var url = hlsUrl(cameraId, camera && camera.hlsUrl);
    if (!url) {
      mostrarErro("Ao vivo indisponível (HLS não configurado).");
      return;
    }

    var titulo = "Ao vivo";
    if (camera.nome) titulo += " — " + camera.nome;
    else if (cameraId) titulo += " — câmera " + cameraId;
    $("#mapaLiveTitulo").text(titulo);

    var video = document.getElementById("mapaLiveVideo");
    pararPlayer();
    $overlay.removeClass("hidden");

    if (window.Hls && Hls.isSupported()) {
      hlsInstance = new Hls({ lowLatencyMode: true });
      hlsInstance.loadSource(url);
      hlsInstance.attachMedia(video);
      hlsInstance.on(Hls.Events.MANIFEST_PARSED, function () {
        video.play().catch(function () {});
      });
      hlsInstance.on(Hls.Events.ERROR, function (_, data) {
        if (data.fatal) {
          mostrarErro(
            "Não foi possível reproduzir ao vivo. Verifique se o stream HLS está ativo."
          );
        }
      });
      return;
    }

    if (video.canPlayType("application/vnd.apple.mpegurl")) {
      video.src = url;
      video.onerror = function () {
        mostrarErro("Erro ao carregar stream ao vivo.");
      };
      video.play().catch(function () {
        mostrarErro("Não foi possível iniciar reprodução ao vivo.");
      });
      return;
    }

    mostrarErro("Seu navegador não suporta HLS.");
  }

  function setorTemCamera(dados) {
    return String(dados && dados.camera || "")
      .trim()
      .toUpperCase() === "S";
  }

  function abrirAoVivo(dados) {
    if (!setorTemCamera(dados)) return;

    ensureModal();
    carregarConfig(function () {
      $.ajax({
        url: "/monitor/camera",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({
          id_setor: dados.idSetor,
          id_dispositivo: dados.idDispositivo,
          id_franqueado: dados.idFranqueado || "",
        }),
      })
        .fail(function (xhr) {
          var msg =
            (xhr.responseJSON && xhr.responseJSON.status) ||
            "Não foi possível localizar a câmera ao vivo.";
          boxErro(msg);
        })
        .done(function (r) {
          if (r.status === "Vazio" || !r.dados || !r.dados.camera) {
            boxErro("Câmera ConfVision não encontrada para este setor.");
            return;
          }
          if (r.dados.hlsBase) {
            hlsBase = String(r.dados.hlsBase).replace(/\/$/, "");
          }
          iniciarPlayer(r.dados.camera);
        });
    });
  }

  function fechar() {
    pararPlayer();
    if ($overlay) $overlay.addClass("hidden");
  }

  function init() {
    ensureModal();
    carregarConfig();
  }

  return {
    init: init,
    abrirAoVivo: abrirAoVivo,
    setorTemCamera: setorTemCamera,
    fechar: fechar,
  };
})();

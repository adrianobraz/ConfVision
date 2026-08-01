const params = new URLSearchParams(window.location.search);
const mapaAmbienteId = params.get("mapa_ambiente_id") || params.get("id_mapa");
const nomeMapa = params.get("nome") || "Mapa";
const idCliente = params.get("idCliente") || params.get("id_cliente");
const nomeCliente = params.get("cliente") || "";
let timerStatus = null;
let layoutCtrl = null;
let setoresAtuais = [];
const INTERVALO_MS = 3000;

$(window).on("load", function () {
  if (!mapaAmbienteId) {
    window.location = "/home";
    return;
  }

  $("#nomeMapa").text(decodeURIComponent(nomeMapa));
  if (nomeCliente) {
    $("#nomeCliente").text(" — " + decodeURIComponent(nomeCliente));
  }
  if (idCliente) {
    $("#linkVoltarMapas").attr(
      "href",
      `/mapas/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}`
    );
  }

  MapaSetorPopup.init("#mapaContainer");
  if (typeof MapaCameraLive !== "undefined") MapaCameraLive.init();

  layoutCtrl = MapaLayout.instalar({
    container: "#mapaContainer",
    imagem: "#mapaImagem",
    camada: "#mapaCamada",
  });

  carregarMapa();
  timerStatus = setInterval(atualizarStatus, INTERVALO_MS);
});

$(window).on("beforeunload", function () {
  if (timerStatus) clearInterval(timerStatus);
  if (layoutCtrl) layoutCtrl.parar();
  if (typeof MapaCameraLive !== "undefined") MapaCameraLive.fechar();
});

function carregarMapa() {
  $.ajax({
    url: "/monitor/carregar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ mapa_ambiente_id: parseInt(mapaAmbienteId, 10) }),
  }).done(function (r) {
    const dados = r.dados || r;
    const mapa = dados.mapa || {};
    setoresAtuais = dados.setores || [];
    if (mapa.imagem_url) {
      $("#mapaImagem").attr("src", mapa.imagem_url).show();
      $("#mapaContainer").removeClass("mapa-sem-imagem");
    } else {
      $("#mapaImagem").hide();
      $("#mapaContainer").addClass("mapa-sem-imagem");
    }
    renderPontos(setoresAtuais, {});
    atualizarStatus();
  });
}

function renderPontos(setores, statusPorSetor) {
  const $container = $("#mapaContainer");
  const $img = $("#mapaImagem");
  const $camada = $("#mapaCamada");
  $camada.empty();

  setores.forEach(function (raw) {
    const p = mapaNormalizarSetor(raw);
    const st = statusPorSetor[p.idSetor] || "normal";
    const cls =
      st === "alarme"
        ? "mapa-ponto-alarme"
        : st === "falha"
        ? "mapa-ponto-falha"
        : "mapa-ponto-normal";
    const icone = mapaIconeClasse(p.tipoSetor);
    const el = $(`
      <div class="mapa-ponto ${cls} mapa-ponto-monitor">
        <i class="${icone}"></i>
      </div>
    `);
    el.attr("data-id-setor", p.idSetor);
    MapaLayout.aplicarPosicao(el, p.posX, p.posY, $container, $img);
    $camada.append(el);
  });

  MapaSetorPopup.vincularPontos($camada, setores);

  if (layoutCtrl) layoutCtrl.atualizar();
}

function atualizarStatus() {
  $.ajax({
    url: "/monitor/status",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ mapa_ambiente_id: parseInt(mapaAmbienteId, 10) }),
  }).done(function (r) {
    const setores = r.setores || r.dados?.setores || [];
    const map = {};
    setores.forEach(function (s) {
      const id = s.idSetor || s.id_setor;
      map[id] = s.status || "normal";
    });
    $(".mapa-ponto-monitor").each(function () {
      const id = $(this).attr("data-id-setor");
      const st = map[id] || "normal";
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
  });
}

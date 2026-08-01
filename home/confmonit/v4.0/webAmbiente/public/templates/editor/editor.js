const params = new URLSearchParams(window.location.search);
const mapaAmbienteId = params.get("mapa_ambiente_id") || params.get("id_mapa");
const idCliente = params.get("idCliente") || params.get("id_cliente");
const nomeCliente = params.get("nome") || params.get("cliente") || "";
let posicoes = [];
let arrastando = null;
let popupIdx = null;
let dragStartX = 0;
let dragStartY = 0;
let dragMoved = false;
let layoutCtrl = null;

$(window).on("load", function () {
  if (!mapaAmbienteId || !idCliente) {
    boxErro("Parâmetros inválidos. Volte pela lista de mapas.");
    return;
  }

  $("#linkVoltar").attr(
    "href",
    `/mapas/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}`
  );

  $("#btnSalvar").on("click", salvarPosicoes);
  $("#dispositivo").on("change", carregarSetores);
  $("#btnFecharPopup").on("click", fecharPopup);
  $("#btnExcluirSetorMapa").on("click", excluirSetorDoMapa);
  $(document).on("click", function (e) {
    if (
      !$(e.target).closest(
        "#setorPopup, .mapa-ponto-editor, .editor-icone-opcao"
      ).length
    ) {
      fecharPopup();
    }
  });

  carregarDispositivos();

  layoutCtrl = MapaLayout.instalar({
    container: "#mapaContainer",
    imagem: "#mapaImagem",
    camada: "#mapaCamada",
  });

  carregarMapa();
});

function carregarDispositivos() {
  $.ajax({
    url: "/editor/dispositivos",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ idCliente: idCliente }),
  }).done(function (r) {
    $("#dispositivo").empty();
    if (r.status === "Vazio" || !r.dados) {
      $("#dispositivo").append('<option value="">Sem dispositivos</option>');
      return;
    }
    r.dados.forEach(function (d) {
      const label = d.nome || d.conta || d.idDispositivo;
      $("#dispositivo").append(
        `<option value="${d.idDispositivo}" data-nome="${escAttr(label)}">${label}</option>`
      );
    });
    carregarSetores();
  });
}

function nomeDispositivoAtual() {
  const opt = $("#dispositivo option:selected");
  return opt.data("nome") || opt.text() || "";
}

function strCampo(v) {
  return v == null || v === undefined ? "" : String(v);
}

function normalizarPosicao(p) {
  var tipo = strCampo(p.tipo_setor || p.tipoSetor);
  var icone = strCampo(p.icone);
  if (!icone && typeof mapaIconeSugerido === "function") {
    icone = mapaIconeSugerido(tipo);
  }
  return {
    id: parseInt(p.id, 10) || 0,
    id_setor: strCampo(p.id_setor),
    id_dispositivo: strCampo(p.id_dispositivo),
    label: strCampo(p.label || p.setor_nome),
    setor_nome: strCampo(p.setor_nome || p.label),
    numero: strCampo(p.numero),
    tipo_setor: tipo,
    icone: icone,
    dispositivo_nome: strCampo(p.dispositivo_nome),
    pos_x: num(p.pos_x),
    pos_y: num(p.pos_y),
  };
}

function carregarSetores() {
  const idDisp = $("#dispositivo").val();
  if (!idDisp) return;
  $.ajax({
    url: "/editor/setores",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ idDispositivo: idDisp }),
  }).done(function (r) {
    $("#listaSetores").empty();
    const dados = r.dados || [];
    let disponiveis = 0;

    dados.forEach(function (s) {
      const id = strCampo(s.idSetor);
      const nome = strCampo(s.nome || s.numero || id);
      const jaNoMapa = posicoes.some((p) => p.id_setor === id);
      if (jaNoMapa) return;

      disponiveis++;
      const icone = mapaIconeClasse({ tipoSetor: s.tipoSetor });
      $("#listaSetores").append(`
        <div class="setor-item setor-item-grid" draggable="true"
          data-id-setor="${escAttr(id)}"
          data-id-dispositivo="${escAttr(idDisp)}"
          data-nome="${escAttr(nome)}"
          data-numero="${escAttr(s.numero || "")}"
          data-tipo-setor="${escAttr(s.tipoSetor || "")}"
          data-dispositivo-nome="${escAttr(nomeDispositivoAtual())}"
          title="${escAttr(nome)}">
          <span class="setor-item-icone"><i class="${icone}"></i></span>
          <span class="setor-item-nome">${escHtml(nome)}</span>
        </div>
      `);
    });

    if (disponiveis === 0) {
      $("#listaSetoresVazia").removeClass("hidden");
    } else {
      $("#listaSetoresVazia").addClass("hidden");
    }

    initDragFromList();
  });
}

function carregarMapa() {
  $.ajax({
    url: "/editor/carregar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ mapa_ambiente_id: parseInt(mapaAmbienteId, 10) }),
  }).done(function (r) {
    const dados = r.dados || r;
    const mapa = dados.mapa || {};
    posicoes = (dados.posicoes || []).map(normalizarPosicao);
    $("#nomeMapa").text(mapaNomeAmbiente(mapa));
    if (mapa.imagem_url) {
      $("#mapaImagem").attr("src", mapa.imagem_url).show();
      $("#mapaContainer").removeClass("mapa-sem-imagem");
    } else {
      $("#mapaImagem").hide();
      $("#mapaContainer").addClass("mapa-sem-imagem");
    }
    renderPosicoes();
    carregarSetores();
  });
}

function initDragFromList() {
  $(".setor-item").on("dragstart", function (e) {
    const $el = $(this);
    e.originalEvent.dataTransfer.setData(
      "text/plain",
      JSON.stringify({
        id_setor: strCampo($el.attr("data-id-setor")),
        id_dispositivo: strCampo($el.attr("data-id-dispositivo")),
        label: strCampo($el.attr("data-nome")),
        setor_nome: strCampo($el.attr("data-nome")),
        numero: strCampo($el.attr("data-numero")),
        tipo_setor: strCampo($el.attr("data-tipo-setor")),
        dispositivo_nome: strCampo($el.attr("data-dispositivo-nome")),
      })
    );
  });
}

$("#mapaContainer").on("dragover", function (e) {
  e.preventDefault();
});

$("#mapaContainer").on("drop", function (e) {
  e.preventDefault();
  fecharPopup();
  try {
    const data = JSON.parse(e.originalEvent.dataTransfer.getData("text/plain"));
    const $container = $("#mapaContainer");
    const $img = $("#mapaImagem");
    const pct = MapaLayout.clienteParaPct(
      e.clientX,
      e.clientY,
      $container,
      $img
    );
    posicoes = posicoes.filter((p) => p.id_setor !== strCampo(data.id_setor));
    posicoes.push(
      normalizarPosicao({
        id: 0,
        id_setor: data.id_setor,
        id_dispositivo: data.id_dispositivo,
        label: data.label,
        setor_nome: data.setor_nome || data.label,
        numero: data.numero || "",
        tipo_setor: data.tipo_setor || "",
        icone:
          typeof mapaIconeSugerido === "function"
            ? mapaIconeSugerido(data.tipo_setor)
            : "",
        dispositivo_nome: data.dispositivo_nome || nomeDispositivoAtual(),
        pos_x: pct.posX,
        pos_y: pct.posY,
      })
    );
    renderPosicoes();
    carregarSetores();
  } catch (err) {
    console.log(err);
  }
});

function renderPosicoes() {
  const $container = $("#mapaContainer");
  const $img = $("#mapaImagem");
  $("#mapaCamada").empty();
  posicoes.forEach(function (p, idx) {
    const icone = mapaIconeClasse({
      icone: p.icone,
      tipoSetor: p.tipo_setor,
    });
    const el = $(`
      <div class="mapa-ponto mapa-ponto-normal mapa-ponto-editor" data-idx="${idx}"
        title="${escAttr(p.label || p.setor_nome || p.id_setor)}">
        <i class="${icone}"></i>
      </div>
    `);
    MapaLayout.aplicarPosicao(el, p.pos_x, p.pos_y, $container, $img);
    el.on("mousedown", function (e) {
      iniciarInteracao(e, idx);
    });
    $("#mapaCamada").append(el);
  });
  if (layoutCtrl) layoutCtrl.atualizar();
}

function iniciarInteracao(e, idx) {
  if (e.button !== 0) return;
  e.preventDefault();
  e.stopPropagation();

  dragMoved = false;
  dragStartX = e.clientX;
  dragStartY = e.clientY;
  arrastando = $(e.currentTarget);
  const $container = $("#mapaContainer");
  const $img = $("#mapaImagem");

  $(document).on("mousemove.mapa", function (ev) {
    if (
      Math.abs(ev.clientX - dragStartX) > 4 ||
      Math.abs(ev.clientY - dragStartY) > 4
    ) {
      dragMoved = true;
      fecharPopup();
    }
    if (!dragMoved) return;

    const pct = MapaLayout.clienteParaPct(
      ev.clientX,
      ev.clientY,
      $container,
      $img
    );
    posicoes[idx].pos_x = pct.posX;
    posicoes[idx].pos_y = pct.posY;
    MapaLayout.aplicarPosicao(arrastando, pct.posX, pct.posY, $container, $img);
  });

  $(document).on("mouseup.mapa", function (ev) {
    $(document).off(".mapa");
    if (!dragMoved) {
      abrirPopup(idx, ev);
    }
    arrastando = null;
  });
}

function abrirPopup(idx, ev) {
  const p = posicoes[idx];
  if (!p) return;

  popupIdx = idx;
  const titulo = p.label || p.setor_nome || p.id_setor;
  $("#popupTitulo").text(titulo);
  $("#popupSetor").text(p.setor_nome || p.label || p.id_setor);
  $("#popupNumero").text(p.numero || "—");
  $("#popupDispositivo").text(
    p.dispositivo_nome || p.id_dispositivo || "—"
  );
  $("#popupPosicao").text(
    "X " + p.pos_x + "% · Y " + p.pos_y + "%"
  );

  mapaRenderIconePicker($("#popupIconeGrid"), p.icone || p.tipo_setor, function (
    slug
  ) {
    if (popupIdx === null || popupIdx < 0) return;
    posicoes[popupIdx].icone = slug;
    renderPosicoes();
  });

  const popup = $("#setorPopup");
  popup.removeClass("hidden");

  const container = $("#mapaContainer");
  const rect = container[0].getBoundingClientRect();
  const pw = popup.outerWidth();
  const ph = popup.outerHeight();
  let left = ev.clientX - rect.left + 12;
  let top = ev.clientY - rect.top + 12;

  if (left + pw > rect.width - 8) left = rect.width - pw - 8;
  if (top + ph > rect.height - 8) top = rect.height - ph - 8;
  if (left < 8) left = 8;
  if (top < 8) top = 8;

  popup.css({ left: left + "px", top: top + "px" });
}

function fecharPopup() {
  popupIdx = null;
  $("#setorPopup").addClass("hidden");
}

function excluirSetorDoMapa() {
  if (popupIdx === null || popupIdx < 0) return;
  posicoes.splice(popupIdx, 1);
  fecharPopup();
  renderPosicoes();
  carregarSetores();
}

function salvarPosicoes() {
  fecharPopup();
  $.ajax({
    url: "/editor/salvar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      mapa_ambiente_id: parseInt(mapaAmbienteId, 10),
      idCliente: idCliente,
      nomeCliente: decodeURIComponent(nomeCliente),
      itens: posicoes.map(function (p) {
        var item = normalizarPosicao(p);
        return {
          id: item.id,
          id_setor: item.id_setor,
          id_dispositivo: item.id_dispositivo,
          label: item.label,
          icone: item.icone,
          pos_x: item.pos_x,
          pos_y: item.pos_y,
        };
      }),
    }),
  })
    .fail(function (xhr) {
      const msg =
        (xhr.responseJSON && xhr.responseJSON.status) ||
        xhr.responseText ||
        "Erro ao salvar. Verifique XANO_API e as APIs no Xano.";
      boxErro(msg);
    })
    .done(function (r) {
      if (r.status !== "OK") {
        boxErro("Erro ao salvar posições.");
        return;
      }
      boxSucesso("Posições salvas.");
      carregarMapa();
    });
}

function escAttr(s) {
  return String(s || "")
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;");
}

function escHtml(s) {
  return $("<span>").text(s || "").html();
}

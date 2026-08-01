const params = new URLSearchParams(window.location.search);
const idCliente = params.get("idCliente") || params.get("id_cliente");
const nomeCliente = params.get("nome") || params.get("cliente") || "";
const idFranqueado =
  params.get("idFranqueado") ||
  sessionStorage.getItem("login_idFranqueadoSelecionado") ||
  "";
const idMapaEdit = params.get("mapa_ambiente_id") || params.get("id_mapa") || "0";
const ehMaster = sessionStorage.getItem("login_userMaster") === "S";

let modoPlanta = "url";
let arquivoPendente = null;
let previewObjectUrl = "";
let editorIniciado = false;

$(window).on("load", function () {
  if (!ehMaster) {
    $("#msgMaster").removeClass("hidden");
    $("input, button").prop("disabled", true);
    return;
  }

  if (!idCliente) {
    window.location = "/home";
    return;
  }

  if (idFranqueado) {
    sessionStorage.setItem("login_idFranqueadoSelecionado", idFranqueado);
  }

  $("#idCliente").val(idCliente);
  $("#idFranqueado").val(idFranqueado);
  $("#nomeCliente").val(decodeURIComponent(nomeCliente));
  $("#idMapa").val(idMapaEdit);
  const fraQ = idFranqueado ? `&idFranqueado=${encodeURIComponent(idFranqueado)}` : "";
  $("#linkVoltar").attr(
    "href",
    `/mapas/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}${fraQ}`
  );

  $("#btnModoUrl").on("click", function () {
    if (!podeTrocarParaUrl()) return;
    setModoPlanta("url");
  });
  $("#btnModoArquivo").on("click", function () {
    if (!podeTrocarModo("arquivo")) return;
    setModoPlanta("arquivo");
  });
  $("#btnModoDesenhar").on("click", function () {
    if (!podeTrocarModo("desenhar")) return;
    setModoPlanta("desenhar");
  });
  $("#btnLimparPlanta").on("click", limparPlanta);
  $("#imagem_url").on("input", atualizarPreviewUrl);
  $("#arquivoMapa").on("change", selecionarArquivo);

  $(".planta-tool").on("click", function () {
    var t = $(this).data("tool");
    $(".planta-tool").removeClass("planta-tool-ativo");
    $(this).addClass("planta-tool-ativo");
    PlantaEditor.setTool(t);
  });
  $("#btnPlantaDesfazer").on("click", function () {
    PlantaEditor.undo();
  });
  $("#btnPlantaApagar").on("click", function () {
    PlantaEditor.removeSelected();
  });
  $("#btnUsarDesenho").on("click", usarDesenho);

  if (parseInt(idMapaEdit, 10) > 0) {
    $("#tituloForm").text("Editar mapa");
    $("#btnExcluir").removeClass("hidden");
    carregarMapa();
  } else {
    syncLayoutPlanta();
  }

  $("#btnSalvar").on("click", salvar);
  $("#btnExcluir").on("click", excluir);
});

function isUrlHospedadaContabo(url) {
  if (!url) return false;
  return /contabostorage\.com/i.test(url) && /\/mapas\//i.test(url);
}

function temImagemAplicada() {
  if (arquivoPendente) return true;
  if (modoPlanta === "arquivo" || modoPlanta === "desenhar") {
    return isUrlHospedadaContabo($("#imagem_url").val().trim());
  }
  return false;
}

function podeTrocarParaUrl() {
  if (temImagemAplicada()) {
    boxErro("Clique no ícone de excluir na imagem para usar URL.");
    return false;
  }
  if (modoPlanta === "desenhar" && PlantaEditor.hasContent()) {
    boxErro("Limpe o desenho ou use «Usar este desenho» antes de trocar de aba.");
    return false;
  }
  return true;
}

function podeTrocarModo(novoModo) {
  if (temImagemAplicada()) {
    boxErro("Clique no ícone de excluir na imagem para trocar o modo.");
    return false;
  }
  if (modoPlanta === "desenhar" && novoModo !== "desenhar" && PlantaEditor.hasContent()) {
    boxErro("Limpe o desenho ou use «Usar este desenho» antes de trocar de aba.");
    return false;
  }
  return true;
}

function initEditorSeNecessario() {
  if (editorIniciado) return;
  PlantaEditor.init("#plantaCanvas", {
    onChange: function () {},
    onHint: function (msg) {
      const el = $("#plantaFeedback");
      el.removeClass("hidden err").addClass("ok").text(msg);
      clearTimeout(initEditorSeNecessario._t);
      initEditorSeNecessario._t = setTimeout(function () {
        el.addClass("hidden");
      }, 3500);
    },
  });
  editorIniciado = true;
}

function setModoPlanta(modo) {
  modoPlanta = modo;
  $("#btnModoUrl").toggleClass("ambiente-tab-ativo", modo === "url");
  $("#btnModoArquivo").toggleClass("ambiente-tab-ativo", modo === "arquivo");
  $("#btnModoDesenhar").toggleClass("ambiente-tab-ativo", modo === "desenhar");
  $(".app-main").toggleClass("app-main-narrow", modo !== "desenhar");
  if (modo === "desenhar") initEditorSeNecessario();
  syncLayoutPlanta();
}

function syncLayoutPlanta() {
  const urlAtivo = modoPlanta === "url";
  const arquivoAtivo = modoPlanta === "arquivo";
  const desenharAtivo = modoPlanta === "desenhar";
  const imagemAplicada = temImagemAplicada();
  const urlPreview = urlAtivo && $("#imagem_url").val().trim();

  $("#blocoUrl").toggleClass("hidden", !urlAtivo);
  $("#blocoArquivo").toggleClass("hidden", !arquivoAtivo || imagemAplicada);
  $("#blocoDesenhar").toggleClass("hidden", !desenharAtivo || imagemAplicada);
  $("#btnLimparPlanta").toggleClass("hidden", !imagemAplicada);
  $("#previewPlanta").toggleClass("hidden", !imagemAplicada && !urlPreview);
}

function limparPlanta() {
  arquivoPendente = null;
  $("#arquivoMapa").val("");
  liberarPreviewObjectUrl();
  statusUpload("");
  statusDesenho("");
  $("#imagem_url").val("");
  if (editorIniciado) PlantaEditor.clear();
  mostrarPreview("");
  setModoPlanta("url");
}

function liberarPreviewObjectUrl() {
  if (previewObjectUrl) {
    URL.revokeObjectURL(previewObjectUrl);
    previewObjectUrl = "";
  }
}

function mostrarPreview(url) {
  if (!url) {
    $("#previewImagem").attr("src", "");
    syncLayoutPlanta();
    return;
  }
  $("#previewImagem").attr("src", url);
  syncLayoutPlanta();
}

function atualizarPreviewUrl() {
  if (modoPlanta !== "url") return;
  mostrarPreview($("#imagem_url").val().trim());
}

function statusUpload(msg, tipo) {
  const el = $("#uploadStatus");
  if (!msg) {
    el.addClass("hidden").removeClass("ok err").text("");
    return;
  }
  el.removeClass("hidden ok err");
  if (tipo) el.addClass(tipo);
  el.text(msg);
}

function statusDesenho(msg, tipo) {
  const el = $("#desenhoStatus");
  if (!msg) {
    el.addClass("hidden").removeClass("ok err").text("");
    return;
  }
  el.removeClass("hidden ok err");
  if (tipo) el.addClass(tipo);
  el.text(msg);
}

function selecionarArquivo() {
  const input = document.getElementById("arquivoMapa");
  const file = input.files && input.files[0];
  if (!file) {
    syncLayoutPlanta();
    return;
  }

  if (!file.type.startsWith("image/")) {
    boxErro("Selecione um arquivo de imagem.");
    input.value = "";
    return;
  }

  arquivoPendente = file;
  $("#imagem_url").val("");
  liberarPreviewObjectUrl();
  previewObjectUrl = URL.createObjectURL(file);
  mostrarPreview(previewObjectUrl);
  statusUpload(
    file.name + " — será enviada ao salvar (otimizada para até 150 KB).",
    "ok"
  );
}

function usarDesenho() {
  if (!PlantaEditor.hasContent()) {
    boxErro("Desenhe pelo menos uma parede ou área antes de continuar.");
    return;
  }
  PlantaEditor.exportBlob(function (blob) {
    if (!blob) {
      boxErro("Erro ao exportar desenho.");
      return;
    }
    arquivoPendente = new File([blob], "planta-desenhada.png", { type: "image/png" });
    $("#imagem_url").val("");
    liberarPreviewObjectUrl();
    previewObjectUrl = URL.createObjectURL(blob);
    mostrarPreview(previewObjectUrl);
    statusDesenho("Desenho pronto — será enviado ao salvar (otimizado para até 150 KB).", "ok");
  });
}

function carregarMapa() {
  $.ajax({
    url: "/ambiente/carregar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ id: parseInt($("#idMapa").val(), 10) }),
  }).done(function (r) {
    const m = r.dados || {};
    $("#descricao").val(m.descricao || "");
    $("#imagem_url").val(m.imagem_url || "");
    $("#ordem").val(m.ordem || 0);
    $("#ativo").prop("checked", m.ativo !== false);

    if (m.imagem_url && isUrlHospedadaContabo(m.imagem_url)) {
      setModoPlanta("arquivo");
      mostrarPreview(m.imagem_url);
    } else if (m.imagem_url) {
      setModoPlanta("url");
      mostrarPreview(m.imagem_url);
    } else {
      setModoPlanta("url");
    }

    if (m.idFranqueado && m.idFranqueado !== "CENTRAL") {
      $("#idFranqueado").val(m.idFranqueado);
      sessionStorage.setItem("login_idFranqueadoSelecionado", m.idFranqueado);
    }
  });
}

function payloadSalvar() {
  const usarArquivo = (modoPlanta === "arquivo" || modoPlanta === "desenhar") && arquivoPendente;
  const urlCampo = $("#imagem_url").val().trim();
  let imagem_url;
  if (usarArquivo) {
    imagem_url = "";
  } else if ((modoPlanta === "arquivo" || modoPlanta === "desenhar") && isUrlHospedadaContabo(urlCampo)) {
    imagem_url = urlCampo;
  } else {
    imagem_url = urlCampo;
  }

  return {
    id: parseInt($("#idMapa").val(), 10) || 0,
    descricao: $("#descricao").val().trim(),
    imagem_url: imagem_url,
    idCliente: $("#idCliente").val(),
    idFranqueado: $("#idFranqueado").val() || idFranqueado,
    nomeCliente: $("#nomeCliente").val(),
    ordem: parseInt($("#ordem").val(), 10) || 0,
    ativo: $("#ativo").is(":checked"),
  };
}

function salvar() {
  const dados = payloadSalvar();
  if (!dados.descricao) {
    boxErro("Informe a descrição do mapa.");
    return;
  }
  if (!dados.idCliente) {
    boxErro("Cliente não identificado.");
    return;
  }
  if (!dados.idFranqueado && !idFranqueado) {
    boxErro("Franqueado não identificado.");
    return;
  }

  $("#btnSalvar").prop("disabled", true);

  let ajaxOpts;
  if ((modoPlanta === "arquivo" || modoPlanta === "desenhar") && arquivoPendente) {
    const fd = new FormData();
    fd.append("dados", JSON.stringify(dados));
    fd.append("arquivo", arquivoPendente);
    ajaxOpts = {
      url: "/ambiente/salvar",
      method: "POST",
      data: fd,
      processData: false,
      contentType: false,
    };
  } else {
    ajaxOpts = {
      url: "/ambiente/salvar",
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify(dados),
    };
  }

  $.ajax(ajaxOpts)
    .fail(function (xhr) {
      $("#btnSalvar").prop("disabled", false);
      let msg = "Erro ao salvar mapa.";
      try {
        const j = JSON.parse(xhr.responseText);
        if (j.status) msg = j.status.replace(/^Erro:\s*/, "");
      } catch (e) {}
      boxErro(msg);
    })
    .done(function (r) {
      if (r.status !== "OK") {
        $("#btnSalvar").prop("disabled", false);
        boxErro("Erro ao salvar mapa.");
        return;
      }
      boxSucesso("Mapa salvo.");
      setTimeout(function () {
        window.location = `/mapas/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}`;
      }, 800);
    });
}

function excluir() {
  if (!confirm("Excluir este mapa? Os setores posicionados permanecem no Xano até limpeza manual.")) {
    return;
  }

  $.ajax({
    url: "/ambiente/excluir",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({ id: parseInt($("#idMapa").val(), 10) }),
  })
    .fail(function () {
      boxErro("Erro ao excluir.");
    })
    .done(function () {
      boxSucesso("Mapa excluído.");
      setTimeout(function () {
        window.location = `/mapas/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}`;
      }, 800);
    });
}

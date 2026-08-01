function modEvtDesagrupado_start(idProcesso, codigoFull, zoneUser, retorno) {
  $("#boxDir").empty();
  const uri = "/assets/modulos/modEvtDesagrupado/modEvtDesagrupado.html";
  $("#boxDir").load(uri, () => {
    $("#modEvtDesagrupado_btnFechar").on("click", () => {
      retorno(idProcesso);
    });

    modEvtDesagrupado_buscarDados(idProcesso, codigoFull, zoneUser);
  });
}

function modEvtDesagrupado_buscarDados(idProcesso, codigo, zonaUser) {
  $.ajax({
    url: "/modEvtDesagrupado/buscarDados",
    method: "Post",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      Authorization: "Bearer " + sessionStorage.getItem("token"),
    },
    data: JSON.stringify({
      idProcesso: idProcesso,
      codigo: codigo,
      zonaUser: zonaUser,
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      $("#modEvtDesagrupado_responsivo tbody").empty();
      if (r.status != "Vazio" && r.dados && r.dados.length > 0) {
        modEvtDesagrupado_preencherFixos(r.dados[0]);
        modEvtDesagrupado_configurarBotoesCamera(r.dados[0]);
        r.dados.forEach((item) => {
          modEvtDesagrupado_montarLinha(item);
        });
        $("[tipo=modEvtDesagrupado_btnVisualizarGravacao]")
          .off("click")
          .on("click", modEvtDesagrupado_visualizarGravacaoLinha);
        $("[tipo=modEvtDesagrupado_btnConfVisionCamera]")
          .off("click")
          .on("click", modEvtDesagrupado_abrirConfVisionLinha);
      } else {
        modEvtDesagrupado_preencherFixos({});
        modEvtDesagrupado_configurarBotoesCamera({});
      }
    });
}

function modEvtDesagrupado_ctxFromCell($el) {
  return {
    idProcesso: $el.attr("idProcesso") || "",
    idDispositivo: $el.attr("idDispositivo") || "",
    particao: $el.attr("particao") || "",
    zonaUser: $el.attr("zonaUser") || "",
    idEvento: $el.attr("idEvento") || "",
    img: $el.attr("img") || "",
    provedorVideo: $el.attr("provedorVideo") || "",
    usaConfVision: $el.attr("usaConfVision") || "",
  };
}

function modEvtDesagrupado_temCamera(item) {
  if (item.cameraOn === "S") return true;
  if (item.camera && item.camera !== "SEM IMAGEM") return true;
  return !!confvision_parseImg(item.camera);
}

function modEvtDesagrupado_eConfVision(item) {
  return confvision_eConfVision(item) || !!confvision_parseImg(item.camera);
}

function modEvtDesagrupado_montarLinha(item) {
  const temCamera = modEvtDesagrupado_temCamera(item);
  const isConfVision = modEvtDesagrupado_eConfVision(item);
  const attrs = `
    idProcesso="${item.idProcesso || ""}"
    idDispositivo="${item.idDispositivo || ""}"
    particao="${item.particao || ""}"
    zonaUser="${item.zonaUser || ""}"
    idEvento="${item.idEvento || ""}"
    img="${(item.camera || "").replace(/"/g, "&quot;")}"
    provedorVideo="${item.provedorVideo || ""}"
    usaConfVision="${item.usaConfVision || ""}"
  `;

  let gravacaoCell = "";
  if (!temCamera) {
    gravacaoCell = `
      <button type="button" class="btn btn-sm btn-secondary modEvtDesagrupado-actionBtn" disabled>
        <i class="bi bi-x-lg"></i>
      </button>`;
  } else if (isConfVision) {
    gravacaoCell = `
      <button
        type="button"
        class="btn btn-sm btn-success click modEvtDesagrupado-actionBtn"
        tipo="modEvtDesagrupado_btnConfVisionCamera"
        ${attrs}
        title="ConfVision — eventos, foto e ao vivo"
      ><i class="bi bi-camera-video-fill"></i> Ver</button>`;
  } else {
    const jsonD = (item.camera || "").replaceAll(/"/g, "'");
    gravacaoCell = `
      <button
        type="button"
        class="btn btn-sm btn-success click modEvtDesagrupado-actionBtn"
        tipo="modEvtDesagrupado_btnVisualizarGravacao"
        jsonCamera="${jsonD}"
        vivo="N"
        title="Exibição gravação"
      ><i class="bi bi-camera-video-fill"></i> Gravação</button>`;
  }

  $("#modEvtDesagrupado_responsivo tbody").append(`
      <tr class="table-line modEvtDesagrupado-dataRow">
        <td>${item.entrada || ""}</td>
        <td class="modEvtDesagrupado-gravacaoCell">${gravacaoCell}</td>
      </tr>
  `);
}

function modEvtDesagrupado_preencherFixos(item) {
  $("#modEvtDesagrupado_cpZonaUser").text(item.zonaUser || "");
  $("#modEvtDesagrupado_cpParticao").text(item.particao || "");
  $("#modEvtDesagrupado_cpDescricao").text(item.descricao || "");
  $("#modEvtDesagrupado_cpNomeZonaUser").text(item.nomeZonaUser || "");
}

function modEvtDesagrupado_configurarBotoesCamera(item) {
  const temCamera = modEvtDesagrupado_temCamera(item);
  const isConfVision = modEvtDesagrupado_eConfVision(item);
  const jsonD = temCamera && !isConfVision ? (item.camera || "").replaceAll(/"/g, "'") : "";
  const tempoRealBtn = $("#modEvtDesagrupado_btnTempoReal");

  tempoRealBtn
    .removeAttr("jsonCamera idProcesso idDispositivo particao zonaUser idEvento img provedorVideo usaConfVision tipo")
    .off("click");

  if (isConfVision) {
    const attrs = {
      idProcesso: item.idProcesso || "",
      idDispositivo: item.idDispositivo || "",
      particao: item.particao || "",
      zonaUser: item.zonaUser || "",
      idEvento: item.idEvento || "",
      img: item.camera || "",
      provedorVideo: item.provedorVideo || "",
      usaConfVision: item.usaConfVision || "",
    };
    Object.keys(attrs).forEach(function (k) {
      tempoRealBtn.attr(k, attrs[k]);
    });
    tempoRealBtn
      .attr("tipo", "modEvtDesagrupado_btnConfVisionCamera")
      .html('<i class="bi bi-camera-video-fill"></i> ConfVision')
      .toggleClass("btn-success", temCamera)
      .toggleClass("btn-secondary", !temCamera)
      .prop("disabled", !temCamera)
      .on("click", modEvtDesagrupado_abrirConfVisionTopo);
    return;
  }

  tempoRealBtn.attr("jsonCamera", jsonD);
  tempoRealBtn
    .html('<i class="bi bi-eye-fill"></i> Exibição tempo real')
    .toggleClass("btn-success", temCamera)
    .toggleClass("btn-secondary", !temCamera)
    .prop("disabled", !temCamera)
    .on("click", modEvtDesagrupado_visualizarCameraTempoReal);
}

function modEvtDesagrupado_abrirConfVision(ctx, focusEventoId) {
  confvision_abrirPainelPorSetor(ctx, {
    focusEventoId: focusEventoId || null,
  });
}

function modEvtDesagrupado_abrirConfVisionTopo() {
  modEvtDesagrupado_abrirConfVision(modEvtDesagrupado_ctxFromCell($("#modEvtDesagrupado_btnTempoReal")));
}

function modEvtDesagrupado_abrirConfVisionLinha() {
  var parsed = confvision_parseImg($(this).attr("img"));
  modEvtDesagrupado_abrirConfVision(
    modEvtDesagrupado_ctxFromCell($(this)),
    parsed && parsed.visEventoId ? parsed.visEventoId : null
  );
}

function modEvtDesagrupado_visualizarCameraTempoReal() {
  const jsonCamera = $("#modEvtDesagrupado_btnTempoReal").attr("jsonCamera");
  if (!jsonCamera) return;

  const d = JSON.parse(jsonCamera.replaceAll(/'/g, '"'));
  benuvem_visualizar(d.company_code, d.partition, d.client_code, d.channels);
}

function modEvtDesagrupado_visualizarGravacaoLinha() {
  const jsonCamera = $(this).attr("jsonCamera");
  if (!jsonCamera) return;

  const d = JSON.parse(jsonCamera.replaceAll(/'/g, '"'));
  benuvem_visualizar(d.company_code, d.partition, d.client_code, d.channels, d.date);
}

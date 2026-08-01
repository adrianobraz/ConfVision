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
      if (r.status != "Vazio") {
        r.dados.forEach((item) => {
          modEvtDesagrupado_montarLinha(item);
        });
        $("td[tipo=modEvtDesagrupado_btnVisualizarCamera]").on(
          "click",
          modEvtDesagrupado_visualizarCamera
        );
        $("td[tipo=modEvtDesagrupado_btnConfVisionCamera]").on("click", function () {
          var parsed = confvision_parseImg($(this).attr("img"));
          confvision_abrirPainelPorSetor(modEvtDesagrupado_ctxFromCell($(this)), {
            focusEventoId: (parsed && parsed.visEventoId) ? parsed.visEventoId : null
          });
        });
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

function modEvtDesagrupado_montarLinha(item) {
  const temCamera = modEvtDesagrupado_temCamera(item);
  const isConfVision =
    confvision_eConfVision(item) || !!confvision_parseImg(item.camera);

  let cameraCor = "bg-secondary";
  let cameraOn = "";
  let cameraCells = "";

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

  if (temCamera) {
    cameraCor = "bg-success";

    if (isConfVision) {
      cameraCells = `
        <td colspan="2" class="${cameraCor} click text-center" tipo="modEvtDesagrupado_btnConfVisionCamera" ${attrs} title="ConfVision — eventos, foto e ao vivo">
          <i class="bi bi-camera-video-fill"></i>
        </td>`;
    } else {
      const jsonD = (item.camera || "").replaceAll(/"/g, "'");
      cameraOn = 'tipo="modEvtDesagrupado_btnVisualizarCamera"';
      cameraCells = `
        <td class="${cameraCor} ${cameraOn}" jsonCamera="${jsonD}" vivo="S" title="Exibição tempo real">
          <i class="bi bi-eye-fill"></i>
        </td>
        <td class="${cameraCor} ${cameraOn}" jsonCamera="${jsonD}" vivo="N" title="Exibição gravação">
          <i class="bi bi-camera-video-fill"></i>
        </td>`;
    }
  } else {
    cameraCells = `
      <td colspan="2" class="${cameraCor} text-center"><i class="bi bi-x-lg"></i></td>`;
  }

  $("#modEvtDesagrupado tbody").append(`
        <tr class="table-line">
            ${cameraCells}
            <td>${item.zonaUser}</td>
            <td>${item.particao}</td>
            <td>${item.descricao}</td>
            <td>${item.nomeZonaUser}</td>
            <td>${item.entrada}</td>
        </tr>   
    `);
}

function modEvtDesagrupado_visualizarCamera() {
  const jsonCamera = $(this).attr("jsonCamera");
  const vivo = $(this).attr("vivo");
  const d = JSON.parse(jsonCamera.replaceAll(/'/g, '"'));

  if (vivo == "S") {
    benuvem_visualizar(d.company_code, d.partition, d.client_code, d.channels);
  } else {
    benuvem_visualizar(
      d.company_code,
      d.partition,
      d.client_code,
      d.channels,
      d.date
    );
  }
}

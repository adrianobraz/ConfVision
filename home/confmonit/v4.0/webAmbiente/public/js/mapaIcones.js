// Catálogo de ícones para pontos no mapa (Bootstrap Icons, slug sem prefixo bi-)
var MAPA_ICONES_CATALOGO = [
  { id: "camera-video-fill", label: "Câmera" },
  { id: "bell-fill", label: "Alarme" },
  { id: "lock-fill", label: "Trancado" },
  { id: "door-closed-fill", label: "Campainha" },
  { id: "broadcast-pin", label: "Sensor" },
  { id: "radar", label: "Movimento" },
  { id: "window", label: "Janela" },
  { id: "fire", label: "Incêndio" },
  { id: "lightbulb-fill", label: "Luz" },
  { id: "fingerprint", label: "Digital" },
  { id: "person-badge-fill", label: "Vigilante" },
  { id: "exclamation-octagon-fill", label: "Emergência" },
  { id: "bell", label: "Notificação" },
  { id: "shield-shaded", label: "Cão guarda" },
  { id: "unlock-fill", label: "Destrancado" },
  { id: "sign-stop-fill", label: "Barreira" },
  { id: "display-fill", label: "Monitor" },
  { id: "cloud-fog2-fill", label: "Fumaça" },
  { id: "door-open-fill", label: "Acesso" },
  { id: "phone-fill", label: "App" },
  { id: "shield-check", label: "Proteção" },
  { id: "person-x-fill", label: "Risco" },
  { id: "house-lock-fill", label: "Residência" },
  { id: "key-fill", label: "Chave" },
  { id: "bullseye", label: "Padrão" },
];

function mapaIconeSugerido(tipoSetor) {
  var tipo = String(tipoSetor || "")
    .trim()
    .toUpperCase();
  if (tipo === "CAMERA") return "camera-video-fill";
  if (tipo === "SENSOR") return "broadcast-pin";
  return "bullseye";
}

function mapaIconeCatalogoValido(slug) {
  var id = String(slug || "").trim();
  if (!id) return "";
  if (id.indexOf("bi-") === 0) id = id.slice(3);
  for (var i = 0; i < MAPA_ICONES_CATALOGO.length; i++) {
    if (MAPA_ICONES_CATALOGO[i].id === id) return id;
  }
  return id;
}

function mapaIconeResolver(setor) {
  if (setor && typeof setor === "object") {
    var salvo = mapaIconeCatalogoValido(setor.icone);
    if (salvo) return salvo;
    return mapaIconeSugerido(setor.tipoSetor || setor.tipo_setor);
  }
  return mapaIconeSugerido(setor);
}

function mapaRenderIconePicker($container, iconeAtual, onSelect) {
  $container.empty();
  var atual = mapaIconeCatalogoValido(iconeAtual) || mapaIconeSugerido();

  MAPA_ICONES_CATALOGO.forEach(function (item) {
    var btn = $(
      '<button type="button" class="editor-icone-opcao" title="' +
        escAttrIcone(item.label) +
        '" data-icone="' +
        escAttrIcone(item.id) +
        '"><i class="bi bi-' +
        escAttrIcone(item.id) +
        '"></i></button>'
    );
    if (item.id === atual) btn.addClass("ativo");
    btn.on("click", function (e) {
      e.preventDefault();
      e.stopPropagation();
      var slug = $(this).attr("data-icone");
      $container.find(".editor-icone-opcao").removeClass("ativo");
      $(this).addClass("ativo");
      if (typeof onSelect === "function") onSelect(slug);
    });
    $container.append(btn);
  });
}

function escAttrIcone(s) {
  return String(s || "")
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;");
}

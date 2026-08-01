// Normaliza campos das tabelas Xano mapa_ambiente (#104) e mapa_setor (#105)
function mapaNomeAmbiente(m) {
  return m.descricao || m.nomeCliente || "Mapa";
}

function mapaNormalizarSetor(p) {
  return {
    idSetor: p.idSetor || p.id_setor || "",
    label: p.label || p.setornome || p.setorNome || "",
    posX: num(p.posX ?? p.pos_x),
    posY: num(p.poxY ?? p.posY ?? p.pos_y),
    icone: p.icone || "",
    tipoSetor: p.tipoSetor || p.tipo_setor || "",
    camera: p.camera || "",
  };
}

// Classe CSS do ícone: prioriza icone salvo, depois sugestão por tipoSetor.
function mapaIconeClasse(setorOuTipo) {
  var slug = "";
  if (setorOuTipo && typeof setorOuTipo === "object") {
    if (typeof mapaIconeResolver === "function") {
      slug = mapaIconeResolver(setorOuTipo);
    } else {
      slug = String(setorOuTipo.icone || "").trim();
      if (!slug) {
        var tipoSetor = String(
          setorOuTipo.tipoSetor || setorOuTipo.tipo_setor || ""
        )
          .trim()
          .toUpperCase();
        if (tipoSetor === "CAMERA") slug = "camera-video-fill";
        else slug = "bullseye";
      }
    }
  } else if (typeof mapaIconeSugerido === "function") {
    slug = mapaIconeSugerido(setorOuTipo);
  } else {
    var tipo = String(setorOuTipo || "").trim().toUpperCase();
    slug = tipo === "CAMERA" ? "camera-video-fill" : "bullseye";
  }
  if (slug.indexOf("bi-") === 0) return "bi " + slug;
  return "bi bi-" + slug;
}

function num(v) {
  const n = parseFloat(v);
  return isNaN(n) ? 0 : n;
}

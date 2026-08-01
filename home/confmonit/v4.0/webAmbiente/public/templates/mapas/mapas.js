const params = new URLSearchParams(window.location.search);
const idCliente = params.get("idCliente") || params.get("id_cliente");
const nomeCliente = params.get("nome") || "Cliente";
const idFranqueado =
  params.get("idFranqueado") ||
  sessionStorage.getItem("login_idFranqueadoSelecionado") ||
  "";
const ehMaster = sessionStorage.getItem("login_userMaster") === "S";

function urlComFranqueado(path) {
  if (!idFranqueado) return path;
  const sep = path.includes("?") ? "&" : "?";
  return `${path}${sep}idFranqueado=${encodeURIComponent(idFranqueado)}`;
}

$(window).on("load", function () {
  if (!idCliente) {
    window.location = "/home";
    return;
  }
  if (idFranqueado) {
    sessionStorage.setItem("login_idFranqueadoSelecionado", idFranqueado);
  }
  $("#tituloCliente").text(decodeURIComponent(nomeCliente));

  if (ehMaster) {
    $("#btnNovoMapa")
      .removeClass("hidden")
      .attr(
        "href",
        urlComFranqueado(
          `/ambiente/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}`
        )
      );
  }

  carregarMapas();
});

function carregarMapas() {
  const payload = { idCliente: idCliente };
  if (idFranqueado) payload.idFranqueado = idFranqueado;

  $.ajax({
    url: "/mapas/listar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(payload),
  })
    .fail(function (e) {
      console.log(e);
      $("#listaMapas").html('<p class="msg-error">Erro ao listar mapas.</p>');
    })
    .done(function (r) {
      $("#listaMapas").empty();
      const dados = r.dados || r.items || [];
      if (r.status === "Vazio" || !dados.length) {
        let html = '<p class="msg-empty">Nenhum mapa para este cliente.</p>';
        if (ehMaster) {
          html +=
            `<p style="margin-top:0.75rem;text-align:center;"><a href="${urlComFranqueado(`/ambiente/page?idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}`)}"` +
            ` class="link-accent">Cadastrar primeiro mapa</a></p>`;
        }
        $("#listaMapas").html(html);
        return;
      }
      dados.forEach(function (m) {
        const id = m.id;
        const nome = mapaNomeAmbiente(m);
        const q = `mapa_ambiente_id=${id}&idCliente=${idCliente}&nome=${encodeURIComponent(nomeCliente)}${idFranqueado ? "&idFranqueado=" + encodeURIComponent(idFranqueado) : ""}`;
        const btnMonitor = `
          <a href="/monitor/page?${q}&cliente=${encodeURIComponent(nomeCliente)}"
             class="btn btn-danger btn-sm">Monitor</a>`;
        const btnEditar = ehMaster
          ? `<a href="/editor/page?${q}" class="btn btn-primary btn-sm">Setores</a>
             <a href="/ambiente/page?${q}" class="btn btn-ghost btn-sm">Editar</a>`
          : "";
        $("#listaMapas").append(`
          <div class="card">
            <h3 class="card-title">${nome}</h3>
            <p class="card-meta">${m.descricao || ""}</p>
            <div class="btn-row">${btnMonitor}${btnEditar}</div>
          </div>
        `);
      });
    });
}

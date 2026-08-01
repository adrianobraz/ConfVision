function logar() {
  if ($("#email").val() == "") {
    boxErro("Digite o email");
    return;
  }
  if ($("#senha").val() == "") {
    boxErro("Digite a senha");
    return;
  }

  $.ajax({
    url: "/logar",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      email: $("#email").val(),
      senha: $("#senha").val(),
    }),
  })
    .fail(function (e) {
      const msg =
        (e.responseJSON && e.responseJSON.status) ||
        "Usuario ou senha invalidos";
      boxErro(msg.replace(/^Erro:\s*/i, ""));
      console.log(e);
    })
    .done(function (r) {
      if (r.status !== "OK" || !r.dados) {
        const msg = (r.status || "").replace(/^Erro:\s*/i, "");
        boxErro(msg || "Usuario ou senha invalidos");
        return;
      }
      const d = r.dados;
      sessionStorage.setItem("login_userVinculo", d.idVinculo);
      sessionStorage.setItem("login_userVinculoNome", d.nomeVinculo);
      sessionStorage.setItem("login_userIdOperador", d.idOperador);
      sessionStorage.setItem("login_userNome", d.userNome);
      sessionStorage.setItem("login_userNick", d.userNick);
      sessionStorage.setItem("login_userTipo", d.userTipo);
      sessionStorage.setItem("login_userEmail", d.userEmail);
      sessionStorage.setItem("login_userMaster", d.userMaster || "N");
      if (d.userTipo === "FRA" && d.idVinculo) {
        sessionStorage.setItem("login_idFranqueadoSelecionado", d.idVinculo);
      }
      window.location = "/home";
    });
}

$(window).on("load", function () {
  document.addEventListener("keypress", function (event) {
    if (event.key === "Enter") {
      logar();
    }
  });
});

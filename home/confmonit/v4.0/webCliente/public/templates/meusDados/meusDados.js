$(window).on("load", function () {
  carregarDados();
});

function carregarDados() {
  $.ajax({
    url: `/carregarDados`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r)
      if (r.status == "OK") {
        const d = r.dados;
        $("#nome").val(d.nome);
        $("#nick").val(d.nick);
        $("#cpf").val(formatarDocumento(d.documento1));
        $("#rg").val(d.documento2);
        $("#cep").val(d.cep);
        $("#uf").val(d.uf);
        $("#endereco").val(d.endereco);
        $("#complemento").val(d.complemento);
        $("#bairro").val(d.bairro);
        $("#cidade").val(d.cidade);
        $("#telefone1").val(formatarCelular(d.telefone1));
        $("#telefone2").val(formatarCelular(d.telefone2));
        $("#email1").val(d.email1);
        $("#email2").val(d.email2);
      } else {
        boxErro(r.status);
      }
    });
}

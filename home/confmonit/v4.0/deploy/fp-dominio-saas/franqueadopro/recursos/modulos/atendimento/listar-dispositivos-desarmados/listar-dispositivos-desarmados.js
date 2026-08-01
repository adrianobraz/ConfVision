var fpPagDisp = null

$(document).ready(function () {
  $("#pesquisar-cliente").on("input", function () {
    clearTimeout(window._fpBuscaDisp)
    window._fpBuscaDisp = setTimeout(iniciarLista, 400)
  })
  iniciarLista()
})

function iniciarLista() {
  if (fpPagDisp) fpPagDisp.destroy()
  var termo = ($("#pesquisar-cliente").val() || "").trim()

  fpPagDisp = fpScrollPaginacao({
    url: "/DispositivoListarDesarmado",
    tbodySel: "tbody",
    getPayload: function () {
      return {
        idFranqueado: localStorage.getItem("idFranqueado"),
        armado: "N",
        termo: termo
      }
    },
    renderRows: function (dados) {
      dados.forEach(function (i) {
        var dataUltimoEvento =
          i.dataUltimoEvento == "" ? "NUNCA CONECTOU" : i.dataUltimoEvento
        $("tbody").append(`
          <tr>
            <td>${i.conta}</td>
            <td>${i.nomeCliente}</td>
            <td>${i.nome}</td>
            <td>${i.dataArmado}</td>
            <td>${i.codigoUltimoEvento}</td>
            <td>${dataUltimoEvento}</td>
          </tr>
        `)
      })
    },
    onEmpty: function () {
      $("tbody").append(
        `<tr><td colspan="6" class="text-center">LISTA VAZIA</td></tr>`
      )
    },
    onFail: function () {
      boxErro("Erro ao carregar tabela")
    }
  })

  fpPagDisp.reset()
}

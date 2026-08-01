$(window).on("load", function () {
  sessionStorage.setItem("idProcesso", "0");
  getProcessos();
  $("#descricao").show();
});

function getProcessos() {
  $.ajax({
    url: `/atendimento/listarProcessoToOpenByCliente`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      console.log(r);
      $("tbody").empty();
      if (r.status == "OK") {
        r.dados.map((item) => {
          $("tbody").append(`
            <tr>
              <td style="width: 5%;" title="Pânico" class="p-1 bg-red-700">
                  <i class="bi bi-radioactive"></i>
              </td>       
              
              <td>${item.dispNome}</td>

              
              <td 
                  idProcesso="${item.idProcesso}"
                  tipo="btnAtender" 
                  class="p-1 bg-green-700"
              >
                  <i class="bi bi-headset"></i>
              </td>                  
              
              <td 
                  idDispositivo="${item.dispId}"
                  tipo="btnManutencao" 
                  class="p-1 bg-yellow-700"
              >
                  <i class="bi bi-tools"></i>
              </td>     
                                            
            </tr>
          `);
        });

        $("td[tipo=btnAtender").on("click", function () {
          const id = this.getAttribute("idProcesso");
          atender(id);
        });

        $("td[tipo=btnVisualizar").on("click", function () {
          const id = this.getAttribute("idProcesso");
          visualizar(id);
        });

        $("td[tipo=btnManutencao").on("click", function () {
          const id = this.getAttribute("idDispositivo");
          manutencao(id);
        });
      }
    });
}

function atender(id) {
  sessionStorage.setItem("idProcesso", id);
  window.location = "/atenderProcesso/page";
}

function visualizar() {
  $.ajax({
    url: `/atendimento/visualizar`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r);
    });
}

function manutencao(idDisp, idProc) {
  // boxSucesso("Recurso sera disponibilizado em breve");
  // return;

  $.ajax({
    url: "/atendimento/manutencao",
    method: "Post",
    data: JSON.stringify({
      idDispositivo: idDisp,
      tempo: "30",
      idProcesso: idProc,
      idOperador: sessionStorage.getItem("loginId"),
      nick: sessionStorage.getItem("loginNick"),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      console.log(r);
      msgSucesso();
    });
}

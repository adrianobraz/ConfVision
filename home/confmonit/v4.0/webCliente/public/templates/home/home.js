$(window).on("load", function () {
  //getFranqBloqueado()
  //faturaVencidas()
  carregarDispitivo();
  $("#textoInformativo").html(sessionStorage.getItem("loginInformativo"));
  const cli = sessionStorage.getItem("loginNome");
  $("#boas-vindas").html(`${cli}, bem vindo ao portal`);
  $("#btnGravarSenha").on("click", gravarSenha);
  $("#btnGravarDisp").on("click", gravarDispPanico);
});

function carregarDispitivo() {
  $.ajax({
    url: `/grade/carregarDispositivo`,
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
      $("tbody").empty();
      $("#dispPadrao").empty();

      if (r.status != "Vazio") {
        r.dados.map((i) => {
          let btnCor;
          let btnIcon;
          let btnTxt;
          let acao;

          if (i.armado == "S") {
            btnCor = "bg-green-800";
            btnIcon = '<i class="bi bi-toggle-on">Desarmar</i>';
            btnTxt = "Desarmar";
            acao = "0";
          } else {
            btnCor = "bg-slate-700";
            btnIcon = '<i class="bi bi-toggle-off">Armar</i>';
            btnTxt = "Armar";
            acao = "1";
          }

          $("tbody").append(`
                    <tr>
                        <td>${i.nome}</td>
                        <td 
                            class="${btnCor} btn-click"
                            tipo="armar" 
                            id="${i.idDispositivo}"
                            numero="${i.particao}"
                            acao="${acao}"
                            senha="${i.senha}"
                        >
                        ${btnTxt}
                        </td>
                    </tr>
                `);

          $("#dispPadrao").append(`
                    <option value="${i.idDispositivo}">${i.nome}</option>
                `);
        });

        $("td[tipo=armar]").on("click", function () {
          armarCentral(this);
        });
      }
    });
}

function armarCentral(d) {
  $.ajax({
    start: boxProcessando(),
    url: "/armarDispositivo",
    method: "Post",
    data: JSON.stringify({
      idDispositivo: d.getAttribute("id"),
      numero: Number(d.getAttribute("numero")),
      usuario: "000",
      acao: Number(d.getAttribute("acao")),
      senha: d.getAttribute("senha"),
      senhaWeb: sessionStorage.getItem("loginComandoSenha"),
    }),
  })
    .fail(function (e) {
      console.log(e);
      if (e.responseJSON.status == "Erro: nao conectado") {
        boxErro("Dispositivo esta Offline");
      } else {
        boxErro("Erro ao enviar comando de arme");
      }
    })
    .done(function (r) {
      if (r.status == "OK") {
        setTimeout(aguardar, 3000);
      } else {
        boxErro("Erro ao manipular disopositivo");
      }
    });
}

function aguardar() {
  boxFechar();
  carregarDispitivo();
}

function gravarSenha() {
  const senha = $("#novaSenha").val();
  const confirma = $("#confirmaNovaSenha").val();
  if (senha === confirma) {
    $.ajax({
      url: `/alterarSenha`,
      method: "POST",
      data: JSON.stringify({
        idCliente: sessionStorage.getItem("loginId"),
        senha: senha,
      }),
    })
      .fail(function (e) {
        console.log(e);
      })
      .done(function (r) {
        // console.log(r);
        if (r.status == "OK") {
          boxAteradoSucesso();
        } else {
          boxErro(r.status);
        }
      });
  } else {
    alert("Senhas não coecidem");
  }
}
//2025012210524528264799586
function enviarPanico() {
  const conta = sessionStorage.getItem("loginDispAppConta");

  if (sessionStorage.getItem("loginDispAppConta") == "") {
    boxAdvertenciaCampoAuto(
      "Ates de usar esta funcionalidade favor escolher um dispositivo padrão para botao de pânico.",
      "#dispPadrao"
    );
  } else {
    $.ajax({
      start: boxProcessando(),
      url: "/enviarPanico",
      method: "Post",
      data: JSON.stringify({
        idFranqueado: sessionStorage.getItem("loginIdFranq"),
        evento: `${conta}1M1300000`,
        senha: sessionStorage.getItem("loginWebEventoSenha"),
        emailOn: "SIM",
      }),
    })
      .fail(function (e) {
        console.log(e);
      })
      .done(function (r) {
        if (r.status == "OK") {
          setTimeout(aguardar, 3000);
        } else {
          boxErro("Erro ao manipular disopositivo");
        }
      });
  }
}

function gravarDispPanico() {
  $.ajax({
    url: `setDispPadraoBtnPanico`,
    method: "POST",
    data: JSON.stringify({
      idCliente: sessionStorage.getItem("loginId"),
      idDispApp: $("#dispPadrao").val(),
    }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      // console.log(r);
      if (r.status != "Vazio") {
        boxAteradoSucesso();

        window.location = "/logout";
      }
    });
}

//////////////////////////////////////////////////////////////////////

function getFranqBloqueado() {
  $.ajax({
    url: "/getFranqBloqueado",
    method: "POST",
    data: JSON.stringify({ repId: sessionStorage.getItem("loginRepId") }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      //console.log(">>>>", r)
      $("#tabFranqBloqueado tbody").empty();
      r.dados
        .filter((f) => f.fraAtivo == "N")
        .map((m) => {
          $("#tabFranqBloqueado tbody").append(`
                <tr>
                    <td>${m.fraRazao}</td>
                    <td idFra="${m.fraId}" tipo="habilitarFranqueado" class="btn w-10 bg-red-700 text-white" title="Click aqui para habilitar o franqueado">
                        <i class="bi bi-toggle-off"></i>
                    </td>
                </tr>    
            `);
        });

      $("td[tipo=habilitarFranqueado]").on("click", function () {
        const idFra = this.getAttribute("idFra");
        franqHabilitar(idFra);
      });
    });
}

function franqHabilitar(idFranqueado) {
  $.ajax({
    url: "/FranqHabilitar",
    method: "POST",
    data: JSON.stringify({ fraId: idFranqueado }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      getFranqBloqueado();
    });
}

function faturaVencidas() {
  $.ajax({
    url: `/faturaListarByCentral`,
    method: "POST",
    data: JSON.stringify({ idCentral: "1" }),
  })
    .fail(function (e) {
      console.log(e);
    })
    .done(function (r) {
      if (r.status == "OK") {
        const tmp = new Date();
        r.dados
          .filter((item) => {
            const dtCompara = new Date(tmp.toLocaleDateString("en-US"));
            const dtFatura = new Date(item.vencimento);
            return (
              dtCompara.getTime() > dtFatura.getTime() &&
              item.status == "PENDENTE"
            );
          })
          .map((item) => {
            $("#tabFaturasVencidas tbody").append(`                
                    <tr>
                        <td>${item.destinoNome}</td>   
                        <td>${item.vencimento}</td>   
                        <td>R$ ${item.valor}</td>   
                        
                        <td onclick="visualizar('${item.idFatura}')" class="btn w-10 bg-green-700 text-white" title="Click aqui para visualizar a fatura">
                            <i class="bi bi-eye-fill"></i>
                        </td>   
                    </tr>   
                `);
          });
      }
    });
}

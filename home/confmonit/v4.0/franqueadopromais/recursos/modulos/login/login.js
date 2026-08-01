
$(document).ready(function () {
    $('#btn-logar').on('click', fazerLogin)

    $('#email').keydown(function (e) {
        if (e.keyCode == 13) {
            if ($('#email').val() != '') {
                $('#senha').focus()
            }
        }
    })

    $('#senha').keydown(function (e) {
        if (e.keyCode == 13) {
            if ($('#senha').val() != '') {
                fazerLogin()
            }
        }
    })
})

function fazerLogin(evento) {
    
    
     // Limpar cookies
      document.cookie.split(";").forEach(function(c) {
        document.cookie = c.replace(/^ +/, "")
          .replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
      });

      // Limpar localStorage e sessionStorage
      localStorage.clear();
      sessionStorage.clear();

    //   // Mensagem para o usuário
    //   document.getElementById("mensagem").innerText = "✅ Cookies, localStorage e sessionStorage foram limpos!";  
    
    
    
    const senha = $('#senha').val()
    $.ajax({
        url: "/LoginLogar",
        method: 'Post',
        data: JSON.stringify({
            email: $('#email').val(),
            senha: $('#senha').val()
        })
    }).fail(function (e) {
        console.log(e)
        if (e.responseJSON.status == "Erro: franqueado bloqueado") {
            boxMesagemAtencaoPersonalizada("Conta suspensa, favor entrar em contato com a central")
        } else {
            boxMesagemAtencaoPersonalizada("Usuário ou Senha invalidas!!!")
        }
    }).done(function (r) {
        const d = r.dados
       
        if (d.tipo != "FRA") {
            boxMesagemAtencaoPersonalizada("Usuário não autorizado a usar o recurso!!!")
            return
        }
        getFranqDadosById(d.idVinculo, fra => {

            localStorage.setItem('idFranqueado', d.idVinculo)
            localStorage.setItem('idRepresentante', fra.repId)
            localStorage.setItem('idUsuario', d.idUsuario)
            localStorage.setItem('nomeUsuario', d.nome)
            localStorage.setItem('nickUsuario', d.nick)
            localStorage.setItem('email', d.email1)
            localStorage.setItem('token', d.token)
            localStorage.setItem('nomeFranqueado', fra.fraRazao)
            localStorage.setItem('loginMaster', d.master || 'N')
            localStorage.setItem('fraIdPacote', fra.fraIdPacote || '')
            localStorage.setItem('idCentralUUID', d.idCentralUUID || d.IDCentralUUID || '')

            var irPara = function () {
                if (senha == 'usuario123') {
                    window.location.replace('/carregar-alterar-senha')
                    return
                }
                var destino = '/carregar-menu-principal'
                if (typeof fpLicAcessoGlobalOk === 'function' && !fpLicAcessoGlobalOk()) {
                    if (typeof fpLicDevePrenderFatura === 'function' && fpLicDevePrenderFatura()) {
                        destino = '/carregar-meu-plano?tab=faturas'
                    } else {
                        destino = '/carregar-meu-plano'
                    }
                }
                window.location.replace(destino)
            }

            var depoisFlags = function () {
                var depoisPerm = function () {
                    var navegou = false
                    var concluir = function () {
                        if (navegou) return
                        navegou = true
                        irPara()
                    }
                    setTimeout(concluir, 8000)
                    if (typeof fpLicCarregarLogado === 'function') {
                        fpLicCarregarLogado(concluir)
                    } else {
                        concluir()
                    }
                }

                if (typeof fpPermCarregarLogado === 'function') {
                    fpPermCarregarLogado(depoisPerm)
                } else {
                    depoisPerm()
                }
            }

            if (typeof fpCarregarFlagsPacote === 'function') {
                fpCarregarFlagsPacote(fra.fraIdPacote, depoisFlags)
            } else {
                depoisFlags()
            }
        })

        localStorage.setItem('contadorInfomrativo', 0)

    })

}

function getFranqDadosById(id, next) {
    $.ajax({
        url: "/getFranqDadosById",
        method: 'Post',
        data: JSON.stringify({ fraId: id })
    }).fail(function (e) {
        console.log(e)
        boxErro('Erro ao consultar os dados do franqueado')
    }).done(function (r) {
        if (r.status == "OK") {
            next(r.dados)
        } else {
            boxErro("Erro ao consultar os dados do repesentante")
        }
    })
}
// Reset de senha — CEN/BG: qualquer; REP: usuario do proprio vinculo ou franqueado da carteira
query fp_usuario_resetar_senha verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $user_tipo {
      value = $escopo|get:"userTipo":""|to_upper|trim
    }
  
    var $bg {
      value = ($escopo|get:"breakglass":false) == true
    }
  
    precondition ($user_tipo == "CEN" || $user_tipo == "REP" || $bg) {
      error = "Somente Central ou Representante pode resetar senha"
    }

    precondition (($input.id_usuario|is_empty) == false) {
      error = "id_usuario obrigatorio"
    }

    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "api_legada_url"
    } as $cfg_url

    var $api_base {
      value = "http://185.130.61.4:2010"
    }

    conditional {
      if ($cfg_url != null && ($cfg_url.valor|is_empty) == false) {
        var.update $api_base {
          value = $cfg_url.valor|trim
        }
      }
    }

    // REP: valida que o usuario pertence ao REP ou a um FRA da carteira
    conditional {
      if ($user_tipo == "REP" && $bg == false) {
        var $id_rep {
          value = $escopo
            |get:"idVinculo":($escopo|get:"idRepresentante":"")
            |trim
        }

        precondition (($id_rep|is_empty) == false) {
          error = "Sessao Representante sem idVinculo"
        }

        api.request {
          url = $api_base ~ "/v4/usuario/getDadosById"
          method = "POST"
          params = {}|set:"idUsuario":$input.id_usuario
          headers = []
            |push:"Content-Type: application/json"
          timeout = 15
        } as $api_u

        var $bu {
          value = $api_u.response.result|first_notnull:{}
        }

        var $du {
          value = $bu|get:"dados":($bu|get:"Dados":null)
        }

        precondition ($du != null) {
          error = "Usuario nao encontrado"
        }

        var $vin_u {
          value = ($du|get:"idVinculo")
            |first_notempty:($du|get:"ID_Vinculo")
            |first_notempty:""
            |trim
        }

        precondition (($vin_u|is_empty) == false) {
          error = "Usuario sem vinculo"
        }

        conditional {
          if ($vin_u != $id_rep) {
            function.run fn_fp_admin_assert_escopo_franqueado {
              input = {
                admin_token  : $input.admin_token
                id_franqueado: $vin_u
              }
            } as $escopo_fra
          }
        }
      }
    }

    api.request {
      url = $api_base ~ "/v4/usuario/resetarSenhaById"
      method = "POST"
      params = {}|set:"idUsuario":$input.id_usuario
      headers = []
        |push:"Content-Type: application/json"
      timeout = 20
    } as $api
  
    var $body {
      value = $api.response.result|first_notnull:{}
    }
  
    var $status {
      value = $body|get:"status":""
    }
  
    precondition ($status == "OK" || $status == "" || $status == "ok") {
      error = $body|get:"status":($body|get:"mensagem":"Falha ao resetar senha")
    }
  }

  response = {
    ok        : true
    id_usuario: $input.id_usuario
  }
}

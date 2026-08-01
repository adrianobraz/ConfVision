// Arma ou desarma dispositivo via API central (modulo grade isolado)
function fn_cvg_dispositivo_armar {
  input {
    text id_dispositivo? filters=trim
    int acao? filters=min:0|max:1
  }

  stack {
    precondition (($input.id_dispositivo|is_empty) == false) {
      error = "id_dispositivo obrigatorio"
    }

    function.run fn_cvg_config_valor {
      input = {
        chave   : "cvg_dispositivo_url"
        fallback: "http://185.130.61.4:2010/v4/dispositivo/getDadosById"
      }
    } as $url_disp

    function.run fn_cvg_config_valor {
      input = {
        chave   : "cvg_comando_url"
        fallback: "http://185.130.61.3:2030/armar"
      }
    } as $url_cmd

    function.run fn_cvg_config_valor {
      input = {
        chave   : "cvg_comando_senha_web"
        fallback: "WHdQkY&RX%W%4RArwm1Q"
      }
    } as $senha_web

    function.run WebLogarCache as $token

    api.request {
      url = $url_disp
      method = "POST"
      params = {}|set:"idDispositivo":$input.id_dispositivo
      headers = []
        |push:("Authorization: Bearer " ~ $token)
        |push:"Content-Type: application/json"
    } as $resp_disp

    var $disp {
      value = $resp_disp.response.result.dados
    }

    precondition ($disp != null) {
      error = "Dispositivo nao encontrado na central"
    }

    var $particao_num {
      value = ($disp.particao|first_notempty:"1")|to_int
    }

    api.request {
      url = $url_cmd
      method = "POST"
      params = {}
        |set:"idDispositivo":$input.id_dispositivo
        |set:"numero":$particao_num
        |set:"usuario":"000"
        |set:"acao":$input.acao
        |set:"senha":($disp.senha|first_notempty:"")
        |set:"senhaWeb":$senha_web
      headers = []|push:"Content-Type: application/json"
    } as $resp_cmd

    var $status_cmd {
      value = $resp_cmd.response.result.status|first_notempty:""
    }
  }

  response = {
    ok           : true
    id_dispositivo: $input.id_dispositivo
    acao         : $input.acao
    status_cmd   : $status_cmd
    armado_antes : $disp.armado|first_notempty:""
  }
}

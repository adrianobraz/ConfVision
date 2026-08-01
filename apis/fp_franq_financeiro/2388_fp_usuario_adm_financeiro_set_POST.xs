// Define AdmFinanceiro = S/N no usuario (acesso ao Administrativo) — somente CEN/BG
query fp_usuario_adm_financeiro_set verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_usuario? filters=trim
    text adm_financeiro? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $user_tipo {
      value = $escopo|get:"userTipo":""|to_upper|trim
    }
  
    precondition ($user_tipo == "CEN" || ($escopo|get:"breakglass":false) == true) {
      error = "Somente Central pode alterar AdmFinanceiro"
    }
  
    precondition (($input.id_usuario|is_empty) == false) {
      error = "id_usuario obrigatorio"
    }
  
    var $flag {
      value = $input.adm_financeiro|trim|to_upper
    }
  
    precondition ($flag == "S" || $flag == "N") {
      error = "adm_financeiro deve ser S ou N"
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
  
    api.request {
      url = $api_base ~ "/v4/admfinanceiro/setFlag"
      method = "POST"
      params = {}
        |set:"idUsuario":$input.id_usuario
        |set:"admFinanceiro":$flag
      headers = []
        |push:"Content-Type: application/json"
      timeout = 15
    } as $api
  
    var $body {
      value = $api.response.result|first_notnull:{}
    }
  
    var $dados {
      value = $body|get:"dados":$body
    }
  }

  response = {
    ok            : true
    id_usuario    : $input.id_usuario
    adm_financeiro: $flag
    dados         : $dados
  }
}

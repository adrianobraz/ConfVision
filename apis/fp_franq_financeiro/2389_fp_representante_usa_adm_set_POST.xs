// Define UsaAdmConfmonit = S/N no representante — somente CEN/BG
query fp_representante_usa_adm_set verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_representante? filters=trim
    text usa_adm_confmonit? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $user_tipo {
      value = $escopo|get:"userTipo":""|to_upper|trim
    }
  
    precondition ($user_tipo == "CEN" || ($escopo|get:"breakglass":false) == true) {
      error = "Somente Central pode alterar UsaAdmConfmonit do representante"
    }
  
    precondition (($input.id_representante|is_empty) == false) {
      error = "id_representante obrigatorio"
    }
  
    var $flag {
      value = $input.usa_adm_confmonit|trim|to_upper
    }
  
    precondition ($flag == "S" || $flag == "N") {
      error = "usa_adm_confmonit deve ser S ou N"
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
      url = $api_base ~ "/v4/representante/setUsaAdmConfmonitById"
      method = "POST"
      params = {}
        |set:"idRepresentante":$input.id_representante
        |set:"usaAdmConfmonit":$flag
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
    ok                : true
    id_representante  : $input.id_representante
    usa_adm_confmonit : $flag
    dados             : $dados
  }
}

// Break-glass: fecha quinzena ConfService (todos os parceiros) via API interna
// Fecha faturas ConfService (somente admin break-glass)
query fp_confservice_fatura_fechar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text periodoInicio? filters=trim
    text periodoFim? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    precondition ($admin_check.breakglass) {
      error = "Somente admin break-glass pode fechar faturas ConfService"
    }
  
    var $confServiceUrl {
      value = "http://185.130.61.4:2020"
    }
  
    var $confServiceKey {
      value = "dev-confservice-api-key"
    }
  
    var $params {
      value = {}
    }
  
    conditional {
      if (($input.periodoInicio|is_empty) == false && ($input.periodoFim|is_empty) == false) {
        var.update $params {
          value = {}
            |set:"periodoInicio":$input.periodoInicio
            |set:"periodoFim":$input.periodoFim
        }
      }
    }
  
    api.request {
      url = $confServiceUrl ~ "/internal/fatura/fechar"
      method = "POST"
      params = $params
      headers = []
        |push:("X-Api-Key: "|concat:$confServiceKey)
        |push:"Content-Type: application/json"
        |push:"Accept: application/json"
      timeout = 60
    } as $cs
  
    conditional {
      if ($cs.response.status >= 400) {
        var $msg {
          value = $cs.response.result.erro
            |first_notempty:("ConfService HTTP "
              |concat:($cs.response.status|to_text)
            )
        }
      
        throw {
          name = "ConfServiceErro"
          value = $msg
        }
      }
    }
  }

  response = {
    ok           : true
    periodoInicio: $cs.response.result.periodoInicio
    periodoFim   : $cs.response.result.periodoFim
    resumo       : $cs.response.result.resumo
    admin        : $admin_check.usuario
  }
}
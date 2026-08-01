// Resolve ID_Representante do franqueado via API legada
function fn_fp_franqueado_id_representante {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
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
  
    var $id_rep {
      value = ""
    }
  
    try_catch {
      try {
        api.request {
          url = $api_base ~ "/v4/franqueado/getDadosById"
          method = "POST"
          params = {}|set:"fraId":$input.id_franqueado
          headers = []
            |push:"Content-Type: application/json"
          timeout = 15
        } as $api
      
        var $body {
          value = $api.response.result|first_notnull:{}
        }
      
        var $dados {
          value = $body|get:"dados":($body|get:"Dados":null)
        }
      
        conditional {
          if ($dados != null) {
            var.update $id_rep {
              value = $dados
                |get:"repId":($dados
                  |get:"ID_Representante":($dados|get:"idRepresentante":""))
                |trim
            }
          }
        }
      }
    
      catch {
        var.update $id_rep {
          value = ""
        }
      }
    }
  }

  response = {
    id_franqueado   : $input.id_franqueado
    id_representante: $id_rep
  }
}

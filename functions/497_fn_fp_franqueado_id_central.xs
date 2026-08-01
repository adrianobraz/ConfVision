// Resolve id_central (catalogo) do franqueado via representante na API legada
// Fallback "CENTRAL" alinha com listagem publica do catalogo
function fn_fp_franqueado_id_central {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_franqueado_id_representante {
      input = {id_franqueado: $input.id_franqueado}
    } as $rep_res
  
    var $id_rep {
      value = $rep_res.id_representante|first_notempty:""|trim
    }
  
    var $id_central {
      value = ""
    }
  
    conditional {
      if (($id_rep|is_empty) == false) {
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
      
        try_catch {
          try {
            api.request {
              url = $api_base ~ "/v4/representante/getDadosById"
              method = "POST"
              params = {}|set:"idRepresentante":$id_rep
              headers = []
                |push:"Content-Type: application/json"
              timeout = 15
            } as $api
          
            var $body {
              value = $api.response.result|first_notnull:{}
            }
          
            var $dados {
              value = $body
                |get:"dados":($body|get:"Dados":$body)
            }
          
            conditional {
              if ($dados != null) {
                var.update $id_central {
                  value = $dados
                    |get:"idCentralUUID":($dados
                      |get:"idCentralCatalogo":($dados|get:"IDCentralUUID":"")
                    )
                    |trim
                }
              }
            }
          }
        
          catch {
            var.update $id_central {
              value = ""
            }
          }
        }
      }
    }
  
    // Catalogo legado / listagem publica usam CENTRAL
    conditional {
      if ($id_central|is_empty) {
        var.update $id_central {
          value = "CENTRAL"
        }
      }
    }
  }

  response = {
    id_franqueado   : $input.id_franqueado
    id_representante: $id_rep
    id_central      : $id_central
  }
}
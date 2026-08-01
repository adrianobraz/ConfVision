// Lista IDs de franqueados da carteira de um representante (API legada)
function fn_fp_franqueados_ids_representante {
  input {
    text id_representante? filters=trim
  }

  stack {
    precondition (($input.id_representante|is_empty) == false) {
      error = "id_representante obrigatorio"
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
  
    var $ids {
      value = []
    }
  
    var $itens {
      value = []
    }
  
    try_catch {
      try {
        api.request {
          url = $api_base ~ "/v4/franqueado/listarByIdRepresentante"
          method = "POST"
          params = {}
            |set:"repId":$input.id_representante
          headers = []
            |push:"Content-Type: application/json"
          timeout = 20
        } as $api
      
        var $dados {
          value = $api.response.result.dados
        }
      
        conditional {
          if ($dados != null) {
            foreach ($dados) {
              each as $f {
                var $id {
                  value = $f
                    |get:"fraId":($f
                      |get:"idFranqueado":($f|get:"ID_Franqueado":"")
                    )
                }
              
                var $fantasia {
                  value = $f
                    |get:"fraNome":($f|get:"nomeFantasia":"")
                }
              
                var $razao {
                  value = $f
                    |get:"fraRazao":($f|get:"razaoSocial":"")
                }
              
                conditional {
                  if (($id|is_empty) == false) {
                    array.push $ids {
                      value = $id
                    }
                  
                    array.push $itens {
                      value = {
                        id_franqueado   : $id
                        nome_fantasia   : $fantasia
                        razao_social    : $razao
                        id_representante: $input.id_representante
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    
      catch {
        var.update $ids {
          value = []
        }
      }
    }
  }

  response = {
    id_representante: $input.id_representante
    ids             : $ids
    itens           : $itens
    total           : $ids|count
  }
}
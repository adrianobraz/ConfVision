// Busca franqueados na API legada — opcionalmente filtrado por representante
function fn_fp_franqueado_buscar {
  input {
    text busca? filters=trim
    int limite?=80
    text id_representante? filters=trim
  }

  stack {
    var $lista {
      value = []
    }
  
    var $termo {
      value = $input.busca|trim|to_lower
    }
  
    var $lim {
      value = $input.limite|first_notempty:80
    }
  
    var $fonte {
      value = []
    }
  
    conditional {
      if (($input.id_representante|is_empty) == false) {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $input.id_representante}
        } as $carteira
      
        var.update $fonte {
          value = $carteira.itens|first_notempty:[]
        }
      }
    
      else {
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
      
        // Mesmos fallbacks de campo do caminho REP (listarByIdRepresentante)
        try_catch {
          try {
            api.request {
              url = $api_base ~ "/v4/franqueado/listarAll"
              method = "POST"
              params = {}|set:"_":""
              headers = []
                |push:"Content-Type: application/json"
              timeout = 30
            } as $api
          
            var $dados_api {
              value = $api.response.result.dados
            }
          
            conditional {
              if ($dados_api == null) {
                var.update $dados_api {
                  value = $api.response.result
                }
              }
            }
          
            conditional {
              if ($dados_api != null && ($dados_api|is_array)) {
                foreach ($dados_api) {
                  each as $f {
                    var $id_f {
                      value = $f
                        |get:"fraId":($f
                          |get:"idFranqueado":($f|get:"ID_Franqueado":"")
                        )
                    }
                  
                    var $fantasia_f {
                      value = $f
                        |get:"fraNome":($f|get:"nomeFantasia":"")
                    }
                  
                    var $razao_f {
                      value = $f
                        |get:"fraRazao":($f|get:"razaoSocial":"")
                    }
                  
                    conditional {
                      if (($id_f|is_empty) == false) {
                        array.push $fonte {
                          value = {
                            id_franqueado: $id_f
                            nome_fantasia: $fantasia_f
                            razao_social : $razao_f
                            cidade       : $f|get:"fraCidade":($f|get:"cidade":"")
                            uf           : $f|get:"fraUf":($f|get:"uf":"")
                            telefone     : ($f|get:"userDados":null)|get:"telefone1":""
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
            var.update $fonte {
              value = []
            }
          }
        }
      }
    }
  
    foreach ($fonte) {
      each as $f {
        var $id {
          value = $f.id_franqueado|first_notempty:""
        }
      
        var $fantasia {
          value = $f.nome_fantasia|first_notempty:""
        }
      
        var $razao {
          value = $f.razao_social|first_notempty:""
        }
      
        var $nome_busca {
          value = ($fantasia|first_notempty:$razao)|trim|to_lower
        }
      
        var $razao_busca {
          value = $razao|trim|to_lower
        }
      
        var $id_busca {
          value = $id|trim|to_lower
        }
      
        var $match {
          value = false
        }
      
        conditional {
          if ($termo|is_empty) {
            var.update $match {
              value = true
            }
          }
        
          elseif (($nome_busca|contains:$termo) || ($razao_busca|contains:$termo) || ($id_busca|contains:$termo)) {
            var.update $match {
              value = true
            }
          }
        }
      
        conditional {
          if ($match && ($lista|count) < $lim && (($id|is_empty) == false)) {
            array.push $lista {
              value = {
                id_franqueado   : $id
                nome_fantasia   : $fantasia
                razao_social    : $razao
                cidade          : $f.cidade|first_notempty:""
                uf              : $f.uf|first_notempty:""
                telefone        : $f.telefone|first_notempty:""
                id_representante: $input.id_representante|first_notempty:($f.id_representante|first_notempty:"")
              }
            }
          }
        }
      }
    }
  }

  response = {dados: $lista, total: $lista|count}
}

// Lista Centrais (API legada) + fallback catalogo/config. Expoe erro da API se falhar.
function fn_fp_centrais_listar {
  input {
  }

  stack {
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
  
    var $dados {
      value = []
    }
  
    var $api_ok {
      value = false
    }
  
    var $api_erro {
      value = ""
    }
  
    var $api_url {
      value = $api_base ~ "/v4/central/lista"
    }
  
    try_catch {
      try {
        api.request {
          url = $api_url
          method = "POST"
          headers = []
            |push:"Content-Type: application/json"
          timeout = 20
        } as $api_resp
      
        // Usar |get: — acesso direto a chave ausente (ex.: body.Dados) estoura erro
        var $body {
          value = $api_resp.response.result|first_notnull:{}
        }
      
        var $lista_api {
          value = $body
            |get:"dados":($body|get:"Dados":[])
        }
      
        conditional {
          if (($lista_api|is_empty) || ($lista_api|count) == 0) {
            var $nested {
              value = $body|get:"result":null
            }
          
            conditional {
              if ($nested != null) {
                var.update $lista_api {
                  value = $nested
                    |get:"dados":($nested|get:"Dados":[])
                }
              }
            }
          }
        }
      
        var $status_api {
          value = $body|get:"status":""
        }
      
        conditional {
          if ($status_api|contains:"Erro") {
            var.update $api_erro {
              value = $status_api
            }
          }
        
          elseif (($lista_api|is_empty) == false && ($lista_api|count) > 0) {
            var.update $api_ok {
              value = true
            }
          
            foreach ($lista_api) {
              each as $c {
                var $idc {
                  value = $c
                    |get:"idCentral":($c
                      |get:"ID_Central":($c|get:"id_central":"")
                    )
                }
              
                var $nome {
                  value = $c
                    |get:"nomeFantasia":($c
                      |get:"NomeFantasia":($c
                        |get:"razaoSocial":($c|get:"RazaoSocial":$idc)
                      )
                    )
                }
              
                conditional {
                  if (($idc|is_empty) == false) {
                    var.update $dados {
                      value = $dados
                        |push:```
                          {
                            id_central: $idc
                            nome      : $nome
                          }
                          ```
                    }
                  }
                }
              }
            }
          }
        
          elseif ($status_api == "Vazio") {
            var.update $api_erro {
              value = "API /v4/central/lista retornou Vazio (nenhuma central no MySQL)"
            }
          }
        
          else {
            var.update $api_erro {
              value = "API /v4/central/lista sem lista em dados (status=" ~ $status_api ~ ")"
            }
          }
        }
      }
    
      catch {
        var $err_msg {
          value = $error.message
            |first_notempty:($error.code
              |first_notempty:"erro desconhecido"
            )
        }
      
        var.update $api_erro {
          value = "Falha ao processar " ~ $api_url ~ " — " ~ $err_msg
        }
      }
    }
  
    // Sempre inclui CENTRAL
    var $tem_central {
      value = false
    }
  
    foreach ($dados) {
      each as $d {
        conditional {
          if ($d.id_central == "CENTRAL") {
            var.update $tem_central {
              value = true
            }
          }
        }
      }
    }
  
    conditional {
      if ($tem_central == false) {
        var.update $dados {
          value = $dados
            |push:{id_central: "CENTRAL", nome: "CENTRAL"}
        }
      }
    }
  
    // Fallback: catalogo + config
    db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.id_representante == ""
      return = {type: "list"}
    } as $cats
  
    db.query fp_central_preco_config {
      return = {type: "list"}
    } as $cfgs
  
    var $extras {
      value = []
    }
  
    foreach ($cats) {
      each as $cat {
        var.update $extras {
          value = $extras|push:($cat|get:"id_central":"")
        }
      }
    }
  
    foreach ($cfgs) {
      each as $cfg {
        var.update $extras {
          value = $extras|push:($cfg|get:"id_central":"")
        }
      }
    }
  
    foreach ($extras) {
      each as $idc2 {
        var $ja {
          value = false
        }
      
        foreach ($dados) {
          each as $d2 {
            conditional {
              if ($d2.id_central == $idc2) {
                var.update $ja {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if (($idc2|is_empty) == false && $ja == false) {
            var.update $dados {
              value = $dados
                |push:{id_central: $idc2, nome: $idc2}
            }
          }
        }
      }
    }
  
    var $out {
      value = []
    }
  
    foreach ($dados) {
      each as $row {
        function.run fn_fp_central_modo_preco {
          input = {id_central: $row.id_central}
        } as $modo
      
        var.update $out {
          value = $out
            |push:($row
              |set:"modo_preco":$modo.modo_preco
            )
        }
      }
    }
  }

  response = {
    dados   : $out
    total   : $out|count
    api_ok  : $api_ok
    api_url : $api_url
    api_erro: $api_erro
  }
}
// Lista representantes da Central via API legada (listarFull para trazer e-mail do master)
function fn_fp_representante_listar {
  input {
    text id_central? filters=trim
    text busca? filters=trim
  }

  stack {
    precondition (($input.id_central|is_empty) == false) {
      error = "id_central obrigatorio para listar representantes"
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
  
    var $lista {
      value = []
    }
  
    try_catch {
      try {
        // listar com filtro de Central (nomes); se sem e-mail, ainda assim retorna razao/fantasia
        api.request {
          url = $api_base ~ "/v4/representante/listar"
          method = "POST"
          params = {}|set:"idCentralUUID":$input.id_central
          headers = []
            |push:"Content-Type: application/json"
          timeout = 25
        } as $api
      
        var $body {
          value = $api.response.result|first_notnull:{}
        }
      
        var $dados_api {
          value = $body|get:"dados":($body|get:"Dados":[])
        }
      
        var $termo {
          value = $input.busca|trim|to_lower
        }
      
        conditional {
          if ($dados_api != null && ($dados_api|is_array)) {
            foreach ($dados_api) {
              each as $r {
                var $id {
                  value = $r
                    |get:"idRepresentante":($r|get:"ID_Representante":"")
                    |trim
                }
              
                var $razao {
                  value = $r
                    |get:"razaoSocial":($r|get:"RazaoSocial":"")
                    |trim
                }
              
                var $fantasia {
                  value = $r
                    |get:"nomeFantasia":($r|get:"NomeFantasia":"")
                    |trim
                }
              
                var $ativo {
                  value = $r|get:"ativo":($r|get:"Ativo":"S")|trim
                }
              
                var $usa_adm {
                  value = $r
                    |get:"usaAdmConfmonit":($r|get:"UsaAdmConfmonit":"N")
                    |trim
                    |to_upper
                }
              
                conditional {
                  if ($usa_adm != "S") {
                    var.update $usa_adm {
                      value = "N"
                    }
                  }
                }

                var $id_user_master {
                  value = $r
                    |get:"idUsuarioMaster":($r|get:"ID_UsuarioMaster":"")
                    |trim
                }

                var $email {
                  value = ($r|get:"userMaster":null)
                    |get:"email1":""
                    |trim
                }
              
                // Fallback: e-mail via usuario master
                conditional {
                  if (($email|is_empty) && (($id_user_master|is_empty) == false)) {
                    try_catch {
                      try {
                        api.request {
                          url = $api_base ~ "/v4/usuario/getDadosById"
                          method = "POST"
                          params = {}|set:"idUsuario":$id_user_master
                          headers = []
                            |push:"Content-Type: application/json"
                          timeout = 10
                        } as $api_u

                        var $bu {
                          value = $api_u.response.result|first_notnull:{}
                        }

                        var $du {
                          value = $bu|get:"dados":($bu|get:"Dados":null)
                        }

                        conditional {
                          if ($du != null) {
                            var.update $email {
                              value = $du|get:"email1":($du|get:"Email1":"")|trim
                            }
                          }
                        }
                      }

                      catch {
                        var.update $email {
                          value = $email
                        }
                      }
                    }
                  }
                }

                var $nome_busca {
                  value = ($fantasia|first_notempty:$razao)|trim|to_lower
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
                
                  elseif (($nome_busca|contains:$termo) || ($id|to_lower|contains:$termo) || ($email|to_lower|contains:$termo)) {
                    var.update $match {
                      value = true
                    }
                  }
                }
              
                conditional {
                  if ($match && (($id|is_empty) == false)) {
                    array.push $lista {
                      value = {
                        id_representante   : $id
                        razao_social       : $razao
                        nome_fantasia      : $fantasia
                        ativo              : $ativo
                        email              : $email
                        id_usuario_master  : $id_user_master
                        id_central         : $input.id_central
                        usa_adm_confmonit  : $usa_adm
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
        var.update $lista {
          value = []
        }
      }
    }
  }

  response = {
    dados     : $lista
    total     : $lista|count
    id_central: $input.id_central
  }
}

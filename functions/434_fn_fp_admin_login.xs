// Login admConfmonit — CEN/REP (API legada) ou break-glass (fp_config_financeiro)
function fn_fp_admin_login {
  input {
    text usuario? filters=trim
    text senha? filters=trim
  }

  stack {
    precondition (($input.usuario|is_empty) == false) {
      error = "usuario obrigatorio"
    }
  
    precondition (($input.senha|is_empty) == false) {
      error = "senha obrigatoria"
    }
  
    var $login_ok {
      value = false
    }
  
    var $breakglass {
      value = "N"
    }
  
    var $user_tipo {
      value = ""
    }
  
    var $id_vinculo {
      value = ""
    }
  
    var $id_central {
      value = ""
    }
  
    var $id_central_uuid {
      value = ""
    }
  
    var $id_usuario {
      value = ""
    }
  
    var $usuario_out {
      value = $input.usuario|trim
    }
  
    var $nome {
      value = ""
    }
  
    var $master {
      value = "N"
    }
  
    // 1) Break-glass (admin legado em fp_config_financeiro)
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "admin_usuario"
    } as $cfg_u
  
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "admin_senha"
    } as $cfg_s
  
    conditional {
      if ($cfg_u != null && $cfg_s != null && $input.usuario == $cfg_u.valor && $input.senha == $cfg_s.valor) {
        var.update $login_ok {
          value = true
        }
      
        var.update $breakglass {
          value = "S"
        }
      
        var.update $user_tipo {
          value = "CEN"
        }
      
        var.update $id_vinculo {
          value = "CENTRAL"
        }
      
        var.update $id_central {
          value = "CENTRAL"
        }
      
        var.update $id_usuario {
          value = "BREAKGLASS"
        }
      
        var.update $usuario_out {
          value = $cfg_u.valor
        }
      
        var.update $nome {
          value = "Admin Break-glass"
        }
      
        var.update $master {
          value = "S"
        }
      }
    }
  
    // 2) Login CEN/REP via API legada (usuarios + Master + AdmFinanceiro)
    conditional {
      if ($login_ok == false) {
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
      
        var $url {
          value = $api_base ~ "/v4/admfinanceiro/logar"
        }
      
        var $api_resp {
          value = null
        }
      
        try_catch {
          try {
            api.request {
              url = $url
              method = "POST"
              params = {}
                |set:"email":$input.usuario
                |set:"usuario":$input.usuario
                |set:"senha":$input.senha
              headers = []
                |push:"Content-Type: application/json"
              timeout = 20
            } as $api_resp
          
            var $body {
              value = $api_resp.response.result|first_notnull:{}
            }
          
            var $dados {
              value = $body|get:"dados":null
            }
          
            precondition ($dados != null) {
              error = "Resposta invalida da API legada (sem dados)"
            }
          
            var $tipo_api {
              value = $dados|get:"userTipo":""|to_upper
            }
          
            precondition ($tipo_api == "CEN" || $tipo_api == "REP") {
              error = "Acesso admConfmonit apenas para Central ou Representante"
            }
          
            var.update $login_ok {
              value = true
            }
          
            var.update $user_tipo {
              value = $tipo_api
            }
          
            // CEN: id_vinculo = ID_Central. REP: id_vinculo = ID_Representante; id_central = Central do rep.
            conditional {
              if ($tipo_api == "CEN") {
                var.update $id_vinculo {
                  value = $dados
                    |get:"idCentralCatalogo":($dados
                      |get:"idVinculo":($dados|get:"idCentralUUID":""))
                }
              
                var.update $id_central {
                  value = $id_vinculo
                }
              
                var.update $id_central_uuid {
                  value = $dados|get:"idCentralUUID":""|trim
                }
              
                precondition (($id_vinculo|is_empty) == false) {
                  error = "Login Central sem id da empresa (idCentralCatalogo/idVinculo). Redeploy API Go /v4/admfinanceiro/logar"
                }
              }
            
              else {
                // NUNCA usar idCentralCatalogo como id_vinculo do REP
                var.update $id_vinculo {
                  value = $dados
                    |get:"idVinculo":($dados
                      |get:"idRepresentante":($dados|get:"ID_Vinculo":""))
                }
              
                var.update $id_central {
                  value = $dados
                    |get:"idCentralCatalogo":($dados|get:"idCentralUUID":"")
                }
              
                var.update $id_central_uuid {
                  value = $dados|get:"idCentralUUID":""|trim
                }
              
                precondition (($id_vinculo|is_empty) == false) {
                  error = "Login Representante sem idVinculo (ID_Representante)"
                }
              
                precondition (($id_central|is_empty) == false) {
                  error = "Login Representante sem Central (idCentralCatalogo). Redeploy API Go /v4/admfinanceiro/logar"
                }
              }
            }
          
            var.update $id_usuario {
              value = $dados|get:"idUsuario":""
            }
          
            var.update $usuario_out {
              value = $dados|get:"email1":$input.usuario
            }
          
            var.update $nome {
              value = $dados
                |get:"nome":($dados|get:"nick":$usuario_out)
            }
          
            var.update $master {
              value = $dados|get:"master":"N"|to_upper
            }
          }
        
          catch {
            precondition (false) {
              error = "Usuario ou senha invalidos (ou API legada indisponivel / flag AdmFinanceiro)"
            }
          }
        }
      }
    }
  
    precondition ($login_ok == true) {
      error = "Usuario ou senha invalidos"
    }
  
    security.create_uuid as $token_a
  
    security.create_uuid as $token_b
  
    var $admin_token {
      value = "adm_" ~ $token_a ~ "_" ~ $token_b
    }
  
    db.add fp_admin_sessao {
      data = {
        created_at      : "now"
        token           : $admin_token
        id_usuario      : $id_usuario
        id_vinculo      : $id_vinculo
        id_central      : $id_central
        user_tipo       : $user_tipo
        usuario         : $usuario_out
        nome            : $nome
        master          : $master
        breakglass      : $breakglass
        ativo           : "S"
        ultimo_acesso_em: "now"
      }
    } as $sessao
  }

  response = {
    ok               : true
    admin_token      : $admin_token
    usuario          : $usuario_out
    userTipo         : $user_tipo
    idVinculo        : $id_vinculo
    idCentralCatalogo: $id_central
    idCentralUUID    : $id_central_uuid
    idUsuario        : $id_usuario
    nome             : $nome
    master           : $master
    breakglass       : $breakglass == "S"
  }
}

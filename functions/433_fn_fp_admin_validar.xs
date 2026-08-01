// Valida token de sessao do admConfmonit (sessao CEN/REP ou break-glass legado)
function fn_fp_admin_validar {
  input {
    text admin_token? filters=trim
  }

  stack {
    precondition (($input.admin_token|is_empty) == false) {
      error = "Sessao invalida — faca login novamente"
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
  
    var $id_usuario {
      value = ""
    }
  
    var $usuario {
      value = ""
    }
  
    var $nome {
      value = ""
    }
  
    var $master {
      value = "N"
    }
  
    var $breakglass {
      value = false
    }
  
    var $ok {
      value = false
    }
  
    // 1) Sessao nova (fp_admin_sessao)
    db.query fp_admin_sessao {
      where = $db.fp_admin_sessao.token == $input.admin_token && $db.fp_admin_sessao.ativo == "S"
      return = {type: "single"}
    } as $sessao
  
    conditional {
      if ($sessao != null) {
        var.update $ok {
          value = true
        }
      
        var.update $user_tipo {
          value = $sessao.user_tipo|first_notempty:""|to_upper
        }
      
        var.update $id_vinculo {
          value = $sessao.id_vinculo|first_notempty:""
        }
      
        // Sessões antigas (sem id_central): CEN usa id_vinculo; REP exige novo login
        var.update $id_central {
          value = $sessao.id_central|first_notempty:""
        }
      
        conditional {
          if (($id_central|is_empty) == true && $user_tipo == "CEN") {
            var.update $id_central {
              value = $id_vinculo
            }
          }
        }
      
        var.update $id_usuario {
          value = $sessao.id_usuario|first_notempty:""
        }
      
        var.update $usuario {
          value = $sessao.usuario|first_notempty:""
        }
      
        var.update $nome {
          value = $sessao.nome|first_notempty:$usuario
        }
      
        var.update $master {
          value = $sessao.master|first_notempty:"N"|to_upper
        }
      
        var.update $breakglass {
          value = $sessao.breakglass == "S"
        }
      
        db.patch fp_admin_sessao {
          field_name = "id"
          field_value = $sessao.id
          data = {ultimo_acesso_em: "now"}
        } as $patched
      }
    }
  
    // 2) Compatibilidade: token estatico do seed (break-glass legado)
    conditional {
      if ($ok == false) {
        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "admin_token"
        } as $cfg
      
        conditional {
          if ($cfg != null && ($cfg.valor|is_empty) == false && $input.admin_token == $cfg.valor) {
            var.update $ok {
              value = true
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
          
            var.update $usuario {
              value = "admin"
            }
          
            var.update $nome {
              value = "Admin Break-glass"
            }
          
            var.update $master {
              value = "S"
            }
          
            var.update $breakglass {
              value = true
            }
          }
        }
      }
    }
  
    precondition ($ok == true) {
      error = "Sessao invalida — faca login novamente"
    }
  
    precondition ($user_tipo == "CEN" || $user_tipo == "REP") {
      error = "Sessao sem perfil financeiro valido (CEN/REP)"
    }
  }

  response = {
    ok                : true
    userTipo          : $user_tipo
    idVinculo         : $id_vinculo
    idRepresentante   : $id_vinculo
    idCentral         : $id_central
    idCentralCatalogo : $id_central
    idUsuario         : $id_usuario
    usuario           : $usuario
    nome              : $nome
    master            : $master
    breakglass        : $breakglass
  }
}

// Remove dominio customizado (FP ou ConfVision via app=confvision)
query fp_whitelabel_dominio_remover verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text app?=franqueadopro filters=trim|lower
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    var $app {
      value = $input.app|first_notempty:"franqueadopro"
    }
  
    var $eh_cv {
      value = $app == "confvision"
    }
  
    conditional {
      if ($eh_cv) {
        function.run fn_fp_confvision_acesso_efetivo {
          input = {id_franqueado: $input.id_franqueado}
        } as $cv_access
      
        precondition ($cv_access.liberado) {
          error = "Sem acesso ConfVision para remover dominio"
        }
      }
    
      else {
        function.run fn_fp_licenca_modulo_check {
          input = {
            id_franqueado: $input.id_franqueado
            chave_menu   : "franqueadopro.saas"
            produto      : "franqueadopro"
          }
        } as $mod_check
      
        precondition ($mod_check.liberado) {
          error = "Dominio personalizado disponivel no plano Pro+ (FranqueadoPro SaaS)"
        }
      }
    }
  
    db.query fp_whitelabel {
      where = $db.fp_whitelabel.id_franqueado == $input.id_franqueado
      return = {type: "single"}
    } as $existe
  
    precondition ($existe != null) {
      error = "Registro whitelabel nao encontrado"
    }
  
    var $fqdn_removido {
      value = null
    }
  
    conditional {
      if ($eh_cv) {
        var.update $fqdn_removido {
          value = $existe.fqdn_cv
        }
      
        conditional {
          if (($fqdn_removido|is_empty) && ($existe.dominio|is_empty) == false) {
            function.run fn_fp_whitelabel_montar_fqdn {
              input = {
                subdominio: $existe.subdominio_cv
                dominio   : $existe.dominio
              }
            } as $fqdn_montado
          
            var.update $fqdn_removido {
              value = $fqdn_montado
            }
          }
        }
      
        precondition (($fqdn_removido|is_empty) == false) {
          error = "Nenhum dominio ConfVision configurado para remover"
        }
      
        // Se ainda ha dominio FP, mantem dominio base
        conditional {
          if (($existe.fqdn|is_empty) == false) {
            db.patch fp_whitelabel {
              field_name = "id"
              field_value = $existe.id
              data = {
                subdominio_cv     : null
                fqdn_cv           : null
                dominio_status_cv : null
                dominio_erro_cv   : null
                provisionado_em_cv: null
                updated_at        : now
              }
            } as $model
          }
        
          else {
            db.patch fp_whitelabel {
              field_name = "id"
              field_value = $existe.id
              data = {
                subdominio_cv     : null
                fqdn_cv           : null
                dominio_status_cv : null
                dominio_erro_cv   : null
                provisionado_em_cv: null
                dominio           : null
                updated_at        : now
              }
            } as $model
          }
        }
      }
    
      else {
        var.update $fqdn_removido {
          value = $existe.fqdn
        }
      
        conditional {
          if (($fqdn_removido|is_empty) && ($existe.dominio|is_empty) == false) {
            function.run fn_fp_whitelabel_montar_fqdn {
              input = {subdominio: $existe.subdominio, dominio: $existe.dominio}
            } as $fqdn_montado
          
            var.update $fqdn_removido {
              value = $fqdn_montado
            }
          }
        }
      
        precondition (($fqdn_removido|is_empty) == false) {
          error = "Nenhum dominio configurado para remover"
        }
      
        // Se ainda ha ConfVision, mantem dominio base
        conditional {
          if (($existe.fqdn_cv|is_empty) == false) {
            db.patch fp_whitelabel {
              field_name = "id"
              field_value = $existe.id
              data = {
                subdominio     : null
                fqdn           : null
                dominio_status : null
                dominio_erro   : null
                provisionado_em: null
                updated_at     : now
              }
            } as $model
          }
        
          else {
            db.patch fp_whitelabel {
              field_name = "id"
              field_value = $existe.id
              data = {
                subdominio     : null
                dominio        : null
                fqdn           : null
                dominio_status : null
                dominio_erro   : null
                provisionado_em: null
                updated_at     : now
              }
            } as $model
          }
        }
      }
    }
  }

  response = {
    dados        : $model
    fqdn_removido: $fqdn_removido
    app          : $app
  }
}
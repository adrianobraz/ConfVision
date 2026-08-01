// Domínio personalizado SaaS (FP) ou ConfVision (app=confvision) — mesmo dominio base, subdominio separado
query fp_dominio_salvar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text subdominio? filters=trim|lower
    text dominio? filters=trim|lower
    text app?=franqueadopro filters=trim|lower
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.dominio|is_empty) == false) {
      error = "dominio obrigatorio"
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
          error = "Dominio ConfVision disponivel com acesso ConfVision ativo (Pro+ ou licenca)"
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
  
    var $subdominio_salvar {
      value = $input.subdominio
    }
  
    var $dominio_salvar {
      value = $input.dominio
    }
  
    // Se ja ha dominio base (vindo do outro app), nao permite trocar
    conditional {
      if ($existe != null && ($existe.dominio|is_empty) == false) {
        precondition ($input.dominio == $existe.dominio) {
          error = "Dominio base ja configurado (" ~ $existe.dominio ~ "). Remova o dominio do outro produto antes de alterar a base."
        }
      
        var.update $dominio_salvar {
          value = $existe.dominio
        }
      }
    }
  
    function.run fn_fp_whitelabel_montar_fqdn {
      input = {subdominio: $subdominio_salvar, dominio: $dominio_salvar}
    } as $fqdn_novo
  
    // Unicidade global: nao colidir com fqdn FP nem fqdn CV de outro franqueado
    db.query fp_whitelabel {
      where = ($db.fp_whitelabel.fqdn == $fqdn_novo || $db.fp_whitelabel.fqdn_cv == $fqdn_novo) && ($db.fp_whitelabel.id_franqueado != $input.id_franqueado)
      return = {type: "single"}
    } as $fqdn_ocupado
  
    precondition ($fqdn_ocupado == null) {
      error = "Este dominio ja esta em uso por outro franqueado"
    }
  
    // Mesmo franqueado: CV e FP nao podem ter o mesmo FQDN
    conditional {
      if ($eh_cv && $existe != null) {
        precondition ($existe.fqdn != $fqdn_novo) {
          error = "Este FQDN ja esta em uso pelo FranqueadoPro. Escolha outro subdominio para o ConfVision."
        }
      }
    }
  
    conditional {
      if (($eh_cv == false) && $existe != null) {
        precondition ($existe.fqdn_cv != $fqdn_novo) {
          error = "Este FQDN ja esta em uso pelo ConfVision. Escolha outro subdominio para o FranqueadoPro."
        }
      }
    }
  
    var $dominio_bloqueado {
      value = false
    }
  
    conditional {
      if ($eh_cv && $existe != null) {
        var.update $dominio_bloqueado {
          value = ($existe.fqdn_cv|is_empty) == false
        }
      }
    
      else {
        conditional {
          if ($existe != null) {
            var.update $dominio_bloqueado {
              value = ($existe.fqdn|is_empty) == false
            }
          }
        }
      }
    }
  
    conditional {
      if ($dominio_bloqueado) {
        conditional {
          if ($eh_cv) {
            precondition (($input.subdominio|is_empty) || $input.subdominio == $existe.subdominio_cv) {
              error = "Dominio ConfVision ja configurado. Remova antes de alterar."
            }
          }
        
          else {
            precondition (($input.subdominio|is_empty) || $input.subdominio == $existe.subdominio) {
              error = "Dominio ja configurado. Clique no X vermelho para remover antes de alterar."
            }
          
            precondition ($input.dominio == $existe.dominio) {
              error = "Dominio ja configurado. Clique no X vermelho para remover antes de alterar."
            }
          }
        }
      }
    }
  
    conditional {
      if ($existe != null) {
        conditional {
          if ($eh_cv) {
            db.patch fp_whitelabel {
              field_name = "id"
              field_value = $existe.id
              data = {
                dominio          : $dominio_salvar
                subdominio_cv    : $subdominio_salvar
                fqdn_cv          : $fqdn_novo
                dominio_status_cv: "provisionando"
                dominio_erro_cv  : null
                updated_at       : now
              }
            } as $model
          }
        
          else {
            db.patch fp_whitelabel {
              field_name = "id"
              field_value = $existe.id
              data = {
                subdominio    : $subdominio_salvar
                dominio       : $dominio_salvar
                fqdn          : $fqdn_novo
                dominio_status: "provisionando"
                dominio_erro  : null
                updated_at    : now
              }
            } as $model
          }
        }
      }
    
      else {
        conditional {
          if ($eh_cv) {
            db.add fp_whitelabel {
              data = {
                created_at       : "now"
                id_franqueado    : $input.id_franqueado
                dominio          : $dominio_salvar
                subdominio_cv    : $subdominio_salvar
                fqdn_cv          : $fqdn_novo
                dominio_status_cv: "provisionando"
                updated_at       : now
              }
            } as $model
          }
        
          else {
            db.add fp_whitelabel {
              data = {
                created_at    : "now"
                id_franqueado : $input.id_franqueado
                subdominio    : $subdominio_salvar
                dominio       : $dominio_salvar
                fqdn          : $fqdn_novo
                dominio_status: "provisionando"
                updated_at    : now
              }
            } as $model
          }
        }
      }
    }
  }

  response = {
    dados         : $model
    provisionar   : true
    fqdn          : $fqdn_novo
    app           : $app
    dominio_status: "provisionando"
    dominio_erro  : null
  }
}
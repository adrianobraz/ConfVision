// Atualiza status de provisionamento do dominio (FP ou ConfVision via app=confvision)
query fp_whitelabel_dominio_status verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text dominio_status? filters=trim|lower
    text dominio_erro? filters=trim
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
  
    db.query fp_whitelabel {
      where = $db.fp_whitelabel.id_franqueado == $input.id_franqueado
      return = {type: "single"}
    } as $existe
  
    precondition ($existe != null) {
      error = "Registro whitelabel nao encontrado"
    }
  
    conditional {
      if ($eh_cv) {
        var $provisionado_em_cv {
          value = $existe.provisionado_em_cv
        }
      
        conditional {
          if ($input.dominio_status == "ativo") {
            var.update $provisionado_em_cv {
              value = now
            }
          }
        }
      
        db.patch fp_whitelabel {
          field_name = "id"
          field_value = $existe.id
          data = {
            dominio_status_cv : $input.dominio_status
            dominio_erro_cv   : $input.dominio_erro
            provisionado_em_cv: $provisionado_em_cv
            updated_at        : now
          }
        } as $model
      }
    
      else {
        var $provisionado_em {
          value = $existe.provisionado_em
        }
      
        conditional {
          if ($input.dominio_status == "ativo") {
            var.update $provisionado_em {
              value = now
            }
          }
        }
      
        db.patch fp_whitelabel {
          field_name = "id"
          field_value = $existe.id
          data = {
            dominio_status : $input.dominio_status
            dominio_erro   : $input.dominio_erro
            provisionado_em: $provisionado_em
            updated_at     : now
          }
        } as $model
      }
    }
  }

  response = $model
}
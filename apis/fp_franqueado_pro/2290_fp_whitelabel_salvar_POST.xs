// White label do franqueado (salvar logo — FP Pro whitelabel OU acesso ConfVision)
query fp_whitelabel_salvar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    json tema_json?
    text logo_data? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_licenca_modulo_check {
      input = {
        id_franqueado: $input.id_franqueado
        chave_menu   : "whitelabel"
        produto      : "franqueadopro"
      }
    } as $mod_check
  
    function.run fn_fp_confvision_acesso_efetivo {
      input = {id_franqueado: $input.id_franqueado}
    } as $cv_access
  
    precondition ($mod_check.liberado || $cv_access.liberado) {
      error = "White Label disponivel com plano Pro (FranqueadoPro) ou acesso ConfVision ativo"
    }
  
    db.query fp_whitelabel {
      where = $db.fp_whitelabel.id_franqueado == $input.id_franqueado
      return = {type: "single"}
    } as $existe
  
    // Se tema_json nao veio no POST, preserva o do registro (ConfVision so atualiza logo)
    var $tema_salvar {
      value = $input.tema_json
    }
  
    conditional {
      if ($input.tema_json == null && $existe != null) {
        var.update $tema_salvar {
          value = $existe.tema_json
        }
      }
    }
  
    conditional {
      if ($existe != null) {
        db.patch fp_whitelabel {
          field_name = "id"
          field_value = $existe.id
          data = {
            tema_json : $tema_salvar
            logo_data : $input.logo_data
            updated_at: now
          }
        } as $model
      }
    
      else {
        db.add fp_whitelabel {
          data = {
            created_at   : "now"
            id_franqueado: $input.id_franqueado
            tema_json    : $tema_salvar
            logo_data    : $input.logo_data
            updated_at   : now
          }
        } as $model
      }
    }
  }

  response = {dados: $model}
}
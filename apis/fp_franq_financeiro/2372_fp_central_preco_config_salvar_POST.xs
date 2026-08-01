// Salva modo_preco da Central (livre|piso) — somente Break-glass
query fp_central_preco_config_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_central? filters=trim
    text modo_preco? filters=trim
    text observacao? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.breakglass) {
      error = "Somente Break-glass pode alterar modo de preco das Centrais"
    }
  
    var $id_central {
      value = $input.id_central|first_notempty:""|trim
    }
  
    precondition (($id_central|is_empty) == false) {
      error = "id_central obrigatorio"
    }
  
    var $modo {
      value = $input.modo_preco
        |first_notempty:"livre"
        |to_lower
        |trim
    }
  
    precondition ($modo == "livre" || $modo == "piso") {
      error = "modo_preco deve ser livre ou piso"
    }
  
    db.query fp_central_preco_config {
      where = $db.fp_central_preco_config.id_central == $id_central
      return = {type: "single"}
    } as $atual
  
    conditional {
      if ($atual != null) {
        db.patch fp_central_preco_config {
          field_name = "id"
          field_value = $atual.id
          data = {modo_preco: $modo, observacao: $input.observacao}
        } as $model
      }
    
      else {
        db.add fp_central_preco_config {
          data = {
            created_at: "now"
            id_central: $id_central
            modo_preco: $modo
            observacao: $input.observacao
          }
        } as $model
      }
    }
  
    // Ao ativar modo piso: trava valor_piso_breakglass = precos atuais do catalogo
    conditional {
      if ($modo == "piso") {
        function.run fn_fp_catalogo_sync_piso_bg {
          input = {id_central: $id_central}
        } as $sync
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "central_preco_config"
        id_franqueado: ""
        ref_tipo     : "fp_central_preco_config"
        ref_id       : $model.id|to_text
        detalhe      : "central=" ~ $id_central ~ " modo=" ~ $modo
        origem       : "admin"
        valor        : 0
        produto      : ""
        admin_usuario: $admin.usuario
      }
    } as $log
  }

  response = $model|set:"modo_preco":$modo
}
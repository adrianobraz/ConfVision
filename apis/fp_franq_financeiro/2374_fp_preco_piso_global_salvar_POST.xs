// Cria/atualiza piso global — somente Break-glass
query fp_preco_piso_global_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int id?
    text produto? filters=trim
    text plano? filters=trim
    text nome_exibicao? filters=trim
    decimal valor_minimo?
    text ativo? filters=trim
    text observacao? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.breakglass) {
      error = "Somente Break-glass pode editar piso global"
    }
  
    precondition (($input.produto|is_empty) == false && ($input.plano|is_empty) == false) {
      error = "produto e plano obrigatorios"
    }
  
    var $valor {
      value = $input.valor_minimo|first_notempty:0
    }
  
    precondition ($valor >= 0) {
      error = "valor_minimo invalido"
    }
  
    conditional {
      if ($input.id != null && $input.id > 0) {
        db.patch fp_preco_piso_global {
          field_name = "id"
          field_value = $input.id
          data = ```
            {
              produto       : $input.produto
              plano         : $input.plano
              nome_exibicao : $input.nome_exibicao
              valor_minimo  : $valor
              ativo         : $input.ativo|first_notempty:"S"
              observacao    : $input.observacao
            }
            ```
        } as $model
      }
    
      else {
        db.query fp_preco_piso_global {
          where = $db.fp_preco_piso_global.produto == $input.produto && $db.fp_preco_piso_global.plano == $input.plano
          return = {type: "single"}
        } as $exist
      
        conditional {
          if ($exist != null) {
            db.patch fp_preco_piso_global {
              field_name = "id"
              field_value = $exist.id
              data = ```
                {
                  nome_exibicao: $input.nome_exibicao
                  valor_minimo : $valor
                  ativo        : $input.ativo|first_notempty:"S"
                  observacao   : $input.observacao
                }
                ```
            } as $model
          }
        
          else {
            db.add fp_preco_piso_global {
              data = {
                created_at   : "now"
                produto      : $input.produto
                plano        : $input.plano
                nome_exibicao: $input.nome_exibicao
                valor_minimo : $valor
                ativo        : $input.ativo|first_notempty:"S"
                observacao   : $input.observacao
              }
            } as $model
          }
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "preco_piso_global"
        id_franqueado: ""
        ref_tipo     : "fp_preco_piso_global"
        ref_id       : $model.id|to_text
        detalhe      : $input.produto ~ "/" ~ $input.plano ~ " min=" ~ ($valor|to_text)
        origem       : "admin"
        valor        : $valor
        produto      : $input.produto
        plano        : $input.plano
        admin_usuario: $admin.usuario
      }
    } as $log
  }

  response = $model
}
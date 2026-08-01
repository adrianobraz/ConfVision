// Exclui assinatura do franqueado (irreversivel)
query fp_assinatura_excluir verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int assinatura_id? filters=min:1
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    precondition ($input.assinatura_id != null) {
      error = "assinatura_id obrigatorio"
    }
  
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    } as $assinatura
  
    precondition ($assinatura != null) {
      error = "Assinatura nao encontrada"
    }
  
    function.run fn_fp_fatura_cancelar_abertas_assinatura {
      input = {
        assinatura_id: $input.assinatura_id
        observacao   : "Cancelada ao excluir assinatura — " ~ ($input.observacao|first_notempty:"admConfmonit")
        origem       : "admin"
        admin_usuario: $input.admin_usuario
      }
    } as $faturas_canceladas
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "assinatura_excluir"
        id_franqueado: $assinatura.id_franqueado
        ref_tipo     : "fp_assinatura_produto"
        ref_id       : $input.assinatura_id|to_text
        detalhe      : $input.observacao|first_notempty:"Exclusão manual admConfmonit"
        origem       : "admin"
        valor        : $assinatura.valor|first_notempty:0
        produto      : $assinatura.produto
        plano        : $assinatura.plano
        admin_usuario: $input.admin_usuario
      }
    } as $log
  
    db.del fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    }
  
    var $bundle_sync {
      value = null
    }
  
    conditional {
      if ($assinatura.produto == "franqueadopro") {
        function.run fn_fp_bundle_sincronizar {
          input = {
            id_franqueado: $assinatura.id_franqueado
            admin_usuario: $input.admin_usuario
          }
        } as $bundle_sync
      }
    }
  
    // Sempre recalcula UsaConfVision (plano CV, bundle FP ou licenca camera)
    function.run fn_franqueado_sync_usa_confvision {
      input = {id_franqueado: $assinatura.id_franqueado}
    } as $sync_cv
  
    function.run fn_franqueado_sync_usuario_terminal {
      input = {id_franqueado: $assinatura.id_franqueado}
    } as $sync_term
  }

  response = {
    ok                 : true
    excluido_id        : $input.assinatura_id
    faturas_canceladas : $faturas_canceladas.canceladas
    registro_financeiro: $log
    bundle_sync        : $bundle_sync
    sync_confvision    : $sync_cv
    sync_terminal      : $sync_term
  }
}
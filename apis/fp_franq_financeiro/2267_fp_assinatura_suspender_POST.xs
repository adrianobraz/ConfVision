// Suspender assinatura
query fp_assinatura_suspender verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int assinatura_id? filters=min:1
    text observacao? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    } as $assinatura
  
    precondition ($assinatura != null) {
      error = "Assinatura nao encontrada"
    }
  
    function.run fn_fp_admin_assert_escopo_franqueado {
      input = {
        admin_token              : $input.admin_token
        id_franqueado            : $assinatura.id_franqueado
        id_representante_registro: $assinatura.id_representante
      }
    } as $escopo
  
    db.patch fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
      data = {status: "suspensa", observacao: $input.observacao}
    } as $model
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "assinatura_suspender"
        id_franqueado: $assinatura.id_franqueado
        ref_tipo     : "fp_assinatura_produto"
        ref_id       : $input.assinatura_id|to_text
        detalhe      : $input.observacao
        origem       : "admin"
      }
    } as $log
  
    conditional {
      if ($assinatura.produto == "franqueadopro") {
        function.run fn_fp_bundle_sincronizar {
          input = {id_franqueado: $assinatura.id_franqueado}
        } as $bundle_sync
      }
    }
  
    // Recalcula UsaConfVision apos suspender FP/CV (Lite nao inclui ConfVision)
    conditional {
      if ($assinatura.produto == "franqueadopro" || $assinatura.produto == "confvision") {
        function.run fn_franqueado_sync_usa_confvision {
          input = {id_franqueado: $assinatura.id_franqueado}
        } as $sync_cv
      }
    }
  
    // Recalcula UsuarioTeminal do master (WT/TM/WA ou bundle FP)
    conditional {
      if ($assinatura.produto == "franqueadopro" || $assinatura.produto == "webterminal" || $assinatura.produto == "terminalmovel" || $assinatura.produto == "webambiente") {
        function.run fn_franqueado_sync_usuario_terminal {
          input = {id_franqueado: $assinatura.id_franqueado}
        } as $sync_term
      }
    }
  }

  response = $model
}
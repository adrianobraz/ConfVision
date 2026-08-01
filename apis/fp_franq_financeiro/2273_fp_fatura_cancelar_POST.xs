// Cancelar ou excluir fatura aberta (nao paga — nao entra no caixa)
query fp_fatura_cancelar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int fatura_id? filters=min:1
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    db.get fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
    } as $fatura
  
    precondition ($fatura != null) {
      error = "Fatura nao encontrada"
    }
  
    precondition ($fatura.status == "aberta") {
      error = "Somente faturas abertas podem ser excluidas"
    }
  
    db.query fp_pagamento {
      where = ($db.fp_pagamento.fp_fatura_id == $input.fatura_id)
      return = {type: "list"}
    } as $pagamentos
  
    precondition (($pagamentos|count) == 0) {
      error = "Fatura com pagamento registrado nao pode ser excluida"
    }
  
    db.patch fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
      data = {
        status    : "cancelada"
        observacao: $input.observacao|first_notempty:"Exclusao manual admConfmonit"
      }
    } as $model
  
    db.query fp_fatura_item {
      where = ($db.fp_fatura_item.fp_fatura_id == $input.fatura_id) && ($db.fp_fatura_item.ref_tipo == "fp_assinatura_produto")
      return = {type: "list"}
    } as $itens_assinatura
  
    foreach ($itens_assinatura) {
      each as $item {
        db.get fp_assinatura_produto {
          field_name = "id"
          field_value = $item.ref_id|to_int
        } as $assinatura
      
        conditional {
          if ($assinatura != null && $assinatura.ciclo_fatura_ref == $fatura.ciclo_ref) {
            db.patch fp_assinatura_produto {
              field_name = "id"
              field_value = $assinatura.id
              data = {ciclo_fatura_ref: ""}
            }
          }
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "fatura_cancelar"
        id_franqueado: $fatura.id_franqueado
        ref_tipo     : "fp_fatura"
        ref_id       : $input.fatura_id|to_text
        detalhe      : $input.observacao|first_notempty:"Fatura aberta excluida"
        origem       : "admin"
        valor        : $fatura.valor_total
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {fatura: $model, registro_financeiro: $log}
}
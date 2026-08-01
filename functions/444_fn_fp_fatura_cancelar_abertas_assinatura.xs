// Cancela faturas abertas vinculadas a uma assinatura (nao entra no caixa)
function fn_fp_fatura_cancelar_abertas_assinatura {
  input {
    int assinatura_id? filters=min:1
    text observacao? filters=trim
    text origem?=admin filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    db.query fp_fatura_item {
      where = $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto" && $db.fp_fatura_item.ref_id == ($input.assinatura_id|to_text)
      return = {type: "list"}
    } as $itens
  
    var $canceladas {
      value = 0
    }
  
    foreach ($itens) {
      each as $item {
        db.get fp_fatura {
          field_name = "id"
          field_value = $item.fp_fatura_id
        } as $fatura
      
        conditional {
          if ($fatura != null && $fatura.status == "aberta") {
            db.patch fp_fatura {
              field_name = "id"
              field_value = $fatura.id
              data = {
                status    : "cancelada"
                observacao: $input.observacao|first_notempty:"Cancelada automaticamente — assinatura excluida"
              }
            } as $fatura_upd
          
            var.update $canceladas {
              value = $canceladas + 1
            }
          
            function.run fn_fp_financeiro_log {
              input = {
                acao         : "fatura_cancelar"
                id_franqueado: $fatura.id_franqueado
                ref_tipo     : "fp_fatura"
                ref_id       : $fatura.id|to_text
                detalhe      : $input.observacao|first_notempty:"Fatura aberta cancelada"
                origem       : $input.origem|first_notempty:"admin"
                valor        : $fatura.valor_total
                admin_usuario: $input.admin_usuario
              }
            } as $log_item
          }
        }
      }
    }
  }

  response = {canceladas: $canceladas}
}
// Cancela cupom ativo (antes do uso) — admConfmonit
function fn_fp_cupom_cancelar {
  input {
    int cupom_id? filters=min:1
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    db.get fp_cupom_desconto {
      field_name = "id"
      field_value = $input.cupom_id
    } as $cupom
  
    precondition ($cupom != null) {
      error = "Cupom nao encontrado"
    }
  
    precondition ($cupom.status == "ativo") {
      error = "Somente cupons ativos podem ser cancelados"
    }
  
    var $obs {
      value = $cupom.observacao|first_notempty:""
    }
  
    conditional {
      if (($input.observacao|is_empty) == false) {
        conditional {
          if (($obs|is_empty) == false) {
            var.update $obs {
              value = $obs ~ " | Cancelado: " ~ $input.observacao
            }
          }
        
          else {
            var.update $obs {
              value = "Cancelado: " ~ $input.observacao
            }
          }
        }
      }
    }
  
    db.patch fp_cupom_desconto {
      field_name = "id"
      field_value = $cupom.id
      data = {status: "cancelado", observacao: $obs}
    } as $cupom_upd
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "cupom_cancelar"
        id_franqueado: $cupom.id_franqueado
        ref_tipo     : "fp_cupom_desconto"
        ref_id       : $cupom.id|to_text
        detalhe      : "codigo=" ~ ($cupom.codigo|first_notempty:"") ~ " " ~ ($input.observacao|first_notempty:"")
        origem       : "admin"
        valor        : $cupom.valor
        produto      : $cupom.produto
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {cupom: $cupom_upd, registro_financeiro: $log}
}
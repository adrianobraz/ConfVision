// Aplica cupom CEN→REP em fatura de repasse aberta
query fp_repasse_aplicar_cupom verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int fatura_id? filters=min:1
    text codigo? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    db.get fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
    } as $fatura
  
    precondition ($fatura != null) {
      error = "Fatura nao encontrada"
    }
  
    conditional {
      if ($admin.userTipo == "REP") {
        precondition ($fatura.id_representante == $admin.idVinculo) {
          error = "Repasse fora da carteira deste Representante"
        }
      }
    }
  
    function.run fn_fp_repasse_aplicar_cupom {
      input = {
        fatura_id       : $input.fatura_id
        codigo          : $input.codigo
        admin_usuario   : $input.admin_usuario|first_notempty:$admin.usuario
        id_representante: $fatura.id_representante
      }
    } as $resultado
  }

  response = $resultado
}


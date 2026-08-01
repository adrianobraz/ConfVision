// REP define preco de venda do pacote de cota (>= piso)
query fp_preco_pacote_cota_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int fp_pacote_cota_id?
    decimal valor_venda?
    text id_representante? filters=trim
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "REP") {
      error = "Somente Representante define preco de venda de cota"
    }
  
    var $id_rep {
      value = $input.id_representante
        |first_notempty:($admin.idVinculo|trim)
    }
  
    precondition (($id_rep|is_empty) == false) {
      error = "id_representante obrigatorio"
    }
  
    precondition ($input.fp_pacote_cota_id != null && $input.fp_pacote_cota_id > 0) {
      error = "fp_pacote_cota_id obrigatorio"
    }
  
    db.get fp_pacote_cota {
      field_name = "id"
      field_value = $input.fp_pacote_cota_id
    } as $pacote
  
    precondition ($pacote != null && $pacote.ativo == "S") {
      error = "Pacote de cota nao encontrado"
    }
  
    var $piso {
      value = $pacote|get:"valor":0|first_notempty:0
    }
  
    var $venda {
      value = $input.valor_venda|first_notempty:0
    }
  
    precondition ($venda >= $piso) {
      error = "valor_venda nao pode ser menor que o piso da Central (" ~ ($piso|to_text) ~ ")"
    }
  
    db.query fp_preco_pacote_cota {
      where = $db.fp_preco_pacote_cota.id_representante == $id_rep && $db.fp_preco_pacote_cota.fp_pacote_cota_id == $pacote.id
      return = {type: "single"}
    } as $existe
  
    var $salvo {
      value = null
    }
  
    conditional {
      if ($existe != null) {
        db.patch fp_preco_pacote_cota {
          field_name = "id"
          field_value = $existe.id
          data = {
            valor_venda  : $venda
            ativo        : "S"
            observacao   : $input.observacao
            atualizado_em: "now"
          }
        } as $salvo
      }
    
      else {
        db.add fp_preco_pacote_cota {
          data = {
            created_at       : "now"
            id_representante : $id_rep
            fp_pacote_cota_id: $pacote.id
            valor_venda      : $venda
            ativo            : "S"
            observacao       : $input.observacao
            atualizado_em    : "now"
          }
        } as $salvo
      }
    }
  }

  response = $salvo
}
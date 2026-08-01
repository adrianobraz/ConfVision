// Aplica cupom da Central (CEN→REP) em fatura de repasse aberta
function fn_fp_repasse_aplicar_cupom {
  input {
    int fatura_id? filters=min:1
    text codigo? filters=trim
    text admin_usuario? filters=trim
    text id_representante? filters=trim
  }

  stack {
    db.get fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
    } as $fatura
  
    precondition ($fatura != null) {
      error = "Fatura nao encontrada"
    }
  
    precondition ($fatura.tipo == "repasse_rep_central") {
      error = "Cupom da Central so se aplica a fatura de repasse REP→Central"
    }
  
    precondition ($fatura.status == "aberta") {
      error = "Somente repasse aberto pode receber cupom"
    }
  
    var $id_rep {
      value = $input.id_representante
        |first_notempty:($fatura.id_representante|first_notempty:"")
    }
  
    precondition (($id_rep|is_empty) == false) {
      error = "id_representante obrigatorio no repasse"
    }
  
    var $valor_base {
      value = $fatura.valor_total|first_notempty:0
    }
  
    precondition ($valor_base > 0) {
      error = "Repasse sem valor"
    }
  
    var $codigo {
      value = $input.codigo
        |first_notempty:""
        |trim
        |to_upper
    }
  
    db.query fp_cupom_desconto {
      where = $db.fp_cupom_desconto.codigo == $codigo
      return = {type: "single"}
    } as $cupom_peek
  
    precondition ($cupom_peek != null) {
      error = "Cupom nao encontrado"
    }
  
    var $produto_cupom {
      value = $cupom_peek.produto|first_notempty:"franqueadopro"
    }
  
    function.run fn_fp_cupom_aplicar {
      input = {
        codigo          : $codigo
        id_representante: $id_rep
        produto         : $produto_cupom
        valor_base      : $valor_base
        ref_tipo        : "fp_fatura"
        ref_id          : $fatura.id|to_text
        origem          : "admin"
        admin_usuario   : $input.admin_usuario
        contexto        : "representante"
      }
    } as $aplicado
  
    db.patch fp_fatura {
      field_name = "id"
      field_value = $fatura.id
      data = {
        valor_total: $aplicado.valor_final
        observacao : (($fatura.observacao|first_notempty:"") ~ " | Cupom CEN " ~ $codigo ~ " -R$ " ~ ($aplicado.valor_desconto|to_text))|trim
      }
    } as $fatura_upd
  
    db.add fp_fatura_item {
      data = {
        created_at    : "now"
        fp_fatura_id  : $fatura.id
        descricao     : "Cupom Central " ~ $codigo
        quantidade    : 1
        valor_unitario: (0 - $aplicado.valor_desconto)
        valor_total   : (0 - $aplicado.valor_desconto)
        ref_tipo      : "fp_cupom_desconto"
        ref_id        : $aplicado.cupom.id|to_text
      }
    } as $item_cupom
  }

  response = {
    fatura        : $fatura_upd
    valor_base    : $aplicado.valor_base
    valor_desconto: $aplicado.valor_desconto
    valor_final   : $aplicado.valor_final
    cupom         : $aplicado.cupom
  }
}
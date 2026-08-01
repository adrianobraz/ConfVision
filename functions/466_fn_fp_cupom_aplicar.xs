// Consome cupom (uso unico), calcula desconto e registra auditoria financeira
function fn_fp_cupom_aplicar {
  input {
    text codigo? filters=trim
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text produto?=franqueadopro filters=trim
    decimal valor_base?
    decimal valor_piso_minimo?
    text ref_tipo? filters=trim
    text ref_id? filters=trim
    text origem?=franqueado filters=trim
    text admin_usuario? filters=trim
    text contexto?=franqueado filters=trim
  }

  stack {
    precondition (($input.valor_base|first_notempty:0) > 0) {
      error = "valor_base obrigatorio para aplicar cupom"
    }
  
    function.run fn_fp_cupom_validar {
      input = {
        codigo            : $input.codigo
        id_franqueado     : $input.id_franqueado
        id_representante  : $input.id_representante
        produto           : $input.produto
        valor_base        : $input.valor_base
        valor_piso_minimo : $input.valor_piso_minimo
        contexto          : $input.contexto|first_notempty:"franqueado"
      }
    } as $validacao
  
    var $cupom {
      value = $validacao.cupom
    }
  
    var $calc {
      value = $validacao.calculo
    }
  
    db.patch fp_cupom_desconto {
      field_name = "id"
      field_value = $cupom.id
      data = {
        status                 : "usado"
        usado_em               : "now"
        usado_por_id_franqueado: $input.id_franqueado
        ref_tipo               : $input.ref_tipo
        ref_id                 : $input.ref_id
        valor_base             : $calc.valor_base
        valor_desconto         : $calc.valor_desconto
        valor_final            : $calc.valor_final
      }
    } as $cupom_upd
  
    var $detalhe {
      value = "codigo=" ~ $validacao.codigo ~ " base=" ~ ($calc.valor_base|to_text) ~ " desconto=" ~ ($calc.valor_desconto|to_text) ~ " final=" ~ ($calc.valor_final|to_text)
    }
  
    conditional {
      if (($input.ref_tipo|is_empty) == false) {
        var.update $detalhe {
          value = $detalhe ~ " " ~ $input.ref_tipo ~ "=" ~ ($input.ref_id|first_notempty:"")
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "cupom_aplicar"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_cupom_desconto"
        ref_id       : $cupom.id|to_text
        detalhe      : $detalhe
        origem       : $input.origem|first_notempty:"franqueado"
        valor        : $calc.valor_desconto
        produto      : $input.produto
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {
    cupom              : $cupom_upd
    valor_base         : $calc.valor_base
    valor_desconto     : $calc.valor_desconto
    valor_final        : $calc.valor_final
    registro_financeiro: $log
  }
}
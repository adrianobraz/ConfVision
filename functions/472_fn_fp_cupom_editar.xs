// Edita cupom ativo (ainda nao usado) — admConfmonit
function fn_fp_cupom_editar {
  input {
    int cupom_id? filters=min:1
    text codigo? filters=trim
    text tipo?=percentual filters=trim
    decimal valor?
    text produto?=franqueadopro filters=trim
    text id_franqueado? filters=trim
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
      error = "Somente cupons ativos (nao usados) podem ser editados"
    }
  
    var $codigo {
      value = $input.codigo
        |first_notempty:""
        |trim
        |to_upper
    }
  
    precondition (($codigo|is_empty) == false) {
      error = "codigo obrigatorio"
    }
  
    var $tipo {
      value = $input.tipo
        |first_notempty:"percentual"
        |to_lower
    }
  
    precondition ($tipo == "percentual" || $tipo == "valor_fixo") {
      error = "tipo deve ser percentual ou valor_fixo"
    }
  
    var $valor {
      value = $input.valor|first_notempty:0
    }
  
    precondition ($valor > 0) {
      error = "valor do cupom deve ser maior que zero"
    }
  
    conditional {
      if ($tipo == "percentual") {
        precondition ($valor <= 100) {
          error = "percentual nao pode ser maior que 100"
        }
      }
    }
  
    var $produto {
      value = $input.produto
        |first_notempty:"franqueadopro"
        |to_lower
    }
  
    db.query fp_cupom_desconto {
      where = $db.fp_cupom_desconto.codigo == $codigo && $db.fp_cupom_desconto.id != $cupom.id
      return = {type: "exists"}
    } as $existe
  
    precondition ($existe == null) {
      error = "Ja existe outro cupom com este codigo"
    }
  
    db.patch fp_cupom_desconto {
      field_name = "id"
      field_value = $cupom.id
      data = {
        codigo       : $codigo
        tipo         : $tipo
        valor        : $valor
        produto      : $produto
        id_franqueado: $input.id_franqueado
        observacao   : $input.observacao
      }
    } as $cupom_upd
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "cupom_editar"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_cupom_desconto"
        ref_id       : $cupom.id|to_text
        detalhe      : "codigo=" ~ $codigo ~ " tipo=" ~ $tipo ~ " valor=" ~ ($valor|to_text) ~ " produto=" ~ $produto
        origem       : "admin"
        valor        : $valor
        produto      : $produto
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {cupom: $cupom_upd, registro_financeiro: $log}
}
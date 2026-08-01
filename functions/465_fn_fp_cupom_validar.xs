// Valida cupom — hierarquia CEN→REP / REP→FRA. FRA: so cupom do seu Representante; piso nao pode ser furado.
function fn_fp_cupom_validar {
  input {
    text codigo? filters=trim
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text produto?=franqueadopro filters=trim
    decimal valor_base?
    decimal valor_piso_minimo?
    text contexto?=franqueado filters=trim
  }

  stack {
    var $codigo {
      value = $input.codigo
        |first_notempty:""
        |trim
        |to_upper
    }
  
    precondition (($codigo|is_empty) == false) {
      error = "codigo do cupom obrigatorio"
    }
  
    var $produto {
      value = $input.produto
        |first_notempty:"franqueadopro"
        |to_lower
    }
  
    var $contexto {
      value = $input.contexto|first_notempty:"franqueado"|to_lower
    }
  
    // Beneficio Pro+ do ConfVision e independente — cupom nao interfere
    conditional {
      if ($produto == "confvision" && (($input.id_franqueado|is_empty) == false)) {
        function.run fn_fp_assinatura_get_ativa {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "franqueadopro"
          }
        } as $fp
      
        conditional {
          if ($fp.liberado && $fp.assinatura != null && $fp.assinatura.plano == "pro_plus") {
            precondition (false) {
              error = "Cupom nao se aplica a licencas ConfVision para assinantes Pro+. O desconto de 20% do plano ja esta ativo."
            }
          }
        }
      }
    }
  
    db.query fp_cupom_desconto {
      where = $db.fp_cupom_desconto.codigo == $codigo
      return = {type: "single"}
    } as $cupom
  
    precondition ($cupom != null) {
      error = "Cupom nao encontrado"
    }
  
    precondition ($cupom.status == "ativo") {
      error = "Cupom indisponivel (status: " ~ ($cupom.status|first_notempty:"?") ~ ")"
    }
  
    var $cupom_produto {
      value = $cupom.produto
        |first_notempty:"franqueadopro"
        |to_lower
    }
  
    conditional {
      if ($cupom_produto != "todos" && $cupom_produto != $produto) {
        precondition (false) {
          error = "Cupom nao vale para o produto " ~ $produto
        }
      }
    }
  
    var $criado_por_tipo {
      value = $cupom.criado_por_tipo|first_notempty:""|to_upper
    }
  
    var $alvo {
      value = $cupom.alvo_nivel|first_notempty:""|to_upper
    }
  
    // Legado sem hierarquia: trata como REP→FRA
    conditional {
      if ($criado_por_tipo|is_empty) {
        var.update $criado_por_tipo {
          value = "REP"
        }
      
        var.update $alvo {
          value = "FRA"
        }
      }
    }
  
    conditional {
      if ($alvo|is_empty) {
        conditional {
          if ($criado_por_tipo == "CEN") {
            var.update $alvo {
              value = "REP"
            }
          }
        
          else {
            var.update $alvo {
              value = "FRA"
            }
          }
        }
      }
    }
  
    var $id_rep_ctx {
      value = $input.id_representante|first_notempty:""
    }
  
    conditional {
      if (($id_rep_ctx|is_empty) && (($input.id_franqueado|is_empty) == false)) {
        function.run fn_fp_franqueado_id_representante {
          input = {id_franqueado: $input.id_franqueado}
        } as $rep_res
      
        var.update $id_rep_ctx {
          value = $rep_res.id_representante|first_notempty:""
        }
      }
    }
  
    // Checkout franqueado: so cupom REP→FRA do seu representante
    conditional {
      if ($contexto == "franqueado" || $contexto == "fra") {
        precondition ($alvo == "FRA" && $criado_por_tipo == "REP") {
          error = "Este cupom nao e para franqueado (hierarquia: Representante → Franqueado)"
        }
      
        conditional {
          if (($cupom.id_representante|is_empty) == false) {
            precondition ($cupom.id_representante == $id_rep_ctx) {
              error = "Cupom nao liberado para a carteira deste franqueado"
            }
          }
        }
      
        conditional {
          if (($cupom.id_franqueado|is_empty) == false) {
            precondition ($cupom.id_franqueado == $input.id_franqueado) {
              error = "Cupom nao liberado para este franqueado"
            }
          }
        }
      }
    }
  
    // Contexto representante: so cupom CEN→REP
    conditional {
      if ($contexto == "representante" || $contexto == "rep") {
        precondition ($alvo == "REP" && $criado_por_tipo == "CEN") {
          error = "Este cupom nao e para representante (hierarquia: Central → Representante)"
        }
      
        conditional {
          if (($cupom.id_representante|is_empty) == false) {
            precondition ($cupom.id_representante == $id_rep_ctx) {
              error = "Cupom da Central nao liberado para este representante"
            }
          }
        }
      }
    }
  
    var $calc {
      value = null
    }
  
    conditional {
      if ($input.valor_base != null && $input.valor_base > 0) {
        function.run fn_fp_cupom_calcular {
          input = {
            tipo       : $cupom.tipo
            valor_cupom: $cupom.valor
            valor_base : $input.valor_base
          }
        } as $calc
      
        // REP→FRA: nao pode furar o piso da Central
        conditional {
          if ($alvo == "FRA" && $input.valor_piso_minimo != null && $input.valor_piso_minimo > 0) {
            precondition ($calc.valor_final >= $input.valor_piso_minimo) {
              error = "Desconto do Representante nao pode deixar o valor abaixo do piso da Central (R$ " ~ ($input.valor_piso_minimo|to_text) ~ ")"
            }
          }
        }
      }
    }
  }

  response = {
    ok            : true
    cupom         : $cupom
    codigo        : $codigo
    produto       : $produto
    calculo       : $calc
    valor_base    : $calc.valor_base
    valor_desconto: $calc.valor_desconto
    valor_final   : $calc.valor_final
    criado_por_tipo: $criado_por_tipo
    alvo_nivel    : $alvo
  }
}

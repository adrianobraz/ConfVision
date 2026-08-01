// Calcula valor da licenca ConfVision com desconto 20% para FranqueadoPro Pro+ ativo
function fn_fp_confvision_valor_com_desconto {
  input {
    text id_franqueado? filters=trim
    decimal valor_base?
  }

  stack {
    var $valor {
      value = $input.valor_base|first_notempty:0
    }
  
    var $desconto_pct {
      value = 0
    }
  
    var $desconto_aplicado {
      value = false
    }
  
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : "franqueadopro"
      }
    } as $fp
  
    conditional {
      if ($fp.liberado && $fp.assinatura != null && $fp.assinatura.plano == "pro_plus") {
        var.update $desconto_pct {
          value = 20
        }
      
        var.update $desconto_aplicado {
          value = true
        }
      
        var.update $valor {
          value = ($valor * 0.8)|round:2
        }
      }
    }
  }

  response = {
    valor            : $valor
    valor_base       : $input.valor_base
    desconto_pct     : $desconto_pct
    desconto_aplicado: $desconto_aplicado
  }
}
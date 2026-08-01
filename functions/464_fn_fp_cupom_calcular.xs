// Calcula valor com desconto de cupom (nao consome o cupom)
function fn_fp_cupom_calcular {
  input {
    text tipo?=percentual filters=trim
    decimal valor_cupom?
    decimal valor_base?
  }

  stack {
    var $base {
      value = $input.valor_base|first_notempty:0
    }
  
    precondition ($base > 0) {
      error = "valor_base deve ser maior que zero"
    }
  
    var $tipo {
      value = $input.tipo
        |first_notempty:"percentual"
        |to_lower
    }
  
    var $cupom_val {
      value = $input.valor_cupom|first_notempty:0
    }
  
    precondition ($cupom_val > 0) {
      error = "valor do cupom deve ser maior que zero"
    }
  
    var $desconto {
      value = 0
    }
  
    conditional {
      if ($tipo == "percentual") {
        precondition ($cupom_val <= 100) {
          error = "desconto percentual nao pode ser maior que 100"
        }
      
        var.update $desconto {
          value = ($base * $cupom_val / 100)|round:2
        }
      }
    
      elseif ($tipo == "valor_fixo") {
        var.update $desconto {
          value = $cupom_val|round:2
        }
      }
    
      else {
        precondition (false) {
          error = "tipo de cupom invalido (percentual ou valor_fixo)"
        }
      }
    }
  
    conditional {
      if ($desconto > $base) {
        var.update $desconto {
          value = $base
        }
      }
    }
  
    var $final {
      value = ($base - $desconto)|round:2
    }
  
    conditional {
      if ($final < 0) {
        var.update $final {
          value = 0
        }
      }
    }
  }

  response = {
    valor_base    : $base
    valor_desconto: $desconto
    valor_final   : $final
    tipo          : $tipo
  }
}
// Calcula proxima data de cobranca a partir da periodicidade
function fn_fp_calcular_proxima_cobranca {
  input {
    timestamp? base_em?
    text periodicidade? filters=trim
  }

  stack {
    var $base {
      value = $input.base_em
    }
  
    conditional {
      if ($base == null) {
        var.update $base {
          value = now
        }
      }
    }
  
    var $periodo {
      value = $input.periodicidade|first_notempty:"mensal"
    }
  
    var $dias {
      value = 30
    }
  
    conditional {
      if ($periodo == "trimestral") {
        var.update $dias {
          value = 90
        }
      }
    
      elseif ($periodo == "semestral") {
        var.update $dias {
          value = 180
        }
      }
    
      elseif ($periodo == "anual") {
        var.update $dias {
          value = 365
        }
      }
    }
  
    var $proxima {
      value = $base
        |add_secs_to_timestamp:$dias * 86400
    }
  }

  response = {proxima_cobranca_em: $proxima, dias_periodo: $dias}
}
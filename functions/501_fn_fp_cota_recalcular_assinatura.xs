// Soma cotas_json da assinatura → quantidade_total, limites e valor_cotas
function fn_fp_cota_recalcular_assinatura {
  input {
    json cotas_json?
    decimal valor_licenca?
  }

  stack {
    var $itens {
      value = $input.cotas_json|first_notempty:[]
    }
  
    var $qtd_total {
      value = 0
    }
  
    var $valor_cotas {
      value = 0
    }
  
    foreach ($itens) {
      each as $c {
        var.update $qtd_total {
          value = $qtd_total + ($c.quantidade|first_notempty:0)
        }
      
        var.update $valor_cotas {
          value = $valor_cotas + ($c.valor|first_notempty:0)
        }
      }
    }
  
    function.run fn_fp_cota_limites_de_quantidade {
      input = {quantidade: $qtd_total}
    } as $lim
  
    var $valor_total {
      value = ($input.valor_licenca|first_notempty:0) + $valor_cotas
    }
  }

  response = {
    cotas_json      : $itens
    quantidade_total: $qtd_total
    valor_cotas     : $valor_cotas
    valor_licenca   : $input.valor_licenca|first_notempty:0
    valor_mensal    : $valor_total
    limites_json    : $lim.limites_json
  }
}
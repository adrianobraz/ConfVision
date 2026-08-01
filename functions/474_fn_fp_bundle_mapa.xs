// Retorna produtos/planos inclusos no plano FranqueadoPro (bundle) — fonte canonica
function fn_fp_bundle_mapa {
  input {
    text plano_fp? filters=trim
  }

  stack {
    var $plano {
      value = $input.plano_fp|first_notempty:""|to_lower
    }
  
    var $itens {
      value = []
    }
  
    conditional {
      if ($plano == "pro") {
        var.update $itens {
          value = [
            {produto: "webterminal", plano: "lite"}
            {produto: "terminalmovel", plano: "padrao"}
          ]
        }
      }
    
      elseif ($plano == "pro_plus") {
        var.update $itens {
          value = [
            {produto: "webterminal", plano: "pro"}
            {produto: "terminalmovel", plano: "padrao"}
            {produto: "confvision", plano: "padrao"}
            {produto: "webambiente", plano: "pro"}
            {produto: "dialyze", plano: "padrao"}
          ]
        }
      }
    }
  }

  response = {plano_fp: $plano, itens: $itens, total: $itens|count}
}
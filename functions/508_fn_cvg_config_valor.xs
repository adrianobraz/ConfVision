// Le valor de configuracao do modulo grade (fp_config_financeiro)
function fn_cvg_config_valor {
  input {
    text chave? filters=trim
    text fallback? filters=trim
  }

  stack {
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = $input.chave
    } as $cfg

    var $valor {
      value = $input.fallback|first_notempty:""
    }

    conditional {
      if ($cfg != null && ($cfg.valor|is_empty) == false) {
        var.update $valor {
          value = $cfg.valor
        }
      }
    }
  }

  response = $valor
}

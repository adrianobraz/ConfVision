// Valida header X-Worker-Key contra fp_config_financeiro
function fn_fp_worker_validar {
  input {
    text worker_key? filters=trim
  }

  stack {
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "worker_secret"
    } as $cfg
  
    var $esperado {
      value = ""
    }
  
    conditional {
      if ($cfg != null) {
        var.update $esperado {
          value = $cfg.valor
        }
      }
    }
  
    precondition (($esperado|is_empty) == false) {
      error = "worker_secret nao configurado"
    }
  
    precondition ($input.worker_key == $esperado) {
      error = "Worker key invalida"
    }
  }

  response = {ok: true}
}
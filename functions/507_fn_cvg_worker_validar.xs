// Valida header X-CVG-Worker-Key contra fp_config_financeiro (chave cvg_worker_secret)
function fn_cvg_worker_validar {
  input {
    text worker_key? filters=trim
  }

  stack {
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "cvg_worker_secret"
    } as $cfg

    var $esperado {
      value = ""
    }

    conditional {
      if ($cfg != null) {
        var.update $esperado {
          value = $cfg.valor|first_notempty:""
        }
      }
    }

    precondition (($esperado|is_empty) == false) {
      error = "cvg_worker_secret nao configurado em fp_config_financeiro"
    }

    precondition ($input.worker_key == $esperado) {
      error = "Worker key invalida"
    }
  }

  response = {ok: true}
}

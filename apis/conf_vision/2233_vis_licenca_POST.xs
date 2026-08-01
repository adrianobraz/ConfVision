// Criar licenca prepaga (admin / financeiro)
query vis_licenca verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_licenca"
    }
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.plano|is_empty) == false) {
      error = "plano obrigatorio"
    }
  
    function.run fn_vis_plano_flags {
      input = {plano: $input.plano}
    } as $flags
  
    precondition ($flags.plano_label != "Nenhum") {
      error = "Plano invalido. Use slug valido (ex.: online, sensor_foto, analitico_24h_foto_video, gravacao_7d, gravacao_15d, gravacao_30d, gravacao_movimento_7d, gravacao_movimento_15d, gravacao_movimento_30d)"
    }
  
    var $valor {
      value = $input.valor
    }
  
    conditional {
      if ($valor == null || $valor == 0) {
        var.update $valor {
          value = $flags.valor
        }
      }
    }
  
    var $pago_em {
      value = $input.pago_em
    }
  
    conditional {
      if ($pago_em == null) {
        var.update $pago_em {
          value = now
        }
      }
    }
  
    var $unidade {
      value = $input.unidade
    }
  
    conditional {
      if ($unidade|is_empty) {
        var.update $unidade {
          value = $flags.unidade
        }
      }
    }
  
    db.add vis_licenca {
      enforce_hidden_fields = false
      data = {
        created_at    : "now"
        id_franqueado : $input.id_franqueado
        plano         : $input.plano
        unidade       : $unidade
        valor         : $valor
        pago_em       : $pago_em
        valido_ate    : $input.valido_ate
        status        : "disponivel"
        id_dispositivo: $input.id_dispositivo
        id_fatura     : $input.id_fatura
        id_pagamento  : $input.id_pagamento
        observacao    : $input.observacao
      }
    } as $model
  }

  response = $model
}
query "autofim/success" verb=POST {
  api_group = "task"

  input {
    int id
    text lock_token
    text motivo_final?
    text acao_final?
    json payload_resultado?
  }

  stack {
    precondition (($input.id|is_empty) == false) {
      error = "id obrigatório"
      payload = false
    }
  
    precondition (($input.lock_token|trim|is_empty) == false) {
      error = "lock_token obrigatório"
      payload = false
    }
  
    db.get bot_finalizaeventoauto {
      field_name = "id"
      field_value = $input.id
    } as $row
  
    precondition (($row.id|is_empty) == false) {
      error = "registro não encontrado"
      payload = false
    }
  
    var $ok {
      value = true
    }
  
    var $skipped {
      value = false
    }
  
    var $reason {
      value = ""
    }
  
    // Já finalizado? tratar como sucesso idempotente
    conditional {
      if ($row.status == "FINALIZADO") {
        var.update $skipped {
          value = true
        }
      
        var.update $reason {
          value = "already_finalized"
        }
      }
    }
  
    // Só atualiza se estiver PROCESSANDO e lock_token bater
    conditional {
      if (($skipped == false) && ($row.status == "PROCESSANDO") && ($row.lock_token == $input.lock_token)) {
        db.edit bot_finalizaeventoauto {
          field_name = "id"
          field_value = $row.id
          enforce_hidden_fields = false
          data = {
            status           : "FINALIZADO"
            motivo_final     : $input.motivo_final
            acao_final       : $input.acao_final
            payload_resultado: $input.payload_resultado
            erro             : ""
            lock_token       : ""
            lock_at          : null
            updated_at       : "now"
          }
        } as $updated
      }
    
      elseif ($skipped == false) {
        var.update $skipped {
          value = true
        }
      
        var.update $reason {
          value = "status_or_lock_mismatch"
        }
      }
    }
  }

  response = {
    dados: ""|set:"ok":$ok|set:"id":$input.id|set:"skipped":$skipped|set:"reason":$reason
  }
}
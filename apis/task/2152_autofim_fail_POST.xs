query "autofim/fail" verb=POST {
  api_group = "task"

  input {
    int id
    text lock_token
    text erro?
    int retry_delay_seconds?
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
  
    precondition ($row.status == "PROCESSANDO") {
      error = "status inválido para fail"
      payload = false
    }
  
    precondition ($row.lock_token == $input.lock_token) {
      error = "lock_token inválido"
      payload = false
    }
  
    var $nextTentativas {
      value = ($row.tentativas + 0) + 1
    }
  
    var $delaySec {
      value = 15
    }
  
    conditional {
      if (($input.retry_delay_seconds|is_empty) == false && $input.retry_delay_seconds > 0) {
        var.update $delaySec {
          value = $input.retry_delay_seconds
        }
      }
    
      elseif ($nextTentativas == 2) {
        var.update $delaySec {
          value = 30
        }
      }
    
      elseif ($nextTentativas == 3) {
        var.update $delaySec {
          value = 60
        }
      }
    
      elseif ($nextTentativas >= 4) {
        var.update $delaySec {
          value = 120
        }
      }
    }
  
    // Agora em milissegundos
    var $agoraMs {
      value = "now"|to_ms
    }
  
    var $rodarEmMs {
      value = $agoraMs + (($delaySec + 0) * 1000)
    }
  
    db.edit bot_finalizaeventoauto {
      field_name = "id"
      field_value = $row.id
      enforce_hidden_fields = false
      data = {
        status    : "PENDENTE"
        tentativas: $nextTentativas
        erro      : $input.erro|default:""
        lock_token: ""
        lock_at   : null
        rodar_em  : $rodarEmMs
        updated_at: "now"
      }
    } as $updated
  }

  response = {
    dados: ""|set:"ok":true|set:"id":$updated.id|set:"status":$updated.status|set:"tentativas":$updated.tentativas|set:"rodar_em":$updated.rodar_em
  }
}
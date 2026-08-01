query "autofim/claim" verb=POST {
  api_group = "task"

  input {
    int id
    text lock_token
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
  
    var $agoraMs {
      value = ("now"|to_timestamp) * 1000
    }
  
    db.query bot_finalizaeventoauto {
      where = $db.bot_finalizaeventoauto.id == $input.id && $db.bot_finalizaeventoauto.status == "PENDENTE" && $db.bot_finalizaeventoauto.rodar_em <= $agoraMs
      return = {type: "single"}
    } as $cand
  
    var $claimed {
      value = false
    }
  
    var $row {
      value = {}
    }
  
    conditional {
      if (($cand.id|is_empty) == false) {
        db.edit bot_finalizaeventoauto {
          field_name = "id"
          field_value = $cand.id
          enforce_hidden_fields = false
          data = {
            status    : "PROCESSANDO"
            lock_token: $input.lock_token
            lock_at   : "now"
            updated_at: "now"
            erro      : ""
          }
        } as $upd
      
        var.update $claimed {
          value = true
        }
      
        var.update $row {
          value = $upd
        }
      }
    }
  }

  response = {dados: ""|set:"claimed":$claimed|set:"row":$row}
}
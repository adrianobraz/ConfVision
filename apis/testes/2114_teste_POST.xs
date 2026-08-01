query teste verb=POST {
  api_group = "testes"

  input {
    int whatsappeventocad_id?
    enum[] dias? {
      values = ["0", "1", "2", "3", "4", "5", "6", "7"]
    }
  
    text hora_inicio? filters=trim
    text hora_fim? filters=trim
  }

  stack {
    var $x1 {
      value = now
    }
  
    var $x2 {
      value = ```
        ($x1|format_timestamp:"N":"America/Sao_Paulo") ~ " - " ~
        ($x1|to_timestamp|format_timestamp:"H:i":"America/Sao_Paulo")
        ```
    }
  
    !db.add WhatsappEventCadJanela {
      enforce_hidden_fields = false
      data = {
        whatsappeventocad_id: $input.whatsappeventocad_id
        dias                : $input.dias
        hora_inicio         : $input.hora_inicio
        hora_fim            : $input.hora_fim
      }
    } as $WhatsappEventCadJanela1
  
    db.query task {
      where = $db.task.Running == true && $db.task.update < (now|timestamp_add_minutes:-30)
      return = {type: "single"}
    } as $task1
  
    conditional {
      if (($task1|is_empty) == false) {
        db.edit task {
          field_name = "id"
          field_value = $task1.id
          enforce_hidden_fields = false
          data = {Running: false, update: now}
        } as $task2
      }
    }
  }

  response = {dados: $task1}
}
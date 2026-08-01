query retornaLink verb=POST {
  api_group = "n8nBya"

  input {
    text grupo? filters=trim
    text idDispositivo? filters=trim
    text particao? filters=trim
    text zonaUser? filters=trim
  }

  stack {
    db.query status_link {
      where = $db.status_link.idDispositivo == $input.idDispositivo && $db.status_link.particao == $input.particao && $db.status_link.zonauser == $input.zonaUser && $db.status_link.expiresAt > now && $db.status_link.grupo == "ALARME"
      sort = {status_link.id: "desc", status_link.created_at: "desc"}
      return = {type: "single"}
    } as $status_link2
  
    conditional {
      if ($status_link2|is_empty) {
        function.run statuslink_reUUID {
          input = {dispositivo: $input.idDispositivo}
        } as $func1
      
        conditional {
          if ($func1|is_empty) {
            db.add status_link {
              enforce_hidden_fields = false
              data = {
                token          : `|uuid`
                idDispositivo  : $input.idDispositivo
                expiresAt      : now|add_secs_to_timestamp:300
                alarm_events_id: $input.id_tblAlarm_events
                particao       : $input.particao
                zonauser       : $input.zonaUser
                grupo          : "ALARME"
              }
            } as $status_link1
          }
        
          else {
            db.edit status_link {
              field_name = "id"
              field_value = $func1.id
              enforce_hidden_fields = false
              data = {
                created_at     : now
                expiresAt      : now|add_secs_to_timestamp:7200
                alarm_events_id: $input.id_tblAlarm_events
                particao       : $input.particao
                zonauser       : $input.zonaUser
                grupo          : "ALARME"
              }
            } as $status_link4
          }
        }
      }
    
      else {
        db.edit status_link {
          field_name = "id"
          field_value = $status_link2.id
          enforce_hidden_fields = false
          data = {expiresAt: now|add_secs_to_timestamp:300}
        } as $status_link3
      }
    }
  }

  response = null
}
query Data_hora_Evento verb=GET {
  api_group = "GraficoEventos"

  input {
    // 15/08/2025 13:50:00
    text? Data? filters=trim
  
    int? Dias?=0
  }

  stack {
    var $nData {
      value = now
    }
  
    conditional {
      if (($input.Data|is_empty) == false) {
        var.update $nData {
          value = $input.Data
            |parse_timestamp:"d/m/Y H:i:s":"America/Sao_Paulo"
        }
      }
    }
  
    db.query alarm_events {
      where = $db.alarm_events.created_at >= ($nData|timestamp_subtract_days:$input.Dias) && $db.alarm_events.created_at <= $nData
      sort = {DataEvento: "asc"}
      eval = {
        hora: $db.alarm_events.created_at|timestamp_hour:"America/Sao_Paulo"
      }
    
      return = {
        type : "aggregate"
        group: {
          DataEvento            : $db.alarm_events.Data
          hora1                 : $db.alarm_events.created_at|timestamp_hour:"America/Sao_Paulo"
          alarm_events_ctiGrupo1: $db.alarm_events.ctiGrupo
        }
        eval : {Total: $db.alarm_events.ctiGrupo|count}
      }
    } as $alarm_events4
  }

  response = $alarm_events4
}
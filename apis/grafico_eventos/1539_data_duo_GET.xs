query Data_Duo verb=GET {
  api_group = "GraficoEventos"

  input {
    // 15/08/2025 13:50:00
    text DataInicial? filters=trim
  
    text DataFinal? filters=trim
  }

  stack {
    var $nDataIni {
      value = now
    }
  
    var $nDataFin {
      value = now
    }
  
    conditional {
      if (($input.DataInicial|is_empty) == false) {
        var.update $nDataIni {
          value = $input.DataInicial
            |parse_timestamp:"d/m/Y H:i:s":"America/Sao_Paulo"
        }
      }
    }
  
    conditional {
      if (($input.DataFinal|is_empty) == false) {
        var.update $nDataFin {
          value = $input.DataFinal
            |parse_timestamp:"d/m/Y H:i:s":"America/Sao_Paulo"
        }
      }
    }
  
    db.query alarm_events {
      where = $db.alarm_events.created_at >= $nDataIni && $db.alarm_events.created_at <= $nDataFin
      sort = {DataEvento: "asc"}
      return = {
        type : "aggregate"
        group: {DataEvento: $db.alarm_events.Data}
        eval : {Total: $db.alarm_events.Data|count}
      }
    } as $alarm_events2
  }

  response = {data: $alarm_events2}
  cache = {
    ttl       : 3600
    input     : true
    auth      : true
    datasource: true
    ip        : false
    headers   : []
    env       : []
  }
}
query Data verb=GET {
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
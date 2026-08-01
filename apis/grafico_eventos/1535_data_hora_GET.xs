query Data_Hora verb=GET {
  api_group = "GraficoEventos"

  input {
    date? data?
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.Data == $input.data
      sort = {hora: "asc"}
      eval = {
        hora: $db.alarm_events.created_at|timestamp_hour:"America/Sao_Paulo"
      }
    
      return = {
        type : "aggregate"
        group: {hora: $db.alarm_events.created_at|timestamp_hour:"UTC"}
        eval : {Total: $db.hora|count}
      }
    } as $alarm_events3
  }

  response = {data: $alarm_events3}
}
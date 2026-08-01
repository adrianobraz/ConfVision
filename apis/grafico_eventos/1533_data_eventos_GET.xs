query Data_Eventos verb=GET {
  api_group = "GraficoEventos"

  input {
    date? data?
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.Data == $input.data
      sort = {Grupo: "asc"}
      return = {
        type : "aggregate"
        group: {
          Data : $db.alarm_events.Data
          Grupo: $db.alarm_events.ctiGrupo
        }
        eval : {Total: $db.alarm_events.ctiGrupo|count}
      }
    } as $alarm_events1
  }

  response = {data: $alarm_events1}
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
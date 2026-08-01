function func_eventos_terminal_atendimento {
  input {
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.nivel > 0
      sort = {alarm_events_idProcesso1: "desc"}
      return = {
        type  : "aggregate"
        paging: {page: 1, per_page: 25}
        group : {
          alarm_events_idProcesso1 : $db.alarm_events.idProcesso
          alarm_events_nomeCliente1: $db.alarm_events.nomeCliente
        }
      }
    } as $alarm_events1
  }

  response = {data: $alarm_events1}
}
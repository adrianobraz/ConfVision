// Query all alarmEvent_Finalizados records
query alarmevent_finalizados verb=GET {
  api_group = "RoboAtendimento"

  input {
  }

  stack {
    db.query alarmEvent_Finalizados {
      return = {type: "list"}
    } as $alarmevent_finalizados
  }

  response = $alarmevent_finalizados
}